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
	// ModeCommand is for recipe command lines where all spaces are significant.
	ModeCommand
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

// IndentTracker returns the lexer's indentation tracker.
func (l *Lexer) IndentTracker() *IndentTracker {
	return l.indent
}

// SetValueMode switches the lexer to value mode where leading spaces are skipped
// but internal spaces are preserved.
func (l *Lexer) SetValueMode() {
	l.mode = ModeValue
}

// SetNormalMode switches the lexer back to normal mode.
func (l *Lexer) SetNormalMode() {
	l.mode = ModeNormal
}

// SetCommandMode switches the lexer to command mode where ALL spaces are preserved.
// This is used by the parser when lexing command lines in recipes.
func (l *Lexer) SetCommandMode() {
	l.mode = ModeCommand
}

const blockKeyword = "block"

// PeekNextIsDotKeyword returns true if the next non-whitespace content starts
// a dot keyword (like .shell, .after). This is used by the parser to decide
// whether to use command mode or normal mode for recipe lines.
//
// A dot keyword requires an identifier-start character (letter or underscore)
// immediately after the '.'. This distinguishes directives like ".shell:" from
// command lines that begin with "./", "../", or a bare "." (e.g. "./{build_dir}/app"),
// which must be lexed in command mode so internal spaces are preserved.
func (l *Lexer) PeekNextIsDotKeyword() bool {
	// Skip any whitespace to find the next non-space character
	pos := l.pos
	for pos < len(l.input) && l.input[pos] == ' ' {
		pos++
	}
	if pos >= len(l.input) || l.input[pos] != '.' {
		return false
	}
	pos++
	if pos >= len(l.input) {
		return false
	}
	return isIdentStart(l.input[pos])
}

// PeekNextIsBlock returns true if the next non-whitespace content is the block keyword.
func (l *Lexer) PeekNextIsBlock() bool {
	pos := l.pos
	for pos < len(l.input) && l.input[pos] == ' ' {
		pos++
	}
	remaining := l.input[pos:]
	return len(remaining) >= len(blockKeyword) && remaining[:len(blockKeyword)] == blockKeyword
}

// PeekIsVariableLine checks if the current line is a variable definition.
// Returns true if '=' appears before ':' (or there is no ':' at all).
// This peeks ahead without consuming any tokens.
func (l *Lexer) PeekIsVariableLine() bool {
	pos := l.pos

	// Skip any leading whitespace
	for pos < len(l.input) && (l.input[pos] == ' ' || l.input[pos] == '\t') {
		pos++
	}

	equalsPos := -1
	colonPos := -1

	// Scan until end of line
	for pos < len(l.input) {
		ch := l.input[pos]
		if ch == '\n' {
			break
		}
		if ch == '=' && equalsPos < 0 {
			equalsPos = pos
		}
		if ch == ':' && colonPos < 0 {
			colonPos = pos
		}
		pos++
	}

	// It's a variable if = appears and (no : or = comes before :)
	if equalsPos >= 0 {
		if colonPos < 0 || equalsPos < colonPos {
			return true
		}
	}
	return false
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

	// Skip spaces only in normal mode
	// In value, command, and interp modes, spaces are significant
	if l.mode == ModeNormal {
		l.skipSpaces()
	}

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
		return l.lexInterpContent()
	}

	// In command mode, lex as command (like value but no leading space skip)
	if l.mode == ModeCommand {
		return l.lexCommand()
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
		l.skipSpaces() // Skip leading spaces in value, but not internal spaces
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
		l.skipSpaces() // Skip leading spaces in value, but not internal spaces
		l.mode = ModeValue
		return l.makeToken(COLON, ":")

	case '(':
		l.advance()
		l.skipSpaces()     // Skip leading spaces in function argument
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
	l.atSOL = true

	// Empty line or comment-only line
	if l.pos >= len(l.input) || l.input[l.pos] == '\n' || l.input[l.pos] == '#' {
		l.mode = ModeNormal
		if len(indent) > 0 {
			// Return indent token even for blank/comment lines
			return l.makeTokenAt(INDENT, indent, startCol)
		}
		// No indent, proceed to next token
		return l.NextToken()
	}

	// Non-empty indentation - stay in normal mode by default
	// The parser can switch to command mode if needed for recipe content
	if len(indent) > 0 {
		l.mode = ModeNormal

		// Validate indentation
		_, err := l.indent.Process(indent)
		if err != nil {
			return l.makeTokenAt(ERROR, err.Error(), startCol)
		}

		return l.makeTokenAt(INDENT, indent, startCol)
	}

	// No indentation, continue in normal mode
	l.mode = ModeNormal
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
	startCol := l.col
	// Consume until end of line
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.advance()
	}
	return l.makeTokenAt(COMMENT, l.input[start:l.pos], startCol)
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
		tok := l.makeToken(ERROR, result.Error)
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

// lexInterpContent handles tokens inside an interpolation {}.
func (l *Lexer) lexInterpContent() Token {
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

		return l.makeTokenAt(INTERP_MOD, l.input[start:l.pos], startCol)
	}

	// Identifier
	if isIdentStart(ch) {
		return l.lexInterpIdentifier()
	}

	// Unexpected character in interpolation
	return l.makeToken(ERROR, "unexpected character in interpolation")
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

	return l.makeTokenAt(IDENTIFIER, l.input[start:l.pos], startCol)
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

// makeTokenAt creates a token with a specific column position.
// Use this when the token's start column differs from the current lexer position
// (e.g., after consuming characters for the token's literal).
func (l *Lexer) makeTokenAt(typ TokenType, literal string, col int) Token {
	return Token{
		Type:    typ,
		Literal: literal,
		Location: SourceLocation{
			File:   l.file,
			Line:   l.line,
			Column: col,
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
