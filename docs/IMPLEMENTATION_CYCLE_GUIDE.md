# Measurable Implementation Cycle Guide

## Required Inputs

The workflow is designed around four cooperating artifacts:

```text
AGENTS.md
docs/SPECIFICATION.md
docs/ROADMAP.md
.agents/skills/
```

The active cycle state is stored in:

```text
.codex/plans/current.md
```

Completed cycles are summarized in:

```text
.codex/cycles/history.md
```

## Readiness Sequence

### 1. Install The Kit

Copy the kit into the repository root.

### 2. Write And Accept The Specification

Start from:

```text
docs/SPECIFICATION.template.md
```

Read:

```text
docs/SPECIFICATION_GUIDE.md
```

Then run:

```text
$specification-roadmap audit specification docs/SPECIFICATION.md
```

Do not implement while the selected capability is `SPECIFICATION BLOCKED`.

### 3. Write And Accept The Roadmap

Start from:

```text
docs/ROADMAP.template.md
```

Read:

```text
docs/ROADMAP_GUIDE.md
```

Then run:

```text
$specification-roadmap audit roadmap docs/ROADMAP.md
```

Do not implement while the selected slice is `ROADMAP BLOCKED`.

### 4. Run One Slice

```text
$implementation-cycle run the next accepted roadmap slice
```

Or select a specific slice:

```text
$implementation-cycle run M1-S1 from docs/ROADMAP.md
```

One pass means one selected slice, not the entire roadmap.

## Phase Flow

Feature:

```text
DISCOVERY READY
-> CONTRACT READY
-> DESIGN READY or DESIGN NOT REQUIRED
-> RED CONFIRMED
-> BUILD APPLIED
-> GREEN VERIFIED
-> DOCS SYNCED or DOCS NOT REQUIRED
-> REVIEW PASSED
-> CYCLE READY
```

Bug fix:

```text
DISCOVERY READY
-> RED CONFIRMED
-> CONTRACT READY
-> BUILD APPLIED
-> GREEN VERIFIED
-> DOCS SYNCED or DOCS NOT REQUIRED
-> REVIEW PASSED
-> CYCLE READY
```

Refactor:

```text
DISCOVERY READY
-> BASELINE GREEN
-> DESIGN READY or DESIGN NOT REQUIRED
-> BUILD APPLIED
-> GREEN VERIFIED
-> DOCS SYNCED or DOCS NOT REQUIRED
-> REVIEW PASSED
-> CYCLE READY
```

Documentation:

```text
DISCOVERY READY
-> CONTRACT UPDATED or DOCS SYNCED
-> REVIEW PASSED
-> CYCLE READY
```

## Measurement Model

Every phase record contains:

```text
Phase:
Result:
Evidence:
- exact command or observed fact
Changed:
- files changed, or none
Next:
Blocker:
```

Evidence must be one or more of:

- exact command and exit result;
- observed output or response;
- observed state change or side effect;
- focused diff tied to the selected criterion;
- explicit reason a check was not applicable, not run, or blocked.

No phase receives percentage credit.

## Successful Result

`CYCLE READY` means:

- one bounded roadmap slice was selected;
- contract references were resolved;
- meaningful red or baseline evidence was obtained;
- production behavior was implemented when code was in scope;
- focused tests passed;
- the production path was exercised;
- applicable negative paths passed;
- applicable broad checks passed;
- documentation was synchronized or correctly marked not required;
- diff review passed;
- no unresolved blocking finding remains.

It does not automatically mean:

- committed;
- pushed;
- merged;
- released;
- deployed;
- production healthy.

Those require separate authorization and evidence.

## Repair Loops

The orchestrator may repair within the selected slice.

```text
GREEN FAILED -> BUILD
REVIEW FAILED -> owning phase
```

The default limit is three build/green repair loops in one invocation.

After the limit:

```text
CYCLE BLOCKED
```

The result must include the failing command, observed failure, attempted changes,
remaining blocker, and next smallest action.

## Stop Conditions

The cycle stops when:

- the specification or roadmap is materially ambiguous;
- authorities conflict;
- implementation would exceed slice scope;
- a security or data-safety invariant would be bypassed;
- required credentials, hardware, service, dependency, or permission is absent;
- destructive or external action requires authorization;
- the production boundary cannot be exercised;
- verification fails for a reason that cannot be repaired within the slice;
- unrelated user changes would be overwritten.

Stopping with exact evidence is a valid measurable result. Invented progress is
not.

## Review The Result

Inspect:

```text
.codex/plans/current.md
.codex/cycles/history.md
git diff
```

A completed cycle should expose:

- selected IDs;
- phase results;
- exact commands;
- real-path observation;
- changed files;
- skipped checks;
- remaining limitations;
- readiness outcome.

## Recommended First Trial

Choose a small slice with:

- one success path;
- one negative path;
- one observable side effect or return value;
- deterministic focused verification;
- one real-path exercise;
- no external credentials;
- no irreversible operation.

Use the first cycle to improve repository-specific commands and contract detail
before selecting a larger slice.
