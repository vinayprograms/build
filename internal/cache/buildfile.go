package cache

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/vinayprograms/build/internal/ast"
)

// cacheEntry holds a cached parse result with its file metadata.
type cacheEntry struct {
	statements []ast.Statement
	modTime    time.Time
}

// BuildfileCache caches parsed Buildfile ASTs keyed by absolute path.
// Cache entries are invalidated when the file's mtime changes.
type BuildfileCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
}

// NewBuildfileCache creates an empty cache.
func NewBuildfileCache() *BuildfileCache {
	return &BuildfileCache{
		entries: make(map[string]*cacheEntry),
	}
}

// Put stores parsed statements for a file in the cache.
// The file's current modification time is recorded for invalidation.
// Returns an error if the file cannot be stat'd.
func (c *BuildfileCache) Put(path string, statements []ast.Statement) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[absPath] = &cacheEntry{
		statements: statements,
		modTime:    info.ModTime(),
	}

	return nil
}

// Get retrieves cached statements for a file.
// Returns (nil, false, nil) on cache miss.
// Returns (nil, false, nil) if file has been modified since caching.
// Returns (nil, false, nil) if file no longer exists.
// Returns (statements, true, nil) on cache hit.
func (c *BuildfileCache) Get(path string) ([]ast.Statement, bool, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, false, err
	}

	c.mu.RLock()
	entry, ok := c.entries[absPath]
	c.mu.RUnlock()

	if !ok {
		return nil, false, nil
	}

	// Check if file still exists and hasn't been modified
	info, err := os.Stat(absPath)
	if err != nil {
		// File no longer exists or can't be stat'd
		c.Invalidate(absPath)
		return nil, false, nil
	}

	// Compare modification times
	if !info.ModTime().Equal(entry.modTime) {
		c.Invalidate(absPath)
		return nil, false, nil
	}

	return entry.statements, true, nil
}

// Invalidate removes a specific entry from the cache.
func (c *BuildfileCache) Invalidate(path string) {
	absPath, _ := filepath.Abs(path)

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, absPath)
}

// Clear removes all entries from the cache.
func (c *BuildfileCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheEntry)
}

// Size returns the number of entries in the cache.
func (c *BuildfileCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}
