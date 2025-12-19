package environ

import "fmt"

// BinaryNotFoundError indicates a required binary was not found in PATH.
type BinaryNotFoundError struct {
	Name string // Binary name
}

func (e *BinaryNotFoundError) Error() string {
	return fmt.Sprintf("required binary '%s' not found in PATH", e.Name)
}

// VersionMismatchError indicates the found version doesn't match the requirement.
type VersionMismatchError struct {
	Name     string // Binary name
	Required string // Required version string
	Detected string // Detected version string
}

func (e *VersionMismatchError) Error() string {
	return fmt.Sprintf("version mismatch for '%s': required %s, found %s",
		e.Name, e.Required, e.Detected)
}

// VersionDetectionError indicates the version could not be detected.
type VersionDetectionError struct {
	Name    string // Binary name
	Message string // Error message
}

func (e *VersionDetectionError) Error() string {
	return fmt.Sprintf("unable to detect version for '%s': %s", e.Name, e.Message)
}

// EnvironmentNotFoundError indicates a named environment was not found.
type EnvironmentNotFoundError struct {
	Name string // Environment name
}

func (e *EnvironmentNotFoundError) Error() string {
	return fmt.Sprintf("environment '%s' not found", e.Name)
}

// NoDefaultEnvironmentError indicates no default environment is defined
// when only named environments exist.
type NoDefaultEnvironmentError struct {
	Available []string // List of available environment names
}

func (e *NoDefaultEnvironmentError) Error() string {
	if len(e.Available) == 0 {
		return "no environment defined"
	}
	return fmt.Sprintf("no default environment; use --env with one of: %v", e.Available)
}
