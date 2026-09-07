package planner

import (
	"testing"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/eval"
)

// ----------------------------------------------------------------------------
// Dependency Path Resolution Tests
// ----------------------------------------------------------------------------

func TestResolveDependency_LiteralOnly(t *testing.T) {
	ctx := eval.NewContext()

	dep := createDependency([]patternPart{
		{literal: "src/main.c"},
	})

	path, err := ResolveDependency(dep, nil, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "src/main.c" {
		t.Errorf("expected 'src/main.c', got %q", path)
	}
}

func TestResolveDependency_VariableInterpolation(t *testing.T) {
	ctx := eval.NewContext()
	ctx.Set("src_dir", "src")

	dep := createDependency([]patternPart{
		{capture: "src_dir"}, // Will be treated as interpolation since it's defined
		{literal: "/main.c"},
	})

	path, err := ResolveDependency(dep, nil, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "src/main.c" {
		t.Errorf("expected 'src/main.c', got %q", path)
	}
}

func TestResolveDependency_CaptureSubstitution(t *testing.T) {
	ctx := eval.NewContext()

	dep := createDependency([]patternPart{
		{literal: "src/"},
		{capture: "name"},
		{literal: ".c"},
	})

	captures := map[string]string{"name": "utils"}

	path, err := ResolveDependency(dep, captures, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "src/utils.c" {
		t.Errorf("expected 'src/utils.c', got %q", path)
	}
}

func TestResolveDependency_MixedInterpolationAndCapture(t *testing.T) {
	ctx := eval.NewContext()
	ctx.Set("base", "build")

	dep := createDependency([]patternPart{
		{capture: "base"}, // Variable interpolation
		{literal: "/"},
		{capture: "name"}, // Capture from pattern match
		{literal: ".o"},
	})

	captures := map[string]string{"name": "main"}

	path, err := ResolveDependency(dep, captures, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "build/main.o" {
		t.Errorf("expected 'build/main.o', got %q", path)
	}
}

func TestResolveDependency_BuiltinVariable(t *testing.T) {
	ctx := eval.NewContext()

	dep := createDependency([]patternPart{
		{literal: "bin/"},
		{capture: "os"}, // Built-in variable
		{literal: "/app"},
	})

	path, err := ResolveDependency(dep, nil, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// os is a built-in variable, should be substituted
	if path != "bin/"+ctx.Variables()["os"]+"/app" {
		t.Errorf("expected 'bin/<os>/app', got %q", path)
	}
}

func TestResolveDependency_UndefinedVariable(t *testing.T) {
	ctx := eval.NewContext()

	dep := createDependency([]patternPart{
		{literal: "src/"},
		{capture: "undefined_var"}, // Not defined anywhere
		{literal: ".c"},
	})

	_, err := ResolveDependency(dep, nil, ctx)

	if err == nil {
		t.Error("expected error for undefined variable")
	}
}

func TestResolveDependency_CapturePreferredOverVariable(t *testing.T) {
	// If a name is both a capture (from pattern match) and a variable,
	// the capture should take precedence
	ctx := eval.NewContext()
	ctx.Set("name", "from_variable")

	dep := createDependency([]patternPart{
		{literal: "src/"},
		{capture: "name"},
		{literal: ".c"},
	})

	captures := map[string]string{"name": "from_capture"}

	path, err := ResolveDependency(dep, captures, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "src/from_capture.c" {
		t.Errorf("expected 'src/from_capture.c', got %q", path)
	}
}

func TestResolveDependency_MultipleCaptures(t *testing.T) {
	ctx := eval.NewContext()

	dep := createDependency([]patternPart{
		{capture: "dir"},
		{literal: "/"},
		{capture: "name"},
		{literal: ".c"},
	})

	captures := map[string]string{
		"dir":  "src",
		"name": "main",
	}

	path, err := ResolveDependency(dep, captures, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "src/main.c" {
		t.Errorf("expected 'src/main.c', got %q", path)
	}
}

func TestResolveDependency_EmptyCapture(t *testing.T) {
	ctx := eval.NewContext()

	dep := createDependency([]patternPart{
		{literal: "build/"},
		{capture: "prefix"},
		{literal: "main.o"},
	})

	captures := map[string]string{"prefix": ""}

	path, err := ResolveDependency(dep, captures, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "build/main.o" {
		t.Errorf("expected 'build/main.o', got %q", path)
	}
}

// ----------------------------------------------------------------------------
// ResolveDependencies (plural) Tests
// ----------------------------------------------------------------------------

func TestResolveDependencies_MultipleDeps(t *testing.T) {
	ctx := eval.NewContext()

	deps := []ast.Dependency{
		createDependency([]patternPart{{literal: "src/main.c"}}),
		createDependency([]patternPart{{literal: "src/utils.c"}}),
		createDependency([]patternPart{{literal: "include/header.h"}}),
	}

	paths, err := ResolveDependencies(deps, nil, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 3 {
		t.Fatalf("expected 3 paths, got %d", len(paths))
	}
	expected := []string{"src/main.c", "src/utils.c", "include/header.h"}
	for i, exp := range expected {
		if paths[i] != exp {
			t.Errorf("path[%d]: expected %q, got %q", i, exp, paths[i])
		}
	}
}

func TestResolveDependencies_WithCaptures(t *testing.T) {
	ctx := eval.NewContext()

	deps := []ast.Dependency{
		createDependency([]patternPart{
			{literal: "src/"},
			{capture: "name"},
			{literal: ".c"},
		}),
		createDependency([]patternPart{
			{literal: "include/"},
			{capture: "name"},
			{literal: ".h"},
		}),
	}

	captures := map[string]string{"name": "utils"}

	paths, err := ResolveDependencies(deps, captures, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}
	if paths[0] != "src/utils.c" {
		t.Errorf("path[0]: expected 'src/utils.c', got %q", paths[0])
	}
	if paths[1] != "include/utils.h" {
		t.Errorf("path[1]: expected 'include/utils.h', got %q", paths[1])
	}
}

func TestResolveDependencies_Empty(t *testing.T) {
	ctx := eval.NewContext()

	paths, err := ResolveDependencies(nil, nil, ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("expected empty paths, got %d", len(paths))
	}
}

func TestResolveDependencies_ErrorPropagation(t *testing.T) {
	ctx := eval.NewContext()

	deps := []ast.Dependency{
		createDependency([]patternPart{{literal: "src/main.c"}}),
		createDependency([]patternPart{
			{capture: "undefined"}, // Will cause error
		}),
	}

	_, err := ResolveDependencies(deps, nil, ctx)

	if err == nil {
		t.Error("expected error for undefined variable in dependency")
	}
}

// ----------------------------------------------------------------------------
// Test Helpers
// ----------------------------------------------------------------------------

// createDependency creates a Dependency from pattern parts.
func createDependency(parts []patternPart) ast.Dependency {
	segments := make([]ast.PatternSegment, 0, len(parts))
	for _, p := range parts {
		if p.capture != "" {
			segments = append(segments, &ast.BraceExpr{Identifier: p.capture})
		} else {
			segments = append(segments, &ast.LiteralSegment{Text: p.literal})
		}
	}
	return ast.Dependency{Segments: segments}
}
