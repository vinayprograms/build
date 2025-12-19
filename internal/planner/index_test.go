package planner

import (
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

func TestNewTargetIndex(t *testing.T) {
	targets := []*ast.Target{
		makeTarget(t, "build/app"),
		makeTarget(t, "build/{name}.o"),
		makeTarget(t, "@clean"),
	}

	idx := NewTargetIndex(targets)

	if idx == nil {
		t.Fatal("NewTargetIndex returned nil")
	}

	// Should have all targets indexed
	if idx.Size() != 3 {
		t.Errorf("expected 3 targets, got %d", idx.Size())
	}
}

func TestTargetIndex_LookupExact(t *testing.T) {
	targets := []*ast.Target{
		makeTarget(t, "build/app"),
		makeTarget(t, "build/lib"),
		makeTarget(t, "src/main.c"),
	}

	idx := NewTargetIndex(targets)

	// Exact match
	target, captures, err := idx.Lookup("build/app")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if target == nil {
		t.Fatal("expected target, got nil")
	}
	if len(captures) != 0 {
		t.Errorf("expected no captures for exact match, got %v", captures)
	}
}

func TestTargetIndex_LookupPattern(t *testing.T) {
	targets := []*ast.Target{
		makeTarget(t, "build/{name}.o"),
	}

	idx := NewTargetIndex(targets)

	// Pattern match
	target, captures, err := idx.Lookup("build/main.o")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if target == nil {
		t.Fatal("expected target, got nil")
	}
	if captures["name"] != "main" {
		t.Errorf("expected name='main', got %q", captures["name"])
	}
}

func TestTargetIndex_LookupPhony(t *testing.T) {
	targets := []*ast.Target{
		makePhonyTarget(t, "clean"),
		makePhonyTarget(t, "test"),
	}

	idx := NewTargetIndex(targets)

	// Lookup with @ prefix
	target, _, err := idx.Lookup("@clean")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if target == nil {
		t.Fatal("expected target, got nil")
	}

	// Lookup without @ prefix
	target, _, err = idx.Lookup("test")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if target == nil {
		t.Fatal("expected target, got nil")
	}
}

func TestTargetIndex_PrefixOptimization(t *testing.T) {
	// Many targets with same prefix should use prefix grouping
	targets := []*ast.Target{
		makeTarget(t, "build/main.o"),
		makeTarget(t, "build/utils.o"),
		makeTarget(t, "build/parser.o"),
		makeTarget(t, "src/main.c"),
		makeTarget(t, "src/utils.c"),
	}

	idx := NewTargetIndex(targets)

	// Verify lookup still works correctly
	target, _, err := idx.Lookup("build/main.o")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if target == nil {
		t.Fatal("expected target")
	}

	target, _, err = idx.Lookup("src/main.c")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if target == nil {
		t.Fatal("expected target")
	}
}

func TestTargetIndex_NotFound(t *testing.T) {
	targets := []*ast.Target{
		makeTarget(t, "build/app"),
	}

	idx := NewTargetIndex(targets)

	_, _, err := idx.Lookup("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent target")
	}
}

func TestTargetIndex_PreferExactOverPattern(t *testing.T) {
	// Both exact match and pattern could match the same path
	// Exact should be preferred
	targets := []*ast.Target{
		makeTarget(t, "build/{name}.o"), // Pattern
		makeTarget(t, "build/main.o"),   // Exact
	}

	idx := NewTargetIndex(targets)

	target, captures, err := idx.Lookup("build/main.o")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if target == nil {
		t.Fatal("expected target")
	}
	// Should get exact match (no captures)
	if len(captures) != 0 {
		t.Errorf("expected exact match with no captures, got %v", captures)
	}
}

// Helper functions

func makeTarget(t *testing.T, pattern string) *ast.Target {
	t.Helper()
	return &ast.Target{
		Pattern: parsePattern(pattern),
	}
}

func makePhonyTarget(t *testing.T, name string) *ast.Target {
	t.Helper()
	return &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{&ast.LiteralSegment{Text: name}},
			IsPhony:  true,
		},
	}
}

func parsePattern(s string) ast.TargetPattern {
	var segments []ast.PatternSegment
	var current string
	i := 0

	for i < len(s) {
		if s[i] == '{' {
			// Start of capture
			if current != "" {
				segments = append(segments, &ast.LiteralSegment{Text: current})
				current = ""
			}
			// Find closing brace
			j := i + 1
			for j < len(s) && s[j] != '}' {
				j++
			}
			if j < len(s) {
				name := s[i+1 : j]
				segments = append(segments, &ast.BraceExpr{Identifier: name})
				i = j + 1
			} else {
				current += string(s[i])
				i++
			}
		} else {
			current += string(s[i])
			i++
		}
	}

	if current != "" {
		segments = append(segments, &ast.LiteralSegment{Text: current})
	}

	return ast.TargetPattern{Segments: segments}
}
