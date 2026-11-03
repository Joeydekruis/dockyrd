---
github: https://github.com/Joeydekruis/dockyrd/issues/16
title: "[M1] Image deploy, job queue, reconciler, Caddy sync"
milestone: M1 · Sprint 2 — Deploy & reconciler
size: L
---

## Goal

Run image-based app deploys via SQLite job queue, blue/green container swap, periodic reconciler, and Caddy reload from live routes.

## Acceptance criteria

- [x] sqlc for `deployments`, `jobs`, deploy context + proxy routes
- [x] `POST /api/v1/services/{id}/deployments` enqueues deploy job
- [x] `GET /api/v1/services/{id}/deployments`, `POST /api/v1/deployments/{id}/cancel`
- [x] `internal/deploy` image pull, container start, health check, supersede prior live
- [x] `internal/jobs` worker loop on `serve`
- [x] `internal/reconcile` interrupted deploy cleanup + `proxy.SyncFromDB`
- [x] Go test: enqueue deployment after create service

## References

- `docs/BRIEFING.md` §3.6, §11.1 (image source)
