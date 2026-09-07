package semantic

import (
	"testing"

	"github.com/vinayprograms/need/internal/ast"
)

// Helper to create a simple variable
func makeVar(name string) *ast.Variable {
	return &ast.Variable{
		Name:     name,
		Value:    &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "test"}}},
		Location: ast.SourceLocation{File: "test", Line: 1, Column: 1},
	}
}

// Helper to create a target with segments
func makeTarget(segments []ast.PatternSegment, deps [][]ast.PatternSegment) *ast.Target {
	target := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: segments,
		},
		Location: ast.SourceLocation{File: "test", Line: 1, Column: 1},
	}
	for _, depSegs := range deps {
		target.Dependencies = append(target.Dependencies, ast.Dependency{Segments: depSegs})
	}
	return target
}

// Helper to make a literal segment
func lit(text string) *ast.LiteralSegment {
	return &ast.LiteralSegment{Text: text}
}

// Helper to make a brace expr (unresolved)
func brace(name string) *ast.BraceExpr {
	return &ast.BraceExpr{
		Identifier: name,
		Location:   ast.SourceLocation{File: "test", Line: 1, Column: 1},
	}
}

// TestValidateCaptures_NoBraceExprs tests validation with no brace expressions.
func TestValidateCaptures_NoBraceExprs(t *testing.T) {
	st := NewSymbolTable()

	// Target: build/app: src/main.c
	target := makeTarget(
		[]ast.PatternSegment{lit("build/app")},
		[][]ast.PatternSegment{{lit("src/main.c")}},
	)
	_ = st.AddTarget(target)

	result := ValidateCaptures(st)

	if result.HasErrors() {
		t.Errorf("Expected no errors, got: %v", result.Errors)
	}
}

// TestValidateCaptures_SimpleCapture tests that undefined names become captures.
func TestValidateCaptures_SimpleCapture(t *testing.T) {
	st := NewSymbolTable()

	// Target: build/{name}.o: src/{name}.c
	target := makeTarget(
		[]ast.PatternSegment{lit("build/"), brace("name"), lit(".o")},
		[][]ast.PatternSegment{{lit("src/"), brace("name"), lit(".c")}},
	)
	_ = st.AddTarget(target)

	result := ValidateCaptures(st)

	if result.HasErrors() {
		t.Errorf("Expected no errors, got: %v", result.Errors)
	}

	// Verify the target now has a capture named "name"
	if len(result.Captures) != 1 {
		t.Fatalf("Expected 1 target with captures, got %d", len(result.Captures))
	}
	captureInfo := result.Captures[target]
	if captureInfo == nil {
		t.Fatal("Expected capture info for target")
	}
	if len(captureInfo.Names) != 1 || captureInfo.Names[0] != "name" {
		t.Errorf("Expected capture 'name', got %v", captureInfo.Names)
	}
}

// TestValidateCaptures_VariableInterpolation tests that defined variables become interpolations.
func TestValidateCaptures_VariableInterpolation(t *testing.T) {
	st := NewSymbolTable()

	// Variable: dir = build
	_ = st.AddVariable(makeVar("dir"))

	// Target: {dir}/app: src/main.c
	target := makeTarget(
		[]ast.PatternSegment{brace("dir"), lit("/app")},
		[][]ast.PatternSegment{{lit("src/main.c")}},
	)
	_ = st.AddTarget(target)

	result := ValidateCaptures(st)

	if result.HasErrors() {
		t.Errorf("Expected no errors, got: %v", result.Errors)
	}

	// Verify the target has no captures (dir is an interpolation)
	if len(result.Captures) != 0 {
		captureInfo := result.Captures[target]
		if captureInfo != nil && len(captureInfo.Names) > 0 {
			t.Errorf("Expected no captures, got: %v", captureInfo.Names)
		}
	}

	// Verify dir is resolved as interpolation
	if len(result.Interpolations) != 1 {
		t.Fatalf("Expected 1 target with interpolations, got %d", len(result.Interpolations))
	}
	interpInfo := result.Interpolations[target]
	if interpInfo == nil {
		t.Fatal("Expected interpolation info for target")
	}
	if len(interpInfo.Names) != 1 || interpInfo.Names[0] != "dir" {
		t.Errorf("Expected interpolation 'dir', got %v", interpInfo.Names)
	}
}

