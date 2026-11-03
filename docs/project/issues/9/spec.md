---
github: https://github.com/Joeydekruis/dockyrd/issues/9
title: "[foundation] dockyard init and setup-token"
milestone: M0 · Sprint 3
size: M
---

## Goal

Idempotent `dockyard init` + `setup-token`; master key; setup token in SQLite settings.

## Acceptance criteria

- [ ] `internal/crypto` — `secret.key` (0600) under config dir
- [ ] `dockyard init` — dirs, master key, DB migrate, setup token logged once
- [ ] `dockyard setup-token` — new token while no users exist
- [ ] `GET /api/v1/setup/status` reads DB (`complete` when users exist)
- [ ] Tests under `internal/setup/`

## References

- `docs/BRIEFING.md` §6.1, §8
