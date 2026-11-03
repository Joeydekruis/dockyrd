# Sprints (M0 → M5)

Backlog and sprints live on **GitHub** ([workflow](github-workflow.md)). Use **milestones** for sprint boundaries.

| Sprint | GitHub milestone | Goal | Exit |
|--|--|--|--|
| **1** | `M0 · Sprint 1 — Monorepo & CI` | Legal, Go skeleton, Makefile, CI, Next static export + embed | `make build` on CI; binary serves embedded placeholder UI |
| **2** | `M0 · Sprint 2 — SQLite & auth bootstrap` | goose + sqlc, config paths, session/auth stubs | `go test ./internal/db/...`; auth paths stubbed |
| **3** | `M0 · Sprint 3 — Install & setup-token` | `install.sh` skeleton, systemd, `dockyard init`, setup-token | Local `init` + token in logs |
| **4** | `M0 · Sprint 4 — Dashboard shell` | AppShell, sidebar, login/setup, ⌘K skeleton, empty states | Login to empty dashboard (M0 done) |

Later sprints follow `roadmap.md` milestones **M1–M5**.

## M1 · Sprint 1 issues (GitHub) — complete 2026-11-03

Milestone: **M1 · Sprint 1 — Projects API** (closed) — #13–#15

## M1 · Sprint 2 — Deploy & reconciler — complete 2026-11-03

Milestone: **M1 · Sprint 2 — Deploy & reconciler** (closed) — #16

## M1 · Sprint 3 — Logs, env, UI — complete 2026-11-03

Milestone: **M1 · Sprint 3 — Logs, env & Projects UI** (closed) — #17–#18

Post-M1 (briefing M1 exit): Dockerfile builds (#19, M2 Sprint 1), Railpack (#20), app detail screens (§17), load-tested zero-downtime redeploy.

## M2 · Sprint 1 — Dockerfile builds — complete 2026-11-03

Milestone: **M2 · Sprint 1 — Dockerfile builds** (closed) — #19

Local spec: `docs/project/issues/19/`

## M2 · Sprint 2 — Railpack — complete 2026-11-03

Milestone: **M2 · Sprint 2 — Railpack builds** (closed) — #20

## M2 · Sprint 3 — GitHub App — complete 2026-11-03

Milestone: **M2 · Sprint 3 — GitHub App** (closed) — #21

## M2 · Sprint 4 — Static sites — complete 2026-11-03

Milestone: **M2 · Sprint 4 — Static sites** (closed) — #22

**Status (2026-11-03):** All M2 issues closed on GitHub (`status/done`). Gate: `./scripts/verify-m2-smoke.sh`.

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

## Sprint 4 issues (GitHub)

| Issue | Title | Local spec |
|--|--|--|
| #10 | [foundation] AppShell, sidebar, and empty dashboard | `docs/project/issues/10/` |
| #11 | [foundation] Sign-in, setup wizard, and auth API | `docs/project/issues/11/` |
| #12 | [foundation] Command palette skeleton | `docs/project/issues/12/` |

## Pickup order

1. Sprint 1 issues in table order
2. Sprint 2 issues when Sprint 1 milestone is complete

```bash
gh issue list --milestone "M0 · Sprint 1 — Monorepo & CI" --state open
```
