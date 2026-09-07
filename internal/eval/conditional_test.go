package eval

import (
	"testing"

	"github.com/vinayprograms/need/internal/ast"
)

// ----------------------------------------------------------------------------
// Condition Evaluation Tests
// ----------------------------------------------------------------------------

func TestEvaluateCondition_Equals_True(t *testing.T) {
	ctx := NewContext()
	ctx.Set("platform", "linux")
	e := NewEvaluator(ctx)

	cond := &ast.EqualsCondition{
		Left:  &ast.Value{Parts: []ast.ValuePart{&ast.Interpolation{Name: "platform"}}},
		Right: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "linux"}}},
	}

	result, err := e.EvaluateCondition(cond)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result {
		t.Error("Expected condition to be true")
	}
}

func TestEvaluateCondition_Equals_False(t *testing.T) {
	ctx := NewContext()
	ctx.Set("platform", "darwin")
	e := NewEvaluator(ctx)

	cond := &ast.EqualsCondition{
		Left:  &ast.Value{Parts: []ast.ValuePart{&ast.Interpolation{Name: "platform"}}},
		Right: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "linux"}}},
	}

	result, err := e.EvaluateCondition(cond)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result {
		t.Error("Expected condition to be false")
	}
}

func TestEvaluateCondition_NotEquals_True(t *testing.T) {
	ctx := NewContext()
	ctx.Set("platform", "darwin")
	e := NewEvaluator(ctx)

	cond := &ast.NotEqualsCondition{
		Left:  &ast.Value{Parts: []ast.ValuePart{&ast.Interpolation{Name: "platform"}}},
		Right: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "windows"}}},
	}

	result, err := e.EvaluateCondition(cond)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result {
		t.Error("Expected condition to be true")
	}
}

func TestEvaluateCondition_NotEquals_False(t *testing.T) {
	ctx := NewContext()
	ctx.Set("platform", "windows")
	e := NewEvaluator(ctx)

	cond := &ast.NotEqualsCondition{
		Left:  &ast.Value{Parts: []ast.ValuePart{&ast.Interpolation{Name: "platform"}}},
		Right: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "windows"}}},
	}

	result, err := e.EvaluateCondition(cond)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result {
		t.Error("Expected condition to be false")
	}
}

func TestEvaluateCondition_Defined_True(t *testing.T) {
	ctx := NewContext()
	ctx.Set("MY_VAR", "value")
	e := NewEvaluator(ctx)

	cond := &ast.DefinedCondition{Name: "MY_VAR"}

	result, err := e.EvaluateCondition(cond)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result {
		t.Error("Expected defined condition to be true")
	}
}

func TestEvaluateCondition_Defined_False(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	cond := &ast.DefinedCondition{Name: "UNDEFINED_VAR"}

	result, err := e.EvaluateCondition(cond)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result {
		t.Error("Expected defined condition to be false")
	}
}

func TestEvaluateCondition_NotDefined_True(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	cond := &ast.NotDefinedCondition{Name: "UNDEFINED_VAR"}

	result, err := e.EvaluateCondition(cond)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result {
		t.Error("Expected not-defined condition to be true")
	}
}

func TestEvaluateCondition_NotDefined_False(t *testing.T) {
	ctx := NewContext()
	ctx.Set("MY_VAR", "value")
	e := NewEvaluator(ctx)

	cond := &ast.NotDefinedCondition{Name: "MY_VAR"}

	result, err := e.EvaluateCondition(cond)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result {
		t.Error("Expected not-defined condition to be false")
	}
}

func TestEvaluateCondition_BuiltinDefined(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	// os is a built-in, should be defined
	cond := &ast.DefinedCondition{Name: "os"}

	result, err := e.EvaluateCondition(cond)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result {
		t.Error("Expected os to be defined")
	}
}

// ----------------------------------------------------------------------------
// Conditional Block Evaluation Tests
// ----------------------------------------------------------------------------

