package parser

import (
	"strings"
	"testing"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/lexer"
)

// TestParser_ParseRecipe_SimpleCommand tests parsing a single command line.
func TestParser_ParseRecipe_SimpleCommand(t *testing.T) {
	// Input: target with a single command
	input := `build/app: src/main.c
    gcc -o build/app src/main.c
`
	l := lexer.New("test.need", input)
	p := New(l)

	// Parse the target (which should include the recipe)
	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed")
	}

	if len(target.Recipe.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(target.Recipe.Commands))
	}

	// Check it's a line command
	lineCmd, ok := target.Recipe.Commands[0].(*ast.LineCommand)
	if !ok {
		t.Fatalf("expected LineCommand, got %T", target.Recipe.Commands[0])
	}

	// Should have at least one part (the command text)
	if len(lineCmd.Parts) == 0 {
		t.Error("expected command to have parts")
	}
}

// TestParser_ParseRecipe_MultipleCommands tests parsing multiple command lines.
func TestParser_ParseRecipe_MultipleCommands(t *testing.T) {
	input := `build/app: src/main.c
    echo "Building..."
    gcc -c src/main.c -o build/main.o
    gcc -o build/app build/main.o
`
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed")
	}

	if len(target.Recipe.Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(target.Recipe.Commands))
	}
}

// TestParser_ParseRecipe_WithInterpolations tests command lines with {var} interpolations.
func TestParser_ParseRecipe_WithInterpolations(t *testing.T) {
	input := `build/app: src/main.c
    gcc -o {target} {deps}
`
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed")
	}

	if len(target.Recipe.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(target.Recipe.Commands))
	}

	lineCmd, ok := target.Recipe.Commands[0].(*ast.LineCommand)
	if !ok {
		t.Fatalf("expected LineCommand, got %T", target.Recipe.Commands[0])
	}

	// Count interpolations
	interpCount := 0
	for _, part := range lineCmd.Parts {
		if _, ok := part.(*ast.CommandInterpolation); ok {
			interpCount++
		}
	}

	if interpCount != 2 {
		t.Errorf("expected 2 interpolations, got %d", interpCount)
	}
}

// TestParser_ParseRecipe_RawModifier tests {var:raw} modifier in commands.
func TestParser_ParseRecipe_RawModifier(t *testing.T) {
	input := `build/app: src/main.c
    gcc {cflags:raw} -o {target} {deps}
`
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed")
	}

	lineCmd, ok := target.Recipe.Commands[0].(*ast.LineCommand)
	if !ok {
		t.Fatalf("expected LineCommand, got %T", target.Recipe.Commands[0])
	}

	// Find the raw interpolation
	foundRaw := false
	for _, part := range lineCmd.Parts {
		if interp, ok := part.(*ast.CommandInterpolation); ok {
			if interp.Name == "cflags" && interp.Raw {
				foundRaw = true
				break
			}
		}
	}

	if !foundRaw {
		t.Error("expected to find {cflags:raw} interpolation")
	}
}

// TestParser_ParseRecipe_ShellDirective tests .shell: directive in recipe.
func TestParser_ParseRecipe_ShellDirective(t *testing.T) {
	input := `build/app: src/main.c
    .shell: bash
    gcc -o {target} {deps}
`
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed")
	}

	if target.Recipe.Directives.Shell == nil {
		t.Fatal("expected .shell directive to be parsed")
	}

	// Check the shell value
	if len(target.Recipe.Directives.Shell.Parts) == 0 {
		t.Error("expected shell value to have parts")
	}
}

// TestParser_ParseRecipe_AfterDirective tests .after: directive in recipe.
func TestParser_ParseRecipe_AfterDirective(t *testing.T) {
	input := `build/app: build/main.o
    .after: build/
    gcc -o {target} {deps}
`
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed")
	}

	if len(target.Recipe.Directives.After) != 1 {
		t.Fatalf("expected 1 .after directive, got %d", len(target.Recipe.Directives.After))
	}
}

// TestParser_ParseRecipe_AutodepsDirective tests .autodeps: directive in recipe.
func TestParser_ParseRecipe_AutodepsDirective(t *testing.T) {
	input := `build/{name}.o: src/{name}.c
    .autodeps: build/{name}.d
    gcc -MMD -MF build/{name}.d -c {in} -o {out}
`
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed")
	}

	if target.Recipe.Directives.Autodeps == nil {
		t.Fatal("expected .autodeps directive to be parsed")
	}
}

