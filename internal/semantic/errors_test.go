package semantic

import (
	"strings"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

// ----------------------------------------------------------------------------
// DuplicateDefinitionError Tests
// ----------------------------------------------------------------------------

func TestDuplicateDefinitionError_Variable(t *testing.T) {
	err := &DuplicateDefinitionError{
		Kind:   "variable",
		Name:   "cc",
		First:  ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
		Second: ast.SourceLocation{File: "Buildfile", Line: 5, Column: 1},
	}

	msg := err.Error()
	if !strings.Contains(msg, "duplicate variable 'cc'") {
		t.Errorf("Expected error to mention 'duplicate variable', got: %s", msg)
	}
	if !strings.Contains(msg, "Buildfile:1:1") {
		t.Errorf("Expected error to mention first location, got: %s", msg)
	}
	if !strings.Contains(msg, "Buildfile:5:1") {
		t.Errorf("Expected error to mention second location, got: %s", msg)
	}
}

func TestDuplicateDefinitionError_Target(t *testing.T) {
	err := &DuplicateDefinitionError{
		Kind:   "target",
		Name:   "build/app",
		First:  ast.SourceLocation{File: "Buildfile", Line: 10, Column: 1},
		Second: ast.SourceLocation{File: "Buildfile", Line: 20, Column: 1},
	}

	msg := err.Error()
	if !strings.Contains(msg, "duplicate target 'build/app'") {
		t.Errorf("Expected error to mention 'duplicate target', got: %s", msg)
	}
}

func TestDuplicateDefinitionError_Environment(t *testing.T) {
	err := &DuplicateDefinitionError{
		Kind:   "environment",
		Name:   "ci",
		First:  ast.SourceLocation{File: "Buildfile", Line: 15, Column: 1},
		Second: ast.SourceLocation{File: "Buildfile", Line: 25, Column: 1},
	}

	msg := err.Error()
	if !strings.Contains(msg, "duplicate environment 'ci'") {
		t.Errorf("Expected error to mention 'duplicate environment', got: %s", msg)
	}
}

// ----------------------------------------------------------------------------
// AutomaticInPatternError Tests
// ----------------------------------------------------------------------------

func TestAutomaticInPatternError_Target(t *testing.T) {
	err := &AutomaticInPatternError{
		Name:     "target",
		Location: ast.SourceLocation{File: "Buildfile", Line: 5, Column: 10},
	}

	msg := err.Error()
	if !strings.Contains(msg, "automatic variable 'target'") {
		t.Errorf("Expected error to mention 'automatic variable', got: %s", msg)
	}
	if !strings.Contains(msg, "cannot be used") || !strings.Contains(msg, "capture") {
		t.Errorf("Expected error to explain capture prohibition, got: %s", msg)
	}
	if !strings.Contains(msg, "Buildfile:5:10") {
		t.Errorf("Expected error to include location, got: %s", msg)
	}
}

func TestAutomaticInPatternError_AllAutomaticVars(t *testing.T) {
	automaticVars := []string{"target", "deps", "in", "out", "stem", "target.dir", "target.file"}

	for _, name := range automaticVars {
		err := &AutomaticInPatternError{
			Name:     name,
			Location: ast.SourceLocation{File: "test", Line: 1, Column: 1},
		}

		msg := err.Error()
		if !strings.Contains(msg, name) {
			t.Errorf("Expected error for '%s' to contain the variable name, got: %s", name, msg)
		}
	}
}

// ----------------------------------------------------------------------------
// CaptureMismatchError Tests
// ----------------------------------------------------------------------------

func TestCaptureMismatchError_ExtraInDependency(t *testing.T) {
	err := &CaptureMismatchError{
		Name:      "name",
		InTarget:  false,
		Location:  ast.SourceLocation{File: "Buildfile", Line: 3, Column: 20},
		TargetLoc: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 1},
	}

	msg := err.Error()
	if !strings.Contains(msg, "name") {
		t.Errorf("Expected error to mention capture name, got: %s", msg)
	}
	if !strings.Contains(msg, "dependency") {
		t.Errorf("Expected error to mention dependency, got: %s", msg)
	}
	if !strings.Contains(msg, "Buildfile:3:20") {
		t.Errorf("Expected error to include capture location, got: %s", msg)
	}
}

// ----------------------------------------------------------------------------
// UndefinedVariableError Tests
// ----------------------------------------------------------------------------

func TestUndefinedVariableError_Basic(t *testing.T) {
	err := &UndefinedVariableError{
		Name:     "foo",
		Location: ast.SourceLocation{File: "Buildfile", Line: 7, Column: 15},
	}

	msg := err.Error()
	if !strings.Contains(msg, "undefined variable 'foo'") {
		t.Errorf("Expected error to mention 'undefined variable', got: %s", msg)
	}
	if !strings.Contains(msg, "Buildfile:7:15") {
		t.Errorf("Expected error to include location, got: %s", msg)
	}
}

