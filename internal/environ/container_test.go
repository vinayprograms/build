package environ

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

// TestDockerfileDetection tests locating Dockerfile from .source: directive
func TestDockerfileDetection(t *testing.T) {
	// Create a temp directory with a Dockerfile
	tmpDir := t.TempDir()

	t.Run("finds Dockerfile at specified path", func(t *testing.T) {
		dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
		if err := os.WriteFile(dockerfilePath, []byte("FROM alpine\n"), 0644); err != nil {
			t.Fatalf("failed to create Dockerfile: %v", err)
		}

		env := &ast.Environment{
			Runtime: ptrRuntime(ast.RuntimeDocker),
			Source:  &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: dockerfilePath}}},
		}

		detector := NewContainerDetector()
		result, err := detector.DetectDockerfile(env, tmpDir)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result.Path != dockerfilePath {
			t.Errorf("expected path %s, got %s", dockerfilePath, result.Path)
		}
		if !result.Exists {
			t.Error("expected Exists=true")
		}
	})

	t.Run("returns error for missing Dockerfile", func(t *testing.T) {
		env := &ast.Environment{
			Runtime: ptrRuntime(ast.RuntimeDocker),
			Source:  &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "/nonexistent/Dockerfile"}}},
		}

		detector := NewContainerDetector()
		result, err := detector.DetectDockerfile(env, tmpDir)
		if err == nil {
			t.Error("expected error for missing Dockerfile")
		}
		if result.Exists {
			t.Error("expected Exists=false")
		}
	})

	t.Run("resolves relative path from base dir", func(t *testing.T) {
		subDir := filepath.Join(tmpDir, "docker")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("failed to create subdir: %v", err)
		}
		dockerfilePath := filepath.Join(subDir, "ci.Dockerfile")
		if err := os.WriteFile(dockerfilePath, []byte("FROM ubuntu\n"), 0644); err != nil {
			t.Fatalf("failed to create Dockerfile: %v", err)
		}

		env := &ast.Environment{
			Runtime: ptrRuntime(ast.RuntimeDocker),
			Source:  &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "./docker/ci.Dockerfile"}}},
		}

		detector := NewContainerDetector()
		result, err := detector.DetectDockerfile(env, tmpDir)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result.Path != dockerfilePath {
			t.Errorf("expected path %s, got %s", dockerfilePath, result.Path)
		}
	})

	t.Run("supports Podman runtime", func(t *testing.T) {
		containerfilePath := filepath.Join(tmpDir, "Containerfile")
		if err := os.WriteFile(containerfilePath, []byte("FROM alpine\n"), 0644); err != nil {
			t.Fatalf("failed to create Containerfile: %v", err)
		}

		env := &ast.Environment{
			Runtime: ptrRuntime(ast.RuntimePodman),
			Source:  &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: containerfilePath}}},
		}

		detector := NewContainerDetector()
		result, err := detector.DetectDockerfile(env, tmpDir)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result.Path != containerfilePath {
			t.Errorf("expected path %s, got %s", containerfilePath, result.Path)
		}
	})

	t.Run("returns error for nil source", func(t *testing.T) {
		env := &ast.Environment{
			Runtime: ptrRuntime(ast.RuntimeDocker),
			Source:  nil,
		}

		detector := NewContainerDetector()
		_, err := detector.DetectDockerfile(env, tmpDir)
		if err == nil {
			t.Error("expected error for nil source")
		}
		if _, ok := err.(*NoSourceError); !ok {
			t.Errorf("expected NoSourceError, got %T", err)
		}
	})

	t.Run("returns error for non-container runtime", func(t *testing.T) {
		env := &ast.Environment{
			Runtime: ptrRuntime(ast.RuntimeBare),
			Source:  &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "Dockerfile"}}},
		}

		detector := NewContainerDetector()
		_, err := detector.DetectDockerfile(env, tmpDir)
		if err == nil {
			t.Error("expected error for bare runtime")
		}
		if _, ok := err.(*InvalidRuntimeError); !ok {
			t.Errorf("expected InvalidRuntimeError, got %T", err)
		}
	})
}

