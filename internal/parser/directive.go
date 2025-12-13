package parser

import "github.com/vinayprograms/build/internal/lexer"

// directiveScopes maps each directive to its valid scopes.
// This follows the rules from DESIGN.md Section 3.3.3.
var directiveScopes = map[lexer.TokenType][]Scope{
	// Global-only directives
	lexer.DOT_PARALLEL:    {ScopeGlobal},
	lexer.DOT_DEFAULT:     {ScopeGlobal},
	lexer.DOT_INCLUDE:     {ScopeGlobal},
	lexer.DOT_ENVIRONMENT: {ScopeGlobal},

	// .shell is valid at global (default) and recipe (override)
	lexer.DOT_SHELL: {ScopeGlobal, ScopeRecipe},

	// Environment-only directives
	lexer.DOT_USING:  {ScopeEnvironment},
	lexer.DOT_SOURCE: {ScopeEnvironment},
	lexer.DOT_ARGS:   {ScopeEnvironment},

	// .requires is valid in environment (requirements) and recipe (binary requirements)
	lexer.DOT_REQUIRES: {ScopeEnvironment, ScopeRecipe},

	// Recipe-only directives
	lexer.DOT_AFTER:    {ScopeRecipe},
	lexer.DOT_AUTODEPS: {ScopeRecipe},
}

// IsDirectiveValidAtScope returns true if the directive is valid at the given scope.
func IsDirectiveValidAtScope(directive lexer.TokenType, scope Scope) bool {
	validScopes, ok := directiveScopes[directive]
	if !ok {
		return false
	}
	for _, s := range validScopes {
		if s == scope {
			return true
		}
	}
	return false
}

// ValidScopesForDirective returns the list of scopes where a directive is valid.
func ValidScopesForDirective(directive lexer.TokenType) []Scope {
	if scopes, ok := directiveScopes[directive]; ok {
		return scopes
	}
	return nil
}
