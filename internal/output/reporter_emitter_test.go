package output

import (
	"bytes"
	"strings"
	"testing"
)

// Tests for Reporter implementations backed by the OutputWriter system.
// These tests verify that the Reporter interface correctly delegates to
// the new event-based output system.

// ----------------------------------------------------------------------------
// NormalReporter backed by Emitter tests
// ----------------------------------------------------------------------------

func TestEmitterBackedNormalReporter_BuildStarted(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedNormalReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.BuildStarted("build/app")

	got := buf.String()
	if !strings.Contains(got, "build/app") {
		t.Errorf("BuildStarted output should contain target name, got: %q", got)
	}
	if !strings.Contains(strings.ToLower(got), "building") {
		t.Errorf("BuildStarted output should indicate building, got: %q", got)
	}
}

func TestEmitterBackedNormalReporter_BuildCompleted(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedNormalReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.BuildCompleted("build/app", true, "")

	got := buf.String()
	if !strings.Contains(got, "build/app") {
		t.Errorf("BuildCompleted should contain target name, got: %q", got)
	}
	if !strings.Contains(strings.ToLower(got), "built") {
		t.Errorf("BuildCompleted success should indicate built, got: %q", got)
	}
}

func TestEmitterBackedNormalReporter_BuildCompletedFailure(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedNormalReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.BuildCompleted("build/app", false, "compile error: undefined reference")

	got := buf.String()
	if !strings.Contains(got, "build/app") {
		t.Errorf("BuildCompleted failure should contain target name, got: %q", got)
	}
	if !strings.Contains(strings.ToLower(got), "fail") {
		t.Errorf("BuildCompleted failure should indicate failure, got: %q", got)
	}
	if !strings.Contains(got, "compile error") {
		t.Errorf("BuildCompleted failure should contain error message, got: %q", got)
	}
}

func TestEmitterBackedNormalReporter_CommandOutput(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedNormalReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.CommandOutput("gcc -c main.c", "main.c:10: warning: unused variable", "")

	got := buf.String()
	if !strings.Contains(got, "warning: unused variable") {
		t.Errorf("CommandOutput should contain stdout content, got: %q", got)
	}
}

func TestEmitterBackedNormalReporter_CommandOutputStderr(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedNormalReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.CommandOutput("gcc -c main.c", "", "error: expected ';'")

	got := buf.String()
	if !strings.Contains(got, "error: expected ';'") {
		t.Errorf("CommandOutput should contain stderr content, got: %q", got)
	}
}

func TestEmitterBackedNormalReporter_SuppressesEmptyOutput(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedNormalReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	// Empty output should not produce any output
	r.CommandOutput("echo test", "", "")

	got := buf.String()
	if got != "" {
		t.Errorf("CommandOutput with empty stdout/stderr should produce no output, got: %q", got)
	}
}

func TestEmitterBackedNormalReporter_Summary(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedNormalReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.Summary(5, 1)

	got := buf.String()
	if !strings.Contains(got, "5") || !strings.Contains(got, "1") {
		t.Errorf("Summary should contain target counts, got: %q", got)
	}
}

func TestEmitterBackedNormalReporter_SummaryAllSuccess(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedNormalReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.Summary(5, 0)

	got := buf.String()
	if !strings.Contains(strings.ToLower(got), "success") {
		t.Errorf("Summary with no failures should indicate success, got: %q", got)
	}
}

func TestEmitterBackedNormalReporter_NothingToBuild(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedNormalReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.NothingToBuild("build/app")

	got := buf.String()
	if !strings.Contains(got, "build/app") {
		t.Errorf("NothingToBuild should contain target name, got: %q", got)
	}
	if !strings.Contains(strings.ToLower(got), "up to date") {
		t.Errorf("NothingToBuild should indicate target is current, got: %q", got)
	}
}

// ----------------------------------------------------------------------------
// DryRunReporter backed by Emitter tests
// ----------------------------------------------------------------------------

func TestEmitterBackedDryRunReporter_WouldBuild(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedDryRunReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.WouldBuild("build/app")

	got := buf.String()
	if !strings.Contains(got, "Would build: build/app") {
		t.Errorf("WouldBuild should output 'Would build: target', got: %q", got)
	}
}

