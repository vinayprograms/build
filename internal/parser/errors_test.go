package parser

import (
	"fmt"
	"testing"

	"github.com/vinayprograms/need/internal/lexer"
)

func TestParseError_String(t *testing.T) {
	err := &ParseError{
		Message: "unexpected token",
		Location: lexer.SourceLocation{
			File:   "Needfile",
			Line:   10,
			Column: 5,
		},
	}

	want := "Needfile:10:5: unexpected token"
	got := err.Error()
	if got != want {
		t.Errorf("ParseError.Error() = %q, want %q", got, want)
	}
}

func TestParseError_WithHint(t *testing.T) {
	err := &ParseError{
		Message: "directive '.after' invalid at global scope",
		Location: lexer.SourceLocation{
			File:   "Needfile",
			Line:   1,
			Column: 1,
		},
		Hint: ".after is only valid inside a recipe",
	}

	want := "Needfile:1:1: directive '.after' invalid at global scope (hint: .after is only valid inside a recipe)"
	got := err.Error()
	if got != want {
		t.Errorf("ParseError.Error() = %q, want %q", got, want)
	}
}

func TestScopeError(t *testing.T) {
	tests := []struct {
		name      string
		directive lexer.TokenType
		found     Scope
		wantMsg   string
		wantHint  string
	}{
		{
			name:      "after at global",
			directive: lexer.DOT_AFTER,
			found:     ScopeGlobal,
			wantMsg:   "directive '.after' invalid at GLOBAL scope",
			wantHint:  ".after is only valid in: RECIPE",
		},
		{
			name:      "using at global",
			directive: lexer.DOT_USING,
			found:     ScopeGlobal,
			wantMsg:   "directive '.using' invalid at GLOBAL scope",
			wantHint:  ".using is only valid in: ENVIRONMENT",
		},
		{
			name:      "shell at environment",
			directive: lexer.DOT_SHELL,
			found:     ScopeEnvironment,
			wantMsg:   "directive '.shell' invalid at ENVIRONMENT scope",
			wantHint:  ".shell is only valid in: GLOBAL, RECIPE",
		},
		{
			name:      "parallel in recipe",
			directive: lexer.DOT_PARALLEL,
			found:     ScopeRecipe,
			wantMsg:   "directive '.parallel' invalid at RECIPE scope",
			wantHint:  ".parallel is only valid in: GLOBAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewScopeError(tt.directive, tt.found, lexer.SourceLocation{
				File:   "test.need",
				Line:   1,
				Column: 1,
			})

			if err.Message != tt.wantMsg {
				t.Errorf("message = %q, want %q", err.Message, tt.wantMsg)
			}
			if err.Hint != tt.wantHint {
				t.Errorf("hint = %q, want %q", err.Hint, tt.wantHint)
			}
		})
	}
}

func TestParseErrors_Add(t *testing.T) {
	errs := &ParseErrors{}

	errs.Add(&ParseError{Message: "error 1"})
	errs.Add(&ParseError{Message: "error 2"})

	if len(errs.Errors) != 2 {
		t.Errorf("len(Errors) = %d, want 2", len(errs.Errors))
	}
}

func TestParseErrors_HasErrors(t *testing.T) {
	errs := &ParseErrors{}
	if errs.HasErrors() {
		t.Error("empty ParseErrors.HasErrors() should be false")
	}

	errs.Add(&ParseError{Message: "error"})
	if !errs.HasErrors() {
		t.Error("ParseErrors with error should return true for HasErrors()")
	}
}

func TestParseErrors_Error(t *testing.T) {
	errs := &ParseErrors{}
	errs.Add(&ParseError{
		Message:  "first error",
		Location: lexer.SourceLocation{File: "a.need", Line: 1, Column: 1},
	})
	errs.Add(&ParseError{
		Message:  "second error",
		Location: lexer.SourceLocation{File: "a.need", Line: 5, Column: 3},
	})

	got := errs.Error()
	want := "a.need:1:1: first error\na.need:5:3: second error"
	if got != want {
		t.Errorf("ParseErrors.Error() = %q, want %q", got, want)
	}
}

func TestDirectiveNameForError(t *testing.T) {
	tests := []struct {
		tok  lexer.TokenType
		want string
	}{
		{lexer.DOT_SHELL, ".shell"},
		{lexer.DOT_PARALLEL, ".parallel"},
		{lexer.DOT_DEFAULT, ".default"},
		{lexer.DOT_INCLUDE, ".include"},
		{lexer.DOT_ENVIRONMENT, ".environment"},
		{lexer.DOT_USING, ".using"},
		{lexer.DOT_SOURCE, ".source"},
		{lexer.DOT_ARGS, ".args"},
		{lexer.DOT_REQUIRES, ".requires"},
		{lexer.DOT_AFTER, ".after"},
		{lexer.DOT_AUTODEPS, ".autodeps"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.tok), func(t *testing.T) {
			got := DirectiveNameForError(tt.tok)
			if got != tt.want {
				t.Errorf("DirectiveNameForError(%v) = %q, want %q", tt.tok, got, tt.want)
			}
		})
	}
}
