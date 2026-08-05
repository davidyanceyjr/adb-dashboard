# Specification

## Status

- Contract status: Accepted
- Specification version: 1.3.0
- Product or system: ADB Dashboard
- Version or milestone: Local bootstrap, read-only ADB discovery, M3 read-only
  device inspection, M4 read-only package inspection, and M5 local APK artifact
  intake and analysis
- Owners or decision authority: ADB Dashboard Project
- Last reviewed: August 2026
- Supersedes:

This document defines intended externally observable behavior. It does not claim
implementation. Implementation status is proven by tests, real-path execution,
and version-control history.

## Purpose

ADB Dashboard is a local Linux developer tool that serves a browser interface
for current dashboard status and read-only Android Debug Bridge inspection.

The accepted current contract covers local bootstrap behavior, read-only ADB
discovery and device inventory, M3 read-only device inspection through explicit
refresh, device detail, bounded logcat reads, and screenshot capture, M4
read-only package inspection, and M5 local APK artifact intake and analysis.
Interactive sessions, file transfer, device mutation, background jobs, retained
device output, and persistence beyond the artifact contracts are outside this
specification until a later accepted capability defines them.

## Scope

### In Scope

- `adb-dashboard`, `adb-dashboard serve`, `adb-dashboard version`, and
  `adb-dashboard doctor`.
- Documented command-line options and environment variables.
- TOML configuration discovery, precedence, validation, and failure behavior.
- Creation or validation of the current data and temp directories.
- Loopback-only HTTP serving of the embedded browser shell.
- `GET /api/v1/bootstrap`.
- `GET /api/v1/status`.
- Unknown `/api/v1` route handling.
- Browser host and origin rejection for current `/api/v1` routes.
- Current browser shell visible state.
- Discovery of the `adb` executable from the process `PATH`.
- Read-only `adb version` execution.
- Read-only `adb devices -l` execution and parsing.
- `GET /api/v1/devices`.
- Browser-visible ADB and device inventory state.
- Explicit browser/API device inventory refresh.
- Read-only `GET /api/v1/devices/{serial}` device detail derived from current
  inventory.
- Read-only bounded `adb logcat -d` execution for one selected serial.
- Read-only `adb exec-out screencap -p` execution for one selected serial.
- Browser-visible device detail, logcat, and screenshot states.
- Read-only package inventory and package detail for one selected serial.
- Browser-visible package inventory and package detail states.
- Local APK artifact upload, listing, detail, metadata extraction, browser
  artifact states, and explicit artifact deletion.
- Standard `NIY` behavior for recognized unavailable actions and status fields.

### Out Of Scope

- Executing ADB commands other than those named by `CAP-010`, `CAP-011`,
  `CAP-014`, `CAP-015`, `CAP-016`, and `CAP-017`.
- Explicit `adb start-server`, `adb kill-server`, or ADB server lifecycle
  controls. `adb devices -l` may use the host ADB client's normal server
  behavior as documented by ADB.
- Device selection and any device mutation.
- Raw command execution, binary execution, PTY, WebSocket sessions, and shell.
- File push or pull, package workflows, input, reboot, networking, process, or
  system-state operations beyond the read-only package inspection commands
  named by `CAP-016` and `CAP-017`.
- Logcat streaming, clearing, filtering beyond bounded query parameters, or
  retained log storage.
- Screenshot mutation, screen recording, image annotation, or retained
  screenshot storage.
- Host-tool discovery or execution outside local APK analysis through
  `CAP-019`.
- Artifact project storage, uploads, analysis, indexes, reports, caches,
  command history, retained output, and background jobs outside the local APK
  artifact behavior defined by `CAP-018` through `CAP-021`.
- Public OpenAPI documents, routes, commands, UI controls, settings, or
  configuration keys for out-of-scope behavior.

## Actors And Public Boundaries

| Actor or caller | Public interface | Trust or ownership boundary | Relevant capabilities |
|---|---|---|---|
| Local operator | CLI process invocation | User account running the process | `CAP-001` through `CAP-006`, `CAP-010`, `INV-NIY-001` |
| Browser loaded from the local server | Embedded UI and same-origin HTTP requests | Browser origin boundary | `CAP-007`, `CAP-008`, `CAP-009`, `CAP-012` through `CAP-021` |
| HTTP client | Loopback HTTP API | Host and origin policy enforced by backend | `CAP-008`, `CAP-009`, `CAP-011`, `CAP-013` through `CAP-021`, `INV-NIY-001` |
| Host filesystem | Config, data directory, temp directory | Files owned or selected by process user | `CAP-004`, `CAP-005`, `CAP-006` |
| Host filesystem | Local APK artifacts | Files under resolved data directory | `CAP-018`, `CAP-020`, `CAP-021` |
| Host ADB client | `adb` process invocation | Executable resolved from process `PATH` | `CAP-010`, `CAP-011`, `CAP-013` through `CAP-017` |
| Host APK analysis tool | `aapt` process invocation | Executable resolved from process `PATH` | `CAP-019` |

## Terminology

| ID or term | Meaning |
|---|---|
| `NIY` | Not Implemented Yet. A public unavailable-state value, not success. |
| Loopback host | `localhost` or an IP address that the host resolves as loopback. |
| Current route | A route named by this specification. |
| Future behavior | Behavior listed out of scope until a later accepted capability specifies it. |

## Global Invariants

- `INV-SEC-001`: Backend host and origin policy is authoritative for current
  `/api/v1` routes and is enforced before route-specific handlers produce
  response fields or side effects.
- `INV-DATA-001`: Current startup may create only the resolved data directory
  and temp directory. It must not create project, artifact, cache, upload,
  history, job, index, or retained-output paths during startup.
- `INV-NIY-001`: `NIY` is unavailable behavior, not requested-action success.
  Requested unavailable actions fail before external process execution, device
  mutation, project or artifact state writes, or session creation.
- `INV-FRONTEND-001`: The frontend must derive current status and capability
  state from backend responses. It must not invent success, availability, or
  security state.

## Behavioral Contracts

### CAP-001: CLI Command Dispatch And Help

**Contract status:** Implementation-ready

**Goal:** A local operator can discover and invoke the current command surface
without triggering unavailable behavior.

**Actors or callers:** Local shell user or process invoking `adb-dashboard`.

**Inputs and preconditions:**

- Executable invocation with no subcommand.
- Subcommands: `serve`, `version`, `doctor`.
- Options: `--listen ADDRESS`, `--data-dir PATH`, `--temp-dir PATH`,
  `--config PATH`, `--log-level LEVEL`, `--open`, `--no-open`, `--read-only`,
  `--version`, `--help`.
- Unknown commands, unknown options, missing option arguments, and invalid
  option values.

**Successful behavior:** Documented commands dispatch to their specified
capability. `--help`, `--version`, and `version` exit without server startup.
No subcommand follows `CAP-003` server startup behavior.

**Output, response, event, or visible state:** Help, version, diagnostics,
stdout, stderr, and exit statuses match the contracts below.

**Required side effects:** None for help, version, invalid command, invalid
option, missing argument, or invalid option value.

**Forbidden side effects:** Discovery-only and invalid invocations must not bind
a listener, open a browser, execute `adb`, start the ADB server, mutate device
state, create startup directories, write project or artifact state, or open an
interactive session.

**Errors and negative behavior:**

- Unknown command writes `unknown command: NAME` to stderr and exits `2`.
- Unknown option writes `unknown option: --NAME` to stderr and exits `2`.
- Missing argument writes `missing argument for --NAME` to stderr and exits `2`.
- Invalid option values use the matching configuration diagnostic from
  `CAP-004` and exit `2`.
- Invalid invocations write no requested-action stdout.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** CLI
validation completes before server startup, filesystem startup side effects,
external process execution, or browser-open attempts.

**Security, authorization, privacy, and data safety:** CLI diagnostics must not
include tokens, environment dumps, device identifiers, or unrelated filesystem
content.

**Compatibility and versioning:** Current command names, option names, help
sections, and exit-status meanings are stable for this specification version.

#### Acceptance Criteria

- `AC-001-001`: Given `adb-dashboard --help`, when invoked, then stdout contains
  the help text below, stderr is empty, exit status is `0`, and no server is
  started.
- `AC-001-002`: Given an unknown command, unknown option, or missing option
  argument, when invoked, then stderr contains the documented diagnostic, stdout
  contains no requested-action output, exit status is `2`, and no server is
  started.
- `AC-001-003`: Given `adb-dashboard` with no subcommand, when invoked with
  valid startup configuration, then it follows the same startup contract as
  `adb-dashboard serve`.

#### Observable Acceptance Boundary

- Primary boundary: CLI process invocation.
- Focused evidence expected: automated process tests asserting stdout, stderr,
  exit status, and absence of server startup where applicable.
- Real-path exercise: run the built executable with `--help`, an invalid
  invocation, and no subcommand.
- Required environment: Linux host capable of running the built executable.
- Permitted deterministic substitutes: isolated temporary environment for path
  and option tests.
- Evidence that does not count: command-name existence tests or help snapshots
  without exit-status and side-effect assertions.

#### Blocking Open Questions

- None.

### CAP-002: Version Reporting

**Contract status:** Implementation-ready

**Goal:** A local operator or script can read dashboard build metadata without
starting the server or contacting ADB.

**Actors or callers:** Local shell user or process invoking
`adb-dashboard version` or `adb-dashboard --version`.

**Inputs and preconditions:** The executable is available. No ADB executable,
device, browser, config file, or writable state directory is required.

**Successful behavior:** The command writes build information lines to stdout in
stable order and exits `0`.

**Output, response, event, or visible state:** stdout contains these labels in
order: `adb-dashboard`, `commit:`, `buildDate:`, `goVersion:`,
`frontendRevision:`. stderr is empty.

**Required side effects:** None.

**Forbidden side effects:** The command must not start the server, open a
browser, create state directories, execute `adb`, start the ADB server, or read
device state.

**Errors and negative behavior:** Invocation errors before command dispatch use
`CAP-001`.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** The command
performs no long-running lifecycle work.

**Security, authorization, privacy, and data safety:** Output must not include
tokens, environment values, device identifiers, or host filesystem paths except
build metadata explicitly compiled into the binary.

**Compatibility and versioning:** Labels and line order are stable; values may
vary by build.

#### Acceptance Criteria

- `AC-002-001`: Given `adb-dashboard version`, when invoked, then stdout
  contains the five documented labels in order, stderr is empty, exit status is
  `0`, and no server is started.
- `AC-002-002`: Given `adb-dashboard --version`, when invoked, then stdout,
  stderr, exit status, and side effects match `adb-dashboard version`.

#### Observable Acceptance Boundary

- Primary boundary: CLI process invocation.
- Focused evidence expected: automated process test asserting stdout labels,
  stderr, exit status, and no listener creation.
- Real-path exercise: run the built executable with `version` and `--version`.
- Required environment: Linux host capable of running the built executable.
- Permitted deterministic substitutes: development build metadata values.
- Evidence that does not count: package variable inspection without process
  invocation.

#### Blocking Open Questions

- None.

### CAP-003: Local Server Startup And Shutdown

**Contract status:** Implementation-ready

**Goal:** A local operator can start the dashboard on a loopback listener and
stop it cleanly.

**Actors or callers:** Local shell user or process invoking `adb-dashboard` or
`adb-dashboard serve`.

**Inputs and preconditions:** Resolved configuration provides a listen address,
browser-open setting, read-only setting, and startup filesystem paths. The host
can bind the requested loopback address.

**Successful behavior:** The server binds only to a loopback IP address or
`localhost`, serves current HTTP and browser behavior, writes the startup
diagnostic to stderr, and writes no stdout.

**Output, response, event, or visible state:** Startup diagnostic:

```text
TIMESTAMP INFO server started addr=HOST:PORT
```

`TIMESTAMP` is RFC 3339 with numeric timezone offset. `HOST:PORT` is the actual
listener address. When port `0` is configured, the diagnostic reports the
operating-system-assigned port.

Shutdown diagnostic:

```text
TIMESTAMP INFO server stopped signal=SIGNAL
```

**Required side effects:** The process listens on the resolved loopback address
and creates or validates the current startup directories defined by `CAP-005`.

**Forbidden side effects:** Startup must not execute `adb`, start the ADB
server, enumerate devices, run optional host tools, mutate device state, write
project or artifact state, or expose out-of-scope API routes.

**Errors and negative behavior:**

- A non-loopback listen host fails before listening with stderr
  `invalid configuration: server.listen must use a loopback host: HOST` and
  exit status `2`.
- An unavailable loopback address fails with stderr
  `listen address unavailable: HOST:PORT: DETAIL` and exit status `4`.
- Browser-open failure after successful startup writes one warning and does not
  stop the server.

Browser-open failure diagnostic:

```text
TIMESTAMP WARN browser open failed url=URL error=DETAIL
```

