package errors

import (
	"fmt"

	"github.com/vinayprograms/need/internal/ast"
)

// Execution Error Codes (E400-E499)
// These errors occur during recipe execution and build runtime.
const (
	// E400: Recipe command returned non-zero exit code
	CodeRecipeFailed = "E400"

	// E401: Required dependency file does not exist
	CodeMissingDependency = "E401"

	// E402: Required binary not found in PATH
	CodeMissingBinary = "E402"

	// E403: Specified shell not found
	CodeShellNotFound = "E403"

	// E404: Binary version does not meet requirement
	CodeVersionMismatch = "E404"

	// E405: Requested target not defined in Needfile
	CodeTargetNotFound = "E405"

	// E406: No default target and none specified
	CodeNoDefaultTarget = "E406"
)

// NewRecipeFailedError creates an error for failed recipe execution.
func NewRecipeFailedError(target, command string, exitCode int, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeRecipeFailed,
		Message:  fmt.Sprintf("recipe failed for target '%s'", target),
		Location: loc,
		Note:     fmt.Sprintf("command '%s' exited with code %d", truncateCommand(command), exitCode),
	}
}

// truncateCommand truncates a command for display.
func truncateCommand(cmd string) string {
	if len(cmd) > 50 {
		return cmd[:47] + "..."
	}
	return cmd
}

// NewMissingDependencyError creates an error for missing dependency files.
func NewMissingDependencyError(dependency, target string) *FormattedError {
	return &FormattedError{
		Code:    CodeMissingDependency,
		Message: fmt.Sprintf("missing dependency '%s' for target '%s'", dependency, target),
		Note:    "dependency file does not exist and no rule to build it",
		Help:    fmt.Sprintf("create '%s' or add a target to build it", dependency),
	}
}

// NewMissingBinaryError creates an error for required binaries not in PATH.
func NewMissingBinaryError(binary string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeMissingBinary,
		Message:  fmt.Sprintf("missing required binary '%s'", binary),
		Location: loc,
		Note:     fmt.Sprintf("'%s' is required but not found in PATH", binary),
		Help:     fmt.Sprintf("install '%s' or check your PATH", binary),
	}
}

// NewShellNotFoundError creates an error for missing shell.
func NewShellNotFoundError(shell string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeShellNotFound,
		Message:  fmt.Sprintf("shell not found: '%s'", shell),
		Location: loc,
		Help:     "install the shell or use a different one with .shell directive",
	}
}

// NewVersionMismatchError creates an error for version requirement not met.
func NewVersionMismatchError(binary, required, found string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeVersionMismatch,
		Message:  fmt.Sprintf("version mismatch for '%s': requires %s, found %s", binary, required, found),
		Location: loc,
		Help:     fmt.Sprintf("upgrade '%s' to version %s or higher", binary, required),
	}
}

// NewTargetNotFoundError creates an error for undefined targets.
func NewTargetNotFoundError(target string) *FormattedError {
	return &FormattedError{
		Code:    CodeTargetNotFound,
		Message: fmt.Sprintf("target not found: '%s'", target),
		Note:    "no target with this name is defined in the Needfile",
		Help:    "check spelling or add a target definition",
	}
}

// NewNoDefaultTargetError creates an error when no target is specified and no default exists.
func NewNoDefaultTargetError() *FormattedError {
	return &FormattedError{
		Code:    CodeNoDefaultTarget,
		Message: "no default target defined and no target specified",
		Note:    "build requires a target to build",
		Help:    "specify a target on command line or add .default: directive",
	}
}
