package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

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
	overall := "PASS"
	exitCode := 0
	if dataErr != nil || tempErr != nil {
		overall = "FAIL"
		exitCode = 5
	}

	fmt.Fprintln(stdout, "adb-dashboard doctor")
	fmt.Fprintf(stdout, "overall: %s\n", overall)
	fmt.Fprintf(stdout, "config: PASS source=%s listen=%s readOnly=%t logLevel=%s\n", reportSource(cfg), cfg.listen.value, cfg.readOnly.value, cfg.logLevel.value)
	printDirRow(stdout, "dataDir", cfg.dataDir.value, dataErr)
	printDirRow(stdout, "tempDir", cfg.tempDir.value, tempErr)
	fmt.Fprintln(stdout, "cacheDir: NIY storage.cache is not implemented yet")
	fmt.Fprintln(stdout, "projectDir: NIY storage.projects is not implemented yet")
	fmt.Fprintln(stdout, "adbExecutable: NIY adb.discovery is not implemented yet")
	fmt.Fprintln(stdout, "adbVersion: NIY adb.discovery is not implemented yet")
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

	startedAt := time.Now()
	actualAddr := listener.Addr().String()
	server := &http.Server{
		Handler: dashboardHandler(startedAt, actualAddr, cfg.readOnly.value),
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

func dashboardHandler(startedAt time.Time, bind string, readOnly bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/status", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writeUnknownRoute(writer)
			return
		}
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
			ADB: adbStatus{
				Status:           "NIY",
				Executable:       nil,
				Version:          nil,
				ServerResponsive: "NIY",
			},
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
	mux.HandleFunc("/api/v1/", func(writer http.ResponseWriter, request *http.Request) {
		writeUnknownRoute(writer)
	})
	return mux
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

func openDashboardBrowser(stderr *os.File, url string) {
	err := exec.Command("xdg-open", url).Run()
	if err != nil {
		fmt.Fprintf(stderr, "%s WARN browser open failed url=%s error=%s\n", logTimestamp(time.Now()), url, err)
	}
}

func logTimestamp(at time.Time) string {
	return at.Format("2006-01-02T15:04:05-07:00")
}
