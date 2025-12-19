package output

import (
	"os"
	"testing"
)

// TestColorLevel tests color level detection.
func TestColorLevel(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected ColorLevel
	}{
		{
			name:     "no_color_env_disables_color",
			envVars:  map[string]string{"NO_COLOR": "1"},
			expected: ColorLevelNone,
		},
		{
			name:     "term_dumb_disables_color",
			envVars:  map[string]string{"TERM": "dumb"},
			expected: ColorLevelNone,
		},
		{
			name:     "force_color_enables_basic",
			envVars:  map[string]string{"FORCE_COLOR": "1"},
			expected: ColorLevelBasic,
		},
		{
			name:     "colorterm_truecolor",
			envVars:  map[string]string{"COLORTERM": "truecolor"},
			expected: ColorLevelTruecolor,
		},
		{
			name:     "colorterm_24bit",
			envVars:  map[string]string{"COLORTERM": "24bit"},
			expected: ColorLevelTruecolor,
		},
		{
			name:     "term_256color",
			envVars:  map[string]string{"TERM": "xterm-256color"},
			expected: ColorLevel256,
		},
		{
			name:     "term_screen_256color",
			envVars:  map[string]string{"TERM": "screen-256color"},
			expected: ColorLevel256,
		},
		{
			name:     "term_xterm_basic",
			envVars:  map[string]string{"TERM": "xterm"},
			expected: ColorLevelBasic,
		},
		{
			name:     "no_color_takes_precedence",
			envVars:  map[string]string{"NO_COLOR": "1", "COLORTERM": "truecolor"},
			expected: ColorLevelNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and clear environment
			saved := make(map[string]string)
			for _, key := range []string{"NO_COLOR", "FORCE_COLOR", "TERM", "COLORTERM"} {
				saved[key] = os.Getenv(key)
				os.Unsetenv(key)
			}

			// Set test environment
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Test
			got := DetectColorLevel()
			if got != tt.expected {
				t.Errorf("DetectColorLevel() = %v, want %v", got, tt.expected)
			}

			// Restore environment
			for k, v := range saved {
				if v != "" {
					os.Setenv(k, v)
				} else {
					os.Unsetenv(k)
				}
			}
		})
	}
}

// TestColorLevelString tests ColorLevel.String() method.
func TestColorLevelString(t *testing.T) {
	tests := []struct {
		level    ColorLevel
		expected string
	}{
		{ColorLevelNone, "none"},
		{ColorLevelBasic, "basic"},
		{ColorLevel256, "256"},
		{ColorLevelTruecolor, "truecolor"},
		{ColorLevel(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("ColorLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestTerminalCapabilities tests the TerminalCapabilities struct.
func TestTerminalCapabilities(t *testing.T) {
	t.Run("default_capabilities", func(t *testing.T) {
		caps := DetectCapabilities()
		// Just verify it returns something without crashing
		if caps == nil {
			t.Error("DetectCapabilities() returned nil")
		}
	})

	t.Run("supports_color", func(t *testing.T) {
		caps := &TerminalCapabilities{
			ColorLevel: ColorLevelBasic,
		}
		if !caps.SupportsColor() {
			t.Error("ColorLevelBasic should support color")
		}

		capsNone := &TerminalCapabilities{
			ColorLevel: ColorLevelNone,
		}
		if capsNone.SupportsColor() {
			t.Error("ColorLevelNone should not support color")
		}
	})

	t.Run("supports_256_color", func(t *testing.T) {
		caps := &TerminalCapabilities{
			ColorLevel: ColorLevel256,
		}
		if !caps.Supports256Color() {
			t.Error("ColorLevel256 should support 256 colors")
		}

		capsBasic := &TerminalCapabilities{
			ColorLevel: ColorLevelBasic,
		}
		if capsBasic.Supports256Color() {
			t.Error("ColorLevelBasic should not support 256 colors")
		}
	})

	t.Run("supports_truecolor", func(t *testing.T) {
		caps := &TerminalCapabilities{
			ColorLevel: ColorLevelTruecolor,
		}
		if !caps.SupportsTruecolor() {
			t.Error("ColorLevelTruecolor should support truecolor")
		}

		caps256 := &TerminalCapabilities{
			ColorLevel: ColorLevel256,
		}
		if caps256.SupportsTruecolor() {
			t.Error("ColorLevel256 should not support truecolor")
		}
	})
}

// TestUnicodeDetection tests unicode support detection.
func TestUnicodeDetection(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name:     "utf8_in_lang",
			envVars:  map[string]string{"LANG": "en_US.UTF-8"},
			expected: true,
		},
		{
			name:     "utf8_in_lc_all",
			envVars:  map[string]string{"LC_ALL": "en_US.UTF-8"},
			expected: true,
		},
		{
			name:     "utf8_in_lc_ctype",
			envVars:  map[string]string{"LC_CTYPE": "en_US.UTF-8"},
			expected: true,
		},
		{
			name:     "utf8_lowercase",
			envVars:  map[string]string{"LANG": "en_US.utf8"},
			expected: true,
		},
		{
			name:     "no_utf8",
			envVars:  map[string]string{"LANG": "C"},
			expected: false,
		},
		{
			name:     "posix_locale",
			envVars:  map[string]string{"LANG": "POSIX"},
			expected: false,
		},
		{
			name:     "empty_locale",
			envVars:  map[string]string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and clear environment
			saved := make(map[string]string)
			for _, key := range []string{"LANG", "LC_ALL", "LC_CTYPE"} {
				saved[key] = os.Getenv(key)
				os.Unsetenv(key)
			}

			// Set test environment
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Test
			got := DetectUnicodeSupport()
			if got != tt.expected {
				t.Errorf("DetectUnicodeSupport() = %v, want %v", got, tt.expected)
			}

			// Restore environment
			for k, v := range saved {
				if v != "" {
					os.Setenv(k, v)
				} else {
					os.Unsetenv(k)
				}
			}
		})
	}
}

// TestTerminalSize tests terminal size detection.
func TestTerminalSize(t *testing.T) {
	t.Run("returns_non_negative", func(t *testing.T) {
		width, height := GetTerminalSize()
		// Should return non-negative values (0 if detection fails)
		if width < 0 || height < 0 {
			t.Errorf("GetTerminalSize() returned negative: %d, %d", width, height)
		}
	})

	t.Run("default_size", func(t *testing.T) {
		width, height := DefaultTerminalSize()
		if width != 80 {
			t.Errorf("DefaultTerminalSize() width = %d, want 80", width)
		}
		if height != 24 {
			t.Errorf("DefaultTerminalSize() height = %d, want 24", height)
		}
	})
}

// TestCapabilitiesWithFallback tests fallback behavior.
func TestCapabilitiesWithFallback(t *testing.T) {
	t.Run("size_with_fallback", func(t *testing.T) {
		caps := &TerminalCapabilities{
			Width:  0,
			Height: 0,
		}
		width, height := caps.SizeWithFallback()
		if width != 80 {
			t.Errorf("SizeWithFallback() width = %d, want 80", width)
		}
		if height != 24 {
			t.Errorf("SizeWithFallback() height = %d, want 24", height)
		}
	})

	t.Run("size_actual", func(t *testing.T) {
		caps := &TerminalCapabilities{
			Width:  120,
			Height: 40,
		}
		width, height := caps.SizeWithFallback()
		if width != 120 {
			t.Errorf("SizeWithFallback() width = %d, want 120", width)
		}
		if height != 40 {
			t.Errorf("SizeWithFallback() height = %d, want 40", height)
		}
	})
}
