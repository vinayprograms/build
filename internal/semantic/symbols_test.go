// Package semantic provides semantic analysis for Buildfiles.
//
// This file contains tests for the symbol table implementation.
package semantic

import (
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

// TestNewSymbolTable verifies that a new symbol table is initialized correctly
// with empty maps and populated automatic/built-in variable sets.
func TestNewSymbolTable(t *testing.T) {
	st := NewSymbolTable()

	if st == nil {
		t.Fatal("NewSymbolTable returned nil")
	}

	// Check that maps are initialized
	if st.Variables == nil {
		t.Error("Variables map is nil")
	}
	if st.Targets == nil {
		t.Error("Targets slice is nil")
	}
	if st.Environments == nil {
		t.Error("Environments map is nil")
	}

	// Check that automatic variables are populated
	automaticVars := []string{"target", "deps", "in", "out", "stem", "target.dir", "target.file"}
	for _, v := range automaticVars {
		if !st.IsAutomatic(v) {
			t.Errorf("Expected '%s' to be an automatic variable", v)
		}
	}

	// Check that built-in variables are populated
	builtinVars := []string{"os", "arch"}
	for _, v := range builtinVars {
		if !st.IsBuiltin(v) {
			t.Errorf("Expected '%s' to be a built-in variable", v)
		}
	}
}

// TestSymbolTable_AddVariable tests adding variables to the symbol table.
func TestSymbolTable_AddVariable(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		variable *ast.Variable
		wantErr  bool
	}{
		{
			name:    "simple variable",
			varName: "cc",
			variable: &ast.Variable{
				Name:     "cc",
				Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
			},
			wantErr: false,
		},
		{
			name:    "variable with underscore",
			varName: "my_var",
			variable: &ast.Variable{
				Name:     "my_var",
				Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 1},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := NewSymbolTable()
			err := st.AddVariable(tt.variable)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.wantErr {
				if _, ok := st.Variables[tt.varName]; !ok {
					t.Errorf("Variable '%s' not found in symbol table", tt.varName)
				}
			}
		})
	}
}

// TestSymbolTable_DuplicateVariable tests that duplicate variables are detected.
func TestSymbolTable_DuplicateVariable(t *testing.T) {
	st := NewSymbolTable()

	first := &ast.Variable{
		Name:     "cc",
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	}
	second := &ast.Variable{
		Name:     "cc",
		Location: ast.SourceLocation{File: "Buildfile", Line: 5, Column: 1},
	}

	// First add should succeed
	if err := st.AddVariable(first); err != nil {
		t.Fatalf("First add failed: %v", err)
	}

	// Second add should fail
	err := st.AddVariable(second)
	if err == nil {
		t.Error("Expected duplicate variable error")
	}

	// Check error type
	if dupErr, ok := err.(*DuplicateDefinitionError); ok {
		if dupErr.Kind != "variable" {
			t.Errorf("Expected kind 'variable', got '%s'", dupErr.Kind)
		}
		if dupErr.Name != "cc" {
			t.Errorf("Expected name 'cc', got '%s'", dupErr.Name)
		}
		if dupErr.First.Line != 1 {
			t.Errorf("Expected first location line 1, got %d", dupErr.First.Line)
		}
		if dupErr.Second.Line != 5 {
			t.Errorf("Expected second location line 5, got %d", dupErr.Second.Line)
		}
	} else {
		t.Errorf("Expected DuplicateDefinitionError, got %T", err)
	}
}

// TestSymbolTable_AddTarget tests adding targets to the symbol table.
func TestSymbolTable_AddTarget(t *testing.T) {
	st := NewSymbolTable()

	target := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "build/app"},
			},
			IsPhony: false,
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 10, Column: 1},
	}

	err := st.AddTarget(target)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(st.Targets) != 1 {
		t.Errorf("Expected 1 target, got %d", len(st.Targets))
	}
}

// TestSymbolTable_DuplicateTarget tests that duplicate exact targets are detected.
func TestSymbolTable_DuplicateTarget(t *testing.T) {
	st := NewSymbolTable()

	first := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "build/app"},
			},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 10, Column: 1},
	}
	second := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "build/app"},
			},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 20, Column: 1},
	}

	// First add should succeed
	if err := st.AddTarget(first); err != nil {
		t.Fatalf("First add failed: %v", err)
	}

	// Second add should fail
	err := st.AddTarget(second)
	if err == nil {
		t.Error("Expected duplicate target error")
	}

	// Check error type
	if dupErr, ok := err.(*DuplicateDefinitionError); ok {
		if dupErr.Kind != "target" {
			t.Errorf("Expected kind 'target', got '%s'", dupErr.Kind)
		}
	} else {
		t.Errorf("Expected DuplicateDefinitionError, got %T", err)
	}
}

