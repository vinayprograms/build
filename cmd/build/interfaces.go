// Package main provides the CLI for the build tool.
//
// This file defines interfaces for components used by the CLI.
// Internal packages provide concrete implementations that satisfy these interfaces.
// This follows the Go idiom of "accept interfaces, return structs" and allows
// the CLI to remain decoupled from implementation details.
package main

// Token represents a lexical token from source code.
// The CLI uses this interface to inspect tokens without depending on
// the concrete lexer.Token type.
type Token interface {
	// TokenType returns the type of the token as a string.
	TokenType() string

	// TokenLiteral returns the literal value of the token.
	TokenLiteral() string

	// TokenLocation returns the source location as "file:line:column".
	TokenLocation() string

	// IsEOF returns true if this is an end-of-file token.
	IsEOF() bool

	// IsError returns true if this is an error token.
	IsError() bool
}

// Lexer tokenizes source code into a stream of tokens.
// Implementations should be created via a factory function that accepts
// the file path and source content.
type Lexer interface {
	// NextToken returns the next token from the input.
	// Returns a token with IsEOF() == true when input is exhausted.
	NextToken() Token
}

// LexerFactory creates a new Lexer for the given source.
type LexerFactory func(file, input string) Lexer

// Scope represents the parsing context (global, environment, recipe, etc.).
type Scope interface {
	// String returns a human-readable name for the scope.
	String() string
}

// Parser transforms a token stream into an AST.
// Implementations track parsing scope for directive validation.
type Parser interface {
	// CurrentScope returns the current parsing scope.
	CurrentScope() Scope

	// EnterScope pushes a new scope onto the stack.
	EnterScope(scope Scope)

	// ExitScope pops the current scope and returns it.
	ExitScope() Scope

	// CurrentIndentLevel returns the expected indentation level for the current scope.
	CurrentIndentLevel() int

	// HasErrors returns true if any parse errors were encountered.
	HasErrors() bool
}

// ParserFactory creates a new Parser for the given Lexer.
// Note: The factory accepts the concrete lexer type since the parser
// needs to call NextToken() on it during parsing.
type ParserFactory func(lexer Lexer) Parser

// DirectiveValidator checks if a directive is valid at a given scope.
type DirectiveValidator interface {
	// IsValidAtScope returns true if the directive token type is valid at the given scope.
	IsValidAtScope(tokenType string, scope Scope) bool

	// DirectiveName returns the display name for a directive token type (e.g., ".shell").
	DirectiveName(tokenType string) string
}

// Variable represents a parsed variable definition from the AST.
type Variable interface {
	// Name returns the variable name.
	Name() string

	// IsLazy returns true if this is a lazy variable.
	IsLazy() bool

	// ValueParts returns the parts of the variable value.
	// Each part is either a literal string, interpolation, or function call.
	ValueParts() []ValuePart

	// Location returns the source location as "file:line:column".
	Location() string
}

// ValuePart represents a part of a value (literal, interpolation, or function call).
type ValuePart interface {
	// PartType returns "literal", "interpolation", or "function".
	PartType() string

	// Text returns the text content (for literals) or name (for interpolations/functions).
	Text() string

	// IsRaw returns true for interpolations with :raw modifier.
	IsRaw() bool
}

// VariableParser parses variable definitions.
type VariableParser interface {
	// ParseVariable parses a variable definition from the current position.
	// Returns the variable and any error encountered.
	ParseVariable() (Variable, error)
}

// Target represents a parsed target definition from the AST.
type Target interface {
	// PatternText returns a textual representation of the target pattern.
	PatternText() string

	// IsPhony returns true for @name phony targets.
	IsPhony() bool

	// IsDirectory returns true for targets ending with /.
	IsDirectory() bool

	// DependencyCount returns the number of dependencies.
	DependencyCount() int

	// DependencyText returns the textual representation of the i-th dependency.
	DependencyText(i int) string

	// HasCaptures returns true if the pattern contains {name} expressions.
	HasCaptures() bool

	// CaptureNames returns the names of captures in the pattern.
	CaptureNames() []string

	// Location returns the source location as "file:line:column".
	Location() string
}

// TargetParser parses target definitions.
type TargetParser interface {
	// ParseTarget parses a target definition from the current position.
	// Returns the target and any error encountered.
	ParseTarget() (Target, error)

	// IsTargetLine returns true if the current line is a target definition.
	IsTargetLine() bool
}
