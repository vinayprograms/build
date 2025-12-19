package eval

import (
	"bytes"
	"strings"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

// ----------------------------------------------------------------------------
// Variable Evaluation Tests
// ----------------------------------------------------------------------------

func TestEvaluateVariables_EmptyStatements(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	err := e.EvaluateVariables([]ast.Statement{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Only built-ins should be defined
	if len(ctx.Variables()) != 2 { // os, arch
		t.Errorf("Expected 2 built-ins, got %d variables", len(ctx.Variables()))
	}
}

func TestEvaluateVariables_SimpleVariable(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "foo",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.LiteralValue{Text: "bar"},
				},
			},
		},
	}

	err := e.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	val, ok := ctx.Get("foo")
	if !ok {
		t.Fatal("Expected 'foo' to be defined")
	}
	if val != "bar" {
		t.Errorf("Expected 'bar', got '%s'", val)
	}
}

func TestEvaluateVariables_MultipleVariables(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "a",
			Value: &ast.Value{
				Parts: []ast.ValuePart{&ast.LiteralValue{Text: "1"}},
			},
		},
		&ast.Variable{
			Name: "b",
			Value: &ast.Value{
				Parts: []ast.ValuePart{&ast.LiteralValue{Text: "2"}},
			},
		},
		&ast.Variable{
			Name: "c",
			Value: &ast.Value{
				Parts: []ast.ValuePart{&ast.LiteralValue{Text: "3"}},
			},
		},
	}

	err := e.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	for _, name := range []string{"a", "b", "c"} {
		if !ctx.IsDefined(name) {
			t.Errorf("Expected '%s' to be defined", name)
		}
	}
}

func TestEvaluateVariables_VariableReference(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "base",
			Value: &ast.Value{
				Parts: []ast.ValuePart{&ast.LiteralValue{Text: "build"}},
			},
		},
		&ast.Variable{
			Name: "path",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{Name: "base"},
					&ast.LiteralValue{Text: "/app"},
				},
			},
		},
	}

	err := e.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	val, _ := ctx.Get("path")
	if val != "build/app" {
		t.Errorf("Expected 'build/app', got '%s'", val)
	}
}

func TestEvaluateVariables_ForwardReferenceError(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	// 'path' references 'base' which is defined later
	stmts := []ast.Statement{
		&ast.Variable{
			Name: "path",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{
						Name:     "base",
						Location: ast.SourceLocation{File: "test", Line: 1, Column: 1},
					},
					&ast.LiteralValue{Text: "/app"},
				},
			},
			Location: ast.SourceLocation{File: "test", Line: 1, Column: 1},
		},
		&ast.Variable{
			Name: "base",
			Value: &ast.Value{
				Parts: []ast.ValuePart{&ast.LiteralValue{Text: "build"}},
			},
		},
	}

	err := e.EvaluateVariables(stmts)
	if err == nil {
		t.Fatal("Expected error for forward reference")
	}

	// Check it's an undefined variable error
	if _, ok := err.(*UndefinedVariableError); !ok {
		t.Errorf("Expected UndefinedVariableError, got %T: %v", err, err)
	}
}

func TestEvaluateVariables_LazyVariable(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "lazy_var",
			Lazy: true,
			Value: &ast.Value{
				Parts: []ast.ValuePart{&ast.LiteralValue{Text: "lazy value"}},
			},
		},
	}

	err := e.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Lazy variable should not be in regular variables
	_, ok := ctx.Get("lazy_var")
	if ok {
		t.Error("Lazy variable should not be in regular variables")
	}

	// Should be marked as lazy
	if !ctx.IsLazy("lazy_var") {
		t.Error("Expected 'lazy_var' to be marked as lazy")
	}
}

func TestEvaluateVariables_BuiltinReference(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "platform",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{Name: "os"},
					&ast.LiteralValue{Text: "_"},
					&ast.Interpolation{Name: "arch"},
				},
			},
		},
	}

	err := e.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	os, _ := ctx.Get("os")
	arch, _ := ctx.Get("arch")
	expected := os + "_" + arch

	val, _ := ctx.Get("platform")
	if val != expected {
		t.Errorf("Expected '%s', got '%s'", expected, val)
	}
}

func TestEvaluateVariables_SkipsNonVariables(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	// Mix of statements - only variables should be evaluated
	stmts := []ast.Statement{
		&ast.Comment{Text: "# This is a comment"},
		&ast.Variable{
			Name: "foo",
			Value: &ast.Value{
				Parts: []ast.ValuePart{&ast.LiteralValue{Text: "bar"}},
			},
		},
		&ast.Blank{},
		&ast.Directive{
			Kind:  ast.DirectiveShell,
			Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "bash"}}},
		},
	}

	err := e.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Only 'foo' and built-ins should be defined
	if !ctx.IsDefined("foo") {
		t.Error("Expected 'foo' to be defined")
	}
}

func TestEvaluateVariables_ChainedReferences(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "a",
			Value: &ast.Value{
				Parts: []ast.ValuePart{&ast.LiteralValue{Text: "1"}},
			},
		},
		&ast.Variable{
			Name: "b",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{Name: "a"},
					&ast.LiteralValue{Text: "2"},
				},
			},
		},
		&ast.Variable{
			Name: "c",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{Name: "b"},
					&ast.LiteralValue{Text: "3"},
				},
			},
		},
	}

	err := e.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	tests := []struct {
		name     string
		expected string
	}{
		{"a", "1"},
		{"b", "12"},
		{"c", "123"},
	}

	for _, tt := range tests {
		val, _ := ctx.Get(tt.name)
		if val != tt.expected {
			t.Errorf("%s: expected '%s', got '%s'", tt.name, tt.expected, val)
		}
	}
}

