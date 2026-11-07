# Domains and DNS

dockyrd terminates HTTPS at the edge (**dy-proxy** / Caddy). You point DNS at your server; dockyrd provisions certificates automatically when ACME is configured.

## Dashboard hostname

During setup, set the panel domain (or use `IP:8000` for labs). For production, create an **A** or **AAAA** record to the server’s public IP and open ports **80**, **443**, and **8000** (panel) in the firewall.

## Application domains

Each service can have one or more hostnames:

1. Open the service → **Domains** → add `app.example.com`.
2. Create DNS **A/AAAA** to the same server IP (or a wildcard if you configured one at install).
3. Wait for DNS propagation; Caddy obtains certificates on first successful request.

### Wildcard (optional)

If the instance wildcard is `*.apps.example.com`, subdomains like `myapp.apps.example.com` can be assigned without per-host ACME challenges when DNS points correctly.

### sslip.io (labs)

For quick tests without DNS, dockyrd supports sslip.io-style hostnames that encode your IP. Prefer real domains for production.

## Checks

```bash
dockyrd doctor
```

Verify system clock (HTTPS) and that **Caddy admin socket** is present while `dockyrd serve` runs.

## Database and internal services

Do **not** publish Postgres or Redis ports to the public internet. Apps connect over the Docker project network using internal hostnames dockyrd injects into variables.

See [install](./install.md) for firewall notes when `ufw` is enabled.
