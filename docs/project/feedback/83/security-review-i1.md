# Security review - 83 i1
## Verdict: BLOCK
## Boundaries
- OAuth authorize/token: PKCE S256 required; redirect allowlist (loopback, panel host, `cursor`/`claude`); codes/tokens stored as SHA-256 hashes
- `POST /mcp`: Bearer verification + optional `dyk_*` bridge gated by `mcp_allow_api_token`; Origin checked when header present
- Tool RBAC: scope checks + `assertProjectAccess` for project-scoped API tokens (stdio path)
- Login: `/login` → `/sign-in?return=…`; OAuth uses relative `/oauth/authorize?…` return (client `safeReturnPath`)
- Tests: `go test ./internal/server/ -run MCP -count=1` — pass

## Blockers
- `internal/mcp/oauth/server.go:279-310` + `internal/mcp/tools.go:29-47` — HTTP `dyk_*` bridge returns user + scopes but never loads `ApiToken` into context, so `assertProjectAccess` skips `project_id` binding → project API tokens gain instance-wide MCP when `mcp_allow_api_token` is true → mirror REST `requireProjectScope` (attach token row to principal or reject non–user-scoped tokens on `/mcp`)
- `internal/mcp/tools.go:64-76` — `dockyrd_list_projects` returns all projects without project filter; project-scoped `dyk_*` (stdio) can enumerate the whole instance → restrict list to `APIToken.ProjectID` when set (per `mcp-auth.md`)

## Notes
- `POST /oauth/token` does not re-validate `redirect_uri`/`client_id` (PKCE mitigates); consider binding on exchange for defense in depth
- `mcpOriginOK` allows missing `Origin` (non-browser MCP); trusts `X-Forwarded-Host` like other panel routes — ensure reverse proxy strips untrusted forwarded headers
