package environ

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/eval"
)

func TestLimaDetector_DetectConfig_LimaYaml(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "lima-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create lima.yaml
	limaYamlPath := filepath.Join(tmpDir, "lima.yaml")
	limaYaml := `# Lima configuration
arch: default
images:
  - location: "https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img"
    arch: "x86_64"
mounts:
  - location: "~"
    writable: true
`
	if err := os.WriteFile(limaYamlPath, []byte(limaYaml), 0644); err != nil {
		t.Fatal(err)
	}

	// Test detection
	detector := NewLimaDetector()
	result, err := detector.DetectConfig(tmpDir, nil, eval.NewContext())

	if err != nil {
		t.Fatalf("DetectConfig() error = %v", err)
	}
	if !result.Found {
		t.Error("expected config to be found")
	}
	if result.Path != limaYamlPath {
		t.Errorf("Path = %q, want %q", result.Path, limaYamlPath)
	}
}

func TestLimaDetector_DetectConfig_FromSource(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "lima-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create custom lima config
	customPath := filepath.Join(tmpDir, "dev.yaml")
	customYaml := `arch: default`
	if err := os.WriteFile(customPath, []byte(customYaml), 0644); err != nil {
		t.Fatal(err)
	}

	// Create source value that points to dev.yaml
	source := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.LiteralValue{Text: "dev.yaml"},
		},
	}

	// Test detection
	detector := NewLimaDetector()
	result, err := detector.DetectConfig(tmpDir, source, eval.NewContext())

	if err != nil {
		t.Fatalf("DetectConfig() error = %v", err)
	}
	if !result.Found {
		t.Error("expected config to be found")
	}
	if result.Path != customPath {
		t.Errorf("Path = %q, want %q", result.Path, customPath)
	}
}

func TestLimaDetector_DetectConfig_NotFound(t *testing.T) {
	// Create empty temp directory
	tmpDir, err := os.MkdirTemp("", "lima-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test detection
	detector := NewLimaDetector()
	result, err := detector.DetectConfig(tmpDir, nil, eval.NewContext())

	if err != nil {
		t.Fatalf("DetectConfig() error = %v", err)
	}
	if result.Found {
		t.Error("expected config to not be found")
	}
}

func TestLimaRunner_CheckCLI_NotInstalled(t *testing.T) {
	runner := NewLimaRunner("/project", "test")
	// Override lookPath to simulate lima not being installed
	runner.lookPath = func(name string) (string, error) {
		return "", &BinaryNotFoundError{Name: name}
	}

	err := runner.CheckCLI()
	if err == nil {
		t.Error("expected error for missing lima CLI")
	}
}

func TestLimaRunner_VMName(t *testing.T) {
	runner := NewLimaRunner("/project", "dev")
	if runner.vmName != "dev" {
		t.Errorf("vmName = %q, want %q", runner.vmName, "dev")
	}
}
