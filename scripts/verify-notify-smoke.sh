#!/usr/bin/env bash
# Notification channels: unit gates + web build (no live Slack required).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "== Notifications smoke =="
go test ./internal/notify/... ./internal/netsafe/... ./internal/watch/... ./internal/reconcile/...
./scripts/check-no-radix.sh
pnpm --dir web exec tsc --noEmit

cat <<'MANUAL'

== Manual (local serve) ==

1. make build && ./bin/dockyard serve
2. Settings → Server → Channels → add Slack webhook → Test → expect ok
3. Trigger failed deploy (or use test channel) → message within ~10s
4. Delivery log shows sent/retry states; dead delivery can Retry

MANUAL
