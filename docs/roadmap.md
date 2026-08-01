# Roadmap

## Status

- Roadmap status: Accepted
- Roadmap version: 1.2.0
- Specification source: `docs/SPECIFICATION.md`
- Specification version: 1.2.0
- Current milestone: M3
- Next eligible slice: M3-S1
- Last reviewed: 2026-08-01

This roadmap selects implementation order for the accepted local bootstrap and
read-only ADB discovery and M3 read-only device inspection contract. It does
not claim behavior is implemented.

## Slice Rules

Every implementation slice must:

- deliver one observable behavior or one bounded defect correction;
- use exactly one mode;
- reference exact specification and acceptance-criterion IDs;
- name one primary acceptance boundary;
- define useful red or baseline-green evidence;
- define focused verification;
- define a real-path exercise;
- define applicable broad verification;
- state dependencies, in-scope work, and out-of-scope work;
- identify material risks and stop conditions;
- use a binary exit gate;
- avoid horizontal scaffolding without a working path;
- fit one reviewable diff and one implementation cycle.

## Milestone M1: Local Bootstrap

### Slice M1-S1: CLI Discovery And Version Output

- Status: verified
- Mode: feature
- Purpose: Deliver the first usable process boundary for command discovery and
  build metadata without starting the server.
- Specification references: `CAP-001`, `CAP-002`, `AC-001-001`,
  `AC-001-002`, `AC-002-001`, `AC-002-002`, `INV-DATA-003`
- Observable result: A local operator can run help, version, and invalid
  invocations and observe documented stdout, stderr, and exit statuses with no
  listener, browser-open attempt, startup directory creation, or ADB execution.
- Primary acceptance boundary: CLI process invocation.
- Expected red or baseline-green evidence: The focused process test fails
  because no production `adb-dashboard` executable path exists or because help,
  version, or invalid invocation output does not match the contract.
- Focused verification command or deterministic discovery rule: Discover the
  repository's focused process-test command from manifests or test
  documentation; if no source or test harness exists, the implementation cycle
  must create the minimal production executable and acceptance test, then record
  the exact focused command in `.codex/plans/current.md` before build work.
- Real-path exercise: Build or invoke the real `adb-dashboard` command, run
  `--help`, `version`, `--version`, an unknown command, an unknown option, and a
  missing option argument, then record stdout, stderr, exit status, and absence
  of startup side effects.
- Broad verification: Run applicable repository test, formatting, lint,
  type-check, and build commands discovered from repository entry points.
- Required environment: Linux host capable of executing the built command.
- Dependencies: None.
- In scope:
  - Command dispatch for no server-starting discovery paths.
  - Help output, version output, invalid command, invalid option, and missing
    argument diagnostics.
  - Process-level acceptance tests.
- Out of scope:
  - Server startup, configuration files, doctor, filesystem creation, HTTP
    routes, browser UI, ADB execution, device operations, packaging, release,
    and unrelated framework setup.
- Risks:
  - Creating a command skeleton that reports success for behavior not yet
    present.
  - Version metadata defaults leaking host paths or environment values.
- Stop conditions:
  - A production process boundary cannot be built or invoked.
  - Command syntax or version metadata labels conflict with
    `docs/SPECIFICATION.md`.
- Documentation synchronization: Update user-facing command documentation if it
  exists; otherwise record `DOCS NOT REQUIRED` with reason.
- Exit gate: `AC-001-001`, `AC-001-002`, `AC-002-001`, and `AC-002-002` have
  green process-level evidence, forbidden startup side effects are checked,
  applicable broad checks pass or have recorded blockers, and review passes.
- Completion evidence reference: `.codex/plans/current.md` and
  `.codex/cycles/history.md`.

### Slice M1-S2: Configuration Precedence And Successful Doctor Report

- Status: verified
- Mode: feature
- Purpose: Make current configuration resolution observable through `doctor`
  and successful startup-directory validation.
- Specification references: `CAP-004`, `CAP-005`, `CAP-006`, `AC-004-001`,
  `AC-004-002`, `AC-005-001`, `AC-005-002`, `AC-006-001`, `CFG-001` through
  `CFG-010`, `DATA-001` through `DATA-005`, `INV-DATA-001`,
  `INV-DATA-002`, `INV-DATA-003`, `INV-NIY-001-B`
- Observable result: A local operator can run `adb-dashboard doctor` with
  defaults, default files, explicit files, environment overrides, and CLI
  overrides and see the documented successful report while only data and temp
  directories are created or validated.
- Primary acceptance boundary: CLI process invocation with isolated
  configuration files, environment, and filesystem paths.
- Expected red or baseline-green evidence: The focused doctor/config test fails
  because configuration precedence, successful report fields, directory
  creation, or documented `NIY` rows are absent or incorrect.
- Focused verification command or deterministic discovery rule: Discover the
  focused process-test command for doctor/config behavior; record the resolved
  command in `.codex/plans/current.md` before implementation begins.
- Real-path exercise: Run `adb-dashboard doctor` in an isolated temporary
  environment with explicit config files, environment variables, and CLI
  overrides; inspect stdout, stderr, exit status, created directories, and
  absence of ADB or browser side effects.
