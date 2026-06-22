# Phase 02 — Publisher Adapter + Partitioning

**Goal:** A queue abstraction with per-document ordering, implemented for NATS JetStream.
**Depends on:** phase-01 (uses `ChangeEvent` from `internal/model`)

## Tasks
- [ ] `internal/publisher`: define `Publisher` interface (`Publish(ctx, ChangeEvent) error`, `Close() error`).
- [ ] Pure partition-key helper: `partition(documentKey string, n int) int` = `hash(documentKey) % n` (fully unit-tested, deterministic).
- [ ] `nats_jetstream.go`: ensure stream(s); publish to subject `mongo.<conn>.<db>.<coll>.p<N>`; set `Nats-Msg-Id = EventID` for dedup.
- [ ] Keep `internal/publisher` free of any watcher/store imports (clean adapter boundary).

## Files / packages
internal/publisher/publisher.go      Publisher interface + ChangeEvent re-export
internal/publisher/partition.go       pure hash→partition helper
internal/publisher/nats_jetstream.go  JetStream implementation

## Exit criteria
- Unit test: same `_id` → same partition across runs; distribution across partitions is even-ish.
- Integration: publish events, read them back via `nats sub 'mongo.>'`.
