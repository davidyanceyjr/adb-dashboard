package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/BurntSushi/toml"
)

var (
	version          = "dev"
	commit           = "dev"
	buildDate        = "dev"
	frontendRevision = "dev"
)

const helpText = `Usage:
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

const maxPackageOutputBytes = 1 << 20
const maxPackageSummaryLines = 50
const maxArtifactSizeBytes = 256 << 20
const maxArtifactMultipartBytes = maxArtifactSizeBytes + (1 << 20)

var errPackageNotFound = errors.New("package not found")

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr *os.File) int {
	if len(args) == 0 {
		return runStartup(nil, stderr)
	}

	first := args[0]
	switch first {
	case "--help":
		fmt.Fprint(stdout, helpText)
		return 0
	case "--version", "version":
		printVersion(stdout)
		return 0
	case "serve":
		return runStartup(args[1:], stderr)
	case "doctor":
		return runDoctor(args[1:], stdout, stderr)
	default:
		if strings.HasPrefix(first, "--") {
			return validateGlobalOption(args, stderr)
		}
		fmt.Fprintf(stderr, "unknown command: %s\n", first)
		return 2
	}
}

type config struct {
	listen      sourcedString
	readOnly    sourcedBool
	dataDir     sourcedString
	tempDir     sourcedString
	logLevel    sourcedString
	openBrowser sourcedBool
}

type sourcedString struct {
	value  string
	source string
}

type sourcedBool struct {
	value  bool
	source string
}

func defaultConfig() config {
	return config{
		listen:      sourcedString{value: "127.0.0.1:8080", source: "defaults"},
		readOnly:    sourcedBool{value: false, source: "defaults"},
		dataDir:     sourcedString{value: defaultDataDir(), source: "defaults"},
		tempDir:     sourcedString{value: defaultTempDir(), source: "defaults"},
		logLevel:    sourcedString{value: "info", source: "defaults"},
		openBrowser: sourcedBool{value: false, source: "defaults"},
	}
}

type cliOptions struct {
	configPath string
	listen     *string
	dataDir    *string
	tempDir    *string
	logLevel   *string
	readOnly   bool
	open       *bool
}

func runDoctor(args []string, stdout, stderr *os.File) int {
	options, code := parseOptions(args, stderr)
	if code != 0 {
		return code
	}

	cfg, err := resolveConfig(options)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 2
	}

	dataErr := ensureDir(cfg.dataDir.value)
	tempErr := ensureDir(cfg.tempDir.value)
	adb := discoverADBVersion()
	overall := "PASS"
	exitCode := 0
	if dataErr != nil || tempErr != nil {
		overall = "FAIL"
		exitCode = 5
	}
	if adb.hasFailure() {
		overall = "FAIL"
		if exitCode == 0 {
			exitCode = 3
		}
	}

	fmt.Fprintln(stdout, "adb-dashboard doctor")
	fmt.Fprintf(stdout, "overall: %s\n", overall)
	fmt.Fprintf(stdout, "config: PASS source=%s listen=%s readOnly=%t logLevel=%s\n", reportSource(cfg), cfg.listen.value, cfg.readOnly.value, cfg.logLevel.value)
	printDirRow(stdout, "dataDir", cfg.dataDir.value, dataErr)
	printDirRow(stdout, "tempDir", cfg.tempDir.value, tempErr)
	fmt.Fprintln(stdout, "cacheDir: NIY storage.cache is not implemented yet")
	fmt.Fprintln(stdout, "projectDir: NIY storage.projects is not implemented yet")
	printADBDoctorRows(stdout, adb)
	fmt.Fprintln(stdout, "adbServer: NIY adb.server is not implemented yet")
	fmt.Fprintln(stdout, "devices: NIY devices.refresh is not implemented yet")
	fmt.Fprintln(stdout, "hostTools: NIY hosttools.discovery is not implemented yet")
	return exitCode
}

func runStartup(args []string, stderr *os.File) int {
	options, code := parseOptions(args, stderr)
	if code != 0 {
		return code
	}

	cfg, err := resolveConfig(options)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 2
	}
	if err := validateLoopbackListen(cfg.listen.value); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 2
	}
	if failure := ensureStartupDirs(cfg); failure != nil {
		fmt.Fprintf(stderr, "server runtime failure: startup filesystem unavailable: %s directory %s: %s\n", failure.kind, failure.path, failure.err)
		return 5
	}

	return serveDashboard(cfg, stderr)
}

func parseOptions(args []string, stderr *os.File) (cliOptions, int) {
	var options cliOptions
	for i := 0; i < len(args); i++ {
		option := args[i]
		switch option {
		case "--listen", "--data-dir", "--config", "--temp-dir", "--log-level":
			if i+1 >= len(args) {
				fmt.Fprintf(stderr, "missing argument for %s\n", option)
				return options, 2
			}
			value := args[i+1]
			switch option {
			case "--listen":
				options.listen = &value
			case "--data-dir":
				options.dataDir = &value
			case "--config":
				options.configPath = value
			case "--temp-dir":
				options.tempDir = &value
			case "--log-level":
				options.logLevel = &value
			}
			i++
		case "--open":
			value := true
			options.open = &value
		case "--no-open":
			value := false
			options.open = &value
		case "--read-only":
			options.readOnly = true
		default:
			if strings.HasPrefix(option, "--") {
				fmt.Fprintf(stderr, "unknown option: %s\n", option)
				return options, 2
			}
			fmt.Fprintf(stderr, "unknown command: %s\n", option)
			return options, 2
		}
	}
	return options, 0
}

func resolveConfig(options cliOptions) (config, error) {
	cfg := defaultConfig()
	for _, path := range defaultConfigFiles() {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		if err := applyConfigFile(&cfg, path); err != nil {
			return cfg, err
		}
	}
	if path := os.Getenv("ADB_DASHBOARD_CONFIG"); path != "" {
		if err := applyConfigFile(&cfg, path); err != nil {
			return cfg, err
		}
	}
	if options.configPath != "" {
		if err := applyConfigFile(&cfg, options.configPath); err != nil {
			return cfg, err
		}
	}
	applyEnv(&cfg)
	applyCLI(&cfg, options)
	if err := finalizeConfig(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func defaultConfigFiles() []string {
	paths := []string{"/etc/adb-dashboard/config.toml"}
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(os.Getenv("HOME"), ".config")
	}
	if configHome != "" {
		paths = append(paths, filepath.Join(configHome, "adb-dashboard", "config.toml"))
	}
	return paths
}

func applyConfigFile(cfg *config, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("invalid configuration: cannot load %s: %v", path, err)
	}

	var parsed map[string]interface{}
	if _, err := toml.Decode(string(content), &parsed); err != nil {
		return fmt.Errorf("invalid configuration: cannot parse %s: %v", path, err)
	}

	source := "file:" + path
	for section, sectionValue := range parsed {
		values, ok := sectionValue.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid configuration: %s must be table", section)
		}
		switch section {
		case "server":
			if err := applyServerConfig(cfg, values, source); err != nil {
				return err
			}
		case "logging":
			if err := applyLoggingConfig(cfg, values, source); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid configuration: unknown key %s", section)
		}
	}
	return nil
}

func applyServerConfig(cfg *config, values map[string]interface{}, source string) error {
	for key, value := range values {
		switch key {
		case "listen":
			text, ok := value.(string)
			if !ok {
				return fmt.Errorf("invalid configuration: server.listen must be string")
			}
			cfg.listen = sourcedString{value: text, source: source}
		case "open_browser":
			flag, ok := value.(bool)
			if !ok {
				return fmt.Errorf("invalid configuration: server.open_browser must be bool")
			}
			cfg.openBrowser = sourcedBool{value: flag, source: source}
		case "read_only":
			flag, ok := value.(bool)
			if !ok {
				return fmt.Errorf("invalid configuration: server.read_only must be bool")
			}
			cfg.readOnly = sourcedBool{value: flag, source: source}
		case "data_dir":
			text, ok := value.(string)
			if !ok {
				return fmt.Errorf("invalid configuration: server.data_dir must be string")
			}
			cfg.dataDir = sourcedString{value: text, source: source}
		case "temp_dir":
			text, ok := value.(string)
			if !ok {
				return fmt.Errorf("invalid configuration: server.temp_dir must be string")
			}
			cfg.tempDir = sourcedString{value: text, source: source}
		default:
			return fmt.Errorf("invalid configuration: unknown key server.%s", key)
		}
	}
	return nil
}

func applyLoggingConfig(cfg *config, values map[string]interface{}, source string) error {
	for key, value := range values {
		switch key {
		case "level":
			text, ok := value.(string)
			if !ok {
				return fmt.Errorf("invalid configuration: logging.level must be string")
			}
			cfg.logLevel = sourcedString{value: text, source: source}
		default:
			return fmt.Errorf("invalid configuration: unknown key logging.%s", key)
		}
	}
	return nil
}

func applyEnv(cfg *config) {
	if value := os.Getenv("ADB_DASHBOARD_LISTEN"); value != "" {
		cfg.listen = sourcedString{value: value, source: "env"}
	}
	if value := os.Getenv("ADB_DASHBOARD_DATA_DIR"); value != "" {
		cfg.dataDir = sourcedString{value: value, source: "env"}
	}
	if value := os.Getenv("ADB_DASHBOARD_TEMP_DIR"); value != "" {
		cfg.tempDir = sourcedString{value: value, source: "env"}
	}
	if value := os.Getenv("ADB_DASHBOARD_LOG_LEVEL"); value != "" {
		cfg.logLevel = sourcedString{value: value, source: "env"}
	}
}

func applyCLI(cfg *config, options cliOptions) {
	if options.listen != nil {
		cfg.listen = sourcedString{value: *options.listen, source: "cli"}
	}
	if options.dataDir != nil {
		cfg.dataDir = sourcedString{value: *options.dataDir, source: "cli"}
	}
	if options.tempDir != nil {
		cfg.tempDir = sourcedString{value: *options.tempDir, source: "cli"}
	}
	if options.logLevel != nil {
		cfg.logLevel = sourcedString{value: *options.logLevel, source: "cli"}
	}
	if options.readOnly {
		cfg.readOnly = sourcedBool{value: true, source: "cli"}
	}
	if options.open != nil {
		cfg.openBrowser = sourcedBool{value: *options.open, source: "cli"}
	}
}

func finalizeConfig(cfg *config) error {
	if err := validateListen(cfg.listen.value); err != nil {
		return err
	}
	if err := validateLogLevel(cfg.logLevel.value); err != nil {
		return err
	}

	var err error
	cfg.dataDir.value, err = expandPath("server.data_dir", cfg.dataDir.value)
	if err != nil {
		return err
	}
	cfg.tempDir.value, err = expandPath("server.temp_dir", cfg.tempDir.value)
	if err != nil {
		return err
	}
	return nil
}

func validateListen(value string) error {
	host, portText, err := net.SplitHostPort(value)
	if err != nil {
		return fmt.Errorf("invalid configuration: server.listen must be HOST:PORT")
	}
	if host == "" {
		return fmt.Errorf("invalid configuration: server.listen host must not be empty")
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 0 || port > 65535 {
		return fmt.Errorf("invalid configuration: server.listen port must be 0 through 65535: %s", portText)
	}
	return nil
}

func validateLoopbackListen(value string) error {
	host, _, err := net.SplitHostPort(value)
	if err != nil {
		return fmt.Errorf("invalid configuration: server.listen must be HOST:PORT")
	}
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return nil
	}
	return fmt.Errorf("invalid configuration: server.listen must use a loopback host: %s", host)
}

func validateLogLevel(value string) error {
	switch value {
	case "error", "warn", "info", "debug", "trace":
		return nil
	default:
		return fmt.Errorf("invalid configuration: logging.level must be one of error, warn, info, debug, trace: %s", value)
	}
}

func expandPath(key, value string) (string, error) {
	expanded := value
	if strings.HasPrefix(expanded, "~/") {
		home := os.Getenv("HOME")
		if home == "" {
			return "", fmt.Errorf("invalid configuration: %s references unset environment variable: HOME", key)
		}
		expanded = filepath.Join(home, strings.TrimPrefix(expanded, "~/"))
	}
	if strings.Contains(expanded, "~") {
		return "", fmt.Errorf("invalid configuration: %s contains unsupported path expansion: %s", key, value)
	}
	var err error
	expanded, err = expandEnvPath(key, expanded, value)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(expanded) {
		expanded, err = filepath.Abs(expanded)
		if err != nil {
			return "", err
		}
	}
	return filepath.Clean(expanded), nil
}

func expandEnvPath(key, value, original string) (string, error) {
	var builder strings.Builder
	for i := 0; i < len(value); {
		if value[i] != '$' {
			builder.WriteByte(value[i])
			i++
			continue
		}
		nameStart := i + 1
		nameEnd := nameStart
		if nameStart < len(value) && value[nameStart] == '{' {
			nameStart++
			nameEnd = strings.IndexByte(value[nameStart:], '}')
			if nameEnd < 0 {
				return "", fmt.Errorf("invalid configuration: %s contains unsupported path expansion: %s", key, original)
			}
			nameEnd += nameStart
			name := value[nameStart:nameEnd]
			replacement, err := envReplacement(key, name)
			if err != nil {
				return "", err
			}
			builder.WriteString(replacement)
			i = nameEnd + 1
			continue
		}
		for nameEnd < len(value) && isEnvNameChar(value[nameEnd], nameEnd == nameStart) {
			nameEnd++
		}
		if nameEnd == nameStart {
			return "", fmt.Errorf("invalid configuration: %s contains unsupported path expansion: %s", key, original)
		}
		name := value[nameStart:nameEnd]
		replacement, err := envReplacement(key, name)
		if err != nil {
			return "", err
		}
		builder.WriteString(replacement)
		i = nameEnd
	}
	return builder.String(), nil
}

func isEnvNameChar(ch byte, first bool) bool {
	if ch == '_' || ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' {
		return true
	}
	return !first && ch >= '0' && ch <= '9'
}

func envReplacement(key, name string) (string, error) {
	if name == "" || !isEnvNameChar(name[0], true) {
		return "", fmt.Errorf("invalid configuration: %s references unset environment variable: %s", key, name)
	}
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("invalid configuration: %s references unset environment variable: %s", key, name)
	}
	return value, nil
}

func defaultDataDir() string {
	if value := os.Getenv("XDG_STATE_HOME"); value != "" {
		return filepath.Join(value, "adb-dashboard")
	}
	return "$HOME/.local/state/adb-dashboard"
}

func defaultTempDir() string {
	if value := os.Getenv("XDG_RUNTIME_DIR"); value != "" {
		return filepath.Join(value, "adb-dashboard")
	}
	base := os.Getenv("TMPDIR")
	if base == "" {
		base = "/tmp"
	}
	return filepath.Join(base, fmt.Sprintf("adb-dashboard-%d", os.Getuid()))
}

type startupDirFailure struct {
	kind string
	path string
	err  error
}

func ensureStartupDirs(cfg config) *startupDirFailure {
	if err := ensureDir(cfg.dataDir.value); err != nil {
		return &startupDirFailure{kind: "data", path: cfg.dataDir.value, err: err}
	}
	if err := ensureDir(cfg.tempDir.value); err != nil {
		return &startupDirFailure{kind: "temp", path: cfg.tempDir.value, err: err}
	}
	return nil
}

func ensureDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory")
	}
	return nil
}

func printDirRow(stdout *os.File, name, path string, err error) {
	if err == nil {
		fmt.Fprintf(stdout, "%s: PASS path=%s\n", name, path)
		return
	}
	fmt.Fprintf(stdout, "%s: FAIL path=%s error=%s\n", name, path, err.Error())
}

type adbDiscovery struct {
	path       string
	version    string
	execErr    string
	versionErr string
}

func (adb adbDiscovery) hasFailure() bool {
	return adb.execErr != "" || adb.versionErr != ""
}

func discoverADBVersion() adbDiscovery {
	path, err := exec.LookPath("adb")
	if err != nil {
		return adbDiscovery{execErr: "not found in PATH"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "version")
	output, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return adbDiscovery{path: path, versionErr: "timed out"}
	}
	if err != nil {
		return adbDiscovery{path: path, versionErr: err.Error()}
	}
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return adbDiscovery{path: path, version: line}
		}
	}
	return adbDiscovery{path: path, versionErr: "no version output"}
}

func printADBDoctorRows(stdout *os.File, adb adbDiscovery) {
	if adb.execErr != "" {
		fmt.Fprintf(stdout, "adbExecutable: FAIL error=%s\n", adb.execErr)
		fmt.Fprintln(stdout, "adbVersion: NIY adb.version unavailable until adb executable is found")
		return
	}
	fmt.Fprintf(stdout, "adbExecutable: PASS path=%s\n", adb.path)
	if adb.versionErr != "" {
		fmt.Fprintf(stdout, "adbVersion: FAIL error=%s\n", adb.versionErr)
		return
	}
	fmt.Fprintf(stdout, "adbVersion: PASS version=%s\n", adb.version)
}

func (adb adbDiscovery) status() adbStatus {
	if adb.execErr != "" {
		return adbStatus{
			Status:           "unavailable",
			Executable:       nil,
			Version:          nil,
			ServerResponsive: "NIY",
		}
	}
	if adb.versionErr != "" {
		return adbStatus{
			Status:           "error",
			Executable:       &adb.path,
			Version:          nil,
			ServerResponsive: "NIY",
		}
	}
	return adbStatus{
		Status:           "available",
		Executable:       &adb.path,
		Version:          &adb.version,
		ServerResponsive: "NIY",
	}
}

func discoverADBDevices(adbPath string) ([]deviceInventory, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, adbPath, "devices", "-l")
	output, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("timed out")
	}
	if err != nil {
		return nil, err
	}
	devices, err := parseADBDevicesOutput(string(output))
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func discoverADBLogcat(adbPath, serial string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, adbPath, "-s", serial, "logcat", "-d")
	output, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("timed out")
	}
	if err != nil {
		return "", err
	}
	if !utf8.Valid(output) {
		return "", fmt.Errorf("invalid utf-8")
	}
	return string(output), nil
}

func discoverADBScreenshot(adbPath, serial string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, adbPath, "-s", serial, "exec-out", "screencap", "-p")
	output, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("timed out")
	}
	if err != nil {
		return nil, err
	}
	if !isPNG(output) {
		return nil, fmt.Errorf("invalid png")
	}
	return output, nil
}

func discoverADBPackages(adbPath, serial, scope string) ([]packageItem, error) {
	args := []string{"-s", serial, "shell", "pm", "list", "packages"}
	switch scope {
	case "third-party":
		args = append(args, "-3")
	case "system":
		args = append(args, "-s")
	}
	args = append(args, "-f", "-U", "--show-versioncode")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, adbPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	output, readErr := io.ReadAll(io.LimitReader(stdout, maxPackageOutputBytes+1))
	if len(output) > maxPackageOutputBytes {
		cancel()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, fmt.Errorf("output too large")
	}
	waitErr := cmd.Wait()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("timed out")
	}
	if readErr != nil {
		return nil, readErr
	}
	if waitErr != nil {
		return nil, waitErr
	}
	if !utf8.Valid(output) {
		return nil, fmt.Errorf("invalid utf-8")
	}
	return parseADBPackagesOutput(string(output))
}

func discoverADBPackageDetail(adbPath, serial, packageName string) (packageDetailPayload, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, adbPath, "-s", serial, "shell", "dumpsys", "package", packageName)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return packageDetailPayload{}, err
	}
	if err := cmd.Start(); err != nil {
		return packageDetailPayload{}, err
	}
	output, readErr := io.ReadAll(io.LimitReader(stdout, maxPackageOutputBytes+1))
	if len(output) > maxPackageOutputBytes {
		cancel()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return packageDetailPayload{}, fmt.Errorf("output too large")
	}
	waitErr := cmd.Wait()
	if ctx.Err() == context.DeadlineExceeded {
		return packageDetailPayload{}, fmt.Errorf("timed out")
	}
	if readErr != nil {
		return packageDetailPayload{}, readErr
	}
	if waitErr != nil {
		return packageDetailPayload{}, waitErr
	}
	if !utf8.Valid(output) {
		return packageDetailPayload{}, fmt.Errorf("invalid utf-8")
	}
	return parseADBPackageDetailOutput(packageName, string(output))
}

func isPNG(output []byte) bool {
	prefix := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	if len(output) < len(prefix) {
		return false
	}
	for index, value := range prefix {
		if output[index] != value {
			return false
		}
	}
	return true
}

func parseADBPackagesOutput(output string) ([]packageItem, error) {
	items := []packageItem{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		item, err := parseADBPackageLine(line)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return items, nil
}

func parseADBPackageLine(line string) (packageItem, error) {
	if !strings.HasPrefix(line, "package:") {
		return packageItem{}, fmt.Errorf("malformed package row")
	}
	fields := strings.Fields(strings.TrimPrefix(line, "package:"))
	if len(fields) == 0 {
		return packageItem{}, fmt.Errorf("malformed package row")
	}
	item := packageItem{}
	if apkPath, name, ok := strings.Cut(fields[0], "="); ok {
		if apkPath == "" || name == "" {
			return packageItem{}, fmt.Errorf("malformed package row")
		}
		item.APKPath = stringPointer(apkPath)
		item.Name = name
	} else {
		item.Name = fields[0]
	}
	if item.Name == "" {
		return packageItem{}, fmt.Errorf("malformed package row")
	}
	for _, field := range fields[1:] {
		key, value, ok := strings.Cut(field, ":")
		if !ok || key == "" || value == "" {
			return packageItem{}, fmt.Errorf("malformed package row")
		}
		switch key {
		case "uid":
			item.UserID = stringPointer(value)
		case "versionCode":
			item.VersionCode = stringPointer(value)
		case "installer":
			item.Installer = stringPointer(value)
		default:
			return packageItem{}, fmt.Errorf("malformed package row")
		}
	}
	return item, nil
}

func parseADBPackageDetailOutput(packageName, output string) (packageDetailPayload, error) {
	lines := boundedSummaryLines(output, maxPackageSummaryLines)
	if len(lines) == 0 || strings.Contains(output, "Unable to find package: "+packageName) {
		return packageDetailPayload{}, errPackageNotFound
	}
	header := "Package [" + packageName + "]"
	foundPackage := false
	for _, line := range lines {
		if strings.HasPrefix(line, "Package [") {
			if strings.HasPrefix(line, header) {
				foundPackage = true
				break
			}
			return packageDetailPayload{}, fmt.Errorf("malformed package detail")
		}
	}
	if !foundPackage {
		return packageDetailPayload{}, fmt.Errorf("malformed package detail")
	}

	detail := packageDetailPayload{
		Name:         packageName,
		SummaryLines: lines,
	}
	collectPermissions := false
	for _, rawLine := range strings.Split(output, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "versionName="):
			detail.VersionName = stringPointer(strings.TrimPrefix(line, "versionName="))
			collectPermissions = false
		case strings.HasPrefix(line, "versionCode="):
			value := strings.TrimPrefix(line, "versionCode=")
			value, _, _ = strings.Cut(value, " ")
			if value != "" {
				detail.VersionCode = stringPointer(value)
			}
			collectPermissions = false
		case strings.HasPrefix(line, "installerPackageName="):
			detail.Installer = stringPointer(strings.TrimPrefix(line, "installerPackageName="))
			collectPermissions = false
		case strings.HasPrefix(line, "firstInstallTime="):
			detail.FirstInstallTime = stringPointer(strings.TrimPrefix(line, "firstInstallTime="))
			collectPermissions = false
		case strings.HasPrefix(line, "lastUpdateTime="):
			detail.LastUpdateTime = stringPointer(strings.TrimPrefix(line, "lastUpdateTime="))
			collectPermissions = false
		case line == "requested permissions:":
			collectPermissions = true
		case collectPermissions:
			if strings.HasSuffix(line, ":") || strings.Contains(line, ":") {
				collectPermissions = false
				continue
			}
			detail.RequestedPermissions = append(detail.RequestedPermissions, line)
		}
	}
	return detail, nil
}

func boundedSummaryLines(output string, limit int) []string {
	lines := []string{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if len(lines) == limit {
			break
		}
	}
	return lines
}

func lastLogcatLines(output string, limit int) ([]string, bool) {
	output = strings.TrimSuffix(output, "\n")
	if output == "" {
		return []string{}, false
	}
	lines := strings.Split(output, "\n")
	truncated := len(lines) > limit
	if truncated {
		lines = lines[len(lines)-limit:]
	}
	return lines, truncated
}

func parseADBDevicesOutput(output string) ([]deviceInventory, error) {
	lines := strings.Split(output, "\n")
	headerSeen := false
	devices := []deviceInventory{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !headerSeen {
			if line != "List of devices attached" {
				return nil, fmt.Errorf("malformed adb devices output")
			}
			headerSeen = true
			continue
		}
		device, err := parseADBDeviceLine(line)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	if !headerSeen {
		return nil, fmt.Errorf("malformed adb devices output")
	}
	return devices, nil
}

func parseADBDeviceLine(line string) (deviceInventory, error) {
	fields := strings.Fields(line)
	if len(fields) < 2 || strings.Contains(fields[0], ":") || strings.Contains(fields[1], ":") {
		return deviceInventory{}, fmt.Errorf("malformed adb device row")
	}
	device := deviceInventory{
		Serial: fields[0],
		State:  fields[1],
	}
	for _, field := range fields[2:] {
		key, value, ok := strings.Cut(field, ":")
		if !ok || key == "" || value == "" {
			return deviceInventory{}, fmt.Errorf("malformed adb device row")
		}
		switch key {
		case "product":
			device.Product = stringPointer(value)
		case "model":
			device.Model = stringPointer(value)
		case "device":
			device.Device = stringPointer(value)
		case "transport_id":
			device.TransportID = stringPointer(value)
		}
	}
	return device, nil
}

func stringPointer(value string) *string {
	return &value
}

func reportSource(cfg config) string {
	sources := []string{
		cfg.listen.source,
		cfg.readOnly.source,
		cfg.dataDir.source,
		cfg.tempDir.source,
		cfg.logLevel.source,
	}
	first := sources[0]
	for _, source := range sources[1:] {
		if source != first {
			return "mixed"
		}
	}
	return first
}

func validateGlobalOption(args []string, stderr *os.File) int {
	return runStartup(args, stderr)
}

func printVersion(stdout *os.File) {
	fmt.Fprintf(stdout, "adb-dashboard %s\n", version)
	fmt.Fprintf(stdout, "commit: %s\n", commit)
	fmt.Fprintf(stdout, "buildDate: %s\n", buildDate)
	fmt.Fprintf(stdout, "goVersion: %s\n", runtime.Version())
	fmt.Fprintf(stdout, "frontendRevision: %s\n", frontendRevision)
}

type statusResponse struct {
	Application applicationStatus `json:"application"`
	Server      serverStatus      `json:"server"`
	ADB         adbStatus         `json:"adb"`
	Watcher     watcherStatus     `json:"watcher"`
	Jobs        jobsStatus        `json:"jobs"`
	Sessions    sessionsStatus    `json:"sessions"`
	Storage     storageStatus     `json:"storage"`
	HostTools   hostToolsStatus   `json:"hostTools"`
}

type applicationStatus struct {
	Name             string `json:"name"`
	Version          string `json:"version"`
	Commit           string `json:"commit"`
	BuildDate        string `json:"buildDate"`
	GoVersion        string `json:"goVersion"`
	FrontendRevision string `json:"frontendRevision"`
}

type serverStatus struct {
	Status        string `json:"status"`
	UptimeSeconds int64  `json:"uptimeSeconds"`
	ReadOnly      bool   `json:"readOnly"`
	Bind          string `json:"bind"`
}

type adbStatus struct {
	Status           string  `json:"status"`
	Executable       *string `json:"executable"`
	Version          *string `json:"version"`
	ServerResponsive string  `json:"serverResponsive"`
}

type devicesResponse struct {
	ADB     devicesADBStatus  `json:"adb"`
	Devices []deviceInventory `json:"devices"`
}

type deviceDetailResponse struct {
	ADB    devicesADBStatus `json:"adb"`
	Device deviceInventory  `json:"device"`
}

type logcatResponse struct {
	Device logcatDevice  `json:"device"`
	Logcat logcatPayload `json:"logcat"`
}

type logcatDevice struct {
	Serial string `json:"serial"`
	State  string `json:"state"`
}

type logcatPayload struct {
	Format    string   `json:"format"`
	Lines     []string `json:"lines"`
	Truncated bool     `json:"truncated"`
}

type packagesResponse struct {
	Device   packageDevice  `json:"device"`
	Packages packagePayload `json:"packages"`
}

type packageDetailResponse struct {
	Device  packageDevice        `json:"device"`
	Package packageDetailPayload `json:"package"`
}

type packageDevice struct {
	Serial string `json:"serial"`
	State  string `json:"state"`
}

type packagePayload struct {
	Scope string        `json:"scope"`
	Items []packageItem `json:"items"`
	Count int           `json:"count"`
}

type packageItem struct {
	Name        string  `json:"name"`
	APKPath     *string `json:"apkPath,omitempty"`
	Installer   *string `json:"installer,omitempty"`
	UserID      *string `json:"userId,omitempty"`
	VersionCode *string `json:"versionCode,omitempty"`
}

type packageDetailPayload struct {
	Name                 string   `json:"name"`
	VersionName          *string  `json:"versionName,omitempty"`
	VersionCode          *string  `json:"versionCode,omitempty"`
	Installer            *string  `json:"installer,omitempty"`
	FirstInstallTime     *string  `json:"firstInstallTime,omitempty"`
	LastUpdateTime       *string  `json:"lastUpdateTime,omitempty"`
	RequestedPermissions []string `json:"requestedPermissions,omitempty"`
	SummaryLines         []string `json:"summaryLines"`
}

type devicesADBStatus struct {
	Status     string `json:"status"`
	Executable string `json:"executable"`
	Version    string `json:"version"`
}

type deviceInventory struct {
	Serial      string  `json:"serial"`
	State       string  `json:"state"`
	Product     *string `json:"product"`
	Model       *string `json:"model"`
	Device      *string `json:"device"`
	TransportID *string `json:"transportId"`
}

type watcherStatus struct {
	Status             string  `json:"status"`
	LastSuccessfulPoll *string `json:"lastSuccessfulPoll"`
}

type jobsStatus struct {
	Status   string `json:"status"`
	Active   int    `json:"active"`
	Retained int    `json:"retained"`
}

type sessionsStatus struct {
	Status string `json:"status"`
	Active int    `json:"active"`
}

type storageStatus struct {
	Status string `json:"status"`
}

type hostToolsStatus struct {
	Status      string `json:"status"`
	Available   int    `json:"available"`
	Unavailable int    `json:"unavailable"`
}

type bootstrapResponse struct {
	CSRFToken      string `json:"csrfToken"`
	WebSocketToken string `json:"webSocketToken"`
	StatusURL      string `json:"statusUrl"`
}

type artifactUploadResponse struct {
	Artifact artifactMetadata `json:"artifact"`
}

type artifactMetadata struct {
	ID             string `json:"id"`
	OriginalName   string `json:"originalName"`
	SizeBytes      int64  `json:"sizeBytes"`
	SHA256         string `json:"sha256"`
	ContentType    string `json:"contentType,omitempty"`
	CreatedAt      string `json:"createdAt"`
	AnalysisStatus string `json:"analysisStatus"`
}

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details"`
	RequestID *string        `json:"requestId"`
}

