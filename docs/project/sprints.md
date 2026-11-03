# Sprints (M0 → M5)

Backlog and sprints live on **GitHub** ([workflow](github-workflow.md)). Use **milestones** for sprint boundaries.

| Sprint | GitHub milestone | Goal | Exit |
|--|--|--|--|
| **1** | `M0 · Sprint 1 — Monorepo & CI` | Legal, Go skeleton, Makefile, CI, Next static export + embed | `make build` on CI; binary serves embedded placeholder UI |
| **2** | `M0 · Sprint 2 — SQLite & auth bootstrap` | goose + sqlc, config paths, session/auth stubs | `go test ./internal/db/...`; auth paths stubbed |
| **3** | `M0 · Sprint 3 — Install & setup-token` | `install.sh` skeleton, systemd, `dockyard init`, setup-token | Local `init` + token in logs |
| **4** | `M0 · Sprint 4 — Dashboard shell` | AppShell, sidebar, login/setup, ⌘K skeleton, empty states | Login to empty dashboard (M0 done) |

Later sprints follow `roadmap.md` milestones **M1–M5**.

## Sprint 1 issues (GitHub)

Create or track these as GitHub issues under milestone **M0 · Sprint 1 — Monorepo & CI**:

| Issue | Title | Local spec |
|--|--|--|
| #1 | [foundation] Legal docs, README, and CI workflow | `docs/project/issues/1/` |
| #2 | [foundation] Go module and CLI skeleton | `docs/project/issues/2/` |
| #3 | [foundation] Next.js static export and embed pipeline | `docs/project/issues/3/` |

Issue numbers may differ after creation — update this table and rename `docs/project/issues/<n>/` to match.

## Sprint 2 issues (GitHub)

| Issue | Title | Local spec |
|--|--|--|
| #4 | [foundation] SQLite goose migrations and sqlc | `docs/project/issues/4/` |
| #5 | [foundation] Config paths and environment | `docs/project/issues/5/` |
| #6 | [foundation] Auth package stubs | `docs/project/issues/6/` |

## Sprint 3 issues (GitHub)

| Issue | Title | Local spec |
|--|--|--|
| #7 | [foundation] install.sh skeleton and paths | `docs/project/issues/7/` |
| #8 | [foundation] systemd unit and service install | `docs/project/issues/8/` |
| #9 | [foundation] dockyard init and setup-token | `docs/project/issues/9/` |

## Pickup order

1. Sprint 1 issues in table order
2. Sprint 2 issues when Sprint 1 milestone is complete

```bash
gh issue list --milestone "M0 · Sprint 1 — Monorepo & CI" --state open
```
