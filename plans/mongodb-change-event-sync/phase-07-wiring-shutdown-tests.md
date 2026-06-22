# Phase 07 — Wiring, Shutdown, Tests, Docs

**Goal:** Assemble the production-ready single binary with clean lifecycle and green tests.
**Depends on:** all prior phases

## Tasks
- [ ] `cmd/server/main.go`: wire config → store → publisher → manager → metrics → http server.
- [ ] Graceful shutdown: drain watchers, flush resume tokens, close publisher/NATS, close Mongo.
- [ ] Ensure `go test ./...` is green: store, partition-keying, JWT middleware unit tests.
- [ ] README: run instructions, env vars, docker-compose usage, architecture summary.

## Files / packages
cmd/server/main.go   full wiring + signal handling
README.md            run + ops docs

## Exit criteria
- All 8 verification steps in `overview.md` pass end-to-end.
- Clean shutdown leaves no lost events (beyond at-least-once) and no leaked goroutines.
