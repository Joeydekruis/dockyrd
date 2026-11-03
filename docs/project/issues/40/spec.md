---
github: https://github.com/Joeydekruis/dockyrd/issues/40
---

# #40 · GoReleaser + release workflow

## AC

- [x] `.goreleaser.yaml` — linux amd64/arm64 tarballs, checksums, embed web build hook, `main.version` ldflag
- [x] `.github/workflows/release.yml` on tag `v*` with cosign checksum signing
- [x] `scripts/install.sh` downloads latest GitHub release + verifies checksums.txt when present
- [x] `make release-snapshot` for local dry-run
