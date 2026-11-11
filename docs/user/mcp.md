# AI assistants (MCP)

Connect **Claude** or **Cursor** to your Dockyrd panel so an assistant can read deploy status, logs, and analytics, trigger deploys, and help change projects and services—using the same permissions as your team account.

**Public docs:** [dockyrd.io/docs/mcp](https://dockyrd.io/docs/mcp.html)

> Requires a build that includes issue [#83](https://github.com/Joeydekruis/dockyrd/issues/83) (Streamable HTTP `/mcp` + OAuth). Older binaries without `/mcp` can still use the [`dy` CLI](./install.md) and REST API tokens.

## What you get

| Capability | Typical use |
|------------|-------------|
| Read deployments, build output, runtime logs | “Why did this deploy fail?” |
| Read analytics and server metrics | Capacity and traffic questions |
| Trigger a manual deploy | Redeploy after a config change |
| Create or update projects and services | Naming, build settings, domains (within your role) |
| Read instance settings | Domains, SMTP, security policy (admins) |

Mutating actions are audit-logged like changes in the dashboard.

## How it works

Dockyrd runs an **MCP server inside the same `dockyrd` process** as the dashboard—not a separate service.

- **Remote clients (Claude, Cursor):** **Streamable HTTP** at `https://YOUR_PANEL/mcp`, with **OAuth 2.1** sign-in through the panel (PKCE, browser login, MFA if your instance requires it).
- **Same machine as the server:** optional **`dockyrd mcp --stdio`** with an API token in the environment (for editors that launch a local MCP subprocess).

The dashboard **Settings → Integrations → AI assistants** page (planned) walks through OAuth connect and shows the MCP URL for your instance.

## Connect Cursor (recommended: OAuth)

When MCP is enabled on your instance:

1. In Dockyrd, open **Settings → Integrations → AI assistants** and choose **Connect Cursor**.
2. Complete sign-in in the browser (including 2FA if prompted).
3. Approve the requested scopes (`read`, `deploy`, and/or `write`).

Cursor stores the access token and talks to `POST https://YOUR_PANEL/mcp` with **Streamable HTTP**.

If you must configure manually (advanced), use an HTTP MCP entry pointing at your panel’s `/mcp` URL and follow Cursor’s OAuth flow when the server returns `401` with resource metadata—do not put long-lived secrets in project files you commit.

## Connect Claude (Claude Code / Claude Desktop)

Claude remote MCP uses the same pattern:

1. In Dockyrd, use **Connect Claude** on the Integrations page, or add a remote MCP server in Claude’s settings using your panel URL.
2. On first request, Claude discovers authorization via **`/.well-known/oauth-protected-resource`** and opens the panel login.
3. Grant scopes matching what you want the assistant to do (read-only vs deploy vs full write).

Use your **production panel URL** (HTTPS), not `localhost`, unless you are developing on the same machine.

## Scopes and roles

OAuth scopes align with Dockyrd RBAC:

| OAuth scope | Roughly equivalent to |
|-------------|----------------------|
| `mcp:read` | Viewer+: list/get resources, logs, analytics |
| `mcp:deploy` | Member+: trigger deploys and restarts |
| `mcp:write` | Member+: create/patch projects and services |
| `mcp:instance:read` | Admin+: instance settings (no secrets in tool output) |

Instance-changing tools (SMTP passwords, wildcard domains) stay **owner/admin** only and may be read-only in v1.

## Local stdio (SSH session on the server)

For a subprocess MCP on the host:

```bash
export DOCKYRD_URL="https://panel.example.com"
export DOCKYRD_TOKEN="dyk_…"   # from Settings → Account → API tokens
dockyrd mcp --stdio
```

Create the token with the right **project** and **scope** (`read`, `deploy`, or `write`). Stdio mode does not use OAuth; protect the token like a password.

Point Cursor or Claude at this command only when the editor runs **on the server** (or via SSH). For laptops connecting to a remote VPS, prefer **HTTPS + OAuth** instead.

## Security practices

- Prefer **OAuth** for Claude and Cursor on your workstation.
- **Revoke** connections from Integrations or delete API tokens you no longer need.
- Turn off **Allow API token for MCP** on the instance unless you need automation-only access.
- Do not expose the panel on plain HTTP in production; OAuth and cookies depend on a trusted HTTPS origin.

## Related

- [CLI and `dy` tokens](./install.md)
- [Users & security](https://dockyrd.io/docs/security.html)
- [Troubleshooting](./troubleshooting.md)
