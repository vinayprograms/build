package planner

import (
	"strings"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/eval"
)

// CompiledPattern is a pre-processed target pattern for faster matching.
type CompiledPattern struct {
	Target       *ast.Target
	Prefix       string   // Literal prefix before first capture (empty if starts with capture)
	IsLiteral    bool     // True if no captures (after resolving interpolations)
	IsPhony      bool     // True if phony target
	PatternStr   string   // String representation for debugging (with interpolations resolved)
	Captures     []string // Capture names in order (only actual captures, not interpolations)
	ResolvedPath string   // Fully resolved path for literal patterns
}

// TargetIndex provides optimized target lookup using pre-compiled patterns.
// It groups targets by their literal prefix for faster lookup.
type TargetIndex struct {
	// literals maps exact paths to their targets (for exact match shortcut)
	literals map[string]*ast.Target

	// phonies maps phony names (without @) to their targets
	phonies map[string]*ast.Target

	// byPrefix groups pattern targets by their prefix
	// Empty string key contains patterns with no literal prefix
	byPrefix map[string][]*CompiledPattern

	// all contains all compiled patterns for fallback
	all []*CompiledPattern
}

// NewTargetIndex creates an optimized index for fast target lookup.
// The context is used to resolve variable interpolations in target patterns.
func NewTargetIndex(targets []*ast.Target, ctx *eval.Context) *TargetIndex {
	idx := &TargetIndex{
		literals: make(map[string]*ast.Target),
		phonies:  make(map[string]*ast.Target),
		byPrefix: make(map[string][]*CompiledPattern),
		all:      make([]*CompiledPattern, 0, len(targets)),
	}

	for _, target := range targets {
		cp := compilePatternWithContext(target, ctx)
		idx.all = append(idx.all, cp)

		if cp.IsPhony {
			// Extract phony name (with interpolations resolved)
			idx.phonies[cp.ResolvedPath] = target
		} else if cp.IsLiteral {
			// Exact match path (with interpolations resolved)
			idx.literals[cp.ResolvedPath] = target
		} else {
			// Pattern: index by prefix
			prefix := extractPrefix(cp.Prefix)
			idx.byPrefix[prefix] = append(idx.byPrefix[prefix], cp)
		}
	}

	return idx
}

// Size returns the number of targets in the index.
func (idx *TargetIndex) Size() int {
	return len(idx.all)
}

// Lookup finds a target that matches the given path.
// Returns the target, capture values, and any error.
func (idx *TargetIndex) Lookup(path string) (*ast.Target, map[string]string, error) {
	// Strip @ prefix for phony lookup
	lookupPath := path
	isPhonyRef := false
	if strings.HasPrefix(path, "@") {
		lookupPath = path[1:]
		isPhonyRef = true
	}

	// 1. Try exact phony match (most specific)
	if isPhonyRef {
		if target, ok := idx.phonies[lookupPath]; ok {
			return target, make(map[string]string), nil
		}
	}

	// 2. Try exact literal match
	if target, ok := idx.literals[lookupPath]; ok {
		return target, make(map[string]string), nil
	}

	// 3. Try phony match without @ (for backwards compatibility)
	if !isPhonyRef {
		if target, ok := idx.phonies[lookupPath]; ok {
			return target, make(map[string]string), nil
		}
	}

	// 4. Try pattern matches by prefix
	prefix := extractPrefix(lookupPath)
	if patterns, ok := idx.byPrefix[prefix]; ok {
		for _, cp := range patterns {
			if matched, captures := MatchTarget(&cp.Target.Pattern, path); matched {
				return cp.Target, captures, nil
			}
		}
	}

	// 5. Try patterns with empty prefix
	if patterns, ok := idx.byPrefix[""]; ok {
		for _, cp := range patterns {
			if matched, captures := MatchTarget(&cp.Target.Pattern, path); matched {
				return cp.Target, captures, nil
			}
		}
	}

	// 6. Fallback: try all patterns (for complex patterns)
	for _, cp := range idx.all {
		if !cp.IsLiteral && !cp.IsPhony {
			if matched, captures := MatchTarget(&cp.Target.Pattern, path); matched {
				return cp.Target, captures, nil
			}
		}
	}

	return nil, nil, &TargetNotFoundError{Path: path}
}

// compilePatternWithContext creates a CompiledPattern from a target,
// resolving variable interpolations using the context.
// Variables that exist in context are resolved to literals.
// Variables that don't exist are treated as captures.
func compilePatternWithContext(target *ast.Target, ctx *eval.Context) *CompiledPattern {
	cp := &CompiledPattern{
		Target:  target,
		IsPhony: target.Pattern.IsPhony,
	}

	// Build resolved path and identify captures
	var resolved strings.Builder
	var prefix strings.Builder
	hasCapture := false

	for _, seg := range target.Pattern.Segments {
		switch s := seg.(type) {
		case *ast.LiteralSegment:
			resolved.WriteString(s.Text)
			if !hasCapture {
				prefix.WriteString(s.Text)
			}
		case *ast.BraceExpr:
			// Check if this is a defined variable
			if ctx != nil {
				if val, ok := ctx.Get(s.Identifier); ok {
					// It's a variable - resolve it
					resolved.WriteString(val)
					if !hasCapture {
						prefix.WriteString(val)
					}
					continue
				}
			}
			// Not a variable - treat as capture
			cp.Captures = append(cp.Captures, s.Identifier)
			hasCapture = true
			resolved.WriteString("{" + s.Identifier + "}")
		}
	}

	cp.ResolvedPath = resolved.String()
	cp.PatternStr = cp.ResolvedPath
	cp.Prefix = prefix.String()
	cp.IsLiteral = len(cp.Captures) == 0

	return cp
}

// compilePattern creates a CompiledPattern from a target (without context).
// Kept for backward compatibility with tests that don't need interpolation.
func compilePattern(target *ast.Target) *CompiledPattern {
	return compilePatternWithContext(target, nil)
}

// extractPrefix returns a prefix key for indexing.
// Uses the first path segment (up to first /) for grouping.
func extractPrefix(path string) string {
	if path == "" {
		return ""
	}

	// Find first slash
	idx := strings.Index(path, "/")
	if idx > 0 {
		return path[:idx]
	}

	// No slash - use first 4 chars or entire string
	if len(path) <= 4 {
		return path
	}
	return path[:4]
}

// extractPhonyName gets the name of a phony target.
func extractPhonyName(pattern *ast.TargetPattern) string {
	return segmentsToString(pattern.Segments)
}
