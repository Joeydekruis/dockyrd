---
github: https://github.com/Joeydekruis/dockyrd/issues/22
title: "[M2] Static sites, releases, rollback"
milestone: M2 · Sprint 4 — Static sites
size: L
---

## Goal

Build static sites from Git, publish releases on disk, serve via Caddy file_server, rollback via promote.

## Acceptance criteria

- [x] `kind=static_site` services + deploy worker path
- [x] Node container build + output dir detect (`dist`, `out`, …)
- [x] Releases under `{DataDir}/sites/<slug>/releases/<deploy>` + `current` symlink
- [x] Caddy static routes via `ListLiveStaticProxyRoutes`
- [x] `POST /api/v1/deployments/{id}/promote` rollback for static sites
- [x] `GET /api/v1/deployments/{id}/log`, deployment list includes `release_path`
- [x] Image/release retention: static `PruneReleases(5)`; app images pruned to 5 tags per service
