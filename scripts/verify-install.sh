#!/usr/bin/env bash
# Dry-run install path for CI: local binary, temp config/data, no systemd.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

make build

export DOCKYARD_INSTALL_BIN="$ROOT/bin/dockyard"
export DOCKYARD_CONFIG_DIR="$TMP/etc"
export DOCKYARD_DATA_DIR="$TMP/data"
export DOCKYARD_BIN_DEST="$TMP/bin/dockyard"
export DOCKYARD_SKIP_ROOT=1
export DOCKYARD_SKIP_SYSTEMD=1

mkdir -p "$TMP/bin"
sh "$ROOT/scripts/install.sh"

test -x "$DOCKYARD_BIN_DEST"
test -f "$DOCKYARD_CONFIG_DIR/secret.key"
test -f "$DOCKYARD_DATA_DIR/control.db"

echo "verify-install OK"