// TestParser_ParseRecipe_RequiresDirective tests .requires: directive in recipe.
func TestParser_ParseRecipe_RequiresDirective(t *testing.T) {
	input := `@docs: docs/index.html
    .requires: sphinx-build@latest doxygen@latest
    sphinx-build -b html docs/ docs/_build/
`
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed")
	}

	if len(target.Recipe.Directives.Requires) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(target.Recipe.Directives.Requires))
	}

	// Check requirement names
	if target.Recipe.Directives.Requires[0].Name != "sphinx-build" {
		t.Errorf("expected first requirement 'sphinx-build', got %q", target.Recipe.Directives.Requires[0].Name)
	}
	if target.Recipe.Directives.Requires[1].Name != "doxygen" {
		t.Errorf("expected second requirement 'doxygen', got %q", target.Recipe.Directives.Requires[1].Name)
	}
}

// TestParser_ParseRecipe_BlockCommand tests block: with multiple lines.
func TestParser_ParseRecipe_BlockCommand(t *testing.T) {
	input := `build/app: build/main.o
    echo "Starting"
    block:
        if [[ -f {target} ]]; then
            rm {target}
        fi
        gcc -o {target} {deps}
    echo "Finished"
`
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed")
	}

	// Should have: echo, block, echo = 3 commands
	if len(target.Recipe.Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(target.Recipe.Commands))
	}

	// Second command should be a BlockCommand
	blockCmd, ok := target.Recipe.Commands[1].(*ast.BlockCommand)
	if !ok {
		t.Fatalf("expected BlockCommand, got %T", target.Recipe.Commands[1])
	}

	// Block should have 4 lines
	if len(blockCmd.Lines) != 4 {
		t.Errorf("expected 4 lines in block, got %d", len(blockCmd.Lines))
	}
}

// TestParser_ParseRecipe_BlockWithInterpolations tests interpolations inside block.
func TestParser_ParseRecipe_BlockWithInterpolations(t *testing.T) {
	input := `build/app: build/main.o
    block:
        echo "Building {target}"
        {cc} {cflags:raw} -o {target} {deps}
`
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed")
	}

	if len(target.Recipe.Commands) != 1 {
		t.Fatalf("expected 1 command (block), got %d", len(target.Recipe.Commands))
	}

	blockCmd, ok := target.Recipe.Commands[0].(*ast.BlockCommand)
	if !ok {
		t.Fatalf("expected BlockCommand, got %T", target.Recipe.Commands[0])
	}

	// Block should have 2 lines
	if len(blockCmd.Lines) != 2 {
		t.Errorf("expected 2 lines in block, got %d", len(blockCmd.Lines))
	}
}

// TestParser_ParseRecipe_EmptyRecipe tests target with no recipe.
func TestParser_ParseRecipe_EmptyRecipe(t *testing.T) {
	input := `@all: build/app build/test
`
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Recipe should be nil for targets with no indented content
	if target.Recipe != nil {
		t.Error("expected nil recipe for target with no commands")
	}
}

// TestParser_ParseRecipe_Dedent tests that recipe ends on dedent.
func TestParser_ParseRecipe_Dedent(t *testing.T) {
	input := `build/app: src/main.c
    gcc -o {target} {deps}

build/test: src/test.c
    gcc -o {target} {deps}
`
	l := lexer.New("test.need", input)
	p := New(l)

	// Parse first target
	target1, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error on first target: %v", err)
	}

	if target1.Recipe == nil {
		t.Fatal("expected first target to have recipe")
	}
	if len(target1.Recipe.Commands) != 1 {
		t.Errorf("expected first recipe to have 1 command, got %d", len(target1.Recipe.Commands))
	}
}

// TestParser_ParseRecipe_MultipleDirectives tests multiple directives in a recipe.
func TestParser_ParseRecipe_MultipleDirectives(t *testing.T) {
	input := `build/{name}.o: src/{name}.c
    .after: build/
    .autodeps: build/{name}.d
    .shell: bash
    gcc -MMD -MF build/{name}.d -c {in} -o {out}
`
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed")
	}

	if len(target.Recipe.Directives.After) != 1 {
		t.Errorf("expected 1 .after directive, got %d", len(target.Recipe.Directives.After))
	}
	if target.Recipe.Directives.Autodeps == nil {
		t.Error("expected .autodeps directive")
	}
	if target.Recipe.Directives.Shell == nil {
		t.Error("expected .shell directive")
	}
	if len(target.Recipe.Commands) != 1 {
		t.Errorf("expected 1 command, got %d", len(target.Recipe.Commands))
	}
}

