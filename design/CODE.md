# need - Code Design

This document tracks the implementation architecture and design decisions as code is written.

## Package Structure

```
github.com/vinayprograms/need/
├── cmd/
│   └── build/          # CLI entry point
│       ├── main.go
│       ├── main_test.go
│       ├── debug.go            # Debug command implementations
│       ├── interfaces.go       # Interface definitions
│       ├── lexer_adapter.go    # Lexer adapters
│       ├── parser_adapter.go   # Parser adapters
│       └── semantic_adapter.go # Semantic adapters
├── internal/
│   ├── ast/            # Abstract Syntax Tree
│   │   ├── doc.go         # Package documentation
│   │   ├── location.go     # SourceLocation type
│   │   ├── statement.go    # Needfile, Statement, Comment, Blank
│   │   ├── directives.go   # Directive, DirectiveKind
│   │   ├── environment.go  # Environment, Runtime, VersionSpec types
│   │   ├── variables.go    # Variable
│   │   ├── conditionals.go # Condition types, Conditional
│   │   ├── targets.go      # Target, TargetPattern, Dependency, BraceExpr
│   │   ├── recipes.go      # Recipe, Command types
│   │   ├── values.go       # Value, ValuePart, FunctionCall
│   │   └── ast_test.go
│   ├── lexer/          # Lexical analysis
│   │   ├── doc.go         # Package documentation
│   │   ├── indent.go       # Indentation tracking
│   │   ├── indent_test.go
│   │   ├── interp.go       # Interpolation boundary detection
│   │   ├── interp_test.go
│   │   ├── lexer.go        # Core lexer state machine
│   │   ├── lexer_test.go
│   │   ├── string.go       # String/path/identifier lexing
│   │   ├── token.go        # Token types and source location
│   │   ├── token_test.go
│   │   └── value.go        # Value/command mode lexing
│   ├── parser/         # Syntactic analysis
│   │   ├── doc.go         # Package documentation
│   │   ├── command.go      # Command/block parsing
│   │   ├── conditional.go  # Conditional parsing
│   │   ├── conditional_test.go
│   │   ├── directive.go    # Directive scope validation
│   │   ├── directive_test.go
│   │   ├── environment.go  # Environment block parsing
│   │   ├── environment_test.go
│   │   ├── errors.go       # Parse error types
│   │   ├── errors_test.go
│   │   ├── include.go      # Include directive parsing
│   │   ├── include_test.go
│   │   ├── parser.go       # Parser with scope stack
│   │   ├── parser_test.go
│   │   ├── recipe.go       # Recipe structure parsing
│   │   ├── recipe_test.go
│   │   ├── scope.go        # Scope types and stack
│   │   ├── scope_test.go
│   │   ├── target.go       # Target parsing
│   │   ├── target_test.go
│   │   ├── variable.go     # Variable parsing
│   │   ├── variable_test.go
│   │   └── version.go      # Version spec parsing
│   ├── semantic/       # Semantic analysis
│       ├── doc.go         # Package documentation
│       ├── capture.go      # Pass 2: Capture validation
│       ├── capture_test.go
│       ├── collector.go    # Pass 1: Symbol collection
│       ├── collector_test.go
│       ├── depgraph.go     # Pass 4: Dependency graph validation
│       ├── depgraph_test.go
│       ├── errors.go       # All semantic error types
│       ├── errors_test.go
│       ├── reference.go    # Pass 3: Reference validation
│       ├── reference_test.go
│       ├── symbols.go      # Symbol table implementation
│       └── symbols_test.go
│   ├── eval/           # Variable evaluation
│   │   ├── doc.go         # Package documentation
│   │   ├── command.go      # Command interpolation
│   │   ├── command_test.go
│   │   ├── context.go      # Evaluation context
│   │   ├── context_test.go
│   │   ├── evaluator.go    # Value evaluator
│   │   └── evaluator_test.go
│   ├── executor/       # Recipe execution
│   │   ├── doc.go         # Package documentation
│   │   ├── executor.go    # Shell executor
│   │   └── executor_test.go
│   ├── environ/        # Environment management
│   │   ├── doc.go         # Package documentation
│   │   ├── requirements.go # Requirements checking
│   │   ├── requirements_test.go
│   │   ├── version.go     # Version parsing and detection
│   │   ├── version_test.go
│   │   └── errors.go      # Environment error types
│   ├── cache/          # Needfile caching
│   │   ├── doc.go         # Package documentation
│   │   ├── needfile.go   # NeedfileCache implementation
│   │   └── needfile_test.go
│   ├── output/         # Build output formatting
│   │   ├── doc.go         # Package documentation
│   │   ├── reporter.go    # Reporter interface and NormalReporter
│   │   └── reporter_test.go
│   ├── errors/         # Error formatting
│   │   ├── doc.go         # Package documentation
│   │   ├── format.go      # FormattedError and source extraction
│   │   └── format_test.go
│   ├── platform/       # Cross-platform utilities
│   │   ├── doc.go         # Package documentation
│   │   ├── shell.go       # Shell detection, path handling, quoting
│   │   └── platform_test.go
│   └── planner/        # Build planning
│       ├── doc.go         # Package documentation
│       ├── match.go        # Target pattern matching
│       └── match_test.go
├── completions/        # Shell completion scripts
│   ├── build.bash    # Bash completion
│   ├── _build        # Zsh completion
│   └── build.fish    # Fish completion
├── Needfile           # Build configuration for this project
└── go.mod
```

## CLI (`cmd/need/cli`)

The command-line interface for need.

### Architecture

The CLI follows interface-based design where `cmd/need/cli` defines the interfaces and internal packages provide implementations:

```
cmd/need/
├── main.go             # CLI entry point, flag handling, Needfile discovery
├── cache_adapter.go    # Needfile cache adapter and cached parsing
├── debug.go            # Debug command implementations (--debug-*)
├── interfaces.go       # Interface definitions (Lexer, Parser, Token, Scope)
├── lexer_adapter.go    # Lexer adapters (Token, Lexer)
├── parser_adapter.go   # Parser adapters (Scope, Parser, Variable, Target, etc.)
├── semantic_adapter.go # Semantic adapters (SymbolTable, Capture, Reference, etc.)
├── eval_adapter.go     # Eval adapters (EvalContext, EvalResult, CommandContext)
├── planner_adapter.go  # Planner adapters (MatchResult, LookupResult, BuildTask)
├── executor_adapter.go # Executor adapters (ShellConfig, Executor, ExecResult)
├── output_adapter.go   # Output adapters (OutputReporter, NormalReporter, CreateOutputEmitter)
├── environ_adapter.go  # Environment adapters (RequirementsChecker, RequirementResult)
├── error_adapter.go    # Error formatting adapters (FormattedError conversion)
└── environ.go          # Environment commands (--check-env, --list-env)
```

**Key Interfaces:**

| Interface | Description |
|-----------|-------------|
| `Token` | Represents a lexical token with type, literal, and location |
| `Lexer` | Tokenizes source code into a stream of tokens |
| `Scope` | Represents parsing context (global, environment, recipe, block) |
| `Parser` | Transforms token stream into AST with scope tracking |
| `DirectiveValidator` | Validates directive placement at scopes |
| `Variable` | Represents a parsed variable definition |
| `VariableParser` | Parses variable definitions |
| `Target` | Represents a parsed target definition with optional recipe |
| `TargetParser` | Parses target definitions |
| `Recipe` | Represents a parsed recipe with directives and commands |
| `Environment` | Represents a parsed environment block |
| `EnvironmentParser` | Parses environment blocks |
| `Conditional` | Represents a parsed conditional block |
| `ConditionalParser` | Parses conditional blocks |
| `Statement` | Represents a parsed AST statement |
| `ParseError` | Represents a parse error with location and hint |
| `NeedfileResult` | Contains parsed statements and collected errors |
| `NeedfileParser` | Parses complete needfiles with error recovery |
| `EvalContext` | Represents evaluation context with variable values |
| `EvalResult` | Contains evaluation results and any errors |
| `CommandContext` | Extends EvalContext with automatic variables and captures |
| `InterpolateResult` | Contains interpolated command string and any errors |
| `RequirementResult` | Represents result of checking a single requirement |
| `RequirementsChecker` | Checks that required binaries are available |
| `Requirement` | Represents a binary requirement |
| `OutputReporter` | Reports build output (target status, command output, summary) |

**Design Rationale:**

This follows Go's "accept interfaces, return structs" principle:
- CLI defines what it needs (interfaces)
- Internal packages provide implementations (concrete types)
- Adapters bridge the gap without exposing internal types to CLI
- Enables testing with mock implementations
- Decouples CLI from internal package structure

### Usage

```
build [options] [targets...]
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Build failure (recipe returned non-zero) |
| 2 | Usage error (bad arguments) |
| 3 | Parse error (invalid Needfile) |
| 4 | Environment error (missing requirements) |

### Flags

All flags from NEEDFILE_SPEC.md are implemented:

| Flag | Description |
|------|-------------|
| `-f, --file` | Use alternate Needfile |
| `-e, --env` | Use named environment |
| `-j, --jobs` | Parallel jobs |
| `-n, --dry-run` | Show what would execute |
| `-v, --verbose` | Verbose output |
| `-q, --quiet` | Suppress non-error output |
| `--color=MODE` | Color output mode: auto, always, never (default: auto) |
| `--progress=MODE` | Progress output mode: auto, always, never (default: auto) |
| `--check-env` | Verify environment requirements |
| `--show-install` | Show install instructions for missing requirements |
| `--list-env` | List available environments |
| `-V, --version` | Show version |
| `-h, --help` | Show help |

### Needfile Discovery

The `findNeedfile()` function searches for Needfile in the following order:

1. Current directory: `Needfile`, `needfile`, `Needfile.need`
2. Parent directories: same candidates, up to filesystem root
3. Returns the first match found

The `-f` / `--file` flag overrides this discovery.

### Debug Flags

| Flag | Description |
|------|-------------|
| `--debug-lex` | Dump lexer analysis (indentation, interpolations) |
| `--debug-parse` | Dump parser scope validation |
| `--debug-var` | Dump variable parsing (shows parsed variables) |
| `--debug-target` | Dump target parsing (shows parsed targets) |
| `--debug-recipe` | Dump recipe parsing (shows parsed recipes with commands) |
| `--debug-env` | Dump environment parsing (shows parsed environment blocks) |
| `--debug-cond` | Dump conditional parsing (shows parsed conditionals) |
| `--debug-include` | Dump include parsing (shows included files and statements) |
| `--debug-ast` | Dump full AST with error recovery (shows all parsed statements and errors) |
| `--debug-semantic` | Dump semantic analysis (shows symbol table) |
| `--debug-eval` | Dump variable evaluation (shows evaluated variables and lazy variables) |
| `--debug-plan` | Dump build planning / target matching (shows target patterns and matches) |

### Version Information

Version and commit are embedded at build time via `-ldflags`:

```bash
go build -ldflags "-X main.version=v1.0.0 -X main.commit=abc123" ./cmd/need
```

### Shell Completions (`completions/`)

Tab completion scripts for common shells:

| File | Shell | Installation |
|------|-------|--------------|
| `build.bash` | Bash | Copy to `/etc/bash_completion.d/` or source in `~/.bashrc` |
| `_build` | Zsh | Copy to a directory in `$fpath` (e.g., `/usr/local/share/zsh/site-functions`) |
| `build.fish` | Fish | Copy to `~/.config/fish/completions/` |

**Features:**
- Complete all CLI flags with descriptions
- Dynamic target completion from Needfile
- Dynamic environment name completion from Needfile
- Completion for `--color` and `--progress` mode values (auto, always, never)
- Completion for job counts (1, 2, 4, 8, 16)

### Error Formatting (`error_adapter.go`)

The CLI integrates with the `internal/errors` package to provide structured, user-friendly error messages with error codes, source context, and help text.

**Key Functions:**

|| Function | Description |
||----------|-------------|
|| `initFileReader()` | Initializes the errors package with file reading capability for source context |
|| `formatParseErrorFromInterface()` | Converts ParseError interface to FormattedError |
|| `FormatParseError()` | Converts parser.ParseError to FormattedError with source context |
|| `FormatSemanticError()` | Converts semantic errors to FormattedError |
|| `FormatEvaluationError()` | Converts evaluation errors to FormattedError |

**Error Category Integration:**

| Category | Error Codes | Implementation |
|----------|-------------|----------------|
| Lexical | E001-E099 | Handled by lexer, formatted via FormatParseError |
| Syntax | E100-E199 | Handled by parser, formatted via FormatParseError |
| Semantic | E200-E299 | Handled by semantic package, formatted via FormatSemanticError |
| Evaluation | E300-E399 | Handled by eval package, formatted via FormatEvaluationError |
| Execution | E400-E499 | Handled by executor package (future) |

**Error Message Format:**

```
error[E103]: directive '.after' invalid at GLOBAL scope
 --> Needfile:2:1
1 | # Test file
2 | .after: invalid
  | ^
