package environ

import (
	"fmt"
	"os/exec"

	"github.com/vinayprograms/build/internal/ast"
)

// ContainerEnvironment represents a container-based build environment.
type ContainerEnvironment struct {
	env         *ast.Environment
	projectDir  string
	projectName string
	detector    *ContainerDetector
	builder     *ImageBuilder
	runner      *ContainerRunner
	imageTag    string
}

// NewContainerEnvironment creates a new ContainerEnvironment.
func NewContainerEnvironment(env *ast.Environment, projectDir, projectName string) (*ContainerEnvironment, error) {
	if env.Runtime == nil {
		return nil, &InvalidRuntimeError{Message: "runtime not specified"}
	}

	if !isContainerRuntime(*env.Runtime) {
		return nil, &InvalidRuntimeError{Message: "not a container runtime"}
	}

	detector := NewContainerDetector()

	// Find the runtime binary
	runtimePath, err := detector.FindRuntime(*env.Runtime)
	if err != nil {
		return nil, err
	}

	// Parse extra args from .args: directive
	extraArgs := ParseExtraArgs(env.Args)

	// Determine environment name
	envName := ""
	if env.Name != nil {
		envName = *env.Name
	}

	// Generate image tag
	imageTag := GenerateImageTag(projectName, envName)

	// Create builder and runner
	builder := NewImageBuilder(*env.Runtime, runtimePath, extraArgs)
	runner := NewContainerRunner(*env.Runtime, runtimePath, projectDir)
	runner.SetExtraArgs(extraArgs)

	return &ContainerEnvironment{
		env:         env,
		projectDir:  projectDir,
		projectName: projectName,
		detector:    detector,
		builder:     builder,
		runner:      runner,
		imageTag:    imageTag,
	}, nil
}

// Validate validates that the container environment is properly configured.
func (c *ContainerEnvironment) Validate() error {
	// Detect and validate Dockerfile
	result, err := c.detector.DetectDockerfile(c.env, c.projectDir)
	if err != nil {
		return err
	}

	if err := c.detector.ValidateDockerfile(result.Path); err != nil {
		return err
	}

	return nil
}

// EnsureImage ensures the container image is built and available.
func (c *ContainerEnvironment) EnsureImage() error {
	// Check if image already exists
	exists, err := c.builder.ImageExists(c.imageTag)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// Build the image
	result, err := c.detector.DetectDockerfile(c.env, c.projectDir)
	if err != nil {
		return err
	}

	if err := c.builder.Build(result.Path, c.imageTag); err != nil {
		return &ImageBuildError{
			ImageTag: c.imageTag,
			Message:  err.Error(),
		}
	}

	return nil
}

// RunCommand runs a command in the container environment.
func (c *ContainerEnvironment) RunCommand(command []string) *exec.Cmd {
	return c.runner.RunCommand(c.imageTag, command)
}

// Shell opens an interactive shell in the container.
func (c *ContainerEnvironment) Shell(shellPath string) *exec.Cmd {
	if shellPath == "" {
		shellPath = "/bin/sh"
	}
	return c.runner.ShellCommand(c.imageTag, shellPath)
}

// RunCommandKeepAlive runs a command and keeps the container running.
func (c *ContainerEnvironment) RunCommandKeepAlive(command []string) (*exec.Cmd, string) {
	envName := ""
	if c.env.Name != nil {
		envName = *c.env.Name
	}
	containerName := GenerateContainerName(c.projectName, envName)
	return c.runner.RunCommandKeepAlive(c.imageTag, command, containerName), containerName
}

// StopContainer stops a running container by name.
func (c *ContainerEnvironment) StopContainer(containerName string) *exec.Cmd {
	return c.runner.StopContainer(containerName)
}

// RemoveContainer removes a container by name.
func (c *ContainerEnvironment) RemoveContainer(containerName string) *exec.Cmd {
	return c.runner.RemoveContainer(containerName)
}

// ImageTag returns the image tag for this environment.
func (c *ContainerEnvironment) ImageTag() string {
	return c.imageTag
}

// RuntimeName returns the name of the runtime (docker or podman).
func (c *ContainerEnvironment) RuntimeName() string {
	if c.env.Runtime == nil {
		return ""
	}
	return runtimeName(*c.env.Runtime)
}

// PrintKeepInstructions prints instructions for using a kept-alive container.
func PrintKeepInstructions(runtime, containerName string) string {
	return fmt.Sprintf(`Container '%s' is still running.

To enter the container:
  %s exec -it %s /bin/sh

To stop the container:
  %s stop %s

To remove the container:
  %s rm %s
`, containerName, runtime, containerName, runtime, containerName, runtime, containerName)
}
