package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/environ"
	"github.com/vinayprograms/build/internal/eval"
	"github.com/vinayprograms/build/internal/output"
	"github.com/vinayprograms/build/internal/platform"
)

// ShellConfig holds shell configuration for command execution.
type ShellConfig struct {
	Shell   string // Path to shell (default: platform-specific)
	DryRun  bool   // If true, print commands without executing
	Verbose bool   // If true, print commands before executing
	Quiet   bool   // If true, suppress command output (only show errors)
}

// NewShellConfig creates a new ShellConfig with default values.
func NewShellConfig() *ShellConfig {
	return &ShellConfig{
		Shell: platform.DefaultShell(),
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
		Quiet:   c.Quiet,
	}
}

// Validate checks that the shell exists and is executable.
// It handles both absolute paths and shells found via PATH lookup.
func (c *ShellConfig) Validate() error {
	err := platform.ValidateShell(c.Shell)
	if err != nil {
		return &ShellNotFoundError{Shell: c.Shell}
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
	config  *ShellConfig
	output  io.Writer       // Output for dry-run/verbose (defaults to os.Stdout)
	emitter *output.Emitter // Event emitter for output system
	target  string          // Current target being built (for event context)
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

// SetEmitter sets the event emitter for the output system.
func (e *Executor) SetEmitter(emitter *output.Emitter) {
	e.emitter = emitter
}

// SetTarget sets the current target for event context.
func (e *Executor) SetTarget(target string) {
	e.target = target
}

// ExecuteLine executes a single shell command line.
func (e *Executor) ExecuteLine(cmdLine string) (*ExecResult, error) {
	result := &ExecResult{
		Command: cmdLine,
	}
	start := time.Now()

	// Dry-run mode: print and return without executing
	if e.config.DryRun {
		if !e.config.Quiet {
			if e.emitter != nil {
				e.emitter.DryRunCommand(e.target, cmdLine)
			} else {
				fmt.Fprintln(e.output, cmdLine)
			}
		}
		return result, nil
	}

	// Print command before executing (like make does), unless quiet
	if !e.config.Quiet {
		if e.emitter != nil {
			e.emitter.CommandStarted(e.target, cmdLine)
		} else {
			fmt.Fprintln(e.output, cmdLine)
		}
	}

	// Execute the command using platform-appropriate shell args
	args := platform.ShellCommandArgs(e.config.Shell, cmdLine)
	cmd := exec.Command(e.config.Shell, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	// Print output immediately after command (unless quiet)
	if !e.config.Quiet && (result.Stdout != "" || result.Stderr != "") {
		if e.emitter != nil {
			e.emitter.CommandOutput(e.target, result.Stdout, result.Stderr)
		} else {
			if result.Stdout != "" {
				fmt.Fprint(e.output, result.Stdout)
			}
			if result.Stderr != "" {
				fmt.Fprint(e.output, result.Stderr)
			}
		}
	}

	// Get exit code - use cross-platform ExitCode() method
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		// Emit completion event even on error
		if e.emitter != nil {
			e.emitter.CommandCompleted(e.target, cmdLine, result.ExitCode, time.Since(start))
		}
		return result, &CommandError{
			Command:  cmdLine,
			ExitCode: result.ExitCode,
			Stderr:   result.Stderr,
		}
	}

	// Emit completion event
	if e.emitter != nil {
		e.emitter.CommandCompleted(e.target, cmdLine, result.ExitCode, time.Since(start))
	}

	return result, nil
}

// ExecuteBlock executes a shell script block.
func (e *Executor) ExecuteBlock(script string) (*ExecResult, error) {
	result := &ExecResult{
		Command: script,
	}
	start := time.Now()

	// Verbose mode: print script
	if e.config.Verbose && !e.config.DryRun {
		if e.emitter != nil {
			e.emitter.CommandStarted(e.target, script)
		} else {
			fmt.Fprintln(e.output, script)
		}
	}

	// Dry-run mode: don't execute
	if e.config.DryRun {
		if e.emitter != nil {
			e.emitter.DryRunCommand(e.target, script)
		} else {
			fmt.Fprintln(e.output, script)
		}
		return result, nil
	}

	// Execute the script using platform-appropriate shell args
	args := platform.ShellCommandArgs(e.config.Shell, script)
	cmd := exec.Command(e.config.Shell, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	// Print output immediately after command (unless quiet)
	if !e.config.Quiet && (result.Stdout != "" || result.Stderr != "") {
		if e.emitter != nil {
			e.emitter.CommandOutput(e.target, result.Stdout, result.Stderr)
		} else {
			if result.Stdout != "" {
				fmt.Fprint(e.output, result.Stdout)
			}
			if result.Stderr != "" {
				fmt.Fprint(e.output, result.Stderr)
			}
		}
	}

	// Get exit code - use cross-platform ExitCode() method
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		// Emit completion event even on error
		if e.emitter != nil {
			e.emitter.CommandCompleted(e.target, script, result.ExitCode, time.Since(start))
		}
		return result, &CommandError{
			Command:  script,
			ExitCode: result.ExitCode,
			Stderr:   result.Stderr,
		}
	}

	// Emit completion event
	if e.emitter != nil {
		e.emitter.CommandCompleted(e.target, script, result.ExitCode, time.Since(start))
	}

	return result, nil
}

// ExecuteRecipe executes all commands in a recipe.
// Returns results for each command executed and stops on first error.
func (e *Executor) ExecuteRecipe(recipe *ast.Recipe, cmdCtx *eval.CommandContext) ([]*ExecResult, error) {
	var results []*ExecResult

	// Get the target for event context
	target, _ := cmdCtx.Get("target")

	// Set target on executor for event emission
	e.target = target

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
		recipeExec.emitter = e.emitter
		recipeExec.target = target
	}

	// Check .requires directive - verify required binaries exist
	if len(recipe.Directives.Requires) > 0 {
		checker := environ.NewRequirementsChecker()
		checkResults := checker.CheckRequirements(recipe.Directives.Requires)
		var missing []string
		for _, r := range checkResults {
			if r.Error != nil {
				missing = append(missing, r.Requirement.Name)
			}
		}
		if len(missing) > 0 {
			return results, &RequirementError{
				Target:  target,
				Missing: missing,
			}
		}
	}

	// In dry-run mode, print "Would build: target" prefix
	if e.config.DryRun && len(recipe.Commands) > 0 {
		fmt.Fprintf(e.output, "Would build: %s\n", target)
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

// RequirementError represents a missing required binary.
type RequirementError struct {
	Target  string
	Missing []string
}

func (e *RequirementError) Error() string {
	return fmt.Sprintf("missing required binaries for %s: %s", e.Target, strings.Join(e.Missing, ", "))
}
