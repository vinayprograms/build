package eval

import (
	"path/filepath"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
)

// CommandContext extends Context with automatic variables and captures
// for command interpolation during recipe execution.
type CommandContext struct {
	parent    *Context
	automatic map[string]string
	captures  map[string]string
}

// NewCommandContext creates a command context with automatic variables set.
func NewCommandContext(parent *Context, target string, deps []string) *CommandContext {
	ctx := &CommandContext{
		parent:    parent,
		automatic: make(map[string]string),
		captures:  make(map[string]string),
	}

	// Set automatic variables
	ctx.automatic["target"] = target
	ctx.automatic["out"] = target

	// deps is space-separated list
	ctx.automatic["deps"] = strings.Join(deps, " ")

	// in is first dependency
	if len(deps) > 0 {
		ctx.automatic["in"] = deps[0]
	} else {
		ctx.automatic["in"] = ""
	}

	// target.dir and target.file
	ctx.setTargetDirAndFile(target)

	return ctx
}

// setTargetDirAndFile sets target.dir and target.file automatic variables.
func (c *CommandContext) setTargetDirAndFile(target string) {
	// Handle directory targets (ending with /)
	if strings.HasSuffix(target, "/") {
		// For directory targets, target.dir is the directory (without slash)
		// and target.file is empty
		c.automatic["target.dir"] = strings.TrimSuffix(target, "/")
		c.automatic["target.file"] = ""
		return
	}

	dir := filepath.Dir(target)
	file := filepath.Base(target)

	c.automatic["target.dir"] = dir
	c.automatic["target.file"] = file
}

// SetStem sets the stem automatic variable (for pattern targets).
func (c *CommandContext) SetStem(stem string) {
	c.automatic["stem"] = stem
}

// SetCaptures sets capture variables from pattern matching.
func (c *CommandContext) SetCaptures(captures map[string]string) {
	for k, v := range captures {
		c.captures[k] = v
	}
}

// Get retrieves a variable value.
// Priority: automatic > captures > parent context
func (c *CommandContext) Get(name string) (string, bool) {
	// Check automatic variables first
	if val, ok := c.automatic[name]; ok {
		return val, true
	}

	// Check captures
	if val, ok := c.captures[name]; ok {
		return val, true
	}

	// Fall back to parent context
	if c.parent != nil {
		return c.parent.Get(name)
	}

	return "", false
}

// GetCapture retrieves a capture variable.
func (c *CommandContext) GetCapture(name string) (string, bool) {
	val, ok := c.captures[name]
	return val, ok
}

// IsDefined returns true if the variable is defined anywhere.
func (c *CommandContext) IsDefined(name string) bool {
	_, ok := c.Get(name)
	return ok
}

// Parent returns the parent evaluation context.
func (c *CommandContext) Parent() *Context {
	return c.parent
}

// ----------------------------------------------------------------------------
// Command Interpolation
// ----------------------------------------------------------------------------

// InterpolateCommand interpolates a line command and returns the resulting string.
func InterpolateCommand(cmd *ast.LineCommand, ctx *CommandContext) (string, error) {
	return interpolateParts(cmd.Parts, ctx)
}

// InterpolateBlockCommand interpolates a block command and returns the resulting string.
func InterpolateBlockCommand(block *ast.BlockCommand, ctx *CommandContext) (string, error) {
	var lines []string

	for _, lineParts := range block.Lines {
		line, err := interpolateParts(lineParts, ctx)
		if err != nil {
			return "", err
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n"), nil
}

// interpolateParts interpolates a slice of command parts.
func interpolateParts(parts []ast.CommandPart, ctx *CommandContext) (string, error) {
	var result strings.Builder

	for _, part := range parts {
		switch p := part.(type) {
		case *ast.LiteralCommand:
			result.WriteString(p.Text)

		case *ast.CommandInterpolation:
			val, ok := ctx.Get(p.Name)
			if !ok {
				return "", &UndefinedVariableError{
					Name:     p.Name,
					Location: p.Location,
				}
			}

			if p.Raw {
				// Raw: no quoting
				result.WriteString(val)
			} else {
				// Default: shell-quoted
				result.WriteString(ShellQuote(val))
			}
		}
	}

	return result.String(), nil
}

// ----------------------------------------------------------------------------
// Shell Quoting
// ----------------------------------------------------------------------------

// ShellQuote quotes a string for safe use in shell commands.
// It uses single quotes and handles embedded single quotes.
func ShellQuote(s string) string {
	// If string contains single quotes, we need special handling
	if strings.Contains(s, "'") {
		// Replace each ' with '"'"' (end quote, double-quoted quote, start quote)
		escaped := strings.ReplaceAll(s, "'", `'"'"'`)
		return "'" + escaped + "'"
	}

	// Simple case: wrap in single quotes
	return "'" + s + "'"
}
