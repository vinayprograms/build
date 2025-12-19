package output

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestCLIWriter_TargetStarted(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never", Verbose: true})

	w.WriteEvent(TargetStarted{Target: "foo.o", Index: 1, Total: 1})

	got := buf.String()
	if !strings.Contains(got, "Building foo.o") {
		t.Errorf("expected 'Building foo.o', got %q", got)
	}
}

func TestCLIWriter_TargetStarted_Progress(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never", Verbose: true})

	w.WriteEvent(TargetStarted{Target: "foo.o", Index: 3, Total: 5})

	got := buf.String()
	if !strings.Contains(got, "[3/5]") {
		t.Errorf("expected progress [3/5], got %q", got)
	}
}

func TestCLIWriter_TargetCompleted_Success(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never", Verbose: true})

	w.WriteEvent(TargetCompleted{Target: "foo.o", Success: true})

	got := buf.String()
	if !strings.Contains(got, "Built foo.o") {
		t.Errorf("expected 'Built foo.o', got %q", got)
	}
}

func TestCLIWriter_TargetCompleted_Failure(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never"})

	w.WriteEvent(TargetCompleted{Target: "foo.o", Success: false, Error: "compile error"})

	got := buf.String()
	if !strings.Contains(got, "FAILED") {
		t.Errorf("expected 'FAILED', got %q", got)
	}
	if !strings.Contains(got, "compile error") {
		t.Errorf("expected 'compile error', got %q", got)
	}
}

func TestCLIWriter_CommandOutput(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never"})

	w.WriteEvent(CommandOutput{Target: "foo.o", Stdout: "hello\n", Stderr: ""})

	got := buf.String()
	if got != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", got)
	}
}

func TestCLIWriter_BuildSummary_Success(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never", Verbose: true})

	w.WriteEvent(BuildSummary{Total: 5, Succeeded: 5, Failed: 0})

	got := buf.String()
	if !strings.Contains(got, "Build success") {
		t.Errorf("expected 'Build success', got %q", got)
	}
}

func TestCLIWriter_BuildSummary_Failure(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never"})

	w.WriteEvent(BuildSummary{Total: 5, Succeeded: 3, Failed: 2})

	got := buf.String()
	if !strings.Contains(got, "Build failed") {
		t.Errorf("expected 'Build failed', got %q", got)
	}
	if !strings.Contains(got, "2 of 5") {
		t.Errorf("expected '2 of 5', got %q", got)
	}
}

func TestCLIWriter_Error(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never"})

	w.WriteEvent(ErrorOccurred{
		Category: "parse",
		Code:     "E100",
		Message:  "syntax error",
		Location: "Buildfile:10:5",
		Hint:     "check your syntax",
	})

	got := buf.String()
	if !strings.Contains(got, "error") {
		t.Errorf("expected 'error', got %q", got)
	}
	if !strings.Contains(got, "E100") {
		t.Errorf("expected 'E100', got %q", got)
	}
	if !strings.Contains(got, "syntax error") {
		t.Errorf("expected 'syntax error', got %q", got)
	}
	if !strings.Contains(got, "Buildfile:10:5") {
		t.Errorf("expected location, got %q", got)
	}
	if !strings.Contains(got, "help:") {
		t.Errorf("expected 'help:', got %q", got)
	}
}

func TestCLIWriter_Verbose_VariableEvaluated(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never", Verbose: true})

	w.WriteEvent(VariableEvaluated{Name: "cc", Expr: "shell(which gcc)", Result: "/usr/bin/gcc"})

	got := buf.String()
	if !strings.Contains(got, "cc = shell(which gcc)") {
		t.Errorf("expected variable with expression, got %q", got)
	}
	if !strings.Contains(got, "/usr/bin/gcc") {
		t.Errorf("expected result, got %q", got)
	}
}

func TestCLIWriter_Verbose_CommandStarted(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never", Verbose: true})

	w.WriteEvent(CommandStarted{Target: "foo.o", Command: "gcc -c foo.c"})

	got := buf.String()
	if !strings.Contains(got, "gcc -c foo.c") {
		t.Errorf("expected command, got %q", got)
	}
}

func TestCLIWriter_Quiet_SuppressesNonErrors(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never", Quiet: true})

	w.WriteEvent(TargetStarted{Target: "foo.o", Index: 1, Total: 1})
	w.WriteEvent(TargetCompleted{Target: "foo.o", Success: true})
	w.WriteEvent(BuildSummary{Total: 1, Succeeded: 1, Failed: 0})

	got := buf.String()
	if got != "" {
		t.Errorf("expected no output in quiet mode, got %q", got)
	}
}

func TestCLIWriter_Quiet_ShowsErrors(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never", Quiet: true})

	w.WriteEvent(ErrorOccurred{Code: "E100", Message: "error"})

	got := buf.String()
	if got == "" {
		t.Error("expected error output in quiet mode")
	}
}

