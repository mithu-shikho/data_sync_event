# Phase 03 — Watcher + Manager + Resume Tokens

**Goal:** The data plane — watch each source collection, normalize events, publish, and persist resume tokens for durability.
**Depends on:** phase-02 (Publisher interface), phase-01 (store, models)

## Tasks
- [ ] `watcher.go`: open `collection.Watch()` with `fullDocument: updateLookup` and `startAfter`/`resumeAfter` from the persisted token; map raw change → `ChangeEvent`; publish via `Publisher`.
- [ ] Persist resume token to `resume_tokens` after successful publish, batched every N events / T ms.
- [ ] Handle invalid/expired token: surface it, restart from now with a logged warning.
- [ ] `manager.go`: reconcile loop holding `watchID → cancelFunc`; diff enabled watches in store vs running; start/stop goroutines on change.
- [ ] Per-watcher panic isolation + backoff retry.
- [ ] Replica-set precheck on a connection's URI; clear error if standalone `mongod`.

## Files / packages
internal/watcher/watcher.go   single change-stream watcher
internal/watcher/manager.go   reconcile + lifecycle

## Exit criteria
- Verification 2 (event flow), 3 (per-doc ordering), 4 (resume-after-crash) pass.
- Standalone `mongod` connection reports a clear, surfaced error.
