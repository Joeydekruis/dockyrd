---
github: https://github.com/Joeydekruis/dockyrd/issues/1
title: "[foundation] Legal docs, README, and CI workflow"
milestone: M0 · Sprint 1
size: S
---

## Goal

Publish open-source legal baseline and CI so `development` can receive M0 work with clear contribution and security paths.

## Acceptance criteria

- [ ] `LICENSE` (Apache-2.0), `NOTICE`, `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`
- [ ] `README.md`: product one-liner, install curl placeholder, link `docs/project/setup.md`, branch `development`
- [ ] `.github/workflows/ci.yml`: Go vet/test; web lint/build when `web/` present; Radix check when script exists
- [ ] No secrets in repo

## References

- `docs/BRIEFING.md` §14.3, §16, §19