`URL` is the attempted dashboard URL. `DETAIL` is the process-start or
nonzero-exit error text.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Startup
filesystem validation occurs before server startup is reported. If browser
opening is enabled, the browser-open attempt occurs only after server startup
is reported and uses the actual bound dashboard URL. `SIGINT` and `SIGTERM`
initiate shutdown, stop accepting new requests, write the shutdown diagnostic,
and exit `0`.

**Security, authorization, privacy, and data safety:** Host and origin checks
remain backend-authoritative for browser and API requests.

**Compatibility and versioning:** The default listener is `127.0.0.1:8080`.
Exit-status meanings remain stable for current startup failures.

#### Acceptance Criteria

- `AC-003-001`: Given a valid loopback listen address, when
  `adb-dashboard serve` starts, then it listens on that address, writes the
  startup diagnostic to stderr, writes no stdout, and serves current HTTP
  routes.
- `AC-003-002`: Given a non-loopback listen host, when startup is requested,
  then the process fails before listening, writes the invalid-configuration
  diagnostic to stderr, and exits with status `2`.
- `AC-003-003`: Given an unavailable loopback address, when startup is
  requested, then the process fails with the listen-unavailable diagnostic and
  exits with status `4`.
- `AC-003-004`: Given a running server, when it receives `SIGINT` or `SIGTERM`,
  then it stops accepting new requests, writes the shutdown diagnostic to
  stderr, and exits with status `0`.
- `AC-003-005`: Given `--open` is provided and startup succeeds, when the server
  reports startup, then exactly one `xdg-open` attempt is made with the actual
  dashboard URL.
- `AC-003-006`: Given `--open` is provided and the browser-open attempt fails
  after successful startup, when the failure is detected, then the server
  continues running and writes one warning diagnostic to stderr without writing
  stdout.

#### Observable Acceptance Boundary

- Primary boundary: process invocation plus HTTP request through the bound
  listener.
- Focused evidence expected: automated process test asserting listener
  behavior, stderr diagnostics, stdout, exit status, and signal shutdown.
- Real-path exercise: run the built server on `127.0.0.1:0`, request a current
  HTTP route, and terminate it with `SIGTERM`.
- Required environment: Linux host with available loopback networking.
- Permitted deterministic substitutes: isolated temporary startup directories,
  OS-assigned port `0`, and controlled `PATH` for `xdg-open`.
- Evidence that does not count: route registration tests without a running
  process.

#### Blocking Open Questions

- None.

### CAP-004: Configuration Discovery, Precedence, And Validation

**Contract status:** Implementation-ready

**Goal:** A local operator can configure current behavior through files,
environment variables, and command-line options with deterministic precedence.

**Actors or callers:** Local shell user, process environment, and TOML
configuration files.

**Inputs and preconditions:** Recognized current keys, environment variables,
and command-line options are documented in this capability.

**Successful behavior:** Built-in defaults, default configuration files,
explicit configuration files, environment variables, and command-line options
are merged in documented priority order.

**Output, response, event, or visible state:** Resolved values affect server
startup, doctor output, status output, browser-open behavior, read-only state,
log-level validation, and path selection as documented.

**Required side effects:** Missing default configuration files are ignored.
Explicit configuration files are read when provided.

**Forbidden side effects:** Configuration resolution must not execute `adb`,
start the server before validation completes, create startup directories before
validation completes, create project or artifact state, or silently accept
invalid current values.

**Errors and negative behavior:** Configuration failures write exactly one
diagnostic line to stderr, write no requested-action stdout, exit `2`, do not
start the server, do not open a browser, and do not create startup directories.

- `ERR-004-001`: Malformed TOML writes
  `invalid configuration: cannot parse PATH: DETAIL`.
- `ERR-004-002`: An explicit configuration file that cannot be opened or read
  writes `invalid configuration: cannot load PATH: DETAIL`.
- `ERR-004-003`: An unknown configuration key writes
  `invalid configuration: unknown key PATH`.
- `ERR-004-004`: A `server.listen` value that is not `HOST:PORT` writes
  `invalid configuration: server.listen must be HOST:PORT`.
- `ERR-004-005`: A `server.listen` value with an empty host writes
  `invalid configuration: server.listen host must not be empty`.
- `ERR-004-006`: A `server.listen` value with a non-integer port or a port
  outside `0` through `65535` writes
  `invalid configuration: server.listen port must be 0 through 65535: VALUE`.
- `ERR-004-007`: An invalid `logging.level` value writes
  `invalid configuration: logging.level must be one of error, warn, info, debug, trace: VALUE`.
- `ERR-004-008`: A TOML value with the wrong current type writes
  `invalid configuration: KEY must be TYPE`.
- `ERR-004-009`: A path value containing unsupported `~`, `$`, or `${` text
  writes `invalid configuration: KEY contains unsupported path expansion: VALUE`.
- `ERR-004-010`: A path value referencing an unset or empty environment variable
  writes `invalid configuration: KEY references unset environment variable: VAR`.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Resolution
and validation complete before server startup, doctor report generation, browser
opening, or startup directory creation.

**Security, authorization, privacy, and data safety:** Diagnostics must not dump
environment variables, full configuration contents, tokens, or device
identifiers.

**Compatibility and versioning:** Current key names, environment variable names,
defaults, precedence, validation rules, and relative-path resolution against the
dashboard process working directory are stable for this specification version.

#### Configuration Sources And Precedence

Values are merged from lowest to highest priority:

1. Built-in defaults.
2. `/etc/adb-dashboard/config.toml`, if present.
3. `$XDG_CONFIG_HOME/adb-dashboard/config.toml`, or
   `$HOME/.config/adb-dashboard/config.toml`, if present.
4. File named by `ADB_DASHBOARD_CONFIG`, if set.
5. File named by `--config PATH`, if provided.
6. Environment variables.
7. Command-line value options.

When both `ADB_DASHBOARD_CONFIG` and `--config PATH` are provided, the file
named by `--config PATH` has higher priority.

#### Current Settings

| Setting | TOML key | Environment variable | Command-line option | Default | Observable value |
|---|---|---|---|---|---|
| Listen address | `server.listen` | `ADB_DASHBOARD_LISTEN` | `--listen ADDRESS` | `127.0.0.1:8080` | startup stderr, `doctor`, status `server.bind` |
| Open browser | `server.open_browser` | none | `--open`, `--no-open` | `false` | `xdg-open` attempt or no attempt during startup |
| Read-only mode | `server.read_only` | none | `--read-only` | `false` | `doctor`, status `server.readOnly`, UI |
| Data directory | `server.data_dir` | `ADB_DASHBOARD_DATA_DIR` | `--data-dir PATH` | `$XDG_STATE_HOME/adb-dashboard` or `$HOME/.local/state/adb-dashboard` | filesystem side effect, `doctor` |
| Temp directory | `server.temp_dir` | `ADB_DASHBOARD_TEMP_DIR` | `--temp-dir PATH` | `$XDG_RUNTIME_DIR/adb-dashboard` or `$TMPDIR/adb-dashboard-$UID` or `/tmp/adb-dashboard-$UID` | filesystem side effect, `doctor` |
| Log level | `logging.level` | `ADB_DASHBOARD_LOG_LEVEL` | `--log-level LEVEL` | `info` | `doctor` `logLevel=LEVEL` field |

Configuration files use TOML:

```toml
[server]
listen = "127.0.0.1:8080"
open_browser = false
read_only = false
data_dir = "$XDG_STATE_HOME/adb-dashboard"
temp_dir = "$XDG_RUNTIME_DIR/adb-dashboard"

[logging]
level = "info"
```

Path values expand only these forms before validation and filesystem use:

- a leading `~/` expands to `$HOME/`;
- `$VAR` and `${VAR}` expand when `VAR` matches
  `[A-Za-z_][A-Za-z0-9_]*`;
- all other `~`, `$`, and `${` text is invalid configuration.

If a referenced environment variable is unset or empty, path validation fails
with status `2` before directory creation. Relative paths are resolved against
the dashboard process working directory before directory creation or validation.

#### Acceptance Criteria

- `AC-004-001`: Given defaults, default files, explicit files, environment
  variables, and command-line options all provide different values for a setting
  that each source can define, when `adb-dashboard doctor` or
  `adb-dashboard serve` is invoked, then the observed value comes from the
  highest-priority source allowed for that setting.
- `AC-004-002`: Given missing default configuration files, when configuration is
  resolved, then resolution continues without an error for those missing files.
- `AC-004-003`: Given a missing explicit configuration file from
  `ADB_DASHBOARD_CONFIG` or `--config`, when invoked, then the command writes a
  matching `ERR-004-002` diagnostic to stderr, writes no requested-action
  stdout, and exits with status `2`.
- `AC-004-004`: Given a malformed explicit or default configuration file, when
  configuration is resolved, then the command writes a matching `ERR-004-001`
  diagnostic to stderr, writes no requested-action stdout, and exits with
  status `2`.
- `AC-004-005`: Given an unknown configuration key, when the file is loaded,
  then the command writes a matching `ERR-004-003` diagnostic to stderr, writes
  no requested-action stdout, and exits with status `2`.
- `AC-004-006`: Given invalid current values, when configuration is validated,
  then the command writes the matching `ERR-004-004` through `ERR-004-010`
  diagnostic to stderr, writes no requested-action stdout, performs no startup
  directory creation, and exits with status `2`.

#### Observable Acceptance Boundary

- Primary boundary: CLI process invocation with isolated configuration files
  and environment.
- Focused evidence expected: automated process tests asserting resolved
  diagnostics or status output, stderr, stdout, exit status, and absence of
  startup after fatal validation.
- Real-path exercise: run `adb-dashboard doctor` and `adb-dashboard serve` with
  isolated config files, environment overrides, and CLI overrides.
- Required environment: Linux host with readable temporary TOML files.
- Permitted deterministic substitutes: temporary home, XDG, state, and runtime
  directories.
- Evidence that does not count: parser-unit tests without command behavior.

#### Blocking Open Questions

- None.

### CAP-005: Startup Filesystem Directories

**Contract status:** Implementation-ready

**Goal:** Current startup and diagnostics can create or validate only the
directories required by the local bootstrap contract.

**Actors or callers:** `adb-dashboard`, `adb-dashboard serve`, and
`adb-dashboard doctor`.

**Inputs and preconditions:** Resolved `server.data_dir` and `server.temp_dir`
paths are available from configuration resolution.

**Successful behavior:** Missing data and temp directories are created.
Existing directories are accepted.

**Output, response, event, or visible state:** Server startup reports success
only after directory creation or validation. `doctor` reports `PASS` or `FAIL`
rows for `dataDir` and `tempDir`.

**Required side effects:** Create only the resolved data and temp directories
when absent.

**Forbidden side effects:** Do not create cache, project, artifact, history,
upload, index, retained-output, or job paths.

**Errors and negative behavior:** Existing non-directory paths or creation
failures produce:

```text
server runtime failure: startup filesystem unavailable: KIND directory PATH: DETAIL
```

Server startup exits `5`. `doctor` reports failing rows and exits `5`. If one
startup directory is created successfully and a later startup directory fails
validation or creation, the created directory is left in place. Startup still
fails before listener binding, browser opening, or startup-success diagnostics.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Directory
creation or validation happens before startup success is reported and before the
server opens a browser.

**Security, authorization, privacy, and data safety:** Directory creation is
limited to resolved current paths. Path expansion follows `CAP-004`.

**Compatibility and versioning:** Default data and temp path derivation follows
the documented XDG and fallback rules in `CAP-004`.

#### Acceptance Criteria

- `AC-005-001`: Given absent resolved data and temp directories, when server
  startup succeeds, then both directories exist before the startup diagnostic is
  written.
- `AC-005-002`: Given existing directories at the resolved paths, when server
  startup runs, then the directories are accepted and startup may proceed.
- `AC-005-003`: Given a resolved data or temp path that exists as a
  non-directory or cannot be created, when server startup runs, then startup
  fails before reporting success, writes the documented diagnostic, and exits
  with status `5`.
- `AC-005-004`: Given a data or temp directory failure during `doctor`, when the
  report can be produced, then the affected row is `FAIL`, unavailable rows
  remain `NIY`, and exit status is `5`.
- `AC-005-005`: Given the data directory is absent and creatable and the temp
  path fails validation or creation, when server startup runs, then the data
  directory may remain, no listener is opened, no browser is opened, the
  documented filesystem diagnostic is written, and exit status is `5`.

#### Observable Acceptance Boundary

- Primary boundary: filesystem operation through CLI process invocation.
- Focused evidence expected: automated process tests using isolated temporary
  paths and asserting created directories, diagnostics, stdout, stderr, and exit
  status.
- Real-path exercise: run `adb-dashboard serve` or `doctor` against temporary
  data and temp paths.
- Required environment: Linux filesystem with temporary writable and
  intentionally unwritable or non-directory paths.
- Permitted deterministic substitutes: isolated temporary directories.
- Evidence that does not count: direct helper tests that bypass CLI startup or
  doctor behavior.

