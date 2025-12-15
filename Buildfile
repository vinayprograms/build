# Build Tool - Buildfile
#
# This Buildfile builds the build tool itself.
# Use: go run ./cmd/build --debug-lex Buildfile

# ====================
# Configuration
# ====================

.shell: /bin/bash
.default: all

# ====================
# Variables
# ====================

binary = build
build_dir = bin
cmd_dir = cmd/build
go_files = shell(find . -name "*.go" -not -path "./bin/*")

# Platform-specific settings
if {os} == darwin
    ldflags = -s -w
elif {os} == linux
    ldflags = -s -w -linkmode external -extldflags "-static"
else
    ldflags = -s -w
end

version = shell(git describe --tags --always --dirty 2>/dev/null || echo dev)
commit = shell(git rev-parse --short HEAD 2>/dev/null || echo unknown)

# ====================
# Phony Targets
# ====================

@all: {build_dir}/{binary}

@clean:
    rm -rf {build_dir}

@test:
    go test -v ./...

@test-cover:
    go test -cover ./...

@lint:
    go vet ./...
    gofmt -l .

@fmt:
    gofmt -w .

@check: lint test

# ====================
# Build Targets
# ====================

{build_dir}/:
    mkdir -p {build_dir}

{build_dir}/{binary}: {go_files}
    .after: {build_dir}/
    block:
        echo "Building {binary}..."
        go build -ldflags "{ldflags:raw} -X main.version={version} -X main.commit={commit}" -o {target} ./{cmd_dir}
        echo "Built {target}"

# ====================
# Development Targets
# ====================

@run:
    go run ./{cmd_dir} --help

@debug-lex:
    go run ./{cmd_dir} --debug-lex Buildfile

@install: {build_dir}/{binary}
    cp {build_dir}/{binary} ~/bin/
