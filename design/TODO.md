# Build Tool Implementation Plan

This document contains a comprehensive list of implementation tasks, ordered from **highest risk to lowest risk**. Risk is assessed based on:

1. **Technical complexity** - How difficult is the implementation?
2. **Unknowns** - How much discovery/prototyping is needed?
3. **Architectural impact** - Will changes here ripple through the codebase?
4. **Dependency blocking** - Does this block other work?

---

## Phase 1: Core Language Processing (Highest Risk)

These components form the foundation. Errors here cascade through everything else.

### 1.1 Lexer Implementation

The lexer must handle context-free tokenization with careful boundary detection for interpolations.

- [x] **Design token type enumeration**
  - [x] Define all token types from spec (TOKEN_EOF, TOKEN_NEWLINE, TOKEN_INDENT, etc.)
  - [x] Define token structure with source location tracking (file, line, column)
  - [x] Design error token type for error recovery

- [x] **Implement indentation tracking**
  - [x] Track indentation character (space vs tab) from first indented line
  - [x] Calculate logical indentation level (0, 1, 2)
  - [x] Detect and report mixed indentation errors
  - [x] Handle empty lines and comment-only lines correctly

- [x] **Implement interpolation boundary detection**
  - [x] Recognize `{` as INTERP_START when preceded by boundary chars (whitespace, SOL, `:`, `=`, `/`, `"`, `'`, `(`, `)`, `,`, `>`, `<`)
  - [x] Verify following character is valid identifier start
  - [x] Handle `{{` and `}}` escape sequences
  - [x] Parse `:raw` modifier inside interpolations
  - [x] Ensure `${var}` and `x{var}` are NOT recognized as interpolations

- [x] **Implement directive keyword recognition**
  - [x] Recognize all `.keyword` forms (.shell, .parallel, .default, .include, etc.)
  - [x] Distinguish global vs recipe-level directives by indentation
  - [x] Handle `.environment:` with optional name

- [x] **Implement line classification**
  - [x] Classify by first non-whitespace token (directive, keyword, target, variable, etc.)
  - [x] Handle comment lines starting with `#`
  - [x] Handle inline comments after statements
  - [x] Handle blank lines

- [x] **Implement lexer state machine**
  - [x] Define states: LINE_START, NORMAL, INTERPOLATION, STRING_VALUE
  - [x] Implement state transitions per spec
  - [x] Handle end-of-file gracefully

- [x] **Write lexer unit tests**
  - [x] Test all token types individually
  - [x] Test interpolation boundary cases extensively
  - [x] Test indentation edge cases (mixed, inconsistent, tabs)
  - [x] Test escape sequences
  - [x] Test error cases and recovery

### 1.2 Parser Implementation

The parser builds the AST with scope-aware directive validation.

- [x] **Define AST node types**
  - [x] Buildfile (root)
  - [x] Statement enum (Directive, Environment, Variable, Conditional, Target, Comment, Blank)
  - [x] Directive and DirectiveKind
  - [x] Environment with Runtime enum
  - [x] Variable with lazy flag
  - [x] Conditional with branches
  - [x] Target with TargetPattern and Dependency
  - [x] Recipe with RecipeDirectives and Command
  - [x] Value and ValuePart for interpolated strings
  - [x] SourceLocation for all nodes

