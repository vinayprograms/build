# Build Tool - Detailed Design

## 1. Overview

This document describes the detailed design for implementing the `build` tool, a Make-inspired build system with a readable syntax optimized for user simplicity.

### 1.1 Design Priorities

1. **User-friendly syntax** - Minimal punctuation, consistent `{var}` syntax for all substitutions
2. **Deterministic parsing** - Every construct has exactly one interpretation
3. **Clear separation of phases** - Lex → Parse → Analyze → Plan → Execute
4. **Simple data structures** - Minimal AST, no unnecessary abstractions

### 1.2 Context Sensitivity Acknowledgment

The Buildfile language is **mostly context-free** but has **intentional context-sensitive elements** to keep the user-facing syntax simple:

| Construct | Context Required | Rationale |
|-----------|------------------|-----------|
| `{name}` in target patterns | Semantic: Is `name` a defined variable? | Unified syntax for captures and interpolation |
| Directive scope (`.shell:`) | Syntactic: Indentation level + parent block | Same keyword at global and recipe level |
| Automatic variables (`{target}`, `{deps}`) | Semantic: Only valid inside recipes | Natural naming, no special prefix |

**Trade-off**: We accept these context-sensitive cases because:
- They make the language simpler for users (one syntax to learn)
- The context is always locally determinable (indentation, symbol table)
- Errors are caught early with clear messages

## 2. Lexical Design

### 2.1 Design Goal: Context-Free Tokenization

The lexer produces tokens without needing semantic information. Token recognition is based purely on character patterns and local context (previous character, whitespace boundaries).

**What IS context-free at lexical level:**
- Interpolation recognition (`{` boundary rules)
- Directive keywords (`.shell:`, `.parallel:`, etc.)
- Operators (`=`, `:`)
- Comments (`#`)
- Indentation tracking

**What requires parse/semantic context:**
- Whether `{name}` is a capture or variable reference (determined during semantic analysis)
- Which directives are valid at current scope (determined during parsing)

### 2.2 Token Types

```
TOKEN_EOF           // End of file
TOKEN_NEWLINE       // End of line
TOKEN_INDENT        // Indentation (spaces/tabs at line start)
TOKEN_COMMENT       // # to end of line

TOKEN_DOT_KEYWORD   // .shell, .parallel, .default, .include, .environment,
                    // .using, .source, .args, .requires, .after, .autodeps

TOKEN_KEYWORD       // lazy, if, elif, else, end, ifdef, ifndef, block

TOKEN_EQUALS        // =
TOKEN_COLON         // :

TOKEN_IDENTIFIER    // alphanumeric + underscore, not starting with digit
TOKEN_AT_IDENTIFIER // @name (phony target)
TOKEN_PATH          // file path characters (alphanumeric, /, ., -, _)

TOKEN_INTERP_START  // { preceded by whitespace/SOL
TOKEN_INTERP_END    // }
TOKEN_INTERP_MOD    // :raw (modifier inside interpolation)
TOKEN_ESCAPE_BRACE  // {{ or }}

TOKEN_LPAREN        // (
TOKEN_RPAREN        // )

TOKEN_STRING        // Remaining text on line (for values, commands)
```

### 2.3 Lexer Rules

#### 2.3.1 Line Structure

Each line is classified by its first non-whitespace content:

| First Token | Line Type |
|-------------|-----------|
| `TOKEN_DOT_KEYWORD` | Directive line |
| `TOKEN_KEYWORD` (lazy, if, elif, else, end, ifdef, ifndef) | Control line |
| `TOKEN_AT_IDENTIFIER` followed by `:` | Phony target line |
| `TOKEN_PATH` followed by `:` | Target line |
| `TOKEN_IDENTIFIER` followed by `=` | Variable line |
| `TOKEN_INDENT` + any above | Indented variant (recipe context) |
| `#` | Comment line |
| Empty | Blank line |

#### 2.3.2 Interpolation Recognition

**Rule**: `{` is recognized as `TOKEN_INTERP_START` if and only if:
1. It is preceded by a boundary character (see below)
2. It is followed by a valid identifier character (letter or underscore)

**Boundary characters**:
- Whitespace: space, tab
- Start-of-line
- Operators: `:`, `=`
- Path separator: `/`
- Quotes: `"`, `'`
- Parentheses and comma: `(`, `)`, `,`
- Redirections: `>`, `<`

**Examples**:

| Input | Tokens | Reason |
|-------|--------|--------|
| `{var}` | INTERP_START, IDENTIFIER, INTERP_END | SOL + valid identifier |
| `x {var}` | STRING("x "), INTERP_START, IDENTIFIER, INTERP_END | Space boundary |
| `a/{var}` | STRING("a/"), INTERP_START, IDENTIFIER, INTERP_END | `/` boundary |
| `"{var}"` | STRING(`"`), INTERP_START, IDENTIFIER, INTERP_END, STRING(`"`) | `"` boundary |
| `shell({var})` | ..., INTERP_START, IDENTIFIER, INTERP_END, ... | `(` boundary |
| `>{file}` | STRING(">"), INTERP_START, IDENTIFIER, INTERP_END | `>` boundary |
| `${var}` | STRING("${var}") | `$` precedes `{`, not a boundary |
| `x{var}` | STRING("x{var}") | Letter precedes `{`, not a boundary |
| `{"key"}` | STRING(`{"key"}`) | `"` is not valid identifier start |
| `{{var}}` | ESCAPE_BRACE, STRING("var"), ESCAPE_BRACE | Escape sequence |
| `{var:raw}` | INTERP_START, IDENTIFIER, INTERP_MOD, INTERP_END | Modifier syntax |

