// Package semantic provides semantic analysis for Buildfiles.
//
// This file contains tests for Pass 3: Reference Validation.
package semantic

import (
	"strings"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

// ----------------------------------------------------------------------------
// Undefined Variable Tests
// ----------------------------------------------------------------------------

func TestValidateReferences_DefinedVariable(t *testing.T) {
	// Variable is defined, reference should pass
	st := NewSymbolTable()
	_ = st.AddVariable(&ast.Variable{
		Name:     "cc",
		Value:    &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "gcc"}}},
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	})

	// Variable value that references cc
	stmts := []ast.Statement{
		&ast.Variable{
			Name: "flags",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{
						Name:     "cc",
						Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 10},
					},
					&ast.LiteralValue{Text: " -Wall"},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 1},
		},
	}

	result := ValidateReferences(st, stmts)
	if result.HasErrors() {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
}

func TestValidateReferences_UndefinedVariable(t *testing.T) {
	// Reference to undefined variable should error
	st := NewSymbolTable()

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "flags",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{
						Name:     "undefined_var",
						Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 10},
					},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
		},
	}

	result := ValidateReferences(st, stmts)
	if !result.HasErrors() {
		t.Fatal("expected error for undefined variable")
	}

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}

	err, ok := result.Errors[0].(*UndefinedVariableError)
	if !ok {
		t.Fatalf("expected UndefinedVariableError, got %T", result.Errors[0])
	}

	if err.Name != "undefined_var" {
		t.Errorf("expected name 'undefined_var', got %q", err.Name)
	}
}

func TestValidateReferences_BuiltinVariable(t *testing.T) {
	// Built-in variables (os, arch) are always valid
	st := NewSymbolTable()

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "platform",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{
						Name:     "os",
						Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 12},
					},
					&ast.LiteralValue{Text: "-"},
					&ast.Interpolation{
						Name:     "arch",
						Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 20},
					},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
		},
	}

	result := ValidateReferences(st, stmts)
	if result.HasErrors() {
		t.Errorf("expected no errors for built-in variables, got: %v", result.Errors)
	}
}

func TestValidateReferences_ConditionalVariable(t *testing.T) {
	// Variables defined in conditionals should be recognized
	st := NewSymbolTable()
	st.AddConditionalVariable(&ConditionalVarDef{
		Variable: &ast.Variable{
			Name:     "cc",
			Value:    &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "gcc"}}},
			Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 1},
		},
		BranchType:  "if",
		BranchIndex: -1,
	})

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "flags",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{
						Name:     "cc",
						Location: ast.SourceLocation{File: "Buildfile", Line: 5, Column: 10},
					},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 5, Column: 1},
		},
	}

	result := ValidateReferences(st, stmts)
	if result.HasErrors() {
		t.Errorf("expected no errors for conditional variable, got: %v", result.Errors)
	}
}

// ----------------------------------------------------------------------------
// Automatic Variable Scope Tests
// ----------------------------------------------------------------------------

func TestValidateReferences_AutomaticInVariableValue(t *testing.T) {
	// Automatic variables (target, deps, etc.) are invalid outside recipes
	st := NewSymbolTable()

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "output",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{
						Name:     "target",
						Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 10},
					},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
		},
	}

	result := ValidateReferences(st, stmts)
	if !result.HasErrors() {
		t.Fatal("expected error for automatic variable outside recipe")
	}

	err, ok := result.Errors[0].(*AutomaticOutsideRecipeError)
	if !ok {
		t.Fatalf("expected AutomaticOutsideRecipeError, got %T", result.Errors[0])
	}

	if err.Name != "target" {
		t.Errorf("expected name 'target', got %q", err.Name)
	}
}

func TestValidateReferences_AutomaticInRecipe(t *testing.T) {
	// Automatic variables are valid inside recipes
	st := NewSymbolTable()

	target := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "build/app"},
			},
		},
		Dependencies: []ast.Dependency{
			{Segments: []ast.PatternSegment{&ast.LiteralSegment{Text: "src/main.c"}}},
		},
		Recipe: &ast.Recipe{
			Commands: []ast.Command{
				&ast.LineCommand{
					Parts: []ast.CommandPart{
						&ast.LiteralCommand{Text: "gcc -o "},
						&ast.CommandInterpolation{
							Name:     "target",
							Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 12},
						},
						&ast.LiteralCommand{Text: " "},
						&ast.CommandInterpolation{
							Name:     "deps",
							Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 22},
						},
					},
					Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 5},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 5},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	}

	_ = st.AddTarget(target)
	stmts := []ast.Statement{target}

	result := ValidateReferences(st, stmts)
	if result.HasErrors() {
		t.Errorf("expected no errors for automatic variables in recipe, got: %v", result.Errors)
	}
}

