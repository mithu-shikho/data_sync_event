# Phase 01 — Config + Models + Control Store

**Goal:** Typed configuration and the control-MongoDB persistence layer for desired state and durability data.
**Depends on:** phase-00

## Tasks
- [ ] `internal/config`: env-driven config — `CONTROL_MONGO_URI`, `NATS_URL`, `NUM_PARTITIONS`, `JWT_SECRET`, OAuth client id/secret/redirect, session key, allowlist domain. Validate on startup.
- [ ] `internal/model`: `Connection`, `Watch`, `ChangeEvent`, `ResumeToken`, `ErrorLog` types.
- [ ] `internal/store`: Mongo repos with CRUD for `connections`, `watches`, `resume_tokens`.
- [ ] Ensure indexes; encrypt connection URIs at rest.
- [ ] Create capped collections `events_log` and `error_logs`.

## Files / packages
internal/config/config.go    load + validate env
internal/model/*.go          domain types
internal/store/*.go          mongo repos (connections, watches, resume_tokens), capped logs, indexes

## Exit criteria
- Unit tests create/read/update/delete connections & watches against control Mongo (test container or compose).
- Capped collections exist with expected size limits.
