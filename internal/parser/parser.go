package parser

import (
	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/lexer"
)

// Parser transforms a token stream into an AST.
type Parser struct {
	lexer   *lexer.Lexer
	current lexer.Token
	scope   *ScopeStack
	errors  *ParseErrors

	// includeVars tracks the literal values of immediate variables seen so
	// far in this parse, in file order, so that .include: paths can
	// interpolate them. It is shared (same pointer) across an entire
	// include chain - see parseIncludedFile - so an included file sees the
	// variables defined earlier in the file that included it.
	includeVars *includeVarTracker
}

// New creates a new Parser for the given lexer.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:       l,
		scope:       NewScopeStack(),
		errors:      &ParseErrors{},
		includeVars: newIncludeVarTracker(),
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

// CurrentToken returns the current token.
func (p *Parser) CurrentToken() lexer.Token {
	return p.current
}

// EnterScope pushes a new scope onto the stack.
func (p *Parser) EnterScope(scope Scope) {
	p.scope.Push(scope)
}

// ExitScope pops the current scope from the stack.
func (p *Parser) ExitScope() Scope {
	return p.scope.Pop()
}

// CurrentScope returns the current parsing scope.
func (p *Parser) CurrentScope() Scope {
	return p.scope.Current()
}

// validateDirectiveScope checks if a directive token is valid at the current scope.
// Returns a ParseError if invalid, nil if valid.
func (p *Parser) validateDirectiveScope(tok lexer.Token) *ParseError {
	if !tok.Type.IsDotKeyword() {
		return nil // Not a directive
	}

	if IsDirectiveValidAtScope(tok.Type, p.CurrentScope()) {
		return nil // Valid at current scope
	}

	return NewScopeError(tok.Type, p.CurrentScope(), tok.Location)
}

