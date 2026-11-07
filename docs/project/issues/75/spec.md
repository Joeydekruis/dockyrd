---
github: https://github.com/Joeydekruis/dockyrd/issues/75
title: P1a — Host actions framework (internal/hostops)
size: M
parent: https://github.com/Joeydekruis/dockyrd/issues/74
milestone: M7 · Security & performance
---

## Summary

Introduce **`internal/hostops`**: typed host mutations with **Plan → Apply → Verify → optional lockout confirm → Revert**, persisted in SQLite **`host_actions`**, audited on every state transition. Lockout-sensitive actions schedule a **systemd one-shot timer** that survives panel restarts; **startup reconcile** reverts expired `awaiting_confirm` rows. Real SSH/firewall/unattended/proxy behavior is **out of scope** (#76, P1b, P2a) — register v1 action **kinds as stubs** plus one **test action** for unit coverage.

**Product source:** `docs/project/SECURITY_BRIEFING.md` §3.3, §5 (`host_actions`), §6 (confirm/revert routes), §10–§11 (tests/ADR posture), **Bijlage B** (state machine).

**Assumptions (epic #74):** panel continues running as **root** with Docker socket; no privilege-separation helper daemon until a future ADR (§11.1). UI preview/lockout UX is **#76**; this issue is the foundation only.

## Design artefacts

**None** — skip `designer`. Generator uses briefing + existing API/auth patterns.

## Related ADRs

- **Deferred:** panel non-root + helper daemon (§11.1) — not required for P1a.

## Package layout (target)

| Path | Role |
|------|------|
| `internal/hostops/action.go` | `Action` interface, `Plan` preview types, kind constants |
| `internal/hostops/registry.go` | Register known kinds → constructor (stubs return `ErrNotImplemented` or no-op stubs) |
| `internal/hostops/fsm.go` | State machine + allowed transitions (Bijlage B) |
| `internal/hostops/runner.go` | Orchestrate Plan/Apply/Verify/Confirm/Revert; **DB transaction per transition** |
| `internal/hostops/lockout.go` | `LockoutTimer` interface + systemd integration |
| `internal/hostops/reconcile.go` | Startup + periodic safety pass for expired deadlines |
| `internal/hostops/actions/*` | Per-kind files (stubs) + `test_echo` (minimal real action) |
| `internal/db/migrations/00012_host_actions.sql` | Table + indexes |
| `internal/db/queries/host_actions.sql` | sqlc |
| `internal/server/api_security_host_actions.go` | Confirm/revert HTTP stubs |
| `cmd/dockyard` | `dockyard host revert <id>` subcommand for timer callback |

## `Action` contract

Each kind implements (no shell strings, no user-controlled command fragments):

| Method | Behavior |
|--------|----------|
| `Kind() string` | Stable id, e.g. `firewall.ensure` |
| `Plan(ctx) (Plan, error)` | Serializable preview → stored in `plan_json` |
| `Apply(ctx) (RevertPayload, error)` | Mutate host; return opaque revert blob → `revert_json` |
| `Verify(ctx) error` | Post-apply check; **failure triggers automatic Revert** |
| `Revert(ctx, RevertPayload) error` | Restore prior state |

**Lockout flag:** actions that can lock the operator out (SSH, firewall — briefing §3.3) expose `RequiresLockoutConfirm() bool`. After successful `Verify`, transition to `awaiting_confirm`, set `confirm_deadline = now + 120s`, schedule timer. Other actions may skip straight to `confirmed` when product allows (test action: configurable in tests).

**v1 stub kinds** (register only; `Plan` may return static placeholder preview; `Apply`/`Verify`/`Revert` return `ErrNotImplemented` unless/until later issues):

- `firewall.ensure`
- `ssh.passwordauth.disable`
- `unattended.enable`
- `crowdsec.bouncer.install`
- `proxy.swap`

## State machine (Bijlage B)

States: `planned`, `applied`, `verified`, `awaiting_confirm`, `confirmed`, `reverted`, `failed`, `cancelled`.

```
planned ──► applied ──► verified ──► awaiting_confirm ──► confirmed
   │           │            │               │
   │           │            └─► reverted    └─► reverted (timer or manual)
   │           └─► failed (verify failed) ──► reverted
   └──► cancelled
```

- Only `internal/hostops` may change `state` (enforce via unexported transition helpers + tests).
- **Invalid transitions** return a typed error; no partial DB writes.
- Every successful transition writes **`audit_log`** via `internal/audit` (action pattern e.g. `host_action.planned`, `host_action.applied`, … with `target_kind=host_action`, `target_id`, `meta.kind`).

## SQLite: `host_actions`

Migration **`00012_host_actions.sql`** (next goose number after `00011`):

| Column | Type | Notes |
|--------|------|-------|
| `id` | TEXT PK | ULID |
| `kind` | TEXT NOT NULL | Action kind |
| `state` | TEXT NOT NULL | CHECK constrained to known states |
| `plan_json` | TEXT | JSON preview |
| `revert_json` | TEXT | Opaque revert payload |
| `actor_user_id` | TEXT | FK `users(id)` nullable for system/timer revert |
| `applied_at` | INTEGER | Unix |
| `confirm_deadline` | INTEGER | Unix; set when entering `awaiting_confirm` |
| `confirmed_at` | INTEGER | |
| `reverted_at` | INTEGER | |
| `last_error` | TEXT | Optional, for `failed` |
| `created_at` / `updated_at` | INTEGER | |

Indexes: `(state, confirm_deadline)` for reconcile; `(kind, created_at DESC)` for future UI lists.

sqlc queries: insert row, get by id, update state + timestamps + error, list `awaiting_confirm` with `confirm_deadline < ?`, list in-flight for admin diagnostics (optional).

## Lockout timer integration

**Goal:** after `Apply` on lockout actions, a revert runs even if `dockyard serve` crashes (briefing §3.3).

1. **`LockoutTimer` interface** — `Schedule(ctx, actionID string, notAfter time.Time) error`, `Cancel(ctx, actionID string) error`.
2. **Production implementation** — schedule transient systemd unit, aligned with briefing:
   - Prefer `systemd-run --on-active=120s --unit=dockyard-host-revert-<id>` invoking **`dockyard host revert <id>`** (exact CLI spelling in implementation).
   - Document in code comments: requires Dockyard on `PATH` as installed by packaging; timer runs as root (same as panel today).
   - Reuse existing systemd familiarity (`internal/resources` D-Bus patterns) where it reduces fragility; **fallback to `exec` of `systemd-run`** is acceptable for P1a.
3. **On Confirm** — cancel timer, set `confirmed_at`, transition `awaiting_confirm` → `confirmed`.
4. **On manual Revert** (API or timer) — idempotent: if already `reverted`, no-op success.

**Test double:** `hostops` accepts injectable `LockoutTimer` (memory/fake) in tests — no real systemd in unit tests.

## Startup reconcile

- Hook from existing server bootstrap (`startBackground` in `internal/server/runtime.go` or equivalent single startup hook): **`hostops.Reconcile(ctx, q, registry, timer)`** once DB is open.
- For each row in `awaiting_confirm` with `confirm_deadline < now`: run **Revert** path (same as timer), transition → `reverted`, audit `host_action.reverted` with `meta.source=startup_reconcile`.
- Log count at `info`; do not fail `serve` if one revert errors (log `error`, continue others).

Optional light periodic re-check (e.g. 60s) — **nice-to-have**, not required if startup + timer are correct.

## API stubs (Huma / existing mux style)

Prefix **`/api/v1/security`** (repo convention; briefing §6 shorthand `/api/security`).

| Route | Role |
|-------|------|
| `POST /api/v1/security/host-actions/{id}/confirm` | Admin+; session auth; `awaiting_confirm` → `confirmed`; cancel timer |
| `POST /api/v1/security/host-actions/{id}/revert` | Admin+; any revertable state per FSM; used by UI and timer CLI |

**P1a scope for HTTP:**

- Implement confirm + revert handlers wired to `hostops` runner.
- Return RFC 9457-style errors for wrong state / not found (match existing API helpers).
- **No** public “start action” HTTP yet (addons/checks land in P1b) — tests may drive runner directly or add internal test helper.

Regenerate OpenAPI/TS only if project Makefile already does for new routes (`make generate`).

## CLI

- `dockyard host revert <id>` — loads config/DB, runs revert for timer; exit non-zero on hard failure (for systemd `OnFailure` logging).
- Document in brief code comment; user-facing docs optional until #76.

## Minimal test action

Implement **`test.echo`** (or `hostops.test.echo`):

- `Plan`: returns fixed JSON preview.
- `Apply`: writes revert payload (e.g. marker file under `cfg.DataDir` or in-memory store in tests via temp dir).
- `Verify`: checks marker.
- `Revert`: removes marker.
- Used in unit tests for full FSM + lockout path with fake timer.

## Acceptance criteria

### Framework

- [ ] `internal/hostops` package with `Action` interface, registry, runner, FSM enforcing Bijlage B transitions.
- [ ] Unit tests cover valid/invalid transitions and verify-failure → revert.

### Persistence

- [ ] Goose migration `00012_host_actions.sql` + sqlc queries; `make generate` updated artefacts committed.
- [ ] Runner persists `plan_json` / `revert_json` and timestamps per transition.

### Audit

- [ ] Each FSM transition records `audit_log` with stable action names and `host_action` target.

### Lockout + reconcile

- [ ] `LockoutTimer` production adapter documented; confirm cancels scheduled revert.
- [ ] `dockyard host revert <id>` subcommand invokes same revert path as API.
- [ ] Startup reconcile reverts expired `awaiting_confirm` rows (unit test with fake clock or injected `now`).

### Stubs

- [ ] v1 kinds registered; stub `Plan`/`Apply` behavior does not mutate real firewall/SSH.
- [ ] `test.echo` (or equivalent) completes full happy path in tests.

### API

- [ ] `POST .../confirm` and `POST .../revert` registered, Admin+-only, integration/unit tests for auth + state guards.

### Verification gates

- [ ] `go test ./internal/hostops/...`
- [ ] `go test ./internal/db/...` (if queries touched)
- [ ] `make test` and `make build` green

## Out of scope

- Implementing `firewall.ensure`, `ssh.passwordauth.disable`, unattended, CrowdSec bouncer, `proxy.swap` (later issues).
- Security dashboard UI, baseline checks API (#76 / P1b).
- `security_policy`, addons tables, site security tab.
- Helper daemon / non-root panel ADR.
- E2E VM SSH lockout test (briefing §10 — track under P1b/E2E issue).

## Roadmap

On close: M7 row already tracks epic #74; note **P1a complete** in handoff; unblocks **#76** (P1b SSH UI + fix flow).

## Change log

| Date | Note |
|------|------|
| 2026-11-07 | Initial planner spec (M, P1a only) |
