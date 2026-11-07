---
github: https://github.com/Joeydekruis/dockyrd/issues/2
title: "[foundation] Go module and CLI skeleton"
milestone: M0 · Sprint 1
size: S
---

## Goal

Runnable `dockyrd` binary with CLI entrypoints and minimal API health check.

## Acceptance criteria

- [ ] `go.mod` — `github.com/Joeydekruis/dockyrd`, Go ≥ 1.25
- [ ] `cmd/dockyrd/main.go` — subcommands `serve`, `version`, `init` (`init` prints not-implemented OK for now)
- [ ] `internal/server` — `GET /api/v1/health` → `{"status":"ok"}`
- [ ] Unit test for health JSON handler
- [ ] `go test ./...` passes

## References

- `docs/BRIEFING.md` §5
