# Post-M5 backlog (briefing §19 + polish)

Tracked after milestone **M5** / release **v0.0.22+**. Create GitHub issues as you pick these up.

| Item | Notes |
|------|--------|
| Ubuntu/Debian install CI matrix | `.github/workflows/install-matrix.yml` (verify-install on runners) |
| `e2e/fixtures/` deploy smoke | Sample app + Postgres + bucket on Docker host |
| Load during redeploy | `e2e/scripts/load-health.sh` skeleton |
| UI update modal | Instance banner + release notes (server `GET /api/v1/server`) |
| Owner 2FA mandate | **Done** — owners blocked until TOTP enabled |
| Require 2FA for all users | Instance setting `require_2fa_all` |
| RSS targets | `scripts/measure-rss.sh`, doctor warn on Linux |
| Caddy analytics / ECharts depth | Analytics page |
| Service detail tabs | Database/bucket detail vs design-index §17 |
