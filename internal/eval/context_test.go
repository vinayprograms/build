package eval

import (
	"runtime"
	"testing"
)

// ----------------------------------------------------------------------------
// EvaluationContext Tests
// ----------------------------------------------------------------------------

func TestNewContext(t *testing.T) {
	ctx := NewContext()

	if ctx == nil {
		t.Fatal("NewContext returned nil")
	}

	// Should have os and arch built-in
	os, ok := ctx.Get("os")
	if !ok {
		t.Error("Expected 'os' built-in to be defined")
	}
	if os != runtime.GOOS {
		t.Errorf("Expected os=%s, got %s", runtime.GOOS, os)
	}

	arch, ok := ctx.Get("arch")
	if !ok {
		t.Error("Expected 'arch' built-in to be defined")
	}
	if arch != runtime.GOARCH {
		t.Errorf("Expected arch=%s, got %s", runtime.GOARCH, arch)
	}
}

func TestContext_SetAndGet(t *testing.T) {
	ctx := NewContext()

	ctx.Set("foo", "bar")

	val, ok := ctx.Get("foo")
	if !ok {
		t.Error("Expected 'foo' to be defined after Set")
	}
	if val != "bar" {
		t.Errorf("Expected value 'bar', got '%s'", val)
	}
}

func TestContext_GetUndefined(t *testing.T) {
	ctx := NewContext()

	_, ok := ctx.Get("undefined")
	if ok {
		t.Error("Expected undefined variable to return ok=false")
	}
}

func TestContext_IsDefined(t *testing.T) {
	ctx := NewContext()
	ctx.Set("defined", "value")

	tests := []struct {
		name   string
		expect bool
	}{
		{"defined", true},
		{"undefined", false},
		{"os", true},
		{"arch", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ctx.IsDefined(tt.name) != tt.expect {
				t.Errorf("IsDefined(%s) = %v, want %v", tt.name, !tt.expect, tt.expect)
			}
		})
	}
}

func TestContext_SetLazy(t *testing.T) {
	ctx := NewContext()

	// SetLazy stores the unevaluated value
	ctx.SetLazy("lazy_var", "unevaluated value")

	// Should not be in regular variables
	_, ok := ctx.Get("lazy_var")
	if ok {
		t.Error("Lazy variable should not be in regular variables")
	}

	// Should be retrievable as lazy (returns marker since actual value is AST)
	_, ok = ctx.GetLazy("lazy_var")
	if !ok {
		t.Error("Expected lazy variable to be defined")
	}

	// Check that IsLazy returns true
	if !ctx.IsLazy("lazy_var") {
		t.Error("Expected IsLazy to return true")
	}
}

func TestContext_IsLazy(t *testing.T) {
	ctx := NewContext()
	ctx.Set("regular", "value")
	ctx.SetLazy("lazy", "lazy value")

	tests := []struct {
		name   string
		expect bool
	}{
		{"regular", false},
		{"lazy", true},
		{"undefined", false},
		{"os", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ctx.IsLazy(tt.name) != tt.expect {
				t.Errorf("IsLazy(%s) = %v, want %v", tt.name, !tt.expect, tt.expect)
			}
		})
	}
}

func TestContext_Overwrite(t *testing.T) {
	ctx := NewContext()

	ctx.Set("var", "first")
	ctx.Set("var", "second")

	val, _ := ctx.Get("var")
	if val != "second" {
		t.Errorf("Expected 'second' after overwrite, got '%s'", val)
	}
}

func TestContext_BuiltinsAreReadOnly(t *testing.T) {
	ctx := NewContext()

	// Built-ins should not be overwritable
	originalOS, _ := ctx.Get("os")
	ctx.Set("os", "custom")
	newOS, _ := ctx.Get("os")

	if newOS != originalOS {
		t.Errorf("Built-in 'os' should not be overwritable, got '%s'", newOS)
	}
}

func TestContext_Variables(t *testing.T) {
	ctx := NewContext()
	ctx.Set("a", "1")
	ctx.Set("b", "2")
	ctx.Set("c", "3")

	vars := ctx.Variables()
	if len(vars) < 3 {
		t.Errorf("Expected at least 3 variables (a, b, c), got %d", len(vars))
	}

	// Check our variables are present
	found := make(map[string]bool)
	for k, v := range vars {
		if k == "a" && v == "1" {
			found["a"] = true
		}
		if k == "b" && v == "2" {
			found["b"] = true
		}
		if k == "c" && v == "3" {
			found["c"] = true
		}
	}

	for _, name := range []string{"a", "b", "c"} {
		if !found[name] {
			t.Errorf("Expected variable '%s' in Variables()", name)
		}
	}
}

func TestContext_LazyVariables(t *testing.T) {
	ctx := NewContext()
	ctx.SetLazy("x", "lazy x")
	ctx.SetLazy("y", "lazy y")

	lazyVars := ctx.LazyVariables()
	if len(lazyVars) != 2 {
		t.Errorf("Expected 2 lazy variables, got %d", len(lazyVars))
	}

	// LazyVariables now returns "__lazy__" marker since the actual values are AST nodes
	if _, ok := lazyVars["x"]; !ok {
		t.Error("Expected lazy variable 'x' to be present")
	}
	if _, ok := lazyVars["y"]; !ok {
		t.Error("Expected lazy variable 'y' to be present")
	}
}
