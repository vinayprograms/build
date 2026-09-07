package cli

import (
	"os"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/cache"
	"github.com/vinayprograms/need/internal/parser"
)

// NeedfileCache interface abstracts cache operations.
type NeedfileCache interface {
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

// needfileCacheAdapter wraps *cache.NeedfileCache to implement NeedfileCache.
type needfileCacheAdapter struct {
	c *cache.NeedfileCache
}

// NewNeedfileCache creates a new cache instance.
func NewNeedfileCacheImpl() NeedfileCache {
	return &needfileCacheAdapter{c: cache.NewNeedfileCache()}
}

func (a *needfileCacheAdapter) Get(path string) ([]ast.Statement, bool, error) {
	return a.c.Get(path)
}

func (a *needfileCacheAdapter) Put(path string, statements []ast.Statement) error {
	return a.c.Put(path, statements)
}

func (a *needfileCacheAdapter) Invalidate(path string) {
	a.c.Invalidate(path)
}

func (a *needfileCacheAdapter) Clear() {
	a.c.Clear()
}

func (a *needfileCacheAdapter) Size() int {
	return a.c.Size()
}

// globalNeedfileCache is the singleton cache instance used by the CLI.
var globalNeedfileCache NeedfileCache

// initNeedfileCache initializes the global cache.
func initNeedfileCache() {
	globalNeedfileCache = NewNeedfileCacheImpl()
}

// GetNeedfileCache returns the global cache instance.
func GetNeedfileCache() NeedfileCache {
	if globalNeedfileCache == nil {
		initNeedfileCache()
	}
	return globalNeedfileCache
}

// ResetNeedfileCache clears the global cache.
// This is primarily useful for testing.
func ResetNeedfileCache() {
	if globalNeedfileCache != nil {
		globalNeedfileCache.Clear()
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

// cachedNeedfileResultAdapter wraps cached AST statements as a NeedfileResult.
type cachedNeedfileResultAdapter struct {
	statements []ast.Statement
}

func (cr cachedNeedfileResultAdapter) Statements() []Statement {
	result := make([]Statement, len(cr.statements))
	for i, s := range cr.statements {
		result[i] = statementAdapter{s: s}
	}
	return result
}

func (cr cachedNeedfileResultAdapter) ErrorCount() int { return 0 }

func (cr cachedNeedfileResultAdapter) GetError(_ int) ParseError { return nil }

func (cr cachedNeedfileResultAdapter) HasErrors() bool { return false }

func (cr cachedNeedfileResultAdapter) AllErrors() string { return "" }

// ParseNeedfileWithCache parses a Needfile with caching support.
// If the file is cached and unchanged, returns the cached result.
// Otherwise, parses the file and caches the result.
// Returns the NeedfileResult, content (for error formatting), and any read error.
func ParseNeedfileWithCache(needfile string) (NeedfileResult, string, error) {
	cache := GetNeedfileCache()

	// Try cache first
	if stmts, ok, err := cache.Get(needfile); err == nil && ok {
		// Cache hit - still need to read content for potential error formatting later
		content, err := os.ReadFile(needfile)
		if err != nil {
			return nil, "", err
		}
		return cachedNeedfileResultAdapter{statements: stmts}, string(content), nil
	}

	// Cache miss - parse the file
	content, err := os.ReadFile(needfile)
	if err != nil {
		return nil, "", err
	}

	l := NewLexer(needfile, string(content))
	p := NewParser(l)
	bp := NewNeedfileParser(p)
	result := bp.ParseNeedfile()

	// Only cache if no errors
	if !result.HasErrors() {
		astStmts := GetASTStatements(result)
		if astStmts != nil {
			_ = cache.Put(needfile, astStmts)
		}
	}

	return result, string(content), nil
}

// GetASTStatementsFromCachedResult extracts AST statements from a cached result.
func GetASTStatementsFromCachedResult(result NeedfileResult) []ast.Statement {
	// First try the standard adapter
	if stmts := GetASTStatements(result); stmts != nil {
		return stmts
	}
	// Then try the cached adapter
	if cr, ok := result.(cachedNeedfileResultAdapter); ok {
		return cr.statements
	}
	return nil
}

// emptyParseErrors returns an empty ParseErrors for cached results.
func emptyParseErrors() *parser.ParseErrors {
	return &parser.ParseErrors{}
}
