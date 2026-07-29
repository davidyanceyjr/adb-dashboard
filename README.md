# Codex Measurable Implementation Cycle — Guided Kit

This kit combines a project- and language-agnostic `AGENTS.md`, Codex skills,
specification and roadmap authoring guidance, templates, examples, and measurable
cycle state.

Its goal is one bounded implementation pass per selected roadmap slice with
evidence at every phase.

## Layout

```text
AGENTS.md
.agents/skills/
  implementation-cycle/
  specification-roadmap/
  cycle-contract/
  cycle-design/
  cycle-test/
  cycle-build/
  cycle-document/
  cycle-review/
  cycle-handoff/
docs/
  SPECIFICATION_GUIDE.md
  ROADMAP_GUIDE.md
  IMPLEMENTATION_CYCLE_GUIDE.md
  READINESS_CHECKLIST.md
  SPECIFICATION.template.md
  ROADMAP.template.md
  examples/
    EXAMPLE_SPECIFICATION.md
    EXAMPLE_ROADMAP.md
.codex/plans/
  current.md
.codex/cycles/
  history.md
```

## Install

Copy the kit contents into the repository root.

Create the project documents:

```sh
cp docs/SPECIFICATION.template.md docs/SPECIFICATION.md
cp docs/ROADMAP.template.md docs/ROADMAP.md
```

Replace template text with actual project decisions. Do not retain generic
claims that are not true.

## Author The Specification

Read:

```text
docs/SPECIFICATION_GUIDE.md
```

Create or improve the contract:

```text
$specification-roadmap create or improve docs/SPECIFICATION.md
```

Audit it:

```text
$specification-roadmap audit specification docs/SPECIFICATION.md
```

Do not implement the selected capability until the result is:

```text
SPECIFICATION ACCEPTED
```

## Author The Roadmap

Read:

```text
docs/ROADMAP_GUIDE.md
```

Create vertical slices:

```text
$specification-roadmap create roadmap from docs/SPECIFICATION.md
```

Audit them:

```text
$specification-roadmap audit roadmap docs/ROADMAP.md
```

Do not begin the selected slice until the result is:

```text
ROADMAP ACCEPTED
```

## Check Readiness

Use:

```text
docs/READINESS_CHECKLIST.md
```

Or invoke:

```text
$specification-roadmap audit readiness
```

## Run One Implementation Pass

Run the next accepted slice:

```text
$implementation-cycle run the next accepted roadmap slice
```

Or name it:

```text
$implementation-cycle run M1-S1 from docs/ROADMAP.md
```

One pass means one selected slice. The orchestrator continues through contract,
design when needed, red evidence, production implementation, green evidence,
real-path exercise, documentation synchronization, broad verification, and
review.

A successful pass ends with:

```text
CYCLE READY
```

## Measured Results

Each phase records:

- fixed binary result;
- exact commands and results;
- observed output, response, state, or side effect;
- changed files;
- skipped checks and reasons;
- blocker and next phase.

No completion percentages are used.

## State And History

Active evidence:

```text
.codex/plans/current.md
```

Compact completed-cycle history:

```text
.codex/cycles/history.md
```

## Important Limits

`CYCLE READY` does not automatically mean committed, pushed, merged, released,
deployed, or healthy in production.

The cycle stops with exact evidence when the contract is ambiguous, scope must
expand materially, required environment is unavailable, security or data safety
would be bypassed, or the production boundary cannot be exercised.

## Start With The Example

The files under `docs/examples` show one accepted capability and one accepted
vertical slice. They are illustrative and should not be copied as project
requirements.
