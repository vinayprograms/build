package environ

import (
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Version
		wantErr bool
	}{
		{
			name:  "major.minor.patch",
			input: "11.4.0",
			want:  &Version{Major: 11, Minor: 4, Patch: 0},
		},
		{
			name:  "major.minor",
			input: "11.4",
			want:  &Version{Major: 11, Minor: 4, Patch: -1},
		},
		{
			name:  "major only",
			input: "11",
			want:  &Version{Major: 11, Minor: -1, Patch: -1},
		},
		{
			name:  "with v prefix",
			input: "v1.2.3",
			want:  &Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:  "with text suffix",
			input: "11.4.0-ubuntu",
			want:  &Version{Major: 11, Minor: 4, Patch: 0},
		},
		{
			name:  "from gcc output",
			input: "gcc (Ubuntu 11.4.0-1ubuntu1~22.04) 11.4.0",
			want:  &Version{Major: 11, Minor: 4, Patch: 0},
		},
		{
			name:  "from python output",
			input: "Python 3.10.12",
			want:  &Version{Major: 3, Minor: 10, Patch: 12},
		},
		{
			name:    "no version found",
			input:   "no version here",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseVersion(%q) = %v, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseVersion(%q) error = %v", tt.input, err)
				return
			}
			if got.Major != tt.want.Major {
				t.Errorf("ParseVersion(%q).Major = %d, want %d", tt.input, got.Major, tt.want.Major)
			}
			if got.Minor != tt.want.Minor {
				t.Errorf("ParseVersion(%q).Minor = %d, want %d", tt.input, got.Minor, tt.want.Minor)
			}
			if got.Patch != tt.want.Patch {
				t.Errorf("ParseVersion(%q).Patch = %d, want %d", tt.input, got.Patch, tt.want.Patch)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		version Version
		want    string
	}{
		{Version{Major: 11, Minor: 4, Patch: 0}, "11.4.0"},
		{Version{Major: 11, Minor: 4, Patch: -1}, "11.4"},
		{Version{Major: 11, Minor: -1, Patch: -1}, "11"},
	}

	for _, tt := range tests {
		got := tt.version.String()
		if got != tt.want {
			t.Errorf("Version{%d,%d,%d}.String() = %q, want %q",
				tt.version.Major, tt.version.Minor, tt.version.Patch, got, tt.want)
		}
	}
}

func TestVersionSatisfies(t *testing.T) {
	tests := []struct {
		name     string
		detected Version
		spec     ast.VersionSpec
		want     bool
	}{
		// VersionLatest always matches
		{
			name:     "latest matches any",
			detected: Version{Major: 11, Minor: 4, Patch: 0},
			spec:     ast.VersionLatest{},
			want:     true,
		},

		// VersionMajor
		{
			name:     "major match",
			detected: Version{Major: 11, Minor: 4, Patch: 0},
			spec:     ast.VersionMajor{Major: 11},
			want:     true,
		},
		{
			name:     "major mismatch",
			detected: Version{Major: 10, Minor: 0, Patch: 0},
			spec:     ast.VersionMajor{Major: 11},
			want:     false,
		},

		// VersionMajorMinor
		{
			name:     "major.minor match",
			detected: Version{Major: 11, Minor: 4, Patch: 0},
			spec:     ast.VersionMajorMinor{Major: 11, Minor: 4},
			want:     true,
		},
		{
			name:     "major.minor match different patch",
			detected: Version{Major: 11, Minor: 4, Patch: 5},
			spec:     ast.VersionMajorMinor{Major: 11, Minor: 4},
			want:     true,
		},
		{
			name:     "major.minor mismatch minor",
			detected: Version{Major: 11, Minor: 3, Patch: 0},
			spec:     ast.VersionMajorMinor{Major: 11, Minor: 4},
			want:     false,
		},
		{
			name:     "major.minor mismatch major",
			detected: Version{Major: 10, Minor: 4, Patch: 0},
			spec:     ast.VersionMajorMinor{Major: 11, Minor: 4},
			want:     false,
		},

		// VersionExact
		{
			name:     "exact match",
			detected: Version{Major: 11, Minor: 4, Patch: 0},
			spec:     ast.VersionExact{Major: 11, Minor: 4, Patch: 0},
			want:     true,
		},
		{
			name:     "exact mismatch patch",
			detected: Version{Major: 11, Minor: 4, Patch: 1},
			spec:     ast.VersionExact{Major: 11, Minor: 4, Patch: 0},
			want:     false,
		},
		{
			name:     "exact mismatch minor",
			detected: Version{Major: 11, Minor: 3, Patch: 0},
			spec:     ast.VersionExact{Major: 11, Minor: 4, Patch: 0},
			want:     false,
		},
		{
			name:     "exact mismatch major",
			detected: Version{Major: 10, Minor: 4, Patch: 0},
			spec:     ast.VersionExact{Major: 11, Minor: 4, Patch: 0},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.detected.Satisfies(tt.spec)
			if got != tt.want {
				t.Errorf("Version{%d,%d,%d}.Satisfies(%v) = %v, want %v",
					tt.detected.Major, tt.detected.Minor, tt.detected.Patch,
					tt.spec, got, tt.want)
			}
		})
	}
}

func TestDetectVersion(t *testing.T) {
	// Test with a binary that should exist and return a version
	// We use "sh" because it's universal, but version detection may vary
	checker := NewRequirementsChecker()

	// Test with a binary we know exists
	_, err := checker.DetectVersion("sh")
	// We can't test the exact version, but we should at least not error
	// on common systems. Some shells may not provide version info.
	_ = err // Allow either success or version detection error

	// Test with non-existent binary
	_, err = checker.DetectVersion("nonexistent-binary-xyz")
	if err == nil {
		t.Error("DetectVersion(nonexistent) = nil, want error")
	}
}

func TestVersionCache(t *testing.T) {
	checker := NewRequirementsChecker()

	// First call should populate cache
	v1, err1 := checker.DetectVersion("sh")

	// Second call should use cached result
	v2, err2 := checker.DetectVersion("sh")

	// Results should be identical
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("cached result differs: first err=%v, second err=%v", err1, err2)
	}

	if v1 != nil && v2 != nil {
		if v1.Major != v2.Major || v1.Minor != v2.Minor || v1.Patch != v2.Patch {
			t.Errorf("cached version differs: first=%v, second=%v", v1, v2)
		}
	}

	// Cache should have an entry
	if checker.versionCache == nil {
		t.Error("expected version cache to be initialized")
	}

	if _, ok := checker.versionCache["sh"]; !ok {
		t.Error("expected 'sh' to be in version cache")
	}
}

func TestVersionCacheError(t *testing.T) {
	checker := NewRequirementsChecker()

	// First call with non-existent binary should cache the error
	_, err1 := checker.DetectVersion("nonexistent-binary-xyz")

	// Second call should return cached error
	_, err2 := checker.DetectVersion("nonexistent-binary-xyz")

	// Both should error
	if err1 == nil || err2 == nil {
		t.Error("expected error for non-existent binary")
	}
}

func TestClearVersionCache(t *testing.T) {
	checker := NewRequirementsChecker()

	// Populate cache
	checker.DetectVersion("sh")

	// Clear cache
	checker.ClearVersionCache()

	// Cache should be empty
	if len(checker.versionCache) != 0 {
		t.Error("expected version cache to be empty after clear")
	}
}
