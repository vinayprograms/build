package environ

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/eval"
)

// LimaDetector detects and validates Lima VM configurations.
type LimaDetector struct{}

// NewLimaDetector creates a new LimaDetector.
func NewLimaDetector() *LimaDetector {
	return &LimaDetector{}
}

// LimaConfigResult holds the result of detecting a Lima configuration.
type LimaConfigResult struct {
	Path  string // Absolute path to the lima yaml file
	Found bool   // True if configuration was found
}

// DetectConfig searches for a Lima configuration in the given directory.
// If source is provided, it uses that path directly.
// Otherwise, it looks for lima.yaml in the project directory.
func (d *LimaDetector) DetectConfig(baseDir string, source *ast.Value, ctx *eval.Context) (LimaConfigResult, error) {
	result := LimaConfigResult{}

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
				return result, nil
			}
		}
	}

	// Check lima.yaml
	limaYamlPath := filepath.Join(baseDir, "lima.yaml")
	if _, err := os.Stat(limaYamlPath); err == nil {
		result.Path = limaYamlPath
		result.Found = true
		return result, nil
	}

	// Not found
	return result, nil
}

// LimaRunner handles running commands in a Lima VM.
type LimaRunner struct {
	projectDir string
	vmName     string
	configPath string
	args       []string // Extra args from .args:
	lookPath   func(name string) (string, error)
}

// NewLimaRunner creates a new LimaRunner.
func NewLimaRunner(projectDir, vmName string) *LimaRunner {
	return &LimaRunner{
		projectDir: projectDir,
		vmName:     vmName,
		lookPath:   exec.LookPath,
	}
}

// SetConfigPath sets the Lima configuration file path.
func (r *LimaRunner) SetConfigPath(path string) {
	r.configPath = path
}

// SetArgs sets extra arguments from .args: directive.
func (r *LimaRunner) SetArgs(args []string) {
	r.args = args
}

// CheckCLI checks if the Lima CLI is installed.
func (r *LimaRunner) CheckCLI() error {
	_, err := r.lookPath("limactl")
	if err != nil {
		return &BinaryNotFoundError{Name: "limactl"}
	}
	return nil
}

// Start starts the Lima VM.
func (r *LimaRunner) Start() error {
	if err := r.CheckCLI(); err != nil {
		return err
	}

	args := []string{"start"}
	if r.configPath != "" {
		args = append(args, "--name="+r.vmName, r.configPath)
	} else {
		args = append(args, r.vmName)
	}
	args = append(args, r.args...)

	cmd := exec.Command("limactl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Stop stops the Lima VM.
func (r *LimaRunner) Stop() error {
	if err := r.CheckCLI(); err != nil {
		return err
	}

	cmd := exec.Command("limactl", "stop", r.vmName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Exec executes a command inside the Lima VM.
func (r *LimaRunner) Exec(command string) error {
	if err := r.CheckCLI(); err != nil {
		return err
	}

	// Use lima shell to execute command
	// lima <vmname> -- sh -c <command>
	args := []string{r.vmName, "--", "sh", "-c", command}

	cmd := exec.Command("lima", args...)
	cmd.Dir = r.projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// OpenShell opens an interactive shell in the Lima VM.
func (r *LimaRunner) OpenShell() error {
	if err := r.CheckCLI(); err != nil {
		return err
	}

	cmd := exec.Command("lima", r.vmName)
	cmd.Dir = r.projectDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
