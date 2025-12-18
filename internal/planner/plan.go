package planner

import (
	"fmt"
	"strings"
	"time"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/eval"
)

// BuildReason indicates why a target needs to be rebuilt.
type BuildReason int

const (
	BuildReasonTargetMissing BuildReason = iota
	BuildReasonDependencyNewer
	BuildReasonPhonyTarget
	BuildReasonForcedRebuild
)

func (r BuildReason) String() string {
	switch r {
	case BuildReasonTargetMissing:
		return "target missing"
	case BuildReasonDependencyNewer:
		return "dependency newer than target"
	case BuildReasonPhonyTarget:
		return "phony target"
	case BuildReasonForcedRebuild:
		return "forced rebuild"
	default:
		return "unknown"
	}
}

// BuildTask represents a single task in the build plan.
type BuildTask struct {
	// Target is the path to be built (with @ prefix for phony targets)
	Target string

	// Dependencies are the resolved paths of dependencies
	Dependencies []string

	// OrderOnlyDeps are order-only dependencies (from .after: directives)
	// These must exist/be built before this target but don't affect staleness
	OrderOnlyDeps []string

	// Recipe is the recipe to execute (may be nil for phony without recipe)
	Recipe *ast.Recipe

	// Reason indicates why this target needs rebuilding
	Reason BuildReason

	// Captures holds pattern capture values (for pattern targets)
	Captures map[string]string

	// TargetDef is the AST target definition
	TargetDef *ast.Target
}

// BuildPlan contains a topologically sorted list of build tasks.
type BuildPlan struct {
	Tasks []BuildTask
}

// FileSystem abstracts file system operations for testability.
type FileSystem interface {
	Exists(path string) bool
	ModTime(path string) (time.Time, error)
}

// CircularDependencyError is returned when a circular dependency is detected.
type CircularDependencyError struct {
	Path []string
}

func (e *CircularDependencyError) Error() string {
	return fmt.Sprintf("circular dependency detected: %s", strings.Join(e.Path, " -> "))
}

// MissingSourceError is returned when a source file doesn't exist.
type MissingSourceError struct {
	Path string
}

func (e *MissingSourceError) Error() string {
	return fmt.Sprintf("missing source file: %s", e.Path)
}

// PlanBuild creates a build plan for the requested target.
//
// It performs:
//  1. Target lookup (finding matching target definition)
//  2. Dependency resolution (converting patterns to concrete paths)
//  3. Recursive planning (planning all dependencies first)
//  4. Staleness detection (determining if rebuild is needed)
//  5. Topological sorting (ordering tasks for execution)
func PlanBuild(requestedTarget string, targets []*ast.Target, ctx *eval.Context, fs FileSystem) (*BuildPlan, error) {
	planner := &buildPlanner{
		targets:   targets,
		ctx:       ctx,
		fs:        fs,
		visited:   make(map[string]bool),
		inStack:   make(map[string]bool),
		tasks:     make([]BuildTask, 0),
		taskIndex: make(map[string]int),
	}

	if err := planner.planTarget(requestedTarget); err != nil {
		return nil, err
	}

	return &BuildPlan{Tasks: planner.tasks}, nil
}

// buildPlanner handles the recursive planning process.
type buildPlanner struct {
	targets   []*ast.Target
	ctx       *eval.Context
	fs        FileSystem
	visited   map[string]bool // Targets already processed
	inStack   map[string]bool // Targets in current recursion stack (for cycle detection)
	tasks     []BuildTask     // Tasks in topological order
	taskIndex map[string]int  // Index of each target in tasks (for dedup)
}

// planTarget recursively plans a single target.
func (p *buildPlanner) planTarget(targetPath string) error {
	// Check for circular dependency
	if p.inStack[targetPath] {
		return p.buildCycleError(targetPath)
	}

	// Already processed
	if p.visited[targetPath] {
		return nil
	}

	// Find matching target definition
	target, captures, err := LookupTarget(targetPath, p.targets)
	if err != nil {
		// Not a defined target - could be a source file
		if p.fs.Exists(p.normalizePath(targetPath)) {
			return nil // Source file exists, nothing to build
		}
		return err // Target not found and file doesn't exist
	}

	// Mark as in-progress for cycle detection
	p.inStack[targetPath] = true
	defer func() { delete(p.inStack, targetPath) }()

	// Resolve dependencies
	depPaths, err := ResolveDependencies(target.Dependencies, captures, p.ctx)
	if err != nil {
		return err
	}

	// Resolve order-only dependencies from .after: directives
	orderOnlyPaths, err := p.resolveOrderOnlyDeps(target, captures)
	if err != nil {
		return err
	}

	// Recursively plan normal dependencies
	for _, depPath := range depPaths {
		if err := p.planTarget(depPath); err != nil {
			// Check if it's a missing source file
			if _, ok := err.(*TargetNotFoundError); ok {
				if !p.fs.Exists(depPath) {
					return &MissingSourceError{Path: depPath}
				}
			} else {
				return err
			}
		}
	}

	// Recursively plan order-only dependencies (for build order, not staleness)
	for _, orderPath := range orderOnlyPaths {
		if err := p.planTarget(orderPath); err != nil {
			// Check if it's a missing source file
			if _, ok := err.(*TargetNotFoundError); ok {
				if !p.fs.Exists(orderPath) {
					return &MissingSourceError{Path: orderPath}
				}
			} else {
				return err
			}
		}
	}

	// Check if rebuild is needed (only based on normal deps, not order-only)
	reason, needsRebuild := p.needsRebuild(targetPath, target, depPaths)

	// Mark as visited
	p.visited[targetPath] = true

	// Add to plan if rebuild needed
	if needsRebuild {
		task := BuildTask{
			Target:        targetPath,
			Dependencies:  depPaths,
			OrderOnlyDeps: orderOnlyPaths,
			Recipe:        target.Recipe,
			Reason:        reason,
			Captures:      captures,
			TargetDef:     target,
		}
		p.tasks = append(p.tasks, task)
		p.taskIndex[targetPath] = len(p.tasks) - 1
	}

	return nil
}