- [ ] **Implement parser scope stack**
  - [ ] Define Scope enum (GLOBAL, ENVIRONMENT, RECIPE, BLOCK)
  - [ ] Push/pop scope on block entry/exit
  - [ ] Validate directive placement based on current scope
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement variable parsing**
  - [ ] Detect `=` before `:` to distinguish from targets
  - [ ] Parse `lazy` keyword prefix
  - [ ] Parse right-hand side as Value with interpolations and function calls
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement target parsing**
  - [ ] Parse target pattern (file path with `{name}` segments)
  - [ ] Parse phony targets (`@name`)
  - [ ] Parse directory targets (ending with `/`)
  - [ ] Parse dependency list
  - [ ] Handle pattern targets with captures
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement recipe parsing**
  - [ ] Detect recipe start by indentation after target
  - [ ] Parse recipe directives (.shell, .after, .autodeps, .requires)
  - [ ] Parse command lines with interpolations
  - [ ] Parse `block:` with deeper indentation
  - [ ] Handle dedent to end recipe
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement environment block parsing**
  - [ ] Parse `.environment:` with optional name
  - [ ] Enter ENVIRONMENT scope
  - [ ] Parse environment directives (.using, .source, .args, .requires)
  - [ ] Handle dedent to end environment block
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement conditional parsing**
  - [ ] Parse `if`, `elif`, `else`, `end` keywords
  - [ ] Parse conditions (`{var} == value`, `{var} != value`)
  - [ ] Parse `ifdef` and `ifndef` variants
  - [ ] Collect body statements for each branch
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement value parsing**
  - [ ] Parse literal strings
  - [ ] Parse interpolations (`{var}`, `{var:raw}`)
  - [ ] Parse function calls (shell, glob, basename, dirname, replace)
  - [ ] Handle nested parentheses in function arguments
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement `.include:` directive**
  - [ ] Parse include path with variable interpolation
  - [ ] Recursively lex/parse included file
  - [ ] Merge included AST into parent
  - [ ] Detect and prevent circular includes
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement error recovery**
  - [ ] On parse error, skip to next line at indentation level 0
  - [ ] Collect multiple errors before failing
  - [ ] Provide actionable error messages with source locations
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Write parser unit tests**
  - [ ] Test all statement types
  - [ ] Test scope validation for directives
  - [ ] Test nested blocks (recipe → block)
  - [ ] Test conditionals with all branch combinations
  - [ ] Test include directive
  - [ ] Test error recovery and messages
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 1.3 Semantic Analysis

Validates the AST and resolves ambiguous constructs.

- [ ] **Implement symbol table**
  - [ ] Define SymbolTable structure (variables, targets, environments, automatic)
  - [ ] Populate automatic variable set (target, deps, in, out, stem, target.dir, target.file)
  - [ ] Populate built-in variable set (os, arch)
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement Pass 1: Symbol Collection**
  - [ ] Collect all variable definitions
  - [ ] Collect all target definitions
  - [ ] Collect all environment definitions
  - [ ] Detect duplicate definitions with clear error messages
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement Pass 2: Capture Validation**
  - [ ] For each target pattern, identify `{name}` segments
  - [ ] Check if `name` is a defined variable → treat as interpolation
  - [ ] Check if `name` is an automatic variable → error
  - [ ] Otherwise → treat as capture
  - [ ] Verify capture consistency between target and dependencies
  - [ ] Transform AST from BraceExpr to resolved Capture/Interpolation
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement Pass 3: Reference Validation**
  - [ ] For each interpolation in values/commands, verify it references a defined symbol
  - [ ] Check automatic variables are only used in recipe/block scope
  - [ ] Check captures are only referenced in their target's recipe
  - [ ] Check built-in variables (os, arch)
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement Pass 4: Dependency Graph Validation**
  - [ ] Build dependency graph from targets
  - [ ] Detect circular dependencies with cycle path reporting
  - [ ] Validate all dependencies can be satisfied (by explicit target or pattern)
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Define semantic error types**
  - [ ] DuplicateVariable, DuplicateTarget, DuplicateEnvironment
  - [ ] CaptureConflictsWithVariable, CaptureConflictsWithAutomatic
  - [ ] CaptureMismatch between target and dependencies
  - [ ] UndefinedVariable, AutomaticOutsideRecipe
  - [ ] CircularDependency with cycle path
  - [ ] InvalidDirectiveScope
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Write semantic analysis tests**
  - [ ] Test symbol collection and duplicate detection
  - [ ] Test capture vs interpolation resolution
  - [ ] Test automatic variable scope enforcement
  - [ ] Test circular dependency detection
  - [ ] Test comprehensive error messages
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

---

## Phase 2: Evaluation Engine (High Risk)

Variable evaluation, function execution, and conditional logic.

### 2.1 Variable Evaluation

