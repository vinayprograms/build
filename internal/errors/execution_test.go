package errors

import (
	"strings"
	"testing"

	"github.com/vinayprograms/need/internal/ast"
)

// ----------------------------------------------------------------------------
// Execution Error Tests (E400-E499)
// ----------------------------------------------------------------------------

func TestNewRecipeFailedError(t *testing.T) {
	loc := ast.SourceLocation{File: "Needfile", Line: 5, Column: 5}
	err := NewRecipeFailedError("build/app", "gcc -o build/app main.c", 1, loc)

	if err.Code != "E400" {
		t.Errorf("Expected code E400, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "recipe failed") {
		t.Errorf("Message should mention recipe failed: %s", err.Message)
	}
	if !strings.Contains(err.Message, "build/app") {
		t.Errorf("Message should mention target: %s", err.Message)
	}
}

func TestNewMissingDependencyError(t *testing.T) {
	err := NewMissingDependencyError("build/main.o", "build/app")

	if err.Code != "E401" {
		t.Errorf("Expected code E401, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "missing dependency") {
		t.Errorf("Message should mention missing dependency: %s", err.Message)
	}
	if !strings.Contains(err.Message, "build/main.o") {
		t.Errorf("Message should mention the dependency: %s", err.Message)
	}
}

func TestNewMissingBinaryError(t *testing.T) {
	loc := ast.SourceLocation{File: "Needfile", Line: 2, Column: 15}
	err := NewMissingBinaryError("gcc", loc)

	if err.Code != "E402" {
		t.Errorf("Expected code E402, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "missing required binary") {
		t.Errorf("Message should mention missing binary: %s", err.Message)
	}
	if !strings.Contains(err.Message, "gcc") {
		t.Errorf("Message should mention binary name: %s", err.Message)
	}
	if err.Help == "" {
		t.Error("Missing binary error should have help text")
	}
}

func TestNewShellNotFoundError(t *testing.T) {
	loc := ast.SourceLocation{File: "Needfile", Line: 1, Column: 8}
	err := NewShellNotFoundError("/bin/zsh", loc)

	if err.Code != "E403" {
		t.Errorf("Expected code E403, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "shell not found") {
		t.Errorf("Message should mention shell not found: %s", err.Message)
	}
	if !strings.Contains(err.Message, "/bin/zsh") {
		t.Errorf("Message should mention shell path: %s", err.Message)
	}
}

func TestNewVersionMismatchError(t *testing.T) {
	loc := ast.SourceLocation{File: "Needfile", Line: 3, Column: 15}
	err := NewVersionMismatchError("gcc", "11.0.0", "10.2.0", loc)

	if err.Code != "E404" {
		t.Errorf("Expected code E404, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "version mismatch") {
		t.Errorf("Message should mention version mismatch: %s", err.Message)
	}
	if !strings.Contains(err.Message, "11.0.0") {
		t.Errorf("Message should mention required version: %s", err.Message)
	}
	if !strings.Contains(err.Message, "10.2.0") {
		t.Errorf("Message should mention found version: %s", err.Message)
	}
}

func TestNewTargetNotFoundError(t *testing.T) {
	err := NewTargetNotFoundError("clean")

	if err.Code != "E405" {
		t.Errorf("Expected code E405, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "target not found") {
		t.Errorf("Message should mention target not found: %s", err.Message)
	}
	if !strings.Contains(err.Message, "clean") {
		t.Errorf("Message should mention target name: %s", err.Message)
	}
}

func TestNewNoDefaultTargetError(t *testing.T) {
	err := NewNoDefaultTargetError()

	if err.Code != "E406" {
		t.Errorf("Expected code E406, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "no default target") {
		t.Errorf("Message should mention no default target: %s", err.Message)
	}
	if err.Help == "" {
		t.Error("No default target error should have help text")
	}
}

// ----------------------------------------------------------------------------
// Error Code Range Tests
// ----------------------------------------------------------------------------

func TestExecutionErrorCodes_InRange(t *testing.T) {
	testCases := []struct {
		name string
		err  *FormattedError
	}{
		{"RecipeFailed", NewRecipeFailedError("target", "cmd", 1, ast.SourceLocation{})},
		{"MissingDependency", NewMissingDependencyError("dep", "target")},
		{"MissingBinary", NewMissingBinaryError("binary", ast.SourceLocation{})},
		{"ShellNotFound", NewShellNotFoundError("/bin/sh", ast.SourceLocation{})},
		{"VersionMismatch", NewVersionMismatchError("bin", "1.0", "0.9", ast.SourceLocation{})},
		{"TargetNotFound", NewTargetNotFoundError("target")},
		{"NoDefaultTarget", NewNoDefaultTargetError()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := tc.err.Code
			if len(code) != 4 || code[0] != 'E' {
				t.Errorf("Error code should be E### format, got: %s", code)
			}

			num := 0
			for i := 1; i < len(code); i++ {
				if code[i] < '0' || code[i] > '9' {
					t.Errorf("Error code should be numeric after E: %s", code)
					return
				}
				num = num*10 + int(code[i]-'0')
			}

			if num < 400 || num > 499 {
				t.Errorf("Execution error code should be E400-E499, got: %s (num=%d)", code, num)
			}
		})
	}
}
