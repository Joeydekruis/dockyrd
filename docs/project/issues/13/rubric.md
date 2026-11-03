---
github: https://github.com/Joeydekruis/dockyrd/issues/13
reviewers: api-reviewer (lean)
---

## Done when

- All spec AC checked with test or handler evidence
- `go test ./internal/server/...` green
- No secrets logged; auth required on mutating project routes

## Gates

```bash
go test ./internal/server/...
go test ./internal/db/...
```