// TestParser_ParseRecipe_DirectivesInterspersed tests directives can appear anywhere in recipe.
func TestParser_ParseRecipe_DirectivesInterspersed(t *testing.T) {
	// Per spec, recipe directives should come before commands
	// But we should at least parse them even if interspersed (warn later)
	input := `build/app: src/main.c
    .shell: bash
    echo "Building"
    .requires: pkg-config@latest
    gcc -o {target} {deps}
`
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed")
	}

	// Should have shell directive
	if target.Recipe.Directives.Shell == nil {
		t.Error("expected .shell directive")
	}

	// Should have 2 commands (echo and gcc)
	if len(target.Recipe.Commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(target.Recipe.Commands))
	}
}

// TestParser_ParseRecipe_ScopeTracking tests that parser enters recipe scope.
func TestParser_ParseRecipe_ScopeTracking(t *testing.T) {
	input := `build/app: src/main.c
    gcc -o {target} {deps}
`
	l := lexer.New("test.need", input)
	p := New(l)

	// Initially at global scope
	if p.CurrentScope() != ScopeGlobal {
		t.Errorf("expected global scope, got %v", p.CurrentScope())
	}

	// Parse target
	_, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After parsing, should be back at global scope
	if p.CurrentScope() != ScopeGlobal {
		t.Errorf("expected global scope after parsing, got %v", p.CurrentScope())
	}
}

// TestParser_ParseRecipe_Location tests source location tracking for recipe.
func TestParser_ParseRecipe_Location(t *testing.T) {
	input := `build/app: src/main.c
    gcc -o {target} {deps}
`
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed")
	}

	// Recipe location should be at line 2 (first indented line)
	if target.Recipe.Location.Line != 2 {
		t.Errorf("expected recipe location line 2, got %d", target.Recipe.Location.Line)
	}
}

// TestParser_ParseRecipe_BlockLocation tests source location for block command.
func TestParser_ParseRecipe_BlockLocation(t *testing.T) {
	input := `build/app: src/main.c
    block:
        gcc -o {target} {deps}
`
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Recipe == nil || len(target.Recipe.Commands) == 0 {
		t.Fatal("expected recipe with commands")
	}

	blockCmd, ok := target.Recipe.Commands[0].(*ast.BlockCommand)
	if !ok {
		t.Fatalf("expected BlockCommand, got %T", target.Recipe.Commands[0])
	}

	// Block location should be at line 2
	if blockCmd.Location.Line != 2 {
		t.Errorf("expected block location line 2, got %d", blockCmd.Location.Line)
	}
}

// TestParser_ParseRecipe_TabIndentation tests recipe with tab indentation.
func TestParser_ParseRecipe_TabIndentation(t *testing.T) {
	input := "build/app: src/main.c\n\tgcc -o {target} {deps}\n"
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed")
	}

	if len(target.Recipe.Commands) != 1 {
		t.Errorf("expected 1 command, got %d", len(target.Recipe.Commands))
	}
}

// TestParser_ParseRecipe_VersionParsing tests version spec parsing in .requires.
func TestParser_ParseRecipe_VersionParsing(t *testing.T) {
	tests := []struct {
		name        string
		requirement string
		wantName    string
		wantVersion string
	}{
		{
			name:        "no version",
			requirement: "gcc",
			wantName:    "gcc",
			wantVersion: "latest",
		},
		{
			name:        "latest version",
			requirement: "gcc@latest",
			wantName:    "gcc",
			wantVersion: "latest",
		},
		{
			name:        "major version",
			requirement: "gcc@11",
			wantName:    "gcc",
			wantVersion: "11",
		},
		{
			name:        "major.minor version",
			requirement: "python3@3.10",
			wantName:    "python3",
			wantVersion: "3.10",
		},
		{
			name:        "exact version",
			requirement: "cmake@3.20.1",
			wantName:    "cmake",
			wantVersion: "3.20.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "@test:\n    .requires: " + tt.requirement + "\n    echo test\n"
			l := lexer.New("test.need", input)
			p := New(l)

			target, err := p.ParseTarget()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if target.Recipe == nil {
				t.Fatal("expected recipe")
			}

			if len(target.Recipe.Directives.Requires) != 1 {
				t.Fatalf("expected 1 requirement, got %d", len(target.Recipe.Directives.Requires))
			}

			req := target.Recipe.Directives.Requires[0]
			if req.Name != tt.wantName {
				t.Errorf("name = %q, want %q", req.Name, tt.wantName)
			}
			if req.Version.String() != tt.wantVersion {
				t.Errorf("version = %q, want %q", req.Version.String(), tt.wantVersion)
			}
		})
	}
}

