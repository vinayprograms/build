package parser

import (
	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/lexer"
)

// parseVariable parses a variable definition.
// Grammar: variable_def = [ "lazy" ] identifier "=" value NEWLINE ;
func (p *Parser) parseVariable(isLazy bool) *ast.Variable {
	// Placeholder - will be implemented after tests pass
	return nil
}

// parseValue parses a value (string with interpolations and function calls).
// Grammar: value = { value_part } ;
// Grammar: value_part = STRING | interpolation | function_call ;
func (p *Parser) parseValue() *ast.Value {
	// Placeholder - will be implemented after tests pass
	return nil
}

// isVariableLine checks if the current position starts a variable definition.
// A line is a variable if `=` appears before `:`.
func (p *Parser) isVariableLine() bool {
	// Placeholder - will be implemented after tests pass
	return false
}

// ParseVariable is the exported method for parsing a variable.
// It is called when the parser encounters a line that looks like a variable definition.
func (p *Parser) ParseVariable() (*ast.Variable, *ParseError) {
	// Check if we're at a lazy keyword
	isLazy := false
	if p.current.Type == lexer.LAZY {
		isLazy = true
		p.nextToken()
	}

	// Expect identifier
	if p.current.Type != lexer.IDENTIFIER {
		return nil, &ParseError{
			Message:  "expected identifier in variable definition",
			Location: p.current.Location,
		}
	}

	name := p.current.Literal
	loc := ast.SourceLocationFromToken(p.current)
	p.nextToken()

	// Expect equals
	if p.current.Type != lexer.EQUALS {
		return nil, &ParseError{
			Message:  "expected '=' in variable definition",
			Location: p.current.Location,
		}
	}
	p.nextToken()

	// Parse value
	value := p.ParseValue()

	return &ast.Variable{
		Name:     name,
		Value:    value,
		Lazy:     isLazy,
		Location: loc,
	}, nil
}

// ParseValue is the exported method for parsing a value.
func (p *Parser) ParseValue() *ast.Value {
	loc := ast.SourceLocationFromToken(p.current)
	var parts []ast.ValuePart

	// Consume tokens until newline, comment, or EOF
	for p.current.Type != lexer.NEWLINE &&
		p.current.Type != lexer.COMMENT &&
		p.current.Type != lexer.EOF {

		switch p.current.Type {
		case lexer.STRING:
			parts = append(parts, &ast.LiteralValue{Text: p.current.Literal})
			p.nextToken()

		case lexer.INTERP_START:
			interp := p.parseInterpolation()
			if interp != nil {
				parts = append(parts, interp)
			}

		case lexer.FUNC_SHELL, lexer.FUNC_GLOB, lexer.FUNC_BASENAME, lexer.FUNC_DIRNAME, lexer.FUNC_REPLACE:
			funcCall := p.parseFunctionCall()
			if funcCall != nil {
				parts = append(parts, funcCall)
			}

		case lexer.ESCAPE_LBRACE:
			parts = append(parts, &ast.LiteralValue{Text: "{"})
			p.nextToken()

		case lexer.ESCAPE_RBRACE:
			parts = append(parts, &ast.LiteralValue{Text: "}"})
			p.nextToken()

		case lexer.IDENTIFIER:
			// Check if this is a function name followed by (
			tok := LookupKeyword(p.current.Literal)
			if tok.IsFunction() {
				funcCall := p.parseFunctionCall()
				if funcCall != nil {
					parts = append(parts, funcCall)
				}
			} else {
				// Treat as literal text
				parts = append(parts, &ast.LiteralValue{Text: p.current.Literal})
				p.nextToken()
			}

		default:
			// Treat other tokens as literal text
			parts = append(parts, &ast.LiteralValue{Text: p.current.Literal})
			p.nextToken()
		}
	}

	return &ast.Value{
		Parts:    parts,
		Location: loc,
	}
}

// parseInterpolation parses a variable interpolation {name} or {name:raw}.
func (p *Parser) parseInterpolation() *ast.Interpolation {
	if p.current.Type != lexer.INTERP_START {
		return nil
	}
	loc := ast.SourceLocationFromToken(p.current)
	p.nextToken() // consume {

	// Expect identifier
	if p.current.Type != lexer.IDENTIFIER {
		p.addError(&ParseError{
			Message:  "expected identifier in interpolation",
			Location: p.current.Location,
		})
		return nil
	}

	name := p.current.Literal
	p.nextToken()

	// Check for :raw modifier
	raw := false
	if p.current.Type == lexer.INTERP_MOD {
		if p.current.Literal == ":raw" {
			raw = true
		}
		p.nextToken()
	}

	// Expect close brace
	if p.current.Type != lexer.INTERP_END {
		p.addError(&ParseError{
			Message:  "expected '}' to close interpolation",
			Location: p.current.Location,
		})
		return nil
	}
	p.nextToken() // consume }

	return &ast.Interpolation{
		Name:     name,
		Raw:      raw,
		Location: loc,
	}
}

