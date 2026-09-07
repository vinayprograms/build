package environ

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/eval"
)

// LimaEnvironment represents a Lima VM-based build environment.
type LimaEnvironment struct {
	env        *ast.Environment
	projectDir string
	configPath string
	vmName     string
	args       []string
	isStarted  bool
	lookPath   func(string) (string, error)
	runCommand func(name string, args ...string) *exec.Cmd
}

// NewLimaEnvironment creates a new LimaEnvironment. ctx resolves any
// interpolation in the .source: path.
func NewLimaEnvironment(env *ast.Environment, projectDir, projectName string, ctx *eval.Context) (*LimaEnvironment, error) {
	if env.Runtime == nil || *env.Runtime != ast.RuntimeLima {
		return nil, &InvalidRuntimeError{Message: "not a lima runtime"}
	}

	detector := NewLimaDetector()

	// Detect lima configuration
	result, err := detector.DetectConfig(projectDir, env.Source, ctx)
	if err != nil {
		return nil, err
	}
	if !result.Found {
		return nil, fmt.Errorf("no lima configuration found in %s", projectDir)
	}

	// Generate VM name from project and environment name
	vmName := projectName
	if env.Name != nil {
		vmName = projectName + "-" + *env.Name
	}
	// Sanitize VM name (lima requires lowercase, alphanumeric with dashes)
	vmName = strings.ToLower(vmName)
	vmName = strings.ReplaceAll(vmName, " ", "-")

	// Parse extra args from .args: directive
	args := ParseExtraArgs(env.Args)

	return &LimaEnvironment{
		env:        env,
		projectDir: projectDir,
		configPath: result.Path,
		vmName:     vmName,
		args:       args,
		lookPath:   exec.LookPath,
		runCommand: exec.Command,
	}, nil
}

// Close releases resources held by the lima environment.
// Note: We don't stop the VM on close as it may be reused.
func (l *LimaEnvironment) Close() error {
	return nil
}

// Validate validates that the lima environment is properly configured.
func (l *LimaEnvironment) Validate() error {
	// Check if limactl is available
	_, err := l.lookPath("limactl")
	if err != nil {
		return &BinaryNotFoundError{Name: "limactl"}
	}
	return nil
}

// EnsureReady ensures the Lima VM is running.
func (l *LimaEnvironment) EnsureReady(ctx context.Context, output io.Writer) error {
	if l.isStarted {
		return nil
	}

	// Check if VM exists and get its status
	cmd := l.runCommand("limactl", "list", "--format", "{{.Name}}:{{.Status}}")
	cmdOutput, err := cmd.Output()

	vmExists := false
	vmRunning := false
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(cmdOutput)), "\n")
		for _, line := range lines {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 && parts[0] == l.vmName {
				vmExists = true
				vmRunning = parts[1] == "Running"
				break
			}
		}
	}

	if vmRunning {
		l.isStarted = true
		return nil
	}

	// Start or create the VM (use --tty=false to disable interactive prompts)
	if output != nil {
		if vmExists {
			fmt.Fprintf(output, "Starting Lima VM %s...\n", l.vmName)
		} else {
			fmt.Fprintf(output, "Creating and starting Lima VM %s...\n", l.vmName)
		}
	}

	var args []string
	if vmExists {
		// VM exists but is stopped - just start it
		args = []string{"start", "--tty=false", l.vmName}
	} else {
		// VM doesn't exist - create and start it
		args = []string{"start", "--tty=false"}
		if l.configPath != "" {
			args = append(args, "--name="+l.vmName, l.configPath)
		} else {
			args = append(args, l.vmName)
		}
	}
	args = append(args, l.args...)

	cmd = l.runCommand("limactl", args...)
	if output != nil {
		cmd.Stdout = output
		cmd.Stderr = output
		err = cmd.Run()
	} else {
		cmdOutput, err = cmd.CombinedOutput()
	}
	if err != nil {
		return fmt.Errorf("failed to start Lima VM: %w\n%s", err, string(cmdOutput))
	}

	l.isStarted = true
	return nil
}

// RunCommand runs a command in the Lima VM and returns the result.
func (l *LimaEnvironment) RunCommand(ctx context.Context, command []string) (*RunResult, error) {
	result := &RunResult{}

	// Use limactl shell with --tty=false to disable interactive prompts
	// command is already [shell, "-c", script] from the executor
	args := []string{"shell", "--tty=false", l.vmName, "--"}
	args = append(args, command...)

	cmd := l.runCommand("limactl", args...)
	cmd.Dir = l.projectDir
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
			return result, fmt.Errorf("lima command failed: %w", err)
		}
	}

	return result, nil
}

// RunCommandStreaming runs a command in the Lima VM and streams output.
func (l *LimaEnvironment) RunCommandStreaming(ctx context.Context, command []string, stdout, stderr io.Writer) (int, error) {
	// Use limactl shell with --tty=false to disable interactive prompts
	// command is already [shell, "-c", script] from the executor
	args := []string{"shell", "--tty=false", l.vmName, "--"}
	args = append(args, command...)

	cmd := l.runCommand("limactl", args...)
	cmd.Dir = l.projectDir
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, fmt.Errorf("lima command failed: %w", err)
	}

	return 0, nil
}

// RuntimeName returns the name of the runtime.
func (l *LimaEnvironment) RuntimeName() string {
	return "lima"
}

// ProjectDir returns the project directory.
func (l *LimaEnvironment) ProjectDir() string {
	return l.projectDir
}

// VMName returns the Lima VM name.
func (l *LimaEnvironment) VMName() string {
	return l.vmName
}
