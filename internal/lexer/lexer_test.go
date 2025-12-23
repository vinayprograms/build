package lexer

import (
	"testing"
)

func TestNewLexer(t *testing.T) {
	l := New("test.build", "cc = gcc")
	if l == nil {
		t.Fatal("New() returned nil")
	}
	if l.file != "test.build" {
		t.Errorf("file = %q, want %q", l.file, "test.build")
	}
}

func TestLexerEmptyInput(t *testing.T) {
	l := New("test.build", "")
	tok := l.NextToken()
	if tok.Type != EOF {
		t.Errorf("empty input: got %v, want EOF", tok.Type)
	}
}

func TestLexerNewlines(t *testing.T) {
	l := New("test.build", "a\nb\n")

	// First line content
	tok := l.NextToken()
	if tok.Type != IDENTIFIER || tok.Literal != "a" {
		t.Errorf("got %v, want IDENTIFIER(a)", tok)
	}

	tok = l.NextToken()
	if tok.Type != NEWLINE {
		t.Errorf("got %v, want NEWLINE", tok.Type)
	}

	// Second line content
	tok = l.NextToken()
	if tok.Type != IDENTIFIER || tok.Literal != "b" {
		t.Errorf("got %v, want IDENTIFIER(b)", tok)
	}

	tok = l.NextToken()
	if tok.Type != NEWLINE {
		t.Errorf("got %v, want NEWLINE", tok.Type)
	}

	tok = l.NextToken()
	if tok.Type != EOF {
		t.Errorf("got %v, want EOF", tok.Type)
	}
}

func TestLexerComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []TokenType
	}{
		{
			name:  "comment only line",
			input: "# this is a comment",
			want:  []TokenType{COMMENT, EOF},
		},
		{
			name:  "comment after content",
			input: "foo # comment",
			want:  []TokenType{IDENTIFIER, COMMENT, EOF},
		},
		{
			name:  "multiple comment lines",
			input: "# line 1\n# line 2",
			want:  []TokenType{COMMENT, NEWLINE, COMMENT, EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)
			for i, want := range tt.want {
				tok := l.NextToken()
				if tok.Type != want {
					t.Errorf("token %d: got %v, want %v", i, tok.Type, want)
				}
			}
		})
	}
}

func TestLexerIndentation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []TokenType
	}{
		{
			name:  "no indent",
			input: "foo",
			want:  []TokenType{IDENTIFIER, EOF},
		},
		{
			name:  "indented line",
			input: "    foo",
			want:  []TokenType{INDENT, IDENTIFIER, EOF},
		},
		{
			name:  "tab indent",
			input: "\tfoo",
			want:  []TokenType{INDENT, IDENTIFIER, EOF},
		},
		{
			name:  "mixed lines",
			input: "foo\n    bar",
			want:  []TokenType{IDENTIFIER, NEWLINE, INDENT, IDENTIFIER, EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)
			for i, want := range tt.want {
				tok := l.NextToken()
				if tok.Type != want {
					t.Errorf("token %d: got %v, want %v", i, tok.Type, want)
				}
			}
		})
	}
}

func TestLexerDirectives(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []TokenType
	}{
		{
			name:  "shell directive",
			input: ".shell:",
			want:  []TokenType{DOT_SHELL, COLON, EOF},
		},
		{
			name:  "parallel directive",
			input: ".parallel:",
			want:  []TokenType{DOT_PARALLEL, COLON, EOF},
		},
		{
			name:  "default directive",
			input: ".default:",
			want:  []TokenType{DOT_DEFAULT, COLON, EOF},
		},
		{
			name:  "include directive",
			input: ".include:",
			want:  []TokenType{DOT_INCLUDE, COLON, EOF},
		},
		{
			name:  "environment directive",
			input: ".environment:",
			want:  []TokenType{DOT_ENVIRONMENT, COLON, EOF},
		},
		{
			name:  "environment with name",
			input: ".environment: ci",
			want:  []TokenType{DOT_ENVIRONMENT, COLON, STRING, EOF},
		},
		{
			name:  "indented directive",
			input: "    .shell:",
			want:  []TokenType{INDENT, DOT_SHELL, COLON, EOF},
		},
		{
			name:  "using directive",
			input: ".using:",
			want:  []TokenType{DOT_USING, COLON, EOF},
		},
		{
			name:  "source directive",
			input: ".source:",
			want:  []TokenType{DOT_SOURCE, COLON, EOF},
		},
		{
			name:  "after directive",
			input: ".after:",
			want:  []TokenType{DOT_AFTER, COLON, EOF},
		},
		{
			name:  "autodeps directive",
			input: ".autodeps:",
			want:  []TokenType{DOT_AUTODEPS, COLON, EOF},
		},
		{
			name:  "requires directive",
			input: ".requires:",
			want:  []TokenType{DOT_REQUIRES, COLON, EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)
			for i, want := range tt.want {
				tok := l.NextToken()
				if tok.Type != want {
					t.Errorf("token %d: got %v, want %v", i, tok.Type, want)
				}
			}
		})
	}
}

func TestLexerVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     []TokenType
		literals []string
	}{
		{
			name:     "simple assignment",
			input:    "cc = gcc",
			want:     []TokenType{IDENTIFIER, EQUALS, STRING, EOF},
			literals: []string{"cc", "=", "gcc", ""},
		},
		{
			name:     "lazy assignment",
			input:    "lazy flags = -Wall",
			want:     []TokenType{LAZY, IDENTIFIER, EQUALS, STRING, EOF},
			literals: []string{"lazy", "flags", "=", "-Wall", ""},
		},
		{
			name:     "assignment with interpolation",
			input:    "out = {build_dir}/app",
			want:     []TokenType{IDENTIFIER, EQUALS, INTERP_START, IDENTIFIER, INTERP_END, STRING, EOF},
			literals: []string{"out", "=", "{", "build_dir", "}", "/app", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)
			for i, want := range tt.want {
				tok := l.NextToken()
				if tok.Type != want {
					t.Errorf("token %d: got %v, want %v", i, tok.Type, want)
				}
				if i < len(tt.literals) && tt.literals[i] != "" && tok.Literal != tt.literals[i] {
					t.Errorf("token %d literal: got %q, want %q", i, tok.Literal, tt.literals[i])
				}
			}
		})
	}
}

func TestLexerValueMode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []TokenType
	}{
		{
			name:  "escaped open brace in value",
			input: "x = {{literal}}",
			want:  []TokenType{IDENTIFIER, EQUALS, ESCAPE_LBRACE, STRING, ESCAPE_RBRACE, EOF},
		},
		{
			name:  "escaped close brace in value",
			input: "x = a}}b",
			want:  []TokenType{IDENTIFIER, EQUALS, STRING, ESCAPE_RBRACE, STRING, EOF},
		},
		{
			name:  "unclosed interpolation in value",
			input: "x = {unclosed",
			want:  []TokenType{IDENTIFIER, EQUALS, ERROR, EOF},
		},
		{
			name:  "function in value",
			input: "x = shell(cmd)",
			want:  []TokenType{IDENTIFIER, EQUALS, FUNC_SHELL, LPAREN, STRING, RPAREN, EOF},
		},
		{
			name:  "value with comment",
			input: "x = value # comment",
			want:  []TokenType{IDENTIFIER, EQUALS, STRING, COMMENT, EOF},
		},
		{
			name:  "value ending at newline",
			input: "x = value\ny = other",
			want:  []TokenType{IDENTIFIER, EQUALS, STRING, NEWLINE, IDENTIFIER, EQUALS, STRING, EOF},
		},
		{
			name:  "nested parens in value",
			input: "x = shell(echo (nested))",
			want:  []TokenType{IDENTIFIER, EQUALS, FUNC_SHELL, LPAREN, STRING, LPAREN, STRING, RPAREN, RPAREN, EOF},
		},
		{
			name:  "value with single close brace",
			input: "x = a}b",
			want:  []TokenType{IDENTIFIER, EQUALS, STRING, INTERP_END, STRING, EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)
			for i, want := range tt.want {
				tok := l.NextToken()
				if tok.Type != want {
					t.Errorf("token %d: got %v(%q), want %v", i, tok.Type, tok.Literal, want)
				}
			}
		})
	}
}

