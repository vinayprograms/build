package environ

import (
	"runtime"
	"testing"
)

func TestDetectPackageManager(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		goarch   string
		wantNil  bool
		wantName string
	}{
		{
			name:     "linux returns apt or other",
			goos:     "linux",
			goarch:   "amd64",
			wantNil:  false,
			wantName: "", // varies by system
		},
		{
			name:     "darwin returns brew",
			goos:     "darwin",
			goarch:   "arm64",
			wantNil:  false,
			wantName: "brew",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only run for matching OS
			if runtime.GOOS != tt.goos {
				t.Skipf("skipping test for %s on %s", tt.goos, runtime.GOOS)
			}

			pm := DetectPackageManager()
			if tt.wantNil && pm != nil {
				t.Errorf("DetectPackageManager() = %v, want nil", pm)
			}
			if !tt.wantNil && pm == nil {
				t.Errorf("DetectPackageManager() = nil, want non-nil")
			}
			if tt.wantName != "" && pm != nil && pm.Name() != tt.wantName {
				t.Errorf("PackageManager.Name() = %s, want %s", pm.Name(), tt.wantName)
			}
		})
	}
}

func TestPackageManager_GetInstallCommand(t *testing.T) {
	// These tests check the format of install commands, not actual installation
	tests := []struct {
		name    string
		pm      PackageManager
		binary  string
		want    string
		wantErr bool
	}{
		{
			name:   "apt install gcc",
			pm:     &aptPackageManager{},
			binary: "gcc",
			want:   "sudo apt install gcc",
		},
		{
			name:   "brew install gcc",
			pm:     &brewPackageManager{},
			binary: "gcc",
			want:   "brew install gcc",
		},
		{
			name:   "dnf install gcc",
			pm:     &dnfPackageManager{},
			binary: "gcc",
			want:   "sudo dnf install gcc",
		},
		{
			name:   "pacman install gcc",
			pm:     &pacmanPackageManager{},
			binary: "gcc",
			want:   "sudo pacman -S gcc",
		},
		{
			name:   "zypper install gcc",
			pm:     &zypperPackageManager{},
			binary: "gcc",
			want:   "sudo zypper install gcc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pm.GetInstallCommand(tt.binary)
			if got != tt.want {
				t.Errorf("GetInstallCommand(%q) = %q, want %q", tt.binary, got, tt.want)
			}
		})
	}
}

func TestPackageManager_Name(t *testing.T) {
	tests := []struct {
		pm   PackageManager
		want string
	}{
		{&aptPackageManager{}, "apt"},
		{&brewPackageManager{}, "brew"},
		{&dnfPackageManager{}, "dnf"},
		{&pacmanPackageManager{}, "pacman"},
		{&zypperPackageManager{}, "zypper"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.pm.Name(); got != tt.want {
				t.Errorf("Name() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetInstallSuggestion(t *testing.T) {
	tests := []struct {
		name   string
		binary string
		pm     PackageManager
		want   string
	}{
		{
			name:   "gcc with apt",
			binary: "gcc",
			pm:     &aptPackageManager{},
			want:   "sudo apt install gcc",
		},
		{
			name:   "python3 with brew",
			binary: "python3",
			pm:     &brewPackageManager{},
			want:   "brew install python3",
		},
		{
			name:   "nil package manager",
			binary: "gcc",
			pm:     nil,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetInstallSuggestion(tt.binary, tt.pm)
			if got != tt.want {
				t.Errorf("GetInstallSuggestion(%q) = %q, want %q", tt.binary, got, tt.want)
			}
		})
	}
}

func TestBinaryToPackageMapping(t *testing.T) {
	// Test common binary to package name mappings
	tests := []struct {
		binary  string
		pmName  string
		wantPkg string
	}{
		{"gcc", "apt", "gcc"},
		{"g++", "apt", "g++"},
		{"python3", "apt", "python3"},
		{"python", "brew", "python"}, // brew uses just python
		{"make", "apt", "make"},
	}

	for _, tt := range tests {
		t.Run(tt.binary+"_"+tt.pmName, func(t *testing.T) {
			// For now, binary name equals package name
			// Future: implement package name mapping
			got := mapBinaryToPackage(tt.binary, tt.pmName)
			if got != tt.wantPkg {
				t.Errorf("mapBinaryToPackage(%q, %q) = %q, want %q", tt.binary, tt.pmName, got, tt.wantPkg)
			}
		})
	}
}

func TestInstallSuggestions_Integration(t *testing.T) {
	// Skip in CI - this is more of an integration test
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binaries := []string{"gcc", "python3", "make", "cmake"}

	pm := DetectPackageManager()
	if pm == nil {
		t.Skip("no package manager detected")
	}

	for _, binary := range binaries {
		suggestion := GetInstallSuggestion(binary, pm)
		if suggestion == "" {
			t.Errorf("GetInstallSuggestion(%q) returned empty string", binary)
		}
		t.Logf("%s: %s", binary, suggestion)
	}
}