#### Blocking Open Questions

- None.

### CAP-006: Doctor Diagnostics

**Contract status:** Implementation-ready

**Goal:** A local operator can inspect current local readiness and read-only ADB
availability without starting long-running dashboard services or mutating
devices.

**Actors or callers:** Local shell user or process invoking
`adb-dashboard doctor`.

**Inputs and preconditions:** Configuration can be resolved or fails before
report generation. The host filesystem allows or denies current startup
directory checks.

**Successful behavior:** `doctor` writes the current diagnostic report to
stdout, writes no stderr when the report can be produced, and exits `0` when all
required current checks pass.

**Output, response, event, or visible state:** Field names, line order, status
words, `NIY` function names, and `logLevel=LEVEL` are stable. Host paths, ADB
version text, and error details are host-specific.

Successful report shape:

```text
adb-dashboard doctor
overall: PASS
config: PASS source=SOURCE listen=ADDRESS readOnly=BOOL logLevel=LEVEL
dataDir: PASS path=PATH
tempDir: PASS path=PATH
cacheDir: NIY storage.cache is not implemented yet
projectDir: NIY storage.projects is not implemented yet
adbExecutable: PASS path=PATH
adbVersion: PASS version=VERSION
adbServer: NIY adb.server is not implemented yet
devices: NIY devices.refresh is not implemented yet
hostTools: NIY hosttools.discovery is not implemented yet
```

Failure report shape when the report can be produced:

```text
adb-dashboard doctor
overall: FAIL
config: PASS source=SOURCE listen=ADDRESS readOnly=BOOL logLevel=LEVEL
dataDir: FAIL path=PATH error=DETAIL
tempDir: PASS path=PATH
cacheDir: NIY storage.cache is not implemented yet
projectDir: NIY storage.projects is not implemented yet
adbExecutable: PASS path=PATH
adbVersion: PASS version=VERSION
adbServer: NIY adb.server is not implemented yet
devices: NIY devices.refresh is not implemented yet
hostTools: NIY hosttools.discovery is not implemented yet
```

`SOURCE` is one of `defaults`, `file:PATH`, `env`, `cli`, or `mixed`. `ADDRESS`
is the resolved listen address. `BOOL` is `true` or `false`. `LEVEL` is the
resolved log level. If both `dataDir` and `tempDir` fail, both rows report
`FAIL`.

**Required side effects:** Create or validate the resolved data and temp
directories using `CAP-005`. Perform ADB discovery and version checks using
`CAP-010`.

**Forbidden side effects:** `doctor` must not execute ADB commands other than
`adb version`, explicitly start the ADB server, enumerate devices, bind the
listen address, open a browser, open WebSocket sessions, run optional host
tools, start long-running processes, change device state, execute analysis
tools, or modify analysis projects.

**Errors and negative behavior:** Configuration parse and validation failures
write the same stderr diagnostics and exit status as other commands and produce
no stdout report. Filesystem failures produce a report with `overall: FAIL` and
exit status `5`. ADB discovery or version failures produce a report with
`overall: FAIL` and exit status `3` when no higher-priority current failure is
present.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** The report
is produced after configuration resolution and current filesystem checks
complete.

**Security, authorization, privacy, and data safety:** The report must not
include tokens, environment dumps, command arguments, device identifiers, or
resolved executable paths except fields explicitly documented.

**Compatibility and versioning:** Current row names, order, and `NIY` function
names are stable for this specification version.

#### Acceptance Criteria

- `AC-006-001`: Given valid configuration, writable current directories, and
  successful ADB version discovery, when `adb-dashboard doctor` runs, then
  stdout matches the successful report shape, stderr is empty, exit status is
  `0`, required current rows are `PASS`, and unavailable rows outside ADB
  inventory are `NIY`.
- `AC-006-002`: Given one or both current directory checks fail, when `doctor`
  runs, then stdout reports `overall: FAIL`, each failed row is `FAIL`,
  ADB rows follow `CAP-010`, unavailable non-ADB rows remain `NIY`, stderr is
  empty, and exit status is `5`.
- `AC-006-003`: Given invalid configuration, when `doctor` runs, then no stdout
  report is produced, stderr contains the configuration diagnostic, and exit
  status is `2`.

#### Observable Acceptance Boundary

- Primary boundary: CLI process invocation with filesystem observation.
- Focused evidence expected: automated process tests asserting stdout report,
  stderr, exit status, created directories, and forbidden external actions.
- Real-path exercise: run `adb-dashboard doctor` with isolated state and
  runtime paths.
- Required environment: Linux host with temporary filesystem paths.
- Permitted deterministic substitutes: isolated environment variables and
  temporary directories.
- Evidence that does not count: report formatter tests without command
  invocation and filesystem side-effect assertions.

#### Blocking Open Questions

- None.

### CAP-007: Embedded Browser Shell

**Contract status:** Implementation-ready

**Goal:** A local browser can load the current dashboard shell and see backend
status without active controls for unavailable behavior.

**Actors or callers:** Browser loaded from the dashboard server.

**Inputs and preconditions:** The server is running on a loopback listener and
serves embedded frontend assets from the same process.

**Successful behavior:** The browser shell loads from the local server, requests
`/api/v1/bootstrap`, requests the returned `statusUrl`, and renders the current
visible state from backend responses.

**Output, response, event, or visible state:** The loaded page visibly contains:

- application name `adb-dashboard`;
- server state `running`;
- bind address equal to status `server.bind`;
- read-only state `read-only: true` or `read-only: false`, matching status
  `server.readOnly`;
- ADB state `adb: available`, `adb: unavailable`, or `adb: error` matching
  backend status;
- watcher state `watcher: NIY`;
- jobs state `jobs: NIY`;
- sessions state `sessions: NIY`;
- storage state `storage: NIY`;
- host tools state `host tools: NIY`.

**Required side effects:** Browser requests are made to current backend routes
only.

**Forbidden side effects:** The UI must not expose active controls, forms, or
links for device operations, raw command execution, sessions, jobs, transfers,
artifact analysis, host-tool execution, or destructive actions.

**Errors and negative behavior:** If `/api/v1/bootstrap` or `/api/v1/status`
fails, the shell must show `server: unavailable` and must not show success for
any unavailable subsystem. If a future page or control is visible before its
behavior is specified and selected, the shell violates this contract.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Frontend
state is derived from backend responses after page load. The status request uses
the `statusUrl` returned by bootstrap.

**Security, authorization, privacy, and data safety:** Backend security and
capability checks remain authoritative even when the UI hides or disables a
control. The UI must not render bootstrap token values.

**Compatibility and versioning:** Current visible labels and state values are
stable for this specification version.

#### Acceptance Criteria

- `AC-007-001`: Given a running server, when the root browser shell is loaded,
  then the page renders the documented application, server, read-only, and
  subsystem states derived from `/api/v1/status`.
- `AC-007-002`: Given out-of-scope workflows are not specified, when the current
  shell is loaded, then controls, forms, and links for those workflows are
  absent and cannot report success.
- `AC-007-003`: Given `/api/v1/bootstrap` or `/api/v1/status` fails during page
  load, when the shell renders, then it shows `server: unavailable` and does not
  show success for ADB, watcher, jobs, sessions, storage, or host tools.

#### Observable Acceptance Boundary

- Primary boundary: browser load through the running server.
- Focused evidence expected: browser automation or HTTP asset test plus visible
  state assertions for backend-provided status and absence of active
  out-of-scope controls.
- Real-path exercise: start the server, open the root page, and inspect visible
  state.
- Required environment: Linux host with modern JavaScript-capable browser or
  deterministic browser automation.
- Permitted deterministic substitutes: browser automation against
  `127.0.0.1:0`.
- Evidence that does not count: static component rendering without backend
  responses.

#### Blocking Open Questions

- None.

### CAP-008: Browser Security Bootstrap

**Contract status:** Implementation-ready

**Goal:** The embedded browser receives per-process bootstrap tokens while the
backend rejects foreign browser/API requests before handlers run.

**Actors or callers:** Browser or HTTP client requesting `/api/v1/bootstrap` or
other current `/api/v1` routes.

**Inputs and preconditions:** The server is running on a loopback or `localhost`
request target.

**Successful behavior:** `GET /api/v1/bootstrap` returns same-origin bootstrap
JSON with `csrfToken`, `webSocketToken`, and `statusUrl`.

**Output, response, event, or visible state:** Response status is `200`.
Content type is `application/json`. JSON shape:

```json
{
  "csrfToken": "...",
  "webSocketToken": "...",
  "statusUrl": "/api/v1/status"
}
```

`csrfToken` and `webSocketToken` are independent random values generated for
the server process. Each token is encoded as at least 32 URL-safe base64
characters. Tokens issued after a server restart must differ from tokens issued
by the prior process. No current route consumes these tokens. `statusUrl` is
exactly `/api/v1/status`.

**Required side effects:** Tokens are issued for the running process and are not
reused after server restart.

**Forbidden side effects:** Tokens must not be logged, included in status
responses, rendered by the UI, treated as local-user authentication
credentials, or exposed to foreign-origin requests.

**Errors and negative behavior:** Browser and API requests with foreign Host,
absolute-form URL host, or Origin are rejected with the API error envelope and
status `403` before reaching API handlers. Foreign Host rejection uses
`error.code` `forbidden_host`; foreign absolute-form URL host rejection uses
`error.code` `forbidden_absolute_url_host`; foreign Origin rejection uses
`error.code` `forbidden_origin`. Rejected requests return only the standard API
error envelope and do not include bootstrap token fields, status fields,
route-specific response fields, or route-specific side effects.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Tokens are
created per process and invalidated by server restart.

**Security, authorization, privacy, and data safety:** Host and origin rejection
is enforced by the backend before API handler execution.

**Compatibility and versioning:** Current bootstrap response field names,
token-format minimum, and `statusUrl` are stable for this specification version.
Future mutating and WebSocket token enforcement remains out of scope.

#### Acceptance Criteria

- `AC-008-001`: Given a same-origin loopback request, when
  `GET /api/v1/bootstrap` is requested, then the response has status `200`,
  content type `application/json`, non-empty token fields satisfying the
  documented token format, and `statusUrl` equal to `/api/v1/status`.
- `AC-008-002`: Given a foreign Host, absolute-form URL host, or Origin, when a
  current `/api/v1` route is requested, then the response has status `403`,
  content type `application/json`, the documented security error envelope and
  matching `error.code`, contains no route-specific response fields, and
  performs no route-specific side effects.
- `AC-008-003`: Given a server restart, when bootstrap is requested before and
  after restart, then the issued `csrfToken` and `webSocketToken` values differ
  from the token values issued by the prior server process.
- `AC-008-004`: Given status is requested, when the response is inspected, then
  it contains no `csrfToken` or `webSocketToken` fields at any level.

#### Observable Acceptance Boundary

- Primary boundary: HTTP request through production routing.
- Focused evidence expected: automated HTTP tests asserting response status,
  JSON fields, content type, token format, token lifecycle, and handler bypass
  for rejected requests.
- Real-path exercise: start the server and request `/api/v1/bootstrap` with
  accepted and rejected Host/Origin combinations.
- Required environment: Linux host with loopback networking.
- Permitted deterministic substitutes: OS-assigned loopback port.
- Evidence that does not count: token generator tests without HTTP routing and
  rejection checks.

#### Blocking Open Questions

- None.

### CAP-009: Status API

**Contract status:** Implementation-ready

**Goal:** A browser or external developer tool can read current dashboard and
ADB summary status without false success for unavailable future subsystems.

**Actors or callers:** Browser UI or HTTP client requesting
`GET /api/v1/status`.

**Inputs and preconditions:** The server is running and the request passes
current host and origin policy.

**Successful behavior:** The route returns the documented JSON object with
current application, server, and ADB summary fields plus `NIY` status for
unavailable subsystems.

**Output, response, event, or visible state:** Response status is `200`.
Content type is `application/json`. JSON shape:

```json
{
  "application": {
    "name": "adb-dashboard",
    "version": "...",
    "commit": "...",
    "buildDate": "...",
    "goVersion": "...",
    "frontendRevision": "..."
  },
  "server": {
    "status": "running",
    "uptimeSeconds": 0,
    "readOnly": false,
    "bind": "127.0.0.1:8080"
  },
  "adb": {
    "status": "available",
    "executable": "PATH",
    "version": "VERSION",
    "serverResponsive": "NIY"
  },
  "watcher": {
    "status": "NIY",
    "lastSuccessfulPoll": null
  },
  "jobs": {
    "status": "NIY",
    "active": 0,
    "retained": 0
  },
  "sessions": {
    "status": "NIY",
    "active": 0
  },
  "storage": {
    "status": "NIY"
  },
  "hostTools": {
    "status": "NIY",
    "available": 0,
    "unavailable": 0
  }
}
```

Normative current status schema:

