package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
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
		fmt.Fprintln(stderr, "NIY: doctor is not implemented yet")
		return 6
	default:
		if strings.HasPrefix(first, "--") {
			return validateGlobalOption(args, stderr)
		}
		fmt.Fprintf(stderr, "unknown command: %s\n", first)
		return 2
	}
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
