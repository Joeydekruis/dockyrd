---
github: https://github.com/Joeydekruis/dockyrd/issues/3
title: "[foundation] Next.js static export and embed pipeline"
milestone: M0 · Sprint 1
size: M
---

## Goal

Static-export dashboard scaffold embedded in the Go binary via `make build`.

## Acceptance criteria

- [ ] `web/` — Next.js App Router, `output: 'export'`, self-hosted Geist (no Google Fonts CDN)
- [ ] `web/styles/globals.css` — CSS variables from `docs/design-index.json` `design_tokens.colors`
- [ ] Placeholder `/` route with minimal AppShell stub (sidebar wordmark area)
- [ ] `web/embed.go` — `//go:embed all:out`, `FS()` for server
- [ ] `Makefile` — `build`, `test`, `lint` targets; `make build` produces binary
- [ ] `scripts/check-no-radix.sh` — fails on `@radix-ui/*`
- [ ] Server serves embedded UI at `/` when built

## References

- `docs/BRIEFING.md` §4.2, §5
- `docs/design-index.json`
