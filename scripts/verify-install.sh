#!/usr/bin/env bash
# Dry-run install path for CI: local binary, temp config/data, no systemd.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

make build

export dockyrd_INSTALL_BIN="$ROOT/bin/dockyrd"
export dockyrd_INSTALL_DY_BIN="$ROOT/bin/dy"
export dockyrd_CONFIG_DIR="$TMP/etc"
export dockyrd_DATA_DIR="$TMP/data"
export dockyrd_BIN_DEST="$TMP/bin/dockyrd"
export dockyrd_DY_BIN_DEST="$TMP/bin/dy"
export dockyrd_SKIP_ROOT=1
export dockyrd_SKIP_SYSTEMD=1

mkdir -p "$TMP/bin"
sh "$ROOT/scripts/install.sh"

test -x "$dockyrd_BIN_DEST"
test -x "$dockyrd_DY_BIN_DEST"
test -f "$dockyrd_CONFIG_DIR/secret.key"
test -f "$dockyrd_DATA_DIR/control.db"

echo "verify-install OK"
