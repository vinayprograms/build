package parser

import (
	"testing"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/lexer"
)

func TestParser_New(t *testing.T) {
	l := lexer.New("test.build", ".shell: bash")
	p := New(l)

	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.scope == nil {
		t.Error("parser scope stack not initialized")
	}
	if p.scope.Current() != ScopeGlobal {
		t.Errorf("initial scope = %v, want ScopeGlobal", p.scope.Current())
	}
}

func TestParser_ValidateDirective_GlobalScope(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		// Valid global directives
		{".shell at global", ".shell: bash", false},
		{".parallel at global", ".parallel: 4", false},
		{".default at global", ".default: @all", false},
		{".include at global", ".include: ./common.build", false},

		// Invalid directives at global
		{".using at global", ".using: docker", true},
		{".source at global", ".source: ./Dockerfile", true},
		{".args at global", ".args: --platform linux", true},
		{".after at global", ".after: build/", true},
		{".autodeps at global", ".autodeps: deps.d", true},
		{".requires at global", ".requires: gcc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.build", tt.input)
			p := New(l)

			// Parse until we get a directive token
			tok := p.nextToken()
			for tok.Type != lexer.EOF && !tok.Type.IsDotKeyword() {
				tok = p.nextToken()
			}

			if tok.Type.IsDotKeyword() {
				err := p.validateDirectiveScope(tok)
				if tt.wantError && err == nil {
					t.Error("expected error but got nil")
				}
				if !tt.wantError && err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestParser_ScopeTransitions(t *testing.T) {
	// Test that scope changes correctly when entering/exiting blocks
	l := lexer.New("test.build", "")
	p := New(l)

	// Initially at global
	if p.scope.Current() != ScopeGlobal {
		t.Fatalf("initial scope = %v, want ScopeGlobal", p.scope.Current())
	}

	// Enter environment
	p.EnterScope(ScopeEnvironment)
	if p.scope.Current() != ScopeEnvironment {
		t.Errorf("after entering environment, scope = %v, want ScopeEnvironment", p.scope.Current())
	}

	// Exit environment
	p.ExitScope()
	if p.scope.Current() != ScopeGlobal {
		t.Errorf("after exiting environment, scope = %v, want ScopeGlobal", p.scope.Current())
	}

	// Enter recipe
	p.EnterScope(ScopeRecipe)
	if p.scope.Current() != ScopeRecipe {
		t.Errorf("after entering recipe, scope = %v, want ScopeRecipe", p.scope.Current())
	}

	// Enter block inside recipe
	p.EnterScope(ScopeBlock)
	if p.scope.Current() != ScopeBlock {
		t.Errorf("after entering block, scope = %v, want ScopeBlock", p.scope.Current())
	}

	// Exit block
	p.ExitScope()
	if p.scope.Current() != ScopeRecipe {
		t.Errorf("after exiting block, scope = %v, want ScopeRecipe", p.scope.Current())
	}

	// Exit recipe
	p.ExitScope()
	if p.scope.Current() != ScopeGlobal {
		t.Errorf("after exiting recipe, scope = %v, want ScopeGlobal", p.scope.Current())
	}
}

func TestParser_DirectiveInEnvironmentScope(t *testing.T) {
	tests := []struct {
		name      string
		directive lexer.TokenType
		wantValid bool
	}{
		{".using in environment", lexer.DOT_USING, true},
		{".source in environment", lexer.DOT_SOURCE, true},
		{".args in environment", lexer.DOT_ARGS, true},
		{".requires in environment", lexer.DOT_REQUIRES, true},

		// Invalid in environment
		{".shell in environment", lexer.DOT_SHELL, false},
		{".parallel in environment", lexer.DOT_PARALLEL, false},
		{".after in environment", lexer.DOT_AFTER, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.build", "")
			p := New(l)
			p.EnterScope(ScopeEnvironment)

			tok := lexer.Token{
				Type:     tt.directive,
				Literal:  DirectiveNameForError(tt.directive),
				Location: lexer.SourceLocation{File: "test.build", Line: 1, Column: 1},
			}

			err := p.validateDirectiveScope(tok)
			if tt.wantValid && err != nil {
				t.Errorf("expected valid but got error: %v", err)
			}
			if !tt.wantValid && err == nil {
				t.Error("expected error but got nil")
			}
		})
	}
}

func TestParser_DirectiveInRecipeScope(t *testing.T) {
	tests := []struct {
		name      string
		directive lexer.TokenType
		wantValid bool
	}{
		{".shell in recipe", lexer.DOT_SHELL, true},
		{".after in recipe", lexer.DOT_AFTER, true},
		{".autodeps in recipe", lexer.DOT_AUTODEPS, true},
		{".requires in recipe", lexer.DOT_REQUIRES, true},

		// Invalid in recipe
		{".parallel in recipe", lexer.DOT_PARALLEL, false},
		{".using in recipe", lexer.DOT_USING, false},
		{".environment in recipe", lexer.DOT_ENVIRONMENT, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.build", "")
			p := New(l)
			p.EnterScope(ScopeRecipe)

			tok := lexer.Token{
				Type:     tt.directive,
				Literal:  DirectiveNameForError(tt.directive),
				Location: lexer.SourceLocation{File: "test.build", Line: 1, Column: 1},
			}

			err := p.validateDirectiveScope(tok)
			if tt.wantValid && err != nil {
				t.Errorf("expected valid but got error: %v", err)
			}
			if !tt.wantValid && err == nil {
				t.Error("expected error but got nil")
			}
		})
	}
}

