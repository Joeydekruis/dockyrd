---
github: https://github.com/Joeydekruis/dockyrd/issues/71
title: Wildcard DNS verify + default platform hostnames
design: docs/project/issues/71/design-direction.md
wireframe: docs/project/issues/71/assets/domains-urls-wireframe.html
---

## Summary

Encourage and verify a wildcard **apps domain** (`*.apps.<apex>`). When verified, default platform URLs use `{slug}-{4hex}.<suffix>`; otherwise sslip.io. Custom domains unchanged. Databases: no Domains UI in this release.

## Design contract

- `docs/project/issues/71/design-direction.md`
- `docs/project/issues/71/assets/domains-urls-wireframe.html` (open in browser; multi-board canvas)

## Acceptance criteria

### Instance settings

- [ ] **Apps on your domain** block per design state machine (no panel URL → verified → failed).
- [ ] Suggest `*.apps.<apex>` from Dashboard URL (public suffix / apex).
- [ ] Wildcard verify: probe hostname (e.g. `dockyrd-check.<suffix>`) DNS → server `public_ipv4`; persist verified state + timestamp on instance settings.
- [ ] Inline DNS helper (A for `*`, server IPv4); optional AAAA when server has IPv6.
- [ ] Primary CTA: **Verify** when unverified; **Save** otherwise. Re-check + change apps domain (confirm → reset unverified).

### Provisioning (Go)

- [ ] `PlatformHostname` uses verified wildcard suffix; else sslip.io (`internal/proxy/platform.go`).
- [ ] Auto-provision primary **default URL** on create for `app`, `static_site`, `compose`.
- [ ] Platform rows under verified wildcard: DNS status active without per-host pending.
- [ ] API: optional second platform row (sslip) when wildcard primary exists.

### Domains UI (app, static_site, compose)

- [ ] Two sections: **Default URL** (platform) and **Your domains** (custom) per wireframe.
- [ ] Badges/copy per design-direction copy pack.
- [ ] Empty app: add default URL; not custom-only.
- [ ] Admin-only nudge banner (dismiss per user); members see no nudge.

### Setup wizard

- [ ] Skippable step **Default URLs for apps** (use my domain / skip for now) per design.

### Settings chrome (from design open point #4)

- [ ] Settings tab row (Profile & security, Git providers, Instance) on all three settings pages, not Instance-only.

## Out of scope

- Public Domains for databases; per-service wildcard patterns.
- Clean `{slug}.apps…` without random suffix (unless product reverses design decision).

## Product assumptions (from design)

- Last platform URL may be deleted (confirm copy in design-direction).
- Changing apps domain keeps existing service URLs until re-verified.
- Custom primary redirects platform hostnames to it.
