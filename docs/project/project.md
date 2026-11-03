# Dockyard — project

Self-hosted, open-source deploy platform: one Go binary, Docker workloads, Caddy edge, SQLite control plane, embedded Next.js dashboard.

| Doc | Role |
|--|--|
| [`docs/BRIEFING.md`](../BRIEFING.md) | Full technical briefing (architecture, API, milestones) |
| [`docs/design-index.json`](../design-index.json) | UI navigation, tokens, shared blocks |
| [`docs/dockyrd-design.html`](../dockyrd-design.html) | Design export (artboards) |
| [`roadmap.md`](roadmap.md) | Milestones M0–M5 |
| [`setup.md`](setup.md) | Local verify commands |

## Lifecycle stage

**MVP (pre-code → M0)** — first public repo; optimize for install simplicity, stability, and small footprint (briefing §1.3).

## Delivery

- **Branch:** `development` (day-to-day); `main` only when Joey asks (`push-to-main` skill).
- **Features:** Linear issue + harness (`docs/project/issues/<id>/`) when scope needs PM tracking.
- **Direct fixes:** implement on `development` without Linear ceremony.
- **Commits:** only when Joey asks in Cursor chat (unless a Linear handoff batch is explicitly requested).

## Naming

- Product/binary/docs: **Dockyard**
- Dashboard UI / logo: **Dockyrd** (see `.cursor/rules/product-naming.mdc`)

## Quality bar (summary)

Briefing §19 Definition of Done + stack rules in `.cursor/rules/dockyard-*.mdc`. No `@radix-ui/*` in `web/`. Panel must not depend on external font CDNs.
