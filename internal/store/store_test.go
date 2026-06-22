package store

import (
	"context"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"data_sync_event/internal/config"
	"data_sync_event/internal/model"
)

// testStore connects to a control Mongo (TEST_CONTROL_MONGO_URI or localhost),
// using a throwaway database. The test is skipped if Mongo is unreachable.
func testStore(t *testing.T) (*Store, func()) {
	t.Helper()
	uri := os.Getenv("TEST_CONTROL_MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	key := make([]byte, 32)
	cfg := &config.Config{
		ControlMongoURI:  uri,
		ControlDBName:    "data_sync_test_" + strconv.FormatInt(time.Now().UnixNano(), 36),
		URIEncryptionKey: key, // exercise encryption path
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Probe reachability first so we can skip cleanly when Mongo is down.
	probe, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil || probe.Ping(ctx, nil) != nil {
		if probe != nil {
			_ = probe.Disconnect(context.Background())
		}
		t.Skipf("control Mongo not reachable at %s; skipping integration test", uri)
	}
	_ = probe.Disconnect(context.Background())

	s, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	cleanup := func() {
		_ = s.db.Drop(context.Background())
		_ = s.Close(context.Background())
	}
	return s, cleanup
}

func TestConnectionCRUD(t *testing.T) {
	s, cleanup := testStore(t)
	defer cleanup()
	ctx := context.Background()

	c := &model.Connection{Name: "src-1", URI: "mongodb://u:p@src:27017/?replicaSet=rs0", Enabled: true}
	if err := s.CreateConnection(ctx, c); err != nil {
		t.Fatalf("create: %v", err)
	}
	if c.ID.IsZero() {
		t.Fatal("expected generated ID")
	}
	if c.Status != model.StatusPending {
		t.Fatalf("expected pending status, got %s", c.Status)
	}

	got, err := s.GetConnection(ctx, c.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.URI != c.URI {
		t.Fatalf("URI round trip failed: %q", got.URI)
	}
	if len(got.EncURI) == 0 {
		t.Fatal("expected encrypted URI bytes persisted")
	}

	got.Name = "src-1-renamed"
	got.URI = "mongodb://u:p@src2:27017/?replicaSet=rs0"
	if err := s.UpdateConnection(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.SetConnectionStatus(ctx, c.ID, model.StatusRunning, ""); err != nil {
		t.Fatalf("set status: %v", err)
	}

	list, err := s.ListConnections(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].Name != "src-1-renamed" || list[0].Status != model.StatusRunning {
		t.Fatalf("unexpected list: %+v", list)
	}

	if err := s.DeleteConnection(ctx, c.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetConnection(ctx, c.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestWatchAndResumeToken(t *testing.T) {
	s, cleanup := testStore(t)
	defer cleanup()
	ctx := context.Background()

	c := &model.Connection{Name: "src", URI: "mongodb://src:27017"}
	if err := s.CreateConnection(ctx, c); err != nil {
		t.Fatal(err)
	}

	w := &model.Watch{ConnectionID: c.ID, DB: "shop", Collection: "orders", Enabled: true}
	if err := s.CreateWatch(ctx, w); err != nil {
		t.Fatalf("create watch: %v", err)
	}

	enabled, err := s.ListEnabledWatches(ctx)
	if err != nil || len(enabled) != 1 {
		t.Fatalf("list enabled: %v %d", err, len(enabled))
	}

	// Resume token upsert + read back.
	tok, err := bson.Marshal(bson.M{"_data": "8265AABBCC"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SaveResumeToken(ctx, w.ID, bson.Raw(tok)); err != nil {
		t.Fatalf("save token: %v", err)
	}
	rt, err := s.GetResumeToken(ctx, w.ID)
	if err != nil {
		t.Fatalf("get token: %v", err)
	}
	if len(rt.Token) == 0 {
		t.Fatal("expected token bytes")
	}

	// Disable + delete cascade removes the resume token.
	if err := s.SetWatchEnabled(ctx, w.ID, false); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteWatch(ctx, w.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetResumeToken(ctx, w.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected resume token removed, got %v", err)
	}
}

func TestCappedLogs(t *testing.T) {
	s, cleanup := testStore(t)
	defer cleanup()
	ctx := context.Background()

	if err := s.AppendEvent(ctx, &model.EventLogEntry{
		Connection: "src", DB: "shop", Coll: "orders", OpType: model.OpInsert,
		DocumentKey: "abc", Partition: 3, Status: "published",
	}); err != nil {
		t.Fatalf("append event: %v", err)
	}
	if err := s.AppendError(ctx, &model.ErrorLog{Severity: model.SevError, Message: "boom"}); err != nil {
		t.Fatalf("append error: %v", err)
	}

	evs, err := s.ListEvents(ctx, nil, 10)
	if err != nil || len(evs) != 1 {
		t.Fatalf("list events: %v %d", err, len(evs))
	}
	errs, err := s.ListErrors(ctx, nil, 10)
	if err != nil || len(errs) != 1 {
		t.Fatalf("list errors: %v %d", err, len(errs))
	}
}
