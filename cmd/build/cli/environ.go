package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/environ"
)

// checkEnvironment verifies the requirements for an environment.
// If envName is empty, checks the default environment.
// If showInstall is true, shows install suggestions for missing binaries.
// buildfileDir is the directory containing the Buildfile (for resolving relative paths).
// Returns exit code.
func checkEnvironment(result BuildfileResult, envName, buildfileDir string, verbose, showInstall bool) int {
	envs := GetEnvironments(result)

	// Find the selected environment
	var selectedEnv *ast.Environment
	var selectedName string

	if envName == "" {
		// Look for default (unnamed) environment
		for _, env := range envs {
			if env.Name == nil {
				selectedEnv = env
				selectedName = "(default)"
				break
			}
		}
		if selectedEnv == nil {
			// No default environment
			if len(envs) > 0 {
				// Named environments exist, require --env
				fmt.Fprintln(os.Stderr, "error: no default environment; use --env with one of:")
				for _, env := range envs {
					if env.Name != nil {
						fmt.Fprintf(os.Stderr, "  %s\n", *env.Name)
					}
				}
				return exitEnvError
			}
			// No environments defined - bare environment with no requirements
			fmt.Println("No environment defined (bare environment)")
			fmt.Println("All requirements satisfied (none specified)")
			return exitSuccess
		}
	} else {
		// Look for named environment
		for _, env := range envs {
			if env.Name != nil && *env.Name == envName {
				selectedEnv = env
				selectedName = envName
				break
			}
		}
		if selectedEnv == nil {
			fmt.Fprintf(os.Stderr, "error: environment '%s' not found\n", envName)
			if len(envs) > 0 {
				fmt.Fprintln(os.Stderr, "available environments:")
				for _, env := range envs {
					if env.Name == nil {
						fmt.Fprintln(os.Stderr, "  (default)")
					} else {
						fmt.Fprintf(os.Stderr, "  %s\n", *env.Name)
					}
				}
			}
			return exitEnvError
		}
	}

	// Print environment info
	fmt.Printf("Checking environment: %s\n", selectedName)
	if selectedEnv.Runtime != nil {
		fmt.Printf("Runtime: %s\n", selectedEnv.Runtime.String())
	} else {
		fmt.Println("Runtime: bare")
	}

	// Check if it's a bare environment
	isBare := selectedEnv.Runtime == nil || *selectedEnv.Runtime == ast.RuntimeBare

	if !isBare {
		// Container environments - validate Dockerfile and runtime
		if *selectedEnv.Runtime == ast.RuntimeDocker || *selectedEnv.Runtime == ast.RuntimePodman {
			return checkContainerEnvironment(selectedEnv, buildfileDir, verbose, showInstall)
		}

		// Devcontainer environment
		if *selectedEnv.Runtime == ast.RuntimeDevcontainer {
			return checkDevcontainerEnvironment(selectedEnv, buildfileDir, verbose, showInstall)
		}

		// Nix environment
		if *selectedEnv.Runtime == ast.RuntimeNix {
			return checkNixEnvironment(selectedEnv, buildfileDir, verbose, showInstall)
		}

		// Lima environment
		if *selectedEnv.Runtime == ast.RuntimeLima {
			return checkLimaEnvironment(selectedEnv, buildfileDir, verbose, showInstall)
		}

		// Other non-bare environments - just report status for now
		fmt.Printf("Source: %s\n", valueToTextSimple(selectedEnv.Source))
		if selectedEnv.Args != nil {
			fmt.Printf("Args: %s\n", valueToTextSimple(selectedEnv.Args))
		}
		fmt.Println("\nThis environment type checking not yet implemented")
		return exitSuccess
	}

	// Check requirements for bare environment
	if len(selectedEnv.Requires) == 0 {
		fmt.Println("\nNo requirements specified")
		return exitSuccess
	}

	fmt.Printf("\nChecking %d requirement(s)...\n", len(selectedEnv.Requires))

	results := CheckEnvironmentRequirements(selectedEnv, true) // with version checking
	hasErrors := false

	// Detect package manager for install suggestions
	var pm PackageManager
	if showInstall {
		pm = DetectPackageManager()
	}

	for _, r := range results {
		status := "✓"
		if r.Error() != nil {
			status = "✗"
			hasErrors = true
		}

		if verbose {
			fmt.Printf("  %s %s\n", status, r.String())
			if r.Path() != "" {
				fmt.Printf("      path: %s\n", r.Path())
			}
		} else {
			fmt.Printf("  %s %s\n", status, r.String())
		}

		// Show install suggestion if binary not found and showInstall is enabled
		if showInstall && !r.Found() && pm != nil {
			suggestion := GetInstallSuggestion(r.Name(), pm)
			if suggestion != "" {
				fmt.Printf("      install: %s\n", suggestion)
			}
		}
	}

	if hasErrors {
		fmt.Println("\nSome requirements are not met")
		if showInstall && pm == nil {
			fmt.Println("(Unable to detect package manager for install suggestions)")
		}
		return exitEnvError
	}

	fmt.Println("\nAll requirements satisfied")
	return exitSuccess
}

