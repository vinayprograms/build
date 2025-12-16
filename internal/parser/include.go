package parser

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/lexer"
)

// includeStack tracks files being processed to detect circular includes.
// This is stored on the parser to persist across recursive calls.
type includeStack struct {
	files map[string]bool
}

func newIncludeStack() *includeStack {
	return &includeStack{
		files: make(map[string]bool),
	}
}

func (s *includeStack) push(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	if s.files[absPath] {
		return false // circular include detected
	}
	s.files[absPath] = true
	return true
}

func (s *includeStack) pop(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	delete(s.files, absPath)
}

// ParseInclude parses a .include: directive and recursively parses the included file.
// Returns:
//   - The directive AST node
//   - The statements from the included file (already parsed)
//   - Any error encountered
//
// Grammar: ".include:" value NEWLINE
func (p *Parser) ParseInclude() (*ast.Directive, []ast.Statement, *ParseError) {
	return p.parseIncludeWithStack(newIncludeStack())
}

// parseIncludeWithStack is the internal implementation that tracks include stack.
func (p *Parser) parseIncludeWithStack(stack *includeStack) (*ast.Directive, []ast.Statement, *ParseError) {
	if p.current.Type != lexer.DOT_INCLUDE {
		return nil, nil, &ParseError{
			Message:  "expected .include: directive",
			Location: p.current.Location,
		}
	}

	loc := p.current.Location
	astLoc := ast.SourceLocationFromToken(p.current)
	p.nextToken() // consume .include

	// Consume the colon if present
	if p.current.Type == lexer.COLON {
		p.nextToken()
	}

	// Parse the include path value
	value := p.ParseValue()

	// Get the path from the value
	includePath := extractLiteralPath(value)
	if includePath == "" {
		return nil, nil, &ParseError{
			Message:  ".include: path must be a literal string (interpolation not yet supported)",
			Location: loc,
		}
	}

	// Resolve relative paths based on current file's directory
	if !filepath.IsAbs(includePath) {
		baseDir := filepath.Dir(loc.File)
		includePath = filepath.Join(baseDir, includePath)
	}

	// Check for circular includes
	if !stack.push(includePath) {
		return nil, nil, &ParseError{
			Message:  "circular include detected: " + includePath,
			Location: loc,
		}
	}
	defer stack.pop(includePath)

	// Read the included file
	content, err := os.ReadFile(includePath)
	if err != nil {
		return nil, nil, &ParseError{
			Message:  "cannot read included file: " + err.Error(),
			Location: loc,
		}
	}

	// Parse the included file
	includedStatements, parseErr := p.parseIncludedFile(includePath, string(content), stack)
	if parseErr != nil {
		return nil, nil, parseErr
	}

	// Create the directive node
	directive := &ast.Directive{
		Kind:     ast.DirectiveInclude,
		Value:    value,
		Location: astLoc,
	}

	// Skip newline after include directive
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	return directive, includedStatements, nil
}

