---
name: git-workflow
description: >-
  Work on development; commit and push after each GitHub issue handoff.
  Conventional Commits with Fixes #n. Main ship uses push-to-main skill.
---

# Git workflow (Dockyard)

## Branch policy (HARD)

- **Only `development`** for implementation. Never `feature/*` / `fix/*`.
- Before edits: `git rev-parse --abbrev-ref HEAD` → must be `development`.
- **`main`:** only via [push-to-main](../push-to-main/SKILL.md) when Joey asks.

## Commits

- Conventional Commits; link issues: `Fixes #12` or `Refs #12` in body.
- User rule: commit only when Joey asks **unless** completing a Feature issue handoff that requires push.

## After a GitHub issue (Feature track)

1. Tests from `docs/project/setup.md` (capture output).
2. `CHANGELOG.md` + patch bump when the issue cycle closes.
3. Commit + push `origin/development`.
4. Post [github-handoff](../../rules/github-handoff.mdc) comment on the issue.

## Sprint work

One commit (+ push + issue comment) **per issue**, not one commit for the whole sprint.

## Direct track

No mandatory push unless Joey asks or a release batch closes.