3 | 
help: .after is only valid in: RECIPE
```

**Design Decisions:**

1. **File reader initialization**: The `initFileReader()` function is called in `main()` to enable source context extraction from files.

2. **Interface unwrapping**: Parse errors from the NeedfileResult interface are unwrapped to extract the underlying `parser.ParseError` for formatting.

3. **Source context**: All formatting functions accept the source code string to enable extraction of source lines for display.

4. **Error code mapping**: Parse error codes are determined heuristically from error messages using `determineParseErrorCode()`.

5. **Semantic error type switching**: `FormatSemanticError()` uses type switching to handle different semantic error types appropriately.

6. **Evaluation error simplification**: Since evaluation errors often don't have precise source locations, formatting is simplified to show the error message with available context.

## Lexer Package (`internal/lexer`)

### Token Types (`token.go`)

Defines all token types for the Needfile language as specified in `DESIGN.md` Section 2.2.

#### Token Categories

| Category | Token Types | Description |
|----------|-------------|-------------|
| Special | `EOF`, `NEWLINE`, `INDENT`, `COMMENT`, `ERROR` | Control tokens |
| Dot Keywords | `DOT_SHELL`, `DOT_PARALLEL`, `DOT_DEFAULT`, `DOT_INCLUDE`, `DOT_ENVIRONMENT`, `DOT_USING`, `DOT_SOURCE`, `DOT_ARGS`, `DOT_REQUIRES`, `DOT_AFTER`, `DOT_AUTODEPS` | Directives starting with `.` |
| Keywords | `LAZY`, `IF`, `ELIF`, `ELSE`, `END`, `IFDEF`, `IFNDEF`, `BLOCK` | Control flow and modifiers |
| Operators | `EQUALS`, `COLON`, `DOUBLE_EQUALS`, `NOT_EQUALS`, `LPAREN`, `RPAREN`, `COMMA` | Punctuation |
| Identifiers | `IDENTIFIER`, `AT_IDENTIFIER`, `PATH`, `STRING` | Names and literals |
| Interpolation | `INTERP_START`, `INTERP_END`, `INTERP_MOD`, `ESCAPE_LBRACE`, `ESCAPE_RBRACE` | `{var}` syntax |
| Functions | `FUNC_SHELL`, `FUNC_GLOB`, `FUNC_BASENAME`, `FUNC_DIRNAME`, `FUNC_REPLACE` | Built-in functions |

#### SourceLocation

Tracks position in source files for error reporting:
- `File`: Source file path
- `Line`: 1-based line number
- `Column`: 1-based column number

Format: `file:line:column` (e.g., `Needfile:10:5`)

#### Token Structure

```go
type Token struct {
    Type     TokenType
    Literal  string         // The actual text
    Location SourceLocation // Position in source
}
```

#### Lookup Functions

- `LookupKeyword(string) TokenType`: Maps identifiers to keywords/functions or returns `IDENTIFIER`
- `LookupDotKeyword(string) (TokenType, bool)`: Maps `.keyword` strings to directive token types

#### Category Helpers

Category checks use explicit maps instead of fragile range-based checks:

```go
var dotKeywordTypes = map[TokenType]bool{
    DOT_SHELL: true, DOT_PARALLEL: true, // ...
}
func (t TokenType) IsDotKeyword() bool { return dotKeywordTypes[t] }
```

- `TokenType.IsDotKeyword() bool`: True for directive tokens
- `TokenType.IsKeyword() bool`: True for control keywords
- `TokenType.IsFunction() bool`: True for built-in function tokens

### Design Decisions

1. **Explicit token types for each directive**: Rather than a generic `DIRECTIVE` type, each directive (`.shell`, `.parallel`, etc.) has its own token type. This moves validation earlier in the pipeline and makes the parser simpler.

2. **Functions as token types**: Built-in functions (`shell`, `glob`, etc.) are recognized at the lexer level via `LookupKeyword`. This allows the parser to distinguish function calls from regular identifiers without lookahead.

3. **ERROR token for recovery**: The `ERROR` token type supports error recovery by allowing the lexer to continue after encountering invalid input.

4. **1-based line/column**: Matches user expectations and editor conventions for error messages.

### Indentation Tracking (`indent.go`)

Implements indentation state tracking as specified in `DESIGN.md` Section 2.3.3.

#### IndentChar Type

Represents the type of whitespace character used for indentation:

| Value | Description |
|-------|-------------|
| `IndentUnknown` | Not yet determined (initial state) |
| `IndentSpace` | Spaces used for indentation |
| `IndentTab` | Tabs used for indentation |

#### IndentTracker

Stateful tracker that enforces consistent indentation across a Needfile:

```go
type IndentTracker struct {
    char  IndentChar // Character type used for indentation
    width int        // Width of one indentation unit
}
```

**Key Methods:**

| Method | Description |
|--------|-------------|
| `NewIndentTracker()` | Creates tracker with unknown indent type |
| `Process(indent string) (level int, err error)` | Analyzes indentation, returns logical level |
| `Char() IndentChar` | Returns established indent character type |
| `Width() int` | Returns width of one indent unit |
| `Reset()` | Clears state for reuse |

**Indentation Rules (from DESIGN.md):**

1. First indented line establishes the indent unit (e.g., 4 spaces or 1 tab)
2. Subsequent indents must be exact multiples of this unit
3. Mixed tabs and spaces within a single indent string is an error
4. Switching between tabs and spaces after establishing the unit is an error
5. Empty string always returns level 0 without affecting state

#### IndentError

Structured error type for indentation problems:

```go
type IndentError struct {
    Message string
    Line    int
    Column  int
}
```

Format: `indentation error at line X, column Y: message`

#### Design Decisions

1. **Stateful tracking**: The tracker maintains state across lines because the first indented line establishes the unit. This matches the spec requirement that "first indented line establishes the indent unit."

2. **Strict validation**: Mixed indentation (spaces and tabs in the same indent string) is rejected immediately. This catches errors early and provides clear feedback.

3. **Level-based abstraction**: The tracker converts raw character counts to logical levels (0, 1, 2). This simplifies parser logic—it only needs to track levels, not raw widths.

4. **No lookahead required**: Each line's indentation can be validated independently once the unit is established. The tracker doesn't need to see future lines.

5. **Reset capability**: The `Reset()` method allows reuse for different files or testing scenarios without allocating a new tracker.

### Interpolation Boundary Detection (`interp.go`)

Implements interpolation recognition rules as specified in `DESIGN.md` Section 2.3.2.

#### Boundary Rules

Per spec (extended for practical use), `{` is recognized as interpolation start if and only if:
1. Preceded by a boundary character (see table below)
2. Followed by a valid identifier start character (letter or underscore)

**Boundary Characters:**

| Char | Category | Example |
|------|----------|---------|
| ` `, `\t` | Whitespace | `echo {var}` |
| SOL | Start of line | `{var}: dep` |
| `:`, `=` | Operators | `x={var}`, `x:{var}` |
| `/` | Path separator | `{dir}/{file}` |
| `"`, `'` | Quotes | `"{var}"`, `'{var}'` |
| `(`, `)` | Parentheses | `shell({cmd})` |
| `,` | Comma | `replace({a},{b},{c})` |
| `>`, `<` | Redirections | `>{file}`, `<{input}` |

**Not boundaries:** `$`, letters, digits, `_`, `.`, `-`

| Input | Result | Reason |
|-------|--------|--------|
| `{var}` | Valid | SOL |
| `x {var}` | Valid | Space |
| `a/{var}` | Valid | `/` |
| `"{var}"` | Valid | `"` |
| `shell({var})` | Valid | `(` |
| `>{file}` | Valid | `>` |
| `${var}` | Not interpolation | `$` is not a boundary |
| `x{var}` | Not interpolation | Letter is not a boundary |

#### InterpResultKind

Result type for interpolation scanning:

| Value | Description |
|-------|-------------|
| `InterpValid` | Valid interpolation found |
| `InterpEscapedOpen` | Escaped `{{` found |
| `InterpNotInterp` | Not an interpolation (boundary/identifier rules failed) |
| `InterpError` | Malformed interpolation (unclosed, invalid modifier) |

#### InterpResult

Complete result from scanning:

```go
type InterpResult struct {
    Kind  InterpResultKind
    Name  string // Identifier name (for valid/error cases)
    Raw   bool   // Whether :raw modifier present
    Error string // Error message (for InterpError)
}
```

#### Key Functions

| Function | Description |
|----------|-------------|
| `IsValidIdentifierStart(c byte) bool` | True for letters and underscore |
| `IsValidIdentifierChar(c byte) bool` | True for letters, digits, underscore, and dot |
| `IsInterpBoundary(prev byte, atSOL bool) bool` | True if valid interpolation boundary |
| `ScanInterpolation(input, pos, prev, atSOL) (InterpResult, end)` | Scans interpolation at position |
| `ScanEscapedCloseBrace(input, pos) (bool, end)` | Checks for `}}` escape |

#### Identifier Characters

Identifiers can contain:
- Letters (`a-z`, `A-Z`)
- Digits (`0-9`, except as first character)
- Underscore (`_`)
- Dot (`.`) — for automatic variables like `target.dir`, `target.file`

#### Design Decisions

1. **Dot in identifiers**: Dots are allowed in identifiers to support automatic variables like `{target.dir}` and `{target.file}`. This is a slight deviation from typical identifier rules but matches the spec.

2. **Error preservation**: For malformed interpolations (`InterpError`), the parsed identifier name is preserved in the result. This enables better error messages like "unclosed interpolation for 'var'".

3. **Position-based scanning**: `ScanInterpolation` takes a position and previous character rather than operating on a stream. This allows the lexer to decide when to call it and handle the context.

4. **Escape sequence priority**: `{{` is checked before boundary rules. This ensures escape sequences are recognized even at valid interpolation positions.

5. **Strict modifier validation**: Only `:raw` is accepted. Any other modifier (`:foo`) is an error, providing early feedback for typos.

### Lexer Package Structure

The lexer is split across multiple files for maintainability:

| File | Description | Lines |
|------|-------------|-------|
| `lexer.go` | Core state machine, mode switching, interpolation handling | ~350 |
| `value.go` | Value/command mode lexing (`lexValue`, `lexCommand`, etc.) | ~250 |
| `string.go` | String/path/identifier lexing and character classification | ~150 |

### Core Lexer (`lexer.go`)

The main lexer implementation that tokenizes Needfile source code.

#### Lexer Modes

| Mode | Description |
|------|-------------|
| `ModeLineStart` | At beginning of line, consuming indentation |
| `ModeNormal` | Normal token recognition (targets, identifiers, paths) |
| `ModeValue` | After `=` or `:`, consuming value content as strings |
| `ModeInterp` | Inside `{}` interpolation, lexing identifier and modifier |

#### Lexer Structure

```go
type Lexer struct {
    file      string         // Source file name
    input     string         // Input source
    pos       int            // Current position
    line      int            // Current line (1-based)
    col       int            // Current column (1-based)
    mode      LexerMode      // Current lexing mode
    indent    *IndentTracker // Indentation tracker
    prevChar  byte           // Previous character (for boundaries)
    atSOL     bool           // At start of line
    modeStack []LexerMode    // Stack for nested modes
}
```

#### Key Methods

| Method | Description |
|--------|-------------|
| `New(file, input string) *Lexer` | Creates a new lexer |
| `NextToken() Token` | Returns the next token |
| `PeekIsVariableLine() bool` | Peeks ahead to check if line is variable definition (`=` before `:`) |
| `PeekNextIsDotKeyword() bool` | Peeks ahead to check for dot keyword |
| `PeekNextIsBlock() bool` | Peeks ahead to check for block keyword |
| `makeToken(typ, literal)` | Creates token at current position |
| `makeTokenAt(typ, literal, col)` | Creates token at specific column |

#### Token Recognition

**ModeNormal:**
- Dot keywords (`.shell:`, `.default:`, etc.)
- `@identifier` for phony targets
- Identifiers and keywords (`if`, `else`, `lazy`, etc.)
- Paths (tokens containing `/` or `.`)
- Operators (`=`, `:`, `==`, `!=`, `(`, `)`)
- Interpolation start `{` (with boundary check)
- Escape sequences `{{` and `}}`

**ModeValue (after `=` or `:`):**
- String content until newline, comment, or special char
- Interpolations with boundary detection
- Function names followed by `(`

**ModeInterp (inside `{}`):**
- Identifier (with `.` for `target.dir`)
- Modifier (`:raw`)
- Close brace `}`

#### Line Classification

Lines are classified by first non-whitespace token:

| First Token | Line Type |
|-------------|-----------|
| `.keyword` | Directive |
| `@identifier` | Phony target |
| `path:` | File target |
| `identifier =` | Variable |
| `if`, `elif`, `else`, `end` | Conditional |
| `#` | Comment |
| Empty | Blank |

#### Design Decisions

1. **Mode-based lexing**: Different modes handle different contexts (line start, normal, value, interpolation). This simplifies token recognition without complex lookahead.

2. **Mode stack for interpolations**: When entering `{}`, the previous mode is pushed to a stack. After `}`, we pop back to the previous mode. This handles nested contexts correctly.

3. **Value mode simplicity**: After `=` or `:`, most content becomes STRING. The parser handles semantic meaning. This keeps the lexer simple.

4. **Path detection via lookahead**: `looksLikePath()` checks if an identifier-like token contains `/` or `.` ahead. This distinguishes `build/app` (PATH) from `gcc` (IDENTIFIER).

5. **Escape handling in normal mode**: `}}` in normal mode returns ESCAPE_RBRACE. Single `}` in normal mode is treated as string content (unexpected, but recoverable).

6. **Named constant for block keyword**: Uses `const blockKeyword = "block"` instead of magic number for length check in `PeekNextIsBlock()`.

### Value/Command Lexing (`value.go`)

Contains unified methods for lexing values (after `=` or `:`) and commands (recipe lines):

| Method | Description |
|--------|-------------|
| `lexValue()` | Entry point for value mode - skips leading spaces, delegates to `lexValueOrCommand(true)` |
| `lexCommand()` | Entry point for command mode - preserves spaces, delegates to `lexValueOrCommand(false)` |
| `lexValueOrCommand(isValueMode)` | Unified implementation for both modes |
| `lexContentString(isValueMode)` | Unified string consumption for value/command modes |

**Key Difference**: `lexValue()` skips leading whitespace and stops at parentheses/commas for function argument parsing. `lexCommand()` preserves all whitespace for shell command accuracy. Both share ~80% of logic via `lexValueOrCommand()`.

### String/Path Lexing (`string.go`)

Contains methods for lexing identifiers, paths, and general strings:

| Method | Description |
|--------|-------------|
| `lexDotKeyword()` | Handles `.keyword` directives |
| `lexAtIdentifier()` | Handles `@name` phony targets |
| `lexIdentifier()` | Handles identifiers and keywords |
| `lexPath()` | Handles file paths (stops at `{` for interpolation) |
| `lexString()` | General string content in normal mode |

**Character Classification Functions:**

| Function | Description |
|----------|-------------|
| `isIdentStart(ch)` | Delegates to `IsValidIdentifierStart` - letters and underscore |
| `isIdentChar(ch)` | Letters, digits, underscore (excludes dots - see note) |
| `isPathChar(ch)` | Identifier chars plus `/`, `.`, `-` |
| `isPhonyChar(ch)` | Identifier chars plus `-` |

**Note on dot handling**: `isIdentChar` intentionally excludes dots, unlike `IsValidIdentifierChar` in `interp.go` which includes dots for interpolation identifiers like `{target.dir}`. This separation allows proper handling of dots as path separators in normal mode while supporting dotted identifiers inside interpolations.

## AST Package (`internal/ast`)

Defines the Abstract Syntax Tree node types for Needfile parsing. The AST captures syntactic structure without interpretation—no evaluation happens during parsing.

### Package Structure

The AST package is split into logical files by domain:

| File | Contents |
|------|----------|
| `location.go` | `SourceLocation`, `SourceLocationFromToken()` |
| `statement.go` | `Needfile`, `Statement` interface, `Comment`, `Blank` |
| `directives.go` | `DirectiveKind`, `Directive` |
| `environment.go` | `Runtime`, `VersionSpec` types, `Requirement`, `Environment` |
| `variables.go` | `Variable` |
| `conditionals.go` | `Condition` interface and types, `ConditionalBranch`, `Conditional` |
| `targets.go` | `PatternSegment` interface, `LiteralSegment`, `BraceExpr`, `TargetPattern`, `Dependency`, `Target` |
| `recipes.go` | `RecipeDirectives`, `Command`/`CommandPart` interfaces, `LineCommand`, `BlockCommand`, `Recipe` |
| `values.go` | `ValuePart` interface, `LiteralValue`, `Interpolation`, `FunctionName`, `FunctionCall`, `Value` |

### Root Node

```go
type Needfile struct {
    SourcePath string      // Path to the source file
    Statements []Statement // Top-level statements
}
```

### Statement Interface

All top-level AST nodes implement the `Statement` interface:

| Type | Description |
|------|-------------|
| `Directive` | Global directives (`.shell:`, `.parallel:`, `.default:`, `.include:`) |
| `Environment` | Environment block (`.environment:`) |
| `Variable` | Variable definition (immediate or lazy) |
| `Conditional` | If/elif/else/end block |
| `Target` | Target definition with dependencies and recipe |
| `Comment` | Comment line |
| `Blank` | Blank line |

### Directive Types

```go
type DirectiveKind int

const (
    DirectiveShell    DirectiveKind = iota  // .shell:
    DirectiveParallel                        // .parallel:
    DirectiveDefault                         // .default:
    DirectiveInclude                         // .include:
)
```

### Environment Types

#### Runtime

```go
type Runtime int

const (
    RuntimeBare         Runtime = iota  // Host system directly
    RuntimeDocker                        // Docker container
    RuntimePodman                        // Podman container
    RuntimeDevcontainer                  // VS Code devcontainer
    RuntimeNix                           // Nix shell
    RuntimeLima                          // Lima VM (macOS)
)
```

#### VersionSpec Interface

Represents version specifications for requirements:

| Type | Example | Description |
|------|---------|-------------|
| `VersionLatest` | `gcc` or `gcc@latest` | Any version |
| `VersionMajor` | `gcc@11` | Major version 11.x.x |
| `VersionMajorMinor` | `gcc@11.4` | Version 11.4.x |
| `VersionExact` | `gcc@11.4.0` | Exact version |

#### Environment Structure

```go
type Environment struct {
    Name     *string       // nil for default environment
    Runtime  *Runtime      // .using
    Source   *Value        // .source
    Args     *Value        // .args
    Requires []Requirement // .requires
    Location SourceLocation
}
```

### Variable Types

```go
type Variable struct {
    Name     string
    Value    *Value
    Lazy     bool // true for lazy assignment
    Location SourceLocation
}
```

### Conditional Types

#### Condition Interface

| Type | Description | Example |
|------|-------------|---------|
| `EqualsCondition` | `==` comparison | `if {os} == linux` |
| `NotEqualsCondition` | `!=` comparison | `if {os} != windows` |
| `DefinedCondition` | ifdef check | `ifdef DEBUG` |
| `NotDefinedCondition` | ifndef check | `ifndef CC` |

#### Conditional Structure

```go
type Conditional struct {
    IfBranch     ConditionalBranch
    ElifBranches []ConditionalBranch
    ElseBody     []Statement // nil if no else clause
    Location     SourceLocation
}

type ConditionalBranch struct {
    Condition Condition
    Body      []Statement
}
```

### Target Types

#### PatternSegment Interface

Represents segments in a target pattern:

| Type | Description | Example |
|------|-------------|---------|
| `LiteralSegment` | Literal string | `build/`, `.o` |
| `BraceExpr` | Unresolved `{name}` | `{name}` (capture or interpolation TBD) |

**Note:** `BraceExpr` is unresolved during parsing. Semantic analysis determines whether it's a capture or variable interpolation based on the symbol table.

#### Target Structure

```go
type TargetPattern struct {
    Segments    []PatternSegment
    IsPhony     bool // true for @name targets
    IsDirectory bool // true for targets ending with /
}

type Dependency struct {
    Segments []PatternSegment
}

type Target struct {
    Pattern      TargetPattern
    Dependencies []Dependency
    Recipe       *Recipe
    Location     SourceLocation
}
```

### Recipe Types

#### RecipeDirectives

```go
type RecipeDirectives struct {
    Shell    *Value        // .shell override
    After    []*Value      // .after order-only dependencies
    Autodeps *Value        // .autodeps file path
    Requires []Requirement // .requires binaries
}
```

#### Command Interface

| Type | Description |
|------|-------------|
| `LineCommand` | Single command line |
| `BlockCommand` | Block with multiple lines (`block:`) |

#### CommandPart Interface

| Type | Description |
|------|-------------|
| `LiteralCommand` | Literal text in command |
| `CommandInterpolation` | Variable interpolation (`{var}` or `{var:raw}`) |

### Value Types

#### ValuePart Interface

| Type | Description |
|------|-------------|
| `LiteralValue` | Literal text |
| `Interpolation` | Variable interpolation (`{var}` or `{var:raw}`) |
| `FunctionCall` | Function call (`shell()`, `glob()`, etc.) |

#### FunctionName

```go
type FunctionName int

const (
    FuncShell    FunctionName = iota  // shell()
    FuncGlob                           // glob()
    FuncFilename                       // filename()
    FuncDirname                        // dirname()
    FuncReplace                        // replace()
)
```

### SourceLocation

```go
type SourceLocation struct {
    File   string // Source file path
    Line   int    // 1-based line number
    Column int    // 1-based column number
}
```

Can be created from a lexer token using `SourceLocationFromToken(tok)`.

**Note on Duplicate SourceLocation**: Both `lexer.SourceLocation` and `ast.SourceLocation` exist as identical types. This duplication is intentional for layer separation:
- `lexer.SourceLocation` is produced during lexical analysis
- `ast.SourceLocation` is used in AST nodes during parsing
- This keeps packages decoupled (lexer has no dependency on ast)
- `ast.SourceLocationFromToken()` bridges the two types

### Design Decisions

1. **BraceExpr remains unresolved during parsing**: In target patterns, `{name}` could be either a capture or a variable interpolation. The parser produces `BraceExpr` nodes; semantic analysis resolves them based on the symbol table.

2. **Separate Statement and Node interfaces**: All top-level constructs implement `Statement`. This allows type-safe iteration over `Needfile.Statements` with type switches.

3. **Marker method pattern**: Interfaces use unexported marker methods (`statementNode()`, `valuePartNode()`, etc.) to enforce interface implementation at compile time while preventing external implementation.

4. **Nil for optional fields**: Optional fields like `Environment.Name` (nil = default environment), `Recipe.Directives.Shell` (nil = use global default), etc. use nil to indicate absence.

5. **Location on all nodes**: Every AST node carries a `SourceLocation` for error reporting. This enables precise error messages with file:line:column format.

## Parser Package (`internal/parser`)

The parser transforms a token stream from the lexer into an AST. It maintains scope context to validate directive placement.

### Scope Types (`scope.go`)

Defines the parsing scope for directive validation:

```go
type Scope int

const (
    ScopeGlobal      Scope = iota  // Top-level scope
    ScopeEnvironment               // Inside .environment: block
    ScopeRecipe                    // Inside a target's recipe
    ScopeBlock                     // Inside block: within a recipe
)
```

### ScopeStack

Tracks nested scopes during parsing:

| Method | Description |
|--------|-------------|
| `NewScopeStack()` | Creates stack initialized at global scope |
| `Current() Scope` | Returns the topmost scope |
| `Depth() int` | Returns nesting depth |
| `Push(Scope)` | Enters a new scope |
| `Pop() Scope` | Exits current scope (can't pop below global) |
| `IsIn(Scope) bool` | True if scope is anywhere in stack |
| `Reset()` | Returns to global scope |

### Directive Validation (`directive.go`)

Validates directive placement based on current scope per DESIGN.md Section 3.3.3:

| Directive | Valid Scopes |
|-----------|--------------|
| `.shell:` | GLOBAL, RECIPE |
| `.parallel:` | GLOBAL |
| `.default:` | GLOBAL |
| `.include:` | GLOBAL |
| `.environment:` | GLOBAL |
| `.using:` | ENVIRONMENT |
| `.source:` | ENVIRONMENT |
| `.args:` | ENVIRONMENT |
| `.requires:` | ENVIRONMENT, RECIPE |
| `.after:` | RECIPE |
| `.autodeps:` | RECIPE |

**Key Functions:**

| Function | Description |
|----------|-------------|
| `IsDirectiveValidAtScope(tok, scope) bool` | Returns true if directive is valid at scope |
| `ValidScopesForDirective(tok) []Scope` | Returns list of valid scopes for directive |

### Parse Errors (`errors.go`)

Structured error types for parser errors:

```go
type ParseError struct {
    Message  string
    Location lexer.SourceLocation
    Hint     string  // Optional fix suggestion
}

type ParseErrors struct {
    Errors []*ParseError
}
```

Error format: `file:line:column: message (hint: suggestion)`

**Error Creation:**

| Function | Description |
|----------|-------------|
| `NewScopeError(directive, scope, loc)` | Creates error for directive at wrong scope |
| `DirectiveNameForError(tok) string` | Returns `.name` format for error messages |

### Parser Structure (`parser.go`)

```go
type Parser struct {
    lexer   *lexer.Lexer
    current lexer.Token
    scope   *ScopeStack
    errors  *ParseErrors
}
```

**Key Methods:**

| Method | Description |
|--------|-------------|
| `New(lexer) *Parser` | Creates parser, primes first token |
| `CurrentScope() Scope` | Returns current parsing scope |
| `EnterScope(Scope)` | Pushes new scope |
| `ExitScope() Scope` | Pops current scope |
| `CurrentIndentLevel() int` | Expected indent level for scope |
| `Errors() *ParseErrors` | Returns collected errors |
| `HasErrors() bool` | True if any errors collected |

### Indentation Levels

Scopes map to expected indentation:

| Scope | Indent Level |
|-------|--------------|
| GLOBAL | 0 |
| ENVIRONMENT | 1 |
| RECIPE | 1 |
| BLOCK | 2 |

### Design Decisions

1. **Scope stack for context-sensitive parsing**: The scope stack enables directive validation during parsing without semantic analysis. This catches scope errors early.

2. **Separate directive validation**: Directive scope rules are data-driven (`directiveScopes` map), making it easy to add new directives or modify rules.

3. **Error collection**: Parser collects multiple errors rather than failing on first error, enabling better developer experience.

4. **Exported methods for public API**: Scope and state methods (`EnterScope`, `ExitScope`, `CurrentScope`, `CurrentIndentLevel`, `CurrentToken`) are exported for CLI/testing access. Since the package is `internal/`, the exported status primarily serves documentation purposes and adapter access.

5. **Two-tier error handling pattern**: The parser uses two consistent error handling approaches:
   - **Top-level parsing functions** (`ParseVariable`, `ParseTarget`, `ParseConditional`, etc.) return `*ParseError`. These are called by `ParseNeedfile()` which catches errors and performs recovery via `recoverToLevel0()`.
   - **Value/content parsing functions** (`ParseValue`, `parseInterpolation`, `parseFunctionCall`, etc.) use `addError()` internally and return partial results (or nil). This allows continued parsing despite malformed values.
   
   This design enables error recovery: structural errors stop the current block and trigger recovery, while value-level errors are collected but don't halt parsing.

### Variable Parsing (`variable.go`)

Parses variable definitions per DESIGN.md Section 3.2 grammar:
```
variable_def = [ "lazy" ] identifier "=" value NEWLINE ;
value = { value_part } ;
value_part = STRING | interpolation | function_call ;
```

**Key Functions:**

| Function | Description |
|----------|-------------|
| `ParseVariable() (*ast.Variable, *ParseError)` | Parses a complete variable definition |
| `ParseValue() *ast.Value` | Parses a value with interpolations and functions |
| `parseInterpolation() *ast.Interpolation` | Parses `{name}` or `{name:raw}` |
| `parseFunctionCall() *ast.FunctionCall` | Parses `shell(...)`, `glob(...)`, etc. |
| `parseFunctionArg() *ast.Value` | Parses a function argument |

**Variable Detection:**

A line is classified as a variable definition if:
1. It contains `=`
2. The `=` appears before any `:` (or `:` doesn't appear)

This distinguishes variables from targets:
- `cc = gcc` → Variable (= before any :)
- `build/app: deps` → Target (: first)
- `path = /usr/bin:foo` → Variable (= at position 4, : at position 13)

**Value Parsing:**

Values are sequences of:
- **Literal text**: String content
- **Interpolations**: `{var}` or `{var:raw}` for raw (unquoted) output
- **Function calls**: `shell(cmd)`, `glob(pattern)`, `basename(path)`, `dirname(path)`, `replace(str, from, to)`

The parser stops value parsing at:
- Newline
- Comment (`#`)
- End of file

**Interpolation Parsing:**

Handles:
- Simple interpolation: `{varname}`
- Dotted names: `{target.dir}` (for automatic variables)
- Raw modifier: `{flags:raw}` (disables shell quoting)

**Function Call Parsing:**

Recognizes built-in functions:
| Function | AST Type | Description |
|----------|----------|-------------|
| `shell(...)` | `FuncShell` | Execute shell command |
| `glob(...)` | `FuncGlob` | File pattern matching |
| `basename(...)` | `FuncBasename` | Extract filename |
| `dirname(...)` | `FuncDirname` | Extract directory |
| `replace(...)` | `FuncReplace` | String replacement (3 args) |

Function arguments can contain interpolations:
```
sources = shell(find {src_dir} -name *.c)
objects = replace({sources}, .c, .o)
```

**Multi-Argument Functions:**

The `replace(...)` function takes three comma-separated arguments:
```
replace(input, from, to)
```

The parser handles:
- Comma-separated arguments (stops at `,` unless inside nested parentheses)
- Nested parentheses in arguments (e.g., `shell(echo $(date))`)
- Interpolations within any argument

**COMMA Token:**

The lexer emits `COMMA` tokens when in value mode. The parser uses this to separate function arguments:
```
replace({sources}, .c, .o)
        ^       ^  ^  ^  ^
        |       |  |  |  +-- arg3
        |       |  |  +-- COMMA
        |       |  +-- arg2
        |       +-- COMMA
        +-- arg1 (with interpolation)
```

**Design Decisions:**

1. **Lazy prefix as keyword**: The `lazy` keyword is recognized by the lexer as `LAZY` token, making lazy variable detection straightforward.

2. **Value parts as interface**: `ast.ValuePart` interface allows mixing literals, interpolations, and function calls in any order.

3. **Parentheses depth tracking**: Function argument parsing tracks parenthesis depth to handle nested parentheses in function arguments.

4. **Error recovery**: On parse error, the parser records the error and allows continued parsing of subsequent tokens.

### Target Parsing (`target.go`)

Parses target definitions per DESIGN.md Section 3.2 grammar:
```
target_def = target_spec ":" dependency_list NEWLINE [ recipe ] ;
target_spec = phony_target | file_target ;
phony_target = "@" identifier ;
file_target = path_pattern ;
path_pattern = { path_segment | capture } ;
dependency_list = { dependency } ;
dependency = path_pattern | interpolation ;
```

**Key Functions:**

| Function | Description |
|----------|-------------|
| `ParseTarget() (*ast.Target, *ParseError)` | Parses a complete target definition |
| `parseTargetPattern(isPhony bool) (*ast.TargetPattern, *ParseError)` | Parses target pattern (left of `:`) |
| `parseBraceExpr() (*ast.BraceExpr, *ParseError)` | Parses `{name}` in patterns |
| `parseDependencyList() ([]ast.Dependency, *ParseError)` | Parses dependencies (right of `:`) |

**Target Types:**

| Type | Syntax | Example | Detection |
|------|--------|---------|-----------|
| File target | `path:` | `build/app:` | No `@` prefix |
| Phony target | `@name:` | `@clean:` | Starts with `@` |
| Directory target | `path/:` | `build/:` | Ends with `/` |
| Pattern target | `{name}` in path | `build/{name}.o:` | Contains `BraceExpr` |

**Pattern Parsing:**

Target patterns are parsed into segments:
- `LiteralSegment`: Literal path text (e.g., `build/`, `.o`)
- `BraceExpr`: Unresolved `{name}` expression

At parse time, `{name}` in patterns is stored as `BraceExpr`. Semantic analysis later determines if it's:
- A capture (pattern matching variable)
- A variable interpolation (if name is defined)

**Dependency Parsing:**

After the colon, the lexer enters `ModeValue` and returns STRING tokens with interpolations interspersed. Dependencies are space-separated:

```
build/app: build/main.o build/utils.o
          ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
          Dependencies (space-separated)
```

The parser handles this by:
1. Accumulating segments for current dependency
2. Flushing on whitespace boundaries (from STRING tokens containing spaces)
3. Keeping interpolations attached to surrounding literals (e.g., `src/{name}.c` is ONE dependency)

**Directory Detection:**

A target is marked as a directory (`IsDirectory=true`) if its final merged pattern ends with `/`:

```
build/:      # IsDirectory = true
build/{name}.o:  # IsDirectory = false (ends with .o)
```

**Design Decisions:**

1. **BraceExpr deferred resolution**: The parser produces `BraceExpr` nodes without determining if they are captures or interpolations. This keeps parsing context-free and moves semantic decisions to the appropriate phase.

2. **Adjacent literal merging**: After parsing, adjacent `LiteralSegment` nodes are merged for a cleaner AST. This simplifies pattern text reconstruction.

3. **Space handling in dependencies**: The lexer skips spaces between tokens in value mode, so the parser detects dependency boundaries by checking for spaces within STRING tokens.

4. **Phony name extraction**: For phony targets (`@name`), the parser strips the `@` prefix and stores just the name in the pattern.

### Recipe Parsing

Recipe parsing is split across three files:

| File | Description | Key Functions |
|------|-------------|---------------|
| `recipe.go` | Recipe structure and directives | `parseRecipe`, `parseRecipeDirective`, `parseRecipeShell/After/Autodeps/Requires` |
| `command.go` | Command and block parsing | `parseCommandLine`, `parseBlockCommand`, `parseCommandInterpolation` |
| `version.go` | Version spec parsing | `parseRequirementsList`, `parseRequirement`, `parseVersionSpec` |

Parses recipe sections per DESIGN.md Section 3.2 grammar:
```
recipe = INDENT { recipe_line } DEDENT ;
recipe_line = recipe_directive NEWLINE | block_stmt | command NEWLINE ;
recipe_directive = ".shell:" value | ".after:" value | ".autodeps:" value | ".requires:" value ;
block_stmt = "block:" NEWLINE INDENT { raw_line } DEDENT ;
command = { command_part } ;
command_part = STRING | interpolation ;
```

**Key Functions:**

| Function | Description |
|----------|-------------|
| `parseRecipe() (*ast.Recipe, *ParseError)` | Parses complete recipe section |
| `parseRecipeDirective(recipe *ast.Recipe) *ParseError` | Routes directive parsing |
| `parseRecipeShell(recipe *ast.Recipe) *ParseError` | Parses `.shell:` directive |
| `parseRecipeAfter(recipe *ast.Recipe) *ParseError` | Parses `.after:` directive |
| `parseRecipeAutodeps(recipe *ast.Recipe) *ParseError` | Parses `.autodeps:` directive |
| `parseRecipeRequires(recipe *ast.Recipe) *ParseError` | Parses `.requires:` directive |
| `parseRequirementsList() ([]ast.Requirement, *ParseError)` | Parses space-separated requirements |
| `parseCommandLine() (*ast.LineCommand, *ParseError)` | Parses single command line |
| `parseCommandInterpolation() (*ast.CommandInterpolation, *ParseError)` | Parses `{var}` or `{var:raw}` in commands |
| `parseBlockCommand() (*ast.BlockCommand, *ParseError)` | Parses `block:` with nested lines |
| `calculateIndentLevel(indent string) int` | Converts indent string to logical level |

**Recipe Detection:**

A recipe is detected when an INDENT token follows a target line:
```
build/app: src/main.c      # Target line
    gcc -o {target} {deps}  # Recipe starts here (INDENT token)
```

The recipe ends when:
- EOF is reached
- A line with no indentation (level 0) is encountered
- A line with indentation returns to recipe level after deeper nesting

**Recipe Directives:**

| Directive | Purpose | Example |
|-----------|---------|---------|
| `.shell:` | Override shell for this recipe | `.shell: bash` |
| `.after:` | Order-only prerequisites | `.after: build/` |
| `.autodeps:` | Auto-generated dependency file | `.autodeps: build/app.d` |
| `.requires:` | Required binaries | `.requires: gcc@11 pkg-config@latest` |

**Requirement Parsing:**

Requirements are space-separated with optional version specs:
```
.requires: gcc@11 python3@3.10 pkg-config@latest cmake
```

Version formats:
- `name` or `name@latest` → `VersionLatest`
- `name@11` → `VersionMajor{Major: 11}`
- `name@3.10` → `VersionMajorMinor{Major: 3, Minor: 10}`
- `name@11.4.0` → `VersionExact{Major: 11, Minor: 4, Patch: 0}`

**Command Parsing:**

Commands are sequences of:
- **Literal text**: The command text
- **Interpolations**: `{var}` or `{var:raw}` for unquoted expansion

```go
type LineCommand struct {
    Parts    []CommandPart
    Location SourceLocation
}
```

**Block Commands:**

Block commands (`block:`) pass multiple lines as a single script:
```
build/app: src/main.c
    block:
        if [[ -f {target} ]]; then
            rm {target}
        fi
        gcc -o {target} {deps}
```

Block content is at indentation level 2 (deeper than recipe level 1). Each line is parsed independently for interpolations:

```go
type BlockCommand struct {
    Lines    [][]CommandPart
    Location SourceLocation
}
```

**Scope Management:**

Recipe parsing manages scope transitions:
1. Enter `ScopeRecipe` when recipe starts
2. Enter `ScopeBlock` when `block:` encountered
3. Exit `ScopeBlock` when dedenting from block
4. Exit `ScopeRecipe` when dedenting from recipe

**Design Decisions:**

1. **Indent level via lexer tracker**: The parser uses the lexer's `IndentTracker` to calculate logical indent levels, ensuring consistent handling of spaces vs tabs.

2. **Directive order flexibility**: Directives can appear anywhere in the recipe (interspersed with commands), though best practice is to place them first.

3. **Block scope isolation**: Block commands have their own scope (`ScopeBlock`) to potentially enforce different directive rules (none currently allowed in blocks).

4. **Version parsing tolerance**: Invalid version strings fall back to `VersionLatest` rather than erroring, for flexibility in `.requires:` specifications.

5. **Space handling in commands**: The lexer's value mode skips leading spaces, so spaces between command tokens are not preserved in the AST. This is acceptable since command reconstruction for execution will properly quote interpolated values.

### Environment Parsing (`environment.go`)

Parses environment blocks per DESIGN.md Section 3.2 grammar:
```
environment_block = ".environment:" [ identifier ] NEWLINE
                    INDENT { env_directive NEWLINE } DEDENT ;
env_directive = ".using:" value | ".source:" value | ".args:" value | ".requires:" value ;
```

**Key Functions:**

| Function | Description |
|----------|-------------|
| `ParseEnvironment() (*ast.Environment, *ParseError)` | Parses complete environment block |
| `parseEnvironmentDirective(env *ast.Environment) *ParseError` | Routes directive parsing |
| `parseEnvUsing(env *ast.Environment) *ParseError` | Parses `.using:` directive |
| `parseEnvSource(env *ast.Environment) *ParseError` | Parses `.source:` directive |
| `parseEnvArgs(env *ast.Environment) *ParseError` | Parses `.args:` directive |
| `parseEnvRequires(env *ast.Environment) *ParseError` | Parses `.requires:` directive |
| `parseRuntimeType() (ast.Runtime, *ParseError)` | Parses runtime type name |

**Environment Detection:**

An environment block is detected when `.environment:` is encountered at global scope:
```
.environment: ci              # Named environment "ci"
    .using: docker
    .source: Dockerfile.ci
    .args: --platform linux/amd64
    .requires: gcc@11

.environment:                 # Default (unnamed) environment
    .using: bare
    .requires: python3@3.10
```

The environment block ends when:
- EOF is reached
- A line with no indentation (level 0) is encountered
- The parser dedents back to global scope

**Environment Directives:**

| Directive | Purpose | Example |
|-----------|---------|---------|
| `.using:` | Specify runtime type | `.using: docker` |
| `.source:` | Path to runtime configuration | `.source: Dockerfile.ci` |
| `.args:` | Runtime-specific arguments | `.args: --platform linux/amd64` |
| `.requires:` | Required binaries with versions | `.requires: gcc@11 python3@latest` |

**Runtime Types:**

| Type | AST Value | Description |
|------|-----------|-------------|
| `bare` | `RuntimeBare` | Host system directly |
| `docker` | `RuntimeDocker` | Docker container |
| `podman` | `RuntimePodman` | Podman container |
| `devcontainer` | `RuntimeDevcontainer` | VS Code devcontainer |
| `nix` | `RuntimeNix` | Nix shell environment |
| `lima` | `RuntimeLima` | Lima VM (macOS) |

**Scope Management:**

Environment parsing manages scope transitions:
1. Enter `ScopeEnvironment` when `.environment:` is parsed
2. Validate that only environment-specific directives are used
3. Exit `ScopeEnvironment` when dedenting back to global scope

**Value Parsing in Directives:**

Values in environment directives can contain:
- Literal text: `Dockerfile.ci`
- Interpolations: `{docker_dir}/Dockerfile`
- Function calls: Not typically used but supported

**Design Decisions:**

1. **Optional name handling**: Default environments have `Name == nil`, named environments have a pointer to the name string.

2. **Runtime as pointer**: `Runtime` is stored as a pointer to allow distinguishing between "not specified" (nil) and "explicitly set to bare".

3. **Directive reuse**: The `.requires:` directive parsing is shared with recipe parsing via `parseRequirementsList()`.

4. **Strict scope validation**: Invalid directives in environment scope (like `.shell:`) are rejected with clear error messages.

5. **Comment handling**: Comments within environment blocks are skipped, allowing inline documentation.

### Conditional Parsing (`conditional.go`)

Parses conditional blocks per DESIGN.md Section 3.2 grammar:
```
conditional = if_clause { elif_clause } [ else_clause ] "end" NEWLINE ;
if_clause = "if" condition NEWLINE { statement } ;
elif_clause = "elif" condition NEWLINE { statement } ;
else_clause = "else" NEWLINE { statement } ;
condition = interpolation "==" value | interpolation "!=" value ;
ifdef_clause = "ifdef" identifier NEWLINE { statement } "end" NEWLINE ;
ifndef_clause = "ifndef" identifier NEWLINE { statement } "end" NEWLINE ;
```

**Key Functions:**

| Function | Description |
|----------|-------------|
| `IsConditionalLine() bool` | Returns true if current token starts a conditional (if, ifdef, ifndef) |
| `ParseConditional() (*ast.Conditional, *ParseError)` | Parses complete conditional block |
| `parseIfConditional(loc) (*ast.Conditional, *ParseError)` | Parses if/elif/else/end block |
| `parseIfdefConditional(loc, isDefined) (*ast.Conditional, *ParseError)` | Parses ifdef/ifndef/end block |
| `parseElifBranch() (*ast.ConditionalBranch, *ParseError)` | Parses a single elif branch |
| `parseCondition() (ast.Condition, *ParseError)` | Parses condition expression (== or !=) |
| `parseConditionValue() *ast.Value` | Parses left or right side of condition |
| `parseConditionalBody() ([]ast.Statement, *ParseError)` | Parses statements until end, elif, or else |
| `parseBodyStatement() (ast.Statement, *ParseError)` | Parses a single statement in conditional body |

**Conditional Types:**

| Type | Syntax | Example | AST Condition Type |
|------|--------|---------|-------------------|
| If with equals | `if {var} == value` | `if {os} == linux` | `EqualsCondition` |
| If with not-equals | `if {var} != value` | `if {os} != windows` | `NotEqualsCondition` |
| Ifdef | `ifdef VARNAME` | `ifdef DEBUG` | `DefinedCondition` |
| Ifndef | `ifndef VARNAME` | `ifndef CC` | `NotDefinedCondition` |

**Conditional Structure:**

```
if {os} == linux       # if branch
cc = gcc
cflags = -Wall
elif {os} == darwin    # elif branch (0 or more)
cc = clang
cflags = -Wall -Wextra
else                   # else branch (optional)
cc = cc
end                    # required terminator
```

**Condition Parsing:**

The condition expression is parsed in two parts:
1. Left side: Typically an interpolation `{var}` or `{var:raw}`
2. Comparison operator: `==` or `!=`
3. Right side: A value (literal or containing interpolations)

```go
type EqualsCondition struct {
    Left  *Value
    Right *Value
}

type NotEqualsCondition struct {
    Left  *Value
    Right *Value
}
```

**Body Statement Parsing:**

Statements allowed in conditional bodies:
- Variable definitions (immediate or lazy)
- Nested conditionals
- Comments
- Blank lines

**Nested Conditional Support:**

Conditionals can be nested to arbitrary depth:
```
if {os} == linux
    ifdef DEBUG
        debug_flags = -g
    end
    cc = gcc
end
```

**Error Handling:**

| Error | Condition | Message |
|-------|-----------|---------|
| Missing condition | No condition after `if` | "expected condition expression" |
| Missing operator | No `==` or `!=` | "expected '==' or '!=' in condition" |
| Missing end | EOF before `end` | "expected 'end' to close conditional" |
| Missing identifier | No identifier after `ifdef`/`ifndef` | "expected identifier after 'ifdef'/'ifndef'" |

**Design Decisions:**

1. **Unified conditional structure**: Both `if/elif/else/end` and `ifdef/ifndef/end` use the same `ast.Conditional` structure. This simplifies evaluation logic.

2. **DefinedCondition vs interpolation check**: `ifdef` and `ifndef` store just the variable name (string), not a full value. This makes the "is defined" check straightforward.

3. **Body parsing delegation**: `parseBodyStatement()` delegates to existing parsers (`ParseVariable`, `ParseConditional`) for reuse.

4. **Flexible condition values**: Both sides of `==`/`!=` are full values, allowing `{var1} == {var2}` comparisons.

5. **No scope change for conditionals**: Conditionals don't create a new scope. Variables defined in conditional bodies are visible after the conditional ends (matching Make behavior).

### Include Parsing (`include.go`)

Parses `.include:` directives per DESIGN.md Section 3.2 grammar:
```
global_directive = ".include:" value ;
```

**Key Functions:**

| Function | Description |
|----------|-------------|
| `ParseInclude() (*ast.Directive, []ast.Statement, *ParseError)` | Parses .include: directive and returns included statements |
| `parseIncludeWithStack(stack) (...)` | Internal implementation with circular include tracking |
| `parseIncludedFile(path, content, stack) ([]ast.Statement, *ParseError)` | Parses content of included file |
| `parseStatement() (ast.Statement, *ParseError)` | Parses a single top-level statement |
| `extractLiteralPath(v *ast.Value) string` | Extracts literal path from value (interpolation not yet supported) |

**Include Detection:**

An `.include:` directive is detected when `DOT_INCLUDE` token is encountered at global scope:
```
.include: ./common.need
.include: {config_dir}/settings.need
```

**Recursive Parsing:**

The include parser:
1. Extracts the path from the directive value
2. Resolves relative paths based on the including file's directory
3. Reads the included file content
4. Recursively parses the included file
5. Returns both the directive AST node and the parsed statements

**Circular Include Detection:**

Uses an `includeStack` to track files being processed:
```go
type includeStack struct {
    files map[string]bool
}
```

- Before parsing a file, its absolute path is added to the stack
- If the path already exists, a "circular include detected" error is returned
- After parsing completes, the path is removed from the stack

This detects both direct circular includes (A→A) and indirect circular includes (A→B→A).

**Relative Path Resolution:**

Include paths are resolved relative to the including file:
```
# In /project/build/Needfile:
.include: ./common.need    # → /project/build/common.need
.include: ../lib/deps.need # → /project/lib/deps.need
.include: /etc/defaults.need # → /etc/defaults.need (absolute)
```

**Statement Parsing:**

The `parseStatement()` function handles all statement types in included files:
- Variable definitions
- Target definitions
- Directives (`.shell:`, `.parallel:`, `.default:`)
- Environment blocks
- Nested conditionals
- Nested includes

**Design Decisions:**

1. **Literal paths only**: Currently, interpolation in include paths is not supported. The path must be a literal string. This simplifies implementation and avoids chicken-and-egg issues with variable evaluation order.

2. **Included statements merged into parent**: `ParseNeedfile()` uses `parseTopLevelStatements()` which handles includes specially. When an include is encountered, the included statements are prepended to the directive statement, ensuring variables/targets from included files are visible to subsequent content in the including file.

3. **Recursive with stack**: The include stack is passed through recursive calls to detect circular includes at any depth.

4. **Comments preserved**: Comments in included files are preserved in the returned statements.

5. **No indentation in included files**: Included files are parsed at global scope. Indented content in included files follows the same rules as the main file.

### Error Recovery (`parser.go` and `recovery_test.go`)

Implements error recovery to collect multiple parse errors and continue parsing after errors.

**Key Functions:**

| Function | Description |
|----------|-------------|
| `ParseNeedfile() ([]Statement, *ParseErrors)` | Parses complete needfile with error recovery |
| `parseTopLevelStatements() ([]Statement, *ParseError)` | Parses one or more statements (handles includes) |
| `recoverToLevel0()` | Skips to next line at indentation level 0 |
| `looksLikeVariableLine() bool` | Heuristic to detect variable definitions |

**Error Recovery Strategy:**

1. On parse error, record the error in `ParseErrors` collection
2. Skip to the next line at indentation level 0 (global scope)
3. Continue parsing from there
4. Stop after `maxErrors` (10) to avoid infinite loops on badly malformed input

**Recovery Rules:**

| State | Recovery Action |
|-------|-----------------|
| Invalid directive at global scope | Record error, skip line, continue |
| Malformed target | Record error, skip to level 0, continue |
| Unclosed conditional | Record error at EOF, stop |
| Invalid token | Record error, skip line, continue |
| Environment with invalid directive | Record error, skip block, continue |

**Example:**

```
.after: invalid          # Error: .after invalid at GLOBAL scope
cc = gcc                 # Parsed successfully
@test:                   # Parsed successfully
    echo hello
.using: invalid          # Error: .using invalid at GLOBAL scope
suffix = bar             # Parsed successfully
```

Results in:
- 2 errors (both scope errors)
- 3 statements (2 variables, 1 target)

**Constants:**

| Constant | Value | Description |
|----------|-------|-------------|
| `maxErrors` | 10 | Maximum errors before giving up |

**Error Message Format:**

All parse errors include:
- `Message`: Human-readable description
- `Location`: File, line, and column (`SourceLocation`)
- `Hint`: Optional fix suggestion

Example error output:
```
Needfile:1:1: directive '.after' invalid at GLOBAL scope (hint: .after is only valid in: RECIPE)
```

**Design Decisions:**

1. **Skip to level 0**: Recovery skips to indentation level 0 to ensure we're back at global scope. This avoids confusion with partially parsed blocks.

2. **Max error limit**: After 10 errors, parsing stops to avoid runaway error cascades on severely broken input.

3. **Preserve valid statements**: All successfully parsed statements are returned even if errors occurred. This enables partial analysis of broken files.

4. **Scope error hints**: Directive scope errors include hints listing valid scopes, helping users understand where directives belong.

5. **Indented line skipping**: During recovery, any indented lines (part of a block) are skipped until we reach a non-indented line.

## Semantic Package (`internal/semantic`)

The semantic package provides semantic analysis for Needfiles. It validates the AST produced by the parser and resolves context-sensitive constructs.

### Semantic Error Types (`errors.go`)

All semantic error types are consolidated in `errors.go` for maintainability. Each error type corresponds to a specific validation pass:

| Error Type | Pass | Description |
|------------|------|-------------|
| `DuplicateDefinitionError` | Pass 1 | Symbol defined multiple times |
| `AutomaticInPatternError` | Pass 2 | Automatic variable used in target pattern |
| `CaptureMismatchError` | Pass 2 | Capture in dependency not defined in target |
| `UndefinedVariableError` | Pass 3 | Reference to undefined variable |
| `AutomaticOutsideRecipeError` | Pass 3 | Automatic variable used outside recipe |
| `CircularDependencyError` | Pass 4 | Circular dependency in target graph |

#### DuplicateDefinitionError

Returned by Pass 1 (Symbol Collection) when a variable, target, or environment is defined multiple times:

```go
type DuplicateDefinitionError struct {
    Kind   string             // "variable", "target", or "environment"
    Name   string             // The duplicated name
    First  ast.SourceLocation // First definition location
    Second ast.SourceLocation // Duplicate location
}
```

Error format: `duplicate variable 'cc': first defined at Needfile:1:1, redefined at Needfile:5:1`

#### AutomaticInPatternError

Returned by Pass 2 (Capture Validation) when an automatic variable is used in a target pattern:

```go
type AutomaticInPatternError struct {
    Name     string             // The automatic variable name
    Location ast.SourceLocation // Location of the usage
}
```

Error format: `automatic variable 'target' cannot be used as capture in target pattern at Needfile:5:10`

Automatic variables (`target`, `deps`, `in`, `out`, `stem`, `target.dir`, `target.file`) are only available during recipe execution.

#### CaptureMismatchError

Returned by Pass 2 when a capture appears in a dependency but not in the target pattern:

```go
type CaptureMismatchError struct {
    Name      string             // The capture name
    InTarget  bool               // true if capture is in target (allowed)
    Location  ast.SourceLocation // Location of the problematic usage
    TargetLoc ast.SourceLocation // Location of the target definition
}
```

Error format: `capture '{name}' in dependency but not defined in target pattern at Needfile:3:20 (target at Needfile:3:1)`

#### UndefinedVariableError

Returned by Pass 3 (Reference Validation) when a reference points to an undefined variable:

```go
type UndefinedVariableError struct {
    Name     string             // The undefined variable name
    Location ast.SourceLocation // Location of the reference
}
```

Error format: `undefined variable 'foo' at Needfile:7:15`

#### AutomaticOutsideRecipeError

Returned by Pass 3 when an automatic variable is used outside a recipe or block:

```go
type AutomaticOutsideRecipeError struct {
    Name     string             // The automatic variable name
    Location ast.SourceLocation // Location of the reference
}
```

Error format: `automatic variable 'target' is only valid inside recipe or block scope at Needfile:2:10`

#### CircularDependencyError

Returned by Pass 4 (Dependency Graph Validation) when a circular dependency is detected:

```go
type CircularDependencyError struct {
    Cycle []string // e.g., ["a", "b", "c", "a"]
}
```

Error format: `circular dependency detected: a -> b -> c -> a`

The cycle path includes the starting node at both ends to show where the cycle closes.

### Symbol Table (`symbols.go`)

The symbol table tracks all defined symbols in a Needfile for semantic validation.

#### SymbolTable Structure

```go
type SymbolTable struct {
    Variables      map[string]*ast.Variable     // User-defined variables
    Targets        []*ast.Target                // All target definitions
    Environments   map[string]*ast.Environment  // Named environments ("" = default)
    automatic      map[string]bool              // Automatic variable names
    builtin        map[string]bool              // Built-in variable names
}
```

#### Key Methods

| Method | Description |
|--------|-------------|
| `NewSymbolTable()` | Creates initialized table with automatic/built-in vars |
| `AddVariable(v)` | Adds variable, returns error on duplicate |
| `AddTarget(t)` | Adds target, returns error on duplicate pattern |
| `AddEnvironment(e)` | Adds environment, returns error on duplicate name |
| `IsAutomatic(name)` | Returns true if name is an automatic variable |
| `IsBuiltin(name)` | Returns true if name is a built-in variable |
| `IsDefined(name)` | Returns true if name is defined (user, automatic, or built-in) |
| `LookupVariable(name)` | Returns variable or nil if not found |

#### Automatic Variables

Per DESIGN.md Section 3.3.4, these are only valid inside recipe/block scope:

| Variable | Description |
|----------|-------------|
| `target` | Target file path |
| `deps` | All dependencies (space-separated) |
| `in` | First dependency |
| `out` | Alias for target |
| `stem` | Pattern match stem (for pattern targets) |
| `target.dir` | Directory part of target |
| `target.file` | Filename part of target |

#### Built-in Variables

Always available, not user-defined:

| Variable | Description |
|----------|-------------|
| `os` | Operating system name |
| `arch` | Architecture name |

#### PatternString Function

Converts a target pattern to a string for duplicate detection:

```go
func PatternString(p *ast.TargetPattern) string
```

Examples:
- `build/app` → `"build/app"`
- `@clean` → `"@clean"` (phony)
- `build/{name}.o` → `"build/{name}.o"` (pattern)

### Design Decisions

1. **Map for variables and environments**: O(1) lookup by name. Variables and environments have unique names within their respective namespaces.

2. **Slice for targets**: Targets are stored in definition order. Pattern targets may have overlapping matches, so we need to preserve order for match priority.

3. **Separate tracking for target patterns**: The `targetPatterns` map tracks seen patterns for duplicate detection without affecting target storage order.

4. **Automatic vs built-in distinction**: Automatic variables (`target`, `deps`, etc.) are only valid in recipe scope, while built-in variables (`os`, `arch`) are available everywhere. The symbol table distinguishes these categories.

5. **Environment name handling**: The default (unnamed) environment uses empty string `""` as key. Named environments use their name as key.

### Symbol Collection (`collector.go`)

Pass 1 of semantic analysis: collects all symbol definitions and detects duplicates.

#### Collect Function

```go
func Collect(stmts []ast.Statement) (*SymbolTable, []error)
```

The `Collect` function walks the AST and populates a symbol table with all definitions:
- Variable definitions (immediate and lazy)
- Target definitions (file, phony, directory, pattern)
- Environment definitions (named and default)
- Variables inside conditionals (tracked separately in `ConditionalVars`)

#### Conditional Variable Handling

Variables defined in conditional branches receive special handling:

1. **ConditionalVars tracking**: All conditional variable definitions are tracked in `SymbolTable.ConditionalVars` for runtime evaluation to select the correct value.

2. **Variables map**: Each variable name appears in `Variables` map only once (first definition), enabling reference validation without false duplicate errors between branches.

Example:
```
if {os} == linux
cc = gcc          # Added to ConditionalVars[cc] and Variables["cc"]
elif {os} == darwin
cc = clang        # Added to ConditionalVars[cc] only
else
cc = cc           # Added to ConditionalVars[cc] only
end
```

#### ConditionalVarDef Structure

```go
type ConditionalVarDef struct {
    Variable    *ast.Variable       // The variable definition
    Conditional *ast.Conditional    // The containing conditional
    BranchType  string              // "if", "elif", or "else"
    BranchIndex int                 // For elif, the index (0-based); -1 for if/else
}
```

#### Error Collection

The collector continues processing after errors, collecting all duplicate definition errors:

```go
st, errs := Collect(stmts)
if len(errs) > 0 {
    // errs contains DuplicateDefinitionError instances
}
```

#### Nested Conditional Support

Nested conditionals are handled recursively. Each level maintains its own set of conditional variables to avoid false duplicate detection:

```
if {os} == linux
    ifdef DEBUG
        debug_flags = -g  # Nested conditional
    end
    cc = gcc
end
```

#### Design Decisions

1. **Error collection vs early exit**: All errors are collected to provide comprehensive feedback rather than stopping at the first error.

2. **Conditional variable tracking**: Separate tracking allows runtime to evaluate conditions and select the correct definition without requiring re-parsing.

3. **First-definition wins for Variables map**: For reference validation purposes, the first definition is stored in `Variables`. Runtime uses `ConditionalVars` to determine the actual value.

4. **Nested conditional isolation**: Each conditional tracks its own variables to prevent interference between parent and child conditionals.

5. **Targets and environments in conditionals**: Though less common, these are allowed and follow the same duplicate detection rules as top-level definitions.

### Capture Validation (`capture.go`)

Pass 2 of semantic analysis: resolves `BraceExpr` nodes in target patterns to either captures or interpolations.

#### Resolution Rules

For each `{name}` in a target pattern:

1. If `name` is a defined variable (user, built-in, or conditional) → **Interpolation**
2. If `name` is an automatic variable → **Error** (automatic variables are runtime-only)
3. Otherwise → **Capture** (pattern matching variable)

#### CaptureResult Structure

```go
type CaptureResult struct {
    Captures       map[*ast.Target]*CaptureInfo       // Targets with captures
    Interpolations map[*ast.Target]*InterpolationInfo // Targets with interpolations
    Errors         []error                             // Validation errors
}

type CaptureInfo struct {
    Names []string // Unique capture names in order of first appearance
}

type InterpolationInfo struct {
    Names []string // Variable names used as interpolations
}
```

#### Validation Rules

1. **Automatic variables in patterns**: Using automatic variables like `{target}`, `{deps}`, `{in}`, etc. in a target pattern is an error. These are only available during recipe execution.

2. **Capture mismatch**: If a `{name}` appears in a dependency but is not defined in the target pattern and is not a defined variable, it's an error. The capture value would be unknown.

3. **Captures in target with literal dependencies**: A target pattern can have captures even if its dependencies are literal. This is valid for pattern targets.

#### Error Types

```go
type AutomaticInPatternError struct {
    Name     string             // The automatic variable name
    Location ast.SourceLocation // Location of the usage
}

type CaptureMismatchError struct {
    Name      string             // The capture name
    InTarget  bool               // true if capture is in target but not deps
    Location  ast.SourceLocation // Location of the problematic usage
    TargetLoc ast.SourceLocation // Location of the target definition
}
```

#### Example

```
base = build

# {base} is an interpolation (defined variable)
# {name} is a capture (undefined → pattern matching)
{base}/{name}.o: src/{name}.c
    gcc -c {in} -o {out}
```

After capture validation:
- Target `{base}/{name}.o` has:
  - Interpolations: `[base]`
  - Captures: `[name]`

#### Design Decisions

1. **Automatic variables are errors**: Using `{target}` or `{deps}` in a pattern doesn't make sense—they're computed during recipe execution, not pattern matching.

2. **Built-in variables are interpolations**: Variables like `{os}` and `{arch}` can be used in patterns to create platform-specific targets.

3. **Capture order preserved**: Captures are tracked in order of first appearance for deterministic pattern matching.

4. **Dual tracking**: Both captures and interpolations are tracked per-target to enable proper pattern compilation and variable substitution during build planning.

5. **Dependency validation**: Captures in dependencies must be defined in the target pattern. A dependency cannot introduce a new capture.

### Reference Validation (`reference.go`)

Pass 3 of semantic analysis: validates that all variable references in the AST point to defined symbols.

#### Validation Rules

For each interpolation reference in values/commands:

1. **Automatic variables**: `{target}`, `{deps}`, `{in}`, `{out}`, `{stem}`, `{target.dir}`, `{target.file}` are only valid inside recipe or block scope. Using them elsewhere is an error.

2. **Captures**: Capture variables (like `{name}` in pattern targets) are only valid inside the recipe that defines them. They cannot be used in global variable values.

3. **User-defined variables**: References to user-defined variables (immediate or lazy) must point to a defined variable.

4. **Conditional variables**: Variables defined in conditionals are recognized as defined for reference validation purposes.

5. **Built-in variables**: `{os}` and `{arch}` are always valid anywhere.

#### ReferenceResult Structure

```go
type ReferenceResult struct {
    Errors []error // Validation errors
}
```

#### Error Types

```go
type UndefinedVariableError struct {
    Name     string             // The undefined variable name
    Location ast.SourceLocation // Location of the reference
}

type AutomaticOutsideRecipeError struct {
    Name     string             // The automatic variable name
    Location ast.SourceLocation // Location of the reference
}
```

#### Validation Scope

Reference validation checks:
- Variable values (interpolations in RHS of `=`)
- Directive values (`.default:`, `.shell:`, etc.)
- Function arguments (`shell(...)`, `glob(...)`, etc.)
- Conditional conditions (`if {var} == value`)
- Conditional body statements
- Environment directives (`.source:`, `.args:`)
- Recipe directives (`.shell:`, `.after:`, `.autodeps:`)
- Recipe commands (line and block)

#### Options

The `ValidateReferences` function accepts options to configure validation:

```go
// WithCaptures provides capture information from Pass 2
func WithCaptures(captureResult *CaptureResult) ReferenceOption
```

When capture information is provided, capture names are recognized as valid references within their defining recipe.

#### Design Decisions

1. **Capture scope enforcement**: Captures are recipe-scoped. A pattern target's captures can only be used in that target's recipe, not in global variables or other recipes.

2. **Automatic variable scope enforcement**: Automatic variables like `{target}` are only computed during recipe execution, so they cannot be used in variable definitions or directive values.

3. **Error collection continues**: Like other passes, reference validation collects all errors rather than stopping at the first one.

4. **Built-in always valid**: Built-in variables `os` and `arch` are valid everywhere, enabling platform-specific logic in any context.

5. **Conditional variables recognized**: Variables defined inside conditionals are treated as defined for reference validation, even though their actual value depends on which branch executes at runtime.

### Dependency Graph Validation (`depgraph.go`)

Pass 4 of semantic analysis: builds the dependency graph and detects circular dependencies.

#### DependencyGraph Structure

```go
type DependencyGraph struct {
    Nodes map[string]bool    // All target names in the graph
    Edges map[string][]string // Maps each target to its dependencies
}
```

#### DependencyResult Structure

```go
type DependencyResult struct {
    Graph           *DependencyGraph   // The constructed dependency graph
    PatternTargets  []*ast.Target      // Targets with pattern captures (rules)
    UnsatisfiedDeps map[string][]string // Dependencies not defined as targets
    Errors          []error             // Validation errors (e.g., cycles)
}
```

#### ValidateDependencies Function

```go
func ValidateDependencies(targets []*ast.Target) *DependencyResult
```

The function:
1. Identifies pattern targets (targets with `BraceExpr` segments) and tracks them separately
2. Builds a graph of concrete targets (non-pattern) and their dependencies
3. Tracks unsatisfied dependencies (deps that aren't defined as targets)
4. Detects circular dependencies using DFS-based cycle detection

#### Cycle Detection Algorithm

Uses depth-first search (DFS) with a recursion stack:

```go
func findCycle(g *DependencyGraph) []string
func dfs(g *DependencyGraph, node string, visited, recStack map[string]bool, parent map[string]string) []string
func reconstructCycle(parent map[string]string, end, cycleStart string) []string
```

When a cycle is found, it reconstructs the full cycle path for clear error reporting.

#### Error Types

```go
type CircularDependencyError struct {
    Cycle []string // e.g., ["a", "b", "c", "a"]
}
```

Error format: `circular dependency detected: a -> b -> c -> a`

#### Key Behaviors

| Scenario | Result |
|----------|--------|
| Self-loop (`a: a`) | Circular dependency error |
| Two-node cycle (`a: b`, `b: a`) | Circular dependency error |
| Diamond dependency (`a: b c`, `b: d`, `c: d`, `d`) | Valid (no cycle) |
| Pattern target | Stored in `PatternTargets`, not in graph |
| Dependency not defined as target | Stored in `UnsatisfiedDeps` |
| Phony targets | Participate in graph normally |

#### Design Decisions

1. **Pattern targets excluded from graph**: Pattern targets define rules, not concrete nodes. They're tracked separately for build planning to use during pattern matching.

2. **Unsatisfied deps are not errors**: A dependency not defined as a target may be a source file or may be satisfiable by a pattern target. The build planner will resolve this later.

3. **Single cycle reported**: When multiple cycles exist, only the first detected cycle is reported. This keeps error messages focused.

4. **Cycle path includes start node twice**: The cycle path includes the starting node at both ends (e.g., `a -> b -> c -> a`) to clearly show where the cycle closes.

5. **Only graph nodes checked for cycles**: Dependencies that aren't graph nodes (unsatisfied deps) are ignored during cycle detection—they can't form cycles.

## Test Coverage

### Parser Unit Tests

The parser package has comprehensive test coverage across all parsing features:

| Test File | Coverage |
|-----------|----------|
| `parser_test.go` | Parser initialization, scope transitions, directive validation |
| `scope_test.go` | Scope stack operations (push/pop, IsIn, reset, nesting) |
| `directive_test.go` | Directive scope validation for all scopes |
| `errors_test.go` | Error formatting, hints, scope errors, error collection |
| `variable_test.go` | Variables (simple, lazy, interpolation, functions, replace) |
| `target_test.go` | Targets (simple, captures, phony, directory, patterns, dependencies) |
| `recipe_test.go` | Recipes (commands, blocks, directives, interpolations, versions) |
| `environment_test.go` | Environment blocks (all runtimes, named/default, directives, version specs) |
| `conditional_test.go` | Conditionals (if/elif/else/end, ifdef/ifndef, ==, !=, multiple branches) |
| `include_test.go` | Include directive (simple, nested, circular, relative paths, empty files) |
| `recovery_test.go` | Error recovery (skip-to-level-0, actionable messages, max errors) |
| `parser_integration_test.go` | Full `ParseNeedfile` integration tests |
| `edge_cases_test.go` | Edge cases and negative tests |

### Integration Tests (`parser_integration_test.go`)

Tests `ParseNeedfile()` end-to-end parsing:

| Test | Description |
|------|-------------|
| `TestParseNeedfile_AllStatementTypes` | Parses needfile with all statement types |
| `TestParseNeedfile_DirectiveDetails` | Verifies directive parsing details |
| `TestParseNeedfile_VariableDetails` | Verifies variable parsing with interpolations |
| `TestParseNeedfile_TargetDetails` | Verifies target and recipe parsing |
| `TestParseNeedfile_EnvironmentDetails` | Verifies environment block parsing |
| `TestParseNeedfile_ConditionalDetails` | Verifies conditional branch parsing |
| `TestParseNeedfile_NestedBlocks` | Tests recipe with block command |
| `TestParseNeedfile_SourceLocations` | Verifies source location tracking |
| `TestParseNeedfile_EmptyFile` | Tests empty file handling |
| `TestParseNeedfile_OnlyComments` | Tests comment-only file |
| `TestParseNeedfile_MixedWithBlankLines` | Tests blank line handling |
| `TestParseNeedfile_ErrorRecoveryIntegration` | Tests error recovery in full needfile |

### Edge Case Tests (`edge_cases_test.go`)

| Test | Description |
|------|-------------|
| `TestEdgeCase_NestedConditionals` | Conditionals inside conditionals |
| `TestEdgeCase_DirectiveInWrongScope` | Directive scope validation errors |
| `TestEdgeCase_EnvironmentInRecipe` | .environment inside recipe (error) |
| `TestEdgeCase_ParallelInEnvironment` | .parallel inside environment (error) |
| `TestEdgeCase_DefaultInRecipe` | .default inside recipe (error) |
| `TestEdgeCase_DeeplyNestedBlocks` | Three levels of nested conditionals |
| `TestEdgeCase_MultipleErrorsCollected` | Multiple errors collected before max |
| `TestEdgeCase_RecipeWithOnlyDirectives` | Recipe with directives but no commands |
| `TestEdgeCase_TargetWithNoRecipe` | Targets without recipes |
| `TestEdgeCase_VersionSpecFormats` | All version spec formats |
| `TestEdgeCase_FunctionCallsInCommands` | Interpolations in commands |
| `TestEdgeCase_EscapedBracesInValue` | `{{` and `}}` escape sequences |
| `TestEdgeCase_CommentAfterStatement` | Inline comments |
| `TestEdgeCase_PathWithInterpolation` | Patterns with captures |

### CLI Tests (`cmd/need/main_test.go`)

All debug flags have corresponding tests:

| Test | Flag |
|------|------|
| `TestRunDebugLex` | `--debug-lex` |
| `TestRunDebugParse` | `--debug-parse` |
| `TestRunDebugVar` | `--debug-var` |
| `TestRunDebugTarget` | `--debug-target` |
| `TestRunDebugRecipe` | `--debug-recipe` |
| `TestRunDebugEnv` | `--debug-env` |
| `TestRunDebugCond` | `--debug-cond` |
| `TestRunDebugAST` | `--debug-ast` |
| `TestRunDebugSemantic` | `--debug-semantic` |
| `TestRunDebugEval` | `--debug-eval` |
| `TestRunDebugPlan` | `--debug-plan` |

Each debug test includes:
- Success case with valid needfile
- Missing file error case
- (Some) edge cases like empty files or files without target content

#### Command Interpolation Tests

| Test | Description |
|------|-------------|
| `TestRunDebugPlanCommandInterpolation` | Automatic variable resolution in commands |
| `TestRunDebugPlanCaptureVariables` | Capture variable substitution in commands |
| `TestRunDebugPlanTargetDirAndFile` | target.dir and target.file automatic variables |
| `TestRunDebugPlanStemVariable` | Stem variable with pattern targets |
| `TestRunDebugPlanRawModifier` | Raw modifier for unquoted expansion |
| `TestRunDebugPlanBlockCommandInterpolation` | Block command interpolation |
| `TestRunDebugPlanDepsVariable` | deps variable with multiple dependencies |
| `TestRunDebugPlanMixedVariables` | Mixed user-defined, automatic, and builtin variables |

#### Shell Execution Tests

| Test | Description |
|------|-------------|
| `TestRunDebugPlanWithShellDirective` | Global shell directive handling |
| `TestRunDebugPlanWithRecipeShellOverride` | Recipe-level shell override |

#### Parallel Execution Tests

| Test | Description |
|------|-------------|
| `TestRunDebugPlanWithParallelDirective` | Parallel directive handling |
| `TestRunDebugPlanDiamondDependency` | Diamond dependency for parallel scheduling |

### Semantic Unit Tests (`internal/semantic`)

#### Error Type Tests (`errors_test.go`)

| Test | Description |
|------|-------------|
| `TestDuplicateDefinitionError_Variable` | Variable duplicate error message |
| `TestDuplicateDefinitionError_Target` | Target duplicate error message |
| `TestDuplicateDefinitionError_Environment` | Environment duplicate error message |
| `TestAutomaticInPatternError_Target` | Automatic variable in pattern error |
| `TestAutomaticInPatternError_AllAutomaticVars` | All automatic variables in pattern |
| `TestCaptureMismatchError_ExtraInDependency` | Capture in dependency not in target |
| `TestUndefinedVariableError_Basic` | Basic undefined variable error |
| `TestUndefinedVariableError_DottedName` | Dotted name undefined error |
| `TestAutomaticOutsideRecipeError_InVariable` | Automatic variable in variable value |
| `TestAutomaticOutsideRecipeError_AllAutomaticVars` | All automatic variables outside recipe |
| `TestCircularDependencyError_SelfLoop` | Self-loop cycle error |
| `TestCircularDependencyError_TwoNodes` | Two-node cycle error |
| `TestCircularDependencyError_LongCycle` | Long cycle error |
| `TestAllErrorsImplementError` | All error types implement error interface |
| `TestErrorsIncludeLocation` | All errors include source location |

#### Symbol Table Tests (`symbols_test.go`)

| Test | Description |
|------|-------------|
| `TestNewSymbolTable` | Symbol table initialization with automatic/built-in vars |
| `TestSymbolTable_AddVariable` | Adding variables to symbol table |
| `TestSymbolTable_DuplicateVariable` | Duplicate variable detection |
| `TestSymbolTable_AddTarget` | Adding targets to symbol table |
| `TestSymbolTable_DuplicateTarget` | Duplicate target pattern detection |
| `TestSymbolTable_AddPhonyTarget` | Adding phony targets |
| `TestSymbolTable_DuplicatePhonyTarget` | Duplicate phony target detection |
| `TestSymbolTable_PatternTargetsAllowed` | Different patterns don't conflict |
| `TestSymbolTable_AddEnvironment` | Adding environments (default and named) |
| `TestSymbolTable_DuplicateEnvironment` | Duplicate named environment detection |
| `TestSymbolTable_DuplicateDefaultEnvironment` | Duplicate default environment detection |
| `TestSymbolTable_IsAutomatic` | Automatic variable check |
| `TestSymbolTable_IsBuiltin` | Built-in variable check |
| `TestSymbolTable_LookupVariable` | Variable lookup |
| `TestSymbolTable_IsDefined` | Definition check (user, automatic, built-in) |
| `TestSymbolTable_TargetPatternString` | Pattern string generation |
| `TestDuplicateDefinitionError_Error` | Error message format |

#### Collector Tests (`collector_test.go`)

| Test | Description |
|------|-------------|
| `TestCollector_Basic` | Basic collection of variables, targets, environments |
| `TestCollector_DuplicateVariable` | Duplicate variable detection |
| `TestCollector_DuplicateTarget` | Duplicate target detection |
| `TestCollector_DuplicateEnvironment` | Duplicate named environment detection |
| `TestCollector_DuplicateDefaultEnvironment` | Duplicate default environment detection |
| `TestCollector_ConditionalVariables` | Variables in if/elif/else branches |
| `TestCollector_NestedConditionals` | Nested conditionals with variables |
| `TestCollector_MultipleTargets` | Multiple target collection |
| `TestCollector_PatternTargets` | Pattern targets with captures |
| `TestCollector_MultipleErrors` | Multiple errors collected |
| `TestCollector_LazyVariables` | Lazy variable collection |
| `TestCollector_EmptyNeedfile` | Empty needfile handling |
| `TestCollector_CommentsAndBlanks` | Comments and blanks ignored |
| `TestCollector_GlobalDirectives` | Directives don't add symbols |
| `TestCollector_PreservesOrder` | Target order preserved |
| `TestCollector_FileTargets` | File target collection |
| `TestCollector_DirectoryTargets` | Directory target collection |
| `TestCollector_MixedEnvironments` | Named and default environments |

#### Capture Validation Tests (`capture_test.go`)

| Test | Description |
|------|-------------|
| `TestValidateCaptures_NoBraceExprs` | Targets without brace expressions |
| `TestValidateCaptures_SimpleCapture` | Simple capture pattern resolution |
| `TestValidateCaptures_VariableInterpolation` | Defined variables become interpolations |
| `TestValidateCaptures_AutomaticVariableError` | Automatic variables in patterns are errors |
| `TestValidateCaptures_BuiltinInPattern` | Built-in variables (os, arch) in patterns |
| `TestValidateCaptures_MultipleCaptures` | Multiple captures in same pattern |
| `TestValidateCaptures_CaptureMismatch_MissingInDependency` | Pattern target with literal deps (valid) |
| `TestValidateCaptures_CaptureMismatch_ExtraInDependency` | Capture in dep not in target (error) |
| `TestValidateCaptures_CaptureAndInterpolationMixed` | Mix of captures and interpolations |
| `TestValidateCaptures_PhonyTarget` | Phony targets without captures |
| `TestValidateCaptures_PhonyTargetWithBraceExpr` | Phony targets with interpolation |
| `TestValidateCaptures_PhonyTargetWithCapture` | Phony targets with capture |
| `TestValidateCaptures_DuplicateCaptureInPattern` | Duplicate captures treated as single |
| `TestValidateCaptures_ConditionalVariable` | Conditional variables recognized as interpolations |
| `TestAutomaticInPatternError_Error` | Error message format |
| `TestCaptureMismatchError_Error` | Error message format |

#### Reference Validation Tests (`reference_test.go`)

| Test | Description |
|------|-------------|
| `TestValidateReferences_DefinedVariable` | Defined variable references are valid |
| `TestValidateReferences_UndefinedVariable` | Undefined variable references are errors |
| `TestValidateReferences_BuiltinVariable` | Built-in variables (os, arch) are always valid |
| `TestValidateReferences_ConditionalVariable` | Conditional variables are recognized |
| `TestValidateReferences_AutomaticInVariableValue` | Automatic variables in variable values are errors |
| `TestValidateReferences_AutomaticInRecipe` | Automatic variables in recipes are valid |
| `TestValidateReferences_AllAutomaticVariables` | All automatic variables valid in recipe |
| `TestValidateReferences_AutomaticOutsideRecipe_AllVars` | All automatic variables invalid outside recipe |
| `TestValidateReferences_CaptureInRecipe` | Captures are valid in their defining recipe |
| `TestValidateReferences_CaptureOutsideRecipe` | Captures outside recipe are undefined |
| `TestValidateReferences_DirectiveValue` | References in directive values |
| `TestValidateReferences_DirectiveUndefinedVariable` | Undefined in directive is error |
| `TestValidateReferences_FunctionArgument` | References in function arguments |
| `TestValidateReferences_FunctionUndefinedArgument` | Undefined in function arg is error |
| `TestValidateReferences_ConditionalCondition` | References in condition expressions |
| `TestValidateReferences_ConditionalUndefinedCondition` | Undefined in condition is error |
| `TestValidateReferences_ConditionalBody` | References in conditional body |
| `TestValidateReferences_EnvironmentSource` | References in environment source |
| `TestValidateReferences_EnvironmentUndefinedSource` | Undefined in environment is error |
| `TestValidateReferences_RecipeShellDirective` | References in recipe directives |
| `TestValidateReferences_BlockCommand` | References in block commands |
| `TestValidateReferences_MultipleErrors` | Multiple errors collected |
| `TestUndefinedVariableError_Error` | Error message format |
| `TestAutomaticOutsideRecipeError_Error` | Error message format |

#### Dependency Graph Validation Tests (`depgraph_test.go`)

| Test | Description |
|------|-------------|
| `TestValidateDependencies_NoDependencies` | Targets with no dependencies |
| `TestValidateDependencies_SimpleDependency` | A depends on B |
| `TestValidateDependencies_ChainDependencies` | A -> B -> C -> D (linear chain) |
| `TestValidateDependencies_DiamondDependencies` | Diamond dependency (A -> B, A -> C, B -> D, C -> D) |
| `TestValidateDependencies_DirectCycle` | Self-loop detection (A -> A) |
| `TestValidateDependencies_TwoNodeCycle` | Two-node cycle (A -> B -> A) |
| `TestValidateDependencies_ThreeNodeCycle` | Three-node cycle (A -> B -> C -> A) |
| `TestValidateDependencies_CycleInSubgraph` | Cycle in part of graph |
| `TestValidateDependencies_UnsatisfiedDependency` | Dependency not defined as target |
| `TestValidateDependencies_MultipleUnsatisfiedDependencies` | Multiple unsatisfied deps |
| `TestValidateDependencies_PhonyTargets` | Phony targets in graph |
| `TestValidateDependencies_PhonyCycle` | Cycle with phony targets |
| `TestValidateDependencies_PatternTargets` | Pattern targets tracked separately |
| `TestValidateDependencies_MixedPatternAndLiteral` | Mix of pattern and literal targets |
| `TestCircularDependencyError_Error` | Error message format |
| `TestDependencyResult_Graph` | Graph structure verification |
| `TestValidateDependencies_EmptyTargets` | Empty target list handling |
| `TestValidateDependencies_SingleTargetNoDeps` | Single target with no dependencies |
| `TestValidateDependencies_MultipleCycles` | Multiple separate cycles |
| `TestValidateDependencies_LongCycle` | Long cycle (6 nodes) |


## Eval Package (`internal/eval`)

The eval package provides variable evaluation for Needfiles. It evaluates variables after semantic analysis and before build planning.

### Context (`context.go`)

The evaluation context stores all variable values during evaluation.

#### Context Structure

```go
type Context struct {
    variables     map[string]string   // Evaluated variable values
    lazyVariables map[string]*ast.Value // Unevaluated lazy variable AST values
    lazyCache     map[string]string   // Cached lazy variable evaluations
    builtins      map[string]string   // Read-only built-in variables (os, arch)
    shellCache    map[string]string   // Cached shell() function results
}
```

#### Built-in Variables

| Variable | Value | Description |
|----------|-------|-------------|
| `os` | `runtime.GOOS` | Operating system name |
| `arch` | `runtime.GOARCH` | Architecture name |

#### Key Methods

| Method | Description |
|--------|-------------|
| `NewContext()` | Creates context with built-ins initialized |
| `Get(name)` | Returns variable value (built-ins first) |
| `Set(name, value)` | Sets a variable (built-ins protected) |
| `IsDefined(name)` | Returns true if variable is defined |
| `SetLazyValue(name, value)` | Stores a lazy variable's AST value |
| `GetLazyValue(name)` | Gets a lazy variable's AST value |
| `IsLazy(name)` | Returns true if variable is lazy |
| `CacheLazyResult(name, value)` | Caches a lazy variable evaluation result |
| `Variables()` | Returns all evaluated variables |
| `LazyVariables()` | Returns all lazy variables |
| `GetShellCache(cmd)` | Gets cached shell() result for command |
| `SetShellCache(cmd, output)` | Caches shell() result for command |
| `ClearShellCache()` | Clears all cached shell() results |

### Evaluator (`evaluator.go`)

The evaluator evaluates AST values using the context.

#### Key Methods

| Method | Description |
|--------|-------------|
| `NewEvaluator(ctx)` | Creates evaluator with context |
| `SetVerboseOutput(w)` | Sets output writer for verbose mode |
| `EvaluateValue(val)` | Evaluates AST Value to string |
| `EvaluateVariables(stmts)` | Evaluates all variables in statement list |
| `EvaluateCondition(cond)` | Evaluates a condition (==, !=, ifdef, ifndef) |
| `EvaluateConditional(cond)` | Evaluates a conditional block and its body |

#### Verbose Mode

When `SetVerboseOutput()` is called with a writer, the evaluator prints variable evaluation results:
- For immediate variables: `name = value`
- For lazy variables: `name = <lazy>`

#### Value Evaluation

Values are evaluated by processing each part:
- `LiteralValue`: Appended directly
- `Interpolation`: Variable resolved from context
- `FunctionCall`: Function executed

#### Conditional Evaluation

Conditionals are evaluated by:
1. Evaluating the if branch condition
2. If true, execute the if body statements
3. Otherwise, try each elif branch in order
4. If no match, execute the else body (if present)

### Built-in Functions (`functions.go`)

| Function | Description |
|----------|-------------|
| `shell(cmd)` | Executes shell command, returns stdout (cached within build) |
| `glob(pattern)` | Returns space-separated list of matching files |
| `basename(path)` | Extracts filename from path |
| `dirname(path)` | Extracts directory from path |
| `replace(str, from, to)` | Replaces all occurrences of from with to |

#### Shell Quoting in `shell()` Function

The `shell()` function applies shell quoting to interpolated values to ensure safety when values contain spaces or special characters:

| Syntax | Behavior | Example |
|--------|----------|---------|
| `{var}` | Shell-quoted (wrapped in single quotes) | `find '{src_dir}'` |
| `{var:raw}` | Unquoted (allows word splitting) | `gcc -Wall -O2` |

**Implementation:**

The `shell()` function uses `evaluateShellCommand()` instead of `EvaluateValue()` to process its argument. This method:
1. For `{var}` (non-raw): Applies `ShellQuote()` which wraps the value in single quotes and handles embedded single quotes
2. For `{var:raw}`: Inserts the value directly without quoting

**Examples:**

```
src_dir = my sources
flags = -Wall -O2

# Quoted (default) — safe for paths with spaces
sources = shell(find {src_dir} -name "*.c")
# Executes: find 'my sources' -name "*.c"

# Raw — for flags that need word splitting
result = shell(gcc {flags:raw} -c main.c)
# Executes: gcc -Wall -O2 -c main.c
```

**ShellQuote Function:**

```go
func ShellQuote(s string) string
```

Uses single-quote quoting with special handling for embedded single quotes:
- Simple case: `'value'`
- With embedded quotes: `'it'"'"'s working'` (ends quote, adds double-quoted single quote, restarts quote)

#### Shell Caching

The `shell()` function caches results within a build to avoid redundant shell executions:

**Behavior:**
- Results are cached by the evaluated command string
- Same command called multiple times returns cached result
- Only successful executions are cached; errors are not cached
- Different interpolation values produce different cache keys
- Cache is per-Context (cleared between builds)

**Use case:**
```
# Without caching, this would run 'date' twice
timestamp = shell(date +%s)
other = prefix-{timestamp}-suffix
log = shell(date +%s)-log   # Returns same timestamp due to cache
```

**Implementation:**
1. Evaluate command string with interpolations
2. Check `Context.shellCache` for cached result
3. If found, return cached value
4. Otherwise, execute command via `/bin/sh -c "command"`
5. Cache successful result for future calls
6. Do NOT cache errors (allow retry on failure)

**Methods:**

| Method | Description |
|--------|-------------|
| `GetShellCache(cmd)` | Gets cached result for evaluated command |
| `SetShellCache(cmd, output)` | Stores result in cache |
| `ClearShellCache()` | Clears all cached results |

### Error Types

| Error | Description |
|-------|-------------|
| `UndefinedVariableError` | Variable reference could not be resolved |

### Eval Unit Tests

#### Context Tests (`context_test.go`)

| Test | Description |
|------|-------------|
| `TestNewContext` | Context initialization with built-ins |
| `TestContext_SetAndGet` | Variable set/get |
| `TestContext_GetUndefined` | Undefined variable returns ok=false |
| `TestContext_IsDefined` | Definition check for all variable types |
| `TestContext_SetLazy` | Lazy variable storage |
| `TestContext_IsLazy` | Lazy variable detection |
| `TestContext_Overwrite` | Variable overwrite |
| `TestContext_BuiltinsAreReadOnly` | Built-in protection |
| `TestContext_Variables` | Variables() returns all variables |
| `TestContext_LazyVariables` | LazyVariables() returns lazy vars |

#### Evaluator Tests (`evaluator_test.go`)

| Test | Description |
|------|-------------|
| `TestEvaluateValue_Literal` | Simple literal value |
| `TestEvaluateValue_MultipleLiterals` | Multiple literal parts |
| `TestEvaluateValue_Interpolation` | Variable interpolation |
| `TestEvaluateValue_BuiltinInterpolation` | Built-in variable reference |
| `TestEvaluateValue_UndefinedVariable` | Error for undefined variable |
| `TestEvaluateValue_MixedLiteralsAndInterpolations` | Combined parts |
| `TestEvaluateValue_NilValue` | Nil value returns empty string |
| `TestEvaluateValue_EmptyValue` | Empty value returns empty string |
| `TestEvaluateValue_RawModifier` | Raw modifier handling |
| `TestUndefinedVariableError_Error` | Error message format |

#### Conditional Tests (`conditional_test.go`)

| Test | Description |
|------|-------------|
| `TestEvaluateCondition_Equals_True` | Equals condition that evaluates to true |
| `TestEvaluateCondition_Equals_False` | Equals condition that evaluates to false |
| `TestEvaluateCondition_NotEquals_True` | Not-equals condition that evaluates to true |
| `TestEvaluateCondition_NotEquals_False` | Not-equals condition that evaluates to false |
| `TestEvaluateCondition_Defined_True` | Ifdef condition for defined variable |
| `TestEvaluateCondition_Defined_False` | Ifdef condition for undefined variable |
| `TestEvaluateCondition_NotDefined_True` | Ifndef condition for undefined variable |
| `TestEvaluateCondition_NotDefined_False` | Ifndef condition for defined variable |
| `TestEvaluateCondition_BuiltinDefined` | Built-in variables are always defined |
| `TestEvaluateConditional_IfTrue` | If branch executes when condition true |
| `TestEvaluateConditional_IfFalse_NoElse` | No action when condition false and no else |
| `TestEvaluateConditional_ElifBranch` | Elif branch executes when matched |
| `TestEvaluateConditional_ElseBranch` | Else branch executes as fallback |
| `TestEvaluateConditional_Ifdef` | Ifdef conditional execution |
| `TestEvaluateConditional_Ifndef` | Ifndef conditional execution |
| `TestEvaluateConditional_NestedConditionals` | Nested conditional evaluation |

#### Variables Tests (`variables_test.go`)

| Test | Description |
|------|-------------|
| `TestEvaluateVariables_EmptyStatements` | Empty statement list |
| `TestEvaluateVariables_SimpleVariable` | Simple immediate variable |
| `TestEvaluateVariables_MultipleVariables` | Multiple variables in order |
| `TestEvaluateVariables_VariableReference` | Variable referencing another |
| `TestEvaluateVariables_ForwardReferenceError` | Forward reference causes error |
| `TestEvaluateVariables_LazyVariable` | Lazy variable stored for later |
| `TestEvaluateVariables_BuiltinReference` | Built-in variable access |
| `TestEvaluateVariables_SkipsNonVariables` | Non-variable statements skipped |
| `TestEvaluateVariables_ChainedReferences` | Chained variable references |
| `TestEvaluateVariables_LazyVariableOnDemand` | Lazy variable evaluated on demand |
| `TestEvaluateVariables_LazyVariableWithInterpolation` | Lazy variable with interpolation |
| `TestEvaluateVariables_LazyVariableReferencesLater` | Lazy can reference later variables |
| `TestEvaluateVariables_LazyVariableCaching` | Lazy variable result is cached |
| `TestEvaluateVariables_WithConditional` | Conditional in statement list |
| `TestEvaluateVariables_VerboseMode` | Verbose mode outputs variable evaluations |

#### Functions Tests (`functions_test.go`)

| Test | Description |
|------|-------------|
| `TestFuncBasename_Simple` | Basic basename extraction |
| `TestFuncBasename_TrailingSlash` | Basename with trailing slash |
| `TestFuncDirname_Simple` | Basic dirname extraction |
| `TestFuncDirname_RootPath` | Dirname of root path |
| `TestFuncReplace_Simple` | Basic string replacement |
| `TestFuncReplace_Multiple` | Multiple occurrence replacement |
| `TestFuncReplace_NoMatch` | No match returns input unchanged |
| `TestFuncGlob_CurrentDir` | Glob in current directory |
| `TestFuncGlob_NoMatches` | Glob with no matches |
| `TestFuncShell_Echo` | Shell echo command |
| `TestFuncShell_TrimNewline` | Trailing newline trimmed |
| `TestFuncShell_WithInterpolation` | Shell with variable interpolation |
| `TestFuncShell_FailingCommand` | Shell command failure error |
| `TestFuncComposition_DirnameBasename` | Nested function calls |
| `TestFuncComposition_ReplaceInGlob` | Function composition |
| `TestFuncError_UndefinedInArg` | Undefined variable in function arg |
| `TestFuncShell_MultilineOutput` | Shell with multiline output |
| `TestFuncShell_CachingBasic` | Shell result caching for identical commands |
| `TestFuncShell_CachingWithSameCommand` | Same command returns cached result |
| `TestFuncShell_CachingDifferentCommands` | Different commands not sharing cache |
| `TestFuncShell_CachingWithInterpolation` | Cache key includes interpolated values |
| `TestFuncShell_CachingErrorsNotCached` | Errors not cached (allow retry) |
| `TestContext_ShellCacheOperations` | Shell cache get/set/clear operations |

## Planner Package (`internal/planner`)

The planner package provides build planning for Needfiles. It handles target pattern matching, dependency resolution, and build task ordering.

### Target Pattern Matching (`match.go`)

Implements target pattern matching for both literal and pattern targets.

#### MatchTarget Function

```go
func MatchTarget(pattern *ast.TargetPattern, path string) (bool, map[string]string)
```

Matches a concrete path against a target pattern. Returns whether the pattern matches and a map of capture values.

**Matching rules:**
- Literal segments must match exactly
- Captures (`{name}`) match any sequence of characters (including slashes)
- Duplicate capture names must have the same value
- Phony targets match with or without `@` prefix (@ is only for declaration)
- Directory targets can match with or without trailing slash

#### LookupTarget Function

```go
func LookupTarget(path string, targets []*ast.Target) (*ast.Target, map[string]string, error)
```

Finds a target definition that matches the given path.

**Lookup order:**
1. Exact literal matches are preferred over pattern matches
2. Among patterns, first match in definition order wins

Returns the matching target, capture values, and any error.

### Error Types

| Type | Description |
|------|-------------|
| `TargetNotFoundError` | No rule matches the requested target path |

### Design Decisions

1. **Capture greedy matching**: When multiple captures are adjacent (e.g., `{a}{b}`), the first capture is greedy and takes as much as possible while still allowing the rest to match.

2. **Shortest match for literals**: When a capture is followed by a literal, the matcher finds the first position where the literal appears, preferring shorter captures.

3. **Duplicate capture consistency**: If the same capture name appears multiple times in a pattern (e.g., `{name}/{name}.o`), all occurrences must capture the same value.

4. **Directory target flexibility**: Directory targets (`build/`) match both `build/` and `build` for convenience.

5. **Phony target reference flexibility**: Phony targets can be referenced with or without the `@` prefix. The `@` is only required when *declaring* a phony target, not when referencing it in dependencies. This allows natural syntax like `@all: build test` where `build` and `test` are phony targets.

6. **Exact match preference**: When both a literal target and a pattern target could match, the literal target is preferred. This allows overriding pattern rules for specific files.

### Planner Unit Tests (`match_test.go`)

#### Literal Target Matching Tests

| Test | Description |
|------|-------------|
| `TestMatchLiteral_ExactMatch` | Exact literal path matching |
| `TestMatchLiteral_NoMatch` | Different path doesn't match |
| `TestMatchLiteral_PartialNoMatch` | Partial matches don't count |
| `TestMatchLiteral_PhonyTarget` | Phony target with or without @ prefix |
| `TestMatchLiteral_PhonyTargetNoMatch` | Different phony name doesn't match |
| `TestMatchLiteral_DirectoryTarget` | Directory with trailing slash |
| `TestMatchLiteral_DirectoryTargetNoTrailingSlash` | Directory without slash |
| `TestMatchLiteral_EmptyPath` | Empty path handling |
| `TestMatchLiteral_CaseSensitive` | Case-sensitive comparison |

#### Pattern Target Matching Tests

| Test | Description |
|------|-------------|
| `TestMatchPattern_SingleCapture` | Single `{name}` capture |
| `TestMatchPattern_MultipleCaptures` | Multiple captures in pattern |
| `TestMatchPattern_CaptureAtStart` | Capture at pattern start |
| `TestMatchPattern_CaptureAtEnd` | Capture at pattern end |
| `TestMatchPattern_OnlyCapture` | Pattern is just a capture |
| `TestMatchPattern_NoMatchWrongSuffix` | Wrong suffix doesn't match |
| `TestMatchPattern_NoMatchWrongPrefix` | Wrong prefix doesn't match |
| `TestMatchPattern_EmptyCapture` | Empty capture value allowed |
| `TestMatchPattern_CaptureWithSlash` | Captures can contain slashes |
| `TestMatchPattern_DuplicateCaptureName` | Same capture name must match |
| `TestMatchPattern_PhonyCaptureNotAllowed` | Phony patterns work |
| `TestMatchPattern_AdjacentCaptures` | Adjacent captures (greedy) |

#### Target Lookup Tests

| Test | Description |
|------|-------------|
| `TestLookupTarget_ExactMatch` | Find exact literal match |
| `TestLookupTarget_PatternMatch` | Find pattern match |
| `TestLookupTarget_ExactMatchPreferred` | Exact beats pattern |
| `TestLookupTarget_NotFound` | Error for unmatched path |
| `TestLookupTarget_PhonyTarget` | Lookup phony target |
| `TestLookupTarget_DirectoryTarget` | Lookup directory target |
| `TestLookupTarget_MultiplePatternMatches` | First pattern wins |
| `TestLookupTarget_EmptyTargetList` | Error for empty list |

### CLI Integration

The planner adds a new debug flag:

| Flag | Description |
|------|-------------|
| `--debug-plan` | Dump build planning / target matching (for development) |

The debug output shows:
- All defined targets with their types (literal, pattern, phony, directory)
- Pattern matching test results for sample paths
- Target lookup examples based on defined patterns
- Dependency resolution for each target

### Dependency Resolution (`resolve.go`)

Implements dependency path resolution for converting pattern-based dependencies to concrete file paths.

#### ResolveDependency Function

```go
func ResolveDependency(dep ast.Dependency, captures map[string]string, ctx *eval.Context) (string, error)
```

Resolves a single dependency pattern to a concrete path.

**Resolution order for `{name}` in dependency patterns:**
1. If name is in captures (from pattern matching), use capture value
2. If name is defined in context (user variable or built-in), use variable value
3. Otherwise, return error for undefined variable

#### ResolveDependencies Function

```go
func ResolveDependencies(deps []ast.Dependency, captures map[string]string, ctx *eval.Context) ([]string, error)
```

Resolves multiple dependencies to concrete paths. Processes each dependency in order and returns a slice of resolved paths.

#### Error Types

| Type | Description |
|------|-------------|
| `UndefinedVariableError` | Variable in dependency pattern cannot be resolved |

#### Design Decisions

1. **Capture precedence**: When a name exists both as a capture and a variable, the capture takes precedence. This ensures pattern matching works correctly.

2. **Built-in support**: Built-in variables like `os` and `arch` are resolved from the evaluation context.

3. **Error early**: Resolution fails immediately on undefined variables rather than continuing with partial results.

### Build Planning (`plan.go`)

Implements the core build planning logic including recursive dependency planning and topological sorting.

#### BuildReason Enumeration

```go
type BuildReason int

const (
    BuildReasonTargetMissing    BuildReason = iota  // Target file doesn't exist
    BuildReasonDependencyNewer                       // A dependency is newer than target
    BuildReasonPhonyTarget                           // Phony targets always rebuild
    BuildReasonForcedRebuild                         // Explicit rebuild requested
)
```

#### BuildTask Structure

```go
type BuildTask struct {
    Target        string            // Path to be built (@ prefix for phony)
    Dependencies  []string          // Resolved dependency paths
    OrderOnlyDeps []string          // Order-only dependencies (from .after:)
    Recipe        *ast.Recipe       // Recipe to execute (may be nil)
    Reason        BuildReason       // Why this target needs rebuilding
    Captures      map[string]string // Pattern capture values
    TargetDef     *ast.Target       // AST target definition
}
```

**OrderOnlyDeps**: These are dependencies specified via `.after:` directives. They:
- Must exist or be buildable before this target runs
- Do NOT affect staleness checking (their timestamps are ignored)
- Are used purely for build ordering (e.g., creating directories before files)

#### BuildPlan Structure

```go
type BuildPlan struct {
    Tasks []BuildTask  // Topologically sorted list of tasks
}
```

#### FileSystem Interface

```go
type FileSystem interface {
    Exists(path string) bool
    ModTime(path string) (time.Time, error)
}
```

Abstracts file system operations for testability and platform independence.

#### PlanBuild Function

```go
func PlanBuild(requestedTarget string, targets []*ast.Target, ctx *eval.Context, fs FileSystem) (*BuildPlan, error)
```

Creates a build plan for the requested target.

#### PlanBuildWithVerbose Function

```go
func PlanBuildWithVerbose(requestedTarget string, targets []*ast.Target, ctx *eval.Context, fs FileSystem, verboseOutput io.Writer) (*BuildPlan, error)
```

Creates a build plan with optional verbose output. When `verboseOutput` is non-nil, staleness check decisions are written to it:
- For targets needing rebuild: `target: rebuild needed (reason)`
- For up-to-date targets: `target: up to date`

**Planning steps:**
1. **Target lookup**: Find matching target definition (exact or pattern)
2. **Dependency resolution**: Convert patterns to concrete paths
3. **Recursive planning**: Plan all dependencies first (DFS)
4. **Staleness detection**: Determine if rebuild is needed
5. **Topological ordering**: Tasks ordered so dependencies come first

#### Error Types

| Type | Description |
|------|-------------|
| `CircularDependencyError` | Circular dependency detected in target graph |
| `MissingSourceError` | Source file doesn't exist and has no build rule |

#### Staleness Detection Rules

| Condition | Rebuild? | Reason |
|-----------|----------|--------|
| Target is phony | Yes | `BuildReasonPhonyTarget` |
| Target file doesn't exist | Yes | `BuildReasonTargetMissing` |
| Any dependency newer than target | Yes | `BuildReasonDependencyNewer` |
| All dependencies older than target | No | Up to date |
| Dependency was rebuilt this plan | Yes | `BuildReasonDependencyNewer` |

#### Design Decisions

1. **DFS-based planning**: Uses depth-first search to plan dependencies before their dependents, naturally producing topological order.

2. **Cycle detection with stack**: Maintains a recursion stack to detect cycles during the DFS traversal. Returns `CircularDependencyError` with cycle path.

3. **Source file detection**: If a dependency has no build rule and doesn't exist on the filesystem, returns `MissingSourceError`.

4. **Phony targets always rebuild**: Phony targets (prefixed with `@`) are always included in the plan regardless of filesystem state.

5. **Transitive rebuilds**: If a dependency is rebuilt, its dependents are also marked for rebuild even if their files are newer.

### Planner Unit Tests

#### Dependency Resolution Tests (`resolve_test.go`)

| Test | Description |
|------|-------------|
| `TestResolveDependency_LiteralOnly` | Literal dependencies resolve directly |
| `TestResolveDependency_VariableInterpolation` | Variables are substituted |
| `TestResolveDependency_CaptureSubstitution` | Captures are substituted |
| `TestResolveDependency_MixedInterpolationAndCapture` | Mixed captures and variables |
| `TestResolveDependency_BuiltinVariable` | Built-in variables work |
| `TestResolveDependency_UndefinedVariable` | Undefined variables cause error |
| `TestResolveDependency_CapturePreferredOverVariable` | Capture takes precedence |
| `TestResolveDependency_MultipleCaptures` | Multiple captures work |
| `TestResolveDependency_EmptyCapture` | Empty capture values work |
| `TestResolveDependencies_MultipleDeps` | Multiple dependencies resolved |
| `TestResolveDependencies_WithCaptures` | Multiple with captures |
| `TestResolveDependencies_Empty` | Empty dependency list |
| `TestResolveDependencies_ErrorPropagation` | Errors propagate correctly |

#### Build Planning Tests (`plan_test.go`)

| Test | Description |
|------|-------------|
| `TestBuildReason_String` | Reason string representations |
| `TestBuildTask_Structure` | Task structure verification |
| `TestBuildPlan_Empty` | Empty plan handling |
| `TestBuildPlan_AddTask` | Adding tasks to plan |
| `TestPlanBuild_SingleTarget_NoDeps` | Single target without dependencies |
| `TestPlanBuild_SingleTarget_WithDeps` | Single target with source dependencies |
| `TestPlanBuild_ChainedDependencies` | A→B→C dependency chain |
| `TestPlanBuild_DiamondDependency` | A→B,C, B→D, C→D diamond |
| `TestPlanBuild_PatternTarget` | Pattern target with captures |
| `TestPlanBuild_PhonyTarget` | Phony target always rebuilds |
| `TestPlanBuild_TargetUpToDate` | Up-to-date target skipped |
| `TestPlanBuild_DependencyNewer` | Newer dependency triggers rebuild |
| `TestPlanBuild_TargetNotFound` | Error for unknown target |
| `TestPlanBuild_CircularDependency` | Circular dependency detected |
| `TestPlanBuild_SourceFileMissing` | Missing source file error |
| `TestPlanBuild_NoRecipe` | Target without recipe handled |
| `TestPlanBuild_OrderOnlyDeps_BuildOrder` | Order-only deps affect build order |
| `TestPlanBuild_OrderOnlyDeps_NotForStaleness` | Order-only deps don't trigger rebuild |
| `TestPlanBuild_OrderOnlyDeps_MustExist` | Order-only deps must exist or have build rule |
| `TestPlanBuild_OrderOnlyDeps_InTask` | Order-only deps tracked in BuildTask |
| `TestCircularDependencyError_Error` | Error message format |
| `TestPlanBuild_VerboseOutput` | Verbose mode outputs staleness decisions |
| `TestPlanBuild_VerboseOutput_UpToDate` | Verbose mode shows up-to-date targets |

### Command Interpolation (`command.go`)

Implements command interpolation for recipe execution, resolving automatic variables, captures, and user variables with optional shell quoting.

#### CommandContext Structure

```go
type CommandContext struct {
    parent    *Context
    automatic map[string]string
    captures  map[string]string
}
```

Extends the evaluation context with:
- **Automatic variables**: `target`, `out`, `deps`, `in`, `stem`, `target.dir`, `target.file`
- **Captures**: Pattern match values from target pattern matching
- **Parent context**: Inherits user-defined and built-in variables

#### Automatic Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `target` | Target path being built | `build/app` |
| `out` | Alias for target | `build/app` |
| `deps` | Space-separated dependency list | `main.c utils.c` |
| `in` | First dependency | `main.c` |
| `stem` | Pattern match stem (set via SetStem) | `utils` |
| `target.dir` | Directory part of target | `build` |
| `target.file` | Filename part of target | `app` |

#### Key Functions

| Function | Description |
|----------|-------------|
| `NewCommandContext(parent, target, deps)` | Creates context with automatic vars set |
| `SetStem(stem)` | Sets the stem automatic variable |
| `SetCaptures(captures)` | Sets capture variables from pattern matching |
| `Get(name)` | Gets variable (automatic → captures → parent) |
| `InterpolateCommand(cmd, ctx)` | Interpolates a line command |
| `InterpolateBlockCommand(block, ctx)` | Interpolates a block command |
| `ShellQuote(s)` | Quotes string for shell (single quotes) |

#### Variable Resolution Priority

1. Automatic variables (`target`, `deps`, etc.)
2. Captures (from pattern matching)
3. Parent context (user variables, lazy variables, built-ins)

#### Shell Quoting

By default, interpolated values are shell-quoted using single quotes:

```
{target} → 'build/app'
{deps}   → 'main.c utils.c'
```

The `:raw` modifier disables quoting:

```
{FLAGS:raw} → -Wall -O2  (no quotes, allows word splitting)
```

**Quoting rules:**
- Simple strings: `'string'`
- Strings with single quotes: `'it'"'"'s'` (end quote, double-quoted quote, start quote)
- Empty strings: `''`

#### Design Decisions

1. **Automatic variable precedence**: Automatic variables take precedence over captures and user variables. This ensures recipe authors can always rely on `{target}` and `{deps}`.

2. **Single quote quoting**: Single quotes are used because they prevent all shell expansion except for the quote itself. This is safer than double quotes.

3. **Raw modifier for flags**: The `:raw` modifier allows flags and options to be expanded without quoting, enabling word splitting for arguments like `-Wall -O2`.

4. **Context inheritance**: CommandContext wraps the parent Context rather than copying, ensuring lazy variable evaluation works correctly.

5. **Directory target handling**: For directory targets (ending with `/`), `target.dir` is the directory path and `target.file` is empty.

### Command Interpolation Unit Tests (`command_test.go`)

| Test | Description |
|------|-------------|
| `TestNewCommandContext` | Context creation with automatic vars |
| `TestCommandContext_NoDependencies` | Empty deps/in for no-dependency targets |
| `TestCommandContext_WithStem` | Stem variable setting |
| `TestCommandContext_WithCaptures` | Capture variable setting |
| `TestCommandContext_DirectoryTarget` | Directory target handling |
| `TestCommandContext_RootTarget` | Root-level target handling |
| `TestCommandContext_InheritsVariables` | Parent variable inheritance |
| `TestInterpolateCommand_LiteralOnly` | Literal-only commands |
| `TestInterpolateCommand_AutomaticVariables` | Automatic variable substitution |
| `TestInterpolateCommand_RawModifier` | Raw modifier (no quoting) |
| `TestInterpolateCommand_CaptureVariables` | Capture variable substitution |
| `TestInterpolateCommand_StemVariable` | Stem variable substitution |
| `TestInterpolateCommand_UserVariables` | User variable substitution |
| `TestInterpolateCommand_UndefinedVariable` | Error for undefined variables |
| `TestInterpolateCommand_BuiltinVariables` | Built-in variable access |
| `TestInterpolateCommand_TargetDirAndFile` | target.dir and target.file |
| `TestShellQuote_Simple` | Simple string quoting |
| `TestShellQuote_SpecialCharacters` | Special characters (spaces, $, *, etc.) |
| `TestInterpolateBlockCommand` | Block command interpolation |

## Executor Package (`internal/executor`)

The executor package provides recipe execution for Needfiles, handling shell invocation and command orchestration.

### Shell Configuration (`executor.go`)

```go
type ShellConfig struct {
    Shell   string // Path to shell (default: /bin/sh)
    DryRun  bool   // If true, print commands without executing
    Verbose bool   // If true, print commands before executing
}
```

#### Key Functions

| Function | Description |
|----------|-------------|
| `NewShellConfig()` | Creates config with `/bin/sh` default |
| `SetShell(shell)` | Sets the shell path |
| `WithOverride(shell)` | Returns new config with overridden shell |
| `Validate()` | Validates that the shell exists and is executable |

### Shell Validation

The `Validate()` method checks that the configured shell exists before execution:

```go
func (c *ShellConfig) Validate() error
```

**Behavior:**
- Absolute paths (starting with `/`): Checks if the file exists and is executable via `exec.LookPath`
- Relative names (e.g., `bash`, `sh`): Searches PATH for the executable
- Returns `ShellNotFoundError` if the shell cannot be found

**Usage:**

```go
cfg := NewShellConfig()
cfg.SetShell("nonexistent-shell")

if err := cfg.Validate(); err != nil {
    // Handle missing shell error
}
```

The `NewExecutorWithValidation()` function combines creation and validation:

```go
func NewExecutorWithValidation(config *ShellConfig) (*Executor, error)
```

### Executor Structure

```go
type Executor struct {
    config *ShellConfig
    output io.Writer
}
```

#### Key Methods

| Method | Description |
|--------|-------------|
| `NewExecutor(config)` | Creates executor with given config |
| `SetOutput(w)` | Sets output writer for dry-run/verbose |
| `ExecuteLine(cmdLine)` | Executes single command line |
| `ExecuteBlock(script)` | Executes multi-line script block |
| `ExecuteRecipe(recipe, ctx)` | Executes all commands in a recipe |

### Execution Modes

| Mode | Description |
|------|-------------|
| **Line Mode** | Each command line is a separate shell invocation via `shell -c "command"` |
| **Block Mode** | All lines passed as single script to `shell -c "script"` |
| **Dry Run** | Print "Would build: target" followed by commands without executing, always return success |
| **Verbose** | Print commands before executing, then execute |

### ExecResult Structure

```go
type ExecResult struct {
    Command  string // The command that was executed
    Stdout   string // Standard output
    Stderr   string // Standard error
    ExitCode int    // Exit code (0 = success)
}
```

### Recipe Execution Flow

1. Determine shell (recipe `.shell:` overrides global)
2. For each command in recipe:
   - **LineCommand**: Interpolate with CommandContext, execute as single line
   - **BlockCommand**: Interpolate all lines, execute as single script
3. Stop on first command failure (return all results so far)

### Error Types

```go
type CommandError struct {
    Command  string
    ExitCode int
    Stderr   string
}
```

Error format: `command failed with exit code N: command\nstderr...`

```go
type ShellNotFoundError struct {
    Shell string
}
```

Error format: `shell not found: path`

### Design Decisions

1. **Separate invocations for line mode**: Each line command runs in its own shell, so variable assignments don't persist between lines. This matches Make behavior.

2. **Single invocation for block mode**: Block commands preserve shell state (variables, if/fi, loops) by running as a single script.

3. **Recipe shell override**: The recipe's `.shell:` directive creates a new Executor with the overridden shell, leaving the global config unchanged.

4. **Stdout/stderr capture**: Both streams are captured separately, allowing error messages to be extracted from stderr.

5. **Exit code extraction**: Uses `syscall.WaitStatus` to get the exact exit code for error reporting.

### Executor Unit Tests (`executor_test.go`)

| Test | Description |
|------|-------------|
| `TestNewShellConfig_Default` | Default shell is /bin/sh |
| `TestShellConfig_FromGlobalDirective` | SetShell changes shell |
| `TestShellConfig_FromRecipeOverride` | WithOverride creates new config |
| `TestExecuteLine_Simple` | Simple command execution |
| `TestExecuteLine_WithVariables` | Shell variable expansion |
| `TestExecuteLine_Failure` | Command failure handling |
| `TestExecuteLine_Stderr` | Stderr capture |
| `TestExecuteLine_BashSpecific` | Bash-specific syntax |
| `TestExecuteBlock_Simple` | Multi-line script |
| `TestExecuteBlock_WithIfStatement` | If/then/else preserved |
| `TestExecuteBlock_WithLoop` | For loop preserved |
| `TestExecuteBlock_FailsOnError` | Stops on error |
| `TestExecuteRecipe_SingleLineCommand` | Single command execution |
| `TestExecuteRecipe_MultipleLineCommands` | Multiple commands |
| `TestExecuteRecipe_BlockCommand` | Block command execution |
| `TestExecuteRecipe_StopsOnFirstError` | Stops on failure |
| `TestExecuteRecipe_ShellOverride` | Recipe .shell override |
| `TestDryRun_PrintsCommands` | Dry-run prints only |
| `TestDryRun_DoesNotExecute` | Dry-run doesn't execute |
| `TestDryRun_WouldBuildPrefix` | Dry-run prints "Would build: target" prefix |
| `TestVerbose_PrintsCommand` | Verbose shows command |

### Parallel Scheduler (`scheduler.go`)

Implements parallel execution of build tasks with dependency-aware scheduling.

#### Scheduler Structure

```go
type Scheduler struct {
    executor   *Executor
    numWorkers int
    keepGoing  bool // If true, continue building after failures
}
```

#### TaskResult Structure

```go
type TaskResult struct {
    Target  string        // The target that was built
    Results []*ExecResult // Results from recipe execution
    Error   error         // Error if execution failed
    Skipped bool          // True if task was skipped due to dependency failure
}
```

#### Key Functions

| Function | Description |
|----------|-------------|
| `NewScheduler(executor, numWorkers)` | Creates scheduler with N workers |
| `SetKeepGoing(keepGoing)` | Enable/disable keep-going mode |
| `Workers()` | Returns number of workers |
| `Execute(tasks, ctxFactory)` | Executes tasks respecting dependencies |
| `ExecuteWithCallback(tasks, ctxFactory, callback)` | Execute with per-task callback |
| `ResolveWorkerCount(cliJobs, parallelDirective)` | Resolves worker count from CLI and directive |

#### Worker Count Resolution

The `ResolveWorkerCount` function determines the number of workers to use based on CLI `-j` flag and `.parallel:` directive value:

```go
func ResolveWorkerCount(cliJobs, parallelDirective int) int
```

**Resolution rules:**
1. CLI `-j` flag takes precedence when explicitly set (value > 1)
2. `.parallel:` directive is used as fallback when CLI is default (1)
3. Negative values are treated as invalid and ignored
4. Minimum return value is 1

| cliJobs | parallelDirective | Result | Reason |
|---------|-------------------|--------|--------|
| 4 | 2 | 4 | CLI overrides directive |
| 1 | 8 | 8 | Directive used when CLI default |
| 4 | 0 | 4 | CLI used, directive unset |
| 1 | 0 | 1 | Both default/unset |
| -1 | 4 | 4 | Invalid CLI ignored |
| 1 | -2 | 1 | Invalid directive ignored |

#### Keep-Going Mode

When `SetKeepGoing(true)` is called, the scheduler continues building after failures:
- Failed tasks are still marked in the `failed` map
- Dependent tasks are still skipped (their dependencies failed)
- Independent tasks continue to execute
- All results (success, failure, skipped) are returned

#### Scheduling Algorithm

1. **Initialization**: Build dependency tracking map, queue tasks with no dependencies
2. **Worker Pool**: Start N goroutines that pull from ready queue
3. **Dependency Resolution**: When task completes, decrement pending count for dependents
4. **Ready Queue**: Tasks with zero pending deps are queued for execution
5. **Cancellation**: On failure, mark remaining tasks as skipped

#### Dependency Handling

| Scenario | Behavior |
|----------|----------|
| No dependencies | Execute immediately |
| Dependencies complete | Queue when all deps finish |
| Dependency failed | Skip task, mark as failed |
| Diamond dependency | Both paths must complete |

#### Cancellation on Failure

When a task fails:
1. Set `cancelled` flag to prevent new task scheduling
2. Already-running tasks complete normally
3. Pending tasks are marked as `Skipped`
4. All results (success, failure, skipped) are returned

#### Design Decisions

1. **Channel-based coordination**: Ready queue is a buffered channel for efficient task distribution.

2. **Mutex-protected state**: Completed/failed maps are protected by mutex for thread-safety.

3. **Worker count limit**: Max concurrent tasks bounded by numWorkers regardless of ready tasks.

4. **Dependency tracking**: PendingDeps map decremented atomically when deps complete.

5. **Factory pattern for contexts**: ContextFactory callback creates fresh context per target.

### Scheduler Unit Tests (`scheduler_test.go`)

| Test | Description |
|------|-------------|
| `TestNewScheduler` | Scheduler creation with worker count |
| `TestScheduler_SingleTask_NoDeps` | Single task execution |
| `TestScheduler_MultipleTasks_NoDeps` | Multiple independent tasks |
| `TestScheduler_DependencyOrdering` | B depends on A ordering |
| `TestScheduler_ParallelIndependentTasks` | Independent tasks run in parallel |
| `TestScheduler_DiamondDependency` | Diamond A→B,C→D ordering |
| `TestScheduler_FailureCancellation` | Failure cancels pending tasks |
| `TestScheduler_NoRecipe` | Tasks without recipe succeed |
| `TestScheduler_ParallelWorkerCount` | Worker count limits concurrency |
| `TestScheduler_KeepGoing_ContinuesAfterFailure` | Keep-going continues independent tasks |
| `TestScheduler_KeepGoing_SkipsDependentTasks` | Keep-going still skips dependent tasks |
| `TestResolveWorkerCount` | Worker count resolution from CLI and directive |
| `TestScheduler_JobsFlagOverride` | CLI -j flag overrides .parallel directive |

### Autodeps Support (`autodeps.go`)

Implements parsing of Makefile-style dependency files (`.d` files) generated by compilers like gcc -MD.

#### Key Functions

| Function | Description |
|----------|-------------|
| `ParseAutodeps(content)` | Parses .d file content, returns dependency list |
| `ParseAutodepsFile(path)` | Reads and parses .d file from filesystem |

#### .d File Format

```
target: dep1 dep2 dep3
target: dep1 \
  dep2 \
  dep3
```

**Supported features:**
- Basic `target: deps` format
- Backslash line continuations
- Multiple targets per file (all deps merged)
- Escaped spaces in paths (`path\ with\ spaces`)
- Comments (lines starting with `#`)

#### Integration with BuildTask

```go
type BuildTask struct {
    // ...
    AutodepsDeps []string // Learned dependencies from .d file
    AutodepsPath string   // Path to .autodeps file for post-build update
}
```

The planner:
1. Reads existing .d file specified by `.autodeps:` directive
2. Includes learned dependencies in staleness checking
3. Stores AutodepsPath for post-build reference

#### Design Decisions

1. **Non-existent files return empty**: If the .d file doesn't exist (first build), return empty deps without error.

2. **Merge multiple targets**: Some compilers output multiple targets in one file; all dependencies are merged.

3. **Escaped spaces**: Handle `\ ` sequences for paths containing spaces.

4. **Staleness only (not planning)**: Autodeps affect staleness checks but don't create build tasks for headers.

### Autodeps Unit Tests (`autodeps_test.go`)

| Test | Description |
|------|-------------|
| `TestParseAutodeps_Simple` | Basic target: deps format |
| `TestParseAutodeps_Multiline` | Backslash continuation |
| `TestParseAutodeps_MultipleTargets` | Multiple targets merged |
| `TestParseAutodeps_Empty` | Empty content handling |
| `TestParseAutodeps_NoColonLine` | Lines without colon ignored |
| `TestParseAutodeps_SpacesInPath` | Escaped spaces in paths |
| `TestParseAutodepsFile` | File reading and parsing |
| `TestParseAutodepsFile_NotExists` | Non-existent file returns empty |

## Environ Package (`internal/environ`)

The environ package handles environment management for need, including requirements checking and version detection.

### Package Structure

| File | Contents |
|------|----------|
| `doc.go` | Package documentation |
| `requirements.go` | `RequirementsChecker` for binary existence and version checks |
| `version.go` | Version parsing and comparison logic |
| `errors.go` | Error types for environment operations |
| `requirements_test.go` | Tests for requirements checking |
| `version_test.go` | Tests for version parsing and matching |

### Requirements Checker (`requirements.go`)

The `RequirementsChecker` validates that required binaries are available in PATH.

#### RequirementsChecker Structure

```go
type RequirementsChecker struct {
    lookPath     func(file string) (string, error)
    versionCache map[string]versionCacheEntry // Cached version results
}
```

#### Key Methods

| Method | Description |
|--------|-------------|
| `NewRequirementsChecker()` | Creates a new checker (uses `exec.LookPath`) |
| `CheckBinaryExists(name)` | Checks if binary exists in PATH |
| `CheckRequirement(req)` | Checks a single requirement (existence only) |
| `CheckRequirements(reqs)` | Checks multiple requirements |
| `CheckRequirementWithVersion(req)` | Checks requirement including version validation |
| `CheckRequirementsWithVersion(reqs)` | Checks multiple requirements with versions |
| `DetectVersion(name)` | Attempts to detect binary version (cached) |
| `ClearVersionCache()` | Clears the version detection cache |

#### RequirementResult Structure

```go
type RequirementResult struct {
    Requirement     ast.Requirement // The requirement checked
    Found           bool            // True if binary found in PATH
    Path            string          // Full path to binary
    DetectedVersion string          // Version string if detected
    Error           error           // Error if check failed
}
```

### Version Handling (`version.go`)

#### Version Structure

```go
type Version struct {
    Major int // Major version number
    Minor int // Minor version number (-1 if not specified)
    Patch int // Patch version number (-1 if not specified)
}
```

#### ParseVersion Function

```go
func ParseVersion(s string) (*Version, error)
```

Extracts version information from strings using pattern matching. Handles:
- Simple versions: `11.4.0`, `11.4`, `11`
- With prefix: `v1.2.3`
- From tool output: `gcc (Ubuntu 11.4.0-1ubuntu1~22.04) 11.4.0`, `Python 3.10.12`

#### Version.Satisfies Method

```go
func (v Version) Satisfies(spec ast.VersionSpec) bool
```

Checks if a detected version satisfies a requirement:

| VersionSpec Type | Match Logic |
|-----------------|-------------|
| `VersionLatest` | Always matches |
| `VersionMajor{11}` | Major must equal 11 |
| `VersionMajorMinor{11, 4}` | Major=11, Minor=4 (any patch) |
| `VersionExact{11, 4, 0}` | Exact match required |

#### DetectVersion Method

```go
func (c *RequirementsChecker) DetectVersion(name string) (*Version, error)
```

Attempts to detect version by:
1. Checking the version cache (returns cached result if available)
2. Checking if binary exists
3. Trying common version flags: `--version`, `-version`, `-v`
4. Parsing version from stdout/stderr output
5. Caching the result (both successes and errors)

**Caching behavior:**
- Results are cached per binary name
- Both successful detections and errors are cached
- Cache persists for the lifetime of the `RequirementsChecker` instance
- Use `ClearVersionCache()` to reset the cache

### Error Types (`errors.go`)

| Error Type | Description |
|------------|-------------|
| `BinaryNotFoundError` | Required binary not found in PATH |
| `VersionMismatchError` | Found version doesn't match requirement |
| `VersionDetectionError` | Unable to detect version |
| `EnvironmentNotFoundError` | Named environment not found |
| `NoDefaultEnvironmentError` | No default environment when required |

### CLI Integration (`cmd/need/environ.go`, `environ_adapter.go`)

#### --check-env Flag

Verifies environment requirements:

```bash
build --check-env              # Check default environment
build --check-env -e ci        # Check named environment
```

**Behavior:**
1. Parse needfile and extract environments
2. Select environment (default or by `--env` name)
3. For bare environments: check all `.requires:` entries
4. Report status with ✓/✗ symbols

#### --list-env Flag

Lists available environments:

```bash
build --list-env
```

**Output format:**
```
Available environments (2):
  (default)             bare             (2 requirements)
  ci                    docker         
```

### Design Decisions

1. **Separate package for environment logic**: Keeps environment management isolated from parsing and evaluation.

2. **Injectable lookPath function**: `RequirementsChecker` accepts a `lookPath` function, enabling testing without actual filesystem operations.

3. **Version detection via common flags**: Tries `--version`, `-version`, and `-v` to cover most command-line tools.

4. **Regex-based version parsing**: Uses `(?:^|[^0-9])v?(\d+)(?:\.(\d+)(?:\.(\d+))?)?` to handle various version output formats.

5. **Non-fatal version detection**: If version detection fails, the requirement check can still succeed for `VersionLatest` requirements.

6. **Bare environment handling**: Default behavior when no `.environment:` is defined - uses host system directly.

7. **Named environments require explicit selection**: If only named environments exist, `--env` must be specified.

### Unit Tests

#### Requirements Tests (`requirements_test.go`)

| Test | Description |
|------|-------------|
| `TestCheckBinaryExists` | Existing and non-existent binaries |
| `TestCheckRequirement_ExistsNoVersion` | Binary exists with VersionLatest |
| `TestCheckRequirement_NotFound` | Binary not found |
| `TestCheckRequirements` | Multiple requirements |
| `TestCheckRequirements_WithFailures` | Mixed success/failure |
| `TestRequirementResult_String` | Human-readable status |
| `TestBinaryNotFoundError` | Error message format |
| `TestVersionMismatchError` | Error message format |
| `TestVersionDetectionError` | Error message format |

#### Version Tests (`version_test.go`)

| Test | Description |
|------|-------------|
| `TestParseVersion` | Various version string formats |
| `TestVersionString` | Version to string conversion |
| `TestVersionSatisfies` | Version matching against specs |
| `TestDetectVersion` | Live version detection |
| `TestVersionCache` | Version cache population and retrieval |
| `TestVersionCacheError` | Error caching for non-existent binaries |
| `TestClearVersionCache` | Cache clearing |

### Devcontainer Detection (`devcontainer.go`)

Implements detection and parsing of devcontainer configurations for VS Code Development Containers.

#### DevcontainerDetector Structure

```go
type DevcontainerDetector struct{}
```

#### Key Functions

| Function | Description |
|----------|-------------|
| `NewDevcontainerDetector()` | Creates a new detector |
| `DetectConfig(baseDir)` | Searches for devcontainer configuration |
| `LoadConfig(path)` | Loads and parses devcontainer.json |
| `ParseDevcontainerConfig(data)` | Parses devcontainer.json content |

#### Configuration Detection

The detector searches for devcontainer configuration in this order:
1. `.devcontainer/devcontainer.json` (preferred)
2. `devcontainer.json` in project root

#### DevcontainerConfig Structure

```go
type DevcontainerConfig struct {
    Name              string                   // Container name
    Image             string                   // Docker image to use
    Dockerfile        string                   // Path to Dockerfile
    DockerComposeFile string                   // Path to docker-compose file
    Service           string                   // Service name (for docker-compose)
    Build             *DevcontainerBuildConfig // Build configuration
    WorkspaceFolder   string                   // Workspace folder in container
    RemoteUser        string                   // User to run as
}

type DevcontainerBuildConfig struct {
    Dockerfile string // Path to Dockerfile
    Context    string // Build context directory
}
```

#### DevcontainerRunner Structure

```go
type DevcontainerRunner struct {
    projectDir string
    configPath string
    lookPath   func(name string) (string, error)
}
```

Handles running commands in a devcontainer using the `devcontainer` CLI.

#### DevcontainerRunner Methods

| Method | Description |
|--------|-------------|
| `NewDevcontainerRunner(projectDir)` | Creates a new runner |
| `SetConfigPath(path)` | Sets path to devcontainer.json |
| `CheckCLI()` | Verifies devcontainer CLI is installed |
| `Up()` | Starts the devcontainer |
| `Exec(command)` | Executes a command in the container |
| `OpenShell()` | Opens an interactive shell in the container |

#### Devcontainer Tests (`devcontainer_test.go`)

| Test | Description |
|------|-------------|
| `TestDevcontainerDetector_DetectConfig_Directory` | Detect .devcontainer/devcontainer.json |
| `TestDevcontainerDetector_DetectConfig_RootJson` | Detect root devcontainer.json |
| `TestDevcontainerDetector_DetectConfig_NotFound` | Handle missing configuration |
| `TestDevcontainerDetector_DetectConfig_DirectoryPriority` | .devcontainer/ takes priority |
| `TestParseDevcontainerConfig` | Parse various config formats |
| `TestDevcontainerDetector_LoadConfig` | Load and parse file |
| `TestNewDevcontainerRunner` | Runner creation |
| `TestDevcontainerRunner_CheckCLI_NotInstalled` | CLI not installed error |
| `TestDevcontainerConfig_GetImageOrBuildSource` | Config source description |

#### CLI Tests (`main_test.go`)

| Test | Description |
|------|-------------|
| `TestRunListEnv` | List environments with mixed types |
| `TestRunListEnvNoEnvs` | No environments defined |
| `TestRunCheckEnvDefaultSuccess` | Requirements satisfied |
| `TestRunCheckEnvDefaultMissing` | Missing binary |
| `TestRunCheckEnvNamed` | Check named environment |
| `TestRunCheckEnvNamedNotFound` | Named environment doesn't exist |
| `TestRunCheckEnvNoDefaultWithNamedOnly` | Error when no default |
| `TestRunCheckEnvNoEnvironments` | Bare environment (no `.environment:`) |
| `TestShowInstall_MissingBinary` | Show install suggestion for missing binary |
| `TestShowInstall_AllPresent` | All binaries present with --show-install |
| `TestRunCheckEnvDevcontainer_WithConfig` | Devcontainer with valid configuration |
| `TestRunCheckEnvDevcontainer_WithSourcePath` | Devcontainer with custom source path |
| `TestRunCheckEnvDevcontainer_NoConfig` | Devcontainer with missing configuration |
| `TestRunCheckEnvDevcontainer_InvalidConfig` | Devcontainer with invalid JSON config |
| `TestRunListEnv_WithDevcontainer` | List environments including devcontainer type |

#### Devcontainer Environment Checking (`environ.go`)

The `checkDevcontainerEnvironment` function validates devcontainer environments:

```go
func checkDevcontainerEnvironment(env *ast.Environment, needfileDir string, verbose, showInstall bool) int
```

**Validation Steps:**

1. **CLI Check**: Verifies `devcontainer` CLI is installed via `DevcontainerRunner.CheckCLI()`
2. **Configuration Detection**:
   - If `.source:` specified, uses that path
   - Otherwise, auto-detects `.devcontainer/devcontainer.json` or `devcontainer.json`
3. **Configuration Parsing**: Loads and parses devcontainer.json, reports errors for invalid JSON
4. **Configuration Display**: Shows container name and source (image/dockerfile/compose)

**Exit Codes:**

| Condition | Exit Code |
|-----------|-----------|
| Configuration valid | `exitSuccess` (0) |
| CLI not found | Still success (informational only) |
| Configuration not found | `exitEnvError` (4) |
| Invalid configuration JSON | `exitEnvError` (4) |

### Nix Environment Support (`nix.go`)

Implements detection and execution for Nix environments.

#### NixType Enumeration

```go
type NixType int

const (
    NixTypeShell NixType = iota // shell.nix
    NixTypeFlake                // flake.nix
)
```

#### NixDetector Structure

```go
type NixDetector struct{}
```

Detects Nix configurations in a project directory.

#### NixDetector Methods

| Method | Description |
|--------|-------------|
| `NewNixDetector()` | Creates a new detector |
| `DetectConfig(baseDir, source)` | Searches for Nix configuration |

**Detection order:**
1. If `.source:` specified, use that path directly
2. Check for `shell.nix` (preferred)
3. Check for `flake.nix`

#### NixRunner Structure

```go
type NixRunner struct {
    projectDir string
    configPath string
    nixType    NixType
    args       []string
    lookPath   func(name string) (string, error)
}
```

Handles running commands in a Nix environment.

#### NixRunner Methods

| Method | Description |
|--------|-------------|
| `NewNixRunner(projectDir)` | Creates a new runner |
| `SetConfig(path, nixType)` | Sets the nix config path and type |
| `SetArgs(args)` | Sets extra arguments from `.args:` |
| `CheckCLI()` | Verifies nix-shell is installed |
| `Exec(command)` | Executes a command in the nix environment |
| `OpenShell()` | Opens an interactive shell |

**Execution modes:**
- **shell.nix**: Uses `nix-shell --run <command>`
- **flake.nix**: Uses `nix develop -c sh -c <command>`

#### Nix Tests (`nix_test.go`)

| Test | Description |
|------|-------------|
| `TestNixDetector_DetectConfig_ShellNix` | Detect shell.nix |
| `TestNixDetector_DetectConfig_FlakeNix` | Detect flake.nix |
| `TestNixDetector_DetectConfig_FromSource` | Use custom source path |
| `TestNixDetector_DetectConfig_NotFound` | Handle missing configuration |
| `TestNixDetector_DetectConfig_ShellNixPriority` | shell.nix takes priority |
| `TestNixRunner_CheckCLI_NotInstalled` | CLI not installed error |

#### CLI Tests for Nix (`main_test.go`)

| Test | Description |
|------|-------------|
| `TestRunCheckEnvNix_WithShellNix` | Nix environment with shell.nix |
| `TestRunCheckEnvNix_WithFlakeNix` | Nix environment with flake.nix |
| `TestRunCheckEnvNix_WithSourcePath` | Nix environment with custom source path |
| `TestRunCheckEnvNix_NoConfig` | Nix environment with missing configuration |
| `TestRunCheckEnvNix_WithArgs` | Nix environment with .args: directive |
| `TestRunListEnv_WithNix` | List environments including nix type |

#### Nix Environment Checking (`environ.go`)

The `checkNixEnvironment` function validates Nix environments:

```go
func checkNixEnvironment(env *ast.Environment, needfileDir string, verbose, showInstall bool) int
```

**Validation Steps:**

1. **CLI Check**: Verifies `nix-shell` is installed via `NixRunner.CheckCLI()`
2. **Configuration Detection**: Uses `NixDetector.DetectConfig()` to find shell.nix or flake.nix
3. **Configuration Display**: Shows path and type (shell.nix vs flake.nix)
4. **Args Display**: Shows `.args:` if specified

**Exit Codes:**

| Condition | Exit Code |
|-----------|-----------|
| Configuration valid | `exitSuccess` (0) |
| CLI not found | Still success (informational only) |
| Configuration not found | `exitEnvError` (4) |

### Lima Environment Support (`lima.go`)

Implements detection and execution for Lima VM environments (macOS).

#### LimaDetector Structure

```go
type LimaDetector struct{}
```

Detects Lima configurations in a project directory.

#### LimaDetector Methods

| Method | Description |
|--------|-------------|
| `NewLimaDetector()` | Creates a new detector |
| `DetectConfig(baseDir, source)` | Searches for Lima configuration |

**Detection order:**
1. If `.source:` specified, use that path directly
2. Check for `lima.yaml`

#### LimaRunner Structure

```go
type LimaRunner struct {
    projectDir string
    vmName     string
    configPath string
    args       []string
    lookPath   func(name string) (string, error)
}
```

Handles running commands in a Lima VM.

#### LimaRunner Methods

| Method | Description |
|--------|-------------|
| `NewLimaRunner(projectDir, vmName)` | Creates a new runner |
| `SetConfigPath(path)` | Sets the Lima config path |
| `SetArgs(args)` | Sets extra arguments from `.args:` |
| `CheckCLI()` | Verifies limactl is installed |
| `Start()` | Starts the Lima VM |
| `Stop()` | Stops the Lima VM |
| `Exec(command)` | Executes a command in the VM |
| `OpenShell()` | Opens an interactive shell |

#### Lima Tests (`lima_test.go`)

| Test | Description |
|------|-------------|
| `TestLimaDetector_DetectConfig_LimaYaml` | Detect lima.yaml |
| `TestLimaDetector_DetectConfig_FromSource` | Use custom source path |
| `TestLimaDetector_DetectConfig_NotFound` | Handle missing configuration |
| `TestLimaRunner_CheckCLI_NotInstalled` | CLI not installed error |
| `TestLimaRunner_VMName` | VM name assignment |

#### CLI Tests for Lima (`main_test.go`)

| Test | Description |
|------|-------------|
| `TestRunCheckEnvLima_WithLimaYaml` | Lima environment with lima.yaml |
| `TestRunCheckEnvLima_WithSourcePath` | Lima environment with custom source path |
| `TestRunCheckEnvLima_NoConfig` | Lima environment with missing configuration |
| `TestRunListEnv_WithLima` | List environments including lima type |

#### Lima Environment Checking (`environ.go`)

The `checkLimaEnvironment` function validates Lima environments:

```go
func checkLimaEnvironment(env *ast.Environment, needfileDir string, verbose, showInstall bool) int
```

**Validation Steps:**

1. **CLI Check**: Verifies `limactl` is installed via `LimaRunner.CheckCLI()`
2. **Configuration Detection**: Uses `LimaDetector.DetectConfig()` to find lima.yaml
3. **Configuration Display**: Shows path and VM name
4. **Args Display**: Shows `.args:` if specified

**Exit Codes:**

| Condition | Exit Code |
|-----------|-----------|
| Configuration valid | `exitSuccess` (0) |
| CLI not found | Still success (informational only) |
| Configuration not found | `exitEnvError` (4) |

### Environment Selection (`selector.go`)

Implements environment selection logic for determining which environment to use for a build.

#### EnvironmentSelector Structure

```go
type EnvironmentSelector struct{}
```

#### Selection Priority

1. `--env` flag takes highest precedence
2. `NEED_ENV` environment variable is used as fallback
3. Unnamed (default) environment is used if no explicit selection
4. Error if only named environments exist and no selection is made

#### Key Functions

| Function | Description |
|----------|-------------|
| `NewEnvironmentSelector()` | Creates a new selector |
| `Select(envs, envFlag, buildEnv)` | Selects appropriate environment |

**Returns:**
- Selected `*ast.Environment`
- `nil` if no environments defined (bare environment)
- `EnvironmentNotFoundError` if requested environment doesn't exist
- `NoDefaultEnvironmentError` if no default and no selection

#### Environment Selection Tests (`selector_test.go`)

| Test | Description |
|------|-------------|
| `TestSelectEnvironment_EnvFlagPrecedence` | --env flag takes priority |
| `TestSelectEnvironment_BuildEnvFallback` | NEED_ENV used as fallback |
| `TestSelectEnvironment_DefaultEnv` | Default environment selected |
| `TestSelectEnvironment_ErrorWhenOnlyNamedAndNoSelection` | Error when no default |
| `TestSelectEnvironment_EmptyEnvsList` | Bare environment (nil) returned |
| `TestSelectEnvironment_EnvFlagNotFound` | Error for unknown environment |
| `TestSelectEnvironment_EnvFlagOverridesBuildEnv` | Flag overrides NEED_ENV |

### Install Suggestions (`install.go`)

Provides installation suggestions for missing binaries.

#### PackageManager Interface

```go
type PackageManager interface {
    Name() string                        // Package manager name (e.g., "apt", "brew")
    GetInstallCommand(binary string) string  // Install command for binary
}
```

#### Supported Package Managers

| Manager | OS | Install Command |
|---------|-----|-----------------|
| apt | Debian/Ubuntu | `sudo apt install <pkg>` |
| brew | macOS | `brew install <pkg>` |
| dnf | Fedora/RHEL | `sudo dnf install <pkg>` |
| pacman | Arch Linux | `sudo pacman -S <pkg>` |
| zypper | openSUSE | `sudo zypper install <pkg>` |
| apk | Alpine Linux | `apk add <pkg>` |

#### DetectPackageManager Function

```go
func DetectPackageManager() PackageManager
```

Auto-detects the system's package manager by checking for common binaries in PATH.

#### --show-install Flag

When used with `--check-env`, shows install commands for missing binaries:

```bash
build --check-env --show-install
```

**Output:**
```
Checking environment: (default)
Runtime: bare

Checking 2 requirement(s)...
  ✗ gcc: not found
      install: sudo apt install gcc
  ✓ ls: found

Some requirements are not met
```

#### Install Tests (`install_test.go`)

| Test | Description |
|------|-------------|
| `TestDetectPackageManager` | Package manager detection on different OS |
| `TestPackageManager_GetInstallCommand` | Install command format for each manager |
| `TestPackageManager_Name` | Manager name correctness |
| `TestGetInstallSuggestion` | Full install suggestion generation |
| `TestBinaryToPackageMapping` | Binary to package name mapping |
| `TestInstallSuggestions_Integration` | End-to-end install suggestion |

### Container Environments (`container.go`, `container_env.go`, `image.go`, `runner.go`)

The environ package supports container-based build environments using Docker or Podman.

#### Package Structure for Container Support

| File | Contents |
|------|----------|
| `container.go` | `ContainerDetector` for Dockerfile detection and validation |
| `container_env.go` | `ContainerEnvironment` high-level orchestration |
| `image.go` | `ImageBuilder` for building and caching container images |
| `runner.go` | `ContainerRunner` for executing commands in containers |

#### ContainerDetector (`container.go`)

Detects and validates Dockerfiles for container environments.

```go
type ContainerDetector struct {
    lookPath func(name string) (string, error)
}
```

**Key Methods:**

| Method | Description |
|--------|-------------|
| `NewContainerDetector()` | Creates new detector (uses `exec.LookPath`) |
| `DetectDockerfile(env, baseDir)` | Locates Dockerfile from `.source:` directive |
| `ValidateDockerfile(path)` | Validates Dockerfile has FROM instruction |
| `FindRuntime(runtime)` | Finds docker/podman binary in PATH |

**DockerfileResult Structure:**

```go
type DockerfileResult struct {
    Path   string // Absolute path to Dockerfile
    Exists bool   // True if file exists
}
```

**Dockerfile Validation:**
- Checks that file exists
- Validates presence of FROM instruction (required for valid Dockerfile)
- Allows ARG directives before FROM
- Skips comments and empty lines

#### ImageBuilder (`image.go`)

Builds and manages container images.

```go
type ImageBuilder struct {
    runtime    ast.Runtime
    runtimeCmd string
    extraArgs  []string
    runCommand func(name string, args ...string) ([]byte, error)
}
```

**Key Methods:**

| Method | Description |
|--------|-------------|
| `NewImageBuilder(runtime, cmd, extraArgs)` | Creates builder for runtime |
| `BuildCommand(dockerfile, tag)` | Returns exec.Cmd for building |
| `ImageExists(tag)` | Checks if image exists locally |
| `Build(dockerfile, tag)` | Builds image from Dockerfile |

**Image Tag Generation:**

```go
func GenerateImageTag(project, envName string) string
```

Generates deterministic image tags:
- `myproject:latest` for default environment
- `myproject-ci:latest` for named environment "ci"

**Extra Args Parsing:**

```go
func ParseExtraArgs(argsValue *ast.Value) []string
```

Parses `.args:` directive value into command-line arguments (e.g., `--platform linux/amd64`).

#### ContainerRunner (`runner.go`)

Executes commands inside containers with workspace mounting.

```go
type ContainerRunner struct {
    runtime      ast.Runtime
    runtimeCmd   string
    workspaceDir string
    extraArgs    []string
}
```

**Key Methods:**

| Method | Description |
|--------|-------------|
| `NewContainerRunner(runtime, cmd, workspace)` | Creates runner |
| `SetExtraArgs(args)` | Sets additional container args |
| `RunCommand(imageTag, command)` | Runs command with auto-cleanup (`--rm`) |
| `ShellCommand(imageTag, shell)` | Opens interactive shell (`-it`) |
| `RunCommandKeepAlive(imageTag, command, name)` | Runs without cleanup |
| `ExecCommand(containerName, command)` | Runs command in existing container |
| `StopContainer(name)` | Stops running container |
| `RemoveContainer(name)` | Removes container |

**Container Command Generation:**

All commands include:
- `-v workspace:/workspace` - Mount project directory
- `-w /workspace` - Set working directory
- Extra args from `.args:` directive

| Mode | Flags | Behavior |
|------|-------|----------|
| Run | `--rm` | Auto-remove after execution |
| Shell | `-it --rm` | Interactive TTY, auto-remove |
| KeepAlive | `--name` | Named container, keeps running |

#### ContainerEnvironment (`container_env.go`)

High-level orchestration of container environment lifecycle.

```go
type ContainerEnvironment struct {
    env         *ast.Environment
    projectDir  string
    projectName string
    detector    *ContainerDetector
    builder     *ImageBuilder
    runner      *ContainerRunner
    imageTag    string
}
```

**Key Methods:**

| Method | Description |
|--------|-------------|
| `NewContainerEnvironment(env, dir, name)` | Creates environment (validates runtime) |
| `Validate()` | Validates Dockerfile exists and is valid |
| `EnsureImage()` | Builds image if not already present |
| `RunCommand(command)` | Executes command in container |
| `Shell(shellPath)` | Opens interactive shell |
| `RunCommandKeepAlive(command)` | Runs with `--keep` behavior |
| `StopContainer(name)` | Stops container |
| `RemoveContainer(name)` | Removes container |
| `ImageTag()` | Returns generated image tag |
| `RuntimeName()` | Returns "docker" or "podman" |

**Lifecycle:**

1. **Create**: `NewContainerEnvironment()` validates runtime, finds binary, generates image tag
2. **Validate**: `Validate()` checks Dockerfile exists and is valid
3. **Ensure**: `EnsureImage()` builds image if not cached
4. **Execute**: `RunCommand()` / `Shell()` / `RunCommandKeepAlive()`
5. **Cleanup**: `StopContainer()` / `RemoveContainer()` for keep-alive mode

#### Error Types for Containers

| Error Type | Description |
|------------|-------------|
| `NoSourceError` | `.source:` directive missing for container runtime |
| `InvalidRuntimeError` | Runtime not specified or not container type |
| `DockerfileNotFoundError` | Specified Dockerfile doesn't exist |
| `InvalidDockerfileError` | Dockerfile is malformed (missing FROM) |
| `ImageNotFoundError` | Container image not found locally |
| `ImageBuildError` | Error during image build |
| `ContainerRunError` | Error running container |

#### CLI Integration for Containers

The `--check-env` flag validates container environments:

```bash
build --check-env -e ci
```

**Output:**
```
Checking environment: ci
Runtime: docker

✓ Runtime: docker found
✓ Source: /path/to/Dockerfile
✓ Dockerfile is valid
  Image tag: project-ci:latest
  Image not yet built (will be built on first run)

Container environment ready
```

**Keep-Alive Instructions:**

```go
func PrintKeepInstructions(runtime, containerName string) string
```

Returns user-friendly instructions for managing kept-alive containers.

#### Container Tests (`container_test.go`)

| Test | Description |
|------|-------------|
| `TestDockerfileDetection` | Dockerfile location and validation |
| `TestDockerfileValidation` | FROM instruction validation |
| `TestContainerRuntimeDetection` | Runtime binary detection |
| `TestImageBuilder` | Image build command generation |
| `TestImageExists` | Image existence checking |
| `TestContainerRunner` | Container run command generation |
| `TestContainerEnvironment` | High-level environment validation |
| `TestPrintKeepInstructions` | Keep-alive instructions |
| `TestGenerateContainerName` | Container name generation |

#### Design Decisions

1. **Runtime abstraction**: Both Docker and Podman use the same interface, differing only in binary name.

2. **Image caching**: Images are tagged with project-specific names for reuse across builds.

3. **Workspace mounting**: The project directory is mounted at `/workspace` in the container for consistent paths.

4. **Validation-first**: Dockerfile is validated before any build/run operations.

5. **Keep-alive for debugging**: The `--keep` flag preserves the container for troubleshooting.

6. **Literal paths only**: `.source:` paths must be literal (no interpolation) to avoid chicken-and-egg evaluation issues.

## Output Package (`internal/output`)

The output package provides build output formatting and reporting. It uses an event-based architecture with `OutputWriter` and `Emitter` for flexible output across different contexts (CLI, TUI, headless/CI).

### Package Structure

| File | Contents |
|------|----------|
| `doc.go` | Package documentation |
| `events.go` | `OutputEvent` interface and event types |
| `writer.go` | `OutputWriter` interface and factory functions |
| `emitter.go` | `Emitter` for typed event emission |
| `cli.go` | `CLIWriter` for interactive terminal output |
| `headless.go` | `HeadlessWriter` for CI/log output |
| `tui.go` | `TUIWriter` for JSON event stream output |
| `reporter.go` | Legacy `Reporter` interface implementations |
| `reporter_emitter.go` | Emitter-backed reporter implementations |
| `reporter_test.go` | Unit tests for legacy reporters |
| `reporter_emitter_test.go` | Unit tests for emitter-backed reporters |

### Event-Based Output System

The output system uses events for all build output, enabling different rendering contexts.

#### OutputEvent Interface

```go
type OutputEvent interface {
    eventType() string
}
```

#### Event Types

| Event | Description |
|-------|-------------|
| `PhaseStarted` | Build phase begins (parse, semantic, eval, plan, execute) |
| `PhaseCompleted` | Build phase finishes |
| `VariableEvaluated` | Variable evaluation result (verbose mode) |
| `TargetStarted` | Target build begins |
| `TargetCompleted` | Target build finishes |
| `TargetSkipped` | Target is up to date |
| `CommandStarted` | Recipe command begins |
| `CommandOutput` | Command stdout/stderr |
| `CommandCompleted` | Command finishes |
| `StalenessChecked` | Staleness check result (verbose mode) |
| `BuildSummary` | Build completion summary |
| `ErrorOccurred` | Error during any phase |
| `DryRunTarget` | Target in dry-run mode |
| `DryRunCommand` | Command in dry-run mode |

#### OutputWriter Interface

```go
type OutputWriter interface {
    WriteEvent(event OutputEvent)
    Flush()
}
```

#### Emitter

The `Emitter` wraps an `OutputWriter` and provides typed methods for event emission:

```go
type Emitter struct {
    writer OutputWriter
}

func (e *Emitter) TargetStarted(target string, index, total int)
func (e *Emitter) TargetCompleted(target string, success bool, duration time.Duration, errMsg string)
func (e *Emitter) CommandOutput(target, stdout, stderr string)
func (e *Emitter) BuildSummary(total, succeeded, failed, skipped int, duration time.Duration)
// ... etc
```

### Emitter-Backed Reporters

All reporter implementations now delegate to the event-based output system while maintaining backward-compatible interfaces.

#### EmitterBackedNormalReporter

The `EmitterBackedNormalReporter` implements the `Reporter` interface using events:

```go
type EmitterBackedNormalReporter struct {
    emitter *Emitter
    index   int
    total   int
}
```

| Method | Event Emitted |
|--------|---------------|
| `BuildStarted(target)` | `TargetStarted` |
| `BuildCompleted(target, success, errMsg)` | `TargetCompleted` |
| `CommandOutput(cmd, stdout, stderr)` | `CommandOutput` |
| `Summary(total, failed)` | `BuildSummary` |
| `NothingToBuild(target)` | `TargetSkipped` |

#### EmitterBackedDryRunReporter

```go
type EmitterBackedDryRunReporter struct {
    emitter       *Emitter
    index         int
    total         int
    currentTarget string
}
```

| Method | Event Emitted |
|--------|---------------|
| `WouldBuild(target)` | `DryRunTarget` |
| `ShowCommand(command)` | `DryRunCommand` |
| `NothingToBuild(target)` | `TargetSkipped` |

#### EmitterBackedVerboseReporter

```go
type EmitterBackedVerboseReporter struct {
    emitter       *Emitter
    index         int
    total         int
    currentTarget string
}
```

| Method | Event Emitted |
|--------|---------------|
| `StartVariableEvaluation()` | `PhaseStarted("eval")` |
| `VariableEvaluated(name, expr, result)` | `VariableEvaluated` |
| `StartStalenessChecks()` | `PhaseStarted("plan")` |
| `StalenessCheck(target, reason, action)` | `StalenessChecked` |
| `BuildStarted(target)` | `TargetStarted` |
| `CommandExecuted(command)` | `CommandStarted` |

#### EmitterBackedProgressReporter

```go
type EmitterBackedProgressReporter struct {
    emitter   *Emitter
    total     int
    started   int
    completed int
    building  map[string]bool
}
```

| Method | Event Emitted |
|--------|---------------|
| `BuildStarted(target)` | `TargetStarted` (with progress) |
| `BuildCompleted(target, success, errMsg)` | `TargetCompleted` |
| `CurrentlyBuilding()` | Returns active targets map |

### Legacy Reporter Interface

The original `Reporter` interface is still available for backward compatibility:

```go
type Reporter interface {
    BuildStarted(target string)
    BuildCompleted(target string, success bool, errMsg string)
    CommandOutput(command, stdout, stderr string)
    Summary(total, failed int)
    NothingToBuild(target string)
}
```

The legacy `NormalReporter`, `DryRunReporter`, `VerboseReporter`, and `ProgressReporter` remain in `reporter.go` for direct use without the event system.

### CLI Integration

The CLI adapters in `cmd/need/output_adapter.go` now use emitter-backed reporters:

```go
// Creates an emitter-backed normal reporter
func NewNormalReporter(w io.Writer) OutputReporter

// Creates an emitter-backed dry-run reporter
func NewDryRunReporter(w io.Writer) DryRunOutputReporter

// Creates an emitter-backed verbose reporter
func NewVerboseReporter(w io.Writer) VerboseOutputReporter

// Creates an emitter-backed progress reporter
func NewProgressReporter(w io.Writer, total int) ProgressOutputReporter
```

Additional factory functions with config support:
```go
func NewNormalReporterWithConfig(w io.Writer, verbose, quiet bool, color string) OutputReporter
func NewDryRunReporterWithConfig(w io.Writer, verbose, quiet bool, color string) DryRunOutputReporter
func NewVerboseReporterWithConfig(w io.Writer, quiet bool, color string) VerboseOutputReporter
func NewProgressReporterWithConfig(w io.Writer, total int, verbose, quiet bool, color string) ProgressOutputReporter
```

### Unit Tests

#### Legacy Reporter Tests (`reporter_test.go`)

| Test | Description |
|------|-------------|
| `TestNormalReporter_BuildStarted` | Target name in output |
| `TestNormalReporter_BuildCompleted` | Success output |
| `TestNormalReporter_BuildCompletedFailure` | Failure with error message |
| `TestNormalReporter_CommandOutput` | Stdout content displayed |
| `TestNormalReporter_CommandOutputStderr` | Stderr content displayed |
| `TestNormalReporter_SuppressesEmptyOutput` | Empty output suppressed |
| `TestNormalReporter_Summary` | Summary with counts |
| `TestNormalReporter_SummaryAllSuccess` | Success message |
| `TestNormalReporter_NothingToBuild` | Up-to-date message |
| `TestDryRunReporter_*` | Dry-run output tests |
| `TestVerboseReporter_*` | Verbose output tests |
| `TestProgressReporter_*` | Progress output tests |

#### Emitter-Backed Reporter Tests (`reporter_emitter_test.go`)

| Test | Description |
|------|-------------|
| `TestEmitterBackedNormalReporter_*` | Normal reporter with event system |
| `TestEmitterBackedDryRunReporter_*` | Dry-run reporter with event system |
| `TestEmitterBackedVerboseReporter_*` | Verbose reporter with event system |
| `TestEmitterBackedProgressReporter_*` | Progress reporter with event system |
| `TestReporterCompatibility_*` | Compatibility between old and new reporters |

### Design Decisions

1. **Event-based architecture**: All output flows through events, enabling different rendering contexts (CLI, TUI, CI/headless).

2. **Backward-compatible interfaces**: The `Reporter`, `DryRunOutputReporter`, `VerboseOutputReporter`, and `ProgressOutputReporter` interfaces remain unchanged.

3. **Emitter-backed implementations**: All adapter factory functions now create emitter-backed reporters that delegate to `CLIWriter`.

4. **Writer injection**: Reporters accept an `io.Writer` for testability and flexibility.

5. **Empty output suppression**: Commands that produce no output don't generate any events, keeping the build log clean.

6. **Color and mode support**: The new system supports color modes (auto, always, never) and output modes (CLI, TUI, headless).

## Target Resolution (`cmd/need/target_resolve.go`)

The target resolution module handles command-line target argument parsing and resolution to canonical target names.

### Key Functions

| Function | Description |
|----------|-------------|
| `ResolveTargetArgs(args, result)` | Resolves CLI target arguments to canonical names |
| `extractTargetsAndDefault(stmts)` | Extracts targets and `.default:` from AST |
| `resolveDefaultTarget(targets, default)` | Resolves when no args provided |
| `resolveExplicitTargets(args, targets)` | Resolves explicit target arguments |
| `resolveTargetName(arg, targets)` | Resolves single argument to canonical name |
| `matchTargetPattern(target, path)` | Checks if path matches target pattern |

### Target Argument Resolution Rules

| Scenario | Resolution |
|----------|------------|
| No arguments, `.default:` present | Use `.default:` directive value |
| No arguments, no `.default:` | Use first defined target |
| `build/app` (file target) | Exact match against file targets |
| `@clean` (phony with @) | Match phony target with that name |
| `clean` (phony without @) | Match phony target `@clean` if exists |
| Pattern match (e.g., `build/main.o`) | Match against pattern targets |
| Unknown target | Return error |

### Canonical Target Names

Resolved targets are returned in canonical form:
- File targets: Path as-is (e.g., `build/app`)
- Phony targets: With `@` prefix (e.g., `@clean`, even if user typed `clean`)
- Pattern matches: Concrete path (e.g., `build/main.o` for `build/{name}.o`)

### Resolution Priority

When resolving a target argument without `@` prefix:
1. Try exact match against file targets first
2. Try match as phony target name (without @)
3. Try pattern target matching
4. Return "target not found" error

When resolving with `@` prefix:
1. Match only against phony targets
2. Return "target not found" error if no match

### CLI Integration

Target resolution is integrated into the main CLI flow:

```go
// In run() after parsing
resolvedTargets, err := ResolveTargetArgs(targets, result)
if err != nil {
    fmt.Fprintf(os.Stderr, "error: %v\n", err)
    return exitUsageError
}
```

### Unit Tests (`target_resolve_test.go`)

| Test | Description |
|------|-------------|
| `TestResolveTargetsNoArgs` | Default target resolution |
| `TestResolveTargetsSingleArg` | Single target argument |
| `TestResolveTargetsMultipleArgs` | Multiple targets in order |
| `TestResolveTargetArgsEmptySlice` | Empty slice behaves like nil |
| `TestRunWithTargetArgs` | CLI integration with explicit targets |
| `TestRunWithDefaultTarget` | CLI integration with default target |
| `TestRunWithFirstTargetAsDefault` | First target as default |
| `TestRunWithUnknownTarget` | Unknown target error |
| `TestRunWithNoTargets` | No targets defined error |

### Design Decisions

1. **Phony target without @ prefix**: Users can type `clean` instead of `@clean` for convenience. The resolver automatically adds the `@` prefix when returning the canonical name.

2. **File target priority**: When a name could match both a file target and a phony target (without @), file targets are checked first. This matches Make behavior where phony targets are secondary.

3. **Pattern matching support**: Pattern targets (e.g., `build/{name}.o`) are matched using the planner's `MatchTarget` function, ensuring consistent matching behavior.

4. **Error on first failure**: When resolving multiple targets, the first unresolvable target causes an error. This provides immediate feedback rather than proceeding with partial resolution.

5. **Empty args = default**: Both `nil` and empty slice `[]string{}` trigger default target resolution, matching the common case of running `build` with no arguments.


## Error Formatting Package (`internal/errors`)

The errors package provides structured error formatting for user-friendly error messages with source context.

### Package Structure

| File | Contents |
|------|----------|
| `doc.go` | Package documentation |
| `format.go` | `FormattedError`, `SourceLine`, extraction functions |
| `format_test.go` | Unit tests |
| `lexical.go` | Lexical error codes (E001-E099) |
| `lexical_test.go` | Lexical error tests |
| `syntax.go` | Syntax error codes (E100-E199) |
| `syntax_test.go` | Syntax error tests |
| `semantic.go` | Semantic error codes (E200-E299) |
| `semantic_test.go` | Semantic error tests |
| `evaluation.go` | Evaluation error codes (E300-E399) |
| `evaluation_test.go` | Evaluation error tests |
| `execution.go` | Execution error codes (E400-E499) |
| `execution_test.go` | Execution error tests |

### FormattedError Structure

```go
type FormattedError struct {
    Code        string             // Error code (E001, E100, etc.)
    Message     string             // Brief error description
    Location    ast.SourceLocation // Source location
    SourceLines []SourceLine       // Source context (1-3 lines)
    CaretLine   int                // Line to show caret on
    CaretColumn int                // Column for caret (1-based)
    Note        string             // Additional context (optional)
    Help        string             // Fix suggestion (optional)
}
```

### Error Format

Errors are formatted in a Rust-like style:

```
error[E100]: missing ':' in target definition
 --> Needfile:3:10
2 | cc = gcc
3 | build/app deps
  |          ^
4 |     gcc -o build/app deps
note: targets require ':' before dependencies
help: change to: build/app: deps
```

### Key Functions

| Function | Description |
|----------|-------------|
| `NewFormattedError(code, msg, loc)` | Creates basic error |
| `WithNote(note)` | Adds note to error |
| `WithHelp(help)` | Adds help suggestion |
| `WithSourceContext(lines, line, col)` | Adds source lines with caret |
| `ExtractSourceLines(source, line, ctx)` | Extracts lines from source string |
| `FormatCaret(width, column)` | Creates caret line |

### SourceLine Structure

```go
type SourceLine struct {
    Number int    // Line number (1-based)
    Text   string // Line content
}
```

### Error Code Categories

#### Lexical Errors (E001-E099)

| Code | Constructor | Description |
|------|-------------|-------------|
| E001 | `NewInvalidCharacterError` | Invalid character in source |
| E002 | `NewMixedIndentationError` | Mixed tabs and spaces in indentation |
| E003 | `NewInconsistentIndentationError` | Switched between tabs and spaces |
| E004 | `NewInvalidIndentWidthError` | Indentation not multiple of unit |
| E005 | `NewUnclosedInterpolationError` | Interpolation { not closed with } |
| E006 | `NewInvalidModifierError` | Invalid modifier (only :raw is valid) |
| E007 | `NewUnexpectedCharInInterpError` | Unexpected character inside {} |
| E008 | `NewInvalidEscapeSequenceError` | Invalid escape sequence |

#### Syntax Errors (E100-E199)

| Code | Constructor | Description |
|------|-------------|-------------|
| E100 | `NewUnexpectedTokenError` | Unexpected token during parsing |
| E101 | `NewMissingColonError` | Missing : in target definition |
| E102 | `NewMissingEndError` | Missing 'end' to close conditional |
| E103 | `NewInvalidDirectiveScopeError` | Directive used in invalid scope |
| E104 | `NewMissingConditionError` | Missing condition after if/elif |
| E105 | `NewMissingOperatorError` | Missing == or != in condition |
| E106 | `NewMissingIdentifierError` | Missing identifier after ifdef/ifndef |
| E107 | `NewInvalidRuntimeError` | Invalid runtime in .using directive |
| E108 | `NewMissingFunctionArgumentError` | Wrong argument count to function |
| E109 | `NewCircularIncludeError` | Circular include detected |
| E110 | `NewIncludeNotFoundError` | Included file not found |

#### Semantic Errors (E200-E299)

| Code | Constructor | Description |
|------|-------------|-------------|
| E200 | `NewUndefinedVariableError` | Reference to undefined variable |
| E201 | `NewDuplicateVariableError` | Variable defined multiple times |
| E202 | `NewDuplicateTargetError` | Target defined multiple times |
| E203 | `NewDuplicateEnvironmentError` | Environment defined multiple times |
| E204 | `NewCircularDependencyError` | Circular dependency in target graph |
| E205 | `NewCaptureConflictError` | Capture conflicts with variable/automatic |
| E206 | `NewCaptureMismatchError` | Capture in dep not in target pattern |
| E207 | `NewAutomaticOutsideRecipeError` | Automatic var outside recipe scope |
| E208 | `NewAutomaticInPatternError` | Automatic var in target pattern |

#### Evaluation Errors (E300-E399)

| Code | Constructor | Description |
|------|-------------|-------------|
| E300 | `NewShellCommandFailedError` | Shell command returned non-zero |
| E301 | `NewGlobNoMatchError` | Glob pattern matched no files |
| E302 | `NewInvalidFunctionArgumentsError` | Invalid function arguments |
| E303 | `NewForwardReferenceError` | Forward reference in immediate var |
| E304 | `NewLazyEvaluationError` | Error evaluating lazy variable |
| E305 | `NewConditionEvaluationError` | Error evaluating condition |

#### Execution Errors (E400-E499)

| Code | Constructor | Description |
|------|-------------|-------------|
| E400 | `NewRecipeFailedError` | Recipe command returned non-zero |
| E401 | `NewMissingDependencyError` | Required dependency file missing |
| E402 | `NewMissingBinaryError` | Required binary not in PATH |
| E403 | `NewShellNotFoundError` | Specified shell not found |
| E404 | `NewVersionMismatchError` | Binary version doesn't match spec |
| E405 | `NewTargetNotFoundError` | Requested target not defined |
| E406 | `NewNoDefaultTargetError` | No default target and none specified |

### Unit Tests

| Test | Description |
|------|-------------|
| `TestFormattedError_Basic` | Error with code, message, location |
| `TestFormattedError_WithNote` | Error with note section |
| `TestFormattedError_WithHelp` | Error with help section |
| `TestFormattedError_WithSourceContext` | Error with source lines |
| `TestFormattedError_CaretPosition` | Caret positioning |
| `TestFormattedError_Format` | Complete error format |
| `TestSourceLine_Format` | Line number formatting |
| `TestFormatCaret` | Caret line generation |
| `TestExtractSourceLines_*` | Source extraction cases |

### Design Decisions

1. **Rust-like format**: The error format follows Rust's excellent error messages with `error[CODE]: message`, source context, and caret pointers.

2. **Fluent interface**: Methods like `WithNote` and `WithHelp` return the error, enabling chained calls for building errors.

3. **Separate extraction**: Source extraction is separate from error creation, allowing errors to be created without file access (useful for testing).

4. **1-based columns**: Column numbers are 1-based to match editor conventions and SourceLocation.

5. **Flexible context**: The context parameter allows 1-3 lines of surrounding source, adapting to error type needs.

## Complete Execution Pipeline

need now has a fully functional execution pipeline wired up in `cmd/need/cli/main.go`. When a user runs `need target`, the following steps occur:

### Pipeline Stages

1. **Parse & Lex**: Read Needfile and convert to AST
2. **Semantic Analysis**: Run 4-pass semantic validation
   - Pass 1: Symbol collection (variables, targets, environments)
   - Pass 2: Capture validation (resolve {name} to capture vs interpolation)
   - Pass 3: Reference validation (check all variable references are defined)
   - Pass 4: Dependency graph validation (detect circular dependencies)
3. **Evaluation**: Evaluate all immediate variables and conditionals
4. **Planning**: For each requested target:
   - Match target pattern
   - Resolve dependencies recursively
   - Check file staleness (timestamps)
   - Build topologically sorted task list
5. **Execution**: For each task:
   - Create command context with automatic variables
   - Interpolate recipe commands
   - Execute with configured shell
   - Print output
   - Check exit codes

### Key Implementation

The execution loop in `main.go` (lines 295-374):
- Creates build plan for each target
- Configures executor with shell settings
- Loops through tasks in dependency order
- Creates CommandContext with automatic variables (`{target}`, `{deps}`, `{in}`, etc.)
- Sets captures for pattern targets
- Executes recipes
- Prints stdout/stderr
- Reports failures

### RealFileSystem Implementation

A `realFileSystem` type implements the FileSystem interface for actual file system operations:
- `Exists(path)`: Check if file exists
- `ModTime(path)`: Get file modification time

This enables staleness detection by comparing target and dependency timestamps.

### Parser Fix

Fixed parser bug where `IDENTIFIER` followed by `:` was not recognized as a file target (line 207-214 in `parser.go`). Now correctly parses simple file targets like `app:`.

## Integration Tests (`test/integration/`)

Comprehensive end-to-end integration tests verify the complete pipeline:

### Test Infrastructure (`integration_test.go`)

**TestHarness** provides:
- Temporary working directory for each test
- Automatic binary building
- File creation/reading utilities
- Build tool execution with argument passing
- Result assertions (exit code, stdout, stderr)

**RunResult** supports fluent assertions:
```go
result := h.Run("target")
result.AssertSuccess().
    AssertStdoutContains("expected output").
    AssertStderrNotContains("error")
```

### Basic Tests

| Test | Description |
|------|-------------|
| `TestSimpleBuild` | Build phony target with echo command |
| `TestNeedfileDiscovery` | Automatic Needfile discovery in current directory |
| `TestMissingNeedfile` | Error when no Needfile found |
| `TestDefaultTarget` | Build `.default:` target when none specified |
| `TestMultipleTargets` | Build multiple targets in order |
| `TestDryRun` | `--dry-run` shows commands without executing |
| `TestVerboseMode` | `--verbose` shows additional information |
| `TestHelpFlag` | `--help` displays usage |
| `TestVersionFlag` | `--version` shows version info |
| `TestInvalidFlag` | Invalid flags produce usage error |

### End-to-End Tests (`e2e_test.go`)

| Test | Description |
|------|-------------|
| `TestSimpleCCompilation` | Compile C program with gcc |
| `TestPatternTargets` | Pattern matching (`build/{name}.o`) |
| `TestConditionalCompilation` | OS-specific compilation flags |
| `TestPhonyTargets` | Phony targets and dependencies |
| `TestDependencyChain` | Correct dependency order |
| `TestVariableInterpolation` | Variable substitution in recipes |
| `TestLazyVariables` | Lazy variable on-demand evaluation |
| `TestBuiltInFunctions` | glob(), basename(), dirname(), replace() |
| `TestAutomaticVariables` | {target}, {deps}, {in}, {out}, {target.dir}, {target.file} |
| `TestBlockCommands` | Multi-line shell scripts with `block:` |
| `TestIncludeDirective` | `.include:` merges external files |
| `TestNestedIncludes` | Nested include chains (A → B → C) |
| `TestCircularIncludeDetection` | Detects and reports circular includes |
| `TestDeepNestedIncludes` | 4-level deep include chains |
| `TestIncludeWithTargets` | Targets from included files are buildable |
| `TestStalenessDetection` | Timestamp-based rebuild detection |
| `TestErrorHandling` | Failed commands report errors correctly |

## Test Fixtures

The project includes test fixtures for regression testing:

### Valid Fixtures (`test/integration/fixtures/valid/`)

| Fixture | Description |
|---------|-------------|
| `simple.need` | Basic phony target with echo command |
| `variables.need` | Immediate and lazy variable definitions |
| `conditionals.need` | OS-based conditional blocks |
| `patterns.need` | Pattern targets for C compilation |
| `dependencies.need` | Phony target dependency chains |
| `functions.need` | Built-in functions (dirname, basename, replace) |
| `block_commands.need` | Block commands with shell loops |
| `environment.need` | Environment blocks with requirements |

### Invalid Fixtures (`test/integration/fixtures/invalid/`)

| Fixture | Expected Error |
|---------|----------------|
| `missing_end.need` | Missing `end` for conditional |
| `undefined_var.need` | Reference to undefined variable |
| `wrong_scope.need` | Directive in wrong scope |
| `duplicate_target.need` | Duplicate target definition |
| `circular_dep.need` | Circular dependency chain |
| `mixed_indent.need` | Mixed tabs and spaces |

### Environment Tests (`environment_test.go`)

End-to-end tests for bare environment functionality and CLI environment commands:

| Test | Description |
|------|-------------|
| `TestBareEnvironmentNoRequirements` | Bare environment with no requirements builds successfully |
| `TestBareEnvironmentWithSatisfiedRequirements` | Build succeeds when required binaries (bash, sh) exist |
| `TestCheckEnvBareEnvironmentSatisfied` | `--check-env` reports all requirements satisfied |
| `TestCheckEnvBareEnvironmentMissing` | `--check-env` reports missing binary (exit code 4) |
| `TestCheckEnvNamedEnvironment` | `--check-env --env name` checks named environment |
| `TestCheckEnvNoEnvironment` | `--check-env` with no environment defined (bare environment) |
| `TestCheckEnvWithShowInstall` | `--check-env --show-install` shows install suggestions |
| `TestListEnvNoEnvironments` | `--list-env` with no environments defined |
| `TestListEnvSingleDefaultEnvironment` | `--list-env` shows single unnamed environment |
| `TestListEnvMultipleEnvironments` | `--list-env` shows all defined environments |
| `TestCheckEnvEnvironmentNotFound` | Error when `--env` specifies non-existent environment |
| `TestCheckEnvRequiresEnvSelection` | Error when only named environments and no selection |
| `TestCheckEnvVerbose` | `--check-env --verbose` shows path information |
| `TestBareEnvironmentWithVersionedRequirement` | `@latest` version spec always matches |
| `TestListEnvShowsRequirementCount` | Environment list shows requirement counts |
| `TestBuildWithBareEnvironmentSucceeds` | Full build succeeds with satisfied requirements |

### Docker Environment Tests (`environment_test.go`)

Docker environment integration tests verify the container environment CLI commands. These tests require Docker to be available on the host system.

**Skip Helper:**

The `skipIfNoDocker(t)` helper function skips tests when Docker is not available, enabling graceful test execution in environments without Docker.

| Test | Description |
|------|-------------|
| `TestDockerEnvironmentCheckEnv` | `--check-env` validates Docker environment with valid Dockerfile |
| `TestDockerEnvironmentCheckEnvMissingDockerfile` | Error when `.source:` references missing Dockerfile (exit code 4) |
| `TestDockerEnvironmentCheckEnvInvalidDockerfile` | Error when Dockerfile lacks FROM instruction |
| `TestDockerEnvironmentListEnv` | `--list-env` shows Docker environments with runtime type |
| `TestDockerEnvironmentDryRun` | `--dry-run` with Docker environment shows commands without executing |
| `TestDockerEnvironmentVerbose` | `--verbose --check-env` shows verbose Docker validation |
| `TestDockerEnvironmentRecipeFailure` | Recipe failures return exit code 1 |
| `TestDockerEnvironmentWithVariables` | Variable interpolation works in Docker environment recipes |
| `TestDockerEnvironmentSourceInSubdirectory` | Dockerfile in subdirectory path resolved correctly |
| `TestDockerEnvironmentNamedEnvironmentCheckEnv` | `--check-env --env name` validates named Docker environments |
| `TestDockerEnvironmentWithArgsInListEnv` | `--list-env` shows environments with `.args:` directive |
| `TestDockerEnvironmentNoDefaultWithNamedOnly` | Error when only named environments defined and no `--env` flag |
| `TestDockerEnvironmentMixedRuntimes` | `--list-env` shows both bare and Docker environments correctly |

### Podman Environment Tests (`environment_test.go`)

Podman environment integration tests mirror the Docker tests but use Podman as the container runtime. These tests require Podman to be available on the host system.

**Skip Helper:**

The `skipIfNoPodman(t)` helper function skips tests when Podman is not available, enabling graceful test execution in environments without Podman.

| Test | Description |
|------|-------------|
| `TestPodmanEnvironmentCheckEnv` | `--check-env` validates Podman environment with valid Containerfile |
| `TestPodmanEnvironmentCheckEnvMissingContainerfile` | Error when `.source:` references missing Containerfile (exit code 4) |
| `TestPodmanEnvironmentCheckEnvInvalidContainerfile` | Error when Containerfile lacks FROM instruction |
| `TestPodmanEnvironmentListEnv` | `--list-env` shows Podman environments with runtime type |
| `TestPodmanEnvironmentDryRun` | `--dry-run` with Podman environment shows commands without executing |
| `TestPodmanEnvironmentVerbose` | `--verbose --check-env` shows verbose Podman validation |
| `TestPodmanEnvironmentNamedEnvironmentCheckEnv` | `--check-env --env name` validates named Podman environments |
| `TestPodmanEnvironmentSourceInSubdirectory` | Containerfile in subdirectory path resolved correctly |
| `TestPodmanEnvironmentWithArgsInListEnv` | `--list-env` shows environments with `.args:` directive |
| `TestPodmanEnvironmentNoDefaultWithNamedOnly` | Error when only named environments defined and no `--env` flag |
| `TestPodmanEnvironmentMixedRuntimes` | `--list-env` shows both bare and Podman environments correctly |
| `TestMixedDockerPodmanEnvironments` | `--list-env` correctly lists environments using both Docker and Podman |

**Note on Container Execution:** The container execution infrastructure exists (`internal/environ/container_env.go`, `runner.go`, `image.go`) but is not yet wired into the main build execution path. The Docker and Podman tests focus on environment validation (`--check-env`, `--list-env`) which is fully implemented.

### Performance Tests (`performance_test.go`)

Performance tests validate that the parser and planner handle real-world scale Needfiles efficiently:

| Test | Description | Performance Target |
|------|-------------|-------------------|
| `TestLargeNeedfileParsing_1000Targets` | Parse Needfile with 1000 phony targets | < 10s |
| `TestLargeNeedfileParsing_ManyPatternTargets` | Parse 100 targets with file dependencies | < 10s |
| `TestDeepIncludeHierarchy` | Parse 20-level deep include chain | < 5s |
| `TestWideIncludeHierarchy` | Parse 50 sibling include files | < 10s |
| `TestDeepDependencyChain` | Plan 100-deep dependency chain | < 5s |
| `TestWideDependencyGraph` | Plan target with 200 direct dependencies | < 5s |
| `TestDiamondDependencyPerformance` | Plan multi-level diamond pattern (60+ targets) | < 5s |
| `TestManyVariables` | Evaluate 500 variables with references | < 5s |

**Observed Performance (as of implementation):**

| Test | Actual Time |
|------|-------------|
| 1000 targets | ~200ms |
| 100 pattern targets | ~178ms |
| 20-level deep includes | ~199ms |
| 50 sibling includes | ~184ms |
| 100-deep dependency chain | ~179ms |
| 200-wide dependency graph | ~177ms |
| Diamond dependency pattern | ~178ms |
| 500 variables | ~189ms |

All performance tests skip in `-short` mode via `testing.Short()`.

### Current Status

All integration tests pass:
```
=== RUN   TestSimpleBuild
--- PASS: TestSimpleBuild (0.40s)
=== RUN   TestNeedfileDiscovery
--- PASS: TestNeedfileDiscovery (0.40s)
=== RUN   TestBareEnvironmentNoRequirements
--- PASS: TestBareEnvironmentNoRequirements (0.65s)
=== RUN   TestCheckEnvBareEnvironmentSatisfied
--- PASS: TestCheckEnvBareEnvironmentSatisfied (0.40s)
...
PASS
ok  	github.com/vinayprograms/need/test/integration	26.507s
```

The core build pipeline is fully functional with comprehensive test coverage including environment handling.

## Output Beautification System (`internal/output`)

The output beautification system provides context-aware output formatting for CLI, TUI, and headless (CI) environments.

### Package Structure

| File | Contents |
|------|----------|
| `doc.go` | Package documentation |
| `emitter.go` | `Emitter` for typed event emission API |
| `events.go` | `OutputEvent` interface and all event types |
| `mode.go` | `OutputMode` enum and detection logic |
| `writer.go` | `OutputWriter` interface and factory functions |
| `color.go` | ANSI color utilities |
| `cli.go` | `CLIWriter` for interactive terminal output |
| `headless.go` | `HeadlessWriter` for CI/log collectors |
| `tui.go` | `TUIWriter` for structured JSON output |
| `reporter.go` | Legacy `Reporter` interface (to be refactored) |

### Emitter (`emitter.go`)

The `Emitter` provides a typed API for emitting output events. It wraps an `OutputWriter` and provides methods for each event type.

```go
type Emitter struct {
    writer OutputWriter
}

func NewEmitter(writer OutputWriter) *Emitter
func NoOpEmitter() *Emitter
```

#### Emitter Methods

| Method | Event Type | Description |
|--------|------------|-------------|
| `PhaseStarted(phase)` | `PhaseStarted` | Build phase begins |
| `PhaseCompleted(phase, duration)` | `PhaseCompleted` | Build phase finishes |
| `VariableEvaluated(name, expr, result)` | `VariableEvaluated` | Variable evaluated |
| `TargetStarted(target, index, total)` | `TargetStarted` | Target build begins |
| `TargetCompleted(target, success, duration, err)` | `TargetCompleted` | Target build finishes |
| `TargetSkipped(target, reason)` | `TargetSkipped` | Target skipped |
| `CommandStarted(target, command)` | `CommandStarted` | Command begins |
| `CommandOutput(target, stdout, stderr)` | `CommandOutput` | Command output |
| `CommandCompleted(target, cmd, exitCode, duration)` | `CommandCompleted` | Command finishes |
| `StalenessChecked(target, reason, action)` | `StalenessChecked` | Staleness check result |
| `BuildSummary(total, succeeded, failed, skipped, duration)` | `BuildSummary` | Build summary |
| `Error(category, code, message, location, context, hint)` | `ErrorOccurred` | Error occurred |
| `DryRunTarget(target, index, total)` | `DryRunTarget` | Dry-run target |
| `DryRunCommand(target, command)` | `DryRunCommand` | Dry-run command |

### Build Pipeline Integration

The emitter is integrated into the build pipeline components:

#### Executor Integration

The `Executor` accepts an emitter via `SetEmitter()` and emits:
- `CommandStarted` before each command
- `CommandOutput` when command produces output
- `CommandCompleted` after each command with duration and exit code
- `DryRunCommand` in dry-run mode

```go
executor := NewExecutor(config)
executor.SetEmitter(emitter)
executor.ExecuteRecipe(recipe, cmdCtx)
```

#### Planner Integration

The planner accepts an emitter via `PlanBuildWithEmitter()` and emits:
- `StalenessChecked` for each target with rebuild reason or skip status

```go
plan, err := PlanBuildWithEmitter(target, targets, ctx, fs, emitter)
```

#### Evaluator Integration

The `Evaluator` accepts an emitter via `SetEmitter()` and emits:
- `VariableEvaluated` for each variable with name, expression, and result
- Lazy variables show `<lazy>` as result

```go
evaluator := NewEvaluator(ctx)
evaluator.SetEmitter(emitter)
evaluator.EvaluateVariables(statements)
```

### Output Events (`events.go`)

All output is modeled as events that writers render appropriately:

| Event Type | Description |
|------------|-------------|
| `PhaseStarted` | Build phase begins (parse, semantic, eval, plan, execute) |
| `PhaseCompleted` | Build phase finishes with duration |
| `VariableEvaluated` | Variable evaluated (verbose mode) |
| `TargetStarted` | Target build begins with progress (index/total) |
| `TargetCompleted` | Target build finishes (success/failure) |
| `TargetSkipped` | Target skipped (up to date, etc.) |
| `CommandStarted` | Recipe command begins (verbose mode) |
| `CommandOutput` | Command produces stdout/stderr |
| `CommandCompleted` | Command finishes with exit code and duration |
| `StalenessChecked` | Staleness check result (verbose mode) |
| `BuildSummary` | Final build summary |
| `ErrorOccurred` | Error with code, location, context, hint |
| `DryRunTarget` | Target that would be built (dry-run mode) |
| `DryRunCommand` | Command that would run (dry-run mode) |

### Output Modes (`mode.go`)

| Mode | Description |
|------|-------------|
| `ModeCLI` | Interactive terminal with colors and formatting |
| `ModeTUI` | Structured JSON events for terminal UI |
| `ModeHeadless` | Plain text with timestamps for CI/logs |

#### Mode Detection

```go
func DetectOutputMode() OutputMode
```

Detection order:
1. `NEED_OUTPUT_MODE` environment variable (if set)
2. CI environment indicators (`CI`, `GITHUB_ACTIONS`, `GITLAB_CI`, etc.)
3. `TERM=dumb` → headless
4. TTY status of stdout → CLI if TTY, headless otherwise

### Color Utilities (`color.go`)

| Function | Description |
|----------|-------------|
| `ShouldUseColor(setting)` | Determines if colors should be used |
| `Colorize(text, color, enabled)` | Wraps text with ANSI color codes |
| `Bold(text, enabled)` | Makes text bold |
| `Dim(text, enabled)` | Makes text dim |
| `ColorizeStatus(text, success, enabled)` | Green for success, red for failure |

**Color control:**
- `NO_COLOR` environment variable disables colors
- `FORCE_COLOR` environment variable enables colors
- `--color=auto|always|never` CLI flag (to be integrated)

### Terminal Capabilities (`terminal.go`)

Detects terminal capabilities for adaptive output formatting.

**Color Levels:**

| Level | Description |
|-------|-------------|
| `ColorLevelNone` | No color support |
| `ColorLevelBasic` | 16-color ANSI support |
| `ColorLevel256` | 256-color support |
| `ColorLevelTruecolor` | 24-bit truecolor support |

**Detection Functions:**

| Function | Description |
|----------|-------------|
| `DetectCapabilities()` | Returns full terminal capabilities |
| `DetectColorLevel()` | Detects color support level |
| `DetectUnicodeSupport()` | Checks for UTF-8 locale |
| `GetTerminalSize()` | Returns terminal width/height |
| `DefaultTerminalSize()` | Returns default 80x24 |

**TerminalCapabilities struct:**

```go
type TerminalCapabilities struct {
    Width       int        // Terminal width (0 if unknown)
    Height      int        // Terminal height (0 if unknown)
    ColorLevel  ColorLevel // Supported color level
    Unicode     bool       // Unicode support
    Interactive bool       // Is an interactive terminal
}
```

**Helper methods:**
- `SupportsColor()` → true if any color support
- `Supports256Color()` → true if 256+ colors
- `SupportsTruecolor()` → true if 24-bit color
- `SizeWithFallback()` → uses 80x24 defaults if unknown

**Color level detection order:**
1. `NO_COLOR` → none
2. `TERM=dumb` → none
3. `COLORTERM=truecolor` or `24bit` → truecolor
4. `TERM` contains `256color` → 256
5. `TERM` matches known color terminals → basic
6. Otherwise → none

**Unicode detection:**
- Checks `LC_ALL`, `LC_CTYPE`, `LANG` for `UTF-8` or `utf8`

### CLI Writer (`cli.go`)

Interactive terminal output with colors and progress indicators.

**Features:**
- Colored target names and status
- Progress formatting for parallel builds (`[n/total]`)
- Verbose mode shows variable evaluation and staleness checks
- Quiet mode suppresses non-error output
- Error display with source context and hints
- Degraded output for limited terminals (ASCII fallback when Unicode unsupported)

**Symbol Sets:**

| Context | Unicode | ASCII Fallback |
|---------|---------|----------------|
| Success | ✓ | [ok] |
| Failure | ✗ | [FAIL] |
| Arrow | → | -> |
| Bullet | • | * |

**Unicode Control:**
- `--unicode=auto` (default): Detect from locale
- `--unicode=always`: Force Unicode symbols
- `--unicode=never`: Use ASCII-only symbols

**Example output (normal):**
```
Building foo.o
Built foo.o

Build success: 1 target built
```

**Example output (verbose):**
```
Evaluating variables...
  cc → gcc
  sources = shell(find src -name "*.c") → src/main.c

Checking targets...
  foo.o: src/main.c is newer → rebuild

Building foo.o...
  gcc -c src/main.c -o foo.o
Built foo.o (0.2s)

Done.
```

### Headless Writer (`headless.go`)

Plain text output with timestamps for CI/CD and log collectors.

**Features:**
- RFC3339 timestamps on all log lines
- Log levels (DEBUG, INFO, WARN, ERROR)
- Optional JSON log format (`BUILD_LOG_FORMAT=json`)
- No ANSI escape sequences

**Text format:**
```
[2024-01-15T10:30:00Z] [INFO] Building target target=foo.o index=1 total=5
[2024-01-15T10:30:01Z] [INFO] Target built target=foo.o duration_ms=200
```

**JSON format:**
```json
{"time":"2024-01-15T10:30:00Z","level":"INFO","msg":"Building target","target":"foo.o"}
```

### TUI Writer (`tui.go`)

Structured JSON events for terminal UI applications.

**Output format:**
```json
{"type":"target_started","target":"foo.o","index":1,"total":5,"timestamp":"..."}
{"type":"target_completed","target":"foo.o","success":true,"duration_ms":200,"timestamp":"..."}
```

### Writer Configuration

```go
type WriterConfig struct {
    Verbose   bool   // Enable detailed output
    Quiet     bool   // Suppress non-error output
    Color     string // "auto", "always", "never"
    LogLevel  string // "debug", "info", "warn", "error"
    LogFormat string // "text", "json"
}
```

### Writer Factory

```go
func NewWriter(mode OutputMode, w io.Writer, config WriterConfig) OutputWriter
func NewWriterWithMode(mode OutputMode, config WriterConfig) OutputWriter
func NewWriterConfigFromFlags(verbose, quiet bool, color string) WriterConfig
func NewDefaultWriter(config WriterConfig) OutputWriter
func NewNoOpWriter() OutputWriter
```

**CLI Integration:**

```go
// In cmd/need/output_adapter.go
func CreateOutputEmitter(verbose, quiet bool, color string) *output.Emitter
func CreateOutputWriter(verbose, quiet bool, color string) output.OutputWriter
```

### Design Decisions

1. **Event-based architecture**: All output is modeled as events, allowing different renderers for different contexts without changing the build pipeline.

2. **Mode detection**: Automatically detects CLI vs headless based on TTY and CI environment, reducing user configuration burden.

3. **Color control**: Follows standards (`NO_COLOR`, `FORCE_COLOR`) and provides explicit control via config.

4. **Verbose as event filtering**: Verbose mode events are always emitted but filtered by the writer, keeping the pipeline simple.

5. **JSON for TUI**: Structured output enables rich terminal UIs without parsing text.

6. **Timestamp format**: RFC3339 for headless mode ensures consistent log parsing.

### Unit Tests

#### Events Tests (`events_test.go`)

| Test | Description |
|------|-------------|
| `TestPhaseStartedEventType` | Event type string |
| `TestPhaseCompletedEventType` | Event type string |
| `TestVariableEvaluatedEventType` | Event type string |
| `TestTargetStartedEventType` | Event type string |
| `TestTargetCompletedEventType` | Event type string |
| `TestTargetSkippedEventType` | Event type string |
| `TestCommandStartedEventType` | Event type string |
| `TestCommandOutputEventType` | Event type string |
| `TestCommandCompletedEventType` | Event type string |
| `TestStalenessCheckedEventType` | Event type string |
| `TestBuildSummaryEventType` | Event type string |
| `TestErrorOccurredEventType` | Event type string |
| `TestDryRunTargetEventType` | Event type string |
| `TestDryRunCommandEventType` | Event type string |
| `TestAllEventsImplementInterface` | Compile-time interface check |

#### Mode Tests (`mode_test.go`)

| Test | Description |
|------|-------------|
| `TestOutputModeString` | Mode string representation |
| `TestParseOutputMode` | Parse mode from string |
| `TestDetectOutputMode_EnvOverride` | NEED_OUTPUT_MODE override |
| `TestDetectOutputMode_CI` | CI environment detection |
| `TestDetectOutputMode_DumbTerminal` | TERM=dumb detection |
| `TestIsCI` | CI indicator detection |

#### Color Tests (`color_test.go`)

| Test | Description |
|------|-------------|
| `TestShouldUseColor_Always` | Always enable colors |
| `TestShouldUseColor_Never` | Always disable colors |
| `TestShouldUseColor_Auto_NoColor` | NO_COLOR detection |
| `TestShouldUseColor_Auto_ForceColor` | FORCE_COLOR detection |
| `TestColorize_Enabled` | Color application |
| `TestColorize_Disabled` | Color skipped |
| `TestBold_Enabled` | Bold formatting |
| `TestDim_Enabled` | Dim formatting |
| `TestColorizeStatus_Success` | Green for success |
| `TestColorizeStatus_Failure` | Red for failure |

#### CLI Writer Tests (`cli_test.go`)

| Test | Description |
|------|-------------|
| `TestCLIWriter_TargetStarted` | Target start output |
| `TestCLIWriter_TargetStarted_Progress` | Progress format [n/total] |
| `TestCLIWriter_TargetCompleted_Success` | Success output |
| `TestCLIWriter_TargetCompleted_Failure` | Failure output |
| `TestCLIWriter_CommandOutput` | Command stdout/stderr |
| `TestCLIWriter_BuildSummary_Success` | Success summary |
| `TestCLIWriter_BuildSummary_Failure` | Failure summary |
| `TestCLIWriter_Error` | Error with code/location/hint |
| `TestCLIWriter_Verbose_VariableEvaluated` | Verbose variable output |
| `TestCLIWriter_Verbose_CommandStarted` | Verbose command output |
| `TestCLIWriter_Quiet_SuppressesNonErrors` | Quiet mode filtering |
| `TestCLIWriter_Quiet_ShowsErrors` | Errors shown in quiet mode |
| `TestCLIWriter_DryRun` | Dry-run output format |
| `TestCLIWriter_WithColor` | ANSI codes present |
| `TestCLIWriter_StalenessChecked` | Staleness output |
| `TestCLIWriter_TargetSkipped` | Skip output |
| `TestCLIWriter_VerboseDuration` | Duration in output |

#### Headless Writer Tests (`headless_test.go`)

| Test | Description |
|------|-------------|
| `TestHeadlessWriter_TargetStarted_Text` | Text log format |
| `TestHeadlessWriter_TargetStarted_JSON` | JSON log format |
| `TestHeadlessWriter_Error` | Error logging |
| `TestHeadlessWriter_LogLevel_Debug` | Debug level output |
| `TestHeadlessWriter_LogLevel_Info_FiltersDebug` | Level filtering |
| `TestHeadlessWriter_CommandOutput` | Raw output passthrough |
| `TestHeadlessWriter_BuildSummary_Success` | Success summary log |
| `TestHeadlessWriter_BuildSummary_Failure` | Failure summary log |
| `TestHeadlessWriter_Quiet` | Quiet mode |
| `TestHeadlessWriter_Timestamp` | Timestamp presence |

#### TUI Writer Tests (`tui_test.go`)

| Test | Description |
|------|-------------|
| `TestTUIWriter_TargetStarted` | JSON event structure |
| `TestTUIWriter_TargetCompleted` | Success/failure fields |
| `TestTUIWriter_Error` | Error event structure |
| `TestTUIWriter_BuildSummary` | Summary fields |
| `TestTUIWriter_HasTimestamp` | Timestamp field |
| `TestTUIWriter_PhaseEvents` | Phase event structure |
| `TestTUIWriter_DryRunEvents` | Dry-run event structure |

#### Writer Factory Tests (`writer_test.go`)

| Test | Description |
|------|-------------|
| `TestNewWriter_CLI` | CLI mode creates CLIWriter |
| `TestNewWriter_TUI` | TUI mode creates TUIWriter |
| `TestNewWriter_Headless` | Headless mode creates HeadlessWriter |
| `TestNewNoOpWriter` | NoOp writer discards output |
| `TestDefaultWriterConfig` | Default config values |

## Documentation

### Design Documents

| File | Description |
|------|-------------|
| `design/NEEDFILE_SPEC.md` | Complete language specification |
| `design/DESIGN.md` | System architecture and design decisions |
| `design/CODE.md` | Implementation documentation (this file) |
| `design/TODO.md` | Implementation task tracking |
| `design/MIGRATION.md` | Guide for migrating from Make to Needfile |

### Migration Guide (`design/MIGRATION.md`)

The migration guide provides comprehensive documentation for users transitioning from GNU Make to Needfile. It covers:

**Quick Reference Table:**
- Variable syntax mapping (`$(VAR)` → `{var}`)
- Automatic variable mapping (`$@` → `{target}`, etc.)
- Pattern rule conversion (`%` → `{name}`)
- Function mapping (`$(shell)` → `shell()`, etc.)

**Detailed Sections:**
- Variable assignment (simple vs lazy)
- Target and dependency syntax
- Pattern rules with named captures
- Phony and directory targets
- Automatic variables comparison
- Shell commands and block mode
- Built-in functions
- Conditionals (`ifeq` → `if`)
- Include directive
- Common patterns (C project example)
- Auto-dependencies
- Migration checklist

**Key Differences Highlighted:**
- No tab requirement (use any consistent indent)
- Readable automatic variable names
- Named pattern captures vs `%`
- Built-in `{os}` and `{arch}` variables
- Block mode for multi-line scripts
- Environment management features

## Cache Package (`internal/cache`)

The cache package provides caching for parsed Needfiles to avoid re-parsing unchanged files.

### NeedfileCache (`needfile.go`)

#### Purpose

Stores parsed AST keyed by absolute file path. Cache entries are invalidated when the file's modification time changes.

#### Structure

```go
type NeedfileCache struct {
    mu      sync.RWMutex
    entries map[string]*cacheEntry
}

type cacheEntry struct {
    statements []ast.Statement
    modTime    time.Time
}
```

#### Key Methods

| Method | Description |
|--------|-------------|
| `NewNeedfileCache()` | Creates an empty cache |
| `Put(path, statements) error` | Stores parsed statements with file mtime |
| `Get(path) ([]Statement, bool, error)` | Retrieves cached statements, validates mtime |
| `Invalidate(path)` | Removes a specific entry |
| `Clear()` | Removes all entries |
| `Size() int` | Returns the number of entries |

#### Cache Invalidation Rules

1. **File modification**: If file mtime differs from cached mtime, entry is invalidated
2. **File deletion**: If file no longer exists, entry is invalidated
3. **Explicit invalidation**: `Invalidate()` or `Clear()` removes entries

#### Path Normalization

- Paths are normalized to absolute paths before storage
- Both relative and absolute paths work as keys
- Same file accessed via different paths resolves to same cache entry

### CLI Integration (`cmd/need/cache_adapter.go`)

The CLI uses a global cache singleton:

```go
// Global cache singleton
var globalNeedfileCache NeedfileCache

// ParseNeedfileWithCache uses the global cache
func ParseNeedfileWithCache(needfile string) (NeedfileResult, string, error)
```

#### Cached Parsing Flow

1. Try to get from cache
2. On cache hit: return cached statements
3. On cache miss: parse file, cache result (if no errors)
4. Return result

#### cachedNeedfileResultAdapter

Wraps cached AST statements as a `NeedfileResult`:

```go
type cachedNeedfileResultAdapter struct {
    statements []ast.Statement
}
```

- Implements `NeedfileResult` interface
- Always reports no errors (only successful parses are cached)
- `GetASTStatements()` handles both regular and cached results

### Design Decisions

1. **In-memory only**: Cache is not persisted across process invocations. Future versions may add disk-based persistence.

2. **Mtime-based invalidation**: Simple and reliable. File content hash would be more precise but adds overhead.

3. **Global singleton**: Single cache instance for the CLI. Reduces complexity and ensures cache hits across multiple targets.

4. **No error caching**: Only successful parses are cached. Parse errors require fresh parsing to provide accurate error messages.

5. **Thread-safe**: Uses `sync.RWMutex` for concurrent access (future parallel parsing).

### AutodepsCache (`autodeps.go`)

#### Purpose

Caches parsed `.d` file contents (autodeps) keyed by absolute path. These are dependency files generated by compilers (e.g., `gcc -MD`) that list header file dependencies.

#### Structure

```go
type AutodepsCache struct {
    mu      sync.RWMutex
    entries map[string]*autodepsEntry
}

type autodepsEntry struct {
    deps    []string
    modTime time.Time
}
```

#### Key Methods

| Method | Description |
|--------|-------------|
| `NewAutodepsCache()` | Creates an empty cache |
| `Put(path, deps) error` | Stores parsed dependencies with file mtime |
| `Get(path) ([]string, bool, error)` | Retrieves cached dependencies, validates mtime |
| `Invalidate(path)` | Removes a specific entry |
| `Clear()` | Removes all entries |
| `Size() int` | Returns the number of entries |

#### Integration with Planner

The planner's `resolveAutodeps()` function uses the cache:

```go
// Try cache first if available
if p.autodepsCache != nil {
    if deps, ok, err := p.autodepsCache.Get(autodepsPath); err == nil && ok {
        return deps, autodepsPath, nil
    }
}

// Parse the .d file to get learned dependencies
deps, err := ParseAutodepsFile(autodepsPath)

// Cache the result if cache is available
if p.autodepsCache != nil && deps != nil {
    _ = p.autodepsCache.Put(autodepsPath, deps)
}
```

#### CLI Integration

A global singleton is used:

```go
var globalAutodepsCache *cache.AutodepsCache

func GetAutodepsCache() *cache.AutodepsCache {
    if globalAutodepsCache == nil {
        globalAutodepsCache = cache.NewAutodepsCache()
    }
    return globalAutodepsCache
}
```

The cache is passed to `PlanBuildWithOptions()` from the CLI.

#### Design Decisions

1. **Separate from Needfile cache**: Different data structures ([]string vs []ast.Statement) and use cases (planner vs parser).

2. **Mtime-based invalidation**: Same strategy as Needfile cache. When a build regenerates a `.d` file, the cache entry is automatically invalidated.

3. **Missing file handling**: `ParseAutodepsFile()` returns `(nil, nil)` for missing files. These are not cached since the file may be created by the build.

4. **Thread-safe**: Uses `sync.RWMutex` for concurrent access in parallel builds.

### TargetIndex (`internal/planner/index.go`)

#### Purpose

Provides optimized target lookup using pre-compiled patterns and prefix-based indexing. This reduces lookup time from O(n) to O(1) for exact matches and significantly improves pattern matching performance.

#### CompiledPattern

Pre-processed pattern representation:

```go
type CompiledPattern struct {
    Target     *ast.Target
    Prefix     string   // Literal prefix before first capture
    IsLiteral  bool     // True if no captures
    IsPhony    bool     // True if phony target
    PatternStr string   // String representation
    Captures   []string // Capture names in order
}
```

#### TargetIndex Structure

```go
type TargetIndex struct {
    literals map[string]*ast.Target       // Exact path → target
    phonies  map[string]*ast.Target       // Phony name → target
    byPrefix map[string][]*CompiledPattern // Prefix → patterns
    all      []*CompiledPattern            // All patterns for fallback
}
```

#### Lookup Algorithm

1. **Phony check**: If path has `@` prefix, lookup in `phonies` map
2. **Exact match**: Lookup in `literals` map (O(1))
3. **Phony fallback**: Try `phonies` map without `@` prefix
4. **Prefix match**: Get patterns with matching prefix, try each
5. **Empty prefix**: Try patterns with no literal prefix
6. **Fallback**: Try all non-literal, non-phony patterns

#### Performance Characteristics

| Lookup Type | Complexity |
|-------------|------------|
| Exact literal match | O(1) |
| Phony target match | O(1) |
| Pattern with prefix | O(k) where k = patterns with same prefix |
| Pattern without prefix | O(n) worst case |

#### Integration

The planner creates a `TargetIndex` at initialization:

```go
func PlanBuildWithOptions(...) (*BuildPlan, error) {
    planner := &buildPlanner{
        targets:     targets,
        targetIndex: NewTargetIndex(targets),
        // ...
    }
}
```

Target lookup uses the index:

```go
target, captures, err := p.targetIndex.Lookup(targetPath)
```

## Platform Package (`internal/platform`)

The platform package provides cross-platform utilities for path handling, shell execution, and platform detection. It abstracts platform-specific differences between Unix systems (Linux, macOS, BSD) and Windows.

### Package Structure

| File | Contents |
|------|----------|
| `doc.go` | Package documentation |
| `shell.go` | Shell detection, command args, path handling, quoting |
| `platform_test.go` | Unit tests for all platform functions |

### Shell Execution

On Unix systems, the default shell is `/bin/sh` with the `-c` flag. On Windows, `cmd.exe` is used with the `/C` flag, or PowerShell with `-Command`.

#### DefaultShell Function

```go
func DefaultShell() string
```

Returns the platform-appropriate default shell:
- **Unix**: `/bin/sh`
- **Windows**: `cmd.exe`

#### ShellCommandArgs Function

```go
func ShellCommandArgs(shell, command string) []string
```

Returns the command-line arguments to pass a command to a shell:

| Shell | Arguments |
|-------|-----------|
| `/bin/sh`, `bash`, `zsh` | `["-c", command]` |
| `cmd.exe`, `cmd` | `["/C", command]` |
| `powershell.exe`, `pwsh` | `["-Command", command]` |

#### IsWindowsShell Function

```go
func IsWindowsShell(shell string) bool
```

Returns true if the shell is a Windows shell (cmd.exe or PowerShell). Handles:
- Base name extraction from full paths (e.g., `C:\Windows\System32\cmd.exe`)
- Case-insensitive matching
- With or without `.exe` extension

### Path Handling

#### IsAbsolutePath Function

```go
func IsAbsolutePath(path string) bool
```

Returns true if the path is an absolute path. Handles both Unix and Windows styles:

| Path | Result | Reason |
|------|--------|--------|
| `/usr/bin/sh` | true | Unix absolute path |
| `C:\Windows\System32` | true | Windows drive letter |
| `c:/windows` | true | Windows with forward slash |
| `\\server\share` | true | Windows UNC path |
| `./bin/sh` | false | Relative path |
| `Windows\System32` | false | Relative path |

#### IsDirectoryPath Function

```go
func IsDirectoryPath(path string) bool
```

Returns true if the path ends with a directory separator (`/` or `\`).

#### PathSeparator Function

```go
func PathSeparator() byte
```

Returns the OS-specific path separator:
- **Unix**: `/`
- **Windows**: `\`

#### NormalizePath Function

```go
func NormalizePath(path string) string
```

Converts all backslashes to forward slashes for internal consistency.

### Shell Quoting

#### ShellQuote Function

```go
func ShellQuote(shell, value string) string
```

Quotes a string for safe use in shell commands. The quoting style depends on the shell:

| Shell | Quoting Style | Example |
|-------|---------------|---------|
| Unix shells (sh, bash) | Single quotes | `'hello world'` |
| cmd.exe | Double quotes | `"hello world"` |
| PowerShell | Single quotes | `'hello world'` |

**Special characters handled:**
- **cmd.exe**: Space, tab, `&`, `|`, `<`, `>`, `^`, `(`, `)`, `%`, `!`
- **PowerShell**: Space, tab, `'`, `"`, `$`, `` ` ``
- **Unix**: Space, tab, newline, `'`, `"`, `\`, `` ` ``, `$`, `!`, `*`, `?`, `[`, `]`, `{`, `}`, `(`, `)`, `;`, `&`, `|`, `<`, `>`, `#`, `~`

**Embedded quote handling:**
- **Unix**: `it's` → `'it'"'"'s'` (end quote, double-quoted single quote, start quote)
- **cmd.exe**: `hello"world` → `"hello\"world"`
- **PowerShell**: `it's` → `'it''s'` (doubled single quote)

### Shell Validation

#### ValidateShell Function

```go
func ValidateShell(shell string) error
```

Checks that the shell exists and is executable using `exec.LookPath`. Handles both absolute paths and PATH lookup.

### Platform Detection

#### IsWindows Function

```go
func IsWindows() bool
```

Returns true if running on Windows (`runtime.GOOS == "windows"`).

### Windows Package Managers

The `internal/environ/install.go` module includes support for Windows package managers:

| Package Manager | Detection | Install Command |
|----------------|-----------|-----------------|
| winget | `winget` binary | `winget install package` |
| Chocolatey | `choco` binary | `choco install package` |
| Scoop | `scoop` binary | `scoop install package` |

Package manager detection order:
1. winget (Windows Package Manager - official)
2. choco (Chocolatey)
3. scoop (Scoop)

### Integration with Executor

The executor package uses the platform package for cross-platform shell execution:

```go
// In NewShellConfig()
return &ShellConfig{
    Shell: platform.DefaultShell(),
}

// In ExecuteLine()
args := platform.ShellCommandArgs(e.config.Shell, cmdLine)
cmd := exec.Command(e.config.Shell, args...)

// In Validate()
err := platform.ValidateShell(c.Shell)
```

### Integration with Eval

The eval package uses platform functions:

```go
// In funcShell()
shell := platform.DefaultShell()
args := platform.ShellCommandArgs(shell, cmd)
shellCmd := exec.Command(shell, args...)

// In setTargetDirAndFile()
if platform.IsDirectoryPath(target) {
    c.automatic["target.dir"] = strings.TrimSuffix(...)
}
```

### Unit Tests

| Test | Description |
|------|-------------|
| `TestDefaultShell` | Default shell per platform |
| `TestShellCommandArgs` | Command args for all shell types |
| `TestIsAbsolutePath` | Unix and Windows absolute paths |
| `TestIsWindowsShell` | Windows shell detection with full paths |
| `TestIsDirectoryPath` | Directory path detection |
| `TestPathSeparator` | Platform path separator |
| `TestNormalizePath` | Backslash to forward slash conversion |
| `TestShellQuoteWindows` | Shell quoting for cmd.exe, PowerShell, bash |
| `TestIsWindows` | Platform detection |
| `TestCmdExeSpecialChars` | cmd.exe special character handling |

### Design Decisions

1. **Platform abstraction**: All platform-specific code is isolated in the platform package, allowing the rest of the codebase to work consistently across platforms.

2. **Cross-platform path handling**: The `baseName()` function handles both `/` and `\` as separators regardless of the current platform, allowing Windows paths to be processed on Unix and vice versa.

3. **Conservative shell detection**: Uses case-insensitive matching and handles `.exe` extension optionally to work with both Windows and WSL environments.

4. **Exit code portability**: The executor uses Go's cross-platform `ExitCode()` method instead of Unix-specific `syscall.WaitStatus`.

5. **Quoting strategy**: Different quoting strategies per shell ensure command arguments are passed correctly to each shell type.

### Windows Limitations and Future Work

Current Windows support includes:
- Path separator handling (both `/` and `\` recognized)
- Shell selection (cmd.exe, PowerShell)
- Shell-specific command arguments
- Shell-specific quoting
- Windows package managers (winget, choco, scoop)

Future enhancements may include:
- WSL (Windows Subsystem for Linux) integration
- PowerShell Core (pwsh) as default shell option
- Windows-specific environment variable handling
- Long path support (>260 characters)

