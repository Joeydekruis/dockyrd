---
github: https://github.com/Joeydekruis/dockyrd/issues/11
title: "[foundation] Sign-in, setup wizard, and auth API"
milestone: M0 · Sprint 4
size: L
---

## Goal

End-to-end owner setup and login into the empty dashboard.

## Acceptance criteria

- [ ] `POST /api/v1/setup`, `POST /api/v1/auth/login`, `GET /api/v1/auth/me`, `POST /api/v1/auth/logout`
- [ ] Session cookie `dy_session`, argon2id passwords
- [ ] `/sign-in`, `/setup`, `/forgot-password` pages per design auth layout
- [ ] Client auth gate on shell routes

## References

- `docs/BRIEFING.md` §8, route map §10
