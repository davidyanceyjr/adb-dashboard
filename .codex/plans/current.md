# Active Cycle

- Cycle ID: CYCLE-20260805-M5-S3
- Status: verified
- Mode: feature
- Goal: Analyze one stored APK artifact locally through the HTTP API and persist the latest ready analysis.
- Roadmap slice: M5-S3: Artifact Analysis API
- Branch or work context: `feat/m5-s2-artifact-catalog`
- Specification anchors: `CAP-019`, `CAP-020`, `AC-019-001`, `AC-019-002`, `AC-019-003`, ready-analysis portion of `AC-020-002`, `INV-SEC-001`, `INV-SEC-003`, `INV-DATA-004`, `DATA-006`, `DATA-007`
- Acceptance boundary: HTTP request through the running server with isolated artifact storage and deterministic fake `aapt`.
- Focused test command: `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./tests/cli -run 'TestM5S3ArtifactAnalysisAPI'`
- Real-path command or procedure: Start the built server with isolated `--data-dir` and fake `aapt`; upload an artifact; `POST /api/v1/artifacts/{artifactId}/analyze`; inspect JSON, `metadata.json`, artifact detail, unknown artifact, missing tool, nonzero exit, timeout, oversized stdout, output without `package: name=...`, rejected Host/Origin, fake-tool logs, and absence of ADB/browser markers.
- Broad verification commands: `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./...`; `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go vet ./...`; `git diff --check`
- Current gate: review
- Current phase: ready
- Blocker: none
- Next phase: commit.

## In Scope

- `POST /api/v1/artifacts/{artifactId}/analyze`.
- Bounded `aapt dump badging STORED_APK_PATH` execution with 5-second timeout and 1 MiB stdout limit.
- Parsing `packageName`, optional version, SDK, label, launchable activity, and warning metadata from `aapt dump badging`.
- Atomic `metadata.json` replacement with the latest ready analysis.
- Artifact detail response includes ready analysis when present.
- Unknown artifact, missing `aapt`, nonzero exit, timeout, oversized stdout, output without `package: name=...`, and Host/Origin rejection.

## Out Of Scope

- Browser artifact analysis UI.
- APK install, signing verification, malware analysis, network lookups, reports, jobs, artifact deletion, and ADB/device mutation.
- Release.

## Phase Results

Phase: discovery
Result: DISCOVERY READY
Evidence:
- `git status --short --branch` showed branch `feat/m5-s2-artifact-catalog` with pre-existing dirty `docs/SPECIFICATION.md` and `docs/roadmap.md`.
- `.codex/plans/current.md` recorded inactive `CYCLE-20260804-M5-S2`, committed, with next eligible slice `M5-S3`.
- `docs/roadmap.md` contains accepted slice `M5-S3: Artifact Analysis API`.
Changed:
- none
Next: contract
Blocker: none

Phase: contract
Result: CONTRACT READY
Evidence:
- `docs/SPECIFICATION.md` `CAP-019` and `AC-019-001` through `AC-019-003` define route, output fields, local `aapt` command, timeout, 1 MiB stdout limit, persistence, failure cases, and security ordering.
- `docs/SPECIFICATION.md` `AC-020-002` defines detail responses including latest analysis only when present.
- Acceptance boundary selected: HTTP request through running server with isolated artifact storage and deterministic fake `aapt`.
Changed:
- none
Next: red
Blocker: none

Phase: red
Result: RED CONFIRMED
Evidence:
- `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./tests/cli -run 'TestM5S3ArtifactAnalysisAPI'` first exited `2` because the new test compared a struct containing `[]string`; this was a test authoring error, not useful red evidence.
- After fixing the assertion, `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./tests/cli -run 'TestM5S3ArtifactAnalysisAPI'` exited `1`; `POST /api/v1/artifacts/{artifactId}/analyze` returned HTTP `404` with code `not_found`.
Changed:
- `tests/cli/m1_s1_cli_test.go`
Next: build
Blocker: none

