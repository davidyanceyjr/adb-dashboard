# Active Cycle

- Cycle ID: CYCLE-20260805-M5-S4
- Status: active
- Mode: feature
- Goal: Make local artifact upload and catalog inspection usable from the browser.
- Roadmap slice: M5-S4: Browser Artifact Upload And Catalog
- Branch or work context: `feat/m5-s4-browser-artifacts`
- Specification anchors: `CAP-018`, `CAP-020`, `INV-FRONTEND-001`, `INV-SEC-004`, `INV-DATA-004`
- Acceptance criteria: `AC-018-001`, `AC-018-002`, `AC-020-001`, `AC-020-002`, `AC-020-004`
- Acceptance boundary: Browser interaction through the running server with isolated artifact storage.
- In scope: browser artifact upload control; catalog refresh and artifact detail view; pending/no-analysis state; upload and catalog failure states; unsupported-control absence.
- Out of scope: browser analysis trigger; artifact deletion; install; signing verification; reports; jobs; external services; broad UI redesign.
- Focused test command: `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./tests/cli -run 'TestM5S4BrowserArtifactUploadAndCatalog'`
- Real-path command or procedure: build `.codex/cache/adb-dashboard-m5s4`; start `serve --listen 127.0.0.1:0 --data-dir <isolated> --temp-dir <isolated> --no-open`; open the served shell; upload a disposable APK-like ZIP through the browser control; inspect catalog and detail; restart and inspect persistence; trigger invalid upload and catalog failure states; inspect filesystem side effects.
- Broad verification commands: `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./...`; `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go vet ./...`; `git diff --check`
- Current gate: Review
- Current phase: ready
- Blocker: none
- Next phase: commit

## Phase Results

Phase: discovery
Result: DISCOVERY READY
Evidence:
- `git status --short --branch` on branch `main` showed clean working tree and `main...origin/main [ahead 1]`.
- Created local branch `feat/m5-s4-browser-artifacts` before behavior edits.
- `.codex/plans/current.md` named next eligible slice `M5-S4`; `.codex/cycles/history.md` records `M5-S1`, `M5-S2`, and `M5-S3` as committed.
- `docs/roadmap.md` header still says `Next eligible slice: M5-S2`, but slice history and active state show dependencies through `M5-S3` satisfied; treated as stale roadmap metadata for this cycle.
Changed:
- `.codex/plans/current.md`
Next: contract
Blocker: none

Phase: contract
Result: CONTRACT READY
Evidence:
- `docs/SPECIFICATION.md` is Accepted version `1.3.0`; `CAP-018` and `CAP-020` define browser/API artifact upload, catalog, detail, pending/no-analysis, and failure behavior.
- `docs/roadmap.md` M5-S4 is accepted and references `AC-018-001`, `AC-018-002`, `AC-020-001`, `AC-020-002`, `AC-020-004`, `INV-FRONTEND-001`, `INV-SEC-004`, and `INV-DATA-004`.
- Acceptance boundary is browser interaction through the running server with isolated artifact storage.
- Expected red evidence is a focused browser test failing because the current shell lacks artifact upload/catalog workflow.
Changed:
- `.codex/plans/current.md`
Next: red
Blocker: none

Phase: design
Result: DESIGN NOT REQUIRED
Evidence:
- Existing inline browser shell, deterministic Node browser harness, and production artifact APIs provide the obvious implementation and test locations.
Changed:
- none
Next: red
Blocker: none

Phase: red
Result: RED CONFIRMED
Evidence:
- `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./tests/cli -run 'TestM5S4BrowserArtifactUploadAndCatalog'` exited `1`.
- Intended boundary reached: served shell script executed through deterministic browser harness against the running server.
- Observed missing behavior: rendered shell lacked `artifacts-status=artifacts: 0`; invalid browser upload action lacked `artifact-upload-status=upload: unavailable`.
- One initial assertion issue was corrected before confirming red; the first attempt falsely matched existing package `installer` text as an install control.
Changed:
- `.codex/plans/current.md`
- `tests/cli/m1_s1_cli_test.go`
Next: build
Blocker: none

