package errors

import (
	"strings"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

// ----------------------------------------------------------------------------
// Syntax Error Tests (E100-E199)
// ----------------------------------------------------------------------------

func TestNewUnexpectedTokenError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 5, Column: 10}
	err := NewUnexpectedTokenError("IDENTIFIER", ":", loc)

	if err.Code != "E100" {
		t.Errorf("Expected code E100, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "unexpected token") {
		t.Errorf("Message should mention unexpected token: %s", err.Message)
	}
	if !strings.Contains(err.Message, "IDENTIFIER") {
		t.Errorf("Message should mention the token type: %s", err.Message)
	}
}

func TestNewMissingColonError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 3, Column: 15}
	err := NewMissingColonError("build/app", loc)

	if err.Code != "E101" {
		t.Errorf("Expected code E101, got %s", err.Code)
	}
	if !strings.Contains(err.Message, ":") {
		t.Errorf("Message should mention missing colon: %s", err.Message)
	}
	if err.Help == "" {
		t.Error("Missing colon error should have help text")
	}
}

func TestNewMissingEndError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1}
	err := NewMissingEndError(loc)

	if err.Code != "E102" {
		t.Errorf("Expected code E102, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "end") {
		t.Errorf("Message should mention missing 'end': %s", err.Message)
	}
}

func TestNewInvalidDirectiveScopeError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 2, Column: 5}
	err := NewInvalidDirectiveScopeError(".after", "GLOBAL", []string{"RECIPE"}, loc)

	if err.Code != "E103" {
		t.Errorf("Expected code E103, got %s", err.Code)
	}
	if !strings.Contains(err.Message, ".after") {
		t.Errorf("Message should mention directive name: %s", err.Message)
	}
	if !strings.Contains(err.Message, "GLOBAL") {
		t.Errorf("Message should mention current scope: %s", err.Message)
	}
	if !strings.Contains(err.Help, "RECIPE") {
		t.Errorf("Help should list valid scopes: %s", err.Help)
	}
}

func TestNewMissingConditionError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 4, Column: 1}
	err := NewMissingConditionError("if", loc)

	if err.Code != "E104" {
		t.Errorf("Expected code E104, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "condition") {
		t.Errorf("Message should mention condition: %s", err.Message)
	}
}

func TestNewMissingOperatorError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 5, Column: 10}
	err := NewMissingOperatorError(loc)

	if err.Code != "E105" {
		t.Errorf("Expected code E105, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "==") || !strings.Contains(err.Message, "!=") {
		t.Errorf("Message should mention valid operators: %s", err.Message)
	}
}

func TestNewMissingIdentifierError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 3, Column: 7}
	err := NewMissingIdentifierError("ifdef", loc)

	if err.Code != "E106" {
		t.Errorf("Expected code E106, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "identifier") {
		t.Errorf("Message should mention identifier: %s", err.Message)
	}
}

func TestNewInvalidRuntimeError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 2, Column: 12}
	err := NewInvalidRuntimeError("invalid_runtime", loc)

	if err.Code != "E107" {
		t.Errorf("Expected code E107, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "invalid_runtime") {
		t.Errorf("Message should mention the invalid runtime: %s", err.Message)
	}
	if !strings.Contains(err.Help, "bare") {
		t.Errorf("Help should list valid runtimes: %s", err.Help)
	}
}

func TestNewMissingFunctionArgumentError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 6, Column: 20}
	err := NewMissingFunctionArgumentError("replace", 3, 2, loc)

	if err.Code != "E108" {
		t.Errorf("Expected code E108, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "replace") {
		t.Errorf("Message should mention function name: %s", err.Message)
	}
	if !strings.Contains(err.Message, "3") {
		t.Errorf("Message should mention expected count: %s", err.Message)
	}
}

func TestNewCircularIncludeError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1}
	err := NewCircularIncludeError("common.build", loc)

	if err.Code != "E109" {
		t.Errorf("Expected code E109, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "circular") {
		t.Errorf("Message should mention circular include: %s", err.Message)
	}
	if !strings.Contains(err.Message, "common.build") {
		t.Errorf("Message should mention file: %s", err.Message)
	}
}

func TestNewIncludeNotFoundError(t *testing.T) {
	loc := ast.SourceLocation{File: "Buildfile", Line: 2, Column: 11}
	err := NewIncludeNotFoundError("missing.build", loc)

	if err.Code != "E110" {
		t.Errorf("Expected code E110, got %s", err.Code)
	}
	if !strings.Contains(err.Message, "missing.build") {
		t.Errorf("Message should mention file: %s", err.Message)
	}
}

// ----------------------------------------------------------------------------
// Error Code Range Tests
// ----------------------------------------------------------------------------

func TestSyntaxErrorCodes_InRange(t *testing.T) {
	testCases := []struct {
		name string
		err  *FormattedError
	}{
		{"UnexpectedToken", NewUnexpectedTokenError("IDENT", "COLON", ast.SourceLocation{})},
		{"MissingColon", NewMissingColonError("target", ast.SourceLocation{})},
		{"MissingEnd", NewMissingEndError(ast.SourceLocation{})},
		{"InvalidDirectiveScope", NewInvalidDirectiveScopeError(".after", "GLOBAL", []string{"RECIPE"}, ast.SourceLocation{})},
		{"MissingCondition", NewMissingConditionError("if", ast.SourceLocation{})},
		{"MissingOperator", NewMissingOperatorError(ast.SourceLocation{})},
		{"MissingIdentifier", NewMissingIdentifierError("ifdef", ast.SourceLocation{})},
		{"InvalidRuntime", NewInvalidRuntimeError("foo", ast.SourceLocation{})},
		{"MissingFunctionArg", NewMissingFunctionArgumentError("replace", 3, 2, ast.SourceLocation{})},
		{"CircularInclude", NewCircularIncludeError("file.build", ast.SourceLocation{})},
		{"IncludeNotFound", NewIncludeNotFoundError("file.build", ast.SourceLocation{})},
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

			if num < 100 || num > 199 {
				t.Errorf("Syntax error code should be E100-E199, got: %s (num=%d)", code, num)
			}
		})
	}
}
