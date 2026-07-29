# Active Cycle

- Cycle ID: CYCLE-20260729-M1-S5
- Mode: feature
- Goal: Implement browser security bootstrap for current loopback API routes.
- Roadmap slice: M1-S5 Browser Security Bootstrap
- Branch or work context: Git repository on `main`, aligned with
  `origin/main` and `origin/HEAD` at
  `683172071ceadf1c9eeb096bbe490c42dba19449`; working tree had only prior
  `.codex/plans/current.md` handoff metadata drift before this cycle state was
  written.
- Specification anchors: `CAP-008`, `CAP-009`, `INV-SEC-001`,
  `INV-SEC-003`, `INV-SEC-004`, `INV-NIY-001-C`
- Acceptance criteria: `AC-008-001`, `AC-008-002`, `AC-008-003`,
  `AC-008-004`, `AC-009-004`
- Acceptance boundary: HTTP request through production routing.
- In scope: `GET /api/v1/bootstrap`; per-process `csrfToken` and
  `webSocketToken`; host, absolute-form host, and Origin rejection for current
  `/api/v1` routes; token non-disclosure in current status, diagnostics, and
  visible surfaces available in this slice.
- Out of scope: consuming tokens for future mutating requests; WebSocket
  authorization; local-user authentication; browser UI rendering; ADB/device
  behavior; request-correlation IDs beyond the documented `null` value.
- Focused test command:
  `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./...`
- Real-path command or procedure: Build `./cmd/adb-dashboard` to an isolated
  temporary path, start `serve --listen 127.0.0.1:0 --no-open`, request
  `/api/v1/bootstrap` and `/api/v1/status` with accepted loopback headers,
  restart and compare token values, then request current `/api/v1` routes with
  rejected Host, absolute-form URL host, and Origin inputs.
- Broad verification commands:
  `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -race -count=1 ./...`;
  `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go vet ./...`
- Current phase: ready
- Blocker: none
- Next phase: commit when authorized; otherwise M1-S6 is next eligible.

## Phase Results

Phase: discover
Result: DISCOVERY READY
Evidence:
- `git rev-parse HEAD origin/main origin/HEAD` returned
  `683172071ceadf1c9eeb096bbe490c42dba19449` for all three refs.
- `git status --porcelain=v1` showed only `.codex/plans/current.md` modified
  before this cycle state rewrite.
- `docs/roadmap.md` marks `M1-S5` accepted with one HTTP routing boundary,
  expected red evidence, real-path exercise, scope, risks, stop conditions, and
  binary exit gate.
- `docs/SPECIFICATION.md` marks `CAP-008` implementation-ready with acceptance
  criteria `AC-008-001` through `AC-008-004`.
- Repository verification entry points discovered from `go.mod`, existing
  tests, and prior cycle evidence.
Changed:
- `.codex/plans/current.md`
Next: contract
Blocker: none

Phase: contract
Result: CONTRACT READY
Evidence:
- Specification anchors: `CAP-008`, `CAP-009`, `INV-SEC-001`,
  `INV-SEC-003`, `INV-SEC-004`, `INV-NIY-001-C`.
- Acceptance criteria: `AC-008-001`, `AC-008-002`, `AC-008-003`,
  `AC-008-004`, `AC-009-004`.
- Acceptance boundary: HTTP request through production routing.
- Expected red evidence: focused HTTP test reaches the running server and fails
  because `/api/v1/bootstrap` is absent and host/origin rejection is not yet
  enforced.
Changed:
- none
Next: red
Blocker: none

Phase: red
Result: RED CONFIRMED
Evidence:
- `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./...`
  exited `1`.
- Failure reached the production HTTP boundary:
  `TestM1S5BootstrapTokensAndSecurityPolicy` requested
  `/api/v1/bootstrap` from the running server and observed HTTP `404` with
  error code `not_found` instead of the documented bootstrap response.
Changed:
- `tests/cli/m1_s1_cli_test.go`
Next: build
Blocker: none

Phase: build
Result: BUILD APPLIED
Evidence:
- Production path added for per-process bootstrap token generation,
  `/api/v1/bootstrap`, and security policy enforcement before current
  `/api/v1` route handlers.
