package store

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"data_sync_event/internal/model"
)

// ErrNotFound is returned when a document does not exist.
var ErrNotFound = errors.New("not found")

// CreateConnection encrypts the URI and inserts the connection.
func (s *Store) CreateConnection(ctx context.Context, c *model.Connection) error {
	enc, err := s.enc.seal(c.URI)
	if err != nil {
		return fmt.Errorf("encrypt uri: %w", err)
	}
	c.EncURI = enc
	c.Status = model.StatusPending
	c.CreatedAt = now()
	c.UpdatedAt = c.CreatedAt

	res, err := s.col(collConnections).InsertOne(ctx, c)
	if err != nil {
		return fmt.Errorf("insert connection: %w", err)
	}
	c.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// GetConnection loads a connection and decrypts its URI.
func (s *Store) GetConnection(ctx context.Context, id primitive.ObjectID) (*model.Connection, error) {
	var c model.Connection
	err := s.col(collConnections).FindOne(ctx, bson.M{"_id": id}).Decode(&c)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if c.URI, err = s.enc.open(c.EncURI); err != nil {
		return nil, fmt.Errorf("decrypt uri: %w", err)
	}
	return &c, nil
}

// ListConnections returns all connections (URIs decrypted).
func (s *Store) ListConnections(ctx context.Context) ([]model.Connection, error) {
	cur, err := s.col(collConnections).Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []model.Connection
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	for i := range out {
		if out[i].URI, err = s.enc.open(out[i].EncURI); err != nil {
			return nil, fmt.Errorf("decrypt uri: %w", err)
		}
	}
	return out, nil
}

// UpdateConnection updates the name, URI (re-encrypted), and enabled flag.
func (s *Store) UpdateConnection(ctx context.Context, c *model.Connection) error {
	enc, err := s.enc.seal(c.URI)
	if err != nil {
		return fmt.Errorf("encrypt uri: %w", err)
	}
	c.UpdatedAt = now()
	res, err := s.col(collConnections).UpdateOne(ctx, bson.M{"_id": c.ID}, bson.M{"$set": bson.M{
		"name":       c.Name,
		"enc_uri":    enc,
		"enabled":    c.Enabled,
		"updated_at": c.UpdatedAt,
	}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// SetConnectionStatus records the connection's runtime status and last error.
func (s *Store) SetConnectionStatus(ctx context.Context, id primitive.ObjectID, status model.ConnStatus, lastErr string) error {
	_, err := s.col(collConnections).UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{
		"status":     status,
		"last_error": lastErr,
		"updated_at": now(),
	}})
	return err
}

// DeleteConnection removes a connection and all of its watches + resume tokens.
func (s *Store) DeleteConnection(ctx context.Context, id primitive.ObjectID) error {
	watches, err := s.ListWatchesByConnection(ctx, id)
	if err != nil {
		return err
	}
	for _, w := range watches {
		if err := s.DeleteWatch(ctx, w.ID); err != nil {
			return err
		}
	}
	res, err := s.col(collConnections).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) col(name string) *mongo.Collection { return s.db.Collection(name) }
