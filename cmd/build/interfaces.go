// Package main provides the CLI for the build tool.
//
// This file defines interfaces for components used by the CLI.
// Internal packages provide concrete implementations that satisfy these interfaces.
// This follows the Go idiom of "accept interfaces, return structs" and allows
// the CLI to remain decoupled from implementation details.
package main

// Token represents a lexical token from source code.
// The CLI uses this interface to inspect tokens without depending on
// the concrete lexer.Token type.
type Token interface {
	// TokenType returns the type of the token as a string.
	TokenType() string

	// TokenLiteral returns the literal value of the token.
	TokenLiteral() string

	// TokenLocation returns the source location as "file:line:column".
	TokenLocation() string

	// IsEOF returns true if this is an end-of-file token.
	IsEOF() bool

	// IsError returns true if this is an error token.
	IsError() bool
}

// Lexer tokenizes source code into a stream of tokens.
// Implementations should be created via a factory function that accepts
// the file path and source content.
type Lexer interface {
	// NextToken returns the next token from the input.
	// Returns a token with IsEOF() == true when input is exhausted.
	NextToken() Token
}

// LexerFactory creates a new Lexer for the given source.
type LexerFactory func(file, input string) Lexer

// Scope represents the parsing context (global, environment, recipe, etc.).
type Scope interface {
	// String returns a human-readable name for the scope.
	String() string
}

// Parser transforms a token stream into an AST.
// Implementations track parsing scope for directive validation.
type Parser interface {
	// CurrentScope returns the current parsing scope.
	CurrentScope() Scope

	// EnterScope pushes a new scope onto the stack.
	EnterScope(scope Scope)

	// ExitScope pops the current scope and returns it.
	ExitScope() Scope

	// CurrentIndentLevel returns the expected indentation level for the current scope.
	CurrentIndentLevel() int

	// HasErrors returns true if any parse errors were encountered.
	HasErrors() bool
}

// ParserFactory creates a new Parser for the given Lexer.
// Note: The factory accepts the concrete lexer type since the parser
// needs to call NextToken() on it during parsing.
type ParserFactory func(lexer Lexer) Parser

// DirectiveValidator checks if a directive is valid at a given scope.
type DirectiveValidator interface {
	// IsValidAtScope returns true if the directive token type is valid at the given scope.
	IsValidAtScope(tokenType string, scope Scope) bool

	// DirectiveName returns the display name for a directive token type (e.g., ".shell").
	DirectiveName(tokenType string) string
}

// Variable represents a parsed variable definition from the AST.
type Variable interface {
	// Name returns the variable name.
	Name() string

	// IsLazy returns true if this is a lazy variable.
	IsLazy() bool

	// ValueParts returns the parts of the variable value.
	// Each part is either a literal string, interpolation, or function call.
	ValueParts() []ValuePart

	// Location returns the source location as "file:line:column".
	Location() string
}

// ValuePart represents a part of a value (literal, interpolation, or function call).
type ValuePart interface {
	// PartType returns "literal", "interpolation", or "function".
	PartType() string

	// Text returns the text content (for literals) or name (for interpolations/functions).
	Text() string

	// IsRaw returns true for interpolations with :raw modifier.
	IsRaw() bool
}

// VariableParser parses variable definitions.
type VariableParser interface {
	// ParseVariable parses a variable definition from the current position.
	// Returns the variable and any error encountered.
	ParseVariable() (Variable, error)
}

// Target represents a parsed target definition from the AST.
type Target interface {
	// PatternText returns a textual representation of the target pattern.
	PatternText() string

	// IsPhony returns true for @name phony targets.
	IsPhony() bool

	// IsDirectory returns true for targets ending with /.
	IsDirectory() bool

	// DependencyCount returns the number of dependencies.
	DependencyCount() int

	// DependencyText returns the textual representation of the i-th dependency.
	DependencyText(i int) string

	// HasCaptures returns true if the pattern contains {name} expressions.
	HasCaptures() bool

	// CaptureNames returns the names of captures in the pattern.
	CaptureNames() []string

	// Location returns the source location as "file:line:column".
	Location() string

	// HasRecipe returns true if the target has a recipe.
	HasRecipe() bool

	// Recipe returns the recipe information (if any).
	Recipe() Recipe
}

