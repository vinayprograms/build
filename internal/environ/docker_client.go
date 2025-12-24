package environ

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// DockerClient wraps the Docker SDK client for container operations.
type DockerClient struct {
	cli *client.Client
}

// NewDockerClient creates a new Docker client using environment settings.
func NewDockerClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	return &DockerClient{cli: cli}, nil
}

// Close closes the Docker client connection.
func (d *DockerClient) Close() error {
	return d.cli.Close()
}

// Ping checks if the Docker daemon is running.
func (d *DockerClient) Ping(ctx context.Context) error {
	_, err := d.cli.Ping(ctx)
	return err
}

// ImageExists checks if an image exists locally.
func (d *DockerClient) ImageExists(ctx context.Context, imageTag string) (bool, error) {
	_, _, err := d.cli.ImageInspectWithRaw(ctx, imageTag)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ImageCreatedTime returns the creation time of a Docker image.
// Returns zero time if the image doesn't exist.
func (d *DockerClient) ImageCreatedTime(ctx context.Context, imageTag string) (time.Time, error) {
	inspect, _, err := d.cli.ImageInspectWithRaw(ctx, imageTag)
	if err != nil {
		if client.IsErrNotFound(err) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, inspect.Created)
}

// PullImage pulls a Docker image from a registry.
// If output is non-nil, pull progress is written to it.
func (d *DockerClient) PullImage(ctx context.Context, imageRef string, output io.Writer) error {
	reader, err := d.cli.ImagePull(ctx, imageRef, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()

	// Stream pull progress or discard
	if output != nil {
		// Docker returns JSON progress messages, decode and display them
		decoder := json.NewDecoder(reader)
		for {
			var message struct {
				Status   string `json:"status"`
				Progress string `json:"progress"`
				Error    string `json:"error"`
			}
			if err := decoder.Decode(&message); err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("failed to read pull output: %w", err)
			}
			if message.Error != "" {
				return fmt.Errorf("pull failed: %s", message.Error)
			}
			if message.Progress != "" {
				fmt.Fprintf(output, "%s %s\n", message.Status, message.Progress)
			} else if message.Status != "" {
				fmt.Fprintf(output, "%s\n", message.Status)
			}
		}
	} else {
		// Read the output to completion (required for the pull to complete)
		_, err = io.Copy(io.Discard, reader)
		if err != nil {
			return fmt.Errorf("failed to read pull output: %w", err)
		}
	}

	return nil
}

// BuildImage builds a Docker image from a Dockerfile.
// dockerfilePath is the path to the Dockerfile.
// imageTag is the tag to apply to the built image.
// extraArgs are additional build arguments (e.g., --platform).
// If output is non-nil, build progress is written to it.
func (d *DockerClient) BuildImage(ctx context.Context, dockerfilePath, imageTag string, extraArgs []string, output io.Writer) error {
	contextDir := filepath.Dir(dockerfilePath)
	dockerfileName := filepath.Base(dockerfilePath)

	// Create a tar archive of the build context
	buildContext, err := createBuildContext(contextDir)
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}

	// Parse extra args for platform
	var platform string
	for i, arg := range extraArgs {
		if arg == "--platform" && i+1 < len(extraArgs) {
			platform = extraArgs[i+1]
			break
		}
		if strings.HasPrefix(arg, "--platform=") {
			platform = strings.TrimPrefix(arg, "--platform=")
			break
		}
	}

	buildOptions := build.ImageBuildOptions{
		Tags:       []string{imageTag},
		Dockerfile: dockerfileName,
		Remove:     true,
		Platform:   platform,
	}

	resp, err := d.cli.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return fmt.Errorf("failed to start image build: %w", err)
	}
	defer resp.Body.Close()

	// Read build output and check for errors
	// Docker returns a JSON stream with build messages
	decoder := json.NewDecoder(resp.Body)
	for {
		var message struct {
			Stream string `json:"stream"`
			Error  string `json:"error"`
		}
		if err := decoder.Decode(&message); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read build output: %w", err)
		}
		if message.Error != "" {
			return fmt.Errorf("build failed: %s", message.Error)
		}
		if output != nil && message.Stream != "" {
			fmt.Fprint(output, message.Stream)
		}
	}

	return nil
}

// RunResult contains the result of running a command in a container.
type RunResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// RunCommand runs a command in a new container and returns the result.
// The container is automatically removed after the command completes.
func (d *DockerClient) RunCommand(ctx context.Context, imageTag string, command []string, workspaceDir string, extraArgs []string) (*RunResult, error) {
	// Create container
	containerConfig := &container.Config{
		Image:      imageTag,
		Cmd:        command,
		WorkingDir: "/workspace",
		Tty:        false,
	}

	hostConfig := &container.HostConfig{
		AutoRemove: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: workspaceDir,
				Target: "/workspace",
			},
		},
	}

	// Parse extra args for additional host config options
	for i, arg := range extraArgs {
		if arg == "--platform" && i+1 < len(extraArgs) {
			// Platform is handled at image level, skip
			continue
		}
	}

	resp, err := d.cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	containerID := resp.ID

	// Attach to container to capture output
	attachResp, err := d.cli.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to container: %w", err)
	}
	defer attachResp.Close()

	// Start container
	if err := d.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Read stdout/stderr
	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, attachResp.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read container output: %w", err)
	}

	// Wait for container to finish
	statusCh, errCh := d.cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return nil, fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		return &RunResult{
			ExitCode: int(status.StatusCode),
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
		}, nil
	}

	return nil, fmt.Errorf("container wait returned unexpectedly")
}

