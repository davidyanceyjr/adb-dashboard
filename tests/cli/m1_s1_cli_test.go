package cli_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"
)

const expectedHelp = `Usage:
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
`

type commandResult struct {
	stdout string
	stderr string
	code   int
}

func TestM1S1HelpAndVersionDoNotStartServerOrTouchState(t *testing.T) {
	binary := buildDashboard(t)
	for _, args := range [][]string{
		{"--help"},
		{"version"},
		{"--version"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			env := isolatedEnv(t)
			result := runDashboard(t, binary, env, args...)

			if result.code != 0 {
				t.Fatalf("exit status = %d, stderr = %q", result.code, result.stderr)
			}
			if result.stderr != "" {
				t.Fatalf("stderr = %q, want empty", result.stderr)
			}
			if args[0] == "--help" {
				if result.stdout != expectedHelp {
					t.Fatalf("help stdout mismatch\nwant:\n%s\ngot:\n%s", expectedHelp, result.stdout)
				}
			} else {
				assertVersionOutput(t, result.stdout)
			}
			assertNoForbiddenSideEffects(t, env)
		})
	}
}

func TestM1S1InvalidInvocationsFailBeforeStartup(t *testing.T) {
	binary := buildDashboard(t)
	tests := []struct {
		name   string
		args   []string
		stderr string
	}{
		{name: "unknown_command", args: []string{"devices"}, stderr: "unknown command: devices\n"},
		{name: "unknown_option", args: []string{"--bogus"}, stderr: "unknown option: --bogus\n"},
		{name: "unknown_option_after_global_option", args: []string{"--listen", "127.0.0.1:0", "--bogus"}, stderr: "unknown option: --bogus\n"},
		{name: "missing_argument", args: []string{"--listen"}, stderr: "missing argument for --listen\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := isolatedEnv(t)
			result := runDashboard(t, binary, env, test.args...)

			if result.code != 2 {
				t.Fatalf("exit status = %d, want 2; stderr = %q", result.code, result.stderr)
			}
			if result.stdout != "" {
				t.Fatalf("stdout = %q, want empty", result.stdout)
			}
			if result.stderr != test.stderr {
				t.Fatalf("stderr = %q, want %q", result.stderr, test.stderr)
			}
			assertNoForbiddenSideEffects(t, env)
		})
	}
}

func TestM1S2DoctorDefaultsCreateOnlyCurrentStartupDirectories(t *testing.T) {
	binary := buildDashboard(t)
	env := isolatedEnv(t)
	values := envMap(env)

	result := runDashboard(t, binary, env, "doctor")

	if result.code != 0 {
		t.Fatalf("exit status = %d, want 0; stderr = %q; stdout = %q", result.code, result.stderr, result.stdout)
	}
	if result.stderr != "" {
		t.Fatalf("stderr = %q, want empty", result.stderr)
	}

	dataDir := filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard")
	tempDir := filepath.Join(values["XDG_RUNTIME_DIR"], "adb-dashboard")
	want := strings.Join([]string{
		"adb-dashboard doctor",
		"overall: PASS",
		"config: PASS source=defaults listen=127.0.0.1:8080 readOnly=false logLevel=info",
		"dataDir: PASS path=" + dataDir,
		"tempDir: PASS path=" + tempDir,
		"cacheDir: NIY storage.cache is not implemented yet",
		"projectDir: NIY storage.projects is not implemented yet",
		"adbExecutable: PASS path=" + fakeADBPath(t, env),
		"adbVersion: PASS version=Android Debug Bridge version 1.0.41",
		"adbServer: NIY adb.server is not implemented yet",
		"devices: NIY devices.refresh is not implemented yet",
		"hostTools: NIY hosttools.discovery is not implemented yet",
		"",
	}, "\n")
	if result.stdout != want {
		t.Fatalf("doctor stdout mismatch\nwant:\n%s\ngot:\n%s", want, result.stdout)
	}
	assertDirExists(t, dataDir)
	assertDirExists(t, tempDir)
	assertNoFutureStateOrExternalSideEffectsAllowADB(t, env)
}

func TestM2S1DoctorADBVersionDiscoverySuccess(t *testing.T) {
	binary := buildDashboard(t)
	env := isolatedEnv(t)
	values := envMap(env)
	adbPath := writeFakeADB(t, env, `#!/bin/sh
printf '%s\n' "$@" > "$ADB_MARKER"
if [ "$1" != "version" ] || [ "$#" -ne 1 ]; then
  echo "unexpected adb arguments" >&2
  exit 17
fi
printf '\nAndroid Debug Bridge version 1.0.41\nVersion 35.0.2\n'
`)

	result := runDashboard(t, binary, env, "doctor")

	if result.code != 0 {
		t.Fatalf("exit status = %d, want 0; stderr = %q; stdout = %q", result.code, result.stderr, result.stdout)
	}
	if result.stderr != "" {
		t.Fatalf("stderr = %q, want empty", result.stderr)
	}
	dataDir := filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard")
	tempDir := filepath.Join(values["XDG_RUNTIME_DIR"], "adb-dashboard")
	want := strings.Join([]string{
		"adb-dashboard doctor",
		"overall: PASS",
		"config: PASS source=defaults listen=127.0.0.1:8080 readOnly=false logLevel=info",
		"dataDir: PASS path=" + dataDir,
		"tempDir: PASS path=" + tempDir,
		"cacheDir: NIY storage.cache is not implemented yet",
		"projectDir: NIY storage.projects is not implemented yet",
		"adbExecutable: PASS path=" + adbPath,
		"adbVersion: PASS version=Android Debug Bridge version 1.0.41",
		"adbServer: NIY adb.server is not implemented yet",
		"devices: NIY devices.refresh is not implemented yet",
		"hostTools: NIY hosttools.discovery is not implemented yet",
		"",
	}, "\n")
	if result.stdout != want {
		t.Fatalf("doctor stdout mismatch\nwant:\n%s\ngot:\n%s", want, result.stdout)
	}
	assertFileContains(t, values["ADB_MARKER"], "version\n")
	assertDirExists(t, dataDir)
	assertDirExists(t, tempDir)
	assertPathAbsent(t, values["BROWSER_MARKER"])
}

func TestM2S1DoctorADBUnavailableExitsThree(t *testing.T) {
	binary := buildDashboard(t)
	env := isolatedEnv(t)
	values := envMap(env)
	env = removeFakeADB(t, env)

	result := runDashboard(t, binary, env, "doctor")

	if result.code != 3 {
		t.Fatalf("exit status = %d, want 3; stderr = %q; stdout = %q", result.code, result.stderr, result.stdout)
	}
	if result.stderr != "" {
		t.Fatalf("stderr = %q, want empty", result.stderr)
	}
	wantContains := []string{
		"overall: FAIL\n",
		"adbExecutable: FAIL error=not found in PATH\n",
		"adbVersion: NIY adb.version unavailable until adb executable is found\n",
		"adbServer: NIY adb.server is not implemented yet\n",
	}
	for _, want := range wantContains {
		if !strings.Contains(result.stdout, want) {
			t.Fatalf("doctor stdout missing %q\nstdout:\n%s", want, result.stdout)
		}
	}
	assertPathAbsent(t, values["ADB_MARKER"])
	assertPathAbsent(t, values["BROWSER_MARKER"])
}

