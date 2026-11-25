---
github: https://github.com/Joeydekruis/dockyrd/issues/81
title: P2c — Coraza WAF detect-only & security events
size: L
parent: https://github.com/Joeydekruis/dockyrd/issues/74
depends_on:
  - https://github.com/Joeydekruis/dockyrd/issues/76
soft_depends_on:
  - https://github.com/Joeydekruis/dockyrd/issues/79
  - https://github.com/Joeydekruis/dockyrd/issues/80
milestone: M7 · Security & performance
---

## Summary

Ship **M7 P2c**: resolve and render **Coraza WAF** (`coraza_waf`) per app when proxy WAF is installed; **detect-only** by default (Strict profile never auto-blocks); **Block** as explicit per-site opt-in; **per-path exceptions**; persist WAF hits in **`security_events`**; replace the empty `GET …/security/events` stub with real rows; enable Application Security **Traffic protection** WAF UI + events table.

**Product source:** `docs/project/SECURITY_BRIEFING.md` §3.2, §4.4, §4.6, §5 (`security_events`), §6 (events API), §7 (render order), §8 (`ApplicationSecurity`), §9 (P2c), §10 (WAF fixtures).

**Design artefacts:** none (skip `designer`). Visual contract = **`ApplicationSecurity`** in `docs/dockyrd-security-design.html`.

## Dependencies

| Dep | Strength | Why |
|-----|----------|-----|
| **#76** | Hard | Add-ons catalog + `proxy-waf` / `proxy-rate-limit` state; Security tab gating |
| **#77** | Hard (shipped) | `service_security.waf_mode`, Security tab shell, events stub |
| **#79** (P2a proxybuild / `proxy.swap`) | **Soft** | Real Coraza-capable image install. If open: keep install path aligned with #76 catalog/addons patterns; still ship **render + API + events + UI** gated on `addons.proxy-waf.state = installed` (or shared proxybuild flag when #79 introduces it) |
| **#80** (P2b rate limit) | **Soft / shared proxybuild** | May share proxy image/module; **do not** implement `rate_limit` handler in #81 |

**Assumption:** generator may treat `proxy-waf` `installed` as the sole gate for emitting `coraza_waf`. If #79 later merges proxybuild into one add-on id, follow that catalog id without changing product AC.

## Discuss-gate (PASS — folded)

1. **IP retention:** purge `security_events` older than **7 days** (constant or env/config OK). Separate from analytics; **no** retention slider UI. ADR formalization deferred (§11.2) — document privacy note in user docs via `doc-updater`.
2. **SecRuleEngine:** `DetectionOnly` when effective `waf_mode=detect`; `On` when `block`. Assert via golden JSON render tests; note if `caddy adapt` unavailable in unit tests.
3. **Path exceptions:** new JSON column on `service_security`; Admin/Member editable; rendered into Coraza directives (exclude path prefixes from rules).
4. Soft-deps #79/#80 as above; no rate_limit in this issue.
5. **Event ingest:** persist WAF hits (`source=waf`); prefer Coraza audit/security log under dockyrd data dir / proxy volume; minimum = schema + parse/insert path + worker stub testable with fixture lines.
6. Strict profile → WAF **detect** only (never auto-block). Block = explicit site opt-in.
7. **Static sites:** no WAF mode control (§4.6); events list still OK if shared.
8. **UI:** enable WAF segmented control when proxy WAF installed; path exceptions editor; Block with warning; events table when rows exist.

## Size note (L, five workstreams)

| Stream | Scope |
|--------|--------|
| **A** | Migration `00017_*`: `security_events` + `waf_path_exceptions_json`; sqlc |
| **B** | Effective WAF resolve (inherit / Strict→detect / never auto-block); path validation; extend service security GET/PUT |
| **C** | `prependSecurity` insert `coraza_waf` after encode (before access); gate on addon installed; golden JSON |
| **D** | Log ingest worker stub + purge job; `GET …/security/events` returns real rows |
| **E** | Application Security UI: WAF Segmented, exceptions, Block warning, events table |

## Related ADRs

- **Deferred (§11.2):** security-event IP retention ADR — 7-day purge is product assumption for this issue; no binding ADR required to ship.
- Panel non-root (§11.1) — unchanged.

## Data model

Migration **`00017_security_events.sql`** (next after `00016_deploy_break.sql`):

### `service_security` (alter)

| Column | Type | Notes |
|--------|------|--------|
| `waf_path_exceptions_json` | TEXT NOT NULL DEFAULT `'[]'` | JSON **string array** of path **prefixes** (e.g. `"/api/"`, `"/_next/"`). Empty = no exceptions |

