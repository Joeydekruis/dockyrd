#!/usr/bin/env bash
# Manual M2 smoke checklist (Docker host + public URL for GitHub webhooks).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "== M2 smoke (automated gates) =="
go test ./internal/...
./scripts/check-no-radix.sh
pnpm --dir web build

cat <<'MANUAL'

== Manual steps (requires Docker + optional GitHub App) ==

1. make dev (or build + ./bin/dockyrd serve)
2. Settings → Instance → Install GitHub App (needs dockyrd_PUBLIC_URL)
3. Projects → create project → open project → Add app → Deploy (image nginx)
4. Static sites → create site (github repo) → Deploy → Promote older release
5. Push to linked repo → verify webhook deploy (app or static with autoDeploy)

MANUAL
