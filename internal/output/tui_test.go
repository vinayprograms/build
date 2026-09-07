package output

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestTUIWriter_TargetStarted(t *testing.T) {
	var buf bytes.Buffer
	w := NewTUIWriter(&buf)

	w.WriteEvent(TargetStarted{Target: "foo.o", Index: 1, Total: 5})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry["type"] != "target_started" {
		t.Errorf("expected type=target_started, got %v", entry["type"])
	}
	if entry["target"] != "foo.o" {
		t.Errorf("expected target=foo.o, got %v", entry["target"])
	}
	if entry["index"] != float64(1) {
		t.Errorf("expected index=1, got %v", entry["index"])
	}
	if entry["total"] != float64(5) {
		t.Errorf("expected total=5, got %v", entry["total"])
	}
}

func TestTUIWriter_TargetCompleted(t *testing.T) {
	var buf bytes.Buffer
	w := NewTUIWriter(&buf)

	w.WriteEvent(TargetCompleted{
		Target:   "foo.o",
		Success:  true,
		Duration: 250 * time.Millisecond,
	})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry["type"] != "target_completed" {
		t.Errorf("expected type=target_completed, got %v", entry["type"])
	}
	if entry["success"] != true {
		t.Errorf("expected success=true, got %v", entry["success"])
	}
	if entry["duration_ms"] != float64(250) {
		t.Errorf("expected duration_ms=250, got %v", entry["duration_ms"])
	}
}

func TestTUIWriter_Error(t *testing.T) {
	var buf bytes.Buffer
	w := NewTUIWriter(&buf)

	w.WriteEvent(ErrorOccurred{
		Category: "parse",
		Code:     "E100",
		Message:  "syntax error",
		Location: "Needfile:10:5",
		Hint:     "check syntax",
	})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry["type"] != "error" {
		t.Errorf("expected type=error, got %v", entry["type"])
	}
	if entry["category"] != "parse" {
		t.Errorf("expected category=parse, got %v", entry["category"])
	}
	if entry["code"] != "E100" {
		t.Errorf("expected code=E100, got %v", entry["code"])
	}
}

func TestTUIWriter_BuildSummary(t *testing.T) {
	var buf bytes.Buffer
	w := NewTUIWriter(&buf)

	w.WriteEvent(BuildSummary{
		Total:     5,
		Succeeded: 4,
		Failed:    1,
		Skipped:   0,
		Duration:  time.Second,
	})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry["type"] != "build_summary" {
		t.Errorf("expected type=build_summary, got %v", entry["type"])
	}
	if entry["total"] != float64(5) {
		t.Errorf("expected total=5, got %v", entry["total"])
	}
	if entry["failed"] != float64(1) {
		t.Errorf("expected failed=1, got %v", entry["failed"])
	}
}

func TestTUIWriter_HasTimestamp(t *testing.T) {
	var buf bytes.Buffer
	w := NewTUIWriter(&buf)

	w.WriteEvent(TargetStarted{Target: "foo.o", Index: 1, Total: 1})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := entry["timestamp"]; !ok {
		t.Error("expected timestamp field")
	}
}

func TestTUIWriter_PhaseEvents(t *testing.T) {
	var buf bytes.Buffer
	w := NewTUIWriter(&buf)

	now := time.Now()
	w.WriteEvent(PhaseStarted{Phase: "parse", Timestamp: now})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry["type"] != "phase_started" {
		t.Errorf("expected type=phase_started, got %v", entry["type"])
	}
	if entry["phase"] != "parse" {
		t.Errorf("expected phase=parse, got %v", entry["phase"])
	}
}

func TestTUIWriter_DryRunEvents(t *testing.T) {
	var buf bytes.Buffer
	w := NewTUIWriter(&buf)

	w.WriteEvent(DryRunTarget{Target: "foo.o", Index: 1, Total: 1})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry["type"] != "dry_run_target" {
		t.Errorf("expected type=dry_run_target, got %v", entry["type"])
	}
}
