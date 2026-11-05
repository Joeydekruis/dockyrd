#!/usr/bin/env bash
# Starts dockyard for Playwright (fixed port for CI).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PORT="${E2E_PORT:-18080}"
export DOCKYARD_DATA_DIR="${DOCKYARD_DATA_DIR:-/tmp/dockyard-e2e/data}"
export DOCKYARD_CONFIG_DIR="${DOCKYARD_CONFIG_DIR:-/tmp/dockyard-e2e/etc}"
export DOCKYARD_LISTEN=":${PORT}"
# E2E uses a fresh owner; skip mandatory owner 2FA so auth smoke can hit protected routes.
export DOCKYARD_REQUIRE_OWNER_2FA="${DOCKYARD_REQUIRE_OWNER_2FA:-0}"

if [ "${E2E_FRESH:-1}" = "1" ]; then
  rm -rf "$DOCKYARD_DATA_DIR" "$DOCKYARD_CONFIG_DIR"
fi

mkdir -p "$DOCKYARD_DATA_DIR" "$DOCKYARD_CONFIG_DIR"
init_out="$("$ROOT/bin/dockyard" init 2>&1)"
token=$(printf '%s' "$init_out" | sed -n 's/.*token=\([^[:space:]]*\).*/\1/p' | head -n1)
if [ -n "$token" ]; then
  printf '%s' "$token" >"$DOCKYARD_CONFIG_DIR/e2e-setup-token"
fi
exec "$ROOT/bin/dockyard" serve
