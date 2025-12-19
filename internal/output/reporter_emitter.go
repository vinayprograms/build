package output

import (
	"io"
	"time"
)

// EmitterBackedNormalReporter implements the Reporter interface using the
// event-based OutputWriter system. This provides colored output and unified
// output handling across CLI, TUI, and headless modes.
type EmitterBackedNormalReporter struct {
	emitter *Emitter
	index   int
	total   int
}

// NewEmitterBackedNormalReporter creates a reporter backed by the emitter system.
func NewEmitterBackedNormalReporter(w io.Writer, config WriterConfig) *EmitterBackedNormalReporter {
	writer := NewCLIWriter(w, config)
	return &EmitterBackedNormalReporter{
		emitter: NewEmitter(writer),
		index:   0,
		total:   1,
	}
}

// SetTotal sets the total number of targets for progress display.
func (r *EmitterBackedNormalReporter) SetTotal(total int) {
	r.total = total
}

// BuildStarted is called when a target build begins.
func (r *EmitterBackedNormalReporter) BuildStarted(target string) {
	r.index++
	r.emitter.TargetStarted(target, r.index, r.total)
}

// BuildCompleted is called when a target build finishes.
func (r *EmitterBackedNormalReporter) BuildCompleted(target string, success bool, errMsg string) {
	r.emitter.TargetCompleted(target, success, 0, errMsg)
}

// CommandOutput is called to display command output.
func (r *EmitterBackedNormalReporter) CommandOutput(command, stdout, stderr string) {
	if stdout == "" && stderr == "" {
		return
	}
	r.emitter.CommandOutput("", stdout, stderr)
}

// Summary is called at the end to show build summary.
func (r *EmitterBackedNormalReporter) Summary(total, failed int) {
	r.emitter.BuildSummary(total, total-failed, failed, 0, 0)
}

// NothingToBuild is called when a target is already up to date.
func (r *EmitterBackedNormalReporter) NothingToBuild(target string) {
	r.emitter.TargetSkipped(target, "up to date")
}

// ----------------------------------------------------------------------------
// EmitterBackedDryRunReporter
// ----------------------------------------------------------------------------

// EmitterBackedDryRunReporter implements dry-run output using the emitter system.
type EmitterBackedDryRunReporter struct {
	emitter       *Emitter
	index         int
	total         int
	currentTarget string
}

// NewEmitterBackedDryRunReporter creates a dry-run reporter backed by emitter.
func NewEmitterBackedDryRunReporter(w io.Writer, config WriterConfig) *EmitterBackedDryRunReporter {
	writer := NewCLIWriter(w, config)
	return &EmitterBackedDryRunReporter{
		emitter: NewEmitter(writer),
		index:   0,
		total:   1,
	}
}

// SetTotal sets the total number of targets for progress display.
func (r *EmitterBackedDryRunReporter) SetTotal(total int) {
	r.total = total
}

// WouldBuild outputs "Would build: target" line.
func (r *EmitterBackedDryRunReporter) WouldBuild(target string) {
	r.index++
	r.currentTarget = target
	r.emitter.DryRunTarget(target, r.index, r.total)
}

// ShowCommand outputs a command that would be executed.
func (r *EmitterBackedDryRunReporter) ShowCommand(command string) {
	r.emitter.DryRunCommand(r.currentTarget, command)
}

// TargetComplete marks the end of a target's commands.
func (r *EmitterBackedDryRunReporter) TargetComplete() {
	// In event-based system, blank lines between targets handled by writer
}

// Summary outputs a summary of what would be built.
func (r *EmitterBackedDryRunReporter) Summary(total int) {
	r.emitter.BuildSummary(total, total, 0, 0, 0)
}

// NothingToBuild outputs that a target is already up to date.
func (r *EmitterBackedDryRunReporter) NothingToBuild(target string) {
	r.emitter.TargetSkipped(target, "up to date")
}

// ----------------------------------------------------------------------------
// EmitterBackedVerboseReporter
// ----------------------------------------------------------------------------

// EmitterBackedVerboseReporter implements verbose output using the emitter system.
type EmitterBackedVerboseReporter struct {
	emitter       *Emitter
	index         int
	total         int
	currentTarget string
}

// NewEmitterBackedVerboseReporter creates a verbose reporter backed by emitter.
func NewEmitterBackedVerboseReporter(w io.Writer, config WriterConfig) *EmitterBackedVerboseReporter {
	// Force verbose mode in the config
	config.Verbose = true
	writer := NewCLIWriter(w, config)
	return &EmitterBackedVerboseReporter{
		emitter: NewEmitter(writer),
		index:   0,
		total:   1,
	}
}