func TestValidateReferences_AllAutomaticVariables(t *testing.T) {
	// Test all automatic variables are valid in recipe scope
	automaticVars := []string{"target", "deps", "in", "out", "stem", "target.dir", "target.file"}

	for _, varName := range automaticVars {
		t.Run(varName, func(t *testing.T) {
			st := NewSymbolTable()

			target := &ast.Target{
				Pattern: ast.TargetPattern{
					Segments: []ast.PatternSegment{
						&ast.LiteralSegment{Text: "build/app"},
					},
				},
				Recipe: &ast.Recipe{
					Commands: []ast.Command{
						&ast.LineCommand{
							Parts: []ast.CommandPart{
								&ast.LiteralCommand{Text: "echo "},
								&ast.CommandInterpolation{
									Name:     varName,
									Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 10},
								},
							},
							Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 5},
						},
					},
					Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 5},
				},
				Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
			}

			_ = st.AddTarget(target)
			stmts := []ast.Statement{target}

			result := ValidateReferences(st, stmts)
			if result.HasErrors() {
				t.Errorf("expected %s to be valid in recipe, got: %v", varName, result.Errors)
			}
		})
	}
}

func TestValidateReferences_AutomaticOutsideRecipe_AllVars(t *testing.T) {
	// Test all automatic variables are invalid outside recipe scope
	automaticVars := []string{"target", "deps", "in", "out", "stem", "target.dir", "target.file"}

	for _, varName := range automaticVars {
		t.Run(varName, func(t *testing.T) {
			st := NewSymbolTable()

			stmts := []ast.Statement{
				&ast.Variable{
					Name: "test",
					Value: &ast.Value{
						Parts: []ast.ValuePart{
							&ast.Interpolation{
								Name:     varName,
								Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 10},
							},
						},
					},
					Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
				},
			}

			result := ValidateReferences(st, stmts)
			if !result.HasErrors() {
				t.Errorf("expected error for %s outside recipe", varName)
			}

			err, ok := result.Errors[0].(*AutomaticOutsideRecipeError)
			if !ok {
				t.Fatalf("expected AutomaticOutsideRecipeError, got %T", result.Errors[0])
			}

			if err.Name != varName {
				t.Errorf("expected name %q, got %q", varName, err.Name)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Capture Reference Tests
// ----------------------------------------------------------------------------

func TestValidateReferences_CaptureInRecipe(t *testing.T) {
	// Captures are valid inside the recipe that defines them
	st := NewSymbolTable()

	target := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "build/"},
				&ast.BraceExpr{Identifier: "name", Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 7}},
				&ast.LiteralSegment{Text: ".o"},
			},
		},
		Dependencies: []ast.Dependency{
			{Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "src/"},
				&ast.BraceExpr{Identifier: "name", Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 20}},
				&ast.LiteralSegment{Text: ".c"},
			}},
		},
		Recipe: &ast.Recipe{
			Commands: []ast.Command{
				&ast.LineCommand{
					Parts: []ast.CommandPart{
						&ast.LiteralCommand{Text: "gcc -c "},
						&ast.CommandInterpolation{
							Name:     "in",
							Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 12},
						},
						&ast.LiteralCommand{Text: " -o build/"},
						&ast.CommandInterpolation{
							Name:     "name",
							Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 28},
						},
						&ast.LiteralCommand{Text: ".o"},
					},
					Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 5},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 5},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	}

	_ = st.AddTarget(target)
	stmts := []ast.Statement{target}

	// First validate captures to get capture info
	captureResult := ValidateCaptures(st)
	if captureResult.HasErrors() {
		t.Fatalf("capture validation failed: %v", captureResult.Errors)
	}

	result := ValidateReferences(st, stmts, WithCaptures(captureResult))
	if result.HasErrors() {
		t.Errorf("expected no errors for capture in recipe, got: %v", result.Errors)
	}
}

