package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vinayprograms/build/internal/ast"
)

func TestNewBuildfileCache(t *testing.T) {
	c := NewBuildfileCache()
	if c == nil {
		t.Fatal("NewBuildfileCache returned nil")
	}
}

func TestBuildfileCache_PutAndGet(t *testing.T) {
	c := NewBuildfileCache()

	// Create a temp file
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	if err := os.WriteFile(buildfile, []byte("cc = gcc\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Create a simple AST
	stmts := []ast.Statement{
		&ast.Variable{
			Name:     "cc",
			Value:    &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "gcc"}}},
			Location: ast.SourceLocation{File: buildfile, Line: 1, Column: 1},
		},
	}

	// Put into cache
	err := c.Put(buildfile, stmts)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Get from cache
	cached, ok, err := c.Get(buildfile)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit, got miss")
	}
	if len(cached) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(cached))
	}

	v, ok := cached[0].(*ast.Variable)
	if !ok {
		t.Fatal("expected *ast.Variable")
	}
	if v.Name != "cc" {
		t.Errorf("expected variable name 'cc', got %q", v.Name)
	}
}

func TestBuildfileCache_InvalidateOnModification(t *testing.T) {
	c := NewBuildfileCache()

	// Create a temp file
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	if err := os.WriteFile(buildfile, []byte("cc = gcc\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Create initial AST
	stmts := []ast.Statement{
		&ast.Variable{
			Name:     "cc",
			Value:    &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "gcc"}}},
			Location: ast.SourceLocation{File: buildfile, Line: 1, Column: 1},
		},
	}

	// Put into cache
	if err := c.Put(buildfile, stmts); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Verify cache hit
	_, ok, err := c.Get(buildfile)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit before modification")
	}

	// Wait briefly then modify the file
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(buildfile, []byte("cc = clang\n"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// Verify cache miss after modification
	_, ok, err = c.Get(buildfile)
	if err != nil {
		t.Fatalf("Get failed after modification: %v", err)
	}
	if ok {
		t.Fatal("expected cache miss after modification, got hit")
	}
}

func TestBuildfileCache_NonexistentFile(t *testing.T) {
	c := NewBuildfileCache()

	// Try to put a nonexistent file
	err := c.Put("/nonexistent/Buildfile", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestBuildfileCache_DeletedFile(t *testing.T) {
	c := NewBuildfileCache()

	// Create a temp file
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	if err := os.WriteFile(buildfile, []byte("cc = gcc\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Put into cache
	if err := c.Put(buildfile, nil); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Delete the file
	if err := os.Remove(buildfile); err != nil {
		t.Fatalf("failed to delete file: %v", err)
	}

	// Get should return miss (file no longer exists)
	_, ok, err := c.Get(buildfile)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if ok {
		t.Fatal("expected cache miss for deleted file")
	}
}

func TestBuildfileCache_Clear(t *testing.T) {
	c := NewBuildfileCache()

	// Create a temp file
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	if err := os.WriteFile(buildfile, []byte("cc = gcc\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Put into cache
	if err := c.Put(buildfile, nil); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Clear the cache
	c.Clear()

	// Verify cache miss
	_, ok, err := c.Get(buildfile)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if ok {
		t.Fatal("expected cache miss after Clear()")
	}
}

func TestBuildfileCache_Invalidate(t *testing.T) {
	c := NewBuildfileCache()

	// Create a temp file
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	if err := os.WriteFile(buildfile, []byte("cc = gcc\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Put into cache
	if err := c.Put(buildfile, nil); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Invalidate specific entry
	c.Invalidate(buildfile)

	// Verify cache miss
	_, ok, err := c.Get(buildfile)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if ok {
		t.Fatal("expected cache miss after Invalidate()")
	}
}

func TestBuildfileCache_MultipleFiles(t *testing.T) {
	c := NewBuildfileCache()
	tmpDir := t.TempDir()

	// Create multiple files
	files := []string{"Buildfile", "common.build", "deps.build"}
	stmtsMap := make(map[string][]ast.Statement)

	for i, name := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("# file "+name), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}

		stmts := []ast.Statement{
			&ast.Comment{Text: "file " + name, Location: ast.SourceLocation{File: path, Line: 1}},
		}
		stmtsMap[path] = stmts

		if err := c.Put(path, stmts); err != nil {
			t.Fatalf("Put failed for %s: %v", name, err)
		}

		_ = i // unused
	}

	// Verify all files are cached
	for _, name := range files {
		path := filepath.Join(tmpDir, name)
		cached, ok, err := c.Get(path)
		if err != nil {
			t.Fatalf("Get failed for %s: %v", name, err)
		}
		if !ok {
			t.Fatalf("expected cache hit for %s", name)
		}
		if len(cached) != 1 {
			t.Fatalf("expected 1 statement for %s, got %d", name, len(cached))
		}
	}
}

func TestBuildfileCache_Size(t *testing.T) {
	c := NewBuildfileCache()

	if c.Size() != 0 {
		t.Fatalf("expected size 0, got %d", c.Size())
	}

	tmpDir := t.TempDir()

	// Add entries
	for i := 0; i < 3; i++ {
		path := filepath.Join(tmpDir, "file"+string(rune('0'+i))+".build")
		if err := os.WriteFile(path, []byte("# test"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		if err := c.Put(path, nil); err != nil {
			t.Fatalf("Put failed: %v", err)
		}
	}

	if c.Size() != 3 {
		t.Fatalf("expected size 3, got %d", c.Size())
	}

	// Invalidate one
	c.Invalidate(filepath.Join(tmpDir, "file0.build"))

	if c.Size() != 2 {
		t.Fatalf("expected size 2 after invalidate, got %d", c.Size())
	}

	// Clear all
	c.Clear()

	if c.Size() != 0 {
		t.Fatalf("expected size 0 after clear, got %d", c.Size())
	}
}

func TestBuildfileCache_RelativePath(t *testing.T) {
	c := NewBuildfileCache()

	// Create a temp file
	tmpDir := t.TempDir()
	buildfile := filepath.Join(tmpDir, "Buildfile")
	if err := os.WriteFile(buildfile, []byte("cc = gcc\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Change to tmpDir and use relative path
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Put with relative path (should work and normalize to absolute)
	if err := c.Put("Buildfile", nil); err != nil {
		t.Fatalf("Put with relative path failed: %v", err)
	}

	// Get with relative path (same as put)
	_, ok, err := c.Get("Buildfile")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit with relative path")
	}

	// Put with absolute path, get with relative should also work
	c2 := NewBuildfileCache()

	absPath, _ := filepath.Abs("Buildfile")
	if err := c2.Put(absPath, nil); err != nil {
		t.Fatalf("Put with absolute path failed: %v", err)
	}

	_, ok, err = c2.Get("Buildfile")
	if err != nil {
		t.Fatalf("Get with relative path failed: %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit with relative path after absolute put")
	}
}
