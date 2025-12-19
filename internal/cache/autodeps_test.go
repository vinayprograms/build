package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewAutodepsCache(t *testing.T) {
	c := NewAutodepsCache()
	if c == nil {
		t.Fatal("NewAutodepsCache returned nil")
	}
}

func TestAutodepsCache_PutAndGet(t *testing.T) {
	c := NewAutodepsCache()

	// Create a temp .d file
	tmpDir := t.TempDir()
	depfile := filepath.Join(tmpDir, "main.d")
	content := "main.o: main.c main.h utils.h\n"
	if err := os.WriteFile(depfile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	deps := []string{"main.c", "main.h", "utils.h"}

	// Put into cache
	err := c.Put(depfile, deps)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Get from cache
	cached, ok, err := c.Get(depfile)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit, got miss")
	}
	if len(cached) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(cached))
	}
	if cached[0] != "main.c" || cached[1] != "main.h" || cached[2] != "utils.h" {
		t.Errorf("unexpected dependencies: %v", cached)
	}
}

func TestAutodepsCache_InvalidateOnModification(t *testing.T) {
	c := NewAutodepsCache()

	// Create a temp .d file
	tmpDir := t.TempDir()
	depfile := filepath.Join(tmpDir, "main.d")
	if err := os.WriteFile(depfile, []byte("main.o: main.c\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	deps := []string{"main.c"}

	// Put into cache
	if err := c.Put(depfile, deps); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Verify cache hit
	_, ok, err := c.Get(depfile)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit before modification")
	}

	// Wait briefly then modify the file
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(depfile, []byte("main.o: main.c utils.h\n"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// Verify cache miss after modification
	_, ok, err = c.Get(depfile)
	if err != nil {
		t.Fatalf("Get failed after modification: %v", err)
	}
	if ok {
		t.Fatal("expected cache miss after modification, got hit")
	}
}

func TestAutodepsCache_NonexistentFile(t *testing.T) {
	c := NewAutodepsCache()

	// Try to put a nonexistent file
	err := c.Put("/nonexistent/main.d", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestAutodepsCache_DeletedFile(t *testing.T) {
	c := NewAutodepsCache()

	// Create a temp .d file
	tmpDir := t.TempDir()
	depfile := filepath.Join(tmpDir, "main.d")
	if err := os.WriteFile(depfile, []byte("main.o: main.c\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Put into cache
	if err := c.Put(depfile, nil); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Delete the file
	if err := os.Remove(depfile); err != nil {
		t.Fatalf("failed to delete file: %v", err)
	}

	// Get should return miss (file no longer exists)
	_, ok, err := c.Get(depfile)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if ok {
		t.Fatal("expected cache miss for deleted file")
	}
}

func TestAutodepsCache_Clear(t *testing.T) {
	c := NewAutodepsCache()

	// Create a temp .d file
	tmpDir := t.TempDir()
	depfile := filepath.Join(tmpDir, "main.d")
	if err := os.WriteFile(depfile, []byte("main.o: main.c\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Put into cache
	if err := c.Put(depfile, nil); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Clear the cache
	c.Clear()

	// Verify cache miss
	_, ok, err := c.Get(depfile)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if ok {
		t.Fatal("expected cache miss after Clear()")
	}
}

func TestAutodepsCache_Invalidate(t *testing.T) {
	c := NewAutodepsCache()

	// Create a temp .d file
	tmpDir := t.TempDir()
	depfile := filepath.Join(tmpDir, "main.d")
	if err := os.WriteFile(depfile, []byte("main.o: main.c\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Put into cache
	if err := c.Put(depfile, nil); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Invalidate specific entry
	c.Invalidate(depfile)

	// Verify cache miss
	_, ok, err := c.Get(depfile)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if ok {
		t.Fatal("expected cache miss after Invalidate()")
	}
}

func TestAutodepsCache_MultipleFiles(t *testing.T) {
	c := NewAutodepsCache()
	tmpDir := t.TempDir()

	// Create multiple .d files
	files := []string{"main.d", "utils.d", "parser.d"}
	depsMap := make(map[string][]string)

	for _, name := range files {
		path := filepath.Join(tmpDir, name)
		content := "obj/" + name[:len(name)-2] + ".o: src/" + name[:len(name)-2] + ".c\n"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}

		deps := []string{"src/" + name[:len(name)-2] + ".c"}
		depsMap[path] = deps

		if err := c.Put(path, deps); err != nil {
			t.Fatalf("Put failed for %s: %v", name, err)
		}
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
			t.Fatalf("expected 1 dependency for %s, got %d", name, len(cached))
		}
	}
}

func TestAutodepsCache_Size(t *testing.T) {
	c := NewAutodepsCache()

	if c.Size() != 0 {
		t.Fatalf("expected size 0, got %d", c.Size())
	}

	tmpDir := t.TempDir()

	// Add entries
	for i := 0; i < 3; i++ {
		path := filepath.Join(tmpDir, "file"+string(rune('0'+i))+".d")
		if err := os.WriteFile(path, []byte("target: dep\n"), 0644); err != nil {
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
	c.Invalidate(filepath.Join(tmpDir, "file0.d"))

	if c.Size() != 2 {
		t.Fatalf("expected size 2 after invalidate, got %d", c.Size())
	}

	// Clear all
	c.Clear()

	if c.Size() != 0 {
		t.Fatalf("expected size 0 after clear, got %d", c.Size())
	}
}

func TestAutodepsCache_EmptyDeps(t *testing.T) {
	c := NewAutodepsCache()

	tmpDir := t.TempDir()
	depfile := filepath.Join(tmpDir, "empty.d")
	if err := os.WriteFile(depfile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Put empty deps
	if err := c.Put(depfile, []string{}); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Get should return empty slice
	cached, ok, err := c.Get(depfile)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(cached) != 0 {
		t.Fatalf("expected empty slice, got %v", cached)
	}
}
