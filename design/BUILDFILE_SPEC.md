# Buildfile Language Specification v1.0

## Overview

Buildfile is a language for describing build targets, their dependencies, and recipes to produce them. It preserves Make's powerful file-based dependency model while eliminating its syntactic quirks.

**Design principles:**

- Targets are files
- Rebuild only when dependencies are newer than target
- Language-agnostic — it's about files, not code
- Readable, minimal punctuation
- No Turing completeness — use shell for complex logic

---

## Indentation

Buildfile uses indentation to define structure. Unlike Make, **tabs are not required**.

**Rules:**

1. Top-level elements (directives, variables, targets) start at column 0
2. Recipe content is indented from the target line
3. Block content is indented from the `block:` line
4. Consistent indentation within a scope is required

**Accepted indentation:**

- Spaces (any consistent width)
- Tabs
- Mixed tabs and spaces (not recommended)

**Best practices:**

- Use 4 spaces per indentation level
- Configure your editor to insert spaces
- Be consistent within a single Buildfile
- Avoid tabs to prevent invisible character issues

**Examples:**

```
# Good: 4 spaces
build/app: build/main.o
    gcc -o {target} {deps}

# Good: 2 spaces
build/app: build/main.o
  gcc -o {target} {deps}

# Good: tabs (if you must)
build/app: build/main.o
	gcc -o {target} {deps}

# Bad: inconsistent within recipe
build/app: build/main.o
    gcc -o {target} {deps}
      echo "done"
```

**Indentation levels:**

| Level | Contains |
|-------|----------|
| 0 | Directives, variables, conditionals, targets |
| 1 | Recipe directives and commands |
| 2 | Block script content |

---

## Directives

Directives configure build behavior. They are prefixed with `.` to distinguish them from variables and targets.

### Global Directives

| Directive | Purpose | Default |
|-----------|---------|---------|
| `.shell:` | Default shell for recipes | `/bin/sh` |
| `.parallel:` | Maximum parallel jobs | `1` |
| `.default:` | Default target when none specified | First target |
| `.include:` | Include another Buildfile | — |

**Example:**

```
.shell: bash
.parallel: 4
.default: @all
.include: ./common.build
```

### Recipe Directives

| Directive | Purpose |
|-----------|---------|
| `.shell:` | Override shell for this recipe |
| `.after:` | Order-only prerequisites |
| `.autodeps:` | File containing generated dependencies |
| `.requires:` | Additional binaries required for this target |

**Example:**

```
build/app: build/main.o
    .after: build/
    .autodeps: build/app.d
    .shell: bash
    .requires: pkg-config@latest
    gcc -o {target} {deps}
```

---

## Environments

Environments define where and how builds run.

### Environment Directives

| Directive | Purpose |
|-----------|---------|
| `.environment:` | Default (unnamed) environment |
| `.environment: name` | Named environment |
| `.using:` | Runtime type |
| `.source:` | Path to environment definition |
| `.args:` | Additional arguments for runtime |
| `.requires:` | Required binaries (bare environments) |

`.source:` may reference any variable defined earlier in the file, including
nested `{a}/{b}` chains — it is resolved with the same evaluator used for the
rest of the Buildfile:

```
docker_dir = ./docker
.environment: ci
    .using: docker
    .source: {docker_dir}/ci.Dockerfile
```

An automatic variable (`{target}`, `{deps}`, ...) or an undefined variable in
`.source:` is an error. `--check-env` and `--list-env` resolve `.source:` the
same way, so what they report matches what a real build uses.

### Runtime Types

| Runtime | Config file | Description |
|---------|-------------|-------------|
| `bare` | (none) | Host system directly |
| `docker` | Dockerfile | Docker container |
| `podman` | Dockerfile or Containerfile | Podman container |
| `devcontainer` | .devcontainer/ or devcontainer.json | VS Code devcontainer |
| `nix` | shell.nix or flake.nix | Nix shell |
| `lima` | lima.yaml | Lima VM (macOS) |

### Examples

**Bare environment with requirements:**

```
.environment:
    .requires: gcc@latest python3@3.10 pkg-config@latest
```

**Docker environment:**

```
.environment: ci
    .using: docker
    .source: ./docker/ci.Dockerfile
    .args: --platform linux/amd64
```

**Devcontainer environment:**

```
.environment: dev
    .using: devcontainer
    .source: ./.devcontainer
```

**Nix environment:**

```
.environment: nix
    .using: nix
    .source: ./shell.nix
    .args: --pure
```

### Multiple Environments

