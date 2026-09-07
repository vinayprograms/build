package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// ===========================================================================
// Target Argument Resolution Tests (Phase 7.1)
// ===========================================================================

// TestResolveTargetsNoArgs tests that with no target arguments:
// - Uses .default: directive if present
// - Otherwise uses first defined target
func TestResolveTargetsNoArgs(t *testing.T) {
	tests := []struct {
		name         string
		needfile    string
		wantTargets  []string
		wantErr      bool
		wantErrMatch string
	}{
		{
			name: "use default directive",
			needfile: `.default: build/app

build/app: src/main.c
    gcc -o {target} {deps}

@test:
    echo test
`,
			wantTargets: []string{"build/app"},
		},
		{
			name: "use first target when no default",
			needfile: `build/app: src/main.c
    gcc -o {target} {deps}

@test:
    echo test
`,
			wantTargets: []string{"build/app"},
		},
		{
			name: "use first phony target when no default",
			needfile: `@all: build/app

build/app: src/main.c
    gcc -o {target} {deps}
`,
			wantTargets: []string{"@all"},
		},
		{
			name: "no targets defined",
			needfile: `# Just comments
cc = gcc
`,
			wantErr:      true,
			wantErrMatch: "no targets defined",
		},
		{
			name: "default directive with phony target",
			needfile: `.default: @all

@all: build/app

build/app: src/main.c
    gcc -o {target} {deps}
`,
			wantTargets: []string{"@all"},
		},
		{
			name: "default directive without @ prefix for phony",
			needfile: `.default: all

@all: build/app

build/app: src/main.c
    gcc -o {target} {deps}
`,
			wantTargets: []string{"@all"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			needfile := filepath.Join(tmpDir, "Needfile")
			if err := os.WriteFile(needfile, []byte(tt.needfile), 0644); err != nil {
				t.Fatal(err)
			}

			// Parse the needfile
			content, err := os.ReadFile(needfile)
			if err != nil {
				t.Fatalf("failed to read needfile: %v", err)
			}

			l := NewLexer(needfile, string(content))
			p := NewParser(l)
			bp := NewNeedfileParser(p)
			result := bp.ParseNeedfile()

			if result.HasErrors() {
				t.Fatalf("parse errors: %s", result.AllErrors())
			}

			targets, err := ResolveTargetArgs(nil, result, nil)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.wantErrMatch != "" && err.Error() != tt.wantErrMatch {
					t.Errorf("error = %q, want match %q", err.Error(), tt.wantErrMatch)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(targets) != len(tt.wantTargets) {
				t.Errorf("got %d targets, want %d", len(targets), len(tt.wantTargets))
			}

			for i, want := range tt.wantTargets {
				if i >= len(targets) {
					break
				}
				if targets[i] != want {
					t.Errorf("targets[%d] = %q, want %q", i, targets[i], want)
				}
			}
		})
	}
}

// TestResolveTargetsSingleArg tests resolution with a single target argument.
func TestResolveTargetsSingleArg(t *testing.T) {
	tests := []struct {
		name         string
		needfile    string
		args         []string
		wantTargets  []string
		wantErr      bool
		wantErrMatch string
	}{
		{
			name: "explicit file target",
			needfile: `build/app: src/main.c
    gcc -o {target} {deps}
`,
			args:        []string{"build/app"},
			wantTargets: []string{"build/app"},
		},
		{
			name: "phony target with @ prefix",
			needfile: `@clean:
    rm -rf build/
`,
			args:        []string{"@clean"},
			wantTargets: []string{"@clean"},
		},
		{
			name: "phony target without @ prefix",
			needfile: `@clean:
    rm -rf build/
`,
			args:        []string{"clean"},
			wantTargets: []string{"@clean"},
		},
		{
			name: "phony target named same as file target - @ prefers phony",
			needfile: `@test:
    echo phony test

build/test: src/test.c
    gcc -o build/test src/test.c
`,
			args:        []string{"@test"},
			wantTargets: []string{"@test"},
		},
		{
			name: "file target match prefers exact path",
			needfile: `build/test: src/test.c
    gcc -o build/test src/test.c

@test:
    echo phony test
`,
			args:        []string{"build/test"},
			wantTargets: []string{"build/test"},
		},
		{
			name: "target not found",
			needfile: `build/app: src/main.c
    gcc -o {target} {deps}
`,
			args:         []string{"nonexistent"},
			wantErr:      true,
			wantErrMatch: "target 'nonexistent' not found",
		},
		{
			name: "pattern target matches",
			needfile: `build/{name}.o: src/{name}.c
    gcc -c {in} -o {out}
`,
			args:        []string{"build/main.o"},
			wantTargets: []string{"build/main.o"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			needfile := filepath.Join(tmpDir, "Needfile")
			if err := os.WriteFile(needfile, []byte(tt.needfile), 0644); err != nil {
				t.Fatal(err)
			}

			content, err := os.ReadFile(needfile)
			if err != nil {
				t.Fatalf("failed to read needfile: %v", err)
			}

			l := NewLexer(needfile, string(content))
			p := NewParser(l)
			bp := NewNeedfileParser(p)
			result := bp.ParseNeedfile()

			if result.HasErrors() {
				t.Fatalf("parse errors: %s", result.AllErrors())
			}

			targets, err := ResolveTargetArgs(tt.args, result, nil)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.wantErrMatch != "" && err.Error() != tt.wantErrMatch {
					t.Errorf("error = %q, want match %q", err.Error(), tt.wantErrMatch)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(targets) != len(tt.wantTargets) {
				t.Errorf("got %d targets, want %d", len(targets), len(tt.wantTargets))
			}

			for i, want := range tt.wantTargets {
				if i >= len(targets) {
					break
				}
				if targets[i] != want {
					t.Errorf("targets[%d] = %q, want %q", i, targets[i], want)
				}
			}
		})
	}
}

// TestResolveTargetsMultipleArgs tests resolution with multiple target arguments.
func TestResolveTargetsMultipleArgs(t *testing.T) {
	tests := []struct {
		name        string
		needfile   string
		args        []string
		wantTargets []string
		wantErr     bool
	}{
		{
			name: "multiple file targets",
			needfile: `build/app: src/main.c
    gcc -o {target} {deps}

build/lib.a: src/lib.c
    ar rcs {target} {deps}
`,
			args:        []string{"build/app", "build/lib.a"},
			wantTargets: []string{"build/app", "build/lib.a"},
		},
		{
			name: "mixed phony and file targets",
			needfile: `@clean:
    rm -rf build/

build/app: src/main.c
    gcc -o {target} {deps}

@test: build/app
    ./build/app --test
`,
			args:        []string{"clean", "build/app", "test"},
			wantTargets: []string{"@clean", "build/app", "@test"},
		},
		{
			name: "preserve order",
			needfile: `@a:
    echo a

@b:
    echo b

@c:
    echo c
`,
			args:        []string{"c", "a", "b"},
			wantTargets: []string{"@c", "@a", "@b"},
		},
		{
			name: "one missing target fails all",
			needfile: `build/app: src/main.c
    gcc -o {target} {deps}
`,
			args:    []string{"build/app", "missing"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			needfile := filepath.Join(tmpDir, "Needfile")
			if err := os.WriteFile(needfile, []byte(tt.needfile), 0644); err != nil {
				t.Fatal(err)
			}

			content, err := os.ReadFile(needfile)
			if err != nil {
				t.Fatalf("failed to read needfile: %v", err)
			}

			l := NewLexer(needfile, string(content))
			p := NewParser(l)
			bp := NewNeedfileParser(p)
			result := bp.ParseNeedfile()

			if result.HasErrors() {
				t.Fatalf("parse errors: %s", result.AllErrors())
			}

			targets, err := ResolveTargetArgs(tt.args, result, nil)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(targets) != len(tt.wantTargets) {
				t.Errorf("got %d targets, want %d", len(targets), len(tt.wantTargets))
			}

			for i, want := range tt.wantTargets {
				if i >= len(targets) {
					break
				}
				if targets[i] != want {
					t.Errorf("targets[%d] = %q, want %q", i, targets[i], want)
				}
			}
		})
	}
}

// TestResolveTargetArgsEmptySlice tests that empty slice behaves like nil (uses default).
func TestResolveTargetArgsEmptySlice(t *testing.T) {
	needfile := `.default: build/app

build/app: src/main.c
    gcc -o {target} {deps}
`
	tmpDir := t.TempDir()
	needfilePath := filepath.Join(tmpDir, "Needfile")
	if err := os.WriteFile(needfilePath, []byte(needfile), 0644); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(needfilePath)
	if err != nil {
		t.Fatalf("failed to read needfile: %v", err)
	}

	l := NewLexer(needfilePath, string(content))
	p := NewParser(l)
	bp := NewNeedfileParser(p)
	result := bp.ParseNeedfile()

	if result.HasErrors() {
		t.Fatalf("parse errors: %s", result.AllErrors())
	}

	// Empty slice should behave like nil
	targets, err := ResolveTargetArgs([]string{}, result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(targets) != 1 {
		t.Errorf("got %d targets, want 1", len(targets))
	}

	if len(targets) > 0 && targets[0] != "build/app" {
		t.Errorf("targets[0] = %q, want %q", targets[0], "build/app")
	}
}

// ===========================================================================
// CLI Integration Tests for Target Resolution
// ===========================================================================

// TestRunWithTargetArgs tests the CLI with explicit target arguments.
func TestRunWithTargetArgs(t *testing.T) {
	tmpDir := t.TempDir()
	needfile := filepath.Join(tmpDir, "Needfile")
	content := `.shell: bash

@build:
    echo "Building..."

@clean:
    echo "Cleaning..."

@test: @build
    echo "Testing..."
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with explicit phony target
	exitCode := Run([]string{"-f", needfile, "-v", "build"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d (for @build target)", exitCode, exitSuccess)
	}

	// Test with explicit phony target (with @)
	exitCode = Run([]string{"-f", needfile, "-v", "@clean"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d (for @clean target)", exitCode, exitSuccess)
	}

	// Test with phony target (without @)
	exitCode = Run([]string{"-f", needfile, "-v", "clean"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d (for clean target without @)", exitCode, exitSuccess)
	}

	// Test with multiple targets
	exitCode = Run([]string{"-f", needfile, "-v", "clean", "build", "test"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d (for multiple targets)", exitCode, exitSuccess)
	}
}

// TestRunWithDefaultTarget tests the CLI with no target arguments (uses default).
func TestRunWithDefaultTarget(t *testing.T) {
	tmpDir := t.TempDir()
	needfile := filepath.Join(tmpDir, "Needfile")
	content := `.shell: bash
.default: @all

@all: @build
    echo "All done"

@build:
    echo "Building..."
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with no target arguments
	exitCode := Run([]string{"-f", needfile, "-v"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d (for default target)", exitCode, exitSuccess)
	}
}

// TestRunWithFirstTargetAsDefault tests that first target is used when no .default directive.
func TestRunWithFirstTargetAsDefault(t *testing.T) {
	tmpDir := t.TempDir()
	needfile := filepath.Join(tmpDir, "Needfile")
	content := `.shell: bash

@build:
    echo "Building..."

@clean:
    echo "Cleaning..."
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with no target arguments - should use first target (@build)
	exitCode := Run([]string{"-f", needfile, "-v"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d (for first target as default)", exitCode, exitSuccess)
	}
}

// TestRunWithUnknownTarget tests the CLI with an unknown target.
func TestRunWithUnknownTarget(t *testing.T) {
	tmpDir := t.TempDir()
	needfile := filepath.Join(tmpDir, "Needfile")
	content := `build/app: src/main.c
    gcc -o {target} {deps}
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with unknown target - should fail with usage error
	exitCode := Run([]string{"-f", needfile, "nonexistent"})
	if exitCode != exitUsageError {
		t.Errorf("exit code = %d, want %d (for unknown target)", exitCode, exitUsageError)
	}
}

// TestRunWithNoTargets tests the CLI with a needfile that has no targets.
func TestRunWithNoTargets(t *testing.T) {
	tmpDir := t.TempDir()
	needfile := filepath.Join(tmpDir, "Needfile")
	content := `# Just variables, no targets
cc = gcc
cflags = -Wall
`
	if err := os.WriteFile(needfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with no targets defined - should fail with usage error
	exitCode := Run([]string{"-f", needfile})
	if exitCode != exitUsageError {
		t.Errorf("exit code = %d, want %d (for no targets)", exitCode, exitUsageError)
	}
}
