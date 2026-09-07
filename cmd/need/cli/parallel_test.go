package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ===========================================================================
// Parallel Directive Extraction Tests
// ===========================================================================

func TestGetParallelDirective_NotSet(t *testing.T) {
	tmpDir := t.TempDir()
	needfile := filepath.Join(tmpDir, "Needfile")
	content := `@all:
    echo "test"
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, _, err := ParseNeedfileWithCache(needfile)
	if err != nil {
		t.Fatal(err)
	}

	parallel := GetParallelDirective(result)
	if parallel != 0 {
		t.Errorf("parallel = %d, want 0 (not set)", parallel)
	}
}

func TestGetParallelDirective_Set(t *testing.T) {
	tmpDir := t.TempDir()
	needfile := filepath.Join(tmpDir, "Needfile")
	content := `.parallel: 4

@all:
    echo "test"
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, _, err := ParseNeedfileWithCache(needfile)
	if err != nil {
		t.Fatal(err)
	}

	parallel := GetParallelDirective(result)
	if parallel != 4 {
		t.Errorf("parallel = %d, want 4", parallel)
	}
}

func TestGetParallelDirective_Set8(t *testing.T) {
	tmpDir := t.TempDir()
	needfile := filepath.Join(tmpDir, "Needfile")
	content := `.parallel: 8

@all:
    echo "test"
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, _, err := ParseNeedfileWithCache(needfile)
	if err != nil {
		t.Fatal(err)
	}

	parallel := GetParallelDirective(result)
	if parallel != 8 {
		t.Errorf("parallel = %d, want 8", parallel)
	}
}

// ===========================================================================
// Worker Count Resolution Tests
// ===========================================================================

func TestResolveWorkerCount_CLIOverridesDirective(t *testing.T) {
	tests := []struct {
		name        string
		cliJobs     int
		directive   int
		wantWorkers int
	}{
		{"CLI 4, directive 2", 4, 2, 4},
		{"CLI 8, directive 4", 8, 4, 8},
		{"CLI 1, directive 4", 1, 4, 4}, // CLI default, use directive
		{"CLI 1, directive 0", 1, 0, 1}, // CLI default, no directive
		{"CLI 0, directive 4", 0, 4, 4}, // Invalid CLI, use directive
		{"CLI 0, directive 0", 0, 0, 1}, // Both invalid, default to 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveWorkerCount(tt.cliJobs, tt.directive)
			if got != tt.wantWorkers {
				t.Errorf("ResolveWorkerCount(%d, %d) = %d, want %d",
					tt.cliJobs, tt.directive, got, tt.wantWorkers)
			}
		})
	}
}

// ===========================================================================
// Parallel Execution Integration Tests
// ===========================================================================

func TestRunWithParallelDirective_ExecutesInParallel(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source files that the targets depend on
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.c", "b.c", "c.c", "d.c"} {
		if err := os.WriteFile(filepath.Join(srcDir, name), []byte("// "+name), 0644); err != nil {
			t.Fatal(err)
		}
	}

	needfile := filepath.Join(tmpDir, "Needfile")
	// Each target sleeps briefly - with parallelism, total time should be less
	// than sequential execution would take
	content := `.parallel: 4

@all: @a @b @c @d

@a:
    sleep 0.1
    echo "a done"

@b:
    sleep 0.1
    echo "b done"

@c:
    sleep 0.1
    echo "c done"

@d:
    sleep 0.1
    echo "d done"
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	exitCode := Run([]string{"-f", needfile, "@all"})
	elapsed := time.Since(start)

	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}

	// With 4 parallel workers, 4 tasks of 0.1s each should complete in ~0.1-0.2s
	// Sequential would take ~0.4s
	// Allow some overhead, but if it takes more than 0.35s, parallelism isn't working
	if elapsed > 350*time.Millisecond {
		t.Errorf("parallel execution took %v, expected < 350ms (parallelism may not be working)", elapsed)
	}
}

