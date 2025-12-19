package main

import (
	"io"
	"os"

	"github.com/vinayprograms/build/internal/output"
)

// ----------------------------------------------------------------------------
// Output System Setup
// ----------------------------------------------------------------------------

// CreateOutputEmitter creates an output emitter based on CLI flags.
// It sets up the appropriate output writer based on verbosity, quiet mode,
// and color settings.
func CreateOutputEmitter(verbose, quiet bool, color string) *output.Emitter {
	mode := output.DetectOutputMode()
	config := output.NewWriterConfigFromFlags(verbose, quiet, color)
	writer := output.NewWriterWithMode(mode, config)
	return output.NewEmitter(writer)
}

// CreateOutputWriter creates an output writer based on CLI flags.
// This is useful when direct writer access is needed.
func CreateOutputWriter(verbose, quiet bool, color string) output.OutputWriter {
	mode := output.DetectOutputMode()
	config := output.NewWriterConfigFromFlags(verbose, quiet, color)
	return output.NewWriter(mode, os.Stdout, config)
}

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

// ----------------------------------------------------------------------------
// Dry-Run Reporter Adapter
// ----------------------------------------------------------------------------

// DryRunOutputReporter defines the interface for dry-run output reporting.
// It shows what would be built without actually executing commands.
type DryRunOutputReporter interface {
	// WouldBuild outputs "Would build: target" line.
	WouldBuild(target string)

	// ShowCommand outputs a command that would be executed (indented).
	ShowCommand(command string)

	// TargetComplete marks the end of a target's commands.
	TargetComplete()

	// Summary outputs a summary of what would be built.
	Summary(total int)

	// NothingToBuild is called when a target is already up to date.
	NothingToBuild(target string)
}

// dryRunReporterAdapter wraps output.DryRunReporter for CLI use.
type dryRunReporterAdapter struct {
	reporter *output.DryRunReporter
}

// NewDryRunReporter creates a new dry-run output reporter.
func NewDryRunReporter(w io.Writer) DryRunOutputReporter {
	return &dryRunReporterAdapter{
		reporter: output.NewDryRunReporter(w),
	}
}

// WouldBuild implements DryRunOutputReporter.
func (a *dryRunReporterAdapter) WouldBuild(target string) {
	a.reporter.WouldBuild(target)
}

// ShowCommand implements DryRunOutputReporter.
func (a *dryRunReporterAdapter) ShowCommand(command string) {
	a.reporter.ShowCommand(command)
}

// TargetComplete implements DryRunOutputReporter.
func (a *dryRunReporterAdapter) TargetComplete() {
	a.reporter.TargetComplete()
}

// Summary implements DryRunOutputReporter.
func (a *dryRunReporterAdapter) Summary(total int) {
	a.reporter.Summary(total)
}

// NothingToBuild implements DryRunOutputReporter.
func (a *dryRunReporterAdapter) NothingToBuild(target string) {
	a.reporter.NothingToBuild(target)
}

// ----------------------------------------------------------------------------
// Verbose Reporter Adapter
// ----------------------------------------------------------------------------

// VerboseOutputReporter defines the interface for verbose output reporting.
// It shows variable evaluation, staleness checks, and build progress.
type VerboseOutputReporter interface {
	// StartVariableEvaluation outputs a header for the variable evaluation phase.
	StartVariableEvaluation()

	// VariableEvaluated outputs a variable evaluation result.
	VariableEvaluated(name, expr, result string)

	// StartStalenessChecks outputs a header for the staleness check phase.
	StartStalenessChecks()

	// StalenessCheck outputs a staleness check result.
	StalenessCheck(target, reason, action string)

	// BuildStarted outputs a header for building a target.
	BuildStarted(target string)

	// CommandExecuted outputs a command that was executed (indented).
	CommandExecuted(command string)

	// BuildCompleted outputs completion status for a target build.
	BuildCompleted(target string, success bool, errMsg string)

	// CommandOutput outputs command stdout and stderr.
	CommandOutput(command, stdout, stderr string)

	// Summary outputs the build summary.
	Summary(total, failed int)

	// NothingToBuild outputs that a target is already up to date.
	NothingToBuild(target string)
}

// verboseReporterAdapter wraps output.VerboseReporter for CLI use.
type verboseReporterAdapter struct {
	reporter *output.VerboseReporter
}

