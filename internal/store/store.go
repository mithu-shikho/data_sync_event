// Package store provides control-MongoDB persistence for connections,
// watches, resume tokens, and capped operational logs.
package store

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"data_sync_event/internal/config"
)

// Collection names in the control database.
const (
	collConnections  = "connections"
	collWatches      = "watches"
	collResumeTokens = "resume_tokens"
	collEventsLog    = "events_log"
	collErrorLogs    = "error_logs"
)

// Capped collection sizes (bytes) and document caps.
const (
	eventsLogMaxBytes = 256 << 20 // 256 MiB
	eventsLogMaxDocs  = 100_000
	errorLogsMaxBytes = 64 << 20 // 64 MiB
	errorLogsMaxDocs  = 50_000
)

// Store is the control-plane persistence layer.
type Store struct {
	client *mongo.Client
	db     *mongo.Database
	enc    *encryptor
}

// New connects to the control MongoDB, verifies connectivity, and ensures
// indexes + capped collections exist.
func New(ctx context.Context, cfg *config.Config) (*Store, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.ControlMongoURI))
	if err != nil {
		return nil, fmt.Errorf("connect control mongo: %w", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("ping control mongo: %w", err)
	}

	enc, err := newEncryptor(cfg.URIEncryptionKey)
	if err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	s := &Store{
		client: client,
		db:     client.Database(cfg.ControlDBName),
		enc:    enc,
	}
	if err := s.ensure(ctx); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	return s, nil
}

// Close disconnects the underlying client.
func (s *Store) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}

// EncryptionEnabled reports whether connection URIs are encrypted at rest.
func (s *Store) EncryptionEnabled() bool { return s.enc != nil }

// ensure creates indexes and capped collections (idempotent).
func (s *Store) ensure(ctx context.Context) error {
	// Unique connection name.
	if _, err := s.db.Collection(collConnections).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("index connections.name: %w", err)
	}

	// Watches: lookups by connection, and uniqueness per (conn, db, coll).
	if _, err := s.db.Collection(collWatches).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "connection_id", Value: 1}}},
		{
			Keys:    bson.D{{Key: "connection_id", Value: 1}, {Key: "db", Value: 1}, {Key: "collection", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{Keys: bson.D{{Key: "enabled", Value: 1}}},
	}); err != nil {
		return fmt.Errorf("index watches: %w", err)
	}

	if err := s.ensureCapped(ctx, collEventsLog, eventsLogMaxBytes, eventsLogMaxDocs); err != nil {
		return err
	}
	if err := s.ensureCapped(ctx, collErrorLogs, errorLogsMaxBytes, errorLogsMaxDocs); err != nil {
		return err
	}
	return nil
}

func (s *Store) ensureCapped(ctx context.Context, name string, maxBytes, maxDocs int64) error {
	names, err := s.db.ListCollectionNames(ctx, bson.M{"name": name})
	if err != nil {
		return fmt.Errorf("list collections: %w", err)
	}
	if len(names) > 0 {
		return nil // already exists; leave as-is
	}
	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(maxBytes).SetMaxDocuments(maxDocs)
	if err := s.db.CreateCollection(ctx, name, opts); err != nil {
		return fmt.Errorf("create capped %s: %w", name, err)
	}
	return nil
}

func now() time.Time { return time.Now().UTC() }
