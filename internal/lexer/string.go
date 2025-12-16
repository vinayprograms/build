package lexer

// lexDotKeyword handles .keyword directives.
func (l *Lexer) lexDotKeyword() Token {
	start := l.pos
	startCol := l.col

	l.advance() // consume .

	// Read the keyword part
	for l.pos < len(l.input) && isIdentChar(l.input[l.pos]) {
		l.advance()
	}

	keyword := l.input[start:l.pos]
	if tok, ok := LookupDotKeyword(keyword); ok {
		return l.makeTokenAt(tok, keyword, startCol)
	}

	// Unknown dot keyword, treat as path
	return l.makeTokenAt(PATH, keyword, startCol)
}

// lexAtIdentifier handles @name phony targets.
func (l *Lexer) lexAtIdentifier() Token {
	start := l.pos
	startCol := l.col

	l.advance() // consume @

	// Phony target names can include hyphens (like @test-cover, @debug-lex)
	for l.pos < len(l.input) && isPhonyChar(l.input[l.pos]) {
		l.advance()
	}

	return l.makeTokenAt(AT_IDENTIFIER, l.input[start:l.pos], startCol)
}

// lexIdentifier handles identifiers and keywords.
func (l *Lexer) lexIdentifier() Token {
	start := l.pos
	startCol := l.col

	for l.pos < len(l.input) && isIdentChar(l.input[l.pos]) {
		l.advance()
	}

	literal := l.input[start:l.pos]
	tokType := LookupKeyword(literal)

	l.prevChar = l.input[l.pos-1]
	l.atSOL = false

	return l.makeTokenAt(tokType, literal, startCol)
}

// lexPath handles file paths.
func (l *Lexer) lexPath() Token {
	start := l.pos
	startCol := l.col

	for l.pos < len(l.input) && isPathChar(l.input[l.pos]) {
		ch := l.input[l.pos]
		// Stop at interpolation boundary
		if ch == '{' {
			break
		}
		l.advance()
	}

	l.prevChar = l.input[l.pos-1]
	l.atSOL = false

	return l.makeTokenAt(PATH, l.input[start:l.pos], startCol)
}

// lexString handles general string content.
func (l *Lexer) lexString() Token {
	start := l.pos
	startCol := l.col

	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '\n' || ch == ' ' || ch == '\t' || ch == '#' {
			break
		}
		// Check for interpolation
		if ch == '{' {
			result, _ := ScanInterpolation(l.input, l.pos, l.prevChar, l.atSOL)
			if result.Kind == InterpValid || result.Kind == InterpEscapedOpen {
				break
			}
		}
		l.prevChar = ch
		l.atSOL = false
		l.advance()
	}

	return l.makeTokenAt(STRING, l.input[start:l.pos], startCol)
}

// Character classification functions

// isIdentStart returns true if ch can start an identifier (letter or underscore).
// This is equivalent to IsValidIdentifierStart but kept as unexported for internal use.
func isIdentStart(ch byte) bool {
	return IsValidIdentifierStart(ch)
}

// isIdentChar returns true if ch can be part of an identifier (letter, digit, underscore).
// Note: This intentionally excludes dots, unlike IsValidIdentifierChar which includes
// dots for interpolation identifiers like {target.dir}. Dots are handled separately
// in path parsing and interpolation identifier parsing.
func isIdentChar(ch byte) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9')
}

func isPathChar(ch byte) bool {
	return isIdentChar(ch) || ch == '/' || ch == '.' || ch == '-'
}

// isPhonyChar returns true for characters valid in phony target names.
// Phony targets can use hyphens (e.g., @test-cover, @debug-lex).
func isPhonyChar(ch byte) bool {
	return isIdentChar(ch) || ch == '-'
}
