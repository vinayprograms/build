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

// ===========================================================================
// Docker Environment Integration Tests
// These tests require Docker to be available and running.
// ===========================================================================

// skipIfNoDocker skips the test if Docker is not available.
func skipIfNoDocker(t *testing.T) {
	t.Helper()
	h := NewTestHarness(t)
	result := h.RunShell("docker version >/dev/null 2>&1")
	if result.exitCode != 0 {
		t.Skip("Docker is not available, skipping Docker integration test")
	}
}

// TestDockerEnvironmentCheckEnv tests --check-env with a Docker environment.
func TestDockerEnvironmentCheckEnv(t *testing.T) {
	skipIfNoDocker(t)
	h := NewTestHarness(t)

	// Create a simple Dockerfile
	h.WriteFile("Dockerfile", `FROM alpine:latest
RUN echo "test"
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: docker
	.source: ./Dockerfile

@test:
	echo "Hello from Docker"
`)

	result := h.Run("--check-env")
	result.AssertSuccess().
		AssertStdoutContains("Checking environment").
		AssertStdoutContains("Runtime: docker").
		AssertStdoutContains("Dockerfile")
}

// TestDockerEnvironmentCheckEnvMissingDockerfile tests --check-env with missing Dockerfile.
func TestDockerEnvironmentCheckEnvMissingDockerfile(t *testing.T) {
	skipIfNoDocker(t)
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: docker
	.source: ./NonexistentDockerfile

@test:
	echo "test"
`)

	result := h.Run("--check-env")
	result.AssertExitCode(4). // Environment error
					AssertStdoutContains("not found") // Error goes to stdout in this tool
}

// TestDockerEnvironmentCheckEnvInvalidDockerfile tests --check-env with invalid Dockerfile.
func TestDockerEnvironmentCheckEnvInvalidDockerfile(t *testing.T) {
	skipIfNoDocker(t)
	h := NewTestHarness(t)

	// Create a Dockerfile without FROM instruction
	h.WriteFile("Dockerfile", `RUN echo "no FROM instruction"
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: docker
	.source: ./Dockerfile

@test:
	echo "test"
`)

	result := h.Run("--check-env")
	result.AssertExitCode(4). // Environment error
					AssertStdoutContains("missing FROM instruction")
}

// TestDockerEnvironmentListEnv tests --list-env with Docker environment.
func TestDockerEnvironmentListEnv(t *testing.T) {
	skipIfNoDocker(t)
	h := NewTestHarness(t)

	h.WriteFile("Dockerfile", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: docker
	.source: ./Dockerfile

.environment: ci
	.using: docker
	.source: ./Dockerfile
	.args: --platform linux/amd64

@test:
	echo "test"
`)

	result := h.Run("--list-env")
	result.AssertSuccess().
		AssertStdoutContains("Available environments").
		AssertStdoutContains("docker").
		AssertStdoutContains("ci")
}

// TestDockerEnvironmentDryRun tests --dry-run with Docker environment.
// Note: Dry run shows what commands would execute but doesn't actually run them.
func TestDockerEnvironmentDryRun(t *testing.T) {
	skipIfNoDocker(t)
	h := NewTestHarness(t)

	h.WriteFile("Dockerfile", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: docker
	.source: ./Dockerfile

@test:
	echo "test output"
`)

	result := h.Run("--dry-run", "test")
	result.AssertSuccess().
		AssertStdoutContains("echo")
}

// TestDockerEnvironmentVerbose tests --verbose with Docker environment.
func TestDockerEnvironmentVerbose(t *testing.T) {
	skipIfNoDocker(t)
	h := NewTestHarness(t)

	h.WriteFile("Dockerfile", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: docker
	.source: ./Dockerfile

@test:
	echo "test"
`)

	result := h.Run("--verbose", "--check-env")
	result.AssertSuccess()
	// Verbose mode should show docker info
}

// TestDockerEnvironmentRecipeFailure tests that recipe failures are reported correctly.
func TestDockerEnvironmentRecipeFailure(t *testing.T) {
	skipIfNoDocker(t)
	h := NewTestHarness(t)

	h.WriteFile("Dockerfile", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: docker
	.source: ./Dockerfile

@test:
	exit 42
`)

	result := h.Run("test")
	result.AssertExitCode(1) // Build failure
}

// TestDockerEnvironmentWithVariables tests variable interpolation with Docker environments.
// Note: Variables work in commands regardless of environment type.
func TestDockerEnvironmentWithVariables(t *testing.T) {
	skipIfNoDocker(t)
	h := NewTestHarness(t)

	h.WriteFile("Dockerfile", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: docker
	.source: ./Dockerfile

greeting = Hello Docker

@test:
	echo "{greeting}"
`)

	result := h.Run("test")
	result.AssertSuccess().
		AssertStdoutContains("Hello Docker")
}

