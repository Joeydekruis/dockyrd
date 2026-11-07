---
github: https://github.com/Joeydekruis/dockyrd/issues/71
---

## Done bar

- AC in `spec.md` met wireframe structuur op Instance + Domains (1360; spot-check 390).
- Go tests: platform hostname with/without verified wildcard; verify endpoint or service unit tests.
- `make test`, `make build`, `pnpm --dir web exec tsc --noEmit` green.

## Reviewers

- `api-reviewer` — instance verify + domain create contracts
- `security-reviewer` — DNS probe / host validation
- `frontend-evaluator` — vs `assets/domains-urls-wireframe.html` + design-direction states
- Skip `code-reviewer` if specialists pass
