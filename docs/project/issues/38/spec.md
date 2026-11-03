---
github: https://github.com/Joeydekruis/dockyrd/issues/38
---

# #38 · Self-update and rollback

## AC

- [ ] `dockyard update --check` compares current version to GitHub latest release
- [ ] `dockyard update` downloads linux asset, verifies checksums.txt when present, VACUUM INTO backup, swaps binary to `/usr/local/bin/dockyard` (`.prev` kept)
- [ ] `dockyard update --rollback` restores `.prev` binary on Linux
- [ ] `DOCKYARD_UPDATE_DRY_RUN=1` or `--dry-run` skips install