- Broad verification: Run applicable repository test, formatting, lint,
  type-check, and build commands discovered from repository entry points.
- Required environment: Linux host with readable and writable temporary
  directories and TOML fixture files.
- Dependencies: `M1-S1`.
- In scope:
  - Current TOML keys, documented environment variables, CLI option precedence,
    defaults, path expansion, and relative path resolution needed for successful
    doctor output.
  - Successful creation or validation of only resolved data and temp
    directories.
  - Stable successful doctor report shape and `NIY` informational rows.
- Out of scope:
  - Configuration failure cases, filesystem failure cases, server startup,
    HTTP routes, browser UI, ADB discovery or execution, host-tool discovery,
    persisted project state, and migration behavior.
- Risks:
  - Host environment leakage into tests.
  - Directory creation outside isolated paths.
- Stop conditions:
  - Required path fallback semantics are ambiguous in the selected runtime.
  - Tests cannot isolate config files, environment variables, and filesystem
    side effects.
- Documentation synchronization: Update user-facing configuration or doctor
  documentation if it exists; otherwise record `DOCS NOT REQUIRED` with reason.
- Exit gate: Referenced acceptance criteria have green process and filesystem
  evidence, only documented directories are created, unavailable rows remain
  `NIY`, applicable broad checks pass or have recorded blockers, and review
  passes.
- Completion evidence reference: `.codex/plans/current.md` and
  `.codex/cycles/history.md`.

### Slice M1-S3: Configuration And Startup Filesystem Failure Behavior

- Status: verified
- Mode: feature
- Purpose: Fail closed for invalid configuration and unavailable startup
  directories before server startup or browser-open side effects.
- Specification references: `CAP-004`, `CAP-005`, `CAP-006`, `AC-004-003`,
  `AC-004-004`, `AC-004-005`, `AC-004-006`, `AC-005-003`, `AC-005-004`,
  `AC-005-005`, `AC-006-002`, `AC-006-003`, `ERR-004-001` through
  `ERR-004-010`, `INV-DATA-001`, `INV-DATA-002`, `INV-DATA-003`,
  `INV-NIY-001-B`
- Observable result: Invalid config and startup-directory failures produce the
  documented diagnostics, stdout, stderr, exit statuses, and side effects for
  both `doctor` and startup attempts.
- Primary acceptance boundary: CLI process invocation with isolated
  configuration and filesystem failure conditions.
- Expected red or baseline-green evidence: The focused process test fails
  because malformed config, missing explicit config, unknown keys, invalid
  values, non-directory paths, or creation failures do not produce the
  documented behavior.
- Focused verification command or deterministic discovery rule: Discover the
  focused process-test command for config and filesystem negative paths; record
  the resolved command in `.codex/plans/current.md` before implementation
  begins.
- Real-path exercise: Run `adb-dashboard doctor` and `adb-dashboard serve`
  against isolated malformed, missing, and invalid config inputs plus
  non-directory startup paths; inspect stdout, stderr, exit status, listener
  absence, browser-open absence, and directory side effects.
- Broad verification: Run applicable repository test, formatting, lint,
  type-check, and build commands discovered from repository entry points.
- Required environment: Linux host with temporary files, writable directories,
  and deterministic non-directory failure fixtures.
- Dependencies: `M1-S2`.
- In scope:
  - All documented `ERR-004-*` diagnostics.
  - Fatal config validation before directory creation.
  - `doctor` failure report rows and exit status `5`.
  - Startup filesystem failure exit status `5` before listener binding.
- Out of scope:
  - HTTP route behavior, browser UI, browser-open success or warning behavior,
    ADB execution, device workflows, optional host tools, and cleanup policies
    not defined by the specification.
- Risks:
  - Platform-specific permission behavior making unwritable-directory tests
    nondeterministic.
  - Partial directory creation after ordered startup validation.
- Stop conditions:
  - A deterministic filesystem failure cannot be created without touching
    unsafe host paths.
  - Any failure behavior would require silently continuing after a fatal
    configuration or data-safety error.
- Documentation synchronization: Update user-facing diagnostics documentation
  if it exists; otherwise record `DOCS NOT REQUIRED` with reason.
- Exit gate: Referenced acceptance criteria have green process and filesystem
  evidence, fatal paths perform no forbidden startup side effects, applicable
  broad checks pass or have recorded blockers, and review passes.
- Completion evidence reference: `.codex/plans/current.md` and
  `.codex/cycles/history.md`.

### Slice M1-S4: Loopback Server Lifecycle And Status API

- Status: verified
- Mode: feature
- Purpose: Start the local dashboard server on loopback, expose current status,
  reject unavailable listen addresses, and shut down cleanly.
- Specification references: `CAP-003`, `CAP-009`, `AC-001-003`,
  `AC-003-001`, `AC-003-002`, `AC-003-003`, `AC-003-004`, `AC-003-005`,
  `AC-003-006`, `AC-009-001`, `AC-009-002`, `AC-009-003`, `AC-009-004`,
  `INV-SEC-002`, `INV-DATA-001`, `INV-DATA-003`, `INV-NIY-001-B`
- Observable result: A local operator can start `adb-dashboard` or
  `adb-dashboard serve` on a loopback listener, request `/api/v1/status` and an
  unknown `/api/v1` route, observe documented JSON, and stop the process with
  `SIGINT` or `SIGTERM`.
