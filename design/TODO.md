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
  - [x] Prefix with "Would build: target"
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement verbose mode**
  - [x] Print commands before executing
  - [x] Show variable evaluation results
  - [x] Show staleness check decisions
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 4.3 Execution Error Handling

- [x] **Handle command failure**
  - [x] Check exit status of each command
  - [x] Stop build on first failure (default)
  - [x] Implement `--keep-going` flag to continue despite failures
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
  - [x] Test `-j` flag override
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
  - [x] Cache version checks
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

- [x] **Implement devcontainer detection**
  - [x] Check for `.devcontainer/` directory or `devcontainer.json`
  - [x] Parse devcontainer configuration
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement devcontainer CLI integration**
  - [x] Use `devcontainer` CLI to start environment
  - [x] Execute build commands inside devcontainer
  - [x] Handle lifecycle (start, stop)
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 6.4 Nix Environment

- [x] **Implement nix file detection**
  - [x] Locate `shell.nix` or `flake.nix` from `.source:`
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement nix-shell execution**
  - [x] Enter nix-shell with specified configuration
  - [x] Apply `.args:` (e.g., `--pure`)
  - [x] Execute build commands inside nix environment
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 6.5 Lima Environment (macOS VMs)

- [x] **Implement lima.yaml detection**
  - [x] Locate lima configuration from `.source:`
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement Lima VM management**
  - [x] Start Lima VM
  - [x] Mount workspace
  - [x] Execute build commands inside VM
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 6.6 Environment Selection Logic

- [x] **Implement environment selection**
  - [x] `--env name` flag takes precedence
  - [x] `BUILD_ENV` environment variable second
  - [x] Unnamed `.environment:` as default
  - [x] Error if only named environments and no selection
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement `--list-env` flag**
  - [x] List all defined environments
  - [x] Show runtime type and source for each
  - [x] Mark default environment
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement `--check-env` flag**
  - [x] Verify environment requirements are met
  - [x] For containers, verify Dockerfile/runtime exists
  - [x] Print status for each requirement
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Write environment tests**
  - [x] Test bare environment requirements checking
  - [x] Test container environment (requires Docker/Podman)
  - [x] Test environment selection logic
  - [x] Test --list-env and --check-env output
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

---

## Phase 7: CLI Interface (Low-Medium Risk)

Command-line argument parsing and user interaction.

### 7.1 Argument Parsing

- [x] **Implement target argument parsing**
  - [x] No argument → use `.default:` or first target
  - [x] `build target` → build specific file target
  - [x] `build phony` → build phony target. User need not provide `@` to indicate phonyness.
  - [x] Multiple targets → build all in order
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

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

- [x] **Implement normal output**
  - [x] Show target being built
  - [x] Show command output
  - [x] Show completion/failure status
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement dry-run output**
  - [x] "Would build: target"
  - [x] Show each command that would execute
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement verbose output**
  - [x] Show variable evaluation
  - [x] Show staleness check results
  - [x] Show dependency resolution
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement progress for parallel builds**
  - [x] Show currently building targets
  - [x] Show completion count (e.g., [3/10])
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

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

- [x] **Implement source location tracking**
  - [x] All AST nodes carry SourceLocation
  - [x] Include file path, line number, column number
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement error message template**
  - [x] Error code (E001, E100, etc.)
  - [x] Brief description
  - [x] Source context with line numbers
  - [x] Pointer to error location (^^^)
  - [x] "note:" for additional context
  - [x] "help:" for fix suggestions
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement source snippet extraction**
  - [x] Read relevant lines from source file
  - [x] Show 1-3 lines of context around error
  - [x] Highlight error location with carets
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

### 8.2 Error Categories

- [x] **Implement lexical errors (E001-E099)**
  - [x] Invalid character
  - [x] Bad indentation (mixed tabs/spaces)
  - [x] Unterminated interpolation
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement syntax errors (E100-E199)**
  - [x] Unexpected token
  - [x] Missing colon in target
  - [x] Missing `end` for conditional
  - [x] Invalid directive at scope
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement semantic errors (E200-E299)**
  - [x] Undefined variable
  - [x] Duplicate definition
  - [x] Circular dependency
  - [x] Capture conflicts
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement evaluation errors (E300-E399)**
  - [x] Shell command failed
  - [x] Glob matched nothing
  - [x] Invalid function arguments
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

- [x] **Implement execution errors (E400-E499)**
  - [x] Recipe failed
  - [x] Missing dependency file
  - [x] Missing binary
  - [x] Update `cmd/build` CLI to incorporate new feature including updates to all test cases of the CLI.

---

## Phase 9: Testing and Quality (Low Risk)

Comprehensive test coverage and code quality.

### 9.1 Unit Tests

- [x] **Lexer unit tests** (covered in Phase 1)
- [x] **Parser unit tests** (covered in Phase 1)
- [x] **Semantic analyzer unit tests** (covered in Phase 1)
- [x] **Evaluator unit tests** (covered in Phase 2)
- [x] **Planner unit tests** (covered in Phase 3)
- [x] **Executor unit tests** (covered in Phase 4)

### 9.2 Integration Tests

- [x] **End-to-end build tests**
  - [x] Simple C compilation example
  - [x] Multi-file project with patterns
  - [x] Conditional compilation flags
  - [x] Phony targets