func TestM2S1DoctorADBVersionFailuresExitThree(t *testing.T) {
	binary := buildDashboard(t)

	tests := []struct {
		name            string
		script          string
		wantErrorDetail string
	}{
		{
			name: "nonzero",
			script: `#!/bin/sh
printf '%s\n' "$@" > "$ADB_MARKER"
echo "adb exploded" >&2
exit 42
`,
			wantErrorDetail: "exit status 42",
		},
		{
			name: "empty_stdout",
			script: `#!/bin/sh
printf '%s\n' "$@" > "$ADB_MARKER"
printf '\n\n'
`,
			wantErrorDetail: "no version output",
		},
		{
			name: "timeout",
			script: `#!/bin/sh
printf '%s\n' "$@" > "$ADB_MARKER"
exec sleep 10
`,
			wantErrorDetail: "timed out",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := isolatedEnv(t)
			values := envMap(env)
			adbPath := writeFakeADB(t, env, test.script)

			result := runDashboard(t, binary, env, "doctor")

			if result.code != 3 {
				t.Fatalf("exit status = %d, want 3; stderr = %q; stdout = %q", result.code, result.stderr, result.stdout)
			}
			if result.stderr != "" {
				t.Fatalf("stderr = %q, want empty", result.stderr)
			}
			wantContains := []string{
				"overall: FAIL\n",
				"adbExecutable: PASS path=" + adbPath + "\n",
				"adbVersion: FAIL error=",
				test.wantErrorDetail,
				"adbServer: NIY adb.server is not implemented yet\n",
			}
			for _, want := range wantContains {
				if !strings.Contains(result.stdout, want) {
					t.Fatalf("doctor stdout missing %q\nstdout:\n%s", want, result.stdout)
				}
			}
			assertFileContains(t, values["ADB_MARKER"], "version\n")
			assertPathAbsent(t, values["BROWSER_MARKER"])
		})
	}
}

func TestM1S2DoctorReportsHighestPrioritySuccessfulConfiguration(t *testing.T) {
	binary := buildDashboard(t)
	env := isolatedEnv(t)
	values := envMap(env)

	userConfig := filepath.Join(values["XDG_CONFIG_HOME"], "adb-dashboard", "config.toml")
	envConfig := filepath.Join(values["HOME"], "env-config.toml")
	cliConfig := filepath.Join(values["HOME"], "cli-config.toml")
	writeFile(t, userConfig, `[server]
listen = "127.0.0.1:1111"
read_only = false
data_dir = "user-data"
temp_dir = "user-temp"

[logging]
level = "error"
`)
	writeFile(t, envConfig, `[server]
listen = "127.0.0.1:2222"
read_only = false
data_dir = "env-file-data"
temp_dir = "env-file-temp"

[logging]
level = "warn"
`)
	writeFile(t, cliConfig, `[server]
listen = "127.0.0.1:3333"
read_only = false
data_dir = "cli-file-data"
temp_dir = "cli-file-temp"

[logging]
level = "trace"
`)

	envDataDir := filepath.Join(values["HOME"], "env-data")
	cliTempDir := filepath.Join(values["HOME"], "cli-temp")
	env = setEnv(env, "ADB_DASHBOARD_CONFIG", envConfig)
	env = setEnv(env, "ADB_DASHBOARD_DATA_DIR", envDataDir)
	env = setEnv(env, "ADB_DASHBOARD_LISTEN", "127.0.0.1:4444")

	result := runDashboard(t, binary, env,
		"doctor",
		"--config", cliConfig,
		"--listen", "127.0.0.1:5555",
		"--temp-dir", cliTempDir,
		"--read-only",
	)

	if result.code != 0 {
		t.Fatalf("exit status = %d, want 0; stderr = %q; stdout = %q", result.code, result.stderr, result.stdout)
	}
	if result.stderr != "" {
		t.Fatalf("stderr = %q, want empty", result.stderr)
	}
	wantContains := []string{
		"config: PASS source=mixed listen=127.0.0.1:5555 readOnly=true logLevel=trace",
		"dataDir: PASS path=" + envDataDir,
		"tempDir: PASS path=" + cliTempDir,
		"cacheDir: NIY storage.cache is not implemented yet",
		"adbExecutable: PASS path=" + fakeADBPath(t, env),
		"adbVersion: PASS version=Android Debug Bridge version 1.0.41",
		"hostTools: NIY hosttools.discovery is not implemented yet",
	}
	for _, want := range wantContains {
		if !strings.Contains(result.stdout, want) {
			t.Fatalf("doctor stdout missing %q\nstdout:\n%s", want, result.stdout)
		}
	}
	assertDirExists(t, envDataDir)
	assertDirExists(t, cliTempDir)
	assertPathAbsent(t, filepath.Join(values["HOME"], "cli-file-data"))
	assertPathAbsent(t, filepath.Join(values["HOME"], "env-file-temp"))
	assertPathAbsent(t, filepath.Join(values["HOME"], "user-data"))
	assertPathAbsent(t, filepath.Join(values["HOME"], "user-temp"))
	assertNoFutureStateOrExternalSideEffectsAllowADB(t, env)
}

func TestM1S3ConfigurationFailuresExitBeforeReportOrStartup(t *testing.T) {
	binary := buildDashboard(t)

	tests := []struct {
		name       string
		command    []string
		envConfig  bool
		envMutate  func([]string) []string
		configBody string
		wantStderr func(configPath string) string
	}{
		{
			name:    "missing_explicit_cli_config",
			command: []string{"doctor"},
			wantStderr: func(configPath string) string {
				return "invalid configuration: cannot load " + configPath + ": "
			},
		},
		{
			name:      "missing_explicit_env_config",
			command:   []string{"doctor"},
			envConfig: true,
			wantStderr: func(configPath string) string {
				return "invalid configuration: cannot load " + configPath + ": "
			},
		},
		{
			name:       "malformed_toml",
			command:    []string{"doctor"},
			configBody: "[server\nlisten = \"127.0.0.1:8080\"\n",
			wantStderr: func(configPath string) string {
				return "invalid configuration: cannot parse " + configPath + ": "
			},
		},
		{
			name:       "unknown_key",
			command:    []string{"doctor"},
			configBody: "[server]\nunknown = true\n",
			wantStderr: func(string) string {
				return "invalid configuration: unknown key server.unknown\n"
			},
		},
		{
			name:       "listen_not_host_port",
			command:    []string{"doctor"},
			configBody: "[server]\nlisten = \"127.0.0.1\"\n",
			wantStderr: func(string) string {
				return "invalid configuration: server.listen must be HOST:PORT\n"
			},
		},
		{
			name:       "listen_empty_host",
			command:    []string{"doctor"},
			configBody: "[server]\nlisten = \":8080\"\n",
			wantStderr: func(string) string {
				return "invalid configuration: server.listen host must not be empty\n"
			},
		},
		{
			name:       "listen_non_integer_port",
			command:    []string{"doctor"},
			configBody: "[server]\nlisten = \"127.0.0.1:nope\"\n",
			wantStderr: func(string) string {
				return "invalid configuration: server.listen port must be 0 through 65535: nope\n"
			},
		},
		{
			name:       "listen_out_of_range_port",
			command:    []string{"doctor"},
			configBody: "[server]\nlisten = \"127.0.0.1:70000\"\n",
			wantStderr: func(string) string {
				return "invalid configuration: server.listen port must be 0 through 65535: 70000\n"
			},
		},
		{
			name:       "invalid_log_level",
			command:    []string{"doctor"},
			configBody: "[logging]\nlevel = \"verbose\"\n",
			wantStderr: func(string) string {
				return "invalid configuration: logging.level must be one of error, warn, info, debug, trace: verbose\n"
			},
		},
		{
			name:       "wrong_type",
			command:    []string{"doctor"},
			configBody: "[server]\nread_only = \"yes\"\n",
			wantStderr: func(string) string {
				return "invalid configuration: server.read_only must be bool\n"
			},
		},
		{
			name:       "unsupported_path_expansion",
			command:    []string{"doctor"},
			configBody: "[server]\ndata_dir = \"~bad\"\n",
			wantStderr: func(string) string {
				return "invalid configuration: server.data_dir contains unsupported path expansion: ~bad\n"
			},
		},
		{
			name:       "unset_path_environment",
			command:    []string{"doctor"},
			configBody: "[server]\ndata_dir = \"$ADB_DASHBOARD_MISSING/path\"\n",
			wantStderr: func(string) string {
				return "invalid configuration: server.data_dir references unset environment variable: ADB_DASHBOARD_MISSING\n"
			},
		},
		{
			name:       "serve_invalid_config",
			command:    []string{"serve"},
			configBody: "[logging]\nlevel = \"verbose\"\n",
			wantStderr: func(string) string {
				return "invalid configuration: logging.level must be one of error, warn, info, debug, trace: verbose\n"
			},
		},
		{
			name:    "cli_invalid_listen",
			command: []string{"doctor", "--listen", "127.0.0.1"},
			wantStderr: func(string) string {
				return "invalid configuration: server.listen must be HOST:PORT\n"
			},
		},
		{
			name:    "env_invalid_log_level",
			command: []string{"doctor"},
			envMutate: func(env []string) []string {
				return setEnv(env, "ADB_DASHBOARD_LOG_LEVEL", "verbose")
			},
			wantStderr: func(string) string {
				return "invalid configuration: logging.level must be one of error, warn, info, debug, trace: verbose\n"
			},
		},
		{
			name:    "default_home_fallback_unset",
			command: []string{"doctor"},
			envMutate: func(env []string) []string {
				env = setEnv(env, "HOME", "")
				return setEnv(env, "XDG_STATE_HOME", "")
			},
			wantStderr: func(string) string {
				return "invalid configuration: server.data_dir references unset environment variable: HOME\n"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := isolatedEnv(t)
			values := envMap(env)
			configPath := filepath.Join(values["HOME"], test.name+".toml")
			args := append([]string{}, test.command...)
			if test.envMutate != nil {
				env = test.envMutate(env)
			}
			if test.configBody != "" {
				writeFile(t, configPath, test.configBody)
			}
			if test.envConfig {
				env = setEnv(env, "ADB_DASHBOARD_CONFIG", configPath)
			} else if test.configBody != "" || test.name == "missing_explicit_cli_config" {
				args = append(args, "--config", configPath)
			}

			result := runDashboard(t, binary, env, args...)

			if result.code != 2 {
				t.Fatalf("exit status = %d, want 2; stderr = %q; stdout = %q", result.code, result.stderr, result.stdout)
			}
			if result.stdout != "" {
				t.Fatalf("stdout = %q, want empty", result.stdout)
			}
			want := test.wantStderr(configPath)
			if strings.HasSuffix(want, ": ") {
				if !strings.HasPrefix(result.stderr, want) {
					t.Fatalf("stderr = %q, want prefix %q", result.stderr, want)
				}
			} else if result.stderr != want {
				t.Fatalf("stderr = %q, want %q", result.stderr, want)
			}
			assertNoForbiddenSideEffects(t, env)
		})
	}
}

