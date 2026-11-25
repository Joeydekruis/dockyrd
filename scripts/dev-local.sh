#!/usr/bin/env bash
# Local panel: Go API on :8000 reverse-proxies the dashboard to Next.js dev (HMR).
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

pnpm --dir web install

api_pid=""
cleanup() {
	trap - EXIT INT TERM
	if [[ -n "$api_pid" ]]; then
		kill "$api_pid" 2>/dev/null || true
		wait "$api_pid" 2>/dev/null || true
	fi
}
trap cleanup EXIT INT TERM

export dockyrd_DEV_UI="${dockyrd_DEV_UI:-http://localhost:3000}"

echo "→ Panel (API + proxied UI): http://127.0.0.1:8000"
echo "→ Next.js dev (HMR):        $dockyrd_DEV_UI"
echo "   First time: pnpm run init  (or: go run ./cmd/dockyard init)"

go run ./cmd/dockyard serve &
api_pid=$!

# Wait until the API accepts connections (or the go process exits).
for _ in $(seq 1 60); do
	if ! kill -0 "$api_pid" 2>/dev/null; then
		echo "dockyrd serve exited before bind" >&2
		wait "$api_pid" || true
		exit 1
	fi
	if curl -fsS -o /dev/null "http://127.0.0.1:8000/api/v1/health" 2>/dev/null; then
		break
	fi
	sleep 0.25
done

exec pnpm --dir web dev