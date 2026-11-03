# AGENTS.md — Dockyard (dockyrd repo)

Instructions for AI coding agents working in this repository.

## What this repo is

**Dockyard** is a self-hosted deploy platform: one Go binary (systemd), Docker workloads, Caddy edge proxy, SQLite control plane, and an embedded **Next.js static export** dashboard.

| Surface | Name |
|---------|------|
| CLI, binary, installer, domains | **Dockyard** |
| Dashboard UI, logo, in-app copy | **Dockyrd** |

**Product truth:** [`docs/BRIEFING.md`](docs/BRIEFING.md)  
**UI contract:** [`docs/design-index.json`](docs/design-index.json), [`docs/dockyrd-design.html`](docs/dockyrd-design.html)  
**Project index:** [`docs/project/project.md`](docs/project/project.md)  
**Verify commands:** [`docs/project/setup.md`](docs/project/setup.md)  
**Milestones:** [`docs/project/roadmap.md`](docs/project/roadmap.md)

This is **not** a WordPress theme, Studio, or Atelier project. Ignore rules from other Cursor plugins if they conflict (see [`.cursor/rules/dockyard-scope.mdc`](.cursor/rules/dockyard-scope.mdc)).

## Stack (short)

- **Go:** `net/http` + Huma v2, sqlc, goose, modernc SQLite, moby Docker client, embedded UI via `go:embed` of `web/out`
- **Web:** Next.js App Router, `output: 'export'`, shadcn on **Base UI** (no `@radix-ui/*`), Tailwind v4, TanStack Query, openapi-fetch, Biome, pnpm, self-hosted Geist
- **Details:** [`.cursor/rules/dockyard-stack.mdc`](.cursor/rules/dockyard-stack.mdc), skill [`.cursor/skills/dockyard-dev/SKILL.md`](.cursor/skills/dockyard-dev/SKILL.md)

## Git

- **Only branch:** `development` for day-to-day work. Never create `feature/*` or `fix/*`.
- **`main`:** merge only when Joey asks — use skill [`.cursor/skills/push-to-main/SKILL.md`](.cursor/skills/push-to-main/SKILL.md).
- **Commits in Cursor:** only when Joey explicitly asks, except when a finished Linear handoff requires push (see [`docs/project/project.md`](docs/project/project.md)).

## Work tracks

| Track | When | What to do |
|-------|------|------------|
| **Feature** | New feature or big product/architecture change | Linear issue (`DY-nn`) + harness: discuss → plan → [design if needed] → generate → [reviewers if needed] → verify → changelog/handoff |
| **Direct** | Feedback, polish, bugfixes | Research → implement → verify on `development`. No planner, no Linear ceremony, usually no reviewers |

Pick the track yourself; do not ask the human to choose process, size, or design approval.

**Rules:** [`.cursor/rules/issue-tracking.mdc`](.cursor/rules/issue-tracking.mdc), [`.cursor/rules/orchestration.mdc`](.cursor/rules/orchestration.mdc), [`.cursor/rules/lean-loop.mdc`](.cursor/rules/lean-loop.mdc)

## Feature-track loop (parent orchestrates)

1. Discuss product gray areas only when wrong guess would waste work ([`.cursor/rules/discuss-gate.mdc`](.cursor/rules/discuss-gate.mdc)).
2. Decide size (micro-S / S / M / L / XL) and state it.
3. **`planner`** → `docs/project/issues/<id>/spec.md` + `rubric.md`; sync Linear.
4. **`designer`** only if [`.cursor/rules/design-when.mdc`](.cursor/rules/design-when.mdc) applies (briefing §17 gaps, L/XL IA). Most screens already exist in the design export — skip designer.
5. **`generator`** → TDD, update [`docs/project/generator-state.md`](docs/project/generator-state.md). Max **3** generator rounds per issue.
6. **Reviewers** only when [`.cursor/rules/reviewer-defaults.mdc`](.cursor/rules/reviewer-defaults.mdc) or rubric requires — often **skipped** for micro-S/S.
7. Run relevant gates from `setup.md`; capture output for Linear handoff.
8. On Feature close: `CHANGELOG.md` patch bump, commit/push when handoff applies, Linear **In Review** (never **Done**) — [`.cursor/rules/linear-handoff.mdc`](.cursor/rules/linear-handoff.mdc).

Subagents live under [`.cursor/agents/`](.cursor/agents/). Each runs **once** per invocation; the parent owns the loop.

## Agent roster

| Subagent | Role |
|----------|------|
| **planner** | Spec + rubric; Linear sync |
| **designer** | HTML wireframes for §17 / L-XL gaps only |
| **architect** | ADRs in `docs/adr/` — human accepts before binding |
| **generator** | Implementation + tests |
| **code-reviewer** | Light diff; skip by default on small/low-risk work |
| **api-reviewer** | Huma `/api/v1`, webhooks, SSE |
| **database-reviewer** | goose + sqlc + SQLite |
| **security-reviewer** | Auth, secrets, trust boundaries |
| **frontend-evaluator** | UI vs design-index / wireframe + rubric |
| **performance-evaluator** | When rubric requires (footprint, Lighthouse) |
| **mcp-reviewer** | MCP tools if present |
| **doc-updater** | User-facing docs after other reviewers PASS |

Feature feedback files: `docs/project/feedback/<id>/<type>-i<n>.md`. Generator may only write `generator-response-i*.md`.

## UI implementation

- Match [`docs/design-index.json`](docs/design-index.json) (nav, tokens, shared blocks).
- Build from [`web/components/dockyard/`](web/components/dockyard/) per briefing §9.3 when code exists.
- New/changed screens: [`.cursor/rules/dockyard-design-contract.mdc`](.cursor/rules/dockyard-design-contract.mdc), [`.cursor/rules/web/design-quality.mdc`](.cursor/rules/web/design-quality.mdc).
- Static export routing: [`docs/project/patterns/static-export-routing.md`](docs/project/patterns/static-export-routing.md).

## Testing & quality

- [`.cursor/rules/testing.mdc`](.cursor/rules/testing.mdc) — Go unit/integration, web build, `./scripts/check-no-radix.sh`, CI workflows per briefing §14.
- Security: [`.cursor/rules/security.mdc`](.cursor/rules/security.mdc).
- Go style: [`.cursor/rules/go-patterns.mdc`](.cursor/rules/go-patterns.mdc).

## Cursor configuration map

| Path | Purpose |
|------|---------|
| `.cursor/rules/*.mdc` | Always-on and scoped rules |
| `.cursor/agents/*.md` | Subagent prompts |
| `.cursor/skills/` | Deep playbooks (`dockyard-dev`, TDD, push-to-main, …) |
| `.cursor/hooks/` | e.g. block feature branches |

## When to pause and ask Joey

- Product choice that **blocks** correct Feature implementation
- **ADR** acceptance for architectural forks
- Stop-early: security CRITICAL round 1, or frontend evaluator hard FAIL on intent
- Generator round **3** exhausted
- Missing secrets/credentials you cannot obtain

Do **not** pause for: process, size, reviewers, design approval, “shall I continue?”

## Linear

- Issue prefix: **`DY-nn`** (update team URL in handoff when Linear workspace is configured).
- Status after Cursor finish: **`In Review`** with evidence comment — never **`Done`** (PM only).

---

For the full rule set, see [`.cursor/rules/`](.cursor/rules/). For skills index, see [`.cursor/skills/README.md`](.cursor/skills/README.md).