func TestLexerTargets(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []TokenType
	}{
		{
			name:  "simple target",
			input: "build/app: build/main.o",
			want:  []TokenType{PATH, COLON, STRING, EOF},
		},
		{
			name:  "phony target",
			input: "@all: build/app",
			want:  []TokenType{AT_IDENTIFIER, COLON, STRING, EOF},
		},
		{
			name:  "phony target with hyphen",
			input: "@test-cover: @test",
			want:  []TokenType{AT_IDENTIFIER, COLON, STRING, EOF},
		},
		{
			name:  "phony target with multiple hyphens",
			input: "@debug-lex-tokens:",
			want:  []TokenType{AT_IDENTIFIER, COLON, EOF},
		},
		{
			name:  "pattern target",
			input: "build/{name}.o: src/{name}.c",
			want:  []TokenType{PATH, INTERP_START, IDENTIFIER, INTERP_END, PATH, COLON, STRING, INTERP_START, IDENTIFIER, INTERP_END, STRING, EOF},
		},
		{
			name:  "directory target",
			input: "build/:",
			want:  []TokenType{PATH, COLON, EOF},
		},
		{
			name:  "multiple dependencies",
			input: "app: main.o utils.o",
			want:  []TokenType{IDENTIFIER, COLON, STRING, EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)
			for i, want := range tt.want {
				tok := l.NextToken()
				if tok.Type != want {
					t.Errorf("token %d: got %v, want %v", i, tok.Type, want)
				}
			}
		})
	}
}

func TestLexerKeywords(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []TokenType
	}{
		{
			name:  "if keyword",
			input: "if {os} == linux",
			want:  []TokenType{IF, INTERP_START, IDENTIFIER, INTERP_END, DOUBLE_EQUALS, IDENTIFIER, EOF},
		},
		{
			name:  "elif keyword",
			input: "elif {os} == darwin",
			want:  []TokenType{ELIF, INTERP_START, IDENTIFIER, INTERP_END, DOUBLE_EQUALS, IDENTIFIER, EOF},
		},
		{
			name:  "else keyword",
			input: "else",
			want:  []TokenType{ELSE, EOF},
		},
		{
			name:  "end keyword",
			input: "end",
			want:  []TokenType{END, EOF},
		},
		{
			name:  "ifdef keyword",
			input: "ifdef DEBUG",
			want:  []TokenType{IFDEF, IDENTIFIER, EOF},
		},
		{
			name:  "ifndef keyword",
			input: "ifndef CC",
			want:  []TokenType{IFNDEF, IDENTIFIER, EOF},
		},
		{
			name:  "block keyword",
			input: "block:",
			want:  []TokenType{BLOCK, COLON, EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)
			for i, want := range tt.want {
				tok := l.NextToken()
				if tok.Type != want {
					t.Errorf("token %d: got %v, want %v", i, tok.Type, want)
				}
			}
		})
	}
}

func TestLexerInterpolation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     []TokenType
		literals []string
	}{
		{
			name:     "simple interpolation",
			input:    "{var}",
			want:     []TokenType{INTERP_START, IDENTIFIER, INTERP_END, EOF},
			literals: []string{"{", "var", "}", ""},
		},
		{
			name:     "interpolation with raw",
			input:    "{var:raw}",
			want:     []TokenType{INTERP_START, IDENTIFIER, INTERP_MOD, INTERP_END, EOF},
			literals: []string{"{", "var", ":raw", "}", ""},
		},
		{
			name:     "escaped braces",
			input:    "{{escaped}}",
			want:     []TokenType{ESCAPE_LBRACE, IDENTIFIER, ESCAPE_RBRACE, EOF},
			literals: []string{"{{", "escaped", "}}", ""},
		},
		{
			name:     "not interpolation after letter",
			input:    "x{var}",
			want:     []TokenType{IDENTIFIER, STRING, EOF},
			literals: []string{"x", "{var}", ""},
		},
		{
			name:     "not interpolation after dollar",
			input:    "${var}",
			want:     []TokenType{STRING, EOF},
			literals: []string{"${var}", ""},
		},
		{
			name:     "interpolation in path",
			input:    "{dir}/{file}",
			want:     []TokenType{INTERP_START, IDENTIFIER, INTERP_END, PATH, INTERP_START, IDENTIFIER, INTERP_END, EOF},
			literals: []string{"{", "dir", "}", "/", "{", "file", "}", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)
			for i, want := range tt.want {
				tok := l.NextToken()
				if tok.Type != want {
					t.Errorf("token %d: got %v, want %v", i, tok.Type, want)
				}
				if i < len(tt.literals) && tt.literals[i] != "" && tok.Literal != tt.literals[i] {
					t.Errorf("token %d literal: got %q, want %q", i, tok.Literal, tt.literals[i])
				}
			}
		})
	}
}

