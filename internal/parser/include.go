package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/lexer"
)

// automaticVarNames are the automatic variable names (DESIGN.md Section
// 3.3.4) that are only ever bound at recipe-execution time. They can never
// be resolved while parsing, so referencing one in a .include: path is
// always an error.
var automaticVarNames = map[string]bool{
	"target":      true,
	"deps":        true,
	"in":          true,
	"out":         true,
	"stem":        true,
	"target.dir":  true,
	"target.file": true,
}

// includeVarTracker records the literal values of immediate (non-lazy)
// variables as they are parsed, in file order. It lets .include: directives
// interpolate variables defined earlier in the same parse, without needing
// the full evaluator (which does not exist yet at parse time).
//
// A single tracker instance is shared across an entire include chain (see
// parseIncludedFile), so an included file sees variables defined earlier in
// whichever file included it, and its own .include: directives can in turn
// reference them.
type includeVarTracker struct {
	literal map[string]string // name -> fully-resolved literal value
	lazy    map[string]bool   // name -> declared with "lazy"
	dynamic map[string]bool   // name -> immediate, but value isn't a literal
}

func newIncludeVarTracker() *includeVarTracker {
	return &includeVarTracker{
		literal: make(map[string]string),
		lazy:    make(map[string]bool),
		dynamic: make(map[string]bool),
	}
}

// record updates the tracker with a variable statement as it is parsed.
// Redefinition simply overwrites the earlier record, matching normal
// shadowing: only the most recent definition at this point in the file is
// visible to a later .include:.
func (t *includeVarTracker) record(v *ast.Variable) {
	delete(t.literal, v.Name)
	delete(t.lazy, v.Name)
	delete(t.dynamic, v.Name)

	if v.Lazy {
		t.lazy[v.Name] = true
		return
	}

	resolved, ok := t.resolveLiteral(v.Value)
	if !ok {
		t.dynamic[v.Name] = true
		return
	}
	t.literal[v.Name] = resolved
}

// resolveLiteral attempts to fully resolve a Value using only variables
// already known to be literal. It returns ("", false) if the value contains
// a function call or references a variable that isn't a known literal
// (including one not yet defined, or defined non-literally).
func (t *includeVarTracker) resolveLiteral(v *ast.Value) (string, bool) {
	if v == nil {
		return "", true
	}
	var b strings.Builder
	for _, part := range v.Parts {
		switch p := part.(type) {
		case *ast.LiteralValue:
			b.WriteString(p.Text)
		case *ast.Interpolation:
			val, ok := t.literal[p.Name]
			if !ok {
				return "", false
			}
			b.WriteString(val)
		default:
			// FunctionCall (or any future part kind) is never a literal.
			return "", false
		}
	}
	return b.String(), true
}

// resolveIncludePath resolves a .include:-style path value against the
// variables known so far, per the rules above. On success it returns the
// trimmed resolved path. On failure it returns a *ParseError of the form
// "<directive> cannot resolve '<name>': <reason>", located at the offending
// interpolation (or function call).
func (t *includeVarTracker) resolveIncludePath(directive string, v *ast.Value, fallbackLoc lexer.SourceLocation) (string, *ParseError) {
	if v == nil || len(v.Parts) == 0 {
		return "", &ParseError{Message: directive + " requires a path", Location: fallbackLoc}
	}

	var b strings.Builder
	for _, part := range v.Parts {
		switch p := part.(type) {
		case *ast.LiteralValue:
			b.WriteString(p.Text)

		case *ast.Interpolation:
			if val, ok := t.literal[p.Name]; ok {
				b.WriteString(val)
				continue
			}

			reason := "undefined variable"
			switch {
			case automaticVarNames[p.Name]:
				reason = "automatic variable"
			case t.lazy[p.Name]:
				reason = "lazy variable"
			case t.dynamic[p.Name]:
				reason = "value is not a literal"
			}
			return "", &ParseError{
				Message:  fmt.Sprintf("%s cannot resolve '%s': %s", directive, p.Name, reason),
				Location: astLocToLexerLoc(p.Location),
			}

		case *ast.FunctionCall:
			return "", &ParseError{
				Message:  fmt.Sprintf("%s cannot resolve '%s()': function calls are not supported here", directive, p.Name.String()),
				Location: astLocToLexerLoc(p.Location),
			}

		default:
			return "", &ParseError{
				Message:  directive + ": unsupported value part",
				Location: fallbackLoc,
			}
		}
	}

	return strings.TrimSpace(b.String()), nil
}

// astLocToLexerLoc converts an ast.SourceLocation to a lexer.SourceLocation
// (identical fields, distinct types across package boundaries).
func astLocToLexerLoc(loc ast.SourceLocation) lexer.SourceLocation {
	return lexer.SourceLocation{File: loc.File, Line: loc.Line, Column: loc.Column}
}

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

	// Resolve the path, interpolating any variables defined earlier in this
	// parse (see includeVarTracker).
	includePath, resolveErr := p.includeVars.resolveIncludePath(".include:", value, loc)
	if resolveErr != nil {
		return nil, nil, resolveErr
	}
	if includePath == "" {
		return nil, nil, &ParseError{
			Message:  ".include: path must not be empty",
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
	// Share this parser's include-var tracker so the included file sees
	// variables defined earlier in whichever file included it, and any
	// nested .include: inside it sees the same accumulated set.
	includedParser.includeVars = p.includeVars

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
			if v, ok := stmt.(*ast.Variable); ok {
				includedParser.includeVars.record(v)
			}
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