func TestCLIWriter_DryRun(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never"})

	w.WriteEvent(DryRunTarget{Target: "foo.o", Index: 1, Total: 1})
	w.WriteEvent(DryRunCommand{Target: "foo.o", Command: "gcc -c foo.c"})

	got := buf.String()
	if !strings.Contains(got, "Would build: foo.o") {
		t.Errorf("expected 'Would build: foo.o', got %q", got)
	}
	if !strings.Contains(got, "gcc -c foo.c") {
		t.Errorf("expected command, got %q", got)
	}
}

func TestCLIWriter_WithColor(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "always", Verbose: true})

	w.WriteEvent(TargetStarted{Target: "foo.o", Index: 1, Total: 1})

	got := buf.String()
	// Should contain ANSI escape codes
	if !strings.Contains(got, "\033[") {
		t.Errorf("expected ANSI codes with color=always, got %q", got)
	}
}

func TestCLIWriter_StalenessChecked(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never", Verbose: true})

	w.WriteEvent(StalenessChecked{Target: "foo.o", Reason: "src/foo.c is newer", Action: "rebuild"})

	got := buf.String()
	if !strings.Contains(got, "foo.o") {
		t.Errorf("expected target name, got %q", got)
	}
	if !strings.Contains(got, "src/foo.c is newer") {
		t.Errorf("expected reason, got %q", got)
	}
	if !strings.Contains(got, "rebuild") {
		t.Errorf("expected action, got %q", got)
	}
}

func TestCLIWriter_TargetSkipped(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never", Verbose: true})

	w.WriteEvent(TargetSkipped{Target: "foo.o", Reason: "up to date"})

	got := buf.String()
	if !strings.Contains(got, "foo.o") {
		t.Errorf("expected target name, got %q", got)
	}
	if !strings.Contains(got, "up to date") {
		t.Errorf("expected reason, got %q", got)
	}
}

func TestCLIWriter_VerboseDuration(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never", Verbose: true})

	w.WriteEvent(TargetCompleted{Target: "foo.o", Success: true, Duration: 250 * time.Millisecond})

	got := buf.String()
	if !strings.Contains(got, "250ms") {
		t.Errorf("expected duration, got %q", got)
	}
}

// TestCLIWriter_UnicodeAlways tests that Unicode symbols are used when unicode=always.
func TestCLIWriter_UnicodeAlways(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never", Unicode: "always"})

	// Verify Unicode support is enabled
	if !w.useUnicode {
		t.Error("expected Unicode to be enabled with unicode=always")
	}

	// Get symbols and verify they're Unicode
	sym := w.getSymbols()
	if sym.checkMark != "✓" {
		t.Errorf("expected Unicode checkMark, got %q", sym.checkMark)
	}
	if sym.arrow != "→" {
		t.Errorf("expected Unicode arrow, got %q", sym.arrow)
	}
}

// TestCLIWriter_UnicodeNever tests that ASCII symbols are used when unicode=never.
func TestCLIWriter_UnicodeNever(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never", Unicode: "never"})

	// Verify Unicode support is disabled
	if w.useUnicode {
		t.Error("expected Unicode to be disabled with unicode=never")
	}

	// Get symbols and verify they're ASCII
	sym := w.getSymbols()
	if sym.checkMark != "[ok]" {
		t.Errorf("expected ASCII checkMark, got %q", sym.checkMark)
	}
	if sym.arrow != "->" {
		t.Errorf("expected ASCII arrow, got %q", sym.arrow)
	}
	if sym.crossMark != "[FAIL]" {
		t.Errorf("expected ASCII crossMark, got %q", sym.crossMark)
	}
}

// TestCLIWriter_DegradedOutput tests that degraded output uses ASCII symbols.
func TestCLIWriter_DegradedOutput(t *testing.T) {
	var buf bytes.Buffer
	w := NewCLIWriter(&buf, WriterConfig{Color: "never", Unicode: "never"})

	// All output should be ASCII-safe
	sym := w.getSymbols()
	for _, s := range []string{sym.checkMark, sym.crossMark, sym.arrow, sym.bullet, sym.rightArrow, sym.progressBar} {
		for _, r := range s {
			if r > 127 {
				t.Errorf("expected ASCII-only symbol, got %q with rune %U", s, r)
			}
		}
	}
}

// TestShouldUseUnicode tests the ShouldUseUnicode function.
func TestShouldUseUnicode(t *testing.T) {
	tests := []struct {
		name     string
		setting  string
		expected bool
	}{
		{
			name:     "always",
			setting:  "always",
			expected: true,
		},
		{
			name:     "never",
			setting:  "never",
			expected: false,
		},
		// "auto" depends on locale, so we can't reliably test it
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldUseUnicode(tt.setting)
			if got != tt.expected {
				t.Errorf("ShouldUseUnicode(%q) = %v, want %v", tt.setting, got, tt.expected)
			}
		})
	}
}
