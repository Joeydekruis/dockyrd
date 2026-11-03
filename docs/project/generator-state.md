# Generator handoff

_Last updated: 2026-11-03_

## Current focus

**M2 (#19–#22)** backend + minimal UI; build guards, dy-buildkit on serve, prune, installation webhook, `/api/v1/stream?topics=`.

## Blockers

Docker-host E2E + GitHub App credentials for full #21 push deploy proof.

## Next

- Commit/push `development` → GitHub evidence per #19–#22 (maintainer closes M2 after smoke).
- Run `./scripts/verify-m2-smoke.sh` on a Docker host with `DOCKYARD_PUBLIC_URL`.
