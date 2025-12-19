package environ

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
)

// ContainerDetector detects and validates container environments.
type ContainerDetector struct {
	lookPath func(name string) (string, error)
}

// NewContainerDetector creates a new ContainerDetector.
func NewContainerDetector() *ContainerDetector {
	return &ContainerDetector{
		lookPath: exec.LookPath,
	}
}

// DockerfileResult holds the result of detecting a Dockerfile.
type DockerfileResult struct {
	Path   string // Absolute path to the Dockerfile
	Exists bool   // True if the file exists
}

// DetectDockerfile locates the Dockerfile specified in an environment's .source: directive.
func (c *ContainerDetector) DetectDockerfile(env *ast.Environment, baseDir string) (DockerfileResult, error) {
	result := DockerfileResult{}

	// Validate runtime is a container runtime
	if env.Runtime == nil {
		return result, &InvalidRuntimeError{Message: "runtime not specified"}
	}
	if !isContainerRuntime(*env.Runtime) {
		return result, &InvalidRuntimeError{Message: "not a container runtime"}
	}

	// Validate source is specified
	if env.Source == nil {
		return result, &NoSourceError{Runtime: runtimeName(*env.Runtime)}
	}

	// Extract the source path (only literal paths supported for now)
	sourcePath := extractLiteralPath(env.Source)
	if sourcePath == "" {
		return result, &NoSourceError{Runtime: runtimeName(*env.Runtime)}
	}

	// Resolve relative paths
	if !filepath.IsAbs(sourcePath) {
		sourcePath = filepath.Join(baseDir, sourcePath)
	}
	sourcePath = filepath.Clean(sourcePath)
	result.Path = sourcePath

	// Check if file exists
	if _, err := os.Stat(sourcePath); err != nil {
		if os.IsNotExist(err) {
			return result, &DockerfileNotFoundError{Path: sourcePath}
		}
		return result, err
	}

	result.Exists = true
	return result, nil
}

// ValidateDockerfile checks that a Dockerfile is valid.
// Currently only checks that it contains a FROM instruction.
func (c *ContainerDetector) ValidateDockerfile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	hasFrom := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Check for FROM instruction (case-insensitive)
		upper := strings.ToUpper(line)
		if strings.HasPrefix(upper, "FROM ") || upper == "FROM" {
			hasFrom = true
			break
		}
		// ARG is allowed before FROM
		if strings.HasPrefix(upper, "ARG ") || upper == "ARG" {
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if !hasFrom {
		return &InvalidDockerfileError{
			Path:   path,
			Reason: "missing FROM instruction",
		}
	}

	return nil
}

// FindRuntime finds the container runtime binary in PATH.
func (c *ContainerDetector) FindRuntime(runtime ast.Runtime) (string, error) {
	var name string
	switch runtime {
	case ast.RuntimeDocker:
		name = "docker"
	case ast.RuntimePodman:
		name = "podman"
	default:
		return "", &InvalidRuntimeError{Message: "not a container runtime"}
	}

	path, err := c.lookPath(name)
	if err != nil {
		return "", &BinaryNotFoundError{Name: name}
	}
	return path, nil
}

// isContainerRuntime returns true if the runtime is a container runtime.
func isContainerRuntime(r ast.Runtime) bool {
	return r == ast.RuntimeDocker || r == ast.RuntimePodman
}

// runtimeName returns the string name for a runtime.
func runtimeName(r ast.Runtime) string {
	switch r {
	case ast.RuntimeDocker:
		return "docker"
	case ast.RuntimePodman:
		return "podman"
	default:
		return "unknown"
	}
}

// extractLiteralPath extracts a literal path from a Value.
// Returns empty string if the value contains interpolations.
func extractLiteralPath(v *ast.Value) string {
	if v == nil || len(v.Parts) == 0 {
		return ""
	}
	var path strings.Builder
	for _, part := range v.Parts {
		if lit, ok := part.(*ast.LiteralValue); ok {
			path.WriteString(lit.Text)
		} else {
			// Contains non-literal parts (interpolation, function call)
			// For now, we don't support these in .source: paths
			return ""
		}
	}
	return path.String()
}
