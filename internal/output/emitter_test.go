package output

import (
	"testing"
	"time"
)

// TestEmitter_EmitPhaseEvents tests phase event emission.
func TestEmitter_EmitPhaseEvents(t *testing.T) {
	collector := &eventCollector{}
	emitter := NewEmitter(collector)

	emitter.PhaseStarted("parse")
	emitter.PhaseCompleted("parse", 100*time.Millisecond)

	if len(collector.events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(collector.events))
	}

	// Check PhaseStarted event
	started, ok := collector.events[0].(PhaseStarted)
	if !ok {
		t.Fatalf("expected PhaseStarted event, got %T", collector.events[0])
	}
	if started.Phase != "parse" {
		t.Errorf("expected phase 'parse', got '%s'", started.Phase)
	}
	if started.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}

	// Check PhaseCompleted event
	completed, ok := collector.events[1].(PhaseCompleted)
	if !ok {
		t.Fatalf("expected PhaseCompleted event, got %T", collector.events[1])
	}
	if completed.Phase != "parse" {
		t.Errorf("expected phase 'parse', got '%s'", completed.Phase)
	}
	if completed.Duration != 100*time.Millisecond {
		t.Errorf("expected duration 100ms, got %v", completed.Duration)
	}
}

// TestEmitter_EmitVariableEvaluated tests variable evaluation event emission.
func TestEmitter_EmitVariableEvaluated(t *testing.T) {
	collector := &eventCollector{}
	emitter := NewEmitter(collector)

	emitter.VariableEvaluated("cc", "shell(which gcc)", "/usr/bin/gcc")
	emitter.VariableEvaluated("flags", "", "-O2")

	if len(collector.events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(collector.events))
	}

	// Check first variable (with expression)
	v1, ok := collector.events[0].(VariableEvaluated)
	if !ok {
		t.Fatalf("expected VariableEvaluated event, got %T", collector.events[0])
	}
	if v1.Name != "cc" {
		t.Errorf("expected name 'cc', got '%s'", v1.Name)
	}
	if v1.Expr != "shell(which gcc)" {
		t.Errorf("expected expr 'shell(which gcc)', got '%s'", v1.Expr)
	}
	if v1.Result != "/usr/bin/gcc" {
		t.Errorf("expected result '/usr/bin/gcc', got '%s'", v1.Result)
	}

	// Check second variable (without expression)
	v2, ok := collector.events[1].(VariableEvaluated)
	if !ok {
		t.Fatalf("expected VariableEvaluated event, got %T", collector.events[1])
	}
	if v2.Name != "flags" {
		t.Errorf("expected name 'flags', got '%s'", v2.Name)
	}
	if v2.Expr != "" {
		t.Errorf("expected empty expr, got '%s'", v2.Expr)
	}
	if v2.Result != "-O2" {
		t.Errorf("expected result '-O2', got '%s'", v2.Result)
	}
}

// TestEmitter_EmitTargetEvents tests target lifecycle event emission.
func TestEmitter_EmitTargetEvents(t *testing.T) {
	collector := &eventCollector{}
	emitter := NewEmitter(collector)

	emitter.TargetStarted("foo.o", 1, 5)
	emitter.TargetCompleted("foo.o", true, 50*time.Millisecond, "")
	emitter.TargetStarted("bar.o", 2, 5)
	emitter.TargetCompleted("bar.o", false, 30*time.Millisecond, "command failed")
	emitter.TargetSkipped("baz.o", "up to date")

	if len(collector.events) != 5 {
		t.Fatalf("expected 5 events, got %d", len(collector.events))
	}

	// Check TargetStarted
	ts, ok := collector.events[0].(TargetStarted)
	if !ok {
		t.Fatalf("expected TargetStarted event, got %T", collector.events[0])
	}
	if ts.Target != "foo.o" || ts.Index != 1 || ts.Total != 5 {
		t.Errorf("unexpected TargetStarted: %+v", ts)
	}

	// Check successful TargetCompleted
	tc1, ok := collector.events[1].(TargetCompleted)
	if !ok {
		t.Fatalf("expected TargetCompleted event, got %T", collector.events[1])
	}
	if !tc1.Success || tc1.Error != "" {
		t.Errorf("expected success, got: %+v", tc1)
	}

	// Check failed TargetCompleted
	tc2, ok := collector.events[3].(TargetCompleted)
	if !ok {
		t.Fatalf("expected TargetCompleted event, got %T", collector.events[3])
	}
	if tc2.Success || tc2.Error != "command failed" {
		t.Errorf("expected failure with message, got: %+v", tc2)
	}

	// Check TargetSkipped
	skip, ok := collector.events[4].(TargetSkipped)
	if !ok {
		t.Fatalf("expected TargetSkipped event, got %T", collector.events[4])
	}
	if skip.Target != "baz.o" || skip.Reason != "up to date" {
		t.Errorf("unexpected TargetSkipped: %+v", skip)
	}
}