- Primary acceptance boundary: Process invocation plus HTTP requests through
  the bound listener.
- Expected red or baseline-green evidence: The focused process/HTTP test fails
  because the server cannot bind, does not report startup and shutdown
  diagnostics, does not serve the documented status JSON, does not reject
  invalid listen hosts, or does not return the unknown-route envelope.
- Focused verification command or deterministic discovery rule: Discover the
  focused process and HTTP lifecycle test command; record the resolved command
  in `.codex/plans/current.md` before implementation begins.
- Real-path exercise: Start the built command on `127.0.0.1:0`, parse the
  startup diagnostic, request `/api/v1/status` and `/api/v1/unknown`, terminate
  with `SIGTERM`, and record stderr, stdout, statuses, JSON bodies, side
  effects, and process exit status.
- Broad verification: Run applicable repository test, formatting, lint,
  type-check, and build commands discovered from repository entry points.
- Required environment: Linux host with available loopback networking,
  temporary startup directories, and controllable `PATH` for `xdg-open` tests.
- Dependencies: `M1-S3`.
- In scope:
  - Default no-subcommand startup behavior.
  - Loopback listen validation and unavailable listen-address failure.
  - Startup and shutdown diagnostics.
  - Controlled `--open` success attempt and failure warning after startup.
  - Current status JSON and unknown `/api/v1` route envelope.
- Out of scope:
  - Bootstrap tokens, host/origin rejection, browser shell rendering,
    WebSocket behavior, CSRF enforcement, ADB execution, watchers, jobs,
    sessions, storage projects, host-tool execution, and non-loopback serving.
- Risks:
  - Process lifecycle tests hanging if shutdown or listener discovery is
    incomplete.
  - Browser-open behavior accidentally running a host browser instead of a
    controlled test executable.
- Stop conditions:
  - Loopback networking is unavailable.
  - Server startup cannot be exercised as a real process.
  - `xdg-open` cannot be controlled safely for automated or real-path evidence.
- Documentation synchronization: Update user-facing server/API documentation if
  it exists; otherwise record `DOCS NOT REQUIRED` with reason.
- Exit gate: Referenced acceptance criteria have green process and HTTP
  evidence, status JSON contains only documented fields and unavailable values,
  shutdown exits `0`, applicable broad checks pass or have recorded blockers,
  and review passes.
- Completion evidence reference: `.codex/plans/current.md` and
  `.codex/cycles/history.md`.

### Slice M1-S5: Browser Security Bootstrap

- Status: verified
- Mode: feature
- Purpose: Issue per-process bootstrap tokens to same-origin loopback requests
  while backend policy rejects foreign host and origin requests before route
  handlers run.
- Specification references: `CAP-008`, `CAP-009`, `AC-008-001`,
  `AC-008-002`, `AC-008-003`, `AC-008-004`, `AC-009-004`, `INV-SEC-001`,
  `INV-SEC-003`, `INV-SEC-004`, `INV-NIY-001-C`
- Observable result: A same-origin request to `/api/v1/bootstrap` receives the
  documented JSON token response, token values change after restart, status
  never includes tokens, and foreign Host, absolute-form host, or Origin
  requests receive the documented `403` error envelope without route-specific
  fields.
- Primary acceptance boundary: HTTP request through production routing.
- Expected red or baseline-green evidence: The focused HTTP security test fails
  because bootstrap is absent, token format or restart lifecycle is wrong,
  status leaks token fields, or rejected host/origin requests reach route
  handlers or return the wrong envelope.
- Focused verification command or deterministic discovery rule: Discover the
  focused HTTP security test command; record the resolved command in
  `.codex/plans/current.md` before implementation begins.
- Real-path exercise: Start the server, request `/api/v1/bootstrap` and
  `/api/v1/status` with accepted loopback headers, restart the server and
  compare token values, then request current `/api/v1` routes with rejected
  Host, absolute-form URL host, and Origin inputs.
- Broad verification: Run applicable repository test, formatting, lint,
  type-check, and build commands discovered from repository entry points.
- Required environment: Linux host with loopback networking and an HTTP client
  capable of setting Host and Origin headers and absolute-form request targets.
- Dependencies: `M1-S4`.
- In scope:
  - `GET /api/v1/bootstrap`.
  - Per-process `csrfToken` and `webSocketToken` generation and response
    shape.
  - Backend host, absolute-form host, and origin rejection for current
    `/api/v1` routes.
  - Token non-disclosure in status JSON, logs, doctor output, and visible UI
    surfaces available at this slice.
- Out of scope:
  - Consuming tokens for future mutating requests, WebSocket authorization,
    local-user authentication, browser UI rendering, ADB/device behavior, and
    request-correlation IDs beyond the documented `null` value.
- Risks:
  - Host header parsing differences between HTTP server libraries.
  - Token values accidentally appearing in diagnostics or frontend markup.
- Stop conditions:
  - The selected HTTP stack cannot expose enough request-target and header data
    to enforce the documented policy.
  - Token generation cannot use an appropriate random source.
- Documentation synchronization: Update user-facing API or security
  documentation if it exists; otherwise record `DOCS NOT REQUIRED` with reason.
