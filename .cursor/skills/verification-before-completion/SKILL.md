---
name: verification-before-completion
description: Run relevant setup.md gates and capture output before GitHub issue handoff.
---

# Verification before completion

**Capture output** — paste it in the GitHub handoff comment (not “tests green”).

## Before merge to `main` or GitHub release

Run the **CI parity** matrix in [push-to-main](../push-to-main/SKILL.md) §2 (especially `./scripts/verify-install.sh` when `scripts/install.sh` or install verify changed). Capture output for the ship report.

## Before Feature / GitHub issue handoff

- Relevant commands from `docs/project/setup.md` for touched surfaces
- Changes **committed and pushed** to `origin/development` with `Refs #n` — **never** hand off uncommitted work
- GitHub: evidence comment (**Ready for review — #n**) with commit SHA, test output, preview steps for UI — see [github-handoff](../../rules/github-handoff.mdc). Do **not** close the issue.
