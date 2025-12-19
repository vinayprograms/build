package main

import (
	"errors"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
)

// ResolveTargetArgs resolves command-line target arguments to canonical target names.
// If args is nil or empty, uses .default: directive or first defined target.
// Returns resolved target names (with @ prefix for phony targets) or error.
func ResolveTargetArgs(args []string, result BuildfileResult) ([]string, error) {
	statements := GetASTStatements(result)
	if statements == nil {
		return nil, errors.New("no targets defined")
	}

	// Extract targets and default directive from statements
	targets, defaultTarget := extractTargetsAndDefault(statements)

	if len(args) == 0 {
		return resolveDefaultTarget(targets, defaultTarget)
	}

	return resolveExplicitTargets(args, targets)
}

// extractTargetsAndDefault extracts all targets and the .default: directive from statements.
func extractTargetsAndDefault(statements []ast.Statement) ([]*ast.Target, string) {
	var targets []*ast.Target
	var defaultTarget string

	for _, stmt := range statements {
		switch s := stmt.(type) {
		case *ast.Target:
			targets = append(targets, s)
		case *ast.Directive:
			if s.Kind == ast.DirectiveDefault && s.Value != nil {
				defaultTarget = extractLiteralValue(s.Value)
			}
		}
	}

	return targets, defaultTarget
}

// extractLiteralValue extracts the literal string value from an ast.Value.
func extractLiteralValue(v *ast.Value) string {
	if v == nil {
		return ""
	}
	var text string
	for _, part := range v.Parts {
		if lit, ok := part.(*ast.LiteralValue); ok {
			text += lit.Text
		}
	}
	return strings.TrimSpace(text)
}

// resolveDefaultTarget resolves the default target when no arguments provided.
func resolveDefaultTarget(targets []*ast.Target, defaultTarget string) ([]string, error) {
	if len(targets) == 0 {
		return nil, errors.New("no targets defined")
	}

	// If .default: directive is present, resolve it
	if defaultTarget != "" {
		resolved, err := resolveTargetName(defaultTarget, targets)
		if err != nil {
			return nil, err
		}
		return []string{resolved}, nil
	}

	// Otherwise, use first target
	first := targets[0]
	return []string{formatTargetName(first)}, nil
}

// resolveExplicitTargets resolves a list of explicit target arguments.
func resolveExplicitTargets(args []string, targets []*ast.Target) ([]string, error) {
	resolved := make([]string, 0, len(args))

	for _, arg := range args {
		name, err := resolveTargetName(arg, targets)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, name)
	}

	return resolved, nil
}

// resolveTargetName resolves a single target argument to a canonical name.
// The argument may be:
// - A file target path: "build/app"
// - A phony target with @: "@clean"
// - A phony target without @: "clean" (matches @clean if no file target matches)
func resolveTargetName(arg string, targets []*ast.Target) (string, error) {
	// First, try exact match with @ prefix intact
	if strings.HasPrefix(arg, "@") {
		for _, t := range targets {
			if t.Pattern.IsPhony && formatTargetName(t) == arg {
				return arg, nil
			}
		}
		return "", targetNotFoundError(arg)
	}

	// Try exact file target match first
	for _, t := range targets {
		if !t.Pattern.IsPhony && matchTargetPattern(t, arg) {
			return arg, nil
		}
	}

	// Try as phony target (without @)
	for _, t := range targets {
		if t.Pattern.IsPhony {
			phonyName := getPhonyName(t)
			if phonyName == arg {
				return "@" + arg, nil
			}
		}
	}

	// Try pattern targets
	for _, t := range targets {
		if hasCaptures(t) && matchTargetPattern(t, arg) {
			return arg, nil
		}
	}

	return "", targetNotFoundError(arg)
}

// formatTargetName returns the canonical name for a target.
func formatTargetName(t *ast.Target) string {
	var name string
	if t.Pattern.IsPhony {
		name = "@"
	}
	for _, seg := range t.Pattern.Segments {
		switch s := seg.(type) {
		case *ast.LiteralSegment:
			name += s.Text
		case *ast.BraceExpr:
			name += "{" + s.Identifier + "}"
		}
	}
	return name
}

// getPhonyName returns the name part of a phony target (without @).
func getPhonyName(t *ast.Target) string {
	var name string
	for _, seg := range t.Pattern.Segments {
		if lit, ok := seg.(*ast.LiteralSegment); ok {
			name += lit.Text
		}
	}
	return name
}

// matchTargetPattern checks if a path matches a target pattern.
func matchTargetPattern(t *ast.Target, path string) bool {
	// For literal targets, do exact match
	if !hasCaptures(t) {
		pattern := formatTargetName(t)
		// Strip @ for phony targets
		if t.Pattern.IsPhony {
			pattern = pattern[1:]
		}
		return pattern == path
	}

	// For pattern targets, use the planner's matcher
	result := MatchTargetPattern(&t.Pattern, path)
	return result.Matched()
}

// hasCaptures returns true if the target has capture expressions.
func hasCaptures(t *ast.Target) bool {
	for _, seg := range t.Pattern.Segments {
		if _, ok := seg.(*ast.BraceExpr); ok {
			return true
		}
	}
	return false
}

// targetNotFoundError creates an error for a missing target.
func targetNotFoundError(name string) error {
	return errors.New("target '" + name + "' not found")
}