| JSON path | Type | Required value or constraint |
|---|---|---|
| `application.name` | string | exactly `adb-dashboard` |
| `application.version` | string | non-empty build value |
| `application.commit` | string | non-empty build value |
| `application.buildDate` | string | non-empty build value |
| `application.goVersion` | string | non-empty build value |
| `application.frontendRevision` | string | non-empty build value |
| `server.status` | string | exactly `running` |
| `server.uptimeSeconds` | integer | `0` or greater |
| `server.readOnly` | boolean | resolved read-only setting |
| `server.bind` | string | actual bound `HOST:PORT` |
| `adb.status` | string | `available`, `unavailable`, or `error` as defined by `CAP-010` |
| `adb.executable` | string or null | resolved path when available or error after executable discovery; otherwise `null` |
| `adb.version` | string or null | parsed version when available; otherwise `null` |
| `adb.serverResponsive` | string | exactly `NIY` |
| `watcher.status` | string | exactly `NIY` |
| `watcher.lastSuccessfulPoll` | null | exactly `null` |
| `jobs.status` | string | exactly `NIY` |
| `jobs.active` | integer | exactly `0` |
| `jobs.retained` | integer | exactly `0` |
| `sessions.status` | string | exactly `NIY` |
| `sessions.active` | integer | exactly `0` |
| `storage.status` | string | exactly `NIY` |
| `hostTools.status` | string | exactly `NIY` |
| `hostTools.available` | integer | exactly `0` |
| `hostTools.unavailable` | integer | exactly `0` |

Additional top-level fields and additional fields inside current status objects
are forbidden until a later capability specifies them.

**Required side effects:** The route may perform ADB discovery and version
checks defined by `CAP-010`.

**Forbidden side effects:** The route must not execute ADB commands other than
`adb version`, start watchers, start jobs or sessions, run optional host tools,
create project or artifact state, or report unavailable subsystems as
successful.

**Errors and negative behavior:** Unknown API routes return the standard error
envelope with status `404`. Host or origin policy rejection returns the error
envelope with status `403`.

Standard unknown-route envelope:

```json
{
  "error": {
    "code": "not_found",
    "message": "Unknown API route",
    "details": {},
    "requestId": null
  }
}
```

Standard browser-security rejection envelope:

```json
{
  "error": {
    "code": "forbidden_host",
    "message": "Request rejected by dashboard browser security policy",
    "details": {},
    "requestId": null
  }
}
```

For security rejections, `error.code` is replaced with the matching code from
`CAP-008`.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** The response
reflects the running process state at request time.

**Security, authorization, privacy, and data safety:** The status response must
not include host filesystem paths, token values, environment variables, command
arguments, or resolved executable paths except fields explicitly documented.

**Compatibility and versioning:** The current response fields and `NIY` values
are stable until a future capability replaces a field with specified behavior.

#### Acceptance Criteria

- `AC-009-001`: Given a running server, when `GET /api/v1/status` is requested
  through an accepted same-origin loopback request, then the response has status
  `200`, content type `application/json`, and the documented JSON shape.
- `AC-009-002`: Given watcher, jobs, sessions, storage, host-tool behavior, or
  ADB server responsiveness is unavailable, when status is requested, then those
  fields report `NIY` or documented unavailable values and do not report
  success.
- `AC-009-003`: Given an unknown `/api/v1` route, when requested, then the
  response has status `404`, content type `application/json`, and the standard
  error envelope.
- `AC-009-004`: Given status is requested, then the JSON does not include token
  values, environment variables, command arguments, or host filesystem paths
  outside fields explicitly documented.

#### Observable Acceptance Boundary

- Primary boundary: HTTP request through production routing.
- Focused evidence expected: automated HTTP tests asserting status code,
  content type, JSON shape, `NIY` fields, forbidden sensitive fields, and
  unknown-route behavior.
- Real-path exercise: start the server and request `/api/v1/status` and an
  unknown `/api/v1` route.
- Required environment: Linux host with loopback networking.
- Permitted deterministic substitutes: OS-assigned loopback port and
  development build metadata values.
- Evidence that does not count: handler construction tests without a running
  route path.

#### Blocking Open Questions

- None.

### CAP-010: ADB Executable And Version Discovery

**Contract status:** Implementation-ready

**Goal:** A local operator can tell whether the dashboard can find an ADB
client and read its version without performing device operations.

**Actors or callers:** Local shell user invoking `adb-dashboard doctor`, browser
UI or HTTP client requesting `GET /api/v1/status`, and dashboard process
startup logic that builds status responses.

**Inputs and preconditions:** The dashboard process has a `PATH` environment
variable. An executable named `adb` may or may not be present on that path. If
present, the executable may return successful, malformed, slow, or failing
output for `adb version`.

**Successful behavior:** The dashboard resolves the first executable named
`adb` from `PATH`, invokes it as `adb version` using an argument vector, parses
the first non-empty stdout line as the display version, and reports ADB
availability through doctor and status surfaces.

**Output, response, event, or visible state:** `doctor` replaces the prior ADB
`NIY` rows with one of these stable forms:

```text
adbExecutable: PASS path=PATH
adbVersion: PASS version=VERSION
adbServer: NIY adb.server is not implemented yet
```

```text
adbExecutable: FAIL error=not found in PATH
adbVersion: NIY adb.version unavailable until adb executable is found
adbServer: NIY adb.server is not implemented yet
```

```text
adbExecutable: PASS path=PATH
adbVersion: FAIL error=DETAIL
adbServer: NIY adb.server is not implemented yet
```

`GET /api/v1/status` replaces the prior `adb` object with:

```json
{
  "status": "available",
  "executable": "PATH",
  "version": "VERSION",
  "serverResponsive": "NIY"
}
```

When `adb` is absent, `adb.status` is `unavailable`, `adb.executable` is
`null`, `adb.version` is `null`, and `adb.serverResponsive` is `NIY`. When
`adb version` fails, `adb.status` is `error`, `adb.executable` is `PATH`,
`adb.version` is `null`, and `adb.serverResponsive` is `NIY`.

**Required side effects:** One bounded `adb version` process may be executed
when an `adb` executable is found.

**Forbidden side effects:** Discovery and version checks must not run shell
commands through interpolation, execute `adb devices`, start explicit ADB server
lifecycle commands, mutate devices, create project or artifact state, open
sessions, or write outside the existing startup directories.

**Errors and negative behavior:** Missing `adb` is reported as unavailable, not
as process failure. A failing or malformed `adb version` command is reported as
an ADB version failure; `doctor` exits `3`, while status remains HTTP `200` and
reports `adb.status` as `error`. The command is killed and reported as failure
if it does not complete within three seconds.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** ADB
discovery occurs after configuration validation and startup-directory checks.
Each status or doctor request may perform fresh discovery. Timed-out or failed
processes are not retried during the same request.

**Security, authorization, privacy, and data safety:** The dashboard must pass
arguments directly as `["version"]`. Diagnostics must not include environment
dumps or command-line arguments beyond the resolved executable path and bounded
error detail.

**Compatibility and versioning:** `adb` path lookup from process `PATH`,
`doctor` row names, status field names, and timeout semantics are stable for
this specification version.

#### Acceptance Criteria

- `AC-010-001`: Given a deterministic `adb` executable on `PATH` whose
  `version` command succeeds, when `adb-dashboard doctor` runs, then stdout
  reports `adbExecutable: PASS path=PATH`, `adbVersion: PASS version=VERSION`,
  `adbServer: NIY`, stderr is empty, and exit status remains `0` when all other
  required current checks pass.
- `AC-010-002`: Given no `adb` executable on `PATH`, when `doctor` runs, then
  stdout reports the documented unavailable ADB rows, stderr is empty, and exit
  status is `3` when all non-ADB required current checks pass.
- `AC-010-003`: Given `adb version` exits nonzero, times out, or produces no
  non-empty stdout line, when `doctor` runs, then stdout reports
  `adbExecutable: PASS`, `adbVersion: FAIL`, `adbServer: NIY`, stderr is empty,
  and exit status is `3`.
- `AC-010-004`: Given the server is running, when `GET /api/v1/status` is
  requested after successful ADB discovery, then the `adb` object reports
  `status: available`, the resolved executable path, parsed version, and
  `serverResponsive: NIY`.
- `AC-010-005`: Given the server is running and ADB is absent or version
  discovery fails, when status is requested, then the `adb` object reports the
  documented unavailable or error state without returning ADB command stderr,
  environment values, or false success.

#### Observable Acceptance Boundary

- Primary boundary: CLI process invocation for `doctor` and HTTP request
  through production routing for status.
- Focused evidence expected: automated process and HTTP tests with isolated
  `PATH` fixtures asserting stdout, stderr, exit status, JSON fields, timeout,
  and forbidden command side effects.
- Real-path exercise: run `adb-dashboard doctor` and a running server's
  `/api/v1/status` against a deterministic fake `adb` executable and against a
  `PATH` without `adb`.
- Required environment: Linux host capable of process execution and loopback
  HTTP.
- Permitted deterministic substitutes: fake `adb` executables in temporary
  directories.
- Evidence that does not count: unit tests of path lookup without invoking
  `doctor` or status routes.

#### Blocking Open Questions

- None.

### CAP-011: Read-Only Device Inventory API

**Contract status:** Implementation-ready

**Goal:** A browser or external local tool can list attached ADB devices through
the dashboard without mutating device state.

**Actors or callers:** Browser UI or HTTP client requesting
`GET /api/v1/devices`.

**Inputs and preconditions:** The server is running, host and origin checks pass,
and ADB discovery from `CAP-010` succeeds. The resolved ADB client may return
successful, malformed, slow, or failing output for `adb devices -l`.

**Successful behavior:** `GET /api/v1/devices` invokes the resolved ADB client
as `adb devices -l`, parses the standard device list output, and returns a JSON
inventory. The route is read-only from the dashboard perspective.

**Output, response, event, or visible state:** Response status is `200`.
Content type is `application/json`. Successful JSON shape:

```json
{
  "adb": {
    "status": "available",
    "executable": "PATH",
    "version": "VERSION"
  },
  "devices": [
    {
      "serial": "SERIAL",
      "state": "device",
      "product": "PRODUCT",
      "model": "MODEL",
      "device": "DEVICE",
      "transportId": "TRANSPORT_ID"
    }
  ]
}
```

The `devices` array is empty when ADB reports no devices. `serial` and `state`
are required strings for each parsed row. `product`, `model`, `device`, and
`transportId` are strings or `null` when omitted by ADB.

When ADB is unavailable, response status is `503` and the standard API error
envelope uses `error.code` `adb_unavailable`. When `adb devices -l` fails,
times out, or produces malformed device rows, response status is `502` and the
standard API error envelope uses `error.code` `adb_devices_failed`.

**Required side effects:** One bounded `adb devices -l` process may be executed
per accepted request. The host ADB client may communicate with its normal local
ADB server as part of that command.

**Forbidden side effects:** The route must not execute any ADB command other
than `devices -l`, mutate devices, select a device, start interactive sessions,
read files from devices, write files to devices, install packages, capture
screenshots, stream logcat, reboot devices, create jobs, persist inventory, or
write project or artifact state.

**Errors and negative behavior:** Host and origin policy rejection runs before
ADB discovery or command execution. Missing ADB, failed version discovery,
failed device listing, timeout after three seconds, and malformed output are
reported with the documented status and error envelope. Error responses contain
no route-specific `devices` array.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** The route
performs ADB discovery before device listing. Timed-out device-list processes
are killed. No retry is attempted during the same request.

**Security, authorization, privacy, and data safety:** ADB is invoked with an
argument vector exactly equivalent to `["devices", "-l"]`. Device serials and
properties are returned only to accepted loopback requests. Error details must
not dump full environment values or unrelated command output.

**Compatibility and versioning:** Route path, response status meanings, JSON
field names, and error codes are stable for this specification version.

#### Acceptance Criteria

- `AC-011-001`: Given a deterministic `adb` executable whose `version` and
  `devices -l` commands succeed with zero device rows, when
  `GET /api/v1/devices` is requested, then the response has status `200`, the
  documented ADB object, and an empty `devices` array.
- `AC-011-002`: Given `adb devices -l` returns one or more device rows with
  attributes, when the route is requested, then the response has status `200`
  and each row is parsed into the documented `serial`, `state`, `product`,
  `model`, `device`, and `transportId` fields.
- `AC-011-003`: Given ADB discovery is unavailable or version discovery fails,
  when the route is requested, then the response has status `503`, error code
  `adb_unavailable`, no `devices` array, and no `adb devices -l` command is
  executed.
- `AC-011-004`: Given `adb devices -l` exits nonzero, times out, or returns a
  malformed row, when the route is requested, then the response has status
  `502`, error code `adb_devices_failed`, no `devices` array, and no requested
  action is reported as successful.
- `AC-011-005`: Given a foreign Host, absolute-form URL host, or Origin, when
  `GET /api/v1/devices` is requested, then the response has status `403`, the
  documented security error envelope, and no ADB executable is invoked.

