package output

import (
	"os"
	"testing"
)

func TestOutputModeString(t *testing.T) {
	tests := []struct {
		mode     OutputMode
		expected string
	}{
		{ModeCLI, "cli"},
		{ModeTUI, "tui"},
		{ModeHeadless, "headless"},
		{OutputMode(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestParseOutputMode(t *testing.T) {
	tests := []struct {
		input    string
		expected OutputMode
	}{
		{"cli", ModeCLI},
		{"tui", ModeTUI},
		{"headless", ModeHeadless},
		{"unknown", ModeCLI},
		{"", ModeCLI},
		{"CLI", ModeCLI}, // case-sensitive, unknown falls to default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseOutputMode(tt.input)
			if got != tt.expected {
				t.Errorf("ParseOutputMode(%q): expected %v, got %v", tt.input, tt.expected, got)
			}
		})
	}
}

func TestDetectOutputMode_EnvOverride(t *testing.T) {
	tests := []struct {
		envValue string
		expected OutputMode
	}{
		{"cli", ModeCLI},
		{"tui", ModeTUI},
		{"headless", ModeHeadless},
	}

	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			// Save and restore environment
			oldValue := os.Getenv("BUILD_OUTPUT_MODE")
			defer os.Setenv("BUILD_OUTPUT_MODE", oldValue)

			os.Setenv("BUILD_OUTPUT_MODE", tt.envValue)
			got := DetectOutputMode()
			if got != tt.expected {
				t.Errorf("with BUILD_OUTPUT_MODE=%s: expected %v, got %v", tt.envValue, tt.expected, got)
			}
		})
	}
}

func TestDetectOutputMode_CI(t *testing.T) {
	ciVars := []string{
		"CI",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"JENKINS_URL",
		"CIRCLECI",
	}

	for _, ciVar := range ciVars {
		t.Run(ciVar, func(t *testing.T) {
			// Save and restore environment
			oldValue := os.Getenv(ciVar)
			oldModeValue := os.Getenv("BUILD_OUTPUT_MODE")
			defer func() {
				os.Setenv(ciVar, oldValue)
				os.Setenv("BUILD_OUTPUT_MODE", oldModeValue)
			}()

			os.Unsetenv("BUILD_OUTPUT_MODE")
			os.Setenv(ciVar, "true")
			got := DetectOutputMode()
			if got != ModeHeadless {
				t.Errorf("with %s=true: expected ModeHeadless, got %v", ciVar, got)
			}
		})
	}
}

// TestDetectOutputMode_DumbTerminal verifies TERM=dumb, by itself, does NOT
// switch to headless (timestamped log-line) output (C5). Color/progress
// suppression for a dumb terminal is handled separately by
// shouldAutoColor/ShouldUseColor, not by the output mode.
func TestDetectOutputMode_DumbTerminal(t *testing.T) {
	// Save and restore environment
	oldTerm := os.Getenv("TERM")
	oldMode := os.Getenv("BUILD_OUTPUT_MODE")
	oldCI := os.Getenv("CI")
	defer func() {
		os.Setenv("TERM", oldTerm)
		os.Setenv("BUILD_OUTPUT_MODE", oldMode)
		os.Setenv("CI", oldCI)
	}()

	os.Unsetenv("BUILD_OUTPUT_MODE")
	os.Unsetenv("CI")
	os.Setenv("TERM", "dumb")

	got := DetectOutputMode()
	if got != ModeCLI {
		t.Errorf("with TERM=dumb: expected ModeCLI, got %v", got)
	}
}

// TestDetectOutputMode_NonTTYWithoutCI verifies that a plain non-TTY pipe
// (e.g. `build -n | cat`), with no CI indicator and no explicit
// BUILD_OUTPUT_MODE, stays ModeCLI rather than falling back to ModeHeadless
// (C5). Test binaries themselves normally run with stdout not attached to a
// TTY, so this also guards against a regression to the old
// isTerminal()-based check without needing to fake a TTY.
func TestDetectOutputMode_NonTTYWithoutCI(t *testing.T) {
	oldMode := os.Getenv("BUILD_OUTPUT_MODE")
	oldCI := os.Getenv("CI")
	oldTerm := os.Getenv("TERM")
	defer func() {
		os.Setenv("BUILD_OUTPUT_MODE", oldMode)
		os.Setenv("CI", oldCI)
		os.Setenv("TERM", oldTerm)
	}()

	os.Unsetenv("BUILD_OUTPUT_MODE")
	os.Unsetenv("CI")
	os.Setenv("TERM", "xterm-256color")

	got := DetectOutputMode()
	if got != ModeCLI {
		t.Errorf("with non-TTY stdout and no CI indicator: expected ModeCLI, got %v", got)
	}
}

func TestIsCI(t *testing.T) {
	// Test with no CI vars set
	t.Run("no CI", func(t *testing.T) {
		// Clear all CI vars
		ciVars := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL", "CIRCLECI"}
		oldValues := make(map[string]string)
		for _, v := range ciVars {
			oldValues[v] = os.Getenv(v)
			os.Unsetenv(v)
		}
		defer func() {
			for k, v := range oldValues {
				if v != "" {
					os.Setenv(k, v)
				}
			}
		}()

		if isCI() {
			t.Error("expected isCI() = false when no CI vars set")
		}
	})
}
