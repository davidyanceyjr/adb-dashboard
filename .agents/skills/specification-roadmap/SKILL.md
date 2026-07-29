---
name: specification-roadmap
description: Create, improve, or audit an implementation-ready specification and a vertical-slice roadmap for the measurable implementation cycle. Trigger on "write specification", "audit specification", "create roadmap", "audit roadmap", "prepare implementation contract", or "make roadmap slices measurable". Do not implement production code.
---

# Specification And Roadmap Authoring

Follow `AGENTS.md`. This skill owns contract and roadmap authoring only. It must
not claim implementation.

## Read First

Read as applicable:

```text
docs/SPECIFICATION_GUIDE.md
docs/ROADMAP_GUIDE.md
docs/SPECIFICATION.template.md
docs/ROADMAP.template.md
docs/READINESS_CHECKLIST.md
```

Also inspect existing requirements, interfaces, source code, tests, decisions,
issues, and user instructions needed to ground the contract.

## Supported Operations

- `create specification`
- `improve specification`
- `audit specification`
- `create roadmap`
- `improve roadmap`
- `audit roadmap`
- `audit readiness`

## Specification Procedure

1. Identify purpose, scope, actors, and public boundaries.
2. Separate intended behavior from current implementation status.
3. Define stable capability, criterion, invariant, error, and data identifiers.
4. Define inputs, success, outputs, side effects, errors, lifecycle, security,
   data safety, and compatibility.
5. Define binary acceptance criteria.
6. Name the production acceptance boundary and expected real-path evidence.
7. Mark blocking questions explicitly.
8. Audit against `docs/SPECIFICATION_GUIDE.md`.

Return exactly one:

```text
SPECIFICATION ACCEPTED
SPECIFICATION BLOCKED
```

`SPECIFICATION ACCEPTED` means implementation-ready contract text, not software.

## Roadmap Procedure

1. Require an accepted specification.
2. Group criteria into the smallest useful vertical slices.
3. Use exactly one mode per slice.
4. Reference exact specification identifiers.
5. Define one observable result and one primary boundary.
6. Define red or baseline-green evidence.
7. Define focused verification and a real-path exercise.
8. Define broad verification, dependencies, scope exclusions, risks, stop
   conditions, documentation sync, and a binary exit gate.
9. Audit against `docs/ROADMAP_GUIDE.md`.

Return exactly one:

```text
ROADMAP ACCEPTED
ROADMAP BLOCKED
```

## Audit Output

Use:

```text
Artifact:
Result:
Accepted:
- exact capabilities or slices
Blocking gaps:
- identifier or location: exact missing decision
Changed:
- files, or none
Next:
```

Do not soften blockers with percentages.

## Rules

- Do not invent product decisions when decision authority is absent.
- Do not infer success, error, or side-effect semantics from repository
  structure alone.
- Do not create horizontal scaffolding slices.
- Do not use broad verification as the only acceptance evidence.
- Do not expose future interfaces merely to reserve them.
- Do not mark a roadmap accepted while its next slice references a draft or
  ambiguous capability.
- Do not modify production code.
