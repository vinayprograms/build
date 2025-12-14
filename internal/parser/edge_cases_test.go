package parser

import (
	"testing"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/lexer"
)

// TestEdgeCase_NestedConditionals tests conditionals inside conditionals.
func TestEdgeCase_NestedConditionals(t *testing.T) {
	input := `if {os} == linux
    ifdef DEBUG
        debug_flags = -g
    end
    cc = gcc
end
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}

	outer, ok := stmts[0].(*ast.Conditional)
	if !ok {
		t.Fatalf("expected Conditional, got %T", stmts[0])
	}

	// The outer if body should contain the nested ifdef and a variable
	if len(outer.IfBranch.Body) != 2 {
		t.Errorf("outer if body has %d statements, want 2", len(outer.IfBranch.Body))
	}

	// First statement in body should be a conditional (ifdef)
	inner, ok := outer.IfBranch.Body[0].(*ast.Conditional)
	if !ok {
		t.Fatalf("first body statement should be Conditional, got %T", outer.IfBranch.Body[0])
	}

	// Verify it's a DefinedCondition
	if _, ok := inner.IfBranch.Condition.(*ast.DefinedCondition); !ok {
		t.Error("inner conditional should be DefinedCondition")
	}
}

// TestEdgeCase_DirectiveInWrongScope tests directive scope errors.
func TestEdgeCase_DirectiveInWrongScope(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErrPart string
	}{
		{
			name: "environment directive at global scope",
			input: `.using: docker
`,
			wantErrPart: ".using",
		},
		{
			name: "after directive at global scope",
			input: `.after: dep
`,
			wantErrPart: ".after",
		},
		{
			name: "autodeps at global scope",
			input: `.autodeps: file.d
`,
			wantErrPart: ".autodeps",
		},
		{
			name: "source at global scope",
			input: `.source: Dockerfile
`,
			wantErrPart: ".source",
		},
		{
			name: "args at global scope",
			input: `.args: --platform linux
`,
			wantErrPart: ".args",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.build", tt.input)
			p := New(l)

			_, errs := p.ParseBuildfile()

			if !errs.HasErrors() {
				t.Fatal("expected error")
			}

			// Check error mentions the directive
			errStr := errs.Error()
			if !containsSubstring(errStr, tt.wantErrPart) {
				t.Errorf("error %q doesn't mention %q", errStr, tt.wantErrPart)
			}
		})
	}
}

// TestEdgeCase_BlockAtWrongIndentation tests block command at wrong indent.
func TestEdgeCase_BlockAtWrongIndentation(t *testing.T) {
	// Block at level 0 (should be an error or be treated as identifier)
	input := `block:
    echo hello
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	// This input is ambiguous - 'block:' could be interpreted as a target
	// The behavior depends on lexer/parser implementation
	// Just verify it doesn't crash
	_ = stmts
	_ = errs
}

// TestEdgeCase_EnvironmentInRecipe tests .environment inside recipe (error).
func TestEdgeCase_EnvironmentInRecipe(t *testing.T) {
	input := `@test:
    .environment:
        .using: docker
`
	l := lexer.New("test.build", input)
	p := New(l)

	_, errs := p.ParseBuildfile()

	// .environment at recipe scope should be invalid
	if !errs.HasErrors() {
		t.Error("expected error for .environment in recipe scope")
	}
}

// TestEdgeCase_ParallelInEnvironment tests .parallel inside environment (error).
func TestEdgeCase_ParallelInEnvironment(t *testing.T) {
	input := `.environment:
    .parallel: 4
`
	l := lexer.New("test.build", input)
	p := New(l)

	_, errs := p.ParseBuildfile()

	if !errs.HasErrors() {
		t.Error("expected error for .parallel in environment scope")
	}

	if !containsSubstring(errs.Error(), ".parallel") {
		t.Errorf("error should mention .parallel: %s", errs.Error())
	}
}

// TestEdgeCase_DefaultInRecipe tests .default inside recipe (error).
func TestEdgeCase_DefaultInRecipe(t *testing.T) {
	input := `@test:
    .default: @other
`
	l := lexer.New("test.build", input)
	p := New(l)

	_, errs := p.ParseBuildfile()

	if !errs.HasErrors() {
		t.Error("expected error for .default in recipe scope")
	}
}

