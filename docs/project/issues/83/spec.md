---
github: https://github.com/Joeydekruis/dockyrd/issues/83
title: MCP — Streamable HTTP + OAuth (Claude/Cursor)
size: L
---

## Summary

Implement **Dockyrd MCP** in the Go binary: remote agents use **Streamable HTTP** (`POST /mcp`) with **OAuth 2.1** (RFC 9728 PRM, authorization code + PKCE). Local editors may use **`dockyrd mcp --stdio`** with `DOCKYRD_URL` + `DOCKYRD_TOKEN`.

**Binding:** ADR-0002, `docs/user/mcp.md`, `docs/project/patterns/mcp-auth.md`.

**Assumptions:** Single local server MVP; MCP tools call `internal/` domain logic with unified auth (not session-only HTTP gaps). Instance flag default **MCP enabled**; optional **allow API token for MCP**.

## Design artefacts

**Skip designer** — Integrations page follows existing settings patterns (§9 tokens/settings).

## v1 scope

### Transport & OAuth

- [ ] `GET /.well-known/oauth-protected-resource` (`resource` = `{origin}/mcp`)
- [ ] `GET /.well-known/oauth-authorization-server` (issuer = panel origin)
- [ ] `GET /oauth/authorize` — requires dashboard session; consent for requested scopes; redirect with code
- [ ] `POST /oauth/token` — `grant_type=authorization_code`, PKCE S256
- [ ] Opaque access tokens (hashed at rest), TTL configurable (default 1h)
- [ ] `POST /mcp` — `mcp.NewStreamableHTTPHandler` + `auth.RequireBearerToken`
- [ ] Bearer verifier: OAuth access token **or** `dyk_*` when `mcp_allow_api_token` setting true
- [ ] Origin validation on `/mcp` (403 on bad Origin when header present)

### MCP tools (minimum)

| Tool | Scope |
|------|--------|
| `dockyrd_list_projects` | `mcp:read` |
| `dockyrd_list_services` | `mcp:read` |
| `dockyrd_get_service` | `mcp:read` |
| `dockyrd_list_deployments` | `mcp:read` |
| `dockyrd_get_deployment` | `mcp:read` |
| `dockyrd_get_service_logs` | `mcp:read` |
| `dockyrd_trigger_deploy` | `mcp:deploy` |

Tool outputs: JSON text content; no secret values.

### CLI

- [ ] `dockyrd mcp --stdio` — stdio transport, env token auth

### UI

- [ ] `web/app/(shell)/settings/integrations/page.tsx` — MCP URL, OAuth scope table, link to public docs, “Connect” instructions for Cursor/Claude

### Docs

- [x] ADR + user guide (pre-merged); update `docs/user/mcp.md` if URLs/paths change

## Out of scope (v1)

- Full instance settings write via MCP
- Client ID Metadata Documents (use fixed public client IDs in consent copy + PKCE public clients)
- In-panel AI chat
- `mcp:instance:read` tools (defer)

## Related ADRs

- **ADR-0002** (proposed → accept on ship)

## Change log

| Date | Note |
|------|------|
| 2026-11-11 | Initial spec |
