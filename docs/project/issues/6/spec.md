---
github: https://github.com/Joeydekruis/dockyrd/issues/6
title: "[foundation] Auth package stubs"
milestone: M0 · Sprint 2
size: S
---

## Goal

`internal/auth` package with session types and middleware placeholder for future login (Sprint 3–4).

## Acceptance criteria

- [ ] Session + User ID types aligned with briefing §8 (no full login yet)
- [ ] `RequireAuth` middleware stub — passes through when no session store wired; or documents `501` on `/api/v1/me` stub route
- [ ] `GET /api/v1/setup/status` stub → JSON `{ "complete": false }` without DB (until setup-token)
- [ ] Tests for middleware / stub handler JSON shape

## References

- `docs/BRIEFING.md` §8
