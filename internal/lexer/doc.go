// Package lexer implements lexical analysis for Buildfiles.
//
// The lexer tokenizes Buildfile source code into a stream of tokens, handling:
//   - Indentation tracking with consistent space/tab enforcement
//   - Interpolation boundary detection ({var} syntax)
//   - Multiple lexing modes (line start, normal, value, interpolation)
//   - Escape sequences ({{ and }})
//
// # Lexer Modes
//
// The lexer operates in different modes depending on context:
//   - ModeLineStart: At beginning of line, consuming indentation
//   - ModeNormal: Normal token recognition (targets, identifiers, paths)
//   - ModeValue: After = or :, consuming value content as strings
//   - ModeInterp: Inside {} interpolation, lexing identifier and modifier
//
// Mode transitions are managed with a stack to handle nested contexts,
// particularly for interpolations within values.
//
// # Token Types
//
// Token types are categorized as:
//   - Special: EOF, NEWLINE, INDENT, COMMENT, ERROR
//   - Dot Keywords: .shell, .parallel, .default, .include, etc.
//   - Keywords: lazy, if, elif, else, end, ifdef, ifndef, block
//   - Operators: =, :, ==, !=, (, ), ,
//   - Identifiers: IDENTIFIER, AT_IDENTIFIER, PATH, STRING
//   - Interpolation: INTERP_START, INTERP_END, INTERP_MOD, escapes
//   - Functions: shell, glob, basename, dirname, replace
//
// # Indentation Tracking
//
// The IndentTracker enforces consistent indentation across a Buildfile:
//   - First indented line establishes the indent unit (e.g., 4 spaces or 1 tab)
//   - Subsequent indents must be exact multiples of this unit
//   - Mixed tabs and spaces within a single indent string is an error
//
// # Interpolation Boundary Detection
//
// A { character is recognized as interpolation start if and only if:
//   - Preceded by a boundary character (whitespace, :, =, /, quotes, parens, etc.)
//   - Followed by a valid identifier start character (letter or underscore)
//
// This distinguishes interpolations from shell variable syntax ($var) and
// literal braces in commands.
package lexer
