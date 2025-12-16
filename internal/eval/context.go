package eval

import (
	"runtime"
	"sync"
)

// Context holds the evaluation context for a Buildfile.
// It stores evaluated variables, lazy variables, and built-in values.
type Context struct {
	mu sync.RWMutex

	// variables holds evaluated variable values.
	// Keys are variable names, values are the evaluated string values.
	variables map[string]string

	// lazyVariables holds unevaluated lazy variable values.
	// These are stored as strings for now (the AST representation).
	// They will be evaluated on-demand when referenced.
	lazyVariables map[string]string

	// builtins holds read-only built-in variables (os, arch).
	builtins map[string]string
}

// NewContext creates a new evaluation context with built-in variables initialized.
func NewContext() *Context {
	return &Context{
		variables:     make(map[string]string),
		lazyVariables: make(map[string]string),
		builtins: map[string]string{
			"os":   runtime.GOOS,
			"arch": runtime.GOARCH,
		},
	}
}

// Get retrieves the value of a variable.
// It returns the value and true if the variable is defined, or ("", false) otherwise.
// Built-in variables take precedence over user-defined variables.
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
	return false
}

// SetLazy stores a lazy variable for on-demand evaluation.
// The value is stored unevaluated and will be evaluated when first referenced.
func (c *Context) SetLazy(name, unevaluatedValue string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lazyVariables[name] = unevaluatedValue
}

// GetLazy retrieves a lazy variable's unevaluated value.
// Returns the value and true if it's a lazy variable, or ("", false) otherwise.
func (c *Context) GetLazy(name string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.lazyVariables[name]
	return val, ok
}

// IsLazy returns true if the variable is a lazy variable.
func (c *Context) IsLazy(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, ok := c.lazyVariables[name]
	return ok
}

// Variables returns a copy of all evaluated variables (including built-ins).
func (c *Context) Variables() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]string, len(c.variables)+len(c.builtins))

	// Copy built-ins
	for k, v := range c.builtins {
		result[k] = v
	}

	// Copy user variables (may shadow built-ins in the returned map, but Get() protects them)
	for k, v := range c.variables {
		result[k] = v
	}

	return result
}

// LazyVariables returns a copy of all lazy variable definitions.
func (c *Context) LazyVariables() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]string, len(c.lazyVariables))
	for k, v := range c.lazyVariables {
		result[k] = v
	}
	return result
}
