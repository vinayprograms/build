package output

import "time"

// OutputEvent is the interface that all output events implement.
// Events are emitted by the build pipeline and rendered by OutputWriters.
type OutputEvent interface {
	eventType() string
}

// PhaseStarted is emitted when a build phase begins.
type PhaseStarted struct {
	Phase     string    // "parse", "semantic", "eval", "plan", "execute"
	Timestamp time.Time // When the phase started
}

func (e PhaseStarted) eventType() string { return "phase_started" }

// PhaseCompleted is emitted when a build phase finishes.
type PhaseCompleted struct {
	Phase     string        // "parse", "semantic", "eval", "plan", "execute"
	Timestamp time.Time     // When the phase completed
	Duration  time.Duration // How long the phase took
}

func (e PhaseCompleted) eventType() string { return "phase_completed" }

// VariableEvaluated is emitted when a variable is evaluated (verbose mode).
type VariableEvaluated struct {
	Name   string // Variable name
	Expr   string // Original expression (if function call, empty for simple values)
	Result string // Evaluated value
}

func (e VariableEvaluated) eventType() string { return "variable_evaluated" }

// TargetStarted is emitted when a target build begins.
type TargetStarted struct {
	Target string // Target path or name
	Index  int    // 1-based index for progress (e.g., 3 of 10)
	Total  int    // Total number of targets
}

func (e TargetStarted) eventType() string { return "target_started" }

// TargetCompleted is emitted when a target build finishes.
type TargetCompleted struct {
	Target   string        // Target path or name
	Success  bool          // Whether the build succeeded
	Duration time.Duration // How long the build took
	Error    string        // Error message if Success is false
}

func (e TargetCompleted) eventType() string { return "target_completed" }

// TargetSkipped is emitted when a target is skipped (up to date, etc.).
type TargetSkipped struct {
	Target string // Target path or name
	Reason string // "up to date", "already built", etc.
}

func (e TargetSkipped) eventType() string { return "target_skipped" }

// CommandStarted is emitted when a recipe command begins execution.
type CommandStarted struct {
	Target  string // Target being built
	Command string // Command line being executed
}

func (e CommandStarted) eventType() string { return "command_started" }

// CommandOutput is emitted when a command produces output.
type CommandOutput struct {
	Target string // Target being built
	Stdout string // Standard output
	Stderr string // Standard error
}

func (e CommandOutput) eventType() string { return "command_output" }

// CommandCompleted is emitted when a command finishes execution.
type CommandCompleted struct {
	Target   string        // Target being built
	Command  string        // Command that was executed
	ExitCode int           // Exit code (0 = success)
	Duration time.Duration // How long the command took
}

func (e CommandCompleted) eventType() string { return "command_completed" }

// StalenessChecked is emitted during staleness checking (verbose mode).
type StalenessChecked struct {
	Target string // Target being checked
	Reason string // "src/main.c is newer", "target missing", etc.
	Action string // "rebuild", "skip"
}

func (e StalenessChecked) eventType() string { return "staleness_checked" }

// BuildSummary is emitted at the end of a build.
type BuildSummary struct {
	Total     int           // Total number of targets
	Succeeded int           // Number that succeeded
	Failed    int           // Number that failed
	Skipped   int           // Number that were skipped
	Duration  time.Duration // Total build duration
}

func (e BuildSummary) eventType() string { return "build_summary" }

// ErrorOccurred is emitted when an error occurs during any phase.
type ErrorOccurred struct {
	Category string // "parse", "semantic", "eval", "plan", "execute"
	Code     string // "E001", "E200", etc.
	Message  string // Error message
	Location string // "Needfile:10:5" or empty
	Context  string // Source code snippet or empty
	Hint     string // Fix suggestion or empty
}

func (e ErrorOccurred) eventType() string { return "error" }

// DryRunTarget is emitted in dry-run mode to show what would be built.
type DryRunTarget struct {
	Target string // Target that would be built
	Index  int    // 1-based index
	Total  int    // Total targets
}

func (e DryRunTarget) eventType() string { return "dry_run_target" }

// DryRunCommand is emitted in dry-run mode to show commands that would run.
type DryRunCommand struct {
	Target  string // Target this command belongs to
	Command string // Command that would be executed
}

func (e DryRunCommand) eventType() string { return "dry_run_command" }
