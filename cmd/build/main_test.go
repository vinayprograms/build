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
	}

	for _, s := range essentials {
		if !strings.Contains(output, s) {
			t.Errorf("usage output missing %q", s)
		}
	}
}
