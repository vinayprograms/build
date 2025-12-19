package cli

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

func TestParseFlagsQuiet(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"quiet long", []string{"--quiet"}},
		{"quiet short", []string{"-q"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _, err := parseFlags(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !f.quiet {
				t.Error("quiet should be true")
			}
		})
	}
}

func TestParseFlagsColor(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantColor string
	}{
		{"color auto", []string{"--color=auto"}, "auto"},
		{"color always", []string{"--color=always"}, "always"},
		{"color never", []string{"--color=never"}, "never"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _, err := parseFlags(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if f.color != tt.wantColor {
				t.Errorf("color = %q, want %q", f.color, tt.wantColor)
			}
		})
	}
}

func TestParseFlagsProgress(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantProgress string
	}{
		{"progress auto", []string{"--progress=auto"}, "auto"},
		{"progress always", []string{"--progress=always"}, "always"},
		{"progress never", []string{"--progress=never"}, "never"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _, err := parseFlags(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if f.progress != tt.wantProgress {
				t.Errorf("progress = %q, want %q", f.progress, tt.wantProgress)
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
	exitCode := Run([]string{"--help"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunVersion(t *testing.T) {
	exitCode := Run([]string{"--version"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunInvalidFlag(t *testing.T) {
	exitCode := Run([]string{"--invalid-flag"})
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

	exitCode := Run([]string{})
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunWithBuildfile(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `# Test Buildfile
@all:
    echo "all"
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

	exitCode := Run([]string{})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunWithExplicitBuildfile(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "custom.build")
	content := `# Custom Buildfile
@all:
    echo "all"
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-lex"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugLexMissingFile(t *testing.T) {
	exitCode := Run([]string{"-f", "/nonexistent/Buildfile", "--debug-lex"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-parse"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugParseMissingFile(t *testing.T) {
	exitCode := Run([]string{"-f", "/nonexistent/Buildfile", "--debug-parse"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-var"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugVarMissingFile(t *testing.T) {
	exitCode := Run([]string{"-f", "/nonexistent/Buildfile", "--debug-var"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-target"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugTargetMissingFile(t *testing.T) {
	exitCode := Run([]string{"-f", "/nonexistent/Buildfile", "--debug-target"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-recipe"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugRecipeMissingFile(t *testing.T) {
	exitCode := Run([]string{"-f", "/nonexistent/Buildfile", "--debug-recipe"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-env"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugEnvMissingFile(t *testing.T) {
	exitCode := Run([]string{"-f", "/nonexistent/Buildfile", "--debug-env"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-env"})
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

func TestFindBuildfileInParentDirectory(t *testing.T) {
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)

	// Create a directory structure with Buildfile in parent
	parentDir := t.TempDir()
	childDir := filepath.Join(parentDir, "subdir")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Put Buildfile in parent
	buildfilePath := filepath.Join(parentDir, "Buildfile")
	if err := os.WriteFile(buildfilePath, []byte("# test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to child directory
	if err := os.Chdir(childDir); err != nil {
		t.Fatal(err)
	}

	// findBuildfile should find the Buildfile in parent
	got := findBuildfile()
	if got == "" {
		t.Error("findBuildfile() = \"\", want path to parent Buildfile")
	}
	if !strings.Contains(got, "Buildfile") {
		t.Errorf("findBuildfile() = %q, want path containing 'Buildfile'", got)
	}
}

func TestFindBuildfileInGrandparentDirectory(t *testing.T) {
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)

	// Create a directory structure with Buildfile in grandparent
	grandparentDir := t.TempDir()
	parentDir := filepath.Join(grandparentDir, "parent")
	childDir := filepath.Join(parentDir, "child")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Put Buildfile in grandparent
	buildfilePath := filepath.Join(grandparentDir, "Buildfile")
	if err := os.WriteFile(buildfilePath, []byte("# test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to child directory
	if err := os.Chdir(childDir); err != nil {
		t.Fatal(err)
	}

	// findBuildfile should find the Buildfile in grandparent
	got := findBuildfile()
	if got == "" {
		t.Error("findBuildfile() = \"\", want path to grandparent Buildfile")
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

	exitCode := Run([]string{"-f", buildfile, "--debug-cond"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugCondMissingFile(t *testing.T) {
	exitCode := Run([]string{"-f", "/nonexistent/Buildfile", "--debug-cond"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-cond"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-ast"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugASTMissingFile(t *testing.T) {
	exitCode := Run([]string{"-f", "/nonexistent/Buildfile", "--debug-ast"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-ast"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-ast"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-semantic"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugSemanticMissingFile(t *testing.T) {
	exitCode := Run([]string{"-f", "/nonexistent/Buildfile", "--debug-semantic"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-semantic"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-semantic"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-semantic"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-semantic"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-semantic"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-semantic"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-semantic"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-semantic"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-semantic"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-semantic"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-semantic"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-semantic"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-semantic"})
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

	exitCode := Run([]string{"-f", buildfile, "--debug-semantic"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// ===========================================================================
// Variable Evaluation Tests (Phase 2)
// ===========================================================================

func TestParseFlagsDebugEval(t *testing.T) {
	f, _, err := parseFlags([]string{"--debug-eval"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.debugEval {
		t.Error("debugEval should be true")
	}
}

func TestRunDebugEval(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test variable evaluation
	content := `# Test buildfile for variable evaluation
cc = gcc
cflags = -Wall -O2
prefix = /usr/local
build_dir = build

# Variables using interpolation
full_cc = {cc} {cflags}
install_path = {prefix}/bin
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-eval"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugEvalMissingFile(t *testing.T) {
	exitCode := Run([]string{"-f", "/nonexistent/Buildfile", "--debug-eval"})
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugEvalLazyVariables(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test lazy variable evaluation
	content := `# Buildfile with lazy variables
base = build
lazy all_flags = {cflags} -DNDEBUG
cflags = -Wall
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-eval"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugEvalBuiltins(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test built-in variable evaluation
	content := `# Buildfile with built-in variables
platform = {os}-{arch}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-eval"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugEvalEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := ``
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-eval"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugEvalFunctions(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test function evaluation (basename, dirname, replace)
	content := `# Buildfile with functions
path = /usr/local/bin/app
base = basename({path})
dir = dirname({path})
result = replace(foo.c, .c, .o)
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-eval"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugEvalUndefinedVariable(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test undefined variable error (forward reference)
	content := `# Buildfile with undefined variable
foo = {bar}
bar = value
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-eval"})
	// Should return exitParseError because foo references bar before it's defined
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestPrintUsageIncludesDebugEval(t *testing.T) {
	output := captureOutput(printUsage)

	if !strings.Contains(output, "--debug-eval") {
		t.Error("usage output should include --debug-eval")
	}
}

func TestRunDebugEvalWithConditional(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test conditional evaluation - NOTE: We can't override 'os' built-in, so use a different pattern
	// We test with ifdef which doesn't rely on the os built-in
	content := `# Buildfile with conditional
DEBUG = 1
ifdef DEBUG
cflags = -g -O0
end
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-eval"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// ===========================================================================
// Build Planning Tests (Phase 3)
// ===========================================================================

func TestParseFlagsDebugPlan(t *testing.T) {
	f, _, err := parseFlags([]string{"--debug-plan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.debugPlan {
		t.Error("debugPlan should be true")
	}
}

func TestRunDebugPlan(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test build planning with various target types
	content := `# Test buildfile for build planning
cc = gcc
cflags = -Wall

# Literal file targets
build/app: build/main.o build/utils.o
    {cc} {cflags} -o {target} {deps}

build/main.o: src/main.c
    {cc} -c {in} -o {out}

build/utils.o: src/utils.c
    {cc} -c {in} -o {out}

# Phony targets
@all: build/app

@clean:
    rm -rf build/

@test: build/app
    ./build/app --test

# Directory target
build/:
    mkdir -p {target}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-plan"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugPlanMissingFile(t *testing.T) {
	exitCode := Run([]string{"-f", "/nonexistent/Buildfile", "--debug-plan"})
	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
}

func TestRunDebugPlanPatternTargets(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test pattern targets
	content := `# Buildfile with pattern targets
build/{name}.o: src/{name}.c
    gcc -c {in} -o {out}

build/app: build/main.o build/utils.o
    gcc -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-plan"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugPlanEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := ``
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-plan"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestPrintUsageIncludesDebugPlan(t *testing.T) {
	output := captureOutput(printUsage)

	if !strings.Contains(output, "--debug-plan") {
		t.Error("usage output should include --debug-plan")
	}
}

// ===========================================================================
// Command Interpolation Tests (Phase 4.1 - Automatic Variable Resolution)
// ===========================================================================

func TestRunDebugPlanCommandInterpolation(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test command interpolation with automatic variables
	content := `# Buildfile with automatic variables in commands
cc = gcc
cflags = -Wall

build/app: src/main.c
    {cc} {cflags} -o {target} {deps}
    echo "Built {out} from {in}"
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-plan"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugPlanCaptureVariables(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test command interpolation with captures
	content := `# Buildfile with captures in commands
build/{name}.o: src/{name}.c
    gcc -c {in} -o {out}
    echo "Compiling {name}"
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-plan"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugPlanTargetDirAndFile(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test target.dir and target.file automatic variables
	content := `# Buildfile with target.dir and target.file
build/output/app: src/main.c
    mkdir -p {target.dir}
    gcc -o {target} {deps}
    echo "Output file: {target.file}"
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-plan"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugPlanStemVariable(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test stem variable with pattern targets
	content := `# Buildfile with stem variable
build/{name}.o: src/{name}.c
    echo "Stem: {name}"
    gcc -c {in} -o {out}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-plan"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugPlanRawModifier(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test raw modifier for unquoted expansion
	content := `# Buildfile with raw modifier
FLAGS = -Wall -O2

build/app: src/main.c
    gcc {FLAGS:raw} -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-plan"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugPlanBlockCommandInterpolation(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test block command interpolation
	content := `# Buildfile with block command
build/app: src/main.c
    block:
        if [ -f {target} ]; then
            rm {target}
        fi
        gcc -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-plan"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugPlanDepsVariable(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test deps variable with multiple dependencies
	content := `# Buildfile with multiple deps
build/app: src/main.c src/utils.c src/helper.c
    gcc -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-plan"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugPlanMixedVariables(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test mixed user-defined, automatic, and builtin variables
	content := `# Buildfile with mixed variables
cc = gcc
build_dir = build

{build_dir}/app: src/main.c
    echo "Building for {os}-{arch}"
    {cc} -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-plan"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// ===========================================================================
// Shell Execution Tests (Phase 4.2)
// ===========================================================================

func TestRunDebugPlanWithShellDirective(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test global shell directive
	content := `# Buildfile with shell directive
.shell: bash

build/app: src/main.c
    gcc -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-plan"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugPlanWithRecipeShellOverride(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test recipe-level shell override
	content := `# Buildfile with recipe shell override
.shell: sh

build/app: src/main.c
    .shell: bash
    gcc -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-plan"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// ===========================================================================
// Parallel Execution Tests (Phase 5)
// ===========================================================================

func TestRunDebugPlanWithParallelDirective(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test parallel directive
	content := `# Buildfile with parallel directive
.parallel: 4

build/a.o: src/a.c
    gcc -c {in} -o {out}

build/b.o: src/b.c
    gcc -c {in} -o {out}

build/app: build/a.o build/b.o
    gcc -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-plan"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunDebugPlanDiamondDependency(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Test diamond dependency for parallel scheduling
	content := `# Buildfile with diamond dependency
build/app: build/a.o build/b.o
    gcc -o {target} {deps}

build/a.o: build/common.h src/a.c
    gcc -c src/a.c -o {target}

build/b.o: build/common.h src/b.c
    gcc -c src/b.c -o {target}

build/common.h: src/common.h
    cp {in} {out}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--debug-plan"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// ----------------------------------------------------------------------------
// Environment Command Tests
// ----------------------------------------------------------------------------

func TestRunListEnv(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment:
    .requires: ls@latest

.environment: ci
    .using: docker
    .source: ./Dockerfile

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--list-env"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunListEnvNoEnvs(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `# No environments
cc = gcc

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--list-env"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunCheckEnvDefaultSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment:
    .requires: ls@latest sh@latest

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--check-env"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunCheckEnvDefaultMissing(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment:
    .requires: nonexistent-binary-xyz

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--check-env"})
	if exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d", exitCode, exitEnvError)
	}
}

func TestRunCheckEnvNamed(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment:
    .requires: ls@latest

.environment: ci
    .using: docker
    .source: ./Dockerfile

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Create the Dockerfile that the ci environment references
	dockerfile := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfile, []byte("FROM alpine\n"), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--check-env", "-e", "ci"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunCheckEnvNamedNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment:
    .requires: ls@latest

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--check-env", "-e", "nonexistent"})
	if exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d", exitCode, exitEnvError)
	}
}

func TestRunCheckEnvNoDefaultWithNamedOnly(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Only named environments, no default
	content := `.environment: ci
    .using: docker
    .source: ./Dockerfile

.environment: dev
    .using: nix
    .source: ./shell.nix

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Should error when trying to check without specifying --env
	exitCode := Run([]string{"-f", buildfile, "--check-env"})
	if exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d", exitCode, exitEnvError)
	}
}

func TestRunCheckEnvNoEnvironments(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// No environments at all
	content := `# No environments
cc = gcc

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Should succeed - bare environment with no requirements
	exitCode := Run([]string{"-f", buildfile, "--check-env"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

func TestShowInstall_MissingBinary(t *testing.T) {
	// Create a buildfile with a non-existent binary requirement
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment:
    .using: bare
    .requires: this-binary-does-not-exist-xyz123
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Should fail because binary doesn't exist
	exitCode := Run([]string{"-f", buildfile, "--check-env", "--show-install"})
	if exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d", exitCode, exitEnvError)
	}
}

func TestShowInstall_AllPresent(t *testing.T) {
	// Create a buildfile with binaries that should exist
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment:
    .using: bare
    .requires: sh ls
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Should succeed - both sh and ls should exist
	exitCode := Run([]string{"-f", buildfile, "--check-env", "--show-install"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// ===========================================================================
// Devcontainer Environment Tests
// ===========================================================================

func TestRunCheckEnvDevcontainer_WithConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a devcontainer.json file
	devcontainerDir := filepath.Join(tmpDir, ".devcontainer")
	if err := os.MkdirAll(devcontainerDir, 0755); err != nil {
		t.Fatal(err)
	}
	devcontainerConfig := filepath.Join(devcontainerDir, "devcontainer.json")
	configContent := `{
		"name": "Test Container",
		"image": "ubuntu:latest"
	}`
	if err := os.WriteFile(devcontainerConfig, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create buildfile with devcontainer environment
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment: dev
    .using: devcontainer

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Should succeed - devcontainer config exists (even if CLI is not installed, config is valid)
	exitCode := Run([]string{"-f", buildfile, "--check-env", "-e", "dev"})
	// The exit code depends on whether devcontainer CLI is installed
	// If it's not, we still expect the check to be informative
	if exitCode != exitSuccess && exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d or %d", exitCode, exitSuccess, exitEnvError)
	}
}

func TestRunCheckEnvDevcontainer_WithSourcePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a custom devcontainer.json file at a specified path
	customConfig := filepath.Join(tmpDir, "custom-devcontainer.json")
	configContent := `{
		"name": "Custom Container",
		"image": "alpine:latest"
	}`
	if err := os.WriteFile(customConfig, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create buildfile with devcontainer environment specifying source path
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment: dev
    .using: devcontainer
    .source: custom-devcontainer.json

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--check-env", "-e", "dev"})
	// The exit code depends on whether devcontainer CLI is installed
	if exitCode != exitSuccess && exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d or %d", exitCode, exitSuccess, exitEnvError)
	}
}

func TestRunCheckEnvDevcontainer_NoConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create buildfile with devcontainer environment but no config
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment: dev
    .using: devcontainer

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Should fail - no devcontainer configuration found
	exitCode := Run([]string{"-f", buildfile, "--check-env", "-e", "dev"})
	if exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d", exitCode, exitEnvError)
	}
}

func TestRunCheckEnvDevcontainer_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an invalid devcontainer.json file
	devcontainerDir := filepath.Join(tmpDir, ".devcontainer")
	if err := os.MkdirAll(devcontainerDir, 0755); err != nil {
		t.Fatal(err)
	}
	devcontainerConfig := filepath.Join(devcontainerDir, "devcontainer.json")
	invalidContent := `{not valid json}`
	if err := os.WriteFile(devcontainerConfig, []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create buildfile with devcontainer environment
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment: dev
    .using: devcontainer

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Should fail - invalid devcontainer config
	exitCode := Run([]string{"-f", buildfile, "--check-env", "-e", "dev"})
	if exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d", exitCode, exitEnvError)
	}
}

func TestRunListEnv_WithDevcontainer(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment:
    .requires: ls@latest

.environment: devenv
    .using: devcontainer

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--list-env"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// ===========================================================================
// Nix Environment Tests
// ===========================================================================

func TestRunCheckEnvNix_WithShellNix(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a shell.nix file
	shellNixPath := filepath.Join(tmpDir, "shell.nix")
	nixContent := `{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = [ pkgs.gcc ];
}
`
	if err := os.WriteFile(shellNixPath, []byte(nixContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create buildfile with nix environment
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment: nixdev
    .using: nix

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Should succeed - nix config exists (even if nix-shell CLI is not installed)
	exitCode := Run([]string{"-f", buildfile, "--check-env", "-e", "nixdev"})
	// The exit code depends on whether nix-shell is installed
	if exitCode != exitSuccess && exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d or %d", exitCode, exitSuccess, exitEnvError)
	}
}

func TestRunCheckEnvNix_WithFlakeNix(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a flake.nix file
	flakeNixPath := filepath.Join(tmpDir, "flake.nix")
	flakeContent := `{
  description = "Test flake";
  outputs = { self, nixpkgs }: {};
}
`
	if err := os.WriteFile(flakeNixPath, []byte(flakeContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create buildfile with nix environment
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment: nixdev
    .using: nix

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Should succeed - nix config exists
	exitCode := Run([]string{"-f", buildfile, "--check-env", "-e", "nixdev"})
	if exitCode != exitSuccess && exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d or %d", exitCode, exitSuccess, exitEnvError)
	}
}

func TestRunCheckEnvNix_WithSourcePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a custom nix file
	customNixPath := filepath.Join(tmpDir, "dev.nix")
	nixContent := `{ pkgs ? import <nixpkgs> {} }: pkgs.mkShell {}`
	if err := os.WriteFile(customNixPath, []byte(nixContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create buildfile with nix environment specifying source path
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment: nixdev
    .using: nix
    .source: dev.nix

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--check-env", "-e", "nixdev"})
	if exitCode != exitSuccess && exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d or %d", exitCode, exitSuccess, exitEnvError)
	}
}

func TestRunCheckEnvNix_NoConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create buildfile with nix environment but no config
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment: nixdev
    .using: nix

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Should fail - no nix configuration found
	exitCode := Run([]string{"-f", buildfile, "--check-env", "-e", "nixdev"})
	if exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d", exitCode, exitEnvError)
	}
}

func TestRunCheckEnvNix_WithArgs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a shell.nix file
	shellNixPath := filepath.Join(tmpDir, "shell.nix")
	nixContent := `{ pkgs ? import <nixpkgs> {} }: pkgs.mkShell {}`
	if err := os.WriteFile(shellNixPath, []byte(nixContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create buildfile with nix environment with args
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment: nixdev
    .using: nix
    .args: --pure

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--check-env", "-e", "nixdev"})
	if exitCode != exitSuccess && exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d or %d", exitCode, exitSuccess, exitEnvError)
	}
}

func TestRunListEnv_WithNix(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment:
    .requires: ls@latest

.environment: nixdev
    .using: nix

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--list-env"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// ===========================================================================
// Lima Environment Tests
// ===========================================================================

func TestRunCheckEnvLima_WithLimaYaml(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a lima.yaml file
	limaYamlPath := filepath.Join(tmpDir, "lima.yaml")
	limaContent := `images:
  - location: "https://example.com/ubuntu.qcow2"
mounts:
  - location: "~"
    writable: true
`
	if err := os.WriteFile(limaYamlPath, []byte(limaContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create buildfile with lima environment
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment: vm
    .using: lima

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Should succeed - lima config exists (even if limactl is not installed)
	exitCode := Run([]string{"-f", buildfile, "--check-env", "-e", "vm"})
	// The exit code depends on whether limactl is installed
	if exitCode != exitSuccess && exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d or %d", exitCode, exitSuccess, exitEnvError)
	}
}

func TestRunCheckEnvLima_WithSourcePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a custom lima file
	customLimaPath := filepath.Join(tmpDir, "dev-vm.yaml")
	limaContent := `images:
  - location: "https://example.com/arch.qcow2"
`
	if err := os.WriteFile(customLimaPath, []byte(limaContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create buildfile with lima environment specifying source path
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment: vm
    .using: lima
    .source: dev-vm.yaml

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--check-env", "-e", "vm"})
	if exitCode != exitSuccess && exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d or %d", exitCode, exitSuccess, exitEnvError)
	}
}

func TestRunCheckEnvLima_NoConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create buildfile with lima environment but no config
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment: vm
    .using: lima

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Should fail - no lima configuration found
	exitCode := Run([]string{"-f", buildfile, "--check-env", "-e", "vm"})
	if exitCode != exitEnvError {
		t.Errorf("exit code = %d, want %d", exitCode, exitEnvError)
	}
}

func TestRunListEnv_WithLima(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.environment:
    .requires: ls@latest

.environment: vm
    .using: lima

@all: build/app
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	exitCode := Run([]string{"-f", buildfile, "--list-env"})
	if exitCode != exitSuccess {
		t.Errorf("exit code = %d, want %d", exitCode, exitSuccess)
	}
}

// ===========================================================================
// Formatted Error Display Tests (Phase 8 - Error Categories)
// ===========================================================================

// captureStderr captures stderr during function execution
func captureStderr(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestParseErrorShowsErrorCode_SyntaxError(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// E103: Invalid directive at scope
	content := `.after: invalid
cc = gcc
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var exitCode int
	stderr := captureStderr(func() {
		exitCode = Run([]string{"-f", buildfile})
	})

	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
	// Should show error code E103 (invalid directive scope)
	if !strings.Contains(stderr, "error[E103]") {
		t.Errorf("Error output should contain error code 'error[E103]', got: %q", stderr)
	}
}

func TestParseErrorShowsSourceContext(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.after: invalid
cc = gcc
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var exitCode int
	stderr := captureStderr(func() {
		exitCode = Run([]string{"-f", buildfile})
	})

	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
	// Should show source context with line numbers
	if !strings.Contains(stderr, " | ") {
		t.Errorf("Error output should contain source context with ' | ', got: %q", stderr)
	}
}

func TestParseErrorShowsLocation(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	content := `.after: invalid
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var exitCode int
	stderr := captureStderr(func() {
		exitCode = Run([]string{"-f", buildfile})
	})

	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
	// Should show location arrow
	if !strings.Contains(stderr, "-->") {
		t.Errorf("Error output should contain location arrow '-->', got: %q", stderr)
	}
}

func TestSemanticErrorShowsErrorCode_DuplicateVariable(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// E201: Duplicate variable definition
	content := `.shell: bash
cc = gcc
cc = clang

@test:
    echo "test"
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var exitCode int
	stderr := captureStderr(func() {
		exitCode = Run([]string{"-f", buildfile})
	})

	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
	// Should show error code E201 (duplicate variable)
	if !strings.Contains(stderr, "error[E201]") {
		t.Errorf("Error output should contain error code 'error[E201]', got: %q", stderr)
	}
}

func TestSemanticErrorShowsErrorCode_UndefinedVariable(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// E200: Undefined variable
	content := `build/app: src/main.c
    {cc} -o {target} {deps}
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var exitCode int
	stderr := captureStderr(func() {
		exitCode = Run([]string{"-f", buildfile})
	})

	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
	// Should show error code E200 (undefined variable)
	if !strings.Contains(stderr, "error[E200]") {
		t.Errorf("Error output should contain error code 'error[E200]', got: %q", stderr)
	}
}

func TestSemanticErrorShowsErrorCode_AutomaticOutsideRecipe(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// E207: Automatic variable outside recipe
	content := `.shell: bash
output = {target}

@test:
    echo "test"
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var exitCode int
	stderr := captureStderr(func() {
		exitCode = Run([]string{"-f", buildfile})
	})

	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
	// Should show error code E207 (automatic outside recipe)
	if !strings.Contains(stderr, "error[E207]") {
		t.Errorf("Error output should contain error code 'error[E207]', got: %q", stderr)
	}
}

func TestSemanticErrorShowsErrorCode_CircularDependency(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// E204: Circular dependency
	content := `dir/a.o: dir/b.o
    echo a

dir/b.o: dir/a.o
    echo b
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var exitCode int
	stderr := captureStderr(func() {
		exitCode = Run([]string{"-f", buildfile})
	})

	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
	// Should show error code E204 (circular dependency)
	if !strings.Contains(stderr, "error[E204]") {
		t.Errorf("Error output should contain error code 'error[E204]', got: %q", stderr)
	}
}

func TestMultipleErrorsShowAllCodes(t *testing.T) {
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	// Multiple errors: first syntax error, then will not continue
	content := `.after: invalid
.using: invalid
cc = gcc
`
	if err := os.WriteFile(buildfile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var exitCode int
	stderr := captureStderr(func() {
		exitCode = Run([]string{"-f", buildfile})
	})

	if exitCode != exitParseError {
		t.Errorf("exit code = %d, want %d", exitCode, exitParseError)
	}
	// Should show multiple errors
	errorCount := strings.Count(stderr, "error[")
	if errorCount < 2 {
		t.Errorf("Should show at least 2 errors, got %d errors in: %q", errorCount, stderr)
	}
}
