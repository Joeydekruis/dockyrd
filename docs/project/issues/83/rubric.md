---
github: https://github.com/Joeydekruis/dockyrd/issues/83
---

## Done bar

- All v1 checkboxes in `spec.md` with tests cited in GitHub handoff.
- `go test ./internal/mcp/... ./internal/server/ -run MCP` green.
- `make build` green; no secrets in tool output tests.

## Reviewers

| Subagent | Required | Focus |
|----------|----------|--------|
| `mcp-reviewer` | **Yes** | Streamable HTTP, tool schemas, OAuth PRM / 401 |
| `security-reviewer` | **Yes** | PKCE, token storage, Origin, RBAC on tools, `dyk_*` bridge |
| `api-reviewer` | Light | OAuth route shapes |
| `frontend-evaluator` | Skip | S settings page, existing patterns |
| `code-reviewer` | Skip | Specialists cover surface |

## Gates

```bash
go test ./internal/mcp/... ./internal/server/...
make build
pnpm --dir web exec tsc --noEmit
```