func TestM1S3DoctorReportsStartupDirectoryFailures(t *testing.T) {
	binary := buildDashboard(t)
	env := isolatedEnv(t)
	values := envMap(env)

	dataFile := filepath.Join(values["HOME"], "data-file")
	tempDir := filepath.Join(values["HOME"], "temp-dir")
	writeFile(t, dataFile, "not a directory\n")

	result := runDashboard(t, binary, env,
		"doctor",
		"--data-dir", dataFile,
		"--temp-dir", tempDir,
	)

	if result.code != 5 {
		t.Fatalf("exit status = %d, want 5; stderr = %q; stdout = %q", result.code, result.stderr, result.stdout)
	}
	if result.stderr != "" {
		t.Fatalf("stderr = %q, want empty", result.stderr)
	}
	wantContains := []string{
		"adb-dashboard doctor\n",
		"overall: FAIL\n",
		"config: PASS source=mixed listen=127.0.0.1:8080 readOnly=false logLevel=info\n",
		"dataDir: FAIL path=" + dataFile + " error=",
		"tempDir: PASS path=" + tempDir + "\n",
		"cacheDir: NIY storage.cache is not implemented yet\n",
		"adbExecutable: PASS path=" + fakeADBPath(t, env) + "\n",
		"adbVersion: PASS version=Android Debug Bridge version 1.0.41\n",
		"hostTools: NIY hosttools.discovery is not implemented yet\n",
	}
	for _, want := range wantContains {
		if !strings.Contains(result.stdout, want) {
			t.Fatalf("doctor stdout missing %q\nstdout:\n%s", want, result.stdout)
		}
	}
	assertDirExists(t, tempDir)
	assertNoFutureStateOrExternalSideEffectsAllowADB(t, env)
}

func TestM1S3StartupDirectoryFailuresExitBeforeServerSideEffects(t *testing.T) {
	binary := buildDashboard(t)

	tests := []struct {
		name        string
		args        []string
		configure   func(t *testing.T, env []string) (nextEnv []string, dataDir string, tempDir string)
		wantKind    string
		wantPath    func(dataDir, tempDir string) string
		wantCreated func(dataDir, tempDir string) string
	}{
		{
			name: "serve_data_path_is_file",
			args: []string{"serve"},
			configure: func(t *testing.T, env []string) ([]string, string, string) {
				values := envMap(env)
				dataFile := filepath.Join(values["HOME"], "data-file")
				tempDir := filepath.Join(values["HOME"], "temp-dir")
				writeFile(t, dataFile, "not a directory\n")
				return env, dataFile, tempDir
			},
			wantKind: "data",
			wantPath: func(dataDir, tempDir string) string {
				return dataDir
			},
		},
		{
			name: "no_subcommand_temp_path_is_file_after_data_created",
			args: nil,
			configure: func(t *testing.T, env []string) ([]string, string, string) {
				values := envMap(env)
				dataDir := filepath.Join(values["HOME"], "data-dir")
				tempFile := filepath.Join(values["HOME"], "temp-file")
				writeFile(t, tempFile, "not a directory\n")
				env = setEnv(env, "ADB_DASHBOARD_DATA_DIR", dataDir)
				env = setEnv(env, "ADB_DASHBOARD_TEMP_DIR", tempFile)
				return env, dataDir, tempFile
			},
			wantKind: "temp",
			wantPath: func(dataDir, tempDir string) string {
				return tempDir
			},
			wantCreated: func(dataDir, tempDir string) string {
				return dataDir
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := isolatedEnv(t)
			env, dataDir, tempDir := test.configure(t, env)
			args := append([]string{}, test.args...)
			if len(args) > 0 {
				args = append(args, "--data-dir", dataDir, "--temp-dir", tempDir, "--open")
			}

			result := runDashboard(t, binary, env, args...)

			if result.code != 5 {
				t.Fatalf("exit status = %d, want 5; stderr = %q; stdout = %q", result.code, result.stderr, result.stdout)
			}
			if result.stdout != "" {
				t.Fatalf("stdout = %q, want empty", result.stdout)
			}
			wantPrefix := "server runtime failure: startup filesystem unavailable: " + test.wantKind + " directory " + test.wantPath(dataDir, tempDir) + ": "
			if !strings.HasPrefix(result.stderr, wantPrefix) {
				t.Fatalf("stderr = %q, want prefix %q", result.stderr, wantPrefix)
			}
			if test.wantCreated != nil {
				assertDirExists(t, test.wantCreated(dataDir, tempDir))
			}
			assertNoFutureStateOrExternalSideEffects(t, env)
		})
	}
}

