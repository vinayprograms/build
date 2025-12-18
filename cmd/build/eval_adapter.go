package main

import (
	"sort"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/eval"
)

// evalContextAdapter wraps eval.Context to implement the EvalContext interface.
type evalContextAdapter struct {
	ctx *eval.Context
}

func (eca *evalContextAdapter) Get(name string) (string, bool) {
	return eca.ctx.Get(name)
}

func (eca *evalContextAdapter) Set(name, value string) {
	eca.ctx.Set(name, value)
}

func (eca *evalContextAdapter) IsDefined(name string) bool {
	return eca.ctx.IsDefined(name)
}

func (eca *evalContextAdapter) IsLazy(name string) bool {
	return eca.ctx.IsLazy(name)
}

func (eca *evalContextAdapter) Variables() map[string]string {
	return eca.ctx.Variables()
}

func (eca *evalContextAdapter) LazyVariables() map[string]string {
	return eca.ctx.LazyVariables()
}

// evalResultAdapter wraps eval.Context and evaluation errors to implement EvalResult.
type evalResultAdapter struct {
	ctx       *eval.Context
	errors    []error
	evalNames []string // Names of evaluated variables (sorted)
	lazyNames []string // Names of lazy variables (sorted)
}

func (era *evalResultAdapter) Context() EvalContext {
	return &evalContextAdapter{ctx: era.ctx}
}

func (era *evalResultAdapter) HasErrors() bool {
	return len(era.errors) > 0
}

func (era *evalResultAdapter) ErrorCount() int {
	return len(era.errors)
}

func (era *evalResultAdapter) GetError(i int) error {
	if i < 0 || i >= len(era.errors) {
		return nil
	}
	return era.errors[i]
}

func (era *evalResultAdapter) Errors() []error {
	return era.errors
}

func (era *evalResultAdapter) EvaluatedCount() int {
	return len(era.evalNames)
}

func (era *evalResultAdapter) EvaluatedName(i int) string {
	if i < 0 || i >= len(era.evalNames) {
		return ""
	}
	return era.evalNames[i]
}

func (era *evalResultAdapter) EvaluatedValue(i int) string {
	if i < 0 || i >= len(era.evalNames) {
		return ""
	}
	name := era.evalNames[i]
	val, _ := era.ctx.Get(name)
	return val
}

func (era *evalResultAdapter) LazyCount() int {
	return len(era.lazyNames)
}

func (era *evalResultAdapter) LazyName(i int) string {
	if i < 0 || i >= len(era.lazyNames) {
		return ""
	}
	return era.lazyNames[i]
}

// EvaluateVariables evaluates all variables from parsed statements and returns
// an EvalResult containing the evaluation context and any errors.
func EvaluateVariables(result BuildfileResult) EvalResult {
	ctx := eval.NewContext()
	evaluator := eval.NewEvaluator(ctx)

	// Convert Statement interface slice to ast.Statement slice
	astStmts := make([]ast.Statement, len(result.Statements()))
	for i, stmt := range result.Statements() {
		if sa, ok := stmt.(statementAdapter); ok {
			astStmts[i] = sa.s
		}
	}

	// Evaluate all variables
	err := evaluator.EvaluateVariables(astStmts)

	var errors []error
	if err != nil {
		errors = []error{err}
	}

	// Collect evaluated variable names
	vars := ctx.Variables()
	evalNames := make([]string, 0, len(vars))
	for name := range vars {
		// Skip built-in variables
		if name != "os" && name != "arch" {
			evalNames = append(evalNames, name)
		}
	}
	sort.Strings(evalNames)

	// Collect lazy variable names
	lazyVars := ctx.LazyVariables()
	lazyNames := make([]string, 0, len(lazyVars))
	for name := range lazyVars {
		lazyNames = append(lazyNames, name)
	}
	sort.Strings(lazyNames)

	return &evalResultAdapter{
		ctx:       ctx,
		errors:    errors,
		evalNames: evalNames,
		lazyNames: lazyNames,
	}
}