#### 2.3.3 Indentation Tracking

The lexer tracks indentation as a separate concern:

```
struct IndentInfo {
    width: int           // Number of characters
    char: ' ' | '\t'     // Character used (first non-empty line sets this)
    level: int           // Logical level (0, 1, 2)
}
```

**Rules**:
1. First indented line establishes the indent unit (e.g., 4 spaces)
2. Subsequent indents must be multiples of this unit
3. Mixed tabs/spaces after first line is an error
4. Level 0 = column 0, Level 1 = one unit, Level 2 = two units

#### 2.3.4 Directive Keywords

All directives start with `.` followed by a specific keyword:

| Keyword | Valid at Level 0 | Valid at Level 1 |
|---------|------------------|------------------|
| `.shell` | ✓ (global default) | ✓ (recipe override) |
| `.parallel` | ✓ | ✗ |
| `.default` | ✓ | ✗ |
| `.include` | ✓ | ✗ |
| `.environment` | ✓ | ✗ |
| `.using` | ✗ | ✓ (inside .environment) |
| `.source` | ✗ | ✓ (inside .environment) |
| `.args` | ✗ | ✓ (inside .environment) |
| `.requires` | ✗ | ✓ (inside .environment or recipe) |
| `.after` | ✗ | ✓ (recipe only) |
| `.autodeps` | ✗ | ✓ (recipe only) |

### 2.4 Lexer State Machine

The lexer is **stateless** between lines but maintains minimal state within a line:

```
enum LexerMode {
    LINE_START,      // Beginning of line, consume indentation
    NORMAL,          // Normal token recognition
    INTERPOLATION,   // Inside { }, looking for identifier and modifiers
    STRING_VALUE,    // Consuming rest of line as string
}
```

**Key transitions**:
- `LINE_START` → consume whitespace → emit `TOKEN_INDENT` → `NORMAL`
- `NORMAL` + see `=` or `:` → emit token → `STRING_VALUE` (for rest of line)
- `NORMAL` + see `{` (with boundary) → emit `TOKEN_INTERP_START` → `INTERPOLATION`
- `INTERPOLATION` + see `}` → emit `TOKEN_INTERP_END` → `NORMAL`

---

## 3. Syntactic Design

### 3.1 Design Goal: Deterministic Parsing

The parser produces a single, unambiguous AST for any valid input. Where the grammar appears ambiguous, explicit disambiguation rules resolve it.

**Context maintained during parsing:**
- Indentation level (for recipe/block detection)
- Scope stack (GLOBAL, ENVIRONMENT, RECIPE, BLOCK)

**Deferred to semantic analysis:**
- Capture vs variable interpolation distinction
- Automatic variable validation
- Reference resolution

### 3.2 Grammar (EBNF)

```ebnf
buildfile       = { statement } EOF ;

statement       = directive
                | environment_block
                | variable_def
                | conditional
                | target_def
                | NEWLINE
                | COMMENT ;

(* Directives *)
directive       = global_directive NEWLINE ;

global_directive = ".shell:" value
                 | ".parallel:" value
                 | ".default:" value
                 | ".include:" value ;

(* Environment blocks *)
environment_block = ".environment:" [ identifier ] NEWLINE
                    INDENT { env_directive NEWLINE }
                    DEDENT ;

env_directive   = ".using:" value
                | ".source:" value
                | ".args:" value
                | ".requires:" value ;

(* Variables *)
variable_def    = [ "lazy" ] identifier "=" value NEWLINE ;

(* Conditionals *)
conditional     = if_clause { elif_clause } [ else_clause ] "end" NEWLINE ;

if_clause       = "if" condition NEWLINE { statement } ;
elif_clause     = "elif" condition NEWLINE { statement } ;
else_clause     = "else" NEWLINE { statement } ;

condition       = interpolation "==" value
                | interpolation "!=" value ;

ifdef_clause    = "ifdef" identifier NEWLINE { statement } "end" NEWLINE ;
ifndef_clause   = "ifndef" identifier NEWLINE { statement } "end" NEWLINE ;

(* Targets *)
target_def      = target_spec ":" dependency_list NEWLINE
                  [ recipe ] ;

target_spec     = phony_target | file_target ;
phony_target    = "@" identifier ;
file_target     = path_pattern ;

path_pattern    = { path_segment | capture } ;
path_segment    = PATH ;
capture         = "{" identifier "}" ;

dependency_list = { dependency } ;
dependency      = path_pattern | interpolation ;

(* Recipes *)
recipe          = INDENT { recipe_line } DEDENT ;

recipe_line     = recipe_directive NEWLINE
                | block_stmt
                | command NEWLINE ;

recipe_directive = ".shell:" value
                 | ".after:" value
                 | ".autodeps:" value
                 | ".requires:" value ;

block_stmt      = "block:" NEWLINE
                  INDENT { raw_line } DEDENT ;

command         = { command_part } ;
command_part    = STRING | interpolation ;

raw_line        = { raw_part } NEWLINE ;
raw_part        = STRING | interpolation ;

(* Common *)
interpolation   = "{" identifier [ ":" modifier ] "}" ;
modifier        = "raw" ;

value           = { value_part } ;
value_part      = STRING | interpolation | function_call ;

function_call   = func_name "(" { value } ")" ;
func_name       = "shell" | "glob" | "basename" | "dirname" | "replace" ;

identifier      = IDENTIFIER ;
```

### 3.3 Disambiguation Rules

The grammar is unambiguous due to these structural rules:

#### 3.3.1 Variable vs Target

**Rule**: A line is a variable definition if and only if:
1. It contains `=`
2. The `=` appears before any `:`

