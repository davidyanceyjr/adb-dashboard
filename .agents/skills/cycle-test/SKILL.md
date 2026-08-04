---
name: cycle-test
description: Create and run measurable red, baseline-green, green, real-path, negative-path, and broader verification for one selected slice or Test gate. Trigger on failing acceptance tests, regression reproduction, focused verification, real-path exercise, or cycle evidence. Do not repair production code.
---

# Cycle Testing And Evidence

Follow `AGENTS.md`. Own test evidence. Do not edit production code.

## Modes

Choose one from cycle state:

```text
red
baseline-green
green
verify
```

## Red Mode

Use for features and bug fixes.

1. Map each selected acceptance criterion to a focused test.
2. Prefer the real public boundary or the highest deterministic production
   boundary.
3. Add the smallest test that proves the behavior.
4. Run the narrowest command.
5. Confirm failure occurs for the intended missing or defective behavior.
6. Reject compile-only, fixture-only, mock-only, or unrelated failure.

Return:

```text
RED CONFIRMED
RED INVALID
VERIFICATION BLOCKED
```

`RED INVALID` must state whether the contract, test, environment, or existing
implementation invalidated the expected red result.

## Baseline-Green Mode

Use for behavior-preserving refactors.

1. Identify tests that protect the behavior being preserved.
2. Add missing characterization tests only when needed.
3. Run focused and applicable broader checks before production edits.

Return:

```text
BASELINE GREEN
BASELINE FAILED
VERIFICATION BLOCKED
```

Do not refactor from a failing unexplained baseline.

## Green Mode

1. Run the focused test first.
2. Exercise the real production path using the documented command, API,
   interface, UI, job, protocol, or procedure.
3. Verify applicable negative paths and side effects.
4. Run broader regression checks discovered from repository conventions.
5. Record every command, result, relevant output, skipped check, and environment
   assumption.
6. Classify failures as product, test, environment, pre-existing, or unknown.

Return:

```text
GREEN VERIFIED
GREEN FAILED
VERIFICATION BLOCKED
```

A build, lint pass, schema check, mock assertion, or unit test alone does not
prove an external behavior unless that interface is itself the selected product
boundary.

## Verify Mode

Use after repairs or before review to rerun every gate invalidated by later
changes.

Do not broaden testing merely to create activity. Run focused evidence plus the
applicable repository gate.

## Repair Packet

On `GREEN FAILED`, return:

```text
Phase: green
Result: GREEN FAILED
Evidence:
- failed command
- expected result
- observed result
- boundary reached
- failure classification
Changed:
- test files changed, or none
Next: build | contract | test
Blocker: none | exact environment blocker
```

Do not repair production code under this skill.

## Success Packet

```text
Phase: red | baseline-green | green | verify
Result:
Evidence:
- exact commands and exit results
- real-path observation when applicable
- acceptance criteria proven
Changed:
- test files, or none
Next:
Blocker: none | exact blocker
```