// TestValidateCaptures_AutomaticVariableError tests that automatic variables in patterns are errors.
func TestValidateCaptures_AutomaticVariableError(t *testing.T) {
	st := NewSymbolTable()

	// Target: build/{target}.o: src/main.c
	// "target" is an automatic variable, should error
	target := makeTarget(
		[]ast.PatternSegment{lit("build/"), brace("target"), lit(".o")},
		[][]ast.PatternSegment{{lit("src/main.c")}},
	)
	_ = st.AddTarget(target)

	result := ValidateCaptures(st)

	if !result.HasErrors() {
		t.Error("Expected error for automatic variable in pattern")
	}

	// Check error type
	foundErr := false
	for _, err := range result.Errors {
		if ae, ok := err.(*AutomaticInPatternError); ok {
			if ae.Name == "target" {
				foundErr = true
			}
		}
	}
	if !foundErr {
		t.Errorf("Expected AutomaticInPatternError for 'target', got: %v", result.Errors)
	}
}

// TestValidateCaptures_BuiltinInPattern tests built-in variables (os, arch) in patterns.
func TestValidateCaptures_BuiltinInPattern(t *testing.T) {
	st := NewSymbolTable()

	// Target: build/{os}/app: src/main.c
	// "os" is a built-in variable, should be treated as interpolation
	target := makeTarget(
		[]ast.PatternSegment{lit("build/"), brace("os"), lit("/app")},
		[][]ast.PatternSegment{{lit("src/main.c")}},
	)
	_ = st.AddTarget(target)

	result := ValidateCaptures(st)

	if result.HasErrors() {
		t.Errorf("Expected no errors for built-in in pattern, got: %v", result.Errors)
	}

	// Verify "os" is resolved as interpolation
	interpInfo := result.Interpolations[target]
	if interpInfo == nil || len(interpInfo.Names) != 1 || interpInfo.Names[0] != "os" {
		t.Errorf("Expected interpolation 'os', got: %v", result.Interpolations)
	}
}

// TestValidateCaptures_MultipleCaptures tests multiple captures in same pattern.
func TestValidateCaptures_MultipleCaptures(t *testing.T) {
	st := NewSymbolTable()

	// Target: build/{dir}/{name}.o: src/{dir}/{name}.c
	target := makeTarget(
		[]ast.PatternSegment{lit("build/"), brace("dir"), lit("/"), brace("name"), lit(".o")},
		[][]ast.PatternSegment{{lit("src/"), brace("dir"), lit("/"), brace("name"), lit(".c")}},
	)
	_ = st.AddTarget(target)

	result := ValidateCaptures(st)

	if result.HasErrors() {
		t.Errorf("Expected no errors, got: %v", result.Errors)
	}

	captureInfo := result.Captures[target]
	if captureInfo == nil || len(captureInfo.Names) != 2 {
		t.Fatalf("Expected 2 captures, got %v", captureInfo)
	}
	// Order should be preserved
	if captureInfo.Names[0] != "dir" || captureInfo.Names[1] != "name" {
		t.Errorf("Expected captures [dir, name], got %v", captureInfo.Names)
	}
}

// TestValidateCaptures_CaptureMismatch_MissingInDependency tests capture consistency.
func TestValidateCaptures_CaptureMismatch_MissingInDependency(t *testing.T) {
	st := NewSymbolTable()

	// Target: build/{name}.o: src/main.c
	// "name" capture in target but not in dependency
	// This is allowed - it means the dependency is literal
	target := makeTarget(
		[]ast.PatternSegment{lit("build/"), brace("name"), lit(".o")},
		[][]ast.PatternSegment{{lit("src/main.c")}},
	)
	_ = st.AddTarget(target)

	result := ValidateCaptures(st)

	// This should NOT be an error - a pattern target can have literal dependencies
	if result.HasErrors() {
		t.Errorf("Expected no errors for pattern target with literal deps, got: %v", result.Errors)
	}
}

