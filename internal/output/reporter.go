package output

import (
	"fmt"
	"io"
)

// Reporter defines the interface for build output reporting.
// Different implementations can provide different output styles
// (normal, verbose, quiet, etc.).
type Reporter interface {
	// BuildStarted is called when a target build begins.
	BuildStarted(target string)

	// BuildCompleted is called when a target build finishes.
	// success indicates whether the build succeeded.
	// errMsg contains the error message if success is false.
	BuildCompleted(target string, success bool, errMsg string)

	// CommandOutput is called to display command output.
	// stdout and stderr contain the respective output streams.
	CommandOutput(command, stdout, stderr string)

	// Summary is called at the end to show build summary.
	// total is the number of targets built, failed is the number that failed.
	Summary(total, failed int)

	// NothingToBuild is called when a target is already up to date.
	NothingToBuild(target string)
}

// NormalReporter implements Reporter for normal (non-verbose) output.
// It shows target names being built, command output, and completion status.
type NormalReporter struct {
	w io.Writer
}

// NewNormalReporter creates a new NormalReporter that writes to w.
func NewNormalReporter(w io.Writer) *NormalReporter {
	return &NormalReporter{w: w}
}

// BuildStarted outputs the target being built.
func (r *NormalReporter) BuildStarted(target string) {
	fmt.Fprintf(r.w, "Building %s\n", target)
}

// BuildCompleted outputs the build result.
func (r *NormalReporter) BuildCompleted(target string, success bool, errMsg string) {
	if success {
		fmt.Fprintf(r.w, "Built %s\n", target)
	} else {
		fmt.Fprintf(r.w, "FAILED %s: %s\n", target, errMsg)
	}
}

// CommandOutput outputs command stdout and stderr.
// Empty output is suppressed.
func (r *NormalReporter) CommandOutput(command, stdout, stderr string) {
	if stdout != "" {
		fmt.Fprint(r.w, stdout)
		// Ensure output ends with newline
		if len(stdout) > 0 && stdout[len(stdout)-1] != '\n' {
			fmt.Fprintln(r.w)
		}
	}
	if stderr != "" {
		fmt.Fprint(r.w, stderr)
		// Ensure output ends with newline
		if len(stderr) > 0 && stderr[len(stderr)-1] != '\n' {
			fmt.Fprintln(r.w)
		}
	}
}

// Summary outputs the build summary.
func (r *NormalReporter) Summary(total, failed int) {
	if failed == 0 {
		if total == 1 {
			fmt.Fprintf(r.w, "Build success: %d target built\n", total)
		} else {
			fmt.Fprintf(r.w, "Build success: %d targets built\n", total)
		}
	} else {
		fmt.Fprintf(r.w, "Build failed: %d of %d targets failed\n", failed, total)
	}
}

// NothingToBuild outputs that a target is already up to date.
func (r *NormalReporter) NothingToBuild(target string) {
	fmt.Fprintf(r.w, "%s is up to date\n", target)
}

// DryRunReporter implements output formatting for dry-run mode.
// It shows "Would build: target" followed by indented commands.
type DryRunReporter struct {
	w           io.Writer
	targetCount int
}

// NewDryRunReporter creates a new DryRunReporter that writes to w.
func NewDryRunReporter(w io.Writer) *DryRunReporter {
	return &DryRunReporter{w: w}
}

// WouldBuild outputs "Would build: target" line.
func (r *DryRunReporter) WouldBuild(target string) {
	fmt.Fprintf(r.w, "Would build: %s\n", target)
	r.targetCount++
}

// ShowCommand outputs a command that would be executed (indented).
func (r *DryRunReporter) ShowCommand(command string) {
	fmt.Fprintf(r.w, "  %s\n", command)
}

// TargetComplete marks the end of a target's commands.
// Adds a blank line between targets for readability.
func (r *DryRunReporter) TargetComplete() {
	fmt.Fprintln(r.w)
}

// Summary outputs a summary of what would be built.
func (r *DryRunReporter) Summary(total int) {
	if total == 1 {
		fmt.Fprintf(r.w, "Would build %d target\n", total)
	} else {
		fmt.Fprintf(r.w, "Would build %d targets\n", total)
	}
}

// NothingToBuild outputs that a target is already up to date.
func (r *DryRunReporter) NothingToBuild(target string) {
	fmt.Fprintf(r.w, "%s is up to date\n", target)
}

// VerboseReporter implements output formatting for verbose mode (-v).
// It shows variable evaluation results, staleness check results,
// and dependency resolution information in addition to build progress.
type VerboseReporter struct {
	w io.Writer
}

// NewVerboseReporter creates a new VerboseReporter that writes to w.
func NewVerboseReporter(w io.Writer) *VerboseReporter {
	return &VerboseReporter{w: w}
}

// StartVariableEvaluation outputs a header for the variable evaluation phase.
func (r *VerboseReporter) StartVariableEvaluation() {
	fmt.Fprintln(r.w, "Evaluating variables...")
}

