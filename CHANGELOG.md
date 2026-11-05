# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [0.0.52] - 2026-11-05

_Finished: 2026-11-05_

### Added

- **#66 / Notifications F4** — optional `heartbeat_url` in notification settings; `watch.RunHeartbeat` GET ping every 5m via `netsafe` client; field on **Settings → Server → Channels** — https://github.com/Joeydekruis/dockyrd/issues/66
- **Deploy restart alerts** — interrupted deployments marked at panel startup emit `deploy.failed` via notify outbox (`MarkInterruptedDeploymentsReturning` + reconcile hook)

### Changed

- **#65 / Notifications F3** — disk and certificate watchers use configurable warn/critical thresholds and reminder intervals from settings; threshold controls on Channels page — https://github.com/Joeydekruis/dockyrd/issues/65
- **#64 / Notifications F2** — notification settings PATCH uses pointer fields so partial updates do not clear SMTP or watch config — https://github.com/Joeydekruis/dockyrd/issues/64
- **Notify** — Slack/webhook delivery tests (404 disables channel, 429 honors Retry-After)

## [0.0.51] - 2026-11-05

_Finished: 2026-11-05_

### Added

- **#62 / Notifications F0** — SMTP hardening (timeouts, header sanitization, tests); `internal/netsafe` SSRF-safe HTTP client; reconcile marks interrupted deploys **once** at startup — https://github.com/Joeydekruis/dockyrd/issues/62
- **#63 / Notifications F1** — migration `00007_notifications.sql`, `internal/notify` outbox + dispatcher; deploy/backup/disk/cert alerts; admin API + **Settings → Server → Channels** UI; inbox notification prefs; `docs/user/notifications.md` — https://github.com/Joeydekruis/dockyrd/issues/63
- **Backup hardening** — migration `00008_backups_hardening.sql`, restore jobs, SHA256 verify, Garage target, overdue watcher (`internal/watch/backups.go`)

### Changed

- **Backups API/UI** — schedule fields, target test status, restore flow hooks
- **CI / e2e** — webserver and Playwright config updates for notification surfaces

## [0.0.50] - 2026-11-05

_Finished: 2026-11-05_

### Added

- **#57 follow-up** — `buildLocation` `local` / `auto` / `remote:<id>` wired into app + static buildx; `waited_ms` and `frozen_ms` persisted on deployments

### Changed

- **Apps → Deployments** — build resource strip when metrics exist

## [0.0.49] - 2026-11-05

_Finished: 2026-11-05_

### Added

- **#58–#61 / Resource links (M6)** — Directed `service_links`, reference catalog filtered by links, backfill on startup, Connections UI on apps, Connected applications on databases, project List|Map, link counts, ⌘K connection hint — https://github.com/Joeydekruis/dockyrd/issues/58 https://github.com/Joeydekruis/dockyrd/issues/59 https://github.com/Joeydekruis/dockyrd/issues/60 https://github.com/Joeydekruis/dockyrd/issues/61
- **Per-link Docker networks** (`dy-l-*`) reconciled every 30s; Garage attached per bucket link; `links_mode=strict` skips project-wide `dy-p-*` on new deploys
- **API** — `GET/POST /services/{id}/links`, `GET …/linkable`, `DELETE /links/{id}`, `GET /projects/{id}/links`, `POST …/links/backfill`, `PATCH … links_mode`
- **Install** — merge `default-address-pools`, log rotation, and `live-restore` into `/etc/docker/daemon.json` when missing
- **Doctor** — warn when Docker address pools are not configured

### Changed

- **Apps** — join `dy-edge` only when the service has at least one domain
- **Env vars** — `${{slug.VAR}}` rejected unless the slug is linked to the consumer; link-created vars tagged in API/UI

## [0.0.48] - 2026-11-05

_Finished: 2026-11-05_

### Added

- **#56** — Deployment build metrics on `GET …/deployments` (peak memory, budget, wait/freeze, failure kind); dashboard **Waiting** badge; deployment resource strip; Analytics **Build capacity** card; service **Build location** settings; **Server settings → Builders** (remote BuildKit CRUD/test) — https://github.com/Joeydekruis/dockyrd/issues/56
- **#57** — Remote BuildKit API (`internal/build/remote.go`, builders table), `dockyard doctor` build-resilience checks, peak sampler + metrics on successful builds, e2e fixtures `memhog` / `cpuhog` / `spiky-app` — https://github.com/Joeydekruis/dockyrd/issues/57

