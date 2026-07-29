---
name: cycle-contract
description: Define or confirm one bounded implementation contract from a specification and roadmap slice, including acceptance criteria, errors, side effects, compatibility, and observable test boundary. Trigger during contract, requirements, acceptance-criteria, or slice-readiness work. Do not implement production code.
---

# Cycle Contract

Follow `AGENTS.md`. Own intended observable behavior for one selected slice.
Do not implement production code.

## Accepted Inputs

- selected roadmap slice or bounded user request;
- accepted specification and architecture constraints;
- current behavior and tests;
- active cycle state;
- known compatibility, safety, security, and data constraints.

## Procedure

1. Identify the actor or caller and the concrete outcome.
2. Identify the public or internal interface being changed.
3. Define accepted input and preconditions.
4. Define successful output, status, side effects, ordering, and lifecycle.
5. Define errors, negative paths, cancellation, timeout, cleanup, and partial
   failure where applicable.
6. Define compatibility and migration behavior.
7. Define safety, authorization, privacy, and data-integrity requirements where
   applicable.
8. Assign or preserve stable acceptance-criterion IDs.
9. Select the highest practical deterministic acceptance boundary.
10. State what evidence would fail before implementation and pass afterward.
11. Name explicit in-scope and out-of-scope behavior.
12. Update the accepted specification only when the contract is missing,
    ambiguous, contradictory, or intentionally changing.

## Contract Quality

A contract is ready only when an engineer can write a meaningful failing test
without inventing:

- input shape;
- output or response shape;
- status or error behavior;
- side effects;
- ordering or lifecycle;
- compatibility expectations;
- safety or authorization behavior;
- the observable boundary.

Do not broaden a narrow slice into a complete product specification.

Do not weaken the intended behavior merely to fit current code.

Do not create implementation plans, package layouts, interfaces, or schemas
unless they are themselves part of the external contract.

## Result

Return one:

```text
CONTRACT READY
CONTRACT UPDATED
CONTRACT BLOCKED
```

Use this shape:

```text
Phase: contract
Result:
Evidence:
- specification anchors
- acceptance-criterion IDs
- acceptance boundary and expected red evidence
Changed:
- specification files, or none
Next: design | red | documentation | blocked
Blocker: none | exact conflict or missing decision
```

A documentation-only contract is still only `specified`.