func serveDashboard(cfg config, stderr *os.File) int {
	listener, err := net.Listen("tcp", cfg.listen.value)
	if err != nil {
		fmt.Fprintf(stderr, "listen address unavailable: %s: %s\n", cfg.listen.value, err)
		return 4
	}

	tokens, err := newBootstrapResponse()
	if err != nil {
		_ = listener.Close()
		fmt.Fprintf(stderr, "server runtime failure: token generation failed: %s\n", err)
		return 5
	}
	startedAt := time.Now()
	actualAddr := listener.Addr().String()
	server := &http.Server{
		Handler: dashboardHandler(startedAt, actualAddr, cfg.readOnly.value, cfg.dataDir.value, tokens),
	}

	serverErr := make(chan error, 1)
	go func() {
		err := server.Serve(listener)
		if err == http.ErrServerClosed {
			err = nil
		}
		serverErr <- err
	}()

	fmt.Fprintf(stderr, "%s INFO server started addr=%s\n", logTimestamp(time.Now()), actualAddr)
	if cfg.openBrowser.value {
		openDashboardBrowser(stderr, "http://"+actualAddr+"/")
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case received := <-signals:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		if err := <-serverErr; err != nil {
			fmt.Fprintf(stderr, "server runtime failure: shutdown failed: %s\n", err)
			return 5
		}
		fmt.Fprintf(stderr, "%s INFO server stopped signal=%s\n", logTimestamp(time.Now()), received)
		return 0
	case err := <-serverErr:
		if err != nil {
			fmt.Fprintf(stderr, "server runtime failure: server stopped unexpectedly: %s\n", err)
			return 5
		}
		return 0
	}
}