func TestLexerFunctions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []TokenType
	}{
		{
			name:  "shell function",
			input: "shell(echo hello)",
			want:  []TokenType{FUNC_SHELL, LPAREN, STRING, RPAREN, EOF},
		},
		{
			name:  "glob function",
			input: "glob(src/*.c)",
			want:  []TokenType{FUNC_GLOB, LPAREN, STRING, RPAREN, EOF},
		},
		{
			name:  "filename function",
			input: "filename({file})",
			want:  []TokenType{FUNC_FILENAME, LPAREN, INTERP_START, IDENTIFIER, INTERP_END, RPAREN, EOF},
		},
		{
			name:  "dirname function",
			input: "dirname({path})",
			want:  []TokenType{FUNC_DIRNAME, LPAREN, INTERP_START, IDENTIFIER, INTERP_END, RPAREN, EOF},
		},
		{
			name:  "replace function",
			input: "replace({src}, .c, .o)",
			want:  []TokenType{FUNC_REPLACE, LPAREN, INTERP_START, IDENTIFIER, INTERP_END, COMMA, STRING, COMMA, STRING, RPAREN, EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)
			for i, want := range tt.want {
				tok := l.NextToken()
				if tok.Type != want {
					t.Errorf("token %d: got %v, want %v", i, tok.Type, want)
				}
			}
		})
	}
}

func TestLexerOperators(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []TokenType
	}{
		{
			name:  "equals",
			input: "=",
			want:  []TokenType{EQUALS, EOF},
		},
		{
			name:  "double equals",
			input: "==",
			want:  []TokenType{DOUBLE_EQUALS, EOF},
		},
		{
			name:  "not equals",
			input: "!=",
			want:  []TokenType{NOT_EQUALS, EOF},
		},
		{
			name:  "colon",
			input: ":",
			want:  []TokenType{COLON, EOF},
		},
		{
			name:  "parentheses",
			input: "()",
			want:  []TokenType{LPAREN, RPAREN, EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)
			for i, want := range tt.want {
				tok := l.NextToken()
				if tok.Type != want {
					t.Errorf("token %d: got %v, want %v", i, tok.Type, want)
				}
			}
		})
	}
}

func TestLexerSourceLocation(t *testing.T) {
	input := "foo\n  bar"
	l := New("test.build", input)

	// "foo" at line 1, col 1
	tok := l.NextToken()
	if tok.Location.Line != 1 || tok.Location.Column != 1 {
		t.Errorf("foo location: got %d:%d, want 1:1", tok.Location.Line, tok.Location.Column)
	}

	// NEWLINE at line 1
	tok = l.NextToken()
	if tok.Location.Line != 1 {
		t.Errorf("newline location: got line %d, want 1", tok.Location.Line)
	}

	// INDENT at line 2, col 1
	tok = l.NextToken()
	if tok.Location.Line != 2 || tok.Location.Column != 1 {
		t.Errorf("indent location: got %d:%d, want 2:1", tok.Location.Line, tok.Location.Column)
	}

	// "bar" at line 2, col 3
	tok = l.NextToken()
	if tok.Location.Line != 2 || tok.Location.Column != 3 {
		t.Errorf("bar location: got %d:%d, want 2:3", tok.Location.Line, tok.Location.Column)
	}
}

func TestLexerBlankLines(t *testing.T) {
	input := "foo\n\nbar"
	l := New("test.build", input)

	tok := l.NextToken()
	if tok.Type != IDENTIFIER || tok.Literal != "foo" {
		t.Errorf("got %v, want IDENTIFIER(foo)", tok)
	}

	tok = l.NextToken()
	if tok.Type != NEWLINE {
		t.Errorf("got %v, want NEWLINE", tok.Type)
	}

	tok = l.NextToken()
	if tok.Type != NEWLINE {
		t.Errorf("got %v, want NEWLINE (blank line)", tok.Type)
	}

	tok = l.NextToken()
	if tok.Type != IDENTIFIER || tok.Literal != "bar" {
		t.Errorf("got %v, want IDENTIFIER(bar)", tok)
	}
}

