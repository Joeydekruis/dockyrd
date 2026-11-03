---
github: https://github.com/Joeydekruis/dockyrd/issues/40
---

# #40 · GoReleaser + release workflow

## AC

- [ ] `.goreleaser.yaml` — linux amd64/arm64 tarballs, checksums, embed web build hook, `main.version` ldflag
- [ ] `.github/workflows/release.yml` on tag `v*` with cosign checksum signing
- [ ] `scripts/install.sh` downloads latest GitHub release + verifies checksums.txt when present
- [ ] `make release-snapshot` for local dry-run