#### Observable Acceptance Boundary

- Primary boundary: HTTP request through production routing.
- Focused evidence expected: automated HTTP tests with fake ADB executables
  asserting response status, JSON shape, parsed devices, error envelopes,
  timeout, and forbidden command side effects.
- Real-path exercise: start the server with an isolated `PATH` containing a
  deterministic fake `adb`, request `/api/v1/devices` for zero-device,
  multi-device, ADB-unavailable, and command-failure cases, and inspect process
  invocation logs.
- Required environment: Linux host with loopback networking and temporary
  executable fixtures.
- Permitted deterministic substitutes: fake `adb` executables that model ADB
  stdout, stderr, status, and timeout behavior.
- Evidence that does not count: parser-only tests without HTTP routing or ADB
  process invocation checks.

#### Blocking Open Questions

- None.

### CAP-012: Browser ADB And Device Inventory View

**Contract status:** Implementation-ready

**Goal:** A local browser can see ADB availability and read-only device
inventory returned by the backend.

**Actors or callers:** Browser loaded from the dashboard server.

**Inputs and preconditions:** The server is running and serves embedded frontend
assets. `/api/v1/bootstrap`, `/api/v1/status`, and `/api/v1/devices` are
available through same-origin requests.

**Successful behavior:** The browser shell requests bootstrap and status as in
`CAP-007`, then requests `/api/v1/devices` when status reports ADB available,
and renders backend-derived ADB and device state.

**Output, response, event, or visible state:** The loaded page visibly contains:

- ADB state `adb: available`, `adb: unavailable`, or `adb: error` matching
  status.
- A device count matching the `/api/v1/devices` response when device inventory
  succeeds.
- For each returned device, visible serial and state.
- `devices: unavailable` when ADB is unavailable or device inventory fails.

The page must not render ADB executable paths, bootstrap token values, command
stderr, environment values, active mutation controls, forms, or links for
out-of-scope workflows.

**Required side effects:** Browser requests are made only to current backend
routes.

**Forbidden side effects:** The UI must not expose active controls, forms, or
links for shell, logcat, install, uninstall, file transfer, screenshots, input,
reboot, networking, package workflows, artifact analysis, host-tool execution,
or destructive actions.

**Errors and negative behavior:** If device inventory fails, the shell shows
`devices: unavailable` and must not show a stale or invented successful device
list. Bootstrap or status failure behavior remains governed by `CAP-007`.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Frontend
state is derived from backend responses after page load. The device request is
made after status reports ADB availability.

**Security, authorization, privacy, and data safety:** Backend security and
capability checks remain authoritative even when the UI hides or disables a
control.

**Compatibility and versioning:** Current visible labels and state values are
stable for this specification version.

#### Acceptance Criteria

- `AC-012-001`: Given status reports ADB available and `/api/v1/devices`
  returns devices, when the root browser shell is loaded, then the page renders
  `adb: available`, the correct device count, and each device serial and state
  from the backend response.
- `AC-012-002`: Given status reports ADB unavailable or `/api/v1/devices`
  fails, when the shell renders, then it shows `devices: unavailable` and does
  not show a successful or stale device list.
- `AC-012-003`: Given out-of-scope ADB workflows remain unspecified, when the
  shell is loaded, then controls, forms, and links for those workflows are
  absent and cannot report success.
- `AC-012-004`: Given the shell renders ADB and device state, then it does not
  render token values, ADB executable paths, command stderr, or environment
  values.

#### Observable Acceptance Boundary

- Primary boundary: browser load through the running server.
- Focused evidence expected: browser automation or HTTP asset test plus visible
  state assertions for backend-provided ADB and device inventory state.
- Real-path exercise: start the server with a deterministic fake `adb`, open the
  root page, and inspect visible ADB and device state plus absence of forbidden
  controls and sensitive values.
- Required environment: Linux host with loopback networking and a modern
  JavaScript-capable browser or deterministic browser automation.
- Permitted deterministic substitutes: browser automation and fake `adb`
  executables.
- Evidence that does not count: static component rendering without backend
  responses or fake success data unreachable in production.

#### Blocking Open Questions

- None.

### CAP-013: Explicit Device Refresh And Device Detail

**Contract status:** Implementation-ready

**Goal:** A local browser or same-origin HTTP client can explicitly refresh
read-only device inventory and inspect one currently connected device without
mutating device state.

**Actors or callers:** Browser loaded from the dashboard server and local HTTP
clients on the same origin.

**Inputs and preconditions:** The server is running. ADB discovery from
`CAP-010` succeeds. Device inventory from `CAP-011` is available. Device detail
requests include a serial from current inventory using a path segment encoded
for a URL. Browser refresh is initiated by a visible user action.

**Successful behavior:** An explicit refresh requests fresh backend inventory
and replaces the visible device list with the latest returned inventory. Device
detail requests return one device object matching the requested serial from the
fresh inventory source, plus ADB availability metadata.

**Output, response, event, or visible state:**

- `GET /api/v1/devices/{serial}` returns HTTP `200` with JSON:

```json
{
  "adb": {
    "status": "available",
    "executable": "adb",
    "version": "Android Debug Bridge version 1.0.41"
  },
  "device": {
    "serial": "emulator-5554",
    "state": "device",
    "product": "sdk_gphone64_x86_64",
    "model": "sdk_gphone64_x86_64",
    "device": "emu64xa",
    "transportId": "1"
  }
}
```

Optional device properties may be `null` or omitted only when absent from
`adb devices -l` output. The browser shows selected serial, state, optional
inventory properties that are present, refresh success, and refresh/detail
failure states.

**Required side effects:** The backend may execute only `adb version` and
`adb devices -l` for refresh and device detail. Browser refresh updates only
current visible state.

**Forbidden side effects:** Refresh and detail must not select a persistent
active device, mutate a device, start sessions, stream logs, capture
screenshots, write project or artifact state, or create retained inventory.

**Errors and negative behavior:**

- If ADB discovery fails, device detail returns HTTP `503` with code
  `adb_unavailable`.
- If inventory execution fails, times out, or returns malformed output, device
  detail returns HTTP `502` with code `adb_devices_failed`.
- If the serial is absent from the fresh inventory, device detail returns HTTP
  `404` with code `device_not_found`.
- Rejected Host or Origin requests return the standard security error envelope
  before ADB execution.
- Browser refresh or detail failure visibly reports unavailable state and does
  not show stale detail as current.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Each
refresh or detail request performs fresh discovery and inventory work for that
request. Backend ADB commands use bounded timeouts and argument vectors. If a
browser refresh is requested while another refresh is pending, the latest
completed response must not be overwritten by an older response.

**Security, authorization, privacy, and data safety:** Device serials are local
device identifiers and may appear in the browser and same-origin JSON because
inventory already exposes them. Responses must not include bootstrap tokens,
ADB executable paths beyond the documented executable display field, command
stderr, environment values, host filesystem paths, or unsupported workflow
controls.

**Compatibility and versioning:** `GET /api/v1/devices` remains compatible.
`GET /api/v1/devices/{serial}` and browser refresh/detail labels are stable for
this specification version.

#### Acceptance Criteria

- `AC-013-001`: Given a running server with ADB available and two devices, when
  the browser refresh control is activated, then the page requests fresh device
  inventory, renders the latest count and serial/state rows, and exposes no
  mutation controls.
- `AC-013-002`: Given a serial present in fresh inventory, when
  `GET /api/v1/devices/{serial}` is requested from an allowed same-origin
  client, then the response is HTTP `200` with ADB metadata and the matching
  device detail object.
- `AC-013-003`: Given ADB unavailable, inventory failure, malformed inventory,
  timeout, or an absent serial, when device detail is requested, then the
  documented status and error code are returned with no stale successful device
  object.
- `AC-013-004`: Given a rejected Host or Origin request for device detail, when
  the request is handled, then the standard security envelope is returned before
  any ADB process is executed.
- `AC-013-005`: Given the browser renders refresh and device detail states,
  then it does not render token values, command stderr, environment values, host
  filesystem paths, or controls for unsupported workflows.

#### Observable Acceptance Boundary

- Primary boundary: browser interaction and HTTP request through the running
  server.
- Focused evidence expected: automated browser-boundary and HTTP tests with
  deterministic fake ADB fixtures asserting refresh, detail, failures, security
  rejection, and absence of forbidden output.
- Real-path exercise: start the server with fake ADB inventories, load the
  browser shell, activate refresh, open one device detail, request the matching
  API route, then repeat with an absent serial and inventory failure.
- Required environment: Linux host with loopback networking, temporary
  executable fixtures, and a modern JavaScript-capable browser or deterministic
  browser automation.
- Permitted deterministic substitutes: fake ADB executables and browser
  automation.
- Evidence that does not count: static markup checks, mock-only detail
  handlers, or tests that do not prove ADB command restrictions.

#### Blocking Open Questions

- None.

### CAP-014: Read-Only Device Logcat

**Contract status:** Implementation-ready

**Goal:** A local browser or same-origin HTTP client can retrieve bounded
read-only logcat text for one currently connected device.

**Actors or callers:** Browser loaded from the dashboard server and local HTTP
clients on the same origin.

**Inputs and preconditions:** The server is running. ADB discovery succeeds.
The requested serial is present in fresh inventory and has state `device`.
Requests use `GET /api/v1/devices/{serial}/logcat`. Optional query parameters:
`lines` integer `1` through `500`, default `200`; `format` exactly `plain`,
default `plain`.

**Successful behavior:** The backend executes a bounded read-only logcat dump
for the selected serial and returns the last requested number of lines without
persisting output.

**Output, response, event, or visible state:** HTTP `200` JSON:

```json
{
  "device": {
    "serial": "emulator-5554",
    "state": "device"
  },
  "logcat": {
    "format": "plain",
    "lines": ["08-01 10:00:00.000 I Tag: message"],
    "truncated": false
  }
}
```

The browser shows selected serial, log lines, empty-log state, loading state,
and failure state.

**Required side effects:** The backend may execute only `adb version`,
`adb devices -l`, and `adb -s SERIAL logcat -d` with an argument vector.

**Forbidden side effects:** Logcat must not execute shell commands, clear logs,
stream indefinitely, open WebSockets, write retained output, create jobs,
mutate device state, or include command stderr in successful output.

**Errors and negative behavior:**

- Invalid `lines` or `format` returns HTTP `400` with code
  `invalid_logcat_request`.
- ADB unavailable or serial absent follows `CAP-013` errors.
- If the selected device is not in state `device`, return HTTP `409` with code
  `device_not_ready`.
- Logcat command failure, timeout, or stdout that is not valid UTF-8 text
  returns HTTP `502` with code `adb_logcat_failed`.
- Rejected Host or Origin requests return the standard security envelope before
  ADB execution.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Inventory
validation occurs before logcat execution. Logcat execution has a bounded
timeout. Output is retained only in memory for the response and browser state.

**Security, authorization, privacy, and data safety:** Logcat output may contain
application data from the selected device. It is served only through current
same-origin local routes, is not logged by the dashboard, and is not persisted.
Responses must not include bootstrap tokens, command stderr, environment
values, or host filesystem paths.

**Compatibility and versioning:** The route, query parameters, status codes,
and JSON field names are stable for this specification version.

#### Acceptance Criteria

- `AC-014-001`: Given a ready device and logcat output with more than the
  requested number of lines, when the logcat route is requested, then HTTP
  `200` returns only the last requested lines, `truncated: true`, and no command
  stderr or host environment values.
- `AC-014-002`: Given invalid `lines`, invalid `format`, ADB unavailable,
  absent serial, non-ready device state, command failure, stdout that is not
  valid UTF-8 text, or timeout, when logcat is requested, then the documented
  status and error code are returned and no retained output is written.
- `AC-014-003`: Given a rejected Host or Origin request for logcat, when the
  request is handled, then the standard security envelope is returned before
  any ADB process is executed.
- `AC-014-004`: Given the browser opens logcat for a ready device, then it
  renders selected serial, bounded log lines, empty/failure states, and no
  controls for clearing logs, shell, sessions, or streaming.

#### Observable Acceptance Boundary

- Primary boundary: HTTP request and browser interaction through the running
  server.
- Focused evidence expected: process/HTTP and browser-boundary tests with fake
  ADB fixtures for success, truncation, empty output, invalid input,
  unavailable ADB, absent serial, non-ready device, command failure, timeout,
  and security rejection.
- Real-path exercise: start the server with fake ADB logcat output, request the
  logcat API, inspect the browser logcat view, and inspect fake-command logs to
  confirm only allowed commands ran.
- Required environment: Linux host with loopback networking and temporary
  executable fixtures.
- Permitted deterministic substitutes: fake ADB executables and browser
  automation.
- Evidence that does not count: returning fixture logs without a production
  ADB-command path or tests that do not assert command restrictions.

#### Blocking Open Questions

- None.

### CAP-015: Read-Only Device Screenshot Capture

**Contract status:** Implementation-ready

