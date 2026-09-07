package errors

import (
	"strings"
	"testing"

	"github.com/vinayprograms/need/internal/ast"
)

// ----------------------------------------------------------------------------
// Semantic Error Tests (E200-E299)
// ----------------------------------------------------------------------------

func TestNewUndefinedVariableError(t *testing.T) {
	loc := ast.SourceLocation{File: "Needfile", Line: 5, Column: 10}
	err := NewUndefinedVariableError("foo", loc)

	if err.Code != "E200" {
		t.Errorf("Expected code E200, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "undefined variable") {
		t.Errorf("Message should mention undefined variable: %s", err.Message)
	}
	if !strings.Contains(err.Message, "foo") {
		t.Errorf("Message should mention the variable name: %s", err.Message)
	}
}

func TestNewDuplicateVariableError(t *testing.T) {
	first := ast.SourceLocation{File: "Needfile", Line: 1, Column: 1}
	second := ast.SourceLocation{File: "Needfile", Line: 5, Column: 1}
	err := NewDuplicateVariableError("cc", first, second)

	if err.Code != "E201" {
		t.Errorf("Expected code E201, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "duplicate") {
		t.Errorf("Message should mention duplicate: %s", err.Message)
	}
	if !strings.Contains(err.Message, "cc") {
		t.Errorf("Message should mention variable name: %s", err.Message)
	}
	if err.Note == "" {
		t.Error("Duplicate variable error should have note about first definition")
	}
}

func TestNewDuplicateTargetError(t *testing.T) {
	first := ast.SourceLocation{File: "Needfile", Line: 3, Column: 1}
	second := ast.SourceLocation{File: "Needfile", Line: 10, Column: 1}
	err := NewDuplicateTargetError("build/app", first, second)

	if err.Code != "E202" {
		t.Errorf("Expected code E202, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "duplicate") {
		t.Errorf("Message should mention duplicate: %s", err.Message)
	}
	if !strings.Contains(err.Message, "build/app") {
		t.Errorf("Message should mention target name: %s", err.Message)
	}
}

func TestNewDuplicateEnvironmentError(t *testing.T) {
	first := ast.SourceLocation{File: "Needfile", Line: 2, Column: 1}
	second := ast.SourceLocation{File: "Needfile", Line: 8, Column: 1}
	err := NewDuplicateEnvironmentError("ci", first, second)

	if err.Code != "E203" {
		t.Errorf("Expected code E203, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "duplicate") {
		t.Errorf("Message should mention duplicate: %s", err.Message)
	}
	if !strings.Contains(err.Message, "ci") {
		t.Errorf("Message should mention environment name: %s", err.Message)
	}
}

func TestNewCircularDependencyError(t *testing.T) {
	cycle := []string{"a", "b", "c", "a"}
	err := NewCircularDependencyError(cycle)

	if err.Code != "E204" {
		t.Errorf("Expected code E204, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "circular dependency") {
		t.Errorf("Message should mention circular dependency: %s", err.Message)
	}
	if !strings.Contains(err.Message, "a -> b -> c -> a") {
		t.Errorf("Message should show cycle path: %s", err.Message)
	}
}

func TestNewCaptureConflictError(t *testing.T) {
	loc := ast.SourceLocation{File: "Needfile", Line: 5, Column: 10}
	err := NewCaptureConflictError("name", "variable", loc)

	if err.Code != "E205" {
		t.Errorf("Expected code E205, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "capture") && !strings.Contains(err.Message, "conflicts") {
		t.Errorf("Message should mention capture conflict: %s", err.Message)
	}
}

func TestNewCaptureMismatchError(t *testing.T) {
	loc := ast.SourceLocation{File: "Needfile", Line: 6, Column: 20}
	targetLoc := ast.SourceLocation{File: "Needfile", Line: 6, Column: 1}
	err := NewCaptureMismatchError("name", loc, targetLoc)

	if err.Code != "E206" {
		t.Errorf("Expected code E206, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "capture") {
		t.Errorf("Message should mention capture: %s", err.Message)
	}
	if !strings.Contains(err.Message, "name") {
		t.Errorf("Message should mention capture name: %s", err.Message)
	}
}

func TestNewAutomaticOutsideRecipeError(t *testing.T) {
	loc := ast.SourceLocation{File: "Needfile", Line: 2, Column: 10}
	err := NewAutomaticOutsideRecipeError("target", loc)

	if err.Code != "E207" {
		t.Errorf("Expected code E207, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "automatic variable") {
		t.Errorf("Message should mention automatic variable: %s", err.Message)
	}
	if !strings.Contains(err.Message, "target") {
		t.Errorf("Message should mention variable name: %s", err.Message)
	}
}

func TestNewAutomaticInPatternError(t *testing.T) {
	loc := ast.SourceLocation{File: "Needfile", Line: 4, Column: 8}
	err := NewAutomaticInPatternError("deps", loc)

	if err.Code != "E208" {
		t.Errorf("Expected code E208, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "automatic variable") {
		t.Errorf("Message should mention automatic variable: %s", err.Message)
	}
	if !strings.Contains(err.Message, "pattern") {
		t.Errorf("Message should mention pattern: %s", err.Message)
	}
}

// ----------------------------------------------------------------------------
// Error Code Range Tests
// ----------------------------------------------------------------------------

func TestSemanticErrorCodes_InRange(t *testing.T) {
	testCases := []struct {
		name string
		err  *FormattedError
	}{
		{"UndefinedVariable", NewUndefinedVariableError("foo", ast.SourceLocation{})},
		{"DuplicateVariable", NewDuplicateVariableError("cc", ast.SourceLocation{}, ast.SourceLocation{})},
		{"DuplicateTarget", NewDuplicateTargetError("app", ast.SourceLocation{}, ast.SourceLocation{})},
		{"DuplicateEnvironment", NewDuplicateEnvironmentError("ci", ast.SourceLocation{}, ast.SourceLocation{})},
		{"CircularDependency", NewCircularDependencyError([]string{"a", "b", "a"})},
		{"CaptureConflict", NewCaptureConflictError("name", "variable", ast.SourceLocation{})},
		{"CaptureMismatch", NewCaptureMismatchError("name", ast.SourceLocation{}, ast.SourceLocation{})},
		{"AutomaticOutsideRecipe", NewAutomaticOutsideRecipeError("target", ast.SourceLocation{})},
		{"AutomaticInPattern", NewAutomaticInPatternError("deps", ast.SourceLocation{})},
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

			if num < 200 || num > 299 {
				t.Errorf("Semantic error code should be E200-E299, got: %s (num=%d)", code, num)
			}
		})
	}
}
