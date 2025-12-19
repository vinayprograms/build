package main

import (
	"fmt"
	"os"

	"github.com/vinayprograms/build/internal/ast"
)

// checkEnvironment verifies the requirements for an environment.
// If envName is empty, checks the default environment.
// Returns exit code.
func checkEnvironment(result BuildfileResult, envName string, verbose bool) int {
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
		// Non-bare environments - just report status for now
		fmt.Printf("Source: %s\n", valueToTextSimple(selectedEnv.Source))
		if selectedEnv.Args != nil {
			fmt.Printf("Args: %s\n", valueToTextSimple(selectedEnv.Args))
		}
		fmt.Println("\nNon-bare environment checking not yet implemented")
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
	}

	if hasErrors {
		fmt.Println("\nSome requirements are not met")
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
