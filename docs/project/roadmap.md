# Roadmap

From briefing §15. Update when a milestone closes.

| Milestone | Summary | Done when | Status |
|--|--|--|--|
| **M0 · Fundament** | Repo, CI, Go skeleton, SQLite, install/systemd, setup-token, login, AppShell + tokens, ⌘K skeleton | Fresh VPS → install → login to empty dashboard | **Done** |
| **M1 · Apps image/Dockerfile** | Projects, services, Docker labels, Caddy, domains, blue/green, logs SSE, env vars, reconciler | Image + Dockerfile app on HTTPS domain, zero-downtime redeploy | **Done** |
| **M2 · GitHub + Railpack + static** | GitHub App, webhooks, Railpack builds, static sites, rollback | Next.js via push; static site rollback | **Done** |
| **M3 · Databases + storage** | Postgres, Valkey, Garage, reference vars | Next + Postgres + S3 end-to-end | **Done** (2026-11-03, #23–#26) |
| **M4 · Observability + team** | Metrics, events, Analytics, team/2FA/audit/search, SMTP | All 19 design screens functional | **Done** (2026-11-03, #27–#34) |
| **M5 · v0.1.0 release** | Self-update, doctor, docs, e2e green | Public release | In progress (`development` @ 0.0.20) |

**Gates:** `./scripts/verify-m3-smoke.sh`, `./scripts/verify-m4-smoke.sh` (automated); manual Docker checklist printed by M3 script.

**Post-M4 (briefing, not blocking M5):** Caddy request analytics, ECharts depth, per-container CPU/RAM, owner 2FA mandate, full §17 service detail tabs.

## Post-MVP (briefing)

v0.2 backups/notifications/previews; v0.3 multi-server agent.
