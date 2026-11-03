---
github: https://github.com/Joeydekruis/dockyrd/issues/38
---

# #38 · Self-update and rollback

## AC

- [x] `dockyard update --check` compares current version to GitHub latest release
- [x] `dockyard update` downloads linux asset, verifies checksums.txt when present, VACUUM INTO backup, swaps binary to `/usr/local/bin/dockyard` (`.prev` kept)
- [x] `dockyard update --rollback` restores `.prev` binary on Linux
- [x] `DOCKYARD_UPDATE_DRY_RUN=1` or `--dry-run` skips install
