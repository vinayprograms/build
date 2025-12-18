package planner

import (
	"testing"
	"time"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/eval"
)

// ----------------------------------------------------------------------------
// BuildReason Tests
// ----------------------------------------------------------------------------

func TestBuildReason_String(t *testing.T) {
	testCases := []struct {
		reason   BuildReason
		expected string
	}{
		{BuildReasonTargetMissing, "target missing"},
		{BuildReasonDependencyNewer, "dependency newer than target"},
		{BuildReasonPhonyTarget, "phony target"},
		{BuildReasonForcedRebuild, "forced rebuild"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if tc.reason.String() != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, tc.reason.String())
			}
		})
	}
}

// ----------------------------------------------------------------------------
// BuildTask Tests
// ----------------------------------------------------------------------------

func TestBuildTask_Structure(t *testing.T) {
	task := BuildTask{
		Target:       "build/app",
		Dependencies: []string{"src/main.c", "src/utils.c"},
		Recipe:       nil,
		Reason:       BuildReasonTargetMissing,
		Captures:     map[string]string{"name": "app"},
	}

	if task.Target != "build/app" {
		t.Errorf("expected target 'build/app', got %q", task.Target)
	}
	if len(task.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(task.Dependencies))
	}
	if task.Reason != BuildReasonTargetMissing {
		t.Errorf("expected reason TargetMissing, got %v", task.Reason)
	}
}

// ----------------------------------------------------------------------------
// BuildPlan Tests
// ----------------------------------------------------------------------------

func TestBuildPlan_Empty(t *testing.T) {
	plan := &BuildPlan{}

	if len(plan.Tasks) != 0 {
		t.Errorf("expected empty plan, got %d tasks", len(plan.Tasks))
	}
}

func TestBuildPlan_AddTask(t *testing.T) {
	plan := &BuildPlan{}
	task := BuildTask{Target: "build/app"}

	plan.Tasks = append(plan.Tasks, task)

	if len(plan.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(plan.Tasks))
	}
}

// ----------------------------------------------------------------------------
// PlanBuild Tests - Core Functionality
// ----------------------------------------------------------------------------

func TestPlanBuild_SingleTarget_NoDeps(t *testing.T) {
	targets := []*ast.Target{
		createTargetWithRecipe("build/app", false, false, nil, &ast.Recipe{}),
	}
	ctx := eval.NewContext()
	fs := &mockFileSystem{
		missing: map[string]bool{"build/app": true},
	}

	plan, err := PlanBuild("build/app", targets, ctx, fs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(plan.Tasks))
	}
	if plan.Tasks[0].Target != "build/app" {
		t.Errorf("expected target 'build/app', got %q", plan.Tasks[0].Target)
	}
	if plan.Tasks[0].Reason != BuildReasonTargetMissing {
		t.Errorf("expected reason TargetMissing, got %v", plan.Tasks[0].Reason)
	}
}

func TestPlanBuild_SingleTarget_WithDeps(t *testing.T) {
	targets := []*ast.Target{
		createTargetWithRecipe("build/app", false, false, []string{"src/main.c"}, &ast.Recipe{}),
	}
	ctx := eval.NewContext()
	fs := &mockFileSystem{
		missing: map[string]bool{"build/app": true},
		exists:  map[string]bool{"src/main.c": true},
	}

	plan, err := PlanBuild("build/app", targets, ctx, fs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(plan.Tasks))
	}
	task := plan.Tasks[0]
	if len(task.Dependencies) != 1 || task.Dependencies[0] != "src/main.c" {
		t.Errorf("expected dependency 'src/main.c', got %v", task.Dependencies)
	}
}

