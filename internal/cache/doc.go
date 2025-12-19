// Package cache provides caching for parsed Buildfiles.
//
// The cache stores parsed AST keyed by absolute file path. Cache entries are
// invalidated when the file's modification time changes. This avoids re-parsing
// unchanged Buildfiles on subsequent builds.
//
// Cache invalidation strategy:
//   - Store file mtime with each entry
//   - On lookup, compare current mtime with stored mtime
//   - If different, invalidate and return cache miss
//
// The cache is not persistent across process invocations. A future version may
// add disk-based persistence.
package cache
