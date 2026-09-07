package environ

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/eval"
)

// NixType represents the type of Nix configuration.
type NixType int

const (
	NixTypeShell NixType = iota // shell.nix
	NixTypeFlake                // flake.nix
)

// String returns the string representation of the NixType.
func (t NixType) String() string {
	switch t {
	case NixTypeShell:
		return "shell.nix"
	case NixTypeFlake:
		return "flake.nix"
	default:
		return "unknown"
	}
}

// NixDetector detects and validates Nix configurations.
type NixDetector struct{}

// NewNixDetector creates a new NixDetector.
func NewNixDetector() *NixDetector {
	return &NixDetector{}
}

// NixConfigResult holds the result of detecting a Nix configuration.
type NixConfigResult struct {
	Path  string  // Absolute path to the nix file
	Found bool    // True if configuration was found
	Type  NixType // Type of nix file (shell.nix or flake.nix)
}

// DetectConfig searches for a Nix configuration in the given directory.
// If source is provided, it uses that path directly.
// Otherwise, it looks for:
// 1. shell.nix (preferred)
// 2. flake.nix
func (d *NixDetector) DetectConfig(baseDir string, source *ast.Value, ctx *eval.Context) (NixConfigResult, error) {
	result := NixConfigResult{}

	// If source is provided, use that path directly
	if source != nil {
		sourcePath, err := ResolveSourcePath(".source:", source, ctx)
		if err != nil {
			return result, err
		}
		if sourcePath != "" {
			if !filepath.IsAbs(sourcePath) {
				sourcePath = filepath.Join(baseDir, sourcePath)
			}
			sourcePath = filepath.Clean(sourcePath)

			if _, err := os.Stat(sourcePath); err == nil {
				result.Path = sourcePath
				result.Found = true
				// Determine type based on filename
				if filepath.Base(sourcePath) == "flake.nix" {
					result.Type = NixTypeFlake
				} else {
					result.Type = NixTypeShell
				}
				return result, nil
			}
		}
	}

	// Check shell.nix first (takes priority)
	shellNixPath := filepath.Join(baseDir, "shell.nix")
	if _, err := os.Stat(shellNixPath); err == nil {
		result.Path = shellNixPath
		result.Found = true
		result.Type = NixTypeShell
		return result, nil
	}

	// Check flake.nix
	flakeNixPath := filepath.Join(baseDir, "flake.nix")
	if _, err := os.Stat(flakeNixPath); err == nil {
		result.Path = flakeNixPath
		result.Found = true
		result.Type = NixTypeFlake
		return result, nil
	}

	// Not found
	return result, nil
}

// NixRunner handles running commands in a Nix environment.
type NixRunner struct {
	projectDir string
	configPath string
	nixType    NixType
	args       []string // Extra args from .args:
	lookPath   func(name string) (string, error)
}

// NewNixRunner creates a new NixRunner.
func NewNixRunner(projectDir string) *NixRunner {
	return &NixRunner{
		projectDir: projectDir,
		lookPath:   exec.LookPath,
	}
}

// SetConfig sets the nix configuration path and type.
func (r *NixRunner) SetConfig(path string, nixType NixType) {
	r.configPath = path
	r.nixType = nixType
}

// SetArgs sets extra arguments from .args: directive.
func (r *NixRunner) SetArgs(args []string) {
	r.args = args
}

// CheckCLI checks if the nix CLI is installed.
func (r *NixRunner) CheckCLI() error {
	_, err := r.lookPath("nix-shell")
	if err != nil {
		return &BinaryNotFoundError{Name: "nix-shell"}
	}
	return nil
}

// Exec executes a command inside the nix environment.
func (r *NixRunner) Exec(command string) error {
	if err := r.CheckCLI(); err != nil {
		return err
	}

	var args []string

	if r.nixType == NixTypeFlake {
		// For flakes, use: nix develop -c <command>
		args = []string{"develop"}
		if r.configPath != "" {
			args = append(args, filepath.Dir(r.configPath))
		}
		args = append(args, r.args...)
		args = append(args, "-c", "sh", "-c", command)

		cmd := exec.Command("nix", args...)
		cmd.Dir = r.projectDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// For shell.nix, use: nix-shell --run <command>
	args = []string{}
	if r.configPath != "" {
		args = append(args, r.configPath)
	}
	args = append(args, r.args...)
	args = append(args, "--run", command)

	cmd := exec.Command("nix-shell", args...)
	cmd.Dir = r.projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// OpenShell opens an interactive shell in the nix environment.
func (r *NixRunner) OpenShell() error {
	if err := r.CheckCLI(); err != nil {
		return err
	}

	var args []string

	if r.nixType == NixTypeFlake {
		// For flakes, use: nix develop
		args = []string{"develop"}
		if r.configPath != "" {
			args = append(args, filepath.Dir(r.configPath))
		}
		args = append(args, r.args...)

		cmd := exec.Command("nix", args...)
		cmd.Dir = r.projectDir
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// For shell.nix, use: nix-shell
	args = []string{}
	if r.configPath != "" {
		args = append(args, r.configPath)
	}
	args = append(args, r.args...)

	cmd := exec.Command("nix-shell", args...)
	cmd.Dir = r.projectDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