### Changed

- **Failed deploys** — OOM-style errors map to `failure_kind=oom` with user-facing hints (`internal/build.FailureMessage`)

## [0.0.47] - 2026-11-05

_Finished: 2026-11-05_

### Added

- **#52** — Static site builds via BuildKit (`Dockerfile.dockyard` + `buildx --output type=local`) under `dy-buildkit` limits
- **#53** — PSI build watchdog (freeze/thaw build slice) during app and static builds
- **#54** — `MergeBuildArgs` / `InjectBuildEnv` wired into Railpack and Dockerfile deploy builds
- **#55** — GitHub Actions workflow generator, deploy hook token + `POST /api/v1/hooks/deploy/{service_id}`, push webhook skips server build when `buildLocation=github_actions` — https://github.com/Joeydekruis/dockyrd/issues/55

### Changed

- **GET `/api/v1/server/build-budget`** — RAM/swap budget snapshot for admins (#55 API slice)

## [0.0.46] - 2026-11-05

_Finished: 2026-11-05_

### Added

- **#49–#51 / build resilience (foundation)** — product plan in `docs/project/build-resilience.md`; GitHub epic #49–#57; `internal/resources` (cgroup v2, slice limits via systemd D-Bus, freeze); build admission + `waiting` deployment status; migration `00005_build_resilience`; installer slice units, swap/zswap; `dy-buildkit` `buildkitd.toml` + data volume; Docker `CgroupParent`/OOM scores on managed containers — https://github.com/Joeydekruis/dockyrd/issues/50 https://github.com/Joeydekruis/dockyrd/issues/51

### Changed

- **`packaging/dockyard.service`** — `dy-system.slice`, `GOMEMLIMIT`, OOM score — #50

## [0.0.45] - 2026-11-04

_Finished: 2026-11-04_

### Added

- **Dockyrd logo dobber** — inline SVG wordmark with hover wave + cube tilt on sidebar and auth surfaces (`web/components/dockyard/dockyrd-logo.tsx`, `globals.css`).
- **Branded page loader** — five-cube staggered bob animation for dashboard shell and page fetch states (`DockyrdPageLoading`, `dockyrd-cube-icon.tsx`).

### Changed

- **Loading UX** — list, detail, and settings pages use the Dockyrd loader instead of plain “Loading…” text.

## [0.0.44] - 2026-11-04

_Finished: 2026-11-04_

### Added

- **#48 Database platform** — scheduled and manual Postgres/Redis backups (local + S3), remote retention prune, restore succeeded Postgres runs from local or S3 (`internal/backup/`, `POST …/backup-runs/{runId}/restore`).
- **Query Studio** — read-only SQL proxy with audit log and schema sidebar (`POST …/query`, `GET …/schema`, `web/components/dockyard/database-query-panel.tsx`).
- **Database connection UX** — internal/external connection cards and Postgres public access toggle (`PATCH …/public-access`, `web/components/dockyard/database-connection-cards.tsx`).
- **Database lifecycle** — delete/restart managed databases from detail UI (`DELETE`/`POST …/restart`, settings danger zone).

### Changed

- **Connection API** — `GET …/connection` returns `internal` / `external` blocks for Postgres (https://github.com/Joeydekruis/dockyrd/issues/48).

## [0.0.43] - 2026-11-04

_Finished: 2026-11-04_

### Added

- **Static site access logs** — `GET /api/v1/services/{id}/logs?stream=access` tails Caddy JSON access logs filtered by service domains; Logs tab on static sites shows HTTP access lines with refresh (`internal/accesslog`, `web/components/dockyard/service-traffic-panel.tsx`).
- **Static site traffic analytics** — `GET /api/v1/services/{id}/analytics?range=` aggregates requests, bandwidth, status codes, and latency from access logs; Analytics tab charts on static site detail (`internal/server/api_logs.go`).
- **Domains UX (Sevalla-style)** — platform/custom domain tables, Verification status, Show DNS records modal with copyable A record and server `public_ipv4`, add-domain modal, ⋮ menu (make primary, delete), `PATCH /api/v1/domains/{id}` for primary (`web/components/dockyard/service-domains-tab.tsx`, `internal/server/api_domains.go`).
- **Server backup target settings** — S3-compatible backup target API and Instance → Backups UI (foundation for database backups).

### Changed

- **Static site edge routing** — prefer direct `release_path` in Caddy when set; site directory modes `0o755` for proxy readability (`internal/staticsite/release.go`, `internal/docker/proxy.go`).
- **Service lifecycle API** — start/stop/restart handlers for supported service kinds (`internal/server/api_service_lifecycle.go`).

### Fixed

- **Domain delete** — repromote primary platform domain when primary custom domain is removed; proxy resync after domain mutations.

## [0.0.42] - 2026-11-04

_Finished: 2026-11-04_

### Fixed

- **Static site 404 (symlink)** — `current` now symlinks with a **relative** `releases/<id>` target so dy-proxy can follow it under `/srv/sites`; proxy sync **repairs** legacy absolute symlinks on existing sites (`internal/staticsite/release.go`, `internal/proxy/sync.go`)

## [0.0.41] - 2026-11-04

_Finished: 2026-11-04_

### Fixed

- **Static site 404 at edge** — Caddy `file_server` root now uses **`/srv/sites/{slug}/current`** (dy-proxy mount) instead of the host `DataDir` path (`internal/proxy/sync.go`, `staticsite.CaddyFileServerRoot`)

## [0.0.40] - 2026-11-04

_Finished: 2026-11-04_

### Changed

- **Static sites overview** — list rows load **service domains** (primary hostname link), show **project / branch** meta, and **Live/Failed/Deploying** from latest deployment; resource row status badge aligned on narrow viewports (`static-sites-page.tsx`, `resource-row.tsx`)

### Fixed

- **Static deploy publish** — clone tree is no longer deleted before `PublishRelease` copies files (fixes `lstat …/src/site: no such file or directory` on publish-without-build) (`internal/staticsite/build.go`, `internal/deploy/static.go`)
- **Static site settings save** — root `/` normalizes to repo root; default publish dir `.` when build is off (`static-site-settings-tab.tsx`)

## [0.0.39] - 2026-11-04

_Finished: 2026-11-04_

### Added

- **Static site build settings (Sevalla-style)** — section headings with icons, read-only meta grid, **Update** actions; build modal with **Build site before publishing**, root/publish dirs, Node LTS, **index file** / **error file** (`web/components/dockyard/static-site-settings-tab.tsx`, `settings-detail-block.tsx`)
- **`buildEnabled` config** — skip Docker build when off; empty install/build commands persist without defaulting to `npm ci` (`internal/staticsite/buildcmd.go`, `ResolveBuildCommand`)
- **Static site index in Caddy** — `indexFile` → `file_server.index_names` via service config on live static routes (`proxy_routes.sql`, `internal/proxy/sync.go`)
- **Resource provisioning helpers** — shared create/reveal flows for databases, buckets, and apps (`web/lib/resource-provisioning.ts`)

### Changed

- **Static site settings display** — normalize corrupt `"false"` / boolean command fields; split `installCmd` / `buildScript` in API config (`web/lib/static-site-display.ts`)
- **Database detail** — connection credentials panel and logs shell aligned with design patterns (`database-detail-page.tsx`, `service-connection-panel.tsx`)
- **Project resources** — provisioning modals and list rows use shared helpers; application/database/bucket create flows updated

### Fixed

- **Static deploy** — no longer runs shell command `false` after bad config; publish-without-build path when build is disabled

## [0.0.38] - 2026-11-04

_Finished: 2026-11-04_

### Added

- **Static site settings** — read-only Settings tab with **Edit** actions and modals (display name, source, build, auto-deploy, SPA fallback); `SettingsDetailBlock` UI primitive (`web/components/dockyard/static-site-settings-tab.tsx`)
- **Static site PATCH API** — `PATCH /api/v1/services/{id}` accepts `name` and merged `config` for `static_site` services (`UpdateServiceName`, `internal/server/api_services.go`)
- **In-app notifications** — inbox table, dispatch on auth events, `GET|PATCH /api/v1/inbox`, header bell, `/notifications` page (`internal/inbox/`, `web/components/dockyard/notifications-*`)
- **Server email settings** — `/settings/server/email` SMTP form and outbound **email log** (`00003_email_log` migration, `api_email_log.go`)

### Changed

- **Instance general settings** — public URL/TLS only on `/settings/server/general`; SMTP moved to dedicated email route
- **Static site builds** — safer build working directory handling (`internal/staticsite/build.go`, `workdir_test.go`)

## [0.0.37] - 2026-11-04

_Finished: 2026-11-04_

### Added

- **Project detail** — design-aligned header, Resources / Project settings tabs, resource table (status, endpoint, CPU, memory, last deploy), hardware usage block, **Add resource** modal (type toggles, searchable pick from other projects, create new)
- **Move resource API** — `PATCH /api/v1/services/{id}` with `project_id` to attach an existing service to a project (`internal/server/api_services.go`, sqlc)
- **Static site detail** — tabbed overview (deployments, environment, domains, settings) and deployment detail route `/static-sites/{slug}/deployments/{id}`
- **Resource detail routes** — `/applications/{slug}`, `/databases/{slug}`, `/object-storage/{slug}` with SPA fallback for `databases/` and `object-storage/`
- **Add env variable modal** — single and bulk paste for service environment variables

### Changed

- **Project detail navigation** — resource names link to type-specific detail pages (`web/lib/service-routes.ts`)
- **Create static site wizard** — optional `defaultProjectId` when opened from project Add resource

## [0.0.36] - 2026-11-04

_Finished: 2026-11-04_

### Added

- **Static export routing** — Go UI handler maps dynamic dashboard URLs (`/projects/{slug}`, `/static-sites/{slug}`, `/applications/…`) and RSC `index.txt` payloads to `_` placeholders via `http.ServeFileFS`; regression tests in `internal/server/server_ui_test.go`
- **Static site detail** — `/static-sites/[slug]` route and summary page; list rows link to service slug
- **Projects overview** — welcome hero, create-project modal, enriched project cards and overview helpers (`web/lib/projects-overview.ts`)
- **Account regional settings** — `/settings/account/regional` and `GET|PATCH /api/v1/auth/locale` (timezone + 12h/24h); charts and analytics respect user locale via `UserLocaleProvider`

### Fixed

- **Dashboard deep links** — hard refresh and direct URLs such as `/projects/dockyrd` no longer return 404 from the embedded static UI
- **Select / SearchSelect** — dropdown menus use fixed positioning so options are not clipped in scroll containers

### Changed

- **Unknown UI paths** — serve exported `404.html` with HTTP 404 instead of falling back to the home page

## [0.0.35] - 2026-11-04

_Finished: 2026-11-04_

### Added

- **Account settings** — separate sidenav and routes: Profile (`/settings/account/profile`), Password & 2FA, Notifications, Security (sessions); no server tabs in account area
- **Server settings** — separate sidenav: General, Team, Git providers, Updates under `/settings/server/*`
- **Account API** — `PATCH /api/v1/auth/account`, `POST /api/v1/auth/password/change`, `GET|PATCH /api/v1/auth/notifications` (Go + `web/lib/api.ts`)
- **SearchSelectField** — searchable select primitive + Storybook (`web/components/dockyard/ui/search-select*.tsx`)
- **Public site** — `site/source.html`, `scripts/build-public-site.mjs`, `docs/user/README.md`; `make site` output

### Changed

- **Navigation** — main sidebar **Server settings** → `/settings/server/general`; account menu **Account settings** / **Server settings**; legacy `/settings/profile` and `/settings/instance` redirect to new paths
- **GitHub integrations** — callback and manifest setup URL → `/settings/server/integrations`
- **Alert & Snackbar** — shared UI primitives and design stories aligned with dashboard tokens
- **Favicon / PWA icons** — updated assets and `site.webmanifest`

### Removed

- Combined **Profile & security** page component (split into account tabs)

## [0.0.34] - 2026-11-04

_Finished: 2026-11-04_

### Fixed

- **GitHub App manifest** — register flow POSTs `manifest` to GitHub (GET `?manifest=` showed an empty form); `BuildManifestLaunch` + `postGitHubManifestLaunch` in the wizard (`internal/github/manifest.go`, `web/lib/open-external-url.ts`)

### Changed

- **Settings → Git providers** — content width ~1000px; **Recommended** badge beside title (no overlap with **Connect GitHub**)
- **Analytics** — container metrics hook and overview tab polish

## [0.0.33] - 2026-11-04

_Finished: 2026-11-04_

### Added

- **#46 / Container metrics** — `RunContainerCollector` (Docker stats → `metric_container`), `GET /api/v1/server/metrics/containers`, enriched `GET /api/v1/server/containers` with project label and live cpu/memory — https://github.com/Joeydekruis/dockyrd/issues/46
- **#47 / Analytics dashboard** — Overview + CPU/Memory/Disk/Network tabs, ECharts (`#3B93F5` theme), meta chips, per-core CPU, top containers, `ServerContainersTable` — https://github.com/Joeydekruis/dockyrd/issues/47
- **Host metrics API** — `30d` range, `net_rx`/`net_tx` on points; live swap + per-core usage on `GET /api/v1/server` (Linux)

### Changed

- **Analytics** — replaces single-page sparkline layout; events/cleanup/running containers on Overview tab
- **`echarts`** dependency in `web/` for dashboard charts

## [0.0.32] - 2026-11-04

_Finished: 2026-11-04_

### Added

- **Self-update UX** — full-screen backdrop + progress overlay on Settings → Updates apply; health polling until panel returns (`update-applying-overlay.tsx`, `wait-for-panel-health.ts`)
- **GitHub release cache** — background check every 5 minutes; manual **Check again** capped at 40/hour without token; `GET /api/v1/updates?refresh=1`; server banner reads cache only (`internal/updater/release_cache.go`)
- **`DOCKYARD_GITHUB_TOKEN`** — authenticated GitHub API for update checks and `install.sh` (`internal/updater/github_http.go`, `packaging/dockyard.env.example`)
- **Resource catalogs** — create flows (applications, databases, buckets, static sites) with wizard/provisioning modals; hero promo cards and list add links
- **Web icons** — favicon, PWA manifest, apple-touch-icon under `web/public/`
- **Host metrics** — normalized CPU fraction helper for Linux analytics (`internal/metrics/cpu_fraction.go`)

### Changed

- **Settings → Updates** — copy explains cached vs manual checks; apply path uses forced refresh + cached apply
- **Resource list rows / template cards / commit chips** — design-aligned catalog layout across Applications, Databases, Object storage, Static sites
- **Modals & snackbars** — wizard shell, elevated snackbar tone for update banner

## [0.0.31] - 2026-11-04

_Finished: 2026-11-04_

### Added

- **Settings → Integrations (GitHub)** — connect wizard (`ConnectGitHubWizard`), disconnect with confirm, enriched status (public URL, account hints); **`DELETE /api/v1/github/connection`**; manifest POST accepts `app_name` / `public_url` (`internal/github/status_enrich.go`, `urls.go`)
- **Search palette** — grouped results, keyboard navigation, and Storybook coverage updates (`web/components/dockyard/ui/search-palette.tsx`)
- **`SnackbarViewport`** — fixed-bottom save/test confirmations on General settings; Storybook **Viewport** story
- **Auth** — server reset hint when SMTP is off (`auth-server-reset-hint.tsx`); **`DockyrdMark`** logo mark; **`openExternalUrl`** helper

### Changed

- **Settings → General** — **Public URL & TLS** and **SMTP** as separate cards with per-section save; inline green status text replaced by snackbar
- **Forgot password & auth layout** — design-aligned split layout and copy
- **Command palette / dashboard shell** — announcement banner placement and palette wiring
- **Global styles** — auth surface tokens and layout utilities in `globals.css`

## [0.0.30] - 2026-11-04

_Finished: 2026-11-04_

### Added

- **Resource catalog pages** — Applications, Databases, Static sites, and Object storage overviews use design **`PageHeader`**, summary **`StatCard`** tiles, filterable **`ResourceListSection`**, and **`ResourceRow`** lists (`web/lib/service-display.ts`)
- **Settings → Integrations** — GitHub App install/callback and manifest setup URL point to `/settings/integrations`; search palette entry; shared **`SettingsPageBody`** wrapper for settings screens
- **User menu** — header account menu component + Storybook (`web/components/dockyard/user-menu.tsx`)
- **Self-update UX** — **`UpdateApplyingOverlay`** and panel health wait helper during apply/restart

### Changed

- **Settings pages** — Profile, Team, Instance, and Updates layouts aligned to settings shell pattern; instance GitHub UI moved to Integrations
- **Announcement banner & top bar** — dismiss/state styling tweaks; dashboard shell padding with user menu slot
- **`ResourceRow`** — Next.js `Link` for internal hrefs; optional `subMeta` and `trailing` slots

## [0.0.29] - 2026-11-04

_Finished: 2026-11-04_

### Added

- **Design component catalog** — `docs/design/component-boards/` HTML exports + `components-catalog.json`; **Storybook** `Design/Components/*` playgrounds (63 primitives) and board stories (`FoundationsAtoms`, `Molecules`, `Organisms`, `FormsOverlays`)
- **UI primitives (`web/components/dockyard/ui/`)** — custom **Select** / **SelectField**, **SplitButton** with deploy menu, navy **Tooltip** (placements + inline icon chip), **TextareaField**, menus (`.dockyard-menu-panel`), and dozens of molecules/organisms aligned to design boards

### Changed

- **Dashboard & auth** — migrate ad-hoc buttons to shared **`Button`**; native `<select>` → **SelectField** on databases, static sites, object storage, team role; **pointer** cursor on interactive controls (`globals.css`)
- **Switch** — animated thumb (`translateX`) and track color transition; respects `prefers-reduced-motion`
- **Storybook** — legacy `web/stories/dockyrd/*` component stories removed in favor of `web/stories/design/`; Introduction MDX wired for SB10

### Fixed

- **Textarea** — padding `12px 14px` per forms design (`.field-textarea`)

## [0.0.28] - 2026-11-04

_Finished: 2026-11-04_

### Added

- **Dashboard UI shell** — TopBar announcement, fixed sidebar/header (232px), settings sidebar variant with icons, scroll-hide top bar, wider white search trigger, command palette dismiss on backdrop click
- **Design system (`web/components/dockyard/ui/`)** — shared Button, status badges, stat tiles, resource rows, modals, snackbar, search palette, auth fields, empty state; **Storybook 10** catalog (`pnpm --dir web storybook`) with single-source rule `.cursor/rules/storybook-ui-source.mdc`
- **Self-update API** — `GET /api/v1/updates`, `POST /api/v1/updates/apply`; Settings → Updates page and update-available banner in dashboard
- **Auth** — `GET /api/v1/auth/me` returns `mfa_required` for pending 2FA sessions; `probeAuthState()` in web client

### Changed

- **Sign-in / auth pages** — design-aligned split layout, placeholders, password-manager-friendly forms (FormData, separate MFA step, `autocomplete` attributes); settings inputs placeholders
- **Sign-in errors** — surface API error text from login failures

### Fixed

- **2FA login UX** — resume MFA step after refresh when session cookie exists but MFA not completed

## [0.0.27] - 2026-11-04

_Finished: 2026-11-04_

### Fixed

- **Caddy JSON** — move cert `storage` to root config (fixes `proxy sync` 400: unknown field `storage` under `apps.tls`); panel HTTPS on 80/443 loads again

## [0.0.26] - 2026-11-04

_Finished: 2026-11-04_

### Fixed

- **`dy-proxy`** — host networking so 80/443 bind on the VPS; recreate outdated bridge containers; pull image before start; panel upstream `127.0.0.1:8000`
- **`dockyard doctor`** — checks `dy-proxy` container and localhost 80/443 listeners

## [0.0.25] - 2026-11-04

_Finished: 2026-11-04_

### Added

- **`dy-proxy` (Caddy)** — system container on 80/443; panel HTTPS when **Public URL** + **ACME email** are set in Instance settings (proxies to panel on `:8000`)
- **Setup banner** — shows server IPv4 in setup URL when detectable; otherwise token + manual URL hint

### Changed

- **Instance settings** — copy clarifies Public URL enables panel HTTPS via Caddy
- **Proxy sync** — loads panel hostname from instance settings; ACME email from settings

## [0.0.24] - 2026-11-03

_Finished: 2026-11-03_

### Changed

- **`scripts/install.sh`** — auto-install Docker Engine on Debian/Ubuntu; fetch `dockyard.service` when run via `curl | sh`; `DOCKYARD_SKIP_DOCKER_INSTALL` / `--no-docker-install`
- **README** and **`docs/user/install.md`** — interim install URL (raw GitHub) until `get.dockyard.dev` is live

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
