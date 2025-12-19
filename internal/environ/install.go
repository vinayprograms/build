package environ

import (
	"fmt"
	"os/exec"
	"runtime"
)

// PackageManager provides install commands for missing binaries.
type PackageManager interface {
	// Name returns the package manager name (e.g., "apt", "brew").
	Name() string
	// GetInstallCommand returns the command to install a binary.
	GetInstallCommand(binary string) string
}

// DetectPackageManager detects the system's package manager.
// Returns nil if no supported package manager is found.
func DetectPackageManager() PackageManager {
	switch runtime.GOOS {
	case "darwin":
		// macOS uses Homebrew
		if _, err := exec.LookPath("brew"); err == nil {
			return &brewPackageManager{}
		}
		return nil

	case "linux":
		// Try to detect Linux package manager
		return detectLinuxPackageManager()

	default:
		return nil
	}
}

// detectLinuxPackageManager detects the Linux package manager.
func detectLinuxPackageManager() PackageManager {
	// Check for package managers in order of preference
	managers := []struct {
		binary string
		pm     PackageManager
	}{
		{"apt", &aptPackageManager{}},
		{"dnf", &dnfPackageManager{}},
		{"pacman", &pacmanPackageManager{}},
		{"zypper", &zypperPackageManager{}},
		{"apk", &apkPackageManager{}},
	}

	for _, m := range managers {
		if _, err := exec.LookPath(m.binary); err == nil {
			return m.pm
		}
	}

	return nil
}

// GetInstallSuggestion returns the install command for a binary.
// Returns empty string if package manager is nil.
func GetInstallSuggestion(binary string, pm PackageManager) string {
	if pm == nil {
		return ""
	}
	return pm.GetInstallCommand(binary)
}

// mapBinaryToPackage maps a binary name to its package name.
// For most cases, the binary name equals the package name.
// This function can be extended with a mapping table for special cases.
func mapBinaryToPackage(binary, pmName string) string {
	// Common binary to package name mappings
	// Key: binary name, Value: map[pmName]packageName
	mappings := map[string]map[string]string{
		// Example: "cc" might be in "gcc" package
		// For now, most binaries have the same name as their package
	}

	if pkgMap, ok := mappings[binary]; ok {
		if pkg, ok := pkgMap[pmName]; ok {
			return pkg
		}
	}

	// Default: binary name equals package name
	return binary
}

// ----------------------------------------------------------------------------
// Package Manager Implementations
// ----------------------------------------------------------------------------

// aptPackageManager is the apt package manager (Debian/Ubuntu).
type aptPackageManager struct{}

func (p *aptPackageManager) Name() string {
	return "apt"
}

func (p *aptPackageManager) GetInstallCommand(binary string) string {
	pkg := mapBinaryToPackage(binary, "apt")
	return fmt.Sprintf("sudo apt install %s", pkg)
}

// brewPackageManager is the Homebrew package manager (macOS).
type brewPackageManager struct{}

func (p *brewPackageManager) Name() string {
	return "brew"
}

func (p *brewPackageManager) GetInstallCommand(binary string) string {
	pkg := mapBinaryToPackage(binary, "brew")
	return fmt.Sprintf("brew install %s", pkg)
}

// dnfPackageManager is the dnf package manager (Fedora/RHEL).
type dnfPackageManager struct{}

func (p *dnfPackageManager) Name() string {
	return "dnf"
}

func (p *dnfPackageManager) GetInstallCommand(binary string) string {
	pkg := mapBinaryToPackage(binary, "dnf")
	return fmt.Sprintf("sudo dnf install %s", pkg)
}

// pacmanPackageManager is the pacman package manager (Arch Linux).
type pacmanPackageManager struct{}

func (p *pacmanPackageManager) Name() string {
	return "pacman"
}

func (p *pacmanPackageManager) GetInstallCommand(binary string) string {
	pkg := mapBinaryToPackage(binary, "pacman")
	return fmt.Sprintf("sudo pacman -S %s", pkg)
}

// zypperPackageManager is the zypper package manager (openSUSE).
type zypperPackageManager struct{}

func (p *zypperPackageManager) Name() string {
	return "zypper"
}

func (p *zypperPackageManager) GetInstallCommand(binary string) string {
	pkg := mapBinaryToPackage(binary, "zypper")
	return fmt.Sprintf("sudo zypper install %s", pkg)
}

// apkPackageManager is the apk package manager (Alpine Linux).
type apkPackageManager struct{}

func (p *apkPackageManager) Name() string {
	return "apk"
}

func (p *apkPackageManager) GetInstallCommand(binary string) string {
	pkg := mapBinaryToPackage(binary, "apk")
	return fmt.Sprintf("apk add %s", pkg)
}
