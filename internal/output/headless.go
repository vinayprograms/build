package output

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// HeadlessWriter implements OutputWriter for CI/CD and log collection.
// It produces timestamped, plain-text output without ANSI codes.
type HeadlessWriter struct {
	w        io.Writer
	config   WriterConfig
	useJSON  bool
	logLevel logLevel
}

type logLevel int

const (
	logDebug logLevel = iota
	logInfo
	logWarn
	logError
)

// NewHeadlessWriter creates a new HeadlessWriter.
func NewHeadlessWriter(w io.Writer, config WriterConfig) *HeadlessWriter {
	return &HeadlessWriter{
		w:        w,
		config:   config,
		useJSON:  config.LogFormat == "json",
		logLevel: parseLogLevel(config.LogLevel),
	}
}

func parseLogLevel(s string) logLevel {
	switch s {
	case "debug":
		return logDebug
	case "info":
		return logInfo
	case "warn":
		return logWarn
	case "error":
		return logError
	default:
		return logInfo
	}
}

// WriteEvent renders an output event to the log stream.
func (h *HeadlessWriter) WriteEvent(event OutputEvent) {
	if h.config.Quiet && !isErrorEvent(event) {
		return
	}

	switch e := event.(type) {
	case PhaseStarted:
		h.log(logInfo, "Starting phase", map[string]any{"phase": e.Phase})
	case PhaseCompleted:
		h.log(logInfo, "Phase completed", map[string]any{
			"phase":       e.Phase,
			"duration_ms": e.Duration.Milliseconds(),
		})
	case VariableEvaluated:
		h.log(logDebug, "Variable evaluated", map[string]any{
			"name":   e.Name,
			"result": e.Result,
		})
	case TargetStarted:
		h.log(logInfo, "Building target", map[string]any{
			"target": e.Target,
			"index":  e.Index,
			"total":  e.Total,
		})
	case TargetCompleted:
		if e.Success {
			h.log(logInfo, "Target built", map[string]any{
				"target":      e.Target,
				"duration_ms": e.Duration.Milliseconds(),
			})
		} else {
			h.log(logError, "Target failed", map[string]any{
				"target": e.Target,
				"error":  e.Error,
			})
		}
	case TargetSkipped:
		h.log(logInfo, "Target skipped", map[string]any{
			"target": e.Target,
			"reason": e.Reason,
		})
	case CommandStarted:
		h.log(logDebug, "Command started", map[string]any{
			"target":  e.Target,
			"command": e.Command,
		})
	case CommandOutput:
		if e.Stdout != "" {
			fmt.Fprint(h.w, e.Stdout)
			if len(e.Stdout) > 0 && e.Stdout[len(e.Stdout)-1] != '\n' {
				fmt.Fprintln(h.w)
			}
		}
		if e.Stderr != "" {
			fmt.Fprint(h.w, e.Stderr)
			if len(e.Stderr) > 0 && e.Stderr[len(e.Stderr)-1] != '\n' {
				fmt.Fprintln(h.w)
			}
		}
	case CommandCompleted:
		level := logDebug
		if e.ExitCode != 0 {
			level = logError
		}
		h.log(level, "Command completed", map[string]any{
			"target":      e.Target,
			"command":     e.Command,
			"exit_code":   e.ExitCode,
			"duration_ms": e.Duration.Milliseconds(),
		})
	case StalenessChecked:
		h.log(logDebug, "Staleness checked", map[string]any{
			"target": e.Target,
			"reason": e.Reason,
			"action": e.Action,
		})
	case BuildSummary:
		if e.Failed == 0 {
			h.log(logInfo, "Build completed", map[string]any{
				"total":       e.Total,
				"succeeded":   e.Succeeded,
				"skipped":     e.Skipped,
				"duration_ms": e.Duration.Milliseconds(),
			})
		} else {
			h.log(logError, "Build failed", map[string]any{
				"total":       e.Total,
				"succeeded":   e.Succeeded,
				"failed":      e.Failed,
				"skipped":     e.Skipped,
				"duration_ms": e.Duration.Milliseconds(),
			})
		}
	case ErrorOccurred:
		h.log(logError, e.Message, map[string]any{
			"category": e.Category,
			"code":     e.Code,
			"location": e.Location,
			"hint":     e.Hint,
		})
	case DryRunTarget:
		h.log(logInfo, "Would build target", map[string]any{
			"target": e.Target,
			"index":  e.Index,
			"total":  e.Total,
		})
	case DryRunCommand:
		h.log(logInfo, "Would run command", map[string]any{
			"target":  e.Target,
			"command": e.Command,
		})
	}
}

// Flush ensures all output is written.
func (h *HeadlessWriter) Flush() {
	// No buffering in this implementation
}

func (h *HeadlessWriter) log(level logLevel, msg string, fields map[string]any) {
	if level < h.logLevel {
		return
	}

	if h.useJSON {
		h.logJSON(level, msg, fields)
	} else {
		h.logText(level, msg, fields)
	}
}

func (h *HeadlessWriter) logJSON(level logLevel, msg string, fields map[string]any) {
	entry := map[string]any{
		"time":  time.Now().Format(time.RFC3339),
		"level": levelString(level),
		"msg":   msg,
	}
	for k, v := range fields {
		if v != nil && v != "" && v != 0 {
			entry[k] = v
		}
	}
	data, _ := json.Marshal(entry)
	fmt.Fprintln(h.w, string(data))
}

func (h *HeadlessWriter) logText(level logLevel, msg string, fields map[string]any) {
	timestamp := time.Now().Format(time.RFC3339)
	levelStr := levelString(level)

	// Build a simple key=value string for important fields
	extra := ""
	for k, v := range fields {
		if v != nil && v != "" && v != 0 {
			if extra == "" {
				extra = fmt.Sprintf(" %s=%v", k, v)
			} else {
				extra += fmt.Sprintf(" %s=%v", k, v)
			}
		}
	}

	fmt.Fprintf(h.w, "[%s] [%s] %s%s\n", timestamp, levelStr, msg, extra)
}

func levelString(level logLevel) string {
	switch level {
	case logDebug:
		return "DEBUG"
	case logInfo:
		return "INFO"
	case logWarn:
		return "WARN"
	case logError:
		return "ERROR"
	default:
		return "INFO"
	}
}
