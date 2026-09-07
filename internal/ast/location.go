// Package ast defines the Abstract Syntax Tree node types for Needfile parsing.
//
// The AST captures the syntactic structure of a Needfile without interpretation.
// No evaluation happens during parsing - the AST is a pure representation of
// the source structure.
//
// Key design decisions:
//   - BraceExpr nodes in target patterns remain unresolved during parsing.
//     Semantic analysis resolves them to Capture or Interpolation.
//   - All nodes carry SourceLocation for error reporting.
//   - Interfaces use marker methods to ensure type safety at compile time.
package ast

import (
	"fmt"

	"github.com/vinayprograms/need/internal/lexer"
)

// SourceLocation represents a position in source code.
type SourceLocation struct {
	File   string // Source file path
	Line   int    // 1-based line number
	Column int    // 1-based column number
}

// String returns a human-readable representation of the source location.
func (l SourceLocation) String() string {
	return fmt.Sprintf("%s:%d:%d", l.File, l.Line, l.Column)
}

// SourceLocationFromToken creates a SourceLocation from a lexer token.
func SourceLocationFromToken(tok lexer.Token) SourceLocation {
	return SourceLocation{
		File:   tok.Location.File,
		Line:   tok.Location.Line,
		Column: tok.Location.Column,
	}
}