func TestM1S4ServerLifecycleStatusUnknownRouteAndBrowserOpen(t *testing.T) {
	binary := buildDashboard(t)
	env := isolatedEnv(t)
	values := envMap(env)
	dataDir := filepath.Join(values["HOME"], "data-dir")
	tempDir := filepath.Join(values["HOME"], "temp-dir")

	server := startDashboard(t, binary, env,
		"serve",
		"--listen", "127.0.0.1:0",
		"--data-dir", dataDir,
		"--temp-dir", tempDir,
		"--read-only",
		"--open",
	)
	defer server.cleanup(t)

	startedLine := server.waitForStderrLine(t, regexp.MustCompile(`^\S+ INFO server started addr=127\.0\.0\.1:\d+$`))
	addr := strings.TrimPrefix(startedLine[strings.LastIndex(startedLine, "addr="):], "addr=")
	baseURL := "http://" + addr

	assertDirExists(t, dataDir)
	assertDirExists(t, tempDir)
	assertPathAbsent(t, values["ADB_MARKER"])
	waitFileContains(t, values["BROWSER_MARKER"], baseURL+"/\n")

	statusCode, contentType, body := httpGet(t, baseURL+"/api/v1/status")
	if statusCode != http.StatusOK {
		t.Fatalf("status code = %d, body = %s", statusCode, body)
	}
	if !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("content type = %q, want application/json", contentType)
	}
	assertM2S2StatusJSON(t, body, addr, true, map[string]any{
		"status":           "available",
		"executable":       fakeADBPath(t, env),
		"version":          "Android Debug Bridge version 1.0.41",
		"serverResponsive": "NIY",
	})
	assertFileContains(t, values["ADB_MARKER"], "version\n")

	statusCode, contentType, body = httpGet(t, baseURL+"/api/v1/unknown")
	if statusCode != http.StatusNotFound {
		t.Fatalf("unknown route status = %d, body = %s", statusCode, body)
	}
	if !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("unknown route content type = %q, want application/json", contentType)
	}
	assertJSONEqual(t, body, map[string]any{
		"error": map[string]any{
			"code":      "not_found",
			"message":   "Unknown API route",
			"details":   map[string]any{},
			"requestId": nil,
		},
	})

	server.signal(t, syscall.SIGTERM)
	result := server.wait(t)
	if result.code != 0 {
		t.Fatalf("exit status = %d, want 0; stderr = %q; stdout = %q", result.code, result.stderr, result.stdout)
	}
	if result.stdout != "" {
		t.Fatalf("stdout = %q, want empty", result.stdout)
	}
	if !regexp.MustCompile(`\S+ INFO server stopped signal=terminated`).MatchString(result.stderr) {
		t.Fatalf("stderr missing shutdown diagnostic: %q", result.stderr)
	}
	assertNoFutureStateSideEffectsAllowBrowserAndADB(t, env)
}

func TestM2S2StatusAPIADBSummary(t *testing.T) {
	binary := buildDashboard(t)

	t.Run("available", func(t *testing.T) {
		env := isolatedEnv(t)
		values := envMap(env)
		adbPath := writeFakeADB(t, env, `#!/bin/sh
printf '%s\n' "$@" >> "$ADB_MARKER"
if [ "$1" != "version" ] || [ "$#" -ne 1 ]; then
  echo "unexpected adb arguments" >&2
  exit 17
fi
printf '\nAndroid Debug Bridge version status-success\nVersion 35.0.2\n'
`)
		server := startDashboard(t, binary, env, "serve", "--listen", "127.0.0.1:0", "--no-open")
		defer server.cleanup(t)

		addr := serverAddressFromStartLine(t, server.waitForStderrLine(t, regexp.MustCompile(`^\S+ INFO server started addr=127\.0\.0\.1:\d+$`)))
		response := requestJSON(t, httpRequestSpec{method: "GET", url: "http://" + addr + "/api/v1/status"})

		if response.statusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", response.statusCode, response.body)
		}
		if !strings.HasPrefix(response.contentType, "application/json") {
			t.Fatalf("content type = %q, want application/json", response.contentType)
		}
		assertM2S2StatusJSON(t, response.body, addr, false, map[string]any{
			"status":           "available",
			"executable":       adbPath,
			"version":          "Android Debug Bridge version status-success",
			"serverResponsive": "NIY",
		})
		assertFileContains(t, values["ADB_MARKER"], "version\n")
		assertPathAbsent(t, values["BROWSER_MARKER"])
	})

	t.Run("unavailable", func(t *testing.T) {
		env := isolatedEnv(t)
		values := envMap(env)
		env = removeFakeADB(t, env)
		server := startDashboard(t, binary, env, "serve", "--listen", "127.0.0.1:0", "--no-open")
		defer server.cleanup(t)

		addr := serverAddressFromStartLine(t, server.waitForStderrLine(t, regexp.MustCompile(`^\S+ INFO server started addr=127\.0\.0\.1:\d+$`)))
		response := requestJSON(t, httpRequestSpec{method: "GET", url: "http://" + addr + "/api/v1/status"})

		if response.statusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", response.statusCode, response.body)
		}
		assertM2S2StatusJSON(t, response.body, addr, false, map[string]any{
			"status":           "unavailable",
			"executable":       nil,
			"version":          nil,
			"serverResponsive": "NIY",
		})
		assertPathAbsent(t, values["ADB_MARKER"])
		assertPathAbsent(t, values["BROWSER_MARKER"])
	})

	t.Run("version_failure", func(t *testing.T) {
		env := isolatedEnv(t)
		values := envMap(env)
		adbPath := writeFakeADB(t, env, `#!/bin/sh
printf '%s\n' "$@" >> "$ADB_MARKER"
echo "secret stderr $HOME $ADB_DASHBOARD_LOG_LEVEL" >&2
exit 42
`)
		server := startDashboard(t, binary, env, "serve", "--listen", "127.0.0.1:0", "--no-open")
		defer server.cleanup(t)

		addr := serverAddressFromStartLine(t, server.waitForStderrLine(t, regexp.MustCompile(`^\S+ INFO server started addr=127\.0\.0\.1:\d+$`)))
		response := requestJSON(t, httpRequestSpec{method: "GET", url: "http://" + addr + "/api/v1/status"})

		if response.statusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", response.statusCode, response.body)
		}
		assertM2S2StatusJSON(t, response.body, addr, false, map[string]any{
			"status":           "error",
			"executable":       adbPath,
			"version":          nil,
			"serverResponsive": "NIY",
		})
		assertFileContains(t, values["ADB_MARKER"], "version\n")
		for _, forbidden := range []string{"secret stderr", values["HOME"], "ADB_DASHBOARD"} {
			if strings.Contains(string(response.body), forbidden) {
				t.Fatalf("status JSON contains forbidden text %q: %s", forbidden, response.body)
			}
		}
	})

	t.Run("version_timeout", func(t *testing.T) {
		env := isolatedEnv(t)
		values := envMap(env)
		adbPath := writeFakeADB(t, env, `#!/bin/sh
printf '%s\n' "$@" >> "$ADB_MARKER"
exec sleep 10
`)
		server := startDashboard(t, binary, env, "serve", "--listen", "127.0.0.1:0", "--no-open")
		defer server.cleanup(t)

		addr := serverAddressFromStartLine(t, server.waitForStderrLine(t, regexp.MustCompile(`^\S+ INFO server started addr=127\.0\.0\.1:\d+$`)))
		response := requestJSON(t, httpRequestSpec{method: "GET", url: "http://" + addr + "/api/v1/status"})

		if response.statusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", response.statusCode, response.body)
		}
		assertM2S2StatusJSON(t, response.body, addr, false, map[string]any{
			"status":           "error",
			"executable":       adbPath,
			"version":          nil,
			"serverResponsive": "NIY",
		})
		assertFileContains(t, values["ADB_MARKER"], "version\n")
		assertPathAbsent(t, values["BROWSER_MARKER"])
	})

	t.Run("security_rejection_before_adb", func(t *testing.T) {
		env := isolatedEnv(t)
		values := envMap(env)
		server := startDashboard(t, binary, env, "serve", "--listen", "127.0.0.1:0", "--no-open")
		defer server.cleanup(t)

		addr := serverAddressFromStartLine(t, server.waitForStderrLine(t, regexp.MustCompile(`^\S+ INFO server started addr=127\.0\.0\.1:\d+$`)))
		baseURL := "http://" + addr
		for _, test := range []struct {
			name     string
			request  httpRequestSpec
			wantCode string
		}{
			{
				name: "foreign_host",
				request: httpRequestSpec{
					method: "GET",
					url:    baseURL + "/api/v1/status",
					host:   "foreign.example",
				},
				wantCode: "forbidden_host",
			},
			{
				name: "foreign_origin",
				request: httpRequestSpec{
					method: "GET",
					url:    baseURL + "/api/v1/status",
					headers: map[string]string{
						"Origin": "http://foreign.example",
					},
				},
				wantCode: "forbidden_origin",
			},
		} {
			t.Run(test.name, func(t *testing.T) {
				response := requestJSON(t, test.request)
				if response.statusCode != http.StatusForbidden {
					t.Fatalf("status = %d, want 403; body = %s", response.statusCode, response.body)
				}
				assertSecurityErrorEnvelope(t, response.body, test.wantCode)
				assertPathAbsent(t, values["ADB_MARKER"])
			})
		}
	})
}

