package eval

import (
	"runtime"
	"sync"

	"github.com/vinayprograms/need/internal/ast"
)

// Context holds the evaluation context for a Needfile.
// It stores evaluated variables, lazy variables, and built-in values.
type Context struct {
	mu sync.RWMutex

	// workDir is the working directory for shell commands (typically the Needfile's directory).
	// If empty, commands run in the current process directory.
	workDir string

	// variables holds evaluated variable values.
	// Keys are variable names, values are the evaluated string values.
	variables map[string]string

	// lazyVariables holds unevaluated lazy variable AST values.
	// These are evaluated on-demand when first referenced.
	lazyVariables map[string]*ast.Value

	// lazyCache holds cached evaluations of lazy variables.
	// Once a lazy variable is evaluated, the result is cached here.
	lazyCache map[string]string

	// builtins holds read-only built-in variables (os, arch).
	builtins map[string]string

	// shellCache holds cached shell() function results.
	// Keys are the evaluated command strings, values are the output.
	// Only successful executions are cached; errors are not cached.
	shellCache map[string]string
}

// NewContext creates a new evaluation context with built-in variables initialized.
func NewContext() *Context {
	return &Context{
		variables:     make(map[string]string),
		lazyVariables: make(map[string]*ast.Value),
		lazyCache:     make(map[string]string),
		shellCache:    make(map[string]string),
		builtins: map[string]string{
			"os":   runtime.GOOS,
			"arch": runtime.GOARCH,
		},
	}
}

// SetWorkDir sets the working directory for shell commands.
func (c *Context) SetWorkDir(dir string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.workDir = dir
}

// WorkDir returns the working directory for shell commands.
func (c *Context) WorkDir() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.workDir
}

// Get retrieves the value of a variable.
// It returns the value and true if the variable is defined, or ("", false) otherwise.
// Built-in variables take precedence over user-defined variables.
// Note: For lazy variables, use GetOrEvaluateLazy with an evaluator.
func (c *Context) Get(name string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check built-ins first (they take precedence)
	if val, ok := c.builtins[name]; ok {
		return val, true
	}

	// Check user-defined variables
	if val, ok := c.variables[name]; ok {
		return val, true
	}

	// Check lazy cache
	if val, ok := c.lazyCache[name]; ok {
		return val, true
	}

	return "", false
}

// Set sets the value of a variable.
// Built-in variables cannot be overwritten.
func (c *Context) Set(name, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Protect built-ins from being overwritten
	if _, isBuiltin := c.builtins[name]; isBuiltin {
		return
	}

	c.variables[name] = value
}

// IsDefined returns true if the variable is defined (user, lazy, or built-in).
func (c *Context) IsDefined(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if _, ok := c.builtins[name]; ok {
		return true
	}
	if _, ok := c.variables[name]; ok {
		return true
	}
	if _, ok := c.lazyVariables[name]; ok {
		return true
	}
	if _, ok := c.lazyCache[name]; ok {
		return true
	}
	return false
}

// SetLazyValue stores a lazy variable's AST value for on-demand evaluation.
// The value is stored unevaluated and will be evaluated when first referenced.
func (c *Context) SetLazyValue(name string, value *ast.Value) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lazyVariables[name] = value
}

// SetLazy stores a lazy variable for on-demand evaluation (legacy string version).
// Deprecated: Use SetLazyValue for full AST support.
func (c *Context) SetLazy(name, unevaluatedValue string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// For backward compatibility, create a simple literal value
	c.lazyVariables[name] = &ast.Value{
		Parts: []ast.ValuePart{
			&ast.LiteralValue{Text: unevaluatedValue},
		},
	}
}

// GetLazyValue retrieves a lazy variable's AST value.
// Returns the value and true if it's a lazy variable, or (nil, false) otherwise.
func (c *Context) GetLazyValue(name string) (*ast.Value, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.lazyVariables[name]
	return val, ok
}

// GetLazy retrieves a lazy variable's unevaluated value (legacy string version).
// For lazy variables stored with SetLazyValue, returns "__lazy__" as a marker.
func (c *Context) GetLazy(name string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if _, ok := c.lazyVariables[name]; ok {
		return "__lazy__", true
	}
	return "", false
}

// IsLazy returns true if the variable is a lazy variable (not yet evaluated).
func (c *Context) IsLazy(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, ok := c.lazyVariables[name]
	return ok
}

// CacheLazyResult caches the result of a lazy variable evaluation.
func (c *Context) CacheLazyResult(name, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lazyCache[name] = value
}

// GetLazyCache retrieves a cached lazy variable result.
// Returns the value and true if cached, or ("", false) otherwise.
func (c *Context) GetLazyCache(name string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.lazyCache[name]
	return val, ok
}

// Variables returns a copy of all evaluated variables (including built-ins).
func (c *Context) Variables() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]string, len(c.variables)+len(c.builtins)+len(c.lazyCache))

	// Copy built-ins
	for k, v := range c.builtins {
		result[k] = v
	}

	// Copy user variables
	for k, v := range c.variables {
		result[k] = v
	}

	// Copy cached lazy values
	for k, v := range c.lazyCache {
		result[k] = v
	}

	return result
}

// LazyVariables returns a copy of all lazy variable names.
func (c *Context) LazyVariables() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]string, len(c.lazyVariables))
	for k := range c.lazyVariables {
		result[k] = "__lazy__"
	}
	return result
}

// ----------------------------------------------------------------------------
// Shell Cache Operations
// ----------------------------------------------------------------------------

// GetShellCache retrieves a cached shell() result.
// Returns the cached output and true if found, or ("", false) otherwise.
func (c *Context) GetShellCache(cmd string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.shellCache[cmd]
	return val, ok
}

// SetShellCache stores a shell() result in the cache.
// Only successful executions should be cached; errors should not be cached.
func (c *Context) SetShellCache(cmd, output string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.shellCache[cmd] = output
}

// ClearShellCache clears all cached shell() results.
// This can be used between builds or when the environment changes.
func (c *Context) ClearShellCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.shellCache = make(map[string]string)
}
