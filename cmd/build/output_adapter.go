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

// normalReporterAdapter wraps output.EmitterBackedNormalReporter for CLI use.
// It delegates to the new event-based output system while maintaining the
// Reporter interface for backward compatibility.
type normalReporterAdapter struct {
	reporter *output.EmitterBackedNormalReporter
}

// NewNormalReporter creates a new normal output reporter using the emitter system.
func NewNormalReporter(w io.Writer) OutputReporter {
	config := output.WriterConfig{Color: "auto"}
	return &normalReporterAdapter{
		reporter: output.NewEmitterBackedNormalReporter(w, config),
	}
}

// NewNormalReporterWithConfig creates a new normal output reporter with custom config.
func NewNormalReporterWithConfig(w io.Writer, verbose, quiet bool, color string) OutputReporter {
	config := output.NewWriterConfigFromFlags(verbose, quiet, color)
	return &normalReporterAdapter{
		reporter: output.NewEmitterBackedNormalReporter(w, config),
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

// SetTotal sets the total number of targets for progress display.
func (a *normalReporterAdapter) SetTotal(total int) {
	a.reporter.SetTotal(total)
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

// dryRunReporterAdapter wraps output.EmitterBackedDryRunReporter for CLI use.
// It delegates to the new event-based output system while maintaining the
// DryRunReporter interface for backward compatibility.
type dryRunReporterAdapter struct {
	reporter *output.EmitterBackedDryRunReporter
}

// NewDryRunReporter creates a new dry-run output reporter using the emitter system.
func NewDryRunReporter(w io.Writer) DryRunOutputReporter {
	config := output.WriterConfig{Color: "auto"}
	return &dryRunReporterAdapter{
		reporter: output.NewEmitterBackedDryRunReporter(w, config),
	}
}

// NewDryRunReporterWithConfig creates a new dry-run output reporter with custom config.
func NewDryRunReporterWithConfig(w io.Writer, verbose, quiet bool, color string) DryRunOutputReporter {
	config := output.NewWriterConfigFromFlags(verbose, quiet, color)
	return &dryRunReporterAdapter{
		reporter: output.NewEmitterBackedDryRunReporter(w, config),
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

// SetTotal sets the total number of targets for progress display.
func (a *dryRunReporterAdapter) SetTotal(total int) {
	a.reporter.SetTotal(total)
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

// verboseReporterAdapter wraps output.EmitterBackedVerboseReporter for CLI use.
// It delegates to the new event-based output system while maintaining the
// VerboseReporter interface for backward compatibility.
type verboseReporterAdapter struct {
	reporter *output.EmitterBackedVerboseReporter
}

// NewVerboseReporter creates a new verbose output reporter using the emitter system.
func NewVerboseReporter(w io.Writer) VerboseOutputReporter {
	config := output.WriterConfig{Color: "auto", Verbose: true}
	return &verboseReporterAdapter{
		reporter: output.NewEmitterBackedVerboseReporter(w, config),
	}
}

// NewVerboseReporterWithConfig creates a new verbose output reporter with custom config.
func NewVerboseReporterWithConfig(w io.Writer, quiet bool, color string) VerboseOutputReporter {
	config := output.NewWriterConfigFromFlags(true, quiet, color)
	return &verboseReporterAdapter{
		reporter: output.NewEmitterBackedVerboseReporter(w, config),
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

// SetTotal sets the total number of targets for progress display.
func (a *verboseReporterAdapter) SetTotal(total int) {
	a.reporter.SetTotal(total)
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

// progressReporterAdapter wraps output.EmitterBackedProgressReporter for CLI use.
// It delegates to the new event-based output system while maintaining the
// ProgressReporter interface for backward compatibility.
type progressReporterAdapter struct {
	reporter *output.EmitterBackedProgressReporter
}

// NewProgressReporter creates a new progress output reporter using the emitter system.
// total is the total number of targets expected to build.
func NewProgressReporter(w io.Writer, total int) ProgressOutputReporter {
	config := output.WriterConfig{Color: "auto"}
	return &progressReporterAdapter{
		reporter: output.NewEmitterBackedProgressReporter(w, config, total),
	}
}

// NewProgressReporterWithConfig creates a new progress output reporter with custom config.
func NewProgressReporterWithConfig(w io.Writer, total int, verbose, quiet bool, color string) ProgressOutputReporter {
	config := output.NewWriterConfigFromFlags(verbose, quiet, color)
	return &progressReporterAdapter{
		reporter: output.NewEmitterBackedProgressReporter(w, config, total),
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