func TestPlanBuild_ChainedDependencies(t *testing.T) {
	// build/app depends on build/main.o
	// build/main.o depends on src/main.c (source file)
	targets := []*ast.Target{
		createTargetWithRecipe("build/app", false, false, []string{"build/main.o"}, &ast.Recipe{}),
		createTargetWithRecipe("build/main.o", false, false, []string{"src/main.c"}, &ast.Recipe{}),
	}
	ctx := eval.NewContext()
	fs := &mockFileSystem{
		missing: map[string]bool{
			"build/app":    true,
			"build/main.o": true,
		},
		exists: map[string]bool{"src/main.c": true},
	}

	plan, err := PlanBuild("build/app", targets, ctx, fs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(plan.Tasks))
	}
	// Topologically sorted: build/main.o first, then build/app
	if plan.Tasks[0].Target != "build/main.o" {
		t.Errorf("expected first task 'build/main.o', got %q", plan.Tasks[0].Target)
	}
	if plan.Tasks[1].Target != "build/app" {
		t.Errorf("expected second task 'build/app', got %q", plan.Tasks[1].Target)
	}
}

func TestPlanBuild_DiamondDependency(t *testing.T) {
	// build/app depends on build/a.o and build/b.o
	// Both depend on src/common.c
	targets := []*ast.Target{
		createTargetWithRecipe("build/app", false, false, []string{"build/a.o", "build/b.o"}, &ast.Recipe{}),
		createTargetWithRecipe("build/a.o", false, false, []string{"src/common.c"}, &ast.Recipe{}),
		createTargetWithRecipe("build/b.o", false, false, []string{"src/common.c"}, &ast.Recipe{}),
	}
	ctx := eval.NewContext()
	fs := &mockFileSystem{
		missing: map[string]bool{
			"build/app": true,
			"build/a.o": true,
			"build/b.o": true,
		},
		exists: map[string]bool{"src/common.c": true},
	}

	plan, err := PlanBuild("build/app", targets, ctx, fs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(plan.Tasks))
	}
	// build/app must come last
	if plan.Tasks[2].Target != "build/app" {
		t.Errorf("expected last task 'build/app', got %q", plan.Tasks[2].Target)
	}
}

func TestPlanBuild_PatternTarget(t *testing.T) {
	// Pattern: build/{name}.o depends on src/{name}.c
	targets := []*ast.Target{
		createPatternTargetWithDeps(
			[]patternPart{{literal: "build/"}, {capture: "name"}, {literal: ".o"}},
			[][]patternPart{
				{{literal: "src/"}, {capture: "name"}, {literal: ".c"}},
			},
			false, false,
		),
	}
	ctx := eval.NewContext()
	fs := &mockFileSystem{
		missing: map[string]bool{"build/main.o": true},
		exists:  map[string]bool{"src/main.c": true},
	}

	plan, err := PlanBuild("build/main.o", targets, ctx, fs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(plan.Tasks))
	}
	task := plan.Tasks[0]
	if task.Target != "build/main.o" {
		t.Errorf("expected target 'build/main.o', got %q", task.Target)
	}
	if task.Captures["name"] != "main" {
		t.Errorf("expected capture name='main', got %q", task.Captures["name"])
	}
	if len(task.Dependencies) != 1 || task.Dependencies[0] != "src/main.c" {
		t.Errorf("expected dependency 'src/main.c', got %v", task.Dependencies)
	}
}

func TestPlanBuild_PhonyTarget(t *testing.T) {
	targets := []*ast.Target{
		createTargetWithRecipe("all", true, false, []string{"build/app"}, &ast.Recipe{}),
		createTargetWithRecipe("build/app", false, false, nil, &ast.Recipe{}),
	}
	ctx := eval.NewContext()
	fs := &mockFileSystem{
		missing: map[string]bool{"build/app": true},
	}

	plan, err := PlanBuild("@all", targets, ctx, fs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(plan.Tasks))
	}
	// @all is phony, should always be included
	var foundPhony bool
	for _, task := range plan.Tasks {
		if task.Target == "@all" {
			foundPhony = true
			if task.Reason != BuildReasonPhonyTarget {
				t.Errorf("expected reason PhonyTarget for @all, got %v", task.Reason)
			}
		}
	}
	if !foundPhony {
		t.Error("expected @all task in plan")
	}
}

