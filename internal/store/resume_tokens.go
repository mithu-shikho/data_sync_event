package store

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"data_sync_event/internal/model"
)

// GetResumeToken returns the persisted resume token for a watch, or ErrNotFound.
func (s *Store) GetResumeToken(ctx context.Context, watchID primitive.ObjectID) (*model.ResumeToken, error) {
	var rt model.ResumeToken
	err := s.col(collResumeTokens).FindOne(ctx, bson.M{"_id": watchID}).Decode(&rt)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// SaveResumeToken upserts the resume token for a watch.
func (s *Store) SaveResumeToken(ctx context.Context, watchID primitive.ObjectID, token bson.Raw) error {
	_, err := s.col(collResumeTokens).UpdateOne(ctx,
		bson.M{"_id": watchID},
		bson.M{"$set": bson.M{"token": token, "updated_at": now()}},
		options.Update().SetUpsert(true),
	)
	return err
}
