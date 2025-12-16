package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/lexer"
)

// parseRequirementsList parses a space-separated list of requirements.
// Each requirement has the form: name[@version]
func (p *Parser) parseRequirementsList() ([]ast.Requirement, *ParseError) {
	var reqs []ast.Requirement

	// Collect all text from the line
	var text string
	for p.current.Type != lexer.NEWLINE && p.current.Type != lexer.COMMENT && p.current.Type != lexer.EOF {
		text += p.current.Literal
		p.nextToken()
	}

	// Split on whitespace
	parts := strings.Fields(text)
	for _, part := range parts {
		req, err := parseRequirement(part)
		if err != nil {
			return nil, &ParseError{
				Message:  err.Error(),
				Location: p.current.Location,
			}
		}
		reqs = append(reqs, req)
	}

	return reqs, nil
}

// parseRequirement parses a single requirement string like "gcc@11.4".
func parseRequirement(s string) (ast.Requirement, error) {
	// Split on @
	parts := strings.SplitN(s, "@", 2)
	name := parts[0]

	var version ast.VersionSpec = ast.VersionLatest{}

	if len(parts) == 2 {
		verStr := parts[1]
		if verStr == "latest" {
			version = ast.VersionLatest{}
		} else {
			var err error
			version, err = parseVersionSpec(verStr)
			if err != nil {
				return ast.Requirement{}, fmt.Errorf("invalid version '%s' for '%s': %w", verStr, name, err)
			}
		}
	}

	return ast.Requirement{
		Name:    name,
		Version: version,
	}, nil
}

// parseVersionSpec parses a version string into a VersionSpec.
// Returns an error for invalid version formats.
func parseVersionSpec(s string) (ast.VersionSpec, error) {
	parts := strings.Split(s, ".")

	switch len(parts) {
	case 1:
		// Major only
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid major version: %s", parts[0])
		}
		return ast.VersionMajor{Major: major}, nil

	case 2:
		// Major.Minor
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid major version: %s", parts[0])
		}
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid minor version: %s", parts[1])
		}
		return ast.VersionMajorMinor{Major: major, Minor: minor}, nil

	case 3:
		// Exact version
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid major version: %s", parts[0])
		}
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid minor version: %s", parts[1])
		}
		patch, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %s", parts[2])
		}
		return ast.VersionExact{Major: major, Minor: minor, Patch: patch}, nil

	default:
		return nil, fmt.Errorf("invalid version format: expected 1-3 parts, got %d", len(parts))
	}
}
