---
github: https://github.com/Joeydekruis/dockyrd/issues/29
title: "[M4] Team users and invites API"
milestone: M4 · Sprint 2 — Team & audit (planned)
size: M
---

## Goal

Admin-only team management: list users, invite by email, audit log for mutations, settings Team page.

## Acceptance criteria

- [x] `GET /api/v1/users`, `PATCH /api/v1/users/{id}` (role), admin RBAC
- [x] `GET/POST /api/v1/invites` — token returned once on create
- [x] `GET /api/v1/audit` — recent entries
- [x] `/settings/team` UI — users, pending invites, invite form
- [x] `go test ./internal/server/...` includes team smoke test

## Out of scope (later)

- Invite accept flow, 2FA login step, session revoke UI
