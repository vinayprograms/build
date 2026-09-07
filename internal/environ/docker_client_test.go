package environ

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

// TestWrapContainerCreateError_DockerDesktopFileSharing verifies that
// ContainerCreate failures whose daemon text matches a known Docker Desktop
// file-sharing rejection are wrapped with a hint naming the project
// directory and the Settings path to fix it (C7).
func TestWrapContainerCreateError_DockerDesktopFileSharing(t *testing.T) {
	tests := []struct {
		name    string
		errText string
	}{
		{"bind source path does not exist", "Error response from daemon: OCI runtime create failed: bind source path does not exist: /Users/me/project"},
		{"mounts denied", "Error response from daemon: Mounts denied: \nThe path /Users/me/project is not shared from the host and is not known to Docker."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostDir := "myproject"
			err := wrapContainerCreateError(errors.New(tt.errText), ast.RuntimeDocker, hostDir)
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			abs, absErr := filepath.Abs(hostDir)
			if absErr != nil {
				t.Fatalf("filepath.Abs(%q): %v", hostDir, absErr)
			}
			msg := err.Error()
			if !strings.Contains(msg, "failed to create container:") {
				t.Errorf("error %q should still contain the base wrap message", msg)
			}
			if !strings.Contains(msg, abs) {
				t.Errorf("error %q should mention project directory %q", msg, abs)
			}
			if !strings.Contains(msg, "not shared with Docker Desktop") {
				t.Errorf("error %q should mention Docker Desktop file sharing", msg)
			}
			if !strings.Contains(msg, "Settings > Resources > File sharing") {
				t.Errorf("error %q should mention the Settings path", msg)
			}
		})
	}
}

// TestWrapContainerCreateError_PodmanUnaffected verifies the file-sharing
// hint is Docker-Desktop specific: the same daemon text under the Podman
// runtime gets the plain wrap, no hint (C7).
func TestWrapContainerCreateError_PodmanUnaffected(t *testing.T) {
	err := wrapContainerCreateError(errors.New("bind source path does not exist: /some/path"), ast.RuntimePodman, "myproject")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	msg := err.Error()
	if strings.Contains(msg, "Docker Desktop") {
		t.Errorf("podman error should not mention Docker Desktop, got %q", msg)
	}
	want := "failed to create container: bind source path does not exist: /some/path"
	if msg != want {
		t.Errorf("msg = %q, want %q", msg, want)
	}
}

// TestWrapContainerCreateError_OtherErrorsUnchanged verifies errors that
// don't match a known file-sharing marker keep the plain wrap, even under
// the Docker runtime (C7).
func TestWrapContainerCreateError_OtherErrorsUnchanged(t *testing.T) {
	err := wrapContainerCreateError(errors.New("no such image: myimage:latest"), ast.RuntimeDocker, "myproject")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	want := "failed to create container: no such image: myimage:latest"
	if err.Error() != want {
		t.Errorf("msg = %q, want %q", err.Error(), want)
	}
}

// TestWrapContainerCreateError_Nil verifies a nil error passes through.
func TestWrapContainerCreateError_Nil(t *testing.T) {
	if err := wrapContainerCreateError(nil, ast.RuntimeDocker, "myproject"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}
