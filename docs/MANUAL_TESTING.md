# Manual Testing

## Scope

These steps exercise the current verified implementation level through
`M3-S3`: local CLI behavior, configuration and startup checks, loopback server
lifecycle, browser bootstrap/status, ADB discovery, device inventory, device
detail, bounded read-only logcat, and read-only PNG screenshot capture.

## Requirements

- Linux host with loopback networking.
- Go toolchain available on `PATH`.
- `adb` available on `PATH`.
- One connected Android device whose `adb devices` state is `device`.
- Browser available for UI checks.

Run all commands from the repository root:

```sh
cd /home/opsman/project_git/adb-dashboard
```

## Build

Build the command under a disposable repository-local path:

```sh
mkdir -p .codex/tmp
go build -o .codex/tmp/adb-dashboard ./cmd/adb-dashboard
```

Expected result:

- Command exits `0`.
- `.codex/tmp/adb-dashboard` exists and is executable.

## CLI Discovery

Check help output:

```sh
.codex/tmp/adb-dashboard --help
```

Expected result:

- Exit status is `0`.
- Stdout starts with `Usage:`.
- Stdout lists `serve`, `version`, `doctor`, `--listen`, `--no-open`,
  `--read-only`, `--version`, and `--help`.
- No server starts.

Check version output:

```sh
.codex/tmp/adb-dashboard version
.codex/tmp/adb-dashboard --version
```

Expected result for each command:

- Exit status is `0`.
- Stdout includes these labels in order: `adb-dashboard`, `commit:`,
  `buildDate:`, `goVersion:`, and `frontendRevision:`.
- No server starts.

Check invalid command handling:

```sh
.codex/tmp/adb-dashboard devices
```

Expected result:

- Exit status is `2`.
- Stderr is `unknown command: devices`.
- No ADB device command runs as a result of this invalid dashboard command.

Check missing option argument:

```sh
.codex/tmp/adb-dashboard --listen
```

Expected result:

- Exit status is `2`.
- Stderr is `missing argument for --listen`.
- No server starts.

## Live Device Setup

Confirm the host sees a ready device:

```sh
adb version
adb devices -l
SERIAL="$(adb devices | awk 'NR > 1 && $2 == "device" { print $1; exit }')"
echo "$SERIAL"
```

Expected result:

- `adb version` exits `0`.
- `adb devices -l` shows at least one row with state `device`.
- `SERIAL` is non-empty.

Stop if `SERIAL` is empty. The current logcat and detail checks require a ready
device.

## Doctor

Use explicit disposable data and temp directories:

```sh
DATA_DIR="$PWD/.codex/tmp/manual-data"
TEMP_DIR="$PWD/.codex/tmp/manual-temp"

.codex/tmp/adb-dashboard doctor \
  --data-dir "$DATA_DIR" \
  --temp-dir "$TEMP_DIR" \
  --listen 127.0.0.1:18080 \
  --read-only
```

Expected result:

- Exit status is `0`.
- Stdout starts with `adb-dashboard doctor`.
- `config: PASS` shows `listen=127.0.0.1:18080` and `readOnly=true`.
- `dataDir: PASS` references `$DATA_DIR`.
- `tempDir: PASS` references `$TEMP_DIR`.
- `adbExecutable: PASS` references the resolved `adb` executable.
- `adbVersion: PASS` shows the first non-empty `adb version` line.
- Unimplemented rows remain marked `NIY`; they are not failures.
- Stderr is empty.

Check ADB-unavailable doctor behavior without modifying the host:

```sh
PATH=/nonexistent .codex/tmp/adb-dashboard doctor \
  --data-dir "$DATA_DIR" \
  --temp-dir "$TEMP_DIR" \
  --listen 127.0.0.1:18080
```

Expected result:

- Exit status is `3`.
- `adbExecutable: FAIL error=not found in PATH` appears.
- `adbVersion: NIY adb.version unavailable until adb executable is found`
  appears.

## Server Startup And Status API

Start the server in one terminal:

```sh
.codex/tmp/adb-dashboard serve \
  --listen 127.0.0.1:18080 \
  --data-dir "$DATA_DIR" \
  --temp-dir "$TEMP_DIR" \
  --read-only \
  --no-open
```

Expected result:

- Stderr contains `INFO server started addr=127.0.0.1:18080`.
- The process keeps running until interrupted.
- No browser is opened because `--no-open` was supplied.

In another terminal from the repository root, define the address:

```sh
ADDR=127.0.0.1:18080
```

Request status:

```sh
curl -i -sS "http://$ADDR/api/v1/status"
```

Expected result:

