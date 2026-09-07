package parser

import (
	"testing"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/lexer"
)

// TestParseConditionalSimpleIf tests parsing a simple if/end block.
func TestParseConditionalSimpleIf(t *testing.T) {
	input := `if {os} == linux
cc = gcc
end
`
	l := lexer.New("test", input)
	p := New(l)

	cond, err := p.ParseConditional()
	if err != nil {
		t.Fatalf("ParseConditional returned error: %v", err)
	}
	if cond == nil {
		t.Fatal("ParseConditional returned nil")
	}

	// Check if branch condition
	if cond.IfBranch.Condition == nil {
		t.Fatal("IfBranch.Condition is nil")
	}
	eqCond, ok := cond.IfBranch.Condition.(*ast.EqualsCondition)
	if !ok {
		t.Fatalf("expected EqualsCondition, got %T", cond.IfBranch.Condition)
	}

	// Left side should be an interpolation
	if len(eqCond.Left.Parts) != 1 {
		t.Fatalf("expected 1 part in left side, got %d", len(eqCond.Left.Parts))
	}
	interp, ok := eqCond.Left.Parts[0].(*ast.Interpolation)
	if !ok {
		t.Fatalf("expected Interpolation, got %T", eqCond.Left.Parts[0])
	}
	if interp.Name != "os" {
		t.Errorf("expected interpolation name 'os', got %q", interp.Name)
	}

	// Right side should be literal "linux"
	if len(eqCond.Right.Parts) != 1 {
		t.Fatalf("expected 1 part in right side, got %d", len(eqCond.Right.Parts))
	}
	lit, ok := eqCond.Right.Parts[0].(*ast.LiteralValue)
	if !ok {
		t.Fatalf("expected LiteralValue, got %T", eqCond.Right.Parts[0])
	}
	if lit.Text != "linux" {
		t.Errorf("expected literal 'linux', got %q", lit.Text)
	}

	// Check body
	if len(cond.IfBranch.Body) != 1 {
		t.Fatalf("expected 1 statement in body, got %d", len(cond.IfBranch.Body))
	}

	// No elif or else
	if len(cond.ElifBranches) != 0 {
		t.Errorf("expected 0 elif branches, got %d", len(cond.ElifBranches))
	}
	if cond.ElseBody != nil {
		t.Errorf("expected nil ElseBody, got %v", cond.ElseBody)
	}
}

// TestParseConditionalNotEquals tests parsing if with != condition.
func TestParseConditionalNotEquals(t *testing.T) {
	input := `if {os} != windows
cc = gcc
end
`
	l := lexer.New("test", input)
	p := New(l)

	cond, err := p.ParseConditional()
	if err != nil {
		t.Fatalf("ParseConditional returned error: %v", err)
	}

	neqCond, ok := cond.IfBranch.Condition.(*ast.NotEqualsCondition)
	if !ok {
		t.Fatalf("expected NotEqualsCondition, got %T", cond.IfBranch.Condition)
	}

	// Check left side interpolation
	if len(neqCond.Left.Parts) != 1 {
		t.Fatalf("expected 1 part in left side, got %d", len(neqCond.Left.Parts))
	}
	interp, ok := neqCond.Left.Parts[0].(*ast.Interpolation)
	if !ok {
		t.Fatalf("expected Interpolation, got %T", neqCond.Left.Parts[0])
	}
	if interp.Name != "os" {
		t.Errorf("expected interpolation name 'os', got %q", interp.Name)
	}

	// Check right side
	if len(neqCond.Right.Parts) != 1 {
		t.Fatalf("expected 1 part in right side, got %d", len(neqCond.Right.Parts))
	}
	lit, ok := neqCond.Right.Parts[0].(*ast.LiteralValue)
	if !ok {
		t.Fatalf("expected LiteralValue, got %T", neqCond.Right.Parts[0])
	}
	if lit.Text != "windows" {
		t.Errorf("expected literal 'windows', got %q", lit.Text)
	}
}

