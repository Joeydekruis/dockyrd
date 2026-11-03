---
github: https://github.com/Joeydekruis/dockyrd/issues/4
title: "[foundation] SQLite goose migrations and sqlc"
milestone: M0 · Sprint 2
size: M
---

## Goal

Embedded goose migrations and sqlc-generated queries for the control-plane SQLite database.

## Acceptance criteria

- [ ] `internal/db/migrations/00001_init.sql` — schema per `docs/BRIEFING.md` §7.2
- [ ] goose migrations embedded; `MigrateUp` on open in tests
- [ ] `sqlc.yaml` + `internal/db/queries/*.sql` + committed `internal/db/sqlc/`
- [ ] `internal/db` — `Open(path)` returns `*sql.DB` after migrate
- [ ] `go test ./internal/db/...` passes (clean temp DB)
- [ ] `make generate` runs sqlc (document in `setup.md` if new)

## References

- `docs/BRIEFING.md` §7