// TestEmitter_EmitCommandEvents tests command lifecycle event emission.
func TestEmitter_EmitCommandEvents(t *testing.T) {
	collector := &eventCollector{}
	emitter := NewEmitter(collector)

	emitter.CommandStarted("foo.o", "gcc -c foo.c")
	emitter.CommandOutput("foo.o", "compiling...\n", "")
	emitter.CommandCompleted("foo.o", "gcc -c foo.c", 0, 25*time.Millisecond)

	if len(collector.events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(collector.events))
	}

	// Check CommandStarted
	cs, ok := collector.events[0].(CommandStarted)
	if !ok {
		t.Fatalf("expected CommandStarted event, got %T", collector.events[0])
	}
	if cs.Target != "foo.o" || cs.Command != "gcc -c foo.c" {
		t.Errorf("unexpected CommandStarted: %+v", cs)
	}

	// Check CommandOutput
	co, ok := collector.events[1].(CommandOutput)
	if !ok {
		t.Fatalf("expected CommandOutput event, got %T", collector.events[1])
	}
	if co.Target != "foo.o" || co.Stdout != "compiling...\n" {
		t.Errorf("unexpected CommandOutput: %+v", co)
	}

	// Check CommandCompleted
	cc, ok := collector.events[2].(CommandCompleted)
	if !ok {
		t.Fatalf("expected CommandCompleted event, got %T", collector.events[2])
	}
	if cc.ExitCode != 0 || cc.Duration != 25*time.Millisecond {
		t.Errorf("unexpected CommandCompleted: %+v", cc)
	}
}

// TestEmitter_EmitStalenessChecked tests staleness check event emission.
func TestEmitter_EmitStalenessChecked(t *testing.T) {
	collector := &eventCollector{}
	emitter := NewEmitter(collector)

	emitter.StalenessChecked("foo.o", "src/foo.c is newer", "rebuild")
	emitter.StalenessChecked("bar.o", "up to date", "skip")

	if len(collector.events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(collector.events))
	}

	// Check rebuild decision
	s1, ok := collector.events[0].(StalenessChecked)
	if !ok {
		t.Fatalf("expected StalenessChecked event, got %T", collector.events[0])
	}
	if s1.Target != "foo.o" || s1.Action != "rebuild" {
		t.Errorf("unexpected StalenessChecked: %+v", s1)
	}

	// Check skip decision
	s2, ok := collector.events[1].(StalenessChecked)
	if !ok {
		t.Fatalf("expected StalenessChecked event, got %T", collector.events[1])
	}
	if s2.Target != "bar.o" || s2.Action != "skip" {
		t.Errorf("unexpected StalenessChecked: %+v", s2)
	}
}

// TestEmitter_EmitBuildSummary tests build summary event emission.
func TestEmitter_EmitBuildSummary(t *testing.T) {
	collector := &eventCollector{}
	emitter := NewEmitter(collector)

	emitter.BuildSummary(10, 8, 1, 1, 5*time.Second)

	if len(collector.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(collector.events))
	}

	summary, ok := collector.events[0].(BuildSummary)
	if !ok {
		t.Fatalf("expected BuildSummary event, got %T", collector.events[0])
	}
	if summary.Total != 10 || summary.Succeeded != 8 || summary.Failed != 1 || summary.Skipped != 1 {
		t.Errorf("unexpected BuildSummary: %+v", summary)
	}
	if summary.Duration != 5*time.Second {
		t.Errorf("expected duration 5s, got %v", summary.Duration)
	}
}

