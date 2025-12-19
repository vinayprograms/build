package output

import (
	"bytes"
	"testing"
)

func TestNewWriter_CLI(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(ModeCLI, &buf, DefaultWriterConfig())

	_, ok := w.(*CLIWriter)
	if !ok {
		t.Errorf("expected *CLIWriter, got %T", w)
	}
}

func TestNewWriter_TUI(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(ModeTUI, &buf, DefaultWriterConfig())

	_, ok := w.(*TUIWriter)
	if !ok {
		t.Errorf("expected *TUIWriter, got %T", w)
	}
}

func TestNewWriter_Headless(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(ModeHeadless, &buf, DefaultWriterConfig())

	_, ok := w.(*HeadlessWriter)
	if !ok {
		t.Errorf("expected *HeadlessWriter, got %T", w)
	}
}

func TestNewNoOpWriter(t *testing.T) {
	w := NewNoOpWriter()

	// Should not panic
	w.WriteEvent(TargetStarted{Target: "foo.o"})
	w.Flush()
}

func TestDefaultWriterConfig(t *testing.T) {
	config := DefaultWriterConfig()

	if config.Verbose {
		t.Error("expected Verbose=false")
	}
	if config.Quiet {
		t.Error("expected Quiet=false")
	}
	if config.Color != "auto" {
		t.Errorf("expected Color=auto, got %s", config.Color)
	}
	if config.LogLevel != "info" {
		t.Errorf("expected LogLevel=info, got %s", config.LogLevel)
	}
	if config.LogFormat != "text" {
		t.Errorf("expected LogFormat=text, got %s", config.LogFormat)
	}
}
