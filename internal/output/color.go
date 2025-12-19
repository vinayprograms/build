package output

import (
	"os"
)

// Color represents an ANSI color code.
type Color string

// ANSI color codes for terminal output.
const (
	ColorReset   Color = "\033[0m"
	ColorBold    Color = "\033[1m"
	ColorDim     Color = "\033[2m"
	ColorRed     Color = "\033[31m"
	ColorGreen   Color = "\033[32m"
	ColorYellow  Color = "\033[33m"
	ColorBlue    Color = "\033[34m"
	ColorMagenta Color = "\033[35m"
	ColorCyan    Color = "\033[36m"
	ColorWhite   Color = "\033[37m"
	ColorGray    Color = "\033[90m"

	// Bold variants
	ColorBoldRed   Color = "\033[1;31m"
	ColorBoldGreen Color = "\033[1;32m"
	ColorBoldCyan  Color = "\033[1;36m"
)

// ColorConfig determines how colors are applied.
type ColorConfig struct {
	Enabled bool
}

// ShouldUseColor determines if colors should be used based on config.
func ShouldUseColor(colorSetting string) bool {
	switch colorSetting {
	case "always":
		return true
	case "never":
		return false
	default: // "auto"
		return shouldAutoColor()
	}
}

// shouldAutoColor returns true if colors should be enabled automatically.
func shouldAutoColor() bool {
	// Check NO_COLOR environment variable (standard)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check FORCE_COLOR environment variable
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Check if stdout is a terminal
	if !isTerminal(os.Stdout) {
		return false
	}

	// Check for dumb terminal
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	return true
}

// Colorize wraps text with color codes if colors are enabled.
func Colorize(text string, color Color, enabled bool) string {
	if !enabled {
		return text
	}
	return string(color) + text + string(ColorReset)
}

// Bold makes text bold if colors are enabled.
func Bold(text string, enabled bool) string {
	if !enabled {
		return text
	}
	return string(ColorBold) + text + string(ColorReset)
}

// Dim makes text dim if colors are enabled.
func Dim(text string, enabled bool) string {
	if !enabled {
		return text
	}
	return string(ColorDim) + text + string(ColorReset)
}

// ColorizeStatus colors text based on success/failure.
func ColorizeStatus(text string, success bool, enabled bool) string {
	if success {
		return Colorize(text, ColorGreen, enabled)
	}
	return Colorize(text, ColorBoldRed, enabled)
}
