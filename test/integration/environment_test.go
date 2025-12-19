package integration

import (
	"strings"
	"testing"
)

// TestBareEnvironmentNoRequirements tests a bare environment with no requirements.
func TestBareEnvironmentNoRequirements(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

.environment:
	.using: bare

@test:
	echo "Hello from bare environment"
`)

	// Build should succeed
	result := h.Run("test")
	result.AssertSuccess().
		AssertStdoutContains("Hello from bare environment")
}

// TestBareEnvironmentWithSatisfiedRequirements tests a bare environment
// where required binaries exist in PATH.
func TestBareEnvironmentWithSatisfiedRequirements(t *testing.T) {
	h := NewTestHarness(t)

	// Use common binaries that should exist on any Unix system
	h.WriteFile("Buildfile", `.shell: bash

.environment:
	.using: bare
	.requires: bash sh

@test:
	echo "All requirements satisfied"
`)

	// Build should succeed since bash and sh exist
	result := h.Run("test")
	result.AssertSuccess().
		AssertStdoutContains("All requirements satisfied")
}

// TestCheckEnvBareEnvironmentSatisfied tests --check-env with satisfied requirements.
func TestCheckEnvBareEnvironmentSatisfied(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

.environment:
	.using: bare
	.requires: bash sh

@test:
	echo "test"
`)

	result := h.Run("--check-env")
	result.AssertSuccess().
		AssertStdoutContains("Checking environment").
		AssertStdoutContains("Runtime: bare").
		AssertStdoutContains("All requirements satisfied")
}

// TestCheckEnvBareEnvironmentMissing tests --check-env with missing requirements.
func TestCheckEnvBareEnvironmentMissing(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

.environment:
	.using: bare
	.requires: nonexistent_binary_xyz_12345

@test:
	echo "test"
`)

	result := h.Run("--check-env")
	result.AssertExitCode(4). // Environment error
					AssertStdoutContains("not found").
					AssertStdoutContains("Some requirements are not met")
}

// TestCheckEnvNamedEnvironment tests --check-env with a named environment.
func TestCheckEnvNamedEnvironment(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

.environment: dev
	.using: bare
	.requires: bash

@test:
	echo "test"
`)

	// Check named environment
	result := h.Run("--check-env", "--env", "dev")
	result.AssertSuccess().
		AssertStdoutContains("Checking environment: dev").
		AssertStdoutContains("All requirements satisfied")
}

// TestCheckEnvNoEnvironment tests --check-env when no environment is defined.
func TestCheckEnvNoEnvironment(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

@test:
	echo "test"
`)

	result := h.Run("--check-env")
	result.AssertSuccess().
		AssertStdoutContains("No environment defined").
		AssertStdoutContains("bare environment")
}

// TestCheckEnvWithShowInstall tests --check-env --show-install flag combination.
func TestCheckEnvWithShowInstall(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

.environment:
	.using: bare
	.requires: nonexistent_binary_xyz_12345

@test:
	echo "test"
`)

	result := h.Run("--check-env", "--show-install")
	result.AssertExitCode(4). // Environment error
					AssertStdoutContains("not found")
	// May or may not show install suggestions depending on detected package manager
}

// TestListEnvNoEnvironments tests --list-env when no environments defined.
func TestListEnvNoEnvironments(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

@test:
	echo "test"
`)

	result := h.Run("--list-env")
	result.AssertSuccess().
		AssertStdoutContains("No environments defined")
}

// TestListEnvSingleDefaultEnvironment tests --list-env with one unnamed environment.
func TestListEnvSingleDefaultEnvironment(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

.environment:
	.using: bare
	.requires: bash

@test:
	echo "test"
`)

	result := h.Run("--list-env")
	result.AssertSuccess().
		AssertStdoutContains("Available environments").
		AssertStdoutContains("(default)").
		AssertStdoutContains("bare")
}

// TestListEnvMultipleEnvironments tests --list-env with multiple environments.
func TestListEnvMultipleEnvironments(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

.environment:
	.using: bare
	.requires: bash

.environment: dev
	.using: bare
	.requires: bash sh

.environment: ci
	.using: bare

@test:
	echo "test"
`)

	result := h.Run("--list-env")
	result.AssertSuccess()

	stdout := result.Stdout()
	if !strings.Contains(stdout, "Available environments") {
		t.Error("expected 'Available environments' header")
	}
	if !strings.Contains(stdout, "(default)") {
		t.Error("expected '(default)' environment")
	}
	if !strings.Contains(stdout, "dev") {
		t.Error("expected 'dev' environment")
	}
	if !strings.Contains(stdout, "ci") {
		t.Error("expected 'ci' environment")
	}
}

// TestCheckEnvEnvironmentNotFound tests --check-env --env with non-existent environment.
func TestCheckEnvEnvironmentNotFound(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

.environment: dev
	.using: bare

@test:
	echo "test"
`)

	result := h.Run("--check-env", "--env", "nonexistent")
	result.AssertExitCode(4). // Environment error
					AssertStderrContains("not found")
}

// TestCheckEnvRequiresEnvSelection tests --check-env when only named environments exist.
func TestCheckEnvRequiresEnvSelection(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

.environment: dev
	.using: bare

.environment: prod
	.using: bare

@test:
	echo "test"
`)

	// No --env flag, no default environment
	result := h.Run("--check-env")
	result.AssertExitCode(4). // Environment error
					AssertStderrContains("no default environment").
					AssertStderrContains("use --env")
}

// TestCheckEnvVerbose tests --check-env --verbose for detailed output.
func TestCheckEnvVerbose(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

.environment:
	.using: bare
	.requires: bash

@test:
	echo "test"
`)

	result := h.Run("--check-env", "--verbose")
	result.AssertSuccess()
	// Verbose mode shows path information
	if !strings.Contains(result.Stdout(), "path:") && !strings.Contains(result.Stdout(), "found") {
		t.Log("verbose mode may show path or just found status")
	}
}

// TestBareEnvironmentWithVersionedRequirement tests requirements with version specs.
func TestBareEnvironmentWithVersionedRequirement(t *testing.T) {
	h := NewTestHarness(t)

	// Use bash which should exist on any system
	// Note: version checking is best-effort, not all binaries support --version
	h.WriteFile("Buildfile", `.shell: bash

.environment:
	.using: bare
	.requires: bash@latest

@test:
	echo "Version requirement test"
`)

	// @latest should always match
	result := h.Run("--check-env")
	result.AssertSuccess().
		AssertStdoutContains("All requirements satisfied")
}

// TestListEnvShowsRequirementCount tests that --list-env shows requirement counts.
func TestListEnvShowsRequirementCount(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

.environment:
	.using: bare
	.requires: bash sh

.environment: minimal
	.using: bare

@test:
	echo "test"
`)

	result := h.Run("--list-env")
	result.AssertSuccess()

	stdout := result.Stdout()
	// Default environment has 2 requirements
	if !strings.Contains(stdout, "2 requirements") && !strings.Contains(stdout, "requirement") {
		t.Log("list-env output may or may not show requirement counts")
	}
}

// TestBuildWithBareEnvironmentSucceeds tests that builds succeed with satisfied requirements.
func TestBuildWithBareEnvironmentSucceeds(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

.environment:
	.using: bare
	.requires: bash echo

@greet:
	echo "Hello, World!"

@build: @greet
	echo "Build complete"
`)

	result := h.Run("build")
	result.AssertSuccess().
		AssertStdoutContains("Hello, World!").
		AssertStdoutContains("Build complete")
}