func TestLexerRecipe(t *testing.T) {
	input := `build/app: src/main.c
    gcc -o {target} {in}`

	l := New("test.build", input)

	expected := []TokenType{
		PATH,         // build/app
		COLON,        // :
		STRING,       // src/main.c (in value mode after :)
		NEWLINE,      // \n
		INDENT,       // 4 spaces
		IDENTIFIER,   // gcc
		STRING,       // -o
		INTERP_START, // {
		IDENTIFIER,   // target
		INTERP_END,   // }
		INTERP_START, // {
		IDENTIFIER,   // in
		INTERP_END,   // }
		EOF,
	}

	for i, want := range expected {
		tok := l.NextToken()
		if tok.Type != want {
			t.Errorf("token %d: got %v(%q), want %v", i, tok.Type, tok.Literal, want)
		}
	}
}

func TestLexerCompleteExample(t *testing.T) {
	input := `.shell: bash
.default: @all

cc = gcc

@all: build/app

build/app: build/main.o
    {cc} -o {target} {deps}`

	l := New("test.build", input)

	// Just verify we can lex without errors and get EOF
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF || tok.Type == ERROR {
			break
		}
	}

	last := tokens[len(tokens)-1]
	if last.Type != EOF {
		t.Errorf("expected EOF, got %v: %s", last.Type, last.Literal)
	}

	// Verify we got a reasonable number of tokens
	if len(tokens) < 20 {
		t.Errorf("expected at least 20 tokens, got %d", len(tokens))
	}
}

func TestLexerErrorRecovery(t *testing.T) {
	// Unclosed interpolation
	input := "{unclosed"
	l := New("test.build", input)

	tok := l.NextToken()
	if tok.Type != ERROR {
		t.Errorf("expected ERROR for unclosed interpolation, got %v", tok.Type)
	}
}

func TestLexerEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []TokenType
	}{
		{
			name:  "trailing spaces before EOF",
			input: "x   ",
			want:  []TokenType{IDENTIFIER, EOF},
		},
		{
			name:  "exclamation in condition",
			input: "if !debug",
			want:  []TokenType{IF, STRING, EOF},
		},
		{
			name:  "path with interpolation",
			input: "build/{name}.o: src/{name}.c",
			want:  []TokenType{PATH, INTERP_START, IDENTIFIER, INTERP_END, PATH, COLON, STRING, INTERP_START, IDENTIFIER, INTERP_END, STRING, EOF},
		},
		{
			name:  "interpolation at end of input",
			input: "{var}",
			want:  []TokenType{INTERP_START, IDENTIFIER, INTERP_END, EOF},
		},
		{
			name:  "lone brace in value treated as string",
			input: "x = {",
			want:  []TokenType{IDENTIFIER, EQUALS, STRING, EOF},
		},
		{
			name:  "interpolation with invalid char is error",
			input: "{var!}",
			want:  []TokenType{ERROR, STRING, EOF},
		},
		{
			name:  "close paren in normal mode",
			input: "foo)",
			want:  []TokenType{IDENTIFIER, RPAREN, EOF},
		},
		{
			name:  "single close brace in normal mode",
			input: "foo}bar",
			want:  []TokenType{IDENTIFIER, STRING, EOF},
		},
		{
			name:  "indented comment line",
			input: "    # comment",
			want:  []TokenType{INDENT, COMMENT, EOF},
		},
		{
			name:  "value ending with spaces at EOF",
			input: "x =    ",
			want:  []TokenType{IDENTIFIER, EQUALS, EOF},
		},
		{
			name:  "comment after value in recipe",
			input: "target:\n    cmd # inline comment",
			want:  []TokenType{IDENTIFIER, COLON, NEWLINE, INDENT, IDENTIFIER, COMMENT, EOF},
		},
		{
			name:  "newline after colon ends value",
			input: "target:\nfoo",
			want:  []TokenType{IDENTIFIER, COLON, NEWLINE, IDENTIFIER, EOF},
		},
		{
			name:  "comment immediately after colon",
			input: "target: # comment\nfoo",
			want:  []TokenType{IDENTIFIER, COLON, COMMENT, NEWLINE, IDENTIFIER, EOF},
		},
		{
			name:  "interpolation in path stops at brace",
			input: "build/{var}",
			want:  []TokenType{PATH, INTERP_START, IDENTIFIER, INTERP_END, EOF},
		},
		{
			name:  "identifier followed by brace is not interpolation",
			input: "prefix{var}",
			want:  []TokenType{IDENTIFIER, STRING, EOF},
		},
		{
			name:  "space before interpolation makes it valid",
			input: "prefix {var}",
			want:  []TokenType{IDENTIFIER, INTERP_START, IDENTIFIER, INTERP_END, EOF},
		},
		{
			name:  "value string starting with special char",
			input: "x = )",
			want:  []TokenType{IDENTIFIER, EQUALS, RPAREN, EOF},
		},
		{
			name:  "equals at end of file",
			input: "x =",
			want:  []TokenType{IDENTIFIER, EQUALS, EOF},
		},
		{
			name:  "string ends at interpolation boundary",
			input: "}){var}",
			want:  []TokenType{STRING, INTERP_START, IDENTIFIER, INTERP_END, EOF},
		},
		{
			name:  "newline after equals with spaces",
			input: "x = \n",
			want:  []TokenType{IDENTIFIER, EQUALS, NEWLINE, EOF},
		},
		{
			name:  "comment after equals with spaces",
			input: "x = #comment",
			want:  []TokenType{IDENTIFIER, EQUALS, COMMENT, EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)
			for i, want := range tt.want {
				tok := l.NextToken()
				if tok.Type != want {
					t.Errorf("token %d: got %v(%q), want %v", i, tok.Type, tok.Literal, want)
				}
			}
		})
	}
}