// VariableEvaluated outputs a variable evaluation result.
// expr is the expression (e.g., "shell(find src -name \"*.c\")"), result is the evaluated value.
// If expr is empty, only shows "name → result".
func (r *VerboseReporter) VariableEvaluated(name, expr, result string) {
	if expr != "" {
		fmt.Fprintf(r.w, "  %s = %s → %s\n", name, expr, result)
	} else {
		fmt.Fprintf(r.w, "  %s → %s\n", name, result)
	}
}

// StartStalenessChecks outputs a header for the staleness check phase.
func (r *VerboseReporter) StartStalenessChecks() {
	fmt.Fprintln(r.w, "\nChecking targets...")
}

// StalenessCheck outputs a staleness check result.
// reason explains why (e.g., "src/main.c is newer", "up to date").
// action is the decision (e.g., "rebuild", "skip").
func (r *VerboseReporter) StalenessCheck(target, reason, action string) {
	fmt.Fprintf(r.w, "  %s: %s → %s\n", target, reason, action)
}

// BuildStarted outputs a header for building a target.
func (r *VerboseReporter) BuildStarted(target string) {
	fmt.Fprintf(r.w, "\nBuilding %s...\n", target)
}

// CommandExecuted outputs a command that was executed (indented).
func (r *VerboseReporter) CommandExecuted(command string) {
	fmt.Fprintf(r.w, "  %s\n", command)
}

// BuildCompleted outputs completion status for a target build.
func (r *VerboseReporter) BuildCompleted(target string, success bool, errMsg string) {
	if !success {
		fmt.Fprintf(r.w, "FAILED %s: %s\n", target, errMsg)
	}
}

// CommandOutput outputs command stdout and stderr.
func (r *VerboseReporter) CommandOutput(command, stdout, stderr string) {
	if stdout != "" {
		fmt.Fprint(r.w, stdout)
		if len(stdout) > 0 && stdout[len(stdout)-1] != '\n' {
			fmt.Fprintln(r.w)
		}
	}
	if stderr != "" {
		fmt.Fprint(r.w, stderr)
		if len(stderr) > 0 && stderr[len(stderr)-1] != '\n' {
			fmt.Fprintln(r.w)
		}
	}
}

// Summary outputs the build summary.
func (r *VerboseReporter) Summary(total, failed int) {
	fmt.Fprintln(r.w)
	if failed == 0 {
		fmt.Fprintln(r.w, "Done.")
	} else {
		fmt.Fprintf(r.w, "Build failed: %d of %d targets failed\n", failed, total)
	}
}

// NothingToBuild outputs that a target is already up to date.
func (r *VerboseReporter) NothingToBuild(target string) {
	fmt.Fprintf(r.w, "%s is up to date\n", target)
}

// ProgressReporter implements output formatting for parallel builds.
// It shows progress counts and currently building targets.
type ProgressReporter struct {
	w         io.Writer
	total     int             // Total number of targets to build
	started   int             // Number of targets started
	completed int             // Number of targets completed
	building  map[string]bool // Currently building targets
}

// NewProgressReporter creates a new ProgressReporter that writes to w.
// total is the total number of targets expected to build.
func NewProgressReporter(w io.Writer, total int) *ProgressReporter {
	return &ProgressReporter{
		w:        w,
		total:    total,
		building: make(map[string]bool),
	}
}

// BuildStarted outputs progress for a target starting to build.
// Shows "[current/total] Building target"
func (r *ProgressReporter) BuildStarted(target string) {
	r.started++
	r.building[target] = true
	fmt.Fprintf(r.w, "[%d/%d] Building %s\n", r.started, r.total, target)
}

// BuildCompleted handles target build completion.
func (r *ProgressReporter) BuildCompleted(target string, success bool, errMsg string) {
	r.completed++
	delete(r.building, target)
	if !success {
		fmt.Fprintf(r.w, "FAILED %s: %s\n", target, errMsg)
	}
}

// CommandOutput outputs command stdout and stderr (for errors).
func (r *ProgressReporter) CommandOutput(command, stdout, stderr string) {
	if stderr != "" {
		fmt.Fprint(r.w, stderr)
		if len(stderr) > 0 && stderr[len(stderr)-1] != '\n' {
			fmt.Fprintln(r.w)
		}
	}
}

// Summary outputs the build summary.
func (r *ProgressReporter) Summary(total, failed int) {
	if failed == 0 {
		if total == 1 {
			fmt.Fprintf(r.w, "Build success: %d target built\n", total)
		} else {
			fmt.Fprintf(r.w, "Build success: %d targets built\n", total)
		}
	} else {
		fmt.Fprintf(r.w, "Build failed: %d of %d targets failed\n", failed, total)
	}
}

// NothingToBuild outputs that a target is already up to date.
func (r *ProgressReporter) NothingToBuild(target string) {
	fmt.Fprintf(r.w, "%s is up to date\n", target)
}

// CurrentlyBuilding returns the list of currently building targets.
func (r *ProgressReporter) CurrentlyBuilding() []string {
	result := make([]string, 0, len(r.building))
	for target := range r.building {
		result = append(result, target)
	}
	return result
}
