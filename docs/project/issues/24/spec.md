---
github: https://github.com/Joeydekruis/dockyrd/issues/24
title: "[M3] Service reference variables (${{slug.VAR}})"
milestone: M3 · Sprint 2 — Reference variables
size: M
---

## Goal

Resolve `${{service_slug.VAR}}` placeholders in app/static env vars at deploy time within the same project; reject cycles.

## Acceptance criteria

- [x] Parser + resolver using published vars from postgres/redis/bucket/app (briefing §7.3 table)
- [x] Wired into `runtimeEnv` / `buildArgs` before container start and build
- [x] Validation error on unknown slug or unknown key; cycle detection
- [ ] Optional: `POST /services` database create offers to inject `DATABASE_URL` into selected apps
- [x] Unit tests for resolve, cycle, and cross-service fixture

## References

- `docs/BRIEFING.md` §7.3
