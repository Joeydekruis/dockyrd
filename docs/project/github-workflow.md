# GitHub workflow (open source)

Dockyard is managed **entirely on GitHub**: issues, milestones, discussions, and pull requests. There is **no Linear** (or other external PM tool).

## Backlog

| Mechanism | Use |
|--|--|
| **[Issues](https://github.com/Joeydekruis/dockyrd/issues)** | Features, bugs, tasks — single source of truth |
| **Milestones** | Sprints and roadmap slices (e.g. `M0 · Sprint 1 — Monorepo & CI`) |
| **Labels** | `type/feature`, `type/bug`, `area/go`, `area/web`, `status/ready-for-review`, … (create in repo Settings → Labels) |
| **Projects** (optional) | Board view — not required for agents |

## Branches

- **`development`** — day-to-day implementation (agents and contributors).
- **`main`** — release-ready; merges from `development` when maintainers ship (see `.cursor/skills/push-to-main/`).

Do not use long-lived `feature/*` branches for routine work.

## Feature work (agents)

1. Open or pick a GitHub issue; note the number **`#n`**.
2. Planner writes `docs/project/issues/<n>/spec.md` + `rubric.md`.
3. Implement on `development`; run gates from [`setup.md`](setup.md).
4. Commit + push with `Refs #n` or `Fixes #n`.
5. Post a **Ready for review** comment on the issue (template: [`templates/github-handoff-comment.md`](templates/github-handoff-comment.md)).
6. Add label **`status/ready-for-review`**. A maintainer closes the issue after verification.

## Direct fixes (chat / small polish)

Implement on `development` without opening an issue unless you want public tracking. Commits when Joey asks or when batching a release.

## Pull requests

Optional for contributors:

- Target **`development`** for ongoing work.
- Link `Fixes #n` in the PR description.
- Maintainers may land agent work via direct push to `development` instead of PRs.

## CLI

```bash
gh issue list --milestone "M0 · Sprint 1 — Monorepo & CI" --state open
gh issue create --title "[area] Title" --milestone "..." --body-file -
gh issue comment 12 --body-file handoff.md
```

## Agent rules

Cursor enforces this in:

- [`.cursor/rules/issue-tracking.mdc`](../../.cursor/rules/issue-tracking.mdc)
- [`.cursor/rules/github-handoff.mdc`](../../.cursor/rules/github-handoff.mdc)