// TestParseConditionalWithElif tests parsing if/elif/end block.
func TestParseConditionalWithElif(t *testing.T) {
	input := `if {os} == linux
cc = gcc
elif {os} == darwin
cc = clang
end
`
	l := lexer.New("test", input)
	p := New(l)

	cond, err := p.ParseConditional()
	if err != nil {
		t.Fatalf("ParseConditional returned error: %v", err)
	}

	// Check if branch
	if len(cond.IfBranch.Body) != 1 {
		t.Fatalf("expected 1 statement in if body, got %d", len(cond.IfBranch.Body))
	}

	// Check elif branch
	if len(cond.ElifBranches) != 1 {
		t.Fatalf("expected 1 elif branch, got %d", len(cond.ElifBranches))
	}
	elifCond, ok := cond.ElifBranches[0].Condition.(*ast.EqualsCondition)
	if !ok {
		t.Fatalf("expected EqualsCondition in elif, got %T", cond.ElifBranches[0].Condition)
	}

	// Check elif condition left side
	if len(elifCond.Left.Parts) != 1 {
		t.Fatalf("expected 1 part in elif left side, got %d", len(elifCond.Left.Parts))
	}
	interp, ok := elifCond.Left.Parts[0].(*ast.Interpolation)
	if !ok {
		t.Fatalf("expected Interpolation, got %T", elifCond.Left.Parts[0])
	}
	if interp.Name != "os" {
		t.Errorf("expected interpolation name 'os', got %q", interp.Name)
	}

	// Check elif condition right side
	if len(elifCond.Right.Parts) != 1 {
		t.Fatalf("expected 1 part in elif right side, got %d", len(elifCond.Right.Parts))
	}
	lit, ok := elifCond.Right.Parts[0].(*ast.LiteralValue)
	if !ok {
		t.Fatalf("expected LiteralValue, got %T", elifCond.Right.Parts[0])
	}
	if lit.Text != "darwin" {
		t.Errorf("expected literal 'darwin', got %q", lit.Text)
	}

	// Check elif body
	if len(cond.ElifBranches[0].Body) != 1 {
		t.Fatalf("expected 1 statement in elif body, got %d", len(cond.ElifBranches[0].Body))
	}

	// No else
	if cond.ElseBody != nil {
		t.Errorf("expected nil ElseBody, got %v", cond.ElseBody)
	}
}

// TestParseConditionalWithElse tests parsing if/else/end block.
func TestParseConditionalWithElse(t *testing.T) {
	input := `if {os} == linux
cc = gcc
else
cc = clang
end
`
	l := lexer.New("test", input)
	p := New(l)

	cond, err := p.ParseConditional()
	if err != nil {
		t.Fatalf("ParseConditional returned error: %v", err)
	}

	// Check if branch
	if len(cond.IfBranch.Body) != 1 {
		t.Fatalf("expected 1 statement in if body, got %d", len(cond.IfBranch.Body))
	}

	// No elif
	if len(cond.ElifBranches) != 0 {
		t.Errorf("expected 0 elif branches, got %d", len(cond.ElifBranches))
	}

	// Check else body
	if cond.ElseBody == nil {
		t.Fatal("expected ElseBody, got nil")
	}
	if len(cond.ElseBody) != 1 {
		t.Fatalf("expected 1 statement in else body, got %d", len(cond.ElseBody))
	}
}

// TestParseConditionalWithElifAndElse tests parsing if/elif/else/end block.
func TestParseConditionalWithElifAndElse(t *testing.T) {
	input := `if {os} == linux
cc = gcc
elif {os} == darwin
cc = clang
else
cc = cc
end
`
	l := lexer.New("test", input)
	p := New(l)

	cond, err := p.ParseConditional()
	if err != nil {
		t.Fatalf("ParseConditional returned error: %v", err)
	}

	// Check all branches present
	if len(cond.IfBranch.Body) != 1 {
		t.Fatalf("expected 1 statement in if body, got %d", len(cond.IfBranch.Body))
	}
	if len(cond.ElifBranches) != 1 {
		t.Fatalf("expected 1 elif branch, got %d", len(cond.ElifBranches))
	}
	if len(cond.ElifBranches[0].Body) != 1 {
		t.Fatalf("expected 1 statement in elif body, got %d", len(cond.ElifBranches[0].Body))
	}
	if cond.ElseBody == nil {
		t.Fatal("expected ElseBody, got nil")
	}
	if len(cond.ElseBody) != 1 {
		t.Fatalf("expected 1 statement in else body, got %d", len(cond.ElseBody))
	}
}

