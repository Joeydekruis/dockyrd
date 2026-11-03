---
name: verification-before-completion
description: Run relevant setup.md gates and rubric checks before claiming done.
---

# Verification Before Completion

No "done" without evidence. Prefer **relevant** gates over running everything.

## Before claiming done

From `docs/project/setup.md`, run what the **change touched**:

- Direct / tiny fix: focused tests or the single relevant command
- Feature / broader change: typecheck, lint, test, build as configured
- e2e / db migrate only when those surfaces changed

Do **not** re-run the full matrix on every one-line polish.

**Capture output** - you will paste it in the Linear handoff comment (not “tests green”).

## Before Feature / Linear task handoff

- Required reviewer files show PASS in `docs/project/feedback/<id>/` (when reviewers ran)
- No open CRITICAL/HIGH blockers
- `CHANGELOG.md` has a dated `## [x.y.z] - YYYY-MM-DD` for this task when applicable; `package.json` version matches (patch +0.0.1)
- Changes **committed and pushed** to `origin/development` with `Refs DY-nn` - **never** hand off uncommitted work
- Linear: evidence comment posted (**Ready for Review - DY-nn**) with commit SHA, test output, preview URL/recording for UI - then status **In Review** ([linear-handoff](../../rules/linear-handoff.mdc)) - never **Done**

## Generator

List commands actually run (with output summary) in `generator-response-i<i>.md` and reviewer handoff.
