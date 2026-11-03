#!/usr/bin/env bash
# Starts dockyard for Playwright (fixed port for CI).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PORT="${E2E_PORT:-18080}"
export DOCKYARD_DATA_DIR="${DOCKYARD_DATA_DIR:-/tmp/dockyard-e2e/data}"
export DOCKYARD_CONFIG_DIR="${DOCKYARD_CONFIG_DIR:-/tmp/dockyard-e2e/etc}"
export DOCKYARD_LISTEN=":${PORT}"

mkdir -p "$DOCKYARD_DATA_DIR" "$DOCKYARD_CONFIG_DIR"
if [ ! -f "$DOCKYARD_CONFIG_DIR/secret.key" ]; then
  "$ROOT/bin/dockyard" init >/dev/null
fi
exec "$ROOT/bin/dockyard" serve
