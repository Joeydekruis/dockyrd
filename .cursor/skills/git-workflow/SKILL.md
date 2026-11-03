---
name: git-workflow
description: >-
  Always work on development; commit and push after each Linear task finishes.
  Conventional Commits with Refs DY-nn. Main ship uses push-to-main skill.
  Apply when committing, pushing development, or syncing with main.
---

# Git workflow (`development` → `main`)

Long-lived **`development`** branch. Agents implement there, then **commit and push after each Linear task** completes. Shipping to **`main`** is a separate, human-requested flow ([push-to-main](../push-to-main/SKILL.md)).

## Model

```
main              ← release; merge only via push-to-main (full build first)
development       ← daily integration; all agent commits land here
```

1. **`development`** is the only branch for implementation.
2. Do **not** create `feature/*`, `fix/*`, or other per-issue branches.
3. **Linear task finish:** run tests → changelog if applicable → **commit + push** `origin/development` with `Refs DY-nn` → Linear handoff ([linear-handoff](../../rules/linear-handoff.mdc)).
4. **Direct track (no Linear issue):** no mandatory commit/push unless the human asks or a batch cycle closes.
5. Do **not** commit/push **`main`** during normal work - invoke **push-to-main** when the human asks to ship.
6. **Session start / before edits:** if `git rev-parse --abbrev-ref HEAD` ≠ `development`, run `git checkout development` immediately (keep uncommitted work). Delete leftover `feature/*` branches after moving work.
7. Cursor shell hook `.cursor/hooks/block-feature-branches.sh` denies creating/switching to per-issue branches - do not work around it.

## Commits - Conventional Commits

```text
<type>(<optional-scope>): <imperative subject, ~50 chars>

Refs DY-12
https://linear.app/<team>/issue/DY-12/...
```

Common types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `ci`, `revert`.

Body explains **why** when needed. **Linear tasks:** title or body must reference the issue. Align with **`issue-tracking`**.

## Sprint work

One commit (+ push + Linear handoff) **per issue**, not one commit for the whole sprint.

## Staying current

- Prefer `git merge origin/main` into `development` when humans ask to sync, or when conflicts require it.
- Never force-push to `main` or shared `development` unless the human explicitly requests it.

## Pull requests

Optional. Prefer committing on `development`. If a PR is requested: open `development` → `main`. Title/body: What / Why / Testing; Feature work `Refs DY-nn`.

## Conflicts

1. Identify with `git status`.
2. Resolve markers; run local verification from `docs/project/setup.md`.
3. `git add` resolved paths; complete the merge commit.

## Anti-patterns

- Per-feature or per-fix branches.
- Leaving finished **Linear-task** work uncommitted / unpushed on `development`.
- Linear handoff before push (including “uncommitted on development”).
- Pushing to **`main`** without [push-to-main](../push-to-main/SKILL.md).
- **Secrets** committed.
- Huge catch-all commits spanning multiple unrelated Linear issues.
