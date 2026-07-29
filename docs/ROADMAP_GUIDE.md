# Roadmap Authoring Guide

## Purpose

A roadmap converts an accepted specification into an ordered sequence of
measurable implementation slices.

A roadmap is not:

- a calendar promise;
- a component inventory;
- a broad implementation plan;
- an architecture wish list;
- proof of completion.

Use `docs/ROADMAP.template.md` after the specification is accepted.

## Slice Standard

One roadmap slice must produce one reviewable, observable result in one
implementation cycle.

A valid slice has:

- one mode;
- one purpose;
- specific specification and acceptance-criterion references;
- one primary observable boundary;
- useful red evidence or a baseline-green requirement;
- a focused verification command or procedure;
- a real-path exercise;
- applicable broad verification;
- explicit scope exclusions;
- dependencies and risks;
- a binary exit gate.

Keep slices narrow enough that a future implementation cycle can be completed
while the relevant contract, code, tests, and evidence fit comfortably within
the first 30% of a typical Codex token window. This keeps the active context
small enough to reduce drift, missed constraints, and unsupported assumptions.

If the slice cannot satisfy these fields without unrelated work or excessive
context, split it.

## Prefer Vertical Slices

A vertical slice crosses the minimum layers required to deliver one usable
behavior.

Good:

```text
Accept a valid request, persist the record, return its identifier, and prove the
result through the public API.
```

Poor:

```text
Create repository interfaces.
Add handlers.
Build frontend components.
Add tests later.
```

The poor version produces horizontal scaffolding and delays usable behavior.

## Slice Modes

Use exactly one mode.

### Feature

Requires useful failing evidence before implementation when practical.

```text
red -> build -> green -> real path -> broad checks -> review
```

### Bug Fix

Requires reproduction of the defect before applying the fix.

```text
reproduce red -> build -> regression green -> review
```

### Refactor

Requires a passing behavioral baseline before changing code.

```text
baseline green -> refactor -> same behavior green -> review
```

### Documentation

Changes a contract or durable documentation without claiming code exists.

```text
document or contract -> review
```

Documentation-only slices must state whether implementation is intentionally
deferred and which future slice will implement the contract when known.

## Specification References

Reference exact capability and acceptance-criterion IDs:

```text
CAP-004
AC-014
AC-015
INV-002
```

Do not use broad references such as:

```text
configuration section
API work
security requirements
```

unless those sections also contain stable identifiers.

Every acceptance criterion referenced by a slice must be either:

- verified by that slice;
- explicitly excluded with a reason;
- delegated to another named slice.

## Observable Result

State what becomes usable or demonstrably corrected.

Good:

```text
A caller can create a record through the public API and retrieve its generated
identifier.
```

Bad:

```text
Storage foundation is complete.
```

## Acceptance Boundary

Choose the highest practical real boundary that remains deterministic.

Examples:

- command process;
- public library API;
- HTTP route through production middleware;
- UI workflow through backend;
- file transformation;
- event consumer;
- scheduled job;
- deployed smoke test;
- hardware interaction.

One slice may use supporting lower-level tests, but it must identify one primary
boundary that proves delivery.

## Expected Red Or Baseline Evidence

For feature and bug-fix slices, define what should fail before implementation.

Good:

```text
The focused API test returns 404 because the production route does not exist.
```

Bad:

```text
Tests should fail.
```

For refactors, define the exact passing baseline that must remain green.

Do not accept compile errors, missing test fixtures, or intentionally broken
mocks as useful red evidence unless they prove the selected production boundary
is absent.

## Focused Verification

Name the narrowest deterministic command or procedure that proves the slice.

Examples:

```text
pytest tests/api/test_create_record.py
cargo test create_record_contract
go test ./internal/api -run TestCreateRecord
npm test -- create-record
```

When the final command cannot be known before repository inspection, describe
how it will be discovered and require `.codex/plans/current.md` to record the
resolved command before implementation begins.

## Real-Path Exercise

Define a production-path command or procedure separate from automated tests.

Examples:

- invoke the built CLI with isolated temporary state;
- send a request to the actual local server;
- call the installed library entry point;
- run the UI workflow against the production backend;
- process a fixture through the real executable;
- run the migration against an ephemeral database.

