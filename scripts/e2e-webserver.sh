#!/usr/bin/env bash
# Starts dockyrd for Playwright (fixed port for CI).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PORT="${E2E_PORT:-18080}"
export dockyrd_DATA_DIR="${dockyrd_DATA_DIR:-/tmp/dockyrd-e2e/data}"
export dockyrd_CONFIG_DIR="${dockyrd_CONFIG_DIR:-/tmp/dockyrd-e2e/etc}"
export dockyrd_LISTEN=":${PORT}"
# E2E uses a fresh owner; skip mandatory owner 2FA so auth smoke can hit protected routes.
export dockyrd_REQUIRE_OWNER_2FA="${dockyrd_REQUIRE_OWNER_2FA:-0}"

if [ "${E2E_FRESH:-1}" = "1" ]; then
  rm -rf "$dockyrd_DATA_DIR" "$dockyrd_CONFIG_DIR"
fi

mkdir -p "$dockyrd_DATA_DIR" "$dockyrd_CONFIG_DIR"
init_out="$("$ROOT/bin/dockyrd" init 2>&1)"
token=$(printf '%s' "$init_out" | sed -n 's/.*token=\([^[:space:]]*\).*/\1/p' | head -n1)
if [ -n "$token" ]; then
  printf '%s' "$token" >"$dockyrd_CONFIG_DIR/e2e-setup-token"
fi
exec "$ROOT/bin/dockyrd" serve