func TestLexerInterpolationErrors(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		errMsg    string
	}{
		{
			name:      "unclosed interpolation in value",
			input:     "x = {var",
			wantError: true,
			errMsg:    "unclosed",
		},
		{
			name:      "unclosed interpolation at EOF",
			input:     "{var",
			wantError: true,
			errMsg:    "unclosed",
		},
		{
			name:      "interpolation starting with digit",
			input:     "{123}",
			wantError: false, // Not an interpolation, treated as string
		},
		{
			name:      "invalid modifier in interpolation",
			input:     "{var:invalid}",
			wantError: true,
			errMsg:    "invalid modifier",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)
			var foundError bool
			var errorMsg string
			for i := 0; i < 20; i++ {
				tok := l.NextToken()
				if tok.Type == ERROR {
					foundError = true
					errorMsg = tok.Literal
					break
				}
				if tok.Type == EOF {
					break
				}
			}
			if tt.wantError && !foundError {
				t.Errorf("expected an error for input %q", tt.input)
			}
			if !tt.wantError && foundError {
				t.Errorf("unexpected error for input %q: %s", tt.input, errorMsg)
			}
		})
	}
}

func TestLexerIndentationErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "mixed tabs and spaces",
			input: "target:\n\t code",
		},
		{
			name:  "inconsistent indent width",
			input: "target:\n    line1\n  line2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)
			var foundError bool
			for i := 0; i < 20; i++ {
				tok := l.NextToken()
				if tok.Type == ERROR {
					foundError = true
					break
				}
				if tok.Type == EOF {
					break
				}
			}
			if !foundError {
				t.Errorf("expected an indentation error for input %q", tt.input)
			}
		})
	}
}

func TestLexerCommandMode(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantTokens []struct {
			typ TokenType
			lit string
		}
	}{
		{
			name:  "command with spaces preserved",
			input: "rm -rf build",
			wantTokens: []struct {
				typ TokenType
				lit string
			}{
				{STRING, "rm -rf build"},
				{EOF, ""},
			},
		},
		{
			name:  "command with interpolation preserves spaces",
			input: "gcc -o {target} {deps}",
			wantTokens: []struct {
				typ TokenType
				lit string
			}{
				{STRING, "gcc -o "},
				{INTERP_START, "{"},
				{IDENTIFIER, "target"},
				{INTERP_END, "}"},
				{STRING, " "},
				{INTERP_START, "{"},
				{IDENTIFIER, "deps"},
				{INTERP_END, "}"},
				{EOF, ""},
			},
		},
		{
			name:  "command with multiple spaces",
			input: "echo    hello   world",
			wantTokens: []struct {
				typ TokenType
				lit string
			}{
				{STRING, "echo    hello   world"},
				{EOF, ""},
			},
		},
		{
			name:  "command ending with interpolation",
			input: "rm {file}",
			wantTokens: []struct {
				typ TokenType
				lit string
			}{
				{STRING, "rm "},
				{INTERP_START, "{"},
				{IDENTIFIER, "file"},
				{INTERP_END, "}"},
				{EOF, ""},
			},
		},
		{
			name:  "command with escaped braces",
			input: "echo {{literal}}",
			wantTokens: []struct {
				typ TokenType
				lit string
			}{
				{STRING, "echo "},
				{ESCAPE_LBRACE, "{{"},
				{STRING, "literal"},
				{ESCAPE_RBRACE, "}}"},
				{EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)
			l.SetCommandMode() // Explicitly set command mode

			for i, want := range tt.wantTokens {
				tok := l.NextToken()
				if tok.Type != want.typ {
					t.Errorf("token %d: got type %v, want %v", i, tok.Type, want.typ)
				}
				if want.lit != "" && tok.Literal != want.lit {
					t.Errorf("token %d: got literal %q, want %q", i, tok.Literal, want.lit)
				}
			}
		})
	}
}

