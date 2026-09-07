// Package semantic provides semantic analysis for Needfiles.
//
// This file contains tests for Pass 1: Symbol Collection.
package semantic

import (
	"testing"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/lexer"
	"github.com/vinayprograms/need/internal/parser"
)

// Helper to parse a needfile for testing.
func parseNeedfile(t *testing.T, content string) []ast.Statement {
	t.Helper()
	l := lexer.New("test.need", content)
	p := parser.New(l)
	stmts, errs := p.ParseNeedfile()
	if errs.HasErrors() {
		t.Fatalf("Parse error: %s", errs.Error())
	}
	return stmts
}

// TestCollector_Basic tests basic symbol collection from a needfile.
func TestCollector_Basic(t *testing.T) {
	content := `cc = gcc
cflags = -Wall

@build:
    {cc} {cflags} -o app main.c

.environment: ci
    .using: docker
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	st, errs := result.SymbolTable, result.Errors
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	// Check variables
	if len(st.Variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(st.Variables))
	}
	if _, ok := st.Variables["cc"]; !ok {
		t.Error("Expected variable 'cc'")
	}
	if _, ok := st.Variables["cflags"]; !ok {
		t.Error("Expected variable 'cflags'")
	}

	// Check targets
	if len(st.Targets) != 1 {
		t.Errorf("Expected 1 target, got %d", len(st.Targets))
	}

	// Check environments
	if len(st.Environments) != 1 {
		t.Errorf("Expected 1 environment, got %d", len(st.Environments))
	}
	if _, ok := st.Environments["ci"]; !ok {
		t.Error("Expected environment 'ci'")
	}
}

// TestCollector_DuplicateVariable tests that duplicate variables are detected.
func TestCollector_DuplicateVariable(t *testing.T) {
	content := `cc = gcc
cflags = -Wall
cc = clang
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	errs := result.Errors
	if len(errs) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errs))
	}

	err := errs[0]
	if dup, ok := err.(*DuplicateDefinitionError); ok {
		if dup.Kind != "variable" {
			t.Errorf("Expected kind 'variable', got '%s'", dup.Kind)
		}
		if dup.Name != "cc" {
			t.Errorf("Expected name 'cc', got '%s'", dup.Name)
		}
	} else {
		t.Errorf("Expected DuplicateDefinitionError, got %T", err)
	}
}

// TestCollector_DuplicateTarget tests that duplicate targets are detected.
func TestCollector_DuplicateTarget(t *testing.T) {
	content := `@build:
    echo building

@build:
    echo building again
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	errs := result.Errors
	if len(errs) != 1 {
		t.Fatalf("Expected 1 error, got %d: %v", len(errs), errs)
	}

	err := errs[0]
	if dup, ok := err.(*DuplicateDefinitionError); ok {
		if dup.Kind != "target" {
			t.Errorf("Expected kind 'target', got '%s'", dup.Kind)
		}
	} else {
		t.Errorf("Expected DuplicateDefinitionError, got %T", err)
	}
}

// TestCollector_DuplicateEnvironment tests that duplicate environments are detected.
func TestCollector_DuplicateEnvironment(t *testing.T) {
	content := `.environment: ci
    .using: docker

.environment: ci
    .using: podman
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	errs := result.Errors
	if len(errs) != 1 {
		t.Fatalf("Expected 1 error, got %d: %v", len(errs), errs)
	}

	err := errs[0]
	if dup, ok := err.(*DuplicateDefinitionError); ok {
		if dup.Kind != "environment" {
			t.Errorf("Expected kind 'environment', got '%s'", dup.Kind)
		}
		if dup.Name != "ci" {
			t.Errorf("Expected name 'ci', got '%s'", dup.Name)
		}
	} else {
		t.Errorf("Expected DuplicateDefinitionError, got %T", err)
	}
}

