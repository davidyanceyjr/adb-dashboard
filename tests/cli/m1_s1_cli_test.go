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
	} {
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("forbidden side effect exists: %s", path)
		} else if !os.IsNotExist(err) {
			t.Fatalf("cannot inspect side effect path %s: %v", path, err)
		}
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
