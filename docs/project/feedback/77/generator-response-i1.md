# Generator response - issue #77, iteration 1

## Feedback addressed
- N/A (iteration 1)

## Deferred / disputed
- CDN trusted-proxy render (#78): `cdn_*` fields persist on policy; no Caddy `trusted_proxies` in render.
- Rate limit / Coraza WAF handlers (P2): stored on policy/service rows; UI disabled without proxy add-on; not in Caddy JSON.
- `security_events` table: `GET …/security/events` returns `{ events: [] }` only.

## Design compliance (UI)
- Wireframe paths: `docs/dockyrd-security-design.html` (ServerSecurityPolicy, ApplicationSecurity) — layout follows shared Security nav + section blocks; `RadioGroup`/`Switch` used where artboards specify RadioCard/Switch (no separate RadioCard primitive in repo).
- Matched / intentional deltas: Traffic protection is informational + Add-ons link; events use EmptyState stub.

## Commands run
```text
go test ./internal/securitypolicy/... ./internal/proxy/...
go test ./internal/server/... -run 'Security|Policy'
make generate
make test   # all packages green except intermittent internal/hostops (env); securitypolicy/proxy/server pass
make build
pnpm --dir web exec tsc --noEmit
./scripts/check-no-radix.sh
```

## Ready for reviewers
- [x] typecheck / lint / test / build (see above)
- [ ] database-reviewer: `00014_security_policy.sql`, FK cascade, singleton policy
- [ ] api-reviewer: policy + service security routes, basic-auth one-time password, 422 lock shape

## Handoff notes
- New services: `service_security` row is **lazy** on first GET/PUT (defaults `profile=inherit`).
- Resource limits: PUT syncs `memLimitMB` on app config and enqueues redeploy when mem/cpu/pids change; deploy uses `securitypolicy.DeployLimits` for Docker nano CPUs + pids.
- Proxy golden tests: `internal/proxy/security_render_test.go` (Strict + access combo).