// Recipe represents a parsed recipe from the AST.
type Recipe interface {
	// CommandCount returns the number of commands.
	CommandCount() int

	// CommandText returns textual representation of the i-th command.
	CommandText(i int) string

	// IsBlockCommand returns true if the i-th command is a block command.
	IsBlockCommand(i int) bool

	// HasShellDirective returns true if .shell: is specified.
	HasShellDirective() bool

	// HasAfterDirective returns true if .after: is specified.
	HasAfterDirective() bool

	// HasAutodepsDirective returns true if .autodeps: is specified.
	HasAutodepsDirective() bool

	// RequiresCount returns the number of .requires entries.
	RequiresCount() int

	// RequirementName returns the i-th requirement name.
	RequirementName(i int) string

	// RequirementVersion returns the i-th requirement version.
	RequirementVersion(i int) string

	// Location returns the source location as "file:line:column".
	Location() string
}

// TargetParser parses target definitions.
type TargetParser interface {
	// ParseTarget parses a target definition from the current position.
	// Returns the target and any error encountered.
	ParseTarget() (Target, error)
}

// Environment represents a parsed environment block from the AST.
type Environment interface {
	// Name returns the environment name (empty for default environment).
	Name() string

	// IsDefault returns true if this is the default (unnamed) environment.
	IsDefault() bool

	// RuntimeType returns the runtime type (bare, docker, podman, etc.).
	RuntimeType() string

	// HasRuntime returns true if .using: is specified.
	HasRuntime() bool

	// Source returns the source path for the environment.
	Source() string

	// HasSource returns true if .source: is specified.
	HasSource() bool

	// Args returns the runtime args.
	Args() string

	// HasArgs returns true if .args: is specified.
	HasArgs() bool

	// RequiresCount returns the number of requirements.
	RequiresCount() int

	// RequirementName returns the i-th requirement name.
	RequirementName(i int) string

	// RequirementVersion returns the i-th requirement version.
	RequirementVersion(i int) string

	// Location returns the source location as "file:line:column".
	Location() string
}

// EnvironmentParser parses environment blocks.
type EnvironmentParser interface {
	// ParseEnvironment parses an environment block from the current position.
	// Returns the environment and any error encountered.
	ParseEnvironment() (Environment, error)
}

// Conditional represents a parsed conditional block from the AST.
type Conditional interface {
	// ConditionType returns "equals", "not_equals", "defined", or "not_defined".
	ConditionType() string

	// ConditionLeftText returns the left side of the condition.
	ConditionLeftText() string

	// ConditionRightText returns the right side of the condition (for == and !=).
	ConditionRightText() string

	// ConditionVarName returns the variable name for ifdef/ifndef.
	ConditionVarName() string

	// IfBodyCount returns the number of statements in the if body.
	IfBodyCount() int

	// ElifCount returns the number of elif branches.
	ElifCount() int

	// ElifConditionType returns the condition type for the i-th elif.
	ElifConditionType(i int) string

	// ElifConditionLeftText returns the left side of the i-th elif condition.
	ElifConditionLeftText(i int) string

	// ElifConditionRightText returns the right side of the i-th elif condition.
	ElifConditionRightText(i int) string

	// ElifBodyCount returns the number of statements in the i-th elif body.
	ElifBodyCount(i int) int

	// HasElse returns true if there is an else clause.
	HasElse() bool

	// ElseBodyCount returns the number of statements in the else body.
	ElseBodyCount() int

	// Location returns the source location as "file:line:column".
	Location() string
}

// ConditionalParser parses conditional blocks.
type ConditionalParser interface {
	// ParseConditional parses a conditional block from the current position.
	// Returns the conditional and any error encountered.
	ParseConditional() (Conditional, error)

	// IsConditionalLine returns true if the current line starts a conditional.
	IsConditionalLine() bool
}

// IncludeResult represents the result of parsing an .include: directive.
type IncludeResult interface {
	// DirectiveKind returns "include".
	DirectiveKind() string

	// Path returns the include path.
	Path() string

	// IncludedStatementCount returns the number of statements from the included file.
	IncludedStatementCount() int

	// IncludedStatementType returns the type of the i-th included statement.
	IncludedStatementType(i int) string

	// IncludedStatementText returns a text representation of the i-th statement.
	IncludedStatementText(i int) string

	// Location returns the source location as "file:line:column".
	Location() string
}

