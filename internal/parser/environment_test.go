package parser

import (
	"testing"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/lexer"
)

func TestParseEnvironmentBlock_DefaultEnvironment(t *testing.T) {
	input := `.environment:
    .using: bare
    .requires: gcc@11 python3@latest
`
	l := lexer.New("test", input)
	p := New(l)

	env, err := p.ParseEnvironment()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env == nil {
		t.Fatal("expected environment, got nil")
	}

	// Default environment has no name
	if env.Name != nil {
		t.Errorf("expected nil name for default environment, got %q", *env.Name)
	}

	// Check runtime
	if env.Runtime == nil {
		t.Fatal("expected runtime, got nil")
	}
	if *env.Runtime != ast.RuntimeBare {
		t.Errorf("runtime = %v, want %v", *env.Runtime, ast.RuntimeBare)
	}

	// Check requirements
	if len(env.Requires) != 2 {
		t.Fatalf("len(requires) = %d, want 2", len(env.Requires))
	}
	if env.Requires[0].Name != "gcc" {
		t.Errorf("requires[0].Name = %q, want %q", env.Requires[0].Name, "gcc")
	}
	if _, ok := env.Requires[0].Version.(ast.VersionMajor); !ok {
		t.Errorf("requires[0].Version type = %T, want VersionMajor", env.Requires[0].Version)
	}
	if env.Requires[1].Name != "python3" {
		t.Errorf("requires[1].Name = %q, want %q", env.Requires[1].Name, "python3")
	}
	if _, ok := env.Requires[1].Version.(ast.VersionLatest); !ok {
		t.Errorf("requires[1].Version type = %T, want VersionLatest", env.Requires[1].Version)
	}
}

func TestParseEnvironmentBlock_NamedEnvironment(t *testing.T) {
	input := `.environment: ci
    .using: docker
    .source: Dockerfile.ci
    .args: --platform linux/amd64
`
	l := lexer.New("test", input)
	p := New(l)

	env, err := p.ParseEnvironment()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env == nil {
		t.Fatal("expected environment, got nil")
	}

	// Named environment
	if env.Name == nil {
		t.Fatal("expected name, got nil")
	}
	if *env.Name != "ci" {
		t.Errorf("name = %q, want %q", *env.Name, "ci")
	}

	// Check runtime
	if env.Runtime == nil {
		t.Fatal("expected runtime, got nil")
	}
	if *env.Runtime != ast.RuntimeDocker {
		t.Errorf("runtime = %v, want %v", *env.Runtime, ast.RuntimeDocker)
	}

	// Check source
	if env.Source == nil {
		t.Fatal("expected source, got nil")
	}
	if len(env.Source.Parts) != 1 {
		t.Fatalf("len(source.Parts) = %d, want 1", len(env.Source.Parts))
	}
	if lit, ok := env.Source.Parts[0].(*ast.LiteralValue); !ok {
		t.Errorf("source.Parts[0] type = %T, want *LiteralValue", env.Source.Parts[0])
	} else if lit.Text != "Dockerfile.ci" {
		t.Errorf("source text = %q, want %q", lit.Text, "Dockerfile.ci")
	}

	// Check args
	if env.Args == nil {
		t.Fatal("expected args, got nil")
	}
	if len(env.Args.Parts) != 1 {
		t.Fatalf("len(args.Parts) = %d, want 1", len(env.Args.Parts))
	}
	if lit, ok := env.Args.Parts[0].(*ast.LiteralValue); !ok {
		t.Errorf("args.Parts[0] type = %T, want *LiteralValue", env.Args.Parts[0])
	} else if lit.Text != "--platform linux/amd64" {
		t.Errorf("args text = %q, want %q", lit.Text, "--platform linux/amd64")
	}
}

func TestParseEnvironmentBlock_AllRuntimes(t *testing.T) {
	tests := []struct {
		name     string
		using    string
		expected ast.Runtime
	}{
		{"bare", "bare", ast.RuntimeBare},
		{"docker", "docker", ast.RuntimeDocker},
		{"podman", "podman", ast.RuntimePodman},
		{"devcontainer", "devcontainer", ast.RuntimeDevcontainer},
		{"nix", "nix", ast.RuntimeNix},
		{"lima", "lima", ast.RuntimeLima},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := `.environment:
    .using: ` + tt.using + "\n"
			l := lexer.New("test", input)
			p := New(l)

			env, err := p.ParseEnvironment()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if env.Runtime == nil {
				t.Fatal("expected runtime, got nil")
			}
			if *env.Runtime != tt.expected {
				t.Errorf("runtime = %v, want %v", *env.Runtime, tt.expected)
			}
		})
	}
}

