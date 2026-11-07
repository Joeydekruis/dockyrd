---
github: https://github.com/Joeydekruis/dockyrd/issues/77
---

## Done bar

- All `spec.md` AC checked with evidence in GitHub handoff.
- Depends on **#76** merged: Security nav + addons state for proxybuild gating in Policy UI.
- P0 baseline in `internal/proxy/baseline.go` remains; policy-driven render **supersedes** hardcoded headers when configured.
- `CHANGELOG.md` patch bump on cycle close per project rules.

## Global bar

- Admin+ on policy writes; service security writes match existing service RBAC; audit on policy, lock_exempt, basic-auth set/clear.
- Parameterized SQL only (`security_policy`, `service_security`).
- Basic auth secrets: bcrypt hash at rest; one-time password in POST response only; never in logs.
- IP allowlist and CSP validated server-side; no shell in render path.
- Lock floor enforced server-side (422), not UI-only.

## Reviewers

| Subagent | Required | Focus |
|----------|----------|--------|
| `database-reviewer` | **Yes** | `00014` migration constraints, FK cascade, singleton policy, sqlc queries |
| `api-reviewer` | **Yes** | Policy + service security routes, 422 lock shape, basic-auth contract, OpenAPI parity |
| `frontend-evaluator` | **Optional (light)** | Policy + Application Security vs `docs/dockyrd-security-design.html`; run if UI surface is large |
| `security-reviewer` | Skip (default) | Revisit only if reviewer finds auth/hash gaps — specialists listed above are primary |
| `code-reviewer` | Skip | api + database cover contract |

## Verify (paste in handoff)

```bash
go test ./internal/securitypolicy/... ./internal/proxy/...
go test ./internal/server/... -run 'Security|Policy|ServiceSecurity'
make generate
make test
make build
pnpm --dir web exec tsc --noEmit
./scripts/check-no-radix.sh
```

## Handoff evidence

- Full commit SHA on `development`.
- Pointers: `00014_security_policy.sql`, golden proxy fixture path, Policy tab + app Security tab routes.
- Note: CDN render (#78), rate/WAF handlers (P2), events table (P2c+) explicitly deferred.
- Optional: `curl -I` against fixture showing HSTS + CSP after PUT.
