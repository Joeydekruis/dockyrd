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
- [x] Job runner: pg_dump / Valkey RDB, local files + S3 upload, scheduler + **Backup now**
- [x] Restore to **same** Postgres database (`POST …/backup-runs/{runId}/restore`)
- [x] Remote retention prune after S3 upload (per slug prefix + policy retention)
- [ ] Restore to **new** database; Redis restore

### Query Studio (slice 3 — MVP)
- [x] Postgres tab: SQL editor + read-only proxy (`POST …/query`, audit log, row cap)
- [x] Schema browser (`GET …/schema`, sidebar in Query tab)
- [ ] Redis CLI; admin **write** queries

### Connection UX (slice 4)
- [x] Internal / external connection cards; enable Postgres public access (`PATCH …/public-access`)
- [ ] Redis external / public access

## Notes

Briefing v0.2 backup design; user retention default **30 days**, intervals **hourly | daily | weekly**.