func TestParseEnvironmentBlock_WithInterpolation(t *testing.T) {
	input := `.environment:
    .using: docker
    .source: {docker_dir}/Dockerfile
`
	l := lexer.New("test", input)
	p := New(l)

	env, err := p.ParseEnvironment()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env.Source == nil {
		t.Fatal("expected source, got nil")
	}

	// Source should have interpolation
	if len(env.Source.Parts) < 2 {
		t.Fatalf("len(source.Parts) = %d, want at least 2", len(env.Source.Parts))
	}

	// First part should be interpolation
	if interp, ok := env.Source.Parts[0].(*ast.Interpolation); !ok {
		t.Errorf("source.Parts[0] type = %T, want *Interpolation", env.Source.Parts[0])
	} else if interp.Name != "docker_dir" {
		t.Errorf("interpolation name = %q, want %q", interp.Name, "docker_dir")
	}
}

func TestParseEnvironmentBlock_ScopeValidation(t *testing.T) {
	// Environment should be at global scope
	input := `.environment:
    .using: bare
`
	l := lexer.New("test", input)
	p := New(l)

	// Initially at global scope
	if p.CurrentScope() != ScopeGlobal {
		t.Fatalf("initial scope = %v, want %v", p.CurrentScope(), ScopeGlobal)
	}

	env, err := p.ParseEnvironment()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env == nil {
		t.Fatal("expected environment, got nil")
	}

	// After parsing, should be back at global scope
	if p.CurrentScope() != ScopeGlobal {
		t.Errorf("final scope = %v, want %v", p.CurrentScope(), ScopeGlobal)
	}
}

func TestParseEnvironmentBlock_NoDirectives(t *testing.T) {
	input := `.environment:
`
	l := lexer.New("test", input)
	p := New(l)

	env, err := p.ParseEnvironment()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env == nil {
		t.Fatal("expected environment, got nil")
	}

	// All directives should be nil/empty
	if env.Name != nil {
		t.Errorf("name should be nil")
	}
	if env.Runtime != nil {
		t.Errorf("runtime should be nil")
	}
	if env.Source != nil {
		t.Errorf("source should be nil")
	}
	if env.Args != nil {
		t.Errorf("args should be nil")
	}
	if len(env.Requires) != 0 {
		t.Errorf("requires should be empty")
	}
}

