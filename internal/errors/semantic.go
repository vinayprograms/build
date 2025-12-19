package errors

import (
	"fmt"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
)

// Semantic Error Codes (E200-E299)
// These errors occur during semantic analysis (validation after parsing).
const (
	// E200: Reference to undefined variable
	CodeUndefinedVariable = "E200"

	// E201: Variable defined multiple times
	CodeDuplicateVariable = "E201"

	// E202: Target defined multiple times
	CodeDuplicateTarget = "E202"

	// E203: Environment defined multiple times
	CodeDuplicateEnvironment = "E203"

	// E204: Circular dependency in target graph
	CodeCircularDependency = "E204"

	// E205: Capture name conflicts with defined variable or automatic
	CodeCaptureConflict = "E205"

	// E206: Capture in dependency not defined in target pattern
	CodeCaptureMismatch = "E206"

	// E207: Automatic variable used outside recipe/block scope
	CodeAutomaticOutsideRecipe = "E207"

	// E208: Automatic variable used in target pattern
	CodeAutomaticInPattern = "E208"
)

// NewUndefinedVariableError creates an error for undefined variable references.
func NewUndefinedVariableError(name string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeUndefinedVariable,
		Message:  fmt.Sprintf("undefined variable '%s'", name),
		Location: loc,
		Help:     fmt.Sprintf("define the variable before use: %s = value", name),
	}
}

// NewDuplicateVariableError creates an error for duplicate variable definitions.
func NewDuplicateVariableError(name string, first, second ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeDuplicateVariable,
		Message:  fmt.Sprintf("duplicate variable '%s'", name),
		Location: second,
		Note:     fmt.Sprintf("'%s' was first defined at %s", name, first.String()),
	}
}

// NewDuplicateTargetError creates an error for duplicate target definitions.
func NewDuplicateTargetError(name string, first, second ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeDuplicateTarget,
		Message:  fmt.Sprintf("duplicate target '%s'", name),
		Location: second,
		Note:     fmt.Sprintf("target was first defined at %s", first.String()),
	}
}

// NewDuplicateEnvironmentError creates an error for duplicate environment definitions.
func NewDuplicateEnvironmentError(name string, first, second ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeDuplicateEnvironment,
		Message:  fmt.Sprintf("duplicate environment '%s'", name),
		Location: second,
		Note:     fmt.Sprintf("environment was first defined at %s", first.String()),
	}
}

// NewCircularDependencyError creates an error for circular dependencies.
func NewCircularDependencyError(cycle []string) *FormattedError {
	cyclePath := strings.Join(cycle, " -> ")
	return &FormattedError{
		Code:    CodeCircularDependency,
		Message: fmt.Sprintf("circular dependency detected: %s", cyclePath),
		Note:    "each target must not depend on itself directly or indirectly",
		Help:    "review the dependency chain and remove the cycle",
	}
}

// NewCaptureConflictError creates an error when capture conflicts with variable/automatic.
func NewCaptureConflictError(name, conflictType string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeCaptureConflict,
		Message:  fmt.Sprintf("capture '{%s}' conflicts with %s of the same name", name, conflictType),
		Location: loc,
		Help:     fmt.Sprintf("use a different name for the capture, or remove the %s definition", conflictType),
	}
}

// NewCaptureMismatchError creates an error when capture in dependency is not in target.
func NewCaptureMismatchError(name string, loc, targetLoc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeCaptureMismatch,
		Message:  fmt.Sprintf("capture '{%s}' in dependency not defined in target pattern", name),
		Location: loc,
		Note:     fmt.Sprintf("target pattern at %s does not define capture '{%s}'", targetLoc.String(), name),
		Help:     fmt.Sprintf("add '{%s}' to the target pattern, or use a defined variable", name),
	}
}

// NewAutomaticOutsideRecipeError creates an error for automatic vars outside recipe.
func NewAutomaticOutsideRecipeError(name string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeAutomaticOutsideRecipe,
		Message:  fmt.Sprintf("automatic variable '{%s}' is only valid inside recipe or block scope", name),
		Location: loc,
		Note:     "automatic variables like {target}, {deps}, {in}, {out} are only available during recipe execution",
	}
}

// NewAutomaticInPatternError creates an error for automatic vars in target patterns.
func NewAutomaticInPatternError(name string, loc ast.SourceLocation) *FormattedError {
	return &FormattedError{
		Code:     CodeAutomaticInPattern,
		Message:  fmt.Sprintf("automatic variable '{%s}' cannot be used in target pattern", name),
		Location: loc,
		Note:     fmt.Sprintf("'{%s}' is computed at runtime and cannot be used to define a target", name),
	}
}
