---
github: https://github.com/Joeydekruis/dockyrd/issues/62
title: Notifications F0 — hardening
track: Direct
size: S
parent: docs/project/NOTIFICATIONS_BRIEFING.md
---

## Goal

Unblock notification channel work (#63): safe SMTP, SSRF helpers, fix false `interrupted` deploy marking.

## Assumptions (briefing §16)

All briefing defaults accepted: instance-wide channels, stdlib API routes, F0 before F1.

## Acceptance criteria

1. **`internal/mail`**: dial timeout 10s, total deadline 30s on SMTP session; strip `\r`/`\n` from header values; `mime.QEncoding` subject; `Date` + `Message-ID`; STARTTLS on port 587 when `UseTLS`; `PlainAuth` only when username non-empty; unit tests (CRLF injection, hang timeout with fake server).
2. **`internal/netsafe`**: shared `http.Client` factory with no redirects, dial-time private IP block, `RedactURL(err, secretURL)` helper; SSRF table tests.
3. **`reconcile.Loop`**: `MarkInterruptedDeployments` runs **once** on first tick (startup recovery), not on every 30s tick; regression test: deployment stays `building` across multiple ticks after first pass.

## Out of scope

notify package, migrations, UI.

## Verify

```bash
go test ./internal/mail/... ./internal/netsafe/... ./internal/reconcile/...
make test
```
