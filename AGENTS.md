# AGENTS.md

## Mission

Build working software whose behavior can be observed, tested, and reproduced.

Optimize for:

- one small, complete vertical slice at a time;
- externally observable behavior over repository appearance;
- meaningful tests at the real boundary;
- focused diffs and traceable version-control history;
- factual status backed by commands actually run.

Repository appearance is not progress.

## No-SLOP Rule

SLOP is ceremony, scaffolding, abstraction, documentation, or confident status
language that creates the appearance of implementation without verified
behavior.

Do not produce it.

Never substitute any of the following for working behavior:

- plans, specifications, roadmaps, diagrams, ADRs, or checklists;
- interfaces, schemas, routes, handlers, commands, components, packages, or
  directory layouts without a working production path;
- TODOs, placeholder returns, empty handlers, unimplemented exceptions, dead
  feature flags, or commented-out implementations;
- mocked, fabricated, or hard-coded success in production paths;
- UI controls, API endpoints, commands, or configuration for behavior the
  system cannot perform;
- tests that only prove files, symbols, constructors, mocks, snapshots, or help
  text exist;
- compilation, lint, type checking, coverage, schema validation, or builds
  presented as functional proof;
- broad refactoring, dependency churn, or cleanup unrelated to the selected
  behavior;
- completion claims based on commands that were not run.

A behavior is implemented only when its production path exists. It is verified
only when the intended behavior has been exercised at an observable boundary.

When blocked, report the exact blocker. Do not replace failed implementation
with future-work prose or a polished completion summary.

## Status Vocabulary

Use these states literally:

```text
specified    intended behavior is defined
planned      one bounded slice is selected
covered      a meaningful automated test encodes the behavior
implemented  production code exists for the behavior
verified     real behavior and applicable checks passed
committed    verified work exists in version-control history
released     verified work is included in a release
```

Do not collapse states or report a later state without evidence for earlier
states.

Examples:

- Documentation may be `specified`, not `implemented`.
- A mock or test seam may be `covered`, not verified production behavior.
- A build may pass while the behavior remains unverified.
- An unavailable response is not implementation of the underlying feature.
- A merged change is not necessarily released.

## Instruction And Authority Discovery

Before changing behavior, discover the applicable authorities. Common sources
include:

```text
AGENTS.md
README files
user-facing specifications
API or protocol definitions
architecture documents
accepted decision records
roadmaps or issue trackers
active plans
handoff notes
build scripts
package or workspace manifests
continuous-integration configuration
existing tests
```

Do not assume fixed filenames except where this repository explicitly adopts
them. The templates paired with this file use:

```text
docs/SPECIFICATION.md
docs/ROADMAP.md
.codex/plans/current.md
.codex/cycles/history.md
```

Use this default precedence unless the repository defines another:

1. Current user instructions.
2. More specific nested `AGENTS.md` or override instructions.
3. Accepted public behavior contract or specification.
4. Accepted architecture constraints and decisions.
5. Selected roadmap item, issue, or active plan.
6. Existing tests and production behavior.
7. Handoff notes and historical documentation.

Tests and current behavior are evidence of implementation, not permission to
silently contradict the intended contract.

When authorities materially conflict:

- stop implementation;
- identify the exact conflict;
- resolve or update the higher-level authority first;
- do not choose the interpretation that merely makes coding easier.

Do not create specifications, architecture documents, plans, or roadmaps merely
to satisfy ceremony. Create or update them only when they define durable,
necessary behavior or the user explicitly requests them.


## Specification And Roadmap Readiness

This kit provides:

```text
docs/SPECIFICATION_GUIDE.md
docs/ROADMAP_GUIDE.md
docs/READINESS_CHECKLIST.md
```

Before a code-bearing cycle begins:

- the selected specification capability must be accepted;
- it must define stable capability and acceptance-criterion identifiers;
- inputs, success, output, side effects, errors, lifecycle, security, data
  safety, compatibility, and acceptance boundary must be sufficiently exact;
- the selected roadmap slice must be accepted;
- it must reference exact contract identifiers;
- it must define one observable result, one primary boundary, useful red or
  baseline evidence, focused verification, a real-path exercise, explicit scope,
  risks, stop conditions, and a binary exit gate.

When those conditions are absent, improve or audit the documents with
`$specification-roadmap`. Do not allow implementation to fill contract gaps by
guessing.

An accepted specification or roadmap is permission to attempt a slice. It is not
evidence that software exists.

## Skill Workflow

This repository uses the skills under `.agents/skills`.

Use `$implementation-cycle` when the user asks to implement the next roadmap
slice, complete one bounded feature or fix, run an implementation pass, or
advance an active cycle.

This repository separates slice work into three context gates:

