package semantic

import (
	"strings"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

// depTarget creates a target using literal pattern and dependency strings.
// Helper specific to dependency graph tests.
func depTarget(pattern string, deps ...string) *ast.Target {
	isPhony := strings.HasPrefix(pattern, "@")
	name := pattern
	if isPhony {
		name = pattern[1:] // Strip @ for storage
	}

	t := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: name},
			},
			IsPhony:     isPhony,
			IsDirectory: strings.HasSuffix(name, "/"),
		},
		Location: ast.SourceLocation{File: "test", Line: 1, Column: 1},
	}
	for _, d := range deps {
		t.Dependencies = append(t.Dependencies, ast.Dependency{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: d},
			},
		})
	}
	return t
}

// depPatternTarget creates a pattern target with brace expressions.
func depPatternTarget(pattern string, deps ...string) *ast.Target {
	segments := parsePatternString(pattern)

	t := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: segments,
		},
		Location: ast.SourceLocation{File: "test", Line: 1, Column: 1},
	}

	for _, d := range deps {
		t.Dependencies = append(t.Dependencies, ast.Dependency{
			Segments: parsePatternString(d),
		})
	}
	return t
}

// parsePatternString parses a pattern string into segments.
// E.g., "build/{name}.o" -> [LiteralSegment("build/"), BraceExpr("name"), LiteralSegment(".o")]
func parsePatternString(pattern string) []ast.PatternSegment {
	segments := []ast.PatternSegment{}
	remaining := pattern
	for len(remaining) > 0 {
		start := strings.Index(remaining, "{")
		if start == -1 {
			segments = append(segments, &ast.LiteralSegment{Text: remaining})
			break
		}
		if start > 0 {
			segments = append(segments, &ast.LiteralSegment{Text: remaining[:start]})
		}
		end := strings.Index(remaining[start:], "}")
		if end == -1 {
			segments = append(segments, &ast.LiteralSegment{Text: remaining})
			break
		}
		name := remaining[start+1 : start+end]
		segments = append(segments, &ast.BraceExpr{
			Identifier: name,
			Location:   ast.SourceLocation{File: "test", Line: 1, Column: start + 1},
		})
		remaining = remaining[start+end+1:]
	}
	return segments
}

// ===========================================================================
// Basic Graph Construction Tests
// ===========================================================================

