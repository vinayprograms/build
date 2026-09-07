package environ

import (
	"os/exec"
	"testing"
)

// fakeCommand returns a runCommand func that ignores the requested
// name/args and instead runs a shell script producing the given stdout and
// exit code, so tests can control podman's simulated output without a real
// podman binary.
func fakeCommand(stdout string, succeed bool) func(name string, args ...string) *exec.Cmd {
	return func(_ string, _ ...string) *exec.Cmd {
		if !succeed {
			return exec.Command("sh", "-c", "exit 1")
		}
		return exec.Command("sh", "-c", "printf '%s' "+shellQuoteForTest(stdout))
	}
}

// shellQuoteForTest wraps s in single quotes for use in a `sh -c` script.
func shellQuoteForTest(s string) string {
	return "'" + s + "'"
}

// sequencedCommand returns a runCommand func that returns the Nth entry from
// results (by call order), clamped to the last entry if called more times
// than provided.
func sequencedCommand(results ...func(name string, args ...string) *exec.Cmd) func(name string, args ...string) *exec.Cmd {
	call := 0
	return func(name string, args ...string) *exec.Cmd {
		idx := call
		if idx >= len(results) {
			idx = len(results) - 1
		}
		call++
		return results[idx](name, args...)
	}
}

func TestPodmanSocketResolver_ContainerHostEnvTakesPriority(t *testing.T) {
	r := &podmanSocketResolver{
		getenv: func(key string) string {
			if key == "CONTAINER_HOST" {
				return "/tmp/container-host.sock"
			}
			return ""
		},
		runCommand: func(name string, args ...string) *exec.Cmd {
			t.Fatalf("runCommand should not be called when CONTAINER_HOST is set (called %s %v)", name, args)
			return nil
		},
		fileExists: func(string) bool { return true },
	}

	got := r.resolveHost()
	want := "unix:///tmp/container-host.sock"
	if got != want {
		t.Errorf("resolveHost() = %q, want %q", got, want)
	}
}

func TestPodmanSocketResolver_ContainerHostEnvWithScheme(t *testing.T) {
	r := &podmanSocketResolver{
		getenv: func(key string) string {
			if key == "CONTAINER_HOST" {
				return "unix:///tmp/container-host.sock"
			}
			return ""
		},
		runCommand: func(name string, args ...string) *exec.Cmd {
			t.Fatal("runCommand should not be called when CONTAINER_HOST is set")
			return nil
		},
		fileExists: func(string) bool { return true },
	}

	got := r.resolveHost()
	want := "unix:///tmp/container-host.sock"
	if got != want {
		t.Errorf("resolveHost() = %q, want %q", got, want)
	}
}

func TestPodmanSocketResolver_MachineInspectUsedWhenSocketExists(t *testing.T) {
	r := &podmanSocketResolver{
		getenv:     func(string) string { return "" },
		runCommand: fakeCommand("/var/folders/x/podman-machine-default-api.sock", true),
		fileExists: func(path string) bool {
			return path == "/var/folders/x/podman-machine-default-api.sock"
		},
	}

	got := r.resolveHost()
	want := "unix:///var/folders/x/podman-machine-default-api.sock"
	if got != want {
		t.Errorf("resolveHost() = %q, want %q", got, want)
	}
}

func TestPodmanSocketResolver_MachineInspectSocketDoesNotExist_FallsBackToInfo(t *testing.T) {
	r := &podmanSocketResolver{
		getenv: func(string) string { return "" },
		runCommand: sequencedCommand(
			fakeCommand("/run/user/501/podman/podman.sock", true), // machine inspect: succeeds but path is VM-internal
			fakeCommand("/run/user/501/podman/podman.sock", true), // podman info: same path
		),
		fileExists: func(path string) bool {
			// Neither path exists on the host (simulates a path inside the VM).
			return false
		},
	}

	got := r.resolveHost()
	if got != "" {
		t.Errorf("resolveHost() = %q, want empty string (fall back to DOCKER_HOST/FromEnv)", got)
	}
}

func TestPodmanSocketResolver_MachineInspectFails_FallsBackToInfo(t *testing.T) {
	r := &podmanSocketResolver{
		getenv: func(string) string { return "" },
		runCommand: sequencedCommand(
			fakeCommand("", false),                       // machine inspect: command fails (e.g. not on macOS/Windows)
			fakeCommand("/run/podman/podman.sock", true), // podman info: succeeds
		),
		fileExists: func(path string) bool {
			return path == "/run/podman/podman.sock"
		},
	}

	got := r.resolveHost()
	want := "unix:///run/podman/podman.sock"
	if got != want {
		t.Errorf("resolveHost() = %q, want %q", got, want)
	}
}

func TestPodmanSocketResolver_AllFail_ReturnsEmpty(t *testing.T) {
	r := &podmanSocketResolver{
		getenv:     func(string) string { return "" },
		runCommand: fakeCommand("", false),
		fileExists: func(string) bool { return false },
	}

	got := r.resolveHost()
	if got != "" {
		t.Errorf("resolveHost() = %q, want empty string", got)
	}
}

func TestPodmanSocketResolver_InfoPathDoesNotExist_ReturnsEmpty(t *testing.T) {
	r := &podmanSocketResolver{
		getenv: func(string) string { return "" },
		runCommand: sequencedCommand(
			fakeCommand("", false),
			fakeCommand("/run/user/501/podman/podman.sock", true),
		),
		fileExists: func(string) bool { return false }, // path is inside the VM, not on the host
	}

	got := r.resolveHost()
	if got != "" {
		t.Errorf("resolveHost() = %q, want empty string (VM-internal path must not be used)", got)
	}
}
