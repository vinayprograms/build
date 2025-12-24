package environ

import (
	"context"
	"io"
)

// RuntimeEnvironment is the common interface for all runtime environments.
// It abstracts the execution of commands in different runtime contexts
// (Docker, Podman, Devcontainer, Nix, Lima).
type RuntimeEnvironment interface {
	// Close releases resources held by the environment.
	Close() error

	// Validate validates that the environment is properly configured.
	Validate() error

	// EnsureReady ensures the environment is ready to execute commands.
	// For containers, this builds the image. For devcontainers, this runs 'up'.
	// For nix/lima, this may be a no-op or verify the environment exists.
	EnsureReady(ctx context.Context) error

	// RunCommand runs a command in the environment and returns the result.
	RunCommand(ctx context.Context, command []string) (*RunResult, error)

	// RunCommandStreaming runs a command and streams output to the provided writers.
	RunCommandStreaming(ctx context.Context, command []string, stdout, stderr io.Writer) (int, error)

	// RuntimeName returns the name of the runtime (docker, podman, devcontainer, nix, lima).
	RuntimeName() string

	// ProjectDir returns the project directory.
	ProjectDir() string
}
