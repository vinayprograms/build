// Package parser implements the parsing phase for Needfile source code.
// It transforms a token stream from the lexer into an Abstract Syntax Tree (AST).
package parser

// Scope represents the current parsing context.
// Different directives are valid at different scopes.
type Scope int

const (
	// ScopeGlobal is the top-level scope for global directives, variables, targets.
	ScopeGlobal Scope = iota
	// ScopeEnvironment is inside an .environment: block.
	ScopeEnvironment
	// ScopeRecipe is inside a target's recipe section.
	ScopeRecipe
	// ScopeBlock is inside a block: within a recipe.
	ScopeBlock
)

// String returns the string representation of the scope.
func (s Scope) String() string {
	switch s {
	case ScopeGlobal:
		return "GLOBAL"
	case ScopeEnvironment:
		return "ENVIRONMENT"
	case ScopeRecipe:
		return "RECIPE"
	case ScopeBlock:
		return "BLOCK"
	default:
		return "UNKNOWN"
	}
}

// ScopeStack tracks nested scopes during parsing.
// The parser uses this to validate directive placement.
type ScopeStack struct {
	stack []Scope
}

// NewScopeStack creates a new scope stack initialized at global scope.
func NewScopeStack() *ScopeStack {
	return &ScopeStack{
		stack: []Scope{ScopeGlobal},
	}
}

// Current returns the current (topmost) scope.
func (ss *ScopeStack) Current() Scope {
	if len(ss.stack) == 0 {
		return ScopeGlobal
	}
	return ss.stack[len(ss.stack)-1]
}

// Depth returns the current nesting depth.
func (ss *ScopeStack) Depth() int {
	return len(ss.stack)
}

// Push adds a new scope to the stack.
func (ss *ScopeStack) Push(scope Scope) {
	ss.stack = append(ss.stack, scope)
}

// Pop removes and returns the topmost scope.
// If at global scope, returns ScopeGlobal without modifying the stack.
func (ss *ScopeStack) Pop() Scope {
	if len(ss.stack) <= 1 {
		return ScopeGlobal
	}
	top := ss.stack[len(ss.stack)-1]
	ss.stack = ss.stack[:len(ss.stack)-1]
	return top
}

// IsIn returns true if the given scope is anywhere in the stack.
// This is useful for checking if we're "inside" a particular scope.
func (ss *ScopeStack) IsIn(scope Scope) bool {
	for _, s := range ss.stack {
		if s == scope {
			return true
		}
	}
	return false
}

// Reset clears the stack and returns to global scope.
func (ss *ScopeStack) Reset() {
	ss.stack = []Scope{ScopeGlobal}
}