// resolveOrderOnlyDeps resolves order-only dependencies from .after: directives.
func (p *buildPlanner) resolveOrderOnlyDeps(target *ast.Target, captures map[string]string) ([]string, error) {
	if target.Recipe == nil || len(target.Recipe.Directives.After) == 0 {
		return nil, nil
	}

	orderOnlyPaths := make([]string, 0, len(target.Recipe.Directives.After))
	for _, afterValue := range target.Recipe.Directives.After {
		// Evaluate the .after: value
		evaluator := newValueEvaluator(p.ctx, captures)
		path, err := evaluator.evaluate(afterValue)
		if err != nil {
			return nil, err
		}
		orderOnlyPaths = append(orderOnlyPaths, path)
	}

	return orderOnlyPaths, nil
}

// valueEvaluator evaluates AST values to strings.
type valueEvaluator struct {
	ctx      *eval.Context
	captures map[string]string
}

func newValueEvaluator(ctx *eval.Context, captures map[string]string) *valueEvaluator {
	return &valueEvaluator{ctx: ctx, captures: captures}
}

func (e *valueEvaluator) evaluate(v *ast.Value) (string, error) {
	if v == nil {
		return "", nil
	}

	var sb strings.Builder
	for _, part := range v.Parts {
		switch p := part.(type) {
		case *ast.LiteralValue:
			sb.WriteString(p.Text)
		case *ast.Interpolation:
			// Check captures first
			if e.captures != nil {
				if val, ok := e.captures[p.Name]; ok {
					sb.WriteString(val)
					continue
				}
			}
			// Check context
			if val, ok := e.ctx.Get(p.Name); ok {
				sb.WriteString(val)
				continue
			}
			return "", &UndefinedVariableError{Name: p.Name}
		default:
			// Skip unknown parts (e.g., function calls)
		}
	}
	return sb.String(), nil
}

// normalizePath removes @ prefix for file system operations.
func (p *buildPlanner) normalizePath(path string) string {
	if strings.HasPrefix(path, "@") {
		return path[1:]
	}
	return path
}

// needsRebuild determines if a target needs rebuilding.
func (p *buildPlanner) needsRebuild(targetPath string, target *ast.Target, depPaths []string) (BuildReason, bool) {
	// Phony targets always need rebuilding
	if target.Pattern.IsPhony {
		return BuildReasonPhonyTarget, true
	}

	fsPath := p.normalizePath(targetPath)

	// Target doesn't exist - needs building
	if !p.fs.Exists(fsPath) {
		return BuildReasonTargetMissing, true
	}

	// Check if any dependency is newer
	targetMtime, err := p.fs.ModTime(fsPath)
	if err != nil {
		return BuildReasonTargetMissing, true
	}

	for _, depPath := range depPaths {
		depFsPath := p.normalizePath(depPath)

		// Check if dependency was rebuilt (it's in our task list)
		if _, ok := p.taskIndex[depPath]; ok {
			return BuildReasonDependencyNewer, true
		}

		if p.fs.Exists(depFsPath) {
			depMtime, err := p.fs.ModTime(depFsPath)
			if err != nil {
				continue
			}
			if depMtime.After(targetMtime) {
				return BuildReasonDependencyNewer, true
			}
		}
	}

	return 0, false
}

// buildCycleError constructs a CircularDependencyError from the current stack.
func (p *buildPlanner) buildCycleError(targetPath string) error {
	// Reconstruct the cycle path
	cycle := []string{targetPath}

	// We need to find where in the stack the cycle starts
	// For simplicity, just return the target that caused the cycle
	cycle = append(cycle, targetPath)

	return &CircularDependencyError{Path: cycle}
}