// TestDockerfileValidation tests validating Dockerfile content
func TestDockerfileValidation(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("validates Dockerfile has FROM instruction", func(t *testing.T) {
		dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
		if err := os.WriteFile(dockerfilePath, []byte("FROM alpine\nRUN echo hello\n"), 0644); err != nil {
			t.Fatalf("failed to create Dockerfile: %v", err)
		}

		detector := NewContainerDetector()
		err := detector.ValidateDockerfile(dockerfilePath)
		if err != nil {
			t.Errorf("expected valid Dockerfile, got error: %v", err)
		}
	})

	t.Run("rejects Dockerfile without FROM", func(t *testing.T) {
		dockerfilePath := filepath.Join(tmpDir, "BadDockerfile")
		if err := os.WriteFile(dockerfilePath, []byte("RUN echo hello\n"), 0644); err != nil {
			t.Fatalf("failed to create Dockerfile: %v", err)
		}

		detector := NewContainerDetector()
		err := detector.ValidateDockerfile(dockerfilePath)
		if err == nil {
			t.Error("expected error for Dockerfile without FROM")
		}
		if _, ok := err.(*InvalidDockerfileError); !ok {
			t.Errorf("expected InvalidDockerfileError, got %T", err)
		}
	})

	t.Run("handles empty Dockerfile", func(t *testing.T) {
		dockerfilePath := filepath.Join(tmpDir, "EmptyDockerfile")
		if err := os.WriteFile(dockerfilePath, []byte(""), 0644); err != nil {
			t.Fatalf("failed to create Dockerfile: %v", err)
		}

		detector := NewContainerDetector()
		err := detector.ValidateDockerfile(dockerfilePath)
		if err == nil {
			t.Error("expected error for empty Dockerfile")
		}
	})

	t.Run("handles FROM with arguments", func(t *testing.T) {
		dockerfilePath := filepath.Join(tmpDir, "ArgDockerfile")
		content := "ARG BASE=alpine\nFROM ${BASE}\n"
		if err := os.WriteFile(dockerfilePath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create Dockerfile: %v", err)
		}

		detector := NewContainerDetector()
		err := detector.ValidateDockerfile(dockerfilePath)
		if err != nil {
			t.Errorf("expected valid Dockerfile with ARG before FROM, got error: %v", err)
		}
	})
}

// TestContainerRuntimeDetection tests detecting container runtime (docker/podman)
func TestContainerRuntimeDetection(t *testing.T) {
	t.Run("returns docker runtime path when available", func(t *testing.T) {
		detector := &ContainerDetector{
			lookPath: func(name string) (string, error) {
				if name == "docker" {
					return "/usr/bin/docker", nil
				}
				return "", &BinaryNotFoundError{Name: name}
			},
		}

		path, err := detector.FindRuntime(ast.RuntimeDocker)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if path != "/usr/bin/docker" {
			t.Errorf("expected /usr/bin/docker, got %s", path)
		}
	})

	t.Run("returns podman runtime path when available", func(t *testing.T) {
		detector := &ContainerDetector{
			lookPath: func(name string) (string, error) {
				if name == "podman" {
					return "/usr/bin/podman", nil
				}
				return "", &BinaryNotFoundError{Name: name}
			},
		}

		path, err := detector.FindRuntime(ast.RuntimePodman)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if path != "/usr/bin/podman" {
			t.Errorf("expected /usr/bin/podman, got %s", path)
		}
	})

	t.Run("returns error when runtime not found", func(t *testing.T) {
		detector := &ContainerDetector{
			lookPath: func(name string) (string, error) {
				return "", &BinaryNotFoundError{Name: name}
			},
		}

		_, err := detector.FindRuntime(ast.RuntimeDocker)
		if err == nil {
			t.Error("expected error when docker not found")
		}
		if _, ok := err.(*BinaryNotFoundError); !ok {
			t.Errorf("expected BinaryNotFoundError, got %T", err)
		}
	})
}

