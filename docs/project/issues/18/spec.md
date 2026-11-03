---
github: https://github.com/Joeydekruis/dockyrd/issues/18
title: "[M1] Projects dashboard wired to API"
milestone: M1 · Sprint 3 — Logs, env & Projects UI
size: S
---

## Goal

Replace Projects placeholder with live list + create project against `/api/v1/projects`.

## Acceptance criteria

- [x] `/projects` segment renders project grid from API
- [x] Create project form calls `POST /api/v1/projects`
- [x] `web/lib/api.ts` project client helpers
- [x] `make build` green

## References

- `docs/design-index.json` route `/projects`
