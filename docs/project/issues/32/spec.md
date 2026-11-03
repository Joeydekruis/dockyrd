---
github: https://github.com/Joeydekruis/dockyrd/issues/32
title: "[M4] TOTP 2FA, MFA login, profile sessions"
milestone: M4 · Sprint 2 — Team & audit (planned)
size: L
---

## Acceptance criteria

- [x] TOTP setup/enable with encrypted secret + 10 recovery codes
- [x] Login returns 202 `mfa_required`; `POST /api/v1/auth/mfa`
- [x] Sessions blocked until MFA passed when 2FA enabled
- [x] Profile: enable 2FA, list/revoke sessions
- [x] Sign-in MFA step; Team audit log snippet (admin)
- [x] `TestLoginMFAFlow`
