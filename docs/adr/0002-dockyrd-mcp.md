# ADR-0002: Dockyrd MCP (transport and authentication)

**Status:** accepted  
**Date:** 2026-11-11

## Context

Operators want AI assistants (Claude, Cursor, and similar MCP clients) to debug deploys, read logs and analytics, trigger deploys, and manage projects on a self-hosted Dockyrd instance. The Model Context Protocol (MCP) defines how those clients call tools on a server.

Dockyrd already ships:

- One Go binary that serves `/api/v1` and the embedded dashboard on the same origin.
- Team auth (sessions, TOTP, RBAC) and project-scoped API tokens (`dyk_*`, `read` / `deploy` / `write`) for the `dy` CLI.

MCP clients expect **Streamable HTTP** (protocol revision **2026-07-28**) for remote servers and **stdio** with environment credentials for local subprocess servers. HTTP MCP authorization follows **OAuth 2.1** as an OAuth resource server, with **protected resource metadata** (RFC 9728), **PKCE**, and **Bearer** access tokens on every request—not query-string secrets.

## Decision

### 1. Where MCP runs

- **Remote (primary):** The `dockyrd` HTTP server exposes a single MCP endpoint (default path **`POST /mcp`**) using **Streamable HTTP**. No separate Node sidecar.
- **Local (secondary):** **`dockyrd mcp --stdio`** runs an MCP server on the host that calls the panel API using credentials from the environment (`DOCKYRD_URL`, `DOCKYRD_TOKEN`). Per MCP spec, stdio must **not** use the HTTP OAuth flow.
- **Dashboard:** The UI does not embed an MCP server. It provides **Integrations → AI assistants**: OAuth connect for Claude/Cursor, token management, copy-paste client hints, and an instance toggle to disable MCP.

### 2. Authentication (HTTP — Claude, Cursor, other remote clients)

Dockyrd acts as both **authorization server** and **resource server** on the panel origin (same host as `/api/v1`).

| Mechanism | Use |
|-----------|-----|
| **Protected Resource Metadata** | `GET /.well-known/oauth-protected-resource` names the MCP endpoint and authorization server. |
| **Authorization code + PKCE** | Browser sign-in uses existing Dockyrd login (including MFA when required). |
| **Client registration** | **Client ID Metadata Documents** (MCP 2026-07-28); no dynamic client registration in v1. |
| **Access tokens** | Short-lived JWT or opaque tokens; **`Authorization: Bearer`** on **every** `POST /mcp`. Audience (RFC 8707) must match this instance’s MCP resource. |
| **Scopes** | `mcp:read`, `mcp:deploy`, `mcp:write`, `mcp:instance:read` mapped to the same RBAC rules as REST (`viewer` / `member` / `admin` / `owner`). |
| **401 handling** | Unauthenticated MCP requests return `401` with `WWW-Authenticate: Bearer …, resource_metadata="…"`. |

**Automation / break-glass (optional, not the default for Cursor/Claude UI connect):** Pre-issued `dyk_*` tokens may be accepted at the resource server when the instance admin enables “allow API token for MCP,” for scripts and local stdio only. OAuth remains the recommended path for interactive clients.

### 3. Transport hardening

- Validate **`Origin`** on MCP POSTs; reject invalid origins with **403** (DNS rebinding).
- Rate-limit MCP like other authenticated API surfaces.
- Never return secrets in tool results (variables, SMTP passwords, connection strings—same redaction as the dashboard).

### 4. Tools (v1 scope, illustrative)

Implement tools over shared `internal/` logic (not a blind HTTP proxy), including at minimum:

- **Read:** service/project context, deployment status, log tails, build log, events, service analytics, server metrics (role-gated), instance settings (admin read).
- **Write (scoped):** trigger deploy, patch project/service fields allowed to `write` scope; instance settings mutations owner/admin only if exposed at all in v1.

Exact tool list lives in the Feature issue spec; this ADR only fixes transport and auth.

## Consequences

- New routes: `/mcp`, OAuth metadata, authorize/token endpoints, and dashboard Integrations UI.
- **mcp-reviewer** and **security-reviewer** run on implementation; parity with RBAC and audit logging is mandatory for mutating tools.
- Session-only REST gaps (some log streams, server metrics) are closed inside MCP tool handlers via one auth path, without requiring every SSE route to accept `dyk_*` first.

## Alternatives considered

- **stdio-only MCP:** Rejected for Claude/Cursor remote use; does not match how those clients connect to self-hosted panels.
- **Bearer-only (no OAuth):** Rejected as the primary UX; does not match MCP authorization spec current revisions and is weaker for interactive clients.
- **Separate MCP container:** Rejected; violates single-binary control plane and duplicates auth.

## References

- [MCP Authorization](https://modelcontextprotocol.io/specification/2025-11-25/basic/authorization) (OAuth 2.1 resource server)
- [MCP Streamable HTTP](https://modelcontextprotocol.io/specification/2026-07-28/basic/transports/streamable-http) (2026-07-28)
- User guide: [`docs/user/mcp.md`](../user/mcp.md)
- Implementer notes: [`docs/project/patterns/mcp-auth.md`](../project/patterns/mcp-auth.md)