// TestParseConditionalMultipleElif tests parsing if with multiple elif branches.
func TestParseConditionalMultipleElif(t *testing.T) {
	input := `if {os} == linux
cc = gcc
elif {os} == darwin
cc = clang
elif {os} == windows
cc = cl
end
`
	l := lexer.New("test", input)
	p := New(l)

	cond, err := p.ParseConditional()
	if err != nil {
		t.Fatalf("ParseConditional returned error: %v", err)
	}

	if len(cond.ElifBranches) != 2 {
		t.Fatalf("expected 2 elif branches, got %d", len(cond.ElifBranches))
	}

	// Check second elif condition
	elifCond, ok := cond.ElifBranches[1].Condition.(*ast.EqualsCondition)
	if !ok {
		t.Fatalf("expected EqualsCondition in second elif, got %T", cond.ElifBranches[1].Condition)
	}
	if len(elifCond.Right.Parts) != 1 {
		t.Fatalf("expected 1 part in second elif right side, got %d", len(elifCond.Right.Parts))
	}
	lit, ok := elifCond.Right.Parts[0].(*ast.LiteralValue)
	if !ok {
		t.Fatalf("expected LiteralValue, got %T", elifCond.Right.Parts[0])
	}
	if lit.Text != "windows" {
		t.Errorf("expected literal 'windows', got %q", lit.Text)
	}
}

// TestParseIfdef tests parsing ifdef conditional.
func TestParseIfdef(t *testing.T) {
	input := `ifdef DEBUG
cflags = -g
end
`
	l := lexer.New("test", input)
	p := New(l)

	cond, err := p.ParseConditional()
	if err != nil {
		t.Fatalf("ParseConditional returned error: %v", err)
	}

	// Check condition is DefinedCondition
	defCond, ok := cond.IfBranch.Condition.(*ast.DefinedCondition)
	if !ok {
		t.Fatalf("expected DefinedCondition, got %T", cond.IfBranch.Condition)
	}
	if defCond.Name != "DEBUG" {
		t.Errorf("expected name 'DEBUG', got %q", defCond.Name)
	}

	// Check body
	if len(cond.IfBranch.Body) != 1 {
		t.Fatalf("expected 1 statement in body, got %d", len(cond.IfBranch.Body))
	}
}

// TestParseIfndef tests parsing ifndef conditional.
func TestParseIfndef(t *testing.T) {
	input := `ifndef CC
cc = gcc
end
`
	l := lexer.New("test", input)
	p := New(l)

	cond, err := p.ParseConditional()
	if err != nil {
		t.Fatalf("ParseConditional returned error: %v", err)
	}

	// Check condition is NotDefinedCondition
	notDefCond, ok := cond.IfBranch.Condition.(*ast.NotDefinedCondition)
	if !ok {
		t.Fatalf("expected NotDefinedCondition, got %T", cond.IfBranch.Condition)
	}
	if notDefCond.Name != "CC" {
		t.Errorf("expected name 'CC', got %q", notDefCond.Name)
	}
}

// TestParseConditionalEmptyBody tests parsing conditional with empty body.
func TestParseConditionalEmptyBody(t *testing.T) {
	input := `if {os} == linux
end
`
	l := lexer.New("test", input)
	p := New(l)

	cond, err := p.ParseConditional()
	if err != nil {
		t.Fatalf("ParseConditional returned error: %v", err)
	}

	if len(cond.IfBranch.Body) != 0 {
		t.Errorf("expected 0 statements in body, got %d", len(cond.IfBranch.Body))
	}
}

// TestParseConditionalMultipleStatements tests parsing conditional with multiple statements in body.
func TestParseConditionalMultipleStatements(t *testing.T) {
	input := `if {os} == linux
cc = gcc
cflags = -Wall
ldflags = -lpthread
end
`
	l := lexer.New("test", input)
	p := New(l)

	cond, err := p.ParseConditional()
	if err != nil {
		t.Fatalf("ParseConditional returned error: %v", err)
	}

	if len(cond.IfBranch.Body) != 3 {
		t.Fatalf("expected 3 statements in body, got %d", len(cond.IfBranch.Body))
	}

	// Verify each statement is a variable
	for i, stmt := range cond.IfBranch.Body {
		if _, ok := stmt.(*ast.Variable); !ok {
			t.Errorf("statement %d: expected Variable, got %T", i, stmt)
		}
	}
}

