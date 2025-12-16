package eval

import (
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

// ----------------------------------------------------------------------------
// Value Evaluation Tests
// ----------------------------------------------------------------------------

func TestEvaluateValue_Literal(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.LiteralValue{Text: "hello world"},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", result)
	}
}

func TestEvaluateValue_MultipleLiterals(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.LiteralValue{Text: "hello "},
			&ast.LiteralValue{Text: "world"},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", result)
	}
}

func TestEvaluateValue_Interpolation(t *testing.T) {
	ctx := NewContext()
	ctx.Set("name", "Alice")
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.LiteralValue{Text: "hello "},
			&ast.Interpolation{Name: "name"},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "hello Alice" {
		t.Errorf("Expected 'hello Alice', got '%s'", result)
	}
}

func TestEvaluateValue_BuiltinInterpolation(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.LiteralValue{Text: "OS: "},
			&ast.Interpolation{Name: "os"},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should contain the actual OS
	os, _ := ctx.Get("os")
	expected := "OS: " + os
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestEvaluateValue_UndefinedVariable(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.Interpolation{
				Name:     "undefined",
				Location: ast.SourceLocation{File: "test", Line: 1, Column: 1},
			},
		},
	}

	_, err := e.EvaluateValue(val)
	if err == nil {
		t.Fatal("Expected error for undefined variable")
	}

	// Check it's the right error type
	if _, ok := err.(*UndefinedVariableError); !ok {
		t.Errorf("Expected UndefinedVariableError, got %T", err)
	}
}

func TestEvaluateValue_MixedLiteralsAndInterpolations(t *testing.T) {
	ctx := NewContext()
	ctx.Set("dir", "build")
	ctx.Set("name", "app")
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.Interpolation{Name: "dir"},
			&ast.LiteralValue{Text: "/"},
			&ast.Interpolation{Name: "name"},
			&ast.LiteralValue{Text: ".exe"},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "build/app.exe" {
		t.Errorf("Expected 'build/app.exe', got '%s'", result)
	}
}

func TestEvaluateValue_NilValue(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	result, err := e.EvaluateValue(nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty string for nil value, got '%s'", result)
	}
}

func TestEvaluateValue_EmptyValue(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{Parts: []ast.ValuePart{}}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty string for empty value, got '%s'", result)
	}
}

// ----------------------------------------------------------------------------
// Interpolation with :raw Modifier Tests
// ----------------------------------------------------------------------------

func TestEvaluateValue_RawModifier(t *testing.T) {
	ctx := NewContext()
	ctx.Set("flags", "-Wall -Wextra")
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.Interpolation{Name: "flags", Raw: true},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Raw should still substitute the value (raw affects command execution, not evaluation)
	if result != "-Wall -Wextra" {
		t.Errorf("Expected '-Wall -Wextra', got '%s'", result)
	}
}

// ----------------------------------------------------------------------------
// Error Type Tests
// ----------------------------------------------------------------------------

func TestUndefinedVariableError_Error(t *testing.T) {
	err := &UndefinedVariableError{
		Name:     "missing",
		Location: ast.SourceLocation{File: "Buildfile", Line: 5, Column: 10},
	}

	msg := err.Error()
	if msg == "" {
		t.Error("Expected non-empty error message")
	}
	if !containsString(msg, "missing") {
		t.Errorf("Expected error to mention variable name, got: %s", msg)
	}
	if !containsString(msg, "Buildfile:5:10") {
		t.Errorf("Expected error to include location, got: %s", msg)
	}
}

// containsString checks if s contains substr
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || containsString(s[1:], substr)))
}
