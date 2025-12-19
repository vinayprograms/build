// Package integration provides end-to-end integration tests for the build tool.
package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestHarness provides utilities for running integration tests.
type TestHarness struct {
	t          *testing.T
	workDir    string // Temporary working directory
	buildPath  string // Path to build binary
	cleanupFns []func()
}

// NewTestHarness creates a new test harness for integration tests.
func NewTestHarness(t *testing.T) *TestHarness {
	t.Helper()

	// Build the build tool binary
	buildPath := buildBinary(t)

	// Create temporary working directory
	workDir := t.TempDir()

	return &TestHarness{
		t:         t,
		workDir:   workDir,
		buildPath: buildPath,
	}
}

// WriteFile writes a file in the working directory.
func (h *TestHarness) WriteFile(name, content string) {
	h.t.Helper()
	path := filepath.Join(h.workDir, name)

	// Create parent directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		h.t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		h.t.Fatalf("failed to write file %s: %v", path, err)
	}
}

// Mkdir creates a directory in the working directory.
func (h *TestHarness) Mkdir(name string) {
	h.t.Helper()
	path := filepath.Join(h.workDir, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		h.t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

// Run executes the build tool with the given arguments.
func (h *TestHarness) Run(args ...string) *RunResult {
	h.t.Helper()

	cmd := exec.Command(h.buildPath, args...)
	cmd.Dir = h.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			h.t.Fatalf("failed to run build: %v", err)
		}
	}

	return &RunResult{
		t:        h.t,
		exitCode: exitCode,
		stdout:   stdout.String(),
		stderr:   stderr.String(),
	}
}

// RunShell executes a shell command directly (not through the build tool).
func (h *TestHarness) RunShell(command string) *RunResult {
	h.t.Helper()

	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = h.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			h.t.Fatalf("failed to run shell command: %v", err)
		}
	}

	return &RunResult{
		t:        h.t,
		exitCode: exitCode,
		stdout:   stdout.String(),
		stderr:   stderr.String(),
	}
}

// FileExists checks if a file exists in the working directory.
func (h *TestHarness) FileExists(name string) bool {
	path := filepath.Join(h.workDir, name)
	_, err := os.Stat(path)
	return err == nil
}

// ReadFile reads a file from the working directory.
func (h *TestHarness) ReadFile(name string) string {
	h.t.Helper()
	path := filepath.Join(h.workDir, name)
	content, err := os.ReadFile(path)
	if err != nil {
		h.t.Fatalf("failed to read file %s: %v", path, err)
	}
	return string(content)
}

// FileMode returns the file mode of a file in the working directory.
func (h *TestHarness) FileMode(name string) os.FileMode {
	h.t.Helper()
	path := filepath.Join(h.workDir, name)
	info, err := os.Stat(path)
	if err != nil {
		h.t.Fatalf("failed to stat file %s: %v", path, err)
	}
	return info.Mode()
}

// RunResult represents the result of running the build tool.
type RunResult struct {
	t        *testing.T
	exitCode int
	stdout   string
	stderr   string
}

// AssertSuccess asserts that the build succeeded (exit code 0).
func (r *RunResult) AssertSuccess() *RunResult {
	r.t.Helper()
	if r.exitCode != 0 {
		r.t.Errorf("expected success (exit code 0), got exit code %d\nstdout:\n%s\nstderr:\n%s",
			r.exitCode, r.stdout, r.stderr)
	}
	return r
}

// AssertExitCode asserts that the build exited with the given code.
func (r *RunResult) AssertExitCode(code int) *RunResult {
	r.t.Helper()
	if r.exitCode != code {
		r.t.Errorf("expected exit code %d, got %d\nstdout:\n%s\nstderr:\n%s",
			code, r.exitCode, r.stdout, r.stderr)
	}
	return r
}

// AssertStdoutContains asserts that stdout contains the given substring.
func (r *RunResult) AssertStdoutContains(s string) *RunResult {
	r.t.Helper()
	if !strings.Contains(r.stdout, s) {
		r.t.Errorf("expected stdout to contain %q\nstdout:\n%s", s, r.stdout)
	}
	return r
}

// AssertStdoutNotContains asserts that stdout does not contain the given substring.
func (r *RunResult) AssertStdoutNotContains(s string) *RunResult {
	r.t.Helper()
	if strings.Contains(r.stdout, s) {
		r.t.Errorf("expected stdout not to contain %q\nstdout:\n%s", s, r.stdout)
	}
	return r
}

// AssertStderrContains asserts that stderr contains the given substring.
func (r *RunResult) AssertStderrContains(s string) *RunResult {
	r.t.Helper()
	if !strings.Contains(r.stderr, s) {
		r.t.Errorf("expected stderr to contain %q\nstderr:\n%s", s, r.stderr)
	}
	return r
}

// AssertStderrNotContains asserts that stderr does not contain the given substring.
func (r *RunResult) AssertStderrNotContains(s string) *RunResult {
	r.t.Helper()
	if strings.Contains(r.stderr, s) {
		r.t.Errorf("expected stderr not to contain %q\nstderr:\n%s", s, r.stderr)
	}
	return r
}