func dashboardHandler(startedAt time.Time, bind string, readOnly bool, dataDir string, tokens bootstrapResponse) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" || request.Method != http.MethodGet {
			http.NotFound(writer, request)
			return
		}
		writeHTML(writer, dashboardShellHTML)
	})
	mux.HandleFunc("/api/v1/bootstrap", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writeUnknownRoute(writer)
			return
		}
		writeJSON(writer, http.StatusOK, tokens)
	})
	mux.HandleFunc("/api/v1/status", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writeUnknownRoute(writer)
			return
		}
		adb := discoverADBVersion()
		writeJSON(writer, http.StatusOK, statusResponse{
			Application: applicationStatus{
				Name:             "adb-dashboard",
				Version:          version,
				Commit:           commit,
				BuildDate:        buildDate,
				GoVersion:        runtime.Version(),
				FrontendRevision: frontendRevision,
			},
			Server: serverStatus{
				Status:        "running",
				UptimeSeconds: int64(time.Since(startedAt).Seconds()),
				ReadOnly:      readOnly,
				Bind:          bind,
			},
			ADB: adb.status(),
			Watcher: watcherStatus{
				Status:             "NIY",
				LastSuccessfulPoll: nil,
			},
			Jobs: jobsStatus{
				Status:   "NIY",
				Active:   0,
				Retained: 0,
			},
			Sessions: sessionsStatus{
				Status: "NIY",
				Active: 0,
			},
			Storage: storageStatus{
				Status: "NIY",
			},
			HostTools: hostToolsStatus{
				Status:      "NIY",
				Available:   0,
				Unavailable: 0,
			},
		})
	})
	mux.HandleFunc("/api/v1/devices", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writeUnknownRoute(writer)
			return
		}
		adb := discoverADBVersion()
		if adb.hasFailure() {
			writeAPIError(writer, http.StatusServiceUnavailable, "adb_unavailable", "ADB is unavailable")
			return
		}
		devices, err := discoverADBDevices(adb.path)
		if err != nil {
			writeAPIError(writer, http.StatusBadGateway, "adb_devices_failed", "ADB device inventory failed")
			return
		}
		writeJSON(writer, http.StatusOK, devicesResponse{
			ADB: devicesADBStatus{
				Status:     "available",
				Executable: adb.path,
				Version:    adb.version,
			},
			Devices: devices,
		})
	})
	mux.HandleFunc("/api/v1/devices/", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writeUnknownRoute(writer)
			return
		}
		serialText := strings.TrimPrefix(request.URL.Path, "/api/v1/devices/")
		if strings.HasSuffix(serialText, "/logcat") {
			serialText = strings.TrimSuffix(serialText, "/logcat")
			handleDeviceLogcat(writer, request, serialText)
			return
		}
		if strings.Contains(serialText, "/packages/") {
			parts := strings.SplitN(serialText, "/packages/", 2)
			handleDevicePackageDetail(writer, request, parts[0], parts[1])
			return
		}
		if strings.HasSuffix(serialText, "/packages") {
			serialText = strings.TrimSuffix(serialText, "/packages")
			handleDevicePackages(writer, request, serialText)
			return
		}
		if strings.HasSuffix(serialText, "/screenshot") {
			serialText = strings.TrimSuffix(serialText, "/screenshot")
			handleDeviceScreenshot(writer, serialText)
			return
		}
		serial, err := url.PathUnescape(serialText)
		if err != nil || serial == "" || strings.Contains(serial, "/") {
			writeAPIError(writer, http.StatusNotFound, "device_not_found", "Device not found")
			return
		}
		adb := discoverADBVersion()
		if adb.hasFailure() {
			writeAPIError(writer, http.StatusServiceUnavailable, "adb_unavailable", "ADB is unavailable")
			return
		}
		devices, err := discoverADBDevices(adb.path)
		if err != nil {
			writeAPIError(writer, http.StatusBadGateway, "adb_devices_failed", "ADB device inventory failed")
			return
		}
		for _, device := range devices {
			if device.Serial == serial {
				writeJSON(writer, http.StatusOK, deviceDetailResponse{
					ADB: devicesADBStatus{
						Status:     "available",
						Executable: adb.path,
						Version:    adb.version,
					},
					Device: device,
				})
				return
			}
		}
		writeAPIError(writer, http.StatusNotFound, "device_not_found", "Device not found")
	})
	mux.HandleFunc("/api/v1/artifacts", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writeUnknownRoute(writer)
			return
		}
		handleArtifactUpload(writer, request, dataDir)
	})
	mux.HandleFunc("/api/v1/", func(writer http.ResponseWriter, request *http.Request) {
		writeUnknownRoute(writer)
	})
	return securityPolicyHandler(bind, mux)
}

