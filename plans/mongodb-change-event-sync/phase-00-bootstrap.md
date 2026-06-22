# Phase 00 — Bootstrap

**Goal:** Replace the boilerplate, establish the Go module + tooling, and stand up local infra so later phases have something to build on.
**Depends on:** none

## Tasks
- [ ] Delete the GoLand starter body in `main.go`; move entrypoint to `cmd/server/main.go`.
- [ ] Set Go version in `go.mod` (e.g. `go 1.22`).
- [ ] `go get` deps: `go-chi/chi/v5`, `mongo-driver`, `nats.go`, `golang.org/x/oauth2`, `coreos/go-oidc/v3`, `golang-jwt/jwt/v5`.
- [ ] Create empty package dirs: `internal/{config,model,store,publisher,watcher,metrics,auth,api,dashboard}`, `web/{templates,static}`.
- [ ] `docker-compose.yml`: source Mongo (replica set, auto `rs.initiate()`), control Mongo, NATS with JetStream enabled.
- [ ] `Makefile`: `run`, `test`, `up`, `down`.

## Files / packages
cmd/server/main.go     minimal entrypoint (prints + graceful exit for now)
go.mod / go.sum        module + deps
docker-compose.yml     source Mongo RS + control Mongo + NATS JetStream
Makefile               dev shortcuts

## Exit criteria
- `go build ./...` succeeds.
- `docker compose up` brings up Mongo (replica set initiated) and NATS with JetStream.
