---
github: https://github.com/Joeydekruis/dockyrd/issues/75
---

## Done bar

- All checkboxes in `spec.md` satisfied with tests cited in GitHub handoff.
- `go test ./internal/hostops/...` includes FSM, `test.echo`, reconcile (expired deadline), fake lockout timer.
- Migration applies on clean DB; sqlc compiles; no hand-edited `sqlc` output.
- Confirm/revert routes enforce Admin+ and reject invalid states (tests or handler tests).
- `make test` + `make build` green (per `docs/project/setup.md`).

## Global bar

- Parameterized SQL only; secrets not in `plan_json`/`revert_json` logs.
- Invalid FSM transitions never commit partial row updates.
- Timer/revert path does not log raw `revert_json` at info level.

## Reviewers

| Subagent | Required | Focus |
|----------|----------|--------|
| `security-reviewer` | **Yes** | Trust boundaries: Admin+ on mutating routes; timer CLI invocation; no injection via action kinds; audit on transitions; reconcile idempotency |
| `database-reviewer` | **Yes** | `00012_host_actions.sql` constraints, indexes, migration safety, sqlc query shapes |
| `api-reviewer` | Skip | Only two stub routes; security-reviewer covers authz if lean |
| `code-reviewer` | Skip | Specialists sufficient for M framework slice |

## Optional (defer)

- `frontend-evaluator` — no UI in P1a
- E2E systemd timer on real host — P1b+

## Handoff evidence

- Paste `go test ./internal/hostops/...` output.
- Point to migration file and example audit action names.
- Note explicit **non-implementation** of SSH/firewall stubs in handoff comment.
