# dockyrd

Self-hosted, open-source deploy platform: one Go binary, Docker workloads, Caddy edge, SQLite control plane, embedded **Dockyrd** dashboard.

**Status:** M0–M5 shipped; see [releases](https://github.com/Joeydekruis/dockyrd/releases) for the current version.

## Install

Full guide: [dockyrd.io/docs/installation](https://dockyrd.io/docs/installation.html) · [`docs/user/install.md`](docs/user/install.md).

On a Linux VPS (**Ubuntu 24.04** or **Debian 12**, amd64 or arm64):

```bash
curl -fsSL https://raw.githubusercontent.com/Joeydekruis/dockyrd/development/scripts/install.sh | sudo sh
```

The script installs **Docker Engine** when it is missing (official apt repo), downloads the latest [release](https://github.com/Joeydekruis/dockyrd/releases) binary, runs `dockyrd init`, and enables `dockyrd.service`. Use the **setup URL** from the output with your server’s **public IP** on port **8000** — DNS is not required for first setup.

- Other distros: install Docker yourself, then re-run the script (or set `dockyrd_SKIP_DOCKER_INSTALL=1` if Docker is already managed).
- Production hardening: set `dockyrd_REQUIRE_OWNER_2FA=1` in `/etc/dockyrd/dockyrd.env` (see [`packaging/dockyrd.env.example`](packaging/dockyrd.env.example)).

When `https://get.dockyrd.dev` is live, the install one-liner will be:

```bash
curl -fsSL https://get.dockyrd.dev | sudo sh
```

## Local development

**Prerequisites:** Go, pnpm, and (for deploy features) Docker. Details: [`docs/project/setup.md`](docs/project/setup.md).

First-time setup:

```bash
pnpm --dir web install
pnpm run init                    # one-time: prints setup URL
```

### Hot reload (dashboard + API)

One command starts the Go API on **:8000** and Next.js (Turbopack) on **:3000**. Non-API traffic on `:8000` is proxied to the dev server so cookies and `/api/v1/*` stay on one origin.

```bash
pnpm run dev
# Open http://127.0.0.1:8000  (not :3000 alone)
```

Split across two terminals:

```bash
pnpm run dev:api    # dockyrd_DEV_UI=http://localhost:3000
pnpm run dev:web
```

Environment override: `dockyrd_DEV_UI` (default `http://localhost:3000`).

### Embedded UI (no HMR)

Production-like build with the static export embedded in the binary:

```bash
pnpm run build
pnpm run serve
# → http://127.0.0.1:8000  (UI + GET /api/v1/health)
```

UI-only iteration (components, no live API on the same origin): `pnpm run dev:web` or `pnpm --dir web storybook` (:6006).

- **Branch:** day-to-day work on `development`; `main` is release-ready after merge.
- **Verify:** [`docs/project/setup.md`](docs/project/setup.md) · [`docs/user/production-validation.md`](docs/user/production-validation.md)
- **Product brief:** [`docs/BRIEFING.md`](docs/BRIEFING.md)

## License

Apache-2.0 — see [LICENSE](LICENSE) and [NOTICE](NOTICE).
