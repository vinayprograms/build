# Migrating from Make to Buildfile

This guide helps you convert existing Makefiles to Buildfile format. Buildfile preserves Make's powerful dependency-driven build model while providing cleaner, more readable syntax.

---

## Quick Reference

| Make | Buildfile | Notes |
|------|-----------|-------|
| `CC = gcc` | `cc = gcc` | Lowercase convention |
| `$(CC)` | `{cc}` | Braces instead of parens |
| `$@` | `{target}` or `{out}` | Readable names |
| `$<` | `{in}` | First dependency |
| `$^` | `{deps}` | All dependencies |
| `$*` | `{stem}` | Pattern stem |
| `$(@D)` | `{target.dir}` | Target directory |
| `$(@F)` | `{target.file}` | Target filename |
| `.PHONY: clean` | `@clean:` | `@` prefix for phony |
| `%.o: %.c` | `{name}.o: {name}.c` | Named captures |
| Tab indentation | Any consistent indent | Spaces recommended |
| `include file.mk` | `.include: file.build` | Directive syntax |
| `ifeq ($(OS),Linux)` | `if {os} == linux` | Simpler conditionals |
| `$(shell cmd)` | `shell(cmd)` | Function syntax |
| `$(wildcard *.c)` | `glob(*.c)` | File patterns |

---

## Variables

### Simple Assignment

**Make:**
```make
CC = gcc
CFLAGS = -Wall -O2
SRC_DIR = src
```

**Buildfile:**
```
cc = gcc
cflags = -Wall -O2
src_dir = src
```

**Key differences:**
- No `$()` for references — use `{name}` instead
- Lowercase naming convention (not required, but recommended)
- No distinction between `=` and `:=` for simple values

### Variable References

**Make:**
```make
OBJECTS = $(SOURCES:.c=.o)
build/app: $(OBJECTS)
	$(CC) $(CFLAGS) -o $@ $^
```

**Buildfile:**
```
objects = replace({sources}, .c, .o)
build/app: {objects}
    {cc} {cflags:raw} -o {target} {deps}
```

**Key differences:**
- `{var}` instead of `$(VAR)`
- Use `{var:raw}` when you need word splitting (for flags)
- Built-in `replace()` function for pattern substitution

### Lazy vs Immediate Evaluation

**Make:**
```make
# Recursive (lazy)
FLAGS = $(EXTRA_FLAGS) -Wall

# Simple (immediate)  
FLAGS := $(EXTRA_FLAGS) -Wall
```

**Buildfile:**
```
# Immediate (default)
flags = {extra_flags} -Wall

# Lazy (explicit)
lazy flags = {extra_flags} -Wall
```

**Key difference:** Buildfile defaults to immediate evaluation. Use `lazy` keyword when you need deferred evaluation.

---

## Targets and Dependencies

### Basic Target

**Make:**
```make
build/app: build/main.o build/utils.o
	gcc -o $@ $^
```

**Buildfile:**
```
build/app: build/main.o build/utils.o
    gcc -o {target} {deps}
```

### Pattern Rules

**Make:**
```make
%.o: %.c
	$(CC) -c $< -o $@

build/%.o: src/%.c
	$(CC) -c $< -o $@
```

**Buildfile:**
```
{name}.o: {name}.c
    {cc} -c {in} -o {out}

build/{name}.o: src/{name}.c
    {cc} -c {in} -o {out}
```

**Key differences:**
- Named captures `{name}` instead of `%`
- Captures are explicit and can have meaningful names
- Same capture name must be used in target and dependencies

### Phony Targets

**Make:**
```make
.PHONY: all clean test

all: build/app

clean:
	rm -rf build/

test: build/app
	./build/app --test
```

**Buildfile:**
```
@all: build/app

@clean:
    rm -rf build/

@test: build/app
    ./build/app --test
```

**Key difference:** Use `@` prefix instead of `.PHONY` declaration.

### Directory Targets

**Make:**
```make
build:
	mkdir -p $@

build/app: build/main.o | build
	gcc -o $@ $<
```

**Buildfile:**
```
build/:
    mkdir -p {target}

build/app: build/main.o
    .after: build/
    gcc -o {target} {in}
```

**Key differences:**
- Directory targets end with `/`
- Order-only prerequisites use `.after:` directive

---

## Automatic Variables

| Make | Buildfile | Description |
|------|-----------|-------------|
| `$@` | `{target}` | Target being built |
| `$@` | `{out}` | Alias for target |
| `$<` | `{in}` | First prerequisite |
| `$^` | `{deps}` | All prerequisites |
| `$*` | `{stem}` | Pattern match stem |
| `$(@D)` | `{target.dir}` | Directory of target |
| `$(@F)` | `{target.file}` | Filename of target |