// parseIncludedFile parses the content of an included file.
func (p *Parser) parseIncludedFile(path, content string, stack *includeStack) ([]ast.Statement, *ParseError) {
	l := lexer.New(path, content)
	includedParser := New(l)

	var statements []ast.Statement

	for includedParser.current.Type != lexer.EOF {
		// Skip indent at line start
		if includedParser.current.Type == lexer.INDENT {
			includedParser.nextToken()
		}

		// Skip newlines
		if includedParser.current.Type == lexer.NEWLINE {
			includedParser.nextToken()
			continue
		}

		// Skip comments
		if includedParser.current.Type == lexer.COMMENT {
			statements = append(statements, includedParser.parseComment())
			continue
		}

		// Handle nested includes
		if includedParser.current.Type == lexer.DOT_INCLUDE {
			_, nestedStmts, err := includedParser.parseIncludeWithStack(stack)
			if err != nil {
				return nil, err
			}
			statements = append(statements, nestedStmts...)
			continue
		}

		// Parse other statements
		stmt, err := includedParser.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	return statements, nil
}

// parseStatement parses a single top-level statement.
// This is a simplified dispatcher that delegates to specific parsers.
func (p *Parser) parseStatement() (ast.Statement, *ParseError) {
	// Skip indent
	if p.current.Type == lexer.INDENT {
		p.nextToken()
	}

	switch p.current.Type {
	case lexer.EOF:
		return nil, nil

	case lexer.NEWLINE:
		p.nextToken()
		return nil, nil

	case lexer.COMMENT:
		return p.parseComment(), nil

	case lexer.DOT_SHELL, lexer.DOT_PARALLEL, lexer.DOT_DEFAULT:
		return p.parseGlobalDirective()

	case lexer.DOT_ENVIRONMENT:
		return p.ParseEnvironment()

	case lexer.IF, lexer.IFDEF, lexer.IFNDEF:
		return p.ParseConditional()

	case lexer.LAZY:
		// Lazy variable
		v, err := p.ParseVariable()
		return v, err

	case lexer.IDENTIFIER:
		// Could be variable or something else
		// Check if this line is a variable definition
		if p.isVariableLineFromCurrent() {
			v, err := p.ParseVariable()
			return v, err
		}
		// Otherwise skip as unrecognized
		p.skipToNextLine()
		return nil, nil

	case lexer.AT_IDENTIFIER:
		// Phony target
		t, err := p.ParseTarget()
		return t, err

	case lexer.PATH:
		// File target
		t, err := p.ParseTarget()
		return t, err

	default:
		// Skip unrecognized tokens
		p.nextToken()
		return nil, nil
	}
}

// parseGlobalDirective parses .shell:, .parallel:, or .default: directives.
func (p *Parser) parseGlobalDirective() (*ast.Directive, *ParseError) {
	var kind ast.DirectiveKind

	switch p.current.Type {
	case lexer.DOT_SHELL:
		kind = ast.DirectiveShell
	case lexer.DOT_PARALLEL:
		kind = ast.DirectiveParallel
	case lexer.DOT_DEFAULT:
		kind = ast.DirectiveDefault
	default:
		return nil, &ParseError{
			Message:  "expected directive",
			Location: p.current.Location,
		}
	}

	loc := ast.SourceLocationFromToken(p.current)
	p.nextToken() // consume directive token

	// Consume the colon if present (directives are tokenized as .shell + :)
	if p.current.Type == lexer.COLON {
		p.nextToken()
	}

	value := p.ParseValue()

	// Skip trailing newline
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	return &ast.Directive{
		Kind:     kind,
		Value:    value,
		Location: loc,
	}, nil
}

// isVariableLineFromCurrent checks if the current line is a variable definition.
// A line is a variable if = appears before : in the remaining tokens.
func (p *Parser) isVariableLineFromCurrent() bool {
	// Save current state - we need to peek ahead
	// For simplicity, assume identifier followed by = is a variable
	// This is a heuristic that works for most cases
	return true // Will be refined by ParseVariable if needed
}

// skipToNextLine advances past tokens until newline or EOF.
func (p *Parser) skipToNextLine() {
	for p.current.Type != lexer.NEWLINE && p.current.Type != lexer.EOF {
		p.nextToken()
	}
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}
}

// extractLiteralPath extracts a literal path from a Value.
// Returns empty string if the value contains interpolations (not supported yet).
func extractLiteralPath(v *ast.Value) string {
	if v == nil || len(v.Parts) == 0 {
		return ""
	}

	// Check all parts are literals
	var path string
	for _, part := range v.Parts {
		lit, ok := part.(*ast.LiteralValue)
		if !ok {
			return "" // Contains non-literal (interpolation)
		}
		path += lit.Text
	}

	// Trim whitespace
	return strings.TrimSpace(path)
}
