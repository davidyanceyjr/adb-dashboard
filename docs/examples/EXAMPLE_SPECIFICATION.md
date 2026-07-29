# Example Specification: File-Backed Notes CLI

> This is an illustrative contract, not a required project shape.

## Status

- Contract status: Accepted
- Specification version: 1
- Product or system: Example Notes CLI
- Version or milestone: M1
- Owners or decision authority: Example maintainer
- Last reviewed: 2026-07-29

## Purpose

Allow a user to create a plain-text note in an explicitly selected directory
without overwriting existing data.

## Scope

### In scope

- Create one note from command-line input.
- Return the created path.
- Protect existing files.

### Out of scope

- Editing notes.
- Listing notes.
- Encryption.
- Synchronization.
- Interactive prompts.

## Actors And Public Boundaries

| Actor or caller | Public interface | Trust or ownership boundary | Relevant capabilities |
|---|---|---|---|
| Local user | `notes create` process invocation | User-owned filesystem | CAP-001 |

## Global Invariants

- `INV-001`: The command must not modify an existing note unless a future
  contract explicitly introduces overwrite behavior.
- `INV-002`: Diagnostics are written separately from requested output.

## Behavioral Contracts

### CAP-001: Create a note

**Contract status:** Implementation-ready

**Goal:**

Create one UTF-8 note in a selected directory and report its path.

**Actors or callers:**

- Local user.

**Inputs and preconditions:**

- Required title containing at least one non-whitespace character.
- Required body supplied as one argument or standard input.
- Destination directory exists and is writable.
- Filename is the normalized title plus `.txt`.

**Successful behavior:**

- Exactly one new file is created.
- File content is the supplied body followed by one newline.
- The command reports the created path and exits successfully.

**Output, response, event, or visible state:**

- Stdout: absolute created path followed by one newline.
- Stderr: empty.
- Exit status: `0`.

**Required side effects:**

- Create exactly one file in the destination directory.

**Forbidden side effects:**

- Do not modify existing files.
- Do not create directories.
- Do not write outside the destination directory.

**Errors and negative behavior:**

- `ERR-001`: Empty title reports an input error, creates no file, and exits
  nonzero.
- `ERR-002`: Existing destination reports a conflict, preserves the file, and
  exits nonzero.
- `ERR-003`: Unwritable destination reports a filesystem error, creates no file,
  and exits nonzero.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:**

- Validation occurs before file creation.
- The file is written atomically using a temporary file in the destination and
  rename where the platform supports it.
- Temporary files are removed after failure.

**Security, authorization, privacy, and data safety:**

- The normalized path must remain inside the selected destination directory.

**Compatibility and versioning:**

- Filename normalization and stdout shape are stable for specification version 1.

#### Acceptance criteria

- `AC-001`: Given a writable empty destination and valid title and body, when
  creation is requested, then exactly one file with expected content is created,
  stdout contains its path, stderr is empty, and exit status is `0`.
- `AC-002`: Given the target file already exists, when creation is requested,
  then the original content remains unchanged, no additional file remains, and
  the command reports a conflict with nonzero status.
- `AC-003`: Given an empty title, when creation is requested, then no file is
  created and the command reports an input error with nonzero status.

#### Observable acceptance boundary

- Primary boundary: process invocation and isolated filesystem.
- Focused evidence expected: process-level acceptance tests for AC-001 through
  AC-003.
- Real-path exercise: invoke the built command in a temporary directory and
  inspect output, status, files, and content.
- Required environment: local filesystem and project runtime.
- Permitted deterministic substitutes: temporary directories.
- Evidence that does not count: parser unit test alone, file-existence-only test,
  or hard-coded success output.

#### Blocking open questions

- None.
