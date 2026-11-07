# ADR-0001: Docker Compose as a service kind

**Status:** proposed  
**Date:** 2026-11-05

## Context

Self-hosters need existing `docker-compose.yml` stacks without a second orchestrator. dockyrd already manages apps, databases, proxy, and deploy jobs via a single Go binary and SQLite control plane.

## Decision

1. Add `kind=compose` via migration `00011` (SQLite `services` table rebuild with `foreign_keys=OFF`).
2. Run **`docker compose` as a subprocess**; normalize with `docker compose config --format json`.
3. Pipeline: normalize → **policy engine** (deny privileged, host network, docker.sock mounts, `build:` in v1) → mutate (labels, logging, cgroup, strip host `ports`) → `compose.dockyrd.json` → `pull` → `up -d --remove-orphans --wait`.
4. **No blue/green** for stacks; document stateful semantics in UI.
5. **Proxy:** connect `dy-proxy` to each stack network; upstream aliases `dyc-<service_id>-<compose_service_name>` to avoid DNS collisions on shared edge networks.
6. Domains: `target_service` + `target_port` on `domains` for compose upstream selection.

## Consequences

- High-risk migration; must ship with migration tests and pre-migrate backup.
- v1 excludes Compose `build:`, resource links to project DB, previews, and §1 Postgres backups inside stacks.
- Reconciler must re-attach proxy to stack networks after proxy restarts.

## Alternatives considered

- **`app` + compose source:** rejected — wrong health/build/deploy semantics.
- **compose-go library:** rejected — parity and maintenance cost vs CLI.
- **Attach stacks to `dy-edge`:** rejected — duplicate service name aliases across stacks.