// IncludeParser parses .include: directives.
type IncludeParser interface {
	// ParseInclude parses an .include: directive from the current position.
	// Returns the include result and any error encountered.
	ParseInclude() (IncludeResult, error)

	// IsIncludeLine returns true if the current line is an .include: directive.
	IsIncludeLine() bool
}

// Statement represents a parsed AST statement.
type Statement interface {
	// StatementType returns the type of statement (e.g., "variable", "target", "directive").
	StatementType() string

	// Location returns the source location as "file:line:column".
	Location() string

	// Summary returns a brief text summary of the statement content.
	Summary() string
}

// ParseError represents a parse error with location information.
type ParseError interface {
	// Error returns the error message.
	Error() string

	// Location returns the source location.
	ErrorLocation() string

	// Hint returns an optional hint for fixing the error.
	ErrorHint() string
}

// BuildfileResult represents the result of parsing a complete buildfile.
type BuildfileResult interface {
	// Statements returns the successfully parsed statements.
	Statements() []Statement

	// ErrorCount returns the number of parse errors.
	ErrorCount() int

	// GetError returns the i-th parse error.
	GetError(i int) ParseError

	// HasErrors returns true if there are any parse errors.
	HasErrors() bool

	// AllErrors returns all errors as a single string.
	AllErrors() string
}

// BuildfileParser parses complete buildfiles with error recovery.
type BuildfileParser interface {
	// ParseBuildfile parses the complete buildfile with error recovery.
	// Returns statements and errors. Parsing continues after errors to collect
	// as many valid statements as possible.
	ParseBuildfile() BuildfileResult
}

// SymbolTable tracks defined symbols (variables, targets, environments)
// for semantic analysis.
type SymbolTable interface {
	// AddVariable adds a variable to the symbol table.
	// Returns an error if a duplicate variable is detected.
	AddVariable(v interface{}) error

	// AddConditionalVariable adds a variable defined within a conditional.
	// Unlike AddVariable, this allows multiple definitions for the same name
	// (one per conditional branch) since only one branch executes at runtime.
	AddConditionalVariable(varDef interface{}, cond interface{}, branchType string, branchIndex int)

	// AddTarget adds a target to the symbol table.
	// Returns an error if a duplicate target is detected.
	AddTarget(t interface{}) error

	// AddEnvironment adds an environment to the symbol table.
	// Returns an error if a duplicate environment is detected.
	AddEnvironment(e interface{}) error

	// VariableCount returns the number of variables.
	VariableCount() int

	// VariableName returns the name of the i-th variable.
	VariableName(i int) string

	// VariableLocation returns the location of the i-th variable.
	VariableLocation(i int) string

	// VariableIsLazy returns true if the i-th variable is lazy.
	VariableIsLazy(i int) bool

	// ConditionalVarCount returns the number of conditional variable names.
	ConditionalVarCount() int

	// ConditionalVarName returns the name of the i-th conditional variable.
	ConditionalVarName(i int) string

	// ConditionalVarDefCount returns the number of definitions for the i-th conditional variable.
	ConditionalVarDefCount(i int) int

	// ConditionalVarDefLocation returns the location of the j-th definition of the i-th conditional variable.
	ConditionalVarDefLocation(i, j int) string

	// ConditionalVarDefBranch returns the branch info ("if", "elif[0]", "else") for the j-th definition.
	ConditionalVarDefBranch(i, j int) string

	// TargetCount returns the number of targets.
	TargetCount() int

	// TargetPattern returns the pattern of the i-th target.
	TargetPattern(i int) string

	// TargetLocation returns the location of the i-th target.
	TargetLocation(i int) string

	// EnvironmentCount returns the number of environments.
	EnvironmentCount() int

	// EnvironmentName returns the name of the i-th environment.
	EnvironmentName(i int) string

	// EnvironmentLocation returns the location of the i-th environment.
	EnvironmentLocation(i int) string

	// IsAutomatic returns true if name is an automatic variable.
	IsAutomatic(name string) bool

	// IsBuiltin returns true if name is a built-in variable.
	IsBuiltin(name string) bool

	// IsDefined returns true if name is defined.
	IsDefined(name string) bool
}

// CollectResult represents the result of Pass 1: Symbol Collection.
type CollectResult interface {
	// SymbolTable returns the symbol table populated during collection.
	SymbolTable() SymbolTable

	// HasErrors returns true if any errors were encountered during collection.
	HasErrors() bool

	// ErrorCount returns the number of errors.
	ErrorCount() int

	// GetError returns the i-th error.
	GetError(i int) error

	// Errors returns all errors.
	Errors() []error
}

