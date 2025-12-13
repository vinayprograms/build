# Build Tool - Code Design

This document tracks the implementation architecture and design decisions as code is written.

## Package Structure

```
github.com/vinayprograms/build/
├── cmd/
│   └── build/          # CLI entry point
│       ├── main.go
│       └── main_test.go
├── internal/
│   ├── ast/            # Abstract Syntax Tree
│   │   ├── ast.go          # AST node type definitions
│   │   └── ast_test.go
│   ├── lexer/          # Lexical analysis
│   │   ├── indent.go       # Indentation tracking
│   │   ├── indent_test.go
│   │   ├── interp.go       # Interpolation boundary detection
│   │   ├── interp_test.go
│   │   ├── lexer.go        # Main lexer implementation
│   │   ├── lexer_test.go
│   │   ├── token.go        # Token types and source location
│   │   └── token_test.go
│   └── parser/         # Syntactic analysis
│       ├── directive.go    # Directive scope validation
│       ├── directive_test.go
│       ├── errors.go       # Parse error types
│       ├── errors_test.go
│       ├── parser.go       # Parser with scope stack
│       ├── parser_test.go
│       ├── scope.go        # Scope types and stack
│       └── scope_test.go
├── Buildfile           # Build configuration for this project
└── go.mod
```

## CLI (`cmd/build/main.go`)

The command-line interface for the build tool.

### Architecture

The CLI follows interface-based design where `cmd/build` defines the interfaces and internal packages provide implementations:

```
cmd/build/
├── main.go         # CLI entry point and flag handling
├── interfaces.go   # Interface definitions (Lexer, Parser, Token, Scope)
└── adapters.go     # Adapters wrapping internal packages
```

**Key Interfaces:**

| Interface | Description |
|-----------|-------------|
| `Token` | Represents a lexical token with type, literal, and location |
| `Lexer` | Tokenizes source code into a stream of tokens |
| `Scope` | Represents parsing context (global, environment, recipe, block) |
| `Parser` | Transforms token stream into AST with scope tracking |
| `DirectiveValidator` | Validates directive placement at scopes |

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
| 3 | Parse error (invalid Buildfile) |
| 4 | Environment error (missing requirements) |

### Flags

All flags from BUILDFILE_SPEC.md are implemented:

| Flag | Description |
|------|-------------|
| `-f, --file` | Use alternate Buildfile |
| `-e, --env` | Use named environment |
| `-j, --jobs` | Parallel jobs |
| `-n, --dry-run` | Show what would execute |
| `-v, --verbose` | Verbose output |
| `--check-env` | Verify environment requirements |
| `--list-env` | List available environments |
| `-V, --version` | Show version |
| `-h, --help` | Show help |

### Debug Flags

| Flag | Description |
|------|-------------|
| `--debug-lex` | Dump lexer analysis (indentation, interpolations) |
| `--debug-parse` | Dump parser scope validation |

### Version Information

Version and commit are embedded at build time via `-ldflags`:

```bash
go build -ldflags "-X main.version=v1.0.0 -X main.commit=abc123" ./cmd/build
```

## Lexer Package (`internal/lexer`)

### Token Types (`token.go`)

Defines all token types for the Buildfile language as specified in `DESIGN.md` Section 2.2.

#### Token Categories

| Category | Token Types | Description |
|----------|-------------|-------------|
| Special | `EOF`, `NEWLINE`, `INDENT`, `COMMENT`, `ERROR` | Control tokens |
| Dot Keywords | `DOT_SHELL`, `DOT_PARALLEL`, `DOT_DEFAULT`, `DOT_INCLUDE`, `DOT_ENVIRONMENT`, `DOT_USING`, `DOT_SOURCE`, `DOT_ARGS`, `DOT_REQUIRES`, `DOT_AFTER`, `DOT_AUTODEPS` | Directives starting with `.` |
| Keywords | `LAZY`, `IF`, `ELIF`, `ELSE`, `END`, `IFDEF`, `IFNDEF`, `BLOCK` | Control flow and modifiers |
| Operators | `EQUALS`, `COLON`, `DOUBLE_EQUALS`, `NOT_EQUALS`, `LPAREN`, `RPAREN` | Punctuation |
| Identifiers | `IDENTIFIER`, `AT_IDENTIFIER`, `PATH`, `STRING` | Names and literals |
| Interpolation | `INTERP_START`, `INTERP_END`, `INTERP_MOD`, `ESCAPE_LBRACE`, `ESCAPE_RBRACE` | `{var}` syntax |
| Functions | `FUNC_SHELL`, `FUNC_GLOB`, `FUNC_BASENAME`, `FUNC_DIRNAME`, `FUNC_REPLACE` | Built-in functions |

#### SourceLocation

Tracks position in source files for error reporting:
- `File`: Source file path
- `Line`: 1-based line number
- `Column`: 1-based column number

Format: `file:line:column` (e.g., `Buildfile:10:5`)

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

Stateful tracker that enforces consistent indentation across a Buildfile:

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

### Lexer (`lexer.go`)

The main lexer implementation that tokenizes Buildfile source code.

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

## AST Package (`internal/ast`)

Defines the Abstract Syntax Tree node types for Buildfile parsing. The AST captures syntactic structure without interpretation—no evaluation happens during parsing.

### Root Node

```go
type Buildfile struct {
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
    FuncBasename                       // basename()
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

### Design Decisions

1. **BraceExpr remains unresolved during parsing**: In target patterns, `{name}` could be either a capture or a variable interpolation. The parser produces `BraceExpr` nodes; semantic analysis resolves them based on the symbol table.

2. **Separate Statement and Node interfaces**: All top-level constructs implement `Statement`. This allows type-safe iteration over `Buildfile.Statements` with type switches.

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

4. **Exported and unexported methods**: Internal methods are lowercase, with uppercase exported wrappers for CLI/testing access. This maintains encapsulation while enabling debug features.
