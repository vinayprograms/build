package main

import (
	"github.com/vinayprograms/build/internal/lexer"
	"github.com/vinayprograms/build/internal/parser"
)

// ----------------------------------------------------------------------------
// Lexer Adapters
// ----------------------------------------------------------------------------

// tokenAdapter wraps lexer.Token to implement the Token interface.
type tokenAdapter struct {
	tok lexer.Token
}

func (t tokenAdapter) TokenType() string     { return t.tok.Type.String() }
func (t tokenAdapter) TokenLiteral() string  { return t.tok.Literal }
func (t tokenAdapter) TokenLocation() string { return t.tok.Location.String() }
func (t tokenAdapter) IsEOF() bool           { return t.tok.Type == lexer.EOF }
func (t tokenAdapter) IsError() bool         { return t.tok.Type == lexer.ERROR }

// lexerAdapter wraps lexer.Lexer to implement the Lexer interface.
type lexerAdapter struct {
	lex *lexer.Lexer
}

func (l *lexerAdapter) NextToken() Token {
	return tokenAdapter{tok: l.lex.NextToken()}
}

// NewLexer creates a new Lexer for the given source.
// This is the LexerFactory implementation using the internal/lexer package.
func NewLexer(file, input string) Lexer {
	return &lexerAdapter{lex: lexer.New(file, input)}
}

// ----------------------------------------------------------------------------
// Parser Adapters
// ----------------------------------------------------------------------------

// scopeAdapter wraps parser.Scope to implement the Scope interface.
type scopeAdapter struct {
	scope parser.Scope
}

func (s scopeAdapter) String() string { return s.scope.String() }

// Scope constants exposed for external use.
var (
	ScopeGlobal      Scope = scopeAdapter{scope: parser.ScopeGlobal}
	ScopeEnvironment Scope = scopeAdapter{scope: parser.ScopeEnvironment}
	ScopeRecipe      Scope = scopeAdapter{scope: parser.ScopeRecipe}
	ScopeBlock       Scope = scopeAdapter{scope: parser.ScopeBlock}
)

// parserAdapter wraps parser.Parser to implement the Parser interface.
type parserAdapter struct {
	p *parser.Parser
}

func (p *parserAdapter) CurrentScope() Scope {
	return scopeAdapter{scope: p.p.CurrentScope()}
}

func (p *parserAdapter) EnterScope(scope Scope) {
	// Extract the underlying parser.Scope from the adapter
	if sa, ok := scope.(scopeAdapter); ok {
		p.p.EnterScope(sa.scope)
	}
}

func (p *parserAdapter) ExitScope() Scope {
	return scopeAdapter{scope: p.p.ExitScope()}
}

func (p *parserAdapter) CurrentIndentLevel() int {
	return p.p.CurrentIndentLevel()
}

func (p *parserAdapter) HasErrors() bool {
	return p.p.HasErrors()
}

// NewParser creates a new Parser for the given Lexer.
// Note: This function needs access to the underlying lexer.Lexer,
// so it expects a lexerAdapter. For maximum flexibility, it also
// accepts a raw *lexer.Lexer.
func NewParser(lex Lexer) Parser {
	// Extract the underlying lexer.Lexer from the adapter
	if la, ok := lex.(*lexerAdapter); ok {
		return &parserAdapter{p: parser.New(la.lex)}
	}
	// This shouldn't happen in normal use, but handle gracefully
	panic("NewParser requires a Lexer created by NewLexer")
}

// NewParserFromLexer creates a Parser directly from a lexer.Lexer.
// This is useful when you need to work with both interfaces and concrete types.
func NewParserFromLexer(lex *lexer.Lexer) Parser {
	return &parserAdapter{p: parser.New(lex)}
}

// ----------------------------------------------------------------------------
// Directive Validation
// ----------------------------------------------------------------------------

// directiveValidatorImpl implements DirectiveValidator using the parser package.
type directiveValidatorImpl struct{}

func (d directiveValidatorImpl) IsValidAtScope(tokenType string, scope Scope) bool {
	// Convert string token type back to lexer.TokenType
	tokType := tokenTypeFromString(tokenType)
	if tokType == lexer.EOF {
		return false
	}
	// Extract underlying parser.Scope
	if sa, ok := scope.(scopeAdapter); ok {
		return parser.IsDirectiveValidAtScope(tokType, sa.scope)
	}
	return false
}

func (d directiveValidatorImpl) DirectiveName(tokenType string) string {
	tokType := tokenTypeFromString(tokenType)
	return parser.DirectiveNameForError(tokType)
}

// NewDirectiveValidator creates a DirectiveValidator.
func NewDirectiveValidator() DirectiveValidator {
	return directiveValidatorImpl{}
}

// tokenTypeFromString converts a token type string back to lexer.TokenType.
// This is needed for the directive validator interface.
func tokenTypeFromString(s string) lexer.TokenType {
	// Map string names back to token types
	switch s {
	case "DOT_SHELL":
		return lexer.DOT_SHELL
	case "DOT_PARALLEL":
		return lexer.DOT_PARALLEL
	case "DOT_DEFAULT":
		return lexer.DOT_DEFAULT
	case "DOT_INCLUDE":
		return lexer.DOT_INCLUDE
	case "DOT_ENVIRONMENT":
		return lexer.DOT_ENVIRONMENT
	case "DOT_USING":
		return lexer.DOT_USING
	case "DOT_SOURCE":
		return lexer.DOT_SOURCE
	case "DOT_ARGS":
		return lexer.DOT_ARGS
	case "DOT_REQUIRES":
		return lexer.DOT_REQUIRES
	case "DOT_AFTER":
		return lexer.DOT_AFTER
	case "DOT_AUTODEPS":
		return lexer.DOT_AUTODEPS
	default:
		return lexer.EOF
	}
}
