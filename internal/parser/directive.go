package parser

import (
	"fmt"
	"sort"
	"strings"

	"github.com/vinayprograms/need/internal/lexer"
)

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

// validDirectiveNamesForScope returns the sorted names of all directives valid
// at the given scope, for use in "unknown directive" error hints.
func validDirectiveNamesForScope(scope Scope) []string {
	var names []string
	for tok, scopes := range directiveScopes {
		for _, s := range scopes {
			if s == scope {
				names = append(names, DirectiveNameForError(tok))
				break
			}
		}
	}
	sort.Strings(names)
	return names
}

// isUnknownDirectiveCandidate reports whether text has the shape of an
// unrecognized dot-keyword: a leading '.' followed by one or more identifier
// characters (letters, digits, underscore) and nothing else - in particular,
// no '/'. This is exactly the shape the lexer produces (via lexDotKeyword)
// for a dot-prefixed word it doesn't recognize as a directive, since the
// keyword scan stops at the first non-identifier character.
func isUnknownDirectiveCandidate(text string) bool {
	if len(text) < 2 || text[0] != '.' {
		return false
	}
	for i := 1; i < len(text); i++ {
		ch := text[i]
		if !(ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
			return false
		}
	}
	return true
}

// unknownDirectiveError builds a ParseError for a dot-prefixed token that
// looks like a directive but isn't one that's recognized.
func (p *Parser) unknownDirectiveError(name string, loc lexer.SourceLocation) *ParseError {
	valid := validDirectiveNamesForScope(p.CurrentScope())
	const fileTargetHint = "prefix with ./ to use it as a file target"
	hint := fileTargetHint
	if len(valid) > 0 {
		hint = fmt.Sprintf("valid directives here: %s; %s", strings.Join(valid, ", "), fileTargetHint)
	}
	return &ParseError{
		Message:  fmt.Sprintf("unknown directive '%s'", name),
		Location: loc,
		Hint:     hint,
	}
}
