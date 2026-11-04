---
github: https://github.com/Joeydekruis/dockyrd/issues/46
title: "[M4+] Container metrics collector + analytics API"
size: M
---

## Goal

Container-level metrics per briefing §11.4 and enriched analytics APIs.

## Acceptance criteria

- [x] `RunContainerCollector` → `metric_container` (15s tick, minute buckets)
- [x] `GET /api/v1/server/metrics/containers?range=`
- [x] `GET /api/v1/server/containers` enriched (project label, cpu, memory)
- [x] Live swap + per-core CPU on `GET /api/v1/server` (Linux)
- [x] Host metrics `30d` range + `net_rx`/`net_tx` in JSON

## Verify

`go test ./internal/metrics/... ./internal/docker/... ./internal/server/...`
