# Troubleshooting

## `dockyrd doctor`

Run on the server:

```bash
dockyrd doctor
```

| Check | Typical fix |
|-------|-------------|
| Docker Engine | `sudo systemctl start docker`; add user to `docker` group on dev machines |
| Disk space | Free space on `/var/lib/dockyrd` or lower `dockyrd_MIN_FREE_DISK_PERCENT` for dev only |
| Master key / DB | Run `sudo dockyrd init` on a fresh host; do not delete `secret.key` |
| Caddy admin socket | Start `dockyrd serve` so `dy-proxy` loads; workloads keep running in Docker |

Paste doctor output when filing bugs.

## `ERR_CONNECTION_REFUSED` right after install

If the installer printed **`Missing unit file`**, the binary and database were created but **systemd never registered** — nothing listens on port **8000**.

This happens when `curl | sh` cannot fetch `packaging/dockyrd.service` from GitHub (older releases only published `packaging/dockyard.service`). The install script now falls back to that file and pulls slice units over the network when not run from a git checkout.

**On the server until you re-run an updated installer:**

```bash
BASE=https://raw.githubusercontent.com/Joeydekruis/dockyrd/development/packaging
curl -fsSL "$BASE/dy.slice" -o /etc/systemd/system/dy.slice
curl -fsSL "$BASE/dy-system.slice" -o /etc/systemd/system/dy-system.slice
curl -fsSL "$BASE/dy-apps.slice" -o /etc/systemd/system/dy-apps.slice
curl -fsSL "$BASE/dy-build.slice" -o /etc/systemd/system/dy-build.slice
curl -fsSL "$BASE/dockyard.service" -o /etc/systemd/system/dockyrd.service
systemctl daemon-reload
systemctl enable --now dockyrd
curl -sS http://127.0.0.1:8000/api/v1/health
sudo dockyrd setup-token
```

Use **`dockyard.service`** in the URL above — not `dockyrd.service` — when the raw `dockyrd.service` path still returns 404.

If `curl` to `:8000` fails immediately after `enable --now`, wait a second and retry; then check `journalctl -u dockyrd -n 40`.

## `dockyrd.service` fails with `status=226/NAMESPACE`

systemd could not build the service sandbox — the panel never binds to port **8000**.

On **fresh** installs this often means a path in `ReadWritePaths=` does not exist (for example legacy `/var/lib/dockyard` before a rename). Check logs:

```bash
journalctl -u dockyrd -n 40 --no-pager
```

**Quick fix** (creates optional legacy dirs) or reinstall after upgrading to a release that marks them optional in the unit:

```bash
sudo mkdir -p /var/lib/dockyard /etc/dockyard
sudo systemctl restart dockyrd
curl -sS http://127.0.0.1:8000/api/v1/health
```

Or run the panel once without systemd to confirm the binary works: `sudo /usr/local/bin/dockyrd serve` (Ctrl+C to stop).

## Login / 2FA

- **Forgot password:** use **Forgot password?** on the sign-in page (link-based reset until SMTP is configured).
- **Locked out:** on the server, `dockyrd admin reset-password you@example.com` and optionally `dockyrd admin disable-2fa you@example.com`.

## Builds fail with disk errors

Builds refuse to start when free disk on the data volume is below 10% (configurable via `dockyrd_MIN_FREE_DISK_PERCENT`). Prune Docker from **Analytics → Docker cleanup** or free space on the host.

## `mkdir /root/.docker: read-only file system`

The panel service runs with `ProtectHome=true`, so Docker CLI must not write under `/root`. Current releases set `DOCKER_CONFIG` and `BUILDX_CONFIG` under `/var/lib/dockyrd`.

On an older install, add to `/etc/dockyrd/dockyrd.env`:

```bash
DOCKER_CONFIG=/var/lib/dockyrd/docker-cli
BUILDX_CONFIG=/var/lib/dockyrd/buildx
```

Then:

```bash
sudo mkdir -p /var/lib/dockyrd/docker-cli /var/lib/dockyrd/buildx
sudo systemctl restart dockyrd
```

## Setup wizard / setup token again after update

If you already completed setup and the panel suddenly asks for a **setup token**, the control plane is probably reading a **new empty data directory** after the **dockyard → dockyrd** rename.

| Era | Default data | Default config |
|-----|--------------|----------------|
| Before 0.7.1 | `/var/lib/dockyard` | `/etc/dockyard` |
| 0.7.1+ | `/var/lib/dockyrd` | `/etc/dockyrd` |

Your projects and login still live in the old paths until you point the service at them.

**Quick fix (on the server):**

Older installs use **`dockyard.service`** and **`/etc/dockyard/dockyard.env`**. Newer ones use **`dockyrd.service`** and **`/etc/dockyrd/dockyrd.env`**. See which you have:

