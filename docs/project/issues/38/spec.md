---
github: https://github.com/Joeydekruis/dockyrd/issues/38
---

# #38 · Self-update and rollback

## AC

- [x] `dockyrd update --check` compares current version to GitHub latest release
- [x] `dockyrd update` downloads linux asset, verifies checksums.txt when present, VACUUM INTO backup, swaps binary to `/usr/local/bin/dockyrd` (`.prev` kept)
- [x] `dockyrd update --rollback` restores `.prev` binary on Linux
- [x] `dockyrd_UPDATE_DRY_RUN=1` or `--dry-run` skips install
