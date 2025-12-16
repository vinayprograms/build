package ast

import "fmt"

// ----------------------------------------------------------------------------
// Environment Types
// ----------------------------------------------------------------------------

// Runtime represents the runtime type for an environment.
type Runtime int

const (
	RuntimeBare Runtime = iota
	RuntimeDocker
	RuntimePodman
	RuntimeDevcontainer
	RuntimeNix
	RuntimeLima
)

// String returns the string representation of the runtime.
func (r Runtime) String() string {
	switch r {
	case RuntimeBare:
		return "bare"
	case RuntimeDocker:
		return "docker"
	case RuntimePodman:
		return "podman"
	case RuntimeDevcontainer:
		return "devcontainer"
	case RuntimeNix:
		return "nix"
	case RuntimeLima:
		return "lima"
	default:
		return fmt.Sprintf("unknown(%d)", r)
	}
}

// VersionSpec is the interface for version specifications in .requires directives.
//
// The versionSpecNode() marker method is unexported to prevent external packages
// from implementing this interface, ensuring a closed set of version spec types.
//
// Implementers: VersionLatest, VersionMajor, VersionMajorMinor, VersionExact
type VersionSpec interface {
	versionSpecNode()
	String() string
}

// VersionLatest represents "latest" or unspecified version.
type VersionLatest struct{}

func (v VersionLatest) versionSpecNode() {}
func (v VersionLatest) String() string   { return "latest" }

// VersionMajor represents a major version (e.g., "11").
type VersionMajor struct {
	Major int
}

func (v VersionMajor) versionSpecNode() {}
func (v VersionMajor) String() string   { return fmt.Sprintf("%d", v.Major) }

// VersionMajorMinor represents a major.minor version (e.g., "11.4").
type VersionMajorMinor struct {
	Major int
	Minor int
}

func (v VersionMajorMinor) versionSpecNode() {}
func (v VersionMajorMinor) String() string   { return fmt.Sprintf("%d.%d", v.Major, v.Minor) }

// VersionExact represents an exact version (e.g., "11.4.0").
type VersionExact struct {
	Major int
	Minor int
	Patch int
}

func (v VersionExact) versionSpecNode() {}
func (v VersionExact) String() string   { return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch) }

// Requirement represents a required binary with optional version spec.
type Requirement struct {
	Name    string
	Version VersionSpec
}

// Environment represents an environment block.
type Environment struct {
	Name     *string       // nil for default environment
	Runtime  *Runtime      // .using
	Source   *Value        // .source
	Args     *Value        // .args
	Requires []Requirement // .requires
	Location SourceLocation
}

func (e *Environment) statementNode() {}

// String returns a human-readable representation of the environment.
func (e *Environment) String() string {
	name := "(default)"
	if e.Name != nil {
		name = *e.Name
	}
	if e.Runtime != nil {
		return fmt.Sprintf(".environment: %s (%s)", name, e.Runtime.String())
	}
	return fmt.Sprintf(".environment: %s", name)
}
