package lexer

// lexValue handles value mode (after = or :).
// It recognizes function calls but does NOT skip spaces.
// Leading space skipping is handled at the transition point (after = or :).
func (l *Lexer) lexValue() Token {
	// Do NOT skip spaces here - spaces are significant within values.
	// Leading spaces are skipped when transitioning from = or : to ModeValue,
	// but internal spaces (including those after interpolations) must be preserved.
	return l.lexValueOrCommand(true)
}

// lexCommand handles command mode (recipe lines where spaces are significant).
// Unlike lexValue, this does not skip leading spaces and doesn't recognize functions.
func (l *Lexer) lexCommand() Token {
	return l.lexValueOrCommand(false)
}

// lexValueOrCommand is the unified implementation for both value and command modes.
// When isValueMode is true, it recognizes function calls and parentheses/commas.
// When isValueMode is false (command mode), it preserves all content as strings.
func (l *Lexer) lexValueOrCommand(isValueMode bool) Token {
	if l.pos >= len(l.input) {
		return l.makeToken(EOF, "")
	}

	ch := l.input[l.pos]

	// End of value/command on newline
	if ch == '\n' {
		l.mode = ModeNormal
		return l.lexNewline()
	}

	// Comment ends value/command
	if ch == '#' {
		l.mode = ModeNormal
		return l.lexComment()
	}

	// Interpolation
	if ch == '{' {
		result, end := ScanInterpolation(l.input, l.pos, l.prevChar, l.atSOL)
		switch result.Kind {
		case InterpValid:
			return l.lexInterpolation(result, end)
		case InterpEscapedOpen:
			tok := l.makeToken(ESCAPE_LBRACE, "{{")
			l.pos = end
			l.col += 2
			l.prevChar = '{'
			l.atSOL = false
			return tok
		case InterpError:
			tok := l.makeToken(ERROR, result.Error)
			l.pos = end
			return tok
		}
		// InterpNotInterp - fall through to string
	}

	// Escaped close brace
	if ch == '}' {
		if esc, end := ScanEscapedCloseBrace(l.input, l.pos); esc {
			tok := l.makeToken(ESCAPE_RBRACE, "}}")
			l.pos = end
			l.col += 2
			l.prevChar = '}'
			l.atSOL = false
			return tok
		}
		// Single } might be end of interpolation
		l.advance()
		return l.makeToken(INTERP_END, "}")
	}

	// Value mode: handle function calls and parentheses/commas
	if isValueMode {
		// Close paren ends function argument
		if ch == ')' {
			l.advance()
			return l.makeToken(RPAREN, ")")
		}

		// Open paren for function calls
		if ch == '(' {
			l.advance()
			return l.makeToken(LPAREN, "(")
		}

		// Comma separates function arguments
		if ch == ',' {
			l.advance()
			// Skip whitespace after comma (before next argument)
			for l.pos < len(l.input) && l.input[l.pos] == ' ' {
				l.pos++
				l.col++
			}
			return l.makeToken(COMMA, ",")
		}

		// Check for function names (only if followed by '(' to be a function call)
		if isIdentStart(ch) {
			// Peek ahead to see if it's a function call
			ident := l.peekIdentifier()
			if tok := LookupKeyword(ident); tok.IsFunction() {
				// Check if followed by (
				peekPos := l.pos + len(ident)
				// Skip whitespace
				for peekPos < len(l.input) && l.input[peekPos] == ' ' {
					peekPos++
				}
				if peekPos < len(l.input) && l.input[peekPos] == '(' {
					return l.lexIdentifier()
				}
				// Not followed by (, treat as part of string
			}
		}
	}

	// Consume as string until special character
	return l.lexContentString(isValueMode)
}

// lexContentString consumes a string in value/command mode until a special character.
// When isValueMode is true, stops at parentheses and commas (for function args).
// When isValueMode is false (command mode), only stops at newline/comment.
func (l *Lexer) lexContentString(isValueMode bool) Token {
	start := l.pos
	startCol := l.col

	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		// Always stop at newline and comment
		if ch == '\n' || ch == '#' {
			break
		}

		// Value mode: also stop at parentheses and commas
		if isValueMode && (ch == ')' || ch == '(' || ch == ',') {
			break
		}

		// Check for interpolation start
		if ch == '{' {
			result, _ := ScanInterpolation(l.input, l.pos, l.prevChar, l.atSOL)
			if result.Kind == InterpValid || result.Kind == InterpEscapedOpen || result.Kind == InterpError {
				break
			}
		}

		// Check for escaped close brace or single close brace
		if ch == '}' {
			if esc, _ := ScanEscapedCloseBrace(l.input, l.pos); esc {
				break
			}
			// Single } ends the string segment
			break
		}

		l.prevChar = ch
		l.atSOL = false
		l.advance()
	}

	literal := l.input[start:l.pos]

	// Trim trailing whitespace when stopping at a comment.
	// This ensures "value  # comment" produces "value" not "value  ".
	// This applies in both value and command mode.
	if l.pos < len(l.input) && l.input[l.pos] == '#' {
		literal = trimTrailingWhitespace(literal)
	} else if isValueMode && (l.pos >= len(l.input) || l.input[l.pos] == '\n') {
		// C3: in value mode (variable values, directive values, and
		// dependency lists), trailing whitespace at the actual end of the
		// value - end of line or end of file - is trimmed too, e.g.
		// "name = value   " lexes to "value", not "value   ". Internal
		// spaces, and segments that stop because more of the value follows
		// (an interpolation, or a function call's parenthesis/comma), are
		// untouched. Command mode (plain recipe command lines and block
		// lines) is unaffected, since SetCommandMode passes isValueMode=false.
		literal = trimTrailingWhitespace(literal)
	}

	if len(literal) == 0 {
		// If we didn't consume anything, advance one character
		if l.pos < len(l.input) {
			l.advance()
			literal = l.input[start:l.pos]
		}
	}

	return l.makeTokenAt(STRING, literal, startCol)
}

// trimTrailingWhitespace removes trailing spaces and tabs from a string.
func trimTrailingWhitespace(s string) string {
	end := len(s)
	for end > 0 && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[:end]
}
