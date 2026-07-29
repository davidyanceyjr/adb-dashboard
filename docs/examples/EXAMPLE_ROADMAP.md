# Example Roadmap: File-Backed Notes CLI

> This is an illustrative roadmap, not a required project shape.

## Status

- Roadmap status: Accepted
- Roadmap version: 1
- Specification source: `docs/examples/EXAMPLE_SPECIFICATION.md`
- Specification version: 1
- Current milestone: M1
- Next eligible slice: M1-S1
- Last reviewed: 2026-07-29

## Milestone M1: Note Creation

### Slice M1-S1: Create one protected note

- Status: accepted
- Mode: feature
- Purpose: Deliver the first usable note-creation behavior.
- Specification references: `CAP-001`, `AC-001`, `AC-002`, `AC-003`,
  `INV-001`, `INV-002`
- Observable result: A user can create one note through the real command and
  receives its path; invalid and conflicting requests make no data change.
- Primary acceptance boundary: process invocation and isolated filesystem.
- Expected red or baseline-green evidence: The focused process test fails
  because the production `create` behavior is absent or does not create the
  required file.
- Focused verification command or deterministic discovery rule: Resolve the
  repository's focused command for the process-level note-creation test and
  record it in `.codex/plans/current.md` before implementation.
- Real-path exercise: Build or invoke the real command in a temporary directory,
  create a note, inspect stdout/stderr/status and file content, then rerun against
  the existing destination and confirm preservation.
- Broad verification: Run the repository's standard test, lint, formatting,
  type-check, and build commands that apply.
- Required environment: Project runtime and writable temporary filesystem.
- Dependencies: None.
- In scope:
  - Title validation.
  - Filename normalization required by CAP-001.
  - Atomic creation.
  - Existing-file protection.
  - Process-level acceptance tests.
- Out of scope:
  - Listing, editing, deleting, encryption, synchronization, prompts, packaging,
    release, and unrelated refactoring.
- Risks:
  - Path traversal through title normalization.
  - Partial files after write failure.
- Stop conditions:
  - Filename normalization remains ambiguous.
  - The production process boundary cannot be executed.
  - Required atomic behavior is unsupported and the contract does not define a
    fallback.
- Documentation synchronization: Update user-facing command documentation if it
  exists; otherwise record `DOCS NOT REQUIRED` with reason.
- Exit gate: AC-001 through AC-003 have `GREEN VERIFIED` evidence through the
  real process and filesystem boundary, applicable broad checks pass,
  documentation is synchronized or not required with reason, and review passes.
- Completion evidence reference: `.codex/plans/current.md` and
  `.codex/cycles/history.md`.
