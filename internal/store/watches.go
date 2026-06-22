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

// CreateWatch inserts a new watch.
func (s *Store) CreateWatch(ctx context.Context, w *model.Watch) error {
	w.CreatedAt = now()
	w.UpdatedAt = w.CreatedAt
	res, err := s.col(collWatches).InsertOne(ctx, w)
	if err != nil {
		return fmt.Errorf("insert watch: %w", err)
	}
	w.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// GetWatch loads a single watch by ID.
func (s *Store) GetWatch(ctx context.Context, id primitive.ObjectID) (*model.Watch, error) {
	var w model.Watch
	err := s.col(collWatches).FindOne(ctx, bson.M{"_id": id}).Decode(&w)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// ListWatchesByConnection returns all watches for a connection.
func (s *Store) ListWatchesByConnection(ctx context.Context, connID primitive.ObjectID) ([]model.Watch, error) {
	return s.findWatches(ctx, bson.M{"connection_id": connID})
}

// ListEnabledWatches returns all enabled watches (desired running set).
func (s *Store) ListEnabledWatches(ctx context.Context) ([]model.Watch, error) {
	return s.findWatches(ctx, bson.M{"enabled": true})
}

func (s *Store) findWatches(ctx context.Context, filter bson.M) ([]model.Watch, error) {
	cur, err := s.col(collWatches).Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []model.Watch
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetWatchEnabled toggles a watch's enabled flag.
func (s *Store) SetWatchEnabled(ctx context.Context, id primitive.ObjectID, enabled bool) error {
	res, err := s.col(collWatches).UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{
		"enabled":    enabled,
		"updated_at": now(),
	}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteWatch removes a watch and its resume token.
func (s *Store) DeleteWatch(ctx context.Context, id primitive.ObjectID) error {
	res, err := s.col(collWatches).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	_, _ = s.col(collResumeTokens).DeleteOne(ctx, bson.M{"_id": id})
	return nil
}
