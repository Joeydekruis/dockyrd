---
github: https://github.com/Joeydekruis/dockyrd/issues/72
title: Web terminal (container exec)
size: L
---

# Web terminal — container exec

## Goal

Owner/admin users open an interactive shell inside a **running workload container** from Server → Terminal. No host shell.

## Acceptance criteria

- [ ] Sidebar: **Terminal** under Server (owner/admin see page; member/viewer get forbidden UI or redirect).
- [ ] Page matches design language: PageHeader, target picker, connected status, Clear, xterm output.
- [ ] `GET /api/v1/terminal/targets` lists running managed containers (apps, databases, compose members) with kind hints.
- [ ] `POST /api/v1/terminal/sessions` creates a single-use short-lived ticket (≤60s) for WebSocket; body: `service_id`, optional `compose_service`, `mode` (`shell` | `psql` | `valkey-cli`), optional `user` (`root` only for owner).
- [ ] `GET /api/v1/terminal/ws?ticket=` upgrades to WebSocket with **Origin** check (panel origin only).
- [ ] Session bridges stdin/stdout TTY via `docker exec` attach; resize events forwarded when possible.
- [ ] No shell in image (distroless): friendly message in terminal output, no opaque 500.
- [ ] Audit: `terminal.session.start` / `terminal.session.end` with user, service, container, mode (no keystrokes).
- [ ] Limits: 15m idle disconnect, 2h max, max 3 concurrent sessions per user.
- [ ] Database shortcuts: psql on postgres kind, valkey-cli on redis kind.

## Non-goals

Host shell, file upload/download, port-forward, session replay, member/viewer access.

## Security

Single-use ticket; WebSocket Origin validation; admin-only HTTP; cookies still required on ticket minting.
