// Package parser implements syntactic analysis for Buildfiles.
//
// The parser transforms a token stream from the lexer into an AST. It maintains
// scope context to validate directive placement and handles error recovery to
// collect multiple errors from a single parse.
//
// # Scope Stack
//
// The parser tracks nested scopes during parsing:
//   - ScopeGlobal: Top-level scope
//   - ScopeEnvironment: Inside .environment: block
//   - ScopeRecipe: Inside a target's recipe
//   - ScopeBlock: Inside block: within a recipe
//
// Directives are validated against their allowed scopes. For example, .shell:
// is valid at GLOBAL and RECIPE scopes, while .using: is only valid in
// ENVIRONMENT scope.
//
// # Error Handling
//
// The parser uses two complementary error handling patterns:
//
// Top-level parsing functions (ParseVariable, ParseTarget, ParseConditional,
// etc.) return *ParseError. This allows ParseBuildfile to catch errors and
// perform recovery via recoverToLevel0().
//
// Value/content parsing functions (ParseValue, parseInterpolation,
// parseFunctionCall, etc.) use addError() internally and return partial
// results. This enables continued parsing despite malformed values.
//
// This two-tier pattern supports error recovery: structural errors stop
// the current block and trigger recovery, while value-level errors are
// collected but don't halt parsing.
//
// # Error Recovery
//
// On parse error, the parser:
//  1. Records the error in the ParseErrors collection
//  2. Skips to the next line at indentation level 0 (global scope)
//  3. Continues parsing from there
//  4. Stops after maxErrors (10) to avoid infinite loops
//
// This enables partial analysis of broken files and provides comprehensive
// error feedback rather than stopping at the first error.
//
// # Parsing Buildfiles
//
// Use ParseBuildfile() for complete buildfile parsing with error recovery:
//
//	lex := lexer.New("Buildfile", content)
//	p := parser.New(lex)
//	stmts, errs := p.ParseBuildfile()
//
// For parsing individual constructs (useful in testing or tooling), use
// the specific Parse* methods: ParseVariable, ParseTarget, ParseEnvironment,
// ParseConditional, ParseInclude.
package parser
