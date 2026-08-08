# Manual Testing

## Scope

These steps exercise the current implementation level through
`M6-S2`: local CLI behavior, configuration and startup checks, loopback server
lifecycle, browser bootstrap/status, ADB discovery, device inventory, device
detail, bounded read-only logcat, read-only PNG screenshot capture, and
read-only package inventory API/browser behavior plus package detail API
behavior and local APK artifact upload, catalog, detail, analysis API, and
browser analysis and deletion behavior, plus local artifact JSON and Markdown
report API behavior.

## Requirements

- Linux host with loopback networking.
- Go toolchain available on `PATH`.
- `adb` available on `PATH`.
- `aapt` available on `PATH` for artifact analysis checks.
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

## Package Inventory API

Request package inventory for the connected ready device:

```sh
curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/devices/$SERIAL/packages"
```

Expected result:

- HTTP status is `200`.
- JSON includes `device.serial` equal to `$SERIAL`.
- JSON includes `device.state` equal to `device`.
- JSON includes `packages.scope` equal to `all`.
- JSON includes `packages.items` as an array sorted by package name.
- JSON includes `packages.count` equal to the number of returned items.
- Package items include `name` and may include `apkPath`, `userId`,
  `versionCode`, or `installer` when reported by ADB.
- JSON does not include bootstrap tokens, host environment variables, host file
  paths, or command stderr.
- No package output or artifact state is retained in `$DATA_DIR`, `$TEMP_DIR`,
  or `.codex/tmp`.

Repeat the package inventory request for the accepted scopes:

```sh
curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/devices/$SERIAL/packages?scope=all"

curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/devices/$SERIAL/packages?scope=third-party"

curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/devices/$SERIAL/packages?scope=system"
```

Expected result for each request:

- HTTP status is `200`.
- `packages.scope` matches the requested scope.
- `packages.items` is sorted by package name.
- The request is read-only package inspection; it must not install, uninstall,
  clear, stop, launch, or pull packages.

Request an invalid package inventory scope:

```sh
curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/devices/$SERIAL/packages?scope=disabled"
```

Expected result:

- HTTP status is `400`.
- Error code is `invalid_package_request`.
- No package inventory command is executed for the invalid request.

Repeat package inventory requests with an absent serial, a non-ready device,
ADB unavailable, package command failure, timeout, malformed output, invalid
UTF-8 output, and oversized output.

Expected result:

- Absent serial returns HTTP `404` with code `device_not_found`.
- Non-ready device returns HTTP `409` with code `device_not_ready`.
- ADB unavailable returns HTTP `503` with code `adb_unavailable`.
- Package command failure, timeout, malformed output, invalid UTF-8 output, or
  oversized output returns HTTP `502` with code `adb_packages_failed`.
- Error JSON does not include `device`, `packages`, command stderr, host
  environment variables, host filesystem paths, or token values.
- No package output or artifact state is retained in `$DATA_DIR`, `$TEMP_DIR`,
  or `.codex/tmp`.

Request package inventory with a rejected Host or Origin:

```sh
curl -i -sS -H "Host: foreign.example" \
  "http://$ADDR/api/v1/devices/$SERIAL/packages"

curl -i -sS -H "Origin: http://foreign.example" \
  "http://$ADDR/api/v1/devices/$SERIAL/packages"
```

Expected result:

- HTTP status is `403`.
- Error code is `forbidden_host` or `forbidden_origin`.
- The rejection occurs before any ADB process is executed.

## Package Detail API

Request package detail for one package on the connected ready device:

```sh
curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/devices/$SERIAL/packages/com.example.alpha"
```

Expected result:

- HTTP status is `200`.
- JSON includes `device.serial` equal to `$SERIAL`.
- JSON includes `device.state` equal to `device`.
- JSON includes `package.name` equal to the requested package name.
- `package` may include `versionName`, `versionCode`, `installer`,
  `firstInstallTime`, `lastUpdateTime`, and `requestedPermissions` when
  reported by ADB.
