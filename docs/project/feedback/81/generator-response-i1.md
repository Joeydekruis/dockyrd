# Generator response - issue #81, iteration 1

## Feedback addressed

- n/a (i1 — no prior reviewer feedback)

## Deferred / disputed

- **OpenAPI / TS client regen:** `make generate` is sqlc-only in this repo; service security + events DTOs updated in handwritten `web/lib/security-policy-api.ts` (same pattern as #77). No separate openapi-typescript step wired.
- **Live Coraza log tail:** fixture parse + insert + purge + worker skip-when-missing shipped; production volume layout waits on soft-dep **#79**. Mapping: Host header → `domains.hostname` → `service_id`.
- **Custom inherit WAF:** treated as Baseline floor (**off**), documented here per spec.

## Design compliance (UI)

- Wireframe paths: n/a — design skipped; implement vs `ApplicationSecurity` in `docs/dockyrd-security-design.html`
- Matched: Traffic protection WAF Segmented when `proxy_waf_installed`; Block warning; path exceptions; events table / empty state; static hides WAF controls
- Intentional deltas: added `Segmented` primitive under `web/components/dockyard/ui/` (was missing); rate-limit remains disabled CTA copy only

## Commands run

```text
make generate
# go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate → exit 0

go test ./internal/securitypolicy/... ./internal/proxy/... ./internal/security/events/... -count=1
# ok securitypolicy / proxy / security/events

go test ./internal/server/... -run 'Security|WAF|Events' -count=1
# ok

go test ./internal/db/... -count=1
# ok

pnpm --dir web exec tsc --noEmit
# exit 0

./scripts/check-no-radix.sh
# OK: no @radix-ui in web dependencies

make test
# exit 0 — all packages ok (incl. security/events, securitypolicy, proxy, server)

make build
# pnpm web build + go build bin/dockyrd + bin/dy — exit 0
```

## Soft-dep notes

- `#79`: `coraza_waf` gated solely on addon `proxy-waf` state=`installed`
- `#80`: no `rate_limit` handler in diff
- Ingest mode: **fixture-capable worker** (reads `{DataDir}/caddy/coraza-audit.log` when present; skips if missing)

## Ready for reviewers

- [x] typecheck / lint (radix) / scoped tests
- [x] make test / make build
- Reviewers: `database-reviewer`, `api-reviewer`, `security-reviewer`; `doc-updater` after PASS; frontend-eval optional