// TestImageBuilder tests building container images
func TestImageBuilder(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test Dockerfile
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte("FROM alpine\n"), 0644); err != nil {
		t.Fatalf("failed to create Dockerfile: %v", err)
	}

	t.Run("generates correct docker build command", func(t *testing.T) {
		builder := &ImageBuilder{
			runtime:    ast.RuntimeDocker,
			runtimeCmd: "docker",
		}

		cmd := builder.BuildCommand(dockerfilePath, "test-image:latest")
		args := cmd.Args

		// Should have: docker build -t test-image:latest -f /path/to/Dockerfile /path/to
		if args[0] != "docker" {
			t.Errorf("expected docker, got %s", args[0])
		}
		if args[1] != "build" {
			t.Errorf("expected build, got %s", args[1])
		}

		// Find -t flag
		found := false
		for i, arg := range args {
			if arg == "-t" && i+1 < len(args) && args[i+1] == "test-image:latest" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected -t test-image:latest in args: %v", args)
		}
	})

	t.Run("generates correct podman build command", func(t *testing.T) {
		builder := &ImageBuilder{
			runtime:    ast.RuntimePodman,
			runtimeCmd: "podman",
		}

		cmd := builder.BuildCommand(dockerfilePath, "test-image:latest")
		if cmd.Args[0] != "podman" {
			t.Errorf("expected podman, got %s", cmd.Args[0])
		}
	})

	t.Run("generates image tag from project and environment", func(t *testing.T) {
		tag := GenerateImageTag("myproject", "ci")
		if tag != "myproject-ci:latest" {
			t.Errorf("expected myproject-ci:latest, got %s", tag)
		}
	})

	t.Run("generates image tag for default environment", func(t *testing.T) {
		tag := GenerateImageTag("myproject", "")
		if tag != "myproject:latest" {
			t.Errorf("expected myproject:latest, got %s", tag)
		}
	})

	t.Run("includes extra args in build command", func(t *testing.T) {
		builder := &ImageBuilder{
			runtime:    ast.RuntimeDocker,
			runtimeCmd: "docker",
			extraArgs:  []string{"--platform", "linux/amd64"},
		}

		cmd := builder.BuildCommand(dockerfilePath, "test-image:latest")
		args := cmd.Args

		// Find --platform flag
		found := false
		for i, arg := range args {
			if arg == "--platform" && i+1 < len(args) && args[i+1] == "linux/amd64" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected --platform linux/amd64 in args: %v", args)
		}
	})
}

// TestContainerRunner tests running containers
func TestContainerRunner(t *testing.T) {
	t.Run("generates correct run command with workspace mount", func(t *testing.T) {
		runner := NewContainerRunner(ast.RuntimeDocker, "docker", "/project/workspace")

		cmd := runner.RunCommand("myimage:latest", []string{"make", "build"})
		args := cmd.Args

		if args[0] != "docker" {
			t.Errorf("expected docker, got %s", args[0])
		}
		if args[1] != "run" {
			t.Errorf("expected run, got %s", args[1])
		}

		// Should have --rm flag
		foundRm := false
		for _, arg := range args {
			if arg == "--rm" {
				foundRm = true
				break
			}
		}
		if !foundRm {
			t.Errorf("expected --rm flag in args: %v", args)
		}

		// Should have -v mount
		foundMount := false
		for i, arg := range args {
			if arg == "-v" && i+1 < len(args) {
				if strings.Contains(args[i+1], "/project/workspace:") {
					foundMount = true
					break
				}
			}
		}
		if !foundMount {
			t.Errorf("expected workspace mount in args: %v", args)
		}

		// Should have -w workdir
		foundWorkdir := false
		for i, arg := range args {
			if arg == "-w" && i+1 < len(args) {
				foundWorkdir = true
				break
			}
		}
		if !foundWorkdir {
			t.Errorf("expected -w workdir in args: %v", args)
		}

		// Should end with image and command
		if args[len(args)-3] != "myimage:latest" {
			t.Errorf("expected image before command, got args: %v", args)
		}
		if args[len(args)-2] != "make" || args[len(args)-1] != "build" {
			t.Errorf("expected command at end, got args: %v", args)
		}
	})

	t.Run("generates shell command for interactive mode", func(t *testing.T) {
		runner := NewContainerRunner(ast.RuntimeDocker, "docker", "/project/workspace")

		cmd := runner.ShellCommand("myimage:latest", "/bin/bash")
		args := cmd.Args

		// Should have -it flags
		foundIt := false
		for _, arg := range args {
			if arg == "-it" {
				foundIt = true
				break
			}
		}
		if !foundIt {
			t.Errorf("expected -it flag in args: %v", args)
		}

		// Should end with shell
		if args[len(args)-1] != "/bin/bash" {
			t.Errorf("expected /bin/bash at end, got %s", args[len(args)-1])
		}
	})

	t.Run("generates command with extra args", func(t *testing.T) {
		runner := NewContainerRunner(ast.RuntimeDocker, "docker", "/project/workspace")
		runner.SetExtraArgs([]string{"--platform", "linux/amd64"})

		cmd := runner.RunCommand("myimage:latest", []string{"make"})
		args := cmd.Args

		foundPlatform := false
		for i, arg := range args {
			if arg == "--platform" && i+1 < len(args) && args[i+1] == "linux/amd64" {
				foundPlatform = true
				break
			}
		}
		if !foundPlatform {
			t.Errorf("expected --platform linux/amd64 in args: %v", args)
		}
	})

	t.Run("generates keep-alive command without --rm", func(t *testing.T) {
		runner := NewContainerRunner(ast.RuntimeDocker, "docker", "/project/workspace")

		cmd := runner.RunCommandKeepAlive("myimage:latest", []string{"make"}, "build-container")
		args := cmd.Args

		// Should NOT have --rm flag
		for _, arg := range args {
			if arg == "--rm" {
				t.Errorf("should not have --rm flag for keep-alive: %v", args)
			}
		}

		// Should have --name flag
		foundName := false
		for i, arg := range args {
			if arg == "--name" && i+1 < len(args) && args[i+1] == "build-container" {
				foundName = true
				break
			}
		}
		if !foundName {
			t.Errorf("expected --name build-container in args: %v", args)
		}
	})
}