// listEnvironments lists all defined environments.
func listEnvironments(result BuildfileResult) int {
	envs := GetEnvironments(result)

	if len(envs) == 0 {
		fmt.Println("No environments defined")
		return exitSuccess
	}

	fmt.Printf("Available environments (%d):\n", len(envs))

	for _, env := range envs {
		name := "(default)"
		if env.Name != nil {
			name = *env.Name
		}

		runtime := "bare"
		if env.Runtime != nil {
			runtime = env.Runtime.String()
		}

		reqCount := len(env.Requires)
		if reqCount == 0 {
			fmt.Printf("  %-20s  %-15s\n", name, runtime)
		} else if reqCount == 1 {
			fmt.Printf("  %-20s  %-15s  (1 requirement)\n", name, runtime)
		} else {
			fmt.Printf("  %-20s  %-15s  (%d requirements)\n", name, runtime, reqCount)
		}
	}

	return exitSuccess
}

// valueToTextSimple converts an ast.Value to its simple text representation.
func valueToTextSimple(v *ast.Value) string {
	if v == nil {
		return ""
	}
	var text string
	for _, part := range v.Parts {
		switch p := part.(type) {
		case *ast.LiteralValue:
			text += p.Text
		case *ast.Interpolation:
			if p.Raw {
				text += "{" + p.Name + ":raw}"
			} else {
				text += "{" + p.Name + "}"
			}
		case *ast.FunctionCall:
			text += p.Name.String() + "(...)"
		}
	}
	return text
}

