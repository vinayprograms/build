package errors

import (
	"fmt"
	"strings"

	"github.com/vinayprograms/need/internal/ast"
)

// Syntax Error Codes (E100-E199)
// These errors occur during parsing (syntax analysis).
const (
	// E100: Unexpected token during parsing
	CodeUnexpectedToken = "E100"

	// E101: Missing colon in target definition
	CodeMissingColon = "E101"

	// E102: Missing 'end' to close conditional
	CodeMissingEnd = "E102"

	// E103: Directive used in invalid scope
	CodeInvalidDirectiveScope = "E103"

	// E104: Missing condition expression after if/elif
	CodeMissingCondition = "E104"

	// E105: Missing comparison operator in condition
	CodeMissingOperator = "E105"

	// E106: Missing identifier after ifdef/ifndef
	CodeMissingIdentifier = "E106"

	// E107: Invalid runtime type in .using directive
	CodeInvalidRuntime = "E107"

	// E108: Wrong number of arguments to function
	CodeMissingFunctionArgument = "E108"

	// E109: Circular include detected
	CodeCircularInclude = "E109"

	// E110: Included file not found
	CodeIncludeNotFound = "E110"
)

// NewUnexpectedTokenError creates an error for unexpected tokens during parsing.
func NewUnexpectedTokenError(got, expected string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeUnexpectedToken,
		Message:  fmt.Sprintf("unexpected token '%s', expected '%s'", got, expected),
		Location: loc,
	}
}

// NewMissingColonError creates an error for missing colon in target definition.
func NewMissingColonError(target string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeMissingColon,
		Message:  fmt.Sprintf("missing ':' in target definition"),
		Location: loc,
		Note:     "targets require a colon between name and dependencies",
		Help:     fmt.Sprintf("change to: %s: [dependencies]", target),
	}
}

// NewMissingEndError creates an error for missing 'end' keyword.
func NewMissingEndError(loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeMissingEnd,
		Message:  "missing 'end' to close conditional block",
		Location: loc,
		Note:     "every 'if', 'ifdef', or 'ifndef' requires a matching 'end'",
		Help:     "add 'end' at the same indentation level as the opening 'if'",
	}
}

// NewInvalidDirectiveScopeError creates an error for directives in wrong scope.
func NewInvalidDirectiveScopeError(directive, currentScope string, validScopes []string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeInvalidDirectiveScope,
		Message:  fmt.Sprintf("directive '%s' is not valid at %s scope", directive, currentScope),
		Location: loc,
		Note:     fmt.Sprintf("'%s' can only be used in certain contexts", directive),
		Help:     fmt.Sprintf("'%s' is valid in: %s", directive, strings.Join(validScopes, ", ")),
	}
}

// NewMissingConditionError creates an error for missing condition after if/elif.
func NewMissingConditionError(keyword string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeMissingCondition,
		Message:  fmt.Sprintf("expected condition expression after '%s'", keyword),
		Location: loc,
		Help:     fmt.Sprintf("example: %s {os} == linux", keyword),
	}
}

// NewMissingOperatorError creates an error for missing comparison operator.
func NewMissingOperatorError(loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeMissingOperator,
		Message:  "expected '==' or '!=' in condition",
		Location: loc,
		Note:     "conditions must compare values using '==' or '!='",
		Help:     "example: if {os} == linux",
	}
}

// NewMissingIdentifierError creates an error for missing identifier after ifdef/ifndef.
func NewMissingIdentifierError(keyword string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeMissingIdentifier,
		Message:  fmt.Sprintf("expected identifier after '%s'", keyword),
		Location: loc,
		Help:     fmt.Sprintf("example: %s DEBUG", keyword),
	}
}

// NewInvalidRuntimeError creates an error for invalid runtime in .using directive.
func NewInvalidRuntimeError(runtime string, loc ast.SourceLocation) *FormattedError {
	validRuntimes := "bare, docker, podman, devcontainer, nix, lima"
	return &FormattedError{
		Code:     CodeInvalidRuntime,
		Message:  fmt.Sprintf("invalid runtime '%s'", runtime),
		Location: loc,
		Note:     "runtime must be one of the supported types",
		Help:     fmt.Sprintf("valid runtimes are: %s", validRuntimes),
	}
}

// NewMissingFunctionArgumentError creates an error for wrong argument count.
func NewMissingFunctionArgumentError(funcName string, expected, got int, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeMissingFunctionArgument,
		Message:  fmt.Sprintf("function '%s' requires %d arguments, got %d", funcName, expected, got),
		Location: loc,
		Help:     getFunctionUsageHelp(funcName),
	}
}

// getFunctionUsageHelp returns usage help for built-in functions.
func getFunctionUsageHelp(funcName string) string {
	switch funcName {
	case "shell":
		return "usage: shell(command)"
	case "glob":
		return "usage: glob(pattern)"
	case "filename":
		return "usage: filename(path)"
	case "dirname":
		return "usage: dirname(path)"
	case "replace":
		return "usage: replace(input, from, to)"
	default:
		return ""
	}
}

// NewCircularIncludeError creates an error for circular include detection.
func NewCircularIncludeError(file string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeCircularInclude,
		Message:  fmt.Sprintf("circular include detected: '%s'", file),
		Location: loc,
		Note:     "included files cannot include themselves directly or indirectly",
	}
}

// NewIncludeNotFoundError creates an error for missing included files.
func NewIncludeNotFoundError(file string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeIncludeNotFound,
		Message:  fmt.Sprintf("included file not found: '%s'", file),
		Location: loc,
		Help:     "check that the file path is correct and the file exists",
	}
}