type stagedArtifact struct {
	id           string
	dir          string
	tempPath     string
	finalPath    string
	originalName string
	sizeBytes    int64
	sha256       string
	contentType  string
}

var errInvalidArtifactUpload = errors.New("invalid artifact upload")
var errArtifactStorageUnavailable = errors.New("artifact storage unavailable")

func handleArtifactUpload(writer http.ResponseWriter, request *http.Request, dataDir string) {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" {
		writeAPIError(writer, http.StatusUnsupportedMediaType, "unsupported_artifact_media", "Unsupported artifact media type")
		return
	}
	if request.ContentLength > maxArtifactMultipartBytes {
		writeAPIError(writer, http.StatusBadRequest, "invalid_artifact_upload", "Invalid artifact upload")
		return
	}

	request.Body = http.MaxBytesReader(writer, request.Body, maxArtifactMultipartBytes)
	reader, err := request.MultipartReader()
	if err != nil {
		writeAPIError(writer, http.StatusUnsupportedMediaType, "unsupported_artifact_media", "Unsupported artifact media type")
		return
	}

	var staged *stagedArtifact
	artifactParts := 0
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			cleanupStagedArtifact(staged)
			writeAPIError(writer, http.StatusBadRequest, "invalid_artifact_upload", "Invalid artifact upload")
			return
		}

		fileName := part.FileName()
		if fileName == "" {
			_, _ = io.Copy(io.Discard, part)
			_ = part.Close()
			continue
		}
		if part.FormName() != "artifact" || artifactParts > 0 {
			_ = part.Close()
			cleanupStagedArtifact(staged)
			writeAPIError(writer, http.StatusBadRequest, "invalid_artifact_upload", "Invalid artifact upload")
			return
		}
		artifactParts++

		staged, err = stageArtifactPart(dataDir, fileName, part)
		_ = part.Close()
		if errors.Is(err, errArtifactStorageUnavailable) {
			cleanupStagedArtifact(staged)
			writeAPIError(writer, http.StatusInsufficientStorage, "artifact_storage_unavailable", "Artifact storage is unavailable")
			return
		}
		if err != nil {
			cleanupStagedArtifact(staged)
			writeAPIError(writer, http.StatusBadRequest, "invalid_artifact_upload", "Invalid artifact upload")
			return
		}
	}
	if artifactParts != 1 || staged == nil {
		cleanupStagedArtifact(staged)
		writeAPIError(writer, http.StatusBadRequest, "invalid_artifact_upload", "Invalid artifact upload")
		return
	}

	metadata := artifactMetadata{
		ID:             staged.id,
		OriginalName:   staged.originalName,
		SizeBytes:      staged.sizeBytes,
		SHA256:         staged.sha256,
		ContentType:    staged.contentType,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
		AnalysisStatus: "pending",
	}
	if err := finalizeArtifact(staged, metadata); err != nil {
		cleanupStagedArtifact(staged)
		writeAPIError(writer, http.StatusInsufficientStorage, "artifact_storage_unavailable", "Artifact storage is unavailable")
		return
	}
	writeJSON(writer, http.StatusCreated, artifactUploadResponse{Artifact: metadata})
}

