#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
out="$root/web/out"
mkdir -p "$out"

stub_html='<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><title>Dockyrd</title></head><body></body></html>'

if [[ ! -f "$out/index.html" ]]; then
  printf '%s\n' "$stub_html" >"$out/index.html"
fi
if [[ ! -f "$out/404.html" ]]; then
  cat >"$out/404.html" <<'EOF'
<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><title>404</title></head>
<body><h1>404</h1></body></html>
EOF
fi

# Placeholders for static-export dynamic routes (briefing §4.2). Go maps real slugs to these paths.
spa_placeholders=(
  "projects/_/index.html"
  "projects/_/index.txt"
  "static-sites/_/index.html"
  "static-sites/_/deployments/_/index.html"
  "applications/_/index.html"
  "databases/_/index.html"
  "object-storage/_/index.html"
)
for rel in "${spa_placeholders[@]}"; do
  f="$out/$rel"
  if [[ -f "$f" ]]; then
    continue
  fi
  mkdir -p "$(dirname "$f")"
  if [[ "$rel" == *.txt ]]; then
    echo "stub" >"$f"
  else
    printf '%s\n' "$stub_html" >"$f"
  fi
done
