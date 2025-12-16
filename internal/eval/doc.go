// Package eval provides variable evaluation for Buildfiles.
//
// The evaluation phase occurs after semantic analysis and before build planning.
// It evaluates immediate variables, handles conditionals, and prepares the context
// for recipe execution.
//
// The evaluation process:
//  1. Initialize context with built-in variables (os, arch)
//  2. Evaluate conditionals to determine which branches apply
//  3. Evaluate immediate variables in definition order
//  4. Store lazy variables for on-demand evaluation
//
// Built-in variables:
//   - os: Operating system name (runtime.GOOS)
//   - arch: Architecture name (runtime.GOARCH)
package eval
