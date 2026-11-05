#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
mkdir -p "$root/web/out"
if [[ ! -f "$root/web/out/index.html" ]]; then
  cat >"$root/web/out/index.html" <<'EOF'
<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><title>Dockyrd</title></head>
<body><p>Run <code>pnpm --dir web build</code> for the dashboard export.</p></body></html>
EOF
fi
if [[ ! -f "$root/web/out/404.html" ]]; then
  cat >"$root/web/out/404.html" <<'EOF'
<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><title>404</title></head>
<body><h1>404</h1></body></html>
EOF
fi
