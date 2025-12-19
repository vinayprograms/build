// Package errors provides structured error formatting for the build tool.
//
// The package implements error message templates with:
//   - Error codes (E001, E100, E200, etc.)
//   - Source context with line numbers
//   - Caret pointers to error location
//   - Notes for additional context
//   - Help suggestions for fixes
package errors

import (
	"fmt"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
)

// SourceLine represents a single line of source code for error display.
type SourceLine struct {
	Number int    // Line number (1-based)
	Text   string // Line content
}

// Format formats the source line with a line number prefix.
// width is the minimum width for the line number.
func (l SourceLine) Format(width int) string {
	return fmt.Sprintf("%*d | %s", width, l.Number, l.Text)
}

// FormattedError represents a structured error with source context.
type FormattedError struct {
	Code        string             // Error code (E001, E100, etc.)
	Message     string             // Brief error description
	Location    ast.SourceLocation // Source location
	SourceLines []SourceLine       // Source context (1-3 lines)
	CaretLine   int                // Line to show caret on (0 = no caret)
	CaretColumn int                // Column for caret (1-based)
	Note        string             // Additional context (optional)
	Help        string             // Fix suggestion (optional)
}

// Error implements the error interface.
func (e *FormattedError) Error() string {
	return e.Format()
}

// Format returns the fully formatted error message.
func (e *FormattedError) Format() string {
	var b strings.Builder

	// Header: error[CODE]: message
	b.WriteString(fmt.Sprintf("error[%s]: %s\n", e.Code, e.Message))

	// Location: --> file:line:column
	b.WriteString(fmt.Sprintf(" --> %s\n", e.Location.String()))

	// Source context
	if len(e.SourceLines) > 0 {
		// Calculate line number width
		maxLine := 0
		for _, sl := range e.SourceLines {
			if sl.Number > maxLine {
				maxLine = sl.Number
			}
		}
		width := lineNumberWidth(maxLine)

		for _, sl := range e.SourceLines {
			b.WriteString(sl.Format(width))
			b.WriteString("\n")

			// Add caret line if this is the error line
			if sl.Number == e.CaretLine && e.CaretColumn > 0 {
				b.WriteString(FormatCaret(width, e.CaretColumn))
				b.WriteString("\n")
			}
		}
	}

	// Note (optional)
	if e.Note != "" {
		b.WriteString(fmt.Sprintf("note: %s\n", e.Note))
	}

	// Help (optional)
	if e.Help != "" {
		b.WriteString(fmt.Sprintf("help: %s\n", e.Help))
	}

	return b.String()
}

// FormatCaret creates a caret line pointing to a column.
// width is the line number field width, column is 1-based.
func FormatCaret(width, column int) string {
	// Format: "    | " + spaces + "^"
	// The prefix is width + " | "
	prefix := strings.Repeat(" ", width) + " | "
	spaces := strings.Repeat(" ", column-1)
	return prefix + spaces + "^"
}

// lineNumberWidth returns the number of digits needed to display a line number.
func lineNumberWidth(line int) int {
	if line < 10 {
		return 1
	}
	if line < 100 {
		return 2
	}
	if line < 1000 {
		return 3
	}
	return 4
}

// NewFormattedError creates a basic formatted error.
func NewFormattedError(code, message string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     code,
		Message:  message,
		Location: loc,
	}
}

// WithNote adds a note to the error.
func (e *FormattedError) WithNote(note string) *FormattedError {
	e.Note = note
	return e
}

// WithHelp adds a help suggestion to the error.
func (e *FormattedError) WithHelp(help string) *FormattedError {
	e.Help = help
	return e
}

// WithSourceContext adds source lines to the error.
func (e *FormattedError) WithSourceContext(lines []SourceLine, caretLine, caretColumn int) *FormattedError {
	e.SourceLines = lines
	e.CaretLine = caretLine
	e.CaretColumn = caretColumn
	return e
}

// ExtractSourceLines extracts source lines around a location from source content.
// context specifies how many lines before and after to include (1-3 recommended).
func ExtractSourceLines(source string, line, context int) []SourceLine {
	lines := strings.Split(source, "\n")

	// Calculate range
	start := line - context - 1 // Convert to 0-based and subtract context
	if start < 0 {
		start = 0
	}
	end := line + context // Line + context (already accounts for 1-based)
	if end > len(lines) {
		end = len(lines)
	}

	var result []SourceLine
	for i := start; i < end; i++ {
		result = append(result, SourceLine{
			Number: i + 1, // Convert back to 1-based
			Text:   lines[i],
		})
	}
	return result
}

// ExtractSourceLinesFromFile extracts source lines from a file around a location.
// context specifies how many lines before and after to include.
func ExtractSourceLinesFromFile(path string, line, context int) ([]SourceLine, error) {
	content, err := readFile(path)
	if err != nil {
		return nil, err
	}
	return ExtractSourceLines(content, line, context), nil
}

// readFile reads file content. Isolated for testing.
var readFile = func(path string) (string, error) {
	// This will be replaced in tests
	return "", fmt.Errorf("not implemented")
}

// SetFileReader sets the file reader function. Used for testing.
func SetFileReader(fn func(string) (string, error)) {
	readFile = fn
}

// InitFileReader initializes the file reader to use os.ReadFile.
// Call this in main or init to enable file-based source extraction.
func InitFileReader() {
	// Import os in the actual implementation
	// readFile = func(path string) (string, error) {
	//     content, err := os.ReadFile(path)
	//     return string(content), err
	// }
}
