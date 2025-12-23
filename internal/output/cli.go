package output

import (
	"fmt"
	"io"
)

// CLIWriter implements OutputWriter for interactive terminal output.
// It provides colored output, progress indicators, and formatted messages.
type CLIWriter struct {
	w          io.Writer
	config     WriterConfig
	useColor   bool
	useUnicode bool
}

// NewCLIWriter creates a new CLIWriter that writes to w.
func NewCLIWriter(w io.Writer, config WriterConfig) *CLIWriter {
	return &CLIWriter{
		w:          w,
		config:     config,
		useColor:   ShouldUseColor(config.Color),
		useUnicode: ShouldUseUnicode(config.Unicode),
	}
}

// ShouldUseUnicode determines if Unicode should be used based on config.
func ShouldUseUnicode(unicodeSetting string) bool {
	switch unicodeSetting {
	case "always":
		return true
	case "never":
		return false
	default: // "auto"
		return DetectUnicodeSupport()
	}
}

// Symbols for output - Unicode and ASCII fallbacks.
type symbols struct {
	checkMark   string
	crossMark   string
	arrow       string
	bullet      string
	rightArrow  string
	progressBar string
}

var (
	unicodeSymbols = symbols{
		checkMark:   "✓",
		crossMark:   "✗",
		arrow:       "→",
		bullet:      "•",
		rightArrow:  "➜",
		progressBar: "▓",
	}
	asciiSymbols = symbols{
		checkMark:   "[ok]",
		crossMark:   "[FAIL]",
		arrow:       "->",
		bullet:      "*",
		rightArrow:  "->",
		progressBar: "#",
	}
)

// getSymbols returns the appropriate symbol set based on Unicode support.
func (c *CLIWriter) getSymbols() symbols {
	if c.useUnicode {
		return unicodeSymbols
	}
	return asciiSymbols
}

// WriteEvent renders an output event to the terminal.
func (c *CLIWriter) WriteEvent(event OutputEvent) {
	if c.config.Quiet && !isErrorEvent(event) {
		return
	}

	switch e := event.(type) {
	case PhaseStarted:
		c.writePhaseStarted(e)
	case PhaseCompleted:
		// Only show in verbose mode
		if c.config.Verbose {
			c.writePhaseCompleted(e)
		}
	case VariableEvaluated:
		if c.config.Verbose {
			c.writeVariableEvaluated(e)
		}
	case TargetStarted:
		if c.config.Verbose {
			c.writeTargetStarted(e)
		}
	case TargetCompleted:
		if c.config.Verbose || !e.Success {
			c.writeTargetCompleted(e)
		}
	case TargetSkipped:
		if c.config.Verbose {
			c.writeTargetSkipped(e)
		}
	case CommandStarted:
		c.writeCommandStarted(e)
	case CommandOutput:
		c.writeCommandOutput(e)
	case CommandCompleted:
		// Only show duration in verbose mode if there was no error
	case StalenessChecked:
		if c.config.Verbose {
			c.writeStalenessChecked(e)
		}
	case BuildSummary:
		if c.config.Verbose || e.Failed > 0 {
			c.writeBuildSummary(e)
		}
	case ErrorOccurred:
		c.writeError(e)
	case DryRunTarget:
		c.writeDryRunTarget(e)
	case DryRunCommand:
		c.writeDryRunCommand(e)
	}
}

// Flush ensures all output is written.
func (c *CLIWriter) Flush() {
	// No buffering in this implementation
}

func (c *CLIWriter) writePhaseStarted(e PhaseStarted) {
	if c.config.Verbose {
		switch e.Phase {
		case "eval":
			fmt.Fprintln(c.w, "Evaluating variables...")
		case "plan":
			fmt.Fprintln(c.w, "\nChecking targets...")
		}
	}
}

func (c *CLIWriter) writePhaseCompleted(e PhaseCompleted) {
	// Currently no output for phase completion in verbose mode
}

func (c *CLIWriter) writeVariableEvaluated(e VariableEvaluated) {
	if e.Expr != "" {
		fmt.Fprintf(c.w, "  %s = %s → %s\n", e.Name, e.Expr, e.Result)
	} else {
		fmt.Fprintf(c.w, "  %s → %s\n", e.Name, e.Result)
	}
}

func (c *CLIWriter) writeTargetStarted(e TargetStarted) {
	target := Colorize(e.Target, ColorBoldCyan, c.useColor)
	if e.Total > 1 {
		progress := Colorize(fmt.Sprintf("[%d/%d]", e.Index, e.Total), ColorBlue, c.useColor)
		fmt.Fprintf(c.w, "%s Building %s\n", progress, target)
	} else {
		fmt.Fprintf(c.w, "Building %s\n", target)
	}
}

