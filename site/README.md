# Deploying `site/` on Dockyard

Point a **static site** service at this directory’s build output (repository root: **`site/`** after `make site`).

## Build

```bash
make site
# or: node scripts/build-public-site.mjs
```

Requires `web/node_modules/geist` for self-hosted fonts (`pnpm --dir web install`).

## Suggested static site settings

| Setting | Value |
|---------|--------|
| Output directory | `.` (requires `index.html` in that folder — Plain HTML preset) |
| Build command | `make site` (from monorepo root; set working directory accordingly) |
| Domain | `dockyrd.io` (and `www` redirect if desired) |

Publish the **generated** files: `index.html`, `docs/`, `assets/`, `robots.txt`, `sitemap.xml`.

**Important:** the **`assets/`** folder (CSS, fonts, JS) must ship with the HTML. If only `index.html` is published, the page will look unstyled. In Dockyard, point the static site at the **`site/`** directory (or copy the whole folder to the publish root).

## Paths

Asset URLs are **relative** (`assets/…` from home, `../assets/…` from docs) so local preview and domain-root hosting both work.
