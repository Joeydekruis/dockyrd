---
github: https://github.com/Joeydekruis/dockyrd/issues/76
title: P1b — Server Security · Overview & add-ons
size: M
parent: https://github.com/Joeydekruis/dockyrd/issues/74
depends_on: https://github.com/Joeydekruis/dockyrd/issues/75
milestone: M7 · Security & performance
---

## Summary

Ship **Server → Security** for MVP security phase **P1b**: baseline **checks** (read + run API, dashboard Overview), **add-on catalog** with SQLite `addons` state and install/uninstall/plan for baseline host add-ons, and **real** `hostops` implementations for `firewall.ensure`, `ssh.passwordauth.disable`, and `unattended.enable` (replace #75 stubs). Wire **lockout confirm** UX for SSH/firewall fixes from Overview and add-on flows.

**Product source:** `docs/project/SECURITY_BRIEFING.md` §3.3–§3.4, §4.1–§4.2, §5 (`addons`), §6 (checks/addons/host-actions), §8 (UI artboards), §9 (P1b), §10 (tests).

**Foundation (#75):** Use existing `internal/hostops` Runner, FSM, reconcile, `POST /api/v1/security/host-actions/{id}/confirm|revert` — **do not reimplement** the framework.

**Assumption:** P0 proxy defaults (encode, headers, timeouts) already land in `internal/proxy/render.go`; `proxy.defaults` check treats missing pieces as **fail** with message pointing at deploy/reconcile, not an auto-fix in this issue.

## Size note (M, four workstreams)

| Stream | Scope |
|--------|--------|
| **A** | `internal/securitychecks` + checks API + optional doctor hooks |
| **B** | `addons` table + catalog + install job orchestration |
| **C** | Real host actions (firewall, SSH, unattended) |
| **D** | Web: `/server/security` Overview + Add-ons, nav, API client |

CrowdSec / proxybuild install remains **catalog-only** (not installable) until #81+ / P2–P3.

## Design artefacts

**None (skip `designer`).** Visual contract:

- Artboards **Security · Overview** and **Security · Add-ons** in `docs/dockyrd-security-design.html` (briefing §8: `ServerSecurity`, `ServerSecurityAddons`).
- New primitives as needed in `web/components/dockyard/ui/` per briefing §8: `StatusRow`, `StepList`, `CodeBlock` (diff), reuse existing cards/rows from design-index patterns.

## Related ADRs

- **Deferred:** panel non-root + helper daemon (briefing §11.1) — UI copy says “Host actions”, no privilege-separation promise.

## Navigation & routes

| Route | Board |
|-------|--------|
| `/server/security` | Overview — checklist, “Run checks again”, fix flows |
| `/server/security/addons` | Add-ons — catalog, install preview, progress |

**Alternate OK:** single `/server/security` with tabs (Overview | Add-ons) if static-export routing is simpler — must match artboard IA (two boards).

- Add **Security** to `serverNav` in `web/lib/nav.ts` **between Analytics and Terminal** (briefing §8).
- Static export: register paths in the app router (`generateStaticParams` / nested `(shell)` pages) consistent with `docs/project/patterns/static-export-routing.md`.
- Register Go SPA route table entries so deep links work when embedded.

## Package layout (target)

| Path | Role |
|------|------|
| `internal/securitychecks/` | Check definitions, runner, JSON DTOs, fix hints + linked `host_action` kind |
| `internal/securitychecks/store.go` | Optional: persist last run + `ran_at` for GET without re-run (single-row or small table) |
| `internal/security/addons/` | Catalog defs, state machine, Plan/Install/Uninstall/Health for P1b add-ons |
| `internal/hostops/actions/firewall.go` | `firewall.ensure` |
| `internal/hostops/actions/ssh.go` | `ssh.passwordauth.disable` |
| `internal/hostops/actions/unattended.go` | `unattended.enable` |
| `internal/db/migrations/00013_addons.sql` | `addons` per briefing §5 |
| `internal/db/queries/addons.sql` | sqlc |
| `internal/server/api_security_checks.go` | GET + POST run |
| `internal/server/api_security_addons.go` | catalog, plan, install, uninstall |
| `internal/jobs/` | New job kind(s) e.g. `security.addon.install` executing addon install with rollback |
| `web/components/dockyard/security-*` | Overview + Add-ons pages |
| `web/app/(shell)/server/security/...` | Page entry |
| `internal/doctor/doctor.go` | **Optional stretch:** append baseline checks reusing `securitychecks` (briefing §4.10) |

Remove or stop registering **stub** factories in `internal/hostops/registry.go` for the three implemented kinds; keep stubs for `crowdsec.bouncer.install`, `proxy.swap`.

## Baseline checks (§4.1)

Each check returns: `id`, `title`, `status` (`ok` | `warn` | `fail`), `severity`, `detail`, optional `fix` (`none` | `host_action` | `link`), `fix_kind`, `fix_href`, `last_run_at`.

| ID | Detection (summary) | Fix in P1b |
|----|-------------------|------------|
| `firewall.active` | nft `dockyrd` table / expected rules present | `firewall.ensure` via UI |
| `docker.published_ports` | Docker: host-bound ports except proxy 80/443 | **No auto-fix**; list services + deep links |
| `ssh.password_auth` | `sshd -T`: password auth / root login | `ssh.passwordauth.disable` + lockout |
| `updates.unattended` | `unattended-upgrades` + apt config on Debian/Ubuntu | `unattended.enable`; other distros → warn + manual hint |
| `reboot.pending` | `/var/run/reboot-required` | Notice only |
| `panel.2fa` | Instance `require_2fa_all` (or equivalent) | Link `/settings/server/team` or account security |
| `panel.https` | Public URL scheme https | Link `/settings/server/general` |
| `proxy.defaults` | Rendered Caddy JSON includes encode + briefing baseline headers | Info if P0 satisfied; fail with reconcile hint if not |

**Scheduling:** On `GET /api/v1/security/checks`, return cached results if `ran_at` within **24h** unless `?refresh=1`. `POST …/checks/run` always re-runs synchronously (acceptable for M). Optional `runtime.go` ticker every 24h: re-run + create **Needs attention** inbox notification when any check newly fails (reuse `internal/inbox` patterns).

Mirror doctor `Check` shape where practical (`Name`/`Status`/`Detail`/`FixHint`).

## Host actions (real implementations)

Implement briefing §3.3 behaviors on top of #75 `Action` contract. **Unit tests** use temp dirs / fakes; integration tests may skip without root.

### `firewall.ensure`

- `Plan`: ruleset summary (22, 80, 443 in; default drop), mention `DOCKER-USER` / docker-published-port caveat (§2 #8, §11.6).
- `Apply`: idempotent nftables rules in a `dockyrd` table (or documented chain); persist revert payload.
- `Verify`: list rules / default policy; failure → runner auto-revert.
- `RequiresLockoutConfirm`: **true**.

### `ssh.passwordauth.disable`

- Drop-in `/etc/ssh/sshd_config.d/00-dockyrd.conf` (sorts before `50-cloud-init.conf`).
- Pre-checks before `Plan` succeeds: authorized_keys for root or sudo user; current session not password-only (briefing §3.3). Return **422** with clear problem detail when blocked.
- `Apply`: write drop-in; `sshd -t`; reload sshd.
- `Verify`: `sshd -T` shows `passwordauthentication no`; `permitrootlogin` appropriate; else revert.
- `RequiresLockoutConfirm`: **true**.

### `unattended.enable`

- Debian/Ubuntu only; else `Plan` may warn and install path returns distro error.
- `Apply`: install/enable `unattended-upgrades`, security-only config, **no auto-reboot**.
- `Verify`: package active + config present.
- `RequiresLockoutConfirm`: **false** (confirm straight to `confirmed` after verify).

## Add-on framework (P1b subset)

SQLite **`addons`** (briefing §5): `id`, `state`, `version`, `config_json`, `installed_at`, `last_error` with CHECK on `state` ∈ `not_installed|installing|installed|failed|uninstalling`.

**In-code catalog** (Go structs, not user manifests) registers:

| Add-on ID | Kind | P1b install |
|-----------|------|-------------|
| `firewall` | baseline | **Yes** → `firewall.ensure` |
| `unattended-updates` | baseline | **Yes** → `unattended.enable` |
| `crowdsec` | recommended | Catalog row; `install` → **501/422** “available in a later release” |
| `proxy-waf` | optional | Catalog only |
| `proxy-rate-limit` | optional | Catalog only |

States: transitions per briefing §3.4; failed install rolls back host state where applicable and stores `last_error`. Audit log on transitions (`addon.installing`, etc.).

**Install flow:**

1. `POST /api/v1/security/addons/{id}/plan` — Admin+; aggregate `Plan()` previews from host actions; **no mutations**.
2. `POST …/install` — Admin+; set `installing`, enqueue job (prefer existing `jobs` worker); job runs host action chain via `hostops.Runner` (`Plan` row → `Apply` → `Verify` → lockout path); on success `installed` + version string; on failure `failed` + revert.
3. `POST …/uninstall` — Admin+; reverse where implemented (firewall: revert rules; unattended: disable/remove drop-in per safe revert); CrowdSec/WAF: N/A in P1b.

**Health:** `Health()` for installed baseline add-ons used by optional doctor line and UI status dot.

## API (`/api/v1/security`, Admin+ on writes)

| Method | Route | Behavior |
|--------|-------|----------|
| GET | `/api/v1/security/checks` | Latest results (+ cache metadata) |
| POST | `/api/v1/security/checks/run` | Re-run all checks; return full report |
| GET | `/api/v1/security/addons` | Catalog merged with DB state |
| POST | `/api/v1/security/addons/{id}/plan` | Install preview JSON |
| POST | `/api/v1/security/addons/{id}/install` | Start async install; return job/addon id |
| POST | `/api/v1/security/addons/{id}/uninstall` | Start uninstall |
| POST | `/api/v1/security/host-actions/{id}/confirm` | **Existing (#75)** — UI “Keep change” |
| POST | `/api/v1/security/host-actions/{id}/revert` | **Existing (#75)** |

**Fix from check (Overview):** `POST /api/v1/security/checks/{id}/fix` — creates `hostops` planned row for the mapped kind, then `Apply`+`Verify` (same as check’s fix button). Returns `host_action_id` + state for lockout UI. SSH pre-check failures → 422.

Regenerate OpenAPI + TS client via `make generate`.

## UI (Overview)

Per **Security · Overview** artboard:

- Checklist using `StatusRow` (pass/warn/fail).
- Primary **Run checks again** → POST run.
- Rows with host fix: **Preview** (plan JSON / diff in `CodeBlock`) → **Apply** → lockout banner with countdown → **Keep change** / **Revert now** (confirm/revert APIs).
- `docker.published_ports`: table/list with links to affected services.
- Link-only fixes: navigate to Instance / Team settings.
- Protection add-ons summary card (installed baseline count / links to Add-ons tab).

## UI (Add-ons)

Per **Security · Add-ons** artboard:

- Cards/rows per catalog entry with state badge.
- Baseline: Install → plan modal → confirm → installing spinner (poll addon state or job).
- Non-installable entries: disabled Install + short copy (“Coming in P2/P3”).
- Show **Host actions** explainer (no false helper-daemon promise).

## Acceptance criteria

### Checks (stream A)

- [ ] All eight §4.1 checks implemented with stable `id` values and tests for pure detection logic (mocked exec/docker where needed).
- [ ] `GET` and `POST …/checks/run` Admin+ for POST; GET readable by Admin+ (or Member read-only — **default Admin+ only** unless existing server settings pattern says otherwise).
- [ ] 24h cache behavior on GET; `POST run` updates cache.
- [ ] `POST …/checks/{id}/fix` for the three host-action checks; SSH pre-check returns 422 when unsafe.

### Host actions (stream C)

- [ ] `firewall.ensure`, `ssh.passwordauth.disable`, `unattended.enable` registered as real actions; stubs removed from default registry.
- [ ] SSH: drop-in path `00-dockyrd.conf`, `sshd -t`/`sshd -T` verify, lockout timer on success path.
- [ ] Firewall + SSH require confirm within 120s; unattended skips to confirmed.
- [ ] Unit tests cover Plan JSON shape, verify-failure revert (fake action or injected failure).

### Add-ons (stream B)

- [ ] Migration `00013_addons.sql` + sqlc; seeds/upssert catalog rows on first read or migration.
- [ ] GET addons returns merged catalog + state; install/uninstall/plan routes wired.
- [ ] Successful firewall + unattended install sets `installed`; failure sets `failed` with `last_error` and leaves host reverted.
- [ ] crowdsec / proxy-* install attempts return explicit not-available error without mutating host.

### API integration

- [ ] Routes under `/api/v1/security/*`; audit on addon mutations and check fix invocations.
- [ ] Handler tests for auth + 422/404 paths (pattern from `api_security_host_actions_test.go`).

### UI (stream D)

- [ ] Routes `/server/security` and `/server/security/addons` (or tab equivalent) in shell + `serverNav`.
- [ ] Overview and Add-ons match artboard structure (checklist, fix preview, lockout confirm, addon install steps).
- [ ] Uses openapi-fetch types; TanStack Query for checks/addons polling where needed.

### Verification gates

- [ ] `go test ./internal/securitychecks/... ./internal/hostops/... ./internal/security/...`
- [ ] `go test ./internal/server/... -run Security` (or targeted new tests)
- [ ] `pnpm --dir web exec tsc --noEmit`
- [ ] `make generate && make test && make build`

## Out of scope

- `security_policy`, site Security tab (P1c #77+).
- CrowdSec container, bouncer, decisions UI (P3).
- Proxybuild, `proxy.swap`, WAF/rate-limit install (P2).
- Real `crowdsec.bouncer.install` / `proxy.swap` host actions.
- E2E VM SSH lockout on Debian 12 (briefing §10 — follow-up issue or manual gate in handoff).
- New ADR for non-root panel.
- Marketing site / `site/source.html` unless user docs explicitly requested in rubric doc-updater pass.

## Roadmap

On close: M7 epic #74 — mark **P1b** complete; unblocks **P1c** (policy + site tab).

## Change log

| Date | Note |
|------|------|
| 2026-11-07 | Initial planner spec (M, P1b) |
