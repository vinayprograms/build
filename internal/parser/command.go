package parser

import (
	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/lexer"
)

// parseCommandLine parses a single command line.
// Grammar: command = { command_part } ;
// Grammar: command_part = STRING | interpolation ;
func (p *Parser) parseCommandLine() (*ast.LineCommand, *ParseError) {
	loc := ast.SourceLocationFromToken(p.current)
	return p.parseCommandLineContinuation(loc, nil)
}

// parseCommandLineContinuation parses the remainder of a command line into
// parts, given a location and any parts already collected (e.g. a token that
// had to be consumed to look ahead before recognizing this as a command
// line). p.current must be positioned right after those initial parts.
func (p *Parser) parseCommandLineContinuation(loc ast.SourceLocation, parts []ast.CommandPart) (*ast.LineCommand, *ParseError) {
	// Parse command parts until newline
	for p.current.Type != lexer.NEWLINE && p.current.Type != lexer.EOF && p.current.Type != lexer.COMMENT {
		switch p.current.Type {
		case lexer.STRING:
			parts = append(parts, &ast.LiteralCommand{Text: p.current.Literal})
			p.nextToken()

		case lexer.INTERP_START:
			interp, err := p.parseCommandInterpolation()
			if err != nil {
				return nil, err
			}
			if interp != nil {
				parts = append(parts, interp)
			}

		case lexer.ESCAPE_LBRACE:
			parts = append(parts, &ast.LiteralCommand{Text: "{"})
			p.nextToken()

		case lexer.ESCAPE_RBRACE:
			parts = append(parts, &ast.LiteralCommand{Text: "}"})
			p.nextToken()

		case lexer.PATH, lexer.IDENTIFIER:
			parts = append(parts, &ast.LiteralCommand{Text: p.current.Literal})
			p.nextToken()

		case lexer.ERROR:
			return nil, p.errorFromLexerToken(p.current)

		default:
			// Other tokens as literal text
			parts = append(parts, &ast.LiteralCommand{Text: p.current.Literal})
			p.nextToken()
		}
	}

	// Consume trailing comment (if present)
	if p.current.Type == lexer.COMMENT {
		p.nextToken()
	}

	// Consume newline
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	if len(parts) == 0 {
		return nil, nil
	}

	return &ast.LineCommand{
		Parts:    parts,
		Location: loc,
	}, nil
}

