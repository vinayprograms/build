package environ

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDevcontainerDetector_DetectConfig_Directory(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "devcontainer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .devcontainer directory with devcontainer.json
	devcontainerDir := filepath.Join(tmpDir, ".devcontainer")
	if err := os.MkdirAll(devcontainerDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(devcontainerDir, "devcontainer.json")
	config := `{
		"name": "Test Container",
		"image": "ubuntu:latest"
	}`
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	// Test detection
	detector := NewDevcontainerDetector()
	result, err := detector.DetectConfig(tmpDir)

	if err != nil {
		t.Fatalf("DetectConfig() error = %v", err)
	}
	if !result.Found {
		t.Error("expected config to be found")
	}
	if result.Path != configPath {
		t.Errorf("Path = %q, want %q", result.Path, configPath)
	}
}

func TestDevcontainerDetector_DetectConfig_RootJson(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "devcontainer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create devcontainer.json in root
	configPath := filepath.Join(tmpDir, "devcontainer.json")
	config := `{
		"name": "Test Container",
		"dockerFile": "Dockerfile"
	}`
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	// Test detection
	detector := NewDevcontainerDetector()
	result, err := detector.DetectConfig(tmpDir)

	if err != nil {
		t.Fatalf("DetectConfig() error = %v", err)
	}
	if !result.Found {
		t.Error("expected config to be found")
	}
	if result.Path != configPath {
		t.Errorf("Path = %q, want %q", result.Path, configPath)
	}
}

func TestDevcontainerDetector_DetectConfig_NotFound(t *testing.T) {
	// Create empty temp directory
	tmpDir, err := os.MkdirTemp("", "devcontainer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test detection
	detector := NewDevcontainerDetector()
	result, err := detector.DetectConfig(tmpDir)

	if err != nil {
		t.Fatalf("DetectConfig() error = %v", err)
	}
	if result.Found {
		t.Error("expected config to not be found")
	}
}

func TestDevcontainerDetector_DetectConfig_DirectoryPriority(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "devcontainer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create both .devcontainer/devcontainer.json and root devcontainer.json
	devcontainerDir := filepath.Join(tmpDir, ".devcontainer")
	if err := os.MkdirAll(devcontainerDir, 0755); err != nil {
		t.Fatal(err)
	}

	dirConfigPath := filepath.Join(devcontainerDir, "devcontainer.json")
	if err := os.WriteFile(dirConfigPath, []byte(`{"name": "dir"}`), 0644); err != nil {
		t.Fatal(err)
	}

	rootConfigPath := filepath.Join(tmpDir, "devcontainer.json")
	if err := os.WriteFile(rootConfigPath, []byte(`{"name": "root"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Test detection - .devcontainer/ should take priority
	detector := NewDevcontainerDetector()
	result, err := detector.DetectConfig(tmpDir)

	if err != nil {
		t.Fatalf("DetectConfig() error = %v", err)
	}
	if result.Path != dirConfigPath {
		t.Errorf("Path = %q, want %q (directory should take priority)", result.Path, dirConfigPath)
	}
}

func TestParseDevcontainerConfig(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    *DevcontainerConfig
		wantErr bool
	}{
		{
			name: "with image",
			json: `{"name": "Test", "image": "ubuntu:latest"}`,
			want: &DevcontainerConfig{
				Name:  "Test",
				Image: "ubuntu:latest",
			},
		},
		{
			name: "with dockerfile",
			json: `{"name": "Test", "dockerFile": "Dockerfile"}`,
			want: &DevcontainerConfig{
				Name:       "Test",
				Dockerfile: "Dockerfile",
			},
		},
		{
			name: "with docker-compose",
			json: `{"name": "Test", "dockerComposeFile": "docker-compose.yml", "service": "app"}`,
			want: &DevcontainerConfig{
				Name:              "Test",
				DockerComposeFile: "docker-compose.yml",
				Service:           "app",
			},
		},
		{
			name: "with build context",
			json: `{"name": "Test", "build": {"dockerfile": "Dockerfile", "context": "."}}`,
			want: &DevcontainerConfig{
				Name: "Test",
				Build: &DevcontainerBuildConfig{
					Dockerfile: "Dockerfile",
					Context:    ".",
				},
			},
		},
		{
			name:    "invalid json",
			json:    `{not valid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDevcontainerConfig([]byte(tt.json))
			if tt.wantErr {
				if err == nil {
					t.Error("ParseDevcontainerConfig() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseDevcontainerConfig() error = %v", err)
			}

			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Image != tt.want.Image {
				t.Errorf("Image = %q, want %q", got.Image, tt.want.Image)
			}
			if got.Dockerfile != tt.want.Dockerfile {
				t.Errorf("Dockerfile = %q, want %q", got.Dockerfile, tt.want.Dockerfile)
			}
			if got.DockerComposeFile != tt.want.DockerComposeFile {
				t.Errorf("DockerComposeFile = %q, want %q", got.DockerComposeFile, tt.want.DockerComposeFile)
			}
			if got.Service != tt.want.Service {
				t.Errorf("Service = %q, want %q", got.Service, tt.want.Service)
			}
			if tt.want.Build != nil {
				if got.Build == nil {
					t.Error("Build = nil, want non-nil")
				} else {
					if got.Build.Dockerfile != tt.want.Build.Dockerfile {
						t.Errorf("Build.Dockerfile = %q, want %q", got.Build.Dockerfile, tt.want.Build.Dockerfile)
					}
					if got.Build.Context != tt.want.Build.Context {
						t.Errorf("Build.Context = %q, want %q", got.Build.Context, tt.want.Build.Context)
					}
				}
			}
		})
	}
}

func TestDevcontainerDetector_LoadConfig(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "devcontainer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .devcontainer directory with devcontainer.json
	devcontainerDir := filepath.Join(tmpDir, ".devcontainer")
	if err := os.MkdirAll(devcontainerDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(devcontainerDir, "devcontainer.json")
	config := `{
		"name": "Test Container",
		"image": "mcr.microsoft.com/devcontainers/base:ubuntu"
	}`
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	// Test loading
	detector := NewDevcontainerDetector()
	result, err := detector.DetectConfig(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := detector.LoadConfig(result.Path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Name != "Test Container" {
		t.Errorf("Name = %q, want %q", cfg.Name, "Test Container")
	}
	if cfg.Image != "mcr.microsoft.com/devcontainers/base:ubuntu" {
		t.Errorf("Image = %q, want %q", cfg.Image, "mcr.microsoft.com/devcontainers/base:ubuntu")
	}
}

// ----------------------------------------------------------------------------
// Devcontainer Runner Tests
// ----------------------------------------------------------------------------

func TestNewDevcontainerRunner(t *testing.T) {
	runner := NewDevcontainerRunner("/project")

	if runner == nil {
		t.Fatal("NewDevcontainerRunner returned nil")
	}
	if runner.projectDir != "/project" {
		t.Errorf("projectDir = %q, want %q", runner.projectDir, "/project")
	}
}

func TestDevcontainerRunner_CheckCLI_NotInstalled(t *testing.T) {
	runner := NewDevcontainerRunner("/project")
	// Override lookPath to simulate devcontainer CLI not being installed
	runner.lookPath = func(name string) (string, error) {
		return "", &BinaryNotFoundError{Name: name}
	}

	err := runner.CheckCLI()
	if err == nil {
		t.Error("expected error for missing devcontainer CLI")
	}
}

func TestDevcontainerConfig_GetImageOrBuildSource(t *testing.T) {
	tests := []struct {
		name   string
		config DevcontainerConfig
		want   string
	}{
		{
			name:   "with image",
			config: DevcontainerConfig{Image: "ubuntu:latest"},
			want:   "image:ubuntu:latest",
		},
		{
			name:   "with dockerfile",
			config: DevcontainerConfig{Dockerfile: "Dockerfile"},
			want:   "dockerfile:Dockerfile",
		},
		{
			name: "with build dockerfile",
			config: DevcontainerConfig{
				Build: &DevcontainerBuildConfig{Dockerfile: "Dockerfile.dev"},
			},
			want: "dockerfile:Dockerfile.dev",
		},
		{
			name:   "with docker-compose",
			config: DevcontainerConfig{DockerComposeFile: "docker-compose.yml", Service: "app"},
			want:   "compose:docker-compose.yml:app",
		},
		{
			name:   "empty config",
			config: DevcontainerConfig{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetImageOrBuildSource()
			if got != tt.want {
				t.Errorf("GetImageOrBuildSource() = %q, want %q", got, tt.want)
			}
		})
	}
}
