# Dockyard — project

Self-hosted, open-source deploy platform: one Go binary, Docker workloads, Caddy edge, SQLite control plane, embedded Next.js dashboard.

| Doc | Role |
|--|--|
| [`docs/BRIEFING.md`](../BRIEFING.md) | Full technical briefing (architecture, API, milestones) |
| [`docs/design-index.json`](../design-index.json) | UI navigation, tokens, shared blocks |
| [`docs/dockyrd-design.html`](../dockyrd-design.html) | Design export (artboards) |
| [`roadmap.md`](roadmap.md) | Milestones M0–M5 |
| [`sprints.md`](sprints.md) | Sprint ↔ GitHub milestone mapping |
| [`github-workflow.md`](github-workflow.md) | Issues, labels, handoff (open source) |
| [`setup.md`](setup.md) | Local verify commands |

## Lifecycle stage

**MVP (pre-code → M0)** — first public repo; optimize for install simplicity, stability, and small footprint (briefing §1.3).

## Delivery

- **Branch:** `development` (day-to-day); `main` only when Joey asks (`push-to-main` skill).
- **Backlog:** [GitHub Issues](https://github.com/Joeydekruis/dockyrd/issues) only — no Linear.
- **Features:** GitHub issue `#n` + harness (`docs/project/issues/<n>/`) when scope needs tracking.
- **Direct fixes:** implement on `development` without opening an issue unless useful for the community.
- **Commits:** when Joey asks in Cursor chat, or when completing a GitHub issue handoff that requires push.

## Naming

- Product/binary/docs: **Dockyard**
- Dashboard UI / logo: **Dockyrd** (see `.cursor/rules/product-naming.mdc`)

## Quality bar (summary)

Briefing §19 Definition of Done + stack rules in `.cursor/rules/dockyard-*.mdc`. No `@radix-ui/*` in `web/`. Panel must not depend on external font CDNs.