func TestLexerPeekMethods(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		wantIsDotKeyword bool
		wantIsBlock      bool
	}{
		{
			name:             "dot keyword",
			input:            ".shell: bash",
			wantIsDotKeyword: true,
			wantIsBlock:      false,
		},
		{
			name:             "dot keyword with leading spaces",
			input:            "   .after: target",
			wantIsDotKeyword: true,
			wantIsBlock:      false,
		},
		{
			name:             "block keyword",
			input:            "block:",
			wantIsDotKeyword: false,
			wantIsBlock:      true,
		},
		{
			name:             "block keyword with leading spaces",
			input:            "    block:",
			wantIsDotKeyword: false,
			wantIsBlock:      true,
		},
		{
			name:             "regular command",
			input:            "rm -rf build",
			wantIsDotKeyword: false,
			wantIsBlock:      false,
		},
		{
			name:             "command starting with identifier",
			input:            "gcc -o target",
			wantIsDotKeyword: false,
			wantIsBlock:      false,
		},
		{
			name:             "empty input",
			input:            "",
			wantIsDotKeyword: false,
			wantIsBlock:      false,
		},
		{
			name:             "only spaces",
			input:            "   ",
			wantIsDotKeyword: false,
			wantIsBlock:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)

			if got := l.PeekNextIsDotKeyword(); got != tt.wantIsDotKeyword {
				t.Errorf("PeekNextIsDotKeyword() = %v, want %v", got, tt.wantIsDotKeyword)
			}
			if got := l.PeekNextIsBlock(); got != tt.wantIsBlock {
				t.Errorf("PeekNextIsBlock() = %v, want %v", got, tt.wantIsBlock)
			}
		})
	}
}

func TestLexerCommandModeWithNewlines(t *testing.T) {
	input := "rm -rf build\necho done"
	l := New("test.build", input)
	l.SetCommandMode()

	// First command
	tok := l.NextToken()
	if tok.Type != STRING || tok.Literal != "rm -rf build" {
		t.Errorf("got %v(%q), want STRING(\"rm -rf build\")", tok.Type, tok.Literal)
	}

	// Newline - should switch back to normal mode
	tok = l.NextToken()
	if tok.Type != NEWLINE {
		t.Errorf("got %v, want NEWLINE", tok.Type)
	}

	// After newline, mode should be LineStart, then Normal
	tok = l.NextToken()
	// In normal mode, this would be IDENTIFIER
	if tok.Type != IDENTIFIER || tok.Literal != "echo" {
		t.Errorf("got %v(%q), want IDENTIFIER(\"echo\")", tok.Type, tok.Literal)
	}
}

func TestLexerCommandModeInterpolationRestoresMode(t *testing.T) {
	input := "cmd {var} rest"
	l := New("test.build", input)
	l.SetCommandMode()

	// String before interpolation
	tok := l.NextToken()
	if tok.Type != STRING || tok.Literal != "cmd " {
		t.Errorf("got %v(%q), want STRING(\"cmd \")", tok.Type, tok.Literal)
	}

	// Interpolation start
	tok = l.NextToken()
	if tok.Type != INTERP_START {
		t.Errorf("got %v, want INTERP_START", tok.Type)
	}

	// Identifier inside interpolation
	tok = l.NextToken()
	if tok.Type != IDENTIFIER || tok.Literal != "var" {
		t.Errorf("got %v(%q), want IDENTIFIER(\"var\")", tok.Type, tok.Literal)
	}

	// Interpolation end
	tok = l.NextToken()
	if tok.Type != INTERP_END {
		t.Errorf("got %v, want INTERP_END", tok.Type)
	}

	// String after interpolation - should preserve spaces (back in command mode)
	tok = l.NextToken()
	if tok.Type != STRING || tok.Literal != " rest" {
		t.Errorf("got %v(%q), want STRING(\" rest\")", tok.Type, tok.Literal)
	}
}