// TestImageExists tests checking if an image exists
func TestImageExists(t *testing.T) {
	t.Run("returns true when image exists", func(t *testing.T) {
		execCalled := false
		builder := &ImageBuilder{
			runtime:    ast.RuntimeDocker,
			runtimeCmd: "docker",
			runCommand: func(name string, args ...string) ([]byte, error) {
				execCalled = true
				if name == "docker" && len(args) >= 2 && args[0] == "image" && args[1] == "inspect" {
					return []byte("[{}]"), nil
				}
				return nil, &BinaryNotFoundError{Name: "docker"}
			},
		}

		exists, err := builder.ImageExists("myimage:latest")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Error("expected image to exist")
		}
		if !execCalled {
			t.Error("expected exec to be called")
		}
	})

	t.Run("returns false when image not found", func(t *testing.T) {
		builder := &ImageBuilder{
			runtime:    ast.RuntimeDocker,
			runtimeCmd: "docker",
			runCommand: func(name string, args ...string) ([]byte, error) {
				return nil, &ImageNotFoundError{Name: "myimage:latest"}
			},
		}

		exists, err := builder.ImageExists("myimage:latest")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exists {
			t.Error("expected image to not exist")
		}
	})
}

// TestContainerEnvironment tests the high-level container environment
func TestContainerEnvironment(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test Dockerfile
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte("FROM alpine\n"), 0644); err != nil {
		t.Fatalf("failed to create Dockerfile: %v", err)
	}

	t.Run("returns error for non-container runtime", func(t *testing.T) {
		env := &ast.Environment{
			Runtime: ptrRuntime(ast.RuntimeBare),
			Source:  &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: dockerfilePath}}},
		}

		_, err := NewContainerEnvironment(env, tmpDir, "test-project")
		if err == nil {
			t.Error("expected error for bare runtime")
		}
	})

	t.Run("returns error when runtime nil", func(t *testing.T) {
		env := &ast.Environment{
			Runtime: nil,
			Source:  &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: dockerfilePath}}},
		}

		_, err := NewContainerEnvironment(env, tmpDir, "test-project")
		if err == nil {
			t.Error("expected error for nil runtime")
		}
	})
}

// TestPrintKeepInstructions tests the keep-alive instructions
func TestPrintKeepInstructions(t *testing.T) {
	instructions := PrintKeepInstructions("docker", "my-container")

	if !strings.Contains(instructions, "my-container") {
		t.Error("expected container name in instructions")
	}
	if !strings.Contains(instructions, "docker exec") {
		t.Error("expected docker exec command")
	}
	if !strings.Contains(instructions, "docker stop") {
		t.Error("expected docker stop command")
	}
}

// TestGenerateContainerName tests container name generation
func TestGenerateContainerName(t *testing.T) {
	t.Run("with environment name", func(t *testing.T) {
		name := GenerateContainerName("/path/to/project", "ci")
		if name != "build-project-ci" {
			t.Errorf("expected build-project-ci, got %s", name)
		}
	})

	t.Run("without environment name", func(t *testing.T) {
		name := GenerateContainerName("/path/to/project", "")
		if name != "build-project" {
			t.Errorf("expected build-project, got %s", name)
		}
	})
}

// Helper function to create Runtime pointer
func ptrRuntime(r ast.Runtime) *ast.Runtime {
	return &r
}
