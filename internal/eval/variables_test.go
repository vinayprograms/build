package eval

import (
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
