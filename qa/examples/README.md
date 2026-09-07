# need Test Suite

This directory contains test cases for the `need` tool. Each `.need` file is a complete, self-contained test case.

## Test Organization

Tests are organized by number ranges:

| Range | Category | Description |
|-------|----------|-------------|
| 001-009 | Minimal/Edge Cases | Minimal needfiles, empty, single elements |
| 010-019 | Directives | Global and recipe-level directives |
| 020-029 | Variables | Variable assignment, interpolation, modifiers |
| 030-039 | Functions | Built-in functions (shell, glob, etc.) |
| 040-049 | Targets | Phony, file, directory, pattern targets |
| 050-059 | Recipes | Commands, blocks, automatic variables |
| 060-069 | Conditionals | if/elif/else, ifdef/ifndef |
| 070-079 | Environments | Environment definitions and runtimes |
| 100-119 | Error Cases | Invalid needfiles that should produce errors |
| 200-209 | Complex/Integration | Realistic, complex needfiles |
| 210-219 | Edge Cases | Unusual but valid inputs |

## File Format

Each test file includes a header comment with:
- **Test**: Brief description
- **Expected**: Expected behavior
- **Category**: Test category
- **Description**: Detailed description

## Running Tests

To parse/validate all test files:
```bash
for f in qa/examples/*.need; do
    echo "=== $f ==="
    need --debug-ast -f "$f" 2>&1 | head -20
done
```

To check that valid files parse successfully:
```bash
for f in qa/examples/0*.need qa/examples/2*.need; do
    need -n -f "$f" 2>&1 || echo "FAIL: $f"
done
```

To verify error cases detect issues:
```bash
for f in qa/examples/1*.need; do
    if need -n -f "$f" 2>&1 >/dev/null; then
        echo "UNEXPECTED SUCCESS: $f"
    else
        echo "OK (expected failure): $f"
    fi
done
```

## Test Categories

### Minimal/Edge Cases (001-009)
- Empty needfiles
- Minimal targets
- Comments only

### Directives (010-019)
- `.shell:` at global and recipe level
- `.parallel:` values
- `.default:` specification
- Recipe directives (`.after:`, `.autodeps:`, `.requires:`)

### Variables (020-029)
- Immediate assignment (`var = value`)
- Lazy assignment (`lazy var = value`)
- Nested interpolation
- Special characters in values
- Built-in variables (`{os}`, `{arch}`)
- Raw modifier (`{var:raw}`)
- Escaped braces (`{{`, `}}`)

### Functions (030-039)
- `shell()` command execution
- `glob()` pattern matching
- `basename()` path extraction
- `dirname()` directory extraction
- `replace()` string replacement
- Nested function calls
- Interpolation in function arguments

### Targets (040-049)
- Phony targets (`@name`)
- File targets (`path/file`)
- Directory targets (`path/`)
- Pattern targets (`{name}`)
- Multiple captures
- Variable interpolation in patterns
- Target with no dependencies
- Many dependencies
- Phony-to-phony chains
- Mixed phony and file dependencies

### Recipes (050-059)
- Automatic variables (`{target}`, `{deps}`, `{in}`, etc.)
- Line mode (each line separate shell)
- Block mode (multi-line script)
- Shell override
- Order-only dependencies
- Auto-dependencies
- Requirements

### Conditionals (060-069)
- Simple if-end
- if-else-end
- if-elif-else-end
- Not-equals comparison
- ifdef/ifndef
- Nested conditionals
- Variable comparisons
- Targets inside conditionals

### Environments (070-079)
- Default (unnamed) environment
- Named environments
- All runtime types (bare, docker, podman, devcontainer, nix, lima)
- Version specifications
- Interpolation in source paths

### Error Cases (100-119)
- Mixed indentation
- Inconsistent indent width
- Wrong directive scope
- Circular dependencies
- Undefined variables
- Duplicate definitions
- Missing conditional end
- Automatic variable misuse
- Capture mismatches
- Syntax errors

### Complex/Integration (200-209)
- Realistic C project
- Multi-environment project
- Deep dependency chains
- Cross-platform conditionals

### Edge Cases (210-219)
- Very long values
- Unicode content
- Many targets
- Many variables
- Phony names with hyphens
- Shell variables vs recipe variables
- Dot paths and hidden files
- Brace boundary rules
- Multiple blank lines
- Trailing whitespace

## Coverage Goals

This test suite aims to cover:

1. **Lexical Analysis**
   - All token types
   - Indentation tracking
   - Interpolation boundary detection
   - Escape sequences

2. **Parsing**
   - All statement types
   - Scope validation
   - Error recovery

3. **Semantic Analysis**
   - Symbol collection
   - Capture validation
   - Reference resolution
   - Dependency graph validation

4. **Evaluation**
   - Variable expansion
   - Function execution
   - Conditional evaluation
   - Lazy variable deferral

5. **Planning**
   - Target matching
   - Dependency ordering
   - Staleness detection

6. **Execution**
   - Command interpolation
   - Shell invocation
   - Block mode
   - Automatic variables

## Adding New Tests

When adding tests:
1. Choose appropriate number range
2. Use descriptive filename
3. Include header comment with Test/Expected/Category/Description
4. Keep tests focused on one feature
5. Update this README if adding new categories
