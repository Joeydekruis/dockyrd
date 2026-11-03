# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [0.0.23] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **Production validation** — `docs/user/production-validation.md`, `scripts/verify-production-local.sh`, `scripts/measure-rss.sh`, `e2e/scripts/load-health.sh`, `e2e/fixtures/hello-static`
- **Install matrix CI** — `.github/workflows/install-matrix.yml` (ubuntu amd64 + arm64)
- **Update modal** — dashboard banner + release notes from `GET /api/v1/server`
- **`packaging/dockyard.env.example`** — `DOCKYARD_REQUIRE_OWNER_2FA=1`

### Changed

- **Owner 2FA mandate** — when `DOCKYARD_REQUIRE_OWNER_2FA=1` (recommended production); instance `require_2fa_all` for all users
- **Databases UI** — service detail panel + reference var hints
- **Post-M5 backlog** — `docs/project/post-m5-backlog.md`

## [0.0.22] - 2026-11-03

_Finished: 2026-11-03_

### Fixed

- **CI `go vet`** — run `ensure-ui-stub.sh` before vet/test so `web/embed.go` has `web/out` on clean checkouts

## [0.0.21] - 2026-11-03

_Finished: 2026-11-03_

### Fixed

- **SMTP dial address** — `net.JoinHostPort` for `go vet` / CI IPv6 lint

### Changed

