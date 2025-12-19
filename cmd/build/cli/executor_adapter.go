package cli

import (
	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/eval"
	"github.com/vinayprograms/build/internal/executor"
)

// ----------------------------------------------------------------------------
// Shell Configuration Adapter
// ----------------------------------------------------------------------------

// ShellConfig wraps executor.ShellConfig for CLI use.
type ShellConfig struct {
	cfg *executor.ShellConfig
}

// NewShellConfig creates a new shell configuration with defaults.
func NewShellConfig() *ShellConfig {
	return &ShellConfig{
		cfg: executor.NewShellConfig(),
	}
}

// SetShell sets the shell path.
func (c *ShellConfig) SetShell(shell string) {
	c.cfg.SetShell(shell)
}

// Shell returns the current shell path.
func (c *ShellConfig) Shell() string {
	return c.cfg.Shell
}

// SetDryRun enables or disables dry-run mode.
func (c *ShellConfig) SetDryRun(dryRun bool) {
	c.cfg.DryRun = dryRun
}

// DryRun returns whether dry-run mode is enabled.
func (c *ShellConfig) DryRun() bool {
	return c.cfg.DryRun
}

// SetVerbose enables or disables verbose mode.
func (c *ShellConfig) SetVerbose(verbose bool) {
	c.cfg.Verbose = verbose
}

// Verbose returns whether verbose mode is enabled.
func (c *ShellConfig) Verbose() bool {
	return c.cfg.Verbose
}

// SetQuiet enables or disables quiet mode.
func (c *ShellConfig) SetQuiet(quiet bool) {
	c.cfg.Quiet = quiet
}

// Quiet returns whether quiet mode is enabled.
func (c *ShellConfig) Quiet() bool {
	return c.cfg.Quiet
}

// WithOverride returns a new config with the shell overridden.
func (c *ShellConfig) WithOverride(shell string) *ShellConfig {
	return &ShellConfig{
		cfg: c.cfg.WithOverride(shell),
	}
}

// ----------------------------------------------------------------------------
// Executor Adapter
// ----------------------------------------------------------------------------

// Executor wraps executor.Executor for CLI use.
type Executor struct {
	exec *executor.Executor
}

// NewExecutor creates a new executor with the given configuration.
func NewExecutor(config *ShellConfig) *Executor {
	return &Executor{
		exec: executor.NewExecutor(config.cfg),
	}
}

// ExecResult represents the result of a command execution.
type ExecResult struct {
	result *executor.ExecResult
}

// Command returns the command that was executed.
func (r *ExecResult) Command() string {
	if r.result == nil {
		return ""
	}
	return r.result.Command
}

// Stdout returns the standard output.
func (r *ExecResult) Stdout() string {
	if r.result == nil {
		return ""
	}
	return r.result.Stdout
}

// Stderr returns the standard error.
func (r *ExecResult) Stderr() string {
	if r.result == nil {
		return ""
	}
	return r.result.Stderr
}

// ExitCode returns the exit code.
func (r *ExecResult) ExitCode() int {
	if r.result == nil {
		return 0
	}
	return r.result.ExitCode
}

// ExecuteLine executes a single shell command line.
func (e *Executor) ExecuteLine(cmdLine string) (*ExecResult, error) {
	result, err := e.exec.ExecuteLine(cmdLine)
	return &ExecResult{result: result}, err
}

// ExecuteBlock executes a shell script block.
func (e *Executor) ExecuteBlock(script string) (*ExecResult, error) {
	result, err := e.exec.ExecuteBlock(script)
	return &ExecResult{result: result}, err
}

// ExecuteRecipe executes all commands in a recipe.
func (e *Executor) ExecuteRecipe(recipe *ast.Recipe, cmdCtx CommandContext) ([]*ExecResult, error) {
	// Get the underlying command context
	cca, ok := cmdCtx.(*commandContextAdapter)
	if !ok {
		return nil, &eval.UndefinedVariableError{Name: "(invalid command context)"}
	}

	results, err := e.exec.ExecuteRecipe(recipe, cca.ctx)

	// Convert results
	var adapted []*ExecResult
	for _, r := range results {
		adapted = append(adapted, &ExecResult{result: r})
	}

	return adapted, err
}

// GetRecipeShell returns the effective shell for a recipe.
// Returns the recipe's .shell: directive if present, otherwise the global shell.
func GetRecipeShell(recipe *ast.Recipe, globalShell string, ctx EvalContext) string {
	if recipe == nil || recipe.Directives.Shell == nil {
		return globalShell
	}

	// Get the underlying context
	eca, ok := ctx.(*evalContextAdapter)
	if !ok {
		return globalShell
	}

	// Evaluate the shell directive
	evaluator := eval.NewEvaluator(eca.ctx)
	shellVal, err := evaluator.EvaluateValue(recipe.Directives.Shell)
	if err != nil {
		return globalShell
	}

	return shellVal
}
