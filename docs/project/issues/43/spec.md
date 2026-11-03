---
github: https://github.com/Joeydekruis/dockyrd/issues/43
---

# #43 · Install preflight + verify-install

## AC

- [x] `install.sh` preflight (curl, arch, docker warning, systemd when not skipped)
- [x] `DOCKYARD_SKIP_ROOT`, `DOCKYARD_SKIP_SYSTEMD`, `DOCKYARD_BIN_DEST` for CI/dev
- [x] `./scripts/verify-install.sh` passes without root
