# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

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
