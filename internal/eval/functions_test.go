package eval

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/vinayprograms/need/internal/ast"
)

// ----------------------------------------------------------------------------
// basename() Tests
// ----------------------------------------------------------------------------

func TestFuncFilename_Simple(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncFilename,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "/path/to/file.txt"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "file.txt" {
		t.Errorf("Expected 'file.txt', got '%s'", result)
	}
}

func TestFuncFilename_NoDirectory(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncFilename,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "file.txt"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "file.txt" {
		t.Errorf("Expected 'file.txt', got '%s'", result)
	}
}

func TestFuncFilename_WithInterpolation(t *testing.T) {
	ctx := NewContext()
	ctx.Set("path", "/usr/local/bin/app")
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncFilename,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.Interpolation{Name: "path"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "app" {
		t.Errorf("Expected 'app', got '%s'", result)
	}
}

// ----------------------------------------------------------------------------
// dirname() Tests
// ----------------------------------------------------------------------------

func TestFuncDirname_Simple(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncDirname,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "/path/to/file.txt"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "/path/to" {
		t.Errorf("Expected '/path/to', got '%s'", result)
	}
}

func TestFuncDirname_NoDirectory(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncDirname,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "file.txt"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "." {
		t.Errorf("Expected '.', got '%s'", result)
	}
}

func TestFuncDirname_RootPath(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncDirname,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "/file.txt"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "/" {
		t.Errorf("Expected '/', got '%s'", result)
	}
}

// ----------------------------------------------------------------------------
// replace() Tests
// ----------------------------------------------------------------------------

func TestFuncReplace_Simple(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncReplace,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "hello world"}}},
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "world"}}},
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "universe"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "hello universe" {
		t.Errorf("Expected 'hello universe', got '%s'", result)
	}
}

func TestFuncReplace_Multiple(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncReplace,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "a.c b.c c.c"}}},
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: ".c"}}},
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: ".o"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "a.o b.o c.o" {
		t.Errorf("Expected 'a.o b.o c.o', got '%s'", result)
	}
}

func TestFuncReplace_WithInterpolation(t *testing.T) {
	ctx := NewContext()
	ctx.Set("sources", "main.c utils.c")
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncReplace,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.Interpolation{Name: "sources"}}},
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: ".c"}}},
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: ".o"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "main.o utils.o" {
		t.Errorf("Expected 'main.o utils.o', got '%s'", result)
	}
}

func TestFuncReplace_NoMatch(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncReplace,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "hello world"}}},
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "xyz"}}},
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "abc"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", result)
	}
}

// ----------------------------------------------------------------------------
// glob() Tests
// ----------------------------------------------------------------------------

func TestFuncGlob_CurrentDir(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		f, err := os.Create(filepath.Join(tmpDir, name))
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		f.Close()
	}

	// Change to temp dir
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncGlob,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "*.txt"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that all files are included
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if !strings.Contains(result, name) {
			t.Errorf("Expected result to contain '%s', got '%s'", name, result)
		}
	}
}

func TestFuncGlob_NoMatches(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncGlob,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "*.nonexistent"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty string for no matches, got '%s'", result)
	}
}

// ----------------------------------------------------------------------------
// shell() Tests
// ----------------------------------------------------------------------------

func TestFuncShell_Echo(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "echo hello"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("Expected 'hello', got '%s'", result)
	}
}

func TestFuncShell_TrimNewline(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	// Use printf which works consistently across platforms
	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "printf hello"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("Expected 'hello', got '%s'", result)
	}
}

func TestFuncShell_WithInterpolation(t *testing.T) {
	ctx := NewContext()
	ctx.Set("greeting", "world")
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{
						&ast.LiteralValue{Text: "echo hello "},
						&ast.Interpolation{Name: "greeting"},
					}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", result)
	}
}

func TestFuncShell_FailingCommand(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "exit 1"}}},
				},
			},
		},
	}

	_, err := e.EvaluateValue(val)
	if err == nil {
		t.Error("Expected error for failing command")
	}
}

// ----------------------------------------------------------------------------
// Function Composition Tests
// ----------------------------------------------------------------------------

