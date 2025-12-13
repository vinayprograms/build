package lexer

// LexerMode represents the current lexing mode.
type LexerMode int

const (
	// ModeLineStart is at the beginning of a line, consuming indentation.
	ModeLineStart LexerMode = iota
	// ModeNormal is normal token recognition.
	ModeNormal
	// ModeValue is consuming the rest of line as a value (after = or :).
	ModeValue
	// ModeInterp is inside an interpolation {}.
	ModeInterp
)

// Lexer performs lexical analysis on Buildfile source code.
type Lexer struct {
	file   string // Source file name
	input  string // Input source
	pos    int    // Current position in input
	line   int    // Current line number (1-based)
	col    int    // Current column number (1-based)
	mode   LexerMode
	indent *IndentTracker

	// State for tracking previous character (for interpolation boundaries)
	prevChar byte
	atSOL    bool // At start of line (for interpolation)

	// Stack for returning to previous mode after interpolation
	modeStack []LexerMode
}

// New creates a new Lexer for the given source.
func New(file, input string) *Lexer {
	return &Lexer{
		file:   file,
		input:  input,
		pos:    0,
		line:   1,
		col:    1,
		mode:   ModeLineStart,
		indent: NewIndentTracker(),
		atSOL:  true,
	}
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	// Handle end of input
	if l.pos >= len(l.input) {
		return l.makeToken(EOF, "")
	}

	// Handle line start (indentation)
	if l.mode == ModeLineStart {
		return l.lexLineStart()
	}

	// Skip spaces (but not newlines or tabs at line start)
	l.skipSpaces()

	if l.pos >= len(l.input) {
		return l.makeToken(EOF, "")
	}

	ch := l.input[l.pos]

	// Newline
	if ch == '\n' {
		return l.lexNewline()
	}

	// Comment
	if ch == '#' {
		return l.lexComment()
	}

	// In interpolation mode, lex identifier, modifier, or close brace
	if l.mode == ModeInterp {
		return l.lexInsideInterp()
	}

	// In value mode, lex differently
	if l.mode == ModeValue {
		return l.lexValue()
	}

	// Operators and punctuation
	switch ch {
	case '=':
		if l.peek(1) == '=' {
			l.advance()
			l.advance()
			return l.makeToken(DOUBLE_EQUALS, "==")
		}
		l.advance()
		l.mode = ModeValue
		return l.makeToken(EQUALS, "=")

	case '!':
		if l.peek(1) == '=' {
			l.advance()
			l.advance()
			return l.makeToken(NOT_EQUALS, "!=")
		}
		// Single ! is part of a string
		return l.lexString()

	case ':':
		l.advance()
		l.mode = ModeValue
		return l.makeToken(COLON, ":")

	case '(':
		l.advance()
		l.mode = ModeValue // Inside parens, lex as value
		return l.makeToken(LPAREN, "(")

	case ')':
		l.advance()
		return l.makeToken(RPAREN, ")")

	case '{':
		return l.lexBrace()

	case '}':
		// Check for escaped close brace }}
		if l.peek(1) == '}' {
			tok := l.makeToken(ESCAPE_RBRACE, "}}")
			l.advance()
			l.advance()
			return tok
		}
		// Single } in normal mode is unexpected, treat as string
		return l.lexString()

	case '.':
		// Check for dot keyword
		return l.lexDotKeyword()

	case '@':
		return l.lexAtIdentifier()
	}

	// Identifier, keyword, or path
	if isIdentStart(ch) {
		// Look ahead to see if this is a path (contains / or .)
		if l.looksLikePath() {
			return l.lexPath()
		}
		return l.lexIdentifier()
	}

	// Path starting with / or .
	if ch == '/' || ch == '.' {
		return l.lexPath()
	}

	// Fallback: consume as string
	return l.lexString()
}

// lexLineStart handles the start of a line (indentation).
func (l *Lexer) lexLineStart() Token {
	start := l.pos
	startCol := l.col

	// Consume leading whitespace
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch != ' ' && ch != '\t' {
			break
		}
		l.advance()
	}

	indent := l.input[start:l.pos]
	l.mode = ModeNormal
	l.atSOL = true

	// Empty line or comment-only line
	if l.pos >= len(l.input) || l.input[l.pos] == '\n' || l.input[l.pos] == '#' {
		if len(indent) > 0 {
			// Return indent token even for blank/comment lines
			return Token{
				Type:    INDENT,
				Literal: indent,
				Location: SourceLocation{
					File:   l.file,
					Line:   l.line,
					Column: startCol,
				},
			}
		}
		// No indent, proceed to next token
		return l.NextToken()
	}

	// Non-empty indentation
	if len(indent) > 0 {
		// Validate indentation
		_, err := l.indent.Process(indent)
		if err != nil {
			return Token{
				Type:    ERROR,
				Literal: err.Error(),
				Location: SourceLocation{
					File:   l.file,
					Line:   l.line,
					Column: startCol,
				},
			}
		}

		return Token{
			Type:    INDENT,
			Literal: indent,
			Location: SourceLocation{
				File:   l.file,
				Line:   l.line,
				Column: startCol,
			},
		}
	}

	// No indentation, continue to next token
	return l.NextToken()
}

