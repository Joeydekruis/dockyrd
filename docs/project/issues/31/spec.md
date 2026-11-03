---
github: https://github.com/Joeydekruis/dockyrd/issues/31
title: "[M4] Global search API and ⌘K"
milestone: M4 · Sprint 2 — Team & audit (planned)
size: S
---

## Acceptance criteria

- [x] `GET /api/v1/search?q=` — nav pages + project name/slug match
- [x] Command palette calls API when typing (debounced)
- [x] Tests in `api_invites_public_test.go` (search case)
