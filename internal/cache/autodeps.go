package cache

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

// autodepsEntry holds cached autodeps with file metadata.
type autodepsEntry struct {
	deps    []string
	modTime time.Time
}

// AutodepsCache caches parsed .d file contents keyed by absolute path.
// Cache entries are invalidated when the file's mtime changes.
type AutodepsCache struct {
	mu      sync.RWMutex
	entries map[string]*autodepsEntry
}

// NewAutodepsCache creates an empty autodeps cache.
func NewAutodepsCache() *AutodepsCache {
	return &AutodepsCache{
		entries: make(map[string]*autodepsEntry),
	}
}

// Put stores parsed dependencies for a file in the cache.
// The file's current modification time is recorded for invalidation.
// Returns an error if the file cannot be stat'd.
func (c *AutodepsCache) Put(path string, deps []string) error {
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

	c.entries[absPath] = &autodepsEntry{
		deps:    deps,
		modTime: info.ModTime(),
	}

	return nil
}

// Get retrieves cached dependencies for a file.
// Returns (nil, false, nil) on cache miss.
// Returns (nil, false, nil) if file has been modified since caching.
// Returns (nil, false, nil) if file no longer exists.
// Returns (deps, true, nil) on cache hit.
func (c *AutodepsCache) Get(path string) ([]string, bool, error) {
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

	return entry.deps, true, nil
}

// Invalidate removes a specific entry from the cache.
func (c *AutodepsCache) Invalidate(path string) {
	absPath, _ := filepath.Abs(path)

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, absPath)
}

// Clear removes all entries from the cache.
func (c *AutodepsCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*autodepsEntry)
}

// Size returns the number of entries in the cache.
func (c *AutodepsCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}
