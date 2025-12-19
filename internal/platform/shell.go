package platform

import (
	"os/exec"
	"runtime"
	"strings"
)

// IsWindows returns true if running on Windows.
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// DefaultShell returns the default shell for the current platform.
// On Windows, returns cmd.exe. On Unix systems, returns /bin/sh.
func DefaultShell() string {
	if IsWindows() {
		return "cmd.exe"
	}
	return "/bin/sh"
}

// ShellCommandArgs returns the command-line arguments to pass a command to a shell.
// For Unix shells (sh, bash, zsh), this returns ["-c", command].
// For cmd.exe, this returns ["/C", command].
// For PowerShell, this returns ["-Command", command].
func ShellCommandArgs(shell, command string) []string {
	shellBase := baseName(shell)
	shellLower := strings.ToLower(shellBase)

	// Remove .exe extension for comparison
	shellName := strings.TrimSuffix(shellLower, ".exe")

	switch shellName {
	case "cmd":
		return []string{"/C", command}
	case "powershell", "pwsh":
		return []string{"-Command", command}
	default:
		// Unix-style shell
		return []string{"-c", command}
	}
}

// IsWindowsShell returns true if the shell is a Windows shell (cmd.exe or PowerShell).
func IsWindowsShell(shell string) bool {
	shellBase := baseName(shell)
	shellLower := strings.ToLower(shellBase)
	shellName := strings.TrimSuffix(shellLower, ".exe")

	return shellName == "cmd" || shellName == "powershell" || shellName == "pwsh"
}

// baseName returns the last element of path, handling both / and \ as separators.
// This is needed because filepath.Base only treats the native separator as a separator
// on each platform.
func baseName(path string) string {
	// Try forward slash first (works on all platforms)
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	// Then try backslash (for Windows paths on any platform)
	if i := strings.LastIndex(path, "\\"); i >= 0 {
		return path[i+1:]
	}
	return path
}

// IsAbsolutePath returns true if the path is an absolute path.
// This handles both Unix-style paths (/foo) and Windows-style paths (C:\foo, \\server\share).
func IsAbsolutePath(path string) bool {
	if len(path) == 0 {
		return false
	}

	// Unix-style absolute path
	if path[0] == '/' {
		return true
	}

	// Windows UNC path (\\server\share)
	if len(path) >= 2 && path[0] == '\\' && path[1] == '\\' {
		return true
	}

	// Windows drive letter (C:\, D:/, etc.)
	if len(path) >= 2 && isLetter(path[0]) && path[1] == ':' {
		// Must be followed by separator or be at least 3 chars with separator
		if len(path) >= 3 && (path[2] == '\\' || path[2] == '/') {
			return true
		}
		// Just "C:" is also considered absolute on Windows
		if len(path) == 2 {
			return true
		}
	}

	return false
}

// isLetter returns true if c is an ASCII letter.
func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// IsDirectoryPath returns true if the path ends with a directory separator.
func IsDirectoryPath(path string) bool {
	if len(path) == 0 {
		return false
	}
	return path[len(path)-1] == '/' || path[len(path)-1] == '\\'
}

// PathSeparator returns the OS-specific path separator.
func PathSeparator() byte {
	if IsWindows() {
		return '\\'
	}
	return '/'
}

// NormalizePath converts all backslashes to forward slashes for consistency.
// This is useful for internal path representation.
func NormalizePath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// ----------------------------------------------------------------------------
// Shell Quoting
// ----------------------------------------------------------------------------

// cmdSpecialChars are characters that require quoting in cmd.exe.
const cmdSpecialChars = " \t&|<>^()%!"

// psSpecialChars are characters that require quoting in PowerShell.
const psSpecialChars = " \t'\"$`"

// unixSpecialChars are characters that require quoting in Unix shells.
const unixSpecialChars = " \t\n'\"\\`$!*?[]{}();&|<>#~"

// ShellQuote quotes a string for safe use in shell commands.
// The quoting style depends on the shell:
//   - Unix shells: single quotes with escaped embedded single quotes
//   - cmd.exe: double quotes
//   - PowerShell: single quotes
func ShellQuote(shell, value string) string {
	shellBase := baseName(shell)
	shellLower := strings.ToLower(shellBase)
	shellName := strings.TrimSuffix(shellLower, ".exe")

	switch shellName {
	case "cmd":
		return cmdQuote(value)
	case "powershell", "pwsh":
		return psQuote(value)
	default:
		return unixQuote(value)
	}
}

// cmdQuote quotes a string for cmd.exe.
// Uses double quotes if the string contains special characters.
func cmdQuote(s string) string {
	if !strings.ContainsAny(s, cmdSpecialChars) {
		return s
	}
	// cmd.exe uses double quotes
	// Double quotes inside need to be escaped with backslash (not standard, but works)
	escaped := strings.ReplaceAll(s, `"`, `\"`)
	return `"` + escaped + `"`
}

// psQuote quotes a string for PowerShell.
// Uses single quotes (like Unix) for simplicity.
func psQuote(s string) string {
	if !strings.ContainsAny(s, psSpecialChars) {
		return s
	}
	// PowerShell single quotes: escape embedded single quotes by doubling
	escaped := strings.ReplaceAll(s, "'", "''")
	return "'" + escaped + "'"
}

// unixQuote quotes a string for Unix shells.
// Uses single quotes with special handling for embedded single quotes.
func unixQuote(s string) string {
	if !strings.ContainsAny(s, unixSpecialChars) {
		return s
	}
	// If string contains single quotes, we need special handling
	if strings.Contains(s, "'") {
		// Replace each ' with '"'"' (end quote, double-quoted quote, start quote)
		escaped := strings.ReplaceAll(s, "'", `'"'"'`)
		return "'" + escaped + "'"
	}
	return "'" + s + "'"
}

// ----------------------------------------------------------------------------
// Shell Validation
// ----------------------------------------------------------------------------

// ValidateShell checks that the shell exists and is executable.
// Returns an error if the shell cannot be found.
func ValidateShell(shell string) error {
	// If it's an absolute path, check directly
	if IsAbsolutePath(shell) {
		_, err := exec.LookPath(shell)
		return err
	}
	// Otherwise, look up in PATH
	_, err := exec.LookPath(shell)
	return err
}