// TestSymbolTable_AddPhonyTarget tests adding phony targets.
func TestSymbolTable_AddPhonyTarget(t *testing.T) {
	st := NewSymbolTable()

	target := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "clean"},
			},
			IsPhony: true,
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 10, Column: 1},
	}

	err := st.AddTarget(target)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestSymbolTable_DuplicatePhonyTarget tests that duplicate phony targets are detected.
func TestSymbolTable_DuplicatePhonyTarget(t *testing.T) {
	st := NewSymbolTable()

	first := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "clean"},
			},
			IsPhony: true,
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 10, Column: 1},
	}
	second := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "clean"},
			},
			IsPhony: true,
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 20, Column: 1},
	}

	if err := st.AddTarget(first); err != nil {
		t.Fatalf("First add failed: %v", err)
	}

	err := st.AddTarget(second)
	if err == nil {
		t.Error("Expected duplicate phony target error")
	}
}

// TestSymbolTable_PatternTargetsAllowed tests that pattern targets with different
// patterns are allowed.
func TestSymbolTable_PatternTargetsAllowed(t *testing.T) {
	st := NewSymbolTable()

	// Pattern: build/{name}.o
	first := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "build/"},
				&ast.BraceExpr{Identifier: "name"},
				&ast.LiteralSegment{Text: ".o"},
			},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 10, Column: 1},
	}

	// Pattern: build/{name}.a
	second := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "build/"},
				&ast.BraceExpr{Identifier: "name"},
				&ast.LiteralSegment{Text: ".a"},
			},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 20, Column: 1},
	}

	if err := st.AddTarget(first); err != nil {
		t.Errorf("First add failed: %v", err)
	}
	if err := st.AddTarget(second); err != nil {
		t.Errorf("Second add should succeed for different pattern: %v", err)
	}
}

// TestSymbolTable_AddEnvironment tests adding environments to the symbol table.
func TestSymbolTable_AddEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		envName     *string
		description string
	}{
		{
			name:        "default environment (nil name)",
			envName:     nil,
			description: "default",
		},
		{
			name:        "named environment",
			envName:     strPtr("ci"),
			description: "ci",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := NewSymbolTable()
			env := &ast.Environment{
				Name:     tt.envName,
				Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
			}

			err := st.AddEnvironment(env)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// For default env, key is ""
			key := ""
			if tt.envName != nil {
				key = *tt.envName
			}
			if _, ok := st.Environments[key]; !ok {
				t.Errorf("Environment '%s' not found in symbol table", key)
			}
		})
	}
}

// TestSymbolTable_DuplicateEnvironment tests that duplicate environments are detected.
func TestSymbolTable_DuplicateEnvironment(t *testing.T) {
	st := NewSymbolTable()

	ciName := "ci"
	first := &ast.Environment{
		Name:     &ciName,
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	}
	second := &ast.Environment{
		Name:     &ciName,
		Location: ast.SourceLocation{File: "Buildfile", Line: 10, Column: 1},
	}

	if err := st.AddEnvironment(first); err != nil {
		t.Fatalf("First add failed: %v", err)
	}

	err := st.AddEnvironment(second)
	if err == nil {
		t.Error("Expected duplicate environment error")
	}

	if dupErr, ok := err.(*DuplicateDefinitionError); ok {
		if dupErr.Kind != "environment" {
			t.Errorf("Expected kind 'environment', got '%s'", dupErr.Kind)
		}
	} else {
		t.Errorf("Expected DuplicateDefinitionError, got %T", err)
	}
}

// TestSymbolTable_DuplicateDefaultEnvironment tests that duplicate default
// environments are detected.
func TestSymbolTable_DuplicateDefaultEnvironment(t *testing.T) {
	st := NewSymbolTable()

	first := &ast.Environment{
		Name:     nil, // default environment
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	}
	second := &ast.Environment{
		Name:     nil, // another default environment
		Location: ast.SourceLocation{File: "Buildfile", Line: 10, Column: 1},
	}

	if err := st.AddEnvironment(first); err != nil {
		t.Fatalf("First add failed: %v", err)
	}

	err := st.AddEnvironment(second)
	if err == nil {
		t.Error("Expected duplicate default environment error")
	}
}

