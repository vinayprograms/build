package environ

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
)

// Version represents a parsed semantic version.
type Version struct {
	Major int // Major version number
	Minor int // Minor version number (-1 if not specified)
	Patch int // Patch version number (-1 if not specified)
}

// String returns the version as a string.
func (v Version) String() string {
	if v.Minor < 0 {
		return strconv.Itoa(v.Major)
	}
	if v.Patch < 0 {
		return fmt.Sprintf("%d.%d", v.Major, v.Minor)
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Satisfies checks if this version satisfies a version spec.
func (v Version) Satisfies(spec ast.VersionSpec) bool {
	switch s := spec.(type) {
	case ast.VersionLatest:
		return true

	case ast.VersionMajor:
		return v.Major == s.Major

	case ast.VersionMajorMinor:
		return v.Major == s.Major && v.Minor == s.Minor

	case ast.VersionExact:
		return v.Major == s.Major && v.Minor == s.Minor && v.Patch == s.Patch

	default:
		return false
	}
}

// versionPattern matches version strings like "11.4.0", "v1.2.3", etc.
// Captures: 1=major, 2=minor (optional), 3=patch (optional)
var versionPattern = regexp.MustCompile(`(?:^|[^0-9])v?(\d+)(?:\.(\d+)(?:\.(\d+))?)?`)

// ParseVersion extracts version information from a string.
// It searches for version patterns like "11.4.0", "v1.2.3", etc.
func ParseVersion(s string) (*Version, error) {
	if s == "" {
		return nil, fmt.Errorf("empty version string")
	}

	// Try to find a version pattern
	matches := versionPattern.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("no version found in: %s", s)
	}

	major, _ := strconv.Atoi(matches[1])
	v := &Version{
		Major: major,
		Minor: -1,
		Patch: -1,
	}

	if matches[2] != "" {
		minor, _ := strconv.Atoi(matches[2])
		v.Minor = minor
	}

	if matches[3] != "" {
		patch, _ := strconv.Atoi(matches[3])
		v.Patch = patch
	}

	return v, nil
}

// DetectVersion attempts to detect the version of a binary.
// It tries common version flags (--version, -version, -v).
func (c *RequirementsChecker) DetectVersion(name string) (*Version, error) {
	// First check if binary exists
	path, err := c.lookPath(name)
	if err != nil {
		return nil, &BinaryNotFoundError{Name: name}
	}

	// Try common version flags
	versionFlags := [][]string{
		{"--version"},
		{"-version"},
		{"-v"},
	}

	var lastOutput string
	for _, flags := range versionFlags {
		args := append([]string{}, flags...)
		cmd := exec.Command(path, args...)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		// Some tools exit non-zero for --version, that's ok
		_ = cmd.Run()

		output := stdout.String()
		if output == "" {
			output = stderr.String()
		}

		if output != "" {
			lastOutput = output
			// Try to parse version from output
			v, err := ParseVersion(output)
			if err == nil {
				return v, nil
			}
		}
	}

	// Could not detect version
	if lastOutput != "" {
		// We got output but couldn't parse it
		return nil, &VersionDetectionError{
			Name:    name,
			Message: fmt.Sprintf("could not parse version from output: %s", strings.TrimSpace(lastOutput)),
		}
	}

	return nil, &VersionDetectionError{
		Name:    name,
		Message: "no version output from any version flag",
	}
}

// CheckRequirementWithVersion checks a requirement including version validation.
func (c *RequirementsChecker) CheckRequirementWithVersion(req ast.Requirement) RequirementResult {
	result := c.CheckRequirement(req)
	if result.Error != nil {
		return result
	}

	// Skip version check for VersionLatest
	if _, ok := req.Version.(ast.VersionLatest); ok {
		return result
	}

	// Try to detect version
	detected, err := c.DetectVersion(req.Name)
	if err != nil {
		// Version detection failed, but binary exists
		// For non-VersionLatest requirements, this is an error
		result.Error = err
		return result
	}

	result.DetectedVersion = detected.String()

	// Check if version satisfies requirement
	if !detected.Satisfies(req.Version) {
		result.Error = &VersionMismatchError{
			Name:     req.Name,
			Required: req.Version.String(),
			Detected: detected.String(),
		}
	}

	return result
}

// CheckRequirementsWithVersion checks multiple requirements with version validation.
func (c *RequirementsChecker) CheckRequirementsWithVersion(reqs []ast.Requirement) []RequirementResult {
	results := make([]RequirementResult, len(reqs))
	for i, req := range reqs {
		results[i] = c.CheckRequirementWithVersion(req)
	}
	return results
}