- Exit gate: Referenced acceptance criteria have green HTTP routing evidence,
  rejected requests bypass route-specific output, token lifecycle and
  non-disclosure are verified, applicable broad checks pass or have recorded
  blockers, and review passes.
- Completion evidence reference: `.codex/plans/current.md` and
  `.codex/cycles/history.md`.

### Slice M1-S6: Embedded Browser Shell

- Status: verified
- Mode: feature
- Purpose: Serve the current browser dashboard shell and render backend-derived
  status without exposing controls for out-of-scope workflows.
- Specification references: `CAP-007`, `CAP-008`, `CAP-009`, `AC-007-001`,
  `AC-007-002`, `AC-007-003`, `INV-FRONTEND-001`, `INV-SEC-004`,
  `INV-NIY-001-B`
- Observable result: A browser loaded from the running server renders the
  documented application, server, read-only, and subsystem states from
  `/api/v1/bootstrap` and `/api/v1/status`, hides token values, and has no
  active controls for unavailable ADB, device, session, job, transfer,
  artifact, host-tool, or destructive behavior.
- Primary acceptance boundary: Browser load through the running server.
- Expected red or baseline-green evidence: The focused browser test fails
  because the root page is absent, visible state is not derived from backend
  responses, unavailable controls are present, token values render, or backend
  failure does not show `server: unavailable`.
- Focused verification command or deterministic discovery rule: Discover the
  focused browser or HTTP asset test command; record the resolved command in
  `.codex/plans/current.md` before implementation begins.
- Real-path exercise: Start the server on `127.0.0.1:0`, load the root page in
  a modern browser or deterministic browser automation, inspect visible state,
  confirm absence of out-of-scope controls and token values, then simulate
  bootstrap or status failure and inspect the unavailable state.
- Broad verification: Run applicable repository test, formatting, lint,
  type-check, build, and frontend asset checks discovered from repository entry
  points.
- Required environment: Linux host with loopback networking and a modern
  JavaScript-capable browser or deterministic browser automation.
- Dependencies: `M1-S5`.
- In scope:
  - Embedded root browser shell assets served from the local server.
  - Bootstrap request followed by status request using returned `statusUrl`.
  - Visible current state and unavailable fallback state.
  - Absence of active controls, forms, or links for out-of-scope behavior.
- Out of scope:
  - Device operations, ADB execution, interactive sessions, WebSockets, file
    transfer, screenshots, package workflows, artifact analysis, settings,
    authentication, persistence, theming beyond the current shell, packaging,
    and deployment.
- Risks:
  - Frontend tests passing against static markup instead of backend-derived
    state.
  - UI labels drifting from the stable current contract.
- Stop conditions:
  - Browser automation or an equivalent deterministic visible-state boundary is
    unavailable.
  - Implementing the shell would require controls or routes for out-of-scope
    behavior.
- Documentation synchronization: Update user-facing browser/API documentation
  if it exists; otherwise record `DOCS NOT REQUIRED` with reason.
- Exit gate: Referenced acceptance criteria have green browser-boundary
  evidence, backend-derived state and failure state are observed, forbidden
  controls and token rendering are absent, applicable broad checks pass or have
  recorded blockers, and review passes.
- Completion evidence reference: `.codex/plans/current.md` and
  `.codex/cycles/history.md`.

## Milestone M2: Read-Only ADB Discovery

### Slice M2-S1: Doctor ADB Executable And Version Discovery

- Status: verified
- Mode: feature
- Purpose: Make ADB executable and version discovery observable in `doctor`
  without performing device operations.
- Specification references: `CAP-006`, `CAP-010`, `AC-006-001`,
  `AC-010-001`, `AC-010-002`, `AC-010-003`, `INV-DATA-003`
- Observable result: A local operator can run `adb-dashboard doctor` with a
  deterministic `PATH` and observe documented ADB executable, version,
  unavailable, and failure rows plus exit status `0` or `3` as appropriate.
- Primary acceptance boundary: CLI process invocation with isolated `PATH` and
  fake ADB executables.
- Expected red or baseline-green evidence: The focused doctor process test
  fails because ADB rows still report static `NIY`, `adb version` is not
  invoked, missing ADB does not produce exit status `3`, or version failures do
  not produce the documented report.
- Focused verification command or deterministic discovery rule: Discover the
  focused process-test command for doctor behavior; record the resolved command
  in `.codex/plans/current.md` before implementation begins.
- Real-path exercise: Build or invoke the real command, run
  `adb-dashboard doctor` with a temporary fake `adb` that succeeds, with a
  `PATH` containing no `adb`, and with a fake `adb version` failure; inspect
  stdout, stderr, exit status, and fake-command invocation logs.
- Broad verification: Run applicable repository test, formatting, lint,
  type-check, and build commands discovered from repository entry points.
- Required environment: Linux host capable of process execution and temporary
  executable fixtures.
- Dependencies: `M1-S6`.
- In scope:
  - Resolve `adb` from the process `PATH`.
  - Invoke only `adb version` with an argument vector.
  - Parse the first non-empty stdout line as the display version.
  - Report doctor ADB rows and exit status `3` for ADB discovery/version
    failure when no higher-priority failure is present.
  - Timeout and nonzero status handling for `adb version`.
