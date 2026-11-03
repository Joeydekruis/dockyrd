#!/usr/bin/env bash
# M4 smoke: events, host metrics API + analytics web build.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "== M4 smoke (automated gates) =="
go test ./internal/eventlog/... ./internal/metrics/... ./internal/server/...
./scripts/check-no-radix.sh
pnpm --dir web build

cat <<'MANUAL'

== Manual M4 slice (local serve) ==

1. make build && ./bin/dockyard serve
2. Analytics → range pills show CPU/memory/disk/load sparklines after ~1 min
3. Home → project/service counts; Applications lists app services
4. Trigger a deploy → Important events shows deploy row → Dismiss acks it; warning/error rows link via Open service

MANUAL
