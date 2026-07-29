---
name: implementation-cycle
description: Run one complete, measurable implementation or documentation cycle from specification and roadmap through tests, code, documentation sync, and review. Trigger on "implementation pass", "next roadmap slice", "run the cycle", "complete this slice", "cycle status", "resume cycle", or "advance cycle". Do not use for an isolated single-phase request.
---

# Measurable Implementation Cycle

Follow `AGENTS.md`. This skill orchestrates one bounded cycle; it does not relax
any repository rule.

## Objective

Produce one measurable pass for exactly one roadmap slice or equivalently
bounded user request.

Continue through all applicable phases in the same run. Do not stop after
planning, tests, scaffolding, or implementation if the next safe local phase can
be completed now.

Stop only for:

- a genuine contract or authority conflict;
- unavailable required environment, dependency, credential, service, hardware,
  or permission;
- destructive or external action requiring approval;
- material scope expansion;
- a verification failure that cannot be repaired within the selected slice;
- the repair-loop limit.

## Files To Read

Before work:

1. Read applicable `AGENTS.md` files.
2. Discover and read the accepted specification.
3. Discover and read the roadmap or selected issue.
4. Read `.codex/plans/current.md` when present.
5. Read relevant implementation, tests, CI, and build entry points.
6. Read `docs/IMPLEMENTATION_CYCLE_GUIDE.md` when present.
7. Read the sibling phase skill before entering that phase:

```text
../cycle-contract/SKILL.md
../cycle-design/SKILL.md
../cycle-test/SKILL.md
../cycle-build/SKILL.md
../cycle-document/SKILL.md
../cycle-review/SKILL.md
../cycle-handoff/SKILL.md
```

Do not load every sibling skill up front. Read each only when its phase is
reached.

## Supported Operations

Interpret the user's operation from the prompt:

- `run`: select or resume one slice and execute all applicable phases.
- `next`: resume the active cycle and continue to its next valid phase.
- `status`: inspect evidence and report state without editing production code.
- `pause`: update active state and use `cycle-handoff`.
- `resume`: validate active state against repository reality, then continue.
- `close`: close only a cycle that has passed review or is explicitly abandoned.

Default to `run` when implementation intent is clear.

## Cycle Modes

Choose exactly one:

```text
feature
bugfix
refactor
documentation
```

### Feature

```text
discover -> contract -> design-if-needed -> red -> build -> green
         -> documentation-sync -> review -> ready
```

### Bug fix

```text
discover -> reproduce-red -> contract-confirmation -> build -> green
         -> documentation-sync-if-needed -> review -> ready
```

### Refactor

```text
discover -> baseline-green -> design-if-needed -> build -> green
         -> documentation-sync-if-needed -> review -> ready
```

### Documentation

```text
discover -> contract-or-document -> review -> ready
```

A documentation cycle cannot claim implemented behavior.

## State File

Use `.codex/plans/current.md` as the single active state file.

Create it from the repository template when absent and a real cycle is starting.
Do not create it for status-only inspection of a repository that does not use
this workflow.

The state must contain:

```text
Cycle ID
Mode
Goal
Roadmap slice
Branch or work context
Specification anchors
Acceptance criteria
Acceptance boundary
In scope
Out of scope
Focused test command
Real-path command or procedure
Broad verification commands
Current phase
Phase results
Blocker
Next phase
```

Update state after evidence is obtained, not before.

## Phase Result Protocol

Every phase writes one result using this shape:

```text
Phase:
Result:
Evidence:
- exact command or observed fact
Changed:
- relevant files, or none
Next:
Blocker: none | exact blocker
```

Allowed phase results:

```text
DISCOVERY READY
DISCOVERY BLOCKED
CONTRACT READY
CONTRACT UPDATED
CONTRACT BLOCKED
DESIGN READY
DESIGN NOT REQUIRED
DESIGN BLOCKED
RED CONFIRMED
RED INVALID
BASELINE GREEN
BASELINE FAILED
BUILD APPLIED
BUILD PARTIAL
BUILD BLOCKED
GREEN VERIFIED
GREEN FAILED
VERIFICATION BLOCKED
DOCS SYNCED
DOCS NOT REQUIRED
DOCS BLOCKED
REVIEW PASSED
REVIEW FAILED
REVIEW BLOCKED
CYCLE READY
CYCLE BLOCKED
CYCLE ABANDONED
```

Do not invent softer synonyms.

## Execution Algorithm

### 0. Authoring Readiness Gate