// checkContainerEnvironment checks a Docker/Podman container environment.
func checkContainerEnvironment(env *ast.Environment, buildfileDir string, verbose, showInstall bool) int {
	detector := environ.NewContainerDetector()

	// Check if runtime binary exists
	fmt.Println()
	runtimePath, err := detector.FindRuntime(*env.Runtime)
	if err != nil {
		fmt.Printf("✗ Runtime: %s not found in PATH\n", env.Runtime.String())
		if showInstall {
			pm := DetectPackageManager()
			if pm != nil {
				fmt.Printf("  install: %s\n", pm.GetInstallCommand(env.Runtime.String()))
			}
		}
		return exitEnvError
	}
	if verbose {
		fmt.Printf("✓ Runtime: %s (%s)\n", env.Runtime.String(), runtimePath)
	} else {
		fmt.Printf("✓ Runtime: %s found\n", env.Runtime.String())
	}

	// Check if Dockerfile exists
	result, err := detector.DetectDockerfile(env, buildfileDir)
	if err != nil {
		fmt.Printf("✗ Source: %s\n", err)
		return exitEnvError
	}
	fmt.Printf("✓ Source: %s\n", result.Path)

	// Validate Dockerfile
	if err := detector.ValidateDockerfile(result.Path); err != nil {
		fmt.Printf("✗ Dockerfile validation: %s\n", err)
		return exitEnvError
	}
	fmt.Printf("✓ Dockerfile is valid\n")

	// Show extra args if verbose
	if verbose && env.Args != nil {
		fmt.Printf("  Args: %s\n", valueToTextSimple(env.Args))
	}

	// Generate image tag for info
	projectName := filepath.Base(buildfileDir)
	envName := ""
	if env.Name != nil {
		envName = *env.Name
	}
	imageTag := environ.GenerateImageTag(projectName, envName)
	fmt.Printf("  Image tag: %s\n", imageTag)

	// Check for image (optional, just informational)
	builder := environ.NewImageBuilder(*env.Runtime, runtimePath, nil)
	exists, _ := builder.ImageExists(imageTag)
	if exists {
		fmt.Printf("✓ Image exists locally\n")
	} else {
		fmt.Printf("  Image not yet built (will be built on first run)\n")
	}

	fmt.Println("\nContainer environment ready")
	return exitSuccess
}

// checkDevcontainerEnvironment checks a devcontainer environment.
func checkDevcontainerEnvironment(env *ast.Environment, buildfileDir string, verbose, showInstall bool) int {
	detector := environ.NewDevcontainerDetector()
	runner := environ.NewDevcontainerRunner(buildfileDir)

	// Check if devcontainer CLI is installed
	fmt.Println()
	if err := runner.CheckCLI(); err != nil {
		fmt.Println("✗ devcontainer CLI not found in PATH")
		if showInstall {
			fmt.Println("  Install via: npm install -g @devcontainers/cli")
		}
		// Continue checking configuration even if CLI is not installed
	} else {
		fmt.Println("✓ devcontainer CLI found")
	}

	// Determine where to look for devcontainer config
	var configPath string
	if env.Source != nil {
		// Use the specified source path
		sourcePath := valueToTextSimple(env.Source)
		configPath = filepath.Join(buildfileDir, sourcePath)

		// Check if the specified file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Printf("✗ Source: %s not found\n", configPath)
			return exitEnvError
		}

		// Try to load and parse the config
		cfg, err := detector.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("✗ Configuration: failed to parse %s: %v\n", configPath, err)
			return exitEnvError
		}
		fmt.Printf("✓ Configuration: %s\n", configPath)

		// Show configuration details
		printDevcontainerConfig(cfg, verbose)
	} else {
		// Auto-detect devcontainer configuration
		result, err := detector.DetectConfig(buildfileDir)
		if err != nil {
			fmt.Printf("✗ Configuration: error detecting config: %v\n", err)
			return exitEnvError
		}
		if !result.Found {
			fmt.Println("✗ Configuration: no devcontainer.json found")
			fmt.Println("  Expected: .devcontainer/devcontainer.json or devcontainer.json")
			return exitEnvError
		}

		configPath = result.Path
		cfg, err := detector.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("✗ Configuration: failed to parse %s: %v\n", configPath, err)
			return exitEnvError
		}
		fmt.Printf("✓ Configuration: %s\n", configPath)

		// Show configuration details
		printDevcontainerConfig(cfg, verbose)
	}

	// Set the config path on runner for future use
	runner.SetConfigPath(configPath)

	fmt.Println("\nDevcontainer environment ready")
	return exitSuccess
}

// printDevcontainerConfig prints details about the devcontainer configuration.
func printDevcontainerConfig(cfg *environ.DevcontainerConfig, verbose bool) {
	if cfg.Name != "" {
		fmt.Printf("  Name: %s\n", cfg.Name)
	}

	source := cfg.GetImageOrBuildSource()
	if source != "" {
		fmt.Printf("  Source: %s\n", source)
	}

	if verbose {
		if cfg.WorkspaceFolder != "" {
			fmt.Printf("  Workspace: %s\n", cfg.WorkspaceFolder)
		}
		if cfg.RemoteUser != "" {
			fmt.Printf("  User: %s\n", cfg.RemoteUser)
		}
	}
}

