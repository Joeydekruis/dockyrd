---
github: https://github.com/Joeydekruis/dockyrd/issues/14
title: "[M1] Docker client, labels, and network bootstrap"
milestone: M1 · Sprint 1 — Projects API
size: M
---

## Goal

Thin `internal/docker` wrapper: moby client, dockyrd label contract, deterministic resource names, idempotent `dy-edge` / project network ensure.

## Acceptance criteria

- [x] `internal/docker` package with label keys/helpers per briefing §3.3–3.4
- [x] Naming helpers: edge network, project network, app container name
- [x] Client connects via `DOCKER_HOST` or default unix socket; `Ping` for health
- [x] `EnsureEdgeNetwork`, `EnsureProjectNetwork` idempotent (create if missing)
- [x] `ListManaged` filters containers with `dev.dockyrd.managed=true`
- [x] Unit tests for labels/names; optional `integration` test when Docker available

## References

- `docs/BRIEFING.md` §3.3–3.4, §5 `internal/docker`