```
cc = gcc              # Variable: = found, no : before it
foo: bar              # Target: : found, no = before it
path = /usr/bin:foo   # Variable: = at position 4, : at position 13
foo:bar = baz         # Variable: = at position 8, : at position 3...
                      # Wait - : comes first. This is ambiguous!
```

**Resolution**: The `=` vs `:` rule uses first occurrence:
- If `=` appears first → Variable
- If `:` appears first → Target

**Edge case** - `foo:bar = baz`:
- `:` at position 3, `=` at position 8
- `:` comes first → This is a **target** with name `foo` and dependency `bar = baz`
- This is likely a user error, but the parse is deterministic
- Semantic analysis can warn about suspicious dependency names

#### 3.3.2 Capture vs Interpolation in Target Patterns

This is the primary context-sensitive aspect of the language. The syntax `{identifier}` means different things depending on context:

| Location | Meaning | Determined By |
|----------|---------|---------------|
| Target pattern (left of `:`) | Capture OR variable interpolation | Semantic analysis |
| Dependency list (right of `:`) | Variable interpolation only | Always interpolation |
| Recipe commands | Variable/automatic interpolation | Always interpolation |
| Variable values | Variable interpolation | Always interpolation |

**Resolution Algorithm (during semantic analysis):**

```
fn resolve_brace_in_target_pattern(identifier, symbol_table):
    if identifier in symbol_table.variables:
        return VariableInterpolation(identifier)
    else if identifier in AUTOMATIC_VARIABLES:
        return Error("Automatic variable '{identifier}' cannot be used in target pattern")
    else:
        return Capture(identifier)
```

**Examples:**

```
build_dir = build
name = foo                    # User-defined variable

# Case 1: {build_dir} is interpolation (variable exists)
# Case 2: {name} would be interpolation (variable exists) - ERROR if used as capture
# Case 3: {stem} is capture (not a variable, not automatic)

{build_dir}/{stem}.o: src/{stem}.c    # {build_dir} interpolates, {stem} captures
    gcc -c {in} -o {out}              # {in}, {out} are automatic variables
```

**Parser behavior:** The parser does NOT distinguish captures from interpolations. It produces a uniform AST node:

```
// Parser output for: build/{name}.o
TargetPattern {
    segments: [
        Literal("build/"),
        BraceExpr { identifier: "name" },  // Not yet classified
        Literal(".o")
    ]
}

// After semantic analysis:
TargetPattern {
    segments: [
        Literal("build/"),
        Capture("name"),                    // Now classified
        Literal(".o")
    ]
}
```

**Error cases:**

1. Using a defined variable name as capture:
   ```
   name = foo
   build/{name}.o: src/{name}.c    # Error: 'name' is a variable, not a capture
   ```

2. Using automatic variable in target pattern:
   ```
   build/{target}.o: src/{target}.c  # Error: 'target' is automatic, cannot capture
   ```

3. Capture mismatch between target and dependencies:
   ```
   build/{a}.o: src/{b}.c           # Error: capture 'a' not used in dependency
   ```

#### 3.3.3 Directive Scope (Context-Sensitive)

Directives starting with `.` have scope-dependent meaning. This is context-sensitive at the **parser** level (not semantic).

**Rule**: Directive scope is determined by:
1. Indentation level
2. Current scope on the parser's scope stack

| Directive | At Level 0 (GLOBAL) | At Level 1 (ENVIRONMENT) | At Level 1 (RECIPE) |
|-----------|---------------------|--------------------------|---------------------|
| `.shell:` | Global default shell | Invalid | Recipe shell override |
| `.parallel:` | Max parallel jobs | Invalid | Invalid |
| `.default:` | Default target | Invalid | Invalid |
| `.include:` | Include file | Invalid | Invalid |
| `.environment:` | Start env block | Invalid | Invalid |
| `.using:` | Invalid | Runtime type | Invalid |
| `.source:` | Invalid | Source path | Invalid |
| `.args:` | Invalid | Runtime args | Invalid |
| `.requires:` | Invalid | Env requirements | Recipe requirements |
| `.after:` | Invalid | Invalid | Order-only deps |
| `.autodeps:` | Invalid | Invalid | Auto-dep file |

**Parser scope tracking:**

```
enum Scope {
    GLOBAL,
    ENVIRONMENT,
    RECIPE,
    BLOCK,
}

struct Parser {
    scope_stack: Vec<Scope>  // Stack because blocks nest inside recipes
}

fn parse_directive(&mut self, directive: &str) -> Result<Directive, ParseError> {
    let current_scope = self.scope_stack.last().unwrap_or(&Scope::GLOBAL);
    
    match (directive, current_scope) {
        (".shell", Scope::GLOBAL) => self.parse_global_shell(),
        (".shell", Scope::RECIPE) => self.parse_recipe_shell(),
        (".shell", _) => Err(ParseError::InvalidDirectiveScope {
            directive: ".shell",
            scope: current_scope,
        }),
        // ... other directives
    }
}
```

**Scope transitions:**

```
GLOBAL
  │
  ├─ .environment: ──▶ ENVIRONMENT
  │                      │
  │                      └─ (dedent) ──▶ GLOBAL
  │
  └─ target: ──────────▶ RECIPE (on indent)
                           │
                           ├─ block: ──▶ BLOCK
                           │              │
                           │              └─ (dedent) ──▶ RECIPE
                           │
                           └─ (dedent) ──▶ GLOBAL
```

#### 3.3.4 Automatic Variables (Context-Sensitive)

