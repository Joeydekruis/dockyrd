#!/usr/bin/env bash
# Runnable subset of production validation (local / CI). Full VPS checklist: docs/user/production-validation.md
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "== Production validation (automated) =="
./scripts/verify-m3-smoke.sh >/dev/null
./scripts/verify-m4-smoke.sh >/dev/null
./scripts/verify-m5-smoke.sh >/dev/null
./scripts/verify-install.sh
make e2e
if command -v docker >/dev/null 2>&1; then
  go test -tags=integration ./internal/docker/... ./internal/server/... 2>/dev/null || go test -tags=integration ./...
fi
echo "OK: automated production validation gates passed."