func TestFuncComposition_DirnameBasename(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	// dirname(basename(path)) - should get directory of just the filename
	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncDirname,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{
						&ast.FunctionCall{
							Name: ast.FuncFilename,
							Args: []*ast.Value{
								{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "/path/to/file.txt"}}},
							},
						},
					}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// basename returns "file.txt", dirname of that is "."
	if result != "." {
		t.Errorf("Expected '.', got '%s'", result)
	}
}

func TestFuncComposition_ReplaceInGlob(t *testing.T) {
	// This test verifies replace can process glob output
	ctx := NewContext()
	ctx.Set("files", "a.c b.c c.c")
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncReplace,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.Interpolation{Name: "files"}}},
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: ".c"}}},
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: ".o"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "a.o b.o c.o" {
		t.Errorf("Expected 'a.o b.o c.o', got '%s'", result)
	}
}

// ----------------------------------------------------------------------------
// Error Tests
// ----------------------------------------------------------------------------

func TestFuncError_UndefinedInArg(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncFilename,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{
						&ast.Interpolation{
							Name:     "undefined",
							Location: ast.SourceLocation{File: "test", Line: 1, Column: 1},
						},
					}},
				},
			},
		},
	}

	_, err := e.EvaluateValue(val)
	if err == nil {
		t.Error("Expected error for undefined variable in function argument")
	}
}

// ----------------------------------------------------------------------------
// Shell Quoting Tests for Interpolations in shell()
// ----------------------------------------------------------------------------

func TestFuncShell_QuotedInterpolation(t *testing.T) {
	// Per spec: {var} in shell() should be shell-quoted (safe for paths with spaces)
	// shell(echo {src_dir}) with src_dir="my sources" should execute: echo 'my sources'
	ctx := NewContext()
	ctx.Set("src_dir", "my sources")
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{
						&ast.LiteralValue{Text: "echo "},
						&ast.Interpolation{Name: "src_dir", Raw: false},
					}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// With quoting, echo 'my sources' outputs: my sources
	if result != "my sources" {
		t.Errorf("Expected 'my sources', got '%s'", result)
	}
}

func TestFuncShell_RawInterpolation(t *testing.T) {
	// Per spec: {var:raw} in shell() should NOT be quoted (allows word splitting)
	// shell(echo {flags:raw}) with flags="-Wall -O2" should execute: echo -Wall -O2
	ctx := NewContext()
	ctx.Set("flags", "-Wall -O2")
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{
						&ast.LiteralValue{Text: "echo "},
						&ast.Interpolation{Name: "flags", Raw: true},
					}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Without quoting, echo -Wall -O2 outputs: -Wall -O2
	if result != "-Wall -O2" {
		t.Errorf("Expected '-Wall -O2', got '%s'", result)
	}
}

func TestFuncShell_QuotesHandleEmbeddedQuotes(t *testing.T) {
	// Test that embedded single quotes are handled correctly
	ctx := NewContext()
	ctx.Set("msg", "it's working")
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{
						&ast.LiteralValue{Text: "echo "},
						&ast.Interpolation{Name: "msg", Raw: false},
					}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "it's working" {
		t.Errorf("Expected \"it's working\", got '%s'", result)
	}
}

func TestFuncShell_QuotedPreservesSpaces(t *testing.T) {
	// Verify that quoted interpolation preserves the value as a single word
	ctx := NewContext()
	ctx.Set("dir", "path with spaces")
	e := NewEvaluator(ctx)

	// Use printf '%s' which echoes its argument exactly
	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{
						&ast.LiteralValue{Text: "printf '%s' "},
						&ast.Interpolation{Name: "dir", Raw: false},
					}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Should output the value as single string: "path with spaces"
	if result != "path with spaces" {
		t.Errorf("Expected 'path with spaces', got '%s'", result)
	}
}

// Skip this test on Windows where shell behavior differs
func TestFuncShell_MultilineOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "printf 'line1\\nline2\\nline3'"}}},
				},
			},
		},
	}

	result, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should preserve internal newlines but trim trailing
	expected := "line1\nline2\nline3"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// ----------------------------------------------------------------------------
