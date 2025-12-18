package planner

import (
	"os"
	"path/filepath"
	"testing"
)

// ----------------------------------------------------------------------------
// Autodeps Parser Tests
// ----------------------------------------------------------------------------

func TestParseAutodeps_Simple(t *testing.T) {
	// Simple .d file format: target: dep1 dep2
	content := "build/main.o: src/main.c src/main.h\n"

	deps, err := ParseAutodeps(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"src/main.c", "src/main.h"}
	if !stringSliceEqual(deps, expected) {
		t.Errorf("expected %v, got %v", expected, deps)
	}
}

func TestParseAutodeps_Multiline(t *testing.T) {
	// Multiline with backslash continuation
	content := `build/main.o: src/main.c \
  src/main.h \
  include/config.h
`
	deps, err := ParseAutodeps(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"src/main.c", "src/main.h", "include/config.h"}
	if !stringSliceEqual(deps, expected) {
		t.Errorf("expected %v, got %v", expected, deps)
	}
}

func TestParseAutodeps_MultipleTargets(t *testing.T) {
	// Multiple targets in one file (gcc -MD generates this)
	content := `build/main.o: src/main.c src/main.h

build/utils.o: src/utils.c src/utils.h
`
	deps, err := ParseAutodeps(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return all unique dependencies
	if len(deps) != 4 {
		t.Errorf("expected 4 deps, got %d: %v", len(deps), deps)
	}
}

func TestParseAutodeps_Empty(t *testing.T) {
	deps, err := ParseAutodeps("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deps) != 0 {
		t.Errorf("expected empty deps, got %v", deps)
	}
}

func TestParseAutodeps_NoColonLine(t *testing.T) {
	// Lines without colon are ignored
	content := "# comment\nbuild/main.o: src/main.c\nignored line\n"

	deps, err := ParseAutodeps(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"src/main.c"}
	if !stringSliceEqual(deps, expected) {
		t.Errorf("expected %v, got %v", expected, deps)
	}
}

func TestParseAutodeps_SpacesInPath(t *testing.T) {
	// GCC escapes spaces in paths
	content := `build/main.o: src/main.c src/path\ with\ spaces/header.h
`
	deps, err := ParseAutodeps(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should handle escaped spaces
	if len(deps) < 2 {
		t.Errorf("expected at least 2 deps, got %v", deps)
	}
}

func TestParseAutodepsFile(t *testing.T) {
	// Create temp .d file
	tmpDir := t.TempDir()
	dPath := filepath.Join(tmpDir, "test.d")

	content := "build/test.o: src/test.c src/test.h\n"
	if err := os.WriteFile(dPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	deps, err := ParseAutodepsFile(dPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"src/test.c", "src/test.h"}
	if !stringSliceEqual(deps, expected) {
		t.Errorf("expected %v, got %v", expected, deps)
	}
}

func TestParseAutodepsFile_NotExists(t *testing.T) {
	// Non-existent file should return empty (not error)
	deps, err := ParseAutodepsFile("/nonexistent/path.d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deps) != 0 {
		t.Errorf("expected empty deps for non-existent file, got %v", deps)
	}
}

// Helper function
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
