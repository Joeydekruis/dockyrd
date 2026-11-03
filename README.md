# Dockyard

Self-hosted, open-source deploy platform: one Go binary, Docker workloads, Caddy edge, SQLite control plane, embedded **Dockyrd** dashboard.

## Install

See [`docs/user/install.md`](docs/user/install.md). Production:

```bash
curl -fsSL https://get.dockyard.dev | sudo sh
```

Local build:


- **Branch:** day-to-day work on `development`; `main` is release-ready.
- **Verify:** [`docs/project/setup.md`](docs/project/setup.md) · [`docs/user/troubleshooting.md`](docs/user/troubleshooting.md)
- **Product brief:** [`docs/BRIEFING.md`](docs/BRIEFING.md)
- **Sprints:** [`docs/project/sprints.md`](docs/project/sprints.md)

```bash
make build
./bin/dockyard serve
# → http://127.0.0.1:8000  (UI + GET /api/v1/health)
```

## License

Apache-2.0 — see [LICENSE](LICENSE) and [NOTICE](NOTICE).
