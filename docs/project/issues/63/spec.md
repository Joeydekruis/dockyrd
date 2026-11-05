---
github: https://github.com/Joeydekruis/dockyrd/issues/63
title: Notifications F1 — core + channels + deploy
track: Feature
size: L
parent: docs/project/NOTIFICATIONS_BRIEFING.md
blocks: "#64"
depends: "#62"
---

## Goal

Instance-wide alert channels (Slack, Discord, generic webhook) with outbox dispatcher; `deploy.failed` / `deploy.succeeded` via `notify.Emit`.

## Source of truth

`docs/project/NOTIFICATIONS_BRIEFING.md` §4–7, §10–11 (API + UI list/modal/test/delivery log), §13 F1.

**Migration:** `00007_notifications.sql` (00006 is service_links).

## Acceptance criteria (F1)

See briefing §13 F1 — deploy alert within 10s, retries/429/404 behavior, kill -9 recovery, secrets/SSRF tests, RSS +2MB, admin API + `/settings/server/channels` UI, `docs/user/notifications.md`.

## Reviewers

- `security-reviewer` (SSRF, secrets)
- `database-reviewer` (migration, claim query)
- `api-reviewer` (optional, light)

## Design

Reuse email-settings / ResourceRow patterns; no designer wireframe (briefing §11.5).