// TestValidateCaptures_CaptureMismatch_ExtraInDependency tests extra capture in dependency.
func TestValidateCaptures_CaptureMismatch_ExtraInDependency(t *testing.T) {
	st := NewSymbolTable()

	// Target: build/app: src/{name}.c
	// "name" in dependency but not in target - this is an error
	target := makeTarget(
		[]ast.PatternSegment{lit("build/app")},
		[][]ast.PatternSegment{{lit("src/"), brace("name"), lit(".c")}},
	)
	_ = st.AddTarget(target)

	result := ValidateCaptures(st)

	if !result.HasErrors() {
		t.Error("Expected error for capture in dependency not in target")
	}

	// Check error type
	foundErr := false
	for _, err := range result.Errors {
		if ce, ok := err.(*CaptureMismatchError); ok {
			if ce.Name == "name" {
				foundErr = true
			}
		}
	}
	if !foundErr {
		t.Errorf("Expected CaptureMismatchError for 'name', got: %v", result.Errors)
	}
}

// TestValidateCaptures_CaptureAndInterpolationMixed tests mix of captures and interpolations.
func TestValidateCaptures_CaptureAndInterpolationMixed(t *testing.T) {
	st := NewSymbolTable()

	// Variable: base = build
	_ = st.AddVariable(makeVar("base"))

	// Target: {base}/{name}.o: src/{name}.c
	// "base" is a variable (interpolation), "name" is a capture
	target := makeTarget(
		[]ast.PatternSegment{brace("base"), lit("/"), brace("name"), lit(".o")},
		[][]ast.PatternSegment{{lit("src/"), brace("name"), lit(".c")}},
	)
	_ = st.AddTarget(target)

	result := ValidateCaptures(st)

	if result.HasErrors() {
		t.Errorf("Expected no errors, got: %v", result.Errors)
	}

	// Verify "base" is interpolation
	interpInfo := result.Interpolations[target]
	if interpInfo == nil || len(interpInfo.Names) != 1 || interpInfo.Names[0] != "base" {
		t.Errorf("Expected interpolation 'base', got: %v", interpInfo)
	}

	// Verify "name" is capture
	captureInfo := result.Captures[target]
	if captureInfo == nil || len(captureInfo.Names) != 1 || captureInfo.Names[0] != "name" {
		t.Errorf("Expected capture 'name', got: %v", captureInfo)
	}
}

// TestValidateCaptures_PhonyTarget tests phony targets (no captures allowed).
func TestValidateCaptures_PhonyTarget(t *testing.T) {
	st := NewSymbolTable()

	// Target: @clean: (phony target)
	target := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{lit("clean")},
			IsPhony:  true,
		},
		Location: ast.SourceLocation{File: "test", Line: 1, Column: 1},
	}
	_ = st.AddTarget(target)

	result := ValidateCaptures(st)

	if result.HasErrors() {
		t.Errorf("Expected no errors for phony target, got: %v", result.Errors)
	}
}

// TestValidateCaptures_PhonyTargetWithBraceExpr tests phony targets with brace expr.
func TestValidateCaptures_PhonyTargetWithBraceExpr(t *testing.T) {
	st := NewSymbolTable()

	// Variable: action = test
	_ = st.AddVariable(makeVar("action"))

	// Target: @{action}: (phony target with interpolation)
	target := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{brace("action")},
			IsPhony:  true,
		},
		Location: ast.SourceLocation{File: "test", Line: 1, Column: 1},
	}
	_ = st.AddTarget(target)

	result := ValidateCaptures(st)

	// This should be valid - the variable interpolation will resolve to a phony name
	if result.HasErrors() {
		t.Errorf("Expected no errors for phony with variable, got: %v", result.Errors)
	}
}

