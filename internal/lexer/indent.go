package lexer

import "fmt"

// IndentChar represents the type of character used for indentation.
type IndentChar int

const (
	// IndentUnknown indicates indentation character not yet determined.
	IndentUnknown IndentChar = iota
	// IndentSpace indicates spaces are used for indentation.
	IndentSpace
	// IndentTab indicates tabs are used for indentation.
	IndentTab
)

// String returns the string representation of the indent character type.
func (c IndentChar) String() string {
	switch c {
	case IndentUnknown:
		return "unknown"
	case IndentSpace:
		return "space"
	case IndentTab:
		return "tab"
	default:
		return fmt.Sprintf("IndentChar(%d)", c)
	}
}

// IndentError represents an indentation-related error.
type IndentError struct {
	Message string
	Line    int
	Column  int
}

// Error implements the error interface.
func (e *IndentError) Error() string {
	return fmt.Sprintf("indentation error at line %d, column %d: %s", e.Line, e.Column, e.Message)
}

// IndentTracker tracks indentation state across lines in a Needfile.
//
// Per DESIGN.md Section 2.3.3:
//   - First indented line establishes the indent unit (e.g., 4 spaces)
//   - Subsequent indents must be multiples of this unit
//   - Mixed tabs/spaces after first line is an error
//   - Level 0 = column 0, Level 1 = one unit, Level 2 = two units
type IndentTracker struct {
	char  IndentChar // Character type used for indentation
	width int        // Width of one indentation unit (in characters)
}

// NewIndentTracker creates a new IndentTracker with no established indent unit.
func NewIndentTracker() *IndentTracker {
	return &IndentTracker{
		char:  IndentUnknown,
		width: 0,
	}
}

// Char returns the established indentation character type.
func (t *IndentTracker) Char() IndentChar {
	return t.char
}

// Width returns the width of one indentation unit.
func (t *IndentTracker) Width() int {
	return t.width
}

// Reset clears the tracker state, allowing a new indent unit to be established.
func (t *IndentTracker) Reset() {
	t.char = IndentUnknown
	t.width = 0
}

// Process analyzes an indentation string and returns the logical indentation level.
// The input should be the leading whitespace of a line (spaces and/or tabs only).
//
// Returns:
//   - level: The logical indentation level (0, 1, 2, ...)
//   - error: Non-nil if the indentation is invalid (mixed chars, inconsistent width)
//
// Behavior:
//   - Empty string always returns level 0
//   - First non-empty indent establishes the character type and unit width
//   - Subsequent indents must use the same character and be exact multiples of the unit
func (t *IndentTracker) Process(indent string) (int, error) {
	// Empty indentation is always level 0
	if len(indent) == 0 {
		return 0, nil
	}

	// Determine the character type of this indent
	indentChar, err := t.detectChar(indent)
	if err != nil {
		return 0, err
	}

	// If this is the first indented line, establish the unit
	if t.char == IndentUnknown {
		t.char = indentChar
		t.width = len(indent)
		return 1, nil
	}

	// Check that the character type matches
	if indentChar != t.char {
		return 0, &IndentError{
			Message: fmt.Sprintf("inconsistent indentation: expected %s, got %s", t.char, indentChar),
		}
	}

	// Check that the width is an exact multiple of the unit
	if len(indent)%t.width != 0 {
		return 0, &IndentError{
			Message: fmt.Sprintf("indentation width %d is not a multiple of the unit width %d", len(indent), t.width),
		}
	}

	return len(indent) / t.width, nil
}

// detectChar determines whether the indent string uses spaces or tabs.
// Returns an error if the string contains mixed characters.
func (t *IndentTracker) detectChar(indent string) (IndentChar, error) {
	if len(indent) == 0 {
		return IndentUnknown, nil
	}

	// Check the first character to determine type
	var expectedChar IndentChar
	switch indent[0] {
	case ' ':
		expectedChar = IndentSpace
	case '\t':
		expectedChar = IndentTab
	default:
		return IndentUnknown, &IndentError{
			Message: fmt.Sprintf("invalid indentation character: %q", indent[0]),
		}
	}

	// Verify all characters are the same type
	for i := 1; i < len(indent); i++ {
		var charType IndentChar
		switch indent[i] {
		case ' ':
			charType = IndentSpace
		case '\t':
			charType = IndentTab
		default:
			return IndentUnknown, &IndentError{
				Message: fmt.Sprintf("invalid indentation character: %q", indent[i]),
			}
		}

		if charType != expectedChar {
			return IndentUnknown, &IndentError{
				Message: "mixed indentation: spaces and tabs cannot be mixed",
			}
		}
	}

	return expectedChar, nil
}
