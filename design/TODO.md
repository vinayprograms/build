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

- [x] **Implement parser scope stack**
  - [x] Define Scope enum (GLOBAL, ENVIRONMENT, RECIPE, BLOCK)
  - [x] Push/pop scope on block entry/exit
  - [x] Validate directive placement based on current scope
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement variable parsing**
  - [x] Detect `=` before `:` to distinguish from targets
  - [x] Parse `lazy` keyword prefix
  - [x] Parse right-hand side as Value with interpolations and function calls
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement target parsing**
  - [x] Parse target pattern (file path with `{name}` segments)
  - [x] Parse phony targets (`@name`)
  - [x] Parse directory targets (ending with `/`)
  - [x] Parse dependency list
  - [x] Handle pattern targets with captures
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement recipe parsing**
  - [x] Detect recipe start by indentation after target
  - [x] Parse recipe directives (.shell, .after, .autodeps, .requires)
  - [x] Parse command lines with interpolations
  - [x] Parse `block:` with deeper indentation
  - [x] Handle dedent to end recipe
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement environment block parsing**
  - [x] Parse `.environment:` with optional name
  - [x] Enter ENVIRONMENT scope
  - [x] Parse environment directives (.using, .source, .args, .requires)
  - [x] Handle dedent to end environment block
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement conditional parsing**
  - [x] Parse `if`, `elif`, `else`, `end` keywords
  - [x] Parse conditions (`{var} == value`, `{var} != value`)
  - [x] Parse `ifdef` and `ifndef` variants
  - [x] Collect body statements for each branch
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement value parsing**
  - [x] Parse literal strings
  - [x] Parse interpolations (`{var}`, `{var:raw}`)
  - [x] Parse function calls (shell, glob, basename, dirname, replace)
  - [x] Handle nested parentheses in function arguments
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement `.include:` directive**
  - [x] Parse include path with variable interpolation
  - [x] Recursively lex/parse included file
  - [x] Merge included AST into parent
  - [x] Detect and prevent circular includes
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement error recovery**
  - [x] On parse error, skip to next line at indentation level 0
  - [x] Collect multiple errors before failing
  - [x] Provide actionable error messages with source locations
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Write parser unit tests**
  - [x] Test all statement types
  - [x] Test scope validation for directives
  - [x] Test nested blocks (recipe → block)
  - [x] Test conditionals with all branch combinations
  - [x] Test include directive
  - [x] Test error recovery and messages
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 1.3 Semantic Analysis

Validates the AST and resolves ambiguous constructs.

- [x] **Implement symbol table**
  - [x] Define SymbolTable structure (variables, targets, environments, automatic)
  - [x] Populate automatic variable set (target, deps, in, out, stem, target.dir, target.file)
  - [x] Populate built-in variable set (os, arch)
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement Pass 1: Symbol Collection**
  - [x] Collect all variable definitions
  - [x] Collect all target definitions
  - [x] Collect all environment definitions
  - [x] Detect duplicate definitions with clear error messages
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement Pass 2: Capture Validation**
  - [x] For each target pattern, identify `{name}` segments
  - [x] Check if `name` is a defined variable → treat as interpolation
  - [x] Check if `name` is an automatic variable → error
  - [x] Otherwise → treat as capture
  - [x] Verify capture consistency between target and dependencies
  - [x] Transform AST from BraceExpr to resolved Capture/Interpolation
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement Pass 3: Reference Validation**
  - [x] For each interpolation in values/commands, verify it references a defined symbol
  - [x] Check automatic variables are only used in recipe/block scope
  - [x] Check captures are only referenced in their target's recipe
  - [x] Check built-in variables (os, arch)
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement Pass 4: Dependency Graph Validation**
  - [x] Build dependency graph from targets
  - [x] Detect circular dependencies with cycle path reporting
  - [x] Validate all dependencies can be satisfied (by explicit target or pattern)
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Define semantic error types**
  - [x] DuplicateVariable, DuplicateTarget, DuplicateEnvironment
  - [x] CaptureConflictsWithVariable, CaptureConflictsWithAutomatic
  - [x] CaptureMismatch between target and dependencies
  - [x] UndefinedVariable, AutomaticOutsideRecipe
  - [x] CircularDependency with cycle path
  - [x] InvalidDirectiveScope
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Write semantic analysis tests**
  - [x] Test symbol collection and duplicate detection
  - [x] Test capture vs interpolation resolution
  - [x] Test automatic variable scope enforcement
  - [x] Test circular dependency detection
  - [x] Test comprehensive error messages
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