```text
Implementation cycle -> Test cycle -> Review cycle
```

Each gate may be run in a separate invocation to control context usage. Do not
collapse the gates merely to claim readiness.

The implementation cycle is:

```text
discover -> contract -> design-if-needed -> red-or-baseline -> build
         -> docs-if-obviously-required -> implementation-ready-for-test
```

The test cycle is:

```text
focused-green -> real-path-exercise -> negative-path-checks
              -> applicable-broad-checks -> test-ready-for-review
```

The review cycle is:

```text
diff-review -> evidence-review -> docs-review -> ready
```

The gates may route backward:

```text
contract blocked    -> contract
red invalid         -> contract or test
build blocked       -> contract or design
focused test failed -> build
docs conflict       -> contract
review finding      -> owning phase
```

For documentation-only work:

```text
discover -> contract/document -> review -> ready
```

For behavior-preserving refactoring:

```text
discover -> baseline-green -> design-if-needed -> build
test cycle verifies unchanged behavior
review cycle confirms scope and evidence
```

One invocation of `$implementation-cycle` should finish the Implementation gate
for the selected slice. It may stop before Test or Review when continuing would
consume excessive context; record the next gate explicitly in
`.codex/plans/current.md`. Stop only for a genuine blocker, a destructive or
external write requiring approval, a material expansion of scope, or a context
gate boundary.

The orchestrator may read sibling skill files for phase-specific procedure.
Phase skills may also be invoked directly for narrowly scoped work.

## Cycle State

Use `.codex/plans/current.md` as the single active-cycle state file. Do not
create parallel status documents.

The active cycle must record:

- cycle ID and mode;
- goal and selected roadmap slice;
- branch or working context;
- specification and acceptance-criterion references;
- observable acceptance boundary;
- in-scope and out-of-scope work;
- focused test command;
- real-path exercise command or procedure;
- broad verification commands;
- current gate, phase, and result;
- exact evidence from completed phases;
- blocker and next phase.

On successful closure, append one compact row to
`.codex/cycles/history.md`. Do not write a narrative diary.

A phase result is measurable only when it includes at least one of:

- an exact command and exit result;
- an observed output, response, state change, or side effect;
- a diff or file set tied to the selected behavior;
- an explicit `not applicable`, `not run`, or `blocked` reason.

Do not use arbitrary completion percentages.

## Scope Control

Implement exactly one selected roadmap slice or one equivalently bounded user
request per cycle.

Before editing, identify:

- the observable outcome;
- the production boundary;
- the acceptance criteria;
- the minimum files and layers involved;
- explicit out-of-scope behavior;
- the verification needed to prove the result.

Reserved directories, interfaces, schemas, configuration keys, route prefixes,
and package names are not feature permission.

Do not weaken intended behavior merely to match incomplete code unless project
direction has explicitly changed.

Prefer modifying existing code over adding frameworks, registries, plugin
systems, generalized factories, configuration, dependencies, or alternate
architectures.

Do not mix unrelated cleanup into a behavior slice.

## Required Implementation Pass

For each code-bearing slice, the gates together must satisfy:

1. Inspect repository instructions, relevant implementation, tests, and
   version-control state.
2. Preserve unrelated or user-created changes.
3. Select one roadmap slice or bounded request.
4. Confirm the governing contract and acceptance criteria.
5. Clarify contract ambiguity before coding.
6. Choose the real observable boundary.
7. Add or update a meaningful acceptance or regression test.
8. Run it and obtain useful failing evidence when practical.
9. Implement the smallest complete production path.
10. In the Test gate, run the focused test until it passes.
11. In the Test gate, exercise the real path and inspect output, status, side effects, lifecycle,
    and failure behavior.
12. Synchronize behavior-facing documentation when required.
13. In the Test gate, run applicable broader repository checks.
14. In the Review gate, review the diff for SLOP, placeholders, dead code, unsupported claims,
    accidental scope growth, and unrelated changes.
15. Report exact evidence and current status for the completed gate.
16. Commit only when the user or repository workflow authorizes committing.

A meaningful failing test reaches the intended boundary or fails because that
boundary is absent or defective. Compile-only, mock-only, and scaffolding-only
failures are not sufficient evidence.

A step may be skipped only when it is not applicable, cannot be performed, or
the user explicitly directs otherwise. Record the reason. Never use an
exception to support a false completion claim.

## Vertical Slice Rule

Prefer the smallest end-to-end behavior a user, client, operator, or dependent
system can observe.

A valid slice may cross multiple layers when one behavior requires them.
Examples include:

- input parsing through real output;
- API request through persisted state;
- UI action through backend enforcement;
- configuration through runtime behavior;
- file input through generated artifact;
- message receipt through externally visible side effect;
- library call through documented return and error behavior.