func stageArtifactPart(dataDir, fileName string, part *multipart.Part) (*stagedArtifact, error) {
	originalName := artifactOriginalName(fileName)
	if originalName == "" || !strings.EqualFold(filepath.Ext(originalName), ".apk") {
		return nil, errInvalidArtifactUpload
	}

	id, err := randomArtifactID()
	if err != nil {
		return nil, errArtifactStorageUnavailable
	}
	artifactDir := filepath.Join(dataDir, "artifacts", id)
	staged := &stagedArtifact{
		id:           id,
		dir:          artifactDir,
		tempPath:     filepath.Join(artifactDir, "upload.tmp"),
		finalPath:    filepath.Join(artifactDir, "original.apk"),
		originalName: originalName,
	}
	if err := os.MkdirAll(artifactDir, 0o700); err != nil {
		return staged, errArtifactStorageUnavailable
	}
	tempFile, err := os.OpenFile(staged.tempPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return staged, errArtifactStorageUnavailable
	}

	hasher := sha256.New()
	limited := &io.LimitedReader{R: part, N: maxArtifactSizeBytes + 1}
	buffer := make([]byte, 32*1024)
	firstBytes := make([]byte, 0, 4)
	var size int64
	for {
		n, readErr := limited.Read(buffer)
		if n > 0 {
			chunk := buffer[:n]
			if len(firstBytes) < 4 {
				needed := 4 - len(firstBytes)
				if needed > len(chunk) {
					needed = len(chunk)
				}
				firstBytes = append(firstBytes, chunk[:needed]...)
			}
			if _, err := tempFile.Write(chunk); err != nil {
				_ = tempFile.Close()
				return staged, errArtifactStorageUnavailable
			}
			if _, err := hasher.Write(chunk); err != nil {
				_ = tempFile.Close()
				return staged, errArtifactStorageUnavailable
			}
			size += int64(n)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			_ = tempFile.Close()
			return staged, errInvalidArtifactUpload
		}
	}
	if err := tempFile.Close(); err != nil {
		return staged, errArtifactStorageUnavailable
	}
	if limited.N == 0 || size == 0 || !hasZIPSignature(firstBytes) {
		return staged, errInvalidArtifactUpload
	}
	staged.sizeBytes = size
	staged.sha256 = hex.EncodeToString(hasher.Sum(nil))
	partMediaType := part.Header.Get("Content-Type")
	if contentType, _, err := mime.ParseMediaType(partMediaType); err == nil {
		staged.contentType = contentType
	} else {
		staged.contentType = partMediaType
	}
	return staged, nil
}

func artifactOriginalName(fileName string) string {
	normalized := strings.ReplaceAll(fileName, "\\", "/")
	return pathBase(normalized)
}

func pathBase(value string) string {
	index := strings.LastIndex(value, "/")
	if index >= 0 {
		value = value[index+1:]
	}
	if value == "." || value == ".." {
		return ""
	}
	return value
}

func hasZIPSignature(firstBytes []byte) bool {
	if len(firstBytes) < 4 || firstBytes[0] != 'P' || firstBytes[1] != 'K' {
		return false
	}
	return firstBytes[2] == 0x03 && firstBytes[3] == 0x04 ||
		firstBytes[2] == 0x05 && firstBytes[3] == 0x06 ||
		firstBytes[2] == 0x07 && firstBytes[3] == 0x08
}

func finalizeArtifact(staged *stagedArtifact, metadata artifactMetadata) error {
	if err := os.Rename(staged.tempPath, staged.finalPath); err != nil {
		return errArtifactStorageUnavailable
	}
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(artifactUploadResponse{Artifact: metadata}); err != nil {
		return errArtifactStorageUnavailable
	}
	tempMetadataPath := filepath.Join(staged.dir, "metadata.json.tmp")
	if err := os.WriteFile(tempMetadataPath, body.Bytes(), 0o600); err != nil {
		return errArtifactStorageUnavailable
	}
	if err := os.Rename(tempMetadataPath, filepath.Join(staged.dir, "metadata.json")); err != nil {
		return errArtifactStorageUnavailable
	}
	return nil
}

func cleanupStagedArtifact(staged *stagedArtifact) {
	if staged != nil {
		_ = os.RemoveAll(staged.dir)
	}
}

func randomArtifactID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