func TestValidateDependencies_NoDependencies(t *testing.T) {
	// Targets with no dependencies should pass
	targets := []*ast.Target{
		depTarget("build/app"),
		depTarget("build/lib"),
	}

	result := ValidateDependencies(targets)
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestValidateDependencies_SimpleDependency(t *testing.T) {
	// A depends on B, B has no deps
	targets := []*ast.Target{
		depTarget("build/app", "build/lib.o"),
		depTarget("build/lib.o"),
	}

	result := ValidateDependencies(targets)
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestValidateDependencies_ChainDependencies(t *testing.T) {
	// A -> B -> C -> D (linear chain)
	targets := []*ast.Target{
		depTarget("a", "b"),
		depTarget("b", "c"),
		depTarget("c", "d"),
		depTarget("d"),
	}

	result := ValidateDependencies(targets)
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestValidateDependencies_DiamondDependencies(t *testing.T) {
	// Diamond: A -> B, A -> C, B -> D, C -> D
	targets := []*ast.Target{
		depTarget("a", "b", "c"),
		depTarget("b", "d"),
		depTarget("c", "d"),
		depTarget("d"),
	}

	result := ValidateDependencies(targets)
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

// ===========================================================================
// Circular Dependency Detection Tests
// ===========================================================================

func TestValidateDependencies_DirectCycle(t *testing.T) {
	// A -> A (self-loop)
	targets := []*ast.Target{
		depTarget("a", "a"),
	}

	result := ValidateDependencies(targets)
	if len(result.Errors) == 0 {
		t.Error("expected circular dependency error, got none")
		return
	}

	err, ok := result.Errors[0].(*CircularDependencyError)
	if !ok {
		t.Errorf("expected CircularDependencyError, got %T", result.Errors[0])
		return
	}

	if len(err.Cycle) < 2 {
		t.Errorf("expected cycle with at least 2 elements, got %v", err.Cycle)
	}
}

func TestValidateDependencies_TwoNodeCycle(t *testing.T) {
	// A -> B -> A
	targets := []*ast.Target{
		depTarget("a", "b"),
		depTarget("b", "a"),
	}

	result := ValidateDependencies(targets)
	if len(result.Errors) == 0 {
		t.Error("expected circular dependency error, got none")
		return
	}

	err, ok := result.Errors[0].(*CircularDependencyError)
	if !ok {
		t.Errorf("expected CircularDependencyError, got %T", result.Errors[0])
		return
	}

	// Should report cycle: a -> b -> a or b -> a -> b
	if len(err.Cycle) < 3 {
		t.Errorf("expected cycle with at least 3 elements (to show full loop), got %v", err.Cycle)
	}
}

func TestValidateDependencies_ThreeNodeCycle(t *testing.T) {
	// A -> B -> C -> A
	targets := []*ast.Target{
		depTarget("a", "b"),
		depTarget("b", "c"),
		depTarget("c", "a"),
	}

	result := ValidateDependencies(targets)
	if len(result.Errors) == 0 {
		t.Error("expected circular dependency error, got none")
		return
	}

	err, ok := result.Errors[0].(*CircularDependencyError)
	if !ok {
		t.Errorf("expected CircularDependencyError, got %T", result.Errors[0])
		return
	}

	// Should report the full cycle
	if len(err.Cycle) < 4 {
		t.Errorf("expected cycle with at least 4 elements, got %v", err.Cycle)
	}
}

func TestValidateDependencies_CycleInSubgraph(t *testing.T) {
	// D is fine, but B -> C -> B forms a cycle
	// A -> B -> C -> B (cycle), D
	targets := []*ast.Target{
		depTarget("a", "b", "d"),
		depTarget("b", "c"),
		depTarget("c", "b"),
		depTarget("d"),
	}

	result := ValidateDependencies(targets)
	if len(result.Errors) == 0 {
		t.Error("expected circular dependency error, got none")
	}

	err, ok := result.Errors[0].(*CircularDependencyError)
	if !ok {
		t.Errorf("expected CircularDependencyError, got %T", result.Errors[0])
	}

	// Cycle should be b -> c -> b
	cycleStr := err.Error()
	if !strings.Contains(cycleStr, "b") || !strings.Contains(cycleStr, "c") {
		t.Errorf("expected cycle to contain b and c, got: %s", cycleStr)
	}
}

// ===========================================================================
// Unsatisfied Dependency Tests
// ===========================================================================

func TestValidateDependencies_UnsatisfiedDependency(t *testing.T) {
	// A depends on B, but B is not defined and not a source file
	// Note: In semantic analysis, we can't check if files exist,
	// so we track dependencies that aren't targets
	targets := []*ast.Target{
		depTarget("a", "b"),
		// b is not defined
	}

	result := ValidateDependencies(targets)

	// We should track that 'b' is an unsatisfied dependency
	// It will be either a source file or need to be built - tracked for build planning
	if result.UnsatisfiedDeps == nil {
		t.Error("expected UnsatisfiedDeps to be populated")
		return
	}

	deps, ok := result.UnsatisfiedDeps["a"]
	if !ok {
		t.Error("expected 'a' to have unsatisfied dependencies")
		return
	}

	found := false
	for _, d := range deps {
		if d == "b" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'b' to be in unsatisfied deps for 'a', got %v", deps)
	}
}

func TestValidateDependencies_MultipleUnsatisfiedDependencies(t *testing.T) {
	// A depends on b, c, d but only b is defined
	targets := []*ast.Target{
		depTarget("a", "b", "c", "d"),
		depTarget("b"),
		// c and d are not defined
	}

	result := ValidateDependencies(targets)

	if result.UnsatisfiedDeps == nil {
		t.Error("expected UnsatisfiedDeps to be populated")
		return
	}

	deps, ok := result.UnsatisfiedDeps["a"]
	if !ok {
		t.Error("expected 'a' to have unsatisfied dependencies")
		return
	}

	// Should have c and d as unsatisfied
	if len(deps) != 2 {
		t.Errorf("expected 2 unsatisfied deps, got %d: %v", len(deps), deps)
	}
}

// ===========================================================================
// Phony Target Tests
// ===========================================================================

func TestValidateDependencies_PhonyTargets(t *testing.T) {
	// Phony targets should participate in dependency graph
	targets := []*ast.Target{
		depTarget("@all", "build/app"),
		depTarget("build/app", "build/main.o"),
		depTarget("build/main.o"),
	}

	result := ValidateDependencies(targets)
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestValidateDependencies_PhonyCycle(t *testing.T) {
	// @clean -> @build -> @clean
	targets := []*ast.Target{
		depTarget("@clean", "build"),
		depTarget("@build", "clean"),
	}

	result := ValidateDependencies(targets)
	if len(result.Errors) == 0 {
		t.Error("expected circular dependency error for phony targets")
	}

	_, ok := result.Errors[0].(*CircularDependencyError)
	if !ok {
		t.Errorf("expected CircularDependencyError, got %T", result.Errors[0])
	}
}

// ===========================================================================
// Pattern Target Tests
// ===========================================================================

func TestValidateDependencies_PatternTargets(t *testing.T) {
	// Pattern targets: build/{name}.o: src/{name}.c
	// These define a rule, not a concrete dependency
	targets := []*ast.Target{
		depPatternTarget("build/{name}.o", "src/{name}.c"),
	}

	result := ValidateDependencies(targets)

	// Pattern targets should be recorded separately
	if len(result.PatternTargets) != 1 {
		t.Errorf("expected 1 pattern target, got %d", len(result.PatternTargets))
	}
}

func TestValidateDependencies_MixedPatternAndLiteral(t *testing.T) {
	// Mix of literal and pattern targets
	targets := []*ast.Target{
		depTarget("@all", "build/app"),
		depTarget("build/app", "build/main.o", "build/utils.o"),
		depPatternTarget("build/{name}.o", "src/{name}.c"),
	}

	result := ValidateDependencies(targets)

	// build/main.o and build/utils.o should be potentially satisfiable by pattern
	// but we can't verify at semantic analysis time - track for planning
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	// Pattern targets should be tracked
	if len(result.PatternTargets) != 1 {
		t.Errorf("expected 1 pattern target, got %d", len(result.PatternTargets))
	}
}

// ===========================================================================
// Error Message Tests
// ===========================================================================

func TestCircularDependencyError_Error(t *testing.T) {
	err := &CircularDependencyError{
		Cycle: []string{"a", "b", "c", "a"},
	}

	msg := err.Error()
	if !strings.Contains(msg, "circular dependency") {
		t.Errorf("expected error message to contain 'circular dependency', got: %s", msg)
	}
	if !strings.Contains(msg, "a -> b -> c -> a") {
		t.Errorf("expected error message to show cycle path, got: %s", msg)
	}
}

// ===========================================================================
// Graph Structure Tests
// ===========================================================================

func TestDependencyResult_Graph(t *testing.T) {
	// Verify the graph structure is built correctly
	targets := []*ast.Target{
		depTarget("a", "b", "c"),
		depTarget("b", "d"),
		depTarget("c", "d"),
		depTarget("d"),
	}

	result := ValidateDependencies(targets)

	// Check graph has correct edges
	if result.Graph == nil {
		t.Fatal("expected graph to be populated")
	}

	// Check edges for 'a'
	aEdges := result.Graph.Edges["a"]
	if len(aEdges) != 2 {
		t.Errorf("expected 'a' to have 2 edges, got %d", len(aEdges))
	}

	// Check edges for 'b'
	bEdges := result.Graph.Edges["b"]
	if len(bEdges) != 1 {
		t.Errorf("expected 'b' to have 1 edge, got %d", len(bEdges))
	}

	// Check 'd' has no outgoing edges
	dEdges := result.Graph.Edges["d"]
	if len(dEdges) != 0 {
		t.Errorf("expected 'd' to have 0 edges, got %d", len(dEdges))
	}
}

// ===========================================================================
// Edge Cases
// ===========================================================================

func TestValidateDependencies_EmptyTargets(t *testing.T) {
	result := ValidateDependencies(nil)
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors for empty targets, got %d", len(result.Errors))
	}
}

func TestValidateDependencies_SingleTargetNoDeps(t *testing.T) {
	targets := []*ast.Target{
		depTarget("app"),
	}

	result := ValidateDependencies(targets)
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestValidateDependencies_MultipleCycles(t *testing.T) {
	// Two separate cycles: a -> b -> a, c -> d -> c
	targets := []*ast.Target{
		depTarget("a", "b"),
		depTarget("b", "a"),
		depTarget("c", "d"),
		depTarget("d", "c"),
	}

	result := ValidateDependencies(targets)

	// Should detect at least one cycle
	if len(result.Errors) == 0 {
		t.Error("expected circular dependency errors, got none")
	}
}

func TestValidateDependencies_LongCycle(t *testing.T) {
	// Long cycle: a -> b -> c -> d -> e -> f -> a
	targets := []*ast.Target{
		depTarget("a", "b"),
		depTarget("b", "c"),
		depTarget("c", "d"),
		depTarget("d", "e"),
		depTarget("e", "f"),
		depTarget("f", "a"),
	}

	result := ValidateDependencies(targets)
	if len(result.Errors) == 0 {
		t.Error("expected circular dependency error, got none")
		return
	}

	err, ok := result.Errors[0].(*CircularDependencyError)
	if !ok {
		t.Errorf("expected CircularDependencyError, got %T", result.Errors[0])
		return
	}

	// Should report the full cycle
	if len(err.Cycle) < 7 {
		t.Errorf("expected cycle with 7 elements, got %d: %v", len(err.Cycle), err.Cycle)
	}
}
