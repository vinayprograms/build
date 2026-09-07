package parser

import (
	"testing"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/lexer"
)

// TestParseNeedfile_AllStatementTypes tests that ParseNeedfile correctly
// parses all statement types in a complete needfile.
func TestParseNeedfile_AllStatementTypes(t *testing.T) {
	input := `# Comment at top
.shell: bash
.parallel: 4
.default: @all

cc = gcc
lazy cflags = -Wall -O2

if {os} == linux
cc = gcc
elif {os} == darwin
cc = clang
else
cc = cc
end

ifdef DEBUG
debug_flags = -g
end

.environment: dev
    .using: docker
    .source: Dockerfile
    .requires: gcc@11

.environment:
    .using: bare
    .requires: make cmake

src/main.o: src/main.c
    {cc} -c -o {target} {deps}

build/{name}.o: src/{name}.c
    {cc} {cflags:raw} -c -o {target} {deps}

@all: src/main.o
    echo done

@clean:
    rm -rf build/

build/:
    mkdir -p {target}
`
	l := lexer.New("test.need", input)
	p := New(l)

	stmts, errs := p.ParseNeedfile()

	if errs.HasErrors() {
		t.Errorf("unexpected errors: %s", errs.Error())
	}

	// Count statement types
	counts := make(map[string]int)
	for _, stmt := range stmts {
		switch stmt.(type) {
		case *ast.Comment:
			counts["comment"]++
		case *ast.Directive:
			counts["directive"]++
		case *ast.Variable:
			counts["variable"]++
		case *ast.Conditional:
			counts["conditional"]++
		case *ast.Environment:
			counts["environment"]++
		case *ast.Target:
			counts["target"]++
		case *ast.Blank:
			counts["blank"]++
		}
	}

	// Expected counts
	expected := map[string]int{
		"comment":     1, // # Comment at top
		"directive":   3, // .shell, .parallel, .default
		"variable":    2, // cc, lazy cflags
		"conditional": 2, // if/elif/else/end, ifdef/end
		"environment": 2, // dev, default
		"target":      5, // main.o, {name}.o, @all, @clean, build/
	}

	for stmtType, want := range expected {
		if got := counts[stmtType]; got != want {
			t.Errorf("%s count = %d, want %d", stmtType, got, want)
		}
	}
}

// TestParseNeedfile_DirectiveDetails tests that directives are parsed correctly.
func TestParseNeedfile_DirectiveDetails(t *testing.T) {
	input := `.shell: bash
.parallel: 4
.default: @all
`
	l := lexer.New("test.need", input)
	p := New(l)

	stmts, errs := p.ParseNeedfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(stmts))
	}

	// Check .shell directive
	shell, ok := stmts[0].(*ast.Directive)
	if !ok {
		t.Fatalf("stmts[0] is not Directive, got %T", stmts[0])
	}
	if shell.Kind != ast.DirectiveShell {
		t.Errorf("shell.Kind = %v, want DirectiveShell", shell.Kind)
	}

	// Check .parallel directive
	parallel, ok := stmts[1].(*ast.Directive)
	if !ok {
		t.Fatalf("stmts[1] is not Directive, got %T", stmts[1])
	}
	if parallel.Kind != ast.DirectiveParallel {
		t.Errorf("parallel.Kind = %v, want DirectiveParallel", parallel.Kind)
	}

	// Check .default directive
	def, ok := stmts[2].(*ast.Directive)
	if !ok {
		t.Fatalf("stmts[2] is not Directive, got %T", stmts[2])
	}
	if def.Kind != ast.DirectiveDefault {
		t.Errorf("def.Kind = %v, want DirectiveDefault", def.Kind)
	}
}