```
.environment:
    .requires: gcc@latest python3@3.10

.environment: ci
    .using: docker
    .source: ./docker/ci.Dockerfile
    .args: --platform linux/amd64

.environment: ci-arm
    .using: docker
    .source: ./docker/ci.Dockerfile
    .args: --platform linux/arm64

.environment: dev
    .using: devcontainer
    .source: ./.devcontainer
```

### Environment Selection Defaults

| Scenario | Behavior |
|----------|----------|
| No `.environment:` block | Bare, no requirements |
| Unnamed `.environment:` exists | Use as default |
| Only named environments exist | Error, require `--env name` |
| `--env name` specified | Use that environment |
| `BUILD_ENV` variable set | Use that environment (lower precedence than `--env`) |

### Shared Environments

Use `.include:` to share environments across Buildfiles:

```
# environments.build
.environment:
    .requires: gcc@latest python3@3.10

.environment: ci
    .using: docker
    .source: ./docker/ci.Dockerfile
```

```
# Buildfile
.include: ./environments.build

@all: build/app
```

---

## Variables

### Immediate Assignment

Evaluated at definition:

```
cc = gcc
cflags = -Wall -O2
sources = shell(find src -name "*.c")
```

### Lazy Assignment

Evaluated at use:

```
lazy all_flags = {cflags} {extra_flags}
```

### Trailing Whitespace

Trailing spaces and tabs at the end of a line are trimmed from variable
values, directive values, and dependency lists:

```
greeting = hello   
# greeting == "hello", not "hello   "
```

Internal spaces are always preserved (`greeting = hello   world` keeps the
three spaces between the words). This trimming applies only to values — it
does not affect recipe command lines or `block:` lines, where every space is
significant.

### Variable Interpolation

Variables are interpolated using braces:

```
build/app: build/main.o
    {cc} {cflags} -o {target} {deps}
```

### Interpolation in Shell

| Syntax | Behavior |
|--------|----------|
| `{var}` | Context-aware quoting (default, safe) — see below |
| `{var:raw}` | Unquoted, allows word splitting; emitted completely untouched in every context |
| `$var` | Passed to shell as-is |

`{var}` (and automatic variables such as `{target}`, `{deps}`, `{in}`,
`{out}`, `{stem}`, ...) are formatted according to the shell quoting
context they land in. That context is tracked left to right over the
literal text of the command/block line (respecting backslash escapes;
`$(...)`, `` `...` `` and heredoc bodies count as double-quoted context):

| Context | Behavior |
|---------|----------|
| Outside quotes | Emitted bare if it contains only characters from `[A-Za-z0-9_./:@%+=,-]`; otherwise wrapped in single quotes, with an embedded `'` emitted as `'\''` |
| Inside `"..."` | Emitted with `"`, `$`, `` ` `` and `\` each escaped by a backslash. Never wrapped |
| Inside `'...'` | Emitted raw, with `'` turned into `'\''` |

`{deps}` follows the same rule. Outside quotes it expands to one shell word
per dependency, each quoted only when needed; inside quotes it is the
space-joined list, escaped for the surrounding quotes.

**Examples** (`dir = my dir`, `flag = $HOME`, `name = build`, `json =
{"key": "value"}`):

```
ls {dir}                    # → ls 'my dir'
echo "Dir: {dir}"           # prints: Dir: my dir
echo "Home: {flag}"         # prints: Home: $HOME
echo 'Dir: {dir}'           # prints: Dir: my dir
cp {name}.o out/            # → cp build.o out/
echo "JSON: {json}"         # prints: JSON: {"key": "value"}
echo "It's {dir}"           # prints: It's my dir
```

```
src_dir = my sources
flags = -Wall -O2

# Quoted (default) — safe for paths with spaces
sources = shell(find {src_dir} -name "*.c")
# Executes: find 'my sources' -name "*.c"

# Raw — for flags that need word splitting
result = shell(gcc {flags:raw} -c main.c)
# Executes: gcc -Wall -O2 -c main.c

