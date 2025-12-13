// Package ast defines the Abstract Syntax Tree node types for Buildfile parsing.
//
// The AST captures the syntactic structure of a Buildfile without interpretation.
// No evaluation happens during parsing - the AST is a pure representation of
// the source structure.
//
// Key design decisions:
//   - BraceExpr nodes in target patterns remain unresolved during parsing.
//     Semantic analysis resolves them to Capture or Interpolation.
//   - All nodes carry SourceLocation for error reporting.
//   - Interfaces use marker methods to ensure type safety at compile time.
package ast

import (
	"fmt"

	"github.com/vinayprograms/build/internal/lexer"
)

// SourceLocation represents a position in source code.
type SourceLocation struct {
	File   string // Source file path
	Line   int    // 1-based line number
	Column int    // 1-based column number
}

// String returns a human-readable representation of the source location.
func (l SourceLocation) String() string {
	return fmt.Sprintf("%s:%d:%d", l.File, l.Line, l.Column)
}

// SourceLocationFromToken creates a SourceLocation from a lexer token.
func SourceLocationFromToken(tok lexer.Token) SourceLocation {
	return SourceLocation{
		File:   tok.Location.File,
		Line:   tok.Location.Line,
		Column: tok.Location.Column,
	}
}

// ----------------------------------------------------------------------------
// Root Node
// ----------------------------------------------------------------------------

// Buildfile is the root AST node representing an entire Buildfile.
type Buildfile struct {
	SourcePath string      // Path to the source file
	Statements []Statement // Top-level statements
}

// ----------------------------------------------------------------------------
// Statement Interface
// ----------------------------------------------------------------------------

// Statement is the interface for all top-level AST nodes.
// All statement types implement this interface.
type Statement interface {
	statementNode() bool
}

// ----------------------------------------------------------------------------
// Directive Types
// ----------------------------------------------------------------------------

// DirectiveKind represents the type of a directive.
type DirectiveKind int

const (
	DirectiveShell DirectiveKind = iota
	DirectiveParallel
	DirectiveDefault
	DirectiveInclude
)

// String returns the string representation of the directive kind.
func (k DirectiveKind) String() string {
	switch k {
	case DirectiveShell:
		return "shell"
	case DirectiveParallel:
		return "parallel"
	case DirectiveDefault:
		return "default"
	case DirectiveInclude:
		return "include"
	default:
		return fmt.Sprintf("unknown(%d)", k)
	}
}

// Directive represents a global directive statement (.shell, .parallel, etc.).
type Directive struct {
	Kind     DirectiveKind
	Value    *Value
	Location SourceLocation
}

func (d *Directive) statementNode() bool { return true }

// ----------------------------------------------------------------------------
// Environment Types
// ----------------------------------------------------------------------------

// Runtime represents the runtime type for an environment.
type Runtime int

const (
	RuntimeBare Runtime = iota
	RuntimeDocker
	RuntimePodman
	RuntimeDevcontainer
	RuntimeNix
	RuntimeLima
)

// String returns the string representation of the runtime.
func (r Runtime) String() string {
	switch r {
	case RuntimeBare:
		return "bare"
	case RuntimeDocker:
		return "docker"
	case RuntimePodman:
		return "podman"
	case RuntimeDevcontainer:
		return "devcontainer"
	case RuntimeNix:
		return "nix"
	case RuntimeLima:
		return "lima"
	default:
		return fmt.Sprintf("unknown(%d)", r)
	}
}

// VersionSpec is the interface for version specifications.
type VersionSpec interface {
	versionSpecNode() bool
	String() string
}

// VersionLatest represents "latest" or unspecified version.
type VersionLatest struct{}

func (v VersionLatest) versionSpecNode() bool { return true }
func (v VersionLatest) String() string        { return "latest" }

// VersionMajor represents a major version (e.g., "11").
type VersionMajor struct {
	Major int
}

func (v VersionMajor) versionSpecNode() bool { return true }
func (v VersionMajor) String() string        { return fmt.Sprintf("%d", v.Major) }

// VersionMajorMinor represents a major.minor version (e.g., "11.4").
type VersionMajorMinor struct {
	Major int
	Minor int
}

func (v VersionMajorMinor) versionSpecNode() bool { return true }
func (v VersionMajorMinor) String() string        { return fmt.Sprintf("%d.%d", v.Major, v.Minor) }

// VersionExact represents an exact version (e.g., "11.4.0").
type VersionExact struct {
	Major int
	Minor int
	Patch int
}

func (v VersionExact) versionSpecNode() bool { return true }
func (v VersionExact) String() string        { return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch) }

// Requirement represents a required binary with optional version spec.
type Requirement struct {
	Name    string
	Version VersionSpec
}

// Environment represents an environment block.
type Environment struct {
	Name     *string       // nil for default environment
	Runtime  *Runtime      // .using
	Source   *Value        // .source
	Args     *Value        // .args
	Requires []Requirement // .requires
	Location SourceLocation
}

func (e *Environment) statementNode() bool { return true }

// ----------------------------------------------------------------------------
// Variable Types
// ----------------------------------------------------------------------------

// Variable represents a variable definition.
type Variable struct {
	Name     string
	Value    *Value
	Lazy     bool // true for lazy assignment
	Location SourceLocation
}

func (v *Variable) statementNode() bool { return true }

// ----------------------------------------------------------------------------
// Conditional Types
// ----------------------------------------------------------------------------

// Condition is the interface for conditional expressions.
type Condition interface {
	conditionNode() bool
}

// EqualsCondition represents a == comparison.
type EqualsCondition struct {
	Left  *Value
	Right *Value
}

func (c *EqualsCondition) conditionNode() bool { return true }

