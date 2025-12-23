package ast

import (
	"fmt"
	"strings"
)

// ----------------------------------------------------------------------------
// Value Types
// ----------------------------------------------------------------------------

// ValuePart is the interface for parts of a value (right-hand side of =).
//
// The valuePartNode() marker method is unexported to prevent external packages
// from implementing this interface, ensuring a closed set of value part types.
//
// Implementers: LiteralValue, Interpolation, FunctionCall
type ValuePart interface {
	valuePartNode()
}

// LiteralValue represents literal text in a value.
type LiteralValue struct {
	Text string
}

func (v *LiteralValue) valuePartNode() {}

// Interpolation represents a variable interpolation in a value.
type Interpolation struct {
	Name     string
	Raw      bool // true for :raw modifier
	Location SourceLocation
}

func (i *Interpolation) valuePartNode() {}

// FunctionName represents a built-in function name.
type FunctionName int

const (
	FuncShell FunctionName = iota
	FuncGlob
	FuncFilename
	FuncDirname
	FuncReplace
)

// String returns the string representation of the function name.
func (f FunctionName) String() string {
	switch f {
	case FuncShell:
		return "shell"
	case FuncGlob:
		return "glob"
	case FuncFilename:
		return "filename"
	case FuncDirname:
		return "dirname"
	case FuncReplace:
		return "replace"
	default:
		return fmt.Sprintf("unknown(%d)", f)
	}
}

// FunctionCall represents a function call in a value.
type FunctionCall struct {
	Name     FunctionName
	Args     []*Value
	Location SourceLocation
}

func (f *FunctionCall) valuePartNode() {}

// Value represents a composite value (string with interpolations and function calls).
type Value struct {
	Parts    []ValuePart
	Location SourceLocation
}

// String returns a human-readable representation of the value.
func (v *Value) String() string {
	var sb strings.Builder
	for _, part := range v.Parts {
		switch p := part.(type) {
		case *LiteralValue:
			sb.WriteString(p.Text)
		case *Interpolation:
			sb.WriteString("{")
			sb.WriteString(p.Name)
			if p.Raw {
				sb.WriteString(":raw")
			}
			sb.WriteString("}")
		case *FunctionCall:
			sb.WriteString(p.Name.String())
			sb.WriteString("(")
			for i, arg := range p.Args {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(arg.String())
			}
			sb.WriteString(")")
		case *LiteralSegment:
			sb.WriteString(p.Text)
		}
	}
	return sb.String()
}
