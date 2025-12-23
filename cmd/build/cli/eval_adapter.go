package cli

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

// underlyingContext returns the underlying eval.Context.
func (eca *evalContextAdapter) underlyingContext() *eval.Context {
	return eca.ctx
}

// ----------------------------------------------------------------------------
// Command Context Adapter
// ----------------------------------------------------------------------------

// commandContextAdapter wraps eval.CommandContext to implement CommandContext.
type commandContextAdapter struct {
	ctx *eval.CommandContext
}

func (cca *commandContextAdapter) Get(name string) (string, bool) {
	return cca.ctx.Get(name)
}

func (cca *commandContextAdapter) IsDefined(name string) bool {
	return cca.ctx.IsDefined(name)
}

func (cca *commandContextAdapter) SetStem(stem string) {
	cca.ctx.SetStem(stem)
}

func (cca *commandContextAdapter) SetCaptures(captures map[string]string) {
	cca.ctx.SetCaptures(captures)
}

func (cca *commandContextAdapter) AutomaticVarNames() []string {
	// Return the list of automatic variables
	return []string{"target", "out", "deps", "in", "stem", "target.dir", "target.file"}
}

func (cca *commandContextAdapter) GetAutomatic(name string) (string, bool) {
	return cca.ctx.Get(name)
}

func (cca *commandContextAdapter) underlyingContext() *eval.CommandContext {
	return cca.ctx
}

// NewCommandContext creates a CommandContext for command interpolation.
func NewCommandContext(parentCtx EvalContext, target string, deps []string) CommandContext {
	// Get the underlying eval.Context from the parent
	eca, ok := parentCtx.(*evalContextAdapter)
	if !ok {
		// Create a new context if we can't get the underlying one
		return &commandContextAdapter{
			ctx: eval.NewCommandContext(eval.NewContext(), target, deps),
		}
	}

	return &commandContextAdapter{
		ctx: eval.NewCommandContext(eca.ctx, target, deps),
	}
}

// ----------------------------------------------------------------------------
// Interpolation Result Adapter
// ----------------------------------------------------------------------------

// interpolateResultAdapter wraps the result of command interpolation.
type interpolateResultAdapter struct {
	interpolated string
	err          error
}

func (r *interpolateResultAdapter) Interpolated() string {
	return r.interpolated
}

func (r *interpolateResultAdapter) Error() error {
	return r.err
}

// InterpolateLineCommand interpolates a line command.
func InterpolateLineCommand(cmd *ast.LineCommand, ctx CommandContext) InterpolateResult {
	cca, ok := ctx.(*commandContextAdapter)
	if !ok {
		return &interpolateResultAdapter{
			err: &eval.UndefinedVariableError{Name: "(invalid context)"},
		}
	}

	result, err := eval.InterpolateCommand(cmd, cca.ctx)
	return &interpolateResultAdapter{
		interpolated: result,
		err:          err,
	}
}

// InterpolateBlockCommand interpolates a block command.
func InterpolateBlockCommand(cmd *ast.BlockCommand, ctx CommandContext) InterpolateResult {
	cca, ok := ctx.(*commandContextAdapter)
	if !ok {
		return &interpolateResultAdapter{
			err: &eval.UndefinedVariableError{Name: "(invalid context)"},
		}
	}

	result, err := eval.InterpolateBlockCommand(cmd, cca.ctx)
	return &interpolateResultAdapter{
		interpolated: result,
		err:          err,
	}
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
// The workDir parameter specifies the working directory for shell commands
// (typically the directory containing the Buildfile).
func EvaluateVariables(result BuildfileResult, workDir string) EvalResult {
	ctx := eval.NewContext()
	if workDir != "" {
		ctx.SetWorkDir(workDir)
	}
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