**Goal:** A local browser or same-origin HTTP client can capture and view a
single PNG screenshot from one currently connected device without retaining or
mutating state.

**Actors or callers:** Browser loaded from the dashboard server and local HTTP
clients on the same origin.

**Inputs and preconditions:** The server is running. ADB discovery succeeds.
The requested serial is present in fresh inventory and has state `device`.
Requests use `GET /api/v1/devices/{serial}/screenshot`.

**Successful behavior:** The backend executes a bounded read-only screenshot
capture for the selected serial, validates that stdout is PNG data, and returns
the image bytes without persisting them.

**Output, response, event, or visible state:** HTTP `200` with
`Content-Type: image/png`, no JSON envelope, and a PNG byte stream. The browser
shows selected serial, loading state, latest captured screenshot image, and
failure state.

**Required side effects:** The backend may execute only `adb version`,
`adb devices -l`, and `adb -s SERIAL exec-out screencap -p` with an argument
vector.

**Forbidden side effects:** Screenshot capture must not write image files,
create artifact state, start screen recording, use shell interpolation, mutate
device state, or include command stderr in the image response.

**Errors and negative behavior:**

- ADB unavailable or serial absent follows `CAP-013` errors.
- If the selected device is not in state `device`, return HTTP `409` with code
  `device_not_ready`.
- Screenshot command failure, timeout, empty output, or non-PNG output returns
  HTTP `502` with code `adb_screenshot_failed`.
- Rejected Host or Origin requests return the standard security envelope before
  ADB execution.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Inventory
validation occurs before screenshot execution. Screenshot execution has a
bounded timeout. Image bytes are retained only in memory for the response and
browser state.

**Security, authorization, privacy, and data safety:** Screenshots may contain
sensitive device display data. They are served only through current same-origin
local routes, are not logged by the dashboard, and are not persisted. Error
responses must not include command stderr, environment values, host filesystem
paths, or token values.

**Compatibility and versioning:** The route, status codes, PNG success response,
and error envelopes are stable for this specification version.

#### Acceptance Criteria

- `AC-015-001`: Given a ready device and valid PNG screenshot output, when the
  screenshot route is requested, then HTTP `200` returns `Content-Type:
  image/png` with PNG bytes and no JSON envelope.
- `AC-015-002`: Given ADB unavailable, absent serial, non-ready device state,
  command failure, timeout, empty output, or non-PNG output, when screenshot is
  requested, then the documented status and error code are returned and no image
  file or artifact state is written.
- `AC-015-003`: Given a rejected Host or Origin request for screenshot, when
  the request is handled, then the standard security envelope is returned before
  any ADB process is executed.
- `AC-015-004`: Given the browser captures a screenshot for a ready device,
  then it renders selected serial, loading/failure states, the latest image, and
  no controls for screen recording, file transfer, shell, sessions, or device
  mutation.

#### Observable Acceptance Boundary

- Primary boundary: HTTP request and browser interaction through the running
  server.
- Focused evidence expected: process/HTTP and browser-boundary tests with fake
  ADB fixtures for valid PNG, command failure, timeout, empty output, non-PNG
  output, non-ready device, absent serial, and security rejection.
- Real-path exercise: start the server with fake ADB PNG output, request the
  screenshot API, inspect response headers and bytes, inspect the browser image
  view, and inspect fake-command logs to confirm only allowed commands ran.
- Required environment: Linux host with loopback networking and temporary
  executable fixtures.
- Permitted deterministic substitutes: fake ADB executables and browser
  automation.
- Evidence that does not count: static image fixtures served without the
  production ADB-command path or tests that do not assert no retained file was
  written.

#### Blocking Open Questions

- None.

### CAP-016: Read-Only Package Inventory API

**Contract status:** Implementation-ready

**Goal:** A local browser or same-origin HTTP client can inspect installed
package inventory for one current ready device without installing,
uninstalling, clearing, stopping, or otherwise mutating applications.

**Actors or callers:** Browser loaded from the dashboard server and local HTTP
clients on the same origin.

**Inputs and preconditions:** The server is running. ADB discovery succeeds.
The requested serial is present in fresh inventory and has state `device`.
Requests use `GET /api/v1/devices/{serial}/packages`. Optional query
`scope` is absent or one of `all`, `third-party`, or `system`; absent defaults
to `all`.

**Successful behavior:** The backend executes one bounded read-only package
inventory command for the selected serial, parses recognized package rows, and
returns package names, APK paths when reported, installer/user identifiers when
reported, version code when reported, and the selected scope.

**Output, response, event, or visible state:** HTTP `200` JSON with
`device.serial`, `device.state`, and `packages.scope`, `packages.items`, and
`packages.count`. Items are sorted by package name. Empty inventory returns
HTTP `200` with an empty `items` array and count `0`.

**Required side effects:** The backend may execute only `adb version`,
`adb devices -l`, and one of these argument-vector commands:
`adb -s SERIAL shell pm list packages -f -U --show-versioncode`,
`adb -s SERIAL shell pm list packages -3 -f -U --show-versioncode`, or
`adb -s SERIAL shell pm list packages -s -f -U --show-versioncode`.

**Forbidden side effects:** Package inventory must not install, uninstall,
clear, force-stop, disable, enable, launch, or pull files from the device. It
must not retain command output or include command stderr, host filesystem paths,
environment values, or token values in responses.

**Errors and negative behavior:**

- Invalid `scope` returns HTTP `400` with code `invalid_package_request`.
- ADB unavailable or serial absent follows `CAP-013` errors.
- If the selected device is not in state `device`, return HTTP `409` with code
  `device_not_ready`.
- Package command failure, timeout, malformed output, or output exceeding the
  implementation's bounded limit returns HTTP `502` with code
  `adb_packages_failed`.
- Rejected Host or Origin requests return the standard security envelope before
  ADB execution.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Inventory
validation occurs before package command execution. Package execution has a
bounded timeout. Command output is retained only in memory for the response.

**Security, authorization, privacy, and data safety:** Package names and APK
paths may reveal sensitive app inventory. They are served only through current
same-origin local routes and are not logged or persisted by the dashboard.

**Compatibility and versioning:** The route, query values, status codes, JSON
field names, and error codes are stable for this specification version.

#### Acceptance Criteria

- `AC-016-001`: Given a ready device and parseable package output, when package
  inventory is requested, then HTTP `200` returns sorted package items, count,
  scope, selected device fields, and no sensitive host or token values.
- `AC-016-002`: Given `scope` is absent, `all`, `third-party`, or `system`,
  when inventory is requested, then the documented ADB command variant is used
  and no unsupported package mutation command is invoked.
- `AC-016-003`: Given invalid scope, ADB unavailable, absent serial, non-ready
  device state, command failure, timeout, malformed output, or oversized output,
  when inventory is requested, then the documented status and error code are
  returned and no retained output is written.
- `AC-016-004`: Given a rejected Host or Origin package request, when handled,
  then the standard security envelope is returned before any ADB process is
  executed.

#### Observable Acceptance Boundary

- Primary boundary: HTTP request through the running server.
- Focused evidence expected: process/HTTP tests with fake ADB fixtures for
  valid scopes, empty inventory, malformed output, unavailable ADB, absent
  serial, non-ready device, command failure, timeout, oversized output, and
  security rejection.
- Real-path exercise: start the server with fake ADB package output, request
  package inventory for each scope, inspect JSON and fake-command logs, and
  inspect filesystem side effects.
- Required environment: Linux host with loopback networking and temporary
  executable fixtures.
- Permitted deterministic substitutes: fake ADB executables.
- Evidence that does not count: parser-only tests or static fixtures that do
  not exercise production HTTP routing and ADB command restrictions.

#### Blocking Open Questions

- None.

### CAP-017: Read-Only Package Detail

**Contract status:** Implementation-ready

**Goal:** A local browser or same-origin HTTP client can inspect details for
one package on one current ready device without mutating the package or device.

**Actors or callers:** Browser loaded from the dashboard server and local HTTP
clients on the same origin.

**Inputs and preconditions:** The server is running. ADB discovery succeeds.
The requested serial is present in fresh inventory and has state `device`.
Requests use `GET /api/v1/devices/{serial}/packages/{packageName}`.
`packageName` must be non-empty, at most 255 bytes, and match Java-style package
segments using ASCII letters, digits, and underscores separated by dots.

**Successful behavior:** The backend validates the device and package name,
executes a bounded read-only package detail command, extracts stable fields
when present, and returns both parsed detail fields and bounded raw summary
lines useful to an operator.

**Output, response, event, or visible state:** HTTP `200` JSON with
`device.serial`, `device.state`, `package.name`, optional `package.versionName`,
optional `package.versionCode`, optional `package.installer`, optional
`package.firstInstallTime`, optional `package.lastUpdateTime`, optional
`package.requestedPermissions`, and `package.summaryLines` capped to a bounded
line count.

**Required side effects:** The backend may execute only `adb version`,
`adb devices -l`, and
`adb -s SERIAL shell dumpsys package PACKAGE_NAME` with an argument vector.

**Forbidden side effects:** Package detail must not install, uninstall, clear,
force-stop, disable, enable, launch, pull files, or write retained package
state. It must not include command stderr, host filesystem paths, environment
values, or token values in responses.

**Errors and negative behavior:**

- Invalid package names return HTTP `400` with code `invalid_package_request`.
- ADB unavailable or serial absent follows `CAP-013` errors.
- If the selected device is not in state `device`, return HTTP `409` with code
  `device_not_ready`.
- If the package is not reported by the command, return HTTP `404` with code
  `package_not_found`.
- Package command failure, timeout, malformed output, or output exceeding the
  implementation's bounded limit returns HTTP `502` with code
  `adb_package_detail_failed`.
- Rejected Host or Origin requests return the standard security envelope before
  ADB execution.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Device
validation and package-name validation occur before package detail execution.
Package detail execution has a bounded timeout. Command output is retained only
in memory for the response.

**Security, authorization, privacy, and data safety:** Package details may
reveal installed apps and permissions. They are served only through current
same-origin local routes and are not logged or persisted by the dashboard.

**Compatibility and versioning:** The route, validation, status codes, JSON
field names, and error codes are stable for this specification version.

#### Acceptance Criteria

- `AC-017-001`: Given a ready device, valid package name, and parseable package
  detail output, when package detail is requested, then HTTP `200` returns the
  selected package fields, bounded summary lines, selected device fields, and no
  sensitive host or token values.
- `AC-017-002`: Given invalid package name, ADB unavailable, absent serial,
  non-ready device state, package not found, command failure, timeout,
  malformed output, or oversized output, when package detail is requested, then
  the documented status and error code are returned and no retained output is
  written.
- `AC-017-003`: Given a rejected Host or Origin package detail request, when
  handled, then the standard security envelope is returned before any ADB
  process is executed.
- `AC-017-004`: Given the browser opens package detail for a displayed package,
  then it renders selected serial, package name, loading/failure states, parsed
  detail fields when present, bounded summary lines, and no package mutation,
  shell, file-transfer, or install controls.

#### Observable Acceptance Boundary

- Primary boundary: HTTP request and browser interaction through the running
  server.
- Focused evidence expected: process/HTTP and browser-boundary tests with fake
  ADB fixtures for parseable detail, invalid package names, unavailable ADB,
  absent serial, non-ready device, not found, command failure, timeout,
  oversized output, and security rejection.
- Real-path exercise: start the server with fake ADB package detail output,
  request package detail for one package, inspect JSON and browser state, repeat
  negative cases, and inspect fake-command logs plus filesystem side effects.
- Required environment: Linux host with loopback networking and temporary
  executable fixtures.
- Permitted deterministic substitutes: fake ADB executables and browser
  automation.
- Evidence that does not count: parser-only tests or browser tests that do not
  exercise production backend requests.

#### Blocking Open Questions

- None.

### CAP-018: Local APK Artifact Intake

**Contract status:** Implementation-ready

**Goal:** A local browser or same-origin HTTP client can upload one APK-like
artifact into the dashboard data directory for local analysis without sending
it to external services or touching connected devices.

**Actors or callers:** Browser loaded from the dashboard server and local HTTP
clients on the same origin.

**Inputs and preconditions:** The server is running. The resolved data
directory is available. Requests use `POST /api/v1/artifacts` with
`multipart/form-data` containing exactly one file field named `artifact`.
Accepted filenames end in `.apk` using a case-insensitive comparison. Maximum
accepted file size is 256 MiB. The first bytes must identify a ZIP file.

**Successful behavior:** The backend stores the uploaded file under the
resolved data directory using a generated opaque artifact ID, computes SHA-256
and byte size while streaming, and returns the artifact metadata. The write is
atomic: a failed upload leaves no visible artifact record or partial final
file.

**Output, response, event, or visible state:** HTTP `201` JSON with
`artifact.id`, `artifact.originalName`, `artifact.sizeBytes`,
`artifact.sha256`, `artifact.contentType` when supplied, `artifact.createdAt`,
and `artifact.analysisStatus` equal to `pending`.