Automatic variables (`{target}`, `{deps}`, `{in}`, `{out}`, `{stem}`, `{target.dir}`, `{target.file}`) are only valid inside recipes. This is enforced during **semantic analysis**.

**Reserved automatic variable names:**

```
AUTOMATIC_VARIABLES = {
    "target",      // Target file path
    "deps",        // All dependencies (space-separated)
    "in",          // First dependency
    "out",         // Alias for target
    "stem",        // Pattern match stem (for pattern targets)
    "target.dir",  // Directory part of target
    "target.file", // Filename part of target
}
```

**Validation during semantic analysis:**

```
fn validate_interpolation(name: &str, scope: Scope, location: SourceLocation) -> Result<(), SemanticError> {
    if AUTOMATIC_VARIABLES.contains(name) {
        match scope {
            Scope::RECIPE | Scope::BLOCK => Ok(()),
            _ => Err(SemanticError::AutomaticVariableOutsideRecipe {
                name: name.to_string(),
                location,
                hint: "Automatic variables like {target} are only available inside recipes",
            }),
        }
    } else {
        Ok(())  // User variable, checked elsewhere
    }
}
```

**Example errors:**

```
# Error: {target} used in variable definition
output = {target}                        # Error: automatic variable outside recipe

# Error: {deps} used in target pattern  
{deps}.txt: input.txt                    # Error: automatic variable in target pattern
    echo {deps}                          # OK: inside recipe

# OK: {target} inside recipe
build/app: build/main.o
    echo "Building {target}"             # OK
```

### 3.5 Summary: Context-Sensitive Elements

For implementers, here is the complete list of context-sensitive language features:

| Feature | Context Type | Where Determined | Resolution |
|---------|--------------|------------------|------------|
| `{name}` as capture vs interpolation | Semantic | Semantic analysis | Check if `name` is in symbol table |
| Directive validity (`.shell:` etc.) | Syntactic | Parser scope stack | Check current scope allows directive |
| Automatic variable usage | Semantic | Semantic analysis | Check scope is RECIPE or BLOCK |
| `{name}` capture consistency | Semantic | Semantic analysis | Verify captures match in target and deps |

**Key insight**: The lexer is fully context-free. Context sensitivity enters at:
1. **Parser level**: Scope stack for directive validation
2. **Semantic level**: Symbol table for capture/interpolation distinction

### 3.6 Parser Architecture

```
struct Parser {
    lexer: Lexer
    current: Token
    scope_stack: Vec<Scope>
}

impl Parser {
    fn parse(&mut self) -> Result<Buildfile, ParseError>
    fn parse_statement(&mut self) -> Result<Statement, ParseError>
    fn parse_directive(&mut self) -> Result<Directive, ParseError>
    fn parse_environment(&mut self) -> Result<Environment, ParseError>
    fn parse_variable(&mut self) -> Result<Variable, ParseError>
    fn parse_conditional(&mut self) -> Result<Conditional, ParseError>
    fn parse_target(&mut self) -> Result<Target, ParseError>
    fn parse_recipe(&mut self) -> Result<Recipe, ParseError>
    fn parse_value(&mut self) -> Result<Value, ParseError>
}
```

**Error recovery**: On parse error, skip to next line at indentation level 0.

---

## 4. Abstract Syntax Tree

### 4.1 Design Goal: Minimal, Typed Representation

The AST captures the structure without interpretation. No evaluation happens during parsing.

### 4.2 AST Node Definitions

```
struct Buildfile {
    statements: Vec<Statement>
    source_path: PathBuf
}

enum Statement {
    Directive(Directive),
    Environment(Environment),
    Variable(Variable),
    Conditional(Conditional),
    Target(Target),
    Comment(String),
    Blank,
}

// Directives
struct Directive {
    kind: DirectiveKind,
    value: Value,
    location: SourceLocation,
}

enum DirectiveKind {
    Shell,
    Parallel,
    Default,
    Include,
}

// Environments
struct Environment {
    name: Option<String>,          // None = default environment
    runtime: Option<Runtime>,      // .using
    source: Option<Value>,         // .source
    args: Option<Value>,           // .args
    requires: Vec<Requirement>,    // .requires
    location: SourceLocation,
}

enum Runtime {
    Bare,
    Docker,
    Podman,
    Devcontainer,
    Nix,
    Lima,
}

struct Requirement {
    name: String,
    version: VersionSpec,
}

enum VersionSpec {
    Latest,
    Major(u32),
    MajorMinor(u32, u32),
    Exact(u32, u32, u32),
}

// Variables
struct Variable {
    name: String,
    value: Value,
    lazy: bool,
    location: SourceLocation,
}

// Conditionals
struct Conditional {
    if_branch: ConditionalBranch,
    elif_branches: Vec<ConditionalBranch>,
    else_body: Option<Vec<Statement>>,
    location: SourceLocation,
}

struct ConditionalBranch {
    condition: Condition,
    body: Vec<Statement>,
}

enum Condition {
    Equals(Value, Value),
    NotEquals(Value, Value),
    Defined(String),      // ifdef
    NotDefined(String),   // ifndef
}

// Targets
struct Target {
    pattern: TargetPattern,
    dependencies: Vec<Dependency>,
    recipe: Option<Recipe>,
    location: SourceLocation,
}

struct TargetPattern {
    segments: Vec<PatternSegment>,
    is_phony: bool,
    is_directory: bool,    // ends with /
}

enum PatternSegment {
    Literal(String),
    BraceExpr(String),     // {name} - unresolved during parsing
                           // Semantic analysis resolves to Capture or Interpolation
}

// After semantic analysis, BraceExpr is resolved:
enum ResolvedPatternSegment {
    Literal(String),
    Capture(String),       // {name} where name is not a defined variable
    Interpolation(String), // {var} where var is a defined variable
}

struct Dependency {
    segments: Vec<PatternSegment>,  // Same as target, resolved during semantic analysis
}

// Recipes
struct Recipe {
    directives: RecipeDirectives,
    commands: Vec<Command>,
    location: SourceLocation,
}

struct RecipeDirectives {
    shell: Option<Value>,
    after: Vec<Value>,
    autodeps: Option<Value>,
    requires: Vec<Requirement>,
}

enum Command {
    Line(Vec<CommandPart>),
    Block(Vec<Vec<CommandPart>>),
}

enum CommandPart {
    Literal(String),
    Interpolation(Interpolation),
}

struct Interpolation {
    name: String,
    raw: bool,             // :raw modifier
    location: SourceLocation,
}

// Values (used in directives, variables, conditions)
struct Value {
    parts: Vec<ValuePart>,
    location: SourceLocation,
}

enum ValuePart {
    Literal(String),
    Interpolation(Interpolation),
    FunctionCall(FunctionCall),
}

struct FunctionCall {
    name: FunctionName,
    args: Vec<Value>,
}

enum FunctionName {
    Shell,
    Glob,
    Basename,
    Dirname,
    Replace,
}

// Source tracking
struct SourceLocation {
    file: PathBuf,
    line: u32,
    column: u32,
}
```