func TestParser_CurrentIndentLevel(t *testing.T) {
	l := lexer.New("test.build", "")
	p := New(l)

	// Global starts at level 0
	if p.CurrentIndentLevel() != 0 {
		t.Errorf("global indent level = %d, want 0", p.CurrentIndentLevel())
	}

	// Environment is at level 1
	p.EnterScope(ScopeEnvironment)
	if p.CurrentIndentLevel() != 1 {
		t.Errorf("environment indent level = %d, want 1", p.CurrentIndentLevel())
	}

	p.ExitScope()

	// Recipe is at level 1
	p.EnterScope(ScopeRecipe)
	if p.CurrentIndentLevel() != 1 {
		t.Errorf("recipe indent level = %d, want 1", p.CurrentIndentLevel())
	}

	// Block inside recipe is at level 2
	p.EnterScope(ScopeBlock)
	if p.CurrentIndentLevel() != 2 {
		t.Errorf("block indent level = %d, want 2", p.CurrentIndentLevel())
	}
}

func TestParser_RecipeCommandSpaces(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantCmdText string
	}{
		{
			name: "simple command with spaces",
			input: `@clean:
    rm -rf build`,
			wantCmdText: "rm -rf build",
		},
		{
			name: "command with multiple spaces",
			input: `@test:
    go test   -v   ./...`,
			wantCmdText: "go test   -v   ./...",
		},
		{
			name: "command with interpolation and spaces",
			input: `@build:
    gcc -o {target} {deps}`,
			wantCmdText: "gcc -o {target} {deps}",
		},
		{
			name: "command with spaces around interpolation",
			input: `@cmd:
    cmd {var} rest`,
			wantCmdText: "cmd {var} rest",
		},
		{
			name: "complex command with flags",
			input: `@build:
    go build -ldflags "-s -w" -o bin/app ./cmd/app`,
			wantCmdText: "go build -ldflags \"-s -w\" -o bin/app ./cmd/app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.build", tt.input)
			p := New(l)
			stmts, errs := p.ParseBuildfile()

			if errs.HasErrors() {
				t.Fatalf("parse errors: %v", errs.Error())
			}

			if len(stmts) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(stmts))
			}

			target, ok := stmts[0].(*ast.Target)
			if !ok {
				t.Fatalf("expected *ast.Target, got %T", stmts[0])
			}

			if target.Recipe == nil {
				t.Fatal("target has no recipe")
			}

			if len(target.Recipe.Commands) != 1 {
				t.Fatalf("expected 1 command, got %d", len(target.Recipe.Commands))
			}

			lineCmd, ok := target.Recipe.Commands[0].(*ast.LineCommand)
			if !ok {
				t.Fatalf("command is not LineCommand, got %T", target.Recipe.Commands[0])
			}

			// Reconstruct the command text
			var cmdText string
			for _, part := range lineCmd.Parts {
				switch p := part.(type) {
				case *ast.LiteralCommand:
					cmdText += p.Text
				case *ast.CommandInterpolation:
					if p.Raw {
						cmdText += "{" + p.Name + ":raw}"
					} else {
						cmdText += "{" + p.Name + "}"
					}
				}
			}

			if cmdText != tt.wantCmdText {
				t.Errorf("command text = %q, want %q", cmdText, tt.wantCmdText)
			}
		})
	}
}

