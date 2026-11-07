# Four Final skills library (dockyrd)

On-demand playbooks. Rules stay short; agents invoke skills for depth.

## Core workflow

| Skill | Use when |
|-|--|
| [dockyrd-dev](dockyrd-dev/SKILL.md) | Go API, SQLite, embedded web, Makefile, Radix check |
| [subagent-driven-development](subagent-driven-development/SKILL.md) | Parent orchestrating discuss → plan → design → generate → review |
| [test-driven-development](test-driven-development/SKILL.md) | Generator implementing behavior |
| [verification-before-completion](verification-before-completion/SKILL.md) | Before merge-ready / handoff claim |
| [receiving-code-review](receiving-code-review/SKILL.md) | Generator processing feedback |
| [systematic-debugging](systematic-debugging/SKILL.md) | Hard bugs, flaky tests |
| [architecture-decision-records](architecture-decision-records/SKILL.md) | ADR drafts and capture |
| [git-workflow](git-workflow/SKILL.md) | Always `development`; GitHub issue commits when handoff applies |
| [push-to-main](push-to-main/SKILL.md) | Full build gate, then merge `development` → `main` (human-requested) |
| [release-ship](release-ship/SKILL.md) | Tag `v*`, wait for GoReleaser, verify linux amd64+arm64 assets |
| [issue-tracking](issue-tracking/SKILL.md) | GitHub Issues = backlog; Direct = chat fixes |
| [documentation-lookup](documentation-lookup/SKILL.md) | Context7 library docs |
| [mcp-server-patterns](mcp-server-patterns/SKILL.md) | Building MCP servers |

## Stack skills

| Skill | Use when |
|-|--|
| [nextjs-turbopack](nextjs-turbopack/SKILL.md) | Next App Router / static export quirks |
| [frontend-patterns](frontend-patterns/SKILL.md) | React/TanStack patterns in `web/` |
| [api-design](api-design/SKILL.md) | Huma/OpenAPI contract design |
| [security-review](security-review/SKILL.md) | Trust boundaries, auth, secrets |

Removed from this repo: Atelier/WordPress section design, Drizzle/postgres app migrations, Python/FastAPI playbooks.
