# Active Cycle

- Status: inactive
- Last cycle ID: CYCLE-20260803-M5-S1
- Last mode: feature
- Last roadmap slice: M5-S1: Artifact Upload API
- Last result: verified
- Last final phase: ready
- Next eligible slice: M5-S2
- Blocker: none

## Last Evidence

- Focused red: `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./tests/cli -run 'TestM5S1ArtifactUploadAPI'` exited `1`; upload cases reached the running server and returned HTTP `404` with code `not_found` because `POST /api/v1/artifacts` was absent.
- Focused test: `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./tests/cli -run 'TestM5S1ArtifactUploadAPI'` exited `0`, including valid upload persistence, missing, duplicate, empty, non-APK, non-ZIP, oversized, unsupported media type, short-body cleanup, storage failure, and Host/Origin security rejection.
- Real path: built `.codex/cache/adb-dashboard-m5s1`, started `serve --listen 127.0.0.1:0 --data-dir <isolated> --no-open`, uploaded a disposable APK-like ZIP, observed `upload_status=201 artifact_id=eLerbEVgGWmfFiy62QYrqg`, restarted with the same data directory, observed `restart_persistence=ok`, exercised unsupported media, invalid upload, short-body, and foreign Origin cases, and observed `artifact_dirs=1 temp_files=0`.
- Broad test: `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./...` exited `0`.
- Race test: `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -race -count=1 ./...` exited `1` once in existing `TestM4S3BrowserPackageInventoryView/loads_scopes_and_empty_state_from_backend` due fake ADB command ordering under race instrumentation; narrow rerun of that test exited `0`; full race rerun exited `0`.
- Static check: `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go vet ./...` exited `0`.
- Diff whitespace check: `git diff --check` exited `0`.
- Documentation sync: `docs/MANUAL_TESTING.md` documents M5-S1 artifact upload behavior; `docs/roadmap.md` marks M5-S1 verified and M5-S2 next eligible.
- Verification scope note: no further full-suite reruns were performed after the 2026-08-03 broad checks; future non-milestone cycles should prefer narrow checks until a milestone marker requires a full suite.

## Last Review

Phase: review
Result: REVIEW PASSED
Evidence:
- `AC-018-001` is covered by HTTP/filesystem evidence for valid APK-like ZIP upload, HTTP `201` metadata, SHA-256 and byte size, stored `original.apk`, stored `metadata.json`, no ADB/browser side effects, and persistence after server restart.
- `AC-018-002` is covered by HTTP/filesystem evidence for missing, duplicate, empty, non-APK, non-ZIP, oversized, storage-failing, unsupported media type, and short-body uploads with documented status/error codes and no partial artifact directory or temporary upload files.
- `AC-018-003` is covered by Host and Origin rejection evidence returning the standard security envelope before artifact storage.
- Production diff is limited to the `POST /api/v1/artifacts` route, artifact streaming validation, atomic file/metadata persistence, and cleanup helpers in the existing server entry point.
- Test diff is limited to M5-S1 process/HTTP/filesystem integration coverage and reusable HTTP helpers needed for multipart and raw short-body requests.
- Documentation and roadmap are synchronized with verified M5-S1 behavior.
- No placeholders, test-only production hooks, new dependencies, external network calls, ADB/install invocation, host path disclosure, or unrelated cleanup were found in review.
Changed:
- `.codex/plans/current.md`
- `.codex/cycles/history.md`
- `cmd/adb-dashboard/main.go`
- `docs/MANUAL_TESTING.md`
- `docs/roadmap.md`
- `tests/cli/m1_s1_cli_test.go`
Next: commit if explicitly requested; otherwise start next eligible slice `M5-S2` with narrow verification by default.
Blocker: none
