---
github: https://github.com/Joeydekruis/dockyrd/issues/21
title: "[M2] GitHub App manifest, webhooks, auto-deploy"
milestone: M2 · Sprint 3 — GitHub App
size: M
---

## Goal

Link a GitHub App via manifest flow; auto-deploy matching services on push webhooks.

## Acceptance criteria

- [x] `POST /api/v1/github/app/manifest`, `GET /api/v1/github/app/callback`, `GET /api/v1/github/status`
- [x] `POST /api/v1/hooks/github` with HMAC verification when webhook secret stored
- [x] Push events enqueue deploy for `source.type=github` + `autoDeploy` + repo/branch match
- [x] Credentials encrypted in `git_sources`
- [x] `installation` webhook updates `installation_id` when app is installed on a server
