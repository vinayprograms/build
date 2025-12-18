package main

import (
	"time"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/planner"
)

// ----------------------------------------------------------------------------
// Target Matching Adapters
// ----------------------------------------------------------------------------

// MatchResult represents the result of matching a path against a target pattern.
type MatchResult interface {
	// Matched returns true if the pattern matched the path.
	Matched() bool

	// CaptureCount returns the number of captures.
	CaptureCount() int

	// CaptureName returns the i-th capture name.
	CaptureName(i int) string

	// CaptureValue returns the i-th capture value.
	CaptureValue(i int) string
}

// matchResultAdapter wraps the result of planner.MatchTarget.
type matchResultAdapter struct {
	matched   bool
	captures  map[string]string
	nameOrder []string
}

func (m *matchResultAdapter) Matched() bool {
	return m.matched
}

func (m *matchResultAdapter) CaptureCount() int {
	return len(m.nameOrder)
}

func (m *matchResultAdapter) CaptureName(i int) string {
	if i < 0 || i >= len(m.nameOrder) {
		return ""
	}
	return m.nameOrder[i]
}

func (m *matchResultAdapter) CaptureValue(i int) string {
	if i < 0 || i >= len(m.nameOrder) {
		return ""
	}
	return m.captures[m.nameOrder[i]]
}

// MatchTargetPattern matches a path against a target pattern.
// Returns a MatchResult indicating whether the match succeeded and any captures.
func MatchTargetPattern(pattern *ast.TargetPattern, path string) MatchResult {
	matched, captures := planner.MatchTarget(pattern, path)

	// Build stable name order for captures
	nameOrder := make([]string, 0, len(captures))
	for name := range captures {
		nameOrder = append(nameOrder, name)
	}

	return &matchResultAdapter{
		matched:   matched,
		captures:  captures,
		nameOrder: nameOrder,
	}
}

// LookupResult represents the result of looking up a target by path.
type LookupResult interface {
	// Found returns true if a matching target was found.
	Found() bool

	// Target returns the matched target (or nil if not found).
	Target() *ast.Target

	// CaptureCount returns the number of captures.
	CaptureCount() int

	// CaptureName returns the i-th capture name.
	CaptureName(i int) string

	// CaptureValue returns the i-th capture value.
	CaptureValue(i int) string

	// Error returns the error (if not found).
	Error() error
}

// lookupResultAdapter wraps the result of planner.LookupTarget.
type lookupResultAdapter struct {
	target    *ast.Target
	captures  map[string]string
	nameOrder []string
	err       error
}

func (l *lookupResultAdapter) Found() bool {
	return l.target != nil
}

func (l *lookupResultAdapter) Target() *ast.Target {
	return l.target
}

func (l *lookupResultAdapter) CaptureCount() int {
	return len(l.nameOrder)
}

func (l *lookupResultAdapter) CaptureName(i int) string {
	if i < 0 || i >= len(l.nameOrder) {
		return ""
	}
	return l.nameOrder[i]
}

func (l *lookupResultAdapter) CaptureValue(i int) string {
	if i < 0 || i >= len(l.nameOrder) {
		return ""
	}
	return l.captures[l.nameOrder[i]]
}

func (l *lookupResultAdapter) Error() error {
	return l.err
}

// LookupTargetByPath finds a target that matches the given path.
// Returns a LookupResult with the target, captures, and any error.
func LookupTargetByPath(path string, targets []*ast.Target) LookupResult {
	target, captures, err := planner.LookupTarget(path, targets)

	// Build stable name order for captures
	nameOrder := make([]string, 0, len(captures))
	for name := range captures {
		nameOrder = append(nameOrder, name)
	}

	return &lookupResultAdapter{
		target:    target,
		captures:  captures,
		nameOrder: nameOrder,
		err:       err,
	}
}

// ----------------------------------------------------------------------------
// Dependency Resolution Adapters
// ----------------------------------------------------------------------------

// ResolveResult represents the result of resolving dependencies.
type ResolveResult interface {
	// Paths returns the resolved dependency paths.
	Paths() []string

	// Error returns any error during resolution.
	Error() error
}

// resolveResultAdapter wraps the result of planner.ResolveDependencies.
type resolveResultAdapter struct {
	paths []string
	err   error
}

func (r *resolveResultAdapter) Paths() []string {
	return r.paths
}

func (r *resolveResultAdapter) Error() error {
	return r.err
}

