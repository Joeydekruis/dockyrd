# Generator state

**Updated:** 2026-11-05

## In progress

**Backup hardening B-1 (#67)** — implemented locally pending parent verify + handoff:

- Migration `00008_backups_hardening.sql` (schedule fields, sha256/verified_at, managed Garage target + test metadata, `database_restore_runs`).
- `internal/backup`: `NextRunAfter` with `run_at_minute`/`run_weekday`, disk guard, sha256 + `pg_restore --list`, eventlog category `backups` + notify on failure, async `database_restore` job with pre-snapshot and `log_tail`, Garage target helper + S3 probe test.
- API: Garage target POST, target test POST, policy schedule PATCH fields, restore → 202 + restore-runs GET, Redis restore 422.
- UI: server backups (Garage + test), database detail (schedule, verified badge, restore confirm modal).
- Tests: `go test ./internal/backup/...`, `./internal/server/...` (includes restore Redis message check).

**Notifications F1 (#63)** — still pending handoff from prior session (unchanged scope).

## Last ship

**M6 · Resource links** (#58–#61) @ **v0.0.49** — merged to `main`, GitHub release tagged.

## Previous

F0 notifications hardening (#62) — SMTP, `internal/netsafe`, reconcile interrupted-deploy fix.
