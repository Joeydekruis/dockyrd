---
github: https://github.com/Joeydekruis/dockyrd/issues/68
title: API tokens + dy CLI
---

## Acceptance criteria

- [ ] User API tokens (`read` / `deploy` / `write`) scoped to a project; CRUD at `/api/v1/tokens` (session only).
- [ ] Bearer user tokens authenticate selected GET/POST routes; default deny elsewhere; deploy hook tokens unchanged.
- [ ] `GET /api/v1/deployments/{id}` with read scope or session.
- [ ] `cmd/dy`: login, whoami, services list, deploy, logs.
- [ ] Account settings `/settings/account/tokens` wired to token API.
- [ ] Tests: `internal/apitoken`, `internal/server` token/bearer smoke.