// parseCommandInterpolation parses an interpolation in a command ({var} or {var:raw}).
func (p *Parser) parseCommandInterpolation() (*ast.CommandInterpolation, *ParseError) {
	if p.current.Type != lexer.INTERP_START {
		return nil, nil
	}
	loc := ast.SourceLocationFromToken(p.current)
	p.nextToken() // consume {

	// Expect identifier
	if p.current.Type != lexer.IDENTIFIER {
		return nil, &ParseError{
			Message:  "expected identifier in interpolation",
			Location: p.current.Location,
		}
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
		return nil, &ParseError{
			Message:  "expected '}' to close interpolation",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume }

	return &ast.CommandInterpolation{
		Name:     name,
		Raw:      raw,
		Location: loc,
	}, nil
}

// parseBlockCommand parses a block: statement with nested lines.
// Grammar: block_stmt = "block:" NEWLINE INDENT { raw_line } DEDENT ;
func (p *Parser) parseBlockCommand() (*ast.BlockCommand, *ParseError) {
	loc := ast.SourceLocationFromToken(p.current)

	if p.current.Type != lexer.BLOCK {
		return nil, &ParseError{
			Message:  "expected 'block:' keyword",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume block

	// May have a colon after block (block: or block)
	if p.current.Type == lexer.COLON {
		p.nextToken() // consume :
	}

	// Consume newline
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	// Enter block scope
	p.EnterScope(ScopeBlock)
	defer p.ExitScope()

	var lines [][]ast.CommandPart
	baseIndent := -1 // Will be set from first block line

	// Parse block lines until dedent
	for {
		if p.current.Type == lexer.EOF {
			break
		}

		if p.current.Type == lexer.ERROR {
			return nil, p.errorFromLexerToken(p.current)
		}

		// A blank line with no leading whitespace produces a bare NEWLINE (no
		// INDENT token). Blank lines never terminate a block, so record them
		// as an empty line (consistent with the INDENT+NEWLINE case below)
		// and keep scanning.
		if p.current.Type == lexer.NEWLINE {
			lines = append(lines, []ast.CommandPart{})
			p.nextToken()
			continue
		}

		// Check for indent
		if p.current.Type != lexer.INDENT {
			break
		}

		// Check indent level - block content should be at level 2
		indentStr := p.current.Literal
		indentLevel := p.calculateIndentLevel(indentStr)
		if indentLevel < 2 {
			// Back to recipe level - end of block
			break
		}

		// Set base indent from first line
		if baseIndent < 0 {
			baseIndent = len(indentStr)
		}

		// Calculate relative indent (extra spaces beyond base)
		relativeIndent := ""
		if len(indentStr) > baseIndent {
			relativeIndent = indentStr[baseIndent:]
		}

		// Block content is always command lines - switch to command mode
		p.lexer.SetCommandMode()
		p.nextToken() // consume INDENT

		// Empty line in block
		if p.current.Type == lexer.NEWLINE {
			// Preserve empty lines in block output
			lines = append(lines, []ast.CommandPart{})
			p.nextToken()
			continue
		}

		// Comment line in block
		if p.current.Type == lexer.COMMENT {
			p.nextToken()
			if p.current.Type == lexer.NEWLINE {
				p.nextToken()
			}
			continue
		}

		// Parse block line with relative indent prepended
		lineParts, err := p.parseBlockLineWithIndent(relativeIndent)
		if err != nil {
			return nil, err
		}
		if len(lineParts) > 0 {
			lines = append(lines, lineParts)
		}
	}

	return &ast.BlockCommand{
		Lines:    lines,
		Location: loc,
	}, nil
}

// parseBlockLineWithIndent parses a single line in a block, prepending relative indent.
func (p *Parser) parseBlockLineWithIndent(indent string) ([]ast.CommandPart, *ParseError) {
	var parts []ast.CommandPart

	// Prepend relative indent if any
	if indent != "" {
		parts = append(parts, &ast.LiteralCommand{Text: indent})
	}

	// Parse parts until newline
	for p.current.Type != lexer.NEWLINE && p.current.Type != lexer.EOF && p.current.Type != lexer.COMMENT {
		switch p.current.Type {
		case lexer.STRING:
			parts = append(parts, &ast.LiteralCommand{Text: p.current.Literal})
			p.nextToken()

		case lexer.INTERP_START:
			interp, err := p.parseCommandInterpolation()
			if err != nil {
				return nil, err
			}
			if interp != nil {
				parts = append(parts, interp)
			}

		case lexer.ESCAPE_LBRACE:
			parts = append(parts, &ast.LiteralCommand{Text: "{"})
			p.nextToken()

		case lexer.ESCAPE_RBRACE:
			parts = append(parts, &ast.LiteralCommand{Text: "}"})
			p.nextToken()

		case lexer.ERROR:
			return nil, p.errorFromLexerToken(p.current)

		default:
			parts = append(parts, &ast.LiteralCommand{Text: p.current.Literal})
			p.nextToken()
		}
	}

	// Consume trailing comment (if present)
	if p.current.Type == lexer.COMMENT {
		p.nextToken()
	}

	// Consume newline
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	return parts, nil
}

// skipToNewline advances past the current line.
func (p *Parser) skipToNewline() {
	for p.current.Type != lexer.NEWLINE && p.current.Type != lexer.EOF {
		p.nextToken()
	}
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}
}