---

## Phase 2: Evaluation Engine (High Risk)

Variable evaluation, function execution, and conditional logic.

### 2.1 Variable Evaluation

- [x] **Implement evaluation context**
  - [x] Store evaluated immediate variables
  - [x] Store unevaluated lazy variable ASTs
  - [x] Store os and arch built-in values
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement immediate variable evaluation**
  - [x] Evaluate variables in definition order
  - [x] Handle forward references (error if immediate var references later immediate var)
  - [x] Allow lazy variables to reference any other variable
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement value interpolation**
  - [x] Substitute `{var}` with evaluated value
  - [x] Handle `:raw` modifier (affects command execution, not evaluation)
  - [x] Handle missing variable with clear error
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement lazy variable on-demand evaluation**
  - [x] Detect when lazy variable is referenced
  - [x] Evaluate at point of use with current context
  - [x] Cache result? (spec unclear, decide and document)
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 2.2 Built-in Functions

- [x] **Implement `shell()` function**
  - [x] Execute shell command with default shell
  - [x] Capture stdout, trim trailing newline
  - [x] Handle command failure with error message
  - [x] Apply shell quoting for interpolated values (default) vs raw
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement `glob()` function**
  - [x] Parse glob pattern
  - [x] Match files in filesystem
  - [x] Return space-separated list of matches
  - [x] Handle no matches (empty string or error?)
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement `basename()` function**
  - [x] Extract filename without directory
  - [x] Handle edge cases (trailing slash, no directory)
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement `dirname()` function**
  - [x] Extract directory part of path
  - [x] Handle edge cases (no directory, root)
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement `replace()` function**
  - [x] Parse three arguments (input, from, to)
  - [x] Replace all occurrences of `from` with `to`
  - [x] Handle special characters in patterns
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Write function unit tests**
  - [x] Test each function independently
  - [x] Test function composition in values
  - [x] Test error cases (shell failure, bad args)
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 2.3 Conditional Evaluation

- [x] **Implement condition parsing during evaluation**
  - [x] Evaluate left side of comparison
  - [x] Compare with right side literal or evaluated value
  - [x] Handle `==` and `!=` operators
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement `ifdef`/`ifndef` checks**
  - [x] Check if variable name is in symbol table
  - [x] Does NOT evaluate the variable, just checks existence
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement conditional branch execution**
  - [x] Evaluate if condition, execute body if true
  - [x] Otherwise try elif conditions in order
  - [x] Finally execute else body if no match
  - [x] Collect variable definitions from chosen branch
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Write conditional tests**
  - [x] Test os/arch conditionals
  - [x] Test ifdef/ifndef
  - [x] Test nested conditionals
  - [x] Test variable definitions inside conditionals
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

---

## Phase 3: Build Planning (High Risk)

Dependency resolution, staleness detection, and pattern matching.

### 3.1 Target Pattern Matching

- [x] **Implement literal target matching**
  - [x] Exact path comparison
  - [x] Handle phony targets (always match regardless of file)
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement pattern target matching**
  - [x] Match concrete path against pattern
  - [x] Extract capture values (e.g., `{name}` → `"utils"`)
  - [x] Handle multiple captures in single pattern
  - [x] Handle patterns with variable interpolations already resolved
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement target lookup**
  - [x] Given concrete path, find matching target definition
  - [x] Prefer exact match over pattern match
  - [x] Return captures if pattern match
  - [x] Error if no match and not a source file
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Write pattern matching tests**
  - [x] Test single capture patterns
  - [x] Test multiple capture patterns
  - [x] Test patterns with literal path segments
  - [x] Test ambiguous patterns (multiple matches)
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 3.2 Dependency Resolution

