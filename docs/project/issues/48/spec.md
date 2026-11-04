---
github: https://github.com/Joeydekruis/dockyrd/issues/48
title: Database platform (Query Studio, backups, lifecycle)
---

## Acceptance criteria (in progress)

### Lifecycle (slice 1 — implemented)
- [x] `DELETE /api/v1/services/{id}` for postgres/redis with `confirm_slug` + optional `delete_volume`
- [x] `POST /api/v1/services/{id}/restart` recreates container via reconciler
- [x] Database detail: Restart, Delete (danger zone), per-secret connection reveal

### Backups (slice 2 — schema + config UI)
- [x] Migration `00004_database_backups` (targets, policies, runs)
- [x] Server S3 target API + `/settings/server/backups`
- [x] Per-database backup policy API + settings modal (hourly/daily/weekly, retention default 30)
- [ ] Job runner: pg_dump, local retention, S3 upload, restore, manual “Backup now”

### Query Studio (slice 3 — open)
- [ ] Postgres tab: schema browser + SQL editor + audited proxy API

### Connection UX (slice 4 — open)
- [ ] Internal / external connection cards, enable public access

## Notes

Briefing v0.2 backup design; user retention default **30 days**, intervals **hourly | daily | weekly**.
