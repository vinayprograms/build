package errors

import (
	"strings"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

// ----------------------------------------------------------------------------
// Evaluation Error Tests (E300-E399)
// ----------------------------------------------------------------------------

func TestNewShellCommandFailedError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 3, Column: 10}
	err := NewShellCommandFailedError("ls /nonexistent", 1, "No such file or directory", loc)

	if err.Code != "E300" {
		t.Errorf("Expected code E300, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "shell command failed") {
		t.Errorf("Message should mention shell command failed: %s", err.Message)
	}
	if !strings.Contains(err.Note, "exit code") || !strings.Contains(err.Note, "1") {
		t.Errorf("Note should mention exit code: %s", err.Note)
	}
}

func TestNewGlobNoMatchError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 5, Column: 15}
	err := NewGlobNoMatchError("*.xyz", loc)

	if err.Code != "E301" {
		t.Errorf("Expected code E301, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "glob") {
		t.Errorf("Message should mention glob: %s", err.Message)
	}
	if !strings.Contains(err.Message, "*.xyz") {
		t.Errorf("Message should mention pattern: %s", err.Message)
	}
}

func TestNewInvalidFunctionArgumentsError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 2, Column: 20}
	err := NewInvalidFunctionArgumentsError("basename", "path cannot be empty", loc)

	if err.Code != "E302" {
		t.Errorf("Expected code E302, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "invalid argument") {
		t.Errorf("Message should mention invalid argument: %s", err.Message)
	}
	if !strings.Contains(err.Message, "basename") {
		t.Errorf("Message should mention function name: %s", err.Message)
	}
}

func TestNewForwardReferenceError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 1, Column: 10}
	err := NewForwardReferenceError("foo", "bar", loc)

	if err.Code != "E303" {
		t.Errorf("Expected code E303, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "forward reference") {
		t.Errorf("Message should mention forward reference: %s", err.Message)
	}
	if !strings.Contains(err.Message, "foo") {
		t.Errorf("Message should mention variable name: %s", err.Message)
	}
}

func TestNewLazyEvaluationError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 4, Column: 1}
	err := NewLazyEvaluationError("config", "undefined variable 'x'", loc)

	if err.Code != "E304" {
		t.Errorf("Expected code E304, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "lazy variable") {
		t.Errorf("Message should mention lazy variable: %s", err.Message)
	}
	if !strings.Contains(err.Message, "config") {
		t.Errorf("Message should mention variable name: %s", err.Message)
	}
}

func TestNewConditionEvaluationError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 6, Column: 4}
	err := NewConditionEvaluationError("undefined variable 'mode'", loc)

	if err.Code != "E305" {
		t.Errorf("Expected code E305, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "condition") {
		t.Errorf("Message should mention condition: %s", err.Message)
	}
}

// ----------------------------------------------------------------------------
// Error Code Range Tests
// ----------------------------------------------------------------------------

func TestEvaluationErrorCodes_InRange(t *testing.T) {
	testCases := []struct {
		name string
		err  *FormattedError
	}{
		{"ShellCommandFailed", NewShellCommandFailedError("cmd", 1, "error", ast.SourceLocation{})},
		{"GlobNoMatch", NewGlobNoMatchError("*.txt", ast.SourceLocation{})},
		{"InvalidFunctionArgs", NewInvalidFunctionArgumentsError("func", "reason", ast.SourceLocation{})},
		{"ForwardReference", NewForwardReferenceError("a", "b", ast.SourceLocation{})},
		{"LazyEvaluation", NewLazyEvaluationError("var", "reason", ast.SourceLocation{})},
		{"ConditionEvaluation", NewConditionEvaluationError("reason", ast.SourceLocation{})},
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

			if num < 300 || num > 399 {
				t.Errorf("Evaluation error code should be E300-E399, got: %s (num=%d)", code, num)
			}
		})
	}
}