// RunCommandStreaming runs a command in a container and streams output to the provided writers.
func (d *DockerClient) RunCommandStreaming(ctx context.Context, imageTag string, command []string, workspaceDir string, extraArgs []string, stdout, stderr io.Writer) (int, error) {
	// Create container
	containerConfig := &container.Config{
		Image:      imageTag,
		Cmd:        command,
		WorkingDir: "/workspace",
		Tty:        false,
	}

	hostConfig := &container.HostConfig{
		AutoRemove: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: workspaceDir,
				Target: "/workspace",
			},
		},
	}

	resp, err := d.cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return -1, fmt.Errorf("failed to create container: %w", err)
	}
	containerID := resp.ID

	// Attach to container
	attachResp, err := d.cli.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to attach to container: %w", err)
	}
	defer attachResp.Close()

	// Start container
	if err := d.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return -1, fmt.Errorf("failed to start container: %w", err)
	}

	// Stream output
	_, err = stdcopy.StdCopy(stdout, stderr, attachResp.Reader)
	if err != nil {
		return -1, fmt.Errorf("failed to copy container output: %w", err)
	}

	// Wait for container to finish
	statusCh, errCh := d.cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return -1, fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		return int(status.StatusCode), nil
	}

	return -1, fmt.Errorf("container wait returned unexpectedly")
}

// createBuildContext creates a tar archive of the build context directory.
func createBuildContext(contextDir string) (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	err := filepath.Walk(contextDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Get relative path
		relPath, err := filepath.Rel(contextDir, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			header.Linkname = link
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Write file content if it's a regular file
		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return &buf, nil
}

// BuildImageWithContext builds a Docker image with a custom context directory.
// dockerfilePath is the path to the Dockerfile.
// contextDir is the build context directory.
// imageTag is the tag to apply to the built image.
// extraArgs are additional build arguments.
// If output is non-nil, build progress is written to it.
func (d *DockerClient) BuildImageWithContext(ctx context.Context, dockerfilePath, contextDir, imageTag string, extraArgs []string, output io.Writer) error {
	dockerfileName := filepath.Base(dockerfilePath)

	// If Dockerfile is not in the context directory, we need to handle this
	// For now, assume Dockerfile is in or relative to context
	relDockerfile, err := filepath.Rel(contextDir, dockerfilePath)
	if err != nil {
		// Dockerfile is outside context, use absolute path
		relDockerfile = dockerfileName
	}

	// Create a tar archive of the build context
	buildContext, err := createBuildContext(contextDir)
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}

	// Parse extra args for platform
	var platform string
	for i, arg := range extraArgs {
		if arg == "--platform" && i+1 < len(extraArgs) {
			platform = extraArgs[i+1]
			break
		}
		if strings.HasPrefix(arg, "--platform=") {
			platform = strings.TrimPrefix(arg, "--platform=")
			break
		}
	}

	buildOptions := build.ImageBuildOptions{
		Tags:       []string{imageTag},
		Dockerfile: relDockerfile,
		Remove:     true,
		Platform:   platform,
	}

	resp, err := d.cli.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return fmt.Errorf("failed to start image build: %w", err)
	}
	defer resp.Body.Close()

	// Read build output and check for errors
	decoder := json.NewDecoder(resp.Body)
	for {
		var message struct {
			Stream string `json:"stream"`
			Error  string `json:"error"`
		}
		if err := decoder.Decode(&message); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read build output: %w", err)
		}
		if message.Error != "" {
			return fmt.Errorf("build failed: %s", message.Error)
		}
		if output != nil && message.Stream != "" {
			fmt.Fprint(output, message.Stream)
		}
	}

	return nil
}

// RunCommandWithWorkdir runs a command in a container with a custom working directory.
func (d *DockerClient) RunCommandWithWorkdir(ctx context.Context, imageTag string, command []string, hostDir, containerWorkdir string, extraArgs []string) (*RunResult, error) {
	containerConfig := &container.Config{
		Image:      imageTag,
		Cmd:        command,
		WorkingDir: containerWorkdir,
		Tty:        false,
	}

	hostConfig := &container.HostConfig{
		AutoRemove: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: hostDir,
				Target: containerWorkdir,
			},
		},
	}

	resp, err := d.cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	containerID := resp.ID

	attachResp, err := d.cli.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to container: %w", err)
	}
	defer attachResp.Close()

	if err := d.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, attachResp.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read container output: %w", err)
	}

	statusCh, errCh := d.cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return nil, fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		return &RunResult{
			ExitCode: int(status.StatusCode),
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
		}, nil
	}

	return nil, fmt.Errorf("container wait returned unexpectedly")
}

// RunCommandStreamingWithWorkdir runs a command in a container with a custom working directory and streams output.
func (d *DockerClient) RunCommandStreamingWithWorkdir(ctx context.Context, imageTag string, command []string, hostDir, containerWorkdir string, extraArgs []string, stdout, stderr io.Writer) (int, error) {
	containerConfig := &container.Config{
		Image:      imageTag,
		Cmd:        command,
		WorkingDir: containerWorkdir,
		Tty:        false,
	}

	hostConfig := &container.HostConfig{
		AutoRemove: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: hostDir,
				Target: containerWorkdir,
			},
		},
	}

	resp, err := d.cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return -1, fmt.Errorf("failed to create container: %w", err)
	}
	containerID := resp.ID

	attachResp, err := d.cli.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to attach to container: %w", err)
	}
	defer attachResp.Close()

	if err := d.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return -1, fmt.Errorf("failed to start container: %w", err)
	}

	_, err = stdcopy.StdCopy(stdout, stderr, attachResp.Reader)
	if err != nil {
		return -1, fmt.Errorf("failed to copy container output: %w", err)
	}

	statusCh, errCh := d.cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return -1, fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		return int(status.StatusCode), nil
	}

	return -1, fmt.Errorf("container wait returned unexpectedly")
}
