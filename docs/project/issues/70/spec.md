---
github: https://github.com/Joeydekruis/dockyrd/issues/70
title: Docker Compose kind (C-1)
track: Feature
size: XL
---

## Goal

Run multi-container stacks as a first-class `compose` service kind with dockyrd-managed deploy and edge proxy.

## Acceptance criteria

1. Migration `00011_compose.sql` adds `kind=compose` and domain `target_service` / `target_port`.
2. Deploy pipeline runs `docker compose` with mutated compose file (no host port binds).
3. API: validate YAML, list containers; create compose service (admin).
4. Proxy sync routes compose hostnames to stack network upstreams.
5. Dashboard: project create stack + `/stacks/{slug}` detail.

## Verify

`go test ./internal/compose/... ./internal/db/... ./internal/deploy/...`, `pnpm --dir web exec tsc --noEmit`
