package semantic

import (
	"github.com/vinayprograms/build/internal/ast"
)

// Collect performs Pass 1: Symbol Collection on the given statements.
// It walks the AST and populates a symbol table with all definitions.
// Returns a CollectResult containing the symbol table and any errors (duplicate definitions).
//
// Collect handles:
//   - Variable definitions (immediate and lazy)
//   - Target definitions (file, phony, directory, pattern)
//   - Environment definitions (named and default)
//   - Variables inside conditionals (tracked separately)
func Collect(stmts []ast.Statement) *CollectResult {
	c := &collector{
		st:     NewSymbolTable(),
		errors: make([]error, 0),
	}

	for _, stmt := range stmts {
		c.collectStatement(stmt)
	}

	return &CollectResult{
		SymbolTable: c.st,
		Errors:      c.errors,
	}
}

// collector holds the state for symbol collection.
type collector struct {
	st     *SymbolTable
	errors []error
}

// collectStatement processes a single statement and adds its symbols to the table.
func (c *collector) collectStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.Variable:
		c.collectVariable(s)
	case *ast.Target:
		c.collectTarget(s)
	case *ast.Environment:
		c.collectEnvironment(s)
	case *ast.Conditional:
		c.collectConditional(s)
	case *ast.Directive:
		c.collectDirective(s)
	case *ast.Comment:
		// Comments are ignored
	case *ast.Blank:
		// Blank lines are ignored
	}
}

// collectVariable adds a variable definition to the symbol table.
func (c *collector) collectVariable(v *ast.Variable) {
	if err := c.st.AddVariable(v); err != nil {
		c.errors = append(c.errors, err)
	}
}

// collectTarget adds a target definition to the symbol table.
func (c *collector) collectTarget(t *ast.Target) {
	if err := c.st.AddTarget(t); err != nil {
		c.errors = append(c.errors, err)
	}
}

// collectEnvironment adds an environment definition to the symbol table.
func (c *collector) collectEnvironment(e *ast.Environment) {
	if err := c.st.AddEnvironment(e); err != nil {
		c.errors = append(c.errors, err)
	}
}

// collectDirective adds a global directive to the symbol table.
func (c *collector) collectDirective(d *ast.Directive) {
	if err := c.st.AddDirective(d); err != nil {
		c.errors = append(c.errors, err)
	}
}

// collectConditional handles symbol collection for conditional blocks.
// Variables defined in different branches of the same conditional are not
// duplicates of each other (only one branch executes at runtime).
// All variable definitions are tracked in ConditionalVars for later evaluation.
func (c *collector) collectConditional(cond *ast.Conditional) {
	// Track variables defined in this conditional to avoid adding to
	// Variables map multiple times (across branches).
	conditionalVars := make(map[string]bool)

	// Collect from if branch
	c.collectConditionalBranch(cond.IfBranch.Body, cond, "if", -1, conditionalVars)

	// Collect from elif branches
	for i, elif := range cond.ElifBranches {
		c.collectConditionalBranch(elif.Body, cond, "elif", i, conditionalVars)
	}

	// Collect from else branch
	if cond.ElseBody != nil {
		c.collectConditionalBranch(cond.ElseBody, cond, "else", -1, conditionalVars)
	}
}

// collectConditionalBranch collects symbols from a conditional branch.
// For variables, it:
//  1. Always tracks in ConditionalVars (for runtime evaluation to select correct value)
//  2. Adds to Variables map only once per conditional (for reference validation)
func (c *collector) collectConditionalBranch(
	stmts []ast.Statement,
	cond *ast.Conditional,
	branchType string,
	branchIndex int,
	conditionalVars map[string]bool,
) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.Variable:
			// Always track in ConditionalVars for runtime
			def := &ConditionalVarDef{
				Variable:    s,
				Conditional: cond,
				BranchType:  branchType,
				BranchIndex: branchIndex,
			}
			c.st.AddConditionalVariable(def)

			// Add to Variables map only once per conditional
			// This allows reference validation without false duplicate errors
			// between branches of the same conditional.
			if !conditionalVars[s.Name] {
				// Don't report error if already defined - the runtime will
				// resolve which definition to use based on condition evaluation
				_ = c.st.AddVariable(s)
				conditionalVars[s.Name] = true
			}

		case *ast.Target:
			// Targets in conditionals are less common but allowed
			if err := c.st.AddTarget(s); err != nil {
				c.errors = append(c.errors, err)
			}

		case *ast.Environment:
			// Environments in conditionals are less common but allowed
			if err := c.st.AddEnvironment(s); err != nil {
				c.errors = append(c.errors, err)
			}

		case *ast.Conditional:
			// Nested conditional
			c.collectConditional(s)
		}
	}
}

// CollectResult holds the result of symbol collection.
// This is useful for returning both the symbol table and errors.
type CollectResult struct {
	SymbolTable *SymbolTable
	Errors      []error
}

// HasErrors returns true if any errors were encountered during collection.
func (r *CollectResult) HasErrors() bool {
	return len(r.Errors) > 0
}