func TestPlanBuild_TargetUpToDate(t *testing.T) {
	targets := []*ast.Target{
		createTargetWithRecipe("build/app", false, false, []string{"src/main.c"}, &ast.Recipe{}),
	}
	ctx := eval.NewContext()
	now := time.Now()
	fs := &mockFileSystem{
		exists: map[string]bool{
			"build/app":  true,
			"src/main.c": true,
		},
		mtimes: map[string]time.Time{
			"build/app":  now,                     // target is newer
			"src/main.c": now.Add(-1 * time.Hour), // dependency is older
		},
	}

	plan, err := PlanBuild("build/app", targets, ctx, fs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Target is up to date, no tasks needed
	if len(plan.Tasks) != 0 {
		t.Errorf("expected 0 tasks (up to date), got %d", len(plan.Tasks))
	}
}

func TestPlanBuild_DependencyNewer(t *testing.T) {
	targets := []*ast.Target{
		createTargetWithRecipe("build/app", false, false, []string{"src/main.c"}, &ast.Recipe{}),
	}
	ctx := eval.NewContext()
	now := time.Now()
	fs := &mockFileSystem{
		exists: map[string]bool{
			"build/app":  true,
			"src/main.c": true,
		},
		mtimes: map[string]time.Time{
			"build/app":  now.Add(-1 * time.Hour), // target is older
			"src/main.c": now,                     // dependency is newer
		},
	}

	plan, err := PlanBuild("build/app", targets, ctx, fs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(plan.Tasks))
	}
	if plan.Tasks[0].Reason != BuildReasonDependencyNewer {
		t.Errorf("expected reason DependencyNewer, got %v", plan.Tasks[0].Reason)
	}
}

func TestPlanBuild_TargetNotFound(t *testing.T) {
	targets := []*ast.Target{
		createTargetWithRecipe("build/app", false, false, nil, &ast.Recipe{}),
	}
	ctx := eval.NewContext()
	fs := &mockFileSystem{}

	_, err := PlanBuild("build/other", targets, ctx, fs)

	if err == nil {
		t.Error("expected error for unknown target")
	}
}

func TestPlanBuild_CircularDependency(t *testing.T) {
	// a depends on b, b depends on a
	targets := []*ast.Target{
		createTargetWithRecipe("a", false, false, []string{"b"}, &ast.Recipe{}),
		createTargetWithRecipe("b", false, false, []string{"a"}, &ast.Recipe{}),
	}
	ctx := eval.NewContext()
	fs := &mockFileSystem{
		missing: map[string]bool{"a": true, "b": true},
	}

	_, err := PlanBuild("a", targets, ctx, fs)

	if err == nil {
		t.Error("expected error for circular dependency")
	}
	if _, ok := err.(*CircularDependencyError); !ok {
		t.Errorf("expected CircularDependencyError, got %T: %v", err, err)
	}
}

func TestPlanBuild_SourceFileMissing(t *testing.T) {
	targets := []*ast.Target{
		createTargetWithRecipe("build/app", false, false, []string{"src/main.c"}, &ast.Recipe{}),
	}
	ctx := eval.NewContext()
	fs := &mockFileSystem{
		missing: map[string]bool{
			"build/app":  true,
			"src/main.c": true, // source file doesn't exist
		},
	}

	_, err := PlanBuild("build/app", targets, ctx, fs)

	if err == nil {
		t.Error("expected error for missing source file")
	}
}

func TestPlanBuild_NoRecipe(t *testing.T) {
	// Target without recipe, all dependencies are source files
	targets := []*ast.Target{
		createTargetWithRecipe("all", true, false, []string{"src/main.c"}, nil), // No recipe
	}
	ctx := eval.NewContext()
	fs := &mockFileSystem{
		exists: map[string]bool{"src/main.c": true},
	}

	plan, err := PlanBuild("@all", targets, ctx, fs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Phony targets are always in the plan
	if len(plan.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(plan.Tasks))
	}
}

