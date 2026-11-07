#!/usr/bin/env bash
# Measure RSS of dockyrd process (Linux). Briefing §19 targets: panel < 60 MB RSS.
set -euo pipefail
pid="${1:-$(pgrep -xo dockyrd 2>/dev/null || true)}"
if [ -z "$pid" ]; then
  echo "usage: $0 [pid]  (or run dockyrd serve first)" >&2
  exit 1
fi
if [ ! -r "/proc/$pid/status" ]; then
  echo "cannot read /proc/$pid/status" >&2
  exit 1
fi
rss_kb=$(awk '/^VmRSS:/ {print $2}' "/proc/$pid/status")
rss_mb=$((rss_kb / 1024))
echo "dockyrd pid=$pid RSS=${rss_mb}MB (${rss_kb} kB)"
if [ "$rss_mb" -gt 60 ]; then
  echo "warn: above 60 MB briefing target" >&2
  exit 2
fi