func TestValidateReferences_CaptureOutsideRecipe(t *testing.T) {
	// Captures are only valid in their defining recipe
	st := NewSymbolTable()

	// Target with capture
	target := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "build/"},
				&ast.BraceExpr{Identifier: "name", Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 7}},
				&ast.LiteralSegment{Text: ".o"},
			},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	}
	_ = st.AddTarget(target)

	// Variable trying to use the capture
	stmts := []ast.Statement{
		target,
		&ast.Variable{
			Name: "obj",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{
						Name:     "name", // "name" is a capture, not a variable
						Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 10},
					},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 1},
		},
	}

	captureResult := ValidateCaptures(st)
	result := ValidateReferences(st, stmts, WithCaptures(captureResult))
	if !result.HasErrors() {
		t.Fatal("expected error for capture reference outside recipe")
	}

	// Should be undefined variable error (capture "name" is not a global variable)
	err, ok := result.Errors[0].(*UndefinedVariableError)
	if !ok {
		t.Fatalf("expected UndefinedVariableError, got %T", result.Errors[0])
	}

	if err.Name != "name" {
		t.Errorf("expected name 'name', got %q", err.Name)
	}
}

// ----------------------------------------------------------------------------
// Directive Value Tests
// ----------------------------------------------------------------------------

func TestValidateReferences_DirectiveValue(t *testing.T) {
	st := NewSymbolTable()
	_ = st.AddVariable(&ast.Variable{
		Name:     "default_target",
		Value:    &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "all"}}},
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	})

	stmts := []ast.Statement{
		&ast.Directive{
			Kind: ast.DirectiveDefault,
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{
						Name:     "default_target",
						Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 12},
					},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 1},
		},
	}

	result := ValidateReferences(st, stmts)
	if result.HasErrors() {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
}

func TestValidateReferences_DirectiveUndefinedVariable(t *testing.T) {
	st := NewSymbolTable()

	stmts := []ast.Statement{
		&ast.Directive{
			Kind: ast.DirectiveDefault,
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{
						Name:     "undefined",
						Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 12},
					},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
		},
	}

	result := ValidateReferences(st, stmts)
	if !result.HasErrors() {
		t.Fatal("expected error for undefined variable in directive")
	}
}

// ----------------------------------------------------------------------------
// Function Argument Tests
// ----------------------------------------------------------------------------

func TestValidateReferences_FunctionArgument(t *testing.T) {
	st := NewSymbolTable()
	_ = st.AddVariable(&ast.Variable{
		Name:     "src_dir",
		Value:    &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "src"}}},
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	})

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "sources",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.FunctionCall{
						Name: ast.FuncGlob,
						Args: []*ast.Value{
							{
								Parts: []ast.ValuePart{
									&ast.Interpolation{
										Name:     "src_dir",
										Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 20},
									},
									&ast.LiteralValue{Text: "/*.c"},
								},
							},
						},
						Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 12},
					},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 1},
		},
	}

	result := ValidateReferences(st, stmts)
	if result.HasErrors() {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
}

func TestValidateReferences_FunctionUndefinedArgument(t *testing.T) {
	st := NewSymbolTable()

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "sources",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.FunctionCall{
						Name: ast.FuncGlob,
						Args: []*ast.Value{
							{
								Parts: []ast.ValuePart{
									&ast.Interpolation{
										Name:     "undefined_dir",
										Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 20},
									},
									&ast.LiteralValue{Text: "/*.c"},
								},
							},
						},
						Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 12},
					},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
		},
	}

	result := ValidateReferences(st, stmts)
	if !result.HasErrors() {
		t.Fatal("expected error for undefined variable in function argument")
	}

	err, ok := result.Errors[0].(*UndefinedVariableError)
	if !ok {
		t.Fatalf("expected UndefinedVariableError, got %T", result.Errors[0])
	}

	if err.Name != "undefined_dir" {
		t.Errorf("expected name 'undefined_dir', got %q", err.Name)
	}
}

// ----------------------------------------------------------------------------
// Conditional Tests
// ----------------------------------------------------------------------------

