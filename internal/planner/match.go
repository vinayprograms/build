package planner

import (
	"fmt"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
)

// TargetNotFoundError is returned when a target path cannot be matched
// against any defined target.
type TargetNotFoundError struct {
	Path string
}

func (e *TargetNotFoundError) Error() string {
	return fmt.Sprintf("no rule to make target '%s'", e.Path)
}

// MatchTarget attempts to match a concrete path against a target pattern.
// Returns true if the pattern matches, along with a map of capture values.
// For literal patterns (no captures), the map will be empty.
//
// Matching rules:
//   - Literal segments must match exactly
//   - Captures match any sequence of characters (including slashes)
//   - Duplicate capture names must have the same value
//   - Phony targets must be matched with @ prefix in path
//   - Directory targets can match with or without trailing slash
func MatchTarget(pattern *ast.TargetPattern, path string) (bool, map[string]string) {
	// Handle phony targets
	if pattern.IsPhony {
		return matchPhonyTarget(pattern, path)
	}

	// Handle directory targets
	if pattern.IsDirectory {
		return matchDirectoryTarget(pattern, path)
	}

	// Normal file target matching
	return matchPattern(pattern.Segments, path)
}

// matchPhonyTarget matches a phony target pattern against a path.
// Phony targets require the path to start with @.
func matchPhonyTarget(pattern *ast.TargetPattern, path string) (bool, map[string]string) {
	if !strings.HasPrefix(path, "@") {
		return false, nil
	}

	// Remove @ prefix and match rest
	phonyPath := path[1:]
	return matchPattern(pattern.Segments, phonyPath)
}

// matchDirectoryTarget matches a directory target pattern against a path.
// Directory targets can match with or without trailing slash.
func matchDirectoryTarget(pattern *ast.TargetPattern, path string) (bool, map[string]string) {
	// Try matching as-is
	if matched, captures := matchPattern(pattern.Segments, path); matched {
		return true, captures
	}

	// Try matching with trailing slash added if path doesn't have one
	if !strings.HasSuffix(path, "/") {
		if matched, captures := matchPattern(pattern.Segments, path+"/"); matched {
			return true, captures
		}
	}

	// Try matching without trailing slash if path has one
	if strings.HasSuffix(path, "/") {
		patternStr := segmentsToString(pattern.Segments)
		if strings.HasSuffix(patternStr, "/") {
			// Remove trailing slash from both and match
			trimmedPath := strings.TrimSuffix(path, "/")
			trimmedPattern := strings.TrimSuffix(patternStr, "/")
			if trimmedPath == trimmedPattern {
				return true, make(map[string]string)
			}
		}
	}

	return false, nil
}

// matchPattern matches pattern segments against a path.
// Returns whether the match succeeded and the captured values.
func matchPattern(segments []ast.PatternSegment, path string) (bool, map[string]string) {
	if len(segments) == 0 {
		// Empty pattern only matches empty path
		return path == "", make(map[string]string)
	}

	// Check if pattern has any captures
	hasCaptures := false
	for _, seg := range segments {
		if _, ok := seg.(*ast.BraceExpr); ok {
			hasCaptures = true
			break
		}
	}

	if !hasCaptures {
		// Pure literal pattern - exact match only
		patternStr := segmentsToString(segments)
		if patternStr == path {
			return true, make(map[string]string)
		}
		return false, nil
	}

	// Pattern matching with captures
	return matchWithCaptures(segments, path)
}

// matchWithCaptures performs pattern matching with capture extraction.
func matchWithCaptures(segments []ast.PatternSegment, path string) (bool, map[string]string) {
	captures := make(map[string]string)
	return matchSegmentsRecursive(segments, path, captures)
}

// matchSegmentsRecursive recursively matches segments against a path.
func matchSegmentsRecursive(segments []ast.PatternSegment, remaining string, captures map[string]string) (bool, map[string]string) {
	if len(segments) == 0 {
		// All segments consumed; match succeeds if no remaining path
		return remaining == "", captures
	}

	seg := segments[0]
	rest := segments[1:]

	switch s := seg.(type) {
	case *ast.LiteralSegment:
		// Literal must match exactly at the start
		if !strings.HasPrefix(remaining, s.Text) {
			return false, nil
		}
		return matchSegmentsRecursive(rest, remaining[len(s.Text):], captures)

	case *ast.BraceExpr:
		// Capture: try to match
		return matchCapture(s, rest, remaining, captures)

	default:
		// Unknown segment type
		return false, nil
	}
}

