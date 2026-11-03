#!/usr/bin/env bash
# Probe panel health during redeploy (manual / §19). Usage:
#   BASE=http://127.0.0.1:8000 DURATION=120 ./e2e/scripts/load-health.sh
set -euo pipefail
BASE="${BASE:-http://127.0.0.1:8000}"
DURATION="${DURATION:-60}"
end=$((SECONDS + DURATION))
fail=0
while [ "$SECONDS" -lt "$end" ]; do
  if ! curl -sf "$BASE/api/v1/health" >/dev/null; then
    fail=$((fail + 1))
    echo "health fail at $(date -u +%H:%M:%S)" >&2
  fi
  sleep 1
done
echo "done: $fail failed probes in ${DURATION}s"
exit 0