// TestParseNeedfile_VariableDetails tests variable parsing in needfile context.
func TestParseNeedfile_VariableDetails(t *testing.T) {
	input := `cc = gcc
lazy cflags = -Wall {optimization}
path = /usr/bin:{extra_path}
`
	l := lexer.New("test.need", input)
	p := New(l)

	stmts, errs := p.ParseNeedfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(stmts))
	}

	// Check first variable
	v1, ok := stmts[0].(*ast.Variable)
	if !ok {
		t.Fatalf("stmts[0] is not Variable, got %T", stmts[0])
	}
	if v1.Name != "cc" || v1.Lazy {
		t.Errorf("v1: name=%q lazy=%v, want name=cc lazy=false", v1.Name, v1.Lazy)
	}

	// Check lazy variable
	v2, ok := stmts[1].(*ast.Variable)
	if !ok {
		t.Fatalf("stmts[1] is not Variable, got %T", stmts[1])
	}
	if v2.Name != "cflags" || !v2.Lazy {
		t.Errorf("v2: name=%q lazy=%v, want name=cflags lazy=true", v2.Name, v2.Lazy)
	}
	// Check it has interpolation
	hasInterp := false
	for _, part := range v2.Value.Parts {
		if _, ok := part.(*ast.Interpolation); ok {
			hasInterp = true
		}
	}
	if !hasInterp {
		t.Error("v2 should have interpolation")
	}
}

// TestParseNeedfile_TargetDetails tests target parsing with recipes.
func TestParseNeedfile_TargetDetails(t *testing.T) {
	input := `build/app: src/main.o src/utils.o
    .shell: bash
    .requires: gcc@11
    gcc -o {target} {deps}

@test: build/app
    ./build/app --test

build/{name}.o: src/{name}.c
    gcc -c -o {target} {deps}
`
	l := lexer.New("test.need", input)
	p := New(l)

	stmts, errs := p.ParseNeedfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(stmts))
	}

	// Check file target
	t1, ok := stmts[0].(*ast.Target)
	if !ok {
		t.Fatalf("stmts[0] is not Target, got %T", stmts[0])
	}
	if t1.Pattern.IsPhony || t1.Pattern.IsDirectory {
		t.Error("first target should not be phony or directory")
	}
	if len(t1.Dependencies) != 2 {
		t.Errorf("first target has %d deps, want 2", len(t1.Dependencies))
	}
	if t1.Recipe == nil {
		t.Fatal("first target should have recipe")
	}
	if t1.Recipe.Directives.Shell == nil {
		t.Error("first target recipe should have .shell directive")
	}
	if len(t1.Recipe.Directives.Requires) != 1 {
		t.Errorf("first target has %d requires, want 1", len(t1.Recipe.Directives.Requires))
	}

	// Check phony target
	t2, ok := stmts[1].(*ast.Target)
	if !ok {
		t.Fatalf("stmts[1] is not Target, got %T", stmts[1])
	}
	if !t2.Pattern.IsPhony {
		t.Error("second target should be phony")
	}

	// Check pattern target with captures
	t3, ok := stmts[2].(*ast.Target)
	if !ok {
		t.Fatalf("stmts[2] is not Target, got %T", stmts[2])
	}
	hasCapture := false
	for _, seg := range t3.Pattern.Segments {
		if _, ok := seg.(*ast.BraceExpr); ok {
			hasCapture = true
		}
	}
	if !hasCapture {
		t.Error("third target should have capture")
	}
}