The real-path exercise must record output, status, state changes, or side
effects. “Inspect the code” is not a real-path exercise.

## Broad Verification

Name repository-health checks such as:

- full test suite;
- lint;
- formatting;
- type checking;
- build;
- schema or migration validation;
- packaging;
- security checks;
- CI-equivalent command.

Broad checks do not replace focused behavioral evidence.

## Dependencies

A dependency is another accepted slice, existing capability, environment, or
decision required before the selected slice can begin.

Use stable references:

```text
Depends on M1-S2 and accepted INV-004.
```

Do not create implicit dependencies hidden in prose.

## In Scope And Out Of Scope

In-scope items name the minimum work required to prove the observable result.

Out-of-scope items prevent adjacent functionality, cleanup, or generalized
infrastructure from entering the diff.

A useful out-of-scope list often includes:

- adjacent commands or endpoints;
- interactive variants;
- migration or backward-compatibility work assigned elsewhere;
- UI polish unrelated to the behavior;
- broad refactoring;
- performance optimization;
- optional integrations;
- release or deployment.

## Risks And Stop Conditions

Name material risks:

- destructive side effects;
- security boundaries;
- irreversible data changes;
- unavailable hardware or service;
- non-deterministic external dependency;
- compatibility risk;
- required migration;
- expensive verification.

A slice must stop rather than fabricate completion when a required boundary
cannot be exercised.

## Exit Gate

Use a binary exit gate.

Recommended form:

```text
All referenced acceptance criteria have GREEN VERIFIED evidence through the
named real boundary, applicable broad checks pass, documentation is synchronized
or not required with reason, and review passes with no unresolved blocking
finding.
```

Do not use percentage completion.

## Slice Sizing Heuristics

Split a slice when:

- it has multiple independent observable outcomes;
- it requires unrelated public interfaces;
- success and failure paths belong to different subsystems with independent
  release value;
- the focused command would need a broad suite to prove anything;
- it cannot be reviewed without large unrelated context;
- it would require carrying more than roughly the first 30% of a Codex token
  window to keep the contract, implementation, tests, and evidence coherent;
- it requires multiple architectural decisions;
- its out-of-scope list is unclear;
- one failed criterion would not identify the owning behavior.

Do not split so far that a slice only adds scaffolding.

## Roadmap Acceptance Gate

Before changing roadmap status to `Accepted`, confirm:

```text
[ ] The source specification is Accepted.
[ ] Every near-term slice references exact contract IDs.
[ ] Every slice has exactly one mode.
[ ] Every slice delivers one observable result.
[ ] Every slice names a primary acceptance boundary.
[ ] Feature and bug-fix slices define useful red evidence.
[ ] Refactor slices define a baseline-green command.
[ ] Every slice defines focused verification.
[ ] Every slice defines a real-path exercise.
[ ] Applicable broad verification is named.
[ ] Dependencies are explicit and acyclic.
[ ] In-scope and out-of-scope work are explicit.
[ ] Risks and stop conditions are explicit.
[ ] Exit gates are binary and evidence-based.
[ ] No slice is only horizontal scaffolding.
[ ] The next slice is small enough for one implementation cycle.
[ ] The next slice is small enough to keep required context within roughly the
    first 30% of a typical Codex token window.
```

If any applicable item fails, use `Roadmap status: Draft`.

## Anti-Patterns

Reject roadmap items such as:

```text
Build the backend.
Implement the UI.
Add all tests.
Create infrastructure.
Finish configuration.
Polish the system.
```

Reject slices that:

- contain no contract references;
- use only component deliverables;
- defer tests until later;
- use broad checks as the only evidence;
- omit real-path execution;
- have vague exit gates;
- mix feature work, refactoring, documentation, and release;
- depend on future unspecified behavior;
- claim dates or completion without evidence.

## Authoring With Codex

Create or improve a roadmap:

```text
$specification-roadmap create roadmap from docs/SPECIFICATION.md
```

Audit it:

```text
$specification-roadmap audit roadmap docs/ROADMAP.md
```

The skill must return either:

```text
ROADMAP ACCEPTED
```

or:

```text
ROADMAP BLOCKED
```

with exact slice defects.

After both documents are accepted:

```text
$implementation-cycle run the next accepted roadmap slice
```
