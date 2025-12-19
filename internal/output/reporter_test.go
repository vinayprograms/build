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