// TestEmitter_EmitError tests error event emission.
func TestEmitter_EmitError(t *testing.T) {
	collector := &eventCollector{}
	emitter := NewEmitter(collector)

	emitter.Error("parse", "E100", "unexpected token", "Needfile:10:5", "  foo = bar", "check syntax")

	if len(collector.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(collector.events))
	}

	err, ok := collector.events[0].(ErrorOccurred)
	if !ok {
		t.Fatalf("expected ErrorOccurred event, got %T", collector.events[0])
	}
	if err.Category != "parse" {
		t.Errorf("expected category 'parse', got '%s'", err.Category)
	}
	if err.Code != "E100" {
		t.Errorf("expected code 'E100', got '%s'", err.Code)
	}
	if err.Message != "unexpected token" {
		t.Errorf("expected message 'unexpected token', got '%s'", err.Message)
	}
	if err.Location != "Needfile:10:5" {
		t.Errorf("expected location 'Needfile:10:5', got '%s'", err.Location)
	}
	if err.Context != "  foo = bar" {
		t.Errorf("expected context '  foo = bar', got '%s'", err.Context)
	}
	if err.Hint != "check syntax" {
		t.Errorf("expected hint 'check syntax', got '%s'", err.Hint)
	}
}

// TestEmitter_EmitDryRunEvents tests dry-run event emission.
func TestEmitter_EmitDryRunEvents(t *testing.T) {
	collector := &eventCollector{}
	emitter := NewEmitter(collector)

	emitter.DryRunTarget("foo.o", 1, 3)
	emitter.DryRunCommand("foo.o", "gcc -c foo.c -o foo.o")

	if len(collector.events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(collector.events))
	}

	// Check DryRunTarget
	drt, ok := collector.events[0].(DryRunTarget)
	if !ok {
		t.Fatalf("expected DryRunTarget event, got %T", collector.events[0])
	}
	if drt.Target != "foo.o" || drt.Index != 1 || drt.Total != 3 {
		t.Errorf("unexpected DryRunTarget: %+v", drt)
	}

	// Check DryRunCommand
	drc, ok := collector.events[1].(DryRunCommand)
	if !ok {
		t.Fatalf("expected DryRunCommand event, got %T", collector.events[1])
	}
	if drc.Target != "foo.o" || drc.Command != "gcc -c foo.c -o foo.o" {
		t.Errorf("unexpected DryRunCommand: %+v", drc)
	}
}

// TestEmitter_NilWriter tests that emitter handles nil writer gracefully.
func TestEmitter_NilWriter(t *testing.T) {
	emitter := NewEmitter(nil)

	// These should not panic
	emitter.PhaseStarted("parse")
	emitter.VariableEvaluated("x", "", "1")
	emitter.TargetStarted("foo.o", 1, 1)
	emitter.CommandStarted("foo.o", "gcc")
	emitter.Error("parse", "E100", "error", "", "", "")
	emitter.BuildSummary(1, 1, 0, 0, time.Second)
}

// TestNoOpEmitter tests that NoOpEmitter doesn't emit anything.
func TestNoOpEmitter(t *testing.T) {
	collector := &eventCollector{}
	emitter := NoOpEmitter()

	// These should not result in any events
	emitter.PhaseStarted("parse")
	emitter.VariableEvaluated("x", "", "1")
	emitter.TargetStarted("foo.o", 1, 1)

	if len(collector.events) != 0 {
		t.Errorf("expected 0 events, got %d", len(collector.events))
	}
}

// eventCollector is a test helper that collects all events.
type eventCollector struct {
	events []OutputEvent
}

func (c *eventCollector) WriteEvent(event OutputEvent) {
	c.events = append(c.events, event)
}

func (c *eventCollector) Flush() {}
