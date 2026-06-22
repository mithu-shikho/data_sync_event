# Phase 04 — Observability Hub

**Goal:** In-process metrics + event hub that powers the live dashboard, with no extra infrastructure.
**Depends on:** phase-03 (watcher/publisher emit into it)

## Tasks
- [ ] `internal/metrics`: per-watch counters — total processed, events/sec (sliding window), last-event time, status, partition distribution, current resume position.
- [ ] Bounded ring buffer of recent events for the live feed.
- [ ] SSE hub: fan-out of channels to connected browsers (subscribe/unsubscribe, slow-consumer drop).
- [ ] JetStream stats poller (~2s): `StreamInfo`/`ConsumerInfo` per partition → messages, bytes, first/last seq, pending (lag), ack-pending, redelivered, num-consumers.
- [ ] Wire `evProcessed`/`evPublished`/`evError` emissions into watcher + publisher; persist errors to `error_logs`.

## Files / packages
internal/metrics/metrics.go   counters + sliding-window rate
internal/metrics/hub.go       SSE fan-out hub + ring buffer
internal/metrics/jetstream.go JetStream stats poller

## Exit criteria
- Counters move under load; recent-event buffer fills.
- Stats readable (temporary debug endpoint acceptable until phase-06 wires the UI).
