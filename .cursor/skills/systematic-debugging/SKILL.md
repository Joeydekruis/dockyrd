---
name: systematic-debugging
description: Reproduce → minimize → hypothesize → fix → regression test. Use for S bugs and generator blockers.
---

# Systematic Debugging

Use for hard bugs, flaky tests, or unclear failures.

## Loop

1. **Reproduce** - consistent steps; capture error
2. **Minimize** - smallest failing case
3. **Hypothesize** - one theory at a time
4. **Instrument** - logs/tests to confirm or reject
5. **Fix** - minimal change
6. **Regression test** - lock behavior in tests

## Harness

- Most bugs/finetunes are **Direct** track: reproduce → fix on `development` → regression test. No Linear/planner unless it is a **big** adjustment (then Feature track).
- If inside a Feature generator round, document root cause in `generator-response`

## When to escalate

- 3 generator rounds exhausted → human triage per orchestration rule