---

## 5. Semantic Analysis

### 5.1 Design Goal: Validate Before Execution

Semantic analysis catches errors that syntactic parsing cannot, ensuring the build specification is valid before any execution.

### 5.2 Analysis Passes

#### Pass 1: Symbol Collection

Collect all definitions without evaluation:

```
struct SymbolTable {
    variables: HashMap<String, Variable>,
    targets: Vec<Target>,
    environments: HashMap<Option<String>, Environment>,
    
    // Automatic variables (built-in, not user-defined)
    automatic: HashSet<String>,  // {target, deps, in, out, stem, ...}
}
```

**Checks**:
- Duplicate variable definitions → Error
- Duplicate target patterns → Error (or warning for overrides)
- Duplicate environment names → Error

#### Pass 2: Capture Validation

For each target pattern containing `{name}`:

```
fn validate_captures(target: &Target, symbols: &SymbolTable) -> Result<(), SemanticError> {
    for segment in &target.pattern.segments {
        if let PatternSegment::Capture(name) = segment {
            // Rule: Captures must not shadow variables
            if symbols.variables.contains_key(name) {
                return Err(SemanticError::CaptureConflictsWithVariable {
                    capture: name.clone(),
                    variable_location: symbols.variables[name].location,
                    capture_location: target.location,
                });
            }
            
            // Rule: Captures must not shadow automatic variables
            if symbols.automatic.contains(name) {
                return Err(SemanticError::CaptureConflictsWithAutomatic {
                    capture: name.clone(),
                    location: target.location,
                });
            }
        }
    }
    
    // Rule: Same captures must appear in dependencies
    let target_captures = extract_captures(&target.pattern);
    for dep in &target.dependencies {
        let dep_captures = extract_captures(&dep.pattern);
        if dep_captures != target_captures {
            return Err(SemanticError::CaptureMismatch {
                target_captures,
                dep_captures,
                location: target.location,
            });
        }
    }
    
    Ok(())
}
```

#### Pass 3: Reference Validation

Verify all interpolations reference defined symbols:

```
fn validate_references(value: &Value, symbols: &SymbolTable, scope: &Scope) -> Result<(), SemanticError> {
    for part in &value.parts {
        if let ValuePart::Interpolation(interp) = part {
            let name = &interp.name;
            
            // Check: Is it an automatic variable?
            if symbols.automatic.contains(name) {
                // Valid only in recipe scope
                if !matches!(scope, Scope::Recipe | Scope::Block) {
                    return Err(SemanticError::AutomaticOutsideRecipe {
                        name: name.clone(),
                        location: interp.location,
                    });
                }
                continue;
            }
            
            // Check: Is it a capture in scope?
            if let Scope::Recipe { captures } = scope {
                if captures.contains(name) {
                    continue;
                }
            }
            
            // Check: Is it a user-defined variable?
            if symbols.variables.contains_key(name) {
                continue;
            }
            
            // Check: Is it a built-in variable?
            if is_builtin_variable(name) {  // os, arch
                continue;
            }
            
            return Err(SemanticError::UndefinedVariable {
                name: name.clone(),
                location: interp.location,
            });
        }
    }
    Ok(())
}
```

#### Pass 4: Dependency Graph Validation

Build and validate the target dependency graph:

```
fn validate_dependency_graph(targets: &[Target]) -> Result<DependencyGraph, SemanticError> {
    let mut graph = DependencyGraph::new();
    
    for target in targets {
        graph.add_node(target.pattern.clone());
        for dep in &target.dependencies {
            graph.add_edge(target.pattern.clone(), dep.pattern.clone());
        }
    }
    
    // Check for cycles
    if let Some(cycle) = graph.find_cycle() {
        return Err(SemanticError::CircularDependency { cycle });
    }
    
    Ok(graph)
}
```

### 5.3 Semantic Error Types

