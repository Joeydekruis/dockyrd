# E2E fixtures

Sample apps for manual and future automated deploy smoke (briefing §19).

| Fixture | Purpose |
|---------|---------|
| [`hello-static/`](hello-static/) | Minimal static files for static-site deploy tests |
| _(planned)_ `next-postgres/` | Next.js + `DATABASE_URL` reference var |
| _(planned)_ `s3-upload/` | App using bucket published vars |

## Manual smoke (Docker host)

1. `make build && ./bin/dockyard serve`
2. Create project → deploy `hello-static` as static site or copy into `SitesDir` per static-site flow
3. Confirm HTTP 200 on assigned domain/path

Automated wiring lands in a follow-up issue (see `docs/project/post-m5-backlog.md`).
