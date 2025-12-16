package ast

import "fmt"

// ----------------------------------------------------------------------------
// Directive Types
// ----------------------------------------------------------------------------

// DirectiveKind represents the type of a directive.
type DirectiveKind int

const (
	DirectiveShell DirectiveKind = iota
	DirectiveParallel
	DirectiveDefault
	DirectiveInclude
)

// String returns the string representation of the directive kind.
func (k DirectiveKind) String() string {
	switch k {
	case DirectiveShell:
		return "shell"
	case DirectiveParallel:
		return "parallel"
	case DirectiveDefault:
		return "default"
	case DirectiveInclude:
		return "include"
	default:
		return fmt.Sprintf("unknown(%d)", k)
	}
}

// Directive represents a global directive statement (.shell, .parallel, etc.).
type Directive struct {
	Kind     DirectiveKind
	Value    *Value
	Location SourceLocation
}

func (d *Directive) statementNode() {}

// String returns a human-readable representation of the directive.
func (d *Directive) String() string {
	if d.Value != nil {
		return fmt.Sprintf(".%s: %s", d.Kind.String(), d.Value.String())
	}
	return fmt.Sprintf(".%s:", d.Kind.String())
}