Existing `waf_mode` CHECK (`inherit` \| `off` \| `detect` \| `block`) unchanged.

### `security_events` (new)

| Column | Type | Notes |
|--------|------|--------|
| `id` | INTEGER PK AUTOINCREMENT (or TEXT id if repo prefers ULID — match nearby tables) | |
| `ts` | TEXT NOT NULL | RFC3339 / SQLite datetime; index for purge + list |
| `service_id` | TEXT NOT NULL | FK → `services.id` ON DELETE CASCADE |
| `source` | TEXT NOT NULL | CHECK ∈ `ratelimit`, `waf`, `crowdsec` |
| `rule` | TEXT | nullable / empty OK |
| `path` | TEXT | request path |
| `action` | TEXT | e.g. `detect`, `block`, `allow` (normalize from Coraza) |
| `ip` | TEXT | client IP — **security-only**; not analytics |
| `detail` | TEXT | short freeform / truncated log snippet |

Indexes: `(service_id, ts DESC)`, `(ts)` for purge.

**Retention:** background job / periodic tick deletes rows with `ts` older than **7 days**. Configurable constant (e.g. `SecurityEventsRetentionDays = 7`) OK; no UI.

## Effective WAF resolution

Extend `internal/securitypolicy` (and `RouteSnapshot`):

| Stored `waf_mode` | Effective |
|-------------------|-----------|
| `off` | off — no `coraza_waf` handler |
| `detect` | detect — `SecRuleEngine DetectionOnly` |
| `block` | block — `SecRuleEngine On` |
| `inherit` | From **effective profile**: Baseline → **off**; Strict → **detect**; Custom with inherit → treat as Baseline floor (**off**) unless product already maps Custom inherit differently — document choice in generator handoff |

**Hard rules**

