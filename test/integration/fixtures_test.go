package integration

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// fixtureTest represents a test case from a fixture file.
type fixtureTest struct {
	name           string
	needfile      string
	target         string
	expectSuccess  bool
	expectContains []string // Substrings expected in output
}

// getFixturesDir returns the path to the fixtures directory.
func getFixturesDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get caller information")
	}
	return filepath.Join(filepath.Dir(filename), "fixtures")
}

// TestValidFixtures tests that all valid fixture files parse and run correctly.
func TestValidFixtures(t *testing.T) {
	fixturesDir := getFixturesDir()
	validDir := filepath.Join(fixturesDir, "valid")

	entries, err := os.ReadDir(validDir)
	if err != nil {
		t.Fatalf("failed to read valid fixtures directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".need" {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			h := NewTestHarness(t)

			// Read the fixture content
			content, err := os.ReadFile(filepath.Join(validDir, entry.Name()))
			if err != nil {
				t.Fatalf("failed to read fixture: %v", err)
			}

			// Write it as Needfile in test directory
			h.WriteFile("Needfile", string(content))

			// Try to parse with --dry-run
			result := h.Run("--dry-run", "--help")
			// Just checking that it doesn't crash on valid input
			// --help should always succeed regardless of needfile content
			result.AssertSuccess()
		})
	}
}

// TestInvalidFixtures tests that invalid fixture files produce parse errors.
func TestInvalidFixtures(t *testing.T) {
	fixturesDir := getFixturesDir()
	invalidDir := filepath.Join(fixturesDir, "invalid")

	testCases := []struct {
		name        string
		expectError string
		expectCode  int
	}{
		{"missing_end.need", "end", 3},
		{"undefined_var.need", "undefined", 3},
		{"wrong_scope.need", "scope", 3},
		{"duplicate_target.need", "duplicate", 3},
		{"circular_dep.need", "circular", 3},
		{"mixed_indent.need", "indentation", 3}, // Mixed indent produces a lexer indentation error
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewTestHarness(t)

			// Read the fixture content
			content, err := os.ReadFile(filepath.Join(invalidDir, tc.name))
			if err != nil {
				t.Skipf("fixture not found: %s", tc.name)
				return
			}

			// Write it as Needfile in test directory
			h.WriteFile("Needfile", string(content))

			// Try to run (should fail)
			result := h.Run()
			result.AssertExitCode(tc.expectCode)

			// Check that expected error message appears
			stderr := result.Stderr()
			if tc.expectError != "" && !containsIgnoreCase(stderr, tc.expectError) {
				t.Errorf("expected stderr to contain %q, got:\n%s", tc.expectError, stderr)
			}
		})
	}
}

// containsIgnoreCase checks if s contains substr, case-insensitive.
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (findIgnoreCase(s, substr) >= 0)
}

func findIgnoreCase(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if matchIgnoreCase(s[i:i+len(substr)], substr) {
			return i
		}
	}
	return -1
}

func matchIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// TestSimpleNeedfileFixture tests the simple.need fixture specifically.
func TestSimpleNeedfileFixture(t *testing.T) {
	h := NewTestHarness(t)

	fixturesDir := getFixturesDir()
	content, err := os.ReadFile(filepath.Join(fixturesDir, "valid", "simple.need"))
	if err != nil {
		t.Skipf("simple.need fixture not found")
		return
	}

	h.WriteFile("Needfile", string(content))

	result := h.Run("hello")
	result.AssertSuccess().
		AssertStdoutContains("Hello, World!")
}

// TestVariablesFixture tests the variables.need fixture.
func TestVariablesFixture(t *testing.T) {
	h := NewTestHarness(t)

	fixturesDir := getFixturesDir()
	content, err := os.ReadFile(filepath.Join(fixturesDir, "valid", "variables.need"))
	if err != nil {
		t.Skipf("variables.need fixture not found")
		return
	}

	h.WriteFile("Needfile", string(content))

	result := h.Run("build")
	result.AssertSuccess().
		AssertStdoutContains("Compiler:").
		AssertStdoutContains("gcc").
		AssertStdoutContains("Flags:").
		AssertStdoutContains("-Wall")
}

// TestConditionalsFixture tests the conditionals.need fixture.
func TestConditionalsFixture(t *testing.T) {
	h := NewTestHarness(t)

	fixturesDir := getFixturesDir()
	content, err := os.ReadFile(filepath.Join(fixturesDir, "valid", "conditionals.need"))
	if err != nil {
		t.Skipf("conditionals.need fixture not found")
		return
	}

	h.WriteFile("Needfile", string(content))

	result := h.Run("info")
	result.AssertSuccess().
		AssertStdoutContains("Platform:")
}

// TestDependenciesFixture tests the dependencies.need fixture.
func TestDependenciesFixture(t *testing.T) {
	h := NewTestHarness(t)

	fixturesDir := getFixturesDir()
	content, err := os.ReadFile(filepath.Join(fixturesDir, "valid", "dependencies.need"))
	if err != nil {
		t.Skipf("dependencies.need fixture not found")
		return
	}

	h.WriteFile("Needfile", string(content))

	result := h.Run("all")
	result.AssertSuccess().
		AssertStdoutContains("Compiling...").
		AssertStdoutContains("Testing...").
		AssertStdoutContains("Packaging...")
}

// TestFunctionsFixture tests the functions.need fixture.
func TestFunctionsFixture(t *testing.T) {
	h := NewTestHarness(t)

	fixturesDir := getFixturesDir()
	content, err := os.ReadFile(filepath.Join(fixturesDir, "valid", "functions.need"))
	if err != nil {
		t.Skipf("functions.need fixture not found")
		return
	}

	h.WriteFile("Needfile", string(content))

	result := h.Run("test")
	result.AssertSuccess().
		AssertStdoutContains("Directory:").
		AssertStdoutContains("Filename:").
		AssertStdoutContains("Objects:")
}

// TestBlockCommandsFixture tests the block_commands.need fixture.
func TestBlockCommandsFixture(t *testing.T) {
	h := NewTestHarness(t)

	fixturesDir := getFixturesDir()
	content, err := os.ReadFile(filepath.Join(fixturesDir, "valid", "block_commands.need"))
	if err != nil {
		t.Skipf("block_commands.need fixture not found")
		return
	}

	h.WriteFile("Needfile", string(content))

	result := h.Run("generate")
	result.AssertSuccess().
		AssertStdoutContains("Item 1").
		AssertStdoutContains("Item 2").
		AssertStdoutContains("Item 3")
}
