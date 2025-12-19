package environ

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// DevcontainerDetector detects and validates devcontainer configurations.
type DevcontainerDetector struct{}

// NewDevcontainerDetector creates a new DevcontainerDetector.
func NewDevcontainerDetector() *DevcontainerDetector {
	return &DevcontainerDetector{}
}

// DevcontainerConfigResult holds the result of detecting a devcontainer configuration.
type DevcontainerConfigResult struct {
	Path  string // Absolute path to devcontainer.json
	Found bool   // True if configuration was found
}

// DevcontainerConfig represents a parsed devcontainer.json file.
type DevcontainerConfig struct {
	Name              string                   `json:"name,omitempty"`
	Image             string                   `json:"image,omitempty"`
	Dockerfile        string                   `json:"dockerFile,omitempty"`
	DockerComposeFile string                   `json:"dockerComposeFile,omitempty"`
	Service           string                   `json:"service,omitempty"`
	Build             *DevcontainerBuildConfig `json:"build,omitempty"`
	WorkspaceFolder   string                   `json:"workspaceFolder,omitempty"`
	RemoteUser        string                   `json:"remoteUser,omitempty"`
}

// DevcontainerBuildConfig represents the build section of a devcontainer.json.
type DevcontainerBuildConfig struct {
	Dockerfile string `json:"dockerfile,omitempty"`
	Context    string `json:"context,omitempty"`
}

// GetImageOrBuildSource returns a description of the image or build source.
// Format: "image:<name>", "dockerfile:<path>", "compose:<file>:<service>", or empty string.
func (c *DevcontainerConfig) GetImageOrBuildSource() string {
	if c.Image != "" {
		return "image:" + c.Image
	}
	if c.Dockerfile != "" {
		return "dockerfile:" + c.Dockerfile
	}
	if c.Build != nil && c.Build.Dockerfile != "" {
		return "dockerfile:" + c.Build.Dockerfile
	}
	if c.DockerComposeFile != "" && c.Service != "" {
		return fmt.Sprintf("compose:%s:%s", c.DockerComposeFile, c.Service)
	}
	return ""
}

// DetectConfig searches for a devcontainer configuration in the given directory.
// It looks for:
// 1. .devcontainer/devcontainer.json (preferred)
// 2. devcontainer.json in the root directory
func (d *DevcontainerDetector) DetectConfig(baseDir string) (DevcontainerConfigResult, error) {
	result := DevcontainerConfigResult{}

	// Check .devcontainer/devcontainer.json first (takes priority)
	devcontainerDir := filepath.Join(baseDir, ".devcontainer")
	dirConfigPath := filepath.Join(devcontainerDir, "devcontainer.json")
	if _, err := os.Stat(dirConfigPath); err == nil {
		result.Path = dirConfigPath
		result.Found = true
		return result, nil
	}

	// Check root devcontainer.json
	rootConfigPath := filepath.Join(baseDir, "devcontainer.json")
	if _, err := os.Stat(rootConfigPath); err == nil {
		result.Path = rootConfigPath
		result.Found = true
		return result, nil
	}

	// Not found
	return result, nil
}

// LoadConfig loads and parses a devcontainer.json file.
func (d *DevcontainerDetector) LoadConfig(path string) (*DevcontainerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ParseDevcontainerConfig(data)
}

// ParseDevcontainerConfig parses a devcontainer.json content.
func ParseDevcontainerConfig(data []byte) (*DevcontainerConfig, error) {
	var config DevcontainerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// DevcontainerRunner handles running commands in a devcontainer.
type DevcontainerRunner struct {
	projectDir string
	configPath string
	lookPath   func(name string) (string, error)
}

// NewDevcontainerRunner creates a new DevcontainerRunner.
func NewDevcontainerRunner(projectDir string) *DevcontainerRunner {
	return &DevcontainerRunner{
		projectDir: projectDir,
		lookPath:   exec.LookPath,
	}
}

// SetConfigPath sets the path to the devcontainer.json file.
func (r *DevcontainerRunner) SetConfigPath(path string) {
	r.configPath = path
}

// CheckCLI checks if the devcontainer CLI is installed.
func (r *DevcontainerRunner) CheckCLI() error {
	_, err := r.lookPath("devcontainer")
	if err != nil {
		return &BinaryNotFoundError{Name: "devcontainer"}
	}
	return nil
}

// Up starts the devcontainer.
func (r *DevcontainerRunner) Up() error {
	if err := r.CheckCLI(); err != nil {
		return err
	}

	args := []string{"up", "--workspace-folder", r.projectDir}
	if r.configPath != "" {
		args = append(args, "--config", r.configPath)
	}

	cmd := exec.Command("devcontainer", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Exec executes a command inside the devcontainer.
func (r *DevcontainerRunner) Exec(command string) error {
	if err := r.CheckCLI(); err != nil {
		return err
	}

	args := []string{"exec", "--workspace-folder", r.projectDir}
	if r.configPath != "" {
		args = append(args, "--config", r.configPath)
	}
	args = append(args, command)

	cmd := exec.Command("devcontainer", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// OpenShell opens an interactive shell in the devcontainer.
func (r *DevcontainerRunner) OpenShell() error {
	if err := r.CheckCLI(); err != nil {
		return err
	}

	args := []string{"exec", "--workspace-folder", r.projectDir}
	if r.configPath != "" {
		args = append(args, "--config", r.configPath)
	}
	args = append(args, "/bin/bash")

	cmd := exec.Command("devcontainer", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