```
enum SemanticError {
    // Symbol errors
    DuplicateVariable { name: String, first: SourceLocation, second: SourceLocation },
    DuplicateTarget { pattern: String, first: SourceLocation, second: SourceLocation },
    DuplicateEnvironment { name: String, first: SourceLocation, second: SourceLocation },
    
    // Capture errors
    CaptureConflictsWithVariable { capture: String, variable_location: SourceLocation, capture_location: SourceLocation },
    CaptureConflictsWithAutomatic { capture: String, location: SourceLocation },
    CaptureMismatch { target_captures: Vec<String>, dep_captures: Vec<String>, location: SourceLocation },
    
    // Reference errors
    UndefinedVariable { name: String, location: SourceLocation },
    AutomaticOutsideRecipe { name: String, location: SourceLocation },
    
    // Dependency errors
    CircularDependency { cycle: Vec<String> },
    
    // Directive errors
    InvalidDirectiveScope { directive: String, expected: Scope, found: Scope, location: SourceLocation },
    MissingRequiredDirective { directive: String, context: String, location: SourceLocation },
}
```

---

## 6. Execution Model

### 6.1 Phase Overview

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│    Lex      │────▶│    Parse    │────▶│   Analyze   │────▶│    Plan     │────▶│   Execute   │
│             │     │             │     │             │     │             │     │             │
│ Source text │     │ Token stream│     │ Valid AST   │     │ Build plan  │     │ Run recipes │
│ → Tokens    │     │ → AST       │     │ + symbols   │     │ (DAG order) │     │             │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
```

### 6.2 Evaluation Phase

After semantic analysis, evaluate all non-lazy variables:

```
struct EvaluationContext {
    variables: HashMap<String, String>,  // Evaluated values
    lazy_variables: HashMap<String, Value>,  // Unevaluated AST
    os: String,
    arch: String,
}

fn evaluate_variables(ast: &Buildfile, symbols: &SymbolTable) -> Result<EvaluationContext, EvalError> {
    let mut ctx = EvaluationContext::new();
    
    // Process conditionals first (they may define variables)
    for stmt in &ast.statements {
        if let Statement::Conditional(cond) = stmt {
            evaluate_conditional(cond, &mut ctx)?;
        }
    }
    
    // Evaluate immediate variables in order
    for stmt in &ast.statements {
        if let Statement::Variable(var) = stmt {
            if var.lazy {
                ctx.lazy_variables.insert(var.name.clone(), var.value.clone());
            } else {
                let value = evaluate_value(&var.value, &ctx)?;
                ctx.variables.insert(var.name.clone(), value);
            }
        }
    }
    
    Ok(ctx)
}

fn evaluate_value(value: &Value, ctx: &EvaluationContext) -> Result<String, EvalError> {
    let mut result = String::new();
    
    for part in &value.parts {
        match part {
            ValuePart::Literal(s) => result.push_str(s),
            
            ValuePart::Interpolation(interp) => {
                let val = resolve_variable(&interp.name, ctx)?;
                result.push_str(&val);
            }
            
            ValuePart::FunctionCall(call) => {
                let val = evaluate_function(call, ctx)?;
                result.push_str(&val);
            }
        }
    }
    
    Ok(result)
}

fn evaluate_function(call: &FunctionCall, ctx: &EvaluationContext) -> Result<String, EvalError> {
    match call.name {
        FunctionName::Shell => {
            let cmd = evaluate_value(&call.args[0], ctx)?;
            execute_shell_capture(&cmd)
        }
        FunctionName::Glob => {
            let pattern = evaluate_value(&call.args[0], ctx)?;
            glob_files(&pattern)
        }
        FunctionName::Basename => {
            let path = evaluate_value(&call.args[0], ctx)?;
            Ok(PathBuf::from(path).file_name().unwrap().to_string_lossy().into())
        }
        FunctionName::Dirname => {
            let path = evaluate_value(&call.args[0], ctx)?;
            Ok(PathBuf::from(path).parent().unwrap().to_string_lossy().into())
        }
        FunctionName::Replace => {
            let input = evaluate_value(&call.args[0], ctx)?;
            let from = evaluate_value(&call.args[1], ctx)?;
            let to = evaluate_value(&call.args[2], ctx)?;
            Ok(input.replace(&from, &to))
        }
    }
}
```

### 6.3 Planning Phase

Build an execution plan based on requested targets and file timestamps:

```
struct BuildPlan {
    tasks: Vec<BuildTask>,  // Topologically sorted
}

struct BuildTask {
    target: ConcreteTarget,
    dependencies: Vec<ConcreteTarget>,
    recipe: Option<Recipe>,
    reason: BuildReason,
}

enum BuildReason {
    TargetMissing,
    DependencyNewer { dep: PathBuf, dep_mtime: SystemTime, target_mtime: SystemTime },
    PhonyTarget,
    ForcedRebuild,
}

struct ConcreteTarget {
    path: PathBuf,
    pattern_source: Option<TargetPattern>,  // If from pattern
    captures: HashMap<String, String>,       // Resolved captures
}

fn plan_build(
    requested: &[String],
    targets: &[Target],
    ctx: &EvaluationContext,
) -> Result<BuildPlan, PlanError> {
    let mut plan = BuildPlan::new();
    let mut visited = HashSet::new();
    
    for req in requested {
        plan_target(req, targets, ctx, &mut plan, &mut visited)?;
    }
    
    Ok(plan)
}

