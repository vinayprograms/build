package output

import (
	"testing"
	"time"
)

func TestPhaseStartedEventType(t *testing.T) {
	e := PhaseStarted{Phase: "parse", Timestamp: time.Now()}
	if e.eventType() != "phase_started" {
		t.Errorf("expected 'phase_started', got '%s'", e.eventType())
	}
}

func TestPhaseCompletedEventType(t *testing.T) {
	e := PhaseCompleted{Phase: "parse", Timestamp: time.Now(), Duration: time.Second}
	if e.eventType() != "phase_completed" {
		t.Errorf("expected 'phase_completed', got '%s'", e.eventType())
	}
}

func TestVariableEvaluatedEventType(t *testing.T) {
	e := VariableEvaluated{Name: "cc", Expr: "", Result: "gcc"}
	if e.eventType() != "variable_evaluated" {
		t.Errorf("expected 'variable_evaluated', got '%s'", e.eventType())
	}
}

func TestTargetStartedEventType(t *testing.T) {
	e := TargetStarted{Target: "foo.o", Index: 1, Total: 5}
	if e.eventType() != "target_started" {
		t.Errorf("expected 'target_started', got '%s'", e.eventType())
	}
}

func TestTargetCompletedEventType(t *testing.T) {
	e := TargetCompleted{Target: "foo.o", Success: true, Duration: time.Second}
	if e.eventType() != "target_completed" {
		t.Errorf("expected 'target_completed', got '%s'", e.eventType())
	}
}

func TestTargetSkippedEventType(t *testing.T) {
	e := TargetSkipped{Target: "foo.o", Reason: "up to date"}
	if e.eventType() != "target_skipped" {
		t.Errorf("expected 'target_skipped', got '%s'", e.eventType())
	}
}

func TestCommandStartedEventType(t *testing.T) {
	e := CommandStarted{Target: "foo.o", Command: "gcc -c foo.c"}
	if e.eventType() != "command_started" {
		t.Errorf("expected 'command_started', got '%s'", e.eventType())
	}
}

func TestCommandOutputEventType(t *testing.T) {
	e := CommandOutput{Target: "foo.o", Stdout: "output", Stderr: ""}
	if e.eventType() != "command_output" {
		t.Errorf("expected 'command_output', got '%s'", e.eventType())
	}
}

func TestCommandCompletedEventType(t *testing.T) {
	e := CommandCompleted{Target: "foo.o", Command: "gcc", ExitCode: 0, Duration: time.Second}
	if e.eventType() != "command_completed" {
		t.Errorf("expected 'command_completed', got '%s'", e.eventType())
	}
}

func TestStalenessCheckedEventType(t *testing.T) {
	e := StalenessChecked{Target: "foo.o", Reason: "src newer", Action: "rebuild"}
	if e.eventType() != "staleness_checked" {
		t.Errorf("expected 'staleness_checked', got '%s'", e.eventType())
	}
}

func TestBuildSummaryEventType(t *testing.T) {
	e := BuildSummary{Total: 5, Succeeded: 4, Failed: 1, Skipped: 0, Duration: time.Second}
	if e.eventType() != "build_summary" {
		t.Errorf("expected 'build_summary', got '%s'", e.eventType())
	}
}

func TestErrorOccurredEventType(t *testing.T) {
	e := ErrorOccurred{Category: "parse", Code: "E100", Message: "syntax error"}
	if e.eventType() != "error" {
		t.Errorf("expected 'error', got '%s'", e.eventType())
	}
}

func TestDryRunTargetEventType(t *testing.T) {
	e := DryRunTarget{Target: "foo.o", Index: 1, Total: 5}
	if e.eventType() != "dry_run_target" {
		t.Errorf("expected 'dry_run_target', got '%s'", e.eventType())
	}
}

func TestDryRunCommandEventType(t *testing.T) {
	e := DryRunCommand{Target: "foo.o", Command: "gcc -c foo.c"}
	if e.eventType() != "dry_run_command" {
		t.Errorf("expected 'dry_run_command', got '%s'", e.eventType())
	}
}

func TestAllEventsImplementInterface(t *testing.T) {
	// Compile-time check that all events implement OutputEvent
	var _ OutputEvent = PhaseStarted{}
	var _ OutputEvent = PhaseCompleted{}
	var _ OutputEvent = VariableEvaluated{}
	var _ OutputEvent = TargetStarted{}
	var _ OutputEvent = TargetCompleted{}
	var _ OutputEvent = TargetSkipped{}
	var _ OutputEvent = CommandStarted{}
	var _ OutputEvent = CommandOutput{}
	var _ OutputEvent = CommandCompleted{}
	var _ OutputEvent = StalenessChecked{}
	var _ OutputEvent = BuildSummary{}
	var _ OutputEvent = ErrorOccurred{}
	var _ OutputEvent = DryRunTarget{}
	var _ OutputEvent = DryRunCommand{}
}
