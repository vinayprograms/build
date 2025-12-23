package eval

import (
	"testing"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/output"
)

// TestEvaluator_EmitsVariableEvaluated tests that evaluator emits variable events.
func TestEvaluator_EmitsVariableEvaluated(t *testing.T) {
	collector := &eventCollector{}
	emitter := output.NewEmitter(collector)

	ctx := NewContext()
	eval := NewEvaluator(ctx)
	eval.SetEmitter(emitter)

	stmts := []ast.Statement{
		&ast.Variable{
			Name:  "cc",
			Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "gcc"}}},
		},
		&ast.Variable{
			Name:  "flags",
			Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "-O2"}}},
		},
	}

	err := eval.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have emitted 2 VariableEvaluated events
	if len(collector.events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(collector.events))
	}

	// Check first event
	v1, ok := collector.events[0].(output.VariableEvaluated)
	if !ok {
		t.Fatalf("expected VariableEvaluated, got %T", collector.events[0])
	}
	if v1.Name != "cc" || v1.Result != "gcc" {
		t.Errorf("expected cc=gcc, got %s=%s", v1.Name, v1.Result)
	}

	// Check second event
	v2, ok := collector.events[1].(output.VariableEvaluated)
	if !ok {
		t.Fatalf("expected VariableEvaluated, got %T", collector.events[1])
	}
	if v2.Name != "flags" || v2.Result != "-O2" {
		t.Errorf("expected flags=-O2, got %s=%s", v2.Name, v2.Result)
	}
}

// TestEvaluator_EmitsLazyVariable tests that lazy variables emit with <lazy> marker.
func TestEvaluator_EmitsLazyVariable(t *testing.T) {
	collector := &eventCollector{}
	emitter := output.NewEmitter(collector)

	ctx := NewContext()
	eval := NewEvaluator(ctx)
	eval.SetEmitter(emitter)

	stmts := []ast.Statement{
		&ast.Variable{
			Name:  "src",
			Lazy:  true,
			Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "main.c"}}},
		},
	}

	err := eval.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have emitted 1 event
	if len(collector.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(collector.events))
	}

	v, ok := collector.events[0].(output.VariableEvaluated)
	if !ok {
		t.Fatalf("expected VariableEvaluated, got %T", collector.events[0])
	}
	if v.Name != "src" {
		t.Errorf("expected name 'src', got '%s'", v.Name)
	}
	if v.Result != "<lazy>" {
		t.Errorf("expected result '<lazy>', got '%s'", v.Result)
	}
}

// TestEvaluator_EmitsFunctionExpression tests that function expressions are shown.
func TestEvaluator_EmitsFunctionExpression(t *testing.T) {
	collector := &eventCollector{}
	emitter := output.NewEmitter(collector)

	ctx := NewContext()
	eval := NewEvaluator(ctx)
	eval.SetEmitter(emitter)

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "base",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.FunctionCall{
						Name: ast.FuncFilename,
						Args: []*ast.Value{
							{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "/path/to/file.c"}}},
						},
					},
				},
			},
		},
	}

	err := eval.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(collector.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(collector.events))
	}

	v, ok := collector.events[0].(output.VariableEvaluated)
	if !ok {
		t.Fatalf("expected VariableEvaluated, got %T", collector.events[0])
	}
	if v.Expr == "" {
		t.Error("expected expr to be non-empty for function call")
	}
	if v.Result != "file.c" {
		t.Errorf("expected result 'file.c', got '%s'", v.Result)
	}
}

// TestEvaluator_NilEmitter tests that evaluator works without an emitter.
func TestEvaluator_NilEmitter(t *testing.T) {
	ctx := NewContext()
	eval := NewEvaluator(ctx)
	// No emitter set

	stmts := []ast.Statement{
		&ast.Variable{
			Name:  "cc",
			Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "gcc"}}},
		},
	}

	err := eval.EvaluateVariables(stmts)
	if err != nil {
		t.Fatalf("should work without emitter: %v", err)
	}
}

// eventCollector collects output events for testing.
type eventCollector struct {
	events []output.OutputEvent
}

func (c *eventCollector) WriteEvent(event output.OutputEvent) {
	c.events = append(c.events, event)
}

func (c *eventCollector) Flush() {}
