package environ

import (
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

func TestCheckBinaryExists(t *testing.T) {
	tests := []struct {
		name    string
		binary  string
		wantErr bool
		errType string
	}{
		{
			name:    "existing binary sh",
			binary:  "sh",
			wantErr: false,
		},
		{
			name:    "existing binary ls",
			binary:  "ls",
			wantErr: false,
		},
		{
			name:    "non-existent binary",
			binary:  "this-binary-does-not-exist-12345",
			wantErr: true,
			errType: "*environ.BinaryNotFoundError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewRequirementsChecker()
			err := checker.CheckBinaryExists(tt.binary)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CheckBinaryExists(%q) = nil, want error", tt.binary)
				}
				if _, ok := err.(*BinaryNotFoundError); !ok {
					t.Errorf("CheckBinaryExists(%q) error type = %T, want *BinaryNotFoundError", tt.binary, err)
				}
			} else {
				if err != nil {
					t.Errorf("CheckBinaryExists(%q) = %v, want nil", tt.binary, err)
				}
			}
		})
	}
}

func TestCheckRequirement_ExistsNoVersion(t *testing.T) {
	checker := NewRequirementsChecker()

	// Test requirement with VersionLatest (any version)
	req := ast.Requirement{
		Name:    "sh", // Should exist on all Unix systems
		Version: ast.VersionLatest{},
	}

	result := checker.CheckRequirement(req)
	if result.Error != nil {
		t.Errorf("CheckRequirement(%q) = error %v, want success", req.Name, result.Error)
	}
	if !result.Found {
		t.Errorf("CheckRequirement(%q).Found = false, want true", req.Name)
	}
}

func TestCheckRequirement_NotFound(t *testing.T) {
	checker := NewRequirementsChecker()

	req := ast.Requirement{
		Name:    "non-existent-binary-xyz",
		Version: ast.VersionLatest{},
	}

	result := checker.CheckRequirement(req)
	if result.Error == nil {
		t.Errorf("CheckRequirement(%q) = nil error, want BinaryNotFoundError", req.Name)
	}
	if result.Found {
		t.Errorf("CheckRequirement(%q).Found = true, want false", req.Name)
	}
	if _, ok := result.Error.(*BinaryNotFoundError); !ok {
		t.Errorf("CheckRequirement(%q) error type = %T, want *BinaryNotFoundError", req.Name, result.Error)
	}
}

func TestCheckRequirements(t *testing.T) {
	checker := NewRequirementsChecker()

	reqs := []ast.Requirement{
		{Name: "sh", Version: ast.VersionLatest{}},
		{Name: "ls", Version: ast.VersionLatest{}},
	}

	results := checker.CheckRequirements(reqs)
	if len(results) != 2 {
		t.Fatalf("CheckRequirements returned %d results, want 2", len(results))
	}

	for i, result := range results {
		if result.Error != nil {
			t.Errorf("CheckRequirements[%d] (%q) error = %v, want nil", i, reqs[i].Name, result.Error)
		}
		if !result.Found {
			t.Errorf("CheckRequirements[%d] (%q).Found = false, want true", i, reqs[i].Name)
		}
	}
}

func TestCheckRequirements_WithFailures(t *testing.T) {
	checker := NewRequirementsChecker()

	reqs := []ast.Requirement{
		{Name: "sh", Version: ast.VersionLatest{}},
		{Name: "non-existent-xyz", Version: ast.VersionLatest{}},
		{Name: "ls", Version: ast.VersionLatest{}},
	}

	results := checker.CheckRequirements(reqs)
	if len(results) != 3 {
		t.Fatalf("CheckRequirements returned %d results, want 3", len(results))
	}

	// First should succeed
	if results[0].Error != nil || !results[0].Found {
		t.Errorf("CheckRequirements[0] (sh) should succeed")
	}

	// Second should fail
	if results[1].Error == nil || results[1].Found {
		t.Errorf("CheckRequirements[1] (non-existent) should fail")
	}

	// Third should succeed
	if results[2].Error != nil || !results[2].Found {
		t.Errorf("CheckRequirements[2] (ls) should succeed")
	}
}

func TestRequirementResult_String(t *testing.T) {
	tests := []struct {
		name   string
		result RequirementResult
		want   string
	}{
		{
			name: "found with version",
			result: RequirementResult{
				Requirement:     ast.Requirement{Name: "gcc", Version: ast.VersionLatest{}},
				Found:           true,
				DetectedVersion: "11.4.0",
			},
			want: "gcc: found (version 11.4.0)",
		},
		{
			name: "found without version",
			result: RequirementResult{
				Requirement: ast.Requirement{Name: "sh", Version: ast.VersionLatest{}},
				Found:       true,
			},
			want: "sh: found",
		},
		{
			name: "not found",
			result: RequirementResult{
				Requirement: ast.Requirement{Name: "gcc", Version: ast.VersionLatest{}},
				Found:       false,
				Error:       &BinaryNotFoundError{Name: "gcc"},
			},
			want: "gcc: not found",
		},
		{
			name: "version mismatch",
			result: RequirementResult{
				Requirement:     ast.Requirement{Name: "gcc", Version: ast.VersionMajor{Major: 12}},
				Found:           true,
				DetectedVersion: "11.4.0",
				Error:           &VersionMismatchError{Name: "gcc", Required: "12", Detected: "11.4.0"},
			},
			want: "gcc: version mismatch (required 12, found 11.4.0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.String()
			if got != tt.want {
				t.Errorf("RequirementResult.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBinaryNotFoundError(t *testing.T) {
	err := &BinaryNotFoundError{Name: "gcc"}
	want := "required binary 'gcc' not found in PATH"
	if got := err.Error(); got != want {
		t.Errorf("BinaryNotFoundError.Error() = %q, want %q", got, want)
	}
}

func TestVersionMismatchError(t *testing.T) {
	err := &VersionMismatchError{Name: "gcc", Required: "12", Detected: "11.4.0"}
	want := "version mismatch for 'gcc': required 12, found 11.4.0"
	if got := err.Error(); got != want {
		t.Errorf("VersionMismatchError.Error() = %q, want %q", got, want)
	}
}

func TestVersionDetectionError(t *testing.T) {
	err := &VersionDetectionError{Name: "gcc", Message: "unable to parse output"}
	want := "unable to detect version for 'gcc': unable to parse output"
	if got := err.Error(); got != want {
		t.Errorf("VersionDetectionError.Error() = %q, want %q", got, want)
	}
}