- Out of scope:
  - `GET /api/v1/status` ADB fields, `GET /api/v1/devices`, browser device
    rendering, explicit ADB server controls, `adb devices`, shell, logcat,
    install, file transfer, screenshots, device selection, and device mutation.
- Risks:
  - Accidentally invoking host `adb` during tests instead of the fake
    executable.
  - Leaking environment values or command stderr beyond bounded diagnostics.
- Stop conditions:
  - Process tests cannot isolate `PATH` and fake executables.
  - Implementing the slice would require shell interpolation or broader ADB
    command execution.
- Documentation synchronization: Update user-facing doctor/ADB documentation if
  it exists; otherwise record `DOCS NOT REQUIRED` with reason.
- Exit gate: `AC-006-001`, `AC-010-001`, `AC-010-002`, and `AC-010-003` have
  green process-level evidence, only `adb version` is invoked, applicable broad
  checks pass or have recorded blockers, and review passes.
- Completion evidence reference: `.codex/plans/current.md` and
  `.codex/cycles/history.md`.

### Slice M2-S2: Status API ADB Summary

- Status: verified
- Mode: feature
- Purpose: Expose ADB executable and version discovery through the current
  status API without device inventory side effects.
- Specification references: `CAP-009`, `CAP-010`, `AC-009-001`,
  `AC-009-002`, `AC-009-004`, `AC-010-004`, `AC-010-005`, `INV-SEC-001`,
  `INV-DATA-003`
- Observable result: A browser or local HTTP client can request
  `/api/v1/status` and observe `adb.status`, `adb.executable`, `adb.version`,
  and `adb.serverResponsive` values derived from real ADB discovery.
- Primary acceptance boundary: HTTP request through production routing with
  isolated `PATH` and fake ADB executables.
- Expected red or baseline-green evidence: The focused HTTP status test fails
  because the `adb` object still reports static `NIY`, leaks forbidden command
  details, invokes unsupported ADB commands, or fails to represent absent and
  failed ADB discovery as documented.
- Focused verification command or deterministic discovery rule: Discover the
  focused HTTP status test command; record the resolved command in
  `.codex/plans/current.md` before implementation begins.
- Real-path exercise: Start the server with a fake successful `adb`, request
  `/api/v1/status`, restart with no `adb` on `PATH`, request status again, then
  restart with a failing fake `adb version`; inspect status, JSON bodies, and
  fake-command invocation logs.
- Broad verification: Run applicable repository test, formatting, lint,
  type-check, and build commands discovered from repository entry points.
- Required environment: Linux host with loopback networking and temporary
  executable fixtures.
- Dependencies: `M2-S1`.
- In scope:
  - Status-route ADB discovery and version fields.
  - Available, unavailable, and error ADB summary states.
  - Preservation of existing server, watcher, jobs, sessions, storage, and host
    tools fields.
  - Host and origin rejection before ADB command execution.
- Out of scope:
  - `/api/v1/devices`, browser device rendering, explicit ADB server controls,
    device listing, shell, logcat, install, file transfer, screenshots, device
    selection, and device mutation.
- Risks:
  - Turning every status request into unsupported device inventory work.
  - Leaking command stderr, environment values, or token values in status JSON.
- Stop conditions:
  - Host/origin middleware cannot be proven to run before ADB discovery.
  - Status behavior cannot be exercised through the running HTTP server.
- Documentation synchronization: Update user-facing status/API documentation if
  it exists; otherwise record `DOCS NOT REQUIRED` with reason.
- Exit gate: Referenced acceptance criteria have green HTTP evidence, status
  invokes only `adb version`, rejected security requests invoke no ADB process,
  applicable broad checks pass or have recorded blockers, and review passes.
- Completion evidence reference: `.codex/plans/current.md` and
  `.codex/cycles/history.md`.

### Slice M2-S3: Read-Only Devices API

- Status: verified
- Mode: feature
- Purpose: Add a read-only device inventory route backed by `adb devices -l`.
- Specification references: `CAP-008`, `CAP-011`, `AC-011-001`,
  `AC-011-002`, `AC-011-003`, `AC-011-004`, `AC-011-005`, `INV-SEC-001`,
  `INV-SEC-003`, `INV-DATA-003`
- Observable result: A same-origin local HTTP client can request
  `/api/v1/devices` and receive parsed zero-device or multi-device JSON, while
  ADB-unavailable, command-failure, malformed-output, timeout, and rejected
  security requests fail with documented envelopes and no false success.
- Primary acceptance boundary: HTTP request through production routing with
  fake ADB executables.
- Expected red or baseline-green evidence: The focused HTTP devices test fails
  because `/api/v1/devices` is absent, does not invoke `adb devices -l`, does
  not parse rows, returns the wrong error envelope, or invokes ADB after a
  rejected Host or Origin request.
- Focused verification command or deterministic discovery rule: Discover the
  focused HTTP devices test command; record the resolved command in
  `.codex/plans/current.md` before implementation begins.
- Real-path exercise: Start the server with fake `adb` fixtures for zero-device,
  multi-device, ADB-unavailable, `devices -l` failure, and malformed output
  cases; request `/api/v1/devices`; inspect status codes, JSON bodies, and
  fake-command invocation logs.
