# Roadmap

## Status

- Roadmap status: Draft | Accepted
- Roadmap version:
- Specification source: `docs/SPECIFICATION.md`
- Specification version:
- Current milestone:
- Next eligible slice:
- Last reviewed:

This roadmap selects implementation order. It does not claim behavior is
implemented.

## Slice Rules

Every implementation slice must:

- deliver one observable behavior or one bounded defect correction;
- use exactly one mode;
- reference exact specification and acceptance-criterion IDs;
- name one primary acceptance boundary;
- define useful red or baseline-green evidence;
- define focused verification;
- define a real-path exercise;
- define applicable broad verification;
- state dependencies, in-scope work, and out-of-scope work;
- identify material risks and stop conditions;
- use a binary exit gate;
- avoid horizontal scaffolding without a working path;
- fit one reviewable diff across Implementation, Test, and Review gates;
- keep required contract, code, tests, and evidence context for each gate within
  roughly the first 30% of a typical Codex token window.

Split a slice when its acceptance criteria cannot be proven together without
unrelated changes or excessive active context.

## Milestone M1: Name

### Slice M1-S1: Observable behavior

- Status: proposed | accepted | active | completed | blocked
- Mode: feature | bugfix | refactor | documentation
- Purpose:
- Specification references: `CAP-...`, `AC-...`, `INV-...`
- Observable result:
- Primary acceptance boundary:
- Expected red or baseline-green evidence:
- Focused verification command or deterministic discovery rule:
- Real-path exercise:
- Broad verification:
- Required environment:
- Dependencies:
- In scope:
- Out of scope:
- Risks:
- Stop conditions:
- Documentation synchronization:
- Exit gate:
- Completion evidence reference:

### Slice M1-S2: Observable behavior

- Status:
- Mode:
- Purpose:
- Specification references:
- Observable result:
- Primary acceptance boundary:
- Expected red or baseline-green evidence:
- Focused verification command or deterministic discovery rule:
- Real-path exercise:
- Broad verification:
- Required environment:
- Dependencies:
- In scope:
- Out of scope:
- Risks:
- Stop conditions:
- Documentation synchronization:
- Exit gate:
- Completion evidence reference:

## Dependency Order

```text
M1-S1 -> M1-S2
```

Keep dependencies explicit and acyclic.

## Future Milestones

List only enough detail to preserve ordering and dependencies. Expand a future
slice when it becomes a realistic candidate for the next implementation cycle.

## Roadmap Acceptance Record

- Audit result: ROADMAP ACCEPTED | ROADMAP BLOCKED
- Reviewed slices:
- Blocking gaps:
- Evidence or review reference:
