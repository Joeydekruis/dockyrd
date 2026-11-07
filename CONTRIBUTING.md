# Contributing to dockyrd

Thanks for helping build dockyrd. Read [`docs/BRIEFING.md`](docs/BRIEFING.md) for architecture and [`AGENTS.md`](AGENTS.md) for agent/human workflow.

## Branching

- Implement on **`development`**. Do not open long-lived `feature/*` branches.
- `main` merges happen via maintainer release flow.

## Before you open a PR

1. Run relevant checks from [`docs/project/setup.md`](docs/project/setup.md) (`make test`, `make lint`, `make build` when touching Go or `web/`).
2. No `@radix-ui/*` in `web/` — Base UI only.
3. Do not commit secrets or `web/out/`.

## Commits

Use [Conventional Commits](https://www.conventionalcommits.org/) and reference GitHub issues when applicable (`Fixes #n` or `Refs #n`).

## Code of conduct

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