- Broad verification: Run applicable repository test, formatting, lint,
  type-check, and build commands discovered from repository entry points.
- Required environment: Linux host with loopback networking and temporary
  executable fixtures.
- Dependencies: `M2-S2`.
- In scope:
  - `GET /api/v1/devices`.
  - `adb devices -l` invocation with an argument vector.
  - Parsing serial, state, product, model, device, and transport ID fields.
  - `503 adb_unavailable`, `502 adb_devices_failed`, timeout, malformed output,
    and security rejection behavior.
- Out of scope:
  - Browser rendering, persisted inventory, polling or watchers, explicit ADB
    server controls, device selection, shell, logcat, install, file transfer,
    screenshots, package workflows, jobs, WebSockets, and device mutation.
- Risks:
  - Treating malformed or partial ADB output as successful inventory.
  - Host ADB client side effects beyond the normal behavior of
    `adb devices -l`.
- Stop conditions:
  - The route cannot be exercised through production routing.
  - Deterministic fake ADB fixtures cannot model success, failure, malformed,
    and timeout cases without invoking host ADB.
- Documentation synchronization: Update user-facing API documentation if it
  exists; otherwise record `DOCS NOT REQUIRED` with reason.
- Exit gate: Referenced acceptance criteria have green HTTP evidence, only
  `adb version` and `adb devices -l` are invoked, rejected security requests
  invoke no ADB process, applicable broad checks pass or have recorded blockers,
  and review passes.
- Completion evidence reference: `.codex/plans/current.md` and
  `.codex/cycles/history.md`.

### Slice M2-S4: Browser ADB Device Inventory View

- Status: verified
- Mode: feature
- Purpose: Render backend-derived ADB availability and read-only device
  inventory in the embedded browser shell.
- Specification references: `CAP-007`, `CAP-011`, `CAP-012`,
  `AC-012-001`, `AC-012-002`, `AC-012-003`, `AC-012-004`,
  `INV-FRONTEND-001`, `INV-SEC-004`, `INV-DATA-003`
- Observable result: A browser loaded from the running server shows ADB
  availability, device count, and device serial/state values from backend
  responses, while failed inventory shows unavailable state and no unsupported
  controls or sensitive values.
- Primary acceptance boundary: Browser load through the running server.
- Expected red or baseline-green evidence: The focused browser test fails
  because the shell still renders static ADB `NIY`, does not request
  `/api/v1/devices` when ADB is available, does not render returned devices,
  shows stale success after inventory failure, exposes unsupported controls, or
  renders executable paths, tokens, stderr, or environment values.
- Focused verification command or deterministic discovery rule: Discover the
  focused browser or HTTP asset test command; record the resolved command in
  `.codex/plans/current.md` before implementation begins.
- Real-path exercise: Start the server with a deterministic fake `adb`, load
  the root page in a modern browser or deterministic browser automation, inspect
  visible ADB and device state, then repeat with inventory failure and inspect
  unavailable state plus absence of unsupported controls and sensitive values.
- Broad verification: Run applicable repository test, formatting, lint,
  type-check, build, and frontend asset checks discovered from repository entry
  points.
- Required environment: Linux host with loopback networking, temporary
  executable fixtures, and a modern JavaScript-capable browser or deterministic
  browser automation.
- Dependencies: `M2-S3`.
- In scope:
  - Root shell rendering for ADB available, unavailable, and error states.
  - Device inventory request after status reports ADB availability.
  - Visible device count, serial, and state.
  - Inventory failure fallback.
  - Absence of unsupported ADB controls and sensitive values.
- Out of scope:
  - Device actions, polling, watchers, sessions, WebSockets, file transfer,
    screenshots, logcat, package workflows, artifact analysis, settings,
    authentication, persistence, theming beyond the current shell, packaging,
    and deployment.
- Risks:
  - Frontend tests passing against static markup instead of backend-derived
    ADB/device responses.
  - Accidentally exposing ADB executable paths or command diagnostics in the UI.
- Stop conditions:
  - Browser automation or an equivalent deterministic visible-state boundary is
    unavailable.
  - Implementing the view would require controls or routes for out-of-scope ADB
    behavior.
- Documentation synchronization: Update user-facing browser/API documentation
  if it exists; otherwise record `DOCS NOT REQUIRED` with reason.
- Exit gate: Referenced acceptance criteria have green browser-boundary
  evidence, backend-derived ADB and device states are observed, forbidden
  controls and sensitive rendering are absent, applicable broad checks pass or
  have recorded blockers, and review passes.
- Completion evidence reference: `.codex/plans/current.md` and
  `.codex/cycles/history.md`.

## Milestone M3: Read-Only Device Inspection

### Slice M3-S1: Explicit Device Refresh And Detail View

- Status: planned
- Mode: feature
- Purpose: Let a browser user explicitly refresh current device inventory and
  open read-only details for one current device.
- Specification references: `CAP-011`, `CAP-012`, `CAP-013`,
  `AC-013-001`, `AC-013-002`, `AC-013-003`, `AC-013-004`,
  `AC-013-005`, `INV-FRONTEND-001`, `INV-SEC-001`, `INV-SEC-003`,
  `INV-SEC-004`, `INV-DATA-003`
