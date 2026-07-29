# Specification Authoring Guide

## Purpose

A specification is the authoritative description of externally observable
behavior. It tells Codex what must be true without prescribing unnecessary
implementation details.

A specification is not:

- an implementation status report;
- a repository layout proposal;
- a list of hoped-for features;
- a substitute for tests;
- proof that behavior exists.

Use `docs/SPECIFICATION.template.md` to create the repository specification.

## Acceptance Standard

Mark a specification `Accepted` only when a reviewer can derive a meaningful
test and expected result without asking Codex to invent product behavior.

Each capability must answer:

1. Who or what invokes the behavior?
2. What input and preconditions apply?
3. What constitutes success?
4. What output, response, event, or state change is observable?
5. What side effects are permitted or required?
6. What failure cases exist?
7. How are errors represented?
8. What lifecycle, ordering, timeout, cancellation, or cleanup rules apply?
9. What security, authorization, privacy, or safety constraints apply?
10. What compatibility guarantees apply?
11. What acceptance boundary proves the behavior?
12. Which acceptance criteria define completion?

If any answer materially affects implementation or testing and remains
ambiguous, the capability is not implementation-ready.

## Use Stable Identifiers

Use stable identifiers so the roadmap, tests, code review, and cycle evidence can
refer to the same contract.

Recommended forms:

```text
CAP-001    capability
AC-001     acceptance criterion
INV-001    invariant
ERR-001    error contract
DATA-001   persistence or data contract
```

Do not renumber existing identifiers casually. Add new identifiers when
behavior changes.

## Write Capabilities As Observable Contracts

A capability section should describe one coherent externally visible behavior.

Good scope:

```text
CAP-004: Create a configuration file at a requested path.
```

Poor scope:

```text
CAP-004: Implement configuration infrastructure.
```

The first can be exercised. The second invites horizontal scaffolding.

### Goal

State the user, caller, operator, or dependent-system outcome.

Bad:

```text
Add a storage abstraction.
```

Good:

```text
Persist a submitted note so it can be retrieved after process restart.
```

### Inputs And Preconditions

Name exact input forms, required values, limits, defaults, permissions,
environment assumptions, and prerequisite state.

Do not use vague phrases such as:

```text
valid input
appropriate permissions
normal configuration
```

Define what those phrases mean.

### Successful Behavior

Describe what is observably true after success. Include ordering and atomicity
where relevant.

Examples:

- returns the created identifier;
- writes one file and no others;
- emits one event after the transaction commits;
- renders the new state after the backend confirms success;
- exits with the documented status.

### Output Or Response

Define applicable:

- return values;
- stdout and stderr;
- HTTP or protocol status;
- response schema;
- events;
- files or generated artifacts;
- user-visible state;
- logs only when logs are part of the contract.

Distinguish stable machine-facing output from unstable human-facing prose.

### Side Effects

List required and forbidden effects.

Examples:

- files created or modified;
- database changes;
- external calls;
- messages published;
- device or host changes;
- cache invalidation;
- notifications;
- no-op behavior.

A successful response without the required side effect is not success.

### Errors And Negative Behavior

Specify failure categories and observable representation.

Include relevant cases:

- missing or malformed input;
- unauthorized or forbidden operation;
- not found or conflict;
- unavailable dependency;
- timeout or cancellation;
- partial failure;
- persistence failure;
- concurrency conflict;
- unsupported operation.

Define whether failure may leave partial state.

### Lifecycle And Cleanup

When relevant, define:

- startup and shutdown;
- retries;
- idempotency;
- ordering;
- concurrency;
- cancellation;
- timeout;
- rollback;
- resource cleanup;
- resumability.

### Security And Data Safety

State invariants rather than generic aspirations.

Bad:

```text
The endpoint must be secure.
```

Good:

```text
Only an authenticated owner may delete the record.
Authorization is checked before any deletion side effect.
```

### Compatibility

State applicable guarantees:

- backward-compatible inputs;
- versioned API behavior;
- file-format evolution;
- migration behavior;
- supported platforms;
- deprecation rules;
- default preservation.

## Acceptance Criteria

Each acceptance criterion should be:

- binary;
- observable;
- independent enough to diagnose;
- traceable to one capability;
- specific about expected result;
- free of implementation details unless the implementation detail is itself a
  contract.

Use Given/When/Then when useful:

```text
AC-014: Given an existing destination and overwrite disabled, when creation is
requested, then the original file remains unchanged and the operation reports
the documented conflict.
```

Avoid:

```text
AC-014: Configuration works correctly.
```

A criterion that cannot fail clearly is not measurable.

## Observable Acceptance Boundary

Name the real boundary through which the criterion will be proven.

Examples:

- public library call;
- process invocation;
- HTTP request through production routing;
- protocol exchange;
- UI action through the real backend;
- filesystem operation;
- message consumer and emitted event;
- scheduled job run;
- device or hardware operation.

Also name:

- focused evidence expected;
- real-path exercise;
- environment requirements;
- substitutes allowed, such as deterministic fake services.

Mocks or injected seams do not prove production behavior unless the seam itself
is the specified boundary.

## Configuration, Data, And Interfaces

Document configuration only when it changes behavior. Define precedence,
defaults, validation, and failure semantics.

For stored data, define ownership, retention, cleanup, schema compatibility,
atomicity, and corruption behavior where relevant.

For interfaces, define exact public semantics and avoid describing unselected
future endpoints, commands, screens, or settings.

## Open Questions

An open question belongs in the specification only when it blocks or materially
changes the contract.

Every blocking question must identify:

- affected capability or criterion;
- decision needed;
- why it matters;
- who or what can resolve it.

Do not send an implementation cycle into a capability with unresolved blocking
questions.

## Specification Acceptance Gate

Before changing status to `Accepted`, confirm:

```text
[ ] Purpose and scope are explicit.
[ ] Out-of-scope behavior is explicit.
[ ] Actors and public boundaries are identified.
[ ] Every implementation-ready capability has a stable ID.
[ ] Inputs and preconditions are concrete.
[ ] Successful behavior is observable.
[ ] Output and side effects are defined.
[ ] Failure behavior is defined.
[ ] Lifecycle rules are defined where relevant.
[ ] Security and data-safety invariants are concrete.
[ ] Compatibility expectations are stated.
[ ] Every capability has binary acceptance criteria.
[ ] Every acceptance criterion names a real acceptance boundary.
[ ] No blocking open question remains for the next roadmap slice.
[ ] The document does not claim implementation status.
```

If any applicable item fails, use `Contract status: Draft`.

## Anti-Patterns

Reject specifications that:

- describe components but not behavior;
- list features without acceptance criteria;
- say “support,” “handle,” “robust,” or “secure” without exact semantics;
- expose future interfaces before their contracts exist;
- encode repository structure as product behavior;
- define only happy paths;
- omit side effects or failure cleanup;
- let examples contradict normative text;
- present future behavior as implemented;
- contain acceptance criteria that only prove scaffolding.

## Authoring With Codex

Use:

```text
$specification-roadmap create or improve docs/SPECIFICATION.md
```

Then audit it:

```text
$specification-roadmap audit specification docs/SPECIFICATION.md
```

The skill must return either:

```text
SPECIFICATION ACCEPTED
```

or:

```text
SPECIFICATION BLOCKED
```

with exact missing contracts. Acceptance here means implementation-ready
documentation, not implemented software.
