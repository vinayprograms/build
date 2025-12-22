package lexer

import "fmt"

// InterpResultKind indicates the result of scanning for an interpolation.
type InterpResultKind int

const (
	// InterpValid indicates a valid interpolation was found.
	InterpValid InterpResultKind = iota
	// InterpEscapedOpen indicates an escaped open brace {{ was found.
	InterpEscapedOpen
	// InterpNotInterp indicates the { is not an interpolation start.
	InterpNotInterp
	// InterpError indicates a malformed interpolation.
	InterpError
)

// String returns the string representation of the result kind.
func (k InterpResultKind) String() string {
	switch k {
	case InterpValid:
		return "Valid"
	case InterpEscapedOpen:
		return "EscapedOpen"
	case InterpNotInterp:
		return "NotInterp"
	case InterpError:
		return "Error"
	default:
		return fmt.Sprintf("InterpResultKind(%d)", k)
	}
}

// InterpResult holds the result of scanning an interpolation.
type InterpResult struct {
	Kind  InterpResultKind
	Name  string // The identifier name (for InterpValid)
	Raw   bool   // Whether :raw modifier was present
	Error string // Error message (for InterpError)
}

// IsValidIdentifierStart returns true if c can start an identifier.
// Per spec: letter or underscore.
func IsValidIdentifierStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

// IsValidIdentifierChar returns true if c can be part of an identifier.
// Per spec: letter, digit, underscore, or dot (for target.dir, target.file).
func IsValidIdentifierChar(c byte) bool {
	return IsValidIdentifierStart(c) || (c >= '0' && c <= '9') || c == '.'
}

// IsInterpBoundary returns true if prev is a valid boundary character
// for interpolation recognition.
//
// Per DESIGN.md Section 2.3.2 (extended for practical use):
// `{` is recognized as INTERP_START if preceded by:
//   - whitespace (space or tab)
//   - start-of-line (atSOL=true)
//   - `:` or `=` (operators)
//   - `/` (path separator, for patterns like {dir}/{file})
//   - `"` or `'` (quotes, for strings like "{var}")
//   - `(`, `)`, `,` (function call syntax)
//   - `>`, `<` (shell redirections)
//   - `-` (hyphen, for patterns like app-{version})
//   - `}` (closing brace, for consecutive interpolations like {a}{b})
func IsInterpBoundary(prev byte, atSOL bool) bool {
	if atSOL {
		return true
	}
	switch prev {
	case ' ', '\t', ':', '=', '/', '"', '\'', '(', ')', ',', '>', '<', '-', '}':
		return true
	default:
		return false
	}
}

// ScanInterpolation attempts to scan an interpolation starting at pos.
// The character at input[pos] must be '{'.
//
// Returns:
//   - result: The scan result (valid interpolation, escape, not-interp, or error)
//   - end: The position after the scanned content
//
// For InterpNotInterp, end equals pos (nothing consumed).
// For InterpEscapedOpen, end is pos+2 (consumed "{{").
// For InterpValid/InterpError, end is position after the closing '}'.
func ScanInterpolation(input string, pos int, prev byte, atSOL bool) (InterpResult, int) {
	// Must start with '{'
	if pos >= len(input) || input[pos] != '{' {
		return InterpResult{Kind: InterpNotInterp}, pos
	}

	// Check for escape sequence {{
	if pos+1 < len(input) && input[pos+1] == '{' {
		return InterpResult{Kind: InterpEscapedOpen}, pos + 2
	}

	// Check boundary rule
	if !IsInterpBoundary(prev, atSOL) {
		return InterpResult{Kind: InterpNotInterp}, pos
	}

	// Check for valid identifier start after '{'
	if pos+1 >= len(input) {
		return InterpResult{Kind: InterpNotInterp}, pos
	}

	nextChar := input[pos+1]
	if !IsValidIdentifierStart(nextChar) {
		return InterpResult{Kind: InterpNotInterp}, pos
	}

	// Scan the identifier
	identStart := pos + 1
	identEnd := identStart
	for identEnd < len(input) && IsValidIdentifierChar(input[identEnd]) {
		identEnd++
	}

	name := input[identStart:identEnd]
	raw := false

	// Check for modifier or closing brace
	if identEnd >= len(input) {
		return InterpResult{
			Kind:  InterpError,
			Name:  name,
			Error: fmt.Sprintf("unclosed interpolation: {%s", name),
		}, identEnd
	}

	switch input[identEnd] {
	case '}':
		// Valid interpolation without modifier
		return InterpResult{Kind: InterpValid, Name: name, Raw: false}, identEnd + 1

	case ':':
		// Check for :raw modifier
		modStart := identEnd + 1
		modEnd := modStart
		for modEnd < len(input) && input[modEnd] != '}' {
			modEnd++
		}

		if modEnd >= len(input) {
			return InterpResult{
				Kind:  InterpError,
				Name:  name,
				Error: fmt.Sprintf("unclosed interpolation: {%s:", name),
			}, modEnd
		}

		modifier := input[modStart:modEnd]
		if modifier == "raw" {
			raw = true
		} else {
			return InterpResult{
				Kind:  InterpError,
				Name:  name,
				Error: fmt.Sprintf("invalid modifier ':%s', expected ':raw'", modifier),
			}, modEnd + 1
		}

		return InterpResult{Kind: InterpValid, Name: name, Raw: raw}, modEnd + 1

	default:
		// Unexpected character - unclosed interpolation
		return InterpResult{
			Kind:  InterpError,
			Name:  name,
			Error: fmt.Sprintf("unclosed interpolation: {%s", name),
		}, identEnd
	}
}

// ScanEscapedCloseBrace checks if there's an escaped close brace }} at pos.
// Returns (true, pos+2) if escape found, (false, pos) otherwise.
func ScanEscapedCloseBrace(input string, pos int) (bool, int) {
	if pos+1 < len(input) && input[pos] == '}' && input[pos+1] == '}' {
		return true, pos + 2
	}
	return false, pos
}
