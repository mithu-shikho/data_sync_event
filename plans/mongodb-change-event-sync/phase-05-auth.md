# Phase 05 — Auth

**Goal:** Gate the dashboard with Google OAuth (OIDC) and the API with JWT bearer tokens.
**Depends on:** phase-00 (independent of data plane; can run in parallel)

## Tasks
- [ ] `internal/auth`: Google OIDC login flow (`oauth2` + `go-oidc`) → signed session cookie; logout.
- [ ] Email/domain allowlist (e.g. `@shikho.com`) from config or a `users` collection.
- [ ] JWT bearer middleware for `/api/v1/*` (HS256 with config secret, or RS256).
- [ ] Endpoint/page to mint API tokens for programmatic access.
- [ ] Session middleware for dashboard routes; CSRF protection on mutating forms.

## Files / packages
internal/auth/oidc.go         google OIDC login + session cookie
internal/auth/jwt.go          issue/verify JWT, bearer middleware
internal/auth/session.go      signed session cookie + allowlist

## Exit criteria
- Verification 6: dashboard rejects non-allowlisted Google accounts; `/api/v1/*` returns 401 without a valid JWT.
