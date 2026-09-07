package environ

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/eval"
)

// TestResolveSourcePath_Literal covers the common case: a source path with
// no interpolation resolves unchanged.
func TestResolveSourcePath_Literal(t *testing.T) {
	v := &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "docker/ci.Dockerfile"}}}

	got, err := ResolveSourcePath(".source:", v, eval.NewContext())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "docker/ci.Dockerfile" {
		t.Errorf("got %q, want %q", got, "docker/ci.Dockerfile")
	}
}

// TestResolveSourcePath_Interpolation covers C1: a .source: value may
// reference a variable defined in the evaluation context.
func TestResolveSourcePath_Interpolation(t *testing.T) {
	ctx := eval.NewContext()
	ctx.Set("docker_dir", "./docker")

	v := &ast.Value{Parts: []ast.ValuePart{
		&ast.Interpolation{Name: "docker_dir"},
		&ast.LiteralValue{Text: "/ci.Dockerfile"},
	}}

	got, err := ResolveSourcePath(".source:", v, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "./docker/ci.Dockerfile" {
		t.Errorf("got %q, want %q", got, "./docker/ci.Dockerfile")
	}
}

// TestResolveSourcePath_Undefined covers C1's error path: an undefined
// variable in a .source: path is reported naming the directive.
func TestResolveSourcePath_Undefined(t *testing.T) {
	v := &ast.Value{Parts: []ast.ValuePart{&ast.Interpolation{Name: "nope"}}}

	_, err := ResolveSourcePath(".source:", v, eval.NewContext())
	if err == nil {
		t.Fatal("expected an error for undefined variable")
	}
	if !strings.Contains(err.Error(), ".source:") || !strings.Contains(err.Error(), "nope") {
		t.Errorf("error = %q, want to mention '.source:' and 'nope'", err.Error())
	}
}

// TestResolveSourcePath_Nil covers the "directive not given" case.
func TestResolveSourcePath_Nil(t *testing.T) {
	got, err := ResolveSourcePath(".source:", nil, eval.NewContext())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

// TestDockerfileDetection_Interpolation covers C1 end-to-end: DetectDockerfile
// resolves an interpolated .source: path using the evaluation context.
func TestDockerfileDetection_Interpolation(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "docker")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	dockerfilePath := filepath.Join(subDir, "ci.Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte("FROM alpine\n"), 0644); err != nil {
		t.Fatalf("failed to create Dockerfile: %v", err)
	}

	ctx := eval.NewContext()
	ctx.Set("docker_dir", "docker")

	env := &ast.Environment{
		Runtime: ptrRuntime(ast.RuntimeDocker),
		Source: &ast.Value{Parts: []ast.ValuePart{
			&ast.Interpolation{Name: "docker_dir"},
			&ast.LiteralValue{Text: "/ci.Dockerfile"},
		}},
	}

	detector := NewContainerDetector()
	result, err := detector.DetectDockerfile(env, tmpDir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Path != dockerfilePath {
		t.Errorf("Path = %q, want %q", result.Path, dockerfilePath)
	}
	if !result.Exists {
		t.Error("expected Exists=true")
	}
}

// TestDockerfileDetection_UndefinedInterpolation covers C1's error path for
// the container runtime detector: an undefined variable in .source: is a
// clear error, not a silent "no source" fallback.
func TestDockerfileDetection_UndefinedInterpolation(t *testing.T) {
	tmpDir := t.TempDir()

	env := &ast.Environment{
		Runtime: ptrRuntime(ast.RuntimeDocker),
		Source:  &ast.Value{Parts: []ast.ValuePart{&ast.Interpolation{Name: "nope"}}},
	}

	detector := NewContainerDetector()
	_, err := detector.DetectDockerfile(env, tmpDir, eval.NewContext())
	if err == nil {
		t.Fatal("expected an error for undefined variable in .source:")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Errorf("error = %q, want to mention 'nope'", err.Error())
	}
}
