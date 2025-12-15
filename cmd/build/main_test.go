package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantFile    string
		wantEnv     string
		wantJobs    int
		wantDryRun  bool
		wantVerbose bool
		wantTargets []string
		wantErr     bool
	}{
		{
			name:        "no arguments",
			args:        []string{},
			wantJobs:    1,
			wantTargets: []string{},
		},
		{
			name:        "single target",
			args:        []string{"build/app"},
			wantJobs:    1,
			wantTargets: []string{"build/app"},
		},
		{
			name:        "multiple targets",
			args:        []string{"@clean", "build/app", "@test"},
			wantJobs:    1,
			wantTargets: []string{"@clean", "build/app", "@test"},
		},
		{
			name:     "file flag long",
			args:     []string{"--file", "other.build"},
			wantFile: "other.build",
			wantJobs: 1,
		},
		{
			name:     "file flag short",
			args:     []string{"-f", "custom.build"},
			wantFile: "custom.build",
			wantJobs: 1,
		},
		{
			name:     "env flag long",
			args:     []string{"--env", "ci"},
			wantEnv:  "ci",
			wantJobs: 1,
		},
		{
			name:     "env flag short",
			args:     []string{"-e", "docker"},
			wantEnv:  "docker",
			wantJobs: 1,
		},
		{
			name:     "jobs flag long",
			args:     []string{"--jobs", "8"},
			wantJobs: 8,
		},
		{
			name:     "jobs flag short",
			args:     []string{"-j", "4"},
			wantJobs: 4,
		},
		{
			name:       "dry-run flag long",
			args:       []string{"--dry-run"},
			wantDryRun: true,
			wantJobs:   1,
		},
		{
			name:       "dry-run flag short",
			args:       []string{"-n"},
			wantDryRun: true,
			wantJobs:   1,
		},
		{
			name:        "verbose flag long",
			args:        []string{"--verbose"},
			wantVerbose: true,
			wantJobs:    1,
		},
		{
			name:        "verbose flag short",
			args:        []string{"-v"},
			wantVerbose: true,
			wantJobs:    1,
		},
		{
			name:        "combined flags",
			args:        []string{"-v", "-n", "-j", "4", "-f", "test.build", "target1", "target2"},
			wantFile:    "test.build",
			wantJobs:    4,
			wantDryRun:  true,
			wantVerbose: true,
			wantTargets: []string{"target1", "target2"},
		},
		{
			name:    "invalid flag",
			args:    []string{"--invalid-flag"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, targets, err := parseFlags(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if f.file != tt.wantFile {
				t.Errorf("file = %q, want %q", f.file, tt.wantFile)
			}
			if f.env != tt.wantEnv {
				t.Errorf("env = %q, want %q", f.env, tt.wantEnv)
			}
			if f.jobs != tt.wantJobs {
				t.Errorf("jobs = %d, want %d", f.jobs, tt.wantJobs)
			}
			if f.dryRun != tt.wantDryRun {
				t.Errorf("dryRun = %v, want %v", f.dryRun, tt.wantDryRun)
			}
			if f.verbose != tt.wantVerbose {
				t.Errorf("verbose = %v, want %v", f.verbose, tt.wantVerbose)
			}
			if len(targets) != len(tt.wantTargets) {
				t.Errorf("targets = %v, want %v", targets, tt.wantTargets)
			} else {
				for i, target := range targets {
					if target != tt.wantTargets[i] {
						t.Errorf("targets[%d] = %q, want %q", i, target, tt.wantTargets[i])
					}
				}
			}
		})
	}
}

func TestParseFlagsHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"help long", []string{"--help"}},
		{"help short", []string{"-h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _, err := parseFlags(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !f.showHelp {
				t.Error("showHelp should be true")
			}
		})
	}
}

func TestParseFlagsVersion(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"version long", []string{"--version"}},
		{"version short", []string{"-V"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _, err := parseFlags(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !f.showVersion {
				t.Error("showVersion should be true")
			}
		})
	}
}

