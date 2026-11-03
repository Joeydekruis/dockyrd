---
github: https://github.com/Joeydekruis/dockyrd/issues/23
title: "[M3] Postgres & Valkey managed services"
milestone: M3 · Sprint 1 — Postgres & Valkey
size: L
---

## Goal

Provision `kind=postgres` and `kind=redis` (Valkey image) services on project Docker networks with generated credentials, persistent volumes, and reconcile-driven lifecycle (no deploy objects).

## Acceptance criteria

- [x] `POST /api/v1/services` accepts `postgres` and `redis` with optional version / mem limits; stores encrypted password in `config`
- [x] Reconcile loop ensures containers for `desired_state=running`, stops/removes when `stopped`
- [x] Official `postgres:<major>` and `valkey/valkey:<major>` images, `dy-<slug>` container, project network alias = service slug
- [x] `GET /api/v1/services/{id}/connection?reveal=1` returns briefing §7.3 connection keys (redact without `reveal`)
- [x] Go tests: config parse/generate, published vars, API create validation (sqlc + httptest)

## References

- `docs/BRIEFING.md` §7.3, §11.5
- `internal/docker/names.go` (`ServiceContainerName`, `VolumeName`)
