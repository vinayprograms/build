package eval

import (
	"fmt"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
)

// Evaluator evaluates AST values and expressions.
type Evaluator struct {
	ctx *Context
}

// NewEvaluator creates a new evaluator with the given context.
func NewEvaluator(ctx *Context) *Evaluator {
	return &Evaluator{ctx: ctx}
}

// Context returns the evaluation context.
func (e *Evaluator) Context() *Context {
	return e.ctx
}

// EvaluateValue evaluates an AST Value and returns the resulting string.
// It handles literal values, interpolations, and function calls.
func (e *Evaluator) EvaluateValue(val *ast.Value) (string, error) {
	if val == nil {
		return "", nil
	}

	var result strings.Builder

	for _, part := range val.Parts {
		switch p := part.(type) {
		case *ast.LiteralValue:
			result.WriteString(p.Text)

		case *ast.Interpolation:
			resolved, err := e.resolveVariable(p.Name, p.Location)
			if err != nil {
				return "", err
			}
			result.WriteString(resolved)

		case *ast.FunctionCall:
			evaluated, err := e.evaluateFunction(p)
			if err != nil {
				return "", err
			}
			result.WriteString(evaluated)

		default:
			// Unknown part type - skip
		}
	}

	return result.String(), nil
}

// resolveVariable resolves a variable reference to its value.
func (e *Evaluator) resolveVariable(name string, loc ast.SourceLocation) (string, error) {
	// Check if variable is defined
	val, ok := e.ctx.Get(name)
	if ok {
		return val, nil
	}

	// Check if it's a lazy variable that needs evaluation
	if e.ctx.IsLazy(name) {
		// TODO: Evaluate lazy variable on demand
		// For now, return an error - lazy evaluation will be implemented later
		return "", &UndefinedVariableError{Name: name, Location: loc}
	}

	return "", &UndefinedVariableError{Name: name, Location: loc}
}

// evaluateFunction evaluates a function call.
func (e *Evaluator) evaluateFunction(call *ast.FunctionCall) (string, error) {
	// TODO: Implement function evaluation
	// For now, return empty string
	return "", nil
}

// ----------------------------------------------------------------------------
// Error Types
// ----------------------------------------------------------------------------

// UndefinedVariableError is returned when a variable reference cannot be resolved.
type UndefinedVariableError struct {
	Name     string
	Location ast.SourceLocation
}

func (e *UndefinedVariableError) Error() string {
	return fmt.Sprintf("undefined variable '%s' at %s", e.Name, e.Location.String())
}
