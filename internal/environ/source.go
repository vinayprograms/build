package environ

import (
	"fmt"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/eval"
)

// ResolveSourcePath resolves a directive's path value (e.g. a .source:
// directive) to its final string using ctx, interpolating any variables
// defined earlier in the Buildfile. directive names the offending directive
// in error messages (e.g. ".source:"); ctx must be the same evaluation
// context used to run the rest of the Buildfile, so automatic variables and
// undefined variables surface the same errors they would anywhere else.
//
// Returns ("", nil) if v is nil (the directive wasn't given).
func ResolveSourcePath(directive string, v *ast.Value, ctx *eval.Context) (string, error) {
	if v == nil {
		return "", nil
	}
	resolved, err := eval.NewEvaluator(ctx).EvaluateValue(v)
	if err != nil {
		return "", fmt.Errorf("%s %w", directive, err)
	}
	return strings.TrimSpace(resolved), nil
}