// TestCollector_DuplicateDefaultEnvironment tests that duplicate default environments are detected.
func TestCollector_DuplicateDefaultEnvironment(t *testing.T) {
	content := `.environment:
    .using: docker

.environment:
    .using: podman
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	errs := result.Errors
	if len(errs) != 1 {
		t.Fatalf("Expected 1 error, got %d: %v", len(errs), errs)
	}

	err := errs[0]
	if dup, ok := err.(*DuplicateDefinitionError); ok {
		if dup.Kind != "environment" {
			t.Errorf("Expected kind 'environment', got '%s'", dup.Kind)
		}
		if dup.Name != "(default)" {
			t.Errorf("Expected name '(default)', got '%s'", dup.Name)
		}
	} else {
		t.Errorf("Expected DuplicateDefinitionError, got %T", err)
	}
}

// TestCollector_ConditionalVariables tests that variables in conditionals are collected properly.
func TestCollector_ConditionalVariables(t *testing.T) {
	content := `if {os} == linux
cc = gcc
elif {os} == darwin
cc = clang
else
cc = cc
end
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	st, errs := result.SymbolTable, result.Errors
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	// cc should appear in both Variables (first definition) and ConditionalVars
	if len(st.Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(st.Variables))
	}

	// All definitions should be tracked in ConditionalVars
	if !st.IsConditionalVar("cc") {
		t.Error("Expected 'cc' to be tracked as conditional variable")
	}

	defs := st.GetConditionalVarDefs("cc")
	if len(defs) != 3 {
		t.Errorf("Expected 3 conditional definitions for 'cc', got %d", len(defs))
	}

	// Check branch types
	branches := make(map[string]bool)
	for _, def := range defs {
		branches[def.BranchType] = true
	}
	if !branches["if"] {
		t.Error("Expected 'if' branch")
	}
	if !branches["elif"] {
		t.Error("Expected 'elif' branch")
	}
	if !branches["else"] {
		t.Error("Expected 'else' branch")
	}
}

// TestCollector_NestedConditionals tests nested conditionals.
func TestCollector_NestedConditionals(t *testing.T) {
	content := `if {os} == linux
    ifdef DEBUG
        debug_flags = -g -O0
    end
    cc = gcc
end
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	st, errs := result.SymbolTable, result.Errors
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	// Both variables should be collected
	if !st.IsDefined("debug_flags") {
		t.Error("Expected 'debug_flags' to be defined")
	}
	if !st.IsDefined("cc") {
		t.Error("Expected 'cc' to be defined")
	}
}

// TestCollector_MultipleTargets tests collecting multiple targets.
func TestCollector_MultipleTargets(t *testing.T) {
	content := `@build:
    echo building

@test:
    echo testing

@clean:
    rm -rf build/
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	st, errs := result.SymbolTable, result.Errors
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	if len(st.Targets) != 3 {
		t.Errorf("Expected 3 targets, got %d", len(st.Targets))
	}
}

// TestCollector_PatternTargets tests that different pattern targets are allowed.
func TestCollector_PatternTargets(t *testing.T) {
	content := `build/{name}.o: src/{name}.c
    gcc -c {in} -o {out}

build/{name}.a: build/{name}.o
    ar rcs {out} {in}
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	st, errs := result.SymbolTable, result.Errors
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	if len(st.Targets) != 2 {
		t.Errorf("Expected 2 targets, got %d", len(st.Targets))
	}
}

// TestCollector_MultipleErrors tests that multiple errors are collected.
func TestCollector_MultipleErrors(t *testing.T) {
	content := `cc = gcc
cc = clang

@build:
    echo building

@build:
    echo again
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	errs := result.Errors
	if len(errs) != 2 {
		t.Fatalf("Expected 2 errors, got %d: %v", len(errs), errs)
	}
}

