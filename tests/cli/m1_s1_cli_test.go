package cli_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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
		"adbExecutable: NIY adb.discovery is not implemented yet",
		"adbVersion: NIY adb.discovery is not implemented yet",
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
	assertNoFutureStateOrExternalSideEffects(t, env)
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
	assertNoFutureStateOrExternalSideEffects(t, env)
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
		"hostTools: NIY hosttools.discovery is not implemented yet\n",
	}
	for _, want := range wantContains {
		if !strings.Contains(result.stdout, want) {
			t.Fatalf("doctor stdout missing %q\nstdout:\n%s", want, result.stdout)
		}
	}
	assertDirExists(t, tempDir)
	assertNoFutureStateOrExternalSideEffects(t, env)
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

func isolatedEnv(t *testing.T) []string {
	t.Helper()

	root := t.TempDir()
	binDir := filepath.Join(root, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	adbMarker := filepath.Join(root, "adb-invoked")
	adbPath := filepath.Join(binDir, "adb")
	if err := os.WriteFile(adbPath, []byte("#!/bin/sh\necho invoked > "+adbMarker+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	browserMarker := filepath.Join(root, "browser-opened")
	browserPath := filepath.Join(binDir, "xdg-open")
	if err := os.WriteFile(browserPath, []byte("#!/bin/sh\necho opened > "+browserMarker+"\n"), 0o755); err != nil {
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
