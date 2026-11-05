---
github: https://github.com/Joeydekruis/dockyrd/issues/65
title: Notifications F3 — disk, cert, thresholds UI
track: Feature
size: M
depends: "#64"
---

## Goal

Disk and certificate watchers with configurable thresholds; admin UI on Server → Channels.

## Acceptance criteria

1. `watch.Disk` / `watch.Certs` emit stateful alerts with dedup + reminder from settings.
2. PATCH `/api/v1/settings/notifications` for disk/cert threshold percentages.
3. Channels settings page shows threshold form.

## Verify

`go test ./internal/watch/...`
