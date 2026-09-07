// Package environ provides environment management for the build tool.
//
// It handles:
//   - Requirements checking for bare environments
//   - Version detection and validation
//   - Binary existence verification
//   - Container environment support (Docker/Podman)
//   - Image building and caching
//   - Container execution with workspace mounting
//
// # Bare Environments
//
// The package supports version specifications as defined in the NEEDFILE_SPEC.md:
//   - name       → any version (alias for @latest)
//   - name@latest → latest available
//   - name@11    → major version 11.x.x
//   - name@11.4  → version 11.4.x
//   - name@11.4.0 → exact version
//
// # Container Environments
//
// Container environments (Docker/Podman) are configured with:
//   - .using: docker or .using: podman
//   - .source: path to Dockerfile
//   - .args: additional arguments (e.g., --platform linux/amd64)
//
// The container environment:
//   - Validates that the Dockerfile exists and is valid
//   - Builds images with project-specific tags
//   - Mounts the workspace directory at /workspace
//   - Supports interactive shell access (--shell flag)
//   - Supports keep-alive mode (--keep flag)
package environ