- [ ] **Implement evaluation context**
  - [ ] Store evaluated immediate variables
  - [ ] Store unevaluated lazy variable ASTs
  - [ ] Store os and arch built-in values
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement immediate variable evaluation**
  - [ ] Evaluate variables in definition order
  - [ ] Handle forward references (error if immediate var references later immediate var)
  - [ ] Allow lazy variables to reference any other variable
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement value interpolation**
  - [ ] Substitute `{var}` with evaluated value
  - [ ] Handle `:raw` modifier (affects command execution, not evaluation)
  - [ ] Handle missing variable with clear error
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement lazy variable on-demand evaluation**
  - [ ] Detect when lazy variable is referenced
  - [ ] Evaluate at point of use with current context
  - [ ] Cache result? (spec unclear, decide and document)
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 2.2 Built-in Functions

- [ ] **Implement `shell()` function**
  - [ ] Execute shell command with default shell
  - [ ] Capture stdout, trim trailing newline
  - [ ] Handle command failure with error message
  - [ ] Apply shell quoting for interpolated values (default) vs raw
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement `glob()` function**
  - [ ] Parse glob pattern
  - [ ] Match files in filesystem
  - [ ] Return space-separated list of matches
  - [ ] Handle no matches (empty string or error?)
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement `basename()` function**
  - [ ] Extract filename without directory
  - [ ] Handle edge cases (trailing slash, no directory)
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement `dirname()` function**
  - [ ] Extract directory part of path
  - [ ] Handle edge cases (no directory, root)
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement `replace()` function**
  - [ ] Parse three arguments (input, from, to)
  - [ ] Replace all occurrences of `from` with `to`
  - [ ] Handle special characters in patterns
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Write function unit tests**
  - [ ] Test each function independently
  - [ ] Test function composition in values
  - [ ] Test error cases (shell failure, bad args)
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 2.3 Conditional Evaluation

- [ ] **Implement condition parsing during evaluation**
  - [ ] Evaluate left side of comparison
  - [ ] Compare with right side literal or evaluated value
  - [ ] Handle `==` and `!=` operators
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement `ifdef`/`ifndef` checks**
  - [ ] Check if variable name is in symbol table
  - [ ] Does NOT evaluate the variable, just checks existence
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement conditional branch execution**
  - [ ] Evaluate if condition, execute body if true
  - [ ] Otherwise try elif conditions in order
  - [ ] Finally execute else body if no match
  - [ ] Collect variable definitions from chosen branch
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Write conditional tests**
  - [ ] Test os/arch conditionals
  - [ ] Test ifdef/ifndef
  - [ ] Test nested conditionals
  - [ ] Test variable definitions inside conditionals
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

---

## Phase 3: Build Planning (High Risk)

Dependency resolution, staleness detection, and pattern matching.

### 3.1 Target Pattern Matching

- [ ] **Implement literal target matching**
  - [ ] Exact path comparison
  - [ ] Handle phony targets (always match regardless of file)
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement pattern target matching**
  - [ ] Match concrete path against pattern
  - [ ] Extract capture values (e.g., `{name}` → `"utils"`)
  - [ ] Handle multiple captures in single pattern
  - [ ] Handle patterns with variable interpolations already resolved
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement target lookup**
  - [ ] Given concrete path, find matching target definition
  - [ ] Prefer exact match over pattern match
  - [ ] Return captures if pattern match
  - [ ] Error if no match and not a source file
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Write pattern matching tests**
  - [ ] Test single capture patterns
  - [ ] Test multiple capture patterns
  - [ ] Test patterns with literal path segments
  - [ ] Test ambiguous patterns (multiple matches)
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 3.2 Dependency Resolution

- [ ] **Implement dependency path resolution**
  - [ ] For each dependency, resolve pattern with captures
  - [ ] Expand interpolations with evaluation context
  - [ ] Build list of concrete dependency paths
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement recursive dependency planning**
  - [ ] For requested target, find matching definition
  - [ ] For each dependency, recursively plan its build
  - [ ] Handle order-only dependencies (`.after:`)
  - [ ] Detect and report cycles during planning
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement topological sort**
  - [ ] Sort build tasks in dependency order
  - [ ] Tasks with no dependencies first
  - [ ] Identify independent tasks for parallelism
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 3.3 Staleness Detection