- [x] **Implement dependency path resolution**
  - [x] For each dependency, resolve pattern with captures
  - [x] Expand interpolations with evaluation context
  - [x] Build list of concrete dependency paths
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement recursive dependency planning**
  - [x] For requested target, find matching definition
  - [x] For each dependency, recursively plan its build
  - [x] Handle order-only dependencies (`.after:`)
  - [x] Detect and report cycles during planning
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement topological sort**
  - [x] Sort build tasks in dependency order
  - [x] Tasks with no dependencies first
  - [x] Identify independent tasks for parallelism
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 3.3 Staleness Detection

- [x] **Implement file timestamp checking**
  - [x] Get mtime for target and all dependencies
  - [x] Handle missing target (always rebuild)
  - [x] Handle missing dependency (error)
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement rebuild decision**
  - [x] Phony targets always rebuild
  - [x] Missing targets always rebuild
  - [x] Rebuild if any dependency mtime > target mtime
  - [x] Skip if target newer than all dependencies
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement order-only dependency handling**
  - [x] `.after:` dependencies must exist
  - [x] Their timestamps are NOT checked for staleness
  - [x] Only used to ensure build order
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement autodeps support**
  - [x] After successful build, parse `.d` file specified by `.autodeps:`
  - [x] Store learned dependencies for future builds
  - [x] Incorporate into staleness checking
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Build plan structure**
  - [x] List of BuildTask in execution order
  - [x] Each task has target, dependencies, recipe, and build reason
  - [x] Build reason explains why rebuild needed
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Write planning tests**
  - [x] Test simple dependency chains
  - [x] Test diamond dependencies
  - [x] Test phony targets
  - [x] Test order-only dependencies
  - [x] Test staleness detection logic
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

---

## Phase 4: Recipe Execution (Medium-High Risk)

Shell invocation, variable interpolation in commands, and error handling.

### 4.1 Command Interpolation

- [x] **Implement automatic variable resolution**
  - [x] `{target}` / `{out}` → target path
  - [x] `{deps}` → space-separated dependency list
  - [x] `{in}` → first dependency
  - [x] `{stem}` → pattern capture (for pattern targets)
  - [x] `{target.dir}` → directory part of target
  - [x] `{target.file}` → filename part of target
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement capture resolution**
  - [x] Resolve pattern captures from match (e.g., `{name}` → matched value)
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement shell quoting**
  - [x] Default: shell-quote interpolated values (single quotes, escape embedded quotes)
  - [x] With `:raw` modifier: no quoting, allows word splitting
  - [x] Pass through `$var` to shell unchanged
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 4.2 Shell Execution

- [x] **Implement line mode execution**
  - [x] Each command line is separate shell invocation
  - [x] Use global `.shell:` or recipe `.shell:` override
  - [x] Execute via `shell -c "command"`
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement block mode execution**
  - [x] All lines in `block:` passed as single script
  - [x] Preserve internal structure (if/fi, loops, etc.)
  - [x] Execute via `shell -c "script"`
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement shell selection**
  - [x] Default: `/bin/sh`
  - [x] Global override: `.shell: bash`
  - [x] Recipe override: `.shell: zsh` (indented under target)
  - [x] Verify shell exists before execution
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement dry-run mode**
  - [x] Print commands without executing
  - [x] Show interpolated values
  - [ ] Prefix with "Would build: target"
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement verbose mode**
  - [x] Print commands before executing
  - [ ] Show variable evaluation results
  - [ ] Show staleness check decisions
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 4.3 Execution Error Handling

- [x] **Handle command failure**
  - [x] Check exit status of each command
  - [x] Stop build on first failure (default)
  - [ ] Implement `--keep-going` flag to continue despite failures
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Handle missing dependencies**
  - [x] If dependency can't be built and doesn't exist, error
  - [x] Clear error message with dependency path
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Handle missing binaries**
  - [x] If shell not found, error
  - [x] If `.requires:` binary not found, suggest installation
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Write execution tests**
  - [x] Test line mode execution
  - [x] Test block mode execution
  - [x] Test shell selection
  - [x] Test dry-run output
  - [x] Test error handling
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

---

## Phase 5: Parallel Execution (Medium Risk)

