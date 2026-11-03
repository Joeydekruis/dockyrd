---
github: https://github.com/Joeydekruis/dockyrd/issues/20
title: "[M2] Railpack prepare + BuildKit build"
milestone: M2 · Sprint 2 — Railpack builds
size: M
---

## Goal

Railpack auto-detect builds when builder is railpack or no Dockerfile is present.

## Acceptance criteria

- [x] `railpack prepare` + `docker buildx build --load` with pinned `BUILDKIT_SYNTAX` frontend
- [x] Build secrets via `--secret id=K,env=K` for at_build variables
- [x] Integrated in shared `RunGitAppBuild` pipeline
- [x] `dy-buildkit` container + buildx builder ensured when Docker is available on serve
