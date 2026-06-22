# Plans

This folder holds design + development plans, **one subfolder per feature**.

## Convention

For each new feature, create `plans/<feature-slug>/` containing:

- **`overview.md`** — context, architecture, core mechanics, and verification.
- **`progress.md`** — a checklist tracking each phase's status (todo / in-progress / done).
- **`phase-NN-<slug>.md`** — one file per development phase, numbered from `00`.

### Phase file template

```
# Phase NN — <Title>

**Goal:** one sentence.
**Depends on:** <prior phases, or "none">

## Tasks
- [ ] concrete task (file/package it touches)

## Files / packages
internal/<pkg>/...   what lives here

## Exit criteria
- testable condition (maps to a Verification step in overview.md)
```

## Features

- [`mongodb-change-event-sync/`](./mongodb-change-event-sync/overview.md) — connect to N MongoDB deployments, watch change streams, and republish per-document-ordered events to NATS JetStream behind a swappable adapter, with an HTMX+SSE control/ops dashboard.
