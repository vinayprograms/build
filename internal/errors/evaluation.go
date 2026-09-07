package errors

import (
	"fmt"

	"github.com/vinayprograms/need/internal/ast"
)

// Evaluation Error Codes (E300-E399)
// These errors occur during variable evaluation and function execution.
const (
	// E300: Shell command failed during shell() function
	CodeShellCommandFailed = "E300"

	// E301: Glob pattern matched no files
	CodeGlobNoMatch = "E301"

	// E302: Invalid arguments to function
	CodeInvalidFunctionArguments = "E302"

	// E303: Forward reference in immediate variable
	CodeForwardReference = "E303"

	// E304: Error during lazy variable evaluation
	CodeLazyEvaluationError = "E304"

	// E305: Error evaluating condition in conditional
	CodeConditionEvaluationError = "E305"
)

// NewShellCommandFailedError creates an error for shell() function failures.
func NewShellCommandFailedError(cmd string, exitCode int, stderr string, loc ast.SourceLocation) *FormattedError {
	msg := fmt.Sprintf("shell command failed: %s", cmd)
	if len(msg) > 80 {
		msg = fmt.Sprintf("shell command failed: %s...", cmd[:60])
	}
	err := &FormattedError{
		Code:     CodeShellCommandFailed,
		Message:  msg,
		Location: loc,
		Note:     fmt.Sprintf("exit code %d", exitCode),
	}
	if stderr != "" && len(stderr) < 200 {
		err.Help = fmt.Sprintf("command output: %s", stderr)
	}
	return err
}

// NewGlobNoMatchError creates an error for glob patterns matching no files.
func NewGlobNoMatchError(pattern string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeGlobNoMatch,
		Message:  fmt.Sprintf("glob pattern '%s' matched no files", pattern),
		Location: loc,
		Help:     "verify the pattern and ensure matching files exist",
	}
}

// NewInvalidFunctionArgumentsError creates an error for invalid function arguments.
func NewInvalidFunctionArgumentsError(funcName, reason string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeInvalidFunctionArguments,
		Message:  fmt.Sprintf("invalid argument to '%s': %s", funcName, reason),
		Location: loc,
		Help:     getFunctionUsageHelp(funcName),
	}
}

// NewForwardReferenceError creates an error for forward references in immediate variables.
func NewForwardReferenceError(varName, referencedVar string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeForwardReference,
		Message:  fmt.Sprintf("forward reference in variable '%s': references '%s' which is defined later", varName, referencedVar),
		Location: loc,
		Note:     "immediate variables are evaluated in definition order",
		Help:     fmt.Sprintf("define '%s' before '%s', or make '%s' a lazy variable", referencedVar, varName, varName),
	}
}

// NewLazyEvaluationError creates an error for lazy variable evaluation failures.
func NewLazyEvaluationError(varName, reason string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeLazyEvaluationError,
		Message:  fmt.Sprintf("error evaluating lazy variable '%s': %s", varName, reason),
		Location: loc,
		Note:     "lazy variables are evaluated when first used",
	}
}

// NewConditionEvaluationError creates an error for condition evaluation failures.
func NewConditionEvaluationError(reason string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeConditionEvaluationError,
		Message:  fmt.Sprintf("error evaluating condition: %s", reason),
		Location: loc,
	}
}