Concurrent task execution respecting dependencies.

### 5.1 Parallel Scheduler

- [x] **Implement task queue**
  - [x] Tasks ready when all dependencies complete
  - [x] Track in-progress and completed tasks
  - [x] Handle task completion events
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement worker pool**
  - [x] Spawn N workers based on `.parallel:` or `-j` flag
  - [x] Workers pull tasks from ready queue
  - [x] Workers report completion or failure
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement dependency-aware scheduling**
  - [x] Only schedule task when all dependencies satisfied
  - [x] Update ready queue when task completes
  - [x] Handle parallel diamond dependencies correctly
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement cancellation on failure**
  - [x] On task failure, stop scheduling new tasks
  - [x] Wait for in-progress tasks to complete
  - [x] Report all failures
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Write parallel execution tests**
  - [x] Test parallel independent tasks
  - [x] Test parallel with dependencies
  - [x] Test failure propagation
  - [ ] Test `-j` flag override
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

---

## Phase 6: Environment System (Medium Risk)

Runtime environments for build isolation.

### 6.1 Bare Environment

- [x] **Implement requirements checking**
  - [x] Parse `.requires:` list with version specs
  - [x] Check if binaries exist in PATH
  - [x] Check version if specified
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement version parsing**
  - [x] `gcc` or `gcc@latest` → any version
  - [x] `gcc@11` → major version 11.x.x
  - [x] `gcc@11.4` → version 11.4.x
  - [x] `gcc@11.4.0` → exact version
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement version detection**
  - [x] Run `binary --version` and parse output
  - [x] Handle different version output formats
  - [ ] Cache version checks
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement install suggestions**
  - [x] For `--show-install` flag
  - [x] Detect OS and suggest package manager command
  - [x] Map binary names to package names
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 6.2 Container Environments (Docker/Podman)

- [x] **Implement Dockerfile detection**
  - [x] Locate file specified in `.source:`
  - [x] Validate Dockerfile exists
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement image building**
  - [x] Build Docker/Podman image from Dockerfile
  - [x] Tag with project-specific name
  - [x] Cache built images
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement container execution**
  - [x] Run container with workspace mounted
  - [x] Apply `.args:` (e.g., `--platform linux/amd64`)
  - [x] Execute build commands inside container
  - [x] Stream output to terminal
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement `--shell` flag**
  - [x] Open interactive shell in container
  - [x] Mount workspace
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement `--keep` flag**
  - [x] Keep container running after build
  - [x] Print instructions to enter/stop container
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

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

- [x] **Implement `--list-env` flag**
  - [x] List all defined environments
  - [x] Show runtime type and source for each
  - [x] Mark default environment
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement `--check-env` flag**
  - [x] Verify environment requirements are met
  - [ ] For containers, verify Dockerfile/runtime exists
  - [x] Print status for each requirement
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Write environment tests**
  - [x] Test bare environment requirements checking
  - [ ] Test container environment (requires Docker/Podman)
  - [x] Test environment selection logic
  - [x] Test --list-env and --check-env output
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

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

- [x] **Implement option flags**
  - [x] `--env` / `-e` → environment selection
  - [x] `--dry-run` / `-n` → show what would execute
  - [x] `--verbose` / `-v` → verbose output
  - [x] `--jobs N` / `-j N` → parallel jobs
  - [x] `--file path` / `-f path` → alternate Buildfile
  - [x] `--check-env` → verify environment
  - [x] `--show-install` → show install instructions
  - [x] `--list-env` → list environments
  - [x] `--shell` → open shell in environment
  - [x] `--keep` → keep sandbox running
  - [x] `--help` / `-h` → show help
  - [x] `--version` / `-V` → show version
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement Buildfile discovery**
  - [x] Look for `Buildfile` in current directory
  - [x] Look in parent directories up to root
  - [x] Respect `-f` flag override
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

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

- [x] **Define exit codes**
  - [x] 0 → success
  - [x] 1 → build failure (recipe returned non-zero)
  - [x] 2 → usage error (bad arguments)
  - [x] 3 → parse error (invalid Buildfile)
  - [x] 4 → environment error (missing requirements)
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

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
