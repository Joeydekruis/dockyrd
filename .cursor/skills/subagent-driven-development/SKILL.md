---
name: subagent-driven-development
description: >-
  Parent orchestrates Feature track (Linear + plan → design → generate → review)
  or Direct track (implement now). Use for Feature closure and track selection.
---

# Subagent-Driven Development (Final)

Parent owns the loop. Subagents do **one role** per `Task`.

## Pick track first

| Track | When | Loop |
|-|--|--|
| **Feature** | New feature / big adjustment | Below |
| **Direct** | Feedback, finetune, bugfix | Research → implement on `development` → verify. No planner/Linear. |

## Feature loop

1. Create/update **Linear** (always when feature raised in Cursor)
2. Discuss gate only if product gray areas **block** - else assume and note
3. Size **decide** - state it; do not wait
4. `Task(planner)` → local spec + rubric; sync Linear
5. If [design-when](../../rules/design-when.mdc) → `Task(designer)` → HTML (no human approval)
6. `Task(generator)` → code on `development` + handoff (immediately after designer, or after planner if designer skipped)
7. `Task(reviewers)` **only if needed** - often skip for S/micro-S; when run, minimal + lean ([lean-loop](../../rules/lean-loop.mdc))
8. Consolidate; **auto** generator i+1 unless stop-early
9. Repeat generator if ≤3 rounds and not stop-early
10. `Task(doc-updater)` after PASS **if docs need it** (skip if deferred / none required)
11. Testing gate → capture output → changelog version + date → **commit + push `development` (`Refs DY-nn`)** → Linear evidence comment → **`In Review`**; never **Done**. Main ship only via `push-to-main` when asked.

## Sprint pickup

**“Pick up sprint X”** → all open issues in that sprint → each completes with its own test → commit/push → handoff. Parallel subagents OK.

## Rules

- **Efficient + good** - right-sized plans/checks; no over-plan or over-verify
- **No heavy code review by default** - skip Direct/micro-S/many S; light when M+
- Never nest full Feature closure inside one generator run
- Never let generator self-approve when a reviewer *did* run
- Design only per [design-when](../../rules/design-when.mdc); **never** pause for design approval
- Max 3 generator rounds → human triage
- Prefer skip or one specialist over stacking reviewers
- Direct track: no process pauses, no Linear theater
- Do not ask the human to confirm process, size, or design
- **Auto-continue**; pause only per lean-loop pause table
- Always work on **`development`** - no feature branches

Enforced by `.cursor/rules/orchestration.mdc`, `discuss-gate.mdc`, `design-when.mdc`, `reviewer-defaults.mdc`, `issue-tracking.mdc`, `agent-behavior.mdc`, `lean-loop.mdc`.
