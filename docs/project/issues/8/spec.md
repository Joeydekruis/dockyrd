---
github: https://github.com/Joeydekruis/dockyrd/issues/8
title: "[foundation] systemd unit and service install"
milestone: M0 · Sprint 3
size: S
---

## Goal

Ship `packaging/dockyard.service` and wire install script to enable it.

## Acceptance criteria

- [ ] `packaging/dockyard.service` per briefing §6.2 (Type=simple for M0 skeleton)
- [ ] `install.sh` installs unit, `daemon-reload`, `enable`, start/restart
- [ ] `EnvironmentFile=-/etc/dockyard/dockyard.env`

## References

- `docs/BRIEFING.md` §6.2
