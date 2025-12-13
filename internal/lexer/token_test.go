package lexer

import (
	"testing"
)

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		want      string
	}{
		{EOF, "EOF"},
		{NEWLINE, "NEWLINE"},
		{INDENT, "INDENT"},
		{COMMENT, "COMMENT"},
		{DOT_SHELL, "DOT_SHELL"},
		{DOT_PARALLEL, "DOT_PARALLEL"},
		{DOT_DEFAULT, "DOT_DEFAULT"},
		{DOT_INCLUDE, "DOT_INCLUDE"},
		{DOT_ENVIRONMENT, "DOT_ENVIRONMENT"},
		{DOT_USING, "DOT_USING"},
		{DOT_SOURCE, "DOT_SOURCE"},
		{DOT_ARGS, "DOT_ARGS"},
		{DOT_REQUIRES, "DOT_REQUIRES"},
		{DOT_AFTER, "DOT_AFTER"},
		{DOT_AUTODEPS, "DOT_AUTODEPS"},
		{LAZY, "LAZY"},
		{IF, "IF"},
		{ELIF, "ELIF"},
		{ELSE, "ELSE"},
		{END, "END"},
		{IFDEF, "IFDEF"},
		{IFNDEF, "IFNDEF"},
		{BLOCK, "BLOCK"},
		{EQUALS, "EQUALS"},
		{COLON, "COLON"},
		{DOUBLE_EQUALS, "DOUBLE_EQUALS"},
		{NOT_EQUALS, "NOT_EQUALS"},
		{IDENTIFIER, "IDENTIFIER"},
		{AT_IDENTIFIER, "AT_IDENTIFIER"},
		{PATH, "PATH"},
		{INTERP_START, "INTERP_START"},
		{INTERP_END, "INTERP_END"},
		{INTERP_MOD, "INTERP_MOD"},
		{ESCAPE_LBRACE, "ESCAPE_LBRACE"},
		{ESCAPE_RBRACE, "ESCAPE_RBRACE"},
		{LPAREN, "LPAREN"},
		{RPAREN, "RPAREN"},
		{STRING, "STRING"},
		{FUNC_SHELL, "FUNC_SHELL"},
		{FUNC_GLOB, "FUNC_GLOB"},
		{FUNC_BASENAME, "FUNC_BASENAME"},
		{FUNC_DIRNAME, "FUNC_DIRNAME"},
		{FUNC_REPLACE, "FUNC_REPLACE"},
		{ERROR, "ERROR"},
		{TokenType(9999), "UNKNOWN(9999)"}, // Unknown token type returns fallback
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.tokenType.String(); got != tt.want {
				t.Errorf("TokenType.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSourceLocation(t *testing.T) {
	loc := SourceLocation{
		File:   "Buildfile",
		Line:   10,
		Column: 5,
	}

	expected := "Buildfile:10:5"
	if got := loc.String(); got != expected {
		t.Errorf("SourceLocation.String() = %q, want %q", got, expected)
	}
}

func TestSourceLocationZeroBased(t *testing.T) {
	loc := SourceLocation{
		File:   "test.build",
		Line:   1,
		Column: 1,
	}

	expected := "test.build:1:1"
	if got := loc.String(); got != expected {
		t.Errorf("SourceLocation.String() = %q, want %q", got, expected)
	}
}

func TestTokenCreation(t *testing.T) {
	loc := SourceLocation{File: "Buildfile", Line: 5, Column: 10}
	tok := Token{
		Type:     IDENTIFIER,
		Literal:  "foo",
		Location: loc,
	}

	if tok.Type != IDENTIFIER {
		t.Errorf("Token.Type = %v, want %v", tok.Type, IDENTIFIER)
	}
	if tok.Literal != "foo" {
		t.Errorf("Token.Literal = %q, want %q", tok.Literal, "foo")
	}
	if tok.Location.Line != 5 {
		t.Errorf("Token.Location.Line = %d, want %d", tok.Location.Line, 5)
	}
}

func TestTokenString(t *testing.T) {
	tests := []struct {
		name  string
		token Token
		want  string
	}{
		{
			name: "identifier",
			token: Token{
				Type:     IDENTIFIER,
				Literal:  "foo",
				Location: SourceLocation{File: "Buildfile", Line: 1, Column: 1},
			},
			want: "IDENTIFIER(foo)",
		},
		{
			name: "eof",
			token: Token{
				Type:     EOF,
				Literal:  "",
				Location: SourceLocation{File: "Buildfile", Line: 10, Column: 1},
			},
			want: "EOF",
		},
		{
			name: "newline",
			token: Token{
				Type:     NEWLINE,
				Literal:  "\n",
				Location: SourceLocation{File: "Buildfile", Line: 5, Column: 20},
			},
			want: "NEWLINE",
		},
		{
			name: "string with content",
			token: Token{
				Type:     STRING,
				Literal:  "hello world",
				Location: SourceLocation{File: "Buildfile", Line: 3, Column: 5},
			},
			want: "STRING(hello world)",
		},
		{
			name: "dot directive",
			token: Token{
				Type:     DOT_SHELL,
				Literal:  ".shell",
				Location: SourceLocation{File: "Buildfile", Line: 1, Column: 1},
			},
			want: "DOT_SHELL(.shell)",
		},
		{
			name: "error token",
			token: Token{
				Type:     ERROR,
				Literal:  "unexpected character",
				Location: SourceLocation{File: "Buildfile", Line: 7, Column: 3},
			},
			want: "ERROR(unexpected character)",
		},
		{
			name: "operator with empty literal",
			token: Token{
				Type:     EQUALS,
				Literal:  "",
				Location: SourceLocation{File: "Buildfile", Line: 2, Column: 5},
			},
			want: "EQUALS",
		},
		{
			name: "indent token",
			token: Token{
				Type:     INDENT,
				Literal:  "    ",
				Location: SourceLocation{File: "Buildfile", Line: 3, Column: 1},
			},
			want: "INDENT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.String(); got != tt.want {
				t.Errorf("Token.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsDotKeyword(t *testing.T) {
	dotKeywords := []TokenType{
		DOT_SHELL, DOT_PARALLEL, DOT_DEFAULT, DOT_INCLUDE,
		DOT_ENVIRONMENT, DOT_USING, DOT_SOURCE, DOT_ARGS,
		DOT_REQUIRES, DOT_AFTER, DOT_AUTODEPS,
	}

	for _, tt := range dotKeywords {
		if !tt.IsDotKeyword() {
			t.Errorf("%v.IsDotKeyword() = false, want true", tt)
		}
	}

	nonDotKeywords := []TokenType{
		EOF, NEWLINE, IDENTIFIER, IF, ELSE, EQUALS,
	}

	for _, tt := range nonDotKeywords {
		if tt.IsDotKeyword() {
			t.Errorf("%v.IsDotKeyword() = true, want false", tt)
		}
	}
}

func TestIsKeyword(t *testing.T) {
	keywords := []TokenType{
		LAZY, IF, ELIF, ELSE, END, IFDEF, IFNDEF, BLOCK,
	}

	for _, tt := range keywords {
		if !tt.IsKeyword() {
			t.Errorf("%v.IsKeyword() = false, want true", tt)
		}
	}

	nonKeywords := []TokenType{
		EOF, NEWLINE, IDENTIFIER, DOT_SHELL, EQUALS,
	}

	for _, tt := range nonKeywords {
		if tt.IsKeyword() {
			t.Errorf("%v.IsKeyword() = true, want false", tt)
		}
	}
}

func TestIsFunction(t *testing.T) {
	functions := []TokenType{
		FUNC_SHELL, FUNC_GLOB, FUNC_BASENAME, FUNC_DIRNAME, FUNC_REPLACE,
	}

	for _, tt := range functions {
		if !tt.IsFunction() {
			t.Errorf("%v.IsFunction() = false, want true", tt)
		}
	}

	nonFunctions := []TokenType{
		EOF, IDENTIFIER, DOT_SHELL, IF, EQUALS,
	}

	for _, tt := range nonFunctions {
		if tt.IsFunction() {
			t.Errorf("%v.IsFunction() = true, want false", tt)
		}
	}
}

func TestLookupKeyword(t *testing.T) {
	tests := []struct {
		input string
		want  TokenType
	}{
		{"lazy", LAZY},
		{"if", IF},
		{"elif", ELIF},
		{"else", ELSE},
		{"end", END},
		{"ifdef", IFDEF},
		{"ifndef", IFNDEF},
		{"block", BLOCK},
		{"shell", FUNC_SHELL},
		{"glob", FUNC_GLOB},
		{"basename", FUNC_BASENAME},
		{"dirname", FUNC_DIRNAME},
		{"replace", FUNC_REPLACE},
		{"foo", IDENTIFIER},
		{"variable", IDENTIFIER},
		{"LAZY", IDENTIFIER}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := LookupKeyword(tt.input); got != tt.want {
				t.Errorf("LookupKeyword(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLookupDotKeyword(t *testing.T) {
	tests := []struct {
		input string
		want  TokenType
		found bool
	}{
		{".shell", DOT_SHELL, true},
		{".parallel", DOT_PARALLEL, true},
		{".default", DOT_DEFAULT, true},
		{".include", DOT_INCLUDE, true},
		{".environment", DOT_ENVIRONMENT, true},
		{".using", DOT_USING, true},
		{".source", DOT_SOURCE, true},
		{".args", DOT_ARGS, true},
		{".requires", DOT_REQUIRES, true},
		{".after", DOT_AFTER, true},
		{".autodeps", DOT_AUTODEPS, true},
		{".unknown", EOF, false},
		{"shell", EOF, false},  // missing dot
		{".SHELL", EOF, false}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, found := LookupDotKeyword(tt.input)
			if found != tt.found {
				t.Errorf("LookupDotKeyword(%q) found = %v, want %v", tt.input, found, tt.found)
			}
			if found && got != tt.want {
				t.Errorf("LookupDotKeyword(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