func handleDeviceLogcat(writer http.ResponseWriter, request *http.Request, serialText string) {
	linesLimit, format, ok := parseLogcatQuery(request.URL.Query())
	if !ok {
		writeAPIError(writer, http.StatusBadRequest, "invalid_logcat_request", "Invalid logcat request")
		return
	}
	serial, err := url.PathUnescape(serialText)
	if err != nil || serial == "" || strings.Contains(serial, "/") {
		writeAPIError(writer, http.StatusNotFound, "device_not_found", "Device not found")
		return
	}
	adb := discoverADBVersion()
	if adb.hasFailure() {
		writeAPIError(writer, http.StatusServiceUnavailable, "adb_unavailable", "ADB is unavailable")
		return
	}
	devices, err := discoverADBDevices(adb.path)
	if err != nil {
		writeAPIError(writer, http.StatusBadGateway, "adb_devices_failed", "ADB device inventory failed")
		return
	}
	for _, device := range devices {
		if device.Serial != serial {
			continue
		}
		if device.State != "device" {
			writeAPIError(writer, http.StatusConflict, "device_not_ready", "Device is not ready")
			return
		}
		output, err := discoverADBLogcat(adb.path, serial)
		if err != nil {
			writeAPIError(writer, http.StatusBadGateway, "adb_logcat_failed", "ADB logcat failed")
			return
		}
		lines, truncated := lastLogcatLines(output, linesLimit)
		writeJSON(writer, http.StatusOK, logcatResponse{
			Device: logcatDevice{
				Serial: device.Serial,
				State:  device.State,
			},
			Logcat: logcatPayload{
				Format:    format,
				Lines:     lines,
				Truncated: truncated,
			},
		})
		return
	}
	writeAPIError(writer, http.StatusNotFound, "device_not_found", "Device not found")
}

func handleDevicePackages(writer http.ResponseWriter, request *http.Request, serialText string) {
	scope, ok := parsePackageScope(request.URL.Query())
	if !ok {
		writeAPIError(writer, http.StatusBadRequest, "invalid_package_request", "Invalid package request")
		return
	}
	serial, err := url.PathUnescape(serialText)
	if err != nil || serial == "" || strings.Contains(serial, "/") {
		writeAPIError(writer, http.StatusNotFound, "device_not_found", "Device not found")
		return
	}
	adb := discoverADBVersion()
	if adb.hasFailure() {
		writeAPIError(writer, http.StatusServiceUnavailable, "adb_unavailable", "ADB is unavailable")
		return
	}
	devices, err := discoverADBDevices(adb.path)
	if err != nil {
		writeAPIError(writer, http.StatusBadGateway, "adb_devices_failed", "ADB device inventory failed")
		return
	}
	for _, device := range devices {
		if device.Serial != serial {
			continue
		}
		if device.State != "device" {
			writeAPIError(writer, http.StatusConflict, "device_not_ready", "Device is not ready")
			return
		}
		items, err := discoverADBPackages(adb.path, serial, scope)
		if err != nil {
			writeAPIError(writer, http.StatusBadGateway, "adb_packages_failed", "ADB package inventory failed")
			return
		}
		writeJSON(writer, http.StatusOK, packagesResponse{
			Device: packageDevice{
				Serial: device.Serial,
				State:  device.State,
			},
			Packages: packagePayload{
				Scope: scope,
				Items: items,
				Count: len(items),
			},
		})
		return
	}
	writeAPIError(writer, http.StatusNotFound, "device_not_found", "Device not found")
}

func handleDevicePackageDetail(writer http.ResponseWriter, request *http.Request, serialText, packageText string) {
	serial, err := url.PathUnescape(serialText)
	if err != nil || serial == "" || strings.Contains(serial, "/") {
		writeAPIError(writer, http.StatusNotFound, "device_not_found", "Device not found")
		return
	}
	packageName, err := url.PathUnescape(packageText)
	if err != nil || !validPackageName(packageName) {
		writeAPIError(writer, http.StatusBadRequest, "invalid_package_request", "Invalid package request")
		return
	}
	adb := discoverADBVersion()
	if adb.hasFailure() {
		writeAPIError(writer, http.StatusServiceUnavailable, "adb_unavailable", "ADB is unavailable")
		return
	}
	devices, err := discoverADBDevices(adb.path)
	if err != nil {
		writeAPIError(writer, http.StatusBadGateway, "adb_devices_failed", "ADB device inventory failed")
		return
	}
	for _, device := range devices {
		if device.Serial != serial {
			continue
		}
		if device.State != "device" {
			writeAPIError(writer, http.StatusConflict, "device_not_ready", "Device is not ready")
			return
		}
		detail, err := discoverADBPackageDetail(adb.path, serial, packageName)
		if errors.Is(err, errPackageNotFound) {
			writeAPIError(writer, http.StatusNotFound, "package_not_found", "Package not found")
			return
		}
		if err != nil {
			writeAPIError(writer, http.StatusBadGateway, "adb_package_detail_failed", "ADB package detail failed")
			return
		}
		writeJSON(writer, http.StatusOK, packageDetailResponse{
			Device: packageDevice{
				Serial: device.Serial,
				State:  device.State,
			},
			Package: detail,
		})
		return
	}
	writeAPIError(writer, http.StatusNotFound, "device_not_found", "Device not found")
}

func validPackageName(name string) bool {
	if name == "" || len(name) > 255 {
		return false
	}
	segments := strings.Split(name, ".")
	for _, segment := range segments {
		if segment == "" {
			return false
		}
		for index, r := range segment {
			if r > 127 {
				return false
			}
			if index == 0 {
				if !(r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r == '_') {
					return false
				}
				continue
			}
			if !(r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_') {
				return false
			}
		}
	}
	return true
}

func handleDeviceScreenshot(writer http.ResponseWriter, serialText string) {
	serial, err := url.PathUnescape(serialText)
	if err != nil || serial == "" || strings.Contains(serial, "/") {
		writeAPIError(writer, http.StatusNotFound, "device_not_found", "Device not found")
		return
	}
	adb := discoverADBVersion()
	if adb.hasFailure() {
		writeAPIError(writer, http.StatusServiceUnavailable, "adb_unavailable", "ADB is unavailable")
		return
	}
	devices, err := discoverADBDevices(adb.path)
	if err != nil {
		writeAPIError(writer, http.StatusBadGateway, "adb_devices_failed", "ADB device inventory failed")
		return
	}
	for _, device := range devices {
		if device.Serial != serial {
			continue
		}
		if device.State != "device" {
			writeAPIError(writer, http.StatusConflict, "device_not_ready", "Device is not ready")
			return
		}
		output, err := discoverADBScreenshot(adb.path, serial)
		if err != nil {
			writeAPIError(writer, http.StatusBadGateway, "adb_screenshot_failed", "ADB screenshot failed")
			return
		}
		writer.Header().Set("Content-Type", "image/png")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(output)
		return
	}
	writeAPIError(writer, http.StatusNotFound, "device_not_found", "Device not found")
}

func parsePackageScope(values url.Values) (string, bool) {
	scope := "all"
	if scopeValues, ok := values["scope"]; ok {
		if len(scopeValues) != 1 {
			return "", false
		}
		scope = scopeValues[0]
	}
	switch scope {
	case "all", "third-party", "system":
		return scope, true
	default:
		return "", false
	}
}

func parseLogcatQuery(values url.Values) (int, string, bool) {
	linesLimit := 200
	if linesValues, ok := values["lines"]; ok {
		if len(linesValues) != 1 {
			return 0, "", false
		}
		parsed, err := strconv.Atoi(linesValues[0])
		if err != nil || parsed < 1 || parsed > 500 {
			return 0, "", false
		}
		linesLimit = parsed
	}
	format := "plain"
	if formatValues, ok := values["format"]; ok {
		if len(formatValues) != 1 || formatValues[0] != "plain" {
			return 0, "", false
		}
		format = formatValues[0]
	}
	return linesLimit, format, true
}

func newBootstrapResponse() (bootstrapResponse, error) {
	csrfToken, err := randomToken()
	if err != nil {
		return bootstrapResponse{}, err
	}
	webSocketToken, err := randomToken()
	if err != nil {
		return bootstrapResponse{}, err
	}
	return bootstrapResponse{
		CSRFToken:      csrfToken,
		WebSocketToken: webSocketToken,
		StatusURL:      "/api/v1/status",
	}, nil
}

func randomToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

func securityPolicyHandler(bind string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if isCurrentAPIPath(request.URL.Path) {
			if request.URL.Host != "" && !allowedServerHost(request.URL.Host, bind) {
				writeForbidden(writer, "forbidden_absolute_url_host")
				return
			}
			if !allowedServerHost(request.Host, bind) {
				writeForbidden(writer, "forbidden_host")
				return
			}
			if !allowedOrigin(request.Header.Get("Origin"), bind) {
				writeForbidden(writer, "forbidden_origin")
				return
			}
		}
		next.ServeHTTP(writer, request)
	})
}

func isCurrentAPIPath(path string) bool {
	return path == "/api/v1" || strings.HasPrefix(path, "/api/v1/")
}

func allowedOrigin(origin, bind string) bool {
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme != "http" || parsed.User != nil || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}
	return allowedServerHost(parsed.Host, bind)
}

