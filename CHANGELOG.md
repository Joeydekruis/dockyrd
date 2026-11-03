# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

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
