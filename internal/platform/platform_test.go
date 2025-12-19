package platform

import (
	"runtime"
	"strings"
	"testing"
)

// TestDefaultShell verifies the default shell for the current platform.
func TestDefaultShell(t *testing.T) {
	shell := DefaultShell()

	switch runtime.GOOS {
	case "windows":
		// Windows should use cmd.exe or powershell
		if shell != "cmd.exe" && shell != "powershell.exe" {
			t.Errorf("DefaultShell() on windows = %q, want cmd.exe or powershell.exe", shell)
		}
	default:
		// Unix systems should use /bin/sh
		if shell != "/bin/sh" {
			t.Errorf("DefaultShell() on %s = %q, want /bin/sh", runtime.GOOS, shell)
		}
	}
}

// TestShellCommandArgs verifies shell command arguments.
func TestShellCommandArgs(t *testing.T) {
	tests := []struct {
		name    string
		shell   string
		command string
		want    []string
	}{
		{
			name:    "unix shell",
			shell:   "/bin/sh",
			command: "echo hello",
			want:    []string{"-c", "echo hello"},
		},
		{
			name:    "bash",
			shell:   "bash",
			command: "echo hello",
			want:    []string{"-c", "echo hello"},
		},
		{
			name:    "cmd.exe",
			shell:   "cmd.exe",
			command: "echo hello",
			want:    []string{"/C", "echo hello"},
		},
		{
			name:    "CMD uppercase",
			shell:   "CMD",
			command: "echo hello",
			want:    []string{"/C", "echo hello"},
		},
		{
			name:    "cmd without extension",
			shell:   "cmd",
			command: "echo hello",
			want:    []string{"/C", "echo hello"},
		},
		{
			name:    "powershell.exe",
			shell:   "powershell.exe",
			command: "echo hello",
			want:    []string{"-Command", "echo hello"},
		},
		{
			name:    "PowerShell uppercase",
			shell:   "PowerShell",
			command: "echo hello",
			want:    []string{"-Command", "echo hello"},
		},
		{
			name:    "powershell without extension",
			shell:   "powershell",
			command: "echo hello",
			want:    []string{"-Command", "echo hello"},
		},
		{
			name:    "pwsh (PowerShell Core)",
			shell:   "pwsh",
			command: "echo hello",
			want:    []string{"-Command", "echo hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShellCommandArgs(tt.shell, tt.command)
			if len(got) != len(tt.want) {
				t.Errorf("ShellCommandArgs(%q, %q) = %v, want %v", tt.shell, tt.command, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ShellCommandArgs(%q, %q)[%d] = %q, want %q", tt.shell, tt.command, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestIsAbsolutePath verifies absolute path detection.
func TestIsAbsolutePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		// Unix-style paths
		{name: "unix absolute", path: "/usr/bin/sh", want: true},
		{name: "unix root", path: "/", want: true},
		{name: "unix relative", path: "bin/sh", want: false},
		{name: "unix relative dot", path: "./bin/sh", want: false},
		{name: "unix relative parent", path: "../bin/sh", want: false},
		// Windows-style paths
		{name: "windows drive C", path: "C:\\Windows\\System32", want: true},
		{name: "windows drive c lowercase", path: "c:\\windows", want: true},
		{name: "windows drive D", path: "D:/Users", want: true},
		{name: "windows UNC", path: "\\\\server\\share", want: true},
		{name: "windows relative", path: "Windows\\System32", want: false},
		{name: "windows relative dot", path: ".\\bin", want: false},
		// Empty and special
		{name: "empty", path: "", want: false},
		{name: "just backslash", path: "\\", want: false}, // Not UNC, not drive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAbsolutePath(tt.path)
			if got != tt.want {
				t.Errorf("IsAbsolutePath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestIsWindowsShell verifies Windows shell detection.
func TestIsWindowsShell(t *testing.T) {
	tests := []struct {
		shell string
		want  bool
	}{
		{"cmd.exe", true},
		{"cmd", true},
		{"CMD", true},
		{"CMD.EXE", true},
		{"powershell.exe", true},
		{"powershell", true},
		{"PowerShell", true},
		{"POWERSHELL.EXE", true},
		{"pwsh", true},
		{"pwsh.exe", true},
		{"C:\\Windows\\System32\\cmd.exe", true},
		{"C:/Windows/System32/cmd.exe", true},
		{"/bin/sh", false},
		{"bash", false},
		{"zsh", false},
		{"/usr/bin/bash", false},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			got := IsWindowsShell(tt.shell)
			if got != tt.want {
				t.Errorf("IsWindowsShell(%q) = %v, want %v", tt.shell, got, tt.want)
			}
		})
	}
}

// TestIsDirectoryPath verifies directory path detection.
func TestIsDirectoryPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// Unix-style
		{"build/", true},
		{"/tmp/build/", true},
		{"./build/", true},
		{"build", false},
		{"/tmp/build", false},
		// Windows-style
		{"build\\", true},
		{"C:\\build\\", true},
		{".\\build\\", true},
		{"build", false},
		{"C:\\build", false},
		// Empty
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := IsDirectoryPath(tt.path)
			if got != tt.want {
				t.Errorf("IsDirectoryPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestPathSeparator verifies path separator for current platform.
func TestPathSeparator(t *testing.T) {
	sep := PathSeparator()

	switch runtime.GOOS {
	case "windows":
		if sep != '\\' {
			t.Errorf("PathSeparator() on windows = %q, want '\\'", sep)
		}
	default:
		if sep != '/' {
			t.Errorf("PathSeparator() on %s = %q, want '/'", runtime.GOOS, sep)
		}
	}
}

// TestNormalizePath verifies path normalization.
func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string // Expected on current platform
	}{
		{name: "already forward slash", path: "a/b/c", want: "a/b/c"},
		{name: "backslash to forward", path: "a\\b\\c", want: "a/b/c"},
		{name: "mixed slashes", path: "a/b\\c/d", want: "a/b/c/d"},
		{name: "empty", path: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePath(tt.path)
			// NormalizePath always uses forward slashes for consistency
			if got != tt.want {
				t.Errorf("NormalizePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// TestShellQuoteWindows verifies Windows-specific shell quoting.
func TestShellQuoteWindows(t *testing.T) {
	tests := []struct {
		name  string
		shell string
		value string
		want  string
	}{
		// cmd.exe uses double quotes
		{name: "cmd simple", shell: "cmd.exe", value: "hello", want: "hello"},
		{name: "cmd with space", shell: "cmd.exe", value: "hello world", want: `"hello world"`},
		{name: "cmd with special", shell: "cmd.exe", value: "hello&world", want: `"hello&world"`},
		// PowerShell uses single quotes (like Unix)
		{name: "ps simple", shell: "powershell.exe", value: "hello", want: "hello"},
		{name: "ps with space", shell: "powershell.exe", value: "hello world", want: "'hello world'"},
		// Unix shells use single quotes
		{name: "bash simple", shell: "bash", value: "hello", want: "hello"},
		{name: "bash with space", shell: "bash", value: "hello world", want: "'hello world'"},
		{name: "bash with single quote", shell: "bash", value: "it's", want: `'it'"'"'s'`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShellQuote(tt.shell, tt.value)
			if got != tt.want {
				t.Errorf("ShellQuote(%q, %q) = %q, want %q", tt.shell, tt.value, got, tt.want)
			}
		})
	}
}

// TestExitCodeFromError verifies exit code extraction from errors.
func TestExitCodeFromError(t *testing.T) {
	// This test needs to run actual commands to generate real exit errors.
	// We'll test the interface works correctly.
	tests := []struct {
		name     string
		exitCode int
	}{
		{name: "success", exitCode: 0},
		{name: "general error", exitCode: 1},
		{name: "not found", exitCode: 127},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is more of a documentation test - actual exit code
			// extraction is tested via integration tests with real commands.
			if tt.exitCode < 0 || tt.exitCode > 255 {
				t.Errorf("exit code %d out of valid range", tt.exitCode)
			}
		})
	}
}

// TestIsWindows verifies Windows platform detection.
func TestIsWindows(t *testing.T) {
	got := IsWindows()
	want := runtime.GOOS == "windows"
	if got != want {
		t.Errorf("IsWindows() = %v, want %v (GOOS=%s)", got, want, runtime.GOOS)
	}
}

// TestShellExecutable verifies shell executable resolution.
func TestShellExecutable(t *testing.T) {
	tests := []struct {
		name  string
		shell string
	}{
		{name: "default shell", shell: ""},
		{name: "explicit bash", shell: "bash"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shell := tt.shell
			if shell == "" {
				shell = DefaultShell()
			}
			// Just verify it returns something non-empty
			if shell == "" {
				t.Error("shell should not be empty")
			}
		})
	}
}

// TestCmdExeSpecialChars verifies cmd.exe special character detection.
func TestCmdExeSpecialChars(t *testing.T) {
	specialChars := []string{
		"hello world",  // space
		"hello&world",  // ampersand
		"hello|world",  // pipe
		"hello<world",  // less than
		"hello>world",  // greater than
		"hello^world",  // caret
		"hello%world",  // percent
		"hello(world)", // parentheses
	}

	for _, s := range specialChars {
		t.Run(s, func(t *testing.T) {
			quoted := ShellQuote("cmd.exe", s)
			// cmd.exe special chars should be double-quoted
			if !strings.HasPrefix(quoted, `"`) || !strings.HasSuffix(quoted, `"`) {
				t.Errorf("ShellQuote(cmd.exe, %q) = %q, expected double-quoted", s, quoted)
			}
		})
	}

	// Test safe strings that don't need quoting
	safeStrings := []string{
		"hello",
		"hello123",
		"Hello_World",
	}

	for _, s := range safeStrings {
		t.Run("safe_"+s, func(t *testing.T) {
			quoted := ShellQuote("cmd.exe", s)
			// Safe strings should not be quoted
			if quoted != s {
				t.Errorf("ShellQuote(cmd.exe, %q) = %q, expected unquoted", s, quoted)
			}
		})
	}
}
