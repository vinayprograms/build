package planner

import (
	"strings"

	"github.com/vinayprograms/build/internal/ast"
)

// CompiledPattern is a pre-processed target pattern for faster matching.
type CompiledPattern struct {
	Target     *ast.Target
	Prefix     string   // Literal prefix before first capture (empty if starts with capture)
	IsLiteral  bool     // True if no captures
	IsPhony    bool     // True if phony target
	PatternStr string   // String representation for debugging
	Captures   []string // Capture names in order
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
func NewTargetIndex(targets []*ast.Target) *TargetIndex {
	idx := &TargetIndex{
		literals: make(map[string]*ast.Target),
		phonies:  make(map[string]*ast.Target),
		byPrefix: make(map[string][]*CompiledPattern),
		all:      make([]*CompiledPattern, 0, len(targets)),
	}

	for _, target := range targets {
		cp := compilePattern(target)
		idx.all = append(idx.all, cp)

		if cp.IsPhony {
			// Extract phony name
			name := extractPhonyName(&target.Pattern)
			idx.phonies[name] = target
		} else if cp.IsLiteral {
			// Exact match path
			idx.literals[segmentsToString(target.Pattern.Segments)] = target
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

// compilePattern creates a CompiledPattern from a target.
func compilePattern(target *ast.Target) *CompiledPattern {
	cp := &CompiledPattern{
		Target:     target,
		IsPhony:    target.Pattern.IsPhony,
		PatternStr: segmentsToString(target.Pattern.Segments),
	}

	// Check if literal (no captures)
	cp.IsLiteral = isLiteralPattern(&target.Pattern)

	// Extract prefix and capture names
	var prefix strings.Builder
	for _, seg := range target.Pattern.Segments {
		switch s := seg.(type) {
		case *ast.LiteralSegment:
			if len(cp.Captures) == 0 {
				prefix.WriteString(s.Text)
			}
		case *ast.BraceExpr:
			cp.Captures = append(cp.Captures, s.Identifier)
		}
	}
	cp.Prefix = prefix.String()

	return cp
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
