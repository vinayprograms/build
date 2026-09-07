// Package semantic provides semantic analysis for Needfiles.
//
// Semantic analysis validates the AST produced by the parser and resolves
// context-sensitive constructs. It runs in four passes, each building on
// the results of previous passes.
//
// # Pass 1: Symbol Collection (collector.go)
//
// The Collect function walks the AST and populates a symbol table with all
// definitions:
//   - Variables (immediate and lazy)
//   - Targets (file, phony, directory, pattern)
//   - Environments (named and default)
//   - Conditional variables (tracked separately for runtime evaluation)
//
// Duplicate definitions are detected and reported as errors.
//
// # Pass 2: Capture Validation (capture.go)
//
// The ValidateCaptures function resolves BraceExpr nodes in target patterns.
// For each {name} in a pattern:
//   - If name is a defined variable → Interpolation
//   - If name is an automatic variable → Error (runtime-only)
//   - Otherwise → Capture (pattern matching variable)
//
// Captures in dependencies must be defined in the target pattern; a dependency
// cannot introduce a new capture.
//
// # Pass 3: Reference Validation (reference.go)
//
// The ValidateReferences function verifies that all variable references point
// to defined symbols:
//   - Automatic variables (target, deps, in, out, stem, etc.) are only valid
//     inside recipe or block scope
//   - Capture variables are only valid inside their defining recipe
//   - Built-in variables (os, arch) are always valid
//
// # Pass 4: Dependency Graph (depgraph.go)
//
// The ValidateDependencies function builds the dependency graph and detects
// circular dependencies:
//   - Pattern targets are tracked separately (they define rules, not nodes)
//   - Unsatisfied dependencies (deps not defined as targets) are recorded
//     but not errors—they may be source files or satisfiable by patterns
//   - Circular dependencies are detected using DFS-based cycle detection
//
// # Symbol Table (symbols.go)
//
// The SymbolTable tracks all defined symbols with O(1) lookup by name for
// variables and environments. Targets are stored in definition order for
// match priority. The table distinguishes:
//   - User-defined variables
//   - Automatic variables (target, deps, in, out, stem, target.dir, target.file)
//   - Built-in variables (os, arch)
//
// # Usage
//
// Typical usage runs all four passes in sequence:
//
//	// Parse
//	stmts, errs := parser.ParseNeedfile()
//
//	// Pass 1
//	collectResult := semantic.Collect(stmts)
//
//	// Pass 2
//	captureResult := semantic.ValidateCaptures(collectResult.SymbolTable)
//
//	// Pass 3
//	refResult := semantic.ValidateReferences(collectResult.SymbolTable, stmts,
//	    semantic.WithCaptures(captureResult))
//
//	// Pass 4
//	depResult := semantic.ValidateDependencies(collectResult.SymbolTable.Targets)
package semantic
