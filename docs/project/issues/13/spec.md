---
github: https://github.com/Joeydekruis/dockyrd/issues/13
title: "[M1] Local server bootstrap and projects REST API"
milestone: M1 · Sprint 1 — Projects API
size: M
---

## Goal

Bootstrap the local `servers` row on setup and expose authenticated `/api/v1/projects` CRUD so M1 UI and services can attach to projects.

## Acceptance criteria

- [x] Setup completion creates one local `servers` row (idempotent if already present)
- [x] sqlc queries for servers, projects, environments, project_members
- [x] `GET|POST /api/v1/projects`, `GET|PATCH|DELETE /api/v1/projects/{id}` (session cookie auth)
- [x] `POST /projects` creates default `production` environment and adds creator to `project_members`
- [x] Slug derived from name, unique across projects
- [x] Go tests cover setup → create → list → patch → delete

## References

- `docs/BRIEFING.md` §10.2 Projects, §3.3 schema
- `internal/db/migrations/00001_init.sql`