- Strict profile resolution **never** yields `block` unless the site row explicitly stores `waf_mode=block`.
- Floor/lock (if already applied to WAF in #77): weaker-than-instance reject stays server-side; Block is never forced by instance Strict.
- Static kind: API may store/ignore WAF fields; UI hides WAF controls; render **must not** emit `coraza_waf` for static routes.

Path exceptions: validate each entry is a non-empty path starting with `/`, max count (e.g. 64), max length per entry (e.g. 256); reject `..` and newlines → `422`.

## API (`/api/v1`)

| Method | Route | Behavior |
|--------|-------|----------|
| GET | `/api/v1/services/{id}/security` | Include `waf_path_exceptions` (+ effective `waf_mode`); `proxy_waf_installed` already present — keep |
| PUT | `/api/v1/services/{id}/security` | Accept `waf_mode`, `waf_path_exceptions`; validate; audit; `proxy.SyncFromDB` |
| GET | `/api/v1/services/{id}/security/events` | **Real rows** `{ events: [...] }` newest-first; query params optional: `limit` (default 50, max 200), `source` filter |

**Authz**

- Events GET: **Member+** with project read (same as today’s stub).
- Service security PUT (incl. WAF + path exceptions): **Admin or Member** with project write — per discuss-gate (match existing service mutation if Member already can edit service settings; do not elevate Viewer).

**Event DTO (stable)**

```json
{
  "id": "…",
  "ts": "2026-11-25T12:00:00Z",
  "service_id": "…",
  "source": "waf",
  "rule": "…",
  "path": "/api/foo",
  "action": "detect",
  "ip": "203.0.113.10",
  "detail": "…"
}
```

Regenerate OpenAPI + TS client (`make generate`).

**No** public invent-rate-limit routes in this issue. Internal insert helper for ingest/tests OK (not a public write API unless needed for fixture harness — prefer package-private + test inserts).

## Caddy render (§7)

Update `prependSecurity` order to:

1. headers (+ CSP)
2. encode
3. **`(rate_limit)`** — **out of scope**; leave hook/comment only if #80 not merged
4. **`(coraza_waf)`** — when addon installed **and** effective mode ∈ {detect, block}
5. access (basic_auth, IP, maintenance)
6. upstream

**Coraza handler shape** (exact JSON **unconfirmed** in briefing — golden-test the chosen shape):

- Handler name: `coraza_waf` (briefing §4.4)
- Include `order coraza_waf first`, `load_owasp_crs`, and `directives` with `SecRuleEngine DetectionOnly|On`
- Path exceptions → Coraza directives that skip/exclude matching path prefixes (generator picks supported CRS/Coraza syntax; cover in golden + unit parse of directives string)
- Do **not** emit handler when mode `off` or addon not `installed` (sites unchanged after bare install — §4.4)

**Golden tests** (`internal/proxy`):

- detect → `DetectionOnly` in directives
- block → `On`
- off / addon missing → no `coraza_waf` key
- path exceptions appear in directives
- handle order: encode before waf before access
- Optional: `caddy adapt` / validate when binary present; otherwise skip with clear `t.Skip` reason

**WAF fixtures (§10):** unit/fixture lines for known FP shapes (Next.js API path, JSON body markers) proving detect-only render does not set `On`; full live block behavior may wait for #79 image.

Wire `SyncFromDB` after WAF PUT (existing edge sync).

## Event ingest (stream D)

Minimum viable:

1. Document expected log path (e.g. under dockyrd data dir or proxy volume mount used by Coraza audit/security log).
2. Package `internal/security/events` (or similar): parse one log line → `security_events` row (`source=waf`); map `service_id` via host/route labels already known to proxy sync (document mapping strategy — hostname → service).
3. Worker stub: tick or file-tail placeholder callable from jobs; **unit-tested** with fixture lines even if production tail waits on #79 volume layout.
4. Purge job: delete older than 7 days (unit test with frozen clock / inserted old rows).

If live tail is hard without #79: ship schema + parser + insert + GET + purge; wire worker registration with no-op/skip when log file absent.

## UI — Application · Security (`ApplicationSecurity`)

Apps only for WAF controls; static: keep Access/CSP/profile; hide WAF segmented; events section shared.

When `proxy_waf_installed` (or shared proxybuild installed):

- **Traffic protection — WAF:** `Segmented` `off` / `detect` / `block` bound to stored/effective mode
- **Block** selection: warning copy (false positives / downtime risk) before save — match artboard tone
- **Path exceptions:** editable list (add/remove prefixes); save via existing PUT
- Rate-limit controls: leave disabled / P2b CTA — **do not** implement RL handler UX beyond what #77 already greys out

When not installed: keep Install CTA → Add-ons (existing).

**Security events**

- Empty state only when `events.length === 0`
- Else table/list: ts, source, rule, path, action, ip (and detail truncate) — match artboard density
- Poll or refetch on tab focus (TanStack Query); no SSE required in P2c

Reuse `Segmented`, `Switch`, `EmptyState` from `web/components/dockyrd/ui/`.

## Acceptance criteria

### Data (A)

- [ ] Migration `00017_security_events.sql`: `security_events` + `waf_path_exceptions_json`; sqlc queries (insert, list by service, purge)
- [ ] Cascade delete with service

### Resolution & API (B)

- [ ] Effective WAF: inherit/Strict→detect; Block only when stored `block`; unit tests
- [ ] Path exceptions validate + persist on PUT; Member/Admin can edit
- [ ] `GET …/security/events` returns real rows (not always `[]`); authz + 404 tests
- [ ] OpenAPI + TS client regenerated

### Proxy (C)

- [ ] `coraza_waf` emitted only when addon installed and mode detect|block
- [ ] Golden JSON: DetectionOnly vs On; order vs encode/access; exceptions; off/missing addon
- [ ] Static routes never get WAF handler
- [ ] No `rate_limit` handler added by this issue

### Ingest & retention (D)

- [ ] Fixture-line parser inserts `source=waf` rows
- [ ] 7-day purge deletes old rows; documented as separate from analytics
- [ ] Worker skips cleanly when log path missing

### UI (E)

- [ ] App Security: WAF Segmented enabled when proxy WAF installed; Block warning; path exceptions editor
- [ ] Static: no WAF mode control
- [ ] Events table renders when events exist; empty state otherwise
- [ ] No `@radix-ui/*`

### Verification gates

- [ ] `go test ./internal/securitypolicy/... ./internal/proxy/...`
- [ ] `go test ./internal/server/... -run 'Security|WAF|Events'`
- [ ] Ingest/purge package tests
- [ ] `make generate && make test && make build`
- [ ] `pnpm --dir web exec tsc --noEmit`
- [ ] `./scripts/check-no-radix.sh` if web deps touched

## Out of scope

- **`rate_limit` handler** and RL product UX beyond existing disabled state (**#80**)
- Proxybuild release pipeline / `proxy.swap` image install (**#79**) — soft; gate on `installed` flag only
- CrowdSec decisions / bouncer UI (**#82 P3**)
- Retention slider UI; formal accepted ADR for IPs (doc note only)
- Live e2e blocking WAF traffic without proxybuild image (unit/golden suffice)
- Changing analytics IP policy

## Roadmap

On close: M7 epic **#74** — mark **P2c** complete; WAF detect/events unblocked for ops; P3 CrowdSec (#82) still open. Update `docs/project/roadmap.md` P2 row.

## Change log

| Date | Note |
|------|------|
| 2026-11-25 | Initial planner spec (L, P2c); discuss-gate assumptions folded |