func TestUndefinedVariableError_DottedName(t *testing.T) {
	err := &UndefinedVariableError{
		Name:     "some.var",
		Location: ast.SourceLocation{File: "test.build", Line: 1, Column: 1},
	}

	msg := err.Error()
	if !strings.Contains(msg, "some.var") {
		t.Errorf("Expected error to include dotted name, got: %s", msg)
	}
}

// ----------------------------------------------------------------------------
// AutomaticOutsideRecipeError Tests
// ----------------------------------------------------------------------------

func TestAutomaticOutsideRecipeError_InVariable(t *testing.T) {
	err := &AutomaticOutsideRecipeError{
		Name:     "target",
		Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 10},
	}

	msg := err.Error()
	if !strings.Contains(msg, "automatic variable 'target'") {
		t.Errorf("Expected error to mention 'automatic variable', got: %s", msg)
	}
	if !strings.Contains(msg, "recipe") || !strings.Contains(msg, "block") {
		t.Errorf("Expected error to mention valid scopes, got: %s", msg)
	}
	if !strings.Contains(msg, "Buildfile:2:10") {
		t.Errorf("Expected error to include location, got: %s", msg)
	}
}

func TestAutomaticOutsideRecipeError_AllAutomaticVars(t *testing.T) {
	automaticVars := []string{"target", "deps", "in", "out", "stem", "target.dir", "target.file"}

	for _, name := range automaticVars {
		err := &AutomaticOutsideRecipeError{
			Name:     name,
			Location: ast.SourceLocation{File: "test", Line: 1, Column: 1},
		}

		msg := err.Error()
		if !strings.Contains(msg, name) {
			t.Errorf("Expected error for '%s' to contain the variable name, got: %s", name, msg)
		}
	}
}

// ----------------------------------------------------------------------------
// CircularDependencyError Tests
// ----------------------------------------------------------------------------

func TestCircularDependencyError_SelfLoop(t *testing.T) {
	err := &CircularDependencyError{
		Cycle: []string{"a", "a"},
	}

	msg := err.Error()
	if !strings.Contains(msg, "circular dependency") {
		t.Errorf("Expected error to mention 'circular dependency', got: %s", msg)
	}
	if !strings.Contains(msg, "a -> a") {
		t.Errorf("Expected error to show cycle path, got: %s", msg)
	}
}

func TestCircularDependencyError_TwoNodes(t *testing.T) {
	err := &CircularDependencyError{
		Cycle: []string{"a", "b", "a"},
	}

	msg := err.Error()
	if !strings.Contains(msg, "a -> b -> a") {
		t.Errorf("Expected error to show full cycle, got: %s", msg)
	}
}

func TestCircularDependencyError_LongCycle(t *testing.T) {
	err := &CircularDependencyError{
		Cycle: []string{"a", "b", "c", "d", "a"},
	}

	msg := err.Error()
	if !strings.Contains(msg, "a -> b -> c -> d -> a") {
		t.Errorf("Expected error to show full cycle, got: %s", msg)
	}
}

// ----------------------------------------------------------------------------
// Error Interface Tests
// ----------------------------------------------------------------------------

func TestAllErrorsImplementError(t *testing.T) {
	// Verify all error types implement the error interface properly

	errors := []error{
		&DuplicateDefinitionError{Kind: "variable", Name: "x"},
		&AutomaticInPatternError{Name: "target"},
		&CaptureMismatchError{Name: "x"},
		&UndefinedVariableError{Name: "x"},
		&AutomaticOutsideRecipeError{Name: "target"},
		&CircularDependencyError{Cycle: []string{"a", "a"}},
	}

	for i, err := range errors {
		if err.Error() == "" {
			t.Errorf("Error %d returned empty message", i)
		}
	}
}

// ----------------------------------------------------------------------------
// Error Location Tests
// ----------------------------------------------------------------------------

func TestErrorsIncludeLocation(t *testing.T) {
	loc := ast.SourceLocation{File: "test.build", Line: 42, Column: 7}
	expectedLocStr := "test.build:42:7"

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "DuplicateDefinitionError includes first location",
			err: &DuplicateDefinitionError{
				Kind:   "variable",
				Name:   "x",
				First:  loc,
				Second: ast.SourceLocation{File: "test.build", Line: 50, Column: 1},
			},
		},
		{
			name: "AutomaticInPatternError includes location",
			err:  &AutomaticInPatternError{Name: "target", Location: loc},
		},
		{
			name: "CaptureMismatchError includes location",
			err:  &CaptureMismatchError{Name: "x", Location: loc, TargetLoc: loc},
		},
		{
			name: "UndefinedVariableError includes location",
			err:  &UndefinedVariableError{Name: "x", Location: loc},
		},
		{
			name: "AutomaticOutsideRecipeError includes location",
			err:  &AutomaticOutsideRecipeError{Name: "target", Location: loc},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			if !strings.Contains(msg, expectedLocStr) {
				t.Errorf("Expected error to include location %s, got: %s", expectedLocStr, msg)
			}
		})
	}
}