- HTTP status is `200`.
- `Content-Type` is `application/json`.
- JSON includes:
  - `application.name` equal to `adb-dashboard`;
  - `server.status` equal to `running`;
  - `server.readOnly` equal to `true`;
  - `server.bind` equal to `127.0.0.1:18080`;
  - `adb.status` equal to `available`;
  - `adb.executable` and `adb.version` as non-empty values;
  - unavailable future areas such as watcher, jobs, sessions, storage, and host
    tools marked `NIY`.
- JSON does not include bootstrap token values.

Request bootstrap:

```sh
curl -i -sS "http://$ADDR/api/v1/bootstrap"
```

Expected result:

- HTTP status is `200`.
- JSON includes `csrfToken`, `webSocketToken`, and `statusUrl`.
- `statusUrl` is `/api/v1/status`.
- Token values are non-empty.

Request an unknown API route:

```sh
curl -i -sS "http://$ADDR/api/v1/unknown"
```

Expected result:

- HTTP status is `404`.
- JSON error code is `not_found`.
- JSON error message is `Unknown API route`.

Check API security rejection before ADB work:

```sh
curl -i -sS -H "Host: foreign.example" "http://$ADDR/api/v1/status"
curl -i -sS -H "Origin: http://foreign.example" "http://$ADDR/api/v1/status"
```

Expected result:

- Each response has HTTP status `403`.
- Error code is `forbidden_host` for the foreign Host request.
- Error code is `forbidden_origin` for the foreign Origin request.

## Device Inventory API

Request the live inventory:

```sh
curl -i -sS -H "Origin: http://$ADDR" "http://$ADDR/api/v1/devices"
```

Expected result:

- HTTP status is `200`.
- JSON includes `adb.status` equal to `available`.
- JSON includes a `devices` array.
- One `devices` item has `serial` equal to `$SERIAL` and `state` equal to
  `device`.
- Device metadata fields such as `product`, `model`, `device`, and
  `transportId` may be present when ADB reports them.
- JSON does not include bootstrap tokens, host environment variables, or
  command stderr.

Check inventory security rejection:

```sh
curl -i -sS -H "Origin: http://foreign.example" "http://$ADDR/api/v1/devices"
```

Expected result:

- HTTP status is `403`.
- Error code is `forbidden_origin`.

## Device Detail API

Request detail for the connected ready device:

```sh
curl -i -sS -H "Origin: http://$ADDR" "http://$ADDR/api/v1/devices/$SERIAL"
```

Expected result:

- HTTP status is `200`.
- JSON includes `adb.status` equal to `available`.
- JSON includes `device.serial` equal to `$SERIAL`.
- JSON includes `device.state` equal to `device`.
- Optional device metadata matches the current `adb devices -l` row when
  reported by ADB.
- JSON does not include bootstrap tokens, host environment variables, or
  command stderr.

Request detail for an absent serial:

```sh
curl -i -sS -H "Origin: http://$ADDR" "http://$ADDR/api/v1/devices/not-a-real-serial"
```

Expected result:

- HTTP status is `404`.
- Error code is `device_not_found`.
- Response has no route-specific `device` field.

Check detail security rejection:

```sh
curl -i -sS -H "Origin: http://foreign.example" "http://$ADDR/api/v1/devices/$SERIAL"
```

Expected result:

- HTTP status is `403`.
- Error code is `forbidden_origin`.

## Device Logcat API

Request a bounded logcat dump:

```sh
curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/devices/$SERIAL/logcat?lines=5&format=plain"
```

Expected result:

- HTTP status is `200`.
- JSON includes `device.serial` equal to `$SERIAL`.
- JSON includes `device.state` equal to `device`.
- JSON includes `logcat.format` equal to `plain`.
- JSON includes `logcat.lines` as an array with at most 5 strings.
- `logcat.truncated` is `true` when more than 5 log lines were available and
  `false` otherwise.
- JSON does not include bootstrap tokens, host environment variables, host file
  paths, or command stderr.

Request empty-or-default bounded logcat behavior:

```sh
curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/devices/$SERIAL/logcat"
```

Expected result:

- HTTP status is `200`.
- `logcat.format` is `plain`.
- `logcat.lines` contains at most the default 200 lines.

Check invalid logcat query handling:

```sh
curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/devices/$SERIAL/logcat?lines=0"

curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/devices/$SERIAL/logcat?lines=501"

curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/devices/$SERIAL/logcat?format=json"
```

Expected result for each request:

- HTTP status is `400`.
- Error code is `invalid_logcat_request`.
- Response has no route-specific `logcat` field.

Request logcat for an absent serial:

```sh
curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/devices/not-a-real-serial/logcat?lines=5&format=plain"
```

Expected result:

- HTTP status is `404`.
- Error code is `device_not_found`.
- Response has no route-specific `logcat` field.

Check logcat security rejection:

