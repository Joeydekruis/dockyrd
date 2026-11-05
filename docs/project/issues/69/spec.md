---
github: https://github.com/Joeydekruis/dockyrd/issues/69
title: Preview deployments
---

## Acceptance criteria

- [ ] Service config `preview.enabled`; GitHub `pull_request` webhook syncs child services per PR.
- [ ] API: list/delete previews, PATCH preview-config on parent service.
- [ ] Application detail: Previews tab + preview settings on Settings tab.
- [ ] sqlc preview queries + migrations `00010_previews.sql` (with compose `00011` ordering preserved).
- [ ] Test: `internal/server/api_previews_test.go` smoke.