- Observable result: A browser loaded from the running server can activate a
  refresh control, observe updated backend-derived device inventory, open one
  device detail, and see unavailable/error states without unsupported controls
  or sensitive values.
- Primary acceptance boundary: Browser interaction and HTTP request through the
  running server with deterministic fake ADB executables.
- Expected red or baseline-green evidence: The focused browser/API test fails
  because no explicit refresh control exists, `/api/v1/devices/{serial}` is
  absent, stale detail can remain after failure, rejected Host or Origin
  requests reach ADB, or unsupported controls/sensitive values are rendered.
- Focused verification command or deterministic discovery rule: Discover the
  focused browser and HTTP test command for device refresh/detail behavior;
  record the resolved command in `.codex/plans/current.md` before
  implementation begins.
- Real-path exercise: Start the server with fake ADB inventories, load the
  browser shell, activate refresh, open detail for one serial, request the
  device-detail route directly, then repeat with an absent serial and inventory
  failure; inspect visible state, statuses, JSON bodies, and fake-command logs.
- Broad verification: Run applicable repository test, formatting, lint,
  type-check, build, browser-script, and API checks discovered from repository
  entry points.
- Required environment: Linux host with loopback networking, temporary
  executable fixtures, and a modern JavaScript-capable browser or deterministic
  browser automation.
- Dependencies: `M2-S4`.
- In scope:
  - Explicit browser refresh control backed by current `/api/v1/devices`.
  - `GET /api/v1/devices/{serial}` route.
  - Detail rendering for serial, state, and present inventory properties.
  - Available, unavailable, failed, malformed, timeout, absent-serial, and
    rejected-security behavior.
  - Absence of unsupported workflow controls and sensitive values.
- Out of scope:
  - Persistent device selection, mutation, polling, watchers, sessions,
    WebSockets, logcat, screenshots, file transfer, package workflows,
    artifact analysis, retained inventory, settings, packaging, and deployment.
- Risks:
  - Introducing a UI selection state that looks like persistent device control.
  - Serving detail from stale browser state instead of backend inventory.
  - Leaking command stderr, environment values, or executable paths.
- Stop conditions:
  - The detail route cannot be exercised through production routing.
  - Browser automation or an equivalent deterministic visible-state boundary is
    unavailable.
  - Implementing detail would require mutation, persistence, or unsupported ADB
    commands.
- Documentation synchronization: Update user-facing browser/API documentation
  if it exists; otherwise record `DOCS NOT REQUIRED` with reason.
- Exit gate: `AC-013-001` through `AC-013-005` have green browser/API evidence,
  only `adb version` and `adb devices -l` are invoked, rejected security
  requests invoke no ADB process, applicable broad checks pass or have recorded
  blockers, and review passes.
- Completion evidence reference: `.codex/plans/current.md` and
  `.codex/cycles/history.md`.

### Slice M3-S2: Read-Only Device Logcat

- Status: planned
- Mode: feature
- Purpose: Retrieve bounded read-only logcat text for one current ready device
  without streaming, clearing logs, or retaining output.
- Specification references: `CAP-013`, `CAP-014`, `AC-014-001`,
  `AC-014-002`, `AC-014-003`, `AC-014-004`, `INV-SEC-001`,
  `INV-SEC-003`, `INV-DATA-003`
- Observable result: A browser or same-origin HTTP client can request bounded
  logcat for a selected ready device and observe success, empty, invalid-input,
  unavailable, failure, timeout, and security-rejection states.
- Primary acceptance boundary: HTTP request and browser interaction through the
  running server with deterministic fake ADB executables.
- Expected red or baseline-green evidence: The focused HTTP/browser test fails
  because `/api/v1/devices/{serial}/logcat` is absent, invalid query values are
  accepted, the command is not bounded to `adb -s SERIAL logcat -d`, output is
  retained, rejected security requests invoke ADB, or browser controls imply
  unsupported streaming/clear/shell behavior.
- Focused verification command or deterministic discovery rule: Discover the
  focused logcat API and browser test command; record the resolved command in
  `.codex/plans/current.md` before implementation begins.
- Real-path exercise: Start the server with fake ADB inventories and logcat
  output, request bounded logcat for one serial, inspect browser logcat state,
  repeat invalid-input, absent-serial, non-ready-device, command-failure, and
  timeout cases, and inspect fake-command logs.
- Broad verification: Run applicable repository test, formatting, lint,
  type-check, build, browser-script, and API checks discovered from repository
  entry points.
- Required environment: Linux host with loopback networking and temporary
  executable fixtures.
- Dependencies: `M3-S1`.
- In scope:
  - `GET /api/v1/devices/{serial}/logcat`.
  - `lines` and `format` query validation.
  - Bounded execution of `adb -s SERIAL logcat -d`.
  - Last-N-line response and truncation flag.
  - Browser logcat view with loading, empty, success, and failure states.
  - Negative paths and rejected-security behavior.
- Out of scope:
  - Live streaming, WebSockets, log clearing, filters beyond accepted query
    parameters, shell, sessions, jobs, retained logs, file writes, device
    mutation, package workflows, screenshots, packaging, and deployment.