- `package.summaryLines` is present and bounded.
- JSON does not include bootstrap tokens, host environment variables, host file
  paths, or command stderr.
- No package output or artifact state is retained in `$DATA_DIR`, `$TEMP_DIR`,
  or `.codex/tmp`.

Repeat package detail requests with an invalid package name, absent serial,
non-ready device, ADB unavailable, package-not-found output, command failure,
timeout, malformed output, invalid UTF-8 output, and oversized output.

Expected result:

- Invalid package name returns HTTP `400` with code
  `invalid_package_request`.
- Absent serial returns HTTP `404` with code `device_not_found`.
- Non-ready device returns HTTP `409` with code `device_not_ready`.
- ADB unavailable returns HTTP `503` with code `adb_unavailable`.
- Package-not-found output returns HTTP `404` with code `package_not_found`.
- Package command failure, timeout, malformed output, invalid UTF-8 output, or
  oversized output returns HTTP `502` with code `adb_package_detail_failed`.
- Error JSON does not include `device`, `package`, `packages`, command stderr,
  host environment variables, host filesystem paths, or token values.
- No package output or artifact state is retained in `$DATA_DIR`, `$TEMP_DIR`,
  or `.codex/tmp`.

Request package detail with a rejected Host or Origin:

```sh
curl -i -sS -H "Host: foreign.example" \
  "http://$ADDR/api/v1/devices/$SERIAL/packages/com.example.alpha"

curl -i -sS -H "Origin: http://foreign.example" \
  "http://$ADDR/api/v1/devices/$SERIAL/packages/com.example.alpha"
```

Expected result:

- HTTP status is `403`.
- Error code is `forbidden_host` or `forbidden_origin`.
- The rejection occurs before any ADB process is executed.

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
- Artifact upload, refresh, detail, analyze, and delete controls are visible.
- No unsupported controls for shell, install, uninstall, transfer, reboot,
  jobs, sessions, screen recording, or settings appear.

Click `refresh`.

Expected result:

- Device count and list refresh from current ADB inventory.
- Stale detail, logcat, screenshot, and package inventory visible state are
  cleared to unavailable until opened again.

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

Click `packages`.

Expected result:

- Package inventory area first shows a loading state.
- It then shows `packages: all count=N`.
- Package rows are sorted by package name and may include version code, user
  identifier, and APK path when reported by ADB.
- If no package rows are returned, the package list shows `empty`.
- If ADB becomes unavailable, the device is no longer ready, or package
  inventory fails, the package inventory area shows `packages: unavailable`.
- No install, uninstall, clear, stop, launch, pull, shell, file-transfer, or
  mutation control appears.
- No package output is retained in repository data or temp directories.
- Package rows can be opened for read-only package detail.

Click `all`, `third-party`, and `system`.

Expected result:

- The package inventory area updates to the selected scope.
- The count and rows are derived from the backend package inventory response for
  that scope.
- Stale package rows and package detail are cleared during loading and after
  failure.

Click a package row.

Expected result:

- Package detail area first shows a loading state.
- It then shows `package detail: SERIAL PACKAGE_NAME`.
- Parsed fields such as version name, version code, installer, install times,
  update times, and requested permissions appear only when reported by ADB.
- Bounded summary lines from the backend package detail response are visible.
- If ADB becomes unavailable, the device is no longer ready, the package is not
  found, or package detail fails, the package detail area shows
  `package detail: unavailable`.
- No install, uninstall, clear, stop, launch, pull, shell, file-transfer, or
  mutation control appears.
- No package detail output is retained in repository data or temp directories.

Upload a disposable APK-like ZIP with the artifact upload control.

Expected result:

- Upload status shows `upload: FILENAME`.
- Artifact status shows `artifacts: N`.
- The catalog row shows the uploaded file name, size, SHA-256, and
  `analysis: pending`.
- Clicking artifact `details` shows the uploaded file name, size, SHA-256, and
  pending analysis state derived from the backend catalog response.
