package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestNormalReporter_BuildStarted(t *testing.T) {
	var buf bytes.Buffer
	r := NewNormalReporter(&buf)

	r.BuildStarted("build/app")

	got := buf.String()
	if !strings.Contains(got, "build/app") {
		t.Errorf("BuildStarted output should contain target name, got: %q", got)
	}
}

func TestNormalReporter_BuildCompleted(t *testing.T) {
	var buf bytes.Buffer
	r := NewNormalReporter(&buf)

	r.BuildCompleted("build/app", true, "")

	got := buf.String()
	// Success message should be present
	if got == "" {
		t.Error("BuildCompleted should produce output for success")
	}
}

func TestNormalReporter_BuildCompletedFailure(t *testing.T) {
	var buf bytes.Buffer
	r := NewNormalReporter(&buf)

	r.BuildCompleted("build/app", false, "compile error: undefined reference")

	got := buf.String()
	if !strings.Contains(got, "build/app") {
		t.Errorf("BuildCompleted failure should contain target name, got: %q", got)
	}
	if !strings.Contains(got, "error") && !strings.Contains(got, "failed") && !strings.Contains(got, "FAILED") {
		t.Errorf("BuildCompleted failure should indicate failure, got: %q", got)
	}
}

func TestNormalReporter_CommandOutput(t *testing.T) {
	var buf bytes.Buffer
	r := NewNormalReporter(&buf)

	r.CommandOutput("gcc -c main.c", "main.c:10: warning: unused variable", "")

	got := buf.String()
	if !strings.Contains(got, "warning: unused variable") {
		t.Errorf("CommandOutput should contain stdout content, got: %q", got)
	}
}

func TestNormalReporter_CommandOutputStderr(t *testing.T) {
	var buf bytes.Buffer
	r := NewNormalReporter(&buf)

	r.CommandOutput("gcc -c main.c", "", "error: expected ';'")

	got := buf.String()
	if !strings.Contains(got, "error: expected ';'") {
		t.Errorf("CommandOutput should contain stderr content, got: %q", got)
	}
}

func TestNormalReporter_SuppressesEmptyOutput(t *testing.T) {
	var buf bytes.Buffer
	r := NewNormalReporter(&buf)

	// Empty output should not produce any output
	r.CommandOutput("echo test", "", "")

	got := buf.String()
	if got != "" {
		t.Errorf("CommandOutput with empty stdout/stderr should produce no output, got: %q", got)
	}
}

func TestNormalReporter_Summary(t *testing.T) {
	var buf bytes.Buffer
	r := NewNormalReporter(&buf)

	r.Summary(5, 1)

	got := buf.String()
	// Should show counts
	if !strings.Contains(got, "5") || !strings.Contains(got, "1") {
		t.Errorf("Summary should contain target counts, got: %q", got)
	}
}

func TestNormalReporter_SummaryAllSuccess(t *testing.T) {
	var buf bytes.Buffer
	r := NewNormalReporter(&buf)

	r.Summary(5, 0)

	got := buf.String()
	// Should indicate success
	if !strings.Contains(strings.ToLower(got), "success") && !strings.Contains(got, "5") {
		t.Errorf("Summary with no failures should indicate success, got: %q", got)
	}
}

func TestNormalReporter_NothingToBuild(t *testing.T) {
	var buf bytes.Buffer
	r := NewNormalReporter(&buf)

	r.NothingToBuild("build/app")

	got := buf.String()
	if !strings.Contains(got, "build/app") {
		t.Errorf("NothingToBuild should contain target name, got: %q", got)
	}
	// Should indicate up-to-date or nothing to do
	if !strings.Contains(strings.ToLower(got), "up to date") &&
		!strings.Contains(strings.ToLower(got), "nothing") &&
		!strings.Contains(strings.ToLower(got), "up-to-date") {
		t.Errorf("NothingToBuild should indicate target is current, got: %q", got)
	}
}

// ----------------------------------------------------------------------------
// DryRunReporter Tests
// ----------------------------------------------------------------------------

func TestDryRunReporter_WouldBuild(t *testing.T) {
	var buf bytes.Buffer
	r := NewDryRunReporter(&buf)

	r.WouldBuild("build/app")

	got := buf.String()
	if !strings.Contains(got, "Would build: build/app") {
		t.Errorf("WouldBuild should output 'Would build: target', got: %q", got)
	}
}

