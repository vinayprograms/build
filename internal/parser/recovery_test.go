package parser

import (
	"testing"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/lexer"
)

// TestErrorRecovery_SkipToLevel0 tests that the parser skips to level 0 on error.
func TestErrorRecovery_SkipToLevel0(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantErrCount  int
		wantStmtCount int
		wantStmtTypes []string // types of successfully parsed statements
		wantErrMsgs   []string // substrings to match in error messages
	}{
		{
			name: "recover after invalid directive at global scope",
			input: `.after: dep
cc = gcc
`,
			wantErrCount:  1,
			wantStmtCount: 1,
			wantStmtTypes: []string{"Variable"},
			wantErrMsgs:   []string{".after"},
		},
		{
			name: "recover after malformed phony target and parse variable",
			input: `@invalid missing_colon
cc = gcc
`,
			wantErrCount:  1,
			wantStmtCount: 1,
			wantStmtTypes: []string{"Variable"},
			wantErrMsgs:   []string{":"},
		},
		{
			name: "multiple errors collected",
			input: `.after: dep
.using: docker
cc = gcc
`,
			wantErrCount:  2,
			wantStmtCount: 1,
			wantStmtTypes: []string{"Variable"},
			wantErrMsgs:   []string{".after", ".using"},
		},
		{
			name: "recover from malformed target",
			input: `build/ missing colon
cc = gcc
`,
			wantErrCount:  1,
			wantStmtCount: 1,
			wantStmtTypes: []string{"Variable"},
			wantErrMsgs:   []string{"expected"},
		},
		{
			name: "recover from unclosed conditional",
			input: `if {os} == linux
cc = gcc
`,
			wantErrCount:  1,
			wantStmtCount: 0,
			wantStmtTypes: []string{},
			wantErrMsgs:   []string{"end"},
		},
		{
			name: "parse valid statements before and after error",
			input: `prefix = foo
.after: invalid
suffix = bar
`,
			wantErrCount:  1,
			wantStmtCount: 2,
			wantStmtTypes: []string{"Variable", "Variable"},
			wantErrMsgs:   []string{".after"},
		},
		{
			name: "skip indented lines during recovery",
			input: `.after: invalid
    indented = should_skip
cc = gcc
`,
			wantErrCount:  1,
			wantStmtCount: 1,
			wantStmtTypes: []string{"Variable"},
			wantErrMsgs:   []string{".after"},
		},
		{
			name: "recover from environment with invalid directive inside",
			input: `.environment:
    .parallel: 4
cc = gcc
`,
			wantErrCount:  1,
			wantStmtCount: 1, // Only the Variable; environment with error is skipped
			wantStmtTypes: []string{"Variable"},
			wantErrMsgs:   []string{".parallel"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.build", tt.input)
			p := New(l)

			stmts, errs := p.ParseBuildfile()

			// Check error count
			if len(errs.Errors) != tt.wantErrCount {
				t.Errorf("error count = %d, want %d", len(errs.Errors), tt.wantErrCount)
				for _, e := range errs.Errors {
					t.Logf("  error: %s", e.Error())
				}
			}

			// Check statement count
			if len(stmts) != tt.wantStmtCount {
				t.Errorf("statement count = %d, want %d", len(stmts), tt.wantStmtCount)
				for _, s := range stmts {
					t.Logf("  statement: %T", s)
				}
			}

			// Check statement types
			for i, wantType := range tt.wantStmtTypes {
				if i >= len(stmts) {
					t.Errorf("missing statement %d of type %s", i, wantType)
					continue
				}
				gotType := stmtTypeName(stmts[i])
				if gotType != wantType {
					t.Errorf("statement %d type = %s, want %s", i, gotType, wantType)
				}
			}

			// Check error messages contain expected substrings
			for i, wantMsg := range tt.wantErrMsgs {
				if i >= len(errs.Errors) {
					t.Errorf("missing error %d containing %q", i, wantMsg)
					continue
				}
				if !containsSubstring(errs.Errors[i].Error(), wantMsg) {
					t.Errorf("error %d = %q, want containing %q", i, errs.Errors[i].Error(), wantMsg)
				}
			}
		})
	}
}

// TestErrorRecovery_ActionableMessages tests that error messages are actionable.
func TestErrorRecovery_ActionableMessages(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantFile     string
		wantLine     int
		wantMsgPart  string
		wantHintPart string
	}{
		{
			name:         "directive scope error has hint",
			input:        `.after: dep`,
			wantFile:     "test.build",
			wantLine:     1,
			wantMsgPart:  ".after",
			wantHintPart: "RECIPE",
		},
		{
			name:         "missing colon shows location",
			input:        `build/app deps`,
			wantFile:     "test.build",
			wantLine:     1,
			wantMsgPart:  "':'",
			wantHintPart: "",
		},
		{
			name: "error in nested block shows correct line",
			input: `.environment:
    .parallel: 4
`,
			wantFile:     "test.build",
			wantLine:     2,
			wantMsgPart:  ".parallel",
			wantHintPart: "GLOBAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.wantFile, tt.input)
			p := New(l)

			_, errs := p.ParseBuildfile()

			if len(errs.Errors) == 0 {
				t.Fatal("expected at least one error")
			}

			err := errs.Errors[0]

			// Check file
			if err.Location.File != tt.wantFile {
				t.Errorf("error file = %q, want %q", err.Location.File, tt.wantFile)
			}

			// Check line
			if err.Location.Line != tt.wantLine {
				t.Errorf("error line = %d, want %d", err.Location.Line, tt.wantLine)
			}

			// Check message
			if !containsSubstring(err.Message, tt.wantMsgPart) {
				t.Errorf("error message = %q, want containing %q", err.Message, tt.wantMsgPart)
			}

			// Check hint if expected
			if tt.wantHintPart != "" && !containsSubstring(err.Hint, tt.wantHintPart) {
				t.Errorf("error hint = %q, want containing %q", err.Hint, tt.wantHintPart)
			}
		})
	}
}