// lexNewline handles newline characters.
func (l *Lexer) lexNewline() Token {
	tok := l.makeToken(NEWLINE, "\n")
	l.advance()
	l.line++
	l.col = 1
	l.mode = ModeLineStart
	l.atSOL = true
	return tok
}

// lexComment handles comment tokens.
func (l *Lexer) lexComment() Token {
	start := l.pos
	// Consume until end of line
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.advance()
	}
	return Token{
		Type:    COMMENT,
		Literal: l.input[start:l.pos],
		Location: SourceLocation{
			File:   l.file,
			Line:   l.line,
			Column: l.col - (l.pos - start),
		},
	}
}

// lexValue handles value mode (after = or :).
func (l *Lexer) lexValue() Token {
	l.skipSpaces()

	if l.pos >= len(l.input) {
		return l.makeToken(EOF, "")
	}

	ch := l.input[l.pos]

	// End of value on newline
	if ch == '\n' {
		l.mode = ModeNormal
		return l.lexNewline()
	}

	// Comment ends value
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
			tok := Token{
				Type:    ERROR,
				Literal: result.Error,
				Location: SourceLocation{
					File:   l.file,
					Line:   l.line,
					Column: l.col,
				},
			}
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

	// Check for function names
	if isIdentStart(ch) {
		// Peek ahead to see if it's a function
		ident := l.peekIdentifier()
		if tok := LookupKeyword(ident); tok.IsFunction() {
			return l.lexIdentifier()
		}
	}

	// Consume as string until special character
	return l.lexValueString()
}

// lexValueString consumes a string value until a special character.
func (l *Lexer) lexValueString() Token {
	start := l.pos
	startCol := l.col

	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		// Stop at special characters
		if ch == '\n' || ch == '#' || ch == ')' || ch == '(' {
			break
		}

		// Check for interpolation start
		if ch == '{' {
			result, _ := ScanInterpolation(l.input, l.pos, l.prevChar, l.atSOL)
			if result.Kind == InterpValid || result.Kind == InterpEscapedOpen || result.Kind == InterpError {
				break
			}
		}

		// Check for escaped close brace
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
	if len(literal) == 0 {
		// If we didn't consume anything, advance one character
		if l.pos < len(l.input) {
			l.advance()
			literal = l.input[start:l.pos]
		}
	}

	return Token{
		Type:    STRING,
		Literal: literal,
		Location: SourceLocation{
			File:   l.file,
			Line:   l.line,
			Column: startCol,
		},
	}
}

// lexBrace handles { character.
func (l *Lexer) lexBrace() Token {
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
		tok := Token{
			Type:    ERROR,
			Literal: result.Error,
			Location: SourceLocation{
				File:   l.file,
				Line:   l.line,
				Column: l.col,
			},
		}
		l.pos = end
		return tok

	default:
		// Not an interpolation, consume as string
		return l.lexString()
	}
}

// lexInterpolation handles a valid interpolation.
func (l *Lexer) lexInterpolation(result InterpResult, end int) Token {
	startTok := l.makeToken(INTERP_START, "{")
	l.advance() // consume {

	// Save current mode and switch to interpolation mode
	l.modeStack = append(l.modeStack, l.mode)
	l.mode = ModeInterp

	l.prevChar = '{'
	l.atSOL = false

	return startTok
}

