#!/usr/bin/env bash
# M5 smoke: doctor CLI, admin package compile, CI-parity gates.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "== M5 smoke =="
go test ./internal/doctor/... ./internal/admin/... ./internal/updater/...
go vet ./...
go test ./...
./scripts/check-no-radix.sh
pnpm --dir web build

echo "== doctor (expect Docker fail on dev Mac OK) =="
go run ./cmd/dockyard doctor || true

echo "M5 smoke gates passed."
