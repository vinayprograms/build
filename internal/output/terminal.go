package output

import (
	"os"
	"strings"

	"golang.org/x/term"
)

// ColorLevel represents the level of color support.
type ColorLevel int

const (
	// ColorLevelNone indicates no color support.
	ColorLevelNone ColorLevel = iota
	// ColorLevelBasic indicates 16-color support (standard ANSI).
	ColorLevelBasic
	// ColorLevel256 indicates 256-color support.
	ColorLevel256
	// ColorLevelTruecolor indicates 24-bit truecolor support.
	ColorLevelTruecolor
)

// String returns a human-readable name for the color level.
func (c ColorLevel) String() string {
	switch c {
	case ColorLevelNone:
		return "none"
	case ColorLevelBasic:
		return "basic"
	case ColorLevel256:
		return "256"
	case ColorLevelTruecolor:
		return "truecolor"
	default:
		return "unknown"
	}
}

// TerminalCapabilities holds detected terminal capabilities.
type TerminalCapabilities struct {
	Width       int        // Terminal width in columns (0 if unknown)
	Height      int        // Terminal height in rows (0 if unknown)
	ColorLevel  ColorLevel // Supported color level
	Unicode     bool       // Unicode support
	Interactive bool       // Is an interactive terminal
}

// DefaultTerminalSize returns the default terminal dimensions.
func DefaultTerminalSize() (width, height int) {
	return 80, 24
}

// SupportsColor returns true if the terminal supports any color.
func (t *TerminalCapabilities) SupportsColor() bool {
	return t.ColorLevel >= ColorLevelBasic
}

// Supports256Color returns true if the terminal supports 256 colors.
func (t *TerminalCapabilities) Supports256Color() bool {
	return t.ColorLevel >= ColorLevel256
}

// SupportsTruecolor returns true if the terminal supports 24-bit color.
func (t *TerminalCapabilities) SupportsTruecolor() bool {
	return t.ColorLevel >= ColorLevelTruecolor
}

// SizeWithFallback returns the terminal size, using defaults if unknown.
func (t *TerminalCapabilities) SizeWithFallback() (width, height int) {
	width, height = t.Width, t.Height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	return
}

// DetectCapabilities detects terminal capabilities for stdout.
func DetectCapabilities() *TerminalCapabilities {
	caps := &TerminalCapabilities{
		ColorLevel:  DetectColorLevel(),
		Unicode:     DetectUnicodeSupport(),
		Interactive: isTerminal(os.Stdout),
	}
	caps.Width, caps.Height = GetTerminalSize()
	return caps
}

// DetectColorLevel detects the color level supported by the terminal.
func DetectColorLevel() ColorLevel {
	// NO_COLOR takes highest precedence
	if os.Getenv("NO_COLOR") != "" {
		return ColorLevelNone
	}

	// Dumb terminal
	if os.Getenv("TERM") == "dumb" {
		return ColorLevelNone
	}

	// FORCE_COLOR enables at least basic colors
	if os.Getenv("FORCE_COLOR") != "" {
		// Check if truecolor or 256 is also available
		colorterm := os.Getenv("COLORTERM")
		if colorterm == "truecolor" || colorterm == "24bit" {
			return ColorLevelTruecolor
		}
		term := os.Getenv("TERM")
		if strings.Contains(term, "256color") {
			return ColorLevel256
		}
		return ColorLevelBasic
	}

	// Check COLORTERM for truecolor support
	colorterm := os.Getenv("COLORTERM")
	if colorterm == "truecolor" || colorterm == "24bit" {
		return ColorLevelTruecolor
	}

	// Check TERM for 256 color support
	term := os.Getenv("TERM")
	if strings.Contains(term, "256color") {
		return ColorLevel256
	}

	// Check TERM for basic color support
	colorTerms := []string{
		"xterm", "vt100", "vt220", "screen", "rxvt",
		"linux", "cygwin", "ansi", "color",
	}
	for _, ct := range colorTerms {
		if strings.Contains(term, ct) {
			return ColorLevelBasic
		}
	}

	// No color support detected
	return ColorLevelNone
}

// DetectUnicodeSupport detects if the terminal supports Unicode.
func DetectUnicodeSupport() bool {
	// Check locale environment variables for UTF-8
	localeVars := []string{"LC_ALL", "LC_CTYPE", "LANG"}
	for _, v := range localeVars {
		locale := os.Getenv(v)
		if locale != "" {
			upper := strings.ToUpper(locale)
			if strings.Contains(upper, "UTF-8") || strings.Contains(upper, "UTF8") {
				return true
			}
		}
	}
	return false
}

// GetTerminalSize returns the current terminal dimensions.
// Returns (0, 0) if detection fails or stdout is not a terminal.
func GetTerminalSize() (width, height int) {
	fd := int(os.Stdout.Fd())
	w, h, err := term.GetSize(fd)
	if err != nil {
		return 0, 0
	}
	return w, h
}
