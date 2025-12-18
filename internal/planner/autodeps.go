package planner

import (
	"os"
	"strings"
)

// ParseAutodeps parses Makefile-style dependency content (from gcc -MD, etc.)
// and returns a list of dependency paths.
//
// Format:
//
//	target: dep1 dep2 dep3
//	target: dep1 \
//	  dep2 \
//	  dep3
func ParseAutodeps(content string) ([]string, error) {
	if content == "" {
		return nil, nil
	}

	// Handle backslash line continuations
	content = strings.ReplaceAll(content, "\\\n", " ")
	content = strings.ReplaceAll(content, "\\\r\n", " ")

	var deps []string
	seen := make(map[string]bool)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Find colon separating target from dependencies
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}

		// Dependencies are after the colon
		depStr := strings.TrimSpace(line[colonIdx+1:])
		if depStr == "" {
			continue
		}

		// Parse space-separated dependencies, handling escaped spaces
		parts := parseDepList(depStr)
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" && !seen[p] {
				deps = append(deps, p)
				seen[p] = true
			}
		}
	}

	return deps, nil
}

// parseDepList splits a dependency string into individual paths,
// handling escaped spaces (backslash-space).
func parseDepList(s string) []string {
	var parts []string
	var current strings.Builder

	for i := 0; i < len(s); i++ {
		c := s[i]

		if c == '\\' && i+1 < len(s) && s[i+1] == ' ' {
			// Escaped space - include the space in current token
			current.WriteByte(' ')
			i++ // Skip the space
		} else if c == ' ' || c == '\t' {
			// Separator
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(c)
		}
	}

	// Don't forget the last part
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// ParseAutodepsFile reads and parses a .d file.
// Returns empty slice (not error) if file doesn't exist.
func ParseAutodepsFile(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	return ParseAutodeps(string(content))
}