// NotEqualsCondition represents a != comparison.
type NotEqualsCondition struct {
	Left  *Value
	Right *Value
}

func (c *NotEqualsCondition) conditionNode() bool { return true }

// DefinedCondition represents an ifdef check.
type DefinedCondition struct {
	Name string
}

func (c *DefinedCondition) conditionNode() bool { return true }

// NotDefinedCondition represents an ifndef check.
type NotDefinedCondition struct {
	Name string
}

func (c *NotDefinedCondition) conditionNode() bool { return true }

// ConditionalBranch represents a single branch in a conditional.
type ConditionalBranch struct {
	Condition Condition
	Body      []Statement
}

// Conditional represents an if/elif/else/end block.
type Conditional struct {
	IfBranch     ConditionalBranch
	ElifBranches []ConditionalBranch
	ElseBody     []Statement // nil if no else clause
	Location     SourceLocation
}

func (c *Conditional) statementNode() bool { return true }

// ----------------------------------------------------------------------------
// Target Types
// ----------------------------------------------------------------------------

// PatternSegment is the interface for segments in a target pattern.
type PatternSegment interface {
	patternSegmentNode() bool
}

// LiteralSegment represents a literal string in a pattern.
type LiteralSegment struct {
	Text string
}

func (s *LiteralSegment) patternSegmentNode() bool { return true }

// BraceExpr represents an unresolved {name} in a target pattern.
// During parsing, we don't know if this is a capture or a variable interpolation.
// Semantic analysis will resolve this based on the symbol table.
type BraceExpr struct {
	Identifier string
	Location   SourceLocation
}

func (b *BraceExpr) patternSegmentNode() bool { return true }

// LiteralSegment also implements ValuePart for use in .after directives
func (s *LiteralSegment) valuePartNode() bool { return true }

// TargetPattern represents the left side of a target definition.
type TargetPattern struct {
	Segments    []PatternSegment
	IsPhony     bool // true for @name targets
	IsDirectory bool // true for targets ending with /
}

// Dependency represents a dependency in a target definition.
type Dependency struct {
	Segments []PatternSegment
}

// Target represents a target definition with dependencies and optional recipe.
type Target struct {
	Pattern      TargetPattern
	Dependencies []Dependency
	Recipe       *Recipe
	Location     SourceLocation
}

func (t *Target) statementNode() bool { return true }

// ----------------------------------------------------------------------------
// Recipe Types
// ----------------------------------------------------------------------------

// RecipeDirectives holds the directives that can appear in a recipe.
type RecipeDirectives struct {
	Shell    *Value        // .shell override
	After    []*Value      // .after order-only dependencies
	Autodeps *Value        // .autodeps file path
	Requires []Requirement // .requires binaries
}

// Command is the interface for recipe commands.
type Command interface {
	commandNode() bool
}

// CommandPart is the interface for parts of a command line.
type CommandPart interface {
	commandPartNode() bool
}

// LiteralCommand represents literal text in a command.
type LiteralCommand struct {
	Text string
}

func (c *LiteralCommand) commandPartNode() bool { return true }

// CommandInterpolation represents a variable interpolation in a command.
type CommandInterpolation struct {
	Name     string
	Raw      bool // true for :raw modifier
	Location SourceLocation
}

func (c *CommandInterpolation) commandPartNode() bool { return true }

// LineCommand represents a single command line.
type LineCommand struct {
	Parts    []CommandPart
	Location SourceLocation
}

func (c *LineCommand) commandNode() bool { return true }

// BlockCommand represents a block: with multiple lines.
type BlockCommand struct {
	Lines    [][]CommandPart
	Location SourceLocation
}

func (c *BlockCommand) commandNode() bool { return true }

// Recipe represents the recipe section of a target.
type Recipe struct {
	Directives RecipeDirectives
	Commands   []Command
	Location   SourceLocation
}

// ----------------------------------------------------------------------------
// Value Types
// ----------------------------------------------------------------------------

// ValuePart is the interface for parts of a value.
type ValuePart interface {
	valuePartNode() bool
}

// LiteralValue represents literal text in a value.
type LiteralValue struct {
	Text string
}

func (v *LiteralValue) valuePartNode() bool { return true }

// Interpolation represents a variable interpolation in a value.
type Interpolation struct {
	Name     string
	Raw      bool // true for :raw modifier
	Location SourceLocation
}

func (i *Interpolation) valuePartNode() bool { return true }

// FunctionName represents a built-in function name.
type FunctionName int

const (
	FuncShell FunctionName = iota
	FuncGlob
	FuncBasename
	FuncDirname
	FuncReplace
)

// String returns the string representation of the function name.
func (f FunctionName) String() string {
	switch f {
	case FuncShell:
		return "shell"
	case FuncGlob:
		return "glob"
	case FuncBasename:
		return "basename"
	case FuncDirname:
		return "dirname"
	case FuncReplace:
		return "replace"
	default:
		return fmt.Sprintf("unknown(%d)", f)
	}
}

// FunctionCall represents a function call in a value.
type FunctionCall struct {
	Name     FunctionName
	Args     []*Value
	Location SourceLocation
}

func (f *FunctionCall) valuePartNode() bool { return true }

// Value represents a composite value (string with interpolations and function calls).
type Value struct {
	Parts    []ValuePart
	Location SourceLocation
}

// ----------------------------------------------------------------------------
// Comment and Blank
// ----------------------------------------------------------------------------

// Comment represents a comment line.
type Comment struct {
	Text     string
	Location SourceLocation
}

func (c *Comment) statementNode() bool { return true }

// Blank represents a blank line.
type Blank struct {
	Location SourceLocation
}

func (b *Blank) statementNode() bool { return true }
