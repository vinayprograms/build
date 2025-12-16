package ast

// ----------------------------------------------------------------------------
// Root Node
// ----------------------------------------------------------------------------

// Buildfile is the root AST node representing an entire Buildfile.
type Buildfile struct {
	SourcePath string      // Path to the source file
	Statements []Statement // Top-level statements
}

// ----------------------------------------------------------------------------
// Statement Interface
// ----------------------------------------------------------------------------

// Statement is the interface for all top-level AST nodes.
//
// The statementNode() marker method is unexported to prevent external packages
// from implementing this interface. This ensures the set of statement types is
// closed and known at compile time, enabling exhaustive type switches without
// a default case.
//
// Implementers: Directive, Environment, Variable, Conditional, Target, Comment, Blank
type Statement interface {
	statementNode()
}

// ----------------------------------------------------------------------------
// Comment and Blank
// ----------------------------------------------------------------------------

// Comment represents a comment line.
type Comment struct {
	Text     string
	Location SourceLocation
}

func (c *Comment) statementNode() {}

// Blank represents a blank line.
type Blank struct {
	Location SourceLocation
}

func (b *Blank) statementNode() {}