- [ ] **Implement file timestamp checking**
  - [ ] Get mtime for target and all dependencies
  - [ ] Handle missing target (always rebuild)
  - [ ] Handle missing dependency (error)
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement rebuild decision**
  - [ ] Phony targets always rebuild
  - [ ] Missing targets always rebuild
  - [ ] Rebuild if any dependency mtime > target mtime
  - [ ] Skip if target newer than all dependencies
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement order-only dependency handling**
  - [ ] `.after:` dependencies must exist
  - [ ] Their timestamps are NOT checked for staleness
  - [ ] Only used to ensure build order
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement autodeps support**
  - [ ] After successful build, parse `.d` file specified by `.autodeps:`
  - [ ] Store learned dependencies for future builds
  - [ ] Incorporate into staleness checking
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Build plan structure**
  - [ ] List of BuildTask in execution order
  - [ ] Each task has target, dependencies, recipe, and build reason
  - [ ] Build reason explains why rebuild needed
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Write planning tests**
  - [ ] Test simple dependency chains
  - [ ] Test diamond dependencies
  - [ ] Test phony targets
  - [ ] Test order-only dependencies
  - [ ] Test staleness detection logic
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

---

## Phase 4: Recipe Execution (Medium-High Risk)

Shell invocation, variable interpolation in commands, and error handling.

### 4.1 Command Interpolation

- [ ] **Implement automatic variable resolution**
  - [ ] `{target}` / `{out}` → target path
  - [ ] `{deps}` → space-separated dependency list
  - [ ] `{in}` → first dependency
  - [ ] `{stem}` → pattern capture (for pattern targets)
  - [ ] `{target.dir}` → directory part of target
  - [ ] `{target.file}` → filename part of target
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement capture resolution**
  - [ ] Resolve pattern captures from match (e.g., `{name}` → matched value)
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement shell quoting**
  - [ ] Default: shell-quote interpolated values (single quotes, escape embedded quotes)
  - [ ] With `:raw` modifier: no quoting, allows word splitting
  - [ ] Pass through `$var` to shell unchanged
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 4.2 Shell Execution

- [ ] **Implement line mode execution**
  - [ ] Each command line is separate shell invocation
  - [ ] Use global `.shell:` or recipe `.shell:` override
  - [ ] Execute via `shell -c "command"`
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement block mode execution**
  - [ ] All lines in `block:` passed as single script
  - [ ] Preserve internal structure (if/fi, loops, etc.)
  - [ ] Execute via `shell -c "script"`
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement shell selection**
  - [ ] Default: `/bin/sh`
  - [ ] Global override: `.shell: bash`
  - [ ] Recipe override: `.shell: zsh` (indented under target)
  - [ ] Verify shell exists before execution
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement dry-run mode**
  - [ ] Print commands without executing
  - [ ] Show interpolated values
  - [ ] Prefix with "Would build: target"
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement verbose mode**
  - [ ] Print commands before executing
  - [ ] Show variable evaluation results
  - [ ] Show staleness check decisions
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 4.3 Execution Error Handling

- [ ] **Handle command failure**
  - [ ] Check exit status of each command
  - [ ] Stop build on first failure (default)
  - [ ] Implement `--keep-going` flag to continue despite failures
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Handle missing dependencies**
  - [ ] If dependency can't be built and doesn't exist, error
  - [ ] Clear error message with dependency path
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Handle missing binaries**
  - [ ] If shell not found, error
  - [ ] If `.requires:` binary not found, suggest installation
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Write execution tests**
  - [ ] Test line mode execution
  - [ ] Test block mode execution
  - [ ] Test shell selection
  - [ ] Test dry-run output
  - [ ] Test error handling
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

---

## Phase 5: Parallel Execution (Medium Risk)

Concurrent task execution respecting dependencies.

### 5.1 Parallel Scheduler

