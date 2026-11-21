---
github: https://github.com/Joeydekruis/dockyrd/issues/79
title: P2a — Dockyrd proxybuild & proxy.swap
size: L
parent: https://github.com/Joeydekruis/dockyrd/issues/74
milestone: M7 · Security & performance
---

## Summary

Ship **P2a**: signed **proxybuild** image pipeline (xcaddy + Coraza + rate limit modules) and real **`proxy.swap`** host action with validate → blue/green container swap → revert on failure.

**Product source:** `docs/project/SECURITY_BRIEFING.md` §4.4, §9 P2a.

## In scope

- `packaging/proxybuild/Dockerfile` (xcaddy build)
- `.github/workflows/proxybuild.yml` — multi-arch push to `ghcr.io/joeydekruis/dockyrd-proxy`, cosign sign on tags
- `internal/proxybuild` — manifest, cosign verify helper, version → image ref
- `internal/docker` — proxy image state, `ValidateCaddyJSON`, `SwapProxyImage`
- `internal/hostops/proxy_swap.go` — replace stub
- Add-on **`proxybuild`** installable via existing addon job (`HostAction: proxy.swap`)
- API: `proxybuild_installed` on policy + service security payloads
- Migration `00017_proxybuild_addon.sql`

## Out of scope (P2b/c)

- `rate_limit` / `coraza_waf` in Caddy render
- Separate `proxy-waf` / `proxy-rate-limit` addon install flows

## Acceptance criteria

1. Release tag builds and signs `ghcr.io/joeydekruis/dockyrd-proxy:v{version}`.
2. `proxy.swap` Plan shows current + target image; Apply pulls, validates rendered JSON with new image, swaps dy-proxy, retains old container until new admin socket + load succeeds; Revert restores previous image.
3. Installing **proxybuild** add-on runs `proxy.swap` and marks addon installed.
4. Policy/UI treat `proxybuild_installed` as proxy ready for traffic-protection gating (controls still inert until P2b/c handlers).