func TestM1S4NoSubcommandStartsServer(t *testing.T) {
	binary := buildDashboard(t)
	env := isolatedEnv(t)
	values := envMap(env)
	dataDir := filepath.Join(values["HOME"], "data-dir")
	tempDir := filepath.Join(values["HOME"], "temp-dir")

	server := startDashboard(t, binary, env,
		"--listen", "127.0.0.1:0",
		"--data-dir", dataDir,
		"--temp-dir", tempDir,
		"--no-open",
	)
	defer server.cleanup(t)

	startedLine := server.waitForStderrLine(t, regexp.MustCompile(`^\S+ INFO server started addr=127\.0\.0\.1:\d+$`))
	addr := strings.TrimPrefix(startedLine[strings.LastIndex(startedLine, "addr="):], "addr=")
	statusCode, _, body := httpGet(t, "http://"+addr+"/api/v1/status")
	if statusCode != http.StatusOK {
		t.Fatalf("status code = %d, body = %s", statusCode, body)
	}

	server.signal(t, syscall.SIGINT)
	result := server.wait(t)
	if result.code != 0 {
		t.Fatalf("exit status = %d, want 0; stderr = %q; stdout = %q", result.code, result.stderr, result.stdout)
	}
	if result.stdout != "" {
		t.Fatalf("stdout = %q, want empty", result.stdout)
	}
	if !regexp.MustCompile(`\S+ INFO server stopped signal=interrupt`).MatchString(result.stderr) {
		t.Fatalf("stderr missing shutdown diagnostic: %q", result.stderr)
	}
}

func TestM1S4ListenValidationAndUnavailableAddressFailures(t *testing.T) {
	binary := buildDashboard(t)

	t.Run("non_loopback_host", func(t *testing.T) {
		env := isolatedEnv(t)
		result := runDashboard(t, binary, env, "serve", "--listen", "0.0.0.0:0")
		if result.code != 2 {
			t.Fatalf("exit status = %d, want 2; stderr = %q", result.code, result.stderr)
		}
		if result.stdout != "" {
			t.Fatalf("stdout = %q, want empty", result.stdout)
		}
		want := "invalid configuration: server.listen must use a loopback host: 0.0.0.0\n"
		if result.stderr != want {
			t.Fatalf("stderr = %q, want %q", result.stderr, want)
		}
		assertNoForbiddenSideEffects(t, env)
	})

	t.Run("unavailable_loopback_address", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		defer listener.Close()

		env := isolatedEnv(t)
		result := runDashboard(t, binary, env, "serve", "--listen", listener.Addr().String())
		if result.code != 4 {
			t.Fatalf("exit status = %d, want 4; stderr = %q", result.code, result.stderr)
		}
		if result.stdout != "" {
			t.Fatalf("stdout = %q, want empty", result.stdout)
		}
		wantPrefix := "listen address unavailable: " + listener.Addr().String() + ": "
		if !strings.HasPrefix(result.stderr, wantPrefix) {
			t.Fatalf("stderr = %q, want prefix %q", result.stderr, wantPrefix)
		}
		assertNoFutureStateOrExternalSideEffects(t, env)
	})
}

func TestM1S4BrowserOpenFailureWarnsAndServerContinues(t *testing.T) {
	binary := buildDashboard(t)
	env := isolatedEnv(t)
	values := envMap(env)
	binDir := strings.Split(values["PATH"], string(os.PathListSeparator))[0]
	if err := os.WriteFile(filepath.Join(binDir, "xdg-open"), []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$BROWSER_MARKER\"\nexit 9\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	server := startDashboard(t, binary, env, "serve", "--listen", "127.0.0.1:0", "--open")
	defer server.cleanup(t)

	startedLine := server.waitForStderrLine(t, regexp.MustCompile(`^\S+ INFO server started addr=127\.0\.0\.1:\d+$`))
	addr := strings.TrimPrefix(startedLine[strings.LastIndex(startedLine, "addr="):], "addr=")
	warnLine := server.waitForStderrLine(t, regexp.MustCompile(`^\S+ WARN browser open failed url=http://127\.0\.0\.1:\d+/ error=`))
	if !strings.Contains(warnLine, "exit status 9") {
		t.Fatalf("warning line = %q, want exit status 9", warnLine)
	}
	statusCode, _, body := httpGet(t, "http://"+addr+"/api/v1/status")
	if statusCode != http.StatusOK {
		t.Fatalf("status code after browser warning = %d, body = %s", statusCode, body)
	}

	server.signal(t, syscall.SIGTERM)
	result := server.wait(t)
	if result.code != 0 {
		t.Fatalf("exit status = %d, want 0; stderr = %q", result.code, result.stderr)
	}
	if result.stdout != "" {
		t.Fatalf("stdout = %q, want empty", result.stdout)
	}
}

func TestM1S5BootstrapTokensAndSecurityPolicy(t *testing.T) {
	binary := buildDashboard(t)
	env := isolatedEnv(t)

	first := startDashboard(t, binary, env, "serve", "--listen", "127.0.0.1:0", "--no-open")
	firstAddr := serverAddressFromStartLine(t, first.waitForStderrLine(t, regexp.MustCompile(`^\S+ INFO server started addr=127\.0\.0\.1:\d+$`)))
	firstBaseURL := "http://" + firstAddr

	firstBootstrap := requestJSON(t, httpRequestSpec{
		method: "GET",
		url:    firstBaseURL + "/api/v1/bootstrap",
		headers: map[string]string{
			"Origin": firstBaseURL,
		},
	})
	if firstBootstrap.statusCode != http.StatusOK {
		t.Fatalf("bootstrap status = %d, body = %s", firstBootstrap.statusCode, firstBootstrap.body)
	}
	if !strings.HasPrefix(firstBootstrap.contentType, "application/json") {
		t.Fatalf("bootstrap content type = %q, want application/json", firstBootstrap.contentType)
	}
	firstTokens := assertBootstrapJSON(t, firstBootstrap.body)

	status := requestJSON(t, httpRequestSpec{method: "GET", url: firstBaseURL + "/api/v1/status"})
	if status.statusCode != http.StatusOK {
		t.Fatalf("status code = %d, body = %s", status.statusCode, status.body)
	}
	assertNoTokenDisclosure(t, status.body, firstTokens)

	for _, test := range []struct {
		name     string
		request  httpRequestSpec
		wantCode string
	}{
		{
			name: "foreign_host",
			request: httpRequestSpec{
				method: "GET",
				url:    firstBaseURL + "/api/v1/status",
				host:   "foreign.example",
			},
			wantCode: "forbidden_host",
		},
		{
			name: "foreign_absolute_url_host",
			request: httpRequestSpec{
				method:      "GET",
				url:         firstBaseURL + "/api/v1/status",
				requestURI:  "http://foreign.example/api/v1/status",
				rawHTTPHost: firstAddr,
			},
			wantCode: "forbidden_absolute_url_host",
		},
		{
			name: "foreign_origin",
			request: httpRequestSpec{
				method: "GET",
				url:    firstBaseURL + "/api/v1/bootstrap",
				headers: map[string]string{
					"Origin": "http://foreign.example",
				},
			},
			wantCode: "forbidden_origin",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := requestJSON(t, test.request)
			if response.statusCode != http.StatusForbidden {
				t.Fatalf("status = %d, want 403; body = %s", response.statusCode, response.body)
			}
			if !strings.HasPrefix(response.contentType, "application/json") {
				t.Fatalf("content type = %q, want application/json", response.contentType)
			}
			assertSecurityErrorEnvelope(t, response.body, test.wantCode)
			assertNoTokenDisclosure(t, response.body, firstTokens)
		})
	}

	first.signal(t, syscall.SIGTERM)
	result := first.wait(t)
	if result.code != 0 {
		t.Fatalf("first server exit status = %d, want 0; stderr = %q", result.code, result.stderr)
	}

	second := startDashboard(t, binary, env, "serve", "--listen", "127.0.0.1:0", "--no-open")
	defer second.cleanup(t)
	secondAddr := serverAddressFromStartLine(t, second.waitForStderrLine(t, regexp.MustCompile(`^\S+ INFO server started addr=127\.0\.0\.1:\d+$`)))
	secondBootstrap := requestJSON(t, httpRequestSpec{method: "GET", url: "http://" + secondAddr + "/api/v1/bootstrap"})
	if secondBootstrap.statusCode != http.StatusOK {
		t.Fatalf("second bootstrap status = %d, body = %s", secondBootstrap.statusCode, secondBootstrap.body)
	}
	secondTokens := assertBootstrapJSON(t, secondBootstrap.body)
	if firstTokens.csrfToken == secondTokens.csrfToken {
		t.Fatalf("csrfToken reused after restart: %q", firstTokens.csrfToken)
	}
	if firstTokens.webSocketToken == secondTokens.webSocketToken {
		t.Fatalf("webSocketToken reused after restart: %q", firstTokens.webSocketToken)
	}
}

func TestM1S6EmbeddedBrowserShellRendersBackendState(t *testing.T) {
	binary := buildDashboard(t)
	env := isolatedEnv(t)

	server := startDashboard(t, binary, env,
		"serve",
		"--listen", "127.0.0.1:0",
		"--read-only",
		"--no-open",
	)
	defer server.cleanup(t)

	addr := serverAddressFromStartLine(t, server.waitForStderrLine(t, regexp.MustCompile(`^\S+ INFO server started addr=127\.0\.0\.1:\d+$`)))
	baseURL := "http://" + addr

	statusCode, contentType, body := httpGet(t, baseURL+"/")
	if statusCode != http.StatusOK {
		t.Fatalf("root status = %d, body = %s", statusCode, body)
	}
	if !strings.HasPrefix(contentType, "text/html") {
		t.Fatalf("root content type = %q, want text/html", contentType)
	}

	html := string(body)
	for _, forbidden := range []string{
		"csrfToken",
		"webSocketToken",
		"<button",
		"<form",
		"href=",
		"devices",
		"raw command",
		"logcat",
		"transfers",
		"artifacts",
		"reboot",
		"install",
	} {
		if strings.Contains(strings.ToLower(html), strings.ToLower(forbidden)) {
			t.Fatalf("root shell contains forbidden text %q:\n%s", forbidden, html)
		}
	}

	script := extractInlineScript(t, html)
	rendered := runFrontendScript(t, script, baseURL, false)
	for _, want := range []string{
		"adb-dashboard",
		"server: running",
		"bind: " + addr,
		"read-only: true",
		"adb: available",
		"watcher: NIY",
		"jobs: NIY",
		"sessions: NIY",
		"storage: NIY",
		"host tools: NIY",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered shell missing %q\nrendered:\n%s", want, rendered)
		}
	}
	for _, forbidden := range []string{"csrfToken", "webSocketToken"} {
		if strings.Contains(rendered, forbidden) {
			t.Fatalf("rendered shell disclosed %q:\n%s", forbidden, rendered)
		}
	}

	unavailable := runFrontendScript(t, script, baseURL, true)
	if !strings.Contains(unavailable, "server: unavailable") {
		t.Fatalf("failure render missing server unavailable state:\n%s", unavailable)
	}
	for _, forbidden := range []string{
		"server: running",
		"adb: running",
		"watcher: running",
		"jobs: running",
		"sessions: running",
		"storage: running",
		"host tools: running",
	} {
		if strings.Contains(unavailable, forbidden) {
			t.Fatalf("failure render contains success state %q:\n%s", forbidden, unavailable)
		}
	}
}

func buildDashboard(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "adb-dashboard")
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/adb-dashboard")
	cmd.Dir = repoRoot(t)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build adb-dashboard: %v\nstderr:\n%s", err, stderr.String())
	}
	return binary
}

func runDashboard(t *testing.T, binary string, env []string, args ...string) commandResult {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("command timed out, possible server startup: %s %s", binary, strings.Join(args, " "))
	}

	code := 0
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("command failed without exit status: %v", err)
		}
		code = exitErr.ExitCode()
	}

	return commandResult{stdout: stdout.String(), stderr: stderr.String(), code: code}
}

func extractInlineScript(t *testing.T, html string) string {
	t.Helper()

	startMarker := "<script>"
	endMarker := "</script>"
	start := strings.Index(html, startMarker)
	end := strings.LastIndex(html, endMarker)
	if start < 0 || end < 0 || end <= start {
		t.Fatalf("root shell did not contain an inline script:\n%s", html)
	}
	return html[start+len(startMarker) : end]
}

func runFrontendScript(t *testing.T, script, baseURL string, failStatus bool) string {
	t.Helper()

	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Fatalf("node is required for deterministic browser-shell script execution: %v", err)
	}

	harness := fmt.Sprintf(`
const baseURL = %q;
const failStatus = %t;
const elements = {};
function element(id) {
  if (!elements[id]) {
    elements[id] = { textContent: "" };
  }
  return elements[id];
}
global.window = {};
global.document = {
  getElementById: element,
  addEventListener: (event, callback) => {
    if (event === "DOMContentLoaded") {
      Promise.resolve().then(callback);
    }
  },
};
const realFetch = global.fetch;
global.fetch = async (target, options = {}) => {
  const url = new URL(target, baseURL);
  if (failStatus && url.pathname === "/api/v1/status") {
    return { ok: false, status: 503, json: async () => ({}) };
  }
  return realFetch(url.href, {
    ...options,
    headers: {
      ...(options.headers || {}),
      Origin: baseURL,
    },
  });
};
%s
setTimeout(() => {
  for (const key of Object.keys(elements).sort()) {
    console.log(key + "=" + elements[key].textContent);
  }
}, 200);
`, baseURL, failStatus, script)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, nodePath, "-e", harness)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("run frontend script: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("frontend script timed out\nstdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
	}
	return stdout.String()
}

type runningDashboard struct {
	cmd        *exec.Cmd
	stdout     *bytes.Buffer
	stderr     *bytes.Buffer
	stderrLine chan string
	done       chan error
}

func startDashboard(t *testing.T, binary string, env []string, args ...string) *runningDashboard {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("start dashboard: %v", err)
	}

	server := &runningDashboard{
		cmd:        cmd,
		stdout:     &stdout,
		stderr:     &stderr,
		stderrLine: make(chan string, 32),
		done:       make(chan error, 1),
	}
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			stderr.WriteString(line)
			stderr.WriteByte('\n')
			server.stderrLine <- line
		}
	}()
	go func() {
		server.done <- cmd.Wait()
	}()
	return server
}

func (server *runningDashboard) waitForStderrLine(t *testing.T, pattern *regexp.Regexp) string {
	t.Helper()

	timeout := time.After(5 * time.Second)
	for {
		select {
		case line := <-server.stderrLine:
			if pattern.MatchString(line) {
				return line
			}
		case err := <-server.done:
			result := server.resultFromErr(err)
			t.Fatalf("dashboard exited before stderr matched %s: exit=%d stdout=%q stderr=%q", pattern.String(), result.code, result.stdout, result.stderr)
		case <-timeout:
			t.Fatalf("timed out waiting for stderr matching %s; stderr so far: %q", pattern.String(), server.stderr.String())
		}
	}
}

