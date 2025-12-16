package ast

import "strings"

// ----------------------------------------------------------------------------
// Target Types
// ----------------------------------------------------------------------------

// PatternSegment is the interface for segments in a target pattern.
//
// The patternSegmentNode() marker method is unexported to prevent external
// packages from implementing this interface, ensuring a closed set of segment types.
//
// Implementers: LiteralSegment, BraceExpr
type PatternSegment interface {
	patternSegmentNode()
}

// LiteralSegment represents a literal string in a pattern.
type LiteralSegment struct {
	Text string
}

func (s *LiteralSegment) patternSegmentNode() {}

// BraceExpr represents an unresolved {name} in a target pattern.
// During parsing, we don't know if this is a capture or a variable interpolation.
// Semantic analysis will resolve this based on the symbol table.
type BraceExpr struct {
	Identifier string
	Location   SourceLocation
}

func (b *BraceExpr) patternSegmentNode() {}

// LiteralSegment also implements ValuePart for use in .after directives
func (s *LiteralSegment) valuePartNode() {}

// TargetPattern represents the left side of a target definition.
type TargetPattern struct {
	Segments    []PatternSegment
	IsPhony     bool // true for @name targets
	IsDirectory bool // true for targets ending with /
}

// Dependency represents a dependency in a target definition.
type Dependency struct {
	Segments []PatternSegment
}

// Target represents a target definition with dependencies and optional recipe.
type Target struct {
	Pattern      TargetPattern
	Dependencies []Dependency
	Recipe       *Recipe
	Location     SourceLocation
}

func (t *Target) statementNode() {}

// segmentsToString converts pattern segments to a string representation.
func segmentsToString(segments []PatternSegment) string {
	var sb strings.Builder
	for _, seg := range segments {
		switch s := seg.(type) {
		case *LiteralSegment:
			sb.WriteString(s.Text)
		case *BraceExpr:
			sb.WriteString("{")
			sb.WriteString(s.Identifier)
			sb.WriteString("}")
		}
	}
	return sb.String()
}

// String returns a human-readable representation of the target.
func (t *Target) String() string {
	var sb strings.Builder
	if t.Pattern.IsPhony {
		sb.WriteString("@")
	}
	sb.WriteString(segmentsToString(t.Pattern.Segments))
	sb.WriteString(":")
	for i, dep := range t.Dependencies {
		if i > 0 {
			sb.WriteString(" ")
		} else {
			sb.WriteString(" ")
		}
		sb.WriteString(segmentsToString(dep.Segments))
	}
	return sb.String()
}
