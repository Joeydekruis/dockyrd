# MCP authentication pattern (Dockyrd)

Binding spec: **ADR-0002** (`docs/adr/0002-dockyrd-mcp.md`). User-facing steps: `docs/user/mcp.md`.

## Client matrix

| Client | Transport | Auth |
|--------|-----------|------|
| Cursor (remote) | Streamable HTTP `POST /mcp` | OAuth 2.1 + PKCE (primary) |
| Claude Code / Desktop (remote) | Streamable HTTP `POST /mcp` | OAuth 2.1 + PKCE (primary) |
| `dockyrd mcp --stdio` | stdio | `DOCKYRD_URL` + `DOCKYRD_TOKEN` (`dyk_*`) |
| CI / scripts | REST or MCP optional | Project API token; MCP token bridge off by default |

## HTTP request shape (Streamable HTTP)

Per MCP 2026-07-28:

- One URL path (default `/mcp`), **POST only** for client→server JSON-RPC.
- Headers: `Accept: application/json, text/event-stream`, `Authorization: Bearer <access_token>`, valid `Origin` when sent by browser-adjacent clients.
- Response: single JSON body or SSE stream scoped to that request.

## OAuth surfaces (same origin as panel)

| Path | Role |
|------|------|
| `GET /.well-known/oauth-protected-resource` | RFC 9728 metadata; links MCP resource to AS |
| `GET/POST /oauth/authorize` | Login + consent (reuse session after MFA) |
| `POST /oauth/token` | Code exchange (PKCE) |
| `POST /mcp` | MCP resource server |

Access tokens MUST be validated on every MCP POST (audience = this instance’s MCP resource URL).

## Scope → RBAC

| Scope | Token / session check |
|-------|------------------------|
| `mcp:read` | `ActionRead` + project scope where applicable |
| `mcp:deploy` | `ActionDeploy`, not `viewer` |
| `mcp:write` | `ActionWrite` |
| `mcp:instance:read` | `requireAdmin` equivalent |

Map OAuth scopes to sqlc audit entries for mutating tools.

## Implementation notes

- Implement tools in Go calling existing domain code in `internal/` (deploy jobs, log tailers, settings loaders)—avoid duplicating HTTP handler auth splits (`requireUser` vs `requireProjectScope`).
- Redact secrets in tool results using the same helpers as API responses.
- Feature work: `mcp-reviewer` + `security-reviewer` required; document new env toggles in `docs/project/setup.md`.

## Spec links

- Authorization: https://modelcontextprotocol.io/specification/2025-11-25/basic/authorization
- Streamable HTTP: https://modelcontextprotocol.io/specification/2026-07-28/basic/transports/streamable-http
