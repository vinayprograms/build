package environ

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/loft-sh/devpod/pkg/devcontainer/config"
	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/eval"
)

// DevcontainerEnvironment represents a devcontainer-based build environment.
// It uses the devpod config package for parsing devcontainer.json and
// the Docker SDK for building and running containers.
type DevcontainerEnvironment struct {
	env          *ast.Environment
	projectDir   string
	configPath   string
	devConfig    *config.DevContainerConfig
	dockerClient *DockerClient
	imageTag     string
	imageBuilt   bool
}

// NewDevcontainerEnvironment creates a new DevcontainerEnvironment.
func NewDevcontainerEnvironment(env *ast.Environment, projectDir, projectName string, ctx *eval.Context) (*DevcontainerEnvironment, error) {
	if env.Runtime == nil || *env.Runtime != ast.RuntimeDevcontainer {
		return nil, &InvalidRuntimeError{Message: "not a devcontainer runtime"}
	}

	// Determine config path from .source directive or detect automatically
	var relativePath string
	if env.Source != nil {
		sourcePath, err := ResolveSourcePath(".source:", env.Source, ctx)
		if err != nil {
			return nil, err
		}
		if sourcePath != "" {
			// Check if source is a directory - if so, append devcontainer.json
			fullPath := filepath.Join(projectDir, sourcePath)
			if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
				relativePath = filepath.Join(sourcePath, "devcontainer.json")
			} else {
				relativePath = sourcePath
			}
		}
	}

	// Parse devcontainer.json using devpod's config package
	devConfig, err := config.ParseDevContainerJSON(projectDir, relativePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse devcontainer.json: %w", err)
	}
	if devConfig == nil {
		return nil, fmt.Errorf("no devcontainer.json found in %s", projectDir)
	}

	// Create Docker client
	dockerClient, err := NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker: %w", err)
	}

	// Ping to ensure daemon is responsive
	if err := dockerClient.Ping(context.Background()); err != nil {
		dockerClient.Close()
		return nil, fmt.Errorf("Docker daemon not responding: %w", err)
	}

	// Generate image tag
	envName := ""
	if env.Name != nil {
		envName = *env.Name
	}
	imageTag := GenerateImageTag(projectName, envName)

	return &DevcontainerEnvironment{
		env:          env,
		projectDir:   projectDir,
		configPath:   devConfig.Origin,
		devConfig:    devConfig,
		dockerClient: dockerClient,
		imageTag:     imageTag,
	}, nil
}

// Close releases resources held by the devcontainer environment.
func (d *DevcontainerEnvironment) Close() error {
	if d.dockerClient != nil {
		return d.dockerClient.Close()
	}
	return nil
}

// Validate validates that the devcontainer environment is properly configured.
func (d *DevcontainerEnvironment) Validate() error {
	// Check that we have either an image or a Dockerfile
	if d.devConfig.Image == "" && d.devConfig.GetDockerfile() == "" {
		return fmt.Errorf("devcontainer.json must specify either 'image' or 'dockerFile'")
	}
	return nil
}

// EnsureReady ensures the devcontainer image is built and available.
func (d *DevcontainerEnvironment) EnsureReady(ctx context.Context, output io.Writer) error {
	if d.imageBuilt {
		return nil
	}

	// If using a pre-built image (no Dockerfile), check if it exists locally and pull if needed
	if d.devConfig.Image != "" && d.devConfig.GetDockerfile() == "" {
		exists, err := d.dockerClient.ImageExists(ctx, d.devConfig.Image)
		if err != nil {
			return fmt.Errorf("failed to check image: %w", err)
		}
		if !exists {
			if output != nil {
				fmt.Fprintf(output, "Pulling image %s...\n", d.devConfig.Image)
			}
			if err := d.dockerClient.PullImage(ctx, d.devConfig.Image, output); err != nil {
				return fmt.Errorf("failed to pull image %s: %w", d.devConfig.Image, err)
			}
		}
		d.imageBuilt = true
		return nil
	}

	// Build from Dockerfile
	dockerfile := d.devConfig.GetDockerfile()
	if dockerfile == "" {
		return fmt.Errorf("no Dockerfile specified in devcontainer.json")
	}

	// Resolve Dockerfile path relative to devcontainer.json location
	configDir := filepath.Dir(d.configPath)
	dockerfilePath := filepath.Join(configDir, dockerfile)

	// Check if image exists and is up-to-date
	imageTime, err := d.dockerClient.ImageCreatedTime(ctx, d.imageTag)
	if err != nil {
		return err
	}

	needsBuild := imageTime.IsZero() // Image doesn't exist

	// If image exists, check if Dockerfile or devcontainer.json is newer
	if !needsBuild {
		dockerfileInfo, err := os.Stat(dockerfilePath)
		if err != nil {
			return fmt.Errorf("failed to stat Dockerfile: %w", err)
		}
		if dockerfileInfo.ModTime().After(imageTime) {
			needsBuild = true
		}

		// Also check devcontainer.json modification time
		if !needsBuild {
			configInfo, err := os.Stat(d.configPath)
			if err == nil && configInfo.ModTime().After(imageTime) {
				needsBuild = true
			}
		}
	}

	if !needsBuild {
		d.imageBuilt = true
		return nil
	}

	// Get build context - use build.context if specified, otherwise use config directory
	contextDir := configDir
	if d.devConfig.Build != nil && d.devConfig.Build.Context != "" {
		contextDir = filepath.Join(configDir, d.devConfig.Build.Context)
	}

	// Build the image
	if output != nil {
		fmt.Fprintf(output, "Building image %s...\n", d.imageTag)
	}
	if err := d.dockerClient.BuildImageWithContext(ctx, dockerfilePath, contextDir, d.imageTag, nil, output); err != nil {
		return &ImageBuildError{
			ImageTag: d.imageTag,
			Message:  err.Error(),
		}
	}

	d.imageBuilt = true
	return nil
}

// RunCommand runs a command in the devcontainer and returns the result.
func (d *DevcontainerEnvironment) RunCommand(ctx context.Context, command []string) (*RunResult, error) {
	// Determine workspace folder inside container
	workspaceFolder := d.devConfig.WorkspaceFolder
	if workspaceFolder == "" {
		workspaceFolder = "/workspace"
	}

	// Use the pre-built image or the image from devcontainer.json
	imageTag := d.imageTag
	if d.devConfig.Image != "" && d.devConfig.GetDockerfile() == "" {
		imageTag = d.devConfig.Image
	}

	return d.dockerClient.RunCommandWithWorkdir(ctx, imageTag, command, d.projectDir, workspaceFolder, nil)
}

// RunCommandStreaming runs a command in the devcontainer and streams output.
func (d *DevcontainerEnvironment) RunCommandStreaming(ctx context.Context, command []string, stdout, stderr io.Writer) (int, error) {
	// Determine workspace folder inside container
	workspaceFolder := d.devConfig.WorkspaceFolder
	if workspaceFolder == "" {
		workspaceFolder = "/workspace"
	}

	// Use the pre-built image or the image from devcontainer.json
	imageTag := d.imageTag
	if d.devConfig.Image != "" && d.devConfig.GetDockerfile() == "" {
		imageTag = d.devConfig.Image
	}

	return d.dockerClient.RunCommandStreamingWithWorkdir(ctx, imageTag, command, d.projectDir, workspaceFolder, nil, stdout, stderr)
}

// RuntimeName returns the name of the runtime.
func (d *DevcontainerEnvironment) RuntimeName() string {
	return "devcontainer"
}

// ProjectDir returns the project directory.
func (d *DevcontainerEnvironment) ProjectDir() string {
	return d.projectDir
}
