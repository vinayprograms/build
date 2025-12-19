// Package platform provides cross-platform utilities for path handling,
// shell execution, and platform detection.
//
// This package abstracts platform-specific differences between Unix systems
// (Linux, macOS, BSD) and Windows, allowing the build tool to work consistently
// across all supported platforms.
//
// # Shell Execution
//
// On Unix systems, the default shell is /bin/sh with the -c flag.
// On Windows, cmd.exe is used with the /C flag, or PowerShell with -Command.
//
// # Path Handling
//
// Paths are normalized to use forward slashes internally for consistency.
// The package detects absolute paths on both Unix (starting with /) and
// Windows (drive letter or UNC paths).
//
// # Shell Quoting
//
// Shell quoting differs between platforms:
//   - Unix shells use single quotes (')
//   - cmd.exe uses double quotes (")
//   - PowerShell uses single quotes (')
package platform