func TestValidateReferences_ConditionalCondition(t *testing.T) {
	// Interpolations in conditions should be validated
	st := NewSymbolTable()

	stmts := []ast.Statement{
		&ast.Conditional{
			IfBranch: ast.ConditionalBranch{
				Condition: &ast.EqualsCondition{
					Left: &ast.Value{
						Parts: []ast.ValuePart{
							&ast.Interpolation{
								Name:     "os",
								Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 5},
							},
						},
					},
					Right: &ast.Value{
						Parts: []ast.ValuePart{
							&ast.LiteralValue{Text: "linux"},
						},
					},
				},
				Body: []ast.Statement{},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
		},
	}

	result := ValidateReferences(st, stmts)
	if result.HasErrors() {
		t.Errorf("expected no errors for built-in in condition, got: %v", result.Errors)
	}
}

func TestValidateReferences_ConditionalUndefinedCondition(t *testing.T) {
	st := NewSymbolTable()

	stmts := []ast.Statement{
		&ast.Conditional{
			IfBranch: ast.ConditionalBranch{
				Condition: &ast.EqualsCondition{
					Left: &ast.Value{
						Parts: []ast.ValuePart{
							&ast.Interpolation{
								Name:     "undefined_var",
								Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 5},
							},
						},
					},
					Right: &ast.Value{
						Parts: []ast.ValuePart{
							&ast.LiteralValue{Text: "value"},
						},
					},
				},
				Body: []ast.Statement{},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
		},
	}

	result := ValidateReferences(st, stmts)
	if !result.HasErrors() {
		t.Fatal("expected error for undefined variable in condition")
	}
}

func TestValidateReferences_ConditionalBody(t *testing.T) {
	// Variables in conditional body should be validated
	st := NewSymbolTable()

	stmts := []ast.Statement{
		&ast.Conditional{
			IfBranch: ast.ConditionalBranch{
				Condition: &ast.EqualsCondition{
					Left: &ast.Value{
						Parts: []ast.ValuePart{
							&ast.Interpolation{
								Name:     "os",
								Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 5},
							},
						},
					},
					Right: &ast.Value{
						Parts: []ast.ValuePart{
							&ast.LiteralValue{Text: "linux"},
						},
					},
				},
				Body: []ast.Statement{
					&ast.Variable{
						Name: "flags",
						Value: &ast.Value{
							Parts: []ast.ValuePart{
								&ast.Interpolation{
									Name:     "undefined",
									Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 12},
								},
							},
						},
						Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 1},
					},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
		},
	}

	result := ValidateReferences(st, stmts)
	if !result.HasErrors() {
		t.Fatal("expected error for undefined variable in conditional body")
	}
}

// ----------------------------------------------------------------------------
// Environment Tests
// ----------------------------------------------------------------------------

func TestValidateReferences_EnvironmentSource(t *testing.T) {
	st := NewSymbolTable()
	_ = st.AddVariable(&ast.Variable{
		Name:     "docker_dir",
		Value:    &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: ".docker"}}},
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	})

	envName := "ci"
	runtime := ast.RuntimeDocker
	stmts := []ast.Statement{
		&ast.Environment{
			Name:    &envName,
			Runtime: &runtime,
			Source: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{
						Name:     "docker_dir",
						Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 13},
					},
					&ast.LiteralValue{Text: "/Dockerfile"},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 1},
		},
	}

	result := ValidateReferences(st, stmts)
	if result.HasErrors() {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
}

func TestValidateReferences_EnvironmentUndefinedSource(t *testing.T) {
	st := NewSymbolTable()

	envName := "ci"
	runtime := ast.RuntimeDocker
	stmts := []ast.Statement{
		&ast.Environment{
			Name:    &envName,
			Runtime: &runtime,
			Source: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{
						Name:     "undefined_dir",
						Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 13},
					},
					&ast.LiteralValue{Text: "/Dockerfile"},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
		},
	}

	result := ValidateReferences(st, stmts)
	if !result.HasErrors() {
		t.Fatal("expected error for undefined variable in environment source")
	}
}

// ----------------------------------------------------------------------------
// Recipe Directive Tests
// ----------------------------------------------------------------------------

