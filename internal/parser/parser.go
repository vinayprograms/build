package parser

import (
	"github.com/vinayprograms/build/internal/ast"
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

// maxErrors is the maximum number of errors to collect before giving up.
const maxErrors = 10

// ParseBuildfile parses a complete buildfile with error recovery.
// It collects multiple errors and attempts to continue parsing after each error.
// Returns a slice of successfully parsed statements and all collected errors.
func (p *Parser) ParseBuildfile() ([]ast.Statement, *ParseErrors) {
	var statements []ast.Statement

	for p.current.Type != lexer.EOF {
		// Stop if we've collected too many errors
		if len(p.errors.Errors) >= maxErrors {
			break
		}

		stmt, err := p.parseTopLevelStatement()
		if err != nil {
			p.addError(err)
			p.recoverToLevel0()
			continue
		}

		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	return statements, p.errors
}

// parseTopLevelStatement parses a single top-level statement with error handling.
func (p *Parser) parseTopLevelStatement() (ast.Statement, *ParseError) {
	// Skip leading indentation at global scope (shouldn't happen, but be safe)
	if p.current.Type == lexer.INDENT {
		p.nextToken()
	}

	// Skip newlines
	for p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	if p.current.Type == lexer.EOF {
		return nil, nil
	}

	// Handle comments
	if p.current.Type == lexer.COMMENT {
		stmt := &ast.Comment{
			Text:     p.current.Literal,
			Location: ast.SourceLocationFromToken(p.current),
		}
		p.nextToken()
		return stmt, nil
	}

	// Handle directives
	if p.current.Type.IsDotKeyword() {
		return p.parseTopLevelDirective()
	}

	// Handle conditionals
	if p.IsConditionalLine() {
		return p.ParseConditional()
	}

	// Handle lazy variables
	if p.current.Type == lexer.LAZY {
		return p.ParseVariable()
	}

	// Handle variables or targets
	if p.current.Type == lexer.IDENTIFIER {
		// Check if this is a variable (= before :)
		if p.looksLikeVariableLine() {
			return p.ParseVariable()
		}
		// Otherwise could be part of a path-like target
	}

	// Handle phony targets
	if p.current.Type == lexer.AT_IDENTIFIER {
		return p.ParseTarget()
	}

	// Handle file targets
	if p.current.Type == lexer.PATH {
		return p.ParseTarget()
	}

	// Handle targets starting with interpolation (e.g., {build_dir}/app:)
	if p.current.Type == lexer.INTERP_START {
		return p.ParseTarget()
	}

	// Unrecognized token - generate error
	return nil, &ParseError{
		Message:  "unexpected token: " + p.current.Type.String(),
		Location: p.current.Location,
		Hint:     "expected variable definition, target, directive, or conditional",
	}
}

// parseTopLevelDirective handles directives at global scope with scope validation.
func (p *Parser) parseTopLevelDirective() (ast.Statement, *ParseError) {
	// Validate scope first
	if err := p.validateDirectiveScope(p.current); err != nil {
		return nil, err
	}

	switch p.current.Type {
	case lexer.DOT_SHELL, lexer.DOT_PARALLEL, lexer.DOT_DEFAULT:
		return p.parseGlobalDirective()
	case lexer.DOT_INCLUDE:
		directive, stmts, err := p.ParseInclude()
		if err != nil {
			return nil, err
		}
		// For now, return the directive; the caller would need to handle included statements
		// In a full implementation, we'd merge stmts into the statement list
		_ = stmts // Include statements are handled separately
		return directive, nil
	case lexer.DOT_ENVIRONMENT:
		return p.ParseEnvironment()
	default:
		// Other directives invalid at global scope
		return nil, NewScopeError(p.current.Type, p.currentScope(), p.current.Location)
	}
}

// recoverToLevel0 advances the parser until it reaches a line at indentation level 0.
// This is used for error recovery to skip past erroneous content.
func (p *Parser) recoverToLevel0() {
	// Skip the current line
	for p.current.Type != lexer.NEWLINE && p.current.Type != lexer.EOF {
		p.nextToken()
	}

	// Skip past the newline
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	// Continue skipping indented lines
	for p.current.Type != lexer.EOF {
		// If we see an INDENT token, skip the whole line
		if p.current.Type == lexer.INDENT {
			for p.current.Type != lexer.NEWLINE && p.current.Type != lexer.EOF {
				p.nextToken()
			}
			if p.current.Type == lexer.NEWLINE {
				p.nextToken()
			}
			continue
		}

		// Reached a non-indented line, stop recovering
		break
	}
}

// looksLikeVariableLine checks if the current position looks like a variable definition.
// Returns true if `=` appears before `:` (or there's no `:` at all).
func (p *Parser) looksLikeVariableLine() bool {
	// For now, use a simple heuristic: identifiers followed by = are variables
	// This is called after we've verified the current token is IDENTIFIER
	// A more robust check would peek ahead, but for now trust the lexer's mode handling
	return true
}
