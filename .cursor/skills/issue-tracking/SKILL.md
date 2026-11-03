---
name: issue-tracking
description: >-
  Linear = PM for features; Cursor = Direct build/fix. Per-task commit+push
  and evidence handoff. Use when grooming backlog, sprints, or Linear delivery.
---

# Issue tracking workflow

## Mental model

| Artefact | Role |
|--|--|
| **Linear (`DY-nn`)** | **PM** - new features and big adjustments: status, priority, milestone |
| **Cursor** | **Tester/developer** - feedback, finetunes, bugs: implement without Linear ceremony |
| **`docs/project/project.md`** | Product truth |
| **`docs/project/roadmap.md`** | Phases, ordering (update on Feature close) |
| **`docs/project/issues/<id>/`** | Feature AC (planner) - Feature track only |
| **`development`** | All commits; **commit + push after each Linear task**; `main` via push-to-main |

## Tracks

1. **Feature** - create/update Linear when raised in Cursor; plan → generate → review; keep Linear current.
2. **Direct** - feedback / finetune / bugfix: no plans, no Linear updates, no mandatory commit unless asked.

## Per Linear task sequence (automatic)

```text
Pick up → In Progress
Implement → run tests (capture output)
Commit + push development (Refs DY-nn)
Post Linear evidence comment (Ready for Review - DY-nn)
Status → In Review (never Done)
```

See [linear-handoff](../../rules/linear-handoff.mdc). Do **not** wait for the human to ask.

## Sprint pickup

**“Pick up sprint X”** → find **all** open issues in that sprint → complete each with the sequence above. Subagents/subchats per issue OK; **no** single batch commit at sprint end.

## Issue ID

- Features: folder key = Linear ID (`DY-12`).
- Direct/bugs: no Linear required.
- Legacy numeric folders stay valid for in-flight work.

## Sizing (Feature only)

Agent **decides** micro-S / S / M / L / XL. Do not wait for confirm. Escalate only when product-ambiguous and blocking.

## Linear template (Feature)

Keep Linear short; full AC in local spec:

```markdown
## Spec reference

- **Project:** `docs/project/project.md`
- **Roadmap:** `docs/project/roadmap.md`
- **Issue spec:** `docs/project/issues/DY-NN/spec.md`
- **Rubric:** `docs/project/issues/DY-NN/rubric.md`

## Size

M - agent decided

## Acceptance criteria (summary)

- [ ] _Short bullets; full list in local spec_
```

Template: [`docs/project/templates/linear-issue.md`](../../docs/project/templates/linear-issue.md)

## Close / handoff (Feature)

When implementation + local reviewers are done:

1. Run relevant tests - **paste output** in handoff (not “tests green”).
2. **Commit + push** `development` with `Refs DY-nn`.
3. Post one evidence comment ([linear-handoff](../../rules/linear-handoff.mdc)); template: [`docs/project/templates/linear-handoff.md`](../../docs/project/templates/linear-handoff.md).
4. Set Linear status to **`In Review`** - **never `Done`**.
5. Update `roadmap.md` / `generator-state.md` to **In Review**, not Done.
6. Do not wait for PM - they (or Joey) move to Done after checking.

## Grooming checklist (Feature / planner)

- [ ] Roadmap subtasks map to Linear issues.
- [ ] Active Features have local spec/rubric when M+ (micro-S may stay Linear-body-light).
- [ ] Finished Features have **pushed commit SHA** + evidence comment + **In Review** (not silently Done or uncommitted).
- [ ] Direct work is **not** forced into Linear.

## Integration

- **git-workflow** - always `development`; `Refs DY-nn` on every Linear task commit.
- **orchestration.mdc** - Feature harness vs Direct implement.
- **issue-tracking.mdc** / **linear-handoff.mdc** - always-applied policy.

## Anti-patterns

- Dual-tracking Linear + GitHub for the same item.
- Linear ceremony for Direct feedback/bugs.
- Handoff without commit/push (“uncommitted on development”).
- Claims without test output or preview URL for visual work.
- Cherry-picking a few sprint issues when told to pick up the whole sprint.
- Asking the human to confirm size / process / track / handoff.
- Per-feature git branches.
- Setting Linear to **Done** from Cursor (false-completion risk).