func TestValidateReferences_RecipeShellDirective(t *testing.T) {
	st := NewSymbolTable()
	_ = st.AddVariable(&ast.Variable{
		Name:     "shell_path",
		Value:    &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "/bin/bash"}}},
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	})

	target := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "test"},
			},
			IsPhony: true,
		},
		Recipe: &ast.Recipe{
			Directives: ast.RecipeDirectives{
				Shell: &ast.Value{
					Parts: []ast.ValuePart{
						&ast.Interpolation{
							Name:     "shell_path",
							Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 12},
						},
					},
				},
			},
			Commands: []ast.Command{},
			Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 5},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 1},
	}

	_ = st.AddTarget(target)
	stmts := []ast.Statement{target}

	result := ValidateReferences(st, stmts)
	if result.HasErrors() {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
}

// ----------------------------------------------------------------------------
// Block Command Tests
// ----------------------------------------------------------------------------

func TestValidateReferences_BlockCommand(t *testing.T) {
	st := NewSymbolTable()

	target := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "build/app"},
			},
		},
		Recipe: &ast.Recipe{
			Commands: []ast.Command{
				&ast.BlockCommand{
					Lines: [][]ast.CommandPart{
						{
							&ast.LiteralCommand{Text: "if [[ -f "},
							&ast.CommandInterpolation{
								Name:     "target",
								Location: ast.SourceLocation{File: "Buildfile", Line: 4, Column: 15},
							},
							&ast.LiteralCommand{Text: " ]]; then"},
						},
						{
							&ast.LiteralCommand{Text: "    rm "},
							&ast.CommandInterpolation{
								Name:     "target",
								Location: ast.SourceLocation{File: "Buildfile", Line: 5, Column: 12},
							},
						},
						{
							&ast.LiteralCommand{Text: "fi"},
						},
					},
					Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 5},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 5},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	}

	_ = st.AddTarget(target)
	stmts := []ast.Statement{target}

	result := ValidateReferences(st, stmts)
	if result.HasErrors() {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
}

// ----------------------------------------------------------------------------
// Multiple Errors Tests
// ----------------------------------------------------------------------------

func TestValidateReferences_MultipleErrors(t *testing.T) {
	st := NewSymbolTable()

	stmts := []ast.Statement{
		&ast.Variable{
			Name: "a",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{
						Name:     "undefined1",
						Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 5},
					},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
		},
		&ast.Variable{
			Name: "b",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{
						Name:     "undefined2",
						Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 5},
					},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 2, Column: 1},
		},
		&ast.Variable{
			Name: "c",
			Value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{
						Name:     "target", // automatic outside recipe
						Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 5},
					},
				},
			},
			Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 1},
		},
	}

	result := ValidateReferences(st, stmts)
	if !result.HasErrors() {
		t.Fatal("expected multiple errors")
	}

	if len(result.Errors) != 3 {
		t.Errorf("expected 3 errors, got %d", len(result.Errors))
	}
}

// ----------------------------------------------------------------------------
// Error Message Tests
// ----------------------------------------------------------------------------

func TestUndefinedVariableError_Error(t *testing.T) {
	err := &UndefinedVariableError{
		Name:     "foo",
		Location: ast.SourceLocation{File: "Buildfile", Line: 10, Column: 5},
	}

	msg := err.Error()
	if !strings.Contains(msg, "undefined variable") {
		t.Errorf("expected 'undefined variable' in message, got: %s", msg)
	}
	if !strings.Contains(msg, "'foo'") {
		t.Errorf("expected variable name in message, got: %s", msg)
	}
	if !strings.Contains(msg, "Buildfile:10:5") {
		t.Errorf("expected location in message, got: %s", msg)
	}
}

func TestAutomaticOutsideRecipeError_Error(t *testing.T) {
	err := &AutomaticOutsideRecipeError{
		Name:     "target",
		Location: ast.SourceLocation{File: "Buildfile", Line: 5, Column: 10},
	}

	msg := err.Error()
	if !strings.Contains(msg, "automatic variable") {
		t.Errorf("expected 'automatic variable' in message, got: %s", msg)
	}
	if !strings.Contains(msg, "'target'") {
		t.Errorf("expected variable name in message, got: %s", msg)
	}
	if !strings.Contains(msg, "only valid inside recipe") {
		t.Errorf("expected scope hint in message, got: %s", msg)
	}
}