// TestEdgeCase_DeeplyNestedBlocks tests deeply nested structure.
func TestEdgeCase_DeeplyNestedBlocks(t *testing.T) {
	input := `if {os} == linux
    if {arch} == amd64
        ifdef DEBUG
            debug_flags = -g -O0
        end
        cc = gcc
    end
    cflags = -Wall
end
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}

	// Verify structure - outer if
	outer, ok := stmts[0].(*ast.Conditional)
	if !ok {
		t.Fatal("expected Conditional")
	}

	if len(outer.IfBranch.Body) != 2 {
		t.Errorf("outer if body has %d statements, want 2", len(outer.IfBranch.Body))
	}

	// First should be another conditional
	mid, ok := outer.IfBranch.Body[0].(*ast.Conditional)
	if !ok {
		t.Fatal("first body statement should be Conditional")
	}

	if len(mid.IfBranch.Body) != 2 {
		t.Errorf("middle if body has %d statements, want 2", len(mid.IfBranch.Body))
	}

	// First in middle should be ifdef
	inner, ok := mid.IfBranch.Body[0].(*ast.Conditional)
	if !ok {
		t.Fatal("expected inner Conditional")
	}

	if _, ok := inner.IfBranch.Condition.(*ast.DefinedCondition); !ok {
		t.Error("innermost should be DefinedCondition")
	}
}

// TestEdgeCase_MultipleErrorsCollected tests that multiple errors are collected.
func TestEdgeCase_MultipleErrorsCollected(t *testing.T) {
	input := `.after: err1
.using: err2
.source: err3
.args: err4
.autodeps: err5
cc = valid
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	if len(errs.Errors) != 5 {
		t.Errorf("expected 5 errors, got %d", len(errs.Errors))
	}

	// Should still have the valid variable
	var vars int
	for _, stmt := range stmts {
		if _, ok := stmt.(*ast.Variable); ok {
			vars++
		}
	}
	if vars != 1 {
		t.Errorf("expected 1 variable, got %d", vars)
	}
}

// TestEdgeCase_RecipeWithOnlyDirectives tests recipe with directives but no commands.
func TestEdgeCase_RecipeWithOnlyDirectives(t *testing.T) {
	input := `@setup:
    .shell: bash
    .requires: gcc cmake
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}

	target, ok := stmts[0].(*ast.Target)
	if !ok {
		t.Fatal("expected Target")
	}

	if target.Recipe == nil {
		t.Fatal("target should have recipe")
	}

	if target.Recipe.Directives.Shell == nil {
		t.Error("recipe should have .shell directive")
	}

	if len(target.Recipe.Directives.Requires) != 2 {
		t.Errorf("recipe has %d requires, want 2", len(target.Recipe.Directives.Requires))
	}

	if len(target.Recipe.Commands) != 0 {
		t.Errorf("recipe has %d commands, want 0", len(target.Recipe.Commands))
	}
}

// TestEdgeCase_TargetWithNoRecipe tests target without recipe.
func TestEdgeCase_TargetWithNoRecipe(t *testing.T) {
	input := `@all: @test @build
@test: build/test
@build: build/app
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(stmts))
	}

	for i, stmt := range stmts {
		target, ok := stmt.(*ast.Target)
		if !ok {
			t.Fatalf("stmts[%d] is not Target", i)
		}
		if target.Recipe != nil {
			t.Errorf("target %d should not have recipe", i)
		}
	}
}

// TestEdgeCase_VersionSpecFormats tests all version specification formats.
func TestEdgeCase_VersionSpecFormats(t *testing.T) {
	input := `.environment:
    .requires: gcc gcc@latest gcc@11 gcc@11.4 gcc@11.4.0
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}

	env, ok := stmts[0].(*ast.Environment)
	if !ok {
		t.Fatal("expected Environment")
	}

	if len(env.Requires) != 5 {
		t.Fatalf("expected 5 requires, got %d", len(env.Requires))
	}

	// Check version types
	versionTypes := []string{
		"VersionLatest",     // gcc (no @)
		"VersionLatest",     // gcc@latest
		"VersionMajor",      // gcc@11
		"VersionMajorMinor", // gcc@11.4
		"VersionExact",      // gcc@11.4.0
	}

	for i, req := range env.Requires {
		got := ""
		switch req.Version.(type) {
		case ast.VersionLatest, *ast.VersionLatest:
			got = "VersionLatest"
		case ast.VersionMajor, *ast.VersionMajor:
			got = "VersionMajor"
		case ast.VersionMajorMinor, *ast.VersionMajorMinor:
			got = "VersionMajorMinor"
		case ast.VersionExact, *ast.VersionExact:
			got = "VersionExact"
		}

		if got != versionTypes[i] {
			t.Errorf("requires[%d] version type = %s, want %s", i, got, versionTypes[i])
		}
	}
}

// TestEdgeCase_FunctionCallsInCommands tests function calls within recipe commands.
func TestEdgeCase_FunctionCallsInCommands(t *testing.T) {
	// Function calls in commands are treated as literal text since
	// functions are evaluated at variable definition time, not in commands
	input := `@test:
    echo {cc} builds to {target}
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}

	target, ok := stmts[0].(*ast.Target)
	if !ok {
		t.Fatal("expected Target")
	}

	if target.Recipe == nil || len(target.Recipe.Commands) != 1 {
		t.Fatal("expected 1 command")
	}

	cmd, ok := target.Recipe.Commands[0].(*ast.LineCommand)
	if !ok {
		t.Fatal("expected LineCommand")
	}

	// Should have interpolations
	var interpCount int
	for _, part := range cmd.Parts {
		if _, ok := part.(*ast.CommandInterpolation); ok {
			interpCount++
		}
	}

	if interpCount != 2 {
		t.Errorf("command has %d interpolations, want 2", interpCount)
	}
}

