package parser

import (
	"testing"

	"github.com/vinayprograms/build/internal/lexer"
)

func TestDirectiveValidation_GlobalScope(t *testing.T) {
	tests := []struct {
		name      string
		directive lexer.TokenType
		wantValid bool
	}{
		// Global-only directives
		{"shell at global", lexer.DOT_SHELL, true},
		{"parallel at global", lexer.DOT_PARALLEL, true},
		{"default at global", lexer.DOT_DEFAULT, true},
		{"include at global", lexer.DOT_INCLUDE, true},
		{"environment at global", lexer.DOT_ENVIRONMENT, true},

		// Environment-only directives (invalid at global)
		{"using at global", lexer.DOT_USING, false},
		{"source at global", lexer.DOT_SOURCE, false},
		{"args at global", lexer.DOT_ARGS, false},

		// Recipe-only directives (invalid at global)
		{"after at global", lexer.DOT_AFTER, false},
		{"autodeps at global", lexer.DOT_AUTODEPS, false},

		// Context-dependent (valid at global for environment, invalid otherwise)
		{"requires at global", lexer.DOT_REQUIRES, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDirectiveValidAtScope(tt.directive, ScopeGlobal)
			if got != tt.wantValid {
				t.Errorf("IsDirectiveValidAtScope(%v, ScopeGlobal) = %v, want %v",
					tt.directive, got, tt.wantValid)
			}
		})
	}
}

func TestDirectiveValidation_EnvironmentScope(t *testing.T) {
	tests := []struct {
		name      string
		directive lexer.TokenType
		wantValid bool
	}{
		// Global directives (invalid in environment)
		{"shell in environment", lexer.DOT_SHELL, false},
		{"parallel in environment", lexer.DOT_PARALLEL, false},
		{"default in environment", lexer.DOT_DEFAULT, false},
		{"include in environment", lexer.DOT_INCLUDE, false},
		{"environment in environment", lexer.DOT_ENVIRONMENT, false},

		// Environment-specific directives
		{"using in environment", lexer.DOT_USING, true},
		{"source in environment", lexer.DOT_SOURCE, true},
		{"args in environment", lexer.DOT_ARGS, true},
		{"requires in environment", lexer.DOT_REQUIRES, true},

		// Recipe-only directives (invalid in environment)
		{"after in environment", lexer.DOT_AFTER, false},
		{"autodeps in environment", lexer.DOT_AUTODEPS, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDirectiveValidAtScope(tt.directive, ScopeEnvironment)
			if got != tt.wantValid {
				t.Errorf("IsDirectiveValidAtScope(%v, ScopeEnvironment) = %v, want %v",
					tt.directive, got, tt.wantValid)
			}
		})
	}
}

func TestDirectiveValidation_RecipeScope(t *testing.T) {
	tests := []struct {
		name      string
		directive lexer.TokenType
		wantValid bool
	}{
		// Global directives (invalid in recipe except .shell)
		{"shell in recipe", lexer.DOT_SHELL, true}, // Can override shell
		{"parallel in recipe", lexer.DOT_PARALLEL, false},
		{"default in recipe", lexer.DOT_DEFAULT, false},
		{"include in recipe", lexer.DOT_INCLUDE, false},
		{"environment in recipe", lexer.DOT_ENVIRONMENT, false},

		// Environment-specific directives (invalid in recipe)
		{"using in recipe", lexer.DOT_USING, false},
		{"source in recipe", lexer.DOT_SOURCE, false},
		{"args in recipe", lexer.DOT_ARGS, false},

		// Recipe-specific directives
		{"after in recipe", lexer.DOT_AFTER, true},
		{"autodeps in recipe", lexer.DOT_AUTODEPS, true},
		{"requires in recipe", lexer.DOT_REQUIRES, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDirectiveValidAtScope(tt.directive, ScopeRecipe)
			if got != tt.wantValid {
				t.Errorf("IsDirectiveValidAtScope(%v, ScopeRecipe) = %v, want %v",
					tt.directive, got, tt.wantValid)
			}
		})
	}
}

func TestDirectiveValidation_BlockScope(t *testing.T) {
	// Block scope is inside recipe, so directives that are valid in recipe
	// should still be checked, but typically no directives appear in block
	tests := []struct {
		name      string
		directive lexer.TokenType
		wantValid bool
	}{
		// No directives are valid in block scope
		{"shell in block", lexer.DOT_SHELL, false},
		{"parallel in block", lexer.DOT_PARALLEL, false},
		{"default in block", lexer.DOT_DEFAULT, false},
		{"include in block", lexer.DOT_INCLUDE, false},
		{"environment in block", lexer.DOT_ENVIRONMENT, false},
		{"using in block", lexer.DOT_USING, false},
		{"source in block", lexer.DOT_SOURCE, false},
		{"args in block", lexer.DOT_ARGS, false},
		{"requires in block", lexer.DOT_REQUIRES, false},
		{"after in block", lexer.DOT_AFTER, false},
		{"autodeps in block", lexer.DOT_AUTODEPS, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDirectiveValidAtScope(tt.directive, ScopeBlock)
			if got != tt.wantValid {
				t.Errorf("IsDirectiveValidAtScope(%v, ScopeBlock) = %v, want %v",
					tt.directive, got, tt.wantValid)
			}
		})
	}
}

func TestValidScopesForDirective(t *testing.T) {
	tests := []struct {
		directive lexer.TokenType
		want      []Scope
	}{
		{lexer.DOT_SHELL, []Scope{ScopeGlobal, ScopeRecipe}},
		{lexer.DOT_PARALLEL, []Scope{ScopeGlobal}},
		{lexer.DOT_DEFAULT, []Scope{ScopeGlobal}},
		{lexer.DOT_INCLUDE, []Scope{ScopeGlobal}},
		{lexer.DOT_ENVIRONMENT, []Scope{ScopeGlobal}},
		{lexer.DOT_USING, []Scope{ScopeEnvironment}},
		{lexer.DOT_SOURCE, []Scope{ScopeEnvironment}},
		{lexer.DOT_ARGS, []Scope{ScopeEnvironment}},
		{lexer.DOT_REQUIRES, []Scope{ScopeEnvironment, ScopeRecipe}},
		{lexer.DOT_AFTER, []Scope{ScopeRecipe}},
		{lexer.DOT_AUTODEPS, []Scope{ScopeRecipe}},
	}

	for _, tt := range tests {
		t.Run(tt.directive.String(), func(t *testing.T) {
			got := ValidScopesForDirective(tt.directive)
			if len(got) != len(tt.want) {
				t.Errorf("ValidScopesForDirective(%v) = %v, want %v",
					tt.directive, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ValidScopesForDirective(%v) = %v, want %v",
						tt.directive, got, tt.want)
					return
				}
			}
		})
	}
}