// SetTotal sets the total number of targets for progress display.
func (r *EmitterBackedVerboseReporter) SetTotal(total int) {
	r.total = total
}

// StartVariableEvaluation outputs a header for the variable evaluation phase.
func (r *EmitterBackedVerboseReporter) StartVariableEvaluation() {
	r.emitter.PhaseStarted("eval")
}

// VariableEvaluated outputs a variable evaluation result.
func (r *EmitterBackedVerboseReporter) VariableEvaluated(name, expr, result string) {
	r.emitter.VariableEvaluated(name, expr, result)
}

// StartStalenessChecks outputs a header for the staleness check phase.
func (r *EmitterBackedVerboseReporter) StartStalenessChecks() {
	r.emitter.PhaseStarted("plan")
}

// StalenessCheck outputs a staleness check result.
func (r *EmitterBackedVerboseReporter) StalenessCheck(target, reason, action string) {
	r.emitter.StalenessChecked(target, reason, action)
}

// BuildStarted outputs a header for building a target.
func (r *EmitterBackedVerboseReporter) BuildStarted(target string) {
	r.index++
	r.currentTarget = target
	r.emitter.TargetStarted(target, r.index, r.total)
}

// CommandExecuted outputs a command that was executed (indented).
func (r *EmitterBackedVerboseReporter) CommandExecuted(command string) {
	r.emitter.CommandStarted(r.currentTarget, command)
}

// BuildCompleted outputs completion status for a target build.
func (r *EmitterBackedVerboseReporter) BuildCompleted(target string, success bool, errMsg string) {
	r.emitter.TargetCompleted(target, success, 0, errMsg)
}

// CommandOutput outputs command stdout and stderr.
func (r *EmitterBackedVerboseReporter) CommandOutput(command, stdout, stderr string) {
	if stdout == "" && stderr == "" {
		return
	}
	r.emitter.CommandOutput(r.currentTarget, stdout, stderr)
}

// Summary outputs the build summary.
func (r *EmitterBackedVerboseReporter) Summary(total, failed int) {
	r.emitter.BuildSummary(total, total-failed, failed, 0, 0)
}

// NothingToBuild outputs that a target is already up to date.
func (r *EmitterBackedVerboseReporter) NothingToBuild(target string) {
	r.emitter.TargetSkipped(target, "up to date")
}

// ----------------------------------------------------------------------------
// EmitterBackedProgressReporter
// ----------------------------------------------------------------------------

// EmitterBackedProgressReporter implements progress output for parallel builds
// using the emitter system.
type EmitterBackedProgressReporter struct {
	emitter   *Emitter
	total     int
	started   int
	completed int
	building  map[string]bool
}

// NewEmitterBackedProgressReporter creates a progress reporter backed by emitter.
func NewEmitterBackedProgressReporter(w io.Writer, config WriterConfig, total int) *EmitterBackedProgressReporter {
	writer := NewCLIWriter(w, config)
	return &EmitterBackedProgressReporter{
		emitter:  NewEmitter(writer),
		total:    total,
		building: make(map[string]bool),
	}
}

// BuildStarted shows progress for a target starting to build.
func (r *EmitterBackedProgressReporter) BuildStarted(target string) {
	r.started++
	r.building[target] = true
	r.emitter.TargetStarted(target, r.started, r.total)
}

// BuildCompleted handles target build completion.
func (r *EmitterBackedProgressReporter) BuildCompleted(target string, success bool, errMsg string) {
	r.completed++
	delete(r.building, target)
	r.emitter.TargetCompleted(target, success, 0, errMsg)
}

// CommandOutput outputs command stdout and stderr (for errors).
func (r *EmitterBackedProgressReporter) CommandOutput(command, stdout, stderr string) {
	if stderr != "" {
		r.emitter.CommandOutput("", "", stderr)
	}
}

// Summary outputs the build summary.
func (r *EmitterBackedProgressReporter) Summary(total, failed int) {
	r.emitter.BuildSummary(total, total-failed, failed, 0, time.Duration(0))
}

// NothingToBuild outputs that a target is already up to date.
func (r *EmitterBackedProgressReporter) NothingToBuild(target string) {
	r.emitter.TargetSkipped(target, "up to date")
}

// CurrentlyBuilding returns the list of currently building targets.
func (r *EmitterBackedProgressReporter) CurrentlyBuilding() []string {
	result := make([]string, 0, len(r.building))
	for target := range r.building {
		result = append(result, target)
	}
	return result
}