// Stdout returns the stdout output.
func (r *RunResult) Stdout() string {
	return r.stdout
}

// Stderr returns the stderr output.
func (r *RunResult) Stderr() string {
	return r.stderr
}

// buildBinary builds the build tool binary and returns its path.
// The binary is built once and reused across all tests.
func buildBinary(t *testing.T) string {
	t.Helper()

	// Determine project root
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get caller information")
	}
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")

	// Build binary in temp directory
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "build")

	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/build")
	cmd.Dir = projectRoot

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v\nstderr:\n%s", err, stderr.String())
	}

	return binPath
}

// simpleBuildfile returns a minimal valid Buildfile for testing.
func simpleBuildfile() string {
	return `.shell: bash

@test:
	echo "Hello, World!"
`
}

// TestSimpleBuild tests building a simple phony target.
func TestSimpleBuild(t *testing.T) {
	h := NewTestHarness(t)
	h.WriteFile("Buildfile", simpleBuildfile())

	result := h.Run("test")
	result.AssertSuccess().
		AssertStdoutContains("Hello, World!")
}

// TestBuildfileDiscovery tests that the build tool finds Buildfile in the current directory.
func TestBuildfileDiscovery(t *testing.T) {
	h := NewTestHarness(t)
	h.WriteFile("Buildfile", simpleBuildfile())

	// Should find Buildfile without -f flag
	result := h.Run("test")
	result.AssertSuccess()
}

// TestMissingBuildfile tests that the build tool exits with an error when Buildfile is missing.
func TestMissingBuildfile(t *testing.T) {
	h := NewTestHarness(t)

	// No Buildfile written
	result := h.Run("test")
	result.AssertExitCode(3). // Parse error
					AssertStderrContains("Buildfile")
}

// TestDefaultTarget tests building the default target when no target is specified.
func TestDefaultTarget(t *testing.T) {
	h := NewTestHarness(t)
	h.WriteFile("Buildfile", `.shell: bash
.default: all

@all:
	echo "Building all"

@test:
	echo "Running tests"
`)

	// No target specified, should build default
	result := h.Run()
	result.AssertSuccess().
		AssertStdoutContains("Building all").
		AssertStdoutNotContains("Running tests")
}

// TestMultipleTargets tests building multiple targets in order.
func TestMultipleTargets(t *testing.T) {
	h := NewTestHarness(t)
	h.WriteFile("Buildfile", `.shell: bash

@first:
	echo "First"

@second:
	echo "Second"

@third:
	echo "Third"
`)

	result := h.Run("first", "second", "third")
	result.AssertSuccess()

	// Check order (first should appear before second, second before third)
	stdout := result.Stdout()
	firstIdx := strings.Index(stdout, "First")
	secondIdx := strings.Index(stdout, "Second")
	thirdIdx := strings.Index(stdout, "Third")

	if firstIdx == -1 || secondIdx == -1 || thirdIdx == -1 {
		t.Errorf("missing expected output\nstdout:\n%s", stdout)
	} else if !(firstIdx < secondIdx && secondIdx < thirdIdx) {
		t.Errorf("targets executed in wrong order\nstdout:\n%s", stdout)
	}
}

// TestDryRun tests that --dry-run shows what would be executed without executing it.
func TestDryRun(t *testing.T) {
	h := NewTestHarness(t)
	h.WriteFile("Buildfile", `.shell: bash

output.txt:
	echo "Creating output" > output.txt
`)

	result := h.Run("--dry-run", "output.txt")
	result.AssertSuccess().
		AssertStdoutContains("echo \"Creating output\" > output.txt")

	// File should not be created
	if h.FileExists("output.txt") {
		t.Error("output.txt should not exist after dry-run")
	}
}

// TestVerboseMode tests that --verbose shows additional information.
func TestVerboseMode(t *testing.T) {
	h := NewTestHarness(t)
	h.WriteFile("Buildfile", `.shell: bash

@test:
	echo "test"
`)

	result := h.Run("--verbose", "test")
	result.AssertSuccess()
	// Verbose mode should show the command being executed
	// (exact format depends on implementation)
}

// TestHelpFlag tests that --help displays usage information.
func TestHelpFlag(t *testing.T) {
	h := NewTestHarness(t)

	result := h.Run("--help")
	result.AssertSuccess().
		AssertStdoutContains("Usage:").
		AssertStdoutContains("--help")
}

// TestVersionFlag tests that --version displays version information.
func TestVersionFlag(t *testing.T) {
	h := NewTestHarness(t)

	result := h.Run("--version")
	result.AssertSuccess()
	// Version output format depends on implementation
}

// TestInvalidFlag tests that an invalid flag produces an error.
func TestInvalidFlag(t *testing.T) {
	h := NewTestHarness(t)
	h.WriteFile("Buildfile", simpleBuildfile())

	result := h.Run("--invalid-flag")
	result.AssertExitCode(2). // Usage error
					AssertStderrContains("flag")
}
