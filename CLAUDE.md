@AGENTS.md

# Claude Code notes

`AGENTS.md` (imported above) is the shared contract for every coding agent. This file adds what Claude Code needs on top. The Cursor setup in `.cursor/` is the **source of truth**; `.claude/` is generated from it (see "Keeping Cursor and Claude in sync").

## Hard rules (always)

- **Branch:** only `development`. First check `git rev-parse --abbrev-ref HEAD`; if different, `git checkout development` before editing. Never create `feature/*` / `fix/*`. A hook blocks it. No commit/push to `main` except via the `push-to-main` skill when Joey asks.
- **Commits:** only when Joey asks, except finishing a Feature-track GitHub issue (commit + push `development`, `Fixes #n`/`Refs #n`, then the evidence comment). Never close Feature issues.
- **Naming:** CLI/binary/installer/domains = **Dockyard**; dashboard UI/logo/in-app copy = **Dockyrd**. Never show "dockyrd" lowercase in UI.
- **Scope:** Go control plane + embedded Next.js UI only. No WordPress/Studio/Atelier, no Linear (GitHub Issues only), no Postgres as app DB (SQLite + goose + sqlc).
- **Web:** Base UI only, never `@radix-ui/*` (`./scripts/check-no-radix.sh`); no external font CDNs; UI primitives come from `web/components/dockyard/ui/` (shared with Storybook).
- **Secrets:** none in code, logs, or chat. Validate input at trust boundaries; parameterized SQL only.

## How to work

- **Pick the track yourself; never ask about process** (size, reviewers, design approval, "shall I continue?").
  - **Direct** (feedback, polish, bugfix): research → implement → verify on `development`. No planner, no GitHub issue, no reviewers. Add a regression test when behavior changes.
  - **Feature** (new feature / big change): GitHub issue + harness. Load the `subagent-driven-development` skill for the loop, `issue-tracking` for backlog rules.
- Ask only when blocked on a product answer you can't reasonably assume. Batch questions in one message.
- Bias to lean: smallest correct change, surgical diffs, no speculative abstractions, no drive-by cleanups. Verify only the gates relevant to what changed (`docs/project/setup.md`).
- Reviewers: skip by default. Direct/micro-S/docs-only → none. Otherwise one specialist, not a stack. Details: `.cursor/rules/reviewer-defaults.mdc`.
- Stop early (Feature): security CRITICAL or frontend wrong-intent FAIL in round 1 → ask Joey. Max 3 generator rounds.
- Done = `CHANGELOG.md` entry (patch +0.0.1, dated, synced with root `package.json`) → commit + push → issue evidence comment with pasted test output and full SHA. Never hand off before push.

## Cursor → Claude mapping

| Cursor wording | In Claude Code |
|--|--|
| `Task(planner)` / subagents in `.cursor/agents/` | Agent tool with `subagent_type: planner` etc. (`.claude/agents/`) |
| `.cursor/rules/*.mdc` (always-on) | Distilled above; read the `.mdc` on demand (table below) |
| `.cursor/rules/*.mdc` with `globs` | `.claude/rules/*.md` with `paths`, auto-loaded when matching files are touched |
| `.cursor/skills/` | `.claude/skills/` (same content) |
| `.cursor/hooks/` | `.claude/hooks/` |
| `.cursor/mcp.json` | `.mcp.json` (set `CONTEXT7_API_KEY`, `GITHUB_TOKEN` in your shell) |

## Read on demand (not auto-loaded)

| When you are… | Read |
|--|--|
| Running the Feature loop | `.cursor/rules/orchestration.mdc`, `lean-loop.mdc`, `discuss-gate.mdc`, `design-when.mdc` |
| Handing off a GitHub issue | `.cursor/rules/github-handoff.mdc` (comment template) |
| Closing a cycle | `.cursor/rules/changelog.mdc` |
| Writing/running tests | `.cursor/rules/testing.mdc` |
| Touching auth, secrets, API trust boundaries | `.cursor/rules/security.mdc` |
| Writing feedback/spec files | `.cursor/rules/harness-discipline.mdc`, `harness-artifacts.mdc` |
| Deciding on reviewers | `.cursor/rules/reviewer-defaults.mdc` |

Scoped rules for Go (`go-patterns`, `dockyard-stack`), web (`web/*`, `dockyard-design-contract`, `storybook-ui-source`) and TS (`typescript/*`) load automatically from `.claude/rules/`.

## Verify (see `docs/project/setup.md` for the full list)

```bash
make test                      # Go unit tests
make lint                      # Go + Biome
make build                     # web export + embed + go build
pnpm --dir web exec tsc --noEmit
./scripts/check-no-radix.sh    # after web dependency changes
```

## Keeping Cursor and Claude in sync

Edit rules, agents, and skills under `.cursor/` only, then run:

```bash
python3 scripts/sync-claude-config.py          # regenerate .claude/{rules,agents,skills}
python3 scripts/sync-claude-config.py --check  # exit 1 on drift
```

Hand-maintained Claude files: `CLAUDE.md`, `.claude/settings.json`, `.claude/hooks/`, `.mcp.json`. Personal overrides go in `CLAUDE.local.md` / `.claude/settings.local.json` (gitignored).
