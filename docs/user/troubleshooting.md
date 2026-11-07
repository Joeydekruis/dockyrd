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

## Login / 2FA

- **Forgot password:** use **Forgot password?** on the sign-in page (link-based reset until SMTP is configured).
- **Locked out:** on the server, `dockyrd admin reset-password you@example.com` and optionally `dockyrd admin disable-2fa you@example.com`.

## Builds fail with disk errors

Builds refuse to start when free disk on the data volume is below 10% (configurable via `dockyrd_MIN_FREE_DISK_PERCENT`). Prune Docker from **Analytics → Docker cleanup** or free space on the host.

## Proxy / HTTPS

dockyrd configures Caddy via a unix admin socket. Panel restarts do not stop user containers. If HTTPS fails, verify DNS points at the server and system time is correct (`doctor` clock check).

## Logs

```bash
journalctl -u dockyrd -f
docker logs dy-proxy
```

## More help

Product architecture and API details: [`docs/BRIEFING.md`](../BRIEFING.md).
