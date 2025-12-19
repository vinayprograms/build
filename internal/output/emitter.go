package output

import "time"

// Emitter provides a convenient API for emitting output events.
// It wraps an OutputWriter and provides typed methods for each event type.
type Emitter struct {
	writer OutputWriter
}

// NewEmitter creates a new Emitter that writes events to the given writer.
// If writer is nil, events are silently discarded.
func NewEmitter(writer OutputWriter) *Emitter {
	return &Emitter{writer: writer}
}

// NoOpEmitter returns an emitter that discards all events.
func NoOpEmitter() *Emitter {
	return &Emitter{writer: nil}
}

// emit writes an event to the writer if one is configured.
func (e *Emitter) emit(event OutputEvent) {
	if e.writer != nil {
		e.writer.WriteEvent(event)
	}
}

// PhaseStarted emits a phase start event.
func (e *Emitter) PhaseStarted(phase string) {
	e.emit(PhaseStarted{
		Phase:     phase,
		Timestamp: time.Now(),
	})
}

// PhaseCompleted emits a phase completion event.
func (e *Emitter) PhaseCompleted(phase string, duration time.Duration) {
	e.emit(PhaseCompleted{
		Phase:     phase,
		Timestamp: time.Now(),
		Duration:  duration,
	})
}

// VariableEvaluated emits a variable evaluation event.
// expr is the expression (empty for simple values), result is the evaluated value.
func (e *Emitter) VariableEvaluated(name, expr, result string) {
	e.emit(VariableEvaluated{
		Name:   name,
		Expr:   expr,
		Result: result,
	})
}

// TargetStarted emits a target build start event.
func (e *Emitter) TargetStarted(target string, index, total int) {
	e.emit(TargetStarted{
		Target: target,
		Index:  index,
		Total:  total,
	})
}

// TargetCompleted emits a target build completion event.
func (e *Emitter) TargetCompleted(target string, success bool, duration time.Duration, errMsg string) {
	e.emit(TargetCompleted{
		Target:   target,
		Success:  success,
		Duration: duration,
		Error:    errMsg,
	})
}

// TargetSkipped emits a target skipped event.
func (e *Emitter) TargetSkipped(target, reason string) {
	e.emit(TargetSkipped{
		Target: target,
		Reason: reason,
	})
}

// CommandStarted emits a command start event.
func (e *Emitter) CommandStarted(target, command string) {
	e.emit(CommandStarted{
		Target:  target,
		Command: command,
	})
}

// CommandOutput emits a command output event.
func (e *Emitter) CommandOutput(target, stdout, stderr string) {
	e.emit(CommandOutput{
		Target: target,
		Stdout: stdout,
		Stderr: stderr,
	})
}

// CommandCompleted emits a command completion event.
func (e *Emitter) CommandCompleted(target, command string, exitCode int, duration time.Duration) {
	e.emit(CommandCompleted{
		Target:   target,
		Command:  command,
		ExitCode: exitCode,
		Duration: duration,
	})
}

// StalenessChecked emits a staleness check event.
func (e *Emitter) StalenessChecked(target, reason, action string) {
	e.emit(StalenessChecked{
		Target: target,
		Reason: reason,
		Action: action,
	})
}

// BuildSummary emits a build summary event.
func (e *Emitter) BuildSummary(total, succeeded, failed, skipped int, duration time.Duration) {
	e.emit(BuildSummary{
		Total:     total,
		Succeeded: succeeded,
		Failed:    failed,
		Skipped:   skipped,
		Duration:  duration,
	})
}

// Error emits an error event.
func (e *Emitter) Error(category, code, message, location, context, hint string) {
	e.emit(ErrorOccurred{
		Category: category,
		Code:     code,
		Message:  message,
		Location: location,
		Context:  context,
		Hint:     hint,
	})
}

// DryRunTarget emits a dry-run target event.
func (e *Emitter) DryRunTarget(target string, index, total int) {
	e.emit(DryRunTarget{
		Target: target,
		Index:  index,
		Total:  total,
	})
}

// DryRunCommand emits a dry-run command event.
func (e *Emitter) DryRunCommand(target, command string) {
	e.emit(DryRunCommand{
		Target:  target,
		Command: command,
	})
}

// Flush ensures all buffered output is written.
func (e *Emitter) Flush() {
	if e.writer != nil {
		e.writer.Flush()
	}
}
