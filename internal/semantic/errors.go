// Package semantic provides semantic analysis for Buildfiles.
//
// This file defines all semantic error types for the build tool.
// These errors are generated during semantic analysis passes:
//   - Pass 1 (Symbol Collection): DuplicateDefinitionError
//   - Pass 2 (Capture Validation): AutomaticInPatternError, CaptureMismatchError
//   - Pass 3 (Reference Validation): UndefinedVariableError, AutomaticOutsideRecipeError
//   - Pass 4 (Dependency Graph): CircularDependencyError
package semantic

import (
	"fmt"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
)

// ----------------------------------------------------------------------------
// Pass 1: Symbol Collection Errors
// ----------------------------------------------------------------------------

// DuplicateDefinitionError is returned when a symbol (variable, target, or environment)
// is defined multiple times in the same Buildfile or included files.
type DuplicateDefinitionError struct {
	Kind   string             // "variable", "target", or "environment"
	Name   string             // The name that was duplicated
	First  ast.SourceLocation // Location of first definition
	Second ast.SourceLocation // Location of duplicate definition
}

func (e *DuplicateDefinitionError) Error() string {
	return fmt.Sprintf("duplicate %s '%s': first defined at %s, redefined at %s",
		e.Kind, e.Name, e.First.String(), e.Second.String())
}

// ----------------------------------------------------------------------------
// Pass 2: Capture Validation Errors
// ----------------------------------------------------------------------------

// AutomaticInPatternError is returned when an automatic variable is used in a target pattern.
// Automatic variables (target, deps, in, out, stem, target.dir, target.file) are only
// available during recipe execution and cannot be used as capture patterns.
//
// Example error case:
//
//	build/{target}.o: src/main.c  # Error: 'target' is automatic
type AutomaticInPatternError struct {
	Name     string             // The automatic variable name
	Location ast.SourceLocation // Location of the usage
}

func (e *AutomaticInPatternError) Error() string {
	return fmt.Sprintf("automatic variable '%s' cannot be used as capture in target pattern at %s",
		e.Name, e.Location.String())
}

// CaptureMismatchError is returned when a capture appears in dependencies but not in the target.
// Captures must be defined in the target pattern before they can be used in dependencies.
//
// Example error case:
//
//	build/app: src/{name}.c  # Error: 'name' not defined in target
type CaptureMismatchError struct {
	Name      string             // The capture name
	InTarget  bool               // true if capture is in target but not deps (this is allowed)
	Location  ast.SourceLocation // Location of the problematic usage
	TargetLoc ast.SourceLocation // Location of the target definition
}

func (e *CaptureMismatchError) Error() string {
	return fmt.Sprintf("capture '{%s}' in dependency but not defined in target pattern at %s (target at %s)",
		e.Name, e.Location.String(), e.TargetLoc.String())
}

// ----------------------------------------------------------------------------
// Pass 3: Reference Validation Errors
// ----------------------------------------------------------------------------

// UndefinedVariableError is returned when a reference to an undefined variable is found.
// All variable references must point to:
//   - A user-defined variable (immediate or lazy)
//   - A built-in variable (os, arch)
//   - An automatic variable (only in recipe/block scope)
//   - A capture variable (only in the defining recipe)
//
// Example error case:
//
//	output = {undefined_var}  # Error: 'undefined_var' not defined
type UndefinedVariableError struct {
	Name     string             // The undefined variable name
	Location ast.SourceLocation // Location of the reference
}

func (e *UndefinedVariableError) Error() string {
	return fmt.Sprintf("undefined variable '%s' at %s", e.Name, e.Location.String())
}

// AutomaticOutsideRecipeError is returned when an automatic variable is used outside a recipe.
// Automatic variables are only computed during recipe execution and are not available
// in global variable definitions, directive values, or other non-recipe contexts.
//
// Example error case:
//
//	output = {target}  # Error: 'target' only valid in recipes
type AutomaticOutsideRecipeError struct {
	Name     string             // The automatic variable name
	Location ast.SourceLocation // Location of the reference
}

func (e *AutomaticOutsideRecipeError) Error() string {
	return fmt.Sprintf("automatic variable '%s' is only valid inside recipe or block scope at %s",
		e.Name, e.Location.String())
}

// ----------------------------------------------------------------------------
// Pass 4: Dependency Graph Errors
// ----------------------------------------------------------------------------

// CircularDependencyError is returned when a circular dependency is detected
// in the target dependency graph.
//
// Example error case:
//
//	a: b
//	b: a  # Error: circular dependency a -> b -> a
type CircularDependencyError struct {
	// Cycle contains the targets in the cycle, e.g., ["a", "b", "c", "a"]
	// The first and last elements are the same, showing where the cycle closes.
	Cycle []string
}

func (e *CircularDependencyError) Error() string {
	return fmt.Sprintf("circular dependency detected: %s", strings.Join(e.Cycle, " -> "))
}
