// Package lexer implements the lexical analysis phase for Buildfile parsing.
package lexer

import "fmt"

// TokenType represents the type of a lexical token.
type TokenType int

const (
	// Special tokens
	EOF TokenType = iota
	NEWLINE
	INDENT
	COMMENT
	ERROR

	// Dot keywords (directives)
	DOT_SHELL
	DOT_PARALLEL
	DOT_DEFAULT
	DOT_INCLUDE
	DOT_ENVIRONMENT
	DOT_USING
	DOT_SOURCE
	DOT_ARGS
	DOT_REQUIRES
	DOT_AFTER
	DOT_AUTODEPS

	// Keywords
	LAZY
	IF
	ELIF
	ELSE
	END
	IFDEF
	IFNDEF
	BLOCK

	// Operators
	EQUALS        // =
	COLON         // :
	DOUBLE_EQUALS // ==
	NOT_EQUALS    // !=
	LPAREN        // (
	RPAREN        // )

	// Identifiers and literals
	IDENTIFIER    // alphanumeric + underscore
	AT_IDENTIFIER // @name (phony target)
	PATH          // file path
	STRING        // general string content

	// Interpolation
	INTERP_START  // {
	INTERP_END    // }
	INTERP_MOD    // :raw
	ESCAPE_LBRACE // {{
	ESCAPE_RBRACE // }}

	// Built-in functions
	FUNC_SHELL
	FUNC_GLOB
	FUNC_BASENAME
	FUNC_DIRNAME
	FUNC_REPLACE
)

var tokenTypeNames = map[TokenType]string{
	EOF:             "EOF",
	NEWLINE:         "NEWLINE",
	INDENT:          "INDENT",
	COMMENT:         "COMMENT",
	ERROR:           "ERROR",
	DOT_SHELL:       "DOT_SHELL",
	DOT_PARALLEL:    "DOT_PARALLEL",
	DOT_DEFAULT:     "DOT_DEFAULT",
	DOT_INCLUDE:     "DOT_INCLUDE",
	DOT_ENVIRONMENT: "DOT_ENVIRONMENT",
	DOT_USING:       "DOT_USING",
	DOT_SOURCE:      "DOT_SOURCE",
	DOT_ARGS:        "DOT_ARGS",
	DOT_REQUIRES:    "DOT_REQUIRES",
	DOT_AFTER:       "DOT_AFTER",
	DOT_AUTODEPS:    "DOT_AUTODEPS",
	LAZY:            "LAZY",
	IF:              "IF",
	ELIF:            "ELIF",
	ELSE:            "ELSE",
	END:             "END",
	IFDEF:           "IFDEF",
	IFNDEF:          "IFNDEF",
	BLOCK:           "BLOCK",
	EQUALS:          "EQUALS",
	COLON:           "COLON",
	DOUBLE_EQUALS:   "DOUBLE_EQUALS",
	NOT_EQUALS:      "NOT_EQUALS",
	LPAREN:          "LPAREN",
	RPAREN:          "RPAREN",
	IDENTIFIER:      "IDENTIFIER",
	AT_IDENTIFIER:   "AT_IDENTIFIER",
	PATH:            "PATH",
	STRING:          "STRING",
	INTERP_START:    "INTERP_START",
	INTERP_END:      "INTERP_END",
	INTERP_MOD:      "INTERP_MOD",
	ESCAPE_LBRACE:   "ESCAPE_LBRACE",
	ESCAPE_RBRACE:   "ESCAPE_RBRACE",
	FUNC_SHELL:      "FUNC_SHELL",
	FUNC_GLOB:       "FUNC_GLOB",
	FUNC_BASENAME:   "FUNC_BASENAME",
	FUNC_DIRNAME:    "FUNC_DIRNAME",
	FUNC_REPLACE:    "FUNC_REPLACE",
}

// String returns the string representation of the token type.
func (t TokenType) String() string {
	if name, ok := tokenTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(%d)", t)
}

// IsDotKeyword returns true if the token type is a dot keyword (directive).
func (t TokenType) IsDotKeyword() bool {
	return t >= DOT_SHELL && t <= DOT_AUTODEPS
}

// IsKeyword returns true if the token type is a control keyword.
func (t TokenType) IsKeyword() bool {
	return t >= LAZY && t <= BLOCK
}

// IsFunction returns true if the token type is a built-in function.
func (t TokenType) IsFunction() bool {
	return t >= FUNC_SHELL && t <= FUNC_REPLACE
}

var keywords = map[string]TokenType{
	"lazy":     LAZY,
	"if":       IF,
	"elif":     ELIF,
	"else":     ELSE,
	"end":      END,
	"ifdef":    IFDEF,
	"ifndef":   IFNDEF,
	"block":    BLOCK,
	"shell":    FUNC_SHELL,
	"glob":     FUNC_GLOB,
	"basename": FUNC_BASENAME,
	"dirname":  FUNC_DIRNAME,
	"replace":  FUNC_REPLACE,
}

// LookupKeyword checks if an identifier is a keyword or function name.
// Returns IDENTIFIER if not found.
func LookupKeyword(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENTIFIER
}

var dotKeywords = map[string]TokenType{
	".shell":       DOT_SHELL,
	".parallel":    DOT_PARALLEL,
	".default":     DOT_DEFAULT,
	".include":     DOT_INCLUDE,
	".environment": DOT_ENVIRONMENT,
	".using":       DOT_USING,
	".source":      DOT_SOURCE,
	".args":        DOT_ARGS,
	".requires":    DOT_REQUIRES,
	".after":       DOT_AFTER,
	".autodeps":    DOT_AUTODEPS,
}

// LookupDotKeyword checks if a string is a dot keyword.
// Returns the token type and true if found, EOF and false otherwise.
func LookupDotKeyword(s string) (TokenType, bool) {
	if tok, ok := dotKeywords[s]; ok {
		return tok, true
	}
	return EOF, false
}

// SourceLocation represents a position in the source file.
type SourceLocation struct {
	File   string // Source file path
	Line   int    // 1-based line number
	Column int    // 1-based column number
}

// String returns a human-readable representation of the source location.
func (l SourceLocation) String() string {
	return fmt.Sprintf("%s:%d:%d", l.File, l.Line, l.Column)
}

// Token represents a lexical token with its type, literal value, and location.
type Token struct {
	Type     TokenType
	Literal  string
	Location SourceLocation
}

// String returns a human-readable representation of the token.
func (t Token) String() string {
	switch t.Type {
	case EOF, NEWLINE, INDENT:
		return t.Type.String()
	default:
		if t.Literal == "" {
			return t.Type.String()
		}
		return fmt.Sprintf("%s(%s)", t.Type.String(), t.Literal)
	}
}
