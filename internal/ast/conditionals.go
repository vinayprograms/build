package ast

// ----------------------------------------------------------------------------
// Conditional Types
// ----------------------------------------------------------------------------

// Condition is the interface for conditional expressions (if/ifdef/ifndef).
//
// The conditionNode() marker method is unexported to prevent external packages
// from implementing this interface, ensuring a closed set of condition types.
//
// Implementers: EqualsCondition, NotEqualsCondition, DefinedCondition, NotDefinedCondition
type Condition interface {
	conditionNode()
}

// EqualsCondition represents a == comparison.
type EqualsCondition struct {
	Left  *Value
	Right *Value
}

func (c *EqualsCondition) conditionNode() {}

// NotEqualsCondition represents a != comparison.
type NotEqualsCondition struct {
	Left  *Value
	Right *Value
}

func (c *NotEqualsCondition) conditionNode() {}

// DefinedCondition represents an ifdef check.
type DefinedCondition struct {
	Name string
}

func (c *DefinedCondition) conditionNode() {}

// NotDefinedCondition represents an ifndef check.
type NotDefinedCondition struct {
	Name string
}

func (c *NotDefinedCondition) conditionNode() {}

// ConditionalBranch represents a single branch in a conditional.
type ConditionalBranch struct {
	Condition Condition
	Body      []Statement
}

// Conditional represents an if/elif/else/end block.
type Conditional struct {
	IfBranch     ConditionalBranch
	ElifBranches []ConditionalBranch
	ElseBody     []Statement // nil if no else clause
	Location     SourceLocation
}

func (c *Conditional) statementNode() {}
