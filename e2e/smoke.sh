#!/usr/bin/env bash
# Minimal E2E: build binary, init control plane, serve, HTTP checks.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

PORT=$((18000 + RANDOM % 8000))
export dockyrd_DATA_DIR="$TMP/data"
export dockyrd_CONFIG_DIR="$TMP/etc"
export dockyrd_LISTEN=":${PORT}"
BASE="http://127.0.0.1:${PORT}"

echo "== e2e smoke =="
make build

"$ROOT/bin/dockyrd" init >/dev/null

"$ROOT/bin/dockyrd" serve &
PID=$!
cleanup() {
  kill "$PID" 2>/dev/null || true
  wait "$PID" 2>/dev/null || true
  rm -rf "$TMP"
}
trap cleanup EXIT

for _ in $(seq 1 60); do
  if curl -sf "${BASE}/api/v1/health" >/dev/null; then
    break
  fi
  sleep 0.25
done

curl -sf "${BASE}/api/v1/health" | grep -q '"status":"ok"'
curl -sf "${BASE}/api/v1/setup/status" | grep -q '"complete":false'
curl -sf "${BASE}/" | grep -qi dockyrd

echo "e2e smoke OK"
