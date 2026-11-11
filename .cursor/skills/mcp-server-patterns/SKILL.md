---
name: mcp-server-patterns
description: MCP server design - tools, schemas, auth, stdio vs HTTP. Use with mcp-reviewer subagent and Context7 for SDK docs.
---

# MCP Server Patterns

Use when building MCP servers in product repos.

## When to use

- New MCP server or new tools
- Schema/auth design
- stdio vs Streamable HTTP transport choice
- SDK upgrade or registration bugs

## Core

- **Tools** - actions with Zod/schema validation
- **Resources** - read-only data
- **Prompts** - reusable templates (optional)
- **Transport** - stdio local; Streamable HTTP remote (Cursor)

## Best practices

- Schema-first tool inputs
- Structured errors (no stack traces to clients)
- Idempotent tools where possible
- Pin SDK version; use **documentation-lookup** skill for current API

## Review

**mcp-reviewer** subagent owns harness feedback at `docs/project/feedback/<n>/mcp-review-i<i>.md`.

## Dockyrd product MCP

When building **Dockyrd’s** MCP (not dev harness MCPs in `.cursor/mcp.json`):

- **ADR-0002** — Streamable HTTP `/mcp`, OAuth 2.1 resource server (Claude/Cursor), stdio + env for local `dockyrd mcp`.
- **Pattern** — `docs/project/patterns/mcp-auth.md`
- **Users** — `docs/user/mcp.md`
