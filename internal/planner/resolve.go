package planner

import (
	"fmt"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/eval"
)

// UndefinedVariableError is returned when a variable in a dependency pattern
// cannot be resolved.
type UndefinedVariableError struct {
	Name string
}

func (e *UndefinedVariableError) Error() string {
	return fmt.Sprintf("undefined variable '%s' in dependency", e.Name)
}

// ResolveDependency resolves a single dependency pattern to a concrete path.
//
// Resolution order for {name} in dependency patterns:
//  1. If name is in captures (from pattern matching), use capture value
//  2. If name is defined in context (user variable or built-in), use variable value
//  3. Otherwise, return error for undefined variable
//
// This function is used during build planning to convert pattern-based
// dependencies into concrete file paths.
func ResolveDependency(dep ast.Dependency, captures map[string]string, ctx *eval.Context) (string, error) {
	var sb strings.Builder

	for _, seg := range dep.Segments {
		switch s := seg.(type) {
		case *ast.LiteralSegment:
			sb.WriteString(s.Text)

		case *ast.BraceExpr:
			name := s.Identifier

			// 1. Check captures first (from pattern matching)
			if captures != nil {
				if val, ok := captures[name]; ok {
					sb.WriteString(val)
					continue
				}
			}

			// 2. Check evaluation context (variables and built-ins)
			if val, ok := ctx.Get(name); ok {
				sb.WriteString(val)
				continue
			}

			// 3. Undefined variable
			return "", &UndefinedVariableError{Name: name}

		default:
			// Unknown segment type - skip
		}
	}

	return sb.String(), nil
}

// ResolveDependencies resolves multiple dependencies to concrete paths.
//
// It processes each dependency in order and returns a slice of resolved paths.
// If any dependency fails to resolve, an error is returned immediately.
//
// Special handling: if a dependency consists of exactly one BraceExpr (e.g., {objects}),
// and the resolved value contains spaces, it is split into multiple paths.
// This allows variables like "objects = build/main.o build/utils.o" to expand
// into multiple dependencies.
func ResolveDependencies(deps []ast.Dependency, captures map[string]string, ctx *eval.Context) ([]string, error) {
	if len(deps) == 0 {
		return []string{}, nil
	}

	paths := make([]string, 0, len(deps))
	for _, dep := range deps {
		// Check if this dependency is exactly one BraceExpr (variable expansion)
		isSingleVar := len(dep.Segments) == 1
		if isSingleVar {
			if _, ok := dep.Segments[0].(*ast.BraceExpr); !ok {
				isSingleVar = false
			}
		}

		path, err := ResolveDependency(dep, captures, ctx)
		if err != nil {
			return nil, err
		}

		// If it's a single variable expansion, split on spaces
		if isSingleVar && strings.ContainsAny(path, " \t") {
			splitPaths := strings.Fields(path)
			paths = append(paths, splitPaths...)
		} else {
			paths = append(paths, path)
		}
	}

	return paths, nil
}
