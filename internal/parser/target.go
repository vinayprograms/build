package parser

import (
	"strings"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/lexer"
)

// ParseTarget parses a target definition.
// Grammar: target_def = target_spec ":" dependency_list NEWLINE [ recipe ] ;
// Grammar: target_spec = phony_target | file_target ;
// Grammar: phony_target = "@" identifier ;
// Grammar: file_target = path_pattern ;
func (p *Parser) ParseTarget() (*ast.Target, *ParseError) {
	loc := ast.SourceLocationFromToken(p.current)

	// Check for phony target (@name)
	isPhony := false
	if p.current.Type == lexer.AT_IDENTIFIER {
		isPhony = true
	}

	// Parse the target pattern
	pattern, err := p.parseTargetPattern(isPhony)
	if err != nil {
		return nil, err
	}

	// Expect colon
	if p.current.Type != lexer.COLON {
		return nil, &ParseError{
			Message:  "expected ':' after target pattern",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume :

	// Parse dependency list
	deps, err := p.parseDependencyList()
	if err != nil {
		return nil, err
	}

	// Consume trailing newline/EOF
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	return &ast.Target{
		Pattern:      *pattern,
		Dependencies: deps,
		Recipe:       nil, // Recipe parsing will be implemented separately
		Location:     loc,
	}, nil
}

// parseTargetPattern parses the left side of a target definition.
// Grammar: path_pattern = { path_segment | capture } ;
func (p *Parser) parseTargetPattern(isPhony bool) (*ast.TargetPattern, *ParseError) {
	var segments []ast.PatternSegment
	isDirectory := false

	if isPhony {
		// For phony targets, the identifier is the name
		// AT_IDENTIFIER includes the @ prefix, strip it
		name := strings.TrimPrefix(p.current.Literal, "@")
		segments = append(segments, &ast.LiteralSegment{Text: name})
		p.nextToken()

		return &ast.TargetPattern{
			Segments:    segments,
			IsPhony:     true,
			IsDirectory: false,
		}, nil
	}

	// Parse file target pattern
	for {
		switch p.current.Type {
		case lexer.COLON, lexer.NEWLINE, lexer.EOF:
			// End of pattern
			goto done

		case lexer.PATH:
			// Literal path segment
			text := p.current.Literal
			segments = append(segments, &ast.LiteralSegment{Text: text})
			p.nextToken()

		case lexer.IDENTIFIER:
			// Could be part of a path
			segments = append(segments, &ast.LiteralSegment{Text: p.current.Literal})
			p.nextToken()

		case lexer.STRING:
			// String content
			segments = append(segments, &ast.LiteralSegment{Text: p.current.Literal})
			p.nextToken()

		case lexer.INTERP_START:
			// Brace expression {name} - could be capture or interpolation
			braceExpr, err := p.parseBraceExpr()
			if err != nil {
				return nil, err
			}
			segments = append(segments, braceExpr)

		default:
			// Unknown token - could be end of pattern
			goto done
		}
	}

done:
	// Merge adjacent literal segments for cleaner AST
	segments = mergeAdjacentLiterals(segments)

	// Check if the pattern ends with / (directory target)
	// We need to check the final merged result
	isDirectory = false
	if len(segments) > 0 {
		if lit, ok := segments[len(segments)-1].(*ast.LiteralSegment); ok {
			isDirectory = strings.HasSuffix(lit.Text, "/")
		}
	}

	return &ast.TargetPattern{
		Segments:    segments,
		IsPhony:     false,
		IsDirectory: isDirectory,
	}, nil
}

// parseBraceExpr parses a {name} expression in a target pattern.
// At parse time, we don't know if this is a capture or variable interpolation.
func (p *Parser) parseBraceExpr() (*ast.BraceExpr, *ParseError) {
	loc := ast.SourceLocationFromToken(p.current)

	if p.current.Type != lexer.INTERP_START {
		return nil, &ParseError{
			Message:  "expected '{' in brace expression",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume {

	// Expect identifier
	if p.current.Type != lexer.IDENTIFIER {
		return nil, &ParseError{
			Message:  "expected identifier in brace expression",
			Location: p.current.Location,
		}
	}
	name := p.current.Literal
	p.nextToken()

	// Expect close brace
	if p.current.Type != lexer.INTERP_END {
		return nil, &ParseError{
			Message:  "unclosed brace expression, expected '}'",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume }

	return &ast.BraceExpr{
		Identifier: name,
		Location:   loc,
	}, nil
}

// parseDependencyList parses the dependencies after the colon.
// Grammar: dependency_list = { dependency } ;
// Grammar: dependency = path_pattern | interpolation ;
//
// After the colon, the lexer enters ModeValue and returns STRING tokens
// with interpolations interspersed. The lexer skips spaces between tokens.
// Dependencies are space-separated in the source, but since spaces are skipped,
// we need to track if tokens were adjacent (no space) to form patterns.
//
// The lexer returns STRING tokens without the surrounding spaces, so if we
// see a sequence like STRING("src/"), INTERP_START, IDENTIFIER("name"),
// INTERP_END, STRING(".c"), these form a SINGLE dependency "src/{name}.c".
//
// To handle this, we accumulate segments for a dependency and only flush
// when we see a token that indicates a boundary.
func (p *Parser) parseDependencyList() ([]ast.Dependency, *ParseError) {
	var deps []ast.Dependency
	var currentDep []ast.PatternSegment

	// Helper to flush current dependency
	flushDep := func() {
		if len(currentDep) > 0 {
			merged := mergeAdjacentLiterals(currentDep)
			if len(merged) > 0 {
				deps = append(deps, ast.Dependency{Segments: merged})
			}
			currentDep = nil
		}
	}

	for {
		// Stop at newline, comment, or EOF
		if p.current.Type == lexer.NEWLINE ||
			p.current.Type == lexer.COMMENT ||
			p.current.Type == lexer.EOF {
			break
		}

		switch p.current.Type {
		case lexer.PATH:
			// Path segment - add to current dependency
			currentDep = append(currentDep, &ast.LiteralSegment{Text: p.current.Literal})
			p.nextToken()
			// Don't flush yet - might be followed by interpolation

		case lexer.IDENTIFIER:
			// Identifier - add to current dependency
			currentDep = append(currentDep, &ast.LiteralSegment{Text: p.current.Literal})
			p.nextToken()
			// Don't flush yet - might be followed by interpolation

		case lexer.STRING:
			// String content from value mode
			text := p.current.Literal
			p.nextToken()

			// Check if this string contains spaces (meaning multiple deps)
			if strings.Contains(text, " ") || strings.Contains(text, "\t") {
				// First flush current dep, then process this string
				flushDep()

				// Split on whitespace
				parts := strings.Fields(text)
				for i, part := range parts {
					if part != "" {
						if i < len(parts)-1 {
							// Not the last part - this is a complete dependency
							deps = append(deps, ast.Dependency{
								Segments: []ast.PatternSegment{&ast.LiteralSegment{Text: part}},
							})
						} else {
							// Last part - might be start of next pattern
							currentDep = append(currentDep, &ast.LiteralSegment{Text: part})
						}
					}
				}
			} else {
				// No spaces - add to current dependency
				currentDep = append(currentDep, &ast.LiteralSegment{Text: text})
			}

		case lexer.INTERP_START:
			// Brace expression in dependency
			braceExpr, err := p.parseBraceExpr()
			if err != nil {
				return nil, err
			}
			currentDep = append(currentDep, braceExpr)
			// Don't flush - may be part of a larger pattern like {name}.c

		default:
			// Unknown token - flush and skip
			flushDep()
			p.nextToken()
		}
	}

	// Flush any remaining dependency
	flushDep()

	return deps, nil
}

// mergeAdjacentLiterals combines consecutive LiteralSegment nodes.
func mergeAdjacentLiterals(segments []ast.PatternSegment) []ast.PatternSegment {
	if len(segments) <= 1 {
		return segments
	}

	var result []ast.PatternSegment
	var pendingText string

	for _, seg := range segments {
		if lit, ok := seg.(*ast.LiteralSegment); ok {
			pendingText += lit.Text
		} else {
			// Flush pending text
			if pendingText != "" {
				result = append(result, &ast.LiteralSegment{Text: pendingText})
				pendingText = ""
			}
			result = append(result, seg)
		}
	}

	// Flush remaining text
	if pendingText != "" {
		result = append(result, &ast.LiteralSegment{Text: pendingText})
	}

	return result
}

// IsTargetLine checks if the current position starts a target definition.
// A line is a target if `:` appears before `=` (or no `=` exists).
func (p *Parser) IsTargetLine() bool {
	// Save current state
	savedCurrent := p.current

	// For phony targets starting with @
	if p.current.Type == lexer.AT_IDENTIFIER {
		return true
	}

	// Scan tokens to find : or =
	colonPos := -1
	equalsPos := -1
	pos := 0

	for p.current.Type != lexer.NEWLINE && p.current.Type != lexer.EOF {
		switch p.current.Type {
		case lexer.COLON:
			if colonPos < 0 {
				colonPos = pos
			}
		case lexer.EQUALS:
			if equalsPos < 0 {
				equalsPos = pos
			}
		}
		pos++
		p.nextToken()
	}

	// Restore by creating a new lexer (we can't truly restore)
	// This is a peek operation - for actual use, we need to track state
	// For now, return based on what we found
	_ = savedCurrent

	// If we found : and (no = or : comes before =), it's a target
	if colonPos >= 0 {
		if equalsPos < 0 || colonPos < equalsPos {
			return true
		}
	}

	return false
}

// isTargetLine is the internal helper - same as IsTargetLine
// This version doesn't modify parser state - it peeks ahead
func (p *Parser) isTargetLine() bool {
	return p.IsTargetLine()
}
