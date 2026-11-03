# Production validation

Use this after installing Dockyard on a real host (Ubuntu 24.04 or Debian 12, amd64 or arm64).

## Automated (dev machine or CI)

```bash
chmod +x scripts/verify-production-local.sh
./scripts/verify-production-local.sh
```

Runs M3–M5 smoke, `verify-install`, `make e2e`, and integration tests when Docker is available.

## Manual checklist (VPS)

1. **Install** — `curl -fsSL https://get.dockyard.dev | sudo sh` (or `DOCKYARD_INSTALL_BIN` from a release tarball).
2. **Setup** — open setup URL, create owner account.
3. **Owner 2FA** — Settings → Profile & security → enable 2FA (required for owners before the rest of the panel works).
4. **`dockyard doctor`** — resolve any `✗` lines.
5. **SMTP** (optional) — Settings → Instance; test email.
6. **Postgres** — Databases → create → reveal connection.
7. **Garage** — Object storage → install → create bucket in a project.
8. **Deploy** — push or image deploy; confirm HTTPS and logs.
9. **Self-update** — `sudo dockyard update --check` then `sudo dockyard update` on a staging VM; test `--rollback` if needed.
10. **Footprint** — on Linux, `scripts/measure-rss.sh` while idle (targets in briefing §19).

## Load / chaos (advanced)

- **`e2e/scripts/load-health.sh`** — probe `/api/v1/health` during redeploy (requires running panel + load tool).
- Interrupt a build (`kill -9` on `dockyard`); confirm live apps stay up and deployment ends `failed`.

See [`../project/post-m5-backlog.md`](../project/post-m5-backlog.md) for remaining briefing §19 items tracked in GitHub.