- **M5 milestone closed** — doctor, admin CLI, CI, self-update, e2e, GoReleaser, Playwright, server API, install verify, user docs (#35–#44); gates `verify-m5-smoke.sh`, `verify-install.sh`, `make e2e`
- **Issue specs** — AC marked complete for #35–#44; `roadmap.md` / `sprints.md`
- **Release** — `development` merged to `main`; tag **`v0.0.21`** (M3 SMTP, M4 team, M5 ship since v0.0.18)
- **`docs/project/m5-exit.md`** — pre-release checklist

## [0.0.20] - 2026-11-03

_Finished: 2026-11-03_

### Changed

- **M3 milestone closed** — Postgres/Valkey, reference vars, Garage, buckets/UI (#23–#26); gate `./scripts/verify-m3-smoke.sh`
- **M4 milestone closed** — events, metrics, team/2FA/search, cleanup, password reset + SMTP (#27–#34); gate `./scripts/verify-m4-smoke.sh`
- **Project docs** — `roadmap.md`, `sprints.md`, `setup.md`, issue #34 spec (SMTP AC)

## [0.0.19] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **SMTP & instance settings** — `GET|PATCH /api/v1/settings/instance`, `POST /api/v1/settings/smtp/test`, `GET /api/v1/settings/public`; encrypted SMTP password in SQLite; password-reset and invite emails when configured — closes M4 gap (Refs #29, #34)
- **User docs** — `docs/user/email-smtp.md`

### Changed

- **Forgot password** — briefing-compliant copy when SMTP is off; email delivery when SMTP is on (no reset link in API)
- **Instance settings UI** — panel URL, ACME/wildcard fields, SMTP form and test

## [0.0.18] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#43 / Install preflight** — arch/curl checks, `DOCKYARD_SKIP_*`, `verify-install.sh` — https://github.com/Joeydekruis/dockyrd/issues/43
- **#44 / User docs** — `docs/user/nextjs.md`, `domains-dns.md` — https://github.com/Joeydekruis/dockyrd/issues/44
- **E2E setup flow** — Playwright API setup + login; `uptime_sec` on `GET /api/v1/server`

### Changed

- **`install.sh`** — configurable `DOCKYARD_BIN_DEST`; checksum verify fix

## [0.0.17] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#41 / Playwright** — `e2e/` UI tests, `make e2e-ui`, CI Playwright job — https://github.com/Joeydekruis/dockyrd/issues/41
- **#42 / Server info API** — `GET /api/v1/server`, instance settings version block — https://github.com/Joeydekruis/dockyrd/issues/42
- **`packaging/dockyard-rollback.service`** — oneshot rollback unit stub

### Fixed

- **Static UI routes** — `/sign-in/` and similar paths resolve to `*/index.html` without redirect loops

## [0.0.16] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#39 / E2E smoke** — `e2e/smoke.sh`, `e2e.yml`, `make e2e` — https://github.com/Joeydekruis/dockyrd/issues/39
- **#40 / Release pipeline** — `.goreleaser.yaml`, `release.yml`, cosign checksum signing, `install.sh` GitHub download — https://github.com/Joeydekruis/dockyrd/issues/40

### Changed

- **Makefile** — embed `package.json` version via `-ldflags -X main.version=…`

## [0.0.15] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#35 / `dockyard doctor`** — host and control-plane diagnostics — https://github.com/Joeydekruis/dockyrd/issues/35
- **#36 / Admin CLI** — `admin reset-password`, `admin disable-2fa` — https://github.com/Joeydekruis/dockyrd/issues/36
- **#37 / CI** — GitHub Actions `ci.yml`, `./scripts/verify-m5-smoke.sh` — https://github.com/Joeydekruis/dockyrd/issues/37
- **#38 / Self-update** — `update --check`, apply with checksum + DB backup, `--rollback` on Linux — https://github.com/Joeydekruis/dockyrd/issues/38
- **User docs** — `docs/user/install.md`, `first-app.md`, `troubleshooting.md`

## [0.0.14] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **M4 capstone** — `GET /api/v1/applications`, home dashboard counts, Applications screen — https://github.com/Joeydekruis/dockyrd/issues/27 (Analytics/dashboard AC)
- **Event deep links** — `action_href` on service-related events; Analytics “Needs attention” list — #27
- **Static release cleanup** — reclaimable bytes + `old_releases` cleanup target — #33

### Changed

- **Sign-in MFA** — optional recovery code path — #32

## [0.0.13] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#33 / Analytics Docker** — server containers + cleanup API, Analytics UI — https://github.com/Joeydekruis/dockyrd/issues/33
- **#34 / Password reset** — forgot/reset API and pages (link shown in UI until SMTP) — https://github.com/Joeydekruis/dockyrd/issues/34

## [0.0.12] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#32 / 2FA & profile security** — TOTP, MFA login step, sessions API, profile UI — https://github.com/Joeydekruis/dockyrd/issues/32
- **`github.com/pquerna/otp`** — authenticator codes

## [0.0.11] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#30 / Invite accept** — preview + accept API, `/accept-invite` — https://github.com/Joeydekruis/dockyrd/issues/30
- **#31 / Search** — `GET /api/v1/search`, ⌘K project hits — https://github.com/Joeydekruis/dockyrd/issues/31

## [0.0.10] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#29 / Team & invites** — users/invites/audit API, `/settings/team` — https://github.com/Joeydekruis/dockyrd/issues/29

## [0.0.9] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#27 / Event log** — `events` table queries, list/ack API, deploy live/failed events, SSE topic — https://github.com/Joeydekruis/dockyrd/issues/27
- **#28 / Host metrics & Analytics** — `metric_host` collector, `GET /api/v1/server/metrics`, `/analytics` UI — https://github.com/Joeydekruis/dockyrd/issues/28
- **`scripts/verify-m4-smoke.sh`** — M4 Sprint 1 automated gates

## [0.0.8] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#23 / Postgres & Valkey** — managed DB services, reconcile, connection API — https://github.com/Joeydekruis/dockyrd/issues/23
- **#24 / Reference vars** — `${{slug.VAR}}` resolution at deploy/build — https://github.com/Joeydekruis/dockyrd/issues/24
- **#25 / Garage install** — `dy-garage`, `/api/v1/storage`, reconcile — https://github.com/Joeydekruis/dockyrd/issues/25
- **#26 / Buckets & UI** — bucket services, keys API, `/databases` + `/object-storage` pages — https://github.com/Joeydekruis/dockyrd/issues/26
- **`scripts/verify-m3-smoke.sh`** — M3 automated gates + manual Docker checklist

## [0.0.7] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#19 / Dockerfile git builds** — clone + buildx, `building` status, scrubbed build logs, git/github `AppConfig` — https://github.com/Joeydekruis/dockyrd/issues/19
- **#20 / Railpack** — prepare + BuildKit build path, `dy-buildkit` bootstrap on serve — https://github.com/Joeydekruis/dockyrd/issues/20
- **#21 / GitHub App** — manifest, callback, webhooks, encrypted `git_sources`, installation id — https://github.com/Joeydekruis/dockyrd/issues/21
- **#22 / Static sites** — Node build, releases + Caddy `file_server`, promote rollback, deployments API — https://github.com/Joeydekruis/dockyrd/issues/22
- **M2 dashboard slice** — project detail, static sites page, instance GitHub settings, SPA export fallback — `web/`
- **`scripts/verify-m2-smoke.sh`** — automated gates + manual Docker checklist

### Changed

- Build timeout (20m) and disk guard (10% free); image prune (5 tags/service); SSE `topics` filter on `/api/v1/stream`
- Docs: M2 issue specs, `setup.md`, `roadmap.md`, `sprints.md`, `generator-state.md`

## [0.0.6] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#16 / deploy pipeline** — job queue, image deploy blue/green, reconciler, Caddy sync from DB — https://github.com/Joeydekruis/dockyrd/issues/16
- **#17 / env + logs + SSE** — encrypted variables API, runtime logs, `/api/v1/stream` — https://github.com/Joeydekruis/dockyrd/issues/17
- **#18 / Projects UI** — live project list + create on `/projects` — https://github.com/Joeydekruis/dockyrd/issues/18

### Changed

- Sprint/docs: M1 sprint 2–3 specs, roadmap and generator handoff for #16–#18; checked AC on closed #13–#15 specs

## [0.0.5] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#13 / projects API** — local server bootstrap on setup, `GET|POST|PATCH|DELETE /api/v1/projects`, default `production` environment — https://github.com/Joeydekruis/dockyrd/issues/13
- **#14 / Docker layer** — `internal/docker` moby client, Dockyard labels, `dy-edge` / project network ensure, `ListManaged` — https://github.com/Joeydekruis/dockyrd/issues/14
- **#15 / proxy + domains** — Caddy JSON render/load, sslip.io platform hostnames, services + domains REST — https://github.com/Joeydekruis/dockyrd/issues/15

## [0.0.4] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#10 / AppShell** — sidebar, header, empty dashboard, placeholder routes — https://github.com/Joeydekruis/dockyrd/issues/10
- **#11 / auth + setup UI** — login, setup wizard, session API — https://github.com/Joeydekruis/dockyrd/issues/11
- **#12 / command palette** — ⌘K skeleton overlay — https://github.com/Joeydekruis/dockyrd/issues/12

## [0.0.3] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#7 / install skeleton** — `scripts/install.sh`, `scripts/uninstall.sh`, path layout — https://github.com/Joeydekruis/dockyrd/issues/7
- **#8 / systemd** — `packaging/dockyard.service`, install enables unit — https://github.com/Joeydekruis/dockyrd/issues/8
- **#9 / init + setup-token** — master key, `dockyard init`, `dockyard setup-token`, setup status from DB — https://github.com/Joeydekruis/dockyrd/issues/9

## [0.0.2] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#4 / SQLite** — goose `00001_init` migration, sqlc settings queries, `internal/db.Open` — https://github.com/Joeydekruis/dockyrd/issues/4
- **#5 / config** — `DOCKYARD_DATA_DIR`, `DOCKYARD_LISTEN`, database path helper — https://github.com/Joeydekruis/dockyrd/issues/5
- **#6 / auth stubs** — session types, `RequireAuth` pass-through, `GET /api/v1/setup/status` — https://github.com/Joeydekruis/dockyrd/issues/6

## [0.0.1] - 2026-11-03

_Finished: 2026-11-03_

### Added

- **#1 / legal + CI** — Apache-2.0 baseline, README, GitHub Actions CI, GitHub-only workflow docs — https://github.com/Joeydekruis/dockyrd/issues/1
- **#2 / Go skeleton** — `dockyard serve|version|init`, health API — https://github.com/Joeydekruis/dockyrd/issues/2
- **#3 / web embed** — Next.js static export scaffold embedded in binary — https://github.com/Joeydekruis/dockyrd/issues/3