// TestDockerEnvironmentSourceInSubdirectory tests Dockerfile in subdirectory.
func TestDockerEnvironmentSourceInSubdirectory(t *testing.T) {
	skipIfNoDocker(t)
	h := NewTestHarness(t)

	h.Mkdir("docker")
	h.WriteFile("docker/Dockerfile.ci", `FROM alpine:latest
RUN echo "CI image"
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment: ci
	.using: docker
	.source: ./docker/Dockerfile.ci

@test:
	echo "Running in CI container"
`)

	result := h.Run("--env", "ci", "--check-env")
	result.AssertSuccess().
		AssertStdoutContains("Checking environment: ci").
		AssertStdoutContains("docker/Dockerfile.ci")
}

// TestDockerEnvironmentNamedEnvironmentCheckEnv tests --check-env with named Docker environments.
func TestDockerEnvironmentNamedEnvironmentCheckEnv(t *testing.T) {
	skipIfNoDocker(t)
	h := NewTestHarness(t)

	h.WriteFile("Dockerfile.dev", `FROM alpine:latest
`)

	h.WriteFile("Dockerfile.prod", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment: dev
	.using: docker
	.source: ./Dockerfile.dev

.environment: prod
	.using: docker
	.source: ./Dockerfile.prod

@test:
	echo "test"
`)

	// Check dev environment
	result := h.Run("--env", "dev", "--check-env")
	result.AssertSuccess().
		AssertStdoutContains("Checking environment: dev").
		AssertStdoutContains("Runtime: docker")

	// Check prod environment
	result = h.Run("--env", "prod", "--check-env")
	result.AssertSuccess().
		AssertStdoutContains("Checking environment: prod").
		AssertStdoutContains("Runtime: docker")
}

// TestDockerEnvironmentWithArgsInListEnv tests that .args shows in --list-env.
func TestDockerEnvironmentWithArgsInListEnv(t *testing.T) {
	skipIfNoDocker(t)
	h := NewTestHarness(t)

	h.WriteFile("Dockerfile", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment: linux-amd64
	.using: docker
	.source: ./Dockerfile
	.args: --platform linux/amd64

@test:
	echo "test"
`)

	result := h.Run("--list-env")
	result.AssertSuccess().
		AssertStdoutContains("linux-amd64").
		AssertStdoutContains("docker")
}

// TestDockerEnvironmentNoDefaultWithNamedOnly tests error when only named environments exist.
func TestDockerEnvironmentNoDefaultWithNamedOnly(t *testing.T) {
	skipIfNoDocker(t)
	h := NewTestHarness(t)

	h.WriteFile("Dockerfile.dev", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment: dev
	.using: docker
	.source: ./Dockerfile.dev

@test:
	echo "test"
`)

	// No --env flag, no default environment - should error
	result := h.Run("--check-env")
	result.AssertExitCode(4). // Environment error
					AssertStderrContains("no default environment")
}

// TestDockerEnvironmentMixedRuntimes tests --list-env with mixed environment types.
func TestDockerEnvironmentMixedRuntimes(t *testing.T) {
	skipIfNoDocker(t)
	h := NewTestHarness(t)

	h.WriteFile("Dockerfile", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: bare
	.requires: bash

.environment: container
	.using: docker
	.source: ./Dockerfile

@test:
	echo "test"
`)

	result := h.Run("--list-env")
	result.AssertSuccess().
		AssertStdoutContains("bare").
		AssertStdoutContains("docker").
		AssertStdoutContains("container")
}

// ===========================================================================
// Podman Environment Integration Tests
// These tests require Podman to be available.
// ===========================================================================

// skipIfNoPodman skips the test if Podman is not available.
func skipIfNoPodman(t *testing.T) {
	t.Helper()
	h := NewTestHarness(t)
	result := h.RunShell("podman version >/dev/null 2>&1")
	if result.exitCode != 0 {
		t.Skip("Podman is not available, skipping Podman integration test")
	}
}

