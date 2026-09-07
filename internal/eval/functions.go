package eval

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/platform"
)

// evaluateFunction evaluates a function call and returns the result.
func (e *Evaluator) evaluateFunction(call *ast.FunctionCall) (string, error) {
	switch call.Name {
	case ast.FuncShell:
		return e.funcShell(call)
	case ast.FuncGlob:
		return e.funcGlob(call)
	case ast.FuncFilename:
		return e.funcFilename(call)
	case ast.FuncDirname:
		return e.funcDirname(call)
	case ast.FuncReplace:
		return e.funcReplace(call)
	default:
		return "", fmt.Errorf("unknown function: %v", call.Name)
	}
}

// funcShell executes a shell command and returns stdout.
// Trailing newlines are trimmed.
// Interpolated values are shell-quoted by default; use :raw modifier to disable.
// Results are cached by the evaluated command string within a build.
func (e *Evaluator) funcShell(call *ast.FunctionCall) (string, error) {
	if len(call.Args) < 1 {
		return "", fmt.Errorf("shell() requires at least one argument")
	}

	// Evaluate the command with shell quoting for interpolations
	cmd, err := e.evaluateShellCommand(call.Args[0])
	if err != nil {
		return "", err
	}

	// Check cache first
	if cached, ok := e.ctx.GetShellCache(cmd); ok {
		return cached, nil
	}

	// Execute via shell using platform-appropriate shell and args
	shell := platform.DefaultShell()
	args := platform.ShellCommandArgs(shell, cmd)
	shellCmd := exec.Command(shell, args...)

	// Set working directory if specified
	if workDir := e.ctx.WorkDir(); workDir != "" {
		shellCmd.Dir = workDir
	}

	output, err := shellCmd.Output()
	if err != nil {
		// Do NOT cache errors - allow retry on failure
		return "", &ShellError{Command: cmd, Err: err}
	}

	// Trim trailing newline (but preserve internal newlines)
	result := string(output)
	result = strings.TrimSuffix(result, "\n")
	result = strings.TrimSuffix(result, "\r\n")

	// Cache the successful result
	e.ctx.SetShellCache(cmd, result)

	return result, nil
}

// evaluateShellCommand evaluates a value for use as a shell command.
// Unlike EvaluateValue, this applies context-aware shell quoting to non-raw
// interpolations: bare/single/double-quoted state is tracked left to right
// over the literal text, same as recipe command interpolation (see
// command.go's InterpolateCommand and quote.go).
func (e *Evaluator) evaluateShellCommand(val *ast.Value) (string, error) {
	if val == nil {
		return "", nil
	}

	var result strings.Builder
	scanner := newQuoteScanner()

	for _, part := range val.Parts {
		switch p := part.(type) {
		case *ast.LiteralValue:
			result.WriteString(p.Text)
			scanner.scanLiteral(p.Text)

		case *ast.Interpolation:
			resolved, err := e.resolveVariable(p.Name, p.Location)
			if err != nil {
				return "", err
			}
			if p.Raw {
				// Raw: no quoting, allows word splitting
				result.WriteString(resolved)
			} else {
				result.WriteString(formatForContext(scanner.current(), resolved))
			}

		case *ast.FunctionCall:
			evaluated, err := e.evaluateFunction(p)
			if err != nil {
				return "", err
			}
			result.WriteString(evaluated)
			scanner.scanLiteral(evaluated)

		default:
			// Unknown part type - skip
		}
	}

	return result.String(), nil
}

// funcGlob matches files against a pattern.
// Returns space-separated list of matches.
func (e *Evaluator) funcGlob(call *ast.FunctionCall) (string, error) {
	if len(call.Args) < 1 {
		return "", fmt.Errorf("glob() requires at least one argument")
	}

	// Evaluate the pattern
	pattern, err := e.EvaluateValue(call.Args[0])
	if err != nil {
		return "", err
	}

	// Match files
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("glob pattern error: %w", err)
	}

	// Return space-separated list
	return strings.Join(matches, " "), nil
}

// funcFilename extracts the filename from a path.
func (e *Evaluator) funcFilename(call *ast.FunctionCall) (string, error) {
	if len(call.Args) < 1 {
		return "", fmt.Errorf("filename() requires at least one argument")
	}

	// Evaluate the path
	path, err := e.EvaluateValue(call.Args[0])
	if err != nil {
		return "", err
	}

	return filepath.Base(path), nil
}

// funcDirname extracts the directory from a path.
func (e *Evaluator) funcDirname(call *ast.FunctionCall) (string, error) {
	if len(call.Args) < 1 {
		return "", fmt.Errorf("dirname() requires at least one argument")
	}

	// Evaluate the path
	path, err := e.EvaluateValue(call.Args[0])
	if err != nil {
		return "", err
	}

	return filepath.Dir(path), nil
}

// funcReplace replaces all occurrences of 'from' with 'to' in the input.
func (e *Evaluator) funcReplace(call *ast.FunctionCall) (string, error) {
	if len(call.Args) < 3 {
		return "", fmt.Errorf("replace() requires three arguments: input, from, to")
	}

	// Evaluate all arguments
	input, err := e.EvaluateValue(call.Args[0])
	if err != nil {
		return "", err
	}

	from, err := e.EvaluateValue(call.Args[1])
	if err != nil {
		return "", err
	}

	to, err := e.EvaluateValue(call.Args[2])
	if err != nil {
		return "", err
	}

	// Replace all occurrences
	return strings.ReplaceAll(input, from, to), nil
}

// ----------------------------------------------------------------------------
// Error Types
// ----------------------------------------------------------------------------

// ShellError is returned when a shell command fails.
type ShellError struct {
	Command string
	Err     error
}

func (e *ShellError) Error() string {
	return fmt.Sprintf("shell command failed: %s: %v", e.Command, e.Err)
}

func (e *ShellError) Unwrap() error {
	return e.Err
}