- [ ] **Implement task queue**
  - [ ] Tasks ready when all dependencies complete
  - [ ] Track in-progress and completed tasks
  - [ ] Handle task completion events
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement worker pool**
  - [ ] Spawn N workers based on `.parallel:` or `-j` flag
  - [ ] Workers pull tasks from ready queue
  - [ ] Workers report completion or failure
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement dependency-aware scheduling**
  - [ ] Only schedule task when all dependencies satisfied
  - [ ] Update ready queue when task completes
  - [ ] Handle parallel diamond dependencies correctly
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement cancellation on failure**
  - [ ] On task failure, stop scheduling new tasks
  - [ ] Wait for in-progress tasks to complete
  - [ ] Report all failures
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Write parallel execution tests**
  - [ ] Test parallel independent tasks
  - [ ] Test parallel with dependencies
  - [ ] Test failure propagation
  - [ ] Test `-j` flag override
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

---

## Phase 6: Environment System (Medium Risk)

Runtime environments for build isolation.

### 6.1 Bare Environment

- [ ] **Implement requirements checking**
  - [ ] Parse `.requires:` list with version specs
  - [ ] Check if binaries exist in PATH
  - [ ] Check version if specified
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement version parsing**
  - [ ] `gcc` or `gcc@latest` → any version
  - [ ] `gcc@11` → major version 11.x.x
  - [ ] `gcc@11.4` → version 11.4.x
  - [ ] `gcc@11.4.0` → exact version
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement version detection**
  - [ ] Run `binary --version` and parse output
  - [ ] Handle different version output formats
  - [ ] Cache version checks
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement install suggestions**
  - [ ] For `--show-install` flag
  - [ ] Detect OS and suggest package manager command
  - [ ] Map binary names to package names
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 6.2 Container Environments (Docker/Podman)

- [ ] **Implement Dockerfile detection**
  - [ ] Locate file specified in `.source:`
  - [ ] Validate Dockerfile exists
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement image building**
  - [ ] Build Docker/Podman image from Dockerfile
  - [ ] Tag with project-specific name
  - [ ] Cache built images
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement container execution**
  - [ ] Run container with workspace mounted
  - [ ] Apply `.args:` (e.g., `--platform linux/amd64`)
  - [ ] Execute build commands inside container
  - [ ] Stream output to terminal
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement `--shell` flag**
  - [ ] Open interactive shell in container
  - [ ] Mount workspace
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement `--keep` flag**
  - [ ] Keep container running after build
  - [ ] Print instructions to enter/stop container
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 6.3 Devcontainer Environment

- [ ] **Implement devcontainer detection**
  - [ ] Check for `.devcontainer/` directory or `devcontainer.json`
  - [ ] Parse devcontainer configuration
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement devcontainer CLI integration**
  - [ ] Use `devcontainer` CLI to start environment
  - [ ] Execute build commands inside devcontainer
  - [ ] Handle lifecycle (start, stop)
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 6.4 Nix Environment

- [ ] **Implement nix file detection**
  - [ ] Locate `shell.nix` or `flake.nix` from `.source:`
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement nix-shell execution**
  - [ ] Enter nix-shell with specified configuration
  - [ ] Apply `.args:` (e.g., `--pure`)
  - [ ] Execute build commands inside nix environment
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 6.5 Lima Environment (macOS VMs)

- [ ] **Implement lima.yaml detection**
  - [ ] Locate lima configuration from `.source:`
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement Lima VM management**
  - [ ] Start Lima VM
  - [ ] Mount workspace
  - [ ] Execute build commands inside VM
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 6.6 Environment Selection Logic

- [ ] **Implement environment selection**
  - [ ] `--env name` flag takes precedence
  - [ ] `BUILD_ENV` environment variable second
  - [ ] Unnamed `.environment:` as default
  - [ ] Error if only named environments and no selection
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement `--list-env` flag**
  - [ ] List all defined environments
  - [ ] Show runtime type and source for each
  - [ ] Mark default environment
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement `--check-env` flag**
  - [ ] Verify environment requirements are met
  - [ ] For containers, verify Dockerfile/runtime exists
  - [ ] Print status for each requirement
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Write environment tests**
  - [ ] Test bare environment requirements checking
  - [ ] Test container environment (requires Docker/Podman)
  - [ ] Test environment selection logic
  - [ ] Test --list-env and --check-env output
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

