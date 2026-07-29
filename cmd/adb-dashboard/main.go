package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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
		fmt.Fprintln(stderr, "NIY: server.start is not implemented yet")
		return 6
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
		return validateServe(args[1:], stderr)
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

type fileConfig struct {
	Server struct {
		Listen      *string `toml:"listen"`
		OpenBrowser *bool   `toml:"open_browser"`
		ReadOnly    *bool   `toml:"read_only"`
		DataDir     *string `toml:"data_dir"`
		TempDir     *string `toml:"temp_dir"`
	} `toml:"server"`
	Logging struct {
		Level *string `toml:"level"`
	} `toml:"logging"`
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

	dataOK, dataErr := ensureDir(cfg.dataDir.value)
	tempOK, tempErr := ensureDir(cfg.tempDir.value)
	overall := "PASS"
	exitCode := 0
	if !dataOK || !tempOK {
		overall = "FAIL"
		exitCode = 5
	}

	fmt.Fprintln(stdout, "adb-dashboard doctor")
	fmt.Fprintf(stdout, "overall: %s\n", overall)
	fmt.Fprintf(stdout, "config: PASS source=%s listen=%s readOnly=%t logLevel=%s\n", reportSource(cfg), cfg.listen.value, cfg.readOnly.value, cfg.logLevel.value)
	printDirRow(stdout, "dataDir", cfg.dataDir.value, dataOK, dataErr)
	printDirRow(stdout, "tempDir", cfg.tempDir.value, tempOK, tempErr)
	fmt.Fprintln(stdout, "cacheDir: NIY storage.cache is not implemented yet")
	fmt.Fprintln(stdout, "projectDir: NIY storage.projects is not implemented yet")
	fmt.Fprintln(stdout, "adbExecutable: NIY adb.discovery is not implemented yet")
	fmt.Fprintln(stdout, "adbVersion: NIY adb.discovery is not implemented yet")
	fmt.Fprintln(stdout, "adbServer: NIY adb.server is not implemented yet")
	fmt.Fprintln(stdout, "devices: NIY devices.refresh is not implemented yet")
	fmt.Fprintln(stdout, "hostTools: NIY hosttools.discovery is not implemented yet")
	return exitCode
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

	var parsed fileConfig
	metadata, err := toml.Decode(string(content), &parsed)
	if err != nil {
		return fmt.Errorf("invalid configuration: cannot parse %s: %v", path, err)
	}
	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		return fmt.Errorf("invalid configuration: unknown key %s", strings.Join(undecoded[0], "."))
	}

	source := "file:" + path
	if parsed.Server.Listen != nil {
		cfg.listen = sourcedString{value: *parsed.Server.Listen, source: source}
	}
	if parsed.Server.OpenBrowser != nil {
		cfg.openBrowser = sourcedBool{value: *parsed.Server.OpenBrowser, source: source}
	}
	if parsed.Server.ReadOnly != nil {
		cfg.readOnly = sourcedBool{value: *parsed.Server.ReadOnly, source: source}
	}
	if parsed.Server.DataDir != nil {
		cfg.dataDir = sourcedString{value: *parsed.Server.DataDir, source: source}
	}
	if parsed.Server.TempDir != nil {
		cfg.tempDir = sourcedString{value: *parsed.Server.TempDir, source: source}
	}
	if parsed.Logging.Level != nil {
		cfg.logLevel = sourcedString{value: *parsed.Logging.Level, source: source}
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
	return filepath.Join(os.Getenv("HOME"), ".local", "state", "adb-dashboard")
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

func ensureDir(path string) (bool, error) {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return false, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func printDirRow(stdout *os.File, name, path string, ok bool, err error) {
	if ok {
		fmt.Fprintf(stdout, "%s: PASS path=%s\n", name, path)
		return
	}
	detail := "not a directory"
	if err != nil {
		detail = err.Error()
	}
	fmt.Fprintf(stdout, "%s: FAIL path=%s error=%s\n", name, path, detail)
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
	for i := 0; i < len(args); i++ {
		option := args[i]
		switch option {
		case "--listen", "--data-dir", "--config", "--temp-dir", "--log-level":
			if i+1 >= len(args) {
				fmt.Fprintf(stderr, "missing argument for %s\n", option)
				return 2
			}
			i++
		case "--open", "--no-open", "--read-only":
		default:
			if strings.HasPrefix(option, "--") {
				fmt.Fprintf(stderr, "unknown option: %s\n", option)
				return 2
			}
			fmt.Fprintf(stderr, "unknown command: %s\n", option)
			return 2
		}
	}

	fmt.Fprintln(stderr, "NIY: server.start is not implemented yet")
	return 6
}

func validateServe(args []string, stderr *os.File) int {
	for i := 0; i < len(args); i++ {
		option := args[i]
		switch option {
		case "--listen", "--data-dir", "--config", "--temp-dir", "--log-level":
			if i+1 >= len(args) {
				fmt.Fprintf(stderr, "missing argument for %s\n", option)
				return 2
			}
			i++
		case "--open", "--no-open", "--read-only":
		default:
			if strings.HasPrefix(option, "--") {
				fmt.Fprintf(stderr, "unknown option: %s\n", option)
				return 2
			}
			fmt.Fprintf(stderr, "unknown command: %s\n", option)
			return 2
		}
	}

	fmt.Fprintln(stderr, "NIY: server.start is not implemented yet")
	return 6
}

func printVersion(stdout *os.File) {
	fmt.Fprintf(stdout, "adb-dashboard %s\n", version)
	fmt.Fprintf(stdout, "commit: %s\n", commit)
	fmt.Fprintf(stdout, "buildDate: %s\n", buildDate)
	fmt.Fprintf(stdout, "goVersion: %s\n", runtime.Version())
	fmt.Fprintf(stdout, "frontendRevision: %s\n", frontendRevision)
}
