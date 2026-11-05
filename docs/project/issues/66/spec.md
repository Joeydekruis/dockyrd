---
github: https://github.com/Joeydekruis/dockyrd/issues/66
title: Notifications F4 — heartbeat and optional alerts
track: Feature
size: S
depends: "#65"
---

## Goal

External heartbeat ping URL; optional future: crash loop, digest, update alerts.

## Acceptance criteria

1. `heartbeat_url` in notification settings; `watch.RunHeartbeat` GET every 5m via netsafe client.
2. UI field on Server → Channels.

## Out of scope (follow-up)

`service.crash_loop`, update alerts, digest, Test all channels.