The following are incomplete horizontal work unless explicitly requested:

- an interface without a production implementation;
- a schema without a live consumer or producer;
- a route or command returning hard-coded success;
- a UI control without working behavior behind it;
- a process wrapper with no selected operation using it;
- storage types without documented read or write behavior;
- infrastructure that no current feature exercises.

## Testing And Evidence

A test counts as evidence for the current slice only when it would fail if that
slice's intended behavior were broken.

Prefer the highest practical boundary that remains deterministic and useful.
Possible evidence includes:

- process-level command tests;
- public API or protocol tests;
- integration tests through production routing and adapters;
- filesystem tests using isolated temporary locations;
- deterministic fake external services or executables;
- component tests for visible state and interaction;
- end-to-end workflow tests;
- hardware, device, network, or platform smoke tests where substitutes cannot
  prove the contract;
- direct library tests when the library interface is the product boundary.

Non-counting checks include:

- compilation or package existence;
- file, type, route-name, or constructor existence;
- mock-call assertions without an observable result;
- injected seams unreachable in production, unless that seam is the selected
  behavior;
- schema validation without matching runtime behavior;
- UI builds without behavior assertions;
- copied tests from another slice;
- tests that only prove unavailable or placeholder behavior;
- snapshots that bless unstable output without semantic assertions.

Test-only routes, commands, handlers, callbacks, dependencies, and seams must be
impossible to enable accidentally in production.

Fakes must be deterministic and model applicable failure modes such as:

```text
success
invalid input
partial output
standard-error output
nonzero status
exception or rejected operation
timeout
cancellation
signal or shutdown
large or streaming data
concurrency
resource exhaustion
connection loss
corrupt or malformed data
permission failure
```

Do not update snapshots, golden files, fixtures, or expected output merely to
silence a failure. Explain every intentional behavior change.

## Verification Discovery

Do not assume a language, framework, package manager, build tool, operating
system, or fixed command set.

Discover verification commands from:

```text
README files
AGENTS.md files
build scripts
Makefile or task-runner files
package or workspace manifests
continuous-integration configuration
container definitions
existing developer documentation
recent successful project conventions
```

Prefer existing repository entry points such as a documented `check`, `test`,
`verify`, or `ci` task. Do not invent a command and report it as authoritative.

Separate evidence types:

- focused behavioral tests prove the selected behavior;
- broader tests detect regressions;
- formatting, linting, static analysis, type checking, and builds prove
  repository hygiene;
- packaging or deployment checks prove artifacts;
- smoke tests prove the actual runtime path.

None substitutes for all the others.

Run the narrowest relevant test while developing, then applicable broader checks
before readiness.

Do not claim a command passed unless it ran against the current working state.

When a check fails or cannot run, record:

- the exact command;
- the relevant output;
- the concrete reason;
- what remains unverified;
- whether the failure appears to predate the change;
- the strongest valid evidence obtained instead.

Do not silently skip expensive, platform-specific, network, browser, hardware,
concurrency, security, migration, or destructive-path checks.

## Production-Path Integrity

Production behavior must not depend on test-only code, fabricated state, or
silent fallback paths.

Rules:

- Do not add hidden success fallbacks.
- Do not catch broad failures and return success.
- Do not silently discard contract-required errors.
- Do not bypass authorization, validation, safety, transaction, lifecycle, or
  policy boundaries.
- Do not make tests pass by weakening production checks.
- Do not introduce mutable global state merely for test convenience.
- Do not expose test-only entry points in production builds.
- Do not use shell interpolation for command execution when direct APIs or
  argument vectors are available.
- Do not log secrets, tokens, credentials, or sensitive user data.
- Do not add network access, telemetry, or external calls without explicit
  contract and user value.

If behavior is safety-, security-, privacy-, financial-, destructive-, or
availability-sensitive, negative-path tests are required.

## Interface And Output Policy

Every externally visible interface must define and preserve its applicable
contract:

- accepted input;
- successful output;
- error output;
- status or exit semantics;
- side effects;
- ordering and lifecycle;
- compatibility and versioning;
- cancellation, timeout, and cleanup behavior;
- security and authorization behavior.

Do not report success when the documented action did not occur.

Do not hide failure behind friendly wording, empty output, default objects, or
success status.

When output is machine-readable, preserve stable structure and test it.

When output is human-facing, test critical semantic content and destination
without overfitting incidental formatting.

## Dependencies And Architecture

Dependencies and abstractions must be justified by current behavior.

Before adding a dependency, verify:

- the existing stack cannot reasonably satisfy the requirement;
- the dependency is maintained and compatible with project constraints;
- installation, build, licensing, and deployment effects are understood;
- missing or failed dependency behavior is explicit;
- tests cover the integration boundary.

