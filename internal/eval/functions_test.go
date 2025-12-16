package eval

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

// ----------------------------------------------------------------------------
// basename() Tests
// ----------------------------------------------------------------------------

func TestFuncBasename_Simple(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncBasename,
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

func TestFuncBasename_NoDirectory(t *testing.T) {
	ctx := NewContext()
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncBasename,
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

func TestFuncBasename_WithInterpolation(t *testing.T) {
	ctx := NewContext()
	ctx.Set("path", "/usr/local/bin/app")
	e := NewEvaluator(ctx)

	val := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncBasename,
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
							Name: ast.FuncBasename,
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
				Name: ast.FuncBasename,
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
