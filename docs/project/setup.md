# Local verification

Commands below match **briefing §5 / §14**. Until the Makefile and workflows land, run only what exists; expand this file as scripts appear.

## Prerequisites

- Go ≥ 1.25 (when `go.mod` exists)
- pnpm (for `web/`)
- Docker (integration + e2e)
- Optional: `golangci-lint`, `staticcheck`, `cosign`

## Expected targets (Makefile)

| Target | Purpose |
|--|--|
| `make dev` | Local panel + hot reload web |
| `make build` | `pnpm --dir web build` + embed + `go build` |
| `make test` | Unit Go tests |
| `make lint` | Go + Biome |
| `make generate` | sqlc, OpenAPI, TS types |
| `make e2e` | Install + Playwright |

## Go (when present)

```bash
go vet ./...
go test ./...
staticcheck ./...          # if installed
golangci-lint run          # if installed
go test -tags=integration ./...   # Docker required
```

## Web (`web/`)

```bash
pnpm --dir web install
pnpm --dir web lint
pnpm --dir web exec tsc --noEmit   # or package "typecheck" script
pnpm --dir web build
./scripts/check-no-radix.sh        # when script exists
```

## M2 slice (GitHub, builds, static sites)

After M2 backend/UI changes on `development`:

```bash
./scripts/verify-m2-smoke.sh
```

Optional env for builds: `DOCKYARD_BUILD_TIMEOUT` (default `20m`), `DOCKYARD_MIN_FREE_DISK_PERCENT` (default `10`), `DOCKYARD_BUILDX_BUILDER` (default `dy-buildkit`), `DOCKYARD_PUBLIC_URL` for GitHub App callbacks/webhooks.

## M5 slice (doctor, admin CLI, CI)

```bash
./scripts/verify-m5-smoke.sh
go run ./cmd/dockyard doctor
```

## Full pre-release check

Use the matrix in **`.cursor/skills/push-to-main/SKILL.md`** before merging to `main`.

## MCP / env

Document secrets here as they appear (GitHub App, test tokens). Never commit values — use env files listed in `.gitignore`.
