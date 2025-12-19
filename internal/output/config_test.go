package output

import (
	"testing"
)

// TestWriterConfigFromFlags tests creating WriterConfig from CLI flag values.
func TestWriterConfigFromFlags(t *testing.T) {
	tests := []struct {
		name     string
		verbose  bool
		quiet    bool
		color    string
		expected WriterConfig
	}{
		{
			name:    "default values",
			verbose: false,
			quiet:   false,
			color:   "auto",
			expected: WriterConfig{
				Verbose:   false,
				Quiet:     false,
				Color:     "auto",
				Unicode:   "auto",
				LogLevel:  "info",
				LogFormat: "text",
			},
		},
		{
			name:    "verbose mode",
			verbose: true,
			quiet:   false,
			color:   "auto",
			expected: WriterConfig{
				Verbose:   true,
				Quiet:     false,
				Color:     "auto",
				Unicode:   "auto",
				LogLevel:  "info",
				LogFormat: "text",
			},
		},
		{
			name:    "quiet mode",
			verbose: false,
			quiet:   true,
			color:   "auto",
			expected: WriterConfig{
				Verbose:   false,
				Quiet:     true,
				Color:     "auto",
				Unicode:   "auto",
				LogLevel:  "info",
				LogFormat: "text",
			},
		},
		{
			name:    "color always",
			verbose: false,
			quiet:   false,
			color:   "always",
			expected: WriterConfig{
				Verbose:   false,
				Quiet:     false,
				Color:     "always",
				Unicode:   "auto",
				LogLevel:  "info",
				LogFormat: "text",
			},
		},
		{
			name:    "color never",
			verbose: false,
			quiet:   false,
			color:   "never",
			expected: WriterConfig{
				Verbose:   false,
				Quiet:     false,
				Color:     "never",
				Unicode:   "auto",
				LogLevel:  "info",
				LogFormat: "text",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewWriterConfigFromFlags(tt.verbose, tt.quiet, tt.color)
			if config.Verbose != tt.expected.Verbose {
				t.Errorf("Verbose: expected %v, got %v", tt.expected.Verbose, config.Verbose)
			}
			if config.Quiet != tt.expected.Quiet {
				t.Errorf("Quiet: expected %v, got %v", tt.expected.Quiet, config.Quiet)
			}
			if config.Color != tt.expected.Color {
				t.Errorf("Color: expected %v, got %v", tt.expected.Color, config.Color)
			}
		})
	}
}

// TestNewWriterFromFlags tests creating an OutputWriter from CLI flags.
func TestNewWriterFromFlags(t *testing.T) {
	// Test that we get the right writer type based on mode
	tests := []struct {
		name     string
		mode     OutputMode
		verbose  bool
		quiet    bool
		color    string
		wantType string
	}{
		{
			name:     "CLI mode",
			mode:     ModeCLI,
			verbose:  false,
			quiet:    false,
			color:    "auto",
			wantType: "*output.CLIWriter",
		},
		{
			name:     "headless mode",
			mode:     ModeHeadless,
			verbose:  false,
			quiet:    false,
			color:    "auto",
			wantType: "*output.HeadlessWriter",
		},
		{
			name:     "TUI mode",
			mode:     ModeTUI,
			verbose:  false,
			quiet:    false,
			color:    "auto",
			wantType: "*output.TUIWriter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewWriterConfigFromFlags(tt.verbose, tt.quiet, tt.color)
			writer := NewWriterWithMode(tt.mode, config)
			typeName := getTypeName(writer)
			if typeName != tt.wantType {
				t.Errorf("expected type %s, got %s", tt.wantType, typeName)
			}
		})
	}
}

// TestQuietModeOverridesVerbose tests that quiet mode takes precedence over verbose.
func TestQuietModeOverridesVerbose(t *testing.T) {
	// When both quiet and verbose are set, quiet should suppress output
	config := NewWriterConfigFromFlags(true, true, "auto")
	if !config.Quiet {
		t.Error("quiet should be true")
	}
	// Verbose is still true in config, but Quiet takes precedence at render time
	if !config.Verbose {
		t.Error("verbose should still be true in config (quiet handles suppression)")
	}
}

// getTypeName returns the type name for comparison.
func getTypeName(v interface{}) string {
	switch v.(type) {
	case *CLIWriter:
		return "*output.CLIWriter"
	case *HeadlessWriter:
		return "*output.HeadlessWriter"
	case *TUIWriter:
		return "*output.TUIWriter"
	default:
		return "unknown"
	}
}