- Confirm the selected specification capability is accepted and has stable
  capability and acceptance-criterion identifiers.
- Confirm the selected roadmap slice is accepted, references those identifiers,
  defines one observable result, one primary boundary, red or baseline evidence,
  focused verification, a real-path exercise, scope exclusions, and a binary
  exit gate.
- Read `docs/SPECIFICATION_GUIDE.md`, `docs/ROADMAP_GUIDE.md`, and
  `docs/READINESS_CHECKLIST.md` when present.
- When either artifact fails readiness, invoke or follow
  `../specification-roadmap/SKILL.md` and stop with `DISCOVERY BLOCKED` rather
  than inventing missing behavior.

### 1. Discover

- Inspect Git or version-control state.
- Preserve unrelated user work.
- Select exactly one roadmap slice or bounded request.
- Determine the cycle mode.
- Identify contract, acceptance criteria, boundary, expected evidence, and
  repository verification commands.
- Initialize or validate the state file.

If no bounded slice can be selected, return `DISCOVERY BLOCKED`.

### 2. Contract

Read `cycle-contract` and execute it.

The contract phase must end with `CONTRACT READY`, `CONTRACT UPDATED`, or
`CONTRACT BLOCKED`.

Do not proceed from `CONTRACT BLOCKED`.

### 3. Design When Needed

Read `cycle-design` only when implementation placement, ownership, lifecycle,
data flow, compatibility, or safety boundaries are genuinely unclear.

Simple changes with an obvious existing location must record
`DESIGN NOT REQUIRED`.

### 4. Establish Red Or Baseline Green

Read `cycle-test`.

- Feature: obtain `RED CONFIRMED`.
- Bug fix: reproduce the defect and obtain `RED CONFIRMED`.
- Refactor: obtain `BASELINE GREEN` before changing behavior-preserving code.
- Documentation: skip with an explicit not-applicable reason unless the
  documentation itself has executable examples or validation.

Do not proceed from `RED INVALID` or `BASELINE FAILED` without resolving the
reason.

### 5. Build

Read `cycle-build` and implement the smallest complete production path.

`BUILD APPLIED` does not mean verified. Continue immediately to green testing.

### 6. Green And Verification

Read `cycle-test` in green mode.

Require:

- focused behavior passes;
- real production path is exercised;
- applicable negative behavior passes;
- applicable broader checks run.

If green fails, return to build with a repair packet. Limit build-green repair
loops to three per cycle run. After the third failed loop, return
`CYCLE BLOCKED` with exact evidence.

### 7. Documentation Synchronization

Read `cycle-document`.

Synchronize public or durable documentation when behavior, interfaces,
configuration, errors, operations, compatibility, or examples changed.

Record `DOCS NOT REQUIRED` only with a reason.

### 8. Review

Read `cycle-review`.

Review the diff, state evidence, scope, tests, documentation, safety, and SLOP
risk.

A material review finding must identify its owning return phase. Apply at most
two review repair loops in the same cycle run. Rerun every invalidated gate.
After two unsuccessful review repair loops, return `CYCLE BLOCKED`.

### 9. Ready Or Close

`CYCLE READY` requires `REVIEW PASSED` and all applicable gates.

Do not commit, push, open a pull request, deploy, publish, or modify external
systems unless the user or repository workflow explicitly authorizes that
operation.

When commit is authorized:

- stage only in-scope files;
- use a specific commit message;
- rerun any gate invalidated by staging or generated files;
- record the commit identifier.

On successful closure, append one row to `.codex/cycles/history.md` containing:

```text
Date | Cycle ID | Mode | Slice | Result | Focused evidence | Broad evidence | Commit
```

Then reset `.codex/plans/current.md` to an inactive state or the next explicitly
selected slice. Do not silently select another implementation slice in the same
cycle.

## Final Output

Use exactly this structure:

```markdown
## Cycle Result

- Cycle:
- Mode:
- Slice:
- Result: CYCLE READY | CYCLE BLOCKED | CYCLE ABANDONED
- Final phase:

## Delivered

- Exact behavior or documentation result

## Evidence

- Acceptance criterion: result and command/evidence
- Focused test: command and result
- Real path: command/procedure and result
- Broad checks: commands and results
- Review: result

## Changed

- Material files

## Gaps

- Known limitations
- Not run or blocked checks
- Intentionally excluded work

## Next

- Commit/review action, exact blocker, or none
```

Do not add a celebratory summary or unsupported readiness language.
