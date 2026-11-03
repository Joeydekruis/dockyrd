---
github: https://github.com/Joeydekruis/dockyrd/issues/27
title: "[M4] Event log API and deploy events"
milestone: M4 · Sprint 1 — Events & host metrics
size: M
---

## Goal

SQLite-backed event feed with HTTP list/ack and deploy lifecycle entries.

## Acceptance criteria

- [x] sqlc queries: insert, list recent, get by id, ack
- [x] `internal/eventlog` — InsertEvent + SSE publish on `events` topic
- [x] `GET /api/v1/events` (limit, category filter)
- [x] `POST /api/v1/events/{id}/ack` → 204
- [x] Deploy runtime records live/failed with category `deploys`

## References

- `docs/BRIEFING.md` §11.4 events
- Analytics feed: issue #28
