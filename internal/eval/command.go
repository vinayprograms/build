package eval

import (
	"path/filepath"
	"strings"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/platform"
)

// CommandContext extends Context with automatic variables and captures
// for command interpolation during recipe execution.
type CommandContext struct {
	parent    *Context
	evaluator *Evaluator
	automatic map[string]string
	captures  map[string]string
	deps      []string // raw dependency list (for context-aware {deps} quoting)
}

// NewCommandContext creates a command context with automatic variables set.
func NewCommandContext(parent *Context, target string, deps []string) *CommandContext {
	ctx := &CommandContext{
		parent:    parent,
		evaluator: NewEvaluator(parent),
		automatic: make(map[string]string),
		captures:  make(map[string]string),
	}

	// Set automatic variables
	ctx.automatic["target"] = target
	ctx.automatic["out"] = target

	// deps is the raw space-separated list; interpolation quotes each item
	// individually in the bare (unquoted) shell position and escapes the
	// joined list inside quotes (see interpolateParts).
	ctx.automatic["deps"] = strings.Join(deps, " ")
	ctx.deps = deps

	// in is first dependency (not pre-quoted, will be quoted during interpolation)
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
	// Handle phony targets (starting with @)
	if strings.HasPrefix(target, "@") {
		// For phony targets, target.dir is "." and target.file is the name without @
		c.automatic["target.dir"] = "."
		c.automatic["target.file"] = target[1:]
		return
	}

	// Handle directory targets (ending with / or \)
	if platform.IsDirectoryPath(target) {
		// For directory targets, target.dir is the directory (without separator)
		// and target.file is empty
		c.automatic["target.dir"] = strings.TrimSuffix(strings.TrimSuffix(target, "/"), "\\")
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
// Priority: automatic > captures > parent context (including lazy variables)
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
		// First try regular Get
		if val, ok := c.parent.Get(name); ok {
			return val, true
		}

		// Check if it's a lazy variable and evaluate it
		if c.parent.IsLazy(name) && c.evaluator != nil {
			// Use a dummy location since we don't have the actual location here
			val, err := c.evaluator.evaluateLazyVariable(name, ast.SourceLocation{})
			if err == nil {
				return val, true
			}
		}
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
	return interpolateParts(cmd.Parts, ctx, newQuoteScanner())
}

// InterpolateBlockCommand interpolates a block command and returns the
// resulting string. Shell quote state (open double/single quotes, an
// in-progress heredoc body, ...) carries over from one line to the next,
// since a block is executed as a single, continuous shell script.
func InterpolateBlockCommand(block *ast.BlockCommand, ctx *CommandContext) (string, error) {
	var lines []string
	scanner := newQuoteScanner()

	for _, lineParts := range block.Lines {
		scanner.activatePendingHeredoc()
		scanner.checkHeredocTerminator(linePlainText(lineParts))

		line, err := interpolateParts(lineParts, ctx, scanner)
		if err != nil {
			return "", err
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n"), nil
}

// linePlainText concatenates a line's literal text (ignoring any
// interpolations) and reports whether the line contains an interpolation
// at all. It is used only to recognize a heredoc terminator line, which is
// always bare text.
func linePlainText(parts []ast.CommandPart) (text string, hasInterp bool) {
	var b strings.Builder
	for _, part := range parts {
		switch p := part.(type) {
		case *ast.LiteralCommand:
			b.WriteString(p.Text)
		case *ast.CommandInterpolation:
			hasInterp = true
		}
	}
	return b.String(), hasInterp
}

// interpolateParts interpolates a slice of command parts, formatting each
// interpolated value according to the shell quoting context tracked by
// scanner (see quote.go for the rules).
func interpolateParts(parts []ast.CommandPart, ctx *CommandContext, scanner *quoteScanner) (string, error) {
	var result strings.Builder

	for _, part := range parts {
		switch p := part.(type) {
		case *ast.LiteralCommand:
			result.WriteString(p.Text)
			scanner.scanLiteral(p.Text)

		case *ast.CommandInterpolation:
			val, ok := ctx.Get(p.Name)
			if !ok {
				return "", &UndefinedVariableError{
					Name:     p.Name,
					Location: p.Location,
				}
			}

			switch {
			case p.Raw:
				result.WriteString(val)
			case p.Name == "deps" && scanner.current() == ctxUnquoted:
				// Bare {deps} must expand to one shell word per dependency.
				result.WriteString(ShellQuoteList(ctx.deps))
			default:
				result.WriteString(formatForContext(scanner.current(), val))
			}
		}
	}

	return result.String(), nil
}

// ----------------------------------------------------------------------------
// Shell Quoting
// ----------------------------------------------------------------------------

// ShellQuote quotes a string for safe use in an unquoted (bare) shell
// position. It emits the value bare if it contains only characters from
// [A-Za-z0-9_./:@%+=,-]; otherwise it wraps it in single quotes, with an
// embedded ' emitted as '\”. See quote.go for the full context-aware
// interpolation rule this is one case of.
func ShellQuote(s string) string {
	return formatUnquoted(s)
}

// ShellQuoteList quotes each item in a space-separated list individually.
// Returns a space-separated string of quoted items.
func ShellQuoteList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = ShellQuote(item)
	}
	return strings.Join(quoted, " ")
}