**Make:**
```make
build/%.o: src/%.c
	@echo "Compiling $< to $@"
	$(CC) -c $< -o $@
	@echo "Stem: $*"
```

**Buildfile:**
```
build/{name}.o: src/{name}.c
    echo "Compiling {in} to {out}"
    {cc} -c {in} -o {out}
    echo "Stem: {name}"
```

---

## Shell Commands

### Basic Commands

**Make:**
```make
clean:
	rm -rf build/
	@echo "Cleaned"
```

**Buildfile:**
```
@clean:
    rm -rf build/
    echo "Cleaned"
```

**Key difference:** No `@` prefix needed to suppress echo — Buildfile doesn't echo commands by default. Use `--verbose` flag to see commands.

### Multi-line Scripts

**Make:**
```make
deploy:
	if [ -f config.json ]; then \
		cp config.json build/; \
	fi
	./deploy.sh
```

**Buildfile:**
```
@deploy:
    block:
        if [ -f config.json ]; then
            cp config.json build/
        fi
    ./deploy.sh
```

**Key difference:** Use `block:` for multi-line shell scripts that need to run in a single shell invocation.

### Shell Selection

**Make:**
```make
SHELL = /bin/bash

target:
	[[ -f file ]] && echo "exists"
```

**Buildfile:**
```
.shell: bash

target:
    [[ -f file ]] && echo "exists"
```

Or per-recipe:

```
target:
    .shell: bash
    [[ -f file ]] && echo "exists"
```

---

## Functions

### Shell Command Output

**Make:**
```make
GIT_COMMIT := $(shell git rev-parse HEAD)
SOURCES := $(shell find src -name "*.c")
```

**Buildfile:**
```
git_commit = shell(git rev-parse HEAD)
sources = shell(find src -name "*.c")
```

### File Patterns

**Make:**
```make
SOURCES := $(wildcard src/*.c)
HEADERS := $(wildcard include/*.h)
```

**Buildfile:**
```
sources = glob(src/*.c)
headers = glob(include/*.h)
```

### Path Manipulation

**Make:**
```make
BASENAME := $(basename src/main.c)
DIRNAME := $(dir src/main.c)
```

**Buildfile:**
```
base = filename(src/main.c)
dir = dirname(src/main.c)
```

### Pattern Substitution

**Make:**
```make
OBJECTS := $(SOURCES:.c=.o)
OBJECTS := $(patsubst %.c,%.o,$(SOURCES))
```

**Buildfile:**
```
objects = replace({sources}, .c, .o)
```

For complex transformations, use shell:

```
objects = shell(echo {sources:raw} | sed 's/\.c/.o/g')
```

---

## Conditionals

### OS/Platform Detection

**Make:**
```make
ifeq ($(OS),Windows_NT)
    LDFLAGS = -lws2_32
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        LDFLAGS = -lpthread
    endif
    ifeq ($(UNAME_S),Darwin)
        LDFLAGS = -framework CoreFoundation
    endif
endif
```

**Buildfile:**
```
if {os} == linux
    ldflags = -lpthread
elif {os} == darwin
    ldflags = -framework CoreFoundation
elif {os} == windows
    ldflags = -lws2_32
else
    ldflags =
end
```

**Key differences:**
- Built-in `{os}` and `{arch}` variables
- Cleaner `if/elif/else/end` syntax
- No nested `ifeq` needed

### Variable Existence

**Make:**
```make
ifdef DEBUG
    CFLAGS += -g -O0
else
    CFLAGS += -O2
endif

ifndef CC
    CC = gcc
endif
```

**Buildfile:**
```
ifdef DEBUG
    cflags = -g -O0
else
    cflags = -O2
end

ifndef CC
    cc = gcc
end
```

---

## Includes

**Make:**
```make
include config.mk
include rules.mk
-include optional.mk
```

**Buildfile:**
```
.include: config.build
.include: rules.build
```

**Key differences:**
- Use `.include:` directive
- No optional include (file must exist)
- Circular includes are detected and rejected

---

## Common Patterns

### C Project

**Make:**
```make
CC = gcc
CFLAGS = -Wall -O2
LDFLAGS =

SRC_DIR = src
BUILD_DIR = build

SOURCES = $(wildcard $(SRC_DIR)/*.c)
OBJECTS = $(SOURCES:$(SRC_DIR)/%.c=$(BUILD_DIR)/%.o)

.PHONY: all clean

all: $(BUILD_DIR)/app

$(BUILD_DIR)/app: $(OBJECTS) | $(BUILD_DIR)
	$(CC) $(LDFLAGS) -o $@ $^

$(BUILD_DIR)/%.o: $(SRC_DIR)/%.c | $(BUILD_DIR)
	$(CC) $(CFLAGS) -c $< -o $@

$(BUILD_DIR):
	mkdir -p $@

clean:
	rm -rf $(BUILD_DIR)
```