# Shell variables passed through
files = shell(for f in *.c; do echo $f; done)
```

### Built-in Functions

| Function | Purpose | Example |
|----------|---------|---------|
| `shell(...)` | Execute shell command, capture output | `shell(pkg-config --libs gtk)` |
| `glob(...)` | File pattern matching | `glob(src/*.c)` |
| `filename(...)` | Extract filename without directory | `filename(src/main.c)` → `main.c` |
| `dirname(...)` | Extract directory | `dirname(src/main.c)` → `src` |
| `replace(...)` | Pattern replacement | `replace({sources}, .c, .o)` |

For complex transformations, use `shell()`:

```
objects = shell(echo {sources:raw} | sed 's/\.c/.o/g')
```

### Built-in Variables

| Variable | Values |
|----------|--------|
| `{os}` | `linux`, `darwin`, `windows`, `freebsd`, etc. |
| `{arch}` | `amd64`, `arm64`, `386`, etc. |

---

## Targets

### Basic Syntax

```
target: dependency1 dependency2
    command1
    command2
```

### File Targets

```
build/app: build/main.o build/utils.o
    gcc -o {target} {deps}
```

### Pattern Targets

Use `{name}` for captures:

```
build/{name}.o: src/{name}.c
    gcc -c {in} -o {out}
```

When building `build/utils.o`:
- `{name}` resolves to `utils`
- `{in}` resolves to `src/utils.c`
- `{out}` resolves to `build/utils.o`

### Phony Targets

Prefixed with `@`:

```
@all: build/app

@clean:
    rm -rf build/

@test: build/app
    ./build/app --test
```

Phony targets always run — they don't correspond to files.

### Directory Targets

```
build/:
    mkdir -p {target}
```

### Default Target

The default target is determined by:

1. `.default:` directive if specified
2. Otherwise, first target in file

```
.default: @all
```

---

## Automatic Variables

Available in recipes:

| Variable | Meaning | Make equivalent |
|----------|---------|-----------------|
| `{target}` | Target file path | `$@` |
| `{deps}` | All dependencies (space-separated) | `$^` |
| `{in}` | First dependency | `$<` |
| `{out}` | Target file path (alias for `{target}`) | `$@` |
| `{stem}` | Pattern match stem | `$*` |
| `{target.dir}` | Target directory | `$(@D)` |
| `{target.file}` | Target filename | `$(@F)` |

---

## Recipes

### Line Mode (Default)

Each indented line runs as a separate shell invocation:

```
build/app: build/main.o
    echo "Linking..."
    gcc -o {target} {deps}
    echo "Done"
```

### Block Mode

Multiple lines passed as a single script:

```
build/app: build/main.o
    echo "Starting"
    block:
        if [[ -f {target} ]]; then
            rm {target}
        fi
        gcc -o {target} {deps}
    echo "Finished"
```

### Shell Selection

Global default:

```
.shell: bash
```

Per-recipe override:

```
build/app: build/main.o
    .shell: zsh
    gcc -o {target} {deps}
```

---

## Order-only Prerequisites

Dependencies that must exist but whose timestamps are ignored:

```
build/app: build/main.o build/utils.o
    .after: build/
    gcc -o {target} {deps}
```

The `build/` directory must exist before the recipe runs, but changes to its timestamp won't trigger a rebuild.

---

## Automatic Dependency Generation

For languages with compiler-generated dependencies (C/C++):

```
build/{name}.o: src/{name}.c
    .autodeps: build/{name}.d
    gcc -MMD -MF build/{name}.d -c {in} -o {out}
```

build:

1. Runs the recipe
2. Parses the generated `.d` file
3. Stores learned dependencies for future builds

---

## Conditionals

```
if {os} == linux
    ldflags = -lpthread
elif {os} == darwin
    ldflags = -framework CoreFoundation
else
    ldflags =
end
```

**Variable existence:**

```
ifdef DEBUG
    cflags = -g -O0
end

ifndef CC
    cc = gcc
end
```

---

## Includes

```
.include: common.build
.include: config/{platform}.build
.include: ./environments.build
```

Variables are expanded in paths. Included files are processed as if inline.

`.include:` is resolved at parse time, before evaluation, so only variables
that are already fully known at that point in the parse can be used: an
immediate (non-`lazy`) variable assigned earlier in the same file (or, for
an included file, earlier in whichever file included it), with a value that
is itself fully literal after resolving any variables it references. A
`lazy` variable, a function call (`shell()`, `glob()`, ...), an automatic
variable (`{target}`, `{deps}`, ...), or an undefined variable in an
`.include:` path is a parse error: `.include: cannot resolve '<name>':
<reason>`.

---

## Comments

```
# This is a comment

build/app: build/main.o  # inline comment
    gcc -o {target} {deps}
```

---

## Version Syntax

For package requirements:

| Syntax | Meaning |
|--------|---------|
| `gcc` | Any version (alias for `@latest`) |
| `gcc@latest` | Latest available |
| `gcc@11` | Major version 11.x.x |
| `gcc@11.4` | Version 11.4.x |
| `gcc@11.4.0` | Exact version |

---

## Complete Example

```
# ====================
# Environments
# ====================

.environment:
    .requires: gcc@latest cmake@3.20 pkg-config@latest

.environment: ci
    .using: docker
    .source: ./docker/ci.Dockerfile
    .args: --platform linux/amd64

.environment: ci-arm
    .using: docker
    .source: ./docker/ci.Dockerfile
    .args: --platform linux/arm64

.environment: dev
    .using: devcontainer
    .source: ./.devcontainer

# ====================
# Global Configuration
# ====================

.shell: bash
.parallel: 8
.default: @all

# ====================
# Variables
# ====================

cc = gcc
cflags = -Wall -Wextra -O2
ldflags =

src_dir = src
build_dir = build

sources = shell(find {src_dir} -name "*.c")
objects = shell(echo {sources:raw} | sed 's|{src_dir}/|{build_dir}/|g' | sed 's|\.c$|.o|g')

ifdef DEBUG
    cflags = -Wall -Wextra -g -O0 -DDEBUG
end

if {os} == linux
    ldflags = -lpthread
elif {os} == darwin
    ldflags = -framework CoreFoundation
end

# ====================
# Phony Targets
# ====================

@all: {build_dir}/app

@clean:
    rm -rf {build_dir}

@test: {build_dir}/app
    ./{build_dir}/app --run-tests

@install: {build_dir}/app
    cp {build_dir}/app /usr/local/bin/

@docs: docs/index.html
    .requires: sphinx-build@latest doxygen@latest
    sphinx-build -b html docs/ docs/_build/

# ====================
# Build Targets
# ====================

{build_dir}/:
    mkdir -p {target}

{build_dir}/{name}.o: {src_dir}/{name}.c
    .after: {build_dir}/
    .autodeps: {build_dir}/{name}.d
    block:
        echo "Compiling {name}.c..."
        {cc} {cflags} -MMD -MF {build_dir}/{name}.d -c {in} -o {out}

{build_dir}/app: {objects}
    .after: {build_dir}/
    block:
        echo "Linking..."
        {cc} {ldflags} -o {target} {deps}
        echo "Build complete: {target}"
```

---

## CLI Reference

### Basic Commands

| Command | Description |
|---------|-------------|
| `build` | Run default target with default environment |
| `build target` | Build specific target |
| `build @phony` | Run phony target |

### Options

| Option | Short | Description |
|--------|-------|-------------|
| `--env name` | `-e name` | Use named environment |
| `--dry-run` | `-n` | Show what would execute without running |
| `--verbose` | `-v` | Verbose output |
| `--jobs N` | `-j N` | Parallel jobs (overrides `.parallel:`) |
| `--file path` | `-f path` | Use alternate Buildfile |
| `--check-env` | | Verify environment requirements |
| `--show-install` | | Show install instructions for missing requirements |
| `--list-env` | | List available environments |
| `--shell` | | Open shell in sandbox environment |
| `--keep` | | Keep sandbox running after build |
| `--help` | `-h` | Show help |
| `--version` | `-V` | Show version |

### CLI Examples

**Build default target:**

```bash
$ build
```

**Build specific target:**

```bash
$ build build/app
```

**Build phony target:**

```bash
$ build @clean
```

**Use named environment:**

```bash
$ build --env ci
```

**Use environment shorthand:**

```bash
$ build -e ci
```

**Dry run (show commands without executing):**

```bash
$ build -n

Would build: build/
  mkdir -p build/

Would build: build/main.o
  echo "Compiling main.c..."
  gcc -Wall -O2 -MMD -MF build/main.d -c src/main.c -o build/main.o

Would build: build/app
  echo "Linking..."
  gcc -o build/app build/main.o build/utils.o
```

**Verbose output:**

```bash
$ build -v

Evaluating variables...
  sources = shell(find src -name "*.c") → src/main.c src/utils.c
  objects → build/main.o build/utils.o

Checking targets...
  build/main.o: src/main.c is newer → rebuild
  build/utils.o: up to date → skip
  build/app: build/main.o changed → rebuild

Building build/main.o...
  gcc -Wall -O2 -c src/main.c -o build/main.o

Building build/app...
  gcc -o build/app build/main.o build/utils.o

Done.
```

**Parallel build:**

```bash
$ build -j 8
```

**Use alternate Buildfile:**

```bash
$ build -f ./other/Buildfile
```

**Check environment requirements:**

```bash
$ build --check-env

Checking environment...
  ✓ gcc (11.4.0)
  ✓ cmake (3.20.1)
  ✓ pkg-config (0.29.2)

Environment OK.
```

**Check named environment:**

```bash
$ build --check-env -e ci

Checking environment 'ci'...
  Runtime: docker
  Source: ./docker/ci.Dockerfile
  ✓ Dockerfile exists
  ✓ Docker available

Environment OK.
```

**Show install instructions:**

```bash
$ build --show-install

Missing requirements:
  cmake@3.20  → apt install cmake
              → brew install cmake
              → dnf install cmake
```

**List available environments:**

```bash
$ build --list-env

Available environments:
  (default)  bare, requires: gcc@latest cmake@3.20 pkg-config@latest
  ci         docker, source: ./docker/ci.Dockerfile
  ci-arm     docker, source: ./docker/ci.Dockerfile
  dev        devcontainer, source: ./.devcontainer
```

**Open shell in sandbox:**

```bash
$ build --env ci --shell

Starting sandbox 'ci'...
root@container:/workspace#
```

**Keep sandbox after build:**

```bash
$ build --env ci --keep

Building in sandbox 'ci'...
  [build output]

Sandbox kept running. To enter:
  docker exec -it build-ci-a1b2c3 /bin/bash

To stop:
  docker stop build-ci-a1b2c3
```

**Environment variable selection:**

```bash
$ BUILD_ENV=ci build
```

**Combine options:**

```bash
$ build -e ci -j 4 -v @test
```

---

## Quick Reference

### Make to Buildfile

| Make | Buildfile |
|------|-----------|
| `TAB` required | Any consistent indentation |
| `$@` | `{target}` |
| `$<` | `{in}` |
| `$^` | `{deps}` |
| `$*` | `{stem}` |
| `%.o: %.c` | `{name}.o: {name}.c` |
| `.PHONY: clean` | `@clean:` |
| `VAR := value` | `var = value` |
| `VAR = value` | `lazy var = value` |
| `$(shell ...)` | `shell(...)` |
| `$(wildcard ...)` | `glob(...)` |
| `-include file.d` | `.autodeps: file.d` |
| `\| order-only` | `.after: order-only` |
| `ifeq/endif` | `if/end` |
| `include file.mk` | `.include: file.mk` |

### Syntax Summary

| Element | Syntax |
|---------|--------|
| Immediate variable | `name = value` |
| Lazy variable | `lazy name = value` |
| Default target | `.default: target` |
| Global directive | `.directive: value` |
| Recipe directive | `.directive: value` (indented) |
| Phony target | `@name: deps` |
| Pattern target | `path/{name}.o: path/{name}.c` |
| Recipe command | Indented line |
| Recipe block | `block:` + deeper indentation |
| Interpolation | `{var}`, `{var:raw}` |
| Environment | `.environment:` or `.environment: name` |
| Runtime | `.using: docker` |
| Source | `.source: ./Dockerfile` |
| Arguments | `.args: --platform linux/amd64` |
| Requirements | `.requires: gcc@latest python3@3.10` |

---

## Implementation Notes

This section provides guidance for implementing the `build` tool.

### Parser Requirements

1. **Lexer**: Line-oriented, track indentation level
2. **Indentation**: Count leading spaces/tabs, enforce consistency per block
3. **Directives**: Lines starting with `.` at appropriate indent level
4. **Variables**: Lines containing `=` without `:` before it
5. **Targets**: Lines containing `:` with path-like left side
6. **Phony**: Target lines starting with `@`
7. **Pattern**: Target lines containing `{name}` captures
8. **Recipes**: Indented lines following a target
9. **Blocks**: `block:` keyword followed by deeper indentation
10. **Comments**: `#` to end of line

### Execution Model

1. **Parse phase**: Read Buildfile, resolve includes, build AST
2. **Evaluate phase**: Evaluate variables, execute `shell()` calls
3. **Plan phase**: Build dependency graph, determine what needs rebuilding
4. **Execute phase**: Run recipes in dependency order, respecting parallelism

### Dependency Resolution

1. Compare target mtime with all dependency mtimes
2. If any dependency is newer, target needs rebuild
3. Pattern targets generate concrete targets on demand
4. `.autodeps:` files are parsed after successful build, cached for next run

### Environment Execution

1. **Bare**: Verify `.requires:` binaries exist in PATH, then execute directly
2. **Docker/Podman**: Build image from `.source:`, run container with workspace mounted
3. **Devcontainer**: Use devcontainer CLI to start environment
4. **Nix**: Enter nix-shell with specified configuration

### Error Handling

1. Recipe failure: Stop build (unless `-k` keep-going flag)
2. Missing dependency: Error with clear message
3. Circular dependency: Error with cycle path
4. Missing binary: Error with install suggestions

---

This concludes the Buildfile language specification v1.0.
