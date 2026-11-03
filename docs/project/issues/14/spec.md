---
github: https://github.com/Joeydekruis/dockyrd/issues/14
title: "[M1] Docker client, labels, and network bootstrap"
milestone: M1 · Sprint 1 — Projects API
size: M
---

## Goal

Thin `internal/docker` wrapper: moby client, Dockyard label contract, deterministic resource names, idempotent `dy-edge` / project network ensure.

## Acceptance criteria

- [ ] `internal/docker` package with label keys/helpers per briefing §3.3–3.4
- [ ] Naming helpers: edge network, project network, app container name
- [ ] Client connects via `DOCKER_HOST` or default unix socket; `Ping` for health
- [ ] `EnsureEdgeNetwork`, `EnsureProjectNetwork` idempotent (create if missing)
- [ ] `ListManaged` filters containers with `dev.dockyard.managed=true`
- [ ] Unit tests for labels/names; optional `integration` test when Docker available

## References

- `docs/BRIEFING.md` §3.3–3.4, §5 `internal/docker`
