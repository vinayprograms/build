package parser

import (
	"fmt"
	"strings"

	"github.com/vinayprograms/build/internal/lexer"
)

// ParseError represents an error encountered during parsing.
type ParseError struct {
	Message  string
	Location lexer.SourceLocation
	Hint     string // Optional hint for fixing the error
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	msg := fmt.Sprintf("%s: %s", e.Location.String(), e.Message)
	if e.Hint != "" {
		msg += " (hint: " + e.Hint + ")"
	}
	return msg
}

// ParseErrors collects multiple parse errors.
type ParseErrors struct {
	Errors []*ParseError
}

// Add appends an error to the collection.
func (pe *ParseErrors) Add(err *ParseError) {
	pe.Errors = append(pe.Errors, err)
}

// HasErrors returns true if there are any errors.
func (pe *ParseErrors) HasErrors() bool {
	return len(pe.Errors) > 0
}

// Error implements the error interface.
func (pe *ParseErrors) Error() string {
	var msgs []string
	for _, err := range pe.Errors {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "\n")
}

// directiveNames maps directive token types to their display names.
var directiveNames = map[lexer.TokenType]string{
	lexer.DOT_SHELL:       ".shell",
	lexer.DOT_PARALLEL:    ".parallel",
	lexer.DOT_DEFAULT:     ".default",
	lexer.DOT_INCLUDE:     ".include",
	lexer.DOT_ENVIRONMENT: ".environment",
	lexer.DOT_USING:       ".using",
	lexer.DOT_SOURCE:      ".source",
	lexer.DOT_ARGS:        ".args",
	lexer.DOT_REQUIRES:    ".requires",
	lexer.DOT_AFTER:       ".after",
	lexer.DOT_AUTODEPS:    ".autodeps",
}

// DirectiveNameForError returns the directive name for error messages.
func DirectiveNameForError(tok lexer.TokenType) string {
	if name, ok := directiveNames[tok]; ok {
		return name
	}
	return tok.String()
}

// errorFromLexerToken converts a lexer.ERROR token into a ParseError, using the
// token's Literal (which carries the lexer's diagnostic message, e.g. "unclosed
// interpolation: {world") as the error message. This is the single place that
// should be used to surface lexer.ERROR tokens as parse errors.
func (p *Parser) errorFromLexerToken(tok lexer.Token) *ParseError {
	return &ParseError{
		Message:  tok.Literal,
		Location: tok.Location,
	}
}

// NewScopeError creates a parse error for a directive used at an invalid scope.
func NewScopeError(directive lexer.TokenType, found Scope, loc lexer.SourceLocation) *ParseError {
	name := DirectiveNameForError(directive)
	validScopes := ValidScopesForDirective(directive)

	var scopeNames []string
	for _, s := range validScopes {
		scopeNames = append(scopeNames, s.String())
	}

	return &ParseError{
		Message:  fmt.Sprintf("directive '%s' invalid at %s scope", name, found.String()),
		Location: loc,
		Hint:     fmt.Sprintf("%s is only valid in: %s", name, strings.Join(scopeNames, ", ")),
	}
}
