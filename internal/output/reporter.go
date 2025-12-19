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
