---
name: cycle-review
description: Review one completed cycle for contract traceability, meaningful evidence, focused scope, no-SLOP compliance, safety, documentation sync, and commit readiness. Trigger after green verification or documentation completion. Do not treat lint or formatting as functional proof.
---

# Cycle Review

Follow `AGENTS.md`. Own readiness assessment for one cycle.

## Inputs

- active cycle state;
- specification and roadmap slice;
- diff and changed files;
- focused tests and real-path evidence;
- broader verification results;
- documentation result;
- applicable safety, security, data, compatibility, and migration evidence.

## Procedure

1. Confirm exactly one bounded slice was attempted.
2. Trace each acceptance criterion to production code, meaningful tests, and
   evidence.
3. Inspect the complete diff, including untracked and generated files.
4. Reject placeholders, fabricated success, dead paths, test-only production
   hooks, unsupported completion claims, and unrelated churn.
5. Confirm the real path was exercised.
6. Confirm focused and applicable broad checks ran against the current state.
7. Confirm errors, negative paths, cleanup, compatibility, and side effects.
8. Confirm behavior-facing documentation is synchronized or genuinely not
   applicable.
9. Confirm unrelated user work remains untouched.
10. Identify security, privacy, destructive, migration, dependency, and data
    risks where applicable.
11. Run or rerun cheap invalidated hygiene checks when useful.
12. Assign every material finding to an owning return phase.

## Finding Format

```text
Severity: blocker | major | minor
Owner phase: contract | design | test | build | documentation | security
Evidence:
Impact:
Required correction:
Invalidated gates:
```

Do not repair a material finding silently. Return it to the owning phase and
rerun invalidated gates.

## Readiness Rules

`REVIEW PASSED` requires:

- contract and roadmap traceability;
- meaningful focused evidence;
- real-path exercise;
- applicable broad checks;
- synchronized documentation or explicit not-applicable reason;
- focused diff;
- no unsupported completion claims;
- no unresolved blocker or major finding.

Formatting, linting, static analysis, type checking, builds, schema validation,
and coverage are hygiene evidence, not functional proof.

## Result

Return one:

```text
REVIEW PASSED
REVIEW FAILED
REVIEW BLOCKED
```

Use this shape:

```text
Phase: review
Result:
Evidence:
- criteria-to-evidence summary
- diff and repository checks
Changed:
- review-only fixes, normally none
Next: ready | owning return phase | blocked
Blocker: none | exact blocker
```
