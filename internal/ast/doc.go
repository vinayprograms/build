// Package ast defines the Abstract Syntax Tree node types for Buildfile parsing.
//
// The AST captures syntactic structure without interpretation—no evaluation
// happens during parsing. Each node carries a SourceLocation for precise error
// reporting with file:line:column format.
//
// # Statement Types
//
// All top-level AST nodes implement the Statement interface:
//   - Directive: Global directives (.shell:, .parallel:, .default:, .include:)
//   - Environment: Environment block (.environment:)
//   - Variable: Variable definition (immediate or lazy)
//   - Conditional: If/elif/else/end block
//   - Target: Target definition with dependencies and recipe
//   - Comment: Comment line
//   - Blank: Blank line
//
// # Design Decisions
//
// BraceExpr nodes remain unresolved during parsing. In target patterns, {name}
// could be either a capture (pattern matching variable) or a variable
// interpolation. The parser produces BraceExpr nodes; semantic analysis
// resolves them based on the symbol table.
//
// Interfaces use unexported marker methods (statementNode(), valuePartNode(),
// etc.) to enforce interface implementation at compile time while preventing
// external implementation. This ensures the set of node types is closed.
//
// Optional fields use nil pointers to indicate absence. For example,
// Environment.Name is nil for the default environment, and Recipe.Directives.Shell
// is nil when using the global default shell.
package ast