// TestSymbolTable_IsAutomatic tests the automatic variable check.
func TestSymbolTable_IsAutomatic(t *testing.T) {
	st := NewSymbolTable()

	tests := []struct {
		name string
		want bool
	}{
		{"target", true},
		{"deps", true},
		{"in", true},
		{"out", true},
		{"stem", true},
		{"target.dir", true},
		{"target.file", true},
		{"foo", false},
		{"my_var", false},
		{"os", false},   // built-in, not automatic
		{"arch", false}, // built-in, not automatic
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := st.IsAutomatic(tt.name); got != tt.want {
				t.Errorf("IsAutomatic(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// TestSymbolTable_IsBuiltin tests the built-in variable check.
func TestSymbolTable_IsBuiltin(t *testing.T) {
	st := NewSymbolTable()

	tests := []struct {
		name string
		want bool
	}{
		{"os", true},
		{"arch", true},
		{"target", false}, // automatic, not built-in
		{"deps", false},   // automatic, not built-in
		{"foo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := st.IsBuiltin(tt.name); got != tt.want {
				t.Errorf("IsBuiltin(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// TestSymbolTable_LookupVariable tests looking up variables.
func TestSymbolTable_LookupVariable(t *testing.T) {
	st := NewSymbolTable()

	v := &ast.Variable{
		Name:     "cc",
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	}
	_ = st.AddVariable(v)

	// Found
	if found := st.LookupVariable("cc"); found == nil {
		t.Error("Expected to find variable 'cc'")
	}

	// Not found
	if found := st.LookupVariable("cxx"); found != nil {
		t.Error("Did not expect to find variable 'cxx'")
	}
}

// TestSymbolTable_IsDefined tests checking if a variable is defined.
func TestSymbolTable_IsDefined(t *testing.T) {
	st := NewSymbolTable()

	v := &ast.Variable{
		Name:     "cc",
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	}
	_ = st.AddVariable(v)

	// User variable
	if !st.IsDefined("cc") {
		t.Error("Expected 'cc' to be defined")
	}

	// Automatic variable
	if !st.IsDefined("target") {
		t.Error("Expected 'target' to be defined")
	}

	// Built-in variable
	if !st.IsDefined("os") {
		t.Error("Expected 'os' to be defined")
	}

	// Not defined
	if st.IsDefined("undefined_var") {
		t.Error("Did not expect 'undefined_var' to be defined")
	}
}

// TestSymbolTable_TargetPatternString tests generating string representation
// of target patterns for duplicate detection.
func TestSymbolTable_TargetPatternString(t *testing.T) {
	tests := []struct {
		name     string
		pattern  ast.TargetPattern
		expected string
	}{
		{
			name: "simple literal",
			pattern: ast.TargetPattern{
				Segments: []ast.PatternSegment{
					&ast.LiteralSegment{Text: "build/app"},
				},
			},
			expected: "build/app",
		},
		{
			name: "phony target",
			pattern: ast.TargetPattern{
				Segments: []ast.PatternSegment{
					&ast.LiteralSegment{Text: "clean"},
				},
				IsPhony: true,
			},
			expected: "clean",
		},
		{
			name: "pattern with capture",
			pattern: ast.TargetPattern{
				Segments: []ast.PatternSegment{
					&ast.LiteralSegment{Text: "build/"},
					&ast.BraceExpr{Identifier: "name"},
					&ast.LiteralSegment{Text: ".o"},
				},
			},
			expected: "build/{name}.o",
		},
		{
			name: "directory target",
			pattern: ast.TargetPattern{
				Segments: []ast.PatternSegment{
					&ast.LiteralSegment{Text: "build/"},
				},
				IsDirectory: true,
			},
			expected: "build/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PatternString(&tt.pattern)
			if got != tt.expected {
				t.Errorf("PatternString() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestDuplicateDefinitionError_Error tests the error message format.
func TestDuplicateDefinitionError_Error(t *testing.T) {
	err := &DuplicateDefinitionError{
		Kind:   "variable",
		Name:   "cc",
		First:  ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
		Second: ast.SourceLocation{File: "Buildfile", Line: 5, Column: 1},
	}

	msg := err.Error()
	if msg == "" {
		t.Error("Error message should not be empty")
	}

	// Check that key information is in the message
	expectedParts := []string{"duplicate", "variable", "cc", "Buildfile:1:1", "Buildfile:5:1"}
	for _, part := range expectedParts {
		if !contains(msg, part) {
			t.Errorf("Error message should contain %q, got: %s", part, msg)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper to create string pointer
func strPtr(s string) *string {
	return &s
}
