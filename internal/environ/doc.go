// Package environ provides environment management for the build tool.
//
// It handles:
//   - Requirements checking for bare environments
//   - Version detection and validation
//   - Binary existence verification
//
// The package supports version specifications as defined in the BUILDFILE_SPEC.md:
//   - name       → any version (alias for @latest)
//   - name@latest → latest available
//   - name@11    → major version 11.x.x
//   - name@11.4  → version 11.4.x
//   - name@11.4.0 → exact version
package environ