```sh
curl -i -sS -H "Origin: http://foreign.example" \
  "http://$ADDR/api/v1/devices/$SERIAL/logcat?lines=5&format=plain"
```

Expected result:

- HTTP status is `403`.
- Error code is `forbidden_origin`.

## Device Screenshot API

Request a PNG screenshot for the connected ready device:

```sh
curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/devices/$SERIAL/screenshot" \
  --output .codex/tmp/manual-screenshot-response.bin
```

Expected result:

- HTTP status is `200`.
- Response headers include `Content-Type: image/png`.
- `.codex/tmp/manual-screenshot-response.bin` starts with PNG bytes.
- The response body is image bytes, not a JSON envelope.
- The response does not include bootstrap tokens, host environment variables,
  host file paths, or command stderr.
- No screenshot file or artifact state is retained in `$DATA_DIR`, `$TEMP_DIR`,
  or `.codex/tmp` except the explicit curl output file above.

Inspect the PNG signature:

```sh
xxd -l 8 .codex/tmp/manual-screenshot-response.bin
```

Expected result:

- Output starts with `00000000: 8950 4e47 0d0a 1a0a`.

Request screenshot for an absent serial:

```sh
curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/devices/not-a-real-serial/screenshot"
```

Expected result:

- HTTP status is `404`.
- Error code is `device_not_found`.
- Response has no route-specific `screenshot` field.

Check screenshot security rejection:

```sh
curl -i -sS -H "Origin: http://foreign.example" \
  "http://$ADDR/api/v1/devices/$SERIAL/screenshot"
```

Expected result:

- HTTP status is `403`.
- Error code is `forbidden_origin`.

## Browser UI

Open the dashboard in a browser:

```text
http://127.0.0.1:18080/
```

Expected result:

- The page title and heading identify `adb-dashboard`.
- Server status is `server: running`.
- Bind shows `127.0.0.1:18080`.
- Read-only shows `true`.
- ADB status shows `available`.
- Device count is shown.
- The connected serial appears in the device list.
- Future areas such as watcher, jobs, sessions, storage, and host tools show
  `NIY`.
- No bootstrap token values are visible.
- No unsupported controls or text for shell, install, uninstall, transfer,
  reboot, jobs, sessions, artifacts, screen recording, or settings appear.

Click `refresh`.

Expected result:

- Device count and list refresh from current ADB inventory.
- Stale detail, logcat, and screenshot visible state are cleared to unavailable
  until opened again.

Click `details`.

Expected result:

- Detail area shows the selected serial and state.
- Optional product, model, device, and transport metadata appear only when
  reported by ADB.
- No command stderr, executable path, or host environment value is visible.

Click `logcat`.

Expected result:

- Logcat area first shows a loading state.
- It then shows `logcat: SERIAL` followed by bounded log lines, or
  `logcat: empty` when no lines are returned.
- If ADB becomes unavailable or logcat fails, the logcat area shows
  `logcat: unavailable`.
- No clear-log, stream, shell, file-transfer, install, reboot, or mutation
  control appears.
- No logcat output is retained in repository data or temp directories.

Click `screenshot`.

Expected result:

- Screenshot area first shows a loading state.
- It then shows `screenshot: SERIAL`.
- The latest image is visible in the browser as a PNG screenshot.
- If ADB becomes unavailable, the device is no longer ready, or screenshot
  capture fails, the screenshot area shows `screenshot: unavailable`.
- No screen-recording, annotation, shell, file-transfer, install, reboot, or
  mutation control appears.
- No screenshot output is retained in repository data or temp directories.

## Non-Ready Device Observation

If a connected device is visible but not ready, for example with state
`offline` or `unauthorized`, do not run device-mutating commands to force a
state change as part of this manual test.

With that non-ready serial substituted for `$SERIAL`, logcat should return:

- HTTP status `409`.
- Error code `device_not_ready`.
- No route-specific `logcat` field.

With that non-ready serial substituted for `$SERIAL`, screenshot should return:

- HTTP status `409`.
- Error code `device_not_ready`.
- No route-specific `screenshot` field.

## Shutdown

Stop the server terminal with `Ctrl-C`.

Expected result:

- Stderr contains `INFO server stopped signal=interrupt`.
- Process exits with status `0`.

Confirm no server is listening:

```sh
curl -i -sS "http://$ADDR/api/v1/status"
```

Expected result:

- Curl cannot connect to `127.0.0.1:18080`.

## Cleanup

Remove the disposable manual-test files when they are no longer needed:

```sh
rm -rf .codex/tmp/manual-data .codex/tmp/manual-temp \
  .codex/tmp/manual-screenshot-response.bin .codex/tmp/adb-dashboard
```

Do not remove `.codex/cache` directories; they may be used by repository test
commands.
