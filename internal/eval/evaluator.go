package eval

import (
	"fmt"
	"io"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
)

// Evaluator evaluates AST values and expressions.
type Evaluator struct {
	ctx           *Context
	verboseOutput io.Writer // Optional output for verbose mode
}

// NewEvaluator creates a new evaluator with the given context.
func NewEvaluator(ctx *Context) *Evaluator {
	return &Evaluator{ctx: ctx}
}

// SetVerboseOutput sets the output writer for verbose mode.
// When set, variable evaluations will be printed to this writer.
func (e *Evaluator) SetVerboseOutput(w io.Writer) {
	e.verboseOutput = w
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
// For lazy variables, it evaluates on-demand and caches the result.
func (e *Evaluator) resolveVariable(name string, loc ast.SourceLocation) (string, error) {
	// Check if variable is defined (including cached lazy values)
	val, ok := e.ctx.Get(name)
	if ok {
		return val, nil
	}

	// Check if it's a lazy variable that needs evaluation
	if e.ctx.IsLazy(name) {
		return e.evaluateLazyVariable(name, loc)
	}

	return "", &UndefinedVariableError{Name: name, Location: loc}
}

// evaluateLazyVariable evaluates a lazy variable on-demand and caches the result.
func (e *Evaluator) evaluateLazyVariable(name string, loc ast.SourceLocation) (string, error) {
	// Get the lazy variable's AST value
	lazyValue, ok := e.ctx.GetLazyValue(name)
	if !ok {
		return "", &UndefinedVariableError{Name: name, Location: loc}
	}

	// Evaluate the value
	result, err := e.EvaluateValue(lazyValue)
	if err != nil {
		return "", err
	}

	// Cache the result for future references
	e.ctx.CacheLazyResult(name, result)

	return result, nil
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