func TestParseEnvironmentBlock_InvalidDirective(t *testing.T) {
	// .shell is not valid in environment scope
	input := `.environment:
    .shell: bash
`
	l := lexer.New("test", input)
	p := New(l)

	_, err := p.ParseEnvironment()
	if err == nil {
		t.Fatal("expected error for invalid directive in environment scope")
	}

	// Check error message mentions scope
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestParseEnvironmentBlock_MultipleRequires(t *testing.T) {
	input := `.environment:
    .requires: gcc@11
    .requires: python3@3.10 pip@latest
`
	l := lexer.New("test", input)
	p := New(l)

	env, err := p.ParseEnvironment()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All requires should be collected
	if len(env.Requires) != 3 {
		t.Fatalf("len(requires) = %d, want 3", len(env.Requires))
	}

	expected := []struct {
		name    string
		version string
	}{
		{"gcc", "11"},
		{"python3", "3.10"},
		{"pip", "latest"},
	}

	for i, exp := range expected {
		if env.Requires[i].Name != exp.name {
			t.Errorf("requires[%d].Name = %q, want %q", i, env.Requires[i].Name, exp.name)
		}
	}
}

func TestParseEnvironmentBlock_NixEnvironment(t *testing.T) {
	input := `.environment: nix-dev
    .using: nix
    .source: shell.nix
    .args: --pure
`
	l := lexer.New("test", input)
	p := New(l)

	env, err := p.ParseEnvironment()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env.Name == nil || *env.Name != "nix-dev" {
		t.Errorf("name = %v, want %q", env.Name, "nix-dev")
	}

	if env.Runtime == nil || *env.Runtime != ast.RuntimeNix {
		t.Errorf("runtime = %v, want %v", env.Runtime, ast.RuntimeNix)
	}
}

func TestParseEnvironmentBlock_LimaEnvironment(t *testing.T) {
	input := `.environment: linux-vm
    .using: lima
    .source: linux.yaml
`
	l := lexer.New("test", input)
	p := New(l)

	env, err := p.ParseEnvironment()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env.Runtime == nil || *env.Runtime != ast.RuntimeLima {
		t.Errorf("runtime = %v, want %v", env.Runtime, ast.RuntimeLima)
	}
}

func TestParseEnvironmentBlock_DevcontainerEnvironment(t *testing.T) {
	input := `.environment:
    .using: devcontainer
    .source: .devcontainer/devcontainer.json
`
	l := lexer.New("test", input)
	p := New(l)

	env, err := p.ParseEnvironment()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env.Runtime == nil || *env.Runtime != ast.RuntimeDevcontainer {
		t.Errorf("runtime = %v, want %v", env.Runtime, ast.RuntimeDevcontainer)
	}

	if env.Source == nil {
		t.Fatal("expected source")
	}
}

func TestParseEnvironmentBlock_SourceLocation(t *testing.T) {
	input := `.environment: test
    .using: bare
`
	l := lexer.New("test.build", input)
	p := New(l)

	env, err := p.ParseEnvironment()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env.Location.File != "test.build" {
		t.Errorf("location.File = %q, want %q", env.Location.File, "test.build")
	}
	if env.Location.Line != 1 {
		t.Errorf("location.Line = %d, want %d", env.Location.Line, 1)
	}
}

func TestParseEnvironmentBlock_EmptyName(t *testing.T) {
	// Just .environment: with no name should be default environment
	input := `.environment:
    .using: bare
`
	l := lexer.New("test", input)
	p := New(l)

	env, err := p.ParseEnvironment()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env.Name != nil {
		t.Errorf("name should be nil for default environment, got %q", *env.Name)
	}
}

func TestParseEnvironmentBlock_WithComments(t *testing.T) {
	input := `.environment: ci
    # This is the CI environment
    .using: docker
    # Use the CI Dockerfile
    .source: Dockerfile.ci
`
	l := lexer.New("test", input)
	p := New(l)

	env, err := p.ParseEnvironment()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env.Name == nil || *env.Name != "ci" {
		t.Errorf("name = %v, want %q", env.Name, "ci")
	}

	if env.Runtime == nil || *env.Runtime != ast.RuntimeDocker {
		t.Errorf("runtime = %v, want %v", env.Runtime, ast.RuntimeDocker)
	}

	if env.Source == nil {
		t.Fatal("expected source")
	}
}

func TestParseEnvironmentBlock_VersionSpecs(t *testing.T) {
	input := `.environment:
    .requires: gcc@11 python@3.10 node@16.14.0 make@latest cmake
`
	l := lexer.New("test", input)
	p := New(l)

	env, err := p.ParseEnvironment()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(env.Requires) != 5 {
		t.Fatalf("len(requires) = %d, want 5", len(env.Requires))
	}

	// Check version types
	tests := []struct {
		name    string
		verType string
	}{
		{"gcc", "VersionMajor"},
		{"python", "VersionMajorMinor"},
		{"node", "VersionExact"},
		{"make", "VersionLatest"},
		{"cmake", "VersionLatest"},
	}

	for i, tt := range tests {
		switch tt.verType {
		case "VersionMajor":
			if _, ok := env.Requires[i].Version.(ast.VersionMajor); !ok {
				t.Errorf("requires[%d] (%s) version type = %T, want VersionMajor", i, tt.name, env.Requires[i].Version)
			}
		case "VersionMajorMinor":
			if _, ok := env.Requires[i].Version.(ast.VersionMajorMinor); !ok {
				t.Errorf("requires[%d] (%s) version type = %T, want VersionMajorMinor", i, tt.name, env.Requires[i].Version)
			}
		case "VersionExact":
			if _, ok := env.Requires[i].Version.(ast.VersionExact); !ok {
				t.Errorf("requires[%d] (%s) version type = %T, want VersionExact", i, tt.name, env.Requires[i].Version)
			}
		case "VersionLatest":
			if _, ok := env.Requires[i].Version.(ast.VersionLatest); !ok {
				t.Errorf("requires[%d] (%s) version type = %T, want VersionLatest", i, tt.name, env.Requires[i].Version)
			}
		}
	}
}
