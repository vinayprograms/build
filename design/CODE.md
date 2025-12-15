# Build Tool - Code Design

This document tracks the implementation architecture and design decisions as code is written.

## Package Structure

```
github.com/vinayprograms/build/
├── cmd/
│   └── build/          # CLI entry point
│       ├── main.go
│       ├── main_test.go
│       ├── interfaces.go   # Interface definitions
│       └── adapters.go     # Adapters for internal packages
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
│   ├── parser/         # Syntactic analysis
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
│   │   ├── recipe.go       # Recipe parsing
│   │   ├── recipe_test.go
│   │   ├── scope.go        # Scope types and stack
│   │   ├── scope_test.go
│   │   ├── target.go       # Target parsing
│   │   ├── target_test.go
│   │   ├── variable.go     # Variable parsing
│   │   └── variable_test.go
│   └── semantic/       # Semantic analysis
│       ├── capture.go      # Pass 2: Capture validation
│       ├── capture_test.go
│       ├── collector.go    # Pass 1: Symbol collection
│       ├── collector_test.go
│       ├── symbols.go      # Symbol table implementation
│       └── symbols_test.go
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
| `BuildfileResult` | Contains parsed statements and collected errors |
| `BuildfileParser` | Parses complete buildfiles with error recovery |

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
| `--debug-var` | Dump variable parsing (shows parsed variables) |
| `--debug-target` | Dump target parsing (shows parsed targets) |
| `--debug-recipe` | Dump recipe parsing (shows parsed recipes with commands) |
| `--debug-env` | Dump environment parsing (shows parsed environment blocks) |
| `--debug-cond` | Dump conditional parsing (shows parsed conditionals) |
| `--debug-include` | Dump include parsing (shows included files and statements) |
| `--debug-ast` | Dump full AST with error recovery (shows all parsed statements and errors) |
| `--debug-semantic` | Dump semantic analysis (shows symbol table) |

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
| Operators | `EQUALS`, `COLON`, `DOUBLE_EQUALS`, `NOT_EQUALS`, `LPAREN`, `RPAREN`, `COMMA` | Punctuation |
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
| `IsTargetLine() bool` | Checks if current line is a target (`: before =`) |

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

### Recipe Parsing (`recipe.go`)

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
.include: ./common.build
.include: {config_dir}/settings.build
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
# In /project/build/Buildfile:
.include: ./common.build    # → /project/build/common.build
.include: ../lib/deps.build # → /project/lib/deps.build
.include: /etc/defaults.build # → /etc/defaults.build (absolute)
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

2. **Statements returned separately**: `ParseInclude()` returns both the directive node and the parsed statements. This allows the caller to decide how to merge the statements into the parent AST.

3. **Recursive with stack**: The include stack is passed through recursive calls to detect circular includes at any depth.

4. **Comments preserved**: Comments in included files are preserved in the returned statements.

5. **No indentation in included files**: Included files are parsed at global scope. Indented content in included files follows the same rules as the main file.

### Error Recovery (`parser.go` and `recovery_test.go`)

Implements error recovery to collect multiple parse errors and continue parsing after errors.

**Key Functions:**

| Function | Description |
|----------|-------------|
| `ParseBuildfile() ([]Statement, *ParseErrors)` | Parses complete buildfile with error recovery |
| `parseTopLevelStatement() (Statement, *ParseError)` | Parses a single top-level statement |
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
Buildfile:1:1: directive '.after' invalid at GLOBAL scope (hint: .after is only valid in: RECIPE)
```

**Design Decisions:**

1. **Skip to level 0**: Recovery skips to indentation level 0 to ensure we're back at global scope. This avoids confusion with partially parsed blocks.

2. **Max error limit**: After 10 errors, parsing stops to avoid runaway error cascades on severely broken input.

3. **Preserve valid statements**: All successfully parsed statements are returned even if errors occurred. This enables partial analysis of broken files.

4. **Scope error hints**: Directive scope errors include hints listing valid scopes, helping users understand where directives belong.

5. **Indented line skipping**: During recovery, any indented lines (part of a block) are skipped until we reach a non-indented line.

## Semantic Package (`internal/semantic`)

The semantic package provides semantic analysis for Buildfiles. It validates the AST produced by the parser and resolves context-sensitive constructs.

### Symbol Table (`symbols.go`)

The symbol table tracks all defined symbols in a Buildfile for semantic validation.

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

#### DuplicateDefinitionError

Returned when duplicate definitions are detected:

```go
type DuplicateDefinitionError struct {
    Kind   string             // "variable", "target", or "environment"
    Name   string             // The duplicated name
    First  ast.SourceLocation // First definition location
    Second ast.SourceLocation // Duplicate location
}
```

Error format: `duplicate variable 'cc': first defined at Buildfile:1:1, redefined at Buildfile:5:1`

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
| `parser_integration_test.go` | Full `ParseBuildfile` integration tests |
| `edge_cases_test.go` | Edge cases and negative tests |

### Integration Tests (`parser_integration_test.go`)

Tests `ParseBuildfile()` end-to-end parsing:

| Test | Description |
|------|-------------|
| `TestParseBuildfile_AllStatementTypes` | Parses buildfile with all statement types |
| `TestParseBuildfile_DirectiveDetails` | Verifies directive parsing details |
| `TestParseBuildfile_VariableDetails` | Verifies variable parsing with interpolations |
| `TestParseBuildfile_TargetDetails` | Verifies target and recipe parsing |
| `TestParseBuildfile_EnvironmentDetails` | Verifies environment block parsing |
| `TestParseBuildfile_ConditionalDetails` | Verifies conditional branch parsing |
| `TestParseBuildfile_NestedBlocks` | Tests recipe with block command |
| `TestParseBuildfile_SourceLocations` | Verifies source location tracking |
| `TestParseBuildfile_EmptyFile` | Tests empty file handling |
| `TestParseBuildfile_OnlyComments` | Tests comment-only file |
| `TestParseBuildfile_MixedWithBlankLines` | Tests blank line handling |
| `TestParseBuildfile_ErrorRecoveryIntegration` | Tests error recovery in full buildfile |

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

### CLI Tests (`cmd/build/main_test.go`)

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

Each debug test includes:
- Success case with valid buildfile
- Missing file error case
- (Some) edge cases like empty files or files without target content

### Semantic Unit Tests (`internal/semantic`)

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
| `TestCollector_EmptyBuildfile` | Empty buildfile handling |
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
