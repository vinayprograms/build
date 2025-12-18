// Package planner provides build planning for Buildfiles.
//
// Build planning involves:
//   - Target pattern matching (literal and pattern targets)
//   - Dependency resolution
//   - Staleness detection
//   - Build task ordering
//
// The planner operates on the validated AST and semantic analysis results
// to produce an executable build plan.
package planner