func TestEvaluateConditional_IfTrue(t *testing.T) {
	ctx := NewContext()
	ctx.Set("platform", "linux")
	e := NewEvaluator(ctx)

	conditional := &ast.Conditional{
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
	}

	err := e.EvaluateConditional(conditional)
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

func TestEvaluateConditional_IfFalse_NoElse(t *testing.T) {
	ctx := NewContext()
	ctx.Set("platform", "darwin")
	e := NewEvaluator(ctx)

	conditional := &ast.Conditional{
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
	}

	err := e.EvaluateConditional(conditional)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// cc should not be defined since condition was false
	if ctx.IsDefined("cc") {
		t.Error("Expected 'cc' to NOT be defined")
	}
}

func TestEvaluateConditional_ElifBranch(t *testing.T) {
	ctx := NewContext()
	ctx.Set("platform", "darwin")
	e := NewEvaluator(ctx)

	conditional := &ast.Conditional{
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
		ElifBranches: []ast.ConditionalBranch{
			{
				Condition: &ast.EqualsCondition{
					Left:  &ast.Value{Parts: []ast.ValuePart{&ast.Interpolation{Name: "platform"}}},
					Right: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "darwin"}}},
				},
				Body: []ast.Statement{
					&ast.Variable{
						Name:  "cc",
						Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "clang"}}},
					},
				},
			},
		},
	}

	err := e.EvaluateConditional(conditional)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	val, ok := ctx.Get("cc")
	if !ok {
		t.Error("Expected 'cc' to be defined")
	}
	if val != "clang" {
		t.Errorf("Expected 'clang', got '%s'", val)
	}
}

func TestEvaluateConditional_ElseBranch(t *testing.T) {
	ctx := NewContext()
	ctx.Set("platform", "windows")
	e := NewEvaluator(ctx)

	conditional := &ast.Conditional{
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
		ElifBranches: []ast.ConditionalBranch{
			{
				Condition: &ast.EqualsCondition{
					Left:  &ast.Value{Parts: []ast.ValuePart{&ast.Interpolation{Name: "platform"}}},
					Right: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "darwin"}}},
				},
				Body: []ast.Statement{
					&ast.Variable{
						Name:  "cc",
						Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "clang"}}},
					},
				},
			},
		},
		ElseBody: []ast.Statement{
			&ast.Variable{
				Name:  "cc",
				Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "msvc"}}},
			},
		},
	}

	err := e.EvaluateConditional(conditional)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	val, ok := ctx.Get("cc")
	if !ok {
		t.Error("Expected 'cc' to be defined")
	}
	if val != "msvc" {
		t.Errorf("Expected 'msvc', got '%s'", val)
	}
}

func TestEvaluateConditional_Ifdef(t *testing.T) {
	ctx := NewContext()
	ctx.Set("DEBUG", "1")
	e := NewEvaluator(ctx)

	conditional := &ast.Conditional{
		IfBranch: ast.ConditionalBranch{
			Condition: &ast.DefinedCondition{Name: "DEBUG"},
			Body: []ast.Statement{
				&ast.Variable{
					Name:  "cflags",
					Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "-g"}}},
				},
			},
		},
	}

	err := e.EvaluateConditional(conditional)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	val, ok := ctx.Get("cflags")
	if !ok {
		t.Error("Expected 'cflags' to be defined")
	}
	if val != "-g" {
		t.Errorf("Expected '-g', got '%s'", val)
	}
}

func TestEvaluateConditional_Ifndef(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	conditional := &ast.Conditional{
		IfBranch: ast.ConditionalBranch{
			Condition: &ast.NotDefinedCondition{Name: "CC"},
			Body: []ast.Statement{
				&ast.Variable{
					Name:  "cc",
					Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "cc"}}},
				},
			},
		},
	}

	err := e.EvaluateConditional(conditional)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	val, ok := ctx.Get("cc")
	if !ok {
		t.Error("Expected 'cc' to be defined")
	}
	if val != "cc" {
		t.Errorf("Expected 'cc', got '%s'", val)
	}
}

func TestEvaluateConditional_NestedConditionals(t *testing.T) {
	ctx := NewContext()
	ctx.Set("platform", "linux")
	ctx.Set("DEBUG", "1")
	e := NewEvaluator(ctx)

	conditional := &ast.Conditional{
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
				&ast.Conditional{
					IfBranch: ast.ConditionalBranch{
						Condition: &ast.DefinedCondition{Name: "DEBUG"},
						Body: []ast.Statement{
							&ast.Variable{
								Name:  "cflags",
								Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "-g -O0"}}},
							},
						},
					},
					ElseBody: []ast.Statement{
						&ast.Variable{
							Name:  "cflags",
							Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "-O2"}}},
						},
					},
				},
			},
		},
	}

	err := e.EvaluateConditional(conditional)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	cc, _ := ctx.Get("cc")
	if cc != "gcc" {
		t.Errorf("Expected cc='gcc', got '%s'", cc)
	}

	cflags, _ := ctx.Get("cflags")
	if cflags != "-g -O0" {
		t.Errorf("Expected cflags='-g -O0', got '%s'", cflags)
	}
}