fn plan_target(
    target_path: &str,
    targets: &[Target],
    ctx: &EvaluationContext,
    plan: &mut BuildPlan,
    visited: &mut HashSet<String>,
) -> Result<(), PlanError> {
    if visited.contains(target_path) {
        return Ok(());
    }
    visited.insert(target_path.to_string());
    
    // Find matching target definition
    let (target_def, captures) = match_target(target_path, targets)?;
    
    // Plan dependencies first
    for dep in &target_def.dependencies {
        let dep_path = resolve_pattern(&dep.pattern, &captures, ctx)?;
        plan_target(&dep_path, targets, ctx, plan, visited)?;
    }
    
    // Check if rebuild needed
    if let Some(reason) = needs_rebuild(target_path, &target_def, ctx)? {
        plan.tasks.push(BuildTask {
            target: ConcreteTarget {
                path: PathBuf::from(target_path),
                pattern_source: Some(target_def.pattern.clone()),
                captures,
            },
            dependencies: resolve_dependencies(&target_def.dependencies, &captures, ctx)?,
            recipe: target_def.recipe.clone(),
            reason,
        });
    }
    
    Ok(())
}
```

### 6.4 Execution Phase

Execute the build plan, respecting parallelism:

```
struct Executor {
    shell: String,
    parallel: usize,
    dry_run: bool,
    verbose: bool,
}

impl Executor {
    fn execute(&self, plan: BuildPlan, ctx: &EvaluationContext) -> Result<(), ExecError> {
        if self.parallel == 1 {
            self.execute_sequential(plan, ctx)
        } else {
            self.execute_parallel(plan, ctx)
        }
    }
    
    fn execute_sequential(&self, plan: BuildPlan, ctx: &EvaluationContext) -> Result<(), ExecError> {
        for task in plan.tasks {
            self.execute_task(&task, ctx)?;
        }
        Ok(())
    }
    
    fn execute_task(&self, task: &BuildTask, ctx: &EvaluationContext) -> Result<(), ExecError> {
        let Some(recipe) = &task.recipe else {
            return Ok(());  // No recipe, nothing to execute
        };
        
        // Build recipe context with automatic variables
        let recipe_ctx = RecipeContext {
            base: ctx.clone(),
            target: task.target.path.to_string_lossy().into(),
            deps: task.dependencies.iter()
                .map(|d| d.path.to_string_lossy().into())
                .collect::<Vec<_>>()
                .join(" "),
            in_: task.dependencies.first()
                .map(|d| d.path.to_string_lossy().into())
                .unwrap_or_default(),
            out: task.target.path.to_string_lossy().into(),
            stem: task.target.captures.get("name").cloned().unwrap_or_default(),
            captures: task.target.captures.clone(),
        };
        
        // Determine shell
        let shell = recipe.directives.shell
            .as_ref()
            .map(|v| evaluate_value(v, ctx))
            .transpose()?
            .unwrap_or_else(|| self.shell.clone());
        
        // Execute commands
        for cmd in &recipe.commands {
            match cmd {
                Command::Line(parts) => {
                    let line = interpolate_command(parts, &recipe_ctx)?;
                    self.run_shell_line(&shell, &line)?;
                }
                Command::Block(lines) => {
                    let script = lines.iter()
                        .map(|parts| interpolate_command(parts, &recipe_ctx))
                        .collect::<Result<Vec<_>, _>>()?
                        .join("\n");
                    self.run_shell_script(&shell, &script)?;
                }
            }
        }
        
        Ok(())
    }
    
    fn run_shell_line(&self, shell: &str, line: &str) -> Result<(), ExecError> {
        if self.dry_run {
            println!("  {}", line);
            return Ok(());
        }
        
        if self.verbose {
            println!("  {}", line);
        }
        
        let status = std::process::Command::new(shell)
            .arg("-c")
            .arg(line)
            .status()?;
        
        if !status.success() {
            return Err(ExecError::CommandFailed { line: line.into(), status });
        }
        
        Ok(())
    }
}
```

### 6.5 Interpolation in Commands

```
fn interpolate_command(parts: &[CommandPart], ctx: &RecipeContext) -> Result<String, ExecError> {
    let mut result = String::new();
    
    for part in parts {
        match part {
            CommandPart::Literal(s) => result.push_str(s),
            
            CommandPart::Interpolation(interp) => {
                let value = resolve_recipe_variable(&interp.name, ctx)?;
                
                if interp.raw {
                    // No quoting, allows word splitting
                    result.push_str(&value);
                } else {
                    // Shell-quote the value
                    result.push_str(&shell_quote(&value));
                }
            }
        }
    }
    
    Ok(result)
}

fn resolve_recipe_variable(name: &str, ctx: &RecipeContext) -> Result<String, ExecError> {
    // Check automatic variables first
    match name {
        "target" | "out" => return Ok(ctx.target.clone()),
        "deps" => return Ok(ctx.deps.clone()),
        "in" => return Ok(ctx.in_.clone()),
        "stem" => return Ok(ctx.stem.clone()),
        "target.dir" => return Ok(PathBuf::from(&ctx.target).parent().unwrap().to_string_lossy().into()),
        "target.file" => return Ok(PathBuf::from(&ctx.target).file_name().unwrap().to_string_lossy().into()),
        _ => {}
    }
    
    // Check captures
    if let Some(value) = ctx.captures.get(name) {
        return Ok(value.clone());
    }
    
    // Check user variables
    if let Some(value) = ctx.base.variables.get(name) {
        return Ok(value.clone());
    }
    
    // Check lazy variables (evaluate on demand)
    if let Some(ast) = ctx.base.lazy_variables.get(name) {
        return evaluate_value(ast, &ctx.base);
    }
    
    // Check built-ins
    match name {
        "os" => return Ok(ctx.base.os.clone()),
        "arch" => return Ok(ctx.base.arch.clone()),
        _ => {}
    }
    
    Err(ExecError::UndefinedVariable { name: name.into() })
}

fn shell_quote(s: &str) -> String {
    // Use single quotes, escape embedded single quotes
    format!("'{}'", s.replace('\'', "'\\''"))
}
```

---

## 7. Error Reporting

### 7.1 Design Goal: Actionable Error Messages

Every error message should:
1. Identify the exact location (file:line:column)
2. Explain what went wrong
3. Suggest how to fix it

### 7.2 Error Format

```
error[E001]: capture '{name}' conflicts with variable 'name'
  --> Buildfile:15:7
   |
