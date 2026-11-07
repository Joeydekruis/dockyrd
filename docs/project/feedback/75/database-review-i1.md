# Database review - 75 i1
## Verdict: BLOCK
## Gates
- migrations/tests: pass (`go test ./internal/db/...`)
## Blockers
- `internal/db/queries/host_actions.sql:36` — `UpdateHostAction` updates by `id` only with no `AND state = ?` guard; runner reads row outside the txn (`loadAction`) then writes in `transition`/`failAndRevert`, so concurrent confirm + reconcile/revert (or double confirm) can both succeed and skip or overwrite FSM validation → add conditional update (mirror `LockJob` in `jobs.sql`) and treat 0 rows as `ErrInvalidTransition`; optionally wrap read+update in one txn.
## Notes
- Migration shape matches spec: state `CHECK`, FK `actor_user_id` → `users` `ON DELETE SET NULL`, partial index `idx_host_actions_reconcile` fits `ListHostActionsAwaitingConfirmExpired`.
- `00012_host_actions.sql` has no `+goose Down` (repo convention on other migrations); forward-only is fine for MVP if rollback is not required.