// TestParseNeedfile_EnvironmentDetails tests environment block parsing.
func TestParseNeedfile_EnvironmentDetails(t *testing.T) {
	input := `.environment: ci
    .using: docker
    .source: Dockerfile.ci
    .args: --platform linux/amd64
    .requires: gcc@11 python3@3.10

.environment:
    .using: bare
    .requires: gcc make
`
	l := lexer.New("test.need", input)
	p := New(l)

	stmts, errs := p.ParseNeedfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(stmts))
	}

	// Check named environment
	e1, ok := stmts[0].(*ast.Environment)
	if !ok {
		t.Fatalf("stmts[0] is not Environment, got %T", stmts[0])
	}
	if e1.Name == nil || *e1.Name != "ci" {
		t.Error("first env should be named 'ci'")
	}
	if e1.Runtime == nil || *e1.Runtime != ast.RuntimeDocker {
		t.Error("first env should use docker")
	}
	if e1.Source == nil {
		t.Error("first env should have source")
	}
	if e1.Args == nil {
		t.Error("first env should have args")
	}
	if len(e1.Requires) != 2 {
		t.Errorf("first env has %d requires, want 2", len(e1.Requires))
	}

	// Check default environment
	e2, ok := stmts[1].(*ast.Environment)
	if !ok {
		t.Fatalf("stmts[1] is not Environment, got %T", stmts[1])
	}
	if e2.Name != nil {
		t.Error("second env should be default (unnamed)")
	}
	if e2.Runtime == nil || *e2.Runtime != ast.RuntimeBare {
		t.Error("second env should use bare")
	}
}

// TestParseNeedfile_ConditionalDetails tests conditional parsing.
func TestParseNeedfile_ConditionalDetails(t *testing.T) {
	input := `if {os} == linux
cc = gcc
cflags = -Wall
elif {os} == darwin
cc = clang
else
cc = cc
end

ifdef DEBUG
debug = true
end

ifndef RELEASE
optimize = false
end
`
	l := lexer.New("test.need", input)
	p := New(l)

	stmts, errs := p.ParseNeedfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(stmts))
	}

	// Check if/elif/else conditional
	c1, ok := stmts[0].(*ast.Conditional)
	if !ok {
		t.Fatalf("stmts[0] is not Conditional, got %T", stmts[0])
	}
	if len(c1.IfBranch.Body) != 2 {
		t.Errorf("if body has %d stmts, want 2", len(c1.IfBranch.Body))
	}
	if len(c1.ElifBranches) != 1 {
		t.Errorf("has %d elif branches, want 1", len(c1.ElifBranches))
	}
	if c1.ElseBody == nil || len(c1.ElseBody) != 1 {
		t.Error("should have else with 1 statement")
	}

	// Check ifdef
	c2, ok := stmts[1].(*ast.Conditional)
	if !ok {
		t.Fatalf("stmts[1] is not Conditional, got %T", stmts[1])
	}
	if _, ok := c2.IfBranch.Condition.(*ast.DefinedCondition); !ok {
		t.Error("should be DefinedCondition")
	}

	// Check ifndef
	c3, ok := stmts[2].(*ast.Conditional)
	if !ok {
		t.Fatalf("stmts[2] is not Conditional, got %T", stmts[2])
	}
	if _, ok := c3.IfBranch.Condition.(*ast.NotDefinedCondition); !ok {
		t.Error("should be NotDefinedCondition")
	}
}

// TestParseNeedfile_NestedBlocks tests recipe with block command.
func TestParseNeedfile_NestedBlocks(t *testing.T) {
	input := `@setup:
    block:
        if [ ! -d build ]; then
            mkdir -p build
        fi
        echo "Setup complete"
`
	l := lexer.New("test.need", input)
	p := New(l)

	stmts, errs := p.ParseNeedfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}

	target, ok := stmts[0].(*ast.Target)
	if !ok {
		t.Fatalf("stmts[0] is not Target, got %T", stmts[0])
	}

	if target.Recipe == nil {
		t.Fatal("target should have recipe")
	}

	if len(target.Recipe.Commands) != 1 {
		t.Fatalf("recipe has %d commands, want 1", len(target.Recipe.Commands))
	}

	block, ok := target.Recipe.Commands[0].(*ast.BlockCommand)
	if !ok {
		t.Fatalf("command is not BlockCommand, got %T", target.Recipe.Commands[0])
	}

	// Block should have multiple lines
	if len(block.Lines) < 3 {
		t.Errorf("block has %d lines, want at least 3", len(block.Lines))
	}
}