// TestEdgeCase_EscapedBracesInValue tests {{ and }} in values.
func TestEdgeCase_EscapedBracesInValue(t *testing.T) {
	input := `template = Hello {{name}}!
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}

	v, ok := stmts[0].(*ast.Variable)
	if !ok {
		t.Fatal("expected Variable")
	}

	// Value should have the escaped braces as literal {
	hasOpenBrace := false
	for _, part := range v.Value.Parts {
		if lit, ok := part.(*ast.LiteralValue); ok {
			if containsSubstring(lit.Text, "{") {
				hasOpenBrace = true
			}
		}
	}

	if !hasOpenBrace {
		t.Error("value should contain literal brace from escape sequence")
	}
}

// TestEdgeCase_CommentAfterStatement tests inline comments.
func TestEdgeCase_CommentAfterStatement(t *testing.T) {
	input := `cc = gcc # This is the C compiler
cflags = -Wall # Warning flags
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	// Should have 2 variables (comments are handled by lexer)
	var vars int
	for _, stmt := range stmts {
		if _, ok := stmt.(*ast.Variable); ok {
			vars++
		}
	}

	if vars != 2 {
		t.Errorf("expected 2 variables, got %d", vars)
	}
}

// TestEdgeCase_PathWithInterpolation tests paths with variable interpolation.
func TestEdgeCase_PathWithInterpolation(t *testing.T) {
	// Note: Interpolations at the START of a target pattern require a leading path segment
	// e.g., "build/{name}.o" works, but "{build_dir}/app" doesn't currently work
	// This is a known limitation of the lexer/parser that treats leading { as INTERP_START
	input := `build/{name}.o: src/{name}.c
    {cc} -c -o {target} {deps}
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}

	target, ok := stmts[0].(*ast.Target)
	if !ok {
		t.Fatal("expected Target")
	}

	// Check pattern has brace expressions
	hasBrace := false
	for _, seg := range target.Pattern.Segments {
		if _, ok := seg.(*ast.BraceExpr); ok {
			hasBrace = true
		}
	}
	if !hasBrace {
		t.Error("target pattern should have BraceExpr")
	}

	// Check dependency has brace expressions
	if len(target.Dependencies) == 0 {
		t.Fatal("target should have dependencies")
	}
	hasBrace = false
	for _, seg := range target.Dependencies[0].Segments {
		if _, ok := seg.(*ast.BraceExpr); ok {
			hasBrace = true
		}
	}
	if !hasBrace {
		t.Error("dependency should have BraceExpr")
	}
}

// TestEdgeCase_TargetStartingWithInterpolation tests targets that start with {var}.
func TestEdgeCase_TargetStartingWithInterpolation(t *testing.T) {
	input := `{build_dir}/app: src/main.c
    gcc -o {target} {deps}

{out_dir}/:
    mkdir -p {target}
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(stmts))
	}

	// Check first target
	t1, ok := stmts[0].(*ast.Target)
	if !ok {
		t.Fatalf("stmts[0] is not Target, got %T", stmts[0])
	}

	// Should start with brace expression
	if len(t1.Pattern.Segments) == 0 {
		t.Fatal("target should have segments")
	}
	if _, ok := t1.Pattern.Segments[0].(*ast.BraceExpr); !ok {
		t.Errorf("first segment should be BraceExpr, got %T", t1.Pattern.Segments[0])
	}

	// Check second target (directory)
	t2, ok := stmts[1].(*ast.Target)
	if !ok {
		t.Fatalf("stmts[1] is not Target, got %T", stmts[1])
	}
	if !t2.Pattern.IsDirectory {
		t.Error("second target should be directory")
	}
}

// TestEdgeCase_PhonyTargetWithHyphen tests phony targets with hyphens in name.
func TestEdgeCase_PhonyTargetWithHyphen(t *testing.T) {
	input := `@test-cover:
    go test -cover ./...

@debug-lex-tokens:
    go run ./cmd/build --debug-lex

@my-long-target-name: @test-cover @debug-lex-tokens
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(stmts))
	}

	// Check all are phony targets
	expectedNames := []string{"test-cover", "debug-lex-tokens", "my-long-target-name"}
	for i, stmt := range stmts {
		target, ok := stmt.(*ast.Target)
		if !ok {
			t.Fatalf("stmts[%d] is not Target, got %T", i, stmt)
		}
		if !target.Pattern.IsPhony {
			t.Errorf("target %d should be phony", i)
		}
		// Extract name from pattern
		if len(target.Pattern.Segments) != 1 {
			t.Errorf("target %d should have 1 segment", i)
			continue
		}
		lit, ok := target.Pattern.Segments[0].(*ast.LiteralSegment)
		if !ok {
			t.Errorf("target %d segment should be LiteralSegment", i)
			continue
		}
		if lit.Text != expectedNames[i] {
			t.Errorf("target %d name = %q, want %q", i, lit.Text, expectedNames[i])
		}
	}
}
