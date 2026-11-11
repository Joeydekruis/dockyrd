# MCP review - 83 i1
## Verdict: BLOCK
## Scope
- tools: `dockyrd_list_*`, `dockyrd_get_*`, `dockyrd_trigger_deploy`; stdio + `POST /mcp` streamable handler
- OAuth: PRM, authorization server metadata, authorize, token; bearer verifier + optional `dyk_*` bridge
- CLI: `dockyrd mcp` → `RunStdio`
- Schema/DB: `mcp_oauth_codes`, `mcp_access_tokens` (hashed BLOB PKs)
## Gates
- `go test ./internal/server/ -run MCP -count=1` — pass (3 tests; no `POST /mcp` or tool-call coverage)
- `go test ./internal/mcp/...` — no test files
## Blockers
- `internal/mcp/tools.go:125-129,94-98` — `dockyrd_get_service` / `dockyrd_list_services` JSON-marshal raw `sqlc.Service` (`config` includes secrets); spec + rubric require redaction → apply `dbkind.RedactConfig` like `serviceFromRow` in `api_services.go`
- `internal/mcp/oauth/server.go:279-310` + `internal/mcp/tools.go:29-47` — HTTP `dyk_*` bridge sets `TokenInfo` only; `principal()` never loads `APIToken`, so project binding in `assertProjectAccess` is skipped when `mcp_allow_api_token` is true → attach token row (or reject project-scoped tokens on `/mcp`) per `mcp-auth.md`
- `internal/mcp/tools.go:64-76` — `dockyrd_list_projects` lists entire instance; project-scoped `dyk_*` (stdio + HTTP bridge) should match REST `handleListProjects` token branch
- `internal/mcp/oauth/server.go:111-154` — `/oauth/authorize` issues codes immediately after session check with no consent step (`mcp-auth.md` / spec AC); CSRF-style forced grant for logged-in users
## Notes
- PRM + AS metadata + PKCE roundtrip tested; add lean `httptest` for `POST /mcp` 401 (with `resource_metadata`) and Origin 403
