# Build Tool - Code Design

This document tracks the implementation architecture and design decisions as code is written.

## Package Structure

```
github.com/vinay/build/
├── internal/
│   └── lexer/          # Lexical analysis
│       ├── indent.go       # Indentation tracking
│       ├── indent_test.go
│       ├── interp.go       # Interpolation boundary detection
│       ├── interp_test.go
│       ├── token.go        # Token types and source location
│       └── token_test.go
└── go.mod
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

Per spec, `{` is recognized as interpolation start if and only if:
1. Preceded by whitespace (space or tab), start-of-line, `:`, or `=`
2. Followed by a valid identifier start character (letter or underscore)

| Input | Result | Reason |
|-------|--------|--------|
| `{var}` | Valid interpolation | SOL + valid identifier |
| `x {var}` | Valid interpolation | Space boundary |
| `${var}` | Not interpolation | `$` is not a boundary |
| `x{var}` | Not interpolation | No boundary before `{` |
| `{"key"}` | Not interpolation | `"` is not identifier start |
| `{{var}}` | Escaped brace | `{{` is escape sequence |
| `{var:raw}` | Valid with modifier | `:raw` modifier parsed |

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
