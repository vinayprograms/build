package environ

import (
	"os"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

func TestSelectEnvironment_EnvFlagPrecedence(t *testing.T) {
	// Create test environments
	defaultEnv := &ast.Environment{Name: nil} // Default (unnamed) environment
	namedEnv1 := ptr("ci")
	namedEnv2 := ptr("production")
	envCI := &ast.Environment{Name: namedEnv1}
	envProd := &ast.Environment{Name: namedEnv2}

	envs := []*ast.Environment{defaultEnv, envCI, envProd}

	// --env flag should take precedence
	selector := NewEnvironmentSelector()
	result, err := selector.Select(envs, "ci", "")

	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if result != envCI {
		t.Error("expected --env flag to select 'ci' environment")
	}
}

func TestSelectEnvironment_BuildEnvFallback(t *testing.T) {
	// Create test environments
	namedEnv1 := ptr("ci")
	namedEnv2 := ptr("production")
	envCI := &ast.Environment{Name: namedEnv1}
	envProd := &ast.Environment{Name: namedEnv2}

	envs := []*ast.Environment{envCI, envProd}

	// Set BUILD_ENV to select an environment
	os.Setenv("BUILD_ENV", "production")
	defer os.Unsetenv("BUILD_ENV")

	selector := NewEnvironmentSelector()
	result, err := selector.Select(envs, "", "production")

	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if result != envProd {
		t.Error("expected BUILD_ENV to select 'production' environment")
	}
}

func TestSelectEnvironment_DefaultEnv(t *testing.T) {
	// Create test environments
	defaultEnv := &ast.Environment{Name: nil} // Default (unnamed) environment
	namedEnv := ptr("ci")
	envCI := &ast.Environment{Name: namedEnv}

	envs := []*ast.Environment{defaultEnv, envCI}

	// No --env flag and no BUILD_ENV, should use default
	selector := NewEnvironmentSelector()
	result, err := selector.Select(envs, "", "")

	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if result != defaultEnv {
		t.Error("expected default environment to be selected")
	}
}

func TestSelectEnvironment_ErrorWhenOnlyNamedAndNoSelection(t *testing.T) {
	// Create test environments with no default
	namedEnv1 := ptr("ci")
	namedEnv2 := ptr("production")
	envCI := &ast.Environment{Name: namedEnv1}
	envProd := &ast.Environment{Name: namedEnv2}

	envs := []*ast.Environment{envCI, envProd}

	// No --env flag and no BUILD_ENV, should error
	selector := NewEnvironmentSelector()
	_, err := selector.Select(envs, "", "")

	if err == nil {
		t.Error("expected error when only named environments and no selection")
	}
}

func TestSelectEnvironment_EmptyEnvsList(t *testing.T) {
	// Empty environments list - bare environment
	selector := NewEnvironmentSelector()
	result, err := selector.Select(nil, "", "")

	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	// Should return nil, indicating bare environment
	if result != nil {
		t.Error("expected nil for empty environments (bare environment)")
	}
}

func TestSelectEnvironment_EnvFlagNotFound(t *testing.T) {
	// Create test environments
	namedEnv := ptr("ci")
	envCI := &ast.Environment{Name: namedEnv}

	envs := []*ast.Environment{envCI}

	// Request non-existent environment
	selector := NewEnvironmentSelector()
	_, err := selector.Select(envs, "nonexistent", "")

	if err == nil {
		t.Error("expected error when environment not found")
	}
}

func TestSelectEnvironment_EnvFlagOverridesBuildEnv(t *testing.T) {
	// Create test environments
	namedEnv1 := ptr("ci")
	namedEnv2 := ptr("production")
	envCI := &ast.Environment{Name: namedEnv1}
	envProd := &ast.Environment{Name: namedEnv2}

	envs := []*ast.Environment{envCI, envProd}

	// --env flag should override BUILD_ENV
	selector := NewEnvironmentSelector()
	result, err := selector.Select(envs, "ci", "production")

	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if result != envCI {
		t.Error("expected --env flag to override BUILD_ENV")
	}
}

// ptr is a helper to create a pointer to a string
func ptr(s string) *string {
	return &s
}
