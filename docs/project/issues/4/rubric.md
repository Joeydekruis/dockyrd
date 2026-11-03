---
github: https://github.com/Joeydekruis/dockyrd/issues/4
reviewers: database-reviewer
---

## Done when

- Migrations apply on empty DB; sqlc code compiles
- `go test ./internal/db/...` green
- No string-built SQL outside sqlc
