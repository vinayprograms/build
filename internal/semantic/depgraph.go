package semantic

import (
	"github.com/vinayprograms/need/internal/ast"
)

// DependencyGraph represents the dependency relationships between targets.
type DependencyGraph struct {
	// Nodes contains all target names in the graph.
	Nodes map[string]bool

	// Edges maps each target to its dependencies.
	// If A depends on B, then Edges[A] contains B.
	Edges map[string][]string
}

// NewDependencyGraph creates an empty dependency graph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Nodes: make(map[string]bool),
		Edges: make(map[string][]string),
	}
}

// AddNode adds a target node to the graph.
func (g *DependencyGraph) AddNode(name string) {
	g.Nodes[name] = true
}

// AddEdge adds a dependency edge from target to dep.
func (g *DependencyGraph) AddEdge(target, dep string) {
	g.Edges[target] = append(g.Edges[target], dep)
}

// DependencyResult contains the results of dependency validation.
type DependencyResult struct {
	// Graph is the constructed dependency graph.
	Graph *DependencyGraph

	// PatternTargets contains targets that have pattern captures.
	// These define rules rather than concrete targets.
	PatternTargets []*ast.Target

	// UnsatisfiedDeps maps target names to their unsatisfied dependencies.
	// A dependency is unsatisfied if it's not defined as a target
	// (it may be a source file or need to be matched by a pattern).
	UnsatisfiedDeps map[string][]string

	// Errors contains validation errors (e.g., circular dependencies).
	Errors []error
}

// HasErrors returns true if there are any validation errors.
func (r *DependencyResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// ValidateDependencies performs Pass 4 of semantic analysis:
// builds the dependency graph and detects cycles.
func ValidateDependencies(targets []*ast.Target) *DependencyResult {
	result := &DependencyResult{
		Graph:           NewDependencyGraph(),
		PatternTargets:  make([]*ast.Target, 0),
		UnsatisfiedDeps: make(map[string][]string),
		Errors:          make([]error, 0),
	}

	if len(targets) == 0 {
		return result
	}

	// First pass: collect all target names and identify pattern targets
	targetNames := make(map[string]bool)
	for _, t := range targets {
		if isPatternTarget(t) {
			result.PatternTargets = append(result.PatternTargets, t)
			continue
		}

		name := PatternString(&t.Pattern)
		targetNames[name] = true
		result.Graph.AddNode(name)
	}

	// Second pass: build edges and track unsatisfied dependencies
	for _, t := range targets {
		if isPatternTarget(t) {
			// Pattern targets are rules, not concrete nodes
			continue
		}

		targetName := PatternString(&t.Pattern)

		for _, dep := range t.Dependencies {
			depName := dependencyToString(&dep)

			// Add edge
			result.Graph.AddEdge(targetName, depName)

			// Track unsatisfied dependencies
			if !targetNames[depName] {
				// This dependency is not defined as a target
				// It could be a source file or matched by a pattern
				result.UnsatisfiedDeps[targetName] = append(
					result.UnsatisfiedDeps[targetName], depName)
			}
		}
	}

	// Third pass: detect cycles using DFS
	cycle := findCycle(result.Graph)
	if cycle != nil {
		result.Errors = append(result.Errors, &CircularDependencyError{Cycle: cycle})
	}

	return result
}

// isPatternTarget returns true if the target has any BraceExpr segments
// that are captures (not variable interpolations).
func isPatternTarget(t *ast.Target) bool {
	for _, seg := range t.Pattern.Segments {
		if _, ok := seg.(*ast.BraceExpr); ok {
			return true
		}
	}
	return false
}

// dependencyToString converts a dependency to a string.
func dependencyToString(d *ast.Dependency) string {
	return SegmentsToString(d.Segments)
}

// findCycle uses DFS to detect cycles in the graph.
// Returns the cycle path if found, nil otherwise.
func findCycle(g *DependencyGraph) []string {
	// Track visited nodes and nodes in current recursion stack
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	parent := make(map[string]string)

	// DFS from each unvisited node
	for node := range g.Nodes {
		if !visited[node] {
			if cycle := dfs(g, node, visited, recStack, parent); cycle != nil {
				return cycle
			}
		}
	}

	return nil
}

// dfs performs depth-first search to detect cycles.
func dfs(g *DependencyGraph, node string, visited, recStack map[string]bool, parent map[string]string) []string {
	visited[node] = true
	recStack[node] = true

	for _, neighbor := range g.Edges[node] {
		// Only consider neighbors that are in the graph
		// (unsatisfied dependencies aren't nodes)
		if !g.Nodes[neighbor] {
			continue
		}

		if !visited[neighbor] {
			parent[neighbor] = node
			if cycle := dfs(g, neighbor, visited, recStack, parent); cycle != nil {
				return cycle
			}
		} else if recStack[neighbor] {
			// Found a cycle - reconstruct the cycle path
			return reconstructCycle(parent, node, neighbor)
		}
	}

	recStack[node] = false
	return nil
}

// reconstructCycle builds the cycle path from parent map.
func reconstructCycle(parent map[string]string, end, cycleStart string) []string {
	// Build path from end back to cycleStart (in reverse order)
	var reversePath []string
	current := end

	// Walk back through parents until we reach cycleStart
	for current != cycleStart {
		reversePath = append(reversePath, current)
		current = parent[current]
	}

	// Build final cycle: cycleStart -> ... -> end -> cycleStart
	cycle := make([]string, 0, len(reversePath)+2)
	cycle = append(cycle, cycleStart)

	// Reverse the path and append
	for i := len(reversePath) - 1; i >= 0; i-- {
		cycle = append(cycle, reversePath[i])
	}

	// Add cycleStart again to show the complete cycle
	cycle = append(cycle, cycleStart)

	return cycle
}
