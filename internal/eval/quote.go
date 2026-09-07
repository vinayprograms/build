package eval

import "strings"

// ----------------------------------------------------------------------------
// Context-aware interpolation quoting (D1)
//
// {var} interpolated into a recipe command line, block line, or shell()
// argument is formatted according to the shell quoting context it lands in,
// tracked left to right over the surrounding literal text:
//
//   - Outside quotes: emitted bare if it contains only characters from
//     [A-Za-z0-9_./:@%+=,-]; otherwise wrapped in single quotes, with an
//     embedded ' emitted as '\''.
//   - Inside double quotes: emitted with ", $, ` and \ escaped by a
//     backslash. Never wrapped.
//   - Inside single quotes: emitted raw, with ' turned into '\''.
//
// $( ... ), ` ... ` (command substitution) and heredoc bodies are treated
// as double-quoted context for this purpose. {var:raw} bypasses all of this
// and is always emitted untouched.
// ----------------------------------------------------------------------------

// quoteCtx is the effective shell quoting context at a given point in the
// scan, used to select how an interpolated value is formatted.
type quoteCtx int

const (
	ctxUnquoted quoteCtx = iota
	ctxSingle
	ctxDouble
)

// frameKind identifies what opened a nested scanning frame. Several kinds
// (frameDouble, frameSubst, frameBacktick, frameHeredoc) share the same
// effective quoting context (ctxDouble) but differ in how/when they close.
type frameKind int

const (
	frameDouble frameKind = iota
	frameSingle
	frameSubst
	frameBacktick
	frameHeredoc
)

func (k frameKind) effective() quoteCtx {
	if k == frameSingle {
		return ctxSingle
	}
	return ctxDouble
}

// quoteScanner tracks shell quote state left to right across the literal
// text of a command line (or, for a block, across all of its lines), plus
// enough heredoc bookkeeping to treat an open heredoc body as double-quoted
// context. It is reused across calls to scanLiteral so state carries over
// between LiteralCommand parts and (for blocks) between lines.
type quoteScanner struct {
	stack []frameKind

	// heredocActive/heredocDelim/heredocStrip describe an in-progress
	// heredoc body (set once the block-level driver activates a pending
	// heredoc detected on a previous line).
	heredocActive bool
	heredocDelim  string
	heredocStrip  bool // true for <<-, allows leading tabs on the terminator line

	// pendingHeredocDelim/pendingHeredocStrip capture a "<<[-]DELIM" seen
	// while scanning the current line; the block-level driver activates it
	// once the line is finished (a heredoc body starts on the *next* line).
	pendingHeredocDelim string
	pendingHeredocStrip bool
}

func newQuoteScanner() *quoteScanner {
	return &quoteScanner{}
}

// current returns the effective quoting context at the current scan
// position.
func (s *quoteScanner) current() quoteCtx {
	if len(s.stack) == 0 {
		return ctxUnquoted
	}
	return s.stack[len(s.stack)-1].effective()
}

func (s *quoteScanner) topKind() (frameKind, bool) {
	if len(s.stack) == 0 {
		return 0, false
	}
	return s.stack[len(s.stack)-1], true
}

func (s *quoteScanner) push(k frameKind) { s.stack = append(s.stack, k) }

func (s *quoteScanner) pop() {
	if len(s.stack) > 0 {
		s.stack = s.stack[:len(s.stack)-1]
	}
}

