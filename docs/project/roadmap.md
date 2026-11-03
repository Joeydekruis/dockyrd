# Roadmap

From briefing §15. Update when a milestone closes.

| Milestone | Summary | Done when |
|--|--|--|
| **M0 · Fundament** | Repo, CI, Go skeleton, SQLite, install/systemd, setup-token, login, AppShell + tokens, ⌘K skeleton | Fresh VPS → install → login to empty dashboard |
| **M1 · Apps image/Dockerfile** | Projects, services, Docker labels, Caddy, domains, blue/green, logs SSE, env vars, reconciler | Image + Dockerfile app on HTTPS domain, zero-downtime redeploy |
| **M2 · GitHub + Railpack + static** | GitHub App, webhooks, Railpack builds, static sites, rollback | Next.js via push; static site rollback |
| **M3 · Databases + storage** | Postgres, Valkey, Garage, reference vars | Next + Postgres + S3 end-to-end |
| **M4 · Observability + team** | Metrics, events, Analytics, team/2FA/audit/search | All 19 design screens functional |
| **M5 · v0.1.0 release** | Self-update, doctor, docs, e2e green | Public release |

**Progress (2026-11-03):** **M3–M5** implementation on `development` @ **0.0.19** (SMTP for reset/invites, instance settings API/UI). Issues **#23–#44** await maintainer verification on GitHub. **§19 v0.1.0** polish (fixtures, update modal, owner 2FA mandate) remains post-close.

**Post-M4 (briefing, not blocking release):** Caddy request analytics, ECharts depth, per-container CPU/RAM, owner 2FA mandate, full §17 service detail tabs.

## Post-MVP (briefing)

v0.2 backups/notifications/previews; v0.3 multi-server agent.