// checkNixEnvironment checks a Nix environment.
func checkNixEnvironment(env *ast.Environment, buildfileDir string, verbose, showInstall bool) int {
	detector := environ.NewNixDetector()
	runner := environ.NewNixRunner(buildfileDir)

	// Check if nix-shell CLI is installed
	fmt.Println()
	if err := runner.CheckCLI(); err != nil {
		fmt.Println("✗ nix-shell CLI not found in PATH")
		if showInstall {
			fmt.Println("  Install Nix: https://nixos.org/download.html")
		}
		// Continue checking configuration even if CLI is not installed
	} else {
		fmt.Println("✓ nix-shell CLI found")
	}

	// Detect Nix configuration
	result, err := detector.DetectConfig(buildfileDir, env.Source)
	if err != nil {
		fmt.Printf("✗ Configuration: error detecting config: %v\n", err)
		return exitEnvError
	}
	if !result.Found {
		fmt.Println("✗ Configuration: no Nix configuration found")
		fmt.Println("  Expected: shell.nix or flake.nix (or specify with .source:)")
		return exitEnvError
	}

	fmt.Printf("✓ Configuration: %s (%s)\n", result.Path, result.Type.String())

	// Set the config on runner for future use
	runner.SetConfig(result.Path, result.Type)

	// Show args if specified
	if env.Args != nil {
		argsStr := valueToTextSimple(env.Args)
		fmt.Printf("  Args: %s\n", argsStr)
		if verbose {
			runner.SetArgs(parseArgs(argsStr))
		}
	}

	fmt.Println("\nNix environment ready")
	return exitSuccess
}

// parseArgs splits a string into arguments, handling simple space separation.
func parseArgs(argsStr string) []string {
	var args []string
	for _, arg := range strings.Fields(argsStr) {
		if arg != "" {
			args = append(args, arg)
		}
	}
	return args
}

// checkLimaEnvironment checks a Lima VM environment (macOS).
func checkLimaEnvironment(env *ast.Environment, buildfileDir string, verbose, showInstall bool) int {
	detector := environ.NewLimaDetector()

	// Determine VM name from environment name or default
	vmName := "default"
	if env.Name != nil {
		vmName = *env.Name
	}
	runner := environ.NewLimaRunner(buildfileDir, vmName)

	// Check if limactl CLI is installed
	fmt.Println()
	if err := runner.CheckCLI(); err != nil {
		fmt.Println("✗ limactl CLI not found in PATH")
		if showInstall {
			fmt.Println("  Install via: brew install lima (macOS)")
		}
		// Continue checking configuration even if CLI is not installed
	} else {
		fmt.Println("✓ limactl CLI found")
	}

	// Detect Lima configuration
	result, err := detector.DetectConfig(buildfileDir, env.Source)
	if err != nil {
		fmt.Printf("✗ Configuration: error detecting config: %v\n", err)
		return exitEnvError
	}
	if !result.Found {
		fmt.Println("✗ Configuration: no Lima configuration found")
		fmt.Println("  Expected: lima.yaml (or specify with .source:)")
		return exitEnvError
	}

	fmt.Printf("✓ Configuration: %s\n", result.Path)
	fmt.Printf("  VM name: %s\n", vmName)

	// Set the config on runner for future use
	runner.SetConfigPath(result.Path)

	// Show args if specified
	if env.Args != nil {
		argsStr := valueToTextSimple(env.Args)
		fmt.Printf("  Args: %s\n", argsStr)
		if verbose {
			runner.SetArgs(parseArgs(argsStr))
		}
	}

	fmt.Println("\nLima environment ready")
	return exitSuccess
}