- [x] **Include file tests**
  - [x] Single include
  - [x] Nested includes
  - [x] Circular include detection

- [x] **Environment tests**
  - [x] Bare environment with requirements
  - [x] Docker environment (CI only, requires Docker)

### 9.3 Regression Tests

- [x] **Create test fixture directory**
  - [x] Valid Buildfile examples from spec
  - [x] Invalid Buildfiles with expected errors
  - [x] Complex real-world-like examples

- [x] **Implement test harness**
  - [x] Run build tool on fixtures
  - [x] Compare output to expected
  - [x] Verify file creation/modification

### 9.4 Performance Tests

- [x] **Large Buildfile parsing**
  - [x] 1000+ targets
  - [x] Deep include hierarchies
  - [x] Many pattern targets

- [x] **Large build planning**
  - [x] Deep dependency chains
  - [x] Wide dependency graphs

### 9.5 Documentation

- [x] **Write README**
  - [x] Installation instructions
  - [x] Quick start guide
  - [x] Link to spec

- [x] **Write man page or --help text**
  - [x] All command-line options
  - [x] Examples for common tasks

- [x] **Write migration guide**
  - [x] Make to Buildfile translation
  - [x] Common patterns

---

## Phase 10: Polish and Optimization (Lowest Risk)

Performance improvements and user experience polish.

### 10.1 Performance Optimization

- [x] **Cache parsed Buildfiles**
  - [x] Store parsed AST between runs
  - [x] Invalidate on file modification

- [x] **Cache autodeps**
  - [x] Store learned dependencies persistently
  - [x] Load on subsequent builds

- [x] **Optimize pattern matching**
  - [x] Pre-compile patterns for fast matching
  - [x] Index targets by prefix

- [x] **Lazy shell() execution**
  - [x] Defer shell() calls until value is needed
  - [x] Cache shell() results within build

### 10.2 Output Beautification System

The output system supports three contexts: CLI (interactive), TUI (structured), and Headless (CI/logs).
See DESIGN.md Section 10 for full architecture.

- [x] **Implement output event types**
  - [x] Define OutputEvent interface and all event types (PhaseStarted, TargetStarted, etc.)
  - [x] Add timestamps and durations to relevant events
  - [x] Add structured error events with code/location/context/hint

- [x] **Implement OutputMode detection**
  - [x] Create OutputMode enum (CLI, TUI, Headless)
  - [x] Detect mode from TTY status and environment variables
  - [x] Support BUILD_OUTPUT_MODE override
  - [x] Support CI environment detection (GITHUB_ACTIONS, GITLAB_CI, etc.)

- [x] **Implement OutputWriter interface**
  - [x] Define WriteEvent(OutputEvent) method
  - [x] Define Flush() method for buffered output
  - [x] Create factory function to instantiate correct writer

- [x] **Implement terminal capability detection**
  - [x] Query terminal width/height
  - [x] Detect color support (0, 16, 256, truecolor)
  - [x] Detect unicode support
  - [x] Handle TERM=dumb and NO_COLOR

- [x] **Implement ANSI color utilities**
  - [x] Define color constants (Red, Green, Yellow, Cyan, etc.)
  - [x] Create color application functions
  - [x] Support bold, dim, and reset
  - [x] Respect NO_COLOR and FORCE_COLOR environment variables

- [x] **Implement CLIWriter (interactive terminal)**
  - [x] Colored output with proper ANSI codes
  - [x] Progress formatting for parallel builds [n/total]
  - [x] Command output with proper indentation
  - [x] Error display with source context and hints
  - [x] Degraded output for limited terminals

- [x] **Implement HeadlessWriter (CI/logs)**
  - [x] Timestamped log lines
  - [x] Log levels (DEBUG, INFO, WARN, ERROR)
  - [x] Plain text without escape sequences
  - [x] Optional JSON log format (BUILD_LOG_FORMAT=json)

- [x] **Implement TUIWriter (structured output)**
  - [x] JSON event stream output
  - [x] Machine-parseable format
  - [x] Timestamps on all events

- [x] **Integrate output system with build pipeline**
  - [x] Emit events from executor (target/command lifecycle)
  - [x] Emit events from evaluator (variable evaluation in verbose)
  - [x] Emit events from planner (staleness checks in verbose)
  - [x] Emit error events from all stages
  - [x] Emit summary events at completion

- [x] **Add CLI flags for output control**
  - [x] Add --quiet / -q flag to suppress non-error output
  - [x] Add --color=auto|always|never flag
  - [x] Add --progress=auto|always|never flag
  - [x] Wire flags to output writer selection

- [x] **Refactor existing Reporter to use new system**
  - [x] Migrate NormalReporter to CLIWriter
  - [x] Migrate VerboseReporter to CLIWriter (verbose mode)
  - [x] Migrate DryRunReporter to use OutputWriter
  - [x] Migrate ProgressReporter to CLIWriter (parallel mode)

### 10.3 Tab Completion

- [x] **Tab completion**
  - [x] Bash completion script
  - [x] Zsh completion script
  - [x] Fish completion script

### 10.4 Platform Support

- [x] **Linux support** (primary platform)
- [x] **macOS support** (including Lima for VMs)
- [x] **Windows support** (PowerShell? WSL?)
  - [x] Path separator handling
  - [x] Shell selection (cmd.exe? PowerShell? bash via WSL?)

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
