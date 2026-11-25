# API review - #81 i1
## Verdict: PASS
## Scope
- routes: `GET/PUT /api/v1/services/{id}/security`, `GET …/security/events`
- DTOs: `web/lib/security-policy-api.ts` (`waf_path_exceptions`, `SecurityEvent`)
## Smoke
- `go test ./internal/server/... -run 'Security|WAF|Events' -count=1`: pass (PUT exceptions + 422 `../`, events list returns inserted waf row)
## Blockers
- None
## Notes
- Authz: PUT uses `ScopeWrite` → owner/admin/member (Viewer denied); events GET uses `ScopeRead` (project-scoped) — matches stub + discuss-gate.
- Handwritten client fields match server DTOs; `make generate` is sqlc-only here (OpenAPI regen N/A, same as #77).
- Optional `limit`/`source` query params exist on server; TS helper omits them (UI uses defaults) — fine for P2c.