// TestParseNeedfile_SourceLocations tests that source locations are correct.
func TestParseNeedfile_SourceLocations(t *testing.T) {
	input := `cc = gcc
@test:
    echo hello
`
	l := lexer.New("test.need", input)
	p := New(l)

	stmts, errs := p.ParseNeedfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(stmts))
	}

	// Check variable location
	v, ok := stmts[0].(*ast.Variable)
	if !ok {
		t.Fatalf("stmts[0] is not Variable, got %T", stmts[0])
	}
	if v.Location.Line != 1 {
		t.Errorf("variable at line %d, want 1", v.Location.Line)
	}

	// Check target location
	target, ok := stmts[1].(*ast.Target)
	if !ok {
		t.Fatalf("stmts[1] is not Target, got %T", stmts[1])
	}
	if target.Location.Line != 2 {
		t.Errorf("target at line %d, want 2", target.Location.Line)
	}

	// Check recipe location
	if target.Recipe.Location.Line != 3 {
		t.Errorf("recipe at line %d, want 3", target.Recipe.Location.Line)
	}
}

// TestParseNeedfile_EmptyFile tests parsing an empty file.
func TestParseNeedfile_EmptyFile(t *testing.T) {
	input := ``
	l := lexer.New("test.need", input)
	p := New(l)

	stmts, errs := p.ParseNeedfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 0 {
		t.Errorf("expected 0 statements, got %d", len(stmts))
	}
}

// TestParseNeedfile_OnlyComments tests parsing file with only comments.
func TestParseNeedfile_OnlyComments(t *testing.T) {
	input := `# This is a comment
# Another comment
# Yet another
`
	l := lexer.New("test.need", input)
	p := New(l)

	stmts, errs := p.ParseNeedfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	if len(stmts) != 3 {
		t.Errorf("expected 3 statements (comments), got %d", len(stmts))
	}

	for i, stmt := range stmts {
		if _, ok := stmt.(*ast.Comment); !ok {
			t.Errorf("stmts[%d] is not Comment, got %T", i, stmt)
		}
	}
}

// TestParseNeedfile_MixedWithBlankLines tests that blank lines don't cause issues.
func TestParseNeedfile_MixedWithBlankLines(t *testing.T) {
	input := `cc = gcc

cflags = -Wall

@test:
    echo hello

`
	l := lexer.New("test.need", input)
	p := New(l)

	stmts, errs := p.ParseNeedfile()

	if errs.HasErrors() {
		t.Fatalf("unexpected errors: %s", errs.Error())
	}

	// Should have 2 variables and 1 target (blanks are not significant statements)
	var vars, targets int
	for _, stmt := range stmts {
		switch stmt.(type) {
		case *ast.Variable:
			vars++
		case *ast.Target:
			targets++
		}
	}

	if vars != 2 {
		t.Errorf("expected 2 variables, got %d", vars)
	}
	if targets != 1 {
		t.Errorf("expected 1 target, got %d", targets)
	}
}

// TestParseNeedfile_ErrorRecoveryIntegration tests error recovery in full needfile.
func TestParseNeedfile_ErrorRecoveryIntegration(t *testing.T) {
	input := `cc = gcc
.after: invalid
cflags = -Wall
.using: invalid
@test:
    echo hello
`
	l := lexer.New("test.need", input)
	p := New(l)

	stmts, errs := p.ParseNeedfile()

	// Should have errors
	if !errs.HasErrors() {
		t.Fatal("expected errors")
	}
	if len(errs.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errs.Errors))
	}

	// Should still have parsed valid statements
	var vars, targets int
	for _, stmt := range stmts {
		switch stmt.(type) {
		case *ast.Variable:
			vars++
		case *ast.Target:
			targets++
		}
	}

	if vars != 2 {
		t.Errorf("expected 2 variables, got %d", vars)
	}
	if targets != 1 {
		t.Errorf("expected 1 target, got %d", targets)
	}
}
