# build - a simpler build system

`build` is a build system inspired by GNU Make's target and recipe specification philosophy minus the complexities and quirks of GNU Make.

## Features

- **Simple, clean syntax** - Indentation-based, no tabs vs spaces issues
- **Pattern targets** - `build/{name}.o: src/{name}.c` matches any file
- **Variable interpolation** - `{var}` syntax with optional `:raw` modifier
- **Built-in functions** - `shell()`, `glob()`, `filename()`, `dirname()`, `replace()`
- **Conditionals** - `if/elif/else/end` and `ifdef/ifndef` blocks
- **Parallel execution** - Built-in parallel job support with `-j` flag
- **Environment management** - Docker, Podman, devcontainer, Nix, and Lima support
- **Include files** - Split large buildfiles with `.include:`

## Installation

```bash
go install github.com/vinayprograms/build@latest
```

### Prerequisites

- Go 1.25.3 or later (ensure `$GOPATH/bin` is in your `PATH`)

## Quick Start

Create a `Buildfile` in your project:

```bash
# Set shell
.shell: bash

# Variables
cc = gcc
cflags = -Wall -O2

# Phony target (always runs)
@clean:
    rm -rf build/

# File target
app: build/main.o build/utils.o
    {cc} -o {target} {deps}

# Pattern target (matches any .o file)
build/{name}.o: src/{name}.c
    mkdir -p build
    {cc} {cflags} -c {in} -o {out}
```

Build your project:

```bash
build app     # Build the app target
build clean   # Run the clean target
build         # Build the default target (first defined)
```

## Usage

```
build [options] [targets...]

Options:
  -f, --file PATH     Use alternate Buildfile
  -e, --env NAME      Use named environment
  -j, --jobs N        Parallel jobs (default: 1)
  -n, --dry-run       Show what would execute
  -v, --verbose       Verbose output
  --check-env         Verify environment requirements
  --show-install      Show install instructions
  --list-env          List available environments
  -V, --version       Show version
  -h, --help          Show help
```

## Syntax Overview

### Variables

```bash
# Immediate variable (evaluated at definition)
cc = gcc

# Lazy variable (evaluated at each use)
lazy timestamp = shell(date +%s)
```

### Conditionals

```bash
if {os} == linux
    cc = gcc
elif {os} == darwin
    cc = clang
end

ifdef DEBUG
    cflags = -g -DDEBUG
end
```

### Targets

```bash
# File target
output.txt: input.txt
    cp {in} {out}

# Phony target (no output file)
@test:
    go test ./...

# Phony target with phony dependencies
# Note: @ prefix is only needed for declaration, not for references
@all: build test
    echo "Done"

# Pattern target
build/{name}.o: src/{name}.c
    gcc -c {in} -o {out}
```

### Automatic Variables

| Variable | Description |
|----------|-------------|
| `{target}` | Target file path |
| `{out}` | Alias for target |
| `{deps}` | All dependencies (space-separated) |
| `{in}` | First dependency |
| `{stem}` | Matched pattern capture |
| `{target.dir}` | Directory part of target |
| `{target.file}` | Filename part of target |

### Built-in Functions

```bash
sources = glob(src/*.c)
objects = replace({sources}, .c, .o)
dir = dirname(/path/to/file.c)
base = filename(/path/to/file.c)
result = shell(echo hello)
```

### Environment Blocks

```bash
.environment:
    .using: docker
    .source: Dockerfile
    .args: --platform linux/amd64
    .requires: gcc@11 python3
```

## Documentation

- [Buildfile Specification](design/BUILDFILE_SPEC.md)
- [Design Document](design/DESIGN.md)
- [Code Architecture](design/CODE.md)

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Build failure (recipe returned non-zero) |
| 2 | Usage error (bad arguments) |
| 3 | Parse error (invalid Buildfile) |
| 4 | Environment error (missing requirements) |

## License

See [LICENSE](LICENSE) for details.