// TestParser_ParseRecipe_InvalidVersion tests that invalid version specs produce errors.
func TestParser_ParseRecipe_InvalidVersion(t *testing.T) {
	tests := []struct {
		name        string
		requirement string
		wantErr     string
	}{
		{
			name:        "non-numeric major",
			requirement: "gcc@abc",
			wantErr:     "invalid major version",
		},
		{
			name:        "non-numeric minor",
			requirement: "gcc@11.x",
			wantErr:     "invalid minor version",
		},
		{
			name:        "non-numeric patch",
			requirement: "gcc@11.4.x",
			wantErr:     "invalid patch version",
		},
		{
			name:        "too many version parts",
			requirement: "gcc@1.2.3.4",
			wantErr:     "invalid version format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "@test:\n    .requires: " + tt.requirement + "\n    echo test\n"
			l := lexer.New("test.need", input)
			p := New(l)

			_, errs := p.ParseNeedfile()
			if !errs.HasErrors() {
				t.Fatalf("expected error for %q, got none", tt.requirement)
			}

			// Check error message contains expected text
			errMsg := errs.Errors[0].Message
			if !strings.Contains(errMsg, tt.wantErr) {
				t.Errorf("error = %q, want to contain %q", errMsg, tt.wantErr)
			}
		})
	}
}

// TestParser_ParseRecipe_BlankLineDoesNotTerminate covers B1: a blank line
// with no leading whitespace inside a recipe must not end the recipe.
func TestParser_ParseRecipe_BlankLineDoesNotTerminate(t *testing.T) {
	input := "build/app: src/main.c\n    echo one\n\n    echo two\n"
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed")
	}
	if len(target.Recipe.Commands) != 2 {
		t.Fatalf("expected 2 commands (blank line should not terminate recipe), got %d", len(target.Recipe.Commands))
	}
}

// TestParser_ParseRecipe_MultipleBlankLines covers B1: several consecutive
// blank lines inside a recipe must not end the recipe.
func TestParser_ParseRecipe_MultipleBlankLines(t *testing.T) {
	input := "build/app: src/main.c\n    echo one\n\n\n    echo two\n"
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(target.Recipe.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(target.Recipe.Commands))
	}
}

// TestParser_ParseRecipe_BlankLineInBlock covers B1 for block: a blank line
// with no leading whitespace inside a block must not end the block.
func TestParser_ParseRecipe_BlankLineInBlock(t *testing.T) {
	input := "build/app: src/main.c\n    block:\n        echo one\n\n        echo two\n"
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(target.Recipe.Commands) != 1 {
		t.Fatalf("expected 1 command (block), got %d", len(target.Recipe.Commands))
	}
	blockCmd, ok := target.Recipe.Commands[0].(*ast.BlockCommand)
	if !ok {
		t.Fatalf("expected BlockCommand, got %T", target.Recipe.Commands[0])
	}
	// 3 lines: "echo one", "" (blank), "echo two"
	if len(blockCmd.Lines) != 3 {
		t.Fatalf("expected 3 lines in block (including blank), got %d", len(blockCmd.Lines))
	}
	if len(blockCmd.Lines[1]) != 0 {
		t.Errorf("expected blank line to be recorded as empty, got %v", blockCmd.Lines[1])
	}
}

// renderLineCommand joins a LineCommand's literal parts back into a string,
// rendering interpolations as {name}. Used to verify internal spaces survive
// parsing.
func renderLineCommand(cmd *ast.LineCommand) string {
	var b strings.Builder
	for _, part := range cmd.Parts {
		switch v := part.(type) {
		case *ast.LiteralCommand:
			b.WriteString(v.Text)
		case *ast.CommandInterpolation:
			b.WriteString("{" + v.Name + "}")
		}
	}
	return b.String()
}

