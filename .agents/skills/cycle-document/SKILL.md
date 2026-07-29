---
name: cycle-document
description: Synchronize durable technical documentation with an accepted contract or verified implementation cycle. Trigger on documentation-only cycles, public behavior changes, API/config/error changes, examples, migration notes, or post-green documentation sync. Do not invent behavior or evidence.
---

# Cycle Documentation

Follow `AGENTS.md`. Document established contract and verified facts. Do not
invent requirements, architecture, commands, test results, or operational
procedures.

## Modes

```text
documentation-only
post-green-sync
```

Contract decisions belong to `cycle-contract`. This skill may write the accepted
contract into durable documentation, but it must not silently decide missing
behavior.

## Procedure

1. Identify the authoritative document and intended audience.
2. Compare current documentation with the accepted contract and verified
   behavior.
3. Update only affected sections.
4. Preserve stable terminology and acceptance-criterion references.
5. Update inputs, outputs, errors, side effects, configuration, examples,
   compatibility, migration, and operational notes only when affected.
6. Mark unavailable, unverified, deprecated, or future behavior explicitly.
7. Validate executable examples, links, schemas, or generated documentation
   where repository tooling exists.
8. Avoid duplicating the same source-of-truth content across documents.

## Documentation-Only Cycles

A documentation-only cycle may end with specified or synchronized behavior. It
must not claim production implementation or functional verification.

## Post-Green Sync

Use test evidence and observed behavior, not implementation intent.

If documentation conflicts with verified implementation, route to `contract`
rather than silently changing whichever side is easier.

## Result

Return one:

```text
DOCS SYNCED
DOCS NOT REQUIRED
DOCS BLOCKED
```

Use this shape:

```text
Phase: documentation
Result:
Evidence:
- authoritative sections reviewed
- validation commands or not-applicable reason
Changed:
- documentation files, or none
Next: review | contract | blocked
Blocker: none | exact conflict or missing fact
```
