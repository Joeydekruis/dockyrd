---
github: https://github.com/Joeydekruis/dockyrd/issues/5
title: "[foundation] Config paths and environment"
milestone: M0 · Sprint 2
size: S
---

## Goal

Central `internal/config` for data directory, listen address, and env overrides.

## Acceptance criteria

- [ ] `internal/config` — `Load()` from env (`DOCKYARD_DATA_DIR`, `DOCKYARD_LISTEN`, …)
- [ ] Defaults: data under `/var/lib/dockyard` on Linux; temp/XDG-friendly default in tests
- [ ] `DatabasePath()` → `{DataDir}/control.db`
- [ ] Unit tests for override + default paths
- [ ] `dockyard serve` uses config listen addr (default `:8000`)

## References

- `docs/BRIEFING.md` §6