---

## Phase 7: CLI Interface (Low-Medium Risk)

Command-line argument parsing and user interaction.

### 7.1 Argument Parsing

- [ ] **Implement target argument parsing**
  - [ ] No argument → use `.default:` or first target
  - [ ] `build target` → build specific file target
  - [ ] `build @phony` → build phony target
  - [ ] Multiple targets → build all in order
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement option flags**
  - [ ] `--env` / `-e` → environment selection
  - [ ] `--dry-run` / `-n` → show what would execute
  - [ ] `--verbose` / `-v` → verbose output
  - [ ] `--jobs N` / `-j N` → parallel jobs
  - [ ] `--file path` / `-f path` → alternate Buildfile
  - [ ] `--check-env` → verify environment
  - [ ] `--show-install` → show install instructions
  - [ ] `--list-env` → list environments
  - [ ] `--shell` → open shell in environment
  - [ ] `--keep` → keep sandbox running
  - [ ] `--help` / `-h` → show help
  - [ ] `--version` / `-V` → show version
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement Buildfile discovery**
  - [ ] Look for `Buildfile` in current directory
  - [ ] Look in parent directories up to root
  - [ ] Respect `-f` flag override
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 7.2 Output Formatting

- [ ] **Implement normal output**
  - [ ] Show target being built
  - [ ] Show command output
  - [ ] Show completion/failure status
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement dry-run output**
  - [ ] "Would build: target"
  - [ ] Show each command that would execute
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement verbose output**
  - [ ] Show variable evaluation
  - [ ] Show staleness check results
  - [ ] Show dependency resolution
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement progress for parallel builds**
  - [ ] Show currently building targets
  - [ ] Show completion count (e.g., [3/10])
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 7.3 Exit Codes

- [ ] **Define exit codes**
  - [ ] 0 → success
  - [ ] 1 → build failure (recipe returned non-zero)
  - [ ] 2 → usage error (bad arguments)
  - [ ] 3 → parse error (invalid Buildfile)
  - [ ] 4 → environment error (missing requirements)
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

---

## Phase 8: Error Reporting (Low-Medium Risk)

User-friendly, actionable error messages.

### 8.1 Error Message Format

- [ ] **Implement source location tracking**
  - [ ] All AST nodes carry SourceLocation
  - [ ] Include file path, line number, column number
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement error message template**
  - [ ] Error code (E001, E100, etc.)
  - [ ] Brief description
  - [ ] Source context with line numbers
  - [ ] Pointer to error location (^^^)
  - [ ] "note:" for additional context
  - [ ] "help:" for fix suggestions
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement source snippet extraction**
  - [ ] Read relevant lines from source file
  - [ ] Show 1-3 lines of context around error
  - [ ] Highlight error location with carets
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 8.2 Error Categories

- [ ] **Implement lexical errors (E001-E099)**
  - [ ] Invalid character
  - [ ] Bad indentation (mixed tabs/spaces)
  - [ ] Unterminated interpolation
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement syntax errors (E100-E199)**
  - [ ] Unexpected token
  - [ ] Missing colon in target
  - [ ] Missing `end` for conditional
  - [ ] Invalid directive at scope
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement semantic errors (E200-E299)**
  - [ ] Undefined variable
  - [ ] Duplicate definition
  - [ ] Circular dependency
  - [ ] Capture conflicts
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement evaluation errors (E300-E399)**
  - [ ] Shell command failed
  - [ ] Glob matched nothing
  - [ ] Invalid function arguments
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [ ] **Implement execution errors (E400-E499)**
  - [ ] Recipe failed
  - [ ] Missing dependency file
  - [ ] Missing binary
  - [ ] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

---

## Phase 9: Testing and Quality (Low Risk)

Comprehensive test coverage and code quality.

### 9.1 Unit Tests

