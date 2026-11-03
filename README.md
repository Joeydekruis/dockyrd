# Dockyard

Self-hosted, open-source deploy platform: one Go binary, Docker workloads, Caddy edge, SQLite control plane, embedded **Dockyrd** dashboard.

**Status:** M0–M5 shipped; latest release **v0.0.22** on `main` (see [releases](https://github.com/Joeydekruis/dockyrd/releases)).

## Install

See [`docs/user/install.md`](docs/user/install.md). Production:

```bash
curl -fsSL https://get.dockyard.dev | sudo sh
```

Set `DOCKYARD_REQUIRE_OWNER_2FA=1` in `/etc/dockyard/dockyard.env` (see [`packaging/dockyard.env.example`](packaging/dockyard.env.example)).

## Local build

```bash
make build
./bin/dockyard serve
# → http://127.0.0.1:8000  (UI + GET /api/v1/health)
```

- **Branch:** day-to-day work on `development`; `main` is release-ready after merge.
- **Verify:** [`docs/project/setup.md`](docs/project/setup.md) · [`docs/user/production-validation.md`](docs/user/production-validation.md)
- **Product brief:** [`docs/BRIEFING.md`](docs/BRIEFING.md)

## License

Apache-2.0 — see [LICENSE](LICENSE) and [NOTICE](NOTICE).
