---
github: https://github.com/Joeydekruis/dockyrd/issues/72
reviewers:
  - security-reviewer
  - api-reviewer
---

# Rubric #72

## Done when

All spec AC checked; `go test ./internal/terminal/...` and `go test ./internal/server/...` (terminal tests) pass; web typecheck passes.

## Reviewers

| Agent | Why |
|-------|-----|
| security-reviewer | WebSocket + docker exec trust boundary |
| api-reviewer | New `/api/v1/terminal/*` contract |

## Verify

```bash
go test ./internal/terminal/...
go test ./internal/server/ -run Terminal
pnpm --dir web exec tsc --noEmit
```
