package store

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"data_sync_event/internal/model"
)

// AppendEvent records a processed change into the capped events_log.
func (s *Store) AppendEvent(ctx context.Context, e *model.EventLogEntry) error {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now()
	}
	_, err := s.col(collEventsLog).InsertOne(ctx, e)
	return err
}

// ListEvents returns the most recent events (newest first), optionally filtered.
func (s *Store) ListEvents(ctx context.Context, filter bson.M, limit int64) ([]model.EventLogEntry, error) {
	if filter == nil {
		filter = bson.M{}
	}
	cur, err := s.col(collEventsLog).Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "$natural", Value: -1}}).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []model.EventLogEntry
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AppendError records an operational error into the capped error_logs.
func (s *Store) AppendError(ctx context.Context, e *model.ErrorLog) error {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now()
	}
	_, err := s.col(collErrorLogs).InsertOne(ctx, e)
	return err
}

// ListErrors returns the most recent errors (newest first), optionally filtered.
func (s *Store) ListErrors(ctx context.Context, filter bson.M, limit int64) ([]model.ErrorLog, error) {
	if filter == nil {
		filter = bson.M{}
	}
	cur, err := s.col(collErrorLogs).Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "$natural", Value: -1}}).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []model.ErrorLog
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
