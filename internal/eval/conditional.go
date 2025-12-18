package eval

import (
	"fmt"

	"github.com/vinayprograms/build/internal/ast"
)

// EvaluateCondition evaluates a condition expression and returns the result.
func (e *Evaluator) EvaluateCondition(cond ast.Condition) (bool, error) {
	switch c := cond.(type) {
	case *ast.EqualsCondition:
		return e.evaluateEqualsCondition(c)
	case *ast.NotEqualsCondition:
		return e.evaluateNotEqualsCondition(c)
	case *ast.DefinedCondition:
		return e.evaluateDefinedCondition(c), nil
	case *ast.NotDefinedCondition:
		return e.evaluateNotDefinedCondition(c), nil
	default:
		return false, fmt.Errorf("unknown condition type: %T", cond)
	}
}

// evaluateEqualsCondition evaluates a == comparison.
func (e *Evaluator) evaluateEqualsCondition(cond *ast.EqualsCondition) (bool, error) {
	left, err := e.EvaluateValue(cond.Left)
	if err != nil {
		return false, err
	}

	right, err := e.EvaluateValue(cond.Right)
	if err != nil {
		return false, err
	}

	return left == right, nil
}

// evaluateNotEqualsCondition evaluates a != comparison.
func (e *Evaluator) evaluateNotEqualsCondition(cond *ast.NotEqualsCondition) (bool, error) {
	left, err := e.EvaluateValue(cond.Left)
	if err != nil {
		return false, err
	}

	right, err := e.EvaluateValue(cond.Right)
	if err != nil {
		return false, err
	}

	return left != right, nil
}

// evaluateDefinedCondition checks if a variable is defined.
// This does NOT evaluate the variable - it only checks for existence.
func (e *Evaluator) evaluateDefinedCondition(cond *ast.DefinedCondition) bool {
	return e.ctx.IsDefined(cond.Name)
}

// evaluateNotDefinedCondition checks if a variable is NOT defined.
func (e *Evaluator) evaluateNotDefinedCondition(cond *ast.NotDefinedCondition) bool {
	return !e.ctx.IsDefined(cond.Name)
}

// EvaluateConditional evaluates a conditional block.
// It evaluates the condition and executes the appropriate branch's body.
func (e *Evaluator) EvaluateConditional(conditional *ast.Conditional) error {
	// Try if branch
	result, err := e.EvaluateCondition(conditional.IfBranch.Condition)
	if err != nil {
		return err
	}
	if result {
		return e.evaluateBranchBody(conditional.IfBranch.Body)
	}

	// Try elif branches
	for _, elifBranch := range conditional.ElifBranches {
		result, err := e.EvaluateCondition(elifBranch.Condition)
		if err != nil {
			return err
		}
		if result {
			return e.evaluateBranchBody(elifBranch.Body)
		}
	}

	// Try else branch
	if len(conditional.ElseBody) > 0 {
		return e.evaluateBranchBody(conditional.ElseBody)
	}

	return nil
}

// evaluateBranchBody evaluates the statements in a conditional branch.
func (e *Evaluator) evaluateBranchBody(stmts []ast.Statement) error {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.Variable:
			if err := e.evaluateVariable(s); err != nil {
				return err
			}
		case *ast.Conditional:
			if err := e.EvaluateConditional(s); err != nil {
				return err
			}
		default:
			// Skip non-variable statements in conditionals
		}
	}
	return nil
}