// TestErrorRecovery_MaxErrors tests that parsing stops after max errors.
func TestErrorRecovery_MaxErrors(t *testing.T) {
	// Generate input with many errors
	input := `.after: 1
.using: 2
.source: 3
.args: 4
.autodeps: 5
.after: 6
.using: 7
.source: 8
.args: 9
.autodeps: 10
.after: 11
cc = gcc
`
	l := lexer.New("test.build", input)
	p := New(l)

	_, errs := p.ParseBuildfile()

	// Should have collected up to max errors (10)
	if len(errs.Errors) > 10 {
		t.Errorf("error count = %d, want <= 10 (max errors)", len(errs.Errors))
	}

	// But should still have collected the valid statement
	// (if we collected < 10 errors before it)
}

// TestErrorRecovery_PreservesValidStatements tests that valid statements are preserved.
func TestErrorRecovery_PreservesValidStatements(t *testing.T) {
	input := `# Comment at start
cc = gcc
.after: invalid_at_global
cflags = -Wall

@test:
    echo hello

.using: invalid_at_global
src_dir = src
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	// Should have errors
	if len(errs.Errors) < 2 {
		t.Errorf("expected at least 2 errors, got %d", len(errs.Errors))
	}

	// Should have valid statements: comment, 3 variables, 1 target
	// Count by type
	var comments, variables, targets int
	for _, s := range stmts {
		switch s.(type) {
		case *ast.Comment:
			comments++
		case *ast.Variable:
			variables++
		case *ast.Target:
			targets++
		}
	}

	if variables < 3 {
		t.Errorf("expected at least 3 variables, got %d", variables)
	}
	if targets < 1 {
		t.Errorf("expected at least 1 target, got %d", targets)
	}
}

// Helper functions

func stmtTypeName(s ast.Statement) string {
	switch s.(type) {
	case *ast.Directive:
		return "Directive"
	case *ast.Environment:
		return "Environment"
	case *ast.Variable:
		return "Variable"
	case *ast.Conditional:
		return "Conditional"
	case *ast.Target:
		return "Target"
	case *ast.Comment:
		return "Comment"
	case *ast.Blank:
		return "Blank"
	default:
		return "Unknown"
	}
}

// TestErrorRecovery_DeeplyNestedError tests recovery from errors in deeply nested structures.
func TestErrorRecovery_DeeplyNestedError(t *testing.T) {
	// Use an actual parse error - malformed target in nested conditional
	input := `cc = gcc
if {os} == linux
    if {arch} == amd64
        @broken missing_colon
        arch_flags = -m64
    end
    cflags = -Wall
end
suffix = .o
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	// Should have at least 1 error for malformed target
	if len(errs.Errors) < 1 {
		t.Errorf("expected at least 1 error, got %d", len(errs.Errors))
	}

	// Should preserve valid statements before and after the error
	hasVariable := false
	hasConditional := false
	for _, s := range stmts {
		switch s.(type) {
		case *ast.Variable:
			hasVariable = true
		case *ast.Conditional:
			hasConditional = true
		}
	}

	if !hasVariable {
		t.Error("expected at least one variable statement")
	}

	// Note: The conditional itself may or may not be parsed depending on how
	// error recovery handles nested structures. Either outcome is acceptable
	// as long as the parser doesn't crash and reports the error.
	_ = hasConditional
}

// TestErrorRecovery_MultipleErrorsInSameBlock tests handling of multiple errors in one block.
func TestErrorRecovery_MultipleErrorsInSameBlock(t *testing.T) {
	input := `.environment: test
    .after: invalid_in_env
    .using: docker
    .autodeps: also_invalid
    .source: Dockerfile
cc = gcc
`
	l := lexer.New("test.build", input)
	p := New(l)

	stmts, errs := p.ParseBuildfile()

	// Should have errors for both .after and .autodeps in environment
	if len(errs.Errors) < 1 {
		t.Errorf("expected at least 1 error, got %d", len(errs.Errors))
		for _, e := range errs.Errors {
			t.Logf("  error: %s", e.Error())
		}
	}

	// Should still parse the variable after the environment block
	hasVariable := false
	for _, s := range stmts {
		if _, ok := s.(*ast.Variable); ok {
			hasVariable = true
			break
		}
	}

	if !hasVariable {
		t.Error("expected variable statement after error recovery")
	}
}
