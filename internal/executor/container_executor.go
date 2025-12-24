package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/vinayprograms/build/internal/environ"
	"github.com/vinayprograms/build/internal/output"
)

// ContainerExecutor executes commands inside a container environment.
type ContainerExecutor struct {
	containerEnv *environ.ContainerEnvironment
	config       *ShellConfig
	output       io.Writer
	emitter      *output.Emitter
	target       string
	imageBuilt   bool // Track if we've already built the image
}

// NewContainerExecutor creates a new ContainerExecutor for the given environment.
func NewContainerExecutor(containerEnv *environ.ContainerEnvironment, config *ShellConfig) *ContainerExecutor {
	return &ContainerExecutor{
		containerEnv: containerEnv,
		config:       config,
		output:       nil,
	}
}

// SetOutput sets the output writer for dry-run and verbose modes.
func (e *ContainerExecutor) SetOutput(w io.Writer) {
	e.output = w
}

// SetEmitter sets the event emitter for the output system.
func (e *ContainerExecutor) SetEmitter(emitter *output.Emitter) {
	e.emitter = emitter
}

// SetTarget sets the current target for event context.
func (e *ContainerExecutor) SetTarget(target string) {
	e.target = target
}

// EnsureReady builds the container image if needed.
func (e *ContainerExecutor) EnsureReady(ctx context.Context) error {
	if e.imageBuilt {
		return nil
	}

	if err := e.containerEnv.EnsureReady(ctx); err != nil {
		return fmt.Errorf("failed to build container image: %w", err)
	}

	e.imageBuilt = true
	return nil
}

// ExecuteLine executes a single shell command line inside the container.
func (e *ContainerExecutor) ExecuteLine(ctx context.Context, cmdLine string) (*ExecResult, error) {
	result := &ExecResult{
		Command: cmdLine,
	}
	start := time.Now()

	// Dry-run mode: print and return without executing
	if e.config.DryRun {
		if !e.config.Quiet {
			displayCmd := fmt.Sprintf("[%s] %s", e.containerEnv.RuntimeName(), cmdLine)
			if e.emitter != nil {
				e.emitter.DryRunCommand(e.target, displayCmd)
			} else if e.output != nil {
				fmt.Fprintln(e.output, displayCmd)
			}
		}
		return result, nil
	}

	// Ensure image is built
	if err := e.EnsureReady(ctx); err != nil {
		return result, err
	}

	// Print command before executing (like make does), unless quiet
	if !e.config.Quiet {
		displayCmd := fmt.Sprintf("[%s] %s", e.containerEnv.RuntimeName(), cmdLine)
		if e.emitter != nil {
			e.emitter.CommandStarted(e.target, displayCmd)
		} else if e.output != nil {
			fmt.Fprintln(e.output, displayCmd)
		}
	}

	// Build the command to run inside the container
	// We run the command through the container's shell
	shell := e.config.Shell
	if shell == "" {
		shell = "/bin/sh"
	}
	command := []string{shell, "-c", cmdLine}

	// Execute in container
	runResult, err := e.containerEnv.RunCommand(ctx, command)
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
		} else if e.output != nil {
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

// ExecuteBlock executes a multi-line block of commands inside the container.
func (e *ContainerExecutor) ExecuteBlock(ctx context.Context, script string) (*ExecResult, error) {
	result := &ExecResult{
		Command: strings.Split(script, "\n")[0] + " ...",
	}
	start := time.Now()

	// Dry-run mode: print and return without executing
	if e.config.DryRun {
		if !e.config.Quiet {
			displayCmd := fmt.Sprintf("[%s] (block)", e.containerEnv.RuntimeName())
			if e.emitter != nil {
				e.emitter.DryRunCommand(e.target, displayCmd)
			} else if e.output != nil {
				fmt.Fprintln(e.output, displayCmd)
				fmt.Fprintln(e.output, script)
			}
		}
		return result, nil
	}

	// Ensure image is built
	if err := e.EnsureReady(ctx); err != nil {
		return result, err
	}

	// Print command before executing, unless quiet
	if !e.config.Quiet {
		displayCmd := fmt.Sprintf("[%s] (block)", e.containerEnv.RuntimeName())
		if e.emitter != nil {
			e.emitter.CommandStarted(e.target, displayCmd)
		} else if e.output != nil {
			fmt.Fprintln(e.output, displayCmd)
		}
	}

	// Build the command to run inside the container
	shell := e.config.Shell
	if shell == "" {
		shell = "/bin/sh"
	}
	command := []string{shell, "-c", script}

	// Execute in container with streaming output
	var stdout, stderr bytes.Buffer
	exitCode, err := e.containerEnv.RunCommandStreaming(ctx, command, &stdout, &stderr)
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
		} else if e.output != nil {
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
			e.emitter.CommandCompleted(e.target, result.Command, result.ExitCode, time.Since(start))
		}
		return result, &CommandError{
			Command:  result.Command,
			ExitCode: result.ExitCode,
			Stderr:   result.Stderr,
		}
	}

	// Emit completion event
	if e.emitter != nil {
		e.emitter.CommandCompleted(e.target, result.Command, 0, time.Since(start))
	}

	return result, nil
}

// RuntimeName returns the name of the container runtime.
func (e *ContainerExecutor) RuntimeName() string {
	return e.containerEnv.RuntimeName()
}

// Close releases resources held by the container executor.
func (e *ContainerExecutor) Close() error {
	return e.containerEnv.Close()
}