// Shell Caching Tests
// ----------------------------------------------------------------------------

func TestFuncShell_CachingBasic(t *testing.T) {
	// Verify that identical shell commands are cached
	ctx := NewContext()
	e := NewEvaluator(ctx)

	// Create shell call that generates a unique timestamp
	// Running twice should return same result if cached
	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "echo cached"}}},
				},
			},
		},
	}

	result1, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	result2, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result1 != result2 {
		t.Errorf("Expected cached result to match: '%s' != '%s'", result1, result2)
	}
	if result1 != "cached" {
		t.Errorf("Expected 'cached', got '%s'", result1)
	}
}

func TestFuncShell_CachingWithSameCommand(t *testing.T) {
	// Verify that the same command called multiple times returns cached result
	ctx := NewContext()
	e := NewEvaluator(ctx)

	// Use a command that would produce different output each time if not cached
	// We'll use a file counter approach
	tmpDir := t.TempDir()
	counterFile := filepath.Join(tmpDir, "counter")
	os.WriteFile(counterFile, []byte("0"), 0644)

	// Command increments counter and returns new value
	// But with caching, second call should return same value
	cmd := "count=$(cat " + counterFile + "); count=$((count + 1)); echo $count > " + counterFile + "; echo $count"

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: cmd}}},
				},
			},
		},
	}

	result1, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	result2, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// With caching, both results should be "1" (command only executed once)
	if result1 != "1" {
		t.Errorf("Expected first result to be '1', got '%s'", result1)
	}
	if result1 != result2 {
		t.Errorf("Expected cached result, got different values: '%s' vs '%s'", result1, result2)
	}
}

func TestFuncShell_CachingDifferentCommands(t *testing.T) {
	// Different commands should not share cache
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val1 := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "echo first"}}},
				},
			},
		},
	}

	val2 := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "echo second"}}},
				},
			},
		},
	}

	result1, err := e.EvaluateValue(val1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	result2, err := e.EvaluateValue(val2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result1 != "first" {
		t.Errorf("Expected 'first', got '%s'", result1)
	}
	if result2 != "second" {
		t.Errorf("Expected 'second', got '%s'", result2)
	}
}

func TestFuncShell_CachingWithInterpolation(t *testing.T) {
	// Verify cache key includes interpolated values
	ctx := NewContext()
	ctx.Set("name", "world")
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{
						&ast.LiteralValue{Text: "echo hello "},
						&ast.Interpolation{Name: "name"},
					}},
				},
			},
		},
	}

	result1, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Change the variable value
	ctx.Set("name", "universe")

	result2, err := e.EvaluateValue(val)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Different interpolated values should produce different results
	if result1 != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", result1)
	}
	if result2 != "hello universe" {
		t.Errorf("Expected 'hello universe', got '%s'", result2)
	}
}

func TestFuncShell_CachingErrorsNotCached(t *testing.T) {
	// Failing commands should not be cached
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "exit 1"}}},
				},
			},
		},
	}

	_, err1 := e.EvaluateValue(val)
	if err1 == nil {
		t.Fatal("Expected error for failing command")
	}

	// Second call should also fail (not cached)
	_, err2 := e.EvaluateValue(val)
	if err2 == nil {
		t.Fatal("Expected error for second call to failing command")
	}
}

func TestContext_ShellCacheOperations(t *testing.T) {
	ctx := NewContext()

	// Test GetShellCache returns false for missing key
	_, ok := ctx.GetShellCache("echo test")
	if ok {
		t.Error("Expected GetShellCache to return false for missing key")
	}

	// Test SetShellCache stores value
	ctx.SetShellCache("echo test", "test output")

	val, ok := ctx.GetShellCache("echo test")
	if !ok {
		t.Error("Expected GetShellCache to return true after SetShellCache")
	}
	if val != "test output" {
		t.Errorf("Expected 'test output', got '%s'", val)
	}

	// Test ClearShellCache clears cache
	ctx.ClearShellCache()

	_, ok = ctx.GetShellCache("echo test")
	if ok {
		t.Error("Expected GetShellCache to return false after ClearShellCache")
	}
}