**Buildfile:**
```
cc = gcc
cflags = -Wall -O2
ldflags =

src_dir = src
build_dir = build

sources = glob({src_dir}/*.c)
objects = replace({sources}, {src_dir}/, {build_dir}/)
objects = replace({objects}, .c, .o)

.default: @all

@all: {build_dir}/app

{build_dir}/app: {objects}
    .after: {build_dir}/
    {cc} {ldflags:raw} -o {target} {deps}

{build_dir}/{name}.o: {src_dir}/{name}.c
    .after: {build_dir}/
    {cc} {cflags:raw} -c {in} -o {out}

{build_dir}/:
    mkdir -p {target}

@clean:
    rm -rf {build_dir}
```

### Auto-dependencies for C

**Make:**
```make
DEPFLAGS = -MMD -MF $(@:.o=.d)

$(BUILD_DIR)/%.o: $(SRC_DIR)/%.c
	$(CC) $(CFLAGS) $(DEPFLAGS) -c $< -o $@

-include $(OBJECTS:.o=.d)
```

**Buildfile:**
```
{build_dir}/{name}.o: {src_dir}/{name}.c
    .autodeps: {build_dir}/{name}.d
    {cc} {cflags:raw} -MMD -MF {build_dir}/{name}.d -c {in} -o {out}
```

**Key difference:** Use `.autodeps:` directive — no need for manual include.

### Multi-configuration Build

**Make:**
```make
ifeq ($(CONFIG),debug)
    CFLAGS = -g -O0 -DDEBUG
else ifeq ($(CONFIG),release)
    CFLAGS = -O3 -DNDEBUG
else
    CFLAGS = -O2
endif
```

**Buildfile:**
```
ifdef DEBUG
    cflags = -g -O0 -DDEBUG
elif ifdef RELEASE
    cflags = -O3 -DNDEBUG
else
    cflags = -O2
end
```

Or use environments:

```
.environment: debug
    .using: bare

.environment: release
    .using: bare
```

And set variables based on environment selection in your build logic.

---

## What's Different

### Things Buildfile Does Differently

1. **No tab requirement** — Use any consistent indentation
2. **Readable automatic variables** — `{target}` not `$@`
3. **Named pattern captures** — `{name}` not `%`
4. **Explicit phony targets** — `@clean` not `.PHONY: clean`
5. **Built-in OS/arch detection** — `{os}`, `{arch}` variables
6. **Block mode for multi-line scripts** — `block:` keyword
7. **Environment management** — Docker, Podman, Nix, etc.

### Things Make Has That Buildfile Doesn't

1. **Automatic inference rules** — Define patterns explicitly
2. **VPATH** — Use explicit path patterns
3. **Multiple targets per rule** — Use pattern rules or separate targets
4. **Recursive make** — Use `.include:` or shell commands
5. **eval/call functions** — Use shell for complex logic
6. **Double-colon rules** — Not supported

### Things to Watch For

1. **Variable quoting** — `{var}` is shell-quoted by default; use `{var:raw}` for flags
2. **No automatic echo suppression** — Commands don't echo by default (opposite of Make)
3. **Strict dependency checking** — Missing dependencies are errors, not warnings
4. **No implicit rules** — All patterns must be explicitly defined

---

## Migration Checklist

- [ ] Replace `$(VAR)` with `{var}`
- [ ] Replace automatic variables (`$@` → `{target}`, etc.)
- [ ] Convert `%` patterns to `{name}` captures
- [ ] Add `@` prefix to phony targets (remove `.PHONY:`)
- [ ] Convert `ifeq/ifdef` to `if/ifdef` syntax
- [ ] Replace `include` with `.include:`
- [ ] Replace `$(shell ...)` with `shell(...)`
- [ ] Replace `$(wildcard ...)` with `glob(...)`
- [ ] Add `.after:` for order-only prerequisites
- [ ] Convert multi-line commands to `block:` where needed
- [ ] Add `.shell:` if using bash-specific features
- [ ] Set up `.environment:` for build requirements

---

## Getting Help

- Run `build --help` for command-line options
- See `BUILDFILE_SPEC.md` for complete language specification
- Check examples in the `test/integration/fixtures/` directory
