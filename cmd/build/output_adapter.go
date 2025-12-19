package main

import (
	"io"

	"github.com/vinayprograms/build/internal/output"
)

// ----------------------------------------------------------------------------
// Output Reporter Adapter
// ----------------------------------------------------------------------------

// OutputReporter defines the interface for build output reporting.
// Different implementations can provide different output styles.
type OutputReporter interface {
	// BuildStarted is called when a target build begins.
	BuildStarted(target string)

	// BuildCompleted is called when a target build finishes.
	BuildCompleted(target string, success bool, errMsg string)

	// CommandOutput is called to display command output.
	CommandOutput(command, stdout, stderr string)

	// Summary is called at the end to show build summary.
	Summary(total, failed int)

	// NothingToBuild is called when a target is already up to date.
	NothingToBuild(target string)
}

// normalReporterAdapter wraps output.NormalReporter for CLI use.
type normalReporterAdapter struct {
	reporter *output.NormalReporter
}

// NewNormalReporter creates a new normal output reporter.
func NewNormalReporter(w io.Writer) OutputReporter {
	return &normalReporterAdapter{
		reporter: output.NewNormalReporter(w),
	}
}

// BuildStarted implements OutputReporter.
func (a *normalReporterAdapter) BuildStarted(target string) {
	a.reporter.BuildStarted(target)
}

// BuildCompleted implements OutputReporter.
func (a *normalReporterAdapter) BuildCompleted(target string, success bool, errMsg string) {
	a.reporter.BuildCompleted(target, success, errMsg)
}

// CommandOutput implements OutputReporter.
func (a *normalReporterAdapter) CommandOutput(command, stdout, stderr string) {
	a.reporter.CommandOutput(command, stdout, stderr)
}

// Summary implements OutputReporter.
func (a *normalReporterAdapter) Summary(total, failed int) {
	a.reporter.Summary(total, failed)
}

// NothingToBuild implements OutputReporter.
func (a *normalReporterAdapter) NothingToBuild(target string) {
	a.reporter.NothingToBuild(target)
}
