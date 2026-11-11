# User documentation

End-user guides for installing and running **dockyrd** (CLI/binary) and the **Dockyrd** dashboard.

## Public site

The primary, SEO-friendly docs live at **[https://dockyrd.io/docs](https://dockyrd.io/docs)**. Source in this repo:

- **`site/source.html`** — marketing + doc articles (run `make site` to generate `site/docs/*.html`)
- **`docs/user/*.md`** — markdown mirrors for contributors; keep install/CLI paths aligned with the product

When you change onboarding, install, CLI, or user-facing behavior, update **both** the public site source and the relevant file here. See [`.cursor/rules/public-site.mdc`](../../.cursor/rules/public-site.mdc).

## Install (current)

```bash
curl -fsSL https://raw.githubusercontent.com/Joeydekruis/dockyrd/development/scripts/install.sh | sudo sh
```

Until `https://get.dockyrd.dev` is live, this GitHub raw URL is the official one-liner (also shown on dockyrd.io).

## Index

| Topic | Markdown | Public site |
|-------|----------|-------------|
| Install | [install.md](./install.md) | [/docs/installation.html](https://dockyrd.io/docs/installation.html) |
| First app | [first-app.md](./first-app.md) | [/docs/quick-start.html](https://dockyrd.io/docs/quick-start.html) |
| Next.js | [nextjs.md](./nextjs.md) | [/docs/nextjs.html](https://dockyrd.io/docs/nextjs.html) |
| Domains & DNS | [domains-dns.md](./domains-dns.md) | [/docs/domains.html](https://dockyrd.io/docs/domains.html) |
| Email / SMTP | [email-smtp.md](./email-smtp.md) | (panel settings; link from troubleshooting) |
| Troubleshooting | [troubleshooting.md](./troubleshooting.md) | [/docs/troubleshooting.html](https://dockyrd.io/docs/troubleshooting.html) |
| AI assistants (MCP) | [mcp.md](./mcp.md) | [/docs/mcp.html](https://dockyrd.io/docs/mcp.html) |
| Production checks | [production-validation.md](./production-validation.md) | — |
