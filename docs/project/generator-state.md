# Generator state

**Updated:** 2026-11-11

## In progress

**#83 MCP (Claude/Cursor)** — generator i1 on `development`.

- OAuth PRM + authorization server metadata, `/oauth/authorize` + `/oauth/token` (PKCE)
- `POST /mcp` Streamable HTTP (`modelcontextprotocol/go-sdk` v1.8.0)
- Tools: list projects/services, deployments, logs tail, trigger deploy
- `dockyrd mcp --stdio`
- Settings → Server → Integrations: MCP section
- Tests: `TestMCPProtectedResourceMetadata`, `TestMCPOAuthPKCERoundtrip`
- **Next:** `mcp-reviewer` + `security-reviewer`; instance toggles in API/UI; login `return` URL support; handoff #83

## Last ship

Prior: **#77** security policy; **M7** epic slices.
