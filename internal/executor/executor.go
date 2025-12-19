package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/eval"
)

// ShellConfig holds shell configuration for command execution.
type ShellConfig struct {
	Shell   string // Path to shell (default: /bin/sh)
	DryRun  bool   // If true, print commands without executing
	Verbose bool   // If true, print commands before executing
}

// NewShellConfig creates a new ShellConfig with default values.
func NewShellConfig() *ShellConfig {
	return &ShellConfig{
		Shell: "/bin/sh",
	}
}

// SetShell sets the shell path.
func (c *ShellConfig) SetShell(shell string) {
	c.Shell = shell
}

// WithOverride returns a new config with the shell overridden.
// The original config is not modified.
func (c *ShellConfig) WithOverride(shell string) *ShellConfig {
	return &ShellConfig{
		Shell:   shell,
		DryRun:  c.DryRun,
		Verbose: c.Verbose,
	}
}

// Validate checks that the shell exists and is executable.
// It handles both absolute paths and shells found via PATH lookup.
func (c *ShellConfig) Validate() error {
	shell := c.Shell

	// If it's an absolute path, check directly
	if len(shell) > 0 && shell[0] == '/' {
		_, err := exec.LookPath(shell)
		if err != nil {
			return &ShellNotFoundError{Shell: shell}
		}
		return nil
	}

	// Otherwise, look up in PATH
	_, err := exec.LookPath(shell)
	if err != nil {
		return &ShellNotFoundError{Shell: shell}
	}
	return nil
}

// ExecResult holds the result of a shell command execution.
type ExecResult struct {
	Command  string // The command that was executed
	Stdout   string // Standard output
	Stderr   string // Standard error
	ExitCode int    // Exit code (0 = success)
}

// Executor handles shell command execution.
type Executor struct {
	config *ShellConfig
	output io.Writer // Output for dry-run/verbose (defaults to os.Stdout)
}

// NewExecutor creates a new Executor with the given configuration.
func NewExecutor(config *ShellConfig) *Executor {
	return &Executor{
		config: config,
		output: os.Stdout,
	}
}

// NewExecutorWithValidation creates a new Executor and validates the shell.
// Returns an error if the shell is not found.
func NewExecutorWithValidation(config *ShellConfig) (*Executor, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return NewExecutor(config), nil
}

// SetOutput sets the output writer for dry-run and verbose modes.
func (e *Executor) SetOutput(w io.Writer) {
	e.output = w
}

// ExecuteLine executes a single shell command line.
func (e *Executor) ExecuteLine(cmdLine string) (*ExecResult, error) {
	result := &ExecResult{
		Command: cmdLine,
	}

	// Verbose mode: print command
	if e.config.Verbose && !e.config.DryRun {
		fmt.Fprintln(e.output, cmdLine)
	}

	// Dry-run mode: don't execute
	if e.config.DryRun {
		fmt.Fprintln(e.output, cmdLine)
		return result, nil
	}

	// Execute the command
	cmd := exec.Command(e.config.Shell, "-c", cmdLine)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	// Get exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			}
		}
		return result, &CommandError{
			Command:  cmdLine,
			ExitCode: result.ExitCode,
			Stderr:   result.Stderr,
		}
	}

	return result, nil
}

// ExecuteBlock executes a shell script block.
func (e *Executor) ExecuteBlock(script string) (*ExecResult, error) {
	result := &ExecResult{
		Command: script,
	}

	// Verbose mode: print script
	if e.config.Verbose && !e.config.DryRun {
		fmt.Fprintln(e.output, script)
	}

	// Dry-run mode: don't execute
	if e.config.DryRun {
		fmt.Fprintln(e.output, script)
		return result, nil
	}

	// Execute the script as a single command
	cmd := exec.Command(e.config.Shell, "-c", script)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	// Get exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			}
		}
		return result, &CommandError{
			Command:  script,
			ExitCode: result.ExitCode,
			Stderr:   result.Stderr,
		}
	}

	return result, nil
}

// ExecuteRecipe executes all commands in a recipe.
// Returns results for each command executed and stops on first error.
func (e *Executor) ExecuteRecipe(recipe *ast.Recipe, cmdCtx *eval.CommandContext) ([]*ExecResult, error) {
	var results []*ExecResult

	// Determine the shell to use
	shell := e.config.Shell
	if recipe.Directives.Shell != nil {
		// Evaluate the shell directive value
		shellEval := eval.NewEvaluator(cmdCtx.Parent())
		shellVal, err := shellEval.EvaluateValue(recipe.Directives.Shell)
		if err != nil {
			return results, fmt.Errorf("evaluating .shell: %w", err)
		}
		shell = shellVal
	}

	// Create executor with recipe-specific shell
	recipeExec := e
	if shell != e.config.Shell {
		recipeCfg := e.config.WithOverride(shell)
		recipeExec = NewExecutor(recipeCfg)
		recipeExec.output = e.output
	}

	// Execute each command
	for _, cmd := range recipe.Commands {
		switch c := cmd.(type) {
		case *ast.LineCommand:
			// Interpolate the command
			line, err := eval.InterpolateCommand(c, cmdCtx)
			if err != nil {
				return results, fmt.Errorf("interpolating command: %w", err)
			}

			result, err := recipeExec.ExecuteLine(line)
			results = append(results, result)
			if err != nil {
				return results, err
			}

		case *ast.BlockCommand:
			// Interpolate the block
			script, err := eval.InterpolateBlockCommand(c, cmdCtx)
			if err != nil {
				return results, fmt.Errorf("interpolating block: %w", err)
			}

			result, err := recipeExec.ExecuteBlock(script)
			results = append(results, result)
			if err != nil {
				return results, err
			}
		}
	}

	return results, nil
}

// ----------------------------------------------------------------------------
// Error Types
// ----------------------------------------------------------------------------

// CommandError represents a command execution failure.
type CommandError struct {
	Command  string
	ExitCode int
	Stderr   string
}

func (e *CommandError) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("command failed with exit code %d: %s\n%s",
			e.ExitCode, e.Command, e.Stderr)
	}
	return fmt.Sprintf("command failed with exit code %d: %s", e.ExitCode, e.Command)
}

// ShellNotFoundError represents a missing shell error.
type ShellNotFoundError struct {
	Shell string
}

func (e *ShellNotFoundError) Error() string {
	return fmt.Sprintf("shell not found: %s", e.Shell)
}
