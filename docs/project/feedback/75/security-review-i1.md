# Security review - 75 i1
## Verdict: PASS
## Boundaries
- `POST /api/v1/security/host-actions/{id}/confirm|revert`: `requireAdmin` (owner/admin) before any runner mutation; unauthenticated returns 401 (tests).
- sqlc queries use bound parameters for `id` and state updates; migration state CHECK constrains invalid rows.
- `dockyard host revert <id>`: local root trust boundary per spec (timer callback); `exec.Command` argv-style args (ULID id) — no shell interpolation.
- `SystemdLockoutTimer`: unit name derived from action ULID; schedule/cancel via `systemd-run`/`systemctl` without user-controlled command fragments.
- FSM + `Revertable`: confirm only from `awaiting_confirm`; post-`confirmed` not revertable via HTTP; idempotent no-op when already `reverted`.
- Startup `Reconcile`: system path reverts expired `awaiting_confirm` only; failures logged, do not fail serve.
## Blockers
- None.
## Notes
- `test_echo` `Revert` trusts `path` inside stored `revert_json` (acceptable for test kind; real actions should validate payloads in later issues).
- Concurrent confirm/revert lacks row-level locking — acceptable for P1a stubs; tighten before live SSH/firewall `Apply` in P1b+.
