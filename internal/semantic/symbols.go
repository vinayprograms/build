// Package semantic provides semantic analysis for Buildfiles.
//
// Semantic analysis validates the AST produced by the parser and resolves
// context-sensitive constructs like captures vs interpolations in target patterns.
//
// The semantic analyzer performs several passes:
//   - Pass 1: Symbol Collection - collect all definitions
//   - Pass 2: Capture Validation - resolve {name} in target patterns
//   - Pass 3: Reference Validation - verify all references
//   - Pass 4: Dependency Graph Validation - detect cycles
package semantic

import (
	"fmt"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
)

// SymbolTable holds all defined symbols in a Buildfile.
// It tracks user-defined variables, targets, and environments,
// as well as automatic and built-in variables.
type SymbolTable struct {
	// Variables maps variable names to their definitions.
	// For unconditional variables, this is the single definition.
	// For conditional variables, this is the first definition encountered.
	Variables map[string]*ast.Variable

	// ConditionalVars tracks variables that are defined within conditionals.
	// Maps variable name to all conditionals that define it.
	// This allows runtime to evaluate conditions and pick the right value.
	ConditionalVars map[string][]*ConditionalVarDef

	// Targets holds all target definitions.
	// We use a slice because pattern targets may overlap and we need
	// to preserve definition order.
	Targets []*ast.Target

	// targetPatterns tracks seen target pattern strings for duplicate detection.
	targetPatterns map[string]ast.SourceLocation

	// Environments maps environment names to their definitions.
	// The empty string key is used for the default (unnamed) environment.
	Environments map[string]*ast.Environment

	// automatic is the set of automatic variable names.
	// These are only valid inside recipe/block scope.
	automatic map[string]bool

	// builtin is the set of built-in variable names.
	// These are always available (os, arch).
	builtin map[string]bool
}

// ConditionalVarDef represents a variable definition within a conditional.
// It tracks the conditional context so runtime can evaluate the condition.
type ConditionalVarDef struct {
	Variable    *ast.Variable    // The variable definition
	Conditional *ast.Conditional // The containing conditional
	BranchType  string           // "if", "elif", or "else"
	BranchIndex int              // For elif, the index (0-based); -1 for if/else
}

// NewSymbolTable creates an initialized symbol table with automatic
// and built-in variables populated.
func NewSymbolTable() *SymbolTable {
	st := &SymbolTable{
		Variables:       make(map[string]*ast.Variable),
		ConditionalVars: make(map[string][]*ConditionalVarDef),
		Targets:         make([]*ast.Target, 0),
		targetPatterns:  make(map[string]ast.SourceLocation),
		Environments:    make(map[string]*ast.Environment),
		automatic:       make(map[string]bool),
		builtin:         make(map[string]bool),
	}

	// Populate automatic variables per DESIGN.md Section 3.3.4
	automaticVars := []string{
		"target",      // Target file path
		"deps",        // All dependencies (space-separated)
		"in",          // First dependency
		"out",         // Alias for target
		"stem",        // Pattern match stem (for pattern targets)
		"target.dir",  // Directory part of target
		"target.file", // Filename part of target
	}
	for _, v := range automaticVars {
		st.automatic[v] = true
	}

	// Populate built-in variables
	builtinVars := []string{
		"os",   // Operating system
		"arch", // Architecture
	}
	for _, v := range builtinVars {
		st.builtin[v] = true
	}

	return st
}

// AddVariable adds a variable definition to the symbol table.
// Returns an error if a variable with the same name already exists.
func (st *SymbolTable) AddVariable(v *ast.Variable) error {
	if existing, ok := st.Variables[v.Name]; ok {
		return &DuplicateDefinitionError{
			Kind:   "variable",
			Name:   v.Name,
			First:  existing.Location,
			Second: v.Location,
		}
	}
	st.Variables[v.Name] = v
	return nil
}

// AddConditionalVariable adds a variable definition that appears within a conditional.
// Unlike AddVariable, this allows multiple definitions for the same variable name
// (one per conditional branch), since only one branch will execute at runtime.
func (st *SymbolTable) AddConditionalVariable(def *ConditionalVarDef) {
	st.ConditionalVars[def.Variable.Name] = append(
		st.ConditionalVars[def.Variable.Name], def)
}

// IsConditionalVar returns true if the variable has conditional definitions.
func (st *SymbolTable) IsConditionalVar(name string) bool {
	_, ok := st.ConditionalVars[name]
	return ok
}

// GetConditionalVarDefs returns all conditional definitions for a variable.
func (st *SymbolTable) GetConditionalVarDefs(name string) []*ConditionalVarDef {
	return st.ConditionalVars[name]
}

// AddTarget adds a target definition to the symbol table.
// Returns an error if an exact duplicate pattern already exists.
func (st *SymbolTable) AddTarget(t *ast.Target) error {
	patternStr := PatternString(&t.Pattern)

	if firstLoc, ok := st.targetPatterns[patternStr]; ok {
		return &DuplicateDefinitionError{
			Kind:   "target",
			Name:   patternStr,
			First:  firstLoc,
			Second: t.Location,
		}
	}

	st.targetPatterns[patternStr] = t.Location
	st.Targets = append(st.Targets, t)
	return nil
}

// AddEnvironment adds an environment definition to the symbol table.
// Returns an error if an environment with the same name already exists.
func (st *SymbolTable) AddEnvironment(e *ast.Environment) error {
	key := ""
	if e.Name != nil {
		key = *e.Name
	}

	if existing, ok := st.Environments[key]; ok {
		name := key
		if name == "" {
			name = "(default)"
		}
		return &DuplicateDefinitionError{
			Kind:   "environment",
			Name:   name,
			First:  existing.Location,
			Second: e.Location,
		}
	}

	st.Environments[key] = e
	return nil
}

// IsAutomatic returns true if the name is an automatic variable.
func (st *SymbolTable) IsAutomatic(name string) bool {
	return st.automatic[name]
}

// IsBuiltin returns true if the name is a built-in variable.
func (st *SymbolTable) IsBuiltin(name string) bool {
	return st.builtin[name]
}

// LookupVariable returns the variable with the given name, or nil if not found.
func (st *SymbolTable) LookupVariable(name string) *ast.Variable {
	return st.Variables[name]
}

// IsDefined returns true if the name is defined as a user variable,
// conditional variable, automatic variable, or built-in variable.
func (st *SymbolTable) IsDefined(name string) bool {
	if _, ok := st.Variables[name]; ok {
		return true
	}
	if _, ok := st.ConditionalVars[name]; ok {
		return true
	}
	if st.automatic[name] {
		return true
	}
	if st.builtin[name] {
		return true
	}
	return false
}

// PatternString converts a target pattern to a string representation
// for duplicate detection and display purposes.
// Note: Phony targets are stored without the @ prefix; use IsPhony flag to check.
func PatternString(p *ast.TargetPattern) string {
	var sb strings.Builder

	for _, seg := range p.Segments {
		switch s := seg.(type) {
		case *ast.LiteralSegment:
			sb.WriteString(s.Text)
		case *ast.BraceExpr:
			sb.WriteString("{")
			sb.WriteString(s.Identifier)
			sb.WriteString("}")
		}
	}

	return sb.String()
}

// DuplicateDefinitionError is returned when a duplicate definition is detected.
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
