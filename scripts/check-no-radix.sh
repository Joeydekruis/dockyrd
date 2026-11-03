#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
if rg -q '@radix-ui' "$root/web/package.json" 2>/dev/null; then
  echo "Radix UI is forbidden in web/ (use Base UI)." >&2
  exit 1
fi
if [[ -f "$root/web/pnpm-lock.yaml" ]] && rg -q '@radix-ui' "$root/web/pnpm-lock.yaml"; then
  echo "Radix UI found in lockfile." >&2
  exit 1
fi
echo "OK: no @radix-ui in web dependencies"
