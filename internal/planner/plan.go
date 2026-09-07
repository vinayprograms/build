package planner

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/cache"
	"github.com/vinayprograms/build/internal/eval"
	"github.com/vinayprograms/build/internal/output"
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

	// AutodepsDeps are additional dependencies from .autodeps file
	// These are learned from previous builds (e.g., header files)
	AutodepsDeps []string

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

	// AutodepsPath is the .autodeps file path to update after build
	AutodepsPath string
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
	return PlanBuildWithOptions(requestedTarget, targets, ctx, fs, nil, nil)
}

// PlanBuildWithVerbose creates a build plan with optional verbose output.
// If verboseOutput is non-nil, staleness check decisions are written to it.
func PlanBuildWithVerbose(requestedTarget string, targets []*ast.Target, ctx *eval.Context, fs FileSystem, verboseOutput io.Writer) (*BuildPlan, error) {
	return PlanBuildWithOptions(requestedTarget, targets, ctx, fs, nil, verboseOutput)
}

// PlanBuildWithEmitter creates a build plan with event emission.
// If emitter is non-nil, staleness check events are emitted.
func PlanBuildWithEmitter(requestedTarget string, targets []*ast.Target, ctx *eval.Context, fs FileSystem, emitter *output.Emitter) (*BuildPlan, error) {
	return planBuildInternal(requestedTarget, targets, ctx, fs, nil, nil, emitter)
}

// PlanBuildWithOptions creates a build plan with all optional parameters.
// If autodepsCache is non-nil, .d file parsing results are cached.
// If verboseOutput is non-nil, staleness check decisions are written to it.
func PlanBuildWithOptions(requestedTarget string, targets []*ast.Target, ctx *eval.Context, fs FileSystem, autodepsCache *cache.AutodepsCache, verboseOutput io.Writer) (*BuildPlan, error) {
	return planBuildInternal(requestedTarget, targets, ctx, fs, autodepsCache, verboseOutput, nil)
}

// planBuildInternal is the internal implementation that handles all parameters.
func planBuildInternal(requestedTarget string, targets []*ast.Target, ctx *eval.Context, fs FileSystem, autodepsCache *cache.AutodepsCache, verboseOutput io.Writer, emitter *output.Emitter) (*BuildPlan, error) {
	planner := &buildPlanner{
		targets:       targets,
		targetIndex:   NewTargetIndex(targets, ctx),
		ctx:           ctx,
		fs:            fs,
		autodepsCache: autodepsCache,
		visited:       make(map[string]bool),
		inStack:       make(map[string]bool),
		tasks:         make([]BuildTask, 0),
		taskIndex:     make(map[string]int),
		verboseOutput: verboseOutput,
		emitter:       emitter,
	}

	if err := planner.planTarget(requestedTarget); err != nil {
		return nil, err
	}

	return &BuildPlan{Tasks: planner.tasks}, nil
}