// NewVerboseReporter creates a new verbose output reporter.
func NewVerboseReporter(w io.Writer) VerboseOutputReporter {
	return &verboseReporterAdapter{
		reporter: output.NewVerboseReporter(w),
	}
}

// StartVariableEvaluation implements VerboseOutputReporter.
func (a *verboseReporterAdapter) StartVariableEvaluation() {
	a.reporter.StartVariableEvaluation()
}

// VariableEvaluated implements VerboseOutputReporter.
func (a *verboseReporterAdapter) VariableEvaluated(name, expr, result string) {
	a.reporter.VariableEvaluated(name, expr, result)
}

// StartStalenessChecks implements VerboseOutputReporter.
func (a *verboseReporterAdapter) StartStalenessChecks() {
	a.reporter.StartStalenessChecks()
}

// StalenessCheck implements VerboseOutputReporter.
func (a *verboseReporterAdapter) StalenessCheck(target, reason, action string) {
	a.reporter.StalenessCheck(target, reason, action)
}

// BuildStarted implements VerboseOutputReporter.
func (a *verboseReporterAdapter) BuildStarted(target string) {
	a.reporter.BuildStarted(target)
}

// CommandExecuted implements VerboseOutputReporter.
func (a *verboseReporterAdapter) CommandExecuted(command string) {
	a.reporter.CommandExecuted(command)
}

// BuildCompleted implements VerboseOutputReporter.
func (a *verboseReporterAdapter) BuildCompleted(target string, success bool, errMsg string) {
	a.reporter.BuildCompleted(target, success, errMsg)
}

// CommandOutput implements VerboseOutputReporter.
func (a *verboseReporterAdapter) CommandOutput(command, stdout, stderr string) {
	a.reporter.CommandOutput(command, stdout, stderr)
}

// Summary implements VerboseOutputReporter.
func (a *verboseReporterAdapter) Summary(total, failed int) {
	a.reporter.Summary(total, failed)
}

// NothingToBuild implements VerboseOutputReporter.
func (a *verboseReporterAdapter) NothingToBuild(target string) {
	a.reporter.NothingToBuild(target)
}

// ----------------------------------------------------------------------------
// Progress Reporter Adapter (for parallel builds)
// ----------------------------------------------------------------------------

// ProgressOutputReporter defines the interface for parallel build progress reporting.
// It shows progress counts and currently building targets.
type ProgressOutputReporter interface {
	// BuildStarted shows progress for a target starting to build.
	BuildStarted(target string)

	// BuildCompleted handles target build completion.
	BuildCompleted(target string, success bool, errMsg string)

	// CommandOutput outputs command stdout and stderr.
	CommandOutput(command, stdout, stderr string)

	// Summary outputs the build summary.
	Summary(total, failed int)

	// NothingToBuild outputs that a target is already up to date.
	NothingToBuild(target string)

	// CurrentlyBuilding returns the list of currently building targets.
	CurrentlyBuilding() []string
}

// progressReporterAdapter wraps output.ProgressReporter for CLI use.
type progressReporterAdapter struct {
	reporter *output.ProgressReporter
}

// NewProgressReporter creates a new progress output reporter.
// total is the total number of targets expected to build.
func NewProgressReporter(w io.Writer, total int) ProgressOutputReporter {
	return &progressReporterAdapter{
		reporter: output.NewProgressReporter(w, total),
	}
}

// BuildStarted implements ProgressOutputReporter.
func (a *progressReporterAdapter) BuildStarted(target string) {
	a.reporter.BuildStarted(target)
}

// BuildCompleted implements ProgressOutputReporter.
func (a *progressReporterAdapter) BuildCompleted(target string, success bool, errMsg string) {
	a.reporter.BuildCompleted(target, success, errMsg)
}

// CommandOutput implements ProgressOutputReporter.
func (a *progressReporterAdapter) CommandOutput(command, stdout, stderr string) {
	a.reporter.CommandOutput(command, stdout, stderr)
}

// Summary implements ProgressOutputReporter.
func (a *progressReporterAdapter) Summary(total, failed int) {
	a.reporter.Summary(total, failed)
}

// NothingToBuild implements ProgressOutputReporter.
func (a *progressReporterAdapter) NothingToBuild(target string) {
	a.reporter.NothingToBuild(target)
}

// CurrentlyBuilding implements ProgressOutputReporter.
func (a *progressReporterAdapter) CurrentlyBuilding() []string {
	return a.reporter.CurrentlyBuilding()
}