// CaptureResult represents the result of Pass 2: Capture Validation.
type CaptureResult interface {
	// HasErrors returns true if any validation errors were found.
	HasErrors() bool

	// ErrorCount returns the number of errors.
	ErrorCount() int

	// GetError returns the i-th error.
	GetError(i int) error

	// Errors returns all errors.
	Errors() []error

	// CaptureCount returns the number of targets with captures.
	CaptureCount() int

	// TargetPattern returns the pattern of the i-th target with captures.
	TargetPattern(i int) string

	// CaptureNames returns the capture names for the i-th target.
	CaptureNames(i int) []string

	// InterpolationCount returns the number of targets with interpolations.
	InterpolationCount() int

	// InterpolationTargetPattern returns the pattern of the i-th target with interpolations.
	InterpolationTargetPattern(i int) string

	// InterpolationNames returns the interpolation names for the i-th target.
	InterpolationNames(i int) []string
}

// ReferenceResult represents the result of Pass 3: Reference Validation.
type ReferenceResult interface {
	// HasErrors returns true if any validation errors were found.
	HasErrors() bool

	// ErrorCount returns the number of errors.
	ErrorCount() int

	// GetError returns the i-th error.
	GetError(i int) error

	// Errors returns all errors.
	Errors() []error
}

// DependencyResult represents the result of Pass 4: Dependency Graph Validation.
type DependencyResult interface {
	// HasErrors returns true if any validation errors were found (e.g., circular dependencies).
	HasErrors() bool

	// ErrorCount returns the number of errors.
	ErrorCount() int

	// GetError returns the i-th error.
	GetError(i int) error

	// Errors returns all errors.
	Errors() []error

	// NodeCount returns the number of nodes in the dependency graph.
	NodeCount() int

	// NodeName returns the name of the i-th node.
	NodeName(i int) string

	// NodeEdgeCount returns the number of outgoing edges for the i-th node.
	NodeEdgeCount(i int) int

	// NodeEdge returns the j-th edge (dependency) for the i-th node.
	NodeEdge(i, j int) string

	// PatternTargetCount returns the number of pattern targets (not in the graph).
	PatternTargetCount() int

	// PatternTargetPattern returns the pattern of the i-th pattern target.
	PatternTargetPattern(i int) string

	// UnsatisfiedDepsCount returns the number of targets with unsatisfied dependencies.
	UnsatisfiedDepsCount() int

	// UnsatisfiedDepsTarget returns the target name for the i-th entry.
	UnsatisfiedDepsTarget(i int) string

	// UnsatisfiedDepsList returns the unsatisfied dependencies for the i-th entry.
	UnsatisfiedDepsList(i int) []string
}

// EvalContext represents the evaluation context with variable values.
type EvalContext interface {
	// Get retrieves a variable's evaluated value.
	// Returns the value and true if found, or ("", false) if not found.
	Get(name string) (string, bool)

	// Set sets a variable's value.
	Set(name, value string)

	// IsDefined returns true if the variable is defined.
	IsDefined(name string) bool

	// IsLazy returns true if the variable is lazy (not yet evaluated).
	IsLazy(name string) bool

	// Variables returns all evaluated variable names and values.
	Variables() map[string]string

	// LazyVariables returns all lazy variable names.
	LazyVariables() map[string]string
}

// EvalResult represents the result of variable evaluation.
type EvalResult interface {
	// Context returns the evaluation context with variable values.
	Context() EvalContext

	// HasErrors returns true if any errors were encountered during evaluation.
	HasErrors() bool

	// ErrorCount returns the number of errors.
	ErrorCount() int

	// GetError returns the i-th error.
	GetError(i int) error

	// Errors returns all errors.
	Errors() []error

	// EvaluatedCount returns the number of evaluated (non-lazy) variables.
	EvaluatedCount() int

	// EvaluatedName returns the name of the i-th evaluated variable.
	EvaluatedName(i int) string

	// EvaluatedValue returns the value of the i-th evaluated variable.
	EvaluatedValue(i int) string

	// LazyCount returns the number of lazy variables.
	LazyCount() int

	// LazyName returns the name of the i-th lazy variable.
	LazyName(i int) string
}
