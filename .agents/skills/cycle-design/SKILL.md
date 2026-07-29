---
name: cycle-design
description: Choose the smallest maintainable internal structure needed to implement one accepted contract. Trigger when ownership, data flow, lifecycle, module placement, compatibility, or testability is genuinely unclear. Do not create speculative architecture or redesign unrelated areas.
---

# Minimal Cycle Design

Follow `AGENTS.md`. Design exists to enable the selected behavior, not to become
a parallel deliverable.

## Entry Gate

Use this skill only after the contract is ready.

Skip with `DESIGN NOT REQUIRED` when the change has an obvious location and
follows an established pattern.

## Procedure

1. Inspect existing structure, conventions, dependencies, and nearby behavior.
2. Identify the minimum production path from input to observable result.
3. Assign ownership for validation, policy, state, effects, and output.
4. Preserve existing public interfaces unless the accepted contract changes
   them.
5. Identify the test attachment point at the real boundary.
6. Identify failure, cleanup, concurrency, transaction, and lifecycle concerns
   that materially affect the slice.
7. Prefer existing modules and patterns.
8. Introduce an abstraction only when the selected behavior requires it or at
   least two real current uses justify it.
9. Keep test seams inaccessible from production paths.
10. Record a short design note in the active plan; create durable architecture
    documentation only for a cross-cutting decision that future work must obey.

## Reject

- package or directory growth without behavior;
- frameworks or plugin systems for hypothetical future needs;
- generic repositories, factories, managers, services, or adapters with one
  trivial use;
- interfaces introduced only to mock implementation details;
- broad refactoring hidden inside the slice;
- diagrams or ADRs replacing code and tests.

## Result

Return one:

```text
DESIGN READY
DESIGN NOT REQUIRED
DESIGN BLOCKED
```

Use this shape:

```text
Phase: design
Result:
Evidence:
- chosen production path and test boundary
- concrete simplification or reason no design is needed
Changed:
- architecture/ADR files, or none
Next: red | baseline-green | blocked
Blocker: none | exact design conflict
```
