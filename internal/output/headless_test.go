package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestHeadlessWriter_TargetStarted_Text(t *testing.T) {
	var buf bytes.Buffer
	w := NewHeadlessWriter(&buf, WriterConfig{LogFormat: "text", LogLevel: "info"})

	w.WriteEvent(TargetStarted{Target: "foo.o", Index: 1, Total: 5})

	got := buf.String()
	if !strings.Contains(got, "[INFO]") {
		t.Errorf("expected [INFO], got %q", got)
	}
	if !strings.Contains(got, "Building target") {
		t.Errorf("expected 'Building target', got %q", got)
	}
	if !strings.Contains(got, "target=foo.o") {
		t.Errorf("expected target=foo.o, got %q", got)
	}
}

func TestHeadlessWriter_TargetStarted_JSON(t *testing.T) {
	var buf bytes.Buffer
	w := NewHeadlessWriter(&buf, WriterConfig{LogFormat: "json", LogLevel: "info"})

	w.WriteEvent(TargetStarted{Target: "foo.o", Index: 1, Total: 5})

	got := buf.String()
	var entry map[string]any
	if err := json.Unmarshal([]byte(got), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry["level"] != "INFO" {
		t.Errorf("expected level=INFO, got %v", entry["level"])
	}
	if entry["target"] != "foo.o" {
		t.Errorf("expected target=foo.o, got %v", entry["target"])
	}
}

func TestHeadlessWriter_Error(t *testing.T) {
	var buf bytes.Buffer
	w := NewHeadlessWriter(&buf, WriterConfig{LogFormat: "text", LogLevel: "info"})

	w.WriteEvent(ErrorOccurred{
		Category: "parse",
		Code:     "E100",
		Message:  "syntax error",
	})

	got := buf.String()
	if !strings.Contains(got, "[ERROR]") {
		t.Errorf("expected [ERROR], got %q", got)
	}
	if !strings.Contains(got, "syntax error") {
		t.Errorf("expected 'syntax error', got %q", got)
	}
}

func TestHeadlessWriter_LogLevel_Debug(t *testing.T) {
	var buf bytes.Buffer
	w := NewHeadlessWriter(&buf, WriterConfig{LogFormat: "text", LogLevel: "debug"})

	w.WriteEvent(CommandStarted{Target: "foo.o", Command: "gcc -c foo.c"})

	got := buf.String()
	if !strings.Contains(got, "[DEBUG]") {
		t.Errorf("expected debug output, got %q", got)
	}
}

func TestHeadlessWriter_LogLevel_Info_FiltersDebug(t *testing.T) {
	var buf bytes.Buffer
	w := NewHeadlessWriter(&buf, WriterConfig{LogFormat: "text", LogLevel: "info"})

	w.WriteEvent(CommandStarted{Target: "foo.o", Command: "gcc -c foo.c"})

	got := buf.String()
	if got != "" {
		t.Errorf("expected no output for debug at info level, got %q", got)
	}
}

func TestHeadlessWriter_CommandOutput(t *testing.T) {
	var buf bytes.Buffer
	w := NewHeadlessWriter(&buf, WriterConfig{LogFormat: "text", LogLevel: "info"})

	w.WriteEvent(CommandOutput{Target: "foo.o", Stdout: "hello\n", Stderr: ""})

	got := buf.String()
	if got != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", got)
	}
}

func TestHeadlessWriter_BuildSummary_Success(t *testing.T) {
	var buf bytes.Buffer
	w := NewHeadlessWriter(&buf, WriterConfig{LogFormat: "text", LogLevel: "info"})

	w.WriteEvent(BuildSummary{Total: 5, Succeeded: 5, Failed: 0, Duration: time.Second})

	got := buf.String()
	if !strings.Contains(got, "[INFO]") {
		t.Errorf("expected [INFO], got %q", got)
	}
	if !strings.Contains(got, "Build completed") {
		t.Errorf("expected 'Build completed', got %q", got)
	}
}

func TestHeadlessWriter_BuildSummary_Failure(t *testing.T) {
	var buf bytes.Buffer
	w := NewHeadlessWriter(&buf, WriterConfig{LogFormat: "text", LogLevel: "info"})

	w.WriteEvent(BuildSummary{Total: 5, Succeeded: 3, Failed: 2, Duration: time.Second})

	got := buf.String()
	if !strings.Contains(got, "[ERROR]") {
		t.Errorf("expected [ERROR], got %q", got)
	}
	if !strings.Contains(got, "Build failed") {
		t.Errorf("expected 'Build failed', got %q", got)
	}
}

func TestHeadlessWriter_Quiet(t *testing.T) {
	var buf bytes.Buffer
	w := NewHeadlessWriter(&buf, WriterConfig{LogFormat: "text", LogLevel: "info", Quiet: true})

	w.WriteEvent(TargetStarted{Target: "foo.o", Index: 1, Total: 1})
	w.WriteEvent(BuildSummary{Total: 1, Succeeded: 1, Failed: 0})

	got := buf.String()
	if got != "" {
		t.Errorf("expected no output in quiet mode, got %q", got)
	}
}

func TestHeadlessWriter_Quiet_ShowsErrors(t *testing.T) {
	var buf bytes.Buffer
	w := NewHeadlessWriter(&buf, WriterConfig{LogFormat: "text", LogLevel: "info", Quiet: true})

	w.WriteEvent(ErrorOccurred{Code: "E100", Message: "error"})

	got := buf.String()
	if got == "" {
		t.Error("expected error output in quiet mode")
	}
}

func TestHeadlessWriter_Timestamp(t *testing.T) {
	var buf bytes.Buffer
	w := NewHeadlessWriter(&buf, WriterConfig{LogFormat: "text", LogLevel: "info"})

	w.WriteEvent(TargetStarted{Target: "foo.o", Index: 1, Total: 1})

	got := buf.String()
	// Should have RFC3339 timestamp
	if !strings.Contains(got, "T") && !strings.Contains(got, "Z") {
		t.Errorf("expected timestamp, got %q", got)
	}
}
