# Progress — MongoDB Change-Event Sync

Status legend: ⬜ todo · 🟡 in-progress · ✅ done

| Phase | Title | Status |
|-------|-------|--------|
| 00 | Bootstrap (module, deps, skeleton, docker-compose) | ⬜ |
| 01 | Config + models + control store | ⬜ |
| 02 | Publisher adapter + partitioning | ⬜ |
| 03 | Watcher + Manager + resume tokens | ⬜ |
| 04 | Observability hub | ⬜ |
| 05 | Auth (Google OIDC + JWT) | ⬜ |
| 06 | API + Dashboard | ⬜ |
| 07 | Wiring, shutdown, tests, docs | ⬜ |

**Critical path:** 00 → 01 → (02 ∥ 03) → 04 → 06 → 07; phase 05 independent after 00.

Update the status cell as each phase moves. Tick the task checkboxes inside each `phase-NN-*.md` as they complete.