// TestParseConditionalMissingEnd tests error when end is missing.
func TestParseConditionalMissingEnd(t *testing.T) {
	input := `if {os} == linux
cc = gcc
`
	l := lexer.New("test", input)
	p := New(l)

	_, err := p.ParseConditional()
	if err == nil {
		t.Fatal("expected error for missing 'end', got nil")
	}
}

// TestParseConditionalMissingCondition tests error when condition is missing.
func TestParseConditionalMissingCondition(t *testing.T) {
	input := `if
cc = gcc
end
`
	l := lexer.New("test", input)
	p := New(l)

	_, err := p.ParseConditional()
	if err == nil {
		t.Fatal("expected error for missing condition, got nil")
	}
}

// TestParseConditionalTargetInsideHasHint verifies the "targets cannot be
// defined inside conditionals" error carries the hint pointing users at the
// two supported workarounds (C4).
func TestParseConditionalTargetInsideHasHint(t *testing.T) {
	input := `if {os} == darwin
    @test:
        echo "Running on macOS"
end
`
	l := lexer.New("test", input)
	p := New(l)

	_, err := p.ParseConditional()
	if err == nil {
		t.Fatal("expected error for target inside conditional, got nil")
	}
	if err.Message != "targets cannot be defined inside conditionals" {
		t.Errorf("Message = %q, want %q", err.Message, "targets cannot be defined inside conditionals")
	}
	wantHint := "move the condition into the recipe, or set a variable inside the conditional and use it in the dependency list or command"
	if err.Hint != wantHint {
		t.Errorf("Hint = %q, want %q", err.Hint, wantHint)
	}
}

// TestParseConditionalConditionWithRawModifier tests condition with :raw modifier.
func TestParseConditionalConditionWithRawModifier(t *testing.T) {
	input := `if {flags:raw} == -O2
opt = yes
end
`
	l := lexer.New("test", input)
	p := New(l)

	cond, err := p.ParseConditional()
	if err != nil {
		t.Fatalf("ParseConditional returned error: %v", err)
	}

	eqCond, ok := cond.IfBranch.Condition.(*ast.EqualsCondition)
	if !ok {
		t.Fatalf("expected EqualsCondition, got %T", cond.IfBranch.Condition)
	}

	if len(eqCond.Left.Parts) != 1 {
		t.Fatalf("expected 1 part in left side, got %d", len(eqCond.Left.Parts))
	}
	interp, ok := eqCond.Left.Parts[0].(*ast.Interpolation)
	if !ok {
		t.Fatalf("expected Interpolation, got %T", eqCond.Left.Parts[0])
	}
	if interp.Name != "flags" {
		t.Errorf("expected interpolation name 'flags', got %q", interp.Name)
	}
	if !interp.Raw {
		t.Error("expected Raw to be true")
	}
}

// TestParseConditionalSourceLocation tests that source location is captured correctly.
func TestParseConditionalSourceLocation(t *testing.T) {
	input := `if {os} == linux
cc = gcc
end
`
	l := lexer.New("test.need", input)
	p := New(l)

	cond, err := p.ParseConditional()
	if err != nil {
		t.Fatalf("ParseConditional returned error: %v", err)
	}

	if cond.Location.File != "test.need" {
		t.Errorf("expected file 'test.need', got %q", cond.Location.File)
	}
	if cond.Location.Line != 1 {
		t.Errorf("expected line 1, got %d", cond.Location.Line)
	}
	if cond.Location.Column != 1 {
		t.Errorf("expected column 1, got %d", cond.Location.Column)
	}
}

// TestIsConditionalLine tests detection of conditional start.
func TestIsConditionalLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"if keyword", "if {os} == linux\n", true},
		{"ifdef keyword", "ifdef DEBUG\n", true},
		{"ifndef keyword", "ifndef CC\n", true},
		{"variable line", "cc = gcc\n", false},
		{"target line", "app: main.o\n", false},
		{"directive line", ".shell: bash\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test", tt.input)
			p := New(l)

			got := p.IsConditionalLine()
			if got != tt.expected {
				t.Errorf("IsConditionalLine() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
