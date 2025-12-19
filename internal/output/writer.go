package output

import (
	"io"
	"os"
)

// OutputWriter is the interface for rendering build output.
// Different implementations provide different output styles.
type OutputWriter interface {
	// WriteEvent renders an output event to the output stream.
	WriteEvent(event OutputEvent)

	// Flush ensures all buffered output is written.
	Flush()
}

// WriterConfig configures output writer behavior.
type WriterConfig struct {
	// Verbose enables detailed output (variable evaluation, staleness checks)
	Verbose bool

	// Quiet suppresses non-error output
	Quiet bool

	// Color controls color output: "auto", "always", "never"
	Color string

	// LogLevel sets minimum log level for headless mode: "debug", "info", "warn", "error"
	LogLevel string

	// LogFormat sets log format for headless mode: "text", "json"
	LogFormat string
}

// DefaultWriterConfig returns a config with sensible defaults.
func DefaultWriterConfig() WriterConfig {
	return WriterConfig{
		Verbose:   false,
		Quiet:     false,
		Color:     "auto",
		LogLevel:  "info",
		LogFormat: "text",
	}
}

// NewWriter creates an OutputWriter for the given mode and output stream.
func NewWriter(mode OutputMode, w io.Writer, config WriterConfig) OutputWriter {
	switch mode {
	case ModeCLI:
		return NewCLIWriter(w, config)
	case ModeTUI:
		return NewTUIWriter(w)
	case ModeHeadless:
		return NewHeadlessWriter(w, config)
	default:
		return NewCLIWriter(w, config)
	}
}

// NewDefaultWriter creates an OutputWriter with automatic mode detection.
func NewDefaultWriter(config WriterConfig) OutputWriter {
	mode := DetectOutputMode()
	return NewWriter(mode, os.Stdout, config)
}

// noOpWriter is a writer that discards all output.
type noOpWriter struct{}

func (w *noOpWriter) WriteEvent(event OutputEvent) {}
func (w *noOpWriter) Flush()                       {}

// NewNoOpWriter creates a writer that discards all output.
// Useful for testing or when output should be suppressed.
func NewNoOpWriter() OutputWriter {
	return &noOpWriter{}
}
