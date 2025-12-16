package ast

// ----------------------------------------------------------------------------
// Variable Types
// ----------------------------------------------------------------------------

// Variable represents a variable definition.
type Variable struct {
	Name     string
	Value    *Value
	Lazy     bool // true for lazy assignment
	Location SourceLocation
}

func (v *Variable) statementNode() {}

// String returns a human-readable representation of the variable.
func (v *Variable) String() string {
	prefix := ""
	if v.Lazy {
		prefix = "lazy "
	}
	valueStr := ""
	if v.Value != nil {
		valueStr = v.Value.String()
	}
	return prefix + v.Name + " = " + valueStr
}