- Refreshing the browser or restarting the server with the same `$DATA_DIR`
  preserves the catalog and detail state.
- Invalid uploads show `upload: unavailable` and do not render a false
  artifact row or detail.
- If the catalog cannot be read, artifact status shows
  `artifacts: unavailable` and stale catalog/detail text is cleared.
- Clicking artifact `analyze` shows an analysis loading state, then either
  `analysis: ready` with parsed package metadata derived from the backend
  analysis response or `analysis: unavailable` when local analysis fails.
- Clicking artifact `details` after successful analysis shows the persisted
  ready analysis metadata from the artifact detail API.
- Restarting the server with the same `$DATA_DIR` preserves ready analysis
  detail state.
- Clicking artifact `delete` shows a delete loading state and then
  `delete: deleted`; the catalog refreshes to omit the deleted artifact, stale
  detail and analysis text are cleared, and the artifact directory is removed
  from `$DATA_DIR/artifacts`.
- No browser artifact control installs an APK, mutates a device, runs shell
  commands, or exposes stored host paths.

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

With that non-ready serial substituted for `$SERIAL`, package inventory should
return:

- HTTP status `409`.
- Error code `device_not_ready`.
- No route-specific `device` or `packages` field.

With that non-ready serial substituted for `$SERIAL`, package detail should
return:

- HTTP status `409`.
- Error code `device_not_ready`.
- No route-specific `device`, `package`, or `packages` field.

## Artifact Upload API

Create a disposable APK-like ZIP and upload it:

```sh
python3 - <<'PY'
import zipfile
with zipfile.ZipFile(".codex/tmp/manual-upload.apk", "w") as z:
    z.writestr("AndroidManifest.xml", "<manifest package=\"com.example.manual\" />")
PY

curl -i -sS -H "Origin: http://$ADDR" \
  -F "artifact=@.codex/tmp/manual-upload.apk;filename=manual-upload.apk;type=application/vnd.android.package-archive" \
  "http://$ADDR/api/v1/artifacts"
```

Expected result:

- HTTP status is `201`.
- JSON includes `artifact.id`, `artifact.originalName`,
  `artifact.sizeBytes`, `artifact.sha256`, `artifact.createdAt`, and
  `artifact.analysisStatus` equal to `pending`.
- `$DATA_DIR/artifacts/ARTIFACT_ID/original.apk` exists and matches the
  uploaded bytes.
- `$DATA_DIR/artifacts/ARTIFACT_ID/metadata.json` exists.
- Restarting the server with the same `$DATA_DIR` preserves both files.
- No ADB command runs, no APK is installed, and no host filesystem path is
  exposed in the response.

Check invalid and unsupported upload behavior:

```sh
printf 'not a zip' > .codex/tmp/manual-invalid.apk

curl -i -sS \
  -F "artifact=@.codex/tmp/manual-invalid.apk;filename=manual-invalid.apk;type=application/vnd.android.package-archive" \
  "http://$ADDR/api/v1/artifacts"

curl -i -sS -H "Content-Type: application/octet-stream" \
  --data-binary @.codex/tmp/manual-upload.apk \
  "http://$ADDR/api/v1/artifacts"
```

Expected result:

- Non-ZIP `.apk` upload returns HTTP `400` with code
  `invalid_artifact_upload`.
- Unsupported request media returns HTTP `415` with code
  `unsupported_artifact_media`.
- Failed uploads leave no additional artifact directory, no `original.apk`, and
  no temporary upload file.

Check security rejection:

```sh
curl -i -sS -H "Origin: http://foreign.example" \
  -F "artifact=@.codex/tmp/manual-upload.apk;filename=blocked.apk;type=application/vnd.android.package-archive" \
  "http://$ADDR/api/v1/artifacts"
```

Expected result:

- HTTP status is `403`.
- Error code is `forbidden_origin`.
- No request body is stored and no artifact metadata is created.

## Artifact Catalog And Detail API

