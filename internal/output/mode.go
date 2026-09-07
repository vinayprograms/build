package output

import (
	"os"
)

// OutputMode represents the output context the build tool is running in.
type OutputMode int

const (
	// ModeCLI is for interactive terminal output with colors and formatting.
	ModeCLI OutputMode = iota
	// ModeTUI is for structured output consumed by a terminal UI.
	ModeTUI
	// ModeHeadless is for CI/CD pipelines and log collectors.
	ModeHeadless
)

// String returns a human-readable name for the mode.
func (m OutputMode) String() string {
	switch m {
	case ModeCLI:
		return "cli"
	case ModeTUI:
		return "tui"
	case ModeHeadless:
		return "headless"
	default:
		return "unknown"
	}
}

// ParseOutputMode parses a mode string into OutputMode.
// Returns ModeCLI for unrecognized values.
func ParseOutputMode(s string) OutputMode {
	switch s {
	case "cli":
		return ModeCLI
	case "tui":
		return ModeTUI
	case "headless":
		return ModeHeadless
	default:
		return ModeCLI
	}
}

// DetectOutputMode determines the appropriate output mode based on:
// 1. NEED_OUTPUT_MODE environment variable (if set)
// 2. CI environment indicators
//
// A plain non-TTY pipe (e.g. `build -n | cat`) and TERM=dumb are NOT reasons
// to switch to headless (timestamped log-line) output on their own - they
// still use ModeCLI, whose CLIWriter separately auto-disables color and
// progress indicators when stdout isn't a terminal (see ShouldUseColor /
// shouldAutoColor in color.go). Headless mode is reserved for cases that
// actually want structured/logged output: an explicit NEED_OUTPUT_MODE, or
// running under a recognized CI system.
func DetectOutputMode() OutputMode {
	// Check for explicit override
	if mode := os.Getenv("NEED_OUTPUT_MODE"); mode != "" {
		return ParseOutputMode(mode)
	}

	// Check for CI environment indicators
	if isCI() {
		return ModeHeadless
	}

	return ModeCLI
}

// isCI returns true if running in a CI environment.
func isCI() bool {
	ciIndicators := []string{
		"CI",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"JENKINS_URL",
		"CIRCLECI",
		"TRAVIS",
		"BUILDKITE",
		"DRONE",
		"TEAMCITY_VERSION",
		"TF_BUILD",           // Azure Pipelines
		"CODEBUILD_BUILD_ID", // AWS CodeBuild
	}

	for _, indicator := range ciIndicators {
		if os.Getenv(indicator) != "" {
			return true
		}
	}

	return false
}

// isTerminal returns true if the given file is a terminal.
// This is a simplified check that works on Unix-like systems.
// On Windows, additional checks may be needed.
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	mode := stat.Mode()
	// Check if it's a character device (terminal)
	return mode&os.ModeCharDevice != 0
}
