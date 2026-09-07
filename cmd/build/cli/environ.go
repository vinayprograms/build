package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/environ"
	"github.com/vinayprograms/build/internal/eval"
)

// evalContextOrEmpty returns the concrete *eval.Context backing ctx, or a
// fresh empty one if ctx is nil or not the expected concrete adapter type.
// environ constructors need the concrete type to run the evaluator directly;
// an empty context still resolves plain-literal source paths correctly, so
// this degrades gracefully rather than panicking.
func evalContextOrEmpty(ctx EvalContext) *eval.Context {
	if ctx == nil {
		return eval.NewContext()
	}
	if eca, ok := ctx.(*evalContextAdapter); ok {
		return eca.ctx
	}
	return eval.NewContext()
}

// validateEnvironmentRequirements checks that the selected environment's requirements are met.
// Returns nil if no environment is selected, no requirements exist, or all requirements are satisfied.
// Returns an error describing which requirements failed if any are not met.
func validateEnvironmentRequirements(result BuildfileResult, envName string) error {
	envs := GetEnvironments(result)
	if len(envs) == 0 {
		return nil // No environments defined
	}

	// Find the selected environment
	var selectedEnv *ast.Environment
	if envName == "" {
		// Look for default (unnamed) environment
		for _, env := range envs {
			if env.Name == nil {
				selectedEnv = env
				break
			}
		}
	} else {
		// Look for named environment
		for _, env := range envs {
			if env.Name != nil && *env.Name == envName {
				selectedEnv = env
				break
			}
		}
		if selectedEnv == nil {
			return fmt.Errorf("environment '%s' not found", envName)
		}
	}

	if selectedEnv == nil || len(selectedEnv.Requires) == 0 {
		return nil // No requirements to check
	}

	// Check requirements using the same checker as recipe .requires
	checker := environ.NewRequirementsChecker()
	checkResults := checker.CheckRequirementsWithVersion(selectedEnv.Requires)

	var errors []string
	for _, r := range checkResults {
		if r.Error != nil {
			errors = append(errors, r.String())
		}
	}

	if len(errors) > 0 {
		envDisplay := "(default)"
		if envName != "" {
			envDisplay = envName
		}
		return fmt.Errorf("environment '%s' requirements not met:\n  %s", envDisplay, strings.Join(errors, "\n  "))
	}

	return nil
}