// ResolveDependencyPaths resolves dependencies to concrete paths.
func ResolveDependencyPaths(deps []ast.Dependency, captures map[string]string, ctx EvalContext) ResolveResult {
	// Get the underlying eval context
	eca, ok := ctx.(*evalContextAdapter)
	if !ok {
		return &resolveResultAdapter{err: &planner.UndefinedVariableError{Name: "(invalid context)"}}
	}

	paths, err := planner.ResolveDependencies(deps, captures, eca.ctx)
	return &resolveResultAdapter{paths: paths, err: err}
}

// ----------------------------------------------------------------------------
// Build Planning Adapters
// ----------------------------------------------------------------------------

// BuildTask represents a single task in the build plan.
type BuildTask interface {
	// Target returns the path to be built.
	Target() string

	// Dependencies returns the resolved dependency paths.
	Dependencies() []string

	// OrderOnlyDeps returns the order-only dependency paths (from .after:).
	OrderOnlyDeps() []string

	// Reason returns why this target needs rebuilding.
	Reason() string

	// CaptureCount returns the number of captures.
	CaptureCount() int

	// CaptureName returns the i-th capture name.
	CaptureName(i int) string

	// CaptureValue returns the i-th capture value.
	CaptureValue(i int) string

	// HasRecipe returns true if this task has a recipe.
	HasRecipe() bool
}

// buildTaskAdapter wraps a planner.BuildTask.
type buildTaskAdapter struct {
	task      planner.BuildTask
	nameOrder []string
}

func (t *buildTaskAdapter) Target() string {
	return t.task.Target
}

func (t *buildTaskAdapter) Dependencies() []string {
	return t.task.Dependencies
}

func (t *buildTaskAdapter) OrderOnlyDeps() []string {
	return t.task.OrderOnlyDeps
}

func (t *buildTaskAdapter) Reason() string {
	return t.task.Reason.String()
}

func (t *buildTaskAdapter) CaptureCount() int {
	return len(t.nameOrder)
}

func (t *buildTaskAdapter) CaptureName(i int) string {
	if i < 0 || i >= len(t.nameOrder) {
		return ""
	}
	return t.nameOrder[i]
}

func (t *buildTaskAdapter) CaptureValue(i int) string {
	if i < 0 || i >= len(t.nameOrder) {
		return ""
	}
	return t.task.Captures[t.nameOrder[i]]
}

func (t *buildTaskAdapter) HasRecipe() bool {
	return t.task.Recipe != nil
}

// BuildPlanResult represents the result of build planning.
type BuildPlanResult interface {
	// TaskCount returns the number of tasks in the plan.
	TaskCount() int

	// Task returns the i-th task.
	Task(i int) BuildTask

	// Error returns any error during planning.
	Error() error
}

// buildPlanResultAdapter wraps a planner.BuildPlan.
type buildPlanResultAdapter struct {
	plan *planner.BuildPlan
	err  error
}

func (p *buildPlanResultAdapter) TaskCount() int {
	if p.plan == nil {
		return 0
	}
	return len(p.plan.Tasks)
}

func (p *buildPlanResultAdapter) Task(i int) BuildTask {
	if p.plan == nil || i < 0 || i >= len(p.plan.Tasks) {
		return nil
	}
	task := p.plan.Tasks[i]
	nameOrder := make([]string, 0, len(task.Captures))
	for name := range task.Captures {
		nameOrder = append(nameOrder, name)
	}
	return &buildTaskAdapter{task: task, nameOrder: nameOrder}
}

func (p *buildPlanResultAdapter) Error() error {
	return p.err
}

// PlanBuild creates a build plan for the requested target.
func PlanBuild(target string, targets []*ast.Target, ctx EvalContext, fs FileSystem) BuildPlanResult {
	// Get the underlying eval context
	eca, ok := ctx.(*evalContextAdapter)
	if !ok {
		return &buildPlanResultAdapter{err: &planner.UndefinedVariableError{Name: "(invalid context)"}}
	}

	// Create the file system adapter
	fsAdapter := &fileSystemPlannerAdapter{fs: fs}

	plan, err := planner.PlanBuild(target, targets, eca.ctx, fsAdapter)
	return &buildPlanResultAdapter{plan: plan, err: err}
}

// fileSystemPlannerAdapter adapts our FileSystem interface to planner.FileSystem.
type fileSystemPlannerAdapter struct {
	fs FileSystem
}

func (f *fileSystemPlannerAdapter) Exists(path string) bool {
	return f.fs.Exists(path)
}

func (f *fileSystemPlannerAdapter) ModTime(path string) (time.Time, error) {
	return f.fs.ModTime(path)
}

// FileSystem interface for file system operations.
type FileSystem interface {
	Exists(path string) bool
	ModTime(path string) (time.Time, error)
}
