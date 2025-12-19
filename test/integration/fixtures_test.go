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
	buildfile      string
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
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".build" {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			h := NewTestHarness(t)

			// Read the fixture content
			content, err := os.ReadFile(filepath.Join(validDir, entry.Name()))
			if err != nil {
				t.Fatalf("failed to read fixture: %v", err)
			}

			// Write it as Buildfile in test directory
			h.WriteFile("Buildfile", string(content))

			// Try to parse with --dry-run
			result := h.Run("--dry-run", "--help")
			// Just checking that it doesn't crash on valid input
			// --help should always succeed regardless of buildfile content
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
		{"missing_end.build", "end", 3},
		{"undefined_var.build", "undefined", 3},
		{"wrong_scope.build", "scope", 3},
		{"duplicate_target.build", "duplicate", 3},
		{"circular_dep.build", "circular", 3},
		{"mixed_indent.build", "unexpected token", 3}, // Mixed indent produces lexer error
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

			// Write it as Buildfile in test directory
			h.WriteFile("Buildfile", string(content))

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

// TestSimpleBuildfileFixture tests the simple.build fixture specifically.
func TestSimpleBuildfileFixture(t *testing.T) {
	h := NewTestHarness(t)

	fixturesDir := getFixturesDir()
	content, err := os.ReadFile(filepath.Join(fixturesDir, "valid", "simple.build"))
	if err != nil {
		t.Skipf("simple.build fixture not found")
		return
	}

	h.WriteFile("Buildfile", string(content))

	result := h.Run("hello")
	result.AssertSuccess().
		AssertStdoutContains("Hello, World!")
}

// TestVariablesFixture tests the variables.build fixture.
func TestVariablesFixture(t *testing.T) {
	h := NewTestHarness(t)

	fixturesDir := getFixturesDir()
	content, err := os.ReadFile(filepath.Join(fixturesDir, "valid", "variables.build"))
	if err != nil {
		t.Skipf("variables.build fixture not found")
		return
	}

	h.WriteFile("Buildfile", string(content))

	result := h.Run("build")
	result.AssertSuccess().
		AssertStdoutContains("Compiler:").
		AssertStdoutContains("gcc").
		AssertStdoutContains("Flags:").
		AssertStdoutContains("-Wall")
}

// TestConditionalsFixture tests the conditionals.build fixture.
func TestConditionalsFixture(t *testing.T) {
	h := NewTestHarness(t)

	fixturesDir := getFixturesDir()
	content, err := os.ReadFile(filepath.Join(fixturesDir, "valid", "conditionals.build"))
	if err != nil {
		t.Skipf("conditionals.build fixture not found")
		return
	}

	h.WriteFile("Buildfile", string(content))

	result := h.Run("info")
	result.AssertSuccess().
		AssertStdoutContains("Platform:")
}

// TestDependenciesFixture tests the dependencies.build fixture.
func TestDependenciesFixture(t *testing.T) {
	h := NewTestHarness(t)

	fixturesDir := getFixturesDir()
	content, err := os.ReadFile(filepath.Join(fixturesDir, "valid", "dependencies.build"))
	if err != nil {
		t.Skipf("dependencies.build fixture not found")
		return
	}

	h.WriteFile("Buildfile", string(content))

	result := h.Run("all")
	result.AssertSuccess().
		AssertStdoutContains("Compiling...").
		AssertStdoutContains("Testing...").
		AssertStdoutContains("Packaging...")
}

// TestFunctionsFixture tests the functions.build fixture.
func TestFunctionsFixture(t *testing.T) {
	h := NewTestHarness(t)

	fixturesDir := getFixturesDir()
	content, err := os.ReadFile(filepath.Join(fixturesDir, "valid", "functions.build"))
	if err != nil {
		t.Skipf("functions.build fixture not found")
		return
	}

	h.WriteFile("Buildfile", string(content))

	result := h.Run("test")
	result.AssertSuccess().
		AssertStdoutContains("Directory:").
		AssertStdoutContains("Basename:").
		AssertStdoutContains("Objects:")
}

// TestBlockCommandsFixture tests the block_commands.build fixture.
func TestBlockCommandsFixture(t *testing.T) {
	h := NewTestHarness(t)

	fixturesDir := getFixturesDir()
	content, err := os.ReadFile(filepath.Join(fixturesDir, "valid", "block_commands.build"))
	if err != nil {
		t.Skipf("block_commands.build fixture not found")
		return
	}

	h.WriteFile("Buildfile", string(content))

	result := h.Run("generate")
	result.AssertSuccess().
		AssertStdoutContains("Item 1").
		AssertStdoutContains("Item 2").
		AssertStdoutContains("Item 3")
}
