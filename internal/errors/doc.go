// Package errors provides structured error formatting for the build tool.
//
// The package implements error message templates with:
//   - Error codes (E001, E100, E200, etc.)
//   - Source context with line numbers
//   - Caret pointers to error location
//   - Notes for additional context
//   - Help suggestions for fixes
//
// # Error Format
//
// Errors are formatted in a Rust-like style:
//
//	error[E100]: missing ':' in target definition
//	 --> Needfile:3:10
//	2 | cc = gcc
//	3 | build/app deps
//	  |          ^
//	4 |     gcc -o build/app deps
//	note: targets require ':' before dependencies
//	help: change to: build/app: deps
//
// # Error Code Categories
//
//   - E001-E099: Lexical errors (invalid character, bad indentation)
//   - E100-E199: Syntax errors (unexpected token, missing colon)
//   - E200-E299: Semantic errors (undefined variable, duplicate definition)
//   - E300-E399: Evaluation errors (shell failure, bad function args)
//   - E400-E499: Execution errors (recipe failed, missing file)
package errors
