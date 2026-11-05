---
github: https://github.com/Joeydekruis/dockyrd/issues/64
title: Notifications F2 — email, backup alerts
track: Feature
size: M
depends: "#63"
---

## Goal

Email notification channel, backup failed/succeeded/overdue alerts, inbox dedup for deploy.failed.

## Acceptance criteria

1. Email channel uses instance SMTP with `email_log` kind=notification.
2. `backup.failed` / resolved on success; `backup.overdue` via watch.
3. Deploy.failed does not duplicate inbox + email to same address (skip set).

## Verify

`go test ./internal/notify/... ./internal/backup/... ./internal/watch/...`
