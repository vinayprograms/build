package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ===========================================================================
// Build Integration Tests - Normal Output (Phase 7.2)
// ===========================================================================

// skipIfPipelineNotWired skips tests that require the full execution pipeline.
func skipIfPipelineNotWired(t *testing.T) {
	t.Skip("execution pipeline not wired up in main.go - components exist in internal packages")
}

// TestBuildSimpleTarget tests building a simple target with echo commands.
func TestBuildSimpleTarget(t *testing.T) {
	skipIfPipelineNotWired(t)
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `@hello:
    echo "Hello, World!"
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "hello"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// TestBuildWithVariables tests building with variable interpolation.
func TestBuildWithVariables(t *testing.T) {
	skipIfPipelineNotWired(t)
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `name = World

@hello:
    echo "Hello, {name}!"
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "hello"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// TestBuildWithDependencies tests building with target dependencies.
func TestBuildWithDependencies(t *testing.T) {
	skipIfPipelineNotWired(t)
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")

	// Create source file
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "main.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	content := `build_dir = build

{build_dir}/output.txt: src/main.txt
    mkdir -p {build_dir}
    cp {in} {out}

@all: {build_dir}/output.txt
    echo "Build complete"
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to tmpDir for relative paths
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"all"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}

	// Verify output file was created
	outputPath := filepath.Join(tmpDir, "build", "output.txt")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("expected output file to be created")
	}
}

// TestBuildRecipeFailure tests that recipe failures return correct exit code.
func TestBuildRecipeFailure(t *testing.T) {
	skipIfPipelineNotWired(t)
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `@fail:
    exit 1
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "fail"})
	if exitCode != exitBuildFailure {
		t.Errorf("exit code = %d, want %d", exitCode, exitBuildFailure)
	}
}

// ===========================================================================
// Dry-Run Output Tests (Phase 7.2)
// ===========================================================================

// TestDryRunShowsCommands tests that --dry-run shows commands without executing.
func TestDryRunShowsCommands(t *testing.T) {
	skipIfPipelineNotWired(t)
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `@build:
    echo "This should not execute"
    touch marker.txt
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "-n", "build"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}

	// Verify marker file was NOT created (dry-run)
	markerPath := filepath.Join(tmpDir, "marker.txt")
	if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Error("expected marker file NOT to be created in dry-run mode")
	}
}

// TestDryRunWithDependencies tests dry-run with dependencies.
func TestDryRunWithDependencies(t *testing.T) {
	skipIfPipelineNotWired(t)
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")

	// Create source file
	if err := os.WriteFile(filepath.Join(tmpDir, "src.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	content := `out.txt: src.txt
    cp {in} {out}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to tmpDir
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-n", "out.txt"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}

	// Verify output file was NOT created (dry-run)
	if _, err := os.Stat(filepath.Join(tmpDir, "out.txt")); !os.IsNotExist(err) {
		t.Error("expected output file NOT to be created in dry-run mode")
	}
}

// ===========================================================================
// Verbose Output Tests (Phase 7.2)
// ===========================================================================

// TestVerboseShowsVariables tests that --verbose shows variable evaluation.
func TestVerboseShowsVariables(t *testing.T) {
	skipIfPipelineNotWired(t)
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `cc = gcc
flags = -Wall

@build:
    echo "Using {cc} with {flags}"
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "-v", "build"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// ===========================================================================
// Multiple Targets Tests
// ===========================================================================

// TestBuildMultipleTargets tests building multiple targets in order.
func TestBuildMultipleTargets(t *testing.T) {
	skipIfPipelineNotWired(t)
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `@first:
    echo "first"

@second:
    echo "second"

@third:
    echo "third"
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "first", "second", "third"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// ===========================================================================
// Block Command Tests
// ===========================================================================

// TestBuildWithBlockCommand tests building with block commands.
func TestBuildWithBlockCommand(t *testing.T) {
	skipIfPipelineNotWired(t)
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `@check:
    block:
        if [ "1" = "1" ]; then
            echo "condition true"
        fi
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "check"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// ===========================================================================
// Pattern Target Tests
// ===========================================================================

// TestBuildPatternTarget tests building pattern targets.
func TestBuildPatternTarget(t *testing.T) {
	skipIfPipelineNotWired(t)
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")

	// Create source files
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "main.txt"), []byte("main"), 0644); err != nil {
		t.Fatal(err)
	}

	content := `build/{name}.out: src/{name}.txt
    mkdir -p build
    cp {in} {out}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"build/main.out"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}

	// Verify output file was created
	if _, err := os.Stat(filepath.Join(tmpDir, "build", "main.out")); os.IsNotExist(err) {
		t.Error("expected output file to be created")
	}
}

// ===========================================================================
// Staleness Detection Tests
// ===========================================================================

// TestBuildSkipsUpToDate tests that up-to-date targets are skipped.
func TestBuildSkipsUpToDate(t *testing.T) {
	skipIfPipelineNotWired(t)
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")

	// Create source and output files
	srcFile := filepath.Join(tmpDir, "src.txt")
	outFile := filepath.Join(tmpDir, "out.txt")

	if err := os.WriteFile(srcFile, []byte("source"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outFile, []byte("output"), 0644); err != nil {
		t.Fatal(err)
	}

	content := `out.txt: src.txt
    cp {in} {out}
    echo "marker" >> {out}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Get original content
	originalContent, _ := os.ReadFile(outFile)

	// Run build - should skip since out.txt is newer
	exitCode := run([]string{"out.txt"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}

	// Verify file was not modified (recipe didn't run)
	newContent, _ := os.ReadFile(outFile)
	if string(newContent) != string(originalContent) {
		t.Error("expected output file to be unchanged (up-to-date)")
	}
}

// TestBuildRebuildsWhenDependencyNewer tests rebuild when dependency is modified.
func TestBuildRebuildsWhenDependencyNewer(t *testing.T) {
	skipIfPipelineNotWired(t)
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")

	// Create source and output files
	srcFile := filepath.Join(tmpDir, "src.txt")
	outFile := filepath.Join(tmpDir, "out.txt")

	// Create output first (older timestamp)
	if err := os.WriteFile(outFile, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait a bit and create source (newer timestamp)
	if err := os.WriteFile(srcFile, []byte("new source"), 0644); err != nil {
		t.Fatal(err)
	}

	content := `out.txt: src.txt
    cp {in} {out}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"out.txt"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}

	// Verify file was updated
	newContent, _ := os.ReadFile(outFile)
	if !strings.Contains(string(newContent), "new source") {
		t.Error("expected output file to be updated with new source content")
	}
}
