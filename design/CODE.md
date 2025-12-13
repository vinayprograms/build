# Build Tool - Code Design

This document tracks the implementation architecture and design decisions as code is written.

## Package Structure

```
github.com/vinay/build/
├── internal/
│   └── lexer/          # Lexical analysis
│       ├── indent.go       # Indentation tracking
│       ├── indent_test.go
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
