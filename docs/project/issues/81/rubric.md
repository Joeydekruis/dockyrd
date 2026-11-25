---
github: https://github.com/Joeydekruis/dockyrd/issues/81
---

## Done bar

- All `spec.md` AC checked with evidence in GitHub handoff.
- Soft-deps **#79** / **#80** documented: no `rate_limit` in diff; `coraza_waf` gated on `proxy-waf` (or shared proxybuild) `installed`.
- Strict never auto-blocks; Block only via explicit site `waf_mode=block`.
- `GET …/security/events` returns persisted rows (fixture ingest acceptable).
- `CHANGELOG.md` patch bump on cycle close per project rules.
- Design: implement vs `ApplicationSecurity` in `docs/dockyrd-security-design.html` — **no** new wireframes.

## Global bar

- Parameterized SQL only (`security_events`, `service_security` alter).
- Events GET scoped by project membership; no cross-tenant leak.
- Client IPs stored only in `security_events`, purged at **7 days**, documented separate from analytics.
- Path exceptions validated server-side; no unsanitized directive injection into Caddy JSON.
- Addon gate enforced in render (not UI-only).
- No secrets in event `detail`/logs beyond IP + rule metadata needed for ops.

## Reviewers

| Subagent | Required | Focus |
|----------|----------|--------|
| `database-reviewer` | **Yes** | `00017` schema, FK cascade, indexes, sqlc, purge query |
| `api-reviewer` | **Yes** | Events list contract, WAF/exceptions on PUT, authz Member/Admin, OpenAPI |
| `security-reviewer` | **Yes** | IP retention, trust boundary on events, directive injection, addon gate |
| `frontend-evaluator` | **Optional (light)** | Only if WAF/events UI diverges from `ApplicationSecurity`; skip if generator matches design + AC |
| `code-reviewer` | **Skip** | Specialists cover DB/API/trust |
| `doc-updater` | **Yes** (after specialists PASS) | Short privacy note: security-event IPs retained 7 days, separate from analytics (`docs/user/` or security section) |

## Verify (paste in handoff)

```bash
go test ./internal/securitypolicy/... ./internal/proxy/...
go test ./internal/server/... -run 'Security|WAF|Events'
# plus ingest/purge package path chosen by generator
make generate
make test
make build
pnpm --dir web exec tsc --noEmit
./scripts/check-no-radix.sh
```

## Handoff evidence

- Full commit SHA on `development`.
- Pointers: `00017_*.sql`, golden WAF fixture path(s), events list sample, Application Security WAF + events UI.
- Explicit note: #79 soft-dep status; whether live log tail or fixture-only ingest shipped.
- Reviewer verdicts table for database / api / security (+ docs).