- Risks:
  - Returning unbounded or retained log output.
  - Treating non-ready devices as ready for logcat.
  - Leaking stderr or host environment details in responses.
- Stop conditions:
  - Logcat cannot be exercised through production routing with deterministic
    fake ADB fixtures.
  - Implementing the slice would require WebSockets, persistence, shell
    interpolation, or unsupported ADB commands.
- Documentation synchronization: Update user-facing browser/API documentation
  if it exists; otherwise record `DOCS NOT REQUIRED` with reason.
- Exit gate: `AC-014-001` through `AC-014-004` have green HTTP/browser
  evidence, only allowed ADB commands are invoked, no retained output is
  written, rejected security requests invoke no ADB process, applicable broad
  checks pass or have recorded blockers, and review passes.
- Completion evidence reference: `.codex/plans/current.md` and
  `.codex/cycles/history.md`.

### Slice M3-S3: Read-Only Device Screenshot Capture

- Status: planned
- Mode: feature
- Purpose: Capture one PNG screenshot from a current ready device without
  retained files or device mutation.
- Specification references: `CAP-013`, `CAP-015`, `AC-015-001`,
  `AC-015-002`, `AC-015-003`, `AC-015-004`, `INV-SEC-001`,
  `INV-SEC-003`, `INV-DATA-003`
- Observable result: A browser or same-origin HTTP client can request a PNG
  screenshot for a selected ready device and observe success, unavailable,
  failure, invalid-output, timeout, and security-rejection states.
- Primary acceptance boundary: HTTP request and browser interaction through the
  running server with deterministic fake ADB executables.
- Expected red or baseline-green evidence: The focused HTTP/browser test fails
  because `/api/v1/devices/{serial}/screenshot` is absent, non-PNG output is
  returned as success, output is written to files, rejected security requests
  invoke ADB, or browser controls imply screen recording/file-transfer/device
  mutation behavior.
- Focused verification command or deterministic discovery rule: Discover the
  focused screenshot API and browser test command; record the resolved command
  in `.codex/plans/current.md` before implementation begins.
- Real-path exercise: Start the server with fake ADB inventories and PNG
  screenshot output, request screenshot for one serial, inspect content type and
  PNG bytes, inspect browser image state, repeat non-ready-device,
  absent-serial, command-failure, timeout, empty-output, non-PNG, and
  rejected-security cases, and inspect fake-command logs plus filesystem side
  effects.
- Broad verification: Run applicable repository test, formatting, lint,
  type-check, build, browser-script, and API checks discovered from repository
  entry points.
- Required environment: Linux host with loopback networking and temporary
  executable fixtures.
- Dependencies: `M3-S2`.
- In scope:
  - `GET /api/v1/devices/{serial}/screenshot`.
  - Bounded execution of `adb -s SERIAL exec-out screencap -p`.
  - PNG validation and image response.
  - Browser screenshot view with loading, success, and failure states.
  - Negative paths, no retained file output, and rejected-security behavior.
- Out of scope:
  - Screen recording, image annotation, retained screenshots, artifact storage,
    file transfer, shell, sessions, jobs, package workflows, device mutation,
    packaging, and deployment.
- Risks:
  - Returning command error text as an image.
  - Persisting sensitive screenshots by accident.
  - Treating non-ready devices as ready for screenshot capture.
- Stop conditions:
  - Screenshot cannot be exercised through production routing with
    deterministic fake ADB fixtures.
  - Implementing the slice would require persistence, shell interpolation,
    mutation, or unsupported ADB commands.
- Documentation synchronization: Update user-facing browser/API documentation
  if it exists; otherwise record `DOCS NOT REQUIRED` with reason.
- Exit gate: `AC-015-001` through `AC-015-004` have green HTTP/browser
  evidence, only allowed ADB commands are invoked, no retained image file is
  written, rejected security requests invoke no ADB process, applicable broad
  checks pass or have recorded blockers, and review passes.
- Completion evidence reference: `.codex/plans/current.md` and
  `.codex/cycles/history.md`.

## Dependency Order

```text
M1-S1 -> M1-S2 -> M1-S3 -> M1-S4 -> M1-S5 -> M1-S6
M1-S6 -> M2-S1 -> M2-S2 -> M2-S3 -> M2-S4
M2-S4 -> M3-S1 -> M3-S2 -> M3-S3
```

## Future Milestones

Future mutating ADB, interactive device control, WebSocket streaming,
host-tool, artifact, persistence, logging redaction, request-correlation, body
parsing, upload, retention, cleanup, performance, migration, packaging,
release, and deployment behavior requires a later accepted specification before
roadmap slices are added.

## Roadmap Acceptance Record

- Audit result: ROADMAP ACCEPTED
- Reviewed slices: `M1-S1` through `M1-S6`; `M2-S1` through `M2-S4`;
  `M3-S1` through `M3-S3`
- Blocking gaps: None for the accepted local bootstrap, read-only ADB
  discovery, and M3 read-only device inspection contract.
- Evidence or review reference: Authored against `docs/SPECIFICATION.md`
  version `1.2.0`, `docs/ROADMAP_GUIDE.md`,
  `docs/ROADMAP.template.md`, `docs/READINESS_CHECKLIST.md`, `AGENTS.md`, and
  repository source material available on 2026-08-01.