func (server *runningDashboard) signal(t *testing.T, signal os.Signal) {
	t.Helper()

	if err := server.cmd.Process.Signal(signal); err != nil {
		t.Fatalf("signal dashboard: %v", err)
	}
}

func (server *runningDashboard) wait(t *testing.T) commandResult {
	t.Helper()

	select {
	case err := <-server.done:
		return server.resultFromErr(err)
	case <-time.After(5 * time.Second):
		_ = server.cmd.Process.Kill()
		t.Fatal("timed out waiting for dashboard exit")
		return commandResult{}
	}
}

func (server *runningDashboard) cleanup(t *testing.T) {
	t.Helper()

	select {
	case <-server.done:
		return
	default:
		_ = server.cmd.Process.Signal(syscall.SIGTERM)
		select {
		case <-server.done:
		case <-time.After(2 * time.Second):
			_ = server.cmd.Process.Kill()
		}
	}
}

func (server *runningDashboard) resultFromErr(err error) commandResult {
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			code = -1
		}
	}
	return commandResult{stdout: server.stdout.String(), stderr: server.stderr.String(), code: code}
}

func httpGet(t *testing.T, url string) (int, string, []byte) {
	t.Helper()

	client := http.Client{Timeout: 5 * time.Second}
	response, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read %s response: %v", url, err)
	}
	return response.StatusCode, response.Header.Get("Content-Type"), body
}

type httpRequestSpec struct {
	method      string
	url         string
	host        string
	headers     map[string]string
	requestURI  string
	rawHTTPHost string
}

type httpResponse struct {
	statusCode  int
	contentType string
	body        []byte
}

func requestJSON(t *testing.T, spec httpRequestSpec) httpResponse {
	t.Helper()

	if spec.requestURI != "" {
		return rawHTTPRequest(t, spec)
	}

	client := http.Client{Timeout: 5 * time.Second}
	request, err := http.NewRequest(spec.method, spec.url, nil)
	if err != nil {
		t.Fatalf("new request %s %s: %v", spec.method, spec.url, err)
	}
	if spec.host != "" {
		request.Host = spec.host
	}
	for key, value := range spec.headers {
		request.Header.Set(key, value)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%s %s: %v", spec.method, spec.url, err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read %s response: %v", spec.url, err)
	}
	return httpResponse{statusCode: response.StatusCode, contentType: response.Header.Get("Content-Type"), body: body}
}

func rawHTTPRequest(t *testing.T, spec httpRequestSpec) httpResponse {
	t.Helper()

	address := spec.rawHTTPHost
	if address == "" {
		address = strings.TrimPrefix(spec.url, "http://")
		if slash := strings.IndexByte(address, '/'); slash >= 0 {
			address = address[:slash]
		}
	}
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		t.Fatalf("dial %s: %v", address, err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	host := spec.rawHTTPHost
	if spec.host != "" {
		host = spec.host
	}
	if host == "" {
		host = address
	}
	var request bytes.Buffer
	fmt.Fprintf(&request, "%s %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n", spec.method, spec.requestURI, host)
	for key, value := range spec.headers {
		fmt.Fprintf(&request, "%s: %s\r\n", key, value)
	}
	request.WriteString("\r\n")
	if _, err := conn.Write(request.Bytes()); err != nil {
		t.Fatalf("write raw HTTP request: %v", err)
	}
	response, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatalf("read raw HTTP response: %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read raw HTTP body: %v", err)
	}
	return httpResponse{statusCode: response.StatusCode, contentType: response.Header.Get("Content-Type"), body: body}
}

func isolatedEnv(t *testing.T) []string {
	t.Helper()

	root := t.TempDir()
	binDir := filepath.Join(root, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	adbMarker := filepath.Join(root, "adb-invoked")
	adbPath := filepath.Join(binDir, "adb")
	if err := os.WriteFile(adbPath, []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$ADB_MARKER\"\nprintf 'Android Debug Bridge version 1.0.41\\n'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	browserMarker := filepath.Join(root, "browser-opened")
	browserPath := filepath.Join(binDir, "xdg-open")
	if err := os.WriteFile(browserPath, []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$BROWSER_MARKER\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	gocache := filepath.Join(root, "gocache")
	gomodcache := filepath.Join(root, "gomodcache")
	tmp := filepath.Join(root, "tmp")
	for _, dir := range []string{gocache, gomodcache, tmp} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	env := []string{
		"PATH=" + binDir + string(os.PathListSeparator) + os.Getenv("PATH"),
		"HOME=" + filepath.Join(root, "home"),
		"XDG_CONFIG_HOME=" + filepath.Join(root, "config"),
		"XDG_STATE_HOME=" + filepath.Join(root, "state"),
		"XDG_RUNTIME_DIR=" + filepath.Join(root, "runtime"),
		"TMPDIR=" + tmp,
		"GOCACHE=" + gocache,
		"GOMODCACHE=" + gomodcache,
		"ADB_MARKER=" + adbMarker,
		"BROWSER_MARKER=" + browserMarker,
	}
	return append(env, "PWD="+mustGetwd(t))
}

func assertNoForbiddenSideEffects(t *testing.T, env []string) {
	t.Helper()

	values := envMap(env)
	for _, path := range []string{
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard"),
		filepath.Join(values["XDG_RUNTIME_DIR"], "adb-dashboard"),
		filepath.Join(values["TMPDIR"], "adb-dashboard-"+os.Getenv("UID")),
		values["ADB_MARKER"],
		values["BROWSER_MARKER"],
	} {
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("forbidden side effect exists: %s", path)
		} else if !os.IsNotExist(err) {
			t.Fatalf("cannot inspect side effect path %s: %v", path, err)
		}
	}
}

func assertNoFutureStateOrExternalSideEffects(t *testing.T, env []string) {
	t.Helper()

	values := envMap(env)
	for _, path := range []string{
		values["ADB_MARKER"],
		values["BROWSER_MARKER"],
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "cache"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "projects"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "uploads"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "history"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "jobs"),
	} {
		assertPathAbsent(t, path)
	}
}

func assertNoFutureStateOrExternalSideEffectsAllowADB(t *testing.T, env []string) {
	t.Helper()

	values := envMap(env)
	assertFileContains(t, values["ADB_MARKER"], "version\n")
	for _, path := range []string{
		values["BROWSER_MARKER"],
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "cache"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "projects"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "uploads"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "history"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "jobs"),
	} {
		assertPathAbsent(t, path)
	}
}

func assertNoFutureStateSideEffectsAllowBrowser(t *testing.T, env []string) {
	t.Helper()

	values := envMap(env)
	for _, path := range []string{
		values["ADB_MARKER"],
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "cache"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "projects"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "uploads"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "history"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "jobs"),
	} {
		assertPathAbsent(t, path)
	}
}

func assertNoFutureStateSideEffectsAllowBrowserAndADB(t *testing.T, env []string) {
	t.Helper()

	values := envMap(env)
	assertFileContains(t, values["ADB_MARKER"], "version\n")
	for _, path := range []string{
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "cache"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "projects"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "uploads"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "history"),
		filepath.Join(values["XDG_STATE_HOME"], "adb-dashboard", "jobs"),
	} {
		assertPathAbsent(t, path)
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s exists but is not a directory", path)
	}
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(content) != want {
		t.Fatalf("%s = %q, want %q", path, string(content), want)
	}
}

func waitFileContains(t *testing.T, path, want string) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		content, err := os.ReadFile(path)
		if err == nil && string(content) == want {
			return
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("content = %q", string(content))
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("%s did not become %q: %v", path, want, lastErr)
}