// scanLiteral advances the scanner state over a chunk of literal source
// text (never over an interpolated value itself - {var:raw} and friends
// stay untouched precisely because their expansion is not scanned).
func (s *quoteScanner) scanLiteral(text string) {
	n := len(text)
	i := 0
	for i < n {
		top, hasTop := s.topKind()

		if hasTop && top == frameSingle {
			// Inside single quotes only a literal ' closes it; backslash is
			// not an escape character here.
			if text[i] == '\'' {
				s.pop()
			}
			i++
			continue
		}

		if hasTop && top == frameHeredoc {
			// Heredoc body: quote characters are literal; only backslash
			// escaping (for the purposes of this scanner) still applies so
			// an escaped $ inside the body doesn't confuse later scanning.
			if text[i] == '\\' && i+1 < n {
				i += 2
				continue
			}
			i++
			continue
		}

		c := text[i]

		// Backslash escapes the next character (unquoted, double-quoted,
		// $()/backtick contexts all honor this).
		if c == '\\' && i+1 < n {
			i += 2
			continue
		}

		switch {
		case c == '\'':
			if hasTop && top == frameDouble {
				// Literal inside double quotes.
				i++
				continue
			}
			s.push(frameSingle)
			i++

		case c == '"':
			if hasTop && top == frameDouble {
				s.pop()
			} else if hasTop && top == frameHeredoc {
				// unreachable (handled above) but kept for clarity
				i++
				continue
			} else {
				s.push(frameDouble)
			}
			i++

		case c == '`':
			if hasTop && top == frameBacktick {
				s.pop()
			} else {
				s.push(frameBacktick)
			}
			i++

		case c == '$' && i+1 < n && text[i+1] == '(':
			s.push(frameSubst)
			i += 2

		case c == ')':
			if hasTop && top == frameSubst {
				s.pop()
			}
			i++

		case c == '<' && !hasTop && i+1 < n && text[i+1] == '<':
			// Heredoc start, only recognized at the true top level
			// (matches the common case; heredocs inside $()/backticks are
			// out of scope here).
			j := i + 2
			strip := false
			if j < n && text[j] == '-' {
				strip = true
				j++
			}
			for j < n && (text[j] == ' ' || text[j] == '\t') {
				j++
			}
			var quoteChar byte
			if j < n && (text[j] == '\'' || text[j] == '"') {
				quoteChar = text[j]
				j++
			}
			start := j
			for j < n && isHeredocDelimChar(text[j], quoteChar) {
				j++
			}
			delim := text[start:j]
			if quoteChar != 0 && j < n && text[j] == quoteChar {
				j++
			}
			if delim != "" {
				s.pendingHeredocDelim = delim
				s.pendingHeredocStrip = strip
				i = j
				continue
			}
			i++

		default:
			i++
		}
	}
}

func isHeredocDelimChar(c byte, quoteChar byte) bool {
	if quoteChar != 0 {
		return c != quoteChar
	}
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// activatePendingHeredoc, called by block-level drivers between lines,
// turns a "<<[-]DELIM" seen on the previous line into an active heredoc
// body frame for the line about to be scanned.
func (s *quoteScanner) activatePendingHeredoc() {
	if s.pendingHeredocDelim == "" {
		return
	}
	s.push(frameHeredoc)
	s.heredocActive = true
	s.heredocDelim = s.pendingHeredocDelim
	s.heredocStrip = s.pendingHeredocStrip
	s.pendingHeredocDelim = ""
	s.pendingHeredocStrip = false
}

// checkHeredocTerminator closes an active heredoc body if plainLine (the
// concatenation of a line's literal text, ignoring any interpolations) is
// exactly the terminator delimiter. hasInterp must be false for a line to
// ever qualify (a real terminator line is bare text).
func (s *quoteScanner) checkHeredocTerminator(plainLine string, hasInterp bool) {
	if !s.heredocActive || hasInterp {
		return
	}
	term := plainLine
	if s.heredocStrip {
		term = strings.TrimLeft(term, "\t")
	}
	if term != s.heredocDelim {
		return
	}
	if top, ok := s.topKind(); ok && top == frameHeredoc {
		s.pop()
	}
	s.heredocActive = false
	s.heredocDelim = ""
	s.heredocStrip = false
}

// ----------------------------------------------------------------------------
// Value formatting per quoting context
// ----------------------------------------------------------------------------

// bareChars are the characters a value may consist entirely of to be
// emitted without quoting outside of any shell quotes.
const bareChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_./:@%+=,-"

func isBareByte(c byte) bool {
	return strings.IndexByte(bareChars, c) >= 0
}

// isBareValue reports whether s contains only characters that may be
// emitted bare outside of quotes.
func isBareValue(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isBareByte(s[i]) {
			return false
		}
	}
	return true
}

// escapeSingleQuoteBody replaces each ' with '\” - the idiom for
// preserving a literal single quote while remaining (functionally) inside
// a single-quoted string.
func escapeSingleQuoteBody(s string) string {
	if !strings.Contains(s, "'") {
		return s
	}
	return strings.ReplaceAll(s, "'", `'\''`)
}

// escapeDoubleQuoteBody escapes ", $, ` and \ by a backslash, for a value
// landing inside an existing double-quoted string (or an equivalent
// context: $(...), backticks, heredoc bodies).
func escapeDoubleQuoteBody(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"', '$', '`', '\\':
			b.WriteByte('\\')
		}
		b.WriteByte(c)
	}
	return b.String()
}

// formatUnquoted renders a value for an unquoted (bare shell word) context:
// bare if safe, otherwise wrapped in single quotes.
func formatUnquoted(s string) string {
	if isBareValue(s) {
		return s
	}
	return "'" + escapeSingleQuoteBody(s) + "'"
}

// formatForContext renders val for the given quoting context.
func formatForContext(ctx quoteCtx, val string) string {
	switch ctx {
	case ctxSingle:
		return escapeSingleQuoteBody(val)
	case ctxDouble:
		return escapeDoubleQuoteBody(val)
	default:
		return formatUnquoted(val)
	}
}
