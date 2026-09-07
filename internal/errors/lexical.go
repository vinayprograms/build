package errors

import (
	"fmt"

	"github.com/vinayprograms/need/internal/ast"
)

// Lexical Error Codes (E001-E099)
// These errors occur during lexical analysis (tokenization).
const (
	// E001: Invalid character in source
	CodeInvalidCharacter = "E001"

	// E002: Mixed tabs and spaces in indentation
	CodeMixedIndentation = "E002"

	// E003: Inconsistent indentation character (switched from spaces to tabs or vice versa)
	CodeInconsistentIndentation = "E003"

	// E004: Indentation width is not a multiple of the established unit
	CodeInvalidIndentWidth = "E004"

	// E005: Interpolation opened with { but not closed with }
	CodeUnclosedInterpolation = "E005"

	// E006: Invalid modifier in interpolation (only :raw is valid)
	CodeInvalidModifier = "E006"

	// E007: Unexpected character inside interpolation
	CodeUnexpectedCharInInterp = "E007"

	// E008: Invalid escape sequence
	CodeInvalidEscapeSequence = "E008"
)

// NewInvalidCharacterError creates an error for invalid characters in source.
func NewInvalidCharacterError(ch byte, loc ast.SourceLocation) *FormattedError {
	var charRepr string
	if ch >= 32 && ch < 127 {
		charRepr = fmt.Sprintf("'%c'", ch)
	} else {
		charRepr = fmt.Sprintf("0x%02X", ch)
	}
	return &FormattedError{
		Code:     CodeInvalidCharacter,
		Message:  fmt.Sprintf("invalid character %s", charRepr),
		Location: loc,
		Note:     "Needfiles must contain valid UTF-8 text",
	}
}

// NewMixedIndentationError creates an error for mixing tabs and spaces in indentation.
func NewMixedIndentationError(loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeMixedIndentation,
		Message:  "mixed indentation: spaces and tabs cannot be mixed",
		Location: loc,
		Help:     "use either spaces or tabs consistently for indentation, not both",
	}
}

// NewInconsistentIndentationError creates an error for switching between tabs and spaces.
func NewInconsistentIndentationError(expected, got string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeInconsistentIndentation,
		Message:  fmt.Sprintf("inconsistent indentation: expected %s, got %s", expected, got),
		Location: loc,
		Note:     "the first indented line establishes the indentation character",
		Help:     fmt.Sprintf("use %s for indentation throughout the file", expected),
	}
}

// NewInvalidIndentWidthError creates an error for incorrect indentation width.
func NewInvalidIndentWidthError(got, unit int, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeInvalidIndentWidth,
		Message:  fmt.Sprintf("indentation width %d is not a multiple of the unit width %d", got, unit),
		Location: loc,
		Help:     fmt.Sprintf("use %d, %d, %d, etc. spaces for indentation", unit, unit*2, unit*3),
	}
}

// NewUnclosedInterpolationError creates an error for unclosed interpolation.
func NewUnclosedInterpolationError(name string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeUnclosedInterpolation,
		Message:  fmt.Sprintf("unclosed interpolation: {%s", name),
		Location: loc,
		Help:     fmt.Sprintf("add closing brace: {%s}", name),
	}
}

// NewInvalidModifierError creates an error for invalid interpolation modifiers.
func NewInvalidModifierError(modifier string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeInvalidModifier,
		Message:  fmt.Sprintf("invalid modifier ':%s'", modifier),
		Location: loc,
		Note:     "the only valid modifier is ':raw'",
		Help:     "use :raw for unquoted shell expansion, e.g., {flags:raw}",
	}
}

// NewUnexpectedCharInInterpError creates an error for unexpected characters in interpolation.
func NewUnexpectedCharInInterpError(ch byte, loc ast.SourceLocation) *FormattedError {
	var charRepr string
	if ch >= 32 && ch < 127 {
		charRepr = fmt.Sprintf("'%c'", ch)
	} else {
		charRepr = fmt.Sprintf("0x%02X", ch)
	}
	return &FormattedError{
		Code:     CodeUnexpectedCharInInterp,
		Message:  fmt.Sprintf("unexpected character %s in interpolation", charRepr),
		Location: loc,
		Note:     "interpolation identifiers can contain letters, digits, underscores, and dots",
	}
}

// NewInvalidEscapeSequenceError creates an error for invalid escape sequences.
func NewInvalidEscapeSequenceError(seq string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeInvalidEscapeSequence,
		Message:  fmt.Sprintf("invalid escape sequence '%s'", seq),
		Location: loc,
		Note:     "valid escape sequences are {{ and }}",
		Help:     "use {{ for a literal { and }} for a literal }",
	}
}
