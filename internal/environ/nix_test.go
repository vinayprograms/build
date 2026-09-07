package environ

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/eval"
)

func TestNixDetector_DetectConfig_ShellNix(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "nix-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create shell.nix
	shellNixPath := filepath.Join(tmpDir, "shell.nix")
	shellNix := `{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = [ pkgs.go ];
}
`
	if err := os.WriteFile(shellNixPath, []byte(shellNix), 0644); err != nil {
		t.Fatal(err)
	}

	// Test detection
	detector := NewNixDetector()
	result, err := detector.DetectConfig(tmpDir, nil, eval.NewContext())

	if err != nil {
		t.Fatalf("DetectConfig() error = %v", err)
	}
	if !result.Found {
		t.Error("expected config to be found")
	}
	if result.Path != shellNixPath {
		t.Errorf("Path = %q, want %q", result.Path, shellNixPath)
	}
	if result.Type != NixTypeShell {
		t.Errorf("Type = %v, want %v", result.Type, NixTypeShell)
	}
}

func TestNixDetector_DetectConfig_FlakeNix(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "nix-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create flake.nix
	flakeNixPath := filepath.Join(tmpDir, "flake.nix")
	flakeNix := `{
  description = "Test flake";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  outputs = { self, nixpkgs }: {};
}
`
	if err := os.WriteFile(flakeNixPath, []byte(flakeNix), 0644); err != nil {
		t.Fatal(err)
	}

	// Test detection
	detector := NewNixDetector()
	result, err := detector.DetectConfig(tmpDir, nil, eval.NewContext())

	if err != nil {
		t.Fatalf("DetectConfig() error = %v", err)
	}
	if !result.Found {
		t.Error("expected config to be found")
	}
	if result.Path != flakeNixPath {
		t.Errorf("Path = %q, want %q", result.Path, flakeNixPath)
	}
	if result.Type != NixTypeFlake {
		t.Errorf("Type = %v, want %v", result.Type, NixTypeFlake)
	}
}

func TestNixDetector_DetectConfig_FromSource(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "nix-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create custom shell file
	customPath := filepath.Join(tmpDir, "custom-shell.nix")
	customShell := `{ pkgs ? import <nixpkgs> {} }: pkgs.mkShell {}`
	if err := os.WriteFile(customPath, []byte(customShell), 0644); err != nil {
		t.Fatal(err)
	}

	// Create source value that points to custom-shell.nix
	source := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.LiteralValue{Text: "custom-shell.nix"},
		},
	}

	// Test detection
	detector := NewNixDetector()
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

func TestNixDetector_DetectConfig_NotFound(t *testing.T) {
	// Create empty temp directory
	tmpDir, err := os.MkdirTemp("", "nix-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test detection
	detector := NewNixDetector()
	result, err := detector.DetectConfig(tmpDir, nil, eval.NewContext())

	if err != nil {
		t.Fatalf("DetectConfig() error = %v", err)
	}
	if result.Found {
		t.Error("expected config to not be found")
	}
}

func TestNixDetector_DetectConfig_ShellNixPriority(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "nix-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create both shell.nix and flake.nix
	shellNixPath := filepath.Join(tmpDir, "shell.nix")
	if err := os.WriteFile(shellNixPath, []byte("{ }:{}"), 0644); err != nil {
		t.Fatal(err)
	}

	flakeNixPath := filepath.Join(tmpDir, "flake.nix")
	if err := os.WriteFile(flakeNixPath, []byte("{ }"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test detection - shell.nix should take priority
	detector := NewNixDetector()
	result, err := detector.DetectConfig(tmpDir, nil, eval.NewContext())

	if err != nil {
		t.Fatalf("DetectConfig() error = %v", err)
	}
	if result.Path != shellNixPath {
		t.Errorf("Path = %q, want %q (shell.nix should take priority)", result.Path, shellNixPath)
	}
	if result.Type != NixTypeShell {
		t.Errorf("Type = %v, want %v", result.Type, NixTypeShell)
	}
}

func TestNixRunner_CheckCLI_NotInstalled(t *testing.T) {
	runner := NewNixRunner("/project")
	// Override lookPath to simulate nix not being installed
	runner.lookPath = func(name string) (string, error) {
		return "", &BinaryNotFoundError{Name: name}
	}

	err := runner.CheckCLI()
	if err == nil {
		t.Error("expected error for missing nix CLI")
	}
}
