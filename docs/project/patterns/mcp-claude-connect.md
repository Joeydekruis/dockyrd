# MCP OAuth — Claude custom connector checklist

Binding: [Claude connector authentication](https://claude.com/docs/connectors/building/authentication), `docs/project/patterns/mcp-auth.md`, ADR-0002.

## Discovery (must pass before “register” or sign-in)

| Check | Requirement |
|-------|-------------|
| `POST /mcp` without Bearer | **401** + `WWW-Authenticate` with `resource_metadata="…/oauth-protected-resource/mcp"` |
| PRM | `GET /.well-known/oauth-protected-resource/mcp` → JSON; `resource` = `{origin}/mcp` (exact MCP URL in Claude) |
| AS metadata | `GET /.well-known/oauth-authorization-server` (or `openid-configuration`) → `client_id_metadata_document_supported: true`, `token_endpoint_auth_methods_supported` includes `"none"`, `code_challenge_methods_supported` includes `S256`, `registration_endpoint` present |
| Issuer | `issuer` and `authorization_servers[0]` match the HTTPS origin users type in Claude |
| UI | Unknown `/.well-known/*` must **not** return dashboard HTML (404) |

## Client registration

| Method | When |
|--------|------|
| **CIMD** (preferred) | `client_id` = Anthropic HTTPS metadata URL; server fetches JSON, validates `redirect_uris`, loopback port-insensitive match |
| **DCR** (fallback) | `POST /oauth/register` (`application/json`) when CIMD not selected; in-memory client until panel restart |

Pinned Anthropic metadata used when Cloudflare blocks server-side fetch.

## Redirect URIs (must accept)

- `https://claude.ai/api/mcp/auth_callback` (and `https://claude.com/api/mcp/auth_callback`)
- `http://localhost:{port}/callback` and `http://127.0.0.1:{port}/callback` (port ignored vs registered template)

## Token endpoint

- `Content-Type: application/x-www-form-urlencoded`
- PKCE `code_verifier` + `authorization_code`
- `client_id` and `redirect_uri` must match authorize step when sent

## Not implemented (v1 — intentional gaps)

- **Refresh tokens** / `offline_access` (not in `scopes_supported`; access tokens ~1h)
- **DCR persistence** across `dockyrd` restarts
- **RFC 8707 `resource`** parameter on authorize
- Consent screen is GET `consent=1` (approve link); no separate POST consent form

## Ops

- **Public URL** in Server settings must match Claude MCP URL host/scheme (or rely on HTTPS upgrade when request is TLS).
- Panel egress HTTPS to `claude.ai` for CIMD fetch (or pins).
- Allow **Anthropic egress** `160.79.104.0/21` if WAF blocks OAuth discovery.

## Verify on an instance

```bash
PANEL=https://YOUR_PANEL
curl -sS "$PANEL/.well-known/oauth-authorization-server" | jq '{cimd:.client_id_metadata_document_supported,register:.registration_endpoint,pkce:.code_challenge_methods_supported,none:(.token_endpoint_auth_methods_supported|index("none"))}'
curl -sS "$PANEL/.well-known/oauth-protected-resource/mcp" | jq '{resource:.resource,as:.authorization_servers}'
curl -sS -o /dev/null -w "%{http_code}\n" -X POST "$PANEL/mcp" -H 'Content-Type: application/json' -d '{}'
```

Expect: `cimd=true`, `register` set, `pkce` contains `S256`, `none` ≥ 0, `resource` ends with `/mcp`, last line `401`.
