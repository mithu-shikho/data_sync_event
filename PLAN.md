# MongoDB Change-Event Sync System — Plan

## Context

A service that connects to **N MongoDB deployments**, watches each for data changes via Change Streams, and republishes every change as an ordered event onto a message queue so downstream consumers can sync/react. Frequent updates to the same document must not be reordered — a later write must never be applied before an earlier one.

Locked-in decisions:
- **Queue:** NATS JetStream now, behind a **Publisher adapter interface** so Kafka/RabbitMQ/etc. can be added later without touching the watcher engine.
- **Ordering:** **Per-document** — all changes to the same `_id` are strictly ordered; different documents flow in parallel.
- **Control plane:** a **web dashboard** to register/manage connections and watched collections at runtime (not a static config file).
- **Frontend:** Go server-rendered (`html/template` + **HTMX**, live updates via **Server-Sent Events**) — single binary, no separate JS build.
- **Rich operational dashboard:** live feed of documents being processed, live **NATS JetStream queue stats**, **error logs**, per-connection/per-watch metrics (throughput, lag, last-event time, status).
- **Metadata store:** a dedicated **control MongoDB** holding connection configs and resume tokens.
- **Auth:** **Google OAuth (OIDC)** for dashboard login; **JWT bearer** for the API.
- **Durability:** **persisted resume tokens** → at-least-once delivery, no events lost across restarts.

> **Prerequisite:** MongoDB Change Streams require each *source* deployment to be a **replica set or sharded cluster**, not a standalone `mongod`. The dashboard should surface a clear error if a registered URI isn't.

---

## Architecture

A single Go binary running two cooperating planes:
- **Control plane** — HTTP server: dashboard (HTMX) + JSON API for CRUD on connections & watches; stores desired state in the control MongoDB.
- **Data plane** — a Watcher Manager that reconciles running change-stream watchers against the desired state, normalizes each event, and publishes it through the Publisher adapter with a per-document partition key.

```
                ┌─────────────── single Go binary ───────────────┐
  Google OAuth →│  Dashboard (HTMX)   API (JWT)                   │
                │        └────────┬────────┘                      │
                │            Metadata Store ── control MongoDB     │
                │           (connections, watches, resume_tokens)  │
                │                 │ reconcile                      │
   source DB 1 →│  Watcher Manager ─┬─ watcher(conn1/db/coll) ─┐   │
   source DB 2 →│                   ├─ watcher(...)            │   │ → Publisher adapter → NATS JetStream
   source DB N →│                   └─ watcher(...)        partition by _id   (Kafka later)
                └──────────────────────────────────────────────────┘
```

### Package layout
```
cmd/server/main.go            wire config → store → manager → http server
internal/config/              env config (control URI, OAuth, JWT secret, NATS URL, #partitions)
internal/model/               ChangeEvent envelope, Connection, Watch domain types
internal/store/               control-Mongo repos: connections, watches, resume_tokens
internal/watcher/             Manager (reconcile loop) + Watcher (per change stream)
internal/publisher/           Publisher interface + nats_jetstream.go + partition keying
internal/auth/                google OIDC login, signed session cookie, JWT issue/verify middleware
internal/api/                 REST handlers (/api/v1/...) behind JWT
internal/dashboard/           HTML handlers + SSE endpoints behind session auth
internal/metrics/             in-process counters, recent-event ring buffer, SSE hub, JetStream stats poller
web/templates/  web/static/   HTMX templates + assets
```

### Suggested libraries
- Router: `github.com/go-chi/chi/v5`
- Mongo: `go.mongodb.org/mongo-driver`
- NATS: `github.com/nats-io/nats.go` (JetStream)
- OAuth/OIDC: `golang.org/x/oauth2` + `github.com/coreos/go-oidc/v3`
- JWT: `github.com/golang-jwt/jwt/v5`
- Logging: std `log/slog`

---

## Core mechanics

### 1. Publisher adapter (portability)
```go
// internal/publisher
type ChangeEvent struct {
    EventID      string            // resume-token hash / oplog ts — used as dedup id
    Connection   string
    DB, Coll     string
    OpType       string            // insert|update|replace|delete
    DocumentKey  string            // the _id, stringified — the ordering/partition key
    FullDocument bson.Raw          // updateLookup result (nil for delete)
    UpdateDesc   bson.Raw          // updated/removed fields for updates
    ClusterTime  primitive.Timestamp
}

type Publisher interface {
    Publish(ctx context.Context, ev ChangeEvent) error
    Close() error
}
```
NATS JetStream impl: subject `mongo.<conn>.<db>.<coll>.p<N>` where `N = hash(DocumentKey) % numPartitions`; set `Nats-Msg-Id = EventID` for stream dedup. Swapping queues = new impl of this interface only.

### 2. Per-document ordering
Same `_id` → same hash → same partition subject → consumed in order; distinct docs spread across partitions for parallel throughput. `numPartitions` is config. Consumers must be idempotent (at-least-once).

### 3. Watcher + resume tokens
Each watcher opens `collection.Watch()` with `fullDocument: updateLookup` and `startAfter`/`resumeAfter` from the persisted token. After a successful publish, persist the resume token to `resume_tokens` (batched every N events / T ms). On restart the manager resumes exactly where it left off; if a token is too old/invalid, surface it and restart from now with a warning.

### 4. Watcher Manager (reconcile)
Holds a map of `watchID → cancelFunc`. On startup and whenever the dashboard mutates a watch/connection, it diffs desired (enabled watches in store) vs running and starts/stops goroutines. Per-watcher panics are isolated, logged, and retried with backoff.

