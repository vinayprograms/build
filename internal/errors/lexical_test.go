package errors

import (
	"strings"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

// ----------------------------------------------------------------------------
// Lexical Error Tests (E001-E099)
// ----------------------------------------------------------------------------

func TestNewInvalidCharacterError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 5, Column: 10}
	err := NewInvalidCharacterError('\x00', loc)

	if err.Code != "E001" {
		t.Errorf("Expected code E001, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "invalid character") {
		t.Errorf("Message should mention invalid character: %s", err.Message)
	}
	if err.Location != loc {
		t.Errorf("Location mismatch: got %v, want %v", err.Location, loc)
	}
}

func TestNewMixedIndentationError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 3, Column: 1}
	err := NewMixedIndentationError(loc)

	if err.Code != "E002" {
		t.Errorf("Expected code E002, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "mixed") && !strings.Contains(err.Message, "indentation") {
		t.Errorf("Message should mention mixed indentation: %s", err.Message)
	}
	if err.Help == "" {
		t.Error("Mixed indentation error should have help text")
	}
}

func TestNewInconsistentIndentationError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 7, Column: 1}
	err := NewInconsistentIndentationError("space", "tab", loc)

	if err.Code != "E003" {
		t.Errorf("Expected code E003, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "inconsistent") {
		t.Errorf("Message should mention inconsistent indentation: %s", err.Message)
	}
	if !strings.Contains(err.Message, "space") || !strings.Contains(err.Message, "tab") {
		t.Errorf("Message should mention expected and got types: %s", err.Message)
	}
}

func TestNewInvalidIndentWidthError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 5, Column: 1}
	err := NewInvalidIndentWidthError(5, 4, loc)

	if err.Code != "E004" {
		t.Errorf("Expected code E004, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "5") || !strings.Contains(err.Message, "4") {
		t.Errorf("Message should mention actual and expected widths: %s", err.Message)
	}
	if err.Help == "" {
		t.Error("Invalid indent width error should have help text")
	}
}

func TestNewUnclosedInterpolationError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 2, Column: 15}
	err := NewUnclosedInterpolationError("varname", loc)

	if err.Code != "E005" {
		t.Errorf("Expected code E005, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "unclosed") {
		t.Errorf("Message should mention unclosed interpolation: %s", err.Message)
	}
	if !strings.Contains(err.Message, "varname") {
		t.Errorf("Message should mention variable name: %s", err.Message)
	}
	if err.Help == "" {
		t.Error("Unclosed interpolation error should have help text")
	}
}

func TestNewInvalidModifierError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 4, Column: 20}
	err := NewInvalidModifierError("foo", loc)

	if err.Code != "E006" {
		t.Errorf("Expected code E006, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "invalid modifier") {
		t.Errorf("Message should mention invalid modifier: %s", err.Message)
	}
	if !strings.Contains(err.Message, "foo") {
		t.Errorf("Message should mention the invalid modifier: %s", err.Message)
	}
	if !strings.Contains(err.Help, "raw") {
		t.Errorf("Help should suggest :raw: %s", err.Help)
	}
}

func TestNewUnexpectedCharInInterpError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 3, Column: 12}
	err := NewUnexpectedCharInInterpError('$', loc)

	if err.Code != "E007" {
		t.Errorf("Expected code E007, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "unexpected character") {
		t.Errorf("Message should mention unexpected character: %s", err.Message)
	}
	if !strings.Contains(err.Message, "$") {
		t.Errorf("Message should include the character: %s", err.Message)
	}
}

func TestNewInvalidEscapeSequenceError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 1, Column: 5}
	err := NewInvalidEscapeSequenceError("\\n", loc)

	if err.Code != "E008" {
		t.Errorf("Expected code E008, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "invalid escape") {
		t.Errorf("Message should mention invalid escape: %s", err.Message)
	}
}

// ----------------------------------------------------------------------------
// Error Formatting Tests
// ----------------------------------------------------------------------------

func TestLexicalError_Format_WithContext(t *testing.T) {
	source := "cc = gcc\n\t foo = bar"
	loc := ast.SourceLocation{File: "Buildfile", Line: 2, Column: 1}
	err := NewMixedIndentationError(loc)

	// Add source context
	lines := ExtractSourceLines(source, loc.Line, 1)
	err.WithSourceContext(lines, loc.Line, loc.Column)

	formatted := err.Format()

	// Should show error code
	if !strings.Contains(formatted, "E002") {
		t.Errorf("Formatted error should contain code: %s", formatted)
	}

	// Should show location
	if !strings.Contains(formatted, "Buildfile:2:1") {
		t.Errorf("Formatted error should contain location: %s", formatted)
	}

	// Should show help text
	if !strings.Contains(formatted, "help:") {
		t.Errorf("Formatted error should contain help: %s", formatted)
	}
}

// ----------------------------------------------------------------------------
// Error Code Range Tests
// ----------------------------------------------------------------------------

func TestLexicalErrorCodes_InRange(t *testing.T) {
	// All lexical errors should have codes E001-E099
	testCases := []struct {
		name string
		err  *FormattedError
	}{
		{"InvalidCharacter", NewInvalidCharacterError('\x00', ast.SourceLocation{})},
		{"MixedIndentation", NewMixedIndentationError(ast.SourceLocation{})},
		{"InconsistentIndentation", NewInconsistentIndentationError("space", "tab", ast.SourceLocation{})},
		{"InvalidIndentWidth", NewInvalidIndentWidthError(3, 4, ast.SourceLocation{})},
		{"UnclosedInterpolation", NewUnclosedInterpolationError("var", ast.SourceLocation{})},
		{"InvalidModifier", NewInvalidModifierError("bad", ast.SourceLocation{})},
		{"UnexpectedCharInInterp", NewUnexpectedCharInInterpError('$', ast.SourceLocation{})},
		{"InvalidEscapeSequence", NewInvalidEscapeSequenceError("\\x", ast.SourceLocation{})},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := tc.err.Code
			if len(code) != 4 || code[0] != 'E' {
				t.Errorf("Error code should be E### format, got: %s", code)
			}

			// Parse numeric part
			num := 0
			for i := 1; i < len(code); i++ {
				if code[i] < '0' || code[i] > '9' {
					t.Errorf("Error code should be numeric after E: %s", code)
					return
				}
				num = num*10 + int(code[i]-'0')
			}

			if num < 1 || num > 99 {
				t.Errorf("Lexical error code should be E001-E099, got: %s", code)
			}
		})
	}
}