After uploading one or more artifacts with the previous section, request the
catalog and detail API:

```sh
curl -i -sS "http://$ADDR/api/v1/artifacts"
curl -i -sS "http://$ADDR/api/v1/artifacts/$ARTIFACT_ID"
```

Expected result:

- The catalog returns HTTP `200` with `artifacts.items` and
  `artifacts.count`.
- Catalog items are sorted by `createdAt` descending, with `id` ascending when
  timestamps are equal.
- Detail returns HTTP `200` with `artifact` metadata and no `analysis` field
  until a ready analysis exists.
- Responses do not expose stored host paths, `original.apk`, or
  `metadata.json`.
- No ADB command runs, no APK is installed, and no browser is opened.

Check catalog and detail negative behavior:

```sh
curl -i -sS "http://$ADDR/api/v1/artifacts/unknown-artifact"
curl -i -sS -H "Host: foreign.example" "http://$ADDR/api/v1/artifacts"
curl -i -sS -H "Origin: http://foreign.example" \
  "http://$ADDR/api/v1/artifacts/$ARTIFACT_ID"
```

Expected result:

- Unknown artifact detail returns HTTP `404` with code `artifact_not_found`.
- Rejected Host requests return HTTP `403` with code `forbidden_host`.
- Rejected Origin requests return HTTP `403` with code `forbidden_origin`.
- Corrupt stored metadata returns HTTP `500` with code
  `artifact_catalog_unavailable` for catalog or detail requests.

## Artifact Analysis API

After uploading an artifact, request local APK analysis:

```sh
curl -i -sS -X POST -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/artifacts/$ARTIFACT_ID/analyze"

curl -i -sS "http://$ADDR/api/v1/artifacts/$ARTIFACT_ID"
```

Expected result:

- Analysis returns HTTP `200` with `artifact` and `analysis`.
- `artifact.analysisStatus` is `ready`.
- `analysis.tool` is `aapt`, `analysis.packageName` is non-empty, and
  `analysis.analyzedAt` is an RFC3339 timestamp.
- Optional fields such as `versionName`, `versionCode`, `minSdkVersion`,
  `targetSdkVersion`, `applicationLabel`, `launchableActivity`, and `warnings`
  appear only when reported by `aapt dump badging`.
- `$DATA_DIR/artifacts/ARTIFACT_ID/metadata.json` contains the latest ready
  analysis, and the detail response includes that `analysis`.
- Responses do not expose stored host paths, `original.apk`, `metadata.json`,
  command stderr, environment values, or token values.
- No ADB command runs, no APK is installed, and no network lookup is performed
  by the dashboard.

Check analysis negative behavior:

```sh
curl -i -sS -X POST -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/artifacts/unknown-artifact/analyze"

curl -i -sS -X POST -H "Host: foreign.example" \
  "http://$ADDR/api/v1/artifacts/$ARTIFACT_ID/analyze"
```

Expected result:

- Unknown artifact analysis returns HTTP `404` with code `artifact_not_found`.
- Missing `aapt`, nonzero `aapt`, timeout, oversized stdout, or output without
  `package: name=...` returns HTTP `502` with code
  `artifact_analysis_failed`; restart the dashboard with a `PATH` that omits
  `aapt` to check the missing-tool case.
- Failed analysis does not replace a prior ready analysis and does not store a
  false ready result.
- Rejected Host or Origin requests return HTTP `403` before `aapt` is invoked.

## Artifact Report API

After an artifact has ready analysis, request the JSON report:

```sh
curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/artifacts/$ARTIFACT_ID/report"

curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/artifacts/$ARTIFACT_ID/report?format=json"
```

Expected result:

- Each report request returns HTTP `200` with `Content-Type:
  application/json`.
- The response contains only top-level `report`.
- `report.artifact` contains stored artifact metadata and
  `analysisStatus=ready`.
- `report.analysis` contains the latest ready `aapt` analysis stored in
  metadata.
- `report.sections` appears in this order: `artifact`, `package`, `sdk`,
  `activity`, `warnings`, `localNotes`.
