# M5 exit checklist

_Milestone closed 2026-11-03 on `development` @ **v0.0.21** (`3457b7a`). Issues **#35–#44** closed on GitHub._

## Automated gates (run before release tag)

```bash
./scripts/verify-m5-smoke.sh
./scripts/verify-install.sh
make e2e
make lint
make build
```

Optional with Docker: `go test -tags=integration ./...`, `make e2e-ui` (`E2E_FRESH=1`).

## Deliverables (M5 scope)

| Area | Location |
|------|----------|
| Doctor | `dockyard doctor`, `internal/doctor/` |
| Admin CLI | `dockyard admin reset-password`, `disable-2fa` |
| CI | `.github/workflows/ci.yml` |
| Self-update | `dockyard update`, `internal/updater/` |
| E2E | `e2e/smoke.sh`, `.github/workflows/e2e.yml`, `make e2e` |
| Release | `.goreleaser.yaml`, `.github/workflows/release.yml` |
| Playwright | `e2e/tests/`, CI job |
| Server info | `GET /api/v1/server`, instance settings UI |
| Install | `scripts/install.sh`, `scripts/verify-install.sh` |
| User docs | `docs/user/` (install, first-app, nextjs, domains-dns, troubleshooting, email-smtp) |

## After M5 (briefing §19 — not in milestone issues)

- Ubuntu/Debian install matrix in CI
- `e2e/fixtures/` deploy smoke + load during redeploy
- Dashboard update modal (CLI `dockyard update` today)
- Measured idle RSS targets
- Owner 2FA mandate

Ship public binaries: merge `development` → `main`, tag `v0.0.21` (see `.cursor/skills/push-to-main/SKILL.md`).
