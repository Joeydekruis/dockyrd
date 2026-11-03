---
github: https://github.com/Joeydekruis/dockyrd/issues/28
title: "[M4] Host metrics and Analytics UI"
milestone: M4 · Sprint 1 — Events & host metrics
size: L
---

## Goal

Host CPU/memory/disk/load time series and Analytics dashboard (design Server.dc.html).

## Acceptance criteria

- [x] `metric_host` inserts via background collector (linux + stub on other GOOS)
- [x] `GET /api/v1/server/metrics?range=1h|24h|7d`
- [x] `/analytics` — range pills, four metric tiles + sparklines, Important events + dismiss
- [x] Placeholder copy for Needs attention, Docker cleanup, Running containers (later slices)
- [x] `scripts/verify-m4-smoke.sh` green

## References

- `docs/design-index.json` route `/analytics`
- Events API: issue #27