func TestDryRunReporter_ShowCommand(t *testing.T) {
	var buf bytes.Buffer
	r := NewDryRunReporter(&buf)

	r.ShowCommand("gcc -o build/app main.c")

	got := buf.String()
	// Commands should be indented
	if !strings.Contains(got, "  gcc -o build/app main.c") {
		t.Errorf("ShowCommand should indent command with 2 spaces, got: %q", got)
	}
}

func TestDryRunReporter_ShowCommandMultiple(t *testing.T) {
	var buf bytes.Buffer
	r := NewDryRunReporter(&buf)

	r.ShowCommand("echo hello")
	r.ShowCommand("gcc -c main.c")

	got := buf.String()
	if !strings.Contains(got, "  echo hello") {
		t.Errorf("ShowCommand should indent first command, got: %q", got)
	}
	if !strings.Contains(got, "  gcc -c main.c") {
		t.Errorf("ShowCommand should indent second command, got: %q", got)
	}
}

func TestDryRunReporter_WouldBuildWithCommands(t *testing.T) {
	var buf bytes.Buffer
	r := NewDryRunReporter(&buf)

	r.WouldBuild("build/main.o")
	r.ShowCommand("echo \"Compiling main.c...\"")
	r.ShowCommand("gcc -Wall -O2 -c src/main.c -o build/main.o")

	got := buf.String()
	// Should have proper formatting per spec
	want := "Would build: build/main.o\n  echo \"Compiling main.c...\"\n  gcc -Wall -O2 -c src/main.c -o build/main.o\n"
	if got != want {
		t.Errorf("DryRunReporter output mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestDryRunReporter_BlankLineBetweenTargets(t *testing.T) {
	var buf bytes.Buffer
	r := NewDryRunReporter(&buf)

	r.WouldBuild("build/")
	r.ShowCommand("mkdir -p build/")
	r.TargetComplete()

	r.WouldBuild("build/main.o")
	r.ShowCommand("gcc -c src/main.c -o build/main.o")

	got := buf.String()
	// Should have blank line between targets
	if !strings.Contains(got, "mkdir -p build/\n\nWould build:") {
		t.Errorf("Should have blank line between targets, got: %q", got)
	}
}

func TestDryRunReporter_Summary(t *testing.T) {
	var buf bytes.Buffer
	r := NewDryRunReporter(&buf)

	r.Summary(3)

	got := buf.String()
	if !strings.Contains(got, "3") {
		t.Errorf("Summary should contain count, got: %q", got)
	}
	// Should indicate dry-run / would build
	if !strings.Contains(strings.ToLower(got), "would") || !strings.Contains(got, "3") {
		t.Errorf("Summary should indicate dry-run mode, got: %q", got)
	}
}

func TestDryRunReporter_SummarySingular(t *testing.T) {
	var buf bytes.Buffer
	r := NewDryRunReporter(&buf)

	r.Summary(1)

	got := buf.String()
	// Should use singular form
	if strings.Contains(got, "targets") && !strings.Contains(got, "target ") {
		t.Errorf("Summary with 1 should use singular 'target', got: %q", got)
	}
}

func TestDryRunReporter_NothingToBuild(t *testing.T) {
	var buf bytes.Buffer
	r := NewDryRunReporter(&buf)

	r.NothingToBuild("build/app")

	got := buf.String()
	if !strings.Contains(got, "build/app") {
		t.Errorf("NothingToBuild should contain target name, got: %q", got)
	}
	// Should indicate up-to-date
	if !strings.Contains(strings.ToLower(got), "up to date") &&
		!strings.Contains(strings.ToLower(got), "up-to-date") {
		t.Errorf("NothingToBuild should indicate target is current, got: %q", got)
	}
}

// Table-driven test for different output scenarios
func TestNormalReporter_OutputScenarios(t *testing.T) {
	tests := []struct {
		name        string
		action      func(r *NormalReporter)
		wantContain []string
	}{
		{
			name: "phony target build",
			action: func(r *NormalReporter) {
				r.BuildStarted("@test")
			},
			wantContain: []string{"@test"},
		},
		{
			name: "file target build",
			action: func(r *NormalReporter) {
				r.BuildStarted("build/output.o")
			},
			wantContain: []string{"build/output.o"},
		},
		{
			name: "command with stdout and stderr",
			action: func(r *NormalReporter) {
				r.CommandOutput("make all", "Building...", "Warning: something")
			},
			wantContain: []string{"Building...", "Warning: something"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			r := NewNormalReporter(&buf)
			tt.action(r)

			got := buf.String()
			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("output should contain %q, got: %q", want, got)
				}
			}
		})
	}
}