**Required side effects:** Create one artifact directory and one stored APK file
under the resolved data directory, plus the minimal metadata needed to list and
retrieve the artifact after process restart.

**Forbidden side effects:** Artifact intake must not execute ADB, install the
APK, open host tools, send network requests, write outside the resolved data
directory, expose host filesystem paths in responses, or retain partial final
files after failure.

**Errors and negative behavior:**

- Missing, duplicate, empty, non-APK, non-ZIP, or oversized artifact uploads
  return HTTP `400` with code `invalid_artifact_upload`.
- Unsupported media type returns HTTP `415` with code
  `unsupported_artifact_media`.
- Filesystem write or metadata persistence failure returns HTTP `507` with code
  `artifact_storage_unavailable`.
- Rejected Host or Origin requests return the standard security envelope before
  reading or storing the body.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Security
checks run before body parsing. File bytes are streamed to a temporary file in
the selected artifact directory and atomically moved into place only after
validation and hashing succeed. Cancellation or failure removes temporary files.

**Security, authorization, privacy, and data safety:** APK artifacts may contain
sensitive application data. They remain local under the configured data
directory, are not logged, and are not transmitted externally.

**Compatibility and versioning:** Artifact IDs are opaque and stable. Metadata
field names, status codes, and error codes are stable for this specification
version.

#### Acceptance Criteria

- `AC-018-001`: Given a valid APK-like ZIP upload, when artifact intake is
  requested, then HTTP `201` returns artifact metadata and the stored artifact
  file plus metadata remain present after process restart.
- `AC-018-002`: Given missing, duplicate, empty, non-APK, non-ZIP, oversized, or
  storage-failing uploads, when artifact intake is requested, then the
  documented status and error code are returned and no partial final artifact is
  visible.
- `AC-018-003`: Given a rejected Host or Origin upload request, when handled,
  then the standard security envelope is returned before the request body is
  stored or any artifact metadata is created.

#### Observable Acceptance Boundary

- Primary boundary: HTTP request through the running server with isolated data
  directory.
- Focused evidence expected: process/HTTP and filesystem tests for successful
  upload, restart persistence, invalid uploads, storage failure, cancellation
  or short body, and security rejection.
- Real-path exercise: start the server with an isolated data directory, upload
  a disposable APK-like ZIP, inspect response metadata, restart the server,
  inspect stored artifact metadata and files, and inspect absence of partial
  files.
- Required environment: Linux host with loopback networking and writable
  temporary filesystem.
- Permitted deterministic substitutes: generated disposable ZIP/APK fixtures.
- Evidence that does not count: metadata constructor tests without HTTP upload
  and filesystem persistence.

#### Blocking Open Questions

- None.

### CAP-019: APK Artifact Metadata Analysis

**Contract status:** Implementation-ready

**Goal:** A local browser or same-origin HTTP client can analyze a stored APK
artifact and view deterministic package metadata without installing it or
sending it to external services.

**Actors or callers:** Browser loaded from the dashboard server and local HTTP
clients on the same origin.

**Inputs and preconditions:** The server is running. The artifact exists.
Requests use `POST /api/v1/artifacts/{artifactId}/analyze`. The host analysis
tool is `aapt` resolved from `PATH`.

**Successful behavior:** The backend executes bounded local analysis with the
exact argument vector `aapt dump badging STORED_APK_PATH`, reads stdout up to
1 MiB, parses the required `package: name=...` value and optional version name,
version code, SDK values, application label, launchable activity, and warning
lines when reported, and stores the latest analysis result for the artifact.

**Output, response, event, or visible state:** HTTP `200` JSON returns
`artifact` and `analysis`. `artifact.id` matches the selected artifact and
`artifact.analysisStatus` equals `ready`. `analysis` contains required
`tool: "aapt"`, required `packageName`, optional `versionName`, optional
`versionCode`, optional `minSdkVersion`, optional `targetSdkVersion`, optional
`applicationLabel`, optional `launchableActivity`, optional `warnings`, and
required `analyzedAt`. Version and SDK values are strings. `warnings`, when
present, contains at most 20 stdout warning lines, each trimmed to at most 500
bytes without splitting a UTF-8 sequence.

**Required side effects:** Execute only the local host tool command named
above and atomically replace `metadata.json` under the artifact directory with
the existing artifact metadata plus latest ready `analysis`.

**Forbidden side effects:** Analysis must not execute ADB, install or mutate an
APK, send network requests, write outside the artifact directory, expose stored
host paths, command stderr, environment values, or token values in responses, or
delete the artifact file.

**Errors and negative behavior:**

- Unknown artifact ID returns HTTP `404` with code `artifact_not_found`.
- Missing `aapt`, nonzero exit, timeout after 5 seconds, stdout larger than
  1 MiB, or output without a parseable `package: name=...` value returns HTTP
  `502` with code `artifact_analysis_failed`.
- Rejected Host or Origin requests return the standard security envelope before
  host-tool execution.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Security
checks run before artifact lookup and host-tool execution. Analysis times out
after 5 seconds. A failed analysis leaves the prior successful analysis, if any,
unchanged and records no new ready result. If no prior successful analysis
exists, failed analysis leaves `artifact.analysisStatus` equal to `pending` and
stores no `analysis`.

**Security, authorization, privacy, and data safety:** Analysis remains local.
Responses do not expose absolute stored paths or host environment details.

**Compatibility and versioning:** The route, status codes, JSON field names,
tool name, and error codes are stable for this specification version.

#### Acceptance Criteria

- `AC-019-001`: Given a stored artifact and `aapt` output with a parseable
  `package: name=...` value, when analysis is requested, then HTTP `200`
  returns the documented `artifact` and `analysis` fields, stores the latest
  ready analysis in `metadata.json`, and exposes no host paths, stderr,
  environment values, or tokens.
- `AC-019-002`: Given unknown artifact ID, missing `aapt`, nonzero exit,
  timeout after 5 seconds, stdout larger than 1 MiB, or output without a
  parseable `package: name=...` value, when analysis is requested, then the
  documented status and error code are returned and no false ready analysis is
  stored.
- `AC-019-003`: Given a rejected Host or Origin analysis request, when handled,
  then the standard security envelope is returned before any host tool is
  executed.

#### Observable Acceptance Boundary

- Primary boundary: HTTP request through the running server with isolated
  artifact storage and deterministic fake `aapt`.
- Focused evidence expected: process/HTTP and filesystem tests for successful
  analysis, persisted ready metadata, unknown artifact, missing `aapt`, nonzero
  exit, 5-second timeout, stdout larger than 1 MiB, output without
  `package: name=...`, and security rejection.
- Real-path exercise: start the server with one stored artifact and fake
  `aapt`, request analysis, inspect JSON, stored metadata, fake-tool logs,
  artifact detail, unknown artifact, missing tool, nonzero exit, timeout,
  oversized stdout, output without `package: name=...`, and rejected Host or
  Origin cases.
- Required environment: Linux host with loopback networking, writable
  temporary filesystem, and temporary executable fixtures.
- Permitted deterministic substitutes: fake `aapt` executable and disposable
  APK fixtures.
- Evidence that does not count: parser-only tests or static analysis metadata
  not produced through the production route.

#### Blocking Open Questions

- None.

### CAP-020: Artifact Catalog And Browser View

**Contract status:** Implementation-ready

**Goal:** A local browser or same-origin HTTP client can list stored APK
artifacts, inspect one artifact's metadata and latest analysis, and trigger
analysis from the browser.

**Actors or callers:** Browser loaded from the dashboard server and local HTTP
clients on the same origin.

**Inputs and preconditions:** The server is running. Artifact metadata may or
may not exist. Requests use `GET /api/v1/artifacts`,
`GET /api/v1/artifacts/{artifactId}`, and browser actions against those routes
plus `CAP-018` and `CAP-019`.

**Successful behavior:** The backend lists artifacts sorted by `createdAt`
descending, with `id` ascending as the tie-breaker when timestamps are equal,
and returns one artifact's stored metadata and latest analysis when present. The
browser renders upload, list, detail, pending, ready, and failure states using
backend responses.

**Output, response, event, or visible state:** List responses return
`artifacts.items` and `artifacts.count`. Detail responses return `artifact` and
optional `analysis`. Browser state shows artifact names, sizes, SHA-256,
analysis status, parsed metadata when ready, and failure states.

**Required side effects:** Browser upload and analyze actions use only the
production routes defined by `CAP-018` and `CAP-019`.

**Forbidden side effects:** Catalog and browser view must not execute ADB,
install artifacts, send network requests, expose stored host paths, or invent
analysis success when no ready analysis exists.

**Errors and negative behavior:**

- Unknown artifact ID returns HTTP `404` with code `artifact_not_found`.
- Corrupt metadata that cannot be read returns HTTP `500` with code
  `artifact_catalog_unavailable`.
- Rejected Host or Origin requests return the standard security envelope before
  artifact lookup or mutation.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Catalog
reads current artifact metadata from disk for each request. Browser refresh
replaces stale catalog state with the latest response.

**Security, authorization, privacy, and data safety:** Browser and API
responses omit absolute stored paths, host environment values, command stderr,
and tokens.

**Compatibility and versioning:** The routes, status codes, JSON field names,
browser state labels, and error codes are stable for this specification
version.

#### Acceptance Criteria

- `AC-020-001`: Given zero or more stored artifacts, when the artifact catalog
  is requested, then HTTP `200` returns artifact items sorted by `createdAt`
  descending with `id` ascending for equal timestamps, count, and no host paths
  or token values.
- `AC-020-002`: Given a known artifact with or without ready analysis, when
  detail is requested, then HTTP `200` returns artifact metadata and the latest
  analysis only when present.
- `AC-020-003`: Given unknown artifact ID, corrupt metadata, or rejected Host or
  Origin, when catalog or detail is requested, then the documented status and
  error code are returned before forbidden side effects.
- `AC-020-004`: Given the browser uploads an artifact or refreshes the catalog,
  then it renders upload, catalog, detail, pending/no-analysis, and failure
  states from backend responses and exposes no install, device mutation, shell,
  file-transfer, or external-service controls.
- `AC-020-005`: Given the browser triggers analysis for a stored artifact, then
  it renders pending, ready parsed metadata, and failure states from backend
  responses and exposes no install, device mutation, shell, file-transfer, or
  external-service controls.

#### Observable Acceptance Boundary

- Primary boundary: HTTP request and browser interaction through the running
  server with isolated artifact storage.
- Focused evidence expected: process/HTTP and browser-boundary tests for empty
  and populated catalog, detail with and without analysis, corrupt metadata,
  security rejection, browser upload/analyze success, and browser failure
  states.
- Real-path exercise: start the server with isolated artifact storage, upload
  and analyze a disposable artifact, refresh the browser catalog, inspect
  detail state, restart the server, and repeat catalog/detail checks.
- Required environment: Linux host with loopback networking, writable temporary
  filesystem, browser automation, and temporary executable fixtures.
- Permitted deterministic substitutes: fake `aapt` executable and disposable
  APK fixtures.
- Evidence that does not count: static markup checks or catalog tests not
  backed by persisted artifact metadata.

#### Blocking Open Questions

- None.

### CAP-021: Explicit Artifact Deletion

**Contract status:** Implementation-ready

**Goal:** A local browser or same-origin HTTP client can explicitly delete one
stored APK artifact and its analysis metadata without touching connected
devices or unrelated files.

**Actors or callers:** Browser loaded from the dashboard server and local HTTP
clients on the same origin.

**Inputs and preconditions:** The server is running. Requests use
`DELETE /api/v1/artifacts/{artifactId}` for an opaque artifact ID previously
returned by `CAP-018`.

**Successful behavior:** The backend removes only the selected artifact
directory and returns a deletion confirmation. Repeating deletion for the same
ID returns not found.

**Output, response, event, or visible state:** HTTP `200` JSON with
`artifact.id` and `deleted` equal to `true`. Browser catalog no longer shows the
artifact after refresh.

**Required side effects:** Delete the selected artifact file and metadata under
the resolved data directory.

**Forbidden side effects:** Deletion must not follow symlinks out of the
artifact root, delete files outside the selected artifact directory, execute ADB
or host analysis tools, mutate devices, or report success if the selected
artifact remains.

**Errors and negative behavior:**

- Unknown artifact ID returns HTTP `404` with code `artifact_not_found`.
- Invalid artifact ID syntax returns HTTP `400` with code
  `invalid_artifact_request`.
- Filesystem deletion failure returns HTTP `500` with code
  `artifact_delete_failed`.
- Rejected Host or Origin requests return the standard security envelope before
  filesystem deletion.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Security and
ID validation occur before filesystem deletion. Deletion is scoped to one
artifact directory beneath the artifact root.

**Security, authorization, privacy, and data safety:** Deletion is explicit,
local, scoped, and does not expose host paths or sensitive values.