// matchCapture attempts to match a capture segment.
func matchCapture(capture *ast.BraceExpr, rest []ast.PatternSegment, remaining string, captures map[string]string) (bool, map[string]string) {
	name := capture.Identifier

	// If no remaining segments, capture takes everything
	if len(rest) == 0 {
		// Check if this capture name was used before
		if prev, exists := captures[name]; exists && prev != remaining {
			return false, nil
		}
		captures[name] = remaining
		return true, captures
	}

	// Find what the next literal segment starts with
	nextLiteral := findNextLiteral(rest)

	if nextLiteral == "" {
		// Next is another capture - try greedy approach
		// First capture takes everything, subsequent takes nothing
		return matchCaptureGreedy(name, rest, remaining, captures)
	}

	// Try to find where the next literal starts
	// The capture gets everything up to that point
	return matchCaptureUntilLiteral(name, nextLiteral, rest, remaining, captures)
}

// findNextLiteral returns the text of the next literal segment, or empty string.
func findNextLiteral(segments []ast.PatternSegment) string {
	if len(segments) == 0 {
		return ""
	}
	if lit, ok := segments[0].(*ast.LiteralSegment); ok {
		return lit.Text
	}
	return ""
}

// matchCaptureGreedy handles the case of adjacent captures.
// Uses a greedy approach: first capture takes as much as possible
// while still allowing the rest to match.
func matchCaptureGreedy(name string, rest []ast.PatternSegment, remaining string, captures map[string]string) (bool, map[string]string) {
	// Try progressively shorter captures, starting with taking everything
	for i := len(remaining); i >= 0; i-- {
		captureValue := remaining[:i]
		restOfPath := remaining[i:]

		// Check if this capture name conflicts with existing
		if prev, exists := captures[name]; exists && prev != captureValue {
			continue
		}

		// Copy captures for this attempt
		newCaptures := copyCaptures(captures)
		newCaptures[name] = captureValue

		if matched, result := matchSegmentsRecursive(rest, restOfPath, newCaptures); matched {
			return true, result
		}
	}
	return false, nil
}

// matchCaptureUntilLiteral matches a capture by finding where the next literal starts.
func matchCaptureUntilLiteral(name, nextLiteral string, rest []ast.PatternSegment, remaining string, captures map[string]string) (bool, map[string]string) {
	// Find all positions where nextLiteral appears
	// Try each position, starting from the first (shortest capture)
	pos := 0
	for {
		idx := strings.Index(remaining[pos:], nextLiteral)
		if idx < 0 {
			break
		}

		captureValue := remaining[:pos+idx]
		restOfPath := remaining[pos+idx:]

		// Check if this capture name conflicts with existing
		if prev, exists := captures[name]; exists {
			if prev != captureValue {
				// Try next position
				pos = pos + idx + 1
				continue
			}
		}

		// Copy captures for this attempt
		newCaptures := copyCaptures(captures)
		newCaptures[name] = captureValue

		if matched, result := matchSegmentsRecursive(rest, restOfPath, newCaptures); matched {
			return true, result
		}

		// Try next position
		pos = pos + idx + 1
	}
	return false, nil
}

// copyCaptures creates a copy of the captures map.
func copyCaptures(orig map[string]string) map[string]string {
	cp := make(map[string]string, len(orig))
	for k, v := range orig {
		cp[k] = v
	}
	return cp
}

// segmentsToString converts pattern segments to a string.
func segmentsToString(segments []ast.PatternSegment) string {
	var sb strings.Builder
	for _, seg := range segments {
		switch s := seg.(type) {
		case *ast.LiteralSegment:
			sb.WriteString(s.Text)
		case *ast.BraceExpr:
			sb.WriteString("{")
			sb.WriteString(s.Identifier)
			sb.WriteString("}")
		}
	}
	return sb.String()
}

// LookupTarget finds a target definition that matches the given path.
// It returns the matching target, capture values, and any error.
//
// Lookup order:
//  1. Exact literal matches are preferred over pattern matches
//  2. Among patterns, first match in definition order wins
//
// The path should include @ prefix for phony targets.
func LookupTarget(path string, targets []*ast.Target) (*ast.Target, map[string]string, error) {
	if len(targets) == 0 {
		return nil, nil, &TargetNotFoundError{Path: path}
	}

	// First pass: look for exact literal matches
	for _, target := range targets {
		if isLiteralPattern(&target.Pattern) {
			if matched, captures := MatchTarget(&target.Pattern, path); matched {
				return target, captures, nil
			}
		}
	}

	// Second pass: look for pattern matches
	for _, target := range targets {
		if !isLiteralPattern(&target.Pattern) {
			if matched, captures := MatchTarget(&target.Pattern, path); matched {
				return target, captures, nil
			}
		}
	}

	return nil, nil, &TargetNotFoundError{Path: path}
}

// isLiteralPattern returns true if the pattern has no captures.
func isLiteralPattern(pattern *ast.TargetPattern) bool {
	for _, seg := range pattern.Segments {
		if _, ok := seg.(*ast.BraceExpr); ok {
			return false
		}
	}
	return true
}
