---
github: https://github.com/Joeydekruis/dockyrd/issues/17
title: "[M1] Encrypted env vars, runtime logs, SSE stream"
milestone: M1 · Sprint 3 — Logs, env & Projects UI
size: M
---

## Goal

Service environment variables at rest (AES-GCM), runtime log fetch, and authenticated SSE hub for realtime events.

## Acceptance criteria

- [x] `GET|PUT /api/v1/services/{id}/variables` with encrypted storage
- [x] `GET /api/v1/services/{id}/logs` (runtime, docker logs when live)
- [x] `GET /api/v1/stream` SSE with session auth
- [x] `internal/crypto` Encrypt/Decrypt helpers

## References

- `docs/BRIEFING.md` §10.2 variables/logs, §11.3
