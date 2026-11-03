---
github: https://github.com/Joeydekruis/dockyrd/issues/25
title: "[M3] Install Garage (dy-garage system container)"
milestone: M3 · Sprint 3 — Garage install
size: M
---

## Goal

One-click Garage install: start `dy-garage`, persist config under `{DataDir}/garage`, store admin credentials encrypted in `settings`.

## Acceptance criteria

- [x] `GET /api/v1/storage` — installed flag, version, internal endpoint
- [x] `POST /api/v1/storage/install` with `{provider:"garage"}` idempotent; starts container on `dy-edge`
- [ ] Generated `rpc_secret` + admin token; layout assignment via admin API
- [x] Reconcile ensures running when installed
- [x] Go tests for settings round-trip and install handler validation (mock admin client)

## References

- `docs/BRIEFING.md` §11.6