func TestEmitterBackedDryRunReporter_ShowCommand(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedDryRunReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	// Need to set target context first
	r.WouldBuild("build/app")
	buf.Reset()
	r.ShowCommand("gcc -o build/app main.c")

	got := buf.String()
	if !strings.Contains(got, "gcc -o build/app main.c") {
		t.Errorf("ShowCommand should show command, got: %q", got)
	}
}

func TestEmitterBackedDryRunReporter_NothingToBuild(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedDryRunReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.NothingToBuild("build/app")

	got := buf.String()
	if !strings.Contains(got, "build/app") {
		t.Errorf("NothingToBuild should contain target name, got: %q", got)
	}
	if !strings.Contains(strings.ToLower(got), "up to date") {
		t.Errorf("NothingToBuild should indicate target is current, got: %q", got)
	}
}

// ----------------------------------------------------------------------------
// VerboseReporter backed by Emitter tests
// ----------------------------------------------------------------------------

func TestEmitterBackedVerboseReporter_StartVariableEvaluation(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedVerboseReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.StartVariableEvaluation()

	got := buf.String()
	if !strings.Contains(strings.ToLower(got), "evaluat") {
		t.Errorf("StartVariableEvaluation should indicate evaluation phase, got: %q", got)
	}
}

func TestEmitterBackedVerboseReporter_VariableEvaluated(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedVerboseReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.VariableEvaluated("sources", "shell(find src -name \"*.c\")", "src/main.c src/utils.c")

	got := buf.String()
	if !strings.Contains(got, "sources") {
		t.Errorf("VariableEvaluated should contain var name, got: %q", got)
	}
	if !strings.Contains(got, "src/main.c src/utils.c") {
		t.Errorf("VariableEvaluated should contain result value, got: %q", got)
	}
}

func TestEmitterBackedVerboseReporter_StartStalenessChecks(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedVerboseReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.StartStalenessChecks()

	got := buf.String()
	if !strings.Contains(strings.ToLower(got), "check") || !strings.Contains(strings.ToLower(got), "target") {
		t.Errorf("StartStalenessChecks should indicate checking phase, got: %q", got)
	}
}

func TestEmitterBackedVerboseReporter_StalenessCheck(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedVerboseReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.StalenessCheck("build/main.o", "src/main.c is newer", "rebuild")

	got := buf.String()
	if !strings.Contains(got, "build/main.o") {
		t.Errorf("StalenessCheck should contain target name, got: %q", got)
	}
	if !strings.Contains(got, "src/main.c is newer") {
		t.Errorf("StalenessCheck should contain reason, got: %q", got)
	}
	if !strings.Contains(got, "rebuild") {
		t.Errorf("StalenessCheck should contain action, got: %q", got)
	}
}

func TestEmitterBackedVerboseReporter_BuildStarted(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedVerboseReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.BuildStarted("build/main.o")

	got := buf.String()
	if !strings.Contains(got, "build/main.o") {
		t.Errorf("BuildStarted should contain target name, got: %q", got)
	}
	if !strings.Contains(strings.ToLower(got), "building") {
		t.Errorf("BuildStarted should indicate building, got: %q", got)
	}
}

func TestEmitterBackedVerboseReporter_CommandExecuted(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedVerboseReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.CommandExecuted("gcc -Wall -O2 -c src/main.c -o build/main.o")

	got := buf.String()
	if !strings.Contains(got, "gcc") {
		t.Errorf("CommandExecuted should contain command, got: %q", got)
	}
}

func TestEmitterBackedVerboseReporter_Summary(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedVerboseReporter(&buf, WriterConfig{Color: "never", Verbose: true})

	r.Summary(3, 0)

	got := buf.String()
	// Should indicate done/success
	if !strings.Contains(strings.ToLower(got), "success") && !strings.Contains(got, "3") {
		t.Errorf("Summary with no failures should indicate success, got: %q", got)
	}
}

// ----------------------------------------------------------------------------
// ProgressReporter backed by Emitter tests
// ----------------------------------------------------------------------------