- [ ] **Lexer unit tests** (covered in Phase 1)
- [ ] **Parser unit tests** (covered in Phase 1)
- [ ] **Semantic analyzer unit tests** (covered in Phase 1)
- [ ] **Evaluator unit tests** (covered in Phase 2)
- [ ] **Planner unit tests** (covered in Phase 3)
- [ ] **Executor unit tests** (covered in Phase 4)

### 9.2 Integration Tests

- [ ] **End-to-end build tests**
  - [ ] Simple C compilation example
  - [ ] Multi-file project with patterns
  - [ ] Conditional compilation flags
  - [ ] Phony targets

- [ ] **Include file tests**
  - [ ] Single include
  - [ ] Nested includes
  - [ ] Circular include detection

- [ ] **Environment tests**
  - [ ] Bare environment with requirements
  - [ ] Docker environment (CI only, requires Docker)

### 9.3 Regression Tests

- [ ] **Create test fixture directory**
  - [ ] Valid Buildfile examples from spec
  - [ ] Invalid Buildfiles with expected errors
  - [ ] Complex real-world-like examples

- [ ] **Implement test harness**
  - [ ] Run build tool on fixtures
  - [ ] Compare output to expected
  - [ ] Verify file creation/modification

### 9.4 Performance Tests

- [ ] **Large Buildfile parsing**
  - [ ] 1000+ targets
  - [ ] Deep include hierarchies
  - [ ] Many pattern targets

- [ ] **Large build planning**
  - [ ] Deep dependency chains
  - [ ] Wide dependency graphs

### 9.5 Documentation

- [ ] **Write README**
  - [ ] Installation instructions
  - [ ] Quick start guide
  - [ ] Link to spec

- [ ] **Write man page or --help text**
  - [ ] All command-line options
  - [ ] Examples for common tasks

- [ ] **Write migration guide**
  - [ ] Make to Buildfile translation
  - [ ] Common patterns

---

## Phase 10: Polish and Optimization (Lowest Risk)

Performance improvements and user experience polish.

### 10.1 Performance Optimization

- [ ] **Cache parsed Buildfiles**
  - [ ] Store parsed AST between runs
  - [ ] Invalidate on file modification

- [ ] **Cache autodeps**
  - [ ] Store learned dependencies persistently
  - [ ] Load on subsequent builds

- [ ] **Optimize pattern matching**
  - [ ] Pre-compile patterns for fast matching
  - [ ] Index targets by prefix

- [ ] **Lazy shell() execution**
  - [ ] Defer shell() calls until value is needed
  - [ ] Cache shell() results within build

### 10.2 User Experience

- [ ] **Colored output**
  - [ ] Red for errors
  - [ ] Yellow for warnings
  - [ ] Green for success
  - [ ] Detect TTY for automatic disable

- [ ] **Progress indication**
  - [ ] Spinner for long operations
  - [ ] Progress bar for parallel builds

- [ ] **Tab completion**
  - [ ] Bash completion script
  - [ ] Zsh completion script
  - [ ] Fish completion script

### 10.3 Platform Support

- [ ] **Linux support** (primary platform)
- [ ] **macOS support** (including Lima for VMs)
- [ ] **Windows support** (PowerShell? WSL?)
  - [ ] Path separator handling
  - [ ] Shell selection (cmd.exe? PowerShell? bash via WSL?)

---

## Summary: Risk-Ordered Implementation Path

| Phase | Risk Level | Components |
|-------|------------|------------|
| 1 | **Highest** | Lexer, Parser, Semantic Analysis |
| 2 | **High** | Variable Evaluation, Functions, Conditionals |
| 3 | **High** | Pattern Matching, Dependency Resolution, Planning |
| 4 | **Medium-High** | Recipe Execution, Shell Invocation |
| 5 | **Medium** | Parallel Execution |
| 6 | **Medium** | Environment System |
| 7 | **Low-Medium** | CLI Interface |
| 8 | **Low-Medium** | Error Reporting |
| 9 | **Low** | Testing and Quality |
| 10 | **Lowest** | Polish and Optimization |

**Recommended approach**: Complete each phase fully before moving to the next. Early phases block later work and their design impacts everything downstream.
