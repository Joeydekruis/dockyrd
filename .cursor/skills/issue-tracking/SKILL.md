---
name: issue-tracking
description: >-
  GitHub Issues = backlog; Cursor builds on development. Per-task commit+push
  and evidence handoff. Use when grooming backlog, sprints, or delivery.
---

# Issue tracking workflow

## Mental model

| Artefact | Role |
|--|--|
| **GitHub issue `#n`** | Backlog — features, bugs, milestones |
| **Cursor** | Implementation; optional local specs under `docs/project/issues/<n>/` |
| **`docs/project/project.md`** | Product truth |
| **`docs/project/github-workflow.md`** | How issues, labels, and handoff work |
| **`development`** | Day-to-day commits; push when shipping issue work |

**No Linear.** One tracker only.

## Tracks

1. **Feature** — GitHub issue + harness; keep issue body linked to spec.
2. **Direct** — chat fixes; no issue required; no mandatory commit unless asked.

## Per GitHub issue sequence (Feature track)

```text
Pick up → implement → tests (capture output)
Commit + push development (Fixes #n or Refs #n)
Post evidence comment on issue (Ready for review)
Label status/ready-for-review — do not close issue
```

See [github-handoff](../../rules/github-handoff.mdc).

## Sprint pickup

**“Pick up sprint X”** → all open issues in that GitHub **milestone** → complete each with the sequence above.

## Issue folders

- `docs/project/issues/<n>/spec.md` + `rubric.md` — Feature track
- `docs/project/feedback/<n>/` — reviewer iterations

## Sizing (Feature only)

Agent decides micro-S / S / M / L / XL without asking.

## Close / handoff (Feature)

1. Run relevant tests — paste output in issue comment.
2. Commit + push with issue reference.
3. Post handoff comment; add `status/ready-for-review`.
4. Maintainer closes after verification.

Changelog + patch bump per [changelog](../../rules/changelog.mdc).

## Anti-patterns

- Opening duplicate trackers (Linear, etc.).
- Closing issues before maintainer review.
- Cherry-picking part of a sprint when told to pick up the whole milestone.
