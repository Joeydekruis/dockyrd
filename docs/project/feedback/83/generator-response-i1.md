# Generator response - 83 i1

Addressed MCP + security review blockers:

- Service tools: `dbkind.RedactConfig` on list/get
- HTTP `dyk_*`: `TokenInfo.Extra["api_token"]` → `principal()` project binding
- `dockyrd_list_projects`: single project when token is project-scoped
- OAuth: consent HTML before code issue (`consent=1`)
- `redirect_uri` allowlist (loopback + cursor/claude schemes)
- PRM `resource` field; sign-in `?return=`; instance MCP toggles in UI

Tests: `go test ./internal/server/ -run 'MCP|Login' -count=1` (4 tests, pass).