// CurrentIndentLevel returns the expected indentation level for the current scope.
// Level 0 = global, Level 1 = environment/recipe, Level 2 = block.
func (p *Parser) CurrentIndentLevel() int {
	switch p.CurrentScope() {
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

// expectColon consumes the current token (directive keyword) and expects a colon.
// Returns an error if the colon is missing.
// This is a helper for the common directive parsing pattern:
//
//	p.nextToken()  // consume directive
//	if p.current.Type != lexer.COLON { return error }
//	p.nextToken()  // consume colon
func (p *Parser) expectColon(directiveName string) *ParseError {
	p.nextToken() // consume directive keyword

	if p.current.Type != lexer.COLON {
		return &ParseError{
			Message:  "expected ':' after " + directiveName,
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume :

	return nil
}

// consumeNewline advances past a newline if present.
func (p *Parser) consumeNewline() {
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}
}

// parseComment creates a Comment AST node from the current COMMENT token.
// Returns nil if current token is not a comment.
func (p *Parser) parseComment() *ast.Comment {
	if p.current.Type != lexer.COMMENT {
		return nil
	}
	comment := &ast.Comment{
		Text:     p.current.Literal,
		Location: ast.SourceLocationFromToken(p.current),
	}
	p.nextToken()
	return comment
}

// maxErrors is the maximum number of errors to collect before giving up.
const maxErrors = 10

// ParseNeedfile parses a complete needfile with error recovery.
// It collects multiple errors and attempts to continue parsing after each error.
// Returns a slice of successfully parsed statements and all collected errors.
func (p *Parser) ParseNeedfile() ([]ast.Statement, *ParseErrors) {
	var statements []ast.Statement

	for p.current.Type != lexer.EOF {
		// Stop if we've collected too many errors
		if len(p.errors.Errors) >= maxErrors {
			break
		}

		stmts, err := p.parseTopLevelStatements()
		if err != nil {
			p.addError(err)
			p.recoverToLevel0()
			continue
		}

		for _, stmt := range stmts {
			if v, ok := stmt.(*ast.Variable); ok {
				p.includeVars.record(v)
			}
		}

		statements = append(statements, stmts...)
	}

	return statements, p.errors
}

// parseTopLevelStatements parses one or more top-level statements.
// Returns multiple statements when parsing includes (included statements are merged).
func (p *Parser) parseTopLevelStatements() ([]ast.Statement, *ParseError) {
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

	// Lexer-level error (e.g. mixed indentation, unclosed interpolation) -
	// surface it as a parse error instead of falling through to the generic
	// "unexpected token" message.
	if p.current.Type == lexer.ERROR {
		return nil, p.errorFromLexerToken(p.current)
	}

	// Handle comments
	if p.current.Type == lexer.COMMENT {
		return []ast.Statement{p.parseComment()}, nil
	}

	// Handle directives
	if p.current.Type.IsDotKeyword() {
		return p.parseTopLevelDirectiveWithIncludes()
	}

	// Handle conditionals
	if p.IsConditionalLine() {
		stmt, err := p.ParseConditional()
		if err != nil {
			return nil, err
		}
		return []ast.Statement{stmt}, nil
	}

	// Handle lazy variables
	if p.current.Type == lexer.LAZY {
		stmt, err := p.ParseVariable()
		if err != nil {
			return nil, err
		}
		return []ast.Statement{stmt}, nil
	}

	// Handle variables or targets
	if p.current.Type == lexer.IDENTIFIER {
		// Check if this is a variable (= before :)
		if p.looksLikeVariableLine() {
			stmt, err := p.ParseVariable()
			if err != nil {
				return nil, err
			}
			return []ast.Statement{stmt}, nil
		}
		// Otherwise it could be a file target (e.g., "app:" or "build/app:")
		stmt, err := p.ParseTarget()
		if err != nil {
			return nil, err
		}
		return []ast.Statement{stmt}, nil
	}

	// Handle phony targets
	if p.current.Type == lexer.AT_IDENTIFIER {
		stmt, err := p.ParseTarget()
		if err != nil {
			return nil, err
		}
		return []ast.Statement{stmt}, nil
	}

	// Handle file targets
	if p.current.Type == lexer.PATH {
		stmt, err := p.ParseTarget()
		if err != nil {
			return nil, err
		}
		return []ast.Statement{stmt}, nil
	}

	// Handle targets starting with interpolation (e.g., {build_dir}/app:)
	if p.current.Type == lexer.INTERP_START {
		stmt, err := p.ParseTarget()
		if err != nil {
			return nil, err
		}
		return []ast.Statement{stmt}, nil
	}

	// Unrecognized token - generate error
	return nil, &ParseError{
		Message:  "unexpected token: " + p.current.Type.String(),
		Location: p.current.Location,
		Hint:     "expected variable definition, target, directive, or conditional",
	}
}

// parseTopLevelDirectiveWithIncludes handles directives at global scope with scope validation.
// For includes, it returns the included statements along with the directive.
func (p *Parser) parseTopLevelDirectiveWithIncludes() ([]ast.Statement, *ParseError) {
	// Validate scope first
	if err := p.validateDirectiveScope(p.current); err != nil {
		return nil, err
	}

	switch p.current.Type {
	case lexer.DOT_SHELL, lexer.DOT_PARALLEL, lexer.DOT_DEFAULT:
		stmt, err := p.parseGlobalDirective()
		if err != nil {
			return nil, err
		}
		return []ast.Statement{stmt}, nil
	case lexer.DOT_INCLUDE:
		directive, includedStmts, err := p.ParseInclude()
		if err != nil {
			return nil, err
		}
		// Return included statements first, then the directive marker
		// This ensures variables/targets from included files are visible
		// before any subsequent content in the including file
		result := make([]ast.Statement, 0, len(includedStmts)+1)
		result = append(result, includedStmts...)
		result = append(result, directive)
		return result, nil
	case lexer.DOT_ENVIRONMENT:
		stmt, err := p.ParseEnvironment()
		if err != nil {
			return nil, err
		}
		return []ast.Statement{stmt}, nil
	default:
		// Other directives invalid at global scope
		return nil, NewScopeError(p.current.Type, p.CurrentScope(), p.current.Location)
	}
}

// recoverToLevel0 advances the parser until it reaches a line at indentation level 0.
// This is used for error recovery to skip past erroneous content.
func (p *Parser) recoverToLevel0() {
	// Track if we started inside a conditional (saw else/elif/end at level 0)
	nestLevel := 0

	// Skip the current line
	for p.current.Type != lexer.NEWLINE && p.current.Type != lexer.EOF {
		// Track conditional nesting to find matching end
		if p.current.Type == lexer.IF || p.current.Type == lexer.IFDEF || p.current.Type == lexer.IFNDEF {
			nestLevel++
		}
		p.nextToken()
	}

	// Skip past the newline
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	// Continue skipping indented lines and orphaned conditional keywords
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

		// Skip orphaned conditional keywords (else/elif/end) that belong to the
		// broken conditional we're recovering from
		if p.current.Type == lexer.ELSE || p.current.Type == lexer.ELIF {
			// Skip the line with else/elif
			for p.current.Type != lexer.NEWLINE && p.current.Type != lexer.EOF {
				p.nextToken()
			}
			if p.current.Type == lexer.NEWLINE {
				p.nextToken()
			}
			continue
		}

		if p.current.Type == lexer.END {
			if nestLevel > 0 {
				nestLevel--
			}
			// Skip the end keyword and its line
			for p.current.Type != lexer.NEWLINE && p.current.Type != lexer.EOF {
				p.nextToken()
			}
			if p.current.Type == lexer.NEWLINE {
				p.nextToken()
			}
			// If we're back to level 0, we've recovered past the broken conditional
			if nestLevel == 0 {
				break
			}
			continue
		}

		// Reached a non-indented, non-conditional line, stop recovering
		break
	}
}

// looksLikeVariableLine checks if the current position looks like a variable definition.
// Returns true if `=` appears before `:` (or there's no `:` at all).
// Uses the lexer's peek function to determine this without consuming tokens.
func (p *Parser) looksLikeVariableLine() bool {
	return p.lexer.PeekIsVariableLine()
}