Phase: build
Result: BUILD APPLIED
Evidence:
- Production route `POST /api/v1/artifacts/{artifactId}/analyze` added through existing `/api/v1/artifacts/` server mux and security middleware.
- Production analysis path executes `aapt dump badging STORED_APK_PATH`, bounds stdout to 1 MiB, uses a 5-second timeout, discards stderr, parses documented fields, stores ready analysis atomically in `metadata.json`, and leaves existing ready analysis unchanged on failures.
- `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./tests/cli -run 'TestM5S3ArtifactAnalysisAPI'` exited `0`.
Changed:
- `cmd/adb-dashboard/main.go`
- `tests/cli/m1_s1_cli_test.go`
Next: documentation
Blocker: none

Phase: documentation
Result: DOCS SYNCED
Evidence:
- `docs/MANUAL_TESTING.md` now includes M5-S3 artifact analysis API manual checks, expected response fields, persisted metadata, failure cases, and security rejection notes.
- No executable documentation validation command exists for this Markdown file; focused implementation command was rerun after documentation edits and exited `0`.
Changed:
- `docs/MANUAL_TESTING.md`
Next: implementation-ready-for-test
Blocker: none

Phase: implementation-ready-for-test
Result: IMPLEMENTATION READY FOR TEST
Evidence:
- Production and focused test files for M5-S3 are changed.
- Focused command after build and documentation edits: `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./tests/cli -run 'TestM5S3ArtifactAnalysisAPI'` exited `0`.
- `git diff --check` exited `0`.
Changed:
- `.codex/plans/current.md`
- `cmd/adb-dashboard/main.go`
- `docs/MANUAL_TESTING.md`
- `tests/cli/m1_s1_cli_test.go`
Next: test gate focused-green, real-path exercise, negative-path checks, and broad checks.
Blocker: none

Phase: green
Result: GREEN VERIFIED
Evidence:
- Focused command `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./tests/cli -run 'TestM5S3ArtifactAnalysisAPI'` exited `0`.
- Real-path command built `.codex/cache/adb-dashboard-m5s3`, started `serve --listen 127.0.0.1:0 --data-dir <isolated> --temp-dir <isolated> --no-open`, uploaded an artifact with `curl`, requested analysis with fake `aapt`, observed `artifact.analysisStatus=ready`, `analysis.tool=aapt`, `analysis.packageName=com.example.realpath`, detail response with matching analysis, unknown artifact `404`, Host rejection `403`, no host path markers in responses, and exact fake-tool log `dump badging <isolated>/data/artifacts/<artifactId>/original.apk`.
- Broad command `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./...` exited `0`.
- Broad command `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go vet ./...` exited `0`.
- Diff command `git diff --check` exited `0`.
Changed:
- none
Next: review
Blocker: none

Phase: review
Result: REVIEW PASSED
Evidence:
- `AC-019-001` is covered by focused and real-path HTTP/filesystem evidence for successful analysis, parsed fields, persisted ready metadata, detail response analysis, exact `aapt dump badging` invocation, and no response leakage of host paths, stderr, environment values, or tokens.
- `AC-019-002` is covered by focused test evidence for unknown artifact, missing `aapt`, nonzero exit, 5-second timeout, oversized stdout, output without `package: name=...`, and no replacement of prior ready analysis on failed analysis.
- `AC-019-003` is covered by focused and real-path evidence for rejected Host/Origin requests before additional host-tool execution.
- Ready-analysis portion of `AC-020-002` is covered by focused and real-path detail responses returning latest `analysis` only after ready analysis exists.
- Diff reviewed for `.codex/plans/current.md`, `cmd/adb-dashboard/main.go`, `docs/MANUAL_TESTING.md`, and `tests/cli/m1_s1_cli_test.go`; no placeholders, hard-coded production success, test-only production hooks, new dependencies, ADB/install invocation, network lookup, host path disclosure in responses, unsupported browser UI, or unrelated cleanup found.
- Pre-existing dirty `docs/SPECIFICATION.md` and `docs/roadmap.md` remain unstaged for this commit.
Changed:
- `.codex/plans/current.md`
Next: commit
Blocker: none
