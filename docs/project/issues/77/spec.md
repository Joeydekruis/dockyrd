---
github: https://github.com/Joeydekruis/dockyrd/issues/77
title: P1c — Instance policy & application Security tab
size: L
parent: https://github.com/Joeydekruis/dockyrd/issues/74
depends_on:
  - https://github.com/Joeydekruis/dockyrd/issues/76
milestone: M7 · Security & performance
---

## Summary

Ship **M7 P1c**: SQLite **`security_policy`** (singleton) and **`service_security`** (per service), **profile resolution** (Baseline / Strict / Custom, floor + lock), **Caddy render** for resolved headers, CSP, and **Access** (basic auth, IP allowlist, maintenance), plus dashboard **Server → Security → Policy** and **Application / Static site → Security** tab (before Settings).

**Product source:** `docs/project/SECURITY_BRIEFING.md` §3.1–§3.2, §4.5–§4.6, §5 (`security_policy`, `service_security`), §6 (policy + site routes), §7 (render order), §8 (`ServerSecurityPolicy`, `ApplicationSecurity`), §9 (P1c), §10 (golden tests).

**Depends on**

| Prerequisite | Why |
|--------------|-----|
| **P0** proxy baseline | `internal/proxy/baseline.go` encode + default headers/timeouts; `proxy.defaults` check (#76) |
| **#76** checks + add-ons | Policy UI greys rate-limit fields when `proxy-waf` / `proxy-rate-limit` add-ons not `installed`; shared Security nav pattern |

**Assumptions**

- **#76** merged on `development` before generator starts (addons table + `/server/security/*` exists).
- CDN **trusted-proxy render** and Cloudflare range fetch are **#78 (P1d)** — Policy page may **persist** `cdn_*` fields but edge effect waits for P1d.
- Rate limit / WAF **handlers** are **P2** — store instance/site RPM and WAF mode; UI read-only/disabled without proxybuild; no `rate_limit` / `coraza_waf` in rendered JSON in this issue.
- **`security_events`** table + live ingest is **P2c+** — Application Security tab shows **empty state**; optional `GET …/security/events` returns `[]` until events ship.
- Lock **per-site exemption** (§3.1): boolean on `service_security` + explicit Admin action + audit entry (not silent bypass).

## Size note (L, five workstreams)

| Stream | Scope |
|--------|--------|
| **A** | Migrations + sqlc + seed defaults; `internal/securitypolicy` resolve/validate |
| **B** | Policy + service security APIs, audit, authz, OpenAPI |
| **C** | `proxy.Render` per-service chain (headers → encode → access → upstream); `SyncFromDB` loads policy rows |
| **D** | Resource limits (mem/cpu/pids) apply + blue/green redeploy hook |
| **E** | Web: `/server/security/policy`, app + static Security tab, shared components |

## Design artefacts

**None (skip `designer`).** Visual contract:

- **Security · Policy** → `ServerSecurityPolicy` in `docs/dockyrd-security-design.html`
- **Application · Security** → `ApplicationSecurity` in same file
- Reuse #76 primitives: `Switch`, `RadioCard`, `Segmented`, `StatusRow`, `CodeBlock` where artboards show them
- Static sites: same **ApplicationSecurity** layout minus **Resources** block and **WAF mode** control (§4.6)

## Related ADRs

- **Deferred:** panel non-root (§11.1) — unchanged.
- **Deferred:** security-event IP retention ADR (§11.2) — no `security_events` migration in P1c.

## Navigation & routes

| Route | Board |
|-------|--------|
| `/server/security/policy` | Instance policy (tab with Overview / Add-ons) |
| `/applications/{slug}`?tab=**security** | ApplicationSecurity (insert **before** Settings) |
| `/static-sites/{slug}`?tab=**security** | Same, minus Resources / WAF |

- Extend `SecurityNav` (or equivalent) with **Policy** third tab.
- Register shell + Go SPA route table entries (static export pattern).
- Persist tab id via existing `persistDetailTab` / `readInitialDetailTab` helpers.

## Data model

Migration **`00014_security_policy.sql`** (next after `00013_addons`):

### `security_policy` (exactly one row, `id = 1`)

| Column | Type | Notes |
|--------|------|--------|
| `default_profile` | TEXT | `baseline` \| `strict` |
| `lock` | INTEGER | 0/1 — when 1, site cannot be weaker than instance floor |
| `headers_json` | TEXT | JSON map: header name → `{ "enabled": bool, "value": string }` — **no CSP** at instance |
| `rl_default_rpm` | INTEGER | nullable; used when proxybuild installed (P2+) |
| `rl_login_rpm` | INTEGER | nullable |
| `cdn_mode` | TEXT | `direct` \| `cloudflare` \| `custom` |
| `cdn_ranges_json` | TEXT | CIDR list for `custom`; Cloudflare ranges filled in P1d |
| `cdn_header` | TEXT | e.g. `CF-Connecting-IP` for custom |
| `default_mem_mb` | INTEGER | default 512 |
| `default_cpu` | REAL | default 1.0 |
| `default_pids` | INTEGER | default 512 |

Seed on migrate: **Baseline** profile, lock **off**, headers matching current `baselineHeadersHandler()` values, CDN `direct`, defaults per briefing §4.5.

### `service_security` (`service_id` PK, FK → `services.id` ON DELETE CASCADE)

| Column | Type | Notes |
|--------|------|--------|
| `profile` | TEXT | `inherit` \| `baseline` \| `strict` \| `custom` |
| `lock_exempt` | INTEGER | 0/1 — Admin-only; audit when set |
| `csp` | TEXT | nullable |
| `csp_report_only` | INTEGER | 0/1 |
| `basic_user` | TEXT | nullable |
| `basic_hash` | TEXT | bcrypt per Caddy `basic_auth` |
| `ip_allowlist_json` | TEXT | JSON string array of CIDRs (v4 + v6) |
| `maintenance` | INTEGER | 0/1 |
| `rl_mode` | TEXT | `inherit` \| `instance` \| `custom` \| `off` |
| `rl_rpm` / `rl_login_rpm` | INTEGER | nullable |
| `waf_mode` | TEXT | `inherit` \| `off` \| `detect` \| `block` |
| `mem_mb` | INTEGER | nullable → inherit instance default |
| `cpu` | REAL | nullable |
| `pids` | INTEGER | nullable |

**Lazy row:** create `service_security` on first GET/PUT for a service with defaults (`profile = inherit`, null overrides).

## Profile resolution (`internal/securitypolicy`)

Implement §3.1–§3.2:

- **Effective headers** = merge instance `headers_json` with site **Custom** overrides; **Strict** adds Permissions-Policy + longer HSTS per matrix; never emit disabled headers.
- **Floor:** with `lock = 1`, reject PUT when any resolved setting would be **weaker** than instance (`422` + `field` + message naming the setting).
- **`lock_exempt`:** when true (Admin set), skip floor reject for that service; write audit `security.lock_exempt` with actor + service id.
- **Profile `inherit`:** use instance `default_profile` for matrix unless site sets explicit profile.

Unit tests (required): Baseline vs Strict matrix, lock reject, lock_exempt, Custom header toggle cannot disable locked minimum.

## API (`/api/v1`, Admin+ on writes unless noted)

| Method | Route | Behavior |
|--------|-------|----------|
| GET | `/api/v1/security/policy` | Singleton policy + derived header catalog for UI |
| PUT | `/api/v1/security/policy` | Validate JSON; audit; trigger `proxy.SyncFromDB` |
| GET | `/api/v1/services/{id}/security` | Resolved **effective** settings for UI (floor labels, proxybuild flags) + raw stored row |
| PUT | `/api/v1/services/{id}/security` | Validate + floor/lock; audit; proxy sync; optionally enqueue redeploy if resources changed |
| POST | `/api/v1/services/{id}/security/basic-auth` | Set password → returns **one-time** `password` in response body; stores hash only |
| DELETE | `/api/v1/services/{id}/security/basic-auth` | Clear user+hash |
| GET | `/api/v1/services/{id}/security/events` | `[]` until P2c (stable contract) |

**Authz**

- Policy GET/PUT: **Admin+**
- Service security GET: **Member+** with project access (same as service detail)
- Service security PUT / basic-auth: **Admin+** (or project Admin — match existing service mutation pattern)

**Validation highlights**

- CSP: non-empty strings only; max length bound (e.g. 8 KiB)
- IP allowlist: parse with `net.ParseCIDR`; reject overlaps/invalid; empty list = no IP restriction
- Basic auth user: non-empty safe username charset
- Resources: positive bounds aligned with docker deploy limits

Regenerate OpenAPI + TS client (`make generate`).

## Caddy render (§7)

Extend `RenderInput` / per-route types with `ServiceID` + resolved security snapshot (or load inside `SyncFromDB` loop).

**Per app/static route handle order:**

1. **Headers** — from resolved profile (replace hardcoded `baselineHeadersHandler()` when policy row exists); add CSP header(s) when set (`Content-Security-Policy` or `-Report-Only`)
2. **Encode** — keep P0 zstd/gzip
3. **Access** (when configured):
   - `basic_auth` with stored hash
   - IP allowlist: Caddy matcher → `403` for non-matching (support v6)
   - Maintenance: `respond` 503 with small HTML body (inline or bundled asset)
4. **Upstream** — `reverse_proxy` or `file_server` chain unchanged

**Server block:** keep P0 timeouts; **do not** apply `trusted_proxies` until #78 (even if `cdn_mode` stored).

**Golden tests:** `internal/proxy` — fixtures for Baseline app route, Strict headers, CSP report-only, basic_auth + allowlist matcher shapes; `caddy validate` when harness available.

Wire `proxy.SyncFromDB` after policy/service updates (same as domain changes).

## Resources (stream D)

When `mem_mb` / `cpu` / `pids` change on PUT:

- Persist `service_security` + update deploy/appconfig or docker create input used on next promote
- Trigger existing **blue/green redeploy** path (no extra downtime claim beyond current deploy contract)
- Static sites: **Resources block hidden** — no container limits

Defaults for **new** services: read instance `default_*` when creating `service_security` row on first deploy if no row yet (optional hook in service create — document chosen behavior in generator handoff).

## UI — Policy (`/server/security/policy`)

Per **Security · Policy** artboard:

- **Default profile** `RadioCard`: Baseline / Strict
- **Policy lock** `Switch` + short copy (§3.1)
- **Headers** list: toggle + value field per header (CSP absent)
- **Rate limits:** fields visible, **disabled** + helper “Needs the proxy build” unless add-on `proxy-rate-limit` or combined proxy add-on is `installed` (match #76 catalog ids)
- **CDN / trusted proxies:** mode selector + custom ranges/header — copy notes effect in P1d for Cloudflare auto-ranges
- **Default limits** for new services: mem / CPU / pids
- Save → PUT policy; toast on success

## UI — Application · Security

Per **ApplicationSecurity** artboard; tab **before Settings**:

| Block | MVP in P1c |
|-------|------------|
| Profile | `RadioCard` + instance default label + lock badge |
| Content security | CSP textarea + report-only toggle |
| Access | Basic auth (set password modal, one-time reveal), IP list with “Your IP” hint, maintenance toggle |
| Traffic protection | Segmented controls **disabled** + Install CTA linking to Add-ons when proxybuild missing |
| Resources | Mem / CPU / pids — save triggers redeploy banner if needed |
| Security events | Empty state: “Events appear when rate limiting or WAF is enabled” |

**Static site:** omit Resources + WAF mode; Access + CSP + profile remain.

Extract shared `ServiceSecurityTab` used by both detail pages.

## Acceptance criteria

### Data & resolution (A)

- [ ] Migration `00014_security_policy.sql` + sqlc queries; seed row id=1
- [ ] `service_security` CRUD via sqlc; cascade on service delete
- [ ] `internal/securitypolicy` unit tests: profiles, lock, lock_exempt, header merge

### API (B)

- [ ] `GET/PUT /api/v1/security/policy` with validation + audit
- [ ] `GET/PUT /api/v1/services/{id}/security` returns effective + stored shapes
- [ ] Lock violation → `422` with field name; lock_exempt requires Admin + audit
- [ ] Basic auth POST returns one-time password; GET never returns hash/password
- [ ] `GET …/security/events` returns empty list + stable schema
- [ ] Handler tests for auth, 404, 422 paths

### Proxy (C)

- [ ] `SyncFromDB` applies per-service headers, CSP, basic_auth, allowlist, maintenance
- [ ] Golden JSON tests for at least Baseline, Strict, and one Access combo
- [ ] P0 encode/static cache behavior unchanged when no custom site row

### Resources (D)

- [ ] Changing mem/cpu/pids schedules redeploy using existing deploy pipeline
- [ ] Invalid limits rejected at API

### UI (E)

- [ ] `/server/security/policy` + third nav tab
- [ ] Security tab on applications and static sites (tab order before Settings)
- [ ] Rate limit / WAF / events blocks match disabled/empty behaviors above
- [ ] openapi-fetch types; no `@radix-ui/*`

### Verification gates

- [ ] `go test ./internal/securitypolicy/... ./internal/proxy/...`
- [ ] `go test ./internal/server/... -run 'Security|Policy'`
- [ ] `make generate && make test && make build`
- [ ] `pnpm --dir web exec tsc --noEmit`
- [ ] `./scripts/check-no-radix.sh` if web deps touched

## Out of scope

- CDN **render** + Cloudflare range job (**#78 P1d**)
- Proxybuild install, `proxy.swap`, `rate_limit`, Coraza WAF handlers (**P2**)
- `security_events` persistence and live feed (**P2c+**)
- CrowdSec UI / decisions (**#82 P3**)
- Compose-only services without HTTP routes (unless already on proxy list — follow `ListLiveAppProxyRoutes` scope)
- Valkey security presets (**§4.7**, separate work)
- ADR for panel non-root or event IP retention

## Roadmap

On close: M7 epic **#74** — mark **P1c** complete; unblocks **P1d** (#78 CDN/trusted proxies).

## Change log

| Date | Note |
|------|------|
| 2026-11-07 | Initial planner spec (L, P1c) |
