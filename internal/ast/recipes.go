package ast

import "fmt"

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
//
// The commandNode() marker method is unexported to prevent external packages
// from implementing this interface, ensuring a closed set of command types.
//
// Implementers: LineCommand, BlockCommand
type Command interface {
	commandNode()
}

// CommandPart is the interface for parts of a command line.
//
// The commandPartNode() marker method is unexported to prevent external packages
// from implementing this interface, ensuring a closed set of command part types.
//
// Implementers: LiteralCommand, CommandInterpolation
type CommandPart interface {
	commandPartNode()
}

// LiteralCommand represents literal text in a command.
type LiteralCommand struct {
	Text string
}

func (c *LiteralCommand) commandPartNode() {}

// CommandInterpolation represents a variable interpolation in a command.
type CommandInterpolation struct {
	Name     string
	Raw      bool // true for :raw modifier
	Location SourceLocation
}

func (c *CommandInterpolation) commandPartNode() {}

// LineCommand represents a single command line.
type LineCommand struct {
	Parts    []CommandPart
	Location SourceLocation
}

func (c *LineCommand) commandNode() {}

// BlockCommand represents a block: with multiple lines.
type BlockCommand struct {
	Lines    [][]CommandPart
	Location SourceLocation
}

func (c *BlockCommand) commandNode() {}

// Recipe represents the recipe section of a target.
type Recipe struct {
	Directives RecipeDirectives
	Commands   []Command
	Location   SourceLocation
}

// String returns a human-readable representation of the recipe.
func (r *Recipe) String() string {
	return fmt.Sprintf("Recipe(%d commands)", len(r.Commands))
}
