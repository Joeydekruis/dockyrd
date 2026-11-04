# Installing Dockyard

Dockyard ships as a single Go binary plus embedded dashboard UI.

**Public docs:** [dockyrd.io/docs/installation](https://dockyrd.io/docs/installation.html)

## Quick install (production)

On Ubuntu 24.04 or Debian 12 (amd64 or arm64):

```bash
curl -fsSL https://raw.githubusercontent.com/Joeydekruis/dockyrd/development/scripts/install.sh | sudo sh
```

When `https://get.dockyard.dev` is live, the install one-liner on [dockyrd.io](https://dockyrd.io) will point there instead of the raw GitHub script URL.

The installer checks the OS, installs **Docker Engine** on Debian/Ubuntu when missing, installs `/usr/local/bin/dockyard`, registers `dockyard.service`, and prints a **one-time setup token**.

Skip automatic Docker install: `DOCKYARD_SKIP_DOCKER_INSTALL=1` or pass `--no-docker-install` when piping the script.

Open `http://YOUR_HOST:8000/setup`, complete the wizard, then sign in.

## Manual / development

```bash
git clone https://github.com/Joeydekruis/dockyrd.git
cd dockyrd
make build
./bin/dockyard init
DOCKYARD_DATA_DIR=$HOME/.config/dockyard/data \
DOCKYARD_CONFIG_DIR=$HOME/.config/dockyard/etc \
  ./bin/dockyard serve
```

Set `DOCKYARD_DEV_UI=http://localhost:3000` when running the Next.js dev server alongside the API.

## Health check

```bash
dockyard doctor
```

Fix any `✗` lines before going to production. See [troubleshooting](./troubleshooting.md) and [domains & DNS](./domains-dns.md).

## Next steps

- [First app](./first-app.md)
- [Next.js guide](./nextjs.md)
- [Email / SMTP](./email-smtp.md)
- [Production validation](./production-validation.md)

```bash
dockyard update --check
sudo dockyard update
sudo systemctl restart dockyard
```

If an update fails after restart, restore the previous binary:

```bash
sudo dockyard update --rollback
sudo systemctl restart dockyard
```

Pre-update SQLite backups land under `/var/lib/dockyard/backups/`.
