---
github: https://github.com/Joeydekruis/dockyrd/issues/47
title: "Analytics dashboard ECharts tabs + design parity"
size: L
---

## Goal

Analytics page with Overview + CPU/Memory/Disk/Network tabs, ECharts, Storybook primitives.

## Acceptance criteria

- [x] Tabs, time range 1h–30d, server meta chips
- [x] ECharts theme `#3B93F5` / alert orange / navy outbound
- [x] Per-core CPU, top containers, running table from live API
- [x] Events / cleanup on Overview

## Verify

`pnpm --dir web exec tsc --noEmit` · `./scripts/check-no-radix.sh`

Refs #46
