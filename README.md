# Dockyard

Self-hosted, open-source deploy platform: one Go binary, Docker workloads, Caddy edge, SQLite control plane, embedded **Dockyrd** dashboard.

**Status:** M0–M5 complete on `development` (app version **0.0.21**). See [`docs/project/roadmap.md`](docs/project/roadmap.md).

## Install

See [`docs/user/install.md`](docs/user/install.md). Production:

```bash
curl -fsSL https://get.dockyard.dev | sudo sh
```

## Local build

```bash
make build
./bin/dockyard serve
# → http://127.0.0.1:8000  (UI + GET /api/v1/health)
```

- **Branch:** day-to-day work on `development`; `main` is release-ready after merge.
- **Verify:** [`docs/project/setup.md`](docs/project/setup.md) · [`docs/project/m5-exit.md`](docs/project/m5-exit.md)
- **Product brief:** [`docs/BRIEFING.md`](docs/BRIEFING.md)

## License

Apache-2.0 — see [LICENSE](LICENSE) and [NOTICE](NOTICE).
