package eval

import (
	"github.com/vinayprograms/build/internal/ast"
)

// EvaluateVariables evaluates all immediate variables in definition order.
// Lazy variables are stored for on-demand evaluation.
//
// Evaluation rules:
//  1. Variables are evaluated in definition order
//  2. Forward references (referencing a later immediate variable) cause an error
//  3. Lazy variables are stored unevaluated for on-demand evaluation
//  4. Built-in variables (os, arch) are always available
//  5. Non-variable statements are skipped
func (e *Evaluator) EvaluateVariables(stmts []ast.Statement) error {
	for _, stmt := range stmts {
		if err := e.evaluateStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

// evaluateStatement evaluates a single statement if it's a variable.
func (e *Evaluator) evaluateStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.Variable:
		return e.evaluateVariable(s)
	case *ast.Conditional:
		// TODO: Handle conditionals
		return nil
	default:
		// Skip non-variable statements (comments, blanks, directives, targets, etc.)
		return nil
	}
}

// evaluateVariable evaluates a single variable definition.
func (e *Evaluator) evaluateVariable(v *ast.Variable) error {
	if v.Lazy {
		// Store lazy variables with their AST value for on-demand evaluation
		e.ctx.SetLazyValue(v.Name, v.Value)
		return nil
	}

	// Evaluate the value
	result, err := e.EvaluateValue(v.Value)
	if err != nil {
		return err
	}

	// Store the evaluated result
	e.ctx.Set(v.Name, result)
	return nil
}
