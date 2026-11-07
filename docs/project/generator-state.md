# Generator state

**Updated:** 2026-11-07

## In progress

**M7 Security** (epic #74).

- **#77 (P1c policy + app Security tab)** — generator i1 complete; `00014_security_policy.sql`, `internal/securitypolicy`, policy/service security APIs, proxy access render, UI policy + tab. Feedback: `docs/project/feedback/77/generator-response-i1.md`. Ready for `database-reviewer` + `api-reviewer`.
- **#77 (P1c policy + app Security tab)** — generator i1 on `development`; pending api/database review.
- **#76 (P1b)** — i1 + uninstall/SSH/apply fixes; tests + `make build` green.
- **#75 (P1a hostops)** — prior slice; FSM/runner reused.

## #76 deliverables (i1)

- `internal/securitychecks/` — eight baseline checks, 24h SQLite cache, tests.
- `internal/security/addons/` — catalog, merge, install job orchestration.
- `internal/hostops/` — real `firewall.ensure`, `ssh.passwordauth.disable`, `unattended.enable` (dev-host paths under `DataDir/host`).
- `internal/db/migrations/00013_addons.sql` + sqlc queries.
- `/api/v1/security/checks`, `…/run`, `…/checks/{id}/fix`, addons plan/install/uninstall (Admin+).
- Web: `/server/security`, `/server/security/addons`, sidebar Security nav, UI `StatusRow` + `CodeBlock`.

## Verification (2026-11-07)

```text
go test ./internal/securitychecks/... ./internal/hostops/... ./internal/security/...  # ok
go test ./internal/server/ -run 'Security|HostAction'                                 # ok
pnpm --dir web exec tsc --noEmit                                                    # ok
make build                                                                          # ok
```

## Last ship

Prior: **#75** hostops framework; **#72** terminal @ v0.6.0.