func TestParser_BlockCommandSpaces(t *testing.T) {
	input := `@build:
    block:
        echo "Building..."
        go build -ldflags "-s -w" -o {target} ./cmd
        echo "Done"`

	l := lexer.New("test.build", input)
	p := New(l)
	stmts, errs := p.ParseBuildfile()

	if errs.HasErrors() {
		t.Fatalf("parse errors: %v", errs.Error())
	}

	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}

	target, ok := stmts[0].(*ast.Target)
	if !ok {
		t.Fatalf("expected *ast.Target, got %T", stmts[0])
	}

	if target.Recipe == nil {
		t.Fatal("target has no recipe")
	}

	if len(target.Recipe.Commands) != 1 {
		t.Fatalf("expected 1 command (block), got %d", len(target.Recipe.Commands))
	}

	blockCmd, ok := target.Recipe.Commands[0].(*ast.BlockCommand)
	if !ok {
		t.Fatalf("command is not BlockCommand, got %T", target.Recipe.Commands[0])
	}

	if len(blockCmd.Lines) != 3 {
		t.Fatalf("expected 3 block lines, got %d", len(blockCmd.Lines))
	}

	// Check the second line (go build command)
	var line2Text string
	for _, part := range blockCmd.Lines[1] {
		switch p := part.(type) {
		case *ast.LiteralCommand:
			line2Text += p.Text
		case *ast.CommandInterpolation:
			line2Text += "{" + p.Name + "}"
		}
	}

	wantLine2 := "go build -ldflags \"-s -w\" -o {target} ./cmd"
	if line2Text != wantLine2 {
		t.Errorf("block line 2 = %q, want %q", line2Text, wantLine2)
	}
}

func TestParser_RecipeDirectiveWithCommand(t *testing.T) {
	// Test that directives are still parsed correctly alongside commands
	input := `@build:
    .after: clean
    go build -o {target}`

	l := lexer.New("test.build", input)
	p := New(l)
	stmts, errs := p.ParseBuildfile()

	if errs.HasErrors() {
		t.Fatalf("parse errors: %v", errs.Error())
	}

	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}

	target, ok := stmts[0].(*ast.Target)
	if !ok {
		t.Fatalf("expected *ast.Target, got %T", stmts[0])
	}

	if target.Recipe == nil {
		t.Fatal("target has no recipe")
	}

	// Should have .after directive
	if len(target.Recipe.Directives.After) != 1 {
		t.Fatalf("expected 1 .after directive, got %d", len(target.Recipe.Directives.After))
	}

	// Should have 1 command
	if len(target.Recipe.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(target.Recipe.Commands))
	}

	// Check command has proper spaces
	lineCmd, ok := target.Recipe.Commands[0].(*ast.LineCommand)
	if !ok {
		t.Fatalf("command is not LineCommand, got %T", target.Recipe.Commands[0])
	}

	var cmdText string
	for _, part := range lineCmd.Parts {
		switch p := part.(type) {
		case *ast.LiteralCommand:
			cmdText += p.Text
		case *ast.CommandInterpolation:
			cmdText += "{" + p.Name + "}"
		}
	}

	wantCmd := "go build -o {target}"
	if cmdText != wantCmd {
		t.Errorf("command text = %q, want %q", cmdText, wantCmd)
	}
}
