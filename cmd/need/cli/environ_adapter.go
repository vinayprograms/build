package cli

import (
	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/environ"
)

// ----------------------------------------------------------------------------
// Requirement Adapter
// ----------------------------------------------------------------------------

// requirementAdapter wraps ast.Requirement to implement the Requirement interface.
type requirementAdapter struct {
	req ast.Requirement
}

func (r *requirementAdapter) Name() string {
	return r.req.Name
}

func (r *requirementAdapter) VersionSpec() string {
	return r.req.Version.String()
}

// ----------------------------------------------------------------------------
// RequirementResult Adapter
// ----------------------------------------------------------------------------

// requirementResultAdapter wraps environ.RequirementResult to implement the RequirementResult interface.
type requirementResultAdapter struct {
	result environ.RequirementResult
}

func (r *requirementResultAdapter) Name() string {
	return r.result.Requirement.Name
}

func (r *requirementResultAdapter) VersionSpec() string {
	return r.result.Requirement.Version.String()
}

func (r *requirementResultAdapter) Found() bool {
	return r.result.Found
}

func (r *requirementResultAdapter) Path() string {
	return r.result.Path
}

func (r *requirementResultAdapter) DetectedVersion() string {
	return r.result.DetectedVersion
}

func (r *requirementResultAdapter) Error() error {
	return r.result.Error
}

func (r *requirementResultAdapter) String() string {
	return r.result.String()
}

// ----------------------------------------------------------------------------
// RequirementsChecker Adapter
// ----------------------------------------------------------------------------

// requirementsCheckerAdapter wraps environ.RequirementsChecker to implement the RequirementsChecker interface.
type requirementsCheckerAdapter struct {
	checker *environ.RequirementsChecker
}

// NewRequirementsChecker creates a new RequirementsChecker.
func NewRequirementsChecker() RequirementsChecker {
	return &requirementsCheckerAdapter{
		checker: environ.NewRequirementsChecker(),
	}
}

func (c *requirementsCheckerAdapter) CheckBinaryExists(name string) error {
	return c.checker.CheckBinaryExists(name)
}

func (c *requirementsCheckerAdapter) CheckRequirements(reqs []Requirement) []RequirementResult {
	// Convert to ast.Requirement slice
	astReqs := make([]ast.Requirement, len(reqs))
	for i, r := range reqs {
		if ra, ok := r.(*requirementAdapter); ok {
			astReqs[i] = ra.req
		}
	}

	results := c.checker.CheckRequirements(astReqs)
	adapted := make([]RequirementResult, len(results))
	for i, r := range results {
		adapted[i] = &requirementResultAdapter{result: r}
	}
	return adapted
}

func (c *requirementsCheckerAdapter) CheckRequirementsWithVersion(reqs []Requirement) []RequirementResult {
	// Convert to ast.Requirement slice
	astReqs := make([]ast.Requirement, len(reqs))
	for i, r := range reqs {
		if ra, ok := r.(*requirementAdapter); ok {
			astReqs[i] = ra.req
		}
	}

	results := c.checker.CheckRequirementsWithVersion(astReqs)
	adapted := make([]RequirementResult, len(results))
	for i, r := range results {
		adapted[i] = &requirementResultAdapter{result: r}
	}
	return adapted
}

// CheckEnvironmentRequirements checks the requirements for an environment.
// It accepts the environment and returns the results.
func CheckEnvironmentRequirements(env *ast.Environment, withVersion bool) []RequirementResult {
	checker := environ.NewRequirementsChecker()
	var results []environ.RequirementResult

	if withVersion {
		results = checker.CheckRequirementsWithVersion(env.Requires)
	} else {
		results = checker.CheckRequirements(env.Requires)
	}

	adapted := make([]RequirementResult, len(results))
	for i, r := range results {
		adapted[i] = &requirementResultAdapter{result: r}
	}
	return adapted
}

// RequirementFromAST creates a Requirement interface from an ast.Requirement.
func RequirementFromAST(req ast.Requirement) Requirement {
	return &requirementAdapter{req: req}
}

// ----------------------------------------------------------------------------
// PackageManager Adapter
// ----------------------------------------------------------------------------

// packageManagerAdapter wraps environ.PackageManager to implement the PackageManager interface.
type packageManagerAdapter struct {
	pm environ.PackageManager
}

func (p *packageManagerAdapter) Name() string {
	return p.pm.Name()
}

func (p *packageManagerAdapter) GetInstallCommand(binary string) string {
	return p.pm.GetInstallCommand(binary)
}

// DetectPackageManager detects the system's package manager.
// Returns nil if no supported package manager is found.
func DetectPackageManager() PackageManager {
	pm := environ.DetectPackageManager()
	if pm == nil {
		return nil
	}
	return &packageManagerAdapter{pm: pm}
}

// GetInstallSuggestion returns the install command for a binary.
// Returns empty string if package manager is nil.
func GetInstallSuggestion(binary string, pm PackageManager) string {
	if pm == nil {
		return ""
	}
	return pm.GetInstallCommand(binary)
}
