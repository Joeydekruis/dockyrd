# Installing dockyrd

dockyrd ships as a single Go binary plus embedded dashboard UI.

**Public docs:** [dockyrd.io/docs/installation](https://dockyrd.io/docs/installation.html)

## Quick install (production)

On Ubuntu 24.04 or Debian 12 (amd64 or arm64):

```bash
curl -fsSL https://raw.githubusercontent.com/Joeydekruis/dockyrd/development/scripts/install.sh | sudo sh
```

When `https://get.dockyrd.dev` is live, the install one-liner on [dockyrd.io](https://dockyrd.io) will point there instead of the raw GitHub script URL.

The installer checks the OS, installs **Docker Engine** on Debian/Ubuntu when missing, installs the **Railpack** CLI when it is not already on the host (default app builder), installs `/usr/local/bin/dockyrd` (server) and `/usr/local/bin/dy` (API CLI for tokens/CI), registers `dockyrd.service`, and prints a **one-time setup token**.

Skip automatic Docker install: `dockyrd_SKIP_DOCKER_INSTALL=1` or pass `--no-docker-install` when piping the script.

Open `http://YOUR_HOST:8000/setup`, complete the wizard, then sign in.

## Manual / development

```bash
git clone https://github.com/Joeydekruis/dockyrd.git
cd dockyrd
make build
./bin/dockyrd init
dockyrd_DATA_DIR=$HOME/.config/dockyrd/data \
dockyrd_CONFIG_DIR=$HOME/.config/dockyrd/etc \
  ./bin/dockyrd serve
```

Set `dockyrd_DEV_UI=http://localhost:3000` when running the Next.js dev server alongside the API.

## Health check

```bash
dockyrd doctor
```

Fix any `✗` lines before going to production. See [troubleshooting](./troubleshooting.md) and [domains & DNS](./domains-dns.md).

## Next steps

- [First app](./first-app.md)
- [Next.js guide](./nextjs.md)
- [Email / SMTP](./email-smtp.md)
- [AI assistants (MCP)](./mcp.md)
- [Production validation](./production-validation.md)

```bash
dockyrd update --check
sudo dockyrd update
sudo systemctl restart dockyrd
```

If an update fails after restart, restore the previous binary:

```bash
sudo dockyrd update --rollback
sudo systemctl restart dockyrd
```

Pre-update SQLite backups land under `/var/lib/dockyrd/backups/`.
