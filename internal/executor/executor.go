package executor

import (
	"bytes"
	"context"
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
	config       *ShellConfig
	output       io.Writer                 // Output for dry-run/verbose (defaults to os.Stdout)
	emitter      *output.Emitter           // Event emitter for output system
	target       string                    // Current target being built (for event context)
	runtimeEnv   environ.RuntimeEnvironment // Runtime environment for execution (docker, nix, etc.)
	envReady     bool                      // Track if runtime environment is ready
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

// SetRuntimeEnv sets the container environment for containerized execution.
// When set, commands will be executed inside the container.
func (e *Executor) SetRuntimeEnv(runtimeEnv environ.RuntimeEnvironment) {
	e.runtimeEnv = runtimeEnv
}

// HasRuntimeEnv returns true if a container environment is configured.
func (e *Executor) HasRuntimeEnv() bool {
	return e.runtimeEnv != nil
}

// EnsureRuntimeReady builds the container image if not already built.
func (e *Executor) EnsureRuntimeReady(ctx context.Context) error {
	if e.runtimeEnv == nil || e.envReady {
		return nil
	}
	if err := e.runtimeEnv.EnsureReady(ctx); err != nil {
		return fmt.Errorf("failed to build container image: %w", err)
	}
	e.envReady = true
	return nil
}

// Close releases resources held by the executor.
func (e *Executor) Close() error {
	if e.runtimeEnv != nil {
		return e.runtimeEnv.Close()
	}
	return nil
}

// ExecuteLine executes a single shell command line.
func (e *Executor) ExecuteLine(cmdLine string) (*ExecResult, error) {
	return e.ExecuteLineWithContext(context.Background(), cmdLine)
}

// ExecuteLineWithContext executes a single shell command line with context.
func (e *Executor) ExecuteLineWithContext(ctx context.Context, cmdLine string) (*ExecResult, error) {
	result := &ExecResult{
		Command: cmdLine,
	}
	start := time.Now()

	// Determine display command (show container runtime prefix if in container)
	displayCmd := cmdLine
	if e.runtimeEnv != nil {
		displayCmd = fmt.Sprintf("[%s] %s", e.runtimeEnv.RuntimeName(), cmdLine)
	}

	// Dry-run mode: print and return without executing
	if e.config.DryRun {
		if !e.config.Quiet {
			if e.emitter != nil {
				e.emitter.DryRunCommand(e.target, displayCmd)
			} else {
				fmt.Fprintln(e.output, displayCmd)
			}
		}
		return result, nil
	}

	// Print command before executing (like make does), unless quiet
	if !e.config.Quiet {
		if e.emitter != nil {
			e.emitter.CommandStarted(e.target, displayCmd)
		} else {
			fmt.Fprintln(e.output, displayCmd)
		}
	}

	// Execute in container or locally
	if e.runtimeEnv != nil {
		return e.executeLineInContainer(ctx, cmdLine, result, start)
	}

	// Execute locally using platform-appropriate shell args
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

// executeLineInContainer executes a command inside the container.
func (e *Executor) executeLineInContainer(ctx context.Context, cmdLine string, result *ExecResult, start time.Time) (*ExecResult, error) {
	// Ensure image is built
	if err := e.EnsureRuntimeReady(ctx); err != nil {
		return result, err
	}

	// Build command to run in container shell
	shell := e.config.Shell
	if shell == "" {
		shell = "/bin/sh"
	}
	command := []string{shell, "-c", cmdLine}

	// Execute in container
	runResult, err := e.runtimeEnv.RunCommand(ctx, command)
	if err != nil {
		return result, fmt.Errorf("container execution failed: %w", err)
	}

	result.Stdout = runResult.Stdout
	result.Stderr = runResult.Stderr
	result.ExitCode = runResult.ExitCode

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

	// Handle non-zero exit code
	if result.ExitCode != 0 {
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
		e.emitter.CommandCompleted(e.target, cmdLine, 0, time.Since(start))
	}

	return result, nil
}

// ExecuteBlock executes a shell script block.
func (e *Executor) ExecuteBlock(script string) (*ExecResult, error) {
	return e.ExecuteBlockWithContext(context.Background(), script)
}

// ExecuteBlockWithContext executes a shell script block with context.
func (e *Executor) ExecuteBlockWithContext(ctx context.Context, script string) (*ExecResult, error) {
	result := &ExecResult{
		Command: script,
	}
	start := time.Now()

	// Determine display script (show container runtime prefix if in container)
	displayScript := script
	if e.runtimeEnv != nil {
		displayScript = fmt.Sprintf("[%s] (block)\n%s", e.runtimeEnv.RuntimeName(), script)
	}

	// Verbose mode: print script
	if e.config.Verbose && !e.config.DryRun {
		if e.emitter != nil {
			e.emitter.CommandStarted(e.target, displayScript)
		} else {
			fmt.Fprintln(e.output, displayScript)
		}
	}

	// Dry-run mode: don't execute
	if e.config.DryRun {
		if e.emitter != nil {
			e.emitter.DryRunCommand(e.target, displayScript)
		} else {
			fmt.Fprintln(e.output, displayScript)
		}
		return result, nil
	}

	// Execute in container or locally
	if e.runtimeEnv != nil {
		return e.executeBlockInContainer(ctx, script, result, start)
	}

	// Execute locally using platform-appropriate shell args
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

// executeBlockInContainer executes a block script inside the container.
func (e *Executor) executeBlockInContainer(ctx context.Context, script string, result *ExecResult, start time.Time) (*ExecResult, error) {
	// Ensure image is built
	if err := e.EnsureRuntimeReady(ctx); err != nil {
		return result, err
	}

	// Build command to run in container shell
	shell := e.config.Shell
	if shell == "" {
		shell = "/bin/sh"
	}
	command := []string{shell, "-c", script}

	// Execute in container with streaming
	var stdout, stderr bytes.Buffer
	exitCode, err := e.runtimeEnv.RunCommandStreaming(ctx, command, &stdout, &stderr)
	if err != nil {
		return result, fmt.Errorf("container execution failed: %w", err)
	}

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	result.ExitCode = exitCode

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

	// Handle non-zero exit code
	if result.ExitCode != 0 {
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
		e.emitter.CommandCompleted(e.target, script, 0, time.Since(start))
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

	// Check .requires directive - verify required binaries exist and versions match
	if len(recipe.Directives.Requires) > 0 {
		checker := environ.NewRequirementsChecker()
		checkResults := checker.CheckRequirementsWithVersion(recipe.Directives.Requires)
		var errors []string
		for _, r := range checkResults {
			if r.Error != nil {
				errors = append(errors, r.String())
			}
		}
		if len(errors) > 0 {
			return results, &RequirementError{
				Target: target,
				Errors: errors,
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

// RequirementError represents missing or incompatible required binaries.
type RequirementError struct {
	Target string
	Errors []string
}

func (e *RequirementError) Error() string {
	return fmt.Sprintf("requirements not satisfied for %s: %s", e.Target, strings.Join(e.Errors, "; "))
}
