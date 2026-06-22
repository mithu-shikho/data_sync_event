# Phase 06 — API + Dashboard

**Goal:** The control + operations UI: REST CRUD behind JWT and the HTMX/SSE dashboard.
**Depends on:** phase-01 (store), phase-04 (live views), phase-05 (auth)

## Tasks
- [ ] `internal/api`: REST CRUD for connections/watches (behind JWT); enable/disable triggers a manager reconcile.
- [ ] `internal/dashboard` + `web/`: HTMX pages — connections list/detail, add/edit forms, watch management.
- [ ] SSE live views: `/dash/stream/feed` (processing feed), `/dash/stream/queue` (JetStream stat cards), live error panel + filterable history.
- [ ] Per-connection/per-watch metrics grid; replica-set/connection error surfaced in the UI.
- [ ] Optional `GET /metrics` Prometheus endpoint.

## Files / packages
internal/api/*.go             REST handlers behind JWT
internal/dashboard/*.go       HTML + SSE handlers behind session auth
web/templates/  web/static/   HTMX templates + assets

## Exit criteria
- Verification 1: setup via dashboard (login, add connection + watch, shows "running").
- Verification 5: live feed rows stream, JetStream cards update, errors stream live + appear in `error_logs`.
