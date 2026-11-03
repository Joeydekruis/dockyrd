---
github: https://github.com/Joeydekruis/dockyrd/issues/15
title: "[M1] Caddy config renderer, platform domains, and domains API"
milestone: M1 · Sprint 1 — Projects API
size: M
---

## Goal

`internal/proxy` renders full Caddy JSON and can POST `/load` on the admin unix socket; platform `sslip.io` hostnames; REST to attach domains to app services.

## Acceptance criteria

- [x] `proxy.Render` builds Caddy JSON from app reverse-proxy routes (briefing bijlage B)
- [x] `proxy.Load` POSTs config to admin unix socket
- [x] `PlatformHostname(serviceSlug, publicIPv4)` → `<slug>-<rand4>.<ip-dashes>.sslip.io`
- [x] sqlc domains + minimal app `POST /api/v1/services` (image stub config) to attach domains in tests
- [x] `GET|POST /api/v1/services/{id}/domains`, `DELETE /api/v1/domains/{id}`
- [x] Go tests: render golden fragment, platform hostname, domains API flow

## References

- `docs/BRIEFING.md` §3.5, bijlage B, §10.2 domains
