// Package model holds the core domain types shared across the system.
package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OpType is a MongoDB change-stream operation type.
type OpType string

const (
	OpInsert  OpType = "insert"
	OpUpdate  OpType = "update"
	OpReplace OpType = "replace"
	OpDelete  OpType = "delete"
)

// ConnStatus is the lifecycle status of a source connection.
type ConnStatus string

const (
	StatusPending ConnStatus = "pending"
	StatusRunning ConnStatus = "running"
	StatusError   ConnStatus = "error"
	StatusStopped ConnStatus = "stopped"
)

// Connection is a registered source MongoDB deployment. The URI is encrypted
// at rest by the store; in memory it holds the plaintext connection string.
type Connection struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	URI       string             `bson:"-" json:"uri,omitempty"`         // plaintext, never persisted directly
	EncURI    []byte             `bson:"enc_uri" json:"-"`               // AES-GCM ciphertext (or plaintext bytes if encryption disabled)
	Enabled   bool               `bson:"enabled" json:"enabled"`
	Status    ConnStatus         `bson:"status" json:"status"`
	LastError string             `bson:"last_error,omitempty" json:"lastError,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
}

// Watch is a collection (within a Connection) to watch via a change stream.
type Watch struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConnectionID primitive.ObjectID `bson:"connection_id" json:"connectionId"`
	DB           string             `bson:"db" json:"db"`
	Collection   string             `bson:"collection" json:"collection"`
	Pipeline     []bson.M           `bson:"pipeline,omitempty" json:"pipeline,omitempty"` // optional aggregation filter
	Enabled      bool               `bson:"enabled" json:"enabled"`
	CreatedAt    time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updatedAt"`
}

// ResumeToken persists the change-stream position for a watch, keyed by watch ID.
type ResumeToken struct {
	WatchID   primitive.ObjectID `bson:"_id" json:"watchId"`
	Token     bson.Raw           `bson:"token" json:"-"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
}

// ChangeEvent is the normalized envelope published to the queue. DocumentKey
// is the stringified _id and is the ordering / partition key.
type ChangeEvent struct {
	EventID      string              `bson:"event_id" json:"eventId"` // dedup id (resume-token hash / cluster ts)
	Connection   string              `bson:"connection" json:"connection"`
	DB           string              `bson:"db" json:"db"`
	Coll         string              `bson:"coll" json:"coll"`
	OpType       OpType              `bson:"op_type" json:"opType"`
	DocumentKey  string              `bson:"document_key" json:"documentKey"`
	FullDocument bson.Raw            `bson:"full_document,omitempty" json:"-"`
	UpdateDesc   bson.Raw            `bson:"update_desc,omitempty" json:"-"`
	ClusterTime  primitive.Timestamp `bson:"cluster_time" json:"-"`
}

// Severity classifies an ErrorLog entry.
type Severity string

const (
	SevWarn  Severity = "warn"
	SevError Severity = "error"
)

// ErrorLog is a recorded operational error (capped collection).
type ErrorLog struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConnectionID string             `bson:"connection_id,omitempty" json:"connectionId,omitempty"`
	WatchID      string             `bson:"watch_id,omitempty" json:"watchId,omitempty"`
	Severity     Severity           `bson:"severity" json:"severity"`
	Message      string             `bson:"message" json:"message"`
	CreatedAt    time.Time          `bson:"created_at" json:"createdAt"`
}

// EventLogEntry is a compact record of a processed change (capped collection,
// backs the dashboard's searchable feed history).
type EventLogEntry struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Connection  string             `bson:"connection" json:"connection"`
	DB          string             `bson:"db" json:"db"`
	Coll        string             `bson:"coll" json:"coll"`
	OpType      OpType             `bson:"op_type" json:"opType"`
	DocumentKey string             `bson:"document_key" json:"documentKey"`
	Partition   int                `bson:"partition" json:"partition"`
	Status      string             `bson:"status" json:"status"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
}