```bash
systemctl list-unit-files '*dock*'
```

Put the data-dir overrides in the env file your unit already loads (create the file if missing):

```bash
# Pre-rename install (common after self-update from 0.7.0):
sudo mkdir -p /etc/dockyard
sudo tee -a /etc/dockyard/dockyard.env <<'EOF'
# Set both spellings — the unit only loads this file, not /etc/dockyrd/dockyrd.env
DOCKYARD_DATA_DIR=/var/lib/dockyard
DOCKYARD_CONFIG_DIR=/etc/dockyard
dockyrd_DATA_DIR=/var/lib/dockyard
dockyrd_CONFIG_DIR=/etc/dockyard
EOF
sudo systemctl restart dockyard.service

# Fresh install with dockyrd unit:
sudo mkdir -p /etc/dockyrd
sudo tee -a /etc/dockyrd/dockyrd.env <<'EOF'
dockyrd_DATA_DIR=/var/lib/dockyard
dockyrd_CONFIG_DIR=/etc/dockyard
EOF
sudo systemctl restart dockyrd.service
```

If the unit references `/usr/local/bin/dockyard` but only `dockyrd` exists, symlink **only when the target binary exists**:  
`sudo ln -sf /usr/local/bin/dockyrd /usr/local/bin/dockyard`  
A symlink to a missing `dockyrd` causes `203/EXEC` and `ERR_CONNECTION_REFUSED`.

### Panel down: `203/EXEC` / missing `/usr/local/bin/dockyard`

Reinstall the release binary (pick `arm64` or `amd64` from `uname -m`):

```bash
sudo systemctl stop dockyard.service
sudo rm -f /usr/local/bin/dockyard /usr/local/bin/dockyrd
VER=0.7.3
arch=$(uname -m); case "$arch" in aarch64|arm64) arch=arm64;; *) arch=amd64;; esac
tmp=$(mktemp -d)
curl -sfL -o "$tmp/p.tgz" "https://github.com/Joeydekruis/dockyrd/releases/download/v${VER}/dockyard_${VER}_linux_${arch}.tar.gz" \
  || curl -sfL -o "$tmp/p.tgz" "https://github.com/Joeydekruis/dockyrd/releases/download/v${VER}/dockyrd_${VER}_linux_${arch}.tar.gz"
tar -xzf "$tmp/p.tgz" -C "$tmp"
if [ -f "$tmp/dockyard" ]; then install -m 755 "$tmp/dockyard" /usr/local/bin/dockyard; fi
if [ -f "$tmp/dockyrd" ]; then install -m 755 "$tmp/dockyrd" /usr/local/bin/dockyrd; fi
[ -x /usr/local/bin/dockyard ] && [ ! -e /usr/local/bin/dockyrd ] && ln -sf /usr/local/bin/dockyard /usr/local/bin/dockyrd
[ -x /usr/local/bin/dockyrd ] && [ ! -e /usr/local/bin/dockyard ] && ln -sf /usr/local/bin/dockyrd /usr/local/bin/dockyard
rm -rf "$tmp"
sudo systemctl start dockyard.service
```

Then sign in as before (do **not** complete setup again unless `dockyrd setup-token` is intentional on an empty instance).

Newer releases auto-prefer the legacy database when `/var/lib/dockyrd/control.db` is empty but `/var/lib/dockyard/control.db` has users. You can remove the env overrides after upgrading once that build is installed.

## `ERR_QUIC_PROTOCOL_ERROR` in the browser

Chrome often shows this on **HTTPS** (HTTP/3) when the edge proxy and panel disagree after a restart, while the panel may still be fine on the direct port.

1. Try **`http://<server-ip>:8000`** (not `https://`) and sign in there.
2. On the server, confirm the API:

```bash
curl -sS http://127.0.0.1:8000/api/v1/setup/status
curl -sS http://127.0.0.1:8000/api/v1/health
journalctl -u dockyard.service -n 40 --no-pager
```

`{"complete":true}` means your real database is back — use **sign-in**, not setup.

3. If only the panel domain breaks, reload Caddy: `sudo systemctl restart docker` (if dy-proxy runs there) or restart the edge stack from **Settings**. In Chrome you can also test with QUIC off: `chrome://flags/#enable-quic` → Disabled.

Your unit name is **`dockyard.service`** until you install `dockyrd.service`; there is no `dockyrd` binary required if only `/usr/local/bin/dockyard` exists (self-update keeps that path).

## Proxy / HTTPS

dockyrd configures Caddy via a unix admin socket. Panel restarts do not stop user containers. If HTTPS fails, verify DNS points at the server and system time is correct (`doctor` clock check).

## Logs

```bash
journalctl -u dockyrd -f
docker logs dy-proxy
```

## More help

Product architecture and API details: [`docs/BRIEFING.md`](../BRIEFING.md).
