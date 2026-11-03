---
github: https://github.com/Joeydekruis/dockyrd/issues/7
title: "[foundation] install.sh skeleton and paths"
milestone: M0 · Sprint 3
size: S
---

## Goal

POSIX `scripts/install.sh` skeleton: paths, local binary install, calls `dockyard init`.

## Acceptance criteria

- [ ] `scripts/install.sh` — root check, `/etc/dockyard` + `/var/lib/dockyard`, `DOCKYARD_INSTALL_BIN` override
- [ ] `scripts/uninstall.sh` — stop/disable unit, remove binary; keep data
- [ ] Documented dev flow in script header / README

## References

- `docs/BRIEFING.md` §6.1
