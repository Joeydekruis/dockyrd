#!/usr/bin/env bash
# M3 smoke: databases, refs, storage API + web build.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "== M3 smoke (automated gates) =="
go test ./internal/dbkind/... ./internal/refs/... ./internal/deploy/... ./internal/storage/... ./internal/server/...
./scripts/check-no-radix.sh
pnpm --dir web build

cat <<'MANUAL'

== Manual M3 exit (Docker host) ==

1. make build && ./bin/dockyrd serve (Docker running)
2. Databases → create Postgres → Reveal connection → copy DATABASE_URL
3. App env: DATABASE_URL=${{db-slug.DATABASE_URL}} → deploy
4. Object storage → Install Garage → create bucket in a project
5. Optional: Next.js app + Postgres + S3 upload on project network

MANUAL