// TestParser_ParseRecipe_DotSlashCommandPreservesSpaces covers B2b: a recipe
// command line starting with "./" must be lexed in command mode so internal
// spaces are preserved, not merged as if it were a directive line.
func TestParser_ParseRecipe_DotSlashCommandPreservesSpaces(t *testing.T) {
	input := "test: build/app\n    ./{build_dir}/app --test\n"
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(target.Recipe.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(target.Recipe.Commands))
	}
	lineCmd, ok := target.Recipe.Commands[0].(*ast.LineCommand)
	if !ok {
		t.Fatalf("expected LineCommand, got %T", target.Recipe.Commands[0])
	}
	got := renderLineCommand(lineCmd)
	want := "./{build_dir}/app --test"
	if got != want {
		t.Errorf("command = %q, want %q", got, want)
	}
}

// TestParser_ParseRecipe_InvalidModifier covers B4: an invalid interpolation
// modifier inside a recipe command line must surface as a parse error, not
// be executed as literal text.
func TestParser_ParseRecipe_InvalidModifier(t *testing.T) {
	input := "@test:\n    echo \"{msg:invalid}\"\n"
	l := lexer.New("test.need", input)
	p := New(l)

	_, err := p.ParseTarget()
	if err == nil {
		t.Fatal("expected error for invalid modifier, got none")
	}
	want := "invalid modifier ':invalid', expected ':raw'"
	if err.Message != want {
		t.Errorf("error message = %q, want %q", err.Message, want)
	}
}

// TestParser_ParseRecipe_MixedIndentation covers B4/100/101: mixed or
// inconsistent indentation inside a recipe must surface the lexer's
// indentation message, not "unexpected token: ERROR".
func TestParser_ParseRecipe_MixedIndentation(t *testing.T) {
	input := "@test:\n\techo \"with tab\"\n    echo \"with spaces\"\n"
	l := lexer.New("test.need", input)
	p := New(l)

	_, errs := p.ParseNeedfile()
	if !errs.HasErrors() {
		t.Fatal("expected a parse error for mixed indentation, got none")
	}
	for _, e := range errs.Errors {
		if strings.Contains(e.Message, "unexpected token: ERROR") {
			t.Errorf("error message leaked raw ERROR token: %q", e.Message)
		}
	}
	found := false
	for _, e := range errs.Errors {
		if strings.Contains(e.Message, "indentation") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected an error mentioning 'indentation', got %v", errs.Errors)
	}
}

// TestParser_ParseRecipe_UnknownDirective covers B5 inside a recipe: an
// unknown dot-keyword immediately followed by ':' must be reported as an
// unknown directive, not silently treated as a command or a target.
func TestParser_ParseRecipe_UnknownDirective(t *testing.T) {
	input := "@test:\n    .unknownthing: 1\n    echo done\n"
	l := lexer.New("test.need", input)
	p := New(l)

	_, err := p.ParseTarget()
	if err == nil {
		t.Fatal("expected error for unknown directive in recipe, got none")
	}
	if !strings.Contains(err.Message, "unknown directive '.unknownthing'") {
		t.Errorf("error message = %q, want to contain %q", err.Message, "unknown directive '.unknownthing'")
	}
}

// TestParser_ParseRecipe_AfterDirectiveStillWorks is a regression check
// alongside the B5 recipe-scope fix: a real, recognized directive like
// .after: must still parse normally (not be mistaken for an unknown one).
func TestParser_ParseRecipe_AfterDirectiveStillWorks(t *testing.T) {
	input := "build/app: build/main.o\n    .after: build/\n    gcc -o {target} {deps}\n"
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(target.Recipe.Directives.After) != 1 {
		t.Fatalf("expected 1 .after directive, got %d", len(target.Recipe.Directives.After))
	}
}

// TestParser_ParseRecipe_BlankLinesBeforeFirstCommand covers B1's edge case:
// blank lines between the target line and the first recipe line (before the
// recipe-parsing loop even starts) must not be mistaken for "this target has
// no recipe".
func TestParser_ParseRecipe_BlankLinesBeforeFirstCommand(t *testing.T) {
	input := "@test:\n\n\n    echo \"{var1}\"\n\n\n    echo \"{var2}\"\n"
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Recipe == nil {
		t.Fatal("expected recipe to be parsed, got nil (blank lines before first command were mistaken for 'no recipe')")
	}
	if len(target.Recipe.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(target.Recipe.Commands))
	}
}