func allowedServerHost(candidate, bind string) bool {
	host, port, err := net.SplitHostPort(candidate)
	if err != nil {
		return false
	}
	_, bindPort, err := net.SplitHostPort(bind)
	if err != nil || port != bindPort {
		return false
	}
	return isLoopbackHost(host)
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func writeForbidden(writer http.ResponseWriter, code string) {
	writeAPIError(writer, http.StatusForbidden, code, "Request rejected by dashboard browser security policy")
}

func writeAPIError(writer http.ResponseWriter, status int, code, message string) {
	writeJSON(writer, status, errorEnvelope{
		Error: apiError{
			Code:      code,
			Message:   message,
			Details:   map[string]any{},
			RequestID: nil,
		},
	})
}

func writeUnknownRoute(writer http.ResponseWriter) {
	writeJSON(writer, http.StatusNotFound, errorEnvelope{
		Error: apiError{
			Code:      "not_found",
			Message:   "Unknown API route",
			Details:   map[string]any{},
			RequestID: nil,
		},
	})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)
}

func writeHTML(writer http.ResponseWriter, body string) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(body))
}

func openDashboardBrowser(stderr *os.File, url string) {
	err := exec.Command("xdg-open", url).Run()
	if err != nil {
		fmt.Fprintf(stderr, "%s WARN browser open failed url=%s error=%s\n", logTimestamp(time.Now()), url, err)
	}
}

func logTimestamp(at time.Time) string {
	return at.Format("2006-01-02T15:04:05-07:00")
}

const dashboardShellHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>adb-dashboard</title>
  <style>
    :root {
      color-scheme: light;
      font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background: #f5f7fb;
      color: #172033;
    }
    body {
      margin: 0;
      min-height: 100vh;
    }
    main {
      box-sizing: border-box;
      width: min(960px, 100%);
      margin: 0 auto;
      padding: 32px 20px;
    }
    h1 {
      margin: 0 0 20px;
      font-size: 2rem;
      font-weight: 700;
    }
    dl {
      display: grid;
      grid-template-columns: max-content minmax(0, 1fr);
      gap: 12px 18px;
      margin: 0;
      padding: 20px 0;
      border-top: 1px solid #cfd7e6;
      border-bottom: 1px solid #cfd7e6;
    }
    dt {
      font-weight: 700;
      color: #43516b;
    }
    dd {
      margin: 0;
      min-width: 0;
      overflow-wrap: anywhere;
    }
    .device-list {
      white-space: pre-line;
    }
    .controls {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      margin-top: 8px;
    }
    button {
      border: 1px solid #9aa8bf;
      background: #ffffff;
      color: #172033;
      border-radius: 6px;
      padding: 6px 10px;
      font: inherit;
      cursor: pointer;
    }
    .device-detail {
      margin-top: 8px;
      white-space: pre-line;
    }
    .device-screenshot {
      display: block;
      margin-top: 8px;
      max-width: min(100%, 420px);
      height: auto;
      border: 1px solid #cfd7e6;
    }
  </style>
</head>
<body>
  <main>
    <h1 id="application-name">adb-dashboard</h1>
    <dl aria-label="Current dashboard status">
      <dt>server</dt><dd id="server-status">server: unavailable</dd>
      <dt>bind</dt><dd id="server-bind">bind: unavailable</dd>
      <dt>read-only</dt><dd id="server-read-only">read-only: unavailable</dd>
      <dt>adb</dt><dd id="adb-status">adb: unavailable</dd>
      <dt>devices</dt><dd><span id="device-count">devices: unavailable</span><div id="devices-list" class="device-list"></div><div class="controls"><button type="button" id="devices-refresh">refresh</button><button type="button" id="device-detail-first">details</button><button type="button" id="device-logcat-first">logcat</button><button type="button" id="device-screenshot-first">screenshot</button><button type="button" id="device-packages-first">packages</button></div><div id="device-detail" class="device-detail">detail: unavailable</div><div id="device-logcat" class="device-detail">logcat: unavailable</div><div id="device-screenshot" class="device-detail">screenshot: unavailable</div><img id="device-screenshot-image" class="device-screenshot" alt=""><div class="controls"><button type="button" id="package-scope-all">all</button><button type="button" id="package-scope-third-party">third-party</button><button type="button" id="package-scope-system">system</button></div><div id="device-packages" class="device-detail">packages: unavailable</div><div id="device-packages-list" class="device-list"></div><div id="device-package-detail" class="device-detail">package detail: unavailable</div></dd>
      <dt>watcher</dt><dd id="watcher-status">watcher: unavailable</dd>
      <dt>jobs</dt><dd id="jobs-status">jobs: unavailable</dd>
      <dt>sessions</dt><dd id="sessions-status">sessions: unavailable</dd>
      <dt>storage</dt><dd id="storage-status">storage: unavailable</dd>
      <dt>host tools</dt><dd id="host-tools-status">host tools: unavailable</dd>
    </dl>
  </main>
  <script>