Phase: build
Result: BUILD APPLIED
Evidence:
- Production browser shell in `cmd/adb-dashboard/main.go` now renders artifact upload, refresh, catalog, and first-item detail controls.
- Browser upload uses `FormData` against `POST /api/v1/artifacts`; catalog refresh uses `GET /api/v1/artifacts`; detail rendering uses the latest backend catalog item and clears stale detail on refresh or catalog failure.
- Failure handling renders `upload: unavailable` for failed browser upload and `artifacts: unavailable` with cleared list/detail when catalog refresh fails.
- Development focused command `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./tests/cli -run 'TestM5S4BrowserArtifactUploadAndCatalog'` exited `0`.
- `git diff --check` exited `0`.
Changed:
- `.codex/plans/current.md`
- `cmd/adb-dashboard/main.go`
- `tests/cli/m1_s1_cli_test.go`
Next: focused-green
Blocker: none

Phase: implementation gate
Result: IMPLEMENTATION READY FOR TEST
Evidence:
- M5-S4 browser production path exists and the focused test passed once during build.
- Formal Test gate evidence is still required for focused green, real-path exercise, negative paths, applicable broad checks, and documentation synchronization.
Changed:
- `.codex/plans/current.md`
- `cmd/adb-dashboard/main.go`
- `tests/cli/m1_s1_cli_test.go`
Next: focused-green
Blocker: none

Phase: focused-green
Result: FOCUSED GREEN
Evidence:
- `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./tests/cli -run 'TestM5S4BrowserArtifactUploadAndCatalog'` exited `0`; final rerun reported `ok adb-dashboard/tests/cli 11.293s`.
- The focused browser test exercised served shell JavaScript through the deterministic Node browser harness against a running server with isolated artifact storage.
- Evidence covers `AC-018-001`, `AC-018-002`, `AC-020-001`, `AC-020-002`, and `AC-020-004` for browser upload, catalog refresh, detail rendering, persistence after restart, pending/no-analysis state, invalid upload state, corrupt catalog state, and artifact-workflow unsupported-control absence.
Changed:
- `tests/cli/m1_s1_cli_test.go`
Next: real-path-exercise
Blocker: none

Phase: real-path-exercise
Result: REAL PATH VERIFIED
Evidence:
- Built `.codex/cache/adb-dashboard-m5s4` with `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go build -o .codex/cache/adb-dashboard-m5s4 ./cmd/adb-dashboard`; command exited `0`.
- Started `.codex/cache/adb-dashboard-m5s4 serve --listen 127.0.0.1:0 --data-dir /tmp/adb-dashboard-m5s4-realpath-WRStSP/data --temp-dir /tmp/adb-dashboard-m5s4-realpath-WRStSP/temp --no-open`; server reported `addr=127.0.0.1:41817`.
- Uploaded disposable APK-like ZIP with `curl -F artifact=@.../browser-upload.apk` to `POST /api/v1/artifacts`; response status was `201` with `originalName:"browser-upload.apk"`, `sizeBytes:17`, SHA-256 `f0e0c665d232f3fbb045f4da30bafac087cda5bb64223cc011185d943d0e7bc9`, and `analysisStatus:"pending"`.
- `GET /api/v1/artifacts` returned status `200` with `count:1`; `GET /api/v1/artifacts/22hGKqfIbrF7p15dHS1dDA` returned status `200` with pending artifact detail.
- Filesystem side effects under isolated data dir were `artifacts/22hGKqfIbrF7p15dHS1dDA/original.apk` at 17 bytes and `metadata.json` at 290 bytes.
- Restarted server against the same data dir at `127.0.0.1:33063`; catalog and detail returned status `200` from persisted metadata.
- Both smoke servers stopped cleanly on interrupt with `server stopped signal=interrupt`.
Changed:
- none
Next: negative-path-checks
Blocker: none

