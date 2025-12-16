package eval

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
)

// evaluateFunction evaluates a function call and returns the result.
func (e *Evaluator) evaluateFunction(call *ast.FunctionCall) (string, error) {
	switch call.Name {
	case ast.FuncShell:
		return e.funcShell(call)
	case ast.FuncGlob:
		return e.funcGlob(call)
	case ast.FuncBasename:
		return e.funcBasename(call)
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
func (e *Evaluator) funcShell(call *ast.FunctionCall) (string, error) {
	if len(call.Args) < 1 {
		return "", fmt.Errorf("shell() requires at least one argument")
	}

	// Evaluate the command
	cmd, err := e.EvaluateValue(call.Args[0])
	if err != nil {
		return "", err
	}

	// Execute via shell
	shellCmd := exec.Command("/bin/sh", "-c", cmd)
	output, err := shellCmd.Output()
	if err != nil {
		return "", &ShellError{Command: cmd, Err: err}
	}

	// Trim trailing newline (but preserve internal newlines)
	result := string(output)
	result = strings.TrimSuffix(result, "\n")
	result = strings.TrimSuffix(result, "\r\n")

	return result, nil
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

// funcBasename extracts the filename from a path.
func (e *Evaluator) funcBasename(call *ast.FunctionCall) (string, error) {
	if len(call.Args) < 1 {
		return "", fmt.Errorf("basename() requires at least one argument")
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