// TestValidateCaptures_PhonyTargetWithCapture tests phony targets with unresolved name.
func TestValidateCaptures_PhonyTargetWithCapture(t *testing.T) {
	st := NewSymbolTable()

	// Target: @{name}: (phony target with capture - probably an error)
	target := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{brace("name")},
			IsPhony:  true,
		},
		Location: ast.SourceLocation{File: "test", Line: 1, Column: 1},
	}
	_ = st.AddTarget(target)

	result := ValidateCaptures(st)

	// Phony targets with captures are unusual but allowed
	// The pattern matching will apply at runtime
	if result.HasErrors() {
		t.Errorf("Expected no errors for phony with capture, got: %v", result.Errors)
	}
}

// TestValidateCaptures_DuplicateCaptureInPattern tests duplicate capture names.
func TestValidateCaptures_DuplicateCaptureInPattern(t *testing.T) {
	st := NewSymbolTable()

	// Target: build/{name}/{name}.o: src/{name}.c
	// Duplicate "name" in pattern - should be valid (same capture)
	target := makeTarget(
		[]ast.PatternSegment{lit("build/"), brace("name"), lit("/"), brace("name"), lit(".o")},
		[][]ast.PatternSegment{{lit("src/"), brace("name"), lit(".c")}},
	)
	_ = st.AddTarget(target)

	result := ValidateCaptures(st)

	if result.HasErrors() {
		t.Errorf("Expected no errors for duplicate capture, got: %v", result.Errors)
	}

	// Verify only one unique capture name
	captureInfo := result.Captures[target]
	if captureInfo == nil || len(captureInfo.Names) != 1 {
		t.Errorf("Expected 1 unique capture, got %v", captureInfo)
	}
}

// TestValidateCaptures_ConditionalVariable tests conditional variable recognition.
func TestValidateCaptures_ConditionalVariable(t *testing.T) {
	st := NewSymbolTable()

	// Conditional variable: cc (defined somewhere in a conditional)
	condVar := makeVar("cc")
	def := &ConditionalVarDef{
		Variable:    condVar,
		Conditional: &ast.Conditional{},
		BranchType:  "if",
		BranchIndex: -1,
	}
	st.AddConditionalVariable(def)
	_ = st.AddVariable(condVar)

	// Target: {cc}/app: src/main.c
	target := makeTarget(
		[]ast.PatternSegment{brace("cc"), lit("/app")},
		[][]ast.PatternSegment{{lit("src/main.c")}},
	)
	_ = st.AddTarget(target)

	result := ValidateCaptures(st)

	if result.HasErrors() {
		t.Errorf("Expected no errors, got: %v", result.Errors)
	}

	// Verify "cc" is resolved as interpolation (it's a defined variable)
	interpInfo := result.Interpolations[target]
	if interpInfo == nil || len(interpInfo.Names) != 1 || interpInfo.Names[0] != "cc" {
		t.Errorf("Expected interpolation 'cc', got: %v", interpInfo)
	}
}

// TestAutomaticInPatternError_Error tests error message format.
func TestAutomaticInPatternError_Error(t *testing.T) {
	err := &AutomaticInPatternError{
		Name:     "target",
		Location: ast.SourceLocation{File: "Needfile", Line: 5, Column: 10},
	}

	expected := "automatic variable 'target' cannot be used as capture in target pattern at Needfile:5:10"
	if err.Error() != expected {
		t.Errorf("Expected error: %s\nGot: %s", expected, err.Error())
	}
}

// TestCaptureMismatchError_Error tests error message format.
func TestCaptureMismatchError_Error(t *testing.T) {
	err := &CaptureMismatchError{
		Name:      "name",
		InTarget:  false,
		Location:  ast.SourceLocation{File: "Needfile", Line: 3, Column: 20},
		TargetLoc: ast.SourceLocation{File: "Needfile", Line: 3, Column: 1},
	}

	msg := err.Error()
	if msg == "" {
		t.Error("Expected non-empty error message")
	}
	if !containsSubstring(msg, "name") {
		t.Errorf("Expected error to mention capture name, got: %s", msg)
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || containsSubstring(s[1:], substr)))
}