// ----------------------------------------------------------------------------
// VerboseReporter Tests
// ----------------------------------------------------------------------------

func TestVerboseReporter_VariableEvaluation(t *testing.T) {
	var buf bytes.Buffer
	r := NewVerboseReporter(&buf)

	r.VariableEvaluated("sources", "shell(find src -name \"*.c\")", "src/main.c src/utils.c")

	got := buf.String()
	if !strings.Contains(got, "sources") {
		t.Errorf("VariableEvaluated should contain var name, got: %q", got)
	}
	if !strings.Contains(got, "src/main.c src/utils.c") {
		t.Errorf("VariableEvaluated should contain result value, got: %q", got)
	}
}

func TestVerboseReporter_VariableEvaluationHeader(t *testing.T) {
	var buf bytes.Buffer
	r := NewVerboseReporter(&buf)

	r.StartVariableEvaluation()

	got := buf.String()
	if !strings.Contains(strings.ToLower(got), "evaluat") && !strings.Contains(strings.ToLower(got), "variable") {
		t.Errorf("StartVariableEvaluation should indicate evaluation phase, got: %q", got)
	}
}

func TestVerboseReporter_StalenessCheck(t *testing.T) {
	var buf bytes.Buffer
	r := NewVerboseReporter(&buf)

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

func TestVerboseReporter_StalenessCheckUpToDate(t *testing.T) {
	var buf bytes.Buffer
	r := NewVerboseReporter(&buf)

	r.StalenessCheck("build/utils.o", "up to date", "skip")

	got := buf.String()
	if !strings.Contains(got, "build/utils.o") {
		t.Errorf("StalenessCheck should contain target name, got: %q", got)
	}
	if !strings.Contains(strings.ToLower(got), "up to date") {
		t.Errorf("StalenessCheck should indicate up to date, got: %q", got)
	}
	if !strings.Contains(got, "skip") {
		t.Errorf("StalenessCheck should contain skip action, got: %q", got)
	}
}

func TestVerboseReporter_StalenessCheckHeader(t *testing.T) {
	var buf bytes.Buffer
	r := NewVerboseReporter(&buf)

	r.StartStalenessChecks()

	got := buf.String()
	if !strings.Contains(strings.ToLower(got), "check") || !strings.Contains(strings.ToLower(got), "target") {
		t.Errorf("StartStalenessChecks should indicate staleness check phase, got: %q", got)
	}
}

func TestVerboseReporter_BuildStarted(t *testing.T) {
	var buf bytes.Buffer
	r := NewVerboseReporter(&buf)

	r.BuildStarted("build/main.o")

	got := buf.String()
	if !strings.Contains(got, "build/main.o") {
		t.Errorf("BuildStarted should contain target name, got: %q", got)
	}
	// Should indicate building
	if !strings.Contains(strings.ToLower(got), "building") {
		t.Errorf("BuildStarted should indicate building, got: %q", got)
	}
}

func TestVerboseReporter_CommandExecuted(t *testing.T) {
	var buf bytes.Buffer
	r := NewVerboseReporter(&buf)

	r.CommandExecuted("gcc -Wall -O2 -c src/main.c -o build/main.o")

	got := buf.String()
	// Command should be indented
	if !strings.Contains(got, "  gcc") {
		t.Errorf("CommandExecuted should indent command, got: %q", got)
	}
}

func TestVerboseReporter_Summary(t *testing.T) {
	var buf bytes.Buffer
	r := NewVerboseReporter(&buf)

	r.Summary(3, 0)

	got := buf.String()
	// Should indicate done/success
	if !strings.Contains(strings.ToLower(got), "done") && !strings.Contains(strings.ToLower(got), "success") {
		t.Errorf("Summary with no failures should indicate success, got: %q", got)
	}
}

func TestVerboseReporter_SummaryWithFailures(t *testing.T) {
	var buf bytes.Buffer
	r := NewVerboseReporter(&buf)

	r.Summary(3, 1)

	got := buf.String()
	if !strings.Contains(got, "3") || !strings.Contains(got, "1") {
		t.Errorf("Summary should contain counts, got: %q", got)
	}
}

func TestVerboseReporter_CompleteOutput(t *testing.T) {
	var buf bytes.Buffer
	r := NewVerboseReporter(&buf)

	// Simulate complete verbose build output
	r.StartVariableEvaluation()
	r.VariableEvaluated("sources", "shell(find src -name \"*.c\")", "src/main.c src/utils.c")
	r.VariableEvaluated("objects", "", "build/main.o build/utils.o")

	r.StartStalenessChecks()
	r.StalenessCheck("build/main.o", "src/main.c is newer", "rebuild")
	r.StalenessCheck("build/utils.o", "up to date", "skip")
	r.StalenessCheck("build/app", "build/main.o changed", "rebuild")

	r.BuildStarted("build/main.o")
	r.CommandExecuted("gcc -Wall -O2 -c src/main.c -o build/main.o")

	r.BuildStarted("build/app")
	r.CommandExecuted("gcc -o build/app build/main.o build/utils.o")

	r.Summary(2, 0)

	got := buf.String()
	// Verify key sections are present
	if !strings.Contains(got, "sources") {
		t.Errorf("Should show variable evaluation, got: %q", got)
	}
	if !strings.Contains(got, "build/main.o") {
		t.Errorf("Should show build targets, got: %q", got)
	}
}

// ----------------------------------------------------------------------------
// ProgressReporter Tests (for parallel builds)
// ----------------------------------------------------------------------------

func TestProgressReporter_BuildStarted(t *testing.T) {
	var buf bytes.Buffer
	r := NewProgressReporter(&buf, 10)

	r.BuildStarted("build/main.o")

	got := buf.String()
	if !strings.Contains(got, "build/main.o") {
		t.Errorf("BuildStarted should contain target name, got: %q", got)
	}
	// Should show progress count
	if !strings.Contains(got, "[1/10]") && !strings.Contains(got, "1/10") {
		t.Errorf("BuildStarted should show progress count, got: %q", got)
	}
}

func TestProgressReporter_BuildStartedMultiple(t *testing.T) {
	var buf bytes.Buffer
	r := NewProgressReporter(&buf, 5)

	r.BuildStarted("build/a.o")
	r.BuildStarted("build/b.o")
	r.BuildStarted("build/c.o")

	got := buf.String()
	// Should show incremented counts
	if !strings.Contains(got, "[1/5]") || !strings.Contains(got, "[2/5]") || !strings.Contains(got, "[3/5]") {
		t.Errorf("BuildStarted should show incremented counts, got: %q", got)
	}
}

func TestProgressReporter_BuildCompleted(t *testing.T) {
	var buf bytes.Buffer
	r := NewProgressReporter(&buf, 5)

	r.BuildStarted("build/main.o")
	buf.Reset()
	r.BuildCompleted("build/main.o", true, "")

	// On success, may not produce output (progress is shown on BuildStarted)
	_ = buf.String()
}

func TestProgressReporter_BuildCompletedFailure(t *testing.T) {
	var buf bytes.Buffer
	r := NewProgressReporter(&buf, 5)

	r.BuildStarted("build/main.o")
	buf.Reset()
	r.BuildCompleted("build/main.o", false, "compile error")

	got := buf.String()
	if !strings.Contains(got, "build/main.o") {
		t.Errorf("BuildCompleted failure should contain target name, got: %q", got)
	}
	if !strings.Contains(strings.ToLower(got), "fail") && !strings.Contains(strings.ToLower(got), "error") {
		t.Errorf("BuildCompleted failure should indicate failure, got: %q", got)
	}
}

func TestProgressReporter_Summary(t *testing.T) {
	var buf bytes.Buffer
	r := NewProgressReporter(&buf, 5)

	r.Summary(5, 0)

	got := buf.String()
	if !strings.Contains(got, "5") {
		t.Errorf("Summary should contain count, got: %q", got)
	}
}

func TestProgressReporter_SummaryWithFailures(t *testing.T) {
	var buf bytes.Buffer
	r := NewProgressReporter(&buf, 5)

	r.Summary(5, 2)

	got := buf.String()
	if !strings.Contains(got, "2") || !strings.Contains(got, "5") {
		t.Errorf("Summary should contain failure and total counts, got: %q", got)
	}
}

func TestProgressReporter_CurrentlyBuilding(t *testing.T) {
	var buf bytes.Buffer
	r := NewProgressReporter(&buf, 10)

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
