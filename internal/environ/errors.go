package environ

import "fmt"

// BinaryNotFoundError indicates a required binary was not found in PATH.
type BinaryNotFoundError struct {
	Name string // Binary name
}

func (e *BinaryNotFoundError) Error() string {
	return fmt.Sprintf("required binary '%s' not found in PATH", e.Name)
}

// VersionMismatchError indicates the found version doesn't match the requirement.
type VersionMismatchError struct {
	Name     string // Binary name
	Required string // Required version string
	Detected string // Detected version string
}

func (e *VersionMismatchError) Error() string {
	return fmt.Sprintf("version mismatch for '%s': required %s, found %s",
		e.Name, e.Required, e.Detected)
}

// VersionDetectionError indicates the version could not be detected.
type VersionDetectionError struct {
	Name    string // Binary name
	Message string // Error message
}

func (e *VersionDetectionError) Error() string {
	return fmt.Sprintf("unable to detect version for '%s': %s", e.Name, e.Message)
}

// EnvironmentNotFoundError indicates a named environment was not found.
type EnvironmentNotFoundError struct {
	Name string // Environment name
}

func (e *EnvironmentNotFoundError) Error() string {
	return fmt.Sprintf("environment '%s' not found", e.Name)
}

// NoDefaultEnvironmentError indicates no default environment is defined
// when only named environments exist.
type NoDefaultEnvironmentError struct {
	Available []string // List of available environment names
}

func (e *NoDefaultEnvironmentError) Error() string {
	if len(e.Available) == 0 {
		return "no environment defined"
	}
	return fmt.Sprintf("no default environment; use --env with one of: %v", e.Available)
}

// NoSourceError indicates .source: directive is missing for a container runtime.
type NoSourceError struct {
	Runtime string // Runtime name (docker, podman)
}

func (e *NoSourceError) Error() string {
	return fmt.Sprintf("%s environment requires .source: directive specifying Dockerfile path", e.Runtime)
}

// InvalidRuntimeError indicates an invalid or unsupported runtime.
type InvalidRuntimeError struct {
	Message string
}

func (e *InvalidRuntimeError) Error() string {
	return fmt.Sprintf("invalid runtime: %s", e.Message)
}

// DockerfileNotFoundError indicates the specified Dockerfile was not found.
type DockerfileNotFoundError struct {
	Path string // Path to the missing Dockerfile
}

func (e *DockerfileNotFoundError) Error() string {
	return fmt.Sprintf("Dockerfile not found: %s", e.Path)
}

// InvalidDockerfileError indicates the Dockerfile is malformed.
type InvalidDockerfileError struct {
	Path   string // Path to the Dockerfile
	Reason string // Reason for invalidation
}

func (e *InvalidDockerfileError) Error() string {
	return fmt.Sprintf("invalid Dockerfile %s: %s", e.Path, e.Reason)
}

// ImageNotFoundError indicates a container image was not found locally.
type ImageNotFoundError struct {
	Name string // Image name/tag
}

func (e *ImageNotFoundError) Error() string {
	return fmt.Sprintf("image not found: %s", e.Name)
}

// ImageBuildError indicates an error during image build.
type ImageBuildError struct {
	ImageTag string
	Message  string
}

func (e *ImageBuildError) Error() string {
	return fmt.Sprintf("failed to build image %s: %s", e.ImageTag, e.Message)
}

// ContainerRunError indicates an error running a container.
type ContainerRunError struct {
	ImageTag string
	Message  string
}

func (e *ContainerRunError) Error() string {
	return fmt.Sprintf("failed to run container from %s: %s", e.ImageTag, e.Message)
}
