---
github: https://github.com/Joeydekruis/dockyrd/issues/19
title: "[M2] Git clone + Dockerfile build job"
milestone: M2 · Sprint 1 — Dockerfile builds
size: M
---

## Goal

Build app services from a public Git URL and Dockerfile before the existing blue/green deploy runtime runs.

## Acceptance criteria

- [x] `AppConfig` supports `source.type=git` (`url`, `branch`, `rootDir`) and `builder=dockerfile` (`dockerfile` path, default `Dockerfile`)
- [x] Deploy job clones shallow into `{DataDir}/builds/<deployment>/src`, runs `docker buildx build --load`, tags `dockyard/<service-slug>:<deploy-short>`
- [x] Deployment status includes `building`; `build_log_path` set on success or failure
- [x] Build-time variables (`at_build`) passed as build-args; values scrubbed in log file
- [x] Source cleanup of build workspace after build (success or fail)
- [x] Go unit tests for config resolution; deploy test still passes for image source
- [x] Build timeout (default 20m) and disk guard (default 10% free on data volume)

## References

- `docs/BRIEFING.md` §11.1 (dockerfile branch)
- `internal/deploy/runtime.go`
