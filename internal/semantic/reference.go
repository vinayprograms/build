// Package semantic provides semantic analysis for Buildfiles.
//
// This file implements Pass 3: Reference Validation.
// It validates that all variable references point to defined symbols:
//   - User-defined variables
//   - Built-in variables (os, arch)
//   - Automatic variables (target, deps, in, out, stem, target.dir, target.file) - only in recipe scope
//   - Captures - only in the recipe that defines them
package semantic

import (
	"github.com/vinayprograms/build/internal/ast"
)

// ReferenceResult holds the result of reference validation (Pass 3).
type ReferenceResult struct {
	// Errors holds any validation errors encountered.
	Errors []error
}

// HasErrors returns true if any validation errors were found.
func (r *ReferenceResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// ReferenceOption configures reference validation.
type ReferenceOption func(*referenceValidator)

// WithCaptures provides capture information from Pass 2 for reference validation.
// This allows captures to be recognized as valid references in their defining recipe.
func WithCaptures(captureResult *CaptureResult) ReferenceOption {
	return func(v *referenceValidator) {
		v.captureResult = captureResult
	}
}

// ValidateReferences performs Pass 3: Reference Validation.
// It validates that all variable references in the AST are defined.
//
// Validation rules:
//  1. Automatic variables are only valid in recipe/block scope
//  2. Captures are only valid in the recipe that defines them
//  3. All other references must be defined (user, conditional, or built-in)
//
// The opts parameter allows providing capture information from Pass 2.
func ValidateReferences(st *SymbolTable, stmts []ast.Statement, opts ...ReferenceOption) *ReferenceResult {
	v := &referenceValidator{
		st:     st,
		errors: make([]error, 0),
	}

	for _, opt := range opts {
		opt(v)
	}

	for _, stmt := range stmts {
		v.validateStatement(stmt, nil)
	}

	return &ReferenceResult{
		Errors: v.errors,
	}
}

// referenceValidator holds state during reference validation.
type referenceValidator struct {
	st            *SymbolTable
	captureResult *CaptureResult
	errors        []error
}

// isCaptureForTarget returns true if name is a capture defined in the given target.
func (v *referenceValidator) isCaptureForTarget(target *ast.Target, name string) bool {
	if v.captureResult == nil || target == nil {
		return false
	}
	captureInfo, ok := v.captureResult.Captures[target]
	if !ok {
		return false
	}
	for _, captureName := range captureInfo.Names {
		if captureName == name {
			return true
		}
	}
	return false
}

// validateStatement validates references in a statement.
// currentTarget is set when inside a target's recipe to allow capture references.
func (v *referenceValidator) validateStatement(stmt ast.Statement, currentTarget *ast.Target) {
	switch s := stmt.(type) {
	case *ast.Variable:
		v.validateValue(s.Value, false, nil)
	case *ast.Directive:
		if s.Value != nil {
			v.validateValue(s.Value, false, nil)
		}
	case *ast.Target:
		v.validateTarget(s)
	case *ast.Conditional:
		v.validateConditional(s, currentTarget)
	case *ast.Environment:
		v.validateEnvironment(s)
	case *ast.Comment, *ast.Blank:
		// No references to validate
	}
}

// validateTarget validates references in a target definition.
func (v *referenceValidator) validateTarget(t *ast.Target) {
	// Note: Target pattern BraceExprs are validated by Pass 2 (capture validation)
	// We only need to validate recipe references here

	if t.Recipe != nil {
		v.validateRecipe(t.Recipe, t)
	}
}

// validateRecipe validates references in a recipe.
func (v *referenceValidator) validateRecipe(r *ast.Recipe, target *ast.Target) {
	// Validate recipe directives - these are in recipe scope and can use captures
	if r.Directives.Shell != nil {
		v.validateValue(r.Directives.Shell, true, target)
	}

	for _, after := range r.Directives.After {
		v.validateValue(after, true, target)
	}

	if r.Directives.Autodeps != nil {
		v.validateValue(r.Directives.Autodeps, true, target)
	}

	// Validate commands
	for _, cmd := range r.Commands {
		v.validateCommand(cmd, target)
	}
}

// validateCommand validates references in a command.
func (v *referenceValidator) validateCommand(cmd ast.Command, target *ast.Target) {
	switch c := cmd.(type) {
	case *ast.LineCommand:
		for _, part := range c.Parts {
			if interp, ok := part.(*ast.CommandInterpolation); ok {
				v.validateCommandInterpolation(interp, target)
			}
		}
	case *ast.BlockCommand:
		for _, line := range c.Lines {
			for _, part := range line {
				if interp, ok := part.(*ast.CommandInterpolation); ok {
					v.validateCommandInterpolation(interp, target)
				}
			}
		}
	}
}

// validateCommandInterpolation validates a reference in a command (recipe or block scope).
func (v *referenceValidator) validateCommandInterpolation(interp *ast.CommandInterpolation, target *ast.Target) {
	name := interp.Name

	// Automatic variables are valid in recipe scope
	if v.st.IsAutomatic(name) {
		return
	}

	// Check if it's a capture for this target
	if v.isCaptureForTarget(target, name) {
		return
	}

	// Check if it's a defined variable
	if v.st.IsDefined(name) {
		return
	}

	// Undefined
	v.errors = append(v.errors, &UndefinedVariableError{
		Name:     name,
		Location: interp.Location,
	})
}

// validateValue validates references in a value (global scope).
func (v *referenceValidator) validateValue(val *ast.Value, inRecipe bool, target *ast.Target) {
	if val == nil {
		return
	}

	for _, part := range val.Parts {
		switch p := part.(type) {
		case *ast.Interpolation:
			v.validateInterpolation(p, inRecipe, target)
		case *ast.FunctionCall:
			for _, arg := range p.Args {
				v.validateValue(arg, inRecipe, target)
			}
		case *ast.LiteralValue, *ast.LiteralSegment:
			// No references to validate
		}
	}
}

// validateInterpolation validates a reference in an interpolation.
func (v *referenceValidator) validateInterpolation(interp *ast.Interpolation, inRecipe bool, target *ast.Target) {
	name := interp.Name

	// Automatic variables are only valid in recipe scope
	if v.st.IsAutomatic(name) {
		if !inRecipe {
			v.errors = append(v.errors, &AutomaticOutsideRecipeError{
				Name:     name,
				Location: interp.Location,
			})
		}
		return
	}

	// Check if it's a capture for this target (only in recipe scope)
	if inRecipe && v.isCaptureForTarget(target, name) {
		return
	}

	// Check if it's a defined variable (user, conditional, or built-in)
	if v.st.IsDefined(name) {
		return
	}

	// Undefined
	v.errors = append(v.errors, &UndefinedVariableError{
		Name:     name,
		Location: interp.Location,
	})
}

// validateConditional validates references in a conditional.
func (v *referenceValidator) validateConditional(c *ast.Conditional, currentTarget *ast.Target) {
	// Validate if branch condition
	v.validateCondition(c.IfBranch.Condition)

	// Validate if branch body
	for _, stmt := range c.IfBranch.Body {
		v.validateStatement(stmt, currentTarget)
	}

	// Validate elif branches
	for _, elifBranch := range c.ElifBranches {
		v.validateCondition(elifBranch.Condition)
		for _, stmt := range elifBranch.Body {
			v.validateStatement(stmt, currentTarget)
		}
	}

	// Validate else body
	for _, stmt := range c.ElseBody {
		v.validateStatement(stmt, currentTarget)
	}
}

// validateCondition validates references in a condition expression.
func (v *referenceValidator) validateCondition(cond ast.Condition) {
	switch c := cond.(type) {
	case *ast.EqualsCondition:
		v.validateValue(c.Left, false, nil)
		v.validateValue(c.Right, false, nil)
	case *ast.NotEqualsCondition:
		v.validateValue(c.Left, false, nil)
		v.validateValue(c.Right, false, nil)
	case *ast.DefinedCondition, *ast.NotDefinedCondition:
		// These just check if a name is defined, they don't need validation
		// The identifier itself doesn't need to be defined (that's what ifdef checks)
	}
}

// validateEnvironment validates references in an environment block.
func (v *referenceValidator) validateEnvironment(e *ast.Environment) {
	if e.Source != nil {
		v.validateValue(e.Source, false, nil)
	}
	if e.Args != nil {
		v.validateValue(e.Args, false, nil)
	}
	// Requirements don't have interpolations in their names
}
