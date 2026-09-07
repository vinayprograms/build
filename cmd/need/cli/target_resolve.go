package cli

import (
	"errors"
	"strings"

	"github.com/vinayprograms/need/internal/ast"
)

// ResolveTargetArgs resolves command-line target arguments to canonical target names.
// If args is nil or empty, uses .default: directive or first defined target.
// The ctx parameter is used to resolve variable interpolations in target patterns.
// Returns resolved target names (with @ prefix for phony targets) or error.
func ResolveTargetArgs(args []string, result NeedfileResult, ctx EvalContext) ([]string, error) {
	statements := GetASTStatements(result)
	if statements == nil {
		return nil, errors.New("no targets defined")
	}

	// Extract targets and default directive from statements
	targets, defaultTarget := extractTargetsAndDefault(statements)

	if len(args) == 0 {
		return resolveDefaultTarget(targets, defaultTarget, ctx)
	}

	return resolveExplicitTargets(args, targets, ctx)
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
func resolveDefaultTarget(targets []*ast.Target, defaultTarget string, ctx EvalContext) ([]string, error) {
	if len(targets) == 0 {
		return nil, errors.New("no targets defined")
	}

	// If .default: directive is present, resolve it
	if defaultTarget != "" {
		resolved, err := resolveTargetName(defaultTarget, targets, ctx)
		if err != nil {
			return nil, err
		}
		return []string{resolved}, nil
	}

	// Otherwise, use first target
	first := targets[0]
	return []string{formatTargetName(first, ctx)}, nil
}

// resolveExplicitTargets resolves a list of explicit target arguments.
func resolveExplicitTargets(args []string, targets []*ast.Target, ctx EvalContext) ([]string, error) {
	resolved := make([]string, 0, len(args))

	for _, arg := range args {
		name, err := resolveTargetName(arg, targets, ctx)
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
func resolveTargetName(arg string, targets []*ast.Target, ctx EvalContext) (string, error) {
	// First, try exact match with @ prefix intact
	if strings.HasPrefix(arg, "@") {
		for _, t := range targets {
			if t.Pattern.IsPhony && formatTargetName(t, ctx) == arg {
				return arg, nil
			}
		}
		return "", targetNotFoundError(arg)
	}

	// Try exact file target match first
	for _, t := range targets {
		if !t.Pattern.IsPhony && matchTargetPattern(t, arg, ctx) {
			return arg, nil
		}
	}

	// Try as phony target (without @)
	for _, t := range targets {
		if t.Pattern.IsPhony {
			phonyName := getPhonyName(t, ctx)
			if phonyName == arg {
				return "@" + arg, nil
			}
		}
	}

	// Try pattern targets
	for _, t := range targets {
		if hasCapturesOnly(t, ctx) && matchTargetPattern(t, arg, ctx) {
			return arg, nil
		}
	}

	return "", targetNotFoundError(arg)
}

// formatTargetName returns the canonical name for a target.
// BraceExpr nodes are resolved using the context if they refer to defined variables.
func formatTargetName(t *ast.Target, ctx EvalContext) string {
	var name string
	if t.Pattern.IsPhony {
		name = "@"
	}
	for _, seg := range t.Pattern.Segments {
		switch s := seg.(type) {
		case *ast.LiteralSegment:
			name += s.Text
		case *ast.BraceExpr:
			// Try to resolve as variable; if not defined, keep as capture
			if ctx != nil {
				if val, ok := ctx.Get(s.Identifier); ok {
					name += val
					continue
				}
			}
			name += "{" + s.Identifier + "}"
		}
	}
	return name
}

// getPhonyName returns the name part of a phony target (without @).
func getPhonyName(t *ast.Target, ctx EvalContext) string {
	var name string
	for _, seg := range t.Pattern.Segments {
		switch s := seg.(type) {
		case *ast.LiteralSegment:
			name += s.Text
		case *ast.BraceExpr:
			if ctx != nil {
				if val, ok := ctx.Get(s.Identifier); ok {
					name += val
					continue
				}
			}
			name += "{" + s.Identifier + "}"
		}
	}
	return name
}

// matchTargetPattern checks if a path matches a target pattern.
func matchTargetPattern(t *ast.Target, path string, ctx EvalContext) bool {
	// For literal targets (after resolving interpolations), do exact match
	if !hasCapturesOnly(t, ctx) {
		pattern := formatTargetName(t, ctx)
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

// hasCapturesOnly returns true if the target has capture expressions that are not defined variables.
func hasCapturesOnly(t *ast.Target, ctx EvalContext) bool {
	for _, seg := range t.Pattern.Segments {
		if be, ok := seg.(*ast.BraceExpr); ok {
			// If context is available and this is a defined variable, it's not a capture
			if ctx != nil {
				if _, ok := ctx.Get(be.Identifier); ok {
					continue
				}
			}
			// This is a capture (undefined variable)
			return true
		}
	}
	return false
}

// targetNotFoundError creates an error for a missing target.
func targetNotFoundError(name string) error {
	return errors.New("target '" + name + "' not found")
}