func (c *CLIWriter) writeTargetCompleted(e TargetCompleted) {
	if e.Success {
		status := Colorize("Built", ColorGreen, c.useColor)
		if c.config.Verbose && e.Duration > 0 {
			fmt.Fprintf(c.w, "%s %s (%s)\n", status, e.Target, e.Duration)
		} else {
			fmt.Fprintf(c.w, "%s %s\n", status, e.Target)
		}
	} else {
		status := Colorize("FAILED", ColorBoldRed, c.useColor)
		fmt.Fprintf(c.w, "%s %s: %s\n", status, e.Target, e.Error)
	}
}

func (c *CLIWriter) writeTargetSkipped(e TargetSkipped) {
	target := Colorize(e.Target, ColorCyan, c.useColor)
	reason := Dim(e.Reason, c.useColor)
	fmt.Fprintf(c.w, "%s is %s\n", target, reason)
}

func (c *CLIWriter) writeCommandStarted(e CommandStarted) {
	// In verbose mode, indent commands under "Building X"
	// Show commands in bold cyan to distinguish from output (which is dimmed)
	if c.config.Verbose {
		cmd := Colorize("  "+e.Command, ColorBoldCyan, c.useColor)
		fmt.Fprintln(c.w, cmd)
	} else {
		fmt.Fprintln(c.w, Colorize(e.Command, ColorBoldCyan, c.useColor))
	}
}

func (c *CLIWriter) writeCommandOutput(e CommandOutput) {
	// Command output is dimmed (gray) to distinguish from build system messages
	if e.Stdout != "" {
		output := Dim(e.Stdout, c.useColor)
		fmt.Fprint(c.w, output)
		if len(e.Stdout) > 0 && e.Stdout[len(e.Stdout)-1] != '\n' {
			fmt.Fprintln(c.w)
		}
	}
	if e.Stderr != "" {
		output := Dim(e.Stderr, c.useColor)
		fmt.Fprint(c.w, output)
		if len(e.Stderr) > 0 && e.Stderr[len(e.Stderr)-1] != '\n' {
			fmt.Fprintln(c.w)
		}
	}
}

func (c *CLIWriter) writeStalenessChecked(e StalenessChecked) {
	var action string
	if e.Action == "rebuild" {
		action = Colorize("rebuild", ColorYellow, c.useColor)
	} else {
		action = Colorize("skip", ColorGreen, c.useColor)
	}
	fmt.Fprintf(c.w, "  %s: %s → %s\n", e.Target, e.Reason, action)
}

func (c *CLIWriter) writeBuildSummary(e BuildSummary) {
	fmt.Fprintln(c.w)
	if e.Failed == 0 {
		status := Colorize("Build success", ColorGreen, c.useColor)
		if e.Total == 1 {
			fmt.Fprintf(c.w, "%s: 1 target built", status)
		} else {
			fmt.Fprintf(c.w, "%s: %d targets built", status, e.Succeeded)
		}
		if e.Duration > 0 {
			fmt.Fprintf(c.w, " (%s)", e.Duration)
		}
		fmt.Fprintln(c.w)
	} else {
		status := Colorize("Build failed", ColorBoldRed, c.useColor)
		fmt.Fprintf(c.w, "%s: %d of %d targets failed", status, e.Failed, e.Total)
		if e.Duration > 0 {
			fmt.Fprintf(c.w, " (%s)", e.Duration)
		}
		fmt.Fprintln(c.w)
	}
}

func (c *CLIWriter) writeError(e ErrorOccurred) {
	// Error header with code
	errLabel := Colorize("error", ColorBoldRed, c.useColor)
	code := Colorize(fmt.Sprintf("[%s]", e.Code), ColorRed, c.useColor)
	fmt.Fprintf(c.w, "%s%s: %s\n", errLabel, code, e.Message)

	// Location
	if e.Location != "" {
		locLabel := Colorize(" --> ", ColorBlue, c.useColor)
		fmt.Fprintf(c.w, "%s%s\n", locLabel, e.Location)
	}

	// Source context
	if e.Context != "" {
		fmt.Fprintln(c.w, e.Context)
	}

	// Hint
	if e.Hint != "" {
		hintLabel := Colorize("help: ", ColorGreen, c.useColor)
		fmt.Fprintf(c.w, "%s%s\n", hintLabel, e.Hint)
	}
}

func (c *CLIWriter) writeDryRunTarget(e DryRunTarget) {
	target := Colorize(e.Target, ColorBoldCyan, c.useColor)
	fmt.Fprintf(c.w, "Would build: %s\n", target)
}

func (c *CLIWriter) writeDryRunCommand(e DryRunCommand) {
	cmd := Dim("  "+e.Command, c.useColor)
	fmt.Fprintln(c.w, cmd)
}

// isErrorEvent returns true if the event is an error.
func isErrorEvent(event OutputEvent) bool {
	_, ok := event.(ErrorOccurred)
	return ok
}