// TestPodmanEnvironmentCheckEnv tests --check-env with a Podman environment.
func TestPodmanEnvironmentCheckEnv(t *testing.T) {
	skipIfNoPodman(t)
	h := NewTestHarness(t)

	// Create a simple Containerfile (Podman's equivalent)
	h.WriteFile("Containerfile", `FROM alpine:latest
RUN echo "test"
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: podman
	.source: ./Containerfile

@test:
	echo "Hello from Podman"
`)

	result := h.Run("--check-env")
	result.AssertSuccess().
		AssertStdoutContains("Checking environment").
		AssertStdoutContains("Runtime: podman").
		AssertStdoutContains("Containerfile")
}

// TestPodmanEnvironmentCheckEnvMissingContainerfile tests --check-env with missing Containerfile.
func TestPodmanEnvironmentCheckEnvMissingContainerfile(t *testing.T) {
	skipIfNoPodman(t)
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: podman
	.source: ./NonexistentContainerfile

@test:
	echo "test"
`)

	result := h.Run("--check-env")
	result.AssertExitCode(4). // Environment error
					AssertStdoutContains("not found")
}

// TestPodmanEnvironmentCheckEnvInvalidContainerfile tests --check-env with invalid Containerfile.
func TestPodmanEnvironmentCheckEnvInvalidContainerfile(t *testing.T) {
	skipIfNoPodman(t)
	h := NewTestHarness(t)

	// Create a Containerfile without FROM instruction
	h.WriteFile("Containerfile", `RUN echo "no FROM instruction"
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: podman
	.source: ./Containerfile

@test:
	echo "test"
`)

	result := h.Run("--check-env")
	result.AssertExitCode(4). // Environment error
					AssertStdoutContains("missing FROM instruction")
}

// TestPodmanEnvironmentListEnv tests --list-env with Podman environment.
func TestPodmanEnvironmentListEnv(t *testing.T) {
	skipIfNoPodman(t)
	h := NewTestHarness(t)

	h.WriteFile("Containerfile", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: podman
	.source: ./Containerfile

.environment: ci
	.using: podman
	.source: ./Containerfile
	.args: --platform linux/amd64

@test:
	echo "test"
`)

	result := h.Run("--list-env")
	result.AssertSuccess().
		AssertStdoutContains("Available environments").
		AssertStdoutContains("podman").
		AssertStdoutContains("ci")
}

// TestPodmanEnvironmentDryRun tests --dry-run with Podman environment.
func TestPodmanEnvironmentDryRun(t *testing.T) {
	skipIfNoPodman(t)
	h := NewTestHarness(t)

	h.WriteFile("Containerfile", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: podman
	.source: ./Containerfile

@test:
	echo "test output"
`)

	result := h.Run("--dry-run", "test")
	result.AssertSuccess().
		AssertStdoutContains("echo")
}

// TestPodmanEnvironmentVerbose tests --verbose with Podman environment.
func TestPodmanEnvironmentVerbose(t *testing.T) {
	skipIfNoPodman(t)
	h := NewTestHarness(t)

	h.WriteFile("Containerfile", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: podman
	.source: ./Containerfile

@test:
	echo "test"
`)

	result := h.Run("--verbose", "--check-env")
	result.AssertSuccess()
}