// ----------------------------------------------------------------------------
// CircularDependencyError Tests
// ----------------------------------------------------------------------------

func TestCircularDependencyError_Error(t *testing.T) {
	err := &CircularDependencyError{
		Path: []string{"a", "b", "c", "a"},
	}

	expected := "circular dependency detected: a -> b -> c -> a"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

// ----------------------------------------------------------------------------
// Mock FileSystem
// ----------------------------------------------------------------------------

type mockFileSystem struct {
	exists  map[string]bool
	missing map[string]bool
	mtimes  map[string]time.Time
}

func (m *mockFileSystem) Exists(path string) bool {
	if m.missing != nil && m.missing[path] {
		return false
	}
	if m.exists != nil {
		return m.exists[path]
	}
	return false
}

func (m *mockFileSystem) ModTime(path string) (time.Time, error) {
	if m.mtimes != nil {
		if mtime, ok := m.mtimes[path]; ok {
			return mtime, nil
		}
	}
	return time.Time{}, nil
}

// ----------------------------------------------------------------------------
// Test Helpers
// ----------------------------------------------------------------------------

// createTargetWithRecipe creates a target with dependencies and a recipe.
func createTargetWithRecipe(path string, isPhony, isDirectory bool, deps []string, recipe *ast.Recipe) *ast.Target {
	target := &ast.Target{
		Pattern: *createLiteralPattern(path, isPhony, isDirectory),
		Recipe:  recipe,
	}
	for _, dep := range deps {
		target.Dependencies = append(target.Dependencies, ast.Dependency{
			Segments: []ast.PatternSegment{&ast.LiteralSegment{Text: dep}},
		})
	}
	return target
}

// createPatternTargetWithDeps creates a pattern target with pattern dependencies.
func createPatternTargetWithDeps(pattern []patternPart, deps [][]patternPart, isPhony, isDirectory bool) *ast.Target {
	target := &ast.Target{
		Pattern: *createCapturePattern(pattern, isPhony, isDirectory),
		Recipe:  &ast.Recipe{},
	}
	for _, depParts := range deps {
		segments := make([]ast.PatternSegment, 0, len(depParts))
		for _, p := range depParts {
			if p.capture != "" {
				segments = append(segments, &ast.BraceExpr{Identifier: p.capture})
			} else {
				segments = append(segments, &ast.LiteralSegment{Text: p.literal})
			}
		}
		target.Dependencies = append(target.Dependencies, ast.Dependency{Segments: segments})
	}
	return target
}

// createTargetWithOrderDeps creates a target with order-only dependencies.
func createTargetWithOrderDeps(path string, isPhony, isDirectory bool, deps []string, orderDeps []string) *ast.Target {
	// Create .after directives for order-only dependencies
	after := make([]*ast.Value, 0, len(orderDeps))
	for _, od := range orderDeps {
		after = append(after, &ast.Value{
			Parts: []ast.ValuePart{
				&ast.LiteralValue{Text: od},
			},
		})
	}

	recipe := &ast.Recipe{
		Directives: ast.RecipeDirectives{
			After: after,
		},
	}

	target := &ast.Target{
		Pattern: *createLiteralPattern(path, isPhony, isDirectory),
		Recipe:  recipe,
	}
	for _, dep := range deps {
		target.Dependencies = append(target.Dependencies, ast.Dependency{
			Segments: []ast.PatternSegment{&ast.LiteralSegment{Text: dep}},
		})
	}
	return target
}

// ----------------------------------------------------------------------------
// Order-Only Dependency Tests
// ----------------------------------------------------------------------------

func TestPlanBuild_OrderOnlyDeps_BuildOrder(t *testing.T) {
	// build/app has order-only dep on build/ (directory must exist)
	// build/ has no deps
	targets := []*ast.Target{
		createTargetWithOrderDeps("build/app", false, false, []string{"src/main.c"}, []string{"build/"}),
		createTargetWithRecipe("build/", false, true, nil, &ast.Recipe{}),
	}
	ctx := eval.NewContext()
	fs := &mockFileSystem{
		missing: map[string]bool{
			"build/app": true,
			"build/":    true,
		},
		exists: map[string]bool{
			"src/main.c": true,
		},
	}

	plan, err := PlanBuild("build/app", targets, ctx, fs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(plan.Tasks))
	}
	// build/ must come before build/app
	if plan.Tasks[0].Target != "build/" {
		t.Errorf("expected first task 'build/', got %q", plan.Tasks[0].Target)
	}
	if plan.Tasks[1].Target != "build/app" {
		t.Errorf("expected second task 'build/app', got %q", plan.Tasks[1].Target)
	}
}

