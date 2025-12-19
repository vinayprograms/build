package environ

import (
	"os/exec"
	"path/filepath"

	"github.com/vinayprograms/build/internal/ast"
)

// ContainerRunner runs commands in containers.
type ContainerRunner struct {
	runtime      ast.Runtime
	runtimeCmd   string
	workspaceDir string
	extraArgs    []string
}

// NewContainerRunner creates a new ContainerRunner.
func NewContainerRunner(runtime ast.Runtime, runtimeCmd, workspaceDir string) *ContainerRunner {
	return &ContainerRunner{
		runtime:      runtime,
		runtimeCmd:   runtimeCmd,
		workspaceDir: workspaceDir,
	}
}

// SetExtraArgs sets additional arguments for container commands.
func (r *ContainerRunner) SetExtraArgs(args []string) {
	r.extraArgs = args
}

// containerWorkdir returns the workdir path inside the container.
const containerWorkdir = "/workspace"

// RunCommand returns an exec.Cmd for running a command in a container.
// The container is removed after the command completes (--rm).
func (r *ContainerRunner) RunCommand(imageTag string, command []string) *exec.Cmd {
	args := []string{"run"}

	// Auto-remove container after execution
	args = append(args, "--rm")

	// Mount workspace
	mountSpec := r.workspaceDir + ":" + containerWorkdir
	args = append(args, "-v", mountSpec)

	// Set working directory
	args = append(args, "-w", containerWorkdir)

	// Add extra args
	args = append(args, r.extraArgs...)

	// Add image
	args = append(args, imageTag)

	// Add command
	args = append(args, command...)

	return exec.Command(r.runtimeCmd, args...)
}

// ShellCommand returns an exec.Cmd for running an interactive shell in a container.
func (r *ContainerRunner) ShellCommand(imageTag, shell string) *exec.Cmd {
	args := []string{"run"}

	// Interactive with TTY
	args = append(args, "-it")

	// Auto-remove container after shell exits
	args = append(args, "--rm")

	// Mount workspace
	mountSpec := r.workspaceDir + ":" + containerWorkdir
	args = append(args, "-v", mountSpec)

	// Set working directory
	args = append(args, "-w", containerWorkdir)

	// Add extra args
	args = append(args, r.extraArgs...)

	// Add image
	args = append(args, imageTag)

	// Add shell
	args = append(args, shell)

	return exec.Command(r.runtimeCmd, args...)
}

// RunCommandKeepAlive returns an exec.Cmd for running a command in a container
// that is kept running after the command completes.
func (r *ContainerRunner) RunCommandKeepAlive(imageTag string, command []string, containerName string) *exec.Cmd {
	args := []string{"run"}

	// Name the container
	args = append(args, "--name", containerName)

	// Mount workspace
	mountSpec := r.workspaceDir + ":" + containerWorkdir
	args = append(args, "-v", mountSpec)

	// Set working directory
	args = append(args, "-w", containerWorkdir)

	// Add extra args
	args = append(args, r.extraArgs...)

	// Add image
	args = append(args, imageTag)

	// Add command
	args = append(args, command...)

	return exec.Command(r.runtimeCmd, args...)
}

// ExecCommand returns an exec.Cmd for running a command in an existing container.
func (r *ContainerRunner) ExecCommand(containerName string, command []string) *exec.Cmd {
	args := []string{"exec"}

	// Set working directory
	args = append(args, "-w", containerWorkdir)

	// Add container name
	args = append(args, containerName)

	// Add command
	args = append(args, command...)

	return exec.Command(r.runtimeCmd, args...)
}

// StopContainer returns an exec.Cmd for stopping a container.
func (r *ContainerRunner) StopContainer(containerName string) *exec.Cmd {
	return exec.Command(r.runtimeCmd, "stop", containerName)
}

// RemoveContainer returns an exec.Cmd for removing a container.
func (r *ContainerRunner) RemoveContainer(containerName string) *exec.Cmd {
	return exec.Command(r.runtimeCmd, "rm", containerName)
}

// GenerateContainerName generates a unique container name for a build.
func GenerateContainerName(project, envName string) string {
	name := "build-" + filepath.Base(project)
	if envName != "" {
		name += "-" + envName
	}
	return name
}