// getRuntimeEnvironment creates a RuntimeEnvironment for the selected environment.
// Returns nil if the environment is bare (no special runtime needed).
// Returns an error if the environment doesn't exist or setup fails.
// ctx resolves any interpolation in the environment's .source: path.
func getRuntimeEnvironment(result BuildfileResult, envName, projectDir, projectName string, ctx EvalContext) (environ.RuntimeEnvironment, error) {
	envs := GetEnvironments(result)
	if len(envs) == 0 {
		return nil, nil // No environments defined
	}

	// Find the selected environment
	var selectedEnv *ast.Environment
	if envName == "" {
		// Look for default (unnamed) environment
		for _, env := range envs {
			if env.Name == nil {
				selectedEnv = env
				break
			}
		}
	} else {
		// Look for named environment
		for _, env := range envs {
			if env.Name != nil && *env.Name == envName {
				selectedEnv = env
				break
			}
		}
		if selectedEnv == nil {
			return nil, fmt.Errorf("environment '%s' not found", envName)
		}
	}

	if selectedEnv == nil {
		return nil, nil // No environment selected
	}

	// Check runtime type
	if selectedEnv.Runtime == nil {
		return nil, nil // Bare runtime - no special environment needed
	}

	runtime := *selectedEnv.Runtime
	evalCtx := evalContextOrEmpty(ctx)

	var runtimeEnv environ.RuntimeEnvironment
	var err error

	switch runtime {
	case ast.RuntimeBare:
		return nil, nil // Bare runtime - no special environment

	case ast.RuntimeDocker, ast.RuntimePodman:
		runtimeEnv, err = environ.NewContainerEnvironment(selectedEnv, projectDir, projectName, evalCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to create container environment: %w", err)
		}

	case ast.RuntimeDevcontainer:
		runtimeEnv, err = environ.NewDevcontainerEnvironment(selectedEnv, projectDir, projectName, evalCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to create devcontainer environment: %w", err)
		}

	case ast.RuntimeNix:
		runtimeEnv, err = environ.NewNixEnvironment(selectedEnv, projectDir, evalCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to create nix environment: %w", err)
		}

	case ast.RuntimeLima:
		runtimeEnv, err = environ.NewLimaEnvironment(selectedEnv, projectDir, projectName, evalCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to create lima environment: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported runtime: %s", runtime.String())
	}

	// Validate the environment
	if err := runtimeEnv.Validate(); err != nil {
		runtimeEnv.Close()
		return nil, fmt.Errorf("invalid %s environment: %w", runtime.String(), err)
	}

	return runtimeEnv, nil
}

// checkEnvironment verifies the requirements for an environment.
// If envName is empty, checks the default environment.
// If showInstall is true, shows install suggestions for missing binaries.
// buildfileDir is the directory containing the Buildfile (for resolving relative paths).
// ctx resolves any interpolation in the environment's .source: path.
// Returns exit code.
func checkEnvironment(result BuildfileResult, envName, buildfileDir string, verbose, showInstall bool, ctx EvalContext) int {
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
			return checkContainerEnvironment(selectedEnv, buildfileDir, verbose, showInstall, ctx)
		}

		// Devcontainer environment
		if *selectedEnv.Runtime == ast.RuntimeDevcontainer {
			return checkDevcontainerEnvironment(selectedEnv, buildfileDir, verbose, showInstall, ctx)
		}

		// Nix environment
		if *selectedEnv.Runtime == ast.RuntimeNix {
			return checkNixEnvironment(selectedEnv, buildfileDir, verbose, showInstall, ctx)
		}

		// Lima environment
		if *selectedEnv.Runtime == ast.RuntimeLima {
			return checkLimaEnvironment(selectedEnv, buildfileDir, verbose, showInstall, ctx)
		}

		// Other non-bare environments - just report status for now
		sourceText, err := environ.ResolveSourcePath(".source:", selectedEnv.Source, evalContextOrEmpty(ctx))
		if err != nil {
			fmt.Printf("✗ Source: %v\n", err)
			return exitEnvError
		}
		fmt.Printf("Source: %s\n", sourceText)
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

// listEnvironments lists all defined environments. ctx resolves any
// interpolation in each environment's .source: path (e.g.
// `.source: {docker_dir}/ci.Dockerfile`) the same way a real build or
// --check-env would; if resolution fails for a given environment, its
// source is shown as an error rather than aborting the whole listing. The
// full listing is always printed, but the exit code reflects whether any
// environment failed to resolve (exitParseError) so scripts driving
// --list-env can detect the failure without parsing the output.
func listEnvironments(result BuildfileResult, ctx EvalContext) int {
	envs := GetEnvironments(result)

	if len(envs) == 0 {
		fmt.Println("No environments defined")
		return exitSuccess
	}

	evalCtx := evalContextOrEmpty(ctx)
	hadUnresolved := false

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

		detail := ""
		if env.Source != nil {
			if sourceText, err := environ.ResolveSourcePath(".source:", env.Source, evalCtx); err != nil {
				detail = fmt.Sprintf("source: <error: %v>", err)
				hadUnresolved = true
			} else {
				detail = "source: " + sourceText
			}
		} else {
			reqCount := len(env.Requires)
			switch reqCount {
			case 0:
				// no detail
			case 1:
				detail = "1 requirement"
			default:
				detail = fmt.Sprintf("%d requirements", reqCount)
			}
		}

		if detail == "" {
			fmt.Printf("  %-20s  %-15s\n", name, runtime)
		} else {
			fmt.Printf("  %-20s  %-15s  (%s)\n", name, runtime, detail)
		}
	}

	if hadUnresolved {
		return exitParseError
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
func checkContainerEnvironment(env *ast.Environment, buildfileDir string, verbose, showInstall bool, ctx EvalContext) int {
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
	result, err := detector.DetectDockerfile(env, buildfileDir, evalContextOrEmpty(ctx))
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
func checkDevcontainerEnvironment(env *ast.Environment, buildfileDir string, verbose, showInstall bool, ctx EvalContext) int {
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
		sourcePath, err := environ.ResolveSourcePath(".source:", env.Source, evalContextOrEmpty(ctx))
		if err != nil {
			fmt.Printf("✗ Source: %v\n", err)
			return exitEnvError
		}
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
func checkNixEnvironment(env *ast.Environment, buildfileDir string, verbose, showInstall bool, ctx EvalContext) int {
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
	result, err := detector.DetectConfig(buildfileDir, env.Source, evalContextOrEmpty(ctx))
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
func checkLimaEnvironment(env *ast.Environment, buildfileDir string, verbose, showInstall bool, ctx EvalContext) int {
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
	result, err := detector.DetectConfig(buildfileDir, env.Source, evalContextOrEmpty(ctx))
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
