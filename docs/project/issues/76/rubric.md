---
github: https://github.com/Joeydekruis/dockyrd/issues/76
---

## Done bar

- All `spec.md` AC checked with evidence in GitHub handoff.
- `#75` hostops FSM unchanged in behavior except real action registrations; confirm/revert routes still pass tests.
- No Radix; `./scripts/check-no-radix.sh` passes if web deps touched.
- `CHANGELOG.md` patch bump on cycle close per project rules.

## Global bar

- Admin+ on all security write routes; no shell injection in host actions; secrets not logged in plan/revert JSON at info level.
- Parameterized SQL only for `addons` migration.
- Failed verify on host actions triggers revert (runner #75 contract).
- SSH fix blocked without key-based session (pre-check enforced server-side, not UI-only).

## Reviewers

| Subagent | Required | Focus |
|----------|----------|--------|
| `api-reviewer` | **Yes** | `/api/v1/security/checks`, addons plan/install/uninstall, fix endpoint shapes; OpenAPI parity |
| `security-reviewer` | **Yes** | SSH/firewall lockout, pre-checks, nft rules vs Docker, authz, audit, addon install rollback |
| `frontend-evaluator` | **Optional (light)** | Overview + Add-ons vs `docs/dockyrd-security-design.html` boards; run if UI diff is large |
| `database-reviewer` | Light | Only if migration non-trivial — else `security-reviewer` + generator gates suffice |
| `code-reviewer` | Skip | Specialists cover trust boundaries |

## Verify (paste in handoff)

```bash
go test ./internal/securitychecks/... ./internal/hostops/... ./internal/security/...
go test ./internal/server/ -run 'Security|HostAction'
make generate
make test
make build
pnpm --dir web exec tsc --noEmit
./scripts/check-no-radix.sh
```

## Handoff evidence

- Full commit SHA on `development`.
- Pointers: migration `00013_addons.sql`, example check JSON, screenshot or steps for SSH lockout confirm on Overview.
- Note CrowdSec/proxy add-ons **catalog-only** in this slice.