// TestCollector_LazyVariables tests that lazy variables are collected.
func TestCollector_LazyVariables(t *testing.T) {
	content := `lazy sources = shell(find src -name *.c)
objects = {sources}
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	st, errs := result.SymbolTable, result.Errors
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	if len(st.Variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(st.Variables))
	}

	if v := st.LookupVariable("sources"); v == nil {
		t.Error("Expected variable 'sources'")
	} else if !v.Lazy {
		t.Error("Expected 'sources' to be lazy")
	}
}

// TestCollector_EmptyNeedfile tests handling of empty needfile.
func TestCollector_EmptyNeedfile(t *testing.T) {
	content := ``
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	st, errs := result.SymbolTable, result.Errors
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	if len(st.Variables) != 0 {
		t.Errorf("Expected 0 variables, got %d", len(st.Variables))
	}
	if len(st.Targets) != 0 {
		t.Errorf("Expected 0 targets, got %d", len(st.Targets))
	}
	if len(st.Environments) != 0 {
		t.Errorf("Expected 0 environments, got %d", len(st.Environments))
	}
}

// TestCollector_CommentsAndBlanks tests that comments and blanks are ignored.
func TestCollector_CommentsAndBlanks(t *testing.T) {
	content := `# This is a comment
cc = gcc

# Another comment
cflags = -Wall

`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	st, errs := result.SymbolTable, result.Errors
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	if len(st.Variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(st.Variables))
	}
}

// TestCollector_GlobalDirectives tests that global directives are handled.
func TestCollector_GlobalDirectives(t *testing.T) {
	content := `.shell: bash
.parallel: 4
.default: @build

cc = gcc
@build:
    echo building
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	st, errs := result.SymbolTable, result.Errors
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	// Directives don't add to symbol table, only variables and targets do
	if len(st.Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(st.Variables))
	}
	if len(st.Targets) != 1 {
		t.Errorf("Expected 1 target, got %d", len(st.Targets))
	}
}

// TestCollector_PreservesOrder tests that targets are collected in definition order.
func TestCollector_PreservesOrder(t *testing.T) {
	content := `@first:
    echo first

@second:
    echo second

@third:
    echo third
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	st, errs := result.SymbolTable, result.Errors
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	if len(st.Targets) != 3 {
		t.Fatalf("Expected 3 targets, got %d", len(st.Targets))
	}

	patterns := []string{"first", "second", "third"}
	for i, expected := range patterns {
		actual := PatternString(&st.Targets[i].Pattern)
		if actual != expected {
			t.Errorf("Target %d: expected '%s', got '%s'", i, expected, actual)
		}
	}
}

// TestCollector_FileTargets tests file target collection.
func TestCollector_FileTargets(t *testing.T) {
	content := `build/app: build/main.o build/utils.o
    gcc -o {out} {deps}

build/main.o: src/main.c
    gcc -c {in} -o {out}
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	st, errs := result.SymbolTable, result.Errors
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	if len(st.Targets) != 2 {
		t.Errorf("Expected 2 targets, got %d", len(st.Targets))
	}
}

// TestCollector_DirectoryTargets tests directory target collection.
func TestCollector_DirectoryTargets(t *testing.T) {
	content := `build/:
    mkdir -p {target}
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	st, errs := result.SymbolTable, result.Errors
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	if len(st.Targets) != 1 {
		t.Fatalf("Expected 1 target, got %d", len(st.Targets))
	}

	if !st.Targets[0].Pattern.IsDirectory {
		t.Error("Expected target to be directory")
	}
}

// TestCollector_MixedEnvironments tests mixing named and default environments.
func TestCollector_MixedEnvironments(t *testing.T) {
	content := `.environment:
    .using: bare

.environment: ci
    .using: docker

.environment: dev
    .using: nix
`
	stmts := parseNeedfile(t, content)

	result := Collect(stmts)
	st, errs := result.SymbolTable, result.Errors
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	if len(st.Environments) != 3 {
		t.Errorf("Expected 3 environments, got %d", len(st.Environments))
	}

	// Check default environment exists
	if _, ok := st.Environments[""]; !ok {
		t.Error("Expected default environment")
	}

	// Check named environments exist
	if _, ok := st.Environments["ci"]; !ok {
		t.Error("Expected environment 'ci'")
	}
	if _, ok := st.Environments["dev"]; !ok {
		t.Error("Expected environment 'dev'")
	}
}
