package environ

import (
	"github.com/vinayprograms/need/internal/ast"
)

// EnvironmentSelector handles selection of the appropriate build environment.
type EnvironmentSelector struct{}

// NewEnvironmentSelector creates a new EnvironmentSelector.
func NewEnvironmentSelector() *EnvironmentSelector {
	return &EnvironmentSelector{}
}

// Select selects an environment based on the following priority:
// 1. --env flag (envFlagValue)
// 2. NEED_ENV environment variable (buildEnvValue)
// 3. Unnamed (default) environment
//
// Returns nil if there are no environments defined (bare environment).
// Returns an error if only named environments exist and no selection is made.
func (s *EnvironmentSelector) Select(envs []*ast.Environment, envFlagValue, buildEnvValue string) (*ast.Environment, error) {
	// No environments defined - bare environment
	if len(envs) == 0 {
		return nil, nil
	}

	// Build index of environments by name
	byName := make(map[string]*ast.Environment)
	var defaultEnv *ast.Environment
	var availableNames []string

	for _, env := range envs {
		if env.Name == nil {
			defaultEnv = env
		} else {
			byName[*env.Name] = env
			availableNames = append(availableNames, *env.Name)
		}
	}

	// Priority 1: --env flag
	if envFlagValue != "" {
		if env, ok := byName[envFlagValue]; ok {
			return env, nil
		}
		return nil, &EnvironmentNotFoundError{Name: envFlagValue}
	}

	// Priority 2: NEED_ENV environment variable
	if buildEnvValue != "" {
		if env, ok := byName[buildEnvValue]; ok {
			return env, nil
		}
		return nil, &EnvironmentNotFoundError{Name: buildEnvValue}
	}

	// Priority 3: Default (unnamed) environment
	if defaultEnv != nil {
		return defaultEnv, nil
	}

	// No default and no selection - error
	return nil, &NoDefaultEnvironmentError{Available: availableNames}
}
