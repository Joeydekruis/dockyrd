# Roadmap

From briefing §15. Update when a milestone closes.

| Milestone | Summary | Done when | Status |
|--|--|--|--|
| **M0 · Fundament** | Repo, CI, Go skeleton, SQLite, install/systemd, setup-token, login, AppShell + tokens, ⌘K skeleton | Fresh VPS → install → login to empty dashboard | **Done** |
| **M1 · Apps image/Dockerfile** | Projects, services, Docker labels, Caddy, domains, blue/green, logs SSE, env vars, reconciler | Image + Dockerfile app on HTTPS domain, zero-downtime redeploy | **Done** |
| **M2 · GitHub + Railpack + static** | GitHub App, webhooks, Railpack builds, static sites, rollback | Next.js via push; static site rollback | **Done** |
| **M3 · Databases + storage** | Postgres, Valkey, Garage, reference vars | Next + Postgres + S3 end-to-end | **Done** (2026-11-03) |
| **M4 · Observability + team** | Metrics, events, Analytics, team/2FA/audit/search, SMTP | All 19 design screens functional | **Done** (2026-11-03) |
| **M5 · v0.1.0 release** | Self-update, doctor, docs, e2e green | Public release | **Released** — **`v0.0.22`** on `main`; §19 polish ongoing |

**Gate:** `./scripts/verify-m5-smoke.sh`, `./scripts/verify-install.sh`, `make e2e` (Docker host).

**Briefing §19 (post-release backlog):** see [`post-m5-backlog.md`](post-m5-backlog.md) — Ubuntu curl CI matrix, fixture deploy automation, chaos/load in CI.

## Post-MVP (briefing)

v0.2 backups/notifications/previews; v0.3 multi-server agent.
