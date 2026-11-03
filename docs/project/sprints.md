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

## M3 · Sprint 1 — Postgres & Valkey — complete (local)

Milestone: **M3 · Sprint 1 — Postgres & Valkey** — #23

## M3 · Sprint 2 — Reference variables — complete (local)

Milestone: **M3 · Sprint 2 — Reference variables** — #24

## M3 · Sprint 3 — Garage install — complete (local)

Milestone: **M3 · Sprint 3 — Garage install** — #25

## M3 · Sprint 4 — Buckets & UI — complete (local)

Milestone: **M3 · Sprint 4 — Buckets & UI** — #26

**Status (2026-11-03):** **Closed on GitHub** — issues **#23–#26**; gate `./scripts/verify-m3-smoke.sh` @ `development` v0.0.19+.

## M4 · Sprint 1 — Events & host metrics — closed

Milestone: **M4 · Sprint 1 — Events & host metrics** — #27, #28

**Status (2026-11-03):** **Closed on GitHub** — #27, #28.

## M4 · Sprint 2 — Team & audit — closed

Milestone: **M4 · Sprint 2 — Team & audit** — #29, #30

**Status:** **Closed on GitHub** — #29, #30, #31, #32, #33, #34 (SMTP reset/invites v0.0.19).

**M4 milestone:** **done** — maintainer closed 2026-11-03. Gate: `./scripts/verify-m4-smoke.sh`.

## M5 · Sprint 1 — Doctor & admin CLI — complete (ready for review)

Milestone: **M5 · Sprint 1 — Doctor & admin CLI** — #35–#38

| Issue | Title |
|--|--|
| #35 | `dockyard doctor` |
| #36 | Admin reset-password / disable-2fa |
| #37 | CI workflow (`ci.yml`) |
| #38 | Self-update + rollback CLI |

**Status (2026-11-03):** `development` @ **v0.0.15** — gate `./scripts/verify-m5-smoke.sh`. User docs under `docs/user/`.

## M5 · Sprint 2 — E2E & release — complete (ready for review)

Milestone: **M5 · Sprint 2 — E2E & release** — #39, #40

| Issue | Title |
|--|--|
| #39 | E2E smoke + `e2e.yml` |
| #40 | GoReleaser + `release.yml` + install download |

**Status (2026-11-03):** `development` @ **v0.0.16–0.0.17**. #41 Playwright + #42 server info API ready for review.

**Superseded** — see “M5 · Sprint 2 — complete” above.

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
