package main

import (
	"os"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/errors"
	"github.com/vinayprograms/build/internal/parser"
	"github.com/vinayprograms/build/internal/semantic"
)

// initFileReader initializes the error package's file reader for source context.
func initFileReader() {
	errors.SetFileReader(func(path string) (string, error) {
		content, err := os.ReadFile(path)
		return string(content), err
	})
}

// FormatParseError converts a parser.ParseError to a formatted error with source context.
func FormatParseError(e *parser.ParseError, source string) *errors.FormattedError {
	// Determine error code based on message content
	code := determineParseErrorCode(e)

	formatted := &errors.FormattedError{
		Code:     code,
		Message:  e.Message,
		Location: astLocationFromLexer(e.Location),
	}

	// Add hint if present
	if e.Hint != "" {
		formatted.Help = e.Hint
	}

	// Add source context
	if source != "" && e.Location.Line > 0 {
		lines := errors.ExtractSourceLines(source, e.Location.Line, 1)
		formatted.SourceLines = lines
		formatted.CaretLine = e.Location.Line
		formatted.CaretColumn = e.Location.Column
	}

	return formatted
}

// determineParseErrorCode determines the appropriate error code from the message.
func determineParseErrorCode(e *parser.ParseError) string {
	msg := e.Message

	// Detect directive scope errors
	if containsAny(msg, "invalid at", "invalid at GLOBAL scope", "invalid at RECIPE scope") {
		return errors.CodeInvalidDirectiveScope
	}

	// Detect missing end errors
	if containsAny(msg, "missing 'end'", "expected 'end'") {
		return errors.CodeMissingEnd
	}

	// Detect missing colon errors
	if containsAny(msg, "missing ':'", "expected ':'") {
		return errors.CodeMissingColon
	}

	// Detect circular include
	if containsAny(msg, "circular include") {
		return errors.CodeCircularInclude
	}

	// Detect include not found
	if containsAny(msg, "included file not found", "include file not found") {
		return errors.CodeIncludeNotFound
	}

	// Default to unexpected token
	return errors.CodeUnexpectedToken
}

// containsAny returns true if s contains any of the substrings.
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if containsString(s, sub) {
			return true
		}
	}
	return false
}

// containsString is a simple substring check.
func containsString(s, sub string) bool {
	return len(sub) <= len(s) && indexOfString(s, sub) >= 0
}

// indexOfString finds the first occurrence of sub in s.
func indexOfString(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// astLocationFromLexer converts a lexer.SourceLocation to ast.SourceLocation.
func astLocationFromLexer(loc interface{ String() string }) ast.SourceLocation {
	// The lexer.SourceLocation and ast.SourceLocation have the same structure
	// but we need to extract values since they're different types
	str := loc.String()
	// Parse "file:line:column"
	var file string
	var line, column int
	parseLocationString(str, &file, &line, &column)
	return ast.SourceLocation{
		File:   file,
		Line:   line,
		Column: column,
	}
}

// parseLocationString parses a "file:line:column" string.
func parseLocationString(s string, file *string, line, column *int) {
	// Find last two colons
	lastColon := lastIndexOf(s, ':')
	if lastColon < 0 {
		*file = s
		return
	}

	secondLastColon := lastIndexOf(s[:lastColon], ':')
	if secondLastColon < 0 {
		*file = s[:lastColon]
		*line = atoi(s[lastColon+1:])
		return
	}

	*file = s[:secondLastColon]
	*line = atoi(s[secondLastColon+1 : lastColon])
	*column = atoi(s[lastColon+1:])
}

// lastIndexOf finds the last occurrence of ch in s.
func lastIndexOf(s string, ch byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == ch {
			return i
		}
	}
	return -1
}

// atoi converts string to int.
func atoi(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			n = n*10 + int(s[i]-'0')
		}
	}
	return n
}