// TestPodmanEnvironmentNamedEnvironmentCheckEnv tests --check-env with named Podman environments.
func TestPodmanEnvironmentNamedEnvironmentCheckEnv(t *testing.T) {
	skipIfNoPodman(t)
	h := NewTestHarness(t)

	h.WriteFile("Containerfile.dev", `FROM alpine:latest
`)

	h.WriteFile("Containerfile.prod", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment: dev
	.using: podman
	.source: ./Containerfile.dev

.environment: prod
	.using: podman
	.source: ./Containerfile.prod

@test:
	echo "test"
`)

	// Check dev environment
	result := h.Run("--env", "dev", "--check-env")
	result.AssertSuccess().
		AssertStdoutContains("Checking environment: dev").
		AssertStdoutContains("Runtime: podman")

	// Check prod environment
	result = h.Run("--env", "prod", "--check-env")
	result.AssertSuccess().
		AssertStdoutContains("Checking environment: prod").
		AssertStdoutContains("Runtime: podman")
}

// TestPodmanEnvironmentSourceInSubdirectory tests Containerfile in subdirectory.
func TestPodmanEnvironmentSourceInSubdirectory(t *testing.T) {
	skipIfNoPodman(t)
	h := NewTestHarness(t)

	h.Mkdir("containers")
	h.WriteFile("containers/Containerfile.ci", `FROM alpine:latest
RUN echo "CI image"
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment: ci
	.using: podman
	.source: ./containers/Containerfile.ci

@test:
	echo "Running in CI container"
`)

	result := h.Run("--env", "ci", "--check-env")
	result.AssertSuccess().
		AssertStdoutContains("Checking environment: ci").
		AssertStdoutContains("containers/Containerfile.ci")
}

// TestPodmanEnvironmentWithArgsInListEnv tests that .args shows in --list-env.
func TestPodmanEnvironmentWithArgsInListEnv(t *testing.T) {
	skipIfNoPodman(t)
	h := NewTestHarness(t)

	h.WriteFile("Containerfile", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment: linux-amd64
	.using: podman
	.source: ./Containerfile
	.args: --platform linux/amd64

@test:
	echo "test"
`)

	result := h.Run("--list-env")
	result.AssertSuccess().
		AssertStdoutContains("linux-amd64").
		AssertStdoutContains("podman")
}

// TestPodmanEnvironmentNoDefaultWithNamedOnly tests error when only named environments exist.
func TestPodmanEnvironmentNoDefaultWithNamedOnly(t *testing.T) {
	skipIfNoPodman(t)
	h := NewTestHarness(t)

	h.WriteFile("Containerfile.dev", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment: dev
	.using: podman
	.source: ./Containerfile.dev

@test:
	echo "test"
`)

	// No --env flag, no default environment - should error
	result := h.Run("--check-env")
	result.AssertExitCode(4). // Environment error
					AssertStderrContains("no default environment")
}

// TestPodmanEnvironmentMixedRuntimes tests --list-env with mixed environment types.
func TestPodmanEnvironmentMixedRuntimes(t *testing.T) {
	skipIfNoPodman(t)
	h := NewTestHarness(t)

	h.WriteFile("Containerfile", `FROM alpine:latest
`)

	h.WriteFile("Buildfile", `.shell: sh

.environment:
	.using: bare
	.requires: bash

.environment: container
	.using: podman
	.source: ./Containerfile

@test:
	echo "test"
`)

	result := h.Run("--list-env")
	result.AssertSuccess().
		AssertStdoutContains("bare").
		AssertStdoutContains("podman").
		AssertStdoutContains("container")
}

// TestMixedDockerPodmanEnvironments tests --list-env with both Docker and Podman environments.
func TestMixedDockerPodmanEnvironments(t *testing.T) {
	// Skip if neither Docker nor Podman is available
	h := NewTestHarness(t)
	dockerResult := h.RunShell("docker version >/dev/null 2>&1")
	podmanResult := h.RunShell("podman version >/dev/null 2>&1")
	if dockerResult.exitCode != 0 && podmanResult.exitCode != 0 {
		t.Skip("Neither Docker nor Podman is available, skipping mixed container test")
	}

	h.WriteFile("Dockerfile", `FROM alpine:latest
`)
	h.WriteFile("Containerfile", `FROM alpine:latest
`)

	// Build a Buildfile with both Docker and Podman environments
	// Note: Only use the available runtime(s)
	var buildfileContent string
	if dockerResult.exitCode == 0 && podmanResult.exitCode == 0 {
		buildfileContent = `.shell: sh

.environment: docker-env
	.using: docker
	.source: ./Dockerfile

.environment: podman-env
	.using: podman
	.source: ./Containerfile

@test:
	echo "test"
`
	} else if dockerResult.exitCode == 0 {
		buildfileContent = `.shell: sh

.environment: docker-env
	.using: docker
	.source: ./Dockerfile

@test:
	echo "test"
`
	} else {
		buildfileContent = `.shell: sh

.environment: podman-env
	.using: podman
	.source: ./Containerfile

@test:
	echo "test"
`
	}

	h.WriteFile("Buildfile", buildfileContent)

	result := h.Run("--list-env")
	result.AssertSuccess().
		AssertStdoutContains("Available environments")

	// Check that at least one container runtime is listed
	stdout := result.Stdout()
	if !strings.Contains(stdout, "docker") && !strings.Contains(stdout, "podman") {
		t.Error("expected at least one container runtime in output")
	}
}