Phase: negative-path-checks
Result: NEGATIVE PATHS VERIFIED
Evidence:
- Real-path invalid upload using `filename=not-an-apk.txt` and `Content-Type: text/plain` returned status `400` with error code `invalid_artifact_upload`.
- Real-path corrupt metadata check overwrote the isolated artifact `metadata.json` with malformed JSON; `GET /api/v1/artifacts` returned status `500` with error code `artifact_catalog_unavailable`; metadata was restored afterward.
- Focused browser test also asserts invalid upload renders `artifact-upload-status=upload: unavailable`, corrupt catalog renders `artifacts-status=artifacts: unavailable`, stale artifact text is cleared, and no artifact-workflow host paths or unsupported install/analyze/delete/shell controls are rendered.
Changed:
- none
Next: applicable-broad-checks
Blocker: none

Phase: applicable-broad-checks
Result: BROAD CHECKS PASSED
Evidence:
- Initial `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./...` failed because legacy browser tests still forbade the now-supported `artifacts` shell section and had deterministic browser-action timing/order assumptions invalidated by the additional startup catalog fetch.
- Test-only repairs removed stale `artifacts` forbidden tokens from pre-M5 helpers, made browser-action timeout derive from scheduled log delay, waited for affected package/device failure-detail actions to settle, and relaxed one incidental ADB marker order assertion to require the three observable package-scope commands.
- Final `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./...` exited `0`; output included `? adb-dashboard/cmd/adb-dashboard [no test files]` and `ok adb-dashboard/tests/cli 124.308s`.
- Final `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go vet ./...` exited `0`.
- Final `git diff --check` exited `0`.
Changed:
- `tests/cli/m1_s1_cli_test.go`
Next: test-ready-for-review
Blocker: none

Phase: test gate
Result: TEST READY FOR REVIEW
Evidence:
- Focused browser boundary, real-path HTTP/filesystem smoke, negative paths, and applicable broad checks passed against the current working tree.
- No production code was changed during the Test gate; Test gate edits were limited to stale/brittle browser test expectations in `tests/cli/m1_s1_cli_test.go`.
Changed:
- `.codex/plans/current.md`
- `tests/cli/m1_s1_cli_test.go`
Next: docs-review
Blocker: none

Phase: documentation
Result: DOCS SYNCED
Evidence:
- `docs/MANUAL_TESTING.md` Browser UI section now documents visible artifact upload, refresh, and detail controls; pending catalog/detail state; invalid upload state; catalog failure state; restart persistence; and artifact-workflow exclusions for install, device mutation, shell, analysis, deletion, and host path disclosure.
- `git diff --check` after the documentation update exited `0`.
Changed:
- `docs/MANUAL_TESTING.md`
Next: review
Blocker: none

Phase: review
Result: REVIEW PASSED
Evidence:
- One bounded slice reviewed: M5-S4 Browser Artifact Upload And Catalog.
- `AC-018-001` is covered by focused browser upload evidence and real-path `POST /api/v1/artifacts` smoke returning `201`, persisted `original.apk`, persisted `metadata.json`, and restart-visible catalog/detail metadata.
- `AC-018-002` is covered by focused invalid browser upload evidence and real-path invalid upload returning `400 invalid_artifact_upload` without false browser artifact success.
- `AC-020-001` and `AC-020-002` are covered by focused browser catalog/detail evidence plus real-path `GET /api/v1/artifacts` and `GET /api/v1/artifacts/{id}` returning persisted pending metadata with no analysis field.
- `AC-020-004` is covered by browser-shell evidence for upload, catalog, detail, pending/no-analysis, invalid upload, corrupt catalog failure states, and unsupported artifact workflow controls absent.
- Diff is limited to the M5-S4 browser production path in `cmd/adb-dashboard/main.go`, browser/API/filesystem coverage and harness timing fixes in `tests/cli/m1_s1_cli_test.go`, manual behavior documentation, and cycle state.
- Review found no placeholders, fabricated production success, test-only production hooks, new dependencies, ADB/install invocation for artifact workflow, external network lookup, host path disclosure in browser artifact output, or unrelated cleanup.
Changed:
- `.codex/plans/current.md`
Next: ready
Blocker: none

Phase: cycle
Result: CYCLE READY
Evidence:
- Focused, real-path, negative-path, broad-check, documentation, and review evidence are recorded for M5-S4.
Changed:
- `.codex/plans/current.md`
Next: commit
Blocker: none