**Compatibility and versioning:** The route, status codes, JSON field names,
and error codes are stable for this specification version.

#### Acceptance Criteria

- `AC-021-001`: Given an existing artifact, when deletion is requested, then
  HTTP `200` confirms deletion, the artifact disappears from catalog responses,
  and only the selected artifact directory is removed.
- `AC-021-002`: Given invalid artifact ID, unknown artifact ID, filesystem
  deletion failure, or rejected Host or Origin, when deletion is requested, then
  the documented status and error code are returned and no unrelated files are
  removed.
- `AC-021-003`: Given the browser deletes an artifact, then the browser refresh
  shows the artifact absent and exposes no device mutation, install, shell,
  file-transfer, or external-service controls.

#### Observable Acceptance Boundary

- Primary boundary: HTTP request and browser interaction through the running
  server with isolated artifact storage.
- Focused evidence expected: process/HTTP, filesystem, and browser-boundary
  tests for successful deletion, repeated deletion, invalid ID, deletion
  failure, symlink/path escape protection, security rejection, and browser
  refresh.
- Real-path exercise: start the server with isolated artifact storage, upload
  an artifact, delete it through the API and browser, inspect catalog response
  and filesystem side effects, and repeat negative cases.
- Required environment: Linux host with loopback networking, writable temporary
  filesystem, browser automation, and disposable fixtures.
- Permitted deterministic substitutes: generated APK fixtures.
- Evidence that does not count: route-name tests or deletion tests that do not
  inspect filesystem side effects.

#### Blocking Open Questions

- None.

### INV-NIY-001: Standard NIY Behavior

**Contract status:** Current invariant

**Goal:** Recognized but unavailable commands, actions, or status fields fail
closed or report unavailable without pretending out-of-scope behavior
succeeded.

**Actors or callers:** CLI users, browser UI, HTTP clients, and future selected
routes while a function remains unavailable.

**Inputs and preconditions:** A documented command, option, route, field, or UI
state recognizes a function name but that function's production behavior is not
present in the current build.

**Successful behavior:** Informational status surfaces may report documented
`NIY` values. Requested unavailable actions fail closed with the documented
`NIY` diagnostic and status `6`.

**Output, response, event, or visible state:** CLI unavailable behavior writes
`NIY: FUNCTION is not implemented yet` to stderr, writes no requested-action
stdout, and exits with status `6`. API and frontend unavailable states use
documented `NIY` values only where this specification names them.

**Required side effects:** None for requested unavailable actions.

**Forbidden side effects:** `NIY` action handling must not start external
processes, mutate device state, write project or artifact state, open
interactive sessions, or report requested-action success.

**Errors and negative behavior:** Security checks that apply to a command or
route still run before an `NIY` response.

**Ordering, lifecycle, timeout, cancellation, retry, and cleanup:** Fail-closed
behavior happens before process launch, state mutation, or session creation.

**Security, authorization, privacy, and data safety:** `NIY` responses must not
bypass authorization, host/origin checks, CSRF checks, or read-only enforcement
where those checks apply.

**Compatibility and versioning:** Stable function names in `NIY` output remain
stable until the function receives an accepted behavior contract.

#### Invariant Criteria

- `INV-NIY-001-A`: Given a recognized unavailable requested action named by a
  current or future capability, when invoked, then it writes the documented
  `NIY` diagnostic to stderr, writes no requested-action stdout, exits with
  status `6`, and performs no action side effects.
- `INV-NIY-001-B`: Given an informational current status surface with
  unavailable fields, when requested, then those fields report documented `NIY`
  values without changing the overall success semantics of the current status
  surface.
- `INV-NIY-001-C`: Given a security check applies to a command or route, when
  the requested behavior is unavailable, then the security check runs before the
  `NIY` response.

#### Observable Evidence Boundary

- Primary boundary: CLI process invocation, HTTP route, or browser state named
  by the selected roadmap slice.
- Focused evidence expected: automated boundary tests asserting diagnostics,
  status, output, and absence of forbidden side effects.
- Real-path exercise: invoke a recognized unavailable action or request a
  status surface with documented unavailable fields.
- Required environment: Same as the selected current boundary.
- Permitted deterministic substitutes: fake external executables only to prove
  they are not invoked.
- Evidence that does not count: checking that a function name constant exists.

#### Blocking Open Questions

- None.

## Configuration And Environment

Documented environment variables:

| ID | Name | Source and precedence | Default | Validation | Failure behavior |
|---|---|---|---|---|---|
| `CFG-001` | `ADB_DASHBOARD_LISTEN` | Environment, below command-line options and above configuration files | none | same as `server.listen` | `ERR-004-004` through `ERR-004-006` |
| `CFG-002` | `ADB_DASHBOARD_CONFIG` | Explicit configuration file, below `--config` and above default files | none | file must load and parse | `ERR-004-001` or `ERR-004-002` |
| `CFG-003` | `ADB_DASHBOARD_DATA_DIR` | Environment, below command-line options and above configuration files | none | same as `server.data_dir` | `ERR-004-009` or `ERR-004-010`; filesystem failures use `CAP-005` |
| `CFG-004` | `ADB_DASHBOARD_TEMP_DIR` | Environment, below command-line options and above configuration files | none | same as `server.temp_dir` | `ERR-004-009` or `ERR-004-010`; filesystem failures use `CAP-005` |
| `CFG-005` | `ADB_DASHBOARD_LOG_LEVEL` | Environment, below command-line options and above configuration files | none | one of `error`, `warn`, `info`, `debug`, `trace` | `ERR-004-007` |
| `CFG-006` | `HOME` | Default path fallback and `~/` expansion | host environment | non-empty when used for fallback or expansion | `ERR-004-010` |
| `CFG-007` | `XDG_CONFIG_HOME` | default user config path | `$HOME/.config` fallback | path expansion rules | missing default file ignored |
| `CFG-008` | `XDG_STATE_HOME` | default data directory | `$HOME/.local/state` fallback | path expansion rules | `ERR-004-010` if required fallback variable is unavailable |
| `CFG-009` | `XDG_RUNTIME_DIR` | default temp directory | `$TMPDIR` or `/tmp` fallback | path expansion rules | filesystem failures use `CAP-005` |
| `CFG-010` | `TMPDIR` | temp directory fallback | `/tmp` fallback | path expansion rules | filesystem failures use `CAP-005` |

## Files, Data, And Persistence

| ID | Path or data | Read/write | Ownership | Atomicity | Retention or cleanup | Failure behavior |
|---|---|---|---|---|---|---|
| `DATA-001` | `/etc/adb-dashboard/config.toml` | read | host system config | no writes | missing file ignored | parse failure is `ERR-004-001` |
| `DATA-002` | `$XDG_CONFIG_HOME/adb-dashboard/config.toml` or `$HOME/.config/adb-dashboard/config.toml` | read | process user config | no writes | missing file ignored | parse failure is `ERR-004-001` |
| `DATA-003` | explicit config file from `ADB_DASHBOARD_CONFIG` or `--config` | read | selected by operator | no writes | missing file is fatal | `ERR-004-001` or `ERR-004-002` |
| `DATA-004` | resolved `server.data_dir` | create or validate directory | process user | directory may remain if later startup validation fails | no cleanup in current contract | `CAP-005` failure |
| `DATA-005` | resolved `server.temp_dir` | create or validate directory | process user | directory may remain if startup succeeds or fails after creation | no cleanup in current contract | `CAP-005` failure |
| `DATA-006` | `server.data_dir/artifacts/ARTIFACT_ID/original.apk` | write on upload, read for analysis, delete on explicit deletion | process user | written through temporary file then atomic move | retained until explicit artifact deletion | `CAP-018`, `CAP-019`, or `CAP-021` failure |
| `DATA-007` | `server.data_dir/artifacts/ARTIFACT_ID/metadata.json` | write on upload and analysis, read for catalog/detail, delete on explicit deletion | process user | write replacement must not expose partial metadata | retained until explicit artifact deletion | `CAP-018`, `CAP-019`, `CAP-020`, or `CAP-021` failure |

## Public Interfaces And Output

### Help Output

```text
Usage:
  adb-dashboard [OPTIONS]
  adb-dashboard serve [OPTIONS]
  adb-dashboard version
  adb-dashboard doctor

Commands:
  serve      Start the local dashboard server
  version    Print application build information
  doctor     Run current diagnostics

Options:
  --listen ADDRESS
  --data-dir PATH
  --config PATH
  --temp-dir PATH
  --log-level LEVEL
  --open
  --no-open
  --read-only
  --version
  --help
```

### Version Output

```text
adb-dashboard VERSION
commit: COMMIT
buildDate: DATE
goVersion: GOVERSION
frontendRevision: REVISION
```

Values may vary by build. Labels and order are stable.

### HTTP API

The current API root is `/api/v1`. Current responses use JSON with content type
`application/json`. Only these routes are current public behavior:

- `GET /api/v1/bootstrap`
- `GET /api/v1/status`
- `GET /api/v1/devices`
- `GET /api/v1/devices/{serial}`
- `GET /api/v1/devices/{serial}/logcat`
- `GET /api/v1/devices/{serial}/screenshot`
- `GET /api/v1/devices/{serial}/packages`
- `GET /api/v1/devices/{serial}/packages/{packageName}`
- `POST /api/v1/artifacts`
- `GET /api/v1/artifacts`
- `GET /api/v1/artifacts/{artifactId}`
- `POST /api/v1/artifacts/{artifactId}/analyze`
- `DELETE /api/v1/artifacts/{artifactId}`
- unknown `/api/v1` route error handling

Only `POST /api/v1/artifacts` requires a request body.

### Exit Status

| Status | Meaning |
|---|---|
| `0` | Help, version, successful doctor report, or clean server shutdown. |
| `1` | General startup or runtime failure not otherwise classified. |
| `2` | Invalid command-line usage or configuration. |
| `3` | ADB executable or ADB version discovery failure when no higher-priority failure is present. |
| `4` | Listen address unavailable. |
| `5` | Required startup filesystem directory unavailable. |
| `6` | Recognized command or function is `NIY`; no requested action was performed. |

## Security And Data-Safety Invariants

- `INV-SEC-002`: Current server startup rejects non-loopback listen hosts before
  binding.
- `INV-SEC-003`: Current `/api/v1` requests with rejected host or origin data
  return only the standard error envelope and do not reach route handlers.
- `INV-SEC-004`: Bootstrap token values must not appear in logs, status JSON,
  doctor output, CLI diagnostics, or visible UI.
- `INV-DATA-002`: Configuration failure must happen before startup directory
  creation.
- `INV-DATA-003`: Current behavior may execute only the ADB commands named by
  `CAP-010`, `CAP-011`, `CAP-014`, `CAP-015`, `CAP-016`, and `CAP-017`. It
  must not execute optional host tools or other external device-operation
  commands.
- `INV-DATA-004`: Artifact behavior may write only the artifact files and
  metadata named by `DATA-006` and `DATA-007` under the resolved data
  directory, and deletion may remove only one selected artifact directory.

## Compatibility And Migration

- The specification targets Linux hosts.
- The HTTP server is local-only and uses plain HTTP on loopback.
- Current command names, option names, configuration keys, environment variable
  names, exit-status meanings, doctor row names, API route paths, JSON field
  names, and `NIY` function names are stable for specification version `1.3.0`.
- Additional fields, routes, commands, UI controls, files, and configuration
  keys require a later accepted capability.
- No persisted data schema or migration behavior exists in the current
  contract.

## Known Limitations

- The dashboard executes only `adb version`, `adb devices -l`, bounded
  `adb -s SERIAL logcat -d`, bounded
  `adb -s SERIAL exec-out screencap -p`, bounded package inventory/detail ADB
  commands, and bounded local `aapt dump badging` artifact analysis under this
  specification.
- The dashboard does not implement the ADB wire protocol.
- Device mutation, interactive ADB workflows, retained device output, and
  host-tool workflows other than local APK analysis are unavailable until later
  specifications define exact behavior.
- Fastboot is outside the current contract.

## Open Questions

- None for `CAP-001` through `CAP-021` and `INV-NIY-001`.

Future mutating ADB, interactive device control, WebSocket streaming, file
transfer, install/uninstall workflows, artifact reporting beyond local APK
metadata, logging redaction, request-correlation, performance, packaging,
release, deployment, and migration behavior requires separate specification
before it can be implemented.

## Specification Acceptance Record

- Audit result: SPECIFICATION ACCEPTED
- Reviewed capabilities: `CAP-001` through `CAP-021`; `INV-NIY-001`
- Blocking gaps: None for the local bootstrap, read-only ADB discovery, M3
  read-only device inspection, M4 read-only package inspection, and M5 local
  APK artifact intake and analysis contract
- Evidence or review reference: Authored against
  `docs/SPECIFICATION_GUIDE.md`, `docs/SPECIFICATION.template.md`,
  `docs/READINESS_CHECKLIST.md`, `AGENTS.md`, and the source material available
  during the specification rewrite.