func TestRunWithCLIJobsFlag_OverridesDirective(t *testing.T) {
	tmpDir := t.TempDir()

	needfile := filepath.Join(tmpDir, "Needfile")
	content := `.parallel: 1

@all: @a @b

@a:
    sleep 0.1
    echo "a done"

@b:
    sleep 0.1
    echo "b done"
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	// Override with -j2, which should parallelize despite .parallel: 1
	exitCode := Run([]string{"-f", needfile, "-j", "2", "@all"})
	elapsed := time.Since(start)

	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}

	// With -j2 override, 2 tasks of 0.1s should complete in ~0.1-0.15s
	// Sequential (with .parallel: 1) would take ~0.2s
	if elapsed > 180*time.Millisecond {
		t.Errorf("-j flag override took %v, expected < 180ms (override may not be working)", elapsed)
	}
}

func TestRunWithParallel_RespectsDependencies(t *testing.T) {
	tmpDir := t.TempDir()

	needfile := filepath.Join(tmpDir, "Needfile")
	// c depends on a and b, so a and b can run in parallel but c must wait
	content := `.parallel: 4

@c: @a @b
    echo "c done"

@a:
    echo "a done"

@b:
    echo "b done"
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", needfile, "@c"})

	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// ===========================================================================
// Concurrency Safety Tests
// ===========================================================================

func TestParallelExecution_NoDataRace(t *testing.T) {
	// This test is primarily for the race detector
	tmpDir := t.TempDir()

	needfile := filepath.Join(tmpDir, "Needfile")
	content := `.parallel: 8

@all: @t1 @t2 @t3 @t4 @t5 @t6 @t7 @t8

@t1:
    echo "1"

@t2:
    echo "2"

@t3:
    echo "3"

@t4:
    echo "4"

@t5:
    echo "5"

@t6:
    echo "6"

@t7:
    echo "7"

@t8:
    echo "8"
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Run multiple times to increase chance of detecting races
	for i := 0; i < 3; i++ {
		exitCode := Run([]string{"-f", needfile, "@all"})
		if exitCode != exitSuccess {
			t.Errorf("iteration %d: exit code = %d, want %d", i, exitCode, exitSuccess)
		}
	}
}

func TestParallelExecution_FailureStopsOtherTasks(t *testing.T) {
	tmpDir := t.TempDir()

	needfile := filepath.Join(tmpDir, "Needfile")
	// @fail will fail, @slow depends on @fail so should be skipped
	content := `.parallel: 2

@all: @fail @slow

@slow: @fail
    echo "slow done"

@fail:
    exit 1
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", needfile, "@all"})

	if exitCode != exitBuildFailure {
		t.Errorf("exit code = %d, want %d", exitCode, exitBuildFailure)
	}
}

// ===========================================================================
// Keep-Going Mode Tests
// ===========================================================================

func TestParallelExecution_KeepGoingContinuesAfterFailure(t *testing.T) {
	tmpDir := t.TempDir()

	needfile := filepath.Join(tmpDir, "Needfile")
	// With keep-going, @success should still run even though @fail fails
	// Note: --keep flag controls keep-going mode
	content := `.parallel: 2

@all: @fail @success

@fail:
    exit 1

@success:
    echo "success"
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// --keep flag enables keep-going mode
	exitCode := Run([]string{"-f", needfile, "--keep", "@all"})

	// Should still return failure because @fail failed
	if exitCode != exitBuildFailure {
		t.Errorf("exit code = %d, want %d", exitCode, exitBuildFailure)
	}
}

// ===========================================================================
// Sequential Execution Tests (when parallel=1)
// ===========================================================================

func TestRunWithoutParallel_ExecutesSequentially(t *testing.T) {
	tmpDir := t.TempDir()

	needfile := filepath.Join(tmpDir, "Needfile")
	// No .parallel directive, should run sequentially
	content := `@all: @a @b @c

@a:
    echo "a"

@b:
    echo "b"

@c:
    echo "c"
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", needfile, "@all"})

	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// ===========================================================================
// B6: Output printed exactly once (parallel and sequential)
// ===========================================================================

// TestRunWithParallel_OutputPrintedOnce covers B6: with parallel workers,
// each command's output must be printed exactly once, not duplicated by a
// re-report in the results-collection loop.
func TestRunWithParallel_OutputPrintedOnce(t *testing.T) {
	tmpDir := t.TempDir()
	needfile := filepath.Join(tmpDir, "Needfile")
	content := `.parallel: 2

@all: @clean

@clean:
    echo "Cleaning build artifacts"
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	output := captureOutput(func() {
		exitCode := Run([]string{"-f", needfile, "@all"})
		if exitCode != exitSuccess {
			t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
		}
	})

	count := countExactLines(output, "Cleaning build artifacts")
	if count != 1 {
		t.Errorf("output contained %d occurrences of command output, want 1:\n%s", count, output)
	}
}

// TestRunSequential_OutputPrintedOnce covers B6's sequential-path check: -v
// (verbose) mode must not duplicate output either.
func TestRunSequential_OutputPrintedOnce(t *testing.T) {
	tmpDir := t.TempDir()
	needfile := filepath.Join(tmpDir, "Needfile")
	content := `@clean:
    echo "Cleaning build artifacts"
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	output := captureOutput(func() {
		exitCode := Run([]string{"-v", "-f", needfile, "@clean"})
		if exitCode != exitSuccess {
			t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
		}
	})

	count := countExactLines(output, "Cleaning build artifacts")
	if count != 1 {
		t.Errorf("verbose output contained %d occurrences of command output, want 1:\n%s", count, output)
	}
}

// countExactLines counts lines in output whose trimmed content exactly
// equals want. Unlike strings.Count, this doesn't also match want when it
// appears as a substring of a different line - e.g. the echoed command line
// `echo "Cleaning build artifacts"` contains the substring "Cleaning build
// artifacts" but isn't the same line as the command's actual stdout output.
// CLIWriter (the default writer for both TTY and plain non-TTY output as of
// C5) always echoes the command before running it, so a plain substring
// count would over-count by one for any command whose text contains its own
// output.
func countExactLines(output, want string) int {
	count := 0
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == want {
			count++
		}
	}
	return count
}