func TestLexerSetModes(t *testing.T) {
	input := "test value"
	l := New("test.build", input)

	// Default is ModeLineStart, then ModeNormal
	l.SetValueMode()
	tok := l.NextToken()
	// In value mode, leading spaces are skipped by lexValue
	if tok.Type != STRING {
		t.Errorf("value mode: got %v, want STRING", tok.Type)
	}

	l = New("test.build", input)
	l.SetCommandMode()
	tok = l.NextToken()
	// In command mode, the whole string is returned
	if tok.Type != STRING || tok.Literal != "test value" {
		t.Errorf("command mode: got %v(%q), want STRING(\"test value\")", tok.Type, tok.Literal)
	}

	l = New("test.build", input)
	l.SetNormalMode()
	tok = l.NextToken()
	// In normal mode, identifiers are parsed
	if tok.Type != IDENTIFIER || tok.Literal != "test" {
		t.Errorf("normal mode: got %v(%q), want IDENTIFIER(\"test\")", tok.Type, tok.Literal)
	}
}

// TestLexerPeekIsVariableLine tests the peek function for variable line detection.
func TestLexerPeekIsVariableLine(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		isVariable bool
	}{
		// Variables: = appears before : (or no : at all)
		{
			name:       "simple variable",
			input:      "cc = gcc",
			isVariable: true,
		},
		{
			name:       "variable with colon in value",
			input:      "path = /usr/bin:foo",
			isVariable: true,
		},
		{
			name:       "lazy variable",
			input:      "lazy sources = shell(find . -name *.c)",
			isVariable: true,
		},
		{
			name:       "variable no spaces",
			input:      "x=y",
			isVariable: true,
		},
		// Targets: : appears before = (or no = at all)
		{
			name:       "simple target no deps",
			input:      "app:",
			isVariable: false,
		},
		{
			name:       "simple target with deps",
			input:      "app: deps",
			isVariable: false,
		},
		{
			name:       "path target",
			input:      "build/app: src/main.c",
			isVariable: false,
		},
		{
			name:       "target with = in dependency",
			input:      "target: dep=value",
			isVariable: false,
		},
		{
			name:       "phony target",
			input:      "@clean:",
			isVariable: false,
		},
		// Edge cases
		{
			name:       "no = or :",
			input:      "justident",
			isVariable: false,
		},
		{
			name:       "equals and colon at same position impossible - equals first",
			input:      "a=b:c",
			isVariable: true,
		},
		{
			name:       "colon first",
			input:      "a:b=c",
			isVariable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test.build", tt.input)
			got := l.PeekIsVariableLine()
			if got != tt.isVariable {
				t.Errorf("PeekIsVariableLine() = %v, want %v", got, tt.isVariable)
			}
		})
	}
}

// TestLexerPeekIsVariableLine_NoStateChange verifies that PeekIsVariableLine
// doesn't consume any tokens - the lexer state remains unchanged.
func TestLexerPeekIsVariableLine_NoStateChange(t *testing.T) {
	inputs := []string{
		"cc = gcc",
		"build/app: src/main.c",
		"app: deps",
		"path = /usr:bin",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			l := New("test.build", input)

			// Record state before peek
			posBefore := l.pos
			lineBefore := l.line
			colBefore := l.col
			modeBefore := l.mode

			// Call peek
			_ = l.PeekIsVariableLine()

			// Verify state unchanged
			if l.pos != posBefore {
				t.Errorf("pos changed: before=%d, after=%d", posBefore, l.pos)
			}
			if l.line != lineBefore {
				t.Errorf("line changed: before=%d, after=%d", lineBefore, l.line)
			}
			if l.col != colBefore {
				t.Errorf("col changed: before=%d, after=%d", colBefore, l.col)
			}
			if l.mode != modeBefore {
				t.Errorf("mode changed: before=%d, after=%d", modeBefore, l.mode)
			}

			// Verify we can still lex normally
			tok := l.NextToken()
			if tok.Type == ERROR || tok.Type == EOF {
				t.Errorf("NextToken after peek returned %v", tok.Type)
			}
		})
	}
}