// buildPlanner handles the recursive planning process.
type buildPlanner struct {
	targets       []*ast.Target
	targetIndex   *TargetIndex // Optimized target lookup index
	ctx           *eval.Context
	fs            FileSystem
	autodepsCache *cache.AutodepsCache // Cache for .d file parsing
	visited       map[string]bool      // Targets already processed
	inStack       map[string]bool      // Targets in current recursion stack (for cycle detection)
	stack         []string             // Ordered recursion stack, mirrors inStack (for cycle path reporting)
	tasks         []BuildTask          // Tasks in topological order
	taskIndex     map[string]int       // Index of each target in tasks (for dedup)
	verboseOutput io.Writer            // Optional output for verbose mode
	emitter       *output.Emitter      // Optional event emitter for output system
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

	// Find matching target definition using the index
	target, captures, err := p.targetIndex.Lookup(targetPath)
	if err != nil {
		// Not a defined target - could be a source file
		if p.fs.Exists(p.normalizePath(targetPath)) {
			return nil // Source file exists, nothing to build
		}
		return err // Target not found and file doesn't exist
	}

	// Mark as in-progress for cycle detection
	p.inStack[targetPath] = true
	p.stack = append(p.stack, targetPath)
	defer func() {
		delete(p.inStack, targetPath)
		p.stack = p.stack[:len(p.stack)-1]
	}()

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

	// Resolve autodeps (learned dependencies from previous builds)
	autodepsDeps, autodepsPath, err := p.resolveAutodeps(target, captures)
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

	// Check if rebuild is needed (based on normal deps and autodeps, not order-only)
	reason, needsRebuild := p.needsRebuild(targetPath, target, depPaths, autodepsDeps)

	// Verbose output for staleness check decision
	if p.verboseOutput != nil {
		if needsRebuild {
			fmt.Fprintf(p.verboseOutput, "%s: rebuild needed (%s)\n", targetPath, reason.String())
		} else {
			fmt.Fprintf(p.verboseOutput, "%s: up to date\n", targetPath)
		}
	}

	// Emit staleness check event
	if p.emitter != nil {
		if needsRebuild {
			p.emitter.StalenessChecked(targetPath, reason.String(), "rebuild")
		} else {
			p.emitter.StalenessChecked(targetPath, "up to date", "skip")
		}
	}

	// Mark as visited
	p.visited[targetPath] = true

	// Add to plan if rebuild needed
	if needsRebuild {
		task := BuildTask{
			Target:        targetPath,
			Dependencies:  depPaths,
			AutodepsDeps:  autodepsDeps,
			OrderOnlyDeps: orderOnlyPaths,
			Recipe:        target.Recipe,
			Reason:        reason,
			Captures:      captures,
			TargetDef:     target,
			AutodepsPath:  autodepsPath,
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

// resolveAutodeps resolves autodeps (learned dependencies from .autodeps file).
// Returns the list of learned dependencies and the autodeps file path.
func (p *buildPlanner) resolveAutodeps(target *ast.Target, captures map[string]string) ([]string, string, error) {
	if target.Recipe == nil || target.Recipe.Directives.Autodeps == nil {
		return nil, "", nil
	}

	// Evaluate the .autodeps: value to get the path
	evaluator := newValueEvaluator(p.ctx, captures)
	autodepsPath, err := evaluator.evaluate(target.Recipe.Directives.Autodeps)
	if err != nil {
		return nil, "", err
	}

	// Try cache first if available
	if p.autodepsCache != nil {
		if deps, ok, err := p.autodepsCache.Get(autodepsPath); err == nil && ok {
			return deps, autodepsPath, nil
		}
	}

	// Parse the .d file to get learned dependencies
	deps, err := ParseAutodepsFile(autodepsPath)
	if err != nil {
		return nil, autodepsPath, err
	}

	// Cache the result if cache is available
	if p.autodepsCache != nil && deps != nil {
		// Only cache if file exists (ParseAutodepsFile returns nil, nil for missing files)
		_ = p.autodepsCache.Put(autodepsPath, deps)
	}

	return deps, autodepsPath, nil
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
func (p *buildPlanner) needsRebuild(targetPath string, target *ast.Target, depPaths []string, autodepsDeps []string) (BuildReason, bool) {
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

	// Check declared dependencies
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

	// Check autodeps (learned dependencies from previous builds)
	for _, depPath := range autodepsDeps {
		if p.fs.Exists(depPath) {
			depMtime, err := p.fs.ModTime(depPath)
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
// It reports the full cycle path, e.g. "@a -> @b -> @c -> @a", by finding
// where targetPath first appears in the ordered recursion stack and taking
// everything from there, plus targetPath again to close the loop.
func (p *buildPlanner) buildCycleError(targetPath string) error {
	start := -1
	for i, t := range p.stack {
		if t == targetPath {
			start = i
			break
		}
	}

	var cycle []string
	if start >= 0 {
		cycle = append(cycle, p.stack[start:]...)
	}
	cycle = append(cycle, targetPath)

	return &CircularDependencyError{Path: cycle}
}
