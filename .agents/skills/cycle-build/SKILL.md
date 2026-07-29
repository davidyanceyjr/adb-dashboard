---
name: cycle-build
description: Implement the smallest complete production change for one cycle after a confirmed contract and red test or approved baseline. Trigger on the build phase, focused bug fix, or implementation repair packet. Do not claim functional verification.
---

# Cycle Build

Follow `AGENTS.md`. Own production implementation only for the selected slice.

## Entry Gate

Require one of:

```text
RED CONFIRMED
BASELINE GREEN
```

A documented exception is allowed only when a meaningful red test is genuinely
impractical. The active state must record the reason and alternative evidence.

Do not start from `RED INVALID`, `BASELINE FAILED`, or an ambiguous contract.

## Procedure

1. Read the acceptance criteria, expected red evidence, and repair packet when
   present.
2. Inspect the existing production path before editing.
3. Modify the fewest files and layers needed for one complete vertical slice.
4. Preserve behavior outside the slice.
5. Reuse established patterns and dependencies.
6. Implement validation, errors, side effects, cleanup, compatibility, and
   safety required by the contract.
7. Keep test-only seams unreachable in production.
8. Remove placeholders or unavailable behavior only when the real replacement
   is implemented in the same slice.
9. Run the narrowest relevant check while developing.
10. Stop editing when the selected production path exists; return immediately
    to green testing.

## Forbidden

- hard-coded or fabricated success;
- TODOs or stubs standing in for the requested behavior;
- broad architecture or dependency changes unrelated to the slice;
- weakening validation, security, safety, or test assertions;
- silent fallback success;
- unrelated cleanup or formatting churn;
- documenting the feature as verified before test evidence exists.

## Result

Return one:

```text
BUILD APPLIED
BUILD PARTIAL
BUILD BLOCKED
```

`BUILD APPLIED` means production code was changed and is ready for green testing.
It does not mean the behavior is verified.

Use this shape:

```text
Phase: build
Result:
Evidence:
- production path implemented
- focused development command and result, if run
Changed:
- material production files
Next: green | contract | design | blocked
Blocker: none | exact blocker
```
