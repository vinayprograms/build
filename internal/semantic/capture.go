// Package semantic provides semantic analysis for Buildfiles.
//
// This file implements Pass 2: Capture Validation.
// It resolves BraceExpr nodes in target patterns to either:
//   - Capture: undefined name that will be matched at build time
//   - Interpolation: reference to a defined variable
//
// It also validates:
//   - Automatic variables cannot be used as captures
//   - Captures in dependencies must also appear in the target pattern
package semantic

import (
	"fmt"

	"github.com/vinayprograms/build/internal/ast"
)

// CaptureInfo holds information about captures in a target pattern.
type CaptureInfo struct {
	Names []string // Unique capture names in order of first appearance
}

// InterpolationInfo holds information about interpolations in a target pattern.
type InterpolationInfo struct {
	Names []string // Variable names used as interpolations
}

// CaptureResult holds the result of capture validation (Pass 2).
type CaptureResult struct {
	// Captures maps each target to its capture information.
	// Only targets with captures are included.
	Captures map[*ast.Target]*CaptureInfo

	// Interpolations maps each target to its interpolation information.
	// Only targets with interpolations in patterns are included.
	Interpolations map[*ast.Target]*InterpolationInfo

	// Errors holds any validation errors encountered.
	Errors []error
}

// HasErrors returns true if any validation errors were found.
func (r *CaptureResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// AutomaticInPatternError is returned when an automatic variable is used in a target pattern.
// Automatic variables (target, deps, in, out, stem, etc.) are only available during
// recipe execution and cannot be used as capture patterns.
type AutomaticInPatternError struct {
	Name     string             // The automatic variable name
	Location ast.SourceLocation // Location of the usage
}

func (e *AutomaticInPatternError) Error() string {
	return fmt.Sprintf("automatic variable '%s' cannot be used as capture in target pattern at %s",
		e.Name, e.Location.String())
}

// CaptureMismatchError is returned when a capture appears in dependencies but not in the target.
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

// ValidateCaptures performs Pass 2: Capture Validation on the symbol table.
// It resolves BraceExpr nodes in target patterns and validates capture usage.
//
// Resolution rules:
//  1. If name is a defined variable (user, built-in, or conditional) → Interpolation
//  2. If name is an automatic variable → Error
//  3. Otherwise → Capture
//
// Consistency rules:
//   - Captures in dependencies must also appear in the target pattern
//   - Captures in target can have literal dependencies (no matching capture required)
func ValidateCaptures(st *SymbolTable) *CaptureResult {
	v := &captureValidator{
		st:             st,
		captures:       make(map[*ast.Target]*CaptureInfo),
		interpolations: make(map[*ast.Target]*InterpolationInfo),
		errors:         make([]error, 0),
	}

	for _, target := range st.Targets {
		v.validateTarget(target)
	}

	return &CaptureResult{
		Captures:       v.captures,
		Interpolations: v.interpolations,
		Errors:         v.errors,
	}
}

// captureValidator holds state during capture validation.
type captureValidator struct {
	st             *SymbolTable
	captures       map[*ast.Target]*CaptureInfo
	interpolations map[*ast.Target]*InterpolationInfo
	errors         []error
}

// validateTarget validates captures and interpolations for a single target.
func (v *captureValidator) validateTarget(target *ast.Target) {
	// First, collect and classify BraceExprs in the target pattern
	targetCaptures := make(map[string]bool)       // capture names in target
	targetInterpolations := make(map[string]bool) // interpolation names in target
	captureOrder := make([]string, 0)             // preserve order
	interpOrder := make([]string, 0)              // preserve order

	for _, seg := range target.Pattern.Segments {
		if be, ok := seg.(*ast.BraceExpr); ok {
			kind, err := v.classifyBraceExpr(be)
			if err != nil {
				v.errors = append(v.errors, err)
				continue
			}

			switch kind {
			case braceCapture:
				if !targetCaptures[be.Identifier] {
					targetCaptures[be.Identifier] = true
					captureOrder = append(captureOrder, be.Identifier)
				}
			case braceInterpolation:
				if !targetInterpolations[be.Identifier] {
					targetInterpolations[be.Identifier] = true
					interpOrder = append(interpOrder, be.Identifier)
				}
			}
		}
	}

	// Now validate dependencies - captures there must be defined in target
	for _, dep := range target.Dependencies {
		for _, seg := range dep.Segments {
			if be, ok := seg.(*ast.BraceExpr); ok {
				kind, err := v.classifyBraceExpr(be)
				if err != nil {
					v.errors = append(v.errors, err)
					continue
				}

				switch kind {
				case braceCapture:
					// Capture in dependency must exist in target
					if !targetCaptures[be.Identifier] {
						v.errors = append(v.errors, &CaptureMismatchError{
							Name:      be.Identifier,
							InTarget:  false,
							Location:  be.Location,
							TargetLoc: target.Location,
						})
					}
				case braceInterpolation:
					// Interpolations in dependencies are fine
					// Track them if not already tracked from target
					if !targetInterpolations[be.Identifier] {
						targetInterpolations[be.Identifier] = true
						interpOrder = append(interpOrder, be.Identifier)
					}
				}
			}
		}
	}

	// Store results
	if len(captureOrder) > 0 {
		v.captures[target] = &CaptureInfo{Names: captureOrder}
	}
	if len(interpOrder) > 0 {
		v.interpolations[target] = &InterpolationInfo{Names: interpOrder}
	}
}

// braceExprKind represents the kind of a BraceExpr after classification.
type braceExprKind int

const (
	braceCapture       braceExprKind = iota // Undefined name → capture
	braceInterpolation                      // Defined variable → interpolation
)

// classifyBraceExpr determines whether a BraceExpr is a capture or interpolation.
// Returns an error if the name is an automatic variable.
func (v *captureValidator) classifyBraceExpr(be *ast.BraceExpr) (braceExprKind, error) {
	name := be.Identifier

	// Check for automatic variables - error
	if v.st.IsAutomatic(name) {
		return 0, &AutomaticInPatternError{
			Name:     name,
			Location: be.Location,
		}
	}

	// Check for defined variables (user, built-in, conditional) - interpolation
	if v.st.IsDefined(name) {
		return braceInterpolation, nil
	}

	// Otherwise it's a capture
	return braceCapture, nil
}