15 | build/{name}.o: src/{name}.c
   |       ^^^^^^ capture defined here
   |
note: variable 'name' was defined here
  --> Buildfile:3:1
   |
 3 | name = foo
   | ^^^^^^^^^^
   |
help: rename either the capture or the variable to avoid conflict
```

### 7.3 Error Categories

| Code | Category | Example |
|------|----------|---------|
| E001-E099 | Lexical errors | Invalid character, bad indentation |
| E100-E199 | Syntax errors | Unexpected token, missing colon |
| E200-E299 | Semantic errors | Undefined variable, circular dependency |
| E300-E399 | Evaluation errors | Shell command failed, glob no matches |
| E400-E499 | Execution errors | Recipe failed, missing file |

---

## 8. Data Flow Summary

```
                                    ┌─────────────────┐
                                    │   Source File   │
                                    │   (Buildfile)   │
                                    └────────┬────────┘
                                             │
                                             ▼
                              ┌──────────────────────────┐
                              │         LEXER            │
                              │     (Context-Free)       │
                              │                          │
                              │ • Line-oriented scanning │
                              │ • Indentation tracking   │
                              │ • Interpolation boundary │
                              │ • {name} as generic token│
                              └────────────┬─────────────┘
                                           │
                                           ▼
                              ┌──────────────────────────┐
                              │         PARSER           │
                              │  (Scope Stack Context)   │
                              │                          │
                              │ • Scope-aware directives │
                              │ • {name} → BraceExpr AST │
                              │ • AST construction       │
                              └────────────┬─────────────┘
                                           │
                                           ▼
                              ┌──────────────────────────┐
                              │    SEMANTIC ANALYZER     │
                              │   (Symbol Table Context) │
                              │                          │
                              │ • BraceExpr → Capture or │
                              │   Interpolation          │
                              │ • Automatic var checking │
                              │ • Reference validation   │
                              │ • Graph validation       │
                              └────────────┬─────────────┘
                                           │
                                           ▼
                              ┌──────────────────────────┐
                              │       EVALUATOR          │
                              │                          │
                              │ • Conditional evaluation │
                              │ • Immediate variables    │
                              │ • Function calls         │
                              └────────────┬─────────────┘
                                           │
                                           ▼
                              ┌──────────────────────────┐
                              │        PLANNER           │
                              │                          │
                              │ • Target resolution      │
                              │ • Dependency ordering    │
                              │ • Staleness checking     │
                              └────────────┬─────────────┘
                                           │
                                           ▼
                              ┌──────────────────────────┐
                              │        EXECUTOR          │
                              │                          │
                              │ • Recipe interpolation   │
                              │ • Shell invocation       │
                              │ • Parallel execution     │
                              └──────────────────────────┘
```

---

## 9. Appendix: Formal Grammar

### 9.1 Lexical Grammar

```
// Whitespace and structure
NEWLINE     = '\n' | '\r\n'
INDENT      = (SPACE | TAB)+   // At line start only
SPACE       = ' '
TAB         = '\t'

// Comments
COMMENT     = '#' [^\n]*

// Identifiers
IDENTIFIER  = [a-zA-Z_][a-zA-Z0-9_]*
AT_IDENT    = '@' IDENTIFIER

// Literals
PATH        = [a-zA-Z0-9_./-]+
STRING      = [^\n]+           // Rest of line in value context

// Operators
EQUALS      = '='
COLON       = ':'
LPAREN      = '('
RPAREN      = ')'

// Interpolation (context-sensitive recognition)
INTERP_START = '{' when preceded by SPACE | SOL | ':' | '='
                   and followed by IDENTIFIER start
INTERP_END   = '}'
INTERP_MOD   = ':raw'
ESCAPE_BRACE = '{{' | '}}'

// Keywords
DOT_KEYWORD = '.shell' | '.parallel' | '.default' | '.include'
            | '.environment' | '.using' | '.source' | '.args'
            | '.requires' | '.after' | '.autodeps'

KEYWORD     = 'lazy' | 'if' | 'elif' | 'else' | 'end'
            | 'ifdef' | 'ifndef' | 'block'

FUNC_NAME   = 'shell' | 'glob' | 'basename' | 'dirname' | 'replace'

// Comparison
COMP_OP     = '==' | '!='
```

### 9.2 Syntactic Grammar

See Section 3.2 for the full EBNF grammar.

---

## 10. Design Decisions Log

| Decision | Rationale | Alternatives Considered |
|----------|-----------|------------------------|
| Whitespace boundary for interpolation | Shell compatibility, predictability | Always interpolate, require escaping |
| `{{` for literal brace | Simple, familiar from other languages | Backslash escaping |
| Same `{name}` syntax for captures and interpolations | User simplicity - one syntax to learn | Separate syntax like `{*name}` for captures |
| Context-sensitive capture detection | Keeps syntax clean, semantic analysis is straightforward | Different syntax would be context-free but uglier |
| Line-oriented lexer | Matches language structure, simpler implementation | Character-oriented with complex state |
| Directive scope by indentation | Consistent with overall language design | Different keywords per scope (`.recipe-shell:`) |
| Immediate evaluation by default | Matches user expectation, explicit `lazy` keyword | Lazy by default like Make |
| Parser scope stack for directives | Validates directive placement during parsing | Defer all validation to semantic analysis |
| Automatic variables only in recipes | Prevents confusion, clear error messages | Allow everywhere, resolve at execution |