Do not introduce distributed services, databases, plugin systems, message
brokers, caches, code generators, frameworks, persistence models, or generalized
abstractions unless the selected behavior requires them.

Prefer boring, inspectable code over speculative flexibility.

## File And Data Safety

Any behavior that reads, writes, modifies, deletes, migrates, transmits, or
retains data must be explicit and tested.

Required practices when applicable:

- isolate tests in temporary or disposable locations;
- validate paths and scope before destructive actions;
- use atomic or transactional updates where partial writes are unsafe;
- preserve user data on failure;
- define overwrite, backup, conflict, retention, and cleanup behavior;
- avoid writing outside documented or explicitly selected locations;
- avoid implicit destructive defaults;
- redact sensitive values from logs and reports;
- test permission, missing-resource, corrupt-data, and interruption failures.

Do not assume the current directory, home directory, network, device, or shared
environment is safe for tests.

## Version-Control Policy

When the repository uses version control, make small, reviewable changes.

Rules:

- Do not work directly on a protected or primary branch unless explicitly
  instructed.
- Inspect branch and working-tree state before editing.
- Preserve dirty-tree and user-created changes.
- Do not overwrite, move, broadly reformat, stage, or commit unrelated work.
- Do not hide broad changes inside a narrow branch or commit.
- Do not commit generated noise, caches, logs, editor files, temporary files, or
  unrelated dependency churn.

Follow repository conventions for branch and commit naming. When none exist,
use specific names such as:

```text
feat/<observable-behavior>
fix/<specific-defect>
test/<behavior-under-test>
docs/<contract-change>
refactor/<bounded-structural-change>
chore/<maintenance-objective>
```

Use specific, imperative commit subjects. Add a body when reason, risk, or
user-visible behavior is not obvious.

A failing-test commit is allowed when intentional. It does not complete the
feature; a later implementation change must make the test pass.

## Review Gate

A review unit must contain one coherent behavior, fix, test improvement,
documentation contract, or bounded refactor.

Review must verify:

- contract and roadmap traceability;
- meaningful focused evidence;
- real production-path exercise;
- applicable broader checks;
- synchronized behavior-facing documentation;
- focused diff with no unrelated work;
- no placeholders, fabricated success, dead code, or unsupported claims;
- security, compatibility, migration, and data risks where applicable.

A review finding must name its owning return phase:

```text
contract
design
test
build
documentation
security
```

Do not silently repair a material finding under review without recording the
return phase and rerunning affected gates.

## Stop Conditions

Stop implementation and record the exact condition when:

- requested behavior is absent from or conflicts with the governing contract;
- applicable instructions or authorities materially disagree;
- no bounded roadmap slice or equivalent scope can be selected;
- continuing would overwrite or entangle unrelated user work;
- the intended production boundary cannot be exercised meaningfully;
- required verification cannot run and no valid substitute proves behavior;
- a safety, security, privacy, data-integrity, or compatibility invariant would
  be bypassed;
- broad unrelated changes would be required merely to appear complete;
- the environment lacks a required dependency, credential, device, service, or
  permission.

Leave the repository in the most functional and testable state possible. Do not
fabricate progress to avoid reporting a blocker.

## Definition Of Ready

A cycle is ready for commit or handoff only when all applicable conditions hold:

- the intended contract is clear and current;
- one roadmap slice or bounded request is selected;
- a meaningful test covers the real boundary;
- useful failure was observed before implementation when practical;
- production code implements the selected behavior;
- the real path was exercised;
- outputs, statuses, side effects, lifecycle, and negative paths match the
  contract;
- focused tests pass;
- applicable broader tests and repository checks pass;
- behavior-facing documentation is synchronized or explicitly not applicable;
- the diff is focused and contains no placeholder success paths;
- unrelated user changes remain untouched;
- review passes.

Not ready:

- documentation when implementation was requested;
- interfaces, schemas, packages, routes, handlers, components, or configuration
  without a working path;
- unavailable behavior presented as the real feature;
- mock-only or scaffolding-only tests;
- passing compilation, lint, type checking, schema validation, or builds without
  behavioral evidence;
- code inspection without execution;
- claims based on commands not run.

## Completion Report

Report only:

- cycle ID, mode, roadmap slice, and final phase;
- exact behavior implemented or documentation changed;
- materially changed files;
- acceptance criteria and evidence;
- exact commands run and results;
- real-path result;
- known limitations;
- failed, skipped, or unavailable verification;
- intentionally excluded work;
- readiness status using the defined vocabulary.

Do not use promotional claims such as `complete`, `robust`, `production-ready`,
or `fully implemented` unless objective evidence supports them.
