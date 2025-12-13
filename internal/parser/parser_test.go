package parser

import (
	"testing"

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
	p.enterScope(ScopeEnvironment)
	if p.scope.Current() != ScopeEnvironment {
		t.Errorf("after entering environment, scope = %v, want ScopeEnvironment", p.scope.Current())
	}

	// Exit environment
	p.exitScope()
	if p.scope.Current() != ScopeGlobal {
		t.Errorf("after exiting environment, scope = %v, want ScopeGlobal", p.scope.Current())
	}

	// Enter recipe
	p.enterScope(ScopeRecipe)
	if p.scope.Current() != ScopeRecipe {
		t.Errorf("after entering recipe, scope = %v, want ScopeRecipe", p.scope.Current())
	}

	// Enter block inside recipe
	p.enterScope(ScopeBlock)
	if p.scope.Current() != ScopeBlock {
		t.Errorf("after entering block, scope = %v, want ScopeBlock", p.scope.Current())
	}

	// Exit block
	p.exitScope()
	if p.scope.Current() != ScopeRecipe {
		t.Errorf("after exiting block, scope = %v, want ScopeRecipe", p.scope.Current())
	}

	// Exit recipe
	p.exitScope()
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
			p.enterScope(ScopeEnvironment)

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
			p.enterScope(ScopeRecipe)

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
	if p.currentIndentLevel() != 0 {
		t.Errorf("global indent level = %d, want 0", p.currentIndentLevel())
	}

	// Environment is at level 1
	p.enterScope(ScopeEnvironment)
	if p.currentIndentLevel() != 1 {
		t.Errorf("environment indent level = %d, want 1", p.currentIndentLevel())
	}

	p.exitScope()

	// Recipe is at level 1
	p.enterScope(ScopeRecipe)
	if p.currentIndentLevel() != 1 {
		t.Errorf("recipe indent level = %d, want 1", p.currentIndentLevel())
	}

	// Block inside recipe is at level 2
	p.enterScope(ScopeBlock)
	if p.currentIndentLevel() != 2 {
		t.Errorf("block indent level = %d, want 2", p.currentIndentLevel())
	}
}
