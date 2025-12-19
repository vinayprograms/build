package environ

import (
	"fmt"
	"os/exec"

	"github.com/vinayprograms/build/internal/ast"
)

// RequirementsChecker validates that required binaries are available.
type RequirementsChecker struct {
	// lookPath is the function used to find binaries in PATH.
	// Default is exec.LookPath, but can be overridden for testing.
	lookPath func(file string) (string, error)
}

// NewRequirementsChecker creates a new RequirementsChecker.
func NewRequirementsChecker() *RequirementsChecker {
	return &RequirementsChecker{
		lookPath: exec.LookPath,
	}
}

// RequirementResult holds the result of checking a single requirement.
type RequirementResult struct {
	Requirement     ast.Requirement // The requirement that was checked
	Found           bool            // True if binary was found in PATH
	Path            string          // Full path to the binary (if found)
	DetectedVersion string          // Version string detected (if any)
	Error           error           // Error if check failed
}

// String returns a human-readable status string for the result.
func (r RequirementResult) String() string {
	if !r.Found {
		return fmt.Sprintf("%s: not found", r.Requirement.Name)
	}

	// Check for version mismatch error
	if _, ok := r.Error.(*VersionMismatchError); ok {
		return fmt.Sprintf("%s: version mismatch (required %s, found %s)",
			r.Requirement.Name, r.Requirement.Version.String(), r.DetectedVersion)
	}

	if r.DetectedVersion != "" {
		return fmt.Sprintf("%s: found (version %s)", r.Requirement.Name, r.DetectedVersion)
	}
	return fmt.Sprintf("%s: found", r.Requirement.Name)
}

// CheckBinaryExists checks if a binary exists in PATH.
func (c *RequirementsChecker) CheckBinaryExists(name string) error {
	_, err := c.lookPath(name)
	if err != nil {
		return &BinaryNotFoundError{Name: name}
	}
	return nil
}

// CheckRequirement checks a single requirement.
// It verifies the binary exists and optionally checks the version.
func (c *RequirementsChecker) CheckRequirement(req ast.Requirement) RequirementResult {
	result := RequirementResult{
		Requirement: req,
	}

	// Check if binary exists
	path, err := c.lookPath(req.Name)
	if err != nil {
		result.Error = &BinaryNotFoundError{Name: req.Name}
		return result
	}

	result.Found = true
	result.Path = path

	// For VersionLatest, we don't need to check version
	if _, ok := req.Version.(ast.VersionLatest); ok {
		return result
	}

	// Version checking is handled separately
	// (will be implemented in version.go)
	return result
}

// CheckRequirements checks multiple requirements and returns results for each.
// All requirements are checked even if some fail.
func (c *RequirementsChecker) CheckRequirements(reqs []ast.Requirement) []RequirementResult {
	results := make([]RequirementResult, len(reqs))
	for i, req := range reqs {
		results[i] = c.CheckRequirement(req)
	}
	return results
}

// HasErrors returns true if any requirement check failed.
func HasErrors(results []RequirementResult) bool {
	for _, r := range results {
		if r.Error != nil {
			return true
		}
	}
	return false
}

// GetErrors returns all errors from the results.
func GetErrors(results []RequirementResult) []error {
	var errs []error
	for _, r := range results {
		if r.Error != nil {
			errs = append(errs, r.Error)
		}
	}
	return errs
}