func TestPlanBuild_OrderOnlyDeps_NotForStaleness(t *testing.T) {
	// build/app has order-only dep on build/
	// Order-only deps don't trigger rebuilds based on timestamp
	targets := []*ast.Target{
		createTargetWithOrderDeps("build/app", false, false, []string{"src/main.c"}, []string{"build/"}),
		createTargetWithRecipe("build/", false, true, nil, &ast.Recipe{}),
	}
	ctx := eval.NewContext()
	now := time.Now()
	fs := &mockFileSystem{
		exists: map[string]bool{
			"build/app":  true,
			"build/":     true,
			"src/main.c": true,
		},
		mtimes: map[string]time.Time{
			"build/app":  now,                     // target is newer
			"build/":     now.Add(1 * time.Hour),  // order-only dep is newer (should be ignored)
			"src/main.c": now.Add(-1 * time.Hour), // normal dep is older
		},
	}

	plan, err := PlanBuild("build/app", targets, ctx, fs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Target should be up-to-date because order-only deps don't affect staleness
	if len(plan.Tasks) != 0 {
		t.Errorf("expected 0 tasks (up to date), got %d", len(plan.Tasks))
	}
}

func TestPlanBuild_OrderOnlyDeps_MustExist(t *testing.T) {
	// build/app has order-only dep on build/
	// Order-only deps must exist (either as files or have build rules)
	targets := []*ast.Target{
		createTargetWithOrderDeps("build/app", false, false, nil, []string{"nonexistent/"}),
	}
	ctx := eval.NewContext()
	fs := &mockFileSystem{
		missing: map[string]bool{
			"build/app":    true,
			"nonexistent/": true, // Order-only dep doesn't exist and has no rule
		},
	}

	_, err := PlanBuild("build/app", targets, ctx, fs)

	if err == nil {
		t.Error("expected error for missing order-only dependency")
	}
}

func TestPlanBuild_OrderOnlyDeps_InTask(t *testing.T) {
	// Verify order-only deps are tracked separately in BuildTask
	targets := []*ast.Target{
		createTargetWithOrderDeps("build/app", false, false, []string{"src/main.c"}, []string{"build/"}),
		createTargetWithRecipe("build/", false, true, nil, &ast.Recipe{}),
	}
	ctx := eval.NewContext()
	fs := &mockFileSystem{
		missing: map[string]bool{
			"build/app": true,
			"build/":    true,
		},
		exists: map[string]bool{
			"src/main.c": true,
		},
	}

	plan, err := PlanBuild("build/app", targets, ctx, fs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find the build/app task
	var appTask *BuildTask
	for i := range plan.Tasks {
		if plan.Tasks[i].Target == "build/app" {
			appTask = &plan.Tasks[i]
			break
		}
	}

	if appTask == nil {
		t.Fatal("build/app task not found in plan")
	}

	// Check regular dependencies
	if len(appTask.Dependencies) != 1 || appTask.Dependencies[0] != "src/main.c" {
		t.Errorf("expected Dependencies=[src/main.c], got %v", appTask.Dependencies)
	}

	// Check order-only dependencies
	if len(appTask.OrderOnlyDeps) != 1 || appTask.OrderOnlyDeps[0] != "build/" {
		t.Errorf("expected OrderOnlyDeps=[build/], got %v", appTask.OrderOnlyDeps)
	}
}
