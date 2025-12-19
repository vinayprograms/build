package output

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// TUIWriter implements OutputWriter for structured terminal UI output.
// It produces JSON events that can be consumed by a TUI application.
type TUIWriter struct {
	w io.Writer
}

// NewTUIWriter creates a new TUIWriter.
func NewTUIWriter(w io.Writer) *TUIWriter {
	return &TUIWriter{w: w}
}

// WriteEvent outputs a JSON event.
func (t *TUIWriter) WriteEvent(event OutputEvent) {
	data := t.eventToJSON(event)
	if data != nil {
		jsonBytes, _ := json.Marshal(data)
		fmt.Fprintln(t.w, string(jsonBytes))
	}
}

// Flush ensures all output is written.
func (t *TUIWriter) Flush() {
	// No buffering in this implementation
}

func (t *TUIWriter) eventToJSON(event OutputEvent) map[string]any {
	now := time.Now()

	switch e := event.(type) {
	case PhaseStarted:
		return map[string]any{
			"type":      "phase_started",
			"phase":     e.Phase,
			"timestamp": e.Timestamp.Format(time.RFC3339),
		}
	case PhaseCompleted:
		return map[string]any{
			"type":        "phase_completed",
			"phase":       e.Phase,
			"timestamp":   e.Timestamp.Format(time.RFC3339),
			"duration_ms": e.Duration.Milliseconds(),
		}
	case VariableEvaluated:
		return map[string]any{
			"type":      "variable_evaluated",
			"name":      e.Name,
			"expr":      e.Expr,
			"result":    e.Result,
			"timestamp": now.Format(time.RFC3339),
		}
	case TargetStarted:
		return map[string]any{
			"type":      "target_started",
			"target":    e.Target,
			"index":     e.Index,
			"total":     e.Total,
			"timestamp": now.Format(time.RFC3339),
		}
	case TargetCompleted:
		return map[string]any{
			"type":        "target_completed",
			"target":      e.Target,
			"success":     e.Success,
			"duration_ms": e.Duration.Milliseconds(),
			"error":       e.Error,
			"timestamp":   now.Format(time.RFC3339),
		}
	case TargetSkipped:
		return map[string]any{
			"type":      "target_skipped",
			"target":    e.Target,
			"reason":    e.Reason,
			"timestamp": now.Format(time.RFC3339),
		}
	case CommandStarted:
		return map[string]any{
			"type":      "command_started",
			"target":    e.Target,
			"command":   e.Command,
			"timestamp": now.Format(time.RFC3339),
		}
	case CommandOutput:
		return map[string]any{
			"type":      "command_output",
			"target":    e.Target,
			"stdout":    e.Stdout,
			"stderr":    e.Stderr,
			"timestamp": now.Format(time.RFC3339),
		}
	case CommandCompleted:
		return map[string]any{
			"type":        "command_completed",
			"target":      e.Target,
			"command":     e.Command,
			"exit_code":   e.ExitCode,
			"duration_ms": e.Duration.Milliseconds(),
			"timestamp":   now.Format(time.RFC3339),
		}
	case StalenessChecked:
		return map[string]any{
			"type":      "staleness_checked",
			"target":    e.Target,
			"reason":    e.Reason,
			"action":    e.Action,
			"timestamp": now.Format(time.RFC3339),
		}
	case BuildSummary:
		return map[string]any{
			"type":        "build_summary",
			"total":       e.Total,
			"succeeded":   e.Succeeded,
			"failed":      e.Failed,
			"skipped":     e.Skipped,
			"duration_ms": e.Duration.Milliseconds(),
			"timestamp":   now.Format(time.RFC3339),
		}
	case ErrorOccurred:
		return map[string]any{
			"type":      "error",
			"category":  e.Category,
			"code":      e.Code,
			"message":   e.Message,
			"location":  e.Location,
			"context":   e.Context,
			"hint":      e.Hint,
			"timestamp": now.Format(time.RFC3339),
		}
	case DryRunTarget:
		return map[string]any{
			"type":      "dry_run_target",
			"target":    e.Target,
			"index":     e.Index,
			"total":     e.Total,
			"timestamp": now.Format(time.RFC3339),
		}
	case DryRunCommand:
		return map[string]any{
			"type":      "dry_run_command",
			"target":    e.Target,
			"command":   e.Command,
			"timestamp": now.Format(time.RFC3339),
		}
	default:
		return nil
	}
}