### 5. Auth
- **Dashboard:** Google OIDC login → signed session cookie; allowlist by email/domain (e.g. `@shikho.com`) stored in config or a `users` collection.
- **API:** `Authorization: Bearer <jwt>` middleware (HS256 with config secret, or RS256). A dashboard page mints API tokens.

### 6. Metadata collections (control MongoDB)
- `connections` — `{ _id, name, uri (encrypted at rest), enabled, status, lastError, createdAt }`
- `watches` — `{ _id, connectionId, db, collection, pipeline?, enabled, createdAt }`
- `resume_tokens` — `{ _id: watchID, token, updatedAt }`
- `users` (optional) — OAuth allowlist.

---

## Observability dashboard

`internal/metrics` is the hub every component reports into; the dashboard reads from it. Watchers and the publisher emit lightweight events (`evProcessed`, `evPublished`, `evError`) to an in-process **SSE hub** and bump counters — no extra infra, in-memory with a bounded ring buffer of recent items.

1. **Live processing feed** — each watcher pushes `{ts, conn, db.coll, opType, documentKey, partition, status}` to the hub. Dashboard subscribes to `GET /dash/stream/feed` (SSE); HTMX `sse-swap` prepends rows. Capped `events_log` collection backs a searchable history page.
2. **NATS JetStream queue stats** — poller (~2s) calls `js.StreamInfo()`/`ConsumerInfo()` per partition: messages, bytes, first/last seq, **consumer pending (lag)**, ack-pending, redelivered, num-consumers. Live stat cards via `GET /dash/stream/queue` (SSE).
3. **Error logs** — all errors (publish failure, watcher panic/restart, invalid/expired resume token, source unreachable) written to capped `error_logs` and streamed live. Live error panel + filterable history.
4. **Per-connection & per-watch metrics** — total processed, events/sec (sliding window), last-event time, current resume position, partition distribution, status. Optional `GET /metrics` in Prometheus format.

> SSE (not WebSockets) keeps it simple and HTMX-native; each stream endpoint sits behind the same session-auth middleware.

---

## Development milestones

### M0 — Bootstrap
Replace boilerplate `main.go`; set Go version in `go.mod`; `go get` deps; create package skeleton; add `docker-compose.yml` (source Mongo replica set, control Mongo, NATS+JetStream) + `Makefile`.
**Exit:** `go build ./...` succeeds; compose brings up Mongo (RS-initiated) + NATS.

### M1 — Config + models + control store
`internal/config` (env-driven), `internal/model` (Connection, Watch, ChangeEvent, ResumeToken, ErrorLog), `internal/store` (Mongo repos + indexes, encrypted URIs, capped `events_log`/`error_logs`).
**Exit:** CRUD unit tests pass against control Mongo.

### M2 — Publisher adapter + partitioning
`internal/publisher`: interface + `ChangeEvent`; `nats_jetstream.go` (stream ensure, subject `mongo.<conn>.<db>.<coll>.p<N>`, `Nats-Msg-Id` dedup); pure partition-key helper.
**Exit:** unit test — same `_id` → same partition; integration publish/read via `nats sub 'mongo.>'`.

### M3 — Watcher + Manager + resume tokens
`watcher.go` (Watch with `updateLookup` + `startAfter`, map→ChangeEvent, publish, batched token persist); `manager.go` (reconcile `watchID→cancelFunc`, panic isolation + backoff, replica-set precheck).
**Exit:** event flow, per-doc ordering, resume-after-crash verified.

### M4 — Observability hub
`internal/metrics` (counters, sliding-window rate, ring buffer, SSE hub, JetStream poller); wire emissions into watcher + publisher; persist errors.
**Exit:** counters move under load; stats readable.

### M5 — Auth
`internal/auth`: Google OIDC login → signed session cookie, email/domain allowlist; JWT bearer middleware for `/api/v1/*`; API-token mint page.
**Exit:** non-allowlisted Google rejected; API 401 without valid JWT.

### M6 — API + Dashboard
`internal/api` (REST CRUD behind JWT, enable/disable → reconcile); `internal/dashboard` + `web/` (HTMX pages; SSE live feed / queue stats / error panel / metrics grid; optional Prometheus `/metrics`).
**Exit:** setup via dashboard + live feed/stats/error streaming verified.

### M7 — Wiring, shutdown, tests, docs
`cmd/server/main.go` wiring; graceful shutdown (drain watchers, flush tokens, close publisher); `go test ./...` green; README.
**Exit:** all verification steps pass; clean shutdown, no events lost beyond at-least-once.

**Critical path:** M1 → (M2 ∥ M3, M3 depends on M2's interface only). M4 plugs into M3. M5 independent after M0. M6 depends on M1+M4+M5. M7 last.

---

## Verification

`docker-compose.yml` with: a **source** Mongo replica set, the **control** Mongo, and **NATS with JetStream**.

1. **Setup via dashboard:** log in (Google OAuth), add a source connection + a watch; confirm "running".
2. **Event flow:** `nats sub 'mongo.>'`; insert/update/delete in source → matching events with correct `opType`/`fullDocument`.
3. **Per-document ordering:** rapidly update one `_id` → all land on the same `p<N>` subject in write order.
4. **Resume/durability:** kill mid-stream, write while down, restart → events from downtime delivered (no gap) from persisted token.
5. **Live dashboard:** feed rows stream live; JetStream cards update; stop NATS briefly → error streams live + appears in `error_logs`.
6. **Auth:** dashboard rejects non-allowlisted Google accounts; `/api/v1/*` rejects missing/invalid JWT (401).
7. **Adapter portability:** `watcher`/`manager` import only the `Publisher` interface, never `nats.go` directly.
8. `go test ./...` for store, partition-keying, and JWT middleware.