// ----------------------------------------------------------------------------
// Lazy Variable Evaluation Tests
// ----------------------------------------------------------------------------

func TestEvaluateVariables_LazyVariableOnDemand(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	// Define a lazy variable and an immediate variable that references it
	stmts := []ast.Statement{
		&ast.Variable{
			Name: "lazy_greeting",
			Lazy: true,
			Value: &ast.Value{
				Parts: []ast.ValuePart{&ast.LiteralValue{Text: "Hello, World!"}},
			},
		},
	}

	err := e.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Lazy variable should not be evaluated yet
	_, ok := ctx.Get("lazy_greeting")
	if ok {
		t.Error("Lazy variable should not be evaluated until referenced")
	}

	// Now evaluate a value that references the lazy variable
	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.Interpolation{Name: "lazy_greeting"},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error evaluating lazy reference: %v", err)
	}

	if result != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", result)
	}

	// After evaluation, the lazy value should be cached
	cached, ok := ctx.Get("lazy_greeting")
	if !ok {
		t.Error("Lazy variable should be cached after evaluation")
	}
	if cached != "Hello, World!" {
		t.Errorf("Cached value should be 'Hello, World!', got '%s'", cached)
	}
}

func TestEvaluateVariables_LazyVariableWithInterpolation(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	// Define immediate variable first, then lazy variable that uses it
	stmts := []ast.Statement{
		&ast.Variable{
			Name: "greeting",
			Value: &ast.Value{
				Parts: []ast.ValuePart{&ast.LiteralValue{Text: "Hello"}},
			},
		},
		&ast.Variable{
			Name: "message",
			Lazy: true,
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{Name: "greeting"},
					&ast.LiteralValue{Text: ", World!"},
				},
			},
		},
	}

	err := e.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Evaluate the lazy variable
	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.Interpolation{Name: "message"},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", result)
	}
}

func TestEvaluateVariables_LazyVariableReferencesLater(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	// Lazy variable can reference a later-defined immediate variable
	// because lazy evaluation happens at point of use
	stmts := []ast.Statement{
		&ast.Variable{
			Name: "lazy_path",
			Lazy: true,
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{Name: "base"},
					&ast.LiteralValue{Text: "/app"},
				},
			},
		},
		&ast.Variable{
			Name: "base",
			Value: &ast.Value{
				Parts: []ast.ValuePart{&ast.LiteralValue{Text: "build"}},
			},
		},
	}

	err := e.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Now reference the lazy variable - it should work because 'base' is now defined
	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.Interpolation{Name: "lazy_path"},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != "build/app" {
		t.Errorf("Expected 'build/app', got '%s'", result)
	}
}

func TestEvaluateVariables_LazyVariableCaching(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "lazy_var",
			Lazy: true,
			Value: &ast.Value{
				Parts: []ast.ValuePart{&ast.LiteralValue{Text: "cached value"}},
			},
		},
	}

	err := e.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// First reference - should evaluate
	val := &ast.Value{Parts: []ast.ValuePart{&ast.Interpolation{Name: "lazy_var"}}}
	result1, _ := e.EvaluateValue(val)

	// Second reference - should use cache
	result2, _ := e.EvaluateValue(val)

	if result1 != result2 {
		t.Errorf("Cached value should be consistent: %s != %s", result1, result2)
	}

	if result1 != "cached value" {
		t.Errorf("Expected 'cached value', got '%s'", result1)
	}
}

func TestEvaluateVariables_WithConditional(t *testing.T) {
	ctx := NewContext()
	ctx.Set("platform", "linux") // Set a test platform
	e := NewEvaluator(ctx)

	stmts := []ast.Statement{
		&ast.Conditional{
			IfBranch: ast.ConditionalBranch{
				Condition: &ast.EqualsCondition{
					Left:  &ast.Value{Parts: []ast.ValuePart{&ast.Interpolation{Name: "platform"}}},
					Right: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "linux"}}},
				},
				Body: []ast.Statement{
					&ast.Variable{
						Name:  "cc",
						Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "gcc"}}},
					},
				},
			},
			ElseBody: []ast.Statement{
				&ast.Variable{
					Name:  "cc",
					Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "clang"}}},
				},
			},
		},
	}

	err := e.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	val, ok := ctx.Get("cc")
	if !ok {
		t.Error("Expected 'cc' to be defined")
	}
	if val != "gcc" {
		t.Errorf("Expected 'gcc', got '%s'", val)
	}
}

// ----------------------------------------------------------------------------
// Verbose Mode Tests
// ----------------------------------------------------------------------------

func TestEvaluateVariables_VerboseMode(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	var buf bytes.Buffer
	e.SetVerboseOutput(&buf)

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "foo",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.LiteralValue{Text: "bar"},
				},
			},
		},
		&ast.Variable{
			Name: "greeting",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.LiteralValue{Text: "Hello, "},
					&ast.Interpolation{Name: "foo"},
				},
			},
		},
	}

	err := e.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	// Should show variable evaluation results
	if !strings.Contains(output, "foo = bar") {
		t.Errorf("Expected verbose output to contain 'foo = bar', got: %s", output)
	}
	if !strings.Contains(output, "greeting = Hello, bar") {
		t.Errorf("Expected verbose output to contain 'greeting = Hello, bar', got: %s", output)
	}
}