// parseFunctionCall parses a function call like shell(...), glob(...), etc.
func (p *Parser) parseFunctionCall() *ast.FunctionCall {
	loc := ast.SourceLocationFromToken(p.current)

	// Determine function name
	var funcName ast.FunctionName
	switch p.current.Type {
	case lexer.FUNC_SHELL:
		funcName = ast.FuncShell
	case lexer.FUNC_GLOB:
		funcName = ast.FuncGlob
	case lexer.FUNC_BASENAME:
		funcName = ast.FuncBasename
	case lexer.FUNC_DIRNAME:
		funcName = ast.FuncDirname
	case lexer.FUNC_REPLACE:
		funcName = ast.FuncReplace
	default:
		// Check if identifier is a function name
		if p.current.Type == lexer.IDENTIFIER {
			switch p.current.Literal {
			case "shell":
				funcName = ast.FuncShell
			case "glob":
				funcName = ast.FuncGlob
			case "basename":
				funcName = ast.FuncBasename
			case "dirname":
				funcName = ast.FuncDirname
			case "replace":
				funcName = ast.FuncReplace
			default:
				return nil
			}
		} else {
			return nil
		}
	}
	p.nextToken() // consume function name

	// Expect open paren
	if p.current.Type != lexer.LPAREN {
		p.addError(&ParseError{
			Message:  "expected '(' after function name",
			Location: p.current.Location,
		})
		return nil
	}
	p.nextToken() // consume (

	// Parse comma-separated arguments
	var args []*ast.Value
	for p.current.Type != lexer.RPAREN && p.current.Type != lexer.EOF && p.current.Type != lexer.NEWLINE {
		arg := p.parseFunctionArg()
		if arg != nil {
			args = append(args, arg)
		}

		// Check for comma between arguments
		if p.current.Type == lexer.COMMA {
			p.nextToken() // consume comma
		} else {
			break
		}
	}

	// Expect close paren
	if p.current.Type != lexer.RPAREN {
		p.addError(&ParseError{
			Message:  "expected ')' to close function call",
			Location: p.current.Location,
		})
		return nil
	}
	p.nextToken() // consume )

	return &ast.FunctionCall{
		Name:     funcName,
		Args:     args,
		Location: loc,
	}
}

// parseFunctionArg parses a single function argument (until comma, ), newline, or EOF).
func (p *Parser) parseFunctionArg() *ast.Value {
	loc := ast.SourceLocationFromToken(p.current)
	var parts []ast.ValuePart

	// Consume tokens until ), comma, newline, or EOF
	parenDepth := 0
	for {
		if p.current.Type == lexer.EOF || p.current.Type == lexer.NEWLINE {
			break
		}
		if p.current.Type == lexer.RPAREN && parenDepth == 0 {
			break
		}
		// Stop at comma when not inside nested parens
		if p.current.Type == lexer.COMMA && parenDepth == 0 {
			break
		}

		switch p.current.Type {
		case lexer.LPAREN:
			parenDepth++
			parts = append(parts, &ast.LiteralValue{Text: "("})
			p.nextToken()

		case lexer.RPAREN:
			parenDepth--
			parts = append(parts, &ast.LiteralValue{Text: ")"})
			p.nextToken()

		case lexer.STRING:
			parts = append(parts, &ast.LiteralValue{Text: p.current.Literal})
			p.nextToken()

		case lexer.INTERP_START:
			interp := p.parseInterpolation()
			if interp != nil {
				parts = append(parts, interp)
			}

		case lexer.ESCAPE_LBRACE:
			parts = append(parts, &ast.LiteralValue{Text: "{"})
			p.nextToken()

		case lexer.ESCAPE_RBRACE:
			parts = append(parts, &ast.LiteralValue{Text: "}"})
			p.nextToken()

		default:
			parts = append(parts, &ast.LiteralValue{Text: p.current.Literal})
			p.nextToken()
		}
	}

	return &ast.Value{
		Parts:    parts,
		Location: loc,
	}
}

// LookupKeyword is a helper that checks if a literal is a keyword.
// This is used to check for function names.
func LookupKeyword(literal string) lexer.TokenType {
	return lexer.LookupKeyword(literal)
}