- Review repair narrowed security policy matching to the actual `/api/v1`
  route namespace.
- Security rejection envelope repaired to match the accepted standard message
  in `docs/SPECIFICATION.md`.
- `gofmt -w cmd/adb-dashboard/main.go tests/cli/m1_s1_cli_test.go` exited
  `0`.
Changed:
- `cmd/adb-dashboard/main.go`
Next: green
Blocker: none

Phase: green
Result: GREEN VERIFIED
Evidence:
- `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./...`
  exited `0`.
- Real-path exercise built `./cmd/adb-dashboard` to an isolated temporary path,
  started `serve --listen 127.0.0.1:0 --no-open`, requested
  `/api/v1/bootstrap` with same-origin `Origin`, and observed token lengths
  `csrf_len=43` and `websocket_len=43`.
- Real-path `/api/v1/status` returned running status for
  `127.0.0.1:44541` and did not disclose token field names or token values.
- Real-path rejected foreign Host, foreign absolute-form URL host, and foreign
  Origin requests with HTTP `403` and the documented error codes
  `forbidden_host`, `forbidden_absolute_url_host`, and `forbidden_origin`.
- Real-path restart started a second server at `127.0.0.1:45829`; both
  `csrfToken` and `webSocketToken` changed after restart.
- Real-path logs and forbidden responses did not disclose issued token values.
- `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -race -count=1 ./...`
  exited `0`.
- `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go vet ./...`
  exited `0`.
Changed:
- `tests/cli/m1_s1_cli_test.go`
- `cmd/adb-dashboard/main.go`
Next: documentation
Blocker: none

Phase: documentation
Result: DOCS SYNCED
Evidence:
- `rg -n "M1-S5|M1-S4|Status: verified|Next eligible|security|bootstrap|CAP-008" docs README.md .codex -g '*.md'`
  found the accepted specification and roadmap as the behavior-facing docs.
- `docs/SPECIFICATION.md` already defined the accepted `CAP-008` behavior and
  standard security rejection envelope; no specification change was required.
- `docs/roadmap.md` now marks `M1-S5` as `verified` and sets next eligible
  slice to `M1-S6`.
Changed:
- `docs/roadmap.md`
Next: review
Blocker: none

Phase: review
Result: REVIEW PASSED
Evidence:
- Exactly one slice was attempted: `M1-S5` Browser Security Bootstrap.
- `AC-008-001` is covered by `TestM1S5BootstrapTokensAndSecurityPolicy` and
  real-path bootstrap evidence showing HTTP `200`, `application/json`,
  URL-safe 43-character token values, independent token fields, and
  `statusUrl` `/api/v1/status`.
- `AC-008-002` is covered by focused host, absolute-form URL host, and Origin
  rejection tests plus real-path `403` evidence for all three rejection modes.
- `AC-008-003` is covered by focused restart checks and real-path restart
  evidence showing both token values changed.
- `AC-008-004` is covered by focused status non-disclosure checks and
  real-path status/log/forbidden-response non-disclosure checks.
- `git diff --check` exited `0`.
- Final focused command
  `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./...`
  exited `0`.
- Final broad commands
  `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -race -count=1 ./...`
  and
  `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go vet ./...`
  exited `0`.
- Diff is limited to active cycle state, server/bootstrap/security behavior,
  focused HTTP tests, and roadmap status; accidental M1-S1 roadmap churn and
  unused test table field were repaired during review.
- No placeholder success paths, test-only production hooks, new dependencies,
  ADB/device behavior, WebSocket token consumption, or browser UI behavior were
  added.
Changed:
- none
Next: ready
Blocker: none

Phase: ready
Result: CYCLE READY
Evidence:
- `CYCLE-20260729-M1-S5` reached `REVIEW PASSED`.
Changed:
- `.codex/cycles/history.md`
- `.codex/plans/current.md`
- `cmd/adb-dashboard/main.go`
- `docs/roadmap.md`
- `tests/cli/m1_s1_cli_test.go`
Next: commit when authorized; otherwise `M1-S6`.
Blocker: none
