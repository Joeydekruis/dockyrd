# Database review - 81 i1
## Verdict: PASS
## Gates
- migrations/tests: pass (`go test ./internal/db/...`)
## Blockers
- None
## Notes
- `00017` matches AC: `waf_path_exceptions_json` default `'[]'`; `security_events` FK `ON DELETE CASCADE`; `source` CHECK; indexes `(service_id, ts DESC)` + `(ts)`.
- sqlc insert/list/purge all parameterized; purge `WHERE ts < ?` fits 7-day retention (caller must pass comparable RFC3339).
- Down leaves `waf_path_exceptions_json` (documented SQLite DROP COLUMN limit) — acceptable.
