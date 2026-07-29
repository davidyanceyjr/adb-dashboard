---
name: cycle-handoff
description: Pause, resume, or transfer one measurable implementation cycle while preserving exact repository state, evidence, blockers, and next phase. Trigger on pause, resume, handoff, interruption, blocked cycle, or session transfer. Do not use as a substitute for completing available work.
---

# Cycle Handoff And Resume

Follow `AGENTS.md`. A handoff is a factual checkpoint, not a progress narrative
or completion claim.

## Pause

Before pausing:

1. Inspect version-control state.
2. Inspect the active cycle state.
3. Record changed and uncommitted files.
4. Record exact commands run and results.
5. Record passing, failing, blocked, skipped, and not-run checks.
6. Record the exact current phase and next valid phase.
7. Record files or user changes that must not be touched.
8. Update `.codex/plans/current.md`.

Do not create another handoff file unless repository policy explicitly requires
one.

## Resume

Never trust a handoff blindly.

1. Compare the active state with current branch, status, diff, and recent
   history.
2. Confirm referenced files and commands still exist.
3. Confirm prior evidence applies to the current working state.
4. Mark stale evidence invalid when production, tests, configuration,
   dependencies, or generated files changed afterward.
5. Identify the first valid next phase.
6. Resume through `$implementation-cycle`.

## Handoff State

Use this compact section in the active plan:

```markdown
## Pause State

- Current phase:
- Last valid result:
- Changed files:
- Commands run:
- Passing:
- Failing:
- Not run:
- Blocker:
- Next phase:
- Do not touch:
```

## Result

Return one:

```text
HANDOFF READY
RESUME READY
RESUME BLOCKED
```

Use this shape:

```text
Phase: handoff | resume
Result:
Evidence:
- repository state inspected
- evidence validated or invalidated
Changed:
- active state file, or none
Next:
Blocker: none | exact mismatch or missing context
```