- Optional report fields and section items appear only when stored metadata or
  analysis contains values for them.
- The local notes section includes `Generated from local artifact metadata and
  latest ready analysis only.`.
- Report generation does not change
  `$DATA_DIR/artifacts/ARTIFACT_ID/metadata.json`.
- Report generation does not invoke `adb`, `aapt`, install or mutate an APK,
  write report files, send network requests, or expose stored host paths,
  `original.apk`, `metadata.json`, command stderr, environment values, or token
  values.

Request the Markdown report:

```sh
curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/artifacts/$ARTIFACT_ID/report?format=markdown"
```

Expected result:

- The report request returns HTTP `200` with `Content-Type: text/markdown`.
- The Markdown document contains the same semantic sections and values as the
  JSON `report.sections`, plus the artifact name, artifact ID, SHA-256, byte
  size, package name, version fields when present, SDK fields when present,
  application label when present, launchable activity when present, warning
  lines when present, and the local-only note.
- Markdown report generation has the same read-only and sensitive-output
  constraints as JSON report generation.

Check report negative behavior:

```sh
curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/artifacts/$ARTIFACT_ID/report?format=xml"

curl -i -sS -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/artifacts/unknown-artifact/report"

curl -i -sS -H "Host: foreign.example" \
  "http://$ADDR/api/v1/artifacts/$ARTIFACT_ID/report"
```

Expected result:

- Unsupported `format` values other than `json` or `markdown` return HTTP
  `400` with code `invalid_report_format`.
- Unknown artifact IDs return HTTP `404` with code `artifact_not_found`.
- Existing artifacts without ready analysis return HTTP `409` with code
  `artifact_report_unavailable`.
- Corrupt stored metadata returns HTTP `500` with code
  `artifact_catalog_unavailable`.
- Rejected Host or Origin requests return HTTP `403` before artifact lookup or
  report generation.

## Artifact Deletion API

After uploading an artifact, delete it explicitly:

```sh
curl -i -sS -X DELETE -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/artifacts/$ARTIFACT_ID"

curl -i -sS "http://$ADDR/api/v1/artifacts"
curl -i -sS "http://$ADDR/api/v1/artifacts/$ARTIFACT_ID"
```

Expected result:

- Deletion returns HTTP `200` with `artifact.id` equal to `$ARTIFACT_ID` and
  `deleted` equal to `true`.
- The deleted artifact is absent from the catalog response.
- Detail for the deleted artifact returns HTTP `404` with code
  `artifact_not_found`.
- `$DATA_DIR/artifacts/ARTIFACT_ID` no longer exists.
- Other artifact directories under `$DATA_DIR/artifacts` remain present.
- Responses do not expose stored host paths, `original.apk`, `metadata.json`,
  environment values, or token values.

Check deletion negative behavior:

```sh
curl -i -sS -X DELETE -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/artifacts/"

curl -i -sS -X DELETE -H "Origin: http://$ADDR" \
  "http://$ADDR/api/v1/artifacts/unknown-artifact"

curl -i -sS -X DELETE -H "Host: foreign.example" \
  "http://$ADDR/api/v1/artifacts/$OTHER_ARTIFACT_ID"
```

Expected result:

- Invalid artifact ID syntax returns HTTP `400` with code
  `invalid_artifact_request`.
- Unknown or already-deleted artifact IDs return HTTP `404` with code
  `artifact_not_found`.
- Filesystem deletion failure returns HTTP `500` with code
  `artifact_delete_failed`.
- Rejected Host or Origin requests return HTTP `403` before any artifact
  directory is removed.
- Deletion must not remove unrelated files or follow symlinks out of artifact
  storage.

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
  .codex/tmp/manual-screenshot-response.bin .codex/tmp/manual-upload.apk \
  .codex/tmp/manual-invalid.apk .codex/tmp/adb-dashboard
```

Do not remove `.codex/cache` directories; they may be used by repository test
commands.
