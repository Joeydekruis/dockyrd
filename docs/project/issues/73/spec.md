---
title: Terminal v2 (host, themes, copy)
github: https://github.com/Joeydekruis/dockyrd/issues/73
---

## Acceptance

- [x] Owner-only host shell when `terminal_allow_host` enabled; 2FA to enable (if TOTP on) and per host session
- [x] Admin+ container shells unchanged
- [x] Four themes + settings modal; persisted in localStorage
- [x] Copy selection as plain text (button + Ctrl/Cmd+Shift+C)
- [x] Termius-style tabs, menu groups, ROOT warning banner
- [x] Test sandbox shell — Alpine ephemeral container, project/open vs strict link networks, removed on session end
- [x] Terminal theme per user in SQLite — `GET|PATCH /api/v1/terminal/preferences` (+ localStorage cache)

Briefing: `docs/project/briefings/terminal-web.md`
