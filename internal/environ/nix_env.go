package environ

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/eval"
)

// NixEnvironment represents a Nix-based build environment.
type NixEnvironment struct {
	env        *ast.Environment
	projectDir string
	configPath string
	nixType    NixType
	args       []string
	lookPath   func(string) (string, error)
	runCommand func(name string, args ...string) *exec.Cmd
}

// NewNixEnvironment creates a new NixEnvironment. ctx resolves any
// interpolation in the .source: path.
func NewNixEnvironment(env *ast.Environment, projectDir string, ctx *eval.Context) (*NixEnvironment, error) {
	if env.Runtime == nil || *env.Runtime != ast.RuntimeNix {
		return nil, &InvalidRuntimeError{Message: "not a nix runtime"}
	}

	detector := NewNixDetector()

	// Detect nix configuration
	result, err := detector.DetectConfig(projectDir, env.Source, ctx)
	if err != nil {
		return nil, err
	}
	if !result.Found {
		return nil, fmt.Errorf("no nix configuration found in %s", projectDir)
	}

	// Parse extra args from .args: directive
	args := ParseExtraArgs(env.Args)

	return &NixEnvironment{
		env:        env,
		projectDir: projectDir,
		configPath: result.Path,
		nixType:    result.Type,
		args:       args,
		lookPath:   exec.LookPath,
		runCommand: exec.Command,
	}, nil
}

// Close releases resources held by the nix environment.
func (n *NixEnvironment) Close() error {
	return nil
}

// Validate validates that the nix environment is properly configured.
func (n *NixEnvironment) Validate() error {
	// Check if nix-shell is available
	_, err := n.lookPath("nix-shell")
	if err != nil {
		return &BinaryNotFoundError{Name: "nix-shell"}
	}
	return nil
}

// EnsureReady ensures the nix environment is ready.
// For nix, this is a no-op since nix-shell handles everything.
func (n *NixEnvironment) EnsureReady(ctx context.Context, output io.Writer) error {
	return nil
}

// RunCommand runs a command in the nix environment and returns the result.
func (n *NixEnvironment) RunCommand(ctx context.Context, command []string) (*RunResult, error) {
	result := &RunResult{}

	var args []string
	var cmdName string

	if n.nixType == NixTypeFlake {
		// For flakes, use: nix develop -c <command...>
		cmdName = "nix"
		args = []string{"develop"}
		if n.configPath != "" {
			args = append(args, filepath.Dir(n.configPath))
		}
		args = append(args, n.args...)
		args = append(args, "-c")
		args = append(args, command...)
	} else {
		// For shell.nix, use: nix-shell --run "<command>"
		// command is [shell, "-c", script] from executor
		cmdName = "nix-shell"
		if n.configPath != "" {
			args = append(args, n.configPath)
		}
		args = append(args, n.args...)
		// Extract just the script from [shell, "-c", script]
		if len(command) >= 3 && command[1] == "-c" {
			args = append(args, "--run", command[2])
		} else {
			args = append(args, "--run", strings.Join(command, " "))
		}
	}

	cmd := n.runCommand(cmdName, args...)
	cmd.Dir = n.projectDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return result, fmt.Errorf("nix command failed: %w", err)
		}
	}

	return result, nil
}

// RunCommandStreaming runs a command in the nix environment and streams output.
func (n *NixEnvironment) RunCommandStreaming(ctx context.Context, command []string, stdout, stderr io.Writer) (int, error) {
	var args []string
	var cmdName string

	if n.nixType == NixTypeFlake {
		cmdName = "nix"
		args = []string{"develop"}
		if n.configPath != "" {
			args = append(args, filepath.Dir(n.configPath))
		}
		args = append(args, n.args...)
		args = append(args, "-c")
		args = append(args, command...)
	} else {
		cmdName = "nix-shell"
		if n.configPath != "" {
			args = append(args, n.configPath)
		}
		args = append(args, n.args...)
		// Extract just the script from [shell, "-c", script]
		if len(command) >= 3 && command[1] == "-c" {
			args = append(args, "--run", command[2])
		} else {
			args = append(args, "--run", strings.Join(command, " "))
		}
	}

	cmd := n.runCommand(cmdName, args...)
	cmd.Dir = n.projectDir
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, fmt.Errorf("nix command failed: %w", err)
	}

	return 0, nil
}

// RuntimeName returns the name of the runtime.
func (n *NixEnvironment) RuntimeName() string {
	return "nix"
}

// ProjectDir returns the project directory.
func (n *NixEnvironment) ProjectDir() string {
	return n.projectDir
}