(() => {
  const setText = (id, text) => {
    const target = document.getElementById(id);
    if (target) {
      target.textContent = text;
    }
  };

  const setScreenshotImage = (src, alt) => {
    const target = document.getElementById("device-screenshot-image");
    if (!target) {
      return;
    }
    target.src = src || "";
    target.alt = alt || "";
  };

  let latestDevices = [];
  let latestPackages = [];
  let latestPackageScope = "all";
  let refreshSequence = 0;
  let packageDetailSequence = 0;

  const packageDetailButtonID = (packageName) => {
    return "package-detail-" + String(packageName || "");
  };

  const clearPackageDetail = () => {
    packageDetailSequence++;
    setText("device-package-detail", "package detail: unavailable");
  };

  const clearPackageRows = () => {
    latestPackages = [];
    const target = document.getElementById("device-packages-list");
    if (target && target.replaceChildren) {
      target.replaceChildren();
      return;
    }
    setText("device-packages-list", "");
  };

  const renderPackageRows = (items) => {
    latestPackages = items;
    const target = document.getElementById("device-packages-list");
    if (!target || !target.replaceChildren || !document.createElement) {
      setText("device-packages-list", items.map((item) => packageRowText(item)).join("\n"));
      return;
    }
    target.replaceChildren();
    for (const item of items) {
      const row = document.createElement("div");
      const button = document.createElement("button");
      button.type = "button";
      button.id = packageDetailButtonID(item.name);
      button.textContent = packageRowText(item);
      button.addEventListener("click", () => loadPackageDetail(item.name));
      row.appendChild(button);
      target.appendChild(row);
    }
  };

  const packageRowText = (item) => {
    const parts = [String(item.name || "")];
    if (item.versionCode) {
      parts.push("versionCode:" + String(item.versionCode));
    }
    if (item.userId) {
      parts.push("uid:" + String(item.userId));
    }
    if (item.apkPath) {
      parts.push(String(item.apkPath));
    }
    return parts.join(" ");
  };

  const unavailable = () => {
    setText("server-status", "server: unavailable");
    setText("server-bind", "bind: unavailable");
    setText("server-read-only", "read-only: unavailable");
    setText("adb-status", "adb: unavailable");
    setText("device-count", "devices: unavailable");
    setText("devices-list", "");
    setText("device-detail", "detail: unavailable");
    setText("device-logcat", "logcat: unavailable");
    setText("device-screenshot", "screenshot: unavailable");
    setScreenshotImage("", "");
    setText("device-packages", "packages: unavailable");
    clearPackageRows();
    clearPackageDetail();
    setText("watcher-status", "watcher: unavailable");
    setText("jobs-status", "jobs: unavailable");
    setText("sessions-status", "sessions: unavailable");
    setText("storage-status", "storage: unavailable");
    setText("host-tools-status", "host tools: unavailable");
  };

  const devicesUnavailable = () => {
    latestDevices = [];
    setText("device-count", "devices: unavailable");
    setText("devices-list", "");
    setText("device-detail", "detail: unavailable");
    setText("device-logcat", "logcat: unavailable");
    setText("device-screenshot", "screenshot: unavailable");
    setScreenshotImage("", "");
    setText("device-packages", "packages: unavailable");
    clearPackageRows();
    clearPackageDetail();
  };

  const loadDevices = async () => {
    const sequence = ++refreshSequence;
    try {
      const devicesResponse = await fetch("/api/v1/devices", { credentials: "same-origin" });
      if (!devicesResponse.ok) {
        throw new Error("devices unavailable");
      }
      const inventory = await devicesResponse.json();
      const devices = Array.isArray(inventory.devices) ? inventory.devices : [];
      if (sequence !== refreshSequence) {
        return;
      }
      latestDevices = devices;
      setText("device-count", "devices: " + String(devices.length));
      setText("devices-list", devices.map((device) => {
        return String(device.serial || "") + " " + String(device.state || "");
      }).join("\n"));
      setText("device-detail", "detail: unavailable");
      setText("device-logcat", "logcat: unavailable");
      setText("device-screenshot", "screenshot: unavailable");
      setScreenshotImage("", "");
      setText("device-packages", "packages: unavailable");
      clearPackageRows();
      clearPackageDetail();
    } catch (_) {
      if (sequence !== refreshSequence) {
        return;
      }
      devicesUnavailable();
    }
  };

  const loadFirstDeviceDetail = async () => {
    const device = latestDevices[0];
    if (!device || !device.serial) {
      setText("device-detail", "detail: unavailable");
      return;
    }
    try {
      const detailResponse = await fetch("/api/v1/devices/" + encodeURIComponent(device.serial), { credentials: "same-origin" });
      if (!detailResponse.ok) {
        throw new Error("detail unavailable");
      }
      const detail = await detailResponse.json();
      const current = detail.device || {};
      const lines = [
        "serial: " + String(current.serial || ""),
        "state: " + String(current.state || ""),
      ];
      if (current.product) {
        lines.push("product: " + String(current.product));
      }
      if (current.model) {
        lines.push("model: " + String(current.model));
      }
      if (current.device) {
        lines.push("device: " + String(current.device));
      }
      if (current.transportId) {
        lines.push("transport: " + String(current.transportId));
      }
      setText("device-detail", lines.join("\n"));
    } catch (_) {
      setText("device-detail", "detail: unavailable");
    }
  };

  const loadFirstDeviceLogcat = async () => {
    const device = latestDevices[0];
    if (!device || !device.serial) {
      setText("device-logcat", "logcat: unavailable");
      return;
    }
    setText("device-logcat", "logcat: loading");
    try {
      const logcatResponse = await fetch("/api/v1/devices/" + encodeURIComponent(device.serial) + "/logcat?lines=200&format=plain", { credentials: "same-origin" });
      if (!logcatResponse.ok) {
        throw new Error("logcat unavailable");
      }
      const payload = await logcatResponse.json();
      const current = payload.device || {};
      const logcat = payload.logcat || {};
      const lines = Array.isArray(logcat.lines) ? logcat.lines : [];
      if (lines.length === 0) {
        setText("device-logcat", "logcat: empty");
        return;
      }
      setText("device-logcat", ["logcat: " + String(current.serial || device.serial)].concat(lines.map((line) => String(line))).join("\n"));
    } catch (_) {
      setText("device-logcat", "logcat: unavailable");
    }
  };

  const loadFirstDeviceScreenshot = async () => {
    const device = latestDevices[0];
    if (!device || !device.serial) {
      setText("device-screenshot", "screenshot: unavailable");
      setScreenshotImage("", "");
      return;
    }
    setText("device-screenshot", "screenshot: loading");
    setScreenshotImage("", "");
    try {
      const screenshotResponse = await fetch("/api/v1/devices/" + encodeURIComponent(device.serial) + "/screenshot", { credentials: "same-origin" });
      if (!screenshotResponse.ok) {
        throw new Error("screenshot unavailable");
      }
      const buffer = await screenshotResponse.arrayBuffer();
      const bytes = Array.from(new Uint8Array(buffer));
      let binary = "";
      for (const value of bytes) {
        binary += String.fromCharCode(value);
      }
      setScreenshotImage("data:image/png;base64," + btoa(binary), "screenshot: " + String(device.serial));
      setText("device-screenshot", "screenshot: " + String(device.serial));
    } catch (_) {
      setText("device-screenshot", "screenshot: unavailable");
      setScreenshotImage("", "");
    }
  };

  const loadFirstDevicePackages = async (scope) => {
    const device = latestDevices[0];
    latestPackageScope = scope || latestPackageScope || "all";
    if (!device || !device.serial) {
      setText("device-packages", "packages: unavailable");
      clearPackageRows();
      clearPackageDetail();
      return;
    }
    setText("device-packages", "packages: loading");
    clearPackageRows();
    clearPackageDetail();
    try {
      const packagesResponse = await fetch("/api/v1/devices/" + encodeURIComponent(device.serial) + "/packages?scope=" + encodeURIComponent(latestPackageScope), { credentials: "same-origin" });
      if (!packagesResponse.ok) {
        throw new Error("packages unavailable");
      }
      const payload = await packagesResponse.json();
      const packages = payload.packages || {};
      const currentScope = String(packages.scope || latestPackageScope);
      const items = Array.isArray(packages.items) ? packages.items : [];
      setText("device-packages", "packages: " + currentScope + " count=" + String(items.length));
      if (items.length === 0) {
        setText("device-packages-list", "empty");
        return;
      }
      renderPackageRows(items);
    } catch (_) {
      setText("device-packages", "packages: unavailable");
      clearPackageRows();
      clearPackageDetail();
    }
  };

  const loadPackageDetail = async (packageName) => {
    const device = latestDevices[0];
    const selectedName = String(packageName || "");
    const selectedSerial = String((device && device.serial) || "");
    if (!selectedSerial || !selectedName) {
      clearPackageDetail();
      return;
    }
    const sequence = ++packageDetailSequence;
    setText("device-package-detail", "package detail: loading");
    try {
      const detailResponse = await fetch("/api/v1/devices/" + encodeURIComponent(selectedSerial) + "/packages/" + encodeURIComponent(selectedName), { credentials: "same-origin" });
      if (!detailResponse.ok) {
        throw new Error("package detail unavailable");
      }
      const payload = await detailResponse.json();
      const currentDevice = latestDevices[0];
      const selectedPackageVisible = latestPackages.some((item) => String(item.name || "") === selectedName);
      if (sequence !== packageDetailSequence || !currentDevice || String(currentDevice.serial || "") !== selectedSerial || !selectedPackageVisible) {
        return;
      }
      const current = payload.device || {};
      const pkg = payload.package || {};
      const name = String(pkg.name || selectedName);
      const serial = String(current.serial || selectedSerial);
      const lines = [
        "package detail: " + serial + " " + name,
        "package: " + name,
      ];
      if (pkg.versionName) {
        lines.push("versionName: " + String(pkg.versionName));
      }
      if (pkg.versionCode) {
        lines.push("versionCode: " + String(pkg.versionCode));
      }
      if (pkg.installer) {
        lines.push("installer: " + String(pkg.installer));
      }
      if (pkg.firstInstallTime) {
        lines.push("firstInstallTime: " + String(pkg.firstInstallTime));
      }
      if (pkg.lastUpdateTime) {
        lines.push("lastUpdateTime: " + String(pkg.lastUpdateTime));
      }
      const permissions = Array.isArray(pkg.requestedPermissions) ? pkg.requestedPermissions : [];
      for (const permission of permissions) {
        lines.push("permission: " + String(permission));
      }
      const summaryLines = Array.isArray(pkg.summaryLines) ? pkg.summaryLines : [];
      for (const line of summaryLines) {
        lines.push("summary: " + String(line));
      }
      setText("device-package-detail", lines.join("\n"));
    } catch (_) {
      if (sequence === packageDetailSequence) {
        clearPackageDetail();
      }
    }
  };

  const loadStatus = async () => {
    try {
      const bootstrapResponse = await fetch("/api/v1/bootstrap", { credentials: "same-origin" });
      if (!bootstrapResponse.ok) {
        throw new Error("bootstrap unavailable");
      }
      const bootstrap = await bootstrapResponse.json();
      const statusResponse = await fetch(bootstrap.statusUrl, { credentials: "same-origin" });
      if (!statusResponse.ok) {
        throw new Error("status unavailable");
      }
      const status = await statusResponse.json();
      setText("application-name", status.application.name);
      setText("server-status", "server: " + status.server.status);
      setText("server-bind", "bind: " + status.server.bind);
      setText("server-read-only", "read-only: " + String(status.server.readOnly));
      setText("adb-status", "adb: " + status.adb.status);
      if (status.adb.status === "available") {
        await loadDevices();
      } else {
        devicesUnavailable();
      }
      setText("watcher-status", "watcher: " + status.watcher.status);
      setText("jobs-status", "jobs: " + status.jobs.status);
      setText("sessions-status", "sessions: " + status.sessions.status);
      setText("storage-status", "storage: " + status.storage.status);
      setText("host-tools-status", "host tools: " + status.hostTools.status);
    } catch (_) {
      unavailable();
    }
  };

  document.addEventListener("DOMContentLoaded", () => {
    const refresh = document.getElementById("devices-refresh");
    if (refresh) {
      refresh.addEventListener("click", loadDevices);
    }
    const detail = document.getElementById("device-detail-first");
    if (detail) {
      detail.addEventListener("click", loadFirstDeviceDetail);
    }
    const logcat = document.getElementById("device-logcat-first");
    if (logcat) {
      logcat.addEventListener("click", loadFirstDeviceLogcat);
    }
    const screenshot = document.getElementById("device-screenshot-first");
    if (screenshot) {
      screenshot.addEventListener("click", loadFirstDeviceScreenshot);
    }
    const packages = document.getElementById("device-packages-first");
    if (packages) {
      packages.addEventListener("click", () => loadFirstDevicePackages(latestPackageScope));
    }
    const allPackages = document.getElementById("package-scope-all");
    if (allPackages) {
      allPackages.addEventListener("click", () => loadFirstDevicePackages("all"));
    }
    const thirdPartyPackages = document.getElementById("package-scope-third-party");
    if (thirdPartyPackages) {
      thirdPartyPackages.addEventListener("click", () => loadFirstDevicePackages("third-party"));
    }
    const systemPackages = document.getElementById("package-scope-system");
    if (systemPackages) {
      systemPackages.addEventListener("click", () => loadFirstDevicePackages("system"));
    }
    loadStatus();
  });
})();
  </script>
</body>
</html>
`
