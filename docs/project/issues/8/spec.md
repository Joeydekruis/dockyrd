---
github: https://github.com/Joeydekruis/dockyrd/issues/8
title: "[foundation] systemd unit and service install"
milestone: M0 · Sprint 3
size: S
---

## Goal

Ship `packaging/dockyrd.service` and wire install script to enable it.

## Acceptance criteria

- [ ] `packaging/dockyrd.service` per briefing §6.2 (Type=simple for M0 skeleton)
- [ ] `install.sh` installs unit, `daemon-reload`, `enable`, start/restart
- [ ] `EnvironmentFile=-/etc/dockyrd/dockyrd.env`

## References

- `docs/BRIEFING.md` §6.2
