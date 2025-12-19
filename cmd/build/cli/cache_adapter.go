package cli

import (
	"os"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/cache"
	"github.com/vinayprograms/build/internal/parser"
)

// BuildfileCache interface abstracts cache operations.
type BuildfileCache interface {
	// Get retrieves cached statements for a file.
	// Returns (nil, false, nil) on cache miss.
	// Returns (statements, true, nil) on cache hit.
	Get(path string) ([]ast.Statement, bool, error)

	// Put stores parsed statements for a file.
	Put(path string, statements []ast.Statement) error

	// Invalidate removes a specific entry from the cache.
	Invalidate(path string)

	// Clear removes all entries from the cache.
	Clear()

	// Size returns the number of entries in the cache.
	Size() int
}

// buildfileCacheAdapter wraps *cache.BuildfileCache to implement BuildfileCache.
type buildfileCacheAdapter struct {
	c *cache.BuildfileCache
}

// NewBuildfileCache creates a new cache instance.
func NewBuildfileCacheImpl() BuildfileCache {
	return &buildfileCacheAdapter{c: cache.NewBuildfileCache()}
}

func (a *buildfileCacheAdapter) Get(path string) ([]ast.Statement, bool, error) {
	return a.c.Get(path)
}

func (a *buildfileCacheAdapter) Put(path string, statements []ast.Statement) error {
	return a.c.Put(path, statements)
}

func (a *buildfileCacheAdapter) Invalidate(path string) {
	a.c.Invalidate(path)
}

func (a *buildfileCacheAdapter) Clear() {
	a.c.Clear()
}

func (a *buildfileCacheAdapter) Size() int {
	return a.c.Size()
}

// globalBuildfileCache is the singleton cache instance used by the CLI.
var globalBuildfileCache BuildfileCache

// initBuildfileCache initializes the global cache.
func initBuildfileCache() {
	globalBuildfileCache = NewBuildfileCacheImpl()
}

// GetBuildfileCache returns the global cache instance.
func GetBuildfileCache() BuildfileCache {
	if globalBuildfileCache == nil {
		initBuildfileCache()
	}
	return globalBuildfileCache
}

// ResetBuildfileCache clears the global cache.
// This is primarily useful for testing.
func ResetBuildfileCache() {
	if globalBuildfileCache != nil {
		globalBuildfileCache.Clear()
	}
}

// globalAutodepsCache is the singleton autodeps cache instance.
var globalAutodepsCache *cache.AutodepsCache

// GetAutodepsCache returns the global autodeps cache instance.
func GetAutodepsCache() *cache.AutodepsCache {
	if globalAutodepsCache == nil {
		globalAutodepsCache = cache.NewAutodepsCache()
	}
	return globalAutodepsCache
}

// ResetAutodepsCache clears the global autodeps cache.
// This is primarily useful for testing.
func ResetAutodepsCache() {
	if globalAutodepsCache != nil {
		globalAutodepsCache.Clear()
	}
}

// cachedBuildfileResultAdapter wraps cached AST statements as a BuildfileResult.
type cachedBuildfileResultAdapter struct {
	statements []ast.Statement
}

func (cr cachedBuildfileResultAdapter) Statements() []Statement {
	result := make([]Statement, len(cr.statements))
	for i, s := range cr.statements {
		result[i] = statementAdapter{s: s}
	}
	return result
}

func (cr cachedBuildfileResultAdapter) ErrorCount() int { return 0 }

func (cr cachedBuildfileResultAdapter) GetError(_ int) ParseError { return nil }

func (cr cachedBuildfileResultAdapter) HasErrors() bool { return false }

func (cr cachedBuildfileResultAdapter) AllErrors() string { return "" }

// ParseBuildfileWithCache parses a Buildfile with caching support.
// If the file is cached and unchanged, returns the cached result.
// Otherwise, parses the file and caches the result.
// Returns the BuildfileResult, content (for error formatting), and any read error.
func ParseBuildfileWithCache(buildfile string) (BuildfileResult, string, error) {
	cache := GetBuildfileCache()

	// Try cache first
	if stmts, ok, err := cache.Get(buildfile); err == nil && ok {
		// Cache hit - still need to read content for potential error formatting later
		content, err := os.ReadFile(buildfile)
		if err != nil {
			return nil, "", err
		}
		return cachedBuildfileResultAdapter{statements: stmts}, string(content), nil
	}

	// Cache miss - parse the file
	content, err := os.ReadFile(buildfile)
	if err != nil {
		return nil, "", err
	}

	l := NewLexer(buildfile, string(content))
	p := NewParser(l)
	bp := NewBuildfileParser(p)
	result := bp.ParseBuildfile()

	// Only cache if no errors
	if !result.HasErrors() {
		astStmts := GetASTStatements(result)
		if astStmts != nil {
			_ = cache.Put(buildfile, astStmts)
		}
	}

	return result, string(content), nil
}

// GetASTStatementsFromCachedResult extracts AST statements from a cached result.
func GetASTStatementsFromCachedResult(result BuildfileResult) []ast.Statement {
	// First try the standard adapter
	if stmts := GetASTStatements(result); stmts != nil {
		return stmts
	}
	// Then try the cached adapter
	if cr, ok := result.(cachedBuildfileResultAdapter); ok {
		return cr.statements
	}
	return nil
}

// emptyParseErrors returns an empty ParseErrors for cached results.
func emptyParseErrors() *parser.ParseErrors {
	return &parser.ParseErrors{}
}