func writeFakeADB(t *testing.T, env []string, script string) string {
	t.Helper()

	adbPath := fakeADBPath(t, env)
	if err := os.WriteFile(adbPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return adbPath
}

func removeFakeADB(t *testing.T, env []string) []string {
	t.Helper()

	adbPath := fakeADBPath(t, env)
	if err := os.Remove(adbPath); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return setEnv(env, "PATH", filepath.Dir(adbPath))
}

func fakeADBPath(t *testing.T, env []string) string {
	t.Helper()

	pathValue := envMap(env)["PATH"]
	first, _, _ := strings.Cut(pathValue, string(os.PathListSeparator))
	if first == "" {
		t.Fatal("isolated PATH has no first directory")
	}
	return filepath.Join(first, "adb")
}

func assertPathAbsent(t *testing.T, path string) {
	t.Helper()

	if path == "" {
		return
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("forbidden side effect exists: %s", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("cannot inspect side effect path %s: %v", path, err)
	}
}

func assertM2S2StatusJSON(t *testing.T, body []byte, bind string, readOnly bool, wantADB map[string]any) {
	t.Helper()

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal status JSON: %v\nbody: %s", err, body)
	}
	assertKeys(t, got, "application", "server", "adb", "watcher", "jobs", "sessions", "storage", "hostTools")
	assertObject(t, got["application"], map[string]any{
		"name":             "adb-dashboard",
		"version":          nonEmptyString{},
		"commit":           nonEmptyString{},
		"buildDate":        nonEmptyString{},
		"goVersion":        nonEmptyString{},
		"frontendRevision": nonEmptyString{},
	})
	assertObject(t, got["server"], map[string]any{
		"status":        "running",
		"uptimeSeconds": nonNegativeNumber{},
		"readOnly":      readOnly,
		"bind":          bind,
	})
	assertObject(t, got["adb"], wantADB)
	assertObject(t, got["watcher"], map[string]any{
		"status":             "NIY",
		"lastSuccessfulPoll": nil,
	})
	assertObject(t, got["jobs"], map[string]any{
		"status":   "NIY",
		"active":   float64(0),
		"retained": float64(0),
	})
	assertObject(t, got["sessions"], map[string]any{
		"status": "NIY",
		"active": float64(0),
	})
	assertObject(t, got["storage"], map[string]any{
		"status": "NIY",
	})
	assertObject(t, got["hostTools"], map[string]any{
		"status":      "NIY",
		"available":   float64(0),
		"unavailable": float64(0),
	})
	forbidden := []string{"token", "HOME=", "ADB_DASHBOARD"}
	for _, text := range forbidden {
		if strings.Contains(string(body), text) {
			t.Fatalf("status JSON contains forbidden text %q: %s", text, body)
		}
	}
}

type bootstrapTokens struct {
	csrfToken      string
	webSocketToken string
}

func assertBootstrapJSON(t *testing.T, body []byte) bootstrapTokens {
	t.Helper()

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal bootstrap JSON: %v\nbody: %s", err, body)
	}
	assertKeys(t, got, "csrfToken", "webSocketToken", "statusUrl")
	csrfToken := assertTokenField(t, got, "csrfToken")
	webSocketToken := assertTokenField(t, got, "webSocketToken")
	if csrfToken == webSocketToken {
		t.Fatalf("csrfToken and webSocketToken must be independent; both were %q", csrfToken)
	}
	if got["statusUrl"] != "/api/v1/status" {
		t.Fatalf("statusUrl = %#v, want /api/v1/status", got["statusUrl"])
	}
	return bootstrapTokens{csrfToken: csrfToken, webSocketToken: webSocketToken}
}

func assertTokenField(t *testing.T, got map[string]any, key string) string {
	t.Helper()

	value, ok := got[key].(string)
	if !ok {
		t.Fatalf("%s = %#v, want string", key, got[key])
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_-]{32,}$`).MatchString(value) {
		t.Fatalf("%s = %q, want at least 32 URL-safe base64 characters", key, value)
	}
	return value
}

func assertSecurityErrorEnvelope(t *testing.T, body []byte, code string) {
	t.Helper()

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal security error JSON: %v\nbody: %s", err, body)
	}
	assertObject(t, got, map[string]any{
		"error": map[string]any{
			"code":      code,
			"message":   "Request rejected by dashboard browser security policy",
			"details":   map[string]any{},
			"requestId": nil,
		},
	})
}

func assertNoTokenDisclosure(t *testing.T, body []byte, tokens bootstrapTokens) {
	t.Helper()

	text := string(body)
	for _, forbidden := range []string{"csrfToken", "webSocketToken", tokens.csrfToken, tokens.webSocketToken} {
		if forbidden != "" && strings.Contains(text, forbidden) {
			t.Fatalf("response disclosed %q: %s", forbidden, body)
		}
	}
}

func serverAddressFromStartLine(t *testing.T, line string) string {
	t.Helper()

	return strings.TrimPrefix(line[strings.LastIndex(line, "addr="):], "addr=")
}

type nonEmptyString struct{}

type nonNegativeNumber struct{}

func assertJSONEqual(t *testing.T, body []byte, want map[string]any) {
	t.Helper()

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal JSON: %v\nbody: %s", err, body)
	}
	assertObject(t, got, want)
}

func assertObject(t *testing.T, got any, want map[string]any) {
	t.Helper()

	gotObject, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("value = %#v, want object", got)
	}
	keys := make([]string, 0, len(want))
	for key := range want {
		keys = append(keys, key)
	}
	assertKeys(t, gotObject, keys...)
	for key, wantValue := range want {
		gotValue := gotObject[key]
		switch wantValue.(type) {
		case nonEmptyString:
			text, ok := gotValue.(string)
			if !ok || text == "" {
				t.Fatalf("%s = %#v, want non-empty string", key, gotValue)
			}
		case nonNegativeNumber:
			number, ok := gotValue.(float64)
			if !ok || number < 0 || number != float64(int64(number)) {
				t.Fatalf("%s = %#v, want non-negative integer", key, gotValue)
			}
		default:
			if fmt.Sprintf("%#v", gotValue) != fmt.Sprintf("%#v", wantValue) {
				t.Fatalf("%s = %#v, want %#v", key, gotValue, wantValue)
			}
		}
	}
}

func assertKeys(t *testing.T, got map[string]any, keys ...string) {
	t.Helper()

	if len(got) != len(keys) {
		t.Fatalf("object keys = %v, want exactly %v", mapKeys(got), keys)
	}
	for _, key := range keys {
		if _, ok := got[key]; !ok {
			t.Fatalf("object keys = %v, missing %s", mapKeys(got), key)
		}
	}
}

func mapKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func assertVersionOutput(t *testing.T, stdout string) {
	t.Helper()

	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	if len(lines) != 5 {
		t.Fatalf("version line count = %d, want 5; stdout = %q", len(lines), stdout)
	}
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^adb-dashboard \S+$`),
		regexp.MustCompile(`^commit: \S+$`),
		regexp.MustCompile(`^buildDate: \S+$`),
		regexp.MustCompile(`^goVersion: \S+$`),
		regexp.MustCompile(`^frontendRevision: \S+$`),
	}
	for i, pattern := range patterns {
		if !pattern.MatchString(lines[i]) {
			t.Fatalf("version line %d = %q, does not match %s", i+1, lines[i], pattern.String())
		}
	}
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	next := env[:0]
	for _, item := range env {
		if !strings.HasPrefix(item, prefix) {
			next = append(next, item)
		}
	}
	return append(next, prefix+value)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func envMap(env []string) map[string]string {
	values := make(map[string]string)
	for _, item := range env {
		key, value, ok := strings.Cut(item, "=")
		if ok {
			values[key] = value
		}
	}
	return values
}

func mustGetwd(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return wd
}

func repoRoot(t *testing.T) string {
	t.Helper()

	wd := mustGetwd(t)
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("could not find repository root containing go.mod")
		}
		wd = parent
	}
}