// FormatSemanticError converts a semantic error to a formatted error.
func FormatSemanticError(e error, source string) *errors.FormattedError {
	switch err := e.(type) {
	case *semantic.DuplicateDefinitionError:
		return formatDuplicateError(err, source)
	case *semantic.UndefinedVariableError:
		return formatUndefinedError(err, source)
	case *semantic.AutomaticOutsideRecipeError:
		return formatAutomaticOutsideRecipeError(err, source)
	case *semantic.AutomaticInPatternError:
		return formatAutomaticInPatternError(err, source)
	case *semantic.CaptureMismatchError:
		return formatCaptureMismatchError(err, source)
	case *semantic.CircularDependencyError:
		return formatCircularDependencyError(err)
	default:
		// Generic error
		return &errors.FormattedError{
			Code:    "E299",
			Message: e.Error(),
		}
	}
}

func formatDuplicateError(e *semantic.DuplicateDefinitionError, source string) *errors.FormattedError {
	formatted := &errors.FormattedError{
		Code:     errors.CodeDuplicateVariable,
		Message:  e.Error(),
		Location: e.Second,
		Note:     "first defined at " + e.First.String(),
	}

	if source != "" && e.Second.Line > 0 {
		lines := errors.ExtractSourceLines(source, e.Second.Line, 1)
		formatted.SourceLines = lines
		formatted.CaretLine = e.Second.Line
		formatted.CaretColumn = e.Second.Column
	}

	return formatted
}

func formatUndefinedError(e *semantic.UndefinedVariableError, source string) *errors.FormattedError {
	formatted := &errors.FormattedError{
		Code:     errors.CodeUndefinedVariable,
		Message:  e.Error(),
		Location: e.Location,
		Help:     "define the variable before use, or check for typos",
	}

	if source != "" && e.Location.Line > 0 {
		lines := errors.ExtractSourceLines(source, e.Location.Line, 1)
		formatted.SourceLines = lines
		formatted.CaretLine = e.Location.Line
		formatted.CaretColumn = e.Location.Column
	}

	return formatted
}

func formatAutomaticOutsideRecipeError(e *semantic.AutomaticOutsideRecipeError, source string) *errors.FormattedError {
	formatted := &errors.FormattedError{
		Code:     errors.CodeAutomaticOutsideRecipe,
		Message:  e.Error(),
		Location: e.Location,
		Note:     "automatic variables like {target}, {deps}, {in}, etc. are only valid inside recipes",
	}

	if source != "" && e.Location.Line > 0 {
		lines := errors.ExtractSourceLines(source, e.Location.Line, 1)
		formatted.SourceLines = lines
		formatted.CaretLine = e.Location.Line
		formatted.CaretColumn = e.Location.Column
	}

	return formatted
}

func formatAutomaticInPatternError(e *semantic.AutomaticInPatternError, source string) *errors.FormattedError {
	formatted := &errors.FormattedError{
		Code:     errors.CodeAutomaticInPattern,
		Message:  e.Error(),
		Location: e.Location,
		Note:     "automatic variables cannot be used as pattern captures",
	}

	if source != "" && e.Location.Line > 0 {
		lines := errors.ExtractSourceLines(source, e.Location.Line, 1)
		formatted.SourceLines = lines
		formatted.CaretLine = e.Location.Line
		formatted.CaretColumn = e.Location.Column
	}

	return formatted
}

func formatCaptureMismatchError(e *semantic.CaptureMismatchError, source string) *errors.FormattedError {
	formatted := &errors.FormattedError{
		Code:     errors.CodeCaptureMismatch,
		Message:  e.Error(),
		Location: e.Location,
		Note:     "captures must be defined in the target pattern before use in dependencies",
	}

	if source != "" && e.Location.Line > 0 {
		lines := errors.ExtractSourceLines(source, e.Location.Line, 1)
		formatted.SourceLines = lines
		formatted.CaretLine = e.Location.Line
		formatted.CaretColumn = e.Location.Column
	}

	return formatted
}

func formatCircularDependencyError(e *semantic.CircularDependencyError) *errors.FormattedError {
	return &errors.FormattedError{
		Code:    errors.CodeCircularDependency,
		Message: e.Error(),
		Note:    "break the cycle by removing or restructuring dependencies",
	}
}
