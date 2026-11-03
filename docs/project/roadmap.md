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

**Progress (2026-11-03):** GitHub issues **#1–#12 (M0)**, **#13–#18 (M1)**, **#19–#22 (M2)** closed as tested/done (`status/done` on `development` @ `f9f7bf6`). Briefing **milestone exit** for M1/M2 (HTTPS live app, push deploy, static rollback on Docker host) remains optional hardening before calling M1/M2 “product complete” in ops.

## Post-MVP (briefing)

v0.2 backups/notifications/previews; v0.3 multi-server agent.