// lexInsideInterp handles tokens inside an interpolation {}.
func (l *Lexer) lexInsideInterp() Token {
	ch := l.input[l.pos]

	// Close brace ends interpolation
	if ch == '}' {
		tok := l.makeToken(INTERP_END, "}")
		l.advance()

		// Restore previous mode
		if len(l.modeStack) > 0 {
			l.mode = l.modeStack[len(l.modeStack)-1]
			l.modeStack = l.modeStack[:len(l.modeStack)-1]
		} else {
			l.mode = ModeNormal
		}

		l.prevChar = '}'
		return tok
	}

	// Colon for :raw modifier
	if ch == ':' {
		start := l.pos
		startCol := l.col
		l.advance() // consume :

		// Read modifier name
		for l.pos < len(l.input) && isIdentChar(l.input[l.pos]) {
			l.advance()
		}

		return Token{
			Type:    INTERP_MOD,
			Literal: l.input[start:l.pos],
			Location: SourceLocation{
				File:   l.file,
				Line:   l.line,
				Column: startCol,
			},
		}
	}

	// Identifier
	if isIdentStart(ch) {
		return l.lexInterpIdentifier()
	}

	// Unexpected character in interpolation
	return Token{
		Type:    ERROR,
		Literal: "unexpected character in interpolation",
		Location: SourceLocation{
			File:   l.file,
			Line:   l.line,
			Column: l.col,
		},
	}
}

// lexInterpIdentifier handles identifiers inside interpolations.
// Allows dots for target.dir, target.file.
func (l *Lexer) lexInterpIdentifier() Token {
	start := l.pos
	startCol := l.col

	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if isIdentChar(ch) || ch == '.' {
			l.advance()
		} else {
			break
		}
	}

	return Token{
		Type:    IDENTIFIER,
		Literal: l.input[start:l.pos],
		Location: SourceLocation{
			File:   l.file,
			Line:   l.line,
			Column: startCol,
		},
	}
}

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
		return Token{
			Type:    tok,
			Literal: keyword,
			Location: SourceLocation{
				File:   l.file,
				Line:   l.line,
				Column: startCol,
			},
		}
	}

	// Unknown dot keyword, treat as path
	return Token{
		Type:    PATH,
		Literal: keyword,
		Location: SourceLocation{
			File:   l.file,
			Line:   l.line,
			Column: startCol,
		},
	}
}

// lexAtIdentifier handles @name phony targets.
func (l *Lexer) lexAtIdentifier() Token {
	start := l.pos
	startCol := l.col

	l.advance() // consume @

	for l.pos < len(l.input) && isIdentChar(l.input[l.pos]) {
		l.advance()
	}

	return Token{
		Type:    AT_IDENTIFIER,
		Literal: l.input[start:l.pos],
		Location: SourceLocation{
			File:   l.file,
			Line:   l.line,
			Column: startCol,
		},
	}
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

	return Token{
		Type:    tokType,
		Literal: literal,
		Location: SourceLocation{
			File:   l.file,
			Line:   l.line,
			Column: startCol,
		},
	}
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

	return Token{
		Type:    PATH,
		Literal: l.input[start:l.pos],
		Location: SourceLocation{
			File:   l.file,
			Line:   l.line,
			Column: startCol,
		},
	}
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

	return Token{
		Type:    STRING,
		Literal: l.input[start:l.pos],
		Location: SourceLocation{
			File:   l.file,
			Line:   l.line,
			Column: startCol,
		},
	}
}

// Helper methods

func (l *Lexer) advance() {
	if l.pos < len(l.input) {
		l.prevChar = l.input[l.pos]
		l.pos++
		l.col++
		l.atSOL = false
	}
}

func (l *Lexer) peek(offset int) byte {
	pos := l.pos + offset
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

func (l *Lexer) skipSpaces() {
	for l.pos < len(l.input) && l.input[l.pos] == ' ' {
		l.prevChar = ' '
		l.pos++
		l.col++
		l.atSOL = false
	}
}

func (l *Lexer) makeToken(typ TokenType, literal string) Token {
	return Token{
		Type:    typ,
		Literal: literal,
		Location: SourceLocation{
			File:   l.file,
			Line:   l.line,
			Column: l.col,
		},
	}
}

func (l *Lexer) peekIdentifier() string {
	start := l.pos
	end := start
	for end < len(l.input) && isIdentChar(l.input[end]) {
		end++
	}
	return l.input[start:end]
}

// looksLikePath checks if the current position starts a path-like token
// (contains / or . after the identifier part).
func (l *Lexer) looksLikePath() bool {
	pos := l.pos
	// Skip identifier chars
	for pos < len(l.input) && isIdentChar(l.input[pos]) {
		pos++
	}
	// Check if followed by / or .
	if pos < len(l.input) {
		ch := l.input[pos]
		return ch == '/' || ch == '.'
	}
	return false
}

// Character classification

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentChar(ch byte) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9')
}

func isPathChar(ch byte) bool {
	return isIdentChar(ch) || ch == '/' || ch == '.' || ch == '-'
}
