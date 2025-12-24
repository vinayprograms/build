package environ

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/vinayprograms/build/internal/ast"
)

// ContainerEnvironment represents a container-based build environment.
type ContainerEnvironment struct {
	env            *ast.Environment
	projectDir     string
	projectName    string
	detector       *ContainerDetector
	dockerClient   *DockerClient
	imageTag       string
	extraArgs      []string
	dockerfilePath string
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

	// Verify the runtime binary is available in PATH
	_, err := detector.FindRuntime(*env.Runtime)
	if err != nil {
		return nil, err
	}

	// Connect to container daemon (Docker SDK works with both docker and podman)
	dockerClient, err := NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to container daemon: %w", err)
	}

	// Ping to ensure daemon is responsive
	if err := dockerClient.Ping(context.Background()); err != nil {
		dockerClient.Close()
		return nil, fmt.Errorf("container daemon not responding: %w", err)
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

	// Detect Dockerfile path
	result, err := detector.DetectDockerfile(env, projectDir)
	if err != nil {
		dockerClient.Close()
		return nil, err
	}

	return &ContainerEnvironment{
		env:            env,
		projectDir:     projectDir,
		projectName:    projectName,
		detector:       detector,
		dockerClient:   dockerClient,
		imageTag:       imageTag,
		extraArgs:      extraArgs,
		dockerfilePath: result.Path,
	}, nil
}

// Close releases resources held by the container environment.
func (c *ContainerEnvironment) Close() error {
	if c.dockerClient != nil {
		return c.dockerClient.Close()
	}
	return nil
}

// Validate validates that the container environment is properly configured.
func (c *ContainerEnvironment) Validate() error {
	return c.detector.ValidateDockerfile(c.dockerfilePath)
}

// EnsureReady ensures the container image is built and available.
// Implements RuntimeEnvironment interface.
func (c *ContainerEnvironment) EnsureReady(ctx context.Context, output io.Writer) error {
	// Check if image already exists and is up-to-date
	imageTime, err := c.dockerClient.ImageCreatedTime(ctx, c.imageTag)
	if err != nil {
		return err
	}

	needsBuild := imageTime.IsZero() // Image doesn't exist

	// If image exists, check if Dockerfile is newer
	if !needsBuild {
		dockerfileInfo, err := os.Stat(c.dockerfilePath)
		if err != nil {
			return fmt.Errorf("failed to stat Dockerfile: %w", err)
		}
		if dockerfileInfo.ModTime().After(imageTime) {
			needsBuild = true
		}
	}

	if !needsBuild {
		return nil
	}

	// Build the image
	if output != nil {
		fmt.Fprintf(output, "Building image %s...\n", c.imageTag)
	}
	if err := c.dockerClient.BuildImage(ctx, c.dockerfilePath, c.imageTag, c.extraArgs, output); err != nil {
		return &ImageBuildError{
			ImageTag: c.imageTag,
			Message:  err.Error(),
		}
	}

	return nil
}

// RunCommand runs a command in the container environment and returns the result.
func (c *ContainerEnvironment) RunCommand(ctx context.Context, command []string) (*RunResult, error) {
	return c.dockerClient.RunCommand(ctx, c.imageTag, command, c.projectDir, c.extraArgs)
}

// RunCommandStreaming runs a command in the container and streams output.
func (c *ContainerEnvironment) RunCommandStreaming(ctx context.Context, command []string, stdout, stderr io.Writer) (int, error) {
	return c.dockerClient.RunCommandStreaming(ctx, c.imageTag, command, c.projectDir, c.extraArgs, stdout, stderr)
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

// ProjectDir returns the project directory.
func (c *ContainerEnvironment) ProjectDir() string {
	return c.projectDir
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
