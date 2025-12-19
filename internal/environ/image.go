package environ

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
)

// ImageBuilder builds container images from Dockerfiles.
type ImageBuilder struct {
	runtime    ast.Runtime
	runtimeCmd string
	extraArgs  []string
	runCommand func(name string, args ...string) ([]byte, error)
}

// NewImageBuilder creates a new ImageBuilder for the given runtime.
func NewImageBuilder(runtime ast.Runtime, runtimeCmd string, extraArgs []string) *ImageBuilder {
	return &ImageBuilder{
		runtime:    runtime,
		runtimeCmd: runtimeCmd,
		extraArgs:  extraArgs,
		runCommand: func(name string, args ...string) ([]byte, error) {
			return exec.Command(name, args...).Output()
		},
	}
}

// BuildCommand returns an exec.Cmd for building an image.
func (b *ImageBuilder) BuildCommand(dockerfilePath, imageTag string) *exec.Cmd {
	contextDir := filepath.Dir(dockerfilePath)

	args := []string{"build"}
	args = append(args, "-t", imageTag)
	args = append(args, "-f", dockerfilePath)

	// Add extra args (like --platform)
	args = append(args, b.extraArgs...)

	// Add context directory
	args = append(args, contextDir)

	return exec.Command(b.runtimeCmd, args...)
}

// ImageExists checks if an image exists locally.
func (b *ImageBuilder) ImageExists(imageTag string) (bool, error) {
	args := []string{"image", "inspect", imageTag}
	_, err := b.runCommand(b.runtimeCmd, args...)
	if err != nil {
		// Check if it's an exit error (image not found)
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		// Check if it's our custom error
		if _, ok := err.(*ImageNotFoundError); ok {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Build builds the image and returns an error if it fails.
func (b *ImageBuilder) Build(dockerfilePath, imageTag string) error {
	cmd := b.BuildCommand(dockerfilePath, imageTag)
	return cmd.Run()
}

// GenerateImageTag generates an image tag from project name and environment name.
func GenerateImageTag(project, envName string) string {
	// Normalize project name (lowercase, no spaces)
	project = strings.ToLower(strings.ReplaceAll(project, " ", "-"))

	if envName == "" {
		return project + ":latest"
	}

	// Normalize environment name
	envName = strings.ToLower(strings.ReplaceAll(envName, " ", "-"))
	return project + "-" + envName + ":latest"
}

// ParseExtraArgs parses the .args: directive value into command-line arguments.
func ParseExtraArgs(argsValue *ast.Value) []string {
	if argsValue == nil || len(argsValue.Parts) == 0 {
		return nil
	}

	// Extract literal text
	var text strings.Builder
	for _, part := range argsValue.Parts {
		if lit, ok := part.(*ast.LiteralValue); ok {
			text.WriteString(lit.Text)
		}
	}

	// Split on whitespace
	return strings.Fields(text.String())
}
