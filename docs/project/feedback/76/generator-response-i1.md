# Generator response - issue #76, iteration 1

## Feedback addressed
- n/a (iteration 1)

## Deferred / disputed
- Doctor baseline hook (spec stretch): not wired in this slice.
- Addon uninstall host revert: marks DB `not_installed` only; full nft/sshd reversal follow-up if product wants per-addon undo chains.
- 24h background ticker + inbox notification on new failures: cache + manual POST run only.

## Design compliance (UI)
- Wireframe paths: `docs/dockyrd-security-design.html` (Security · Overview / Add-ons); design-index sidebar note updated.
- Matched / intentional deltas: tabs via route links (`/server/security`, `/server/security/addons`); lockout banner as inline copy (no global announcement bar).

## Commands run
```text
make generate # ok
go test ./internal/securitychecks/... ./internal/hostops/... ./internal/security/... # ok (after TestStubApplyNotImplemented → crowdsec stub)
go test ./internal/server/ -run 'Security|HostAction' # ok
pnpm --dir web exec tsc --noEmit # ok
make build # ok
```

## Ready for reviewers
- [x] typecheck / test / build for touched surfaces
- [ ] `api-reviewer` — security checks/addons/fix routes
- [ ] `security-reviewer` — SSH/firewall lockout, addon install job

## Reviewer handoff
- Issue: #76
- Iteration: i1
- Subagents: api-reviewer, security-reviewer (per rubric)
- Design artefacts: `docs/dockyrd-security-design.html`
- Key paths: `00013_addons.sql`, `internal/securitychecks/`, `internal/server/api_security_*.go`, `web/app/(shell)/server/security/`