func TestRunHelp(t *testing.T) {
	exitCode := run([]string{"--help"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunVersion(t *testing.T) {
	exitCode := run([]string{"--version"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunInvalidFlag(t *testing.T) {
	exitCode := run([]string{"--invalid-flag"})
	if exitCode != exitUsageError {
		t.Errorf("exit code = %d, want %d", exitCode, exitUsageError)
	}
}

func TestRunNoBuildfile(t *testing.T) {
	// Change to a temp directory with no Buildfile
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{})
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunWithBuildfile(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	if err := os.WriteFile(buildfile, []byte("# Test Buildfile\n"), 0644); err != nil {
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

	exitCode := run([]string{})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunWithExplicitBuildfile(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "custom.build")
	if err := os.WriteFile(buildfile, []byte("# Custom Buildfile\n"), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugLex(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `# Sample Buildfile
cc = gcc
cflags = -Wall -O2

build/app: build/main.o
    {cc} -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-lex"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugLexMissingFile(t *testing.T) {
	exitCode := run([]string{"-f", "/nonexistent/Buildfile", "--debug-lex"})
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugParse(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `# Sample Buildfile
.shell: bash
.parallel: 4

cc = gcc
cflags = -Wall -O2

build/app: build/main.o
    .shell: zsh
    .after: build/
    {cc} -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-parse"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugParseMissingFile(t *testing.T) {
	exitCode := run([]string{"-f", "/nonexistent/Buildfile", "--debug-parse"})
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestParseFlagsDebugParse(t *testing.T) {
	f, _, err := parseFlags([]string{"--debug-parse"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.debugParse {
		t.Error("debugParse should be true")
	}
}

func TestParseFlagsDebugVar(t *testing.T) {
	f, _, err := parseFlags([]string{"--debug-var"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.debugVar {
		t.Error("debugVar should be true")
	}
}

func TestParseFlagsDebugTarget(t *testing.T) {
	f, _, err := parseFlags([]string{"--debug-target"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.debugTarget {
		t.Error("debugTarget should be true")
	}
}

func TestParseFlagsDebugRecipe(t *testing.T) {
	f, _, err := parseFlags([]string{"--debug-recipe"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.debugRecipe {
		t.Error("debugRecipe should be true")
	}
}

func TestRunDebugVar(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `# Sample Buildfile with variables
cc = gcc
cflags = -Wall -O2
lazy all_flags = {cflags} {extra}
sources = shell(find src -name *.c)
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-var"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugVarMissingFile(t *testing.T) {
	exitCode := run([]string{"-f", "/nonexistent/Buildfile", "--debug-var"})
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugTarget(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `# Sample Buildfile with targets
build/app: build/main.o build/utils.o
    gcc -o {target} {deps}

build/{name}.o: src/{name}.c
    gcc -c {in} -o {out}

@all: build/app

@clean:
    rm -rf build/

build/:
    mkdir -p {target}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-target"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugTargetMissingFile(t *testing.T) {
	exitCode := run([]string{"-f", "/nonexistent/Buildfile", "--debug-target"})
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugRecipe(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `# Sample Buildfile with recipes
build/app: build/main.o
    .after: build/
    .shell: bash
    gcc -o {target} {deps}

build/{name}.o: src/{name}.c
    .autodeps: build/{name}.d
    .requires: gcc@11 pkg-config@latest
    gcc -MMD -c {in} -o {out}

@clean:
    rm -rf build/

@all: build/app

build/:
    mkdir -p {target}

build/complex: src/main.c
    echo "Starting"
    block:
        if [[ -f {target} ]]; then
            rm {target}
        fi
        gcc -o {target} {deps}
    echo "Finished"
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-recipe"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugRecipeMissingFile(t *testing.T) {
	exitCode := run([]string{"-f", "/nonexistent/Buildfile", "--debug-recipe"})
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestParseFlagsDebugEnv(t *testing.T) {
	f, _, err := parseFlags([]string{"--debug-env"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.debugEnv {
		t.Error("debugEnv should be true")
	}
}

func TestRunDebugEnv(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `# Sample Buildfile with environments
.environment:
    .using: bare
    .requires: gcc@11 python3@latest

.environment: ci
    .using: docker
    .source: Dockerfile.ci
    .args: --platform linux/amd64

.environment: nix-dev
    .using: nix
    .source: shell.nix
    .args: --pure
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-env"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugEnvMissingFile(t *testing.T) {
	exitCode := run([]string{"-f", "/nonexistent/Buildfile", "--debug-env"})
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugEnvNoEnvironments(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `# Sample Buildfile without environments
cc = gcc
build/app: src/main.c
    {cc} -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-env"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestFindBuildfile(t *testing.T) {
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)

	// Test that findBuildfile finds files in the expected order
	// Note: On case-insensitive filesystems (macOS), Buildfile and buildfile are the same
	t.Run("finds_Buildfile", func(t *testing.T) {
		testDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(testDir, "Buildfile"), []byte("# test\n"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(testDir); err != nil {
			t.Fatal(err)
		}

		got := findBuildfile()
		if got != "Buildfile" {
			t.Errorf("findBuildfile() = %q, want %q", got, "Buildfile")
		}
	})

	t.Run("finds_Buildfile.build", func(t *testing.T) {
		testDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(testDir, "Buildfile.build"), []byte("# test\n"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(testDir); err != nil {
			t.Fatal(err)
		}

		got := findBuildfile()
		if got != "Buildfile.build" {
			t.Errorf("findBuildfile() = %q, want %q", got, "Buildfile.build")
		}
	})

	t.Run("prefers_Buildfile_over_Buildfile.build", func(t *testing.T) {
		testDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(testDir, "Buildfile"), []byte("# test\n"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(testDir, "Buildfile.build"), []byte("# test\n"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(testDir); err != nil {
			t.Fatal(err)
		}

		got := findBuildfile()
		if got != "Buildfile" {
			t.Errorf("findBuildfile() = %q, want %q", got, "Buildfile")
		}
	})
}

func TestFindBuildfileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	got := findBuildfile()
	if got != "" {
		t.Errorf("findBuildfile() = %q, want empty string", got)
	}
}

// captureOutput captures stdout during function execution
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestPrintUsage(t *testing.T) {
	output := captureOutput(printUsage)

	// Check for essential content
	essentials := []string{
		"Usage: build",
		"--file",
		"--env",
		"--jobs",
		"--dry-run",
		"--verbose",
		"--help",
		"--version",
		"--debug-lex",
		"--debug-parse",
		"--debug-var",
		"--debug-target",
		"--debug-recipe",
		"--debug-env",
		"--debug-cond",
	}

	for _, s := range essentials {
		if !strings.Contains(output, s) {
			t.Errorf("usage output missing %q", s)
		}
	}
}

func TestParseFlagsDebugCond(t *testing.T) {
	f, _, err := parseFlags([]string{"--debug-cond"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.debugCond {
		t.Error("debugCond should be true")
	}
}

func TestRunDebugCond(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `# Sample Buildfile with conditionals
if {os} == linux
cc = gcc
cflags = -Wall
elif {os} == darwin
cc = clang
cflags = -Wall -Wextra
else
cc = cc
end

ifdef DEBUG
debug_flags = -g -O0
end

ifndef CC
default_cc = gcc
end
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-cond"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugCondMissingFile(t *testing.T) {
	exitCode := run([]string{"-f", "/nonexistent/Buildfile", "--debug-cond"})
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugCondNoConditionals(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `# Sample Buildfile without conditionals
cc = gcc
build/app: src/main.c
    {cc} -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-cond"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestParseFlagsDebugAST(t *testing.T) {
	f, _, err := parseFlags([]string{"--debug-ast"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.debugAST {
		t.Error("debugAST should be true")
	}
}

func TestRunDebugAST(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `# Full buildfile test
.shell: bash
.parallel: 4

cc = gcc
lazy cflags = -Wall

if {os} == linux
cc = gcc
end

.environment:
    .using: bare
    .requires: gcc

@test:
    echo test

build/app: src/main.c
    {cc} -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-ast"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugASTMissingFile(t *testing.T) {
	exitCode := run([]string{"-f", "/nonexistent/Buildfile", "--debug-ast"})
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugASTWithErrors(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `# Buildfile with errors
.after: invalid
cc = gcc
.using: invalid
cflags = -Wall
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-ast"})
	// Should return exitParseError because there are parse errors
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugASTEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := ``
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-ast"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestPrintUsageIncludesDebugAST(t *testing.T) {
	output := captureOutput(printUsage)

	if !strings.Contains(output, "--debug-ast") {
		t.Error("usage output should include --debug-ast")
	}
}

func TestRunDebugSemantic(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `# Test buildfile for semantic analysis
cc = gcc
lazy cflags = -Wall

.environment:
    .using: bare
    .requires: gcc

.environment: ci
    .using: docker
    .source: Dockerfile.ci

@clean:
    rm -rf build/

build/app: src/main.c
    {cc} -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-semantic"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugSemanticMissingFile(t *testing.T) {
	exitCode := run([]string{"-f", "/nonexistent/Buildfile", "--debug-semantic"})
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugSemanticDuplicates(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `# Buildfile with duplicate definitions
cc = gcc
cc = clang
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-semantic"})
	// Should return exitParseError because there are semantic errors (duplicate variable)
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugSemanticEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := ``
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-semantic"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestPrintUsageIncludesDebugSemantic(t *testing.T) {
	output := captureOutput(printUsage)

	if !strings.Contains(output, "--debug-semantic") {
		t.Error("usage output should include --debug-semantic")
	}
}

func TestRunDebugSemanticCaptureValidation(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test capture validation with pattern targets
	content := `# Buildfile with pattern targets
base = build

# Pattern target: {name} is a capture
{base}/{name}.o: src/{name}.c
    gcc -c {in} -o {out}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-semantic"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugSemanticCaptureValidationError(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test capture validation with automatic variable in pattern (error)
	content := `# Buildfile with invalid pattern
build/{target}.o: src/main.c
    gcc -c {in} -o {out}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-semantic"})
	// Should return exitParseError because {target} is an automatic variable
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugSemanticCaptureMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test capture validation with capture mismatch (capture in dep but not in target)
	content := `# Buildfile with capture mismatch
build/app: src/{name}.c
    gcc -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-semantic"})
	// Should return exitParseError because {name} is in dependency but not in target
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugSemanticReferenceValidation(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test reference validation with all variable types
	content := `# Buildfile with various variable references
cc = gcc
cflags = -Wall

# Use defined variables and automatic variables in recipe
build/app: src/main.c
    {cc} {cflags} -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-semantic"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugSemanticReferenceUndefined(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test reference validation with undefined variable
	content := `# Buildfile with undefined variable
cc = gcc

build/app: src/main.c
    {cc} {undefined_flags} -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-semantic"})
	// Should return exitParseError because undefined_flags is not defined
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugSemanticReferenceAutomaticOutsideRecipe(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test reference validation with automatic variable outside recipe
	content := `# Buildfile with automatic variable outside recipe
output = {target}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-semantic"})
	// Should return exitParseError because {target} is only valid in recipe
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugSemanticReferenceBuiltin(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test reference validation with built-in variables
	content := `# Buildfile with built-in variables
platform = {os}-{arch}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-semantic"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// ===========================================================================
// Dependency Graph Validation Tests (Pass 4)
// ===========================================================================

func TestRunDebugSemanticDependencyGraph(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test dependency graph with valid dependencies
	content := `# Buildfile with dependency graph
build/app: build/main.o build/utils.o
    gcc -o {target} {deps}

build/main.o: src/main.c
    gcc -c {in} -o {out}

build/utils.o: src/utils.c
    gcc -c {in} -o {out}

@all: build/app

@clean:
    rm -rf build/
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-semantic"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugSemanticCircularDependency(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test circular dependency detection
	content := `# Buildfile with circular dependency
dir/a.o: dir/b.o
    echo a

dir/b.o: dir/c.o
    echo b

dir/c.o: dir/a.o
    echo c
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-semantic"})
	// Should return exitParseError because of circular dependency
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugSemanticSelfLoop(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test self-loop detection
	content := `# Buildfile with self-loop
build/app.o: build/app.o
    echo self
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-semantic"})
	// Should return exitParseError because of self-loop
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugSemanticPatternTarget(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test pattern targets are tracked separately
	content := `# Buildfile with pattern target
build/{name}.o: src/{name}.c
    gcc -c {in} -o {out}

@all: build/main.o
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-semantic"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugSemanticDiamondDependency(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test diamond dependency (valid - no cycle)
	content := `# Buildfile with diamond dependency
build/app: build/a.o build/b.o
    gcc -o {target} {deps}

build/a.o: build/common.o src/a.c
    gcc -c src/a.c -o {target}

build/b.o: build/common.o src/b.c
    gcc -c src/b.c -o {target}

build/common.o: src/common.c
    gcc -c {in} -o {out}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := run([]string{"-f", buildfile, "--debug-semantic"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}