func TestEmitterBackedProgressReporter_BuildStarted(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedProgressReporter(&buf, WriterConfig{Color: "never", Verbose: true}, 10)

	r.BuildStarted("build/main.o")

	got := buf.String()
	if !strings.Contains(got, "build/main.o") {
		t.Errorf("BuildStarted should contain target name, got: %q", got)
	}
	// Should show progress count
	if !strings.Contains(got, "[1/10]") {
		t.Errorf("BuildStarted should show progress count, got: %q", got)
	}
}

func TestEmitterBackedProgressReporter_BuildStartedMultiple(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedProgressReporter(&buf, WriterConfig{Color: "never", Verbose: true}, 5)

	r.BuildStarted("build/a.o")
	r.BuildStarted("build/b.o")
	r.BuildStarted("build/c.o")

	got := buf.String()
	// Should show incremented counts
	if !strings.Contains(got, "[1/5]") || !strings.Contains(got, "[2/5]") || !strings.Contains(got, "[3/5]") {
		t.Errorf("BuildStarted should show incremented counts, got: %q", got)
	}
}

func TestEmitterBackedProgressReporter_BuildCompletedFailure(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedProgressReporter(&buf, WriterConfig{Color: "never", Verbose: true}, 5)

	r.BuildStarted("build/main.o")
	buf.Reset()
	r.BuildCompleted("build/main.o", false, "compile error")

	got := buf.String()
	if !strings.Contains(got, "build/main.o") {
		t.Errorf("BuildCompleted failure should contain target name, got: %q", got)
	}
	if !strings.Contains(strings.ToLower(got), "fail") {
		t.Errorf("BuildCompleted failure should indicate failure, got: %q", got)
	}
}

func TestEmitterBackedProgressReporter_Summary(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedProgressReporter(&buf, WriterConfig{Color: "never", Verbose: true}, 5)

	r.Summary(5, 0)

	got := buf.String()
	if !strings.Contains(strings.ToLower(got), "success") {
		t.Errorf("Summary should indicate success, got: %q", got)
	}
}

func TestEmitterBackedProgressReporter_CurrentlyBuilding(t *testing.T) {
	var buf bytes.Buffer
	r := NewEmitterBackedProgressReporter(&buf, WriterConfig{Color: "never", Verbose: true}, 10)

	// Simulate parallel builds
	r.BuildStarted("build/a.o")
	r.BuildStarted("build/b.o")

	// Get currently building
	building := r.CurrentlyBuilding()
	if len(building) != 2 {
		t.Errorf("CurrentlyBuilding should return 2 targets, got %d", len(building))
	}

	// Complete one
	r.BuildCompleted("build/a.o", true, "")
	building = r.CurrentlyBuilding()
	if len(building) != 1 {
		t.Errorf("CurrentlyBuilding should return 1 target after completion, got %d", len(building))
	}
}

// ----------------------------------------------------------------------------
// Compatibility tests - ensure new reporters produce similar output to old
// ----------------------------------------------------------------------------

func TestReporterCompatibility_NormalBuildFlow(t *testing.T) {
	// Test that the emitter-backed reporter produces functionally equivalent output
	var oldBuf, newBuf bytes.Buffer

	// Old reporter
	oldR := NewNormalReporter(&oldBuf)
	oldR.BuildStarted("build/app")
	oldR.BuildCompleted("build/app", true, "")
	oldR.Summary(1, 0)

	// New reporter
	newR := NewEmitterBackedNormalReporter(&newBuf, WriterConfig{Color: "never", Verbose: true})
	newR.BuildStarted("build/app")
	newR.BuildCompleted("build/app", true, "")
	newR.Summary(1, 0)

	oldOut := oldBuf.String()
	newOut := newBuf.String()

	// Both should contain the target name
	if !strings.Contains(oldOut, "build/app") || !strings.Contains(newOut, "build/app") {
		t.Errorf("Both reporters should mention target.\nOld: %q\nNew: %q", oldOut, newOut)
	}

	// Both should indicate success
	if !strings.Contains(strings.ToLower(oldOut), "success") || !strings.Contains(strings.ToLower(newOut), "success") {
		t.Errorf("Both reporters should indicate success.\nOld: %q\nNew: %q", oldOut, newOut)
	}
}
