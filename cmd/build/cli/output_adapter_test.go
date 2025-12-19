package cli

import (
	"testing"
)

func TestCreateOutputEmitter(t *testing.T) {
	// Test that CreateOutputEmitter returns a valid emitter
	emitter := CreateOutputEmitter(false, false, "auto")
	if emitter == nil {
		t.Fatal("expected non-nil emitter")
	}
}

func TestCreateOutputEmitter_Verbose(t *testing.T) {
	emitter := CreateOutputEmitter(true, false, "auto")
	if emitter == nil {
		t.Fatal("expected non-nil emitter")
	}
}

func TestCreateOutputEmitter_Quiet(t *testing.T) {
	emitter := CreateOutputEmitter(false, true, "auto")
	if emitter == nil {
		t.Fatal("expected non-nil emitter")
	}
}

func TestCreateOutputEmitter_ColorNever(t *testing.T) {
	emitter := CreateOutputEmitter(false, false, "never")
	if emitter == nil {
		t.Fatal("expected non-nil emitter")
	}
}

func TestCreateOutputWriter(t *testing.T) {
	writer := CreateOutputWriter(false, false, "auto")
	if writer == nil {
		t.Fatal("expected non-nil writer")
	}
}
