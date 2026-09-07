package environ

import (
	"os"
	"os/exec"
	"strings"
)

// podmanSocketResolver resolves the Podman API socket to connect the Docker
// SDK client to, since Podman does not sit behind the Docker daemon socket.
// The dependencies here (env lookup, command execution, file existence) are
// injectable so the resolution order can be tested without a real podman
// installation or a live daemon.
type podmanSocketResolver struct {
	getenv     func(string) string
	runCommand func(name string, args ...string) *exec.Cmd
	fileExists func(path string) bool
}

// newPodmanSocketResolver creates a resolver using real OS/process facilities.
func newPodmanSocketResolver() *podmanSocketResolver {
	return &podmanSocketResolver{
		getenv:     os.Getenv,
		runCommand: exec.Command,
		fileExists: fileExistsOnDisk,
	}
}

// fileExistsOnDisk reports whether path exists on the local filesystem.
func fileExistsOnDisk(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// resolveHost determines the Docker SDK client "host" address to use for
// Podman, in this order:
//
//  1. $CONTAINER_HOST, if set.
//  2. `podman machine inspect --format '{{.ConnectionInfo.PodmanSocket.Path}}'`,
//     if it succeeds with a non-empty path that exists on the host (the
//     macOS/Windows podman-machine case).
//  3. `podman info --format '{{.Host.RemoteSocket.Path}}'`, if it succeeds
//     with a non-empty path that exists on the host (Linux rootless/rootful
//     podman, where the daemon runs directly on the host).
//  4. Empty string, signaling the caller should fall back to $DOCKER_HOST /
//     client.FromEnv, same as for the "docker" runtime.
//
// Paths returned by `podman info` can point inside a podman machine VM (e.g.
// /run/user/501/podman/podman.sock) rather than the host, so every candidate
// path is checked for existence on the host before use.
func (r *podmanSocketResolver) resolveHost() string {
	if host := strings.TrimSpace(r.getenv("CONTAINER_HOST")); host != "" {
		return normalizeSocketHost(host)
	}

	if path := r.runFormatCommand("podman", "machine", "inspect", "--format", "{{.ConnectionInfo.PodmanSocket.Path}}"); path != "" {
		if r.fileExists(path) {
			return "unix://" + path
		}
	}

	if path := r.runFormatCommand("podman", "info", "--format", "{{.Host.RemoteSocket.Path}}"); path != "" {
		if r.fileExists(path) {
			return "unix://" + path
		}
	}

	return ""
}

// runFormatCommand runs a command expected to print a single path (optionally
// with surrounding whitespace/quotes) and returns it, or "" if the command
// failed or printed nothing.
func (r *podmanSocketResolver) runFormatCommand(name string, args ...string) string {
	cmd := r.runCommand(name, args...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	path := strings.TrimSpace(string(out))
	path = strings.Trim(path, "'\"")
	return path
}

// normalizeSocketHost ensures a host value has a scheme, defaulting to a
// unix socket path (podman's CONTAINER_HOST is conventionally a bare path or
// a full "unix://..." URI).
func normalizeSocketHost(host string) string {
	if strings.Contains(host, "://") {
		return host
	}
	return "unix://" + host
}
