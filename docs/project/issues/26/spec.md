---
github: https://github.com/Joeydekruis/dockyrd/issues/26
title: "[M3] Buckets, keys, databases & storage UI"
milestone: M3 · Sprint 4 — Buckets & UI
size: L
---

## Goal

`kind=bucket` services, S3 keys via Garage admin API, wire `/databases` and `/storage` dashboard pages to API; M3 exit smoke.

## Acceptance criteria

- [x] Bucket CRUD API + published S3 vars; attach project network to Garage when bucket exists
- [x] Access key create/list/revoke on bucket detail API
- [x] `/databases` lists postgres/redis, create flow, connection reveal
- [x] `/storage` install Garage + bucket list (Garage-only card; no MinIO)
- [x] `scripts/verify-m3-smoke.sh` — go test subset + web build
- [x] E2E path documented: Next app + Postgres + bucket upload (manual checklist)

## References

- `docs/BRIEFING.md` §12, §17 create-flows
- `docs/design-index.json` routes `/databases`, `/storage`
