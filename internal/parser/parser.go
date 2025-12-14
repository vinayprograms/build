package parser

import (
	"github.com/vinayprograms/build/internal/lexer"
)

// Parser transforms a token stream into an AST.
type Parser struct {
	lexer   *lexer.Lexer
	current lexer.Token
	scope   *ScopeStack
	errors  *ParseErrors
}

// New creates a new Parser for the given lexer.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		scope:  NewScopeStack(),
		errors: &ParseErrors{},
	}
	// Prime the parser with the first token
	p.current = p.lexer.NextToken()
	return p
}

// nextToken advances to the next token.
func (p *Parser) nextToken() lexer.Token {
	p.current = p.lexer.NextToken()
	return p.current
}

// currentToken returns the current token.
func (p *Parser) currentToken() lexer.Token {
	return p.current
}

// CurrentToken is the exported version of currentToken for external use.
func (p *Parser) CurrentToken() lexer.Token {
	return p.currentToken()
}

// enterScope pushes a new scope onto the stack.
func (p *Parser) enterScope(scope Scope) {
	p.scope.Push(scope)
}

// EnterScope is the exported version of enterScope for external use.
func (p *Parser) EnterScope(scope Scope) {
	p.enterScope(scope)
}

// exitScope pops the current scope from the stack.
func (p *Parser) exitScope() Scope {
	return p.scope.Pop()
}

// ExitScope is the exported version of exitScope for external use.
func (p *Parser) ExitScope() Scope {
	return p.exitScope()
}

// currentScope returns the current parsing scope.
func (p *Parser) currentScope() Scope {
	return p.scope.Current()
}

// CurrentScope is the exported version of currentScope for external use.
func (p *Parser) CurrentScope() Scope {
	return p.currentScope()
}

// validateDirectiveScope checks if a directive token is valid at the current scope.
// Returns a ParseError if invalid, nil if valid.
func (p *Parser) validateDirectiveScope(tok lexer.Token) *ParseError {
	if !tok.Type.IsDotKeyword() {
		return nil // Not a directive
	}

	if IsDirectiveValidAtScope(tok.Type, p.currentScope()) {
		return nil // Valid at current scope
	}

	return NewScopeError(tok.Type, p.currentScope(), tok.Location)
}

// currentIndentLevel returns the expected indentation level for the current scope.
// Level 0 = global, Level 1 = environment/recipe, Level 2 = block.
func (p *Parser) currentIndentLevel() int {
	switch p.currentScope() {
	case ScopeGlobal:
		return 0
	case ScopeEnvironment, ScopeRecipe:
		return 1
	case ScopeBlock:
		return 2
	default:
		return 0
	}
}

// CurrentIndentLevel is the exported version of currentIndentLevel for external use.
func (p *Parser) CurrentIndentLevel() int {
	return p.currentIndentLevel()
}

// addError adds a parse error to the error collection.
func (p *Parser) addError(err *ParseError) {
	p.errors.Add(err)
}

// Errors returns the collected parse errors.
func (p *Parser) Errors() *ParseErrors {
	return p.errors
}

// HasErrors returns true if any parse errors were encountered.
func (p *Parser) HasErrors() bool {
	return p.errors.HasErrors()
}
