package parser

import (
	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/lexer"
)

// parseRecipe parses the recipe section of a target definition.
// Grammar: recipe = INDENT { recipe_line } DEDENT ;
// Grammar: recipe_line = recipe_directive NEWLINE | block_stmt | command NEWLINE ;
//
// A recipe starts when we see an INDENT token after a target line.
// It ends when we encounter a line at indentation level 0 (dedent).
func (p *Parser) parseRecipe() (*ast.Recipe, *ParseError) {
	// Blank lines (bare NEWLINE, no leading whitespace) between the target
	// line and the first recipe line must not be mistaken for "no recipe
	// follows" - skip them before deciding whether a recipe is present.
	for p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	// Check if next line is indented (recipe start)
	if p.current.Type != lexer.INDENT {
		return nil, nil // No recipe
	}

	// Store the recipe start location
	loc := ast.SourceLocationFromToken(p.current)

	// Enter recipe scope
	p.EnterScope(ScopeRecipe)
	defer p.ExitScope()

	recipe := &ast.Recipe{
		Location: loc,
	}

	// Parse recipe lines until we dedent
	for {
		// Check for end of recipe (no indent, newline, or EOF)
		if p.current.Type == lexer.EOF {
			break
		}

		if p.current.Type == lexer.ERROR {
			return nil, p.errorFromLexerToken(p.current)
		}

		// A blank line with no leading whitespace produces a bare NEWLINE (no
		// INDENT token). Blank lines never terminate a recipe, so skip it and
		// keep scanning for the next line.
		if p.current.Type == lexer.NEWLINE {
			p.nextToken()
			continue
		}

		// At the start of each line, we should have an INDENT token
		if p.current.Type != lexer.INDENT {
			// Not indented - end of recipe
			break
		}

		// Store indent and advance
		indentLevel := p.calculateIndentLevel(p.current.Literal)
		if indentLevel == 0 {
			// Back to global scope - end of recipe
			break
		}

		// Peek ahead to determine if this is a directive, block, or command line
		// Only switch to command mode for actual commands (not directives or block)
		if !p.lexer.PeekNextIsDotKeyword() && !p.lexer.PeekNextIsBlock() {
			p.lexer.SetCommandMode()
		}
		p.nextToken() // consume INDENT

		// Empty line or comment in recipe
		if p.current.Type == lexer.NEWLINE {
			p.nextToken()
			continue
		}
		if p.current.Type == lexer.COMMENT {
			p.nextToken()
			if p.current.Type == lexer.NEWLINE {
				p.nextToken()
			}
			continue
		}

		// Check for recipe directives
		if p.current.Type.IsDotKeyword() {
			if err := p.parseRecipeDirective(recipe); err != nil {
				return nil, err
			}
			continue
		}

		// An unknown dot-keyword (e.g. ".unknownthing") is lexed as a PATH
		// token, since the lexer doesn't recognize it as a directive; this
		// only happens when PeekNextIsDotKeyword kept the lexer in normal
		// mode (i.e. the line looked like a directive, not a "./..." command).
		// If it's immediately followed by ':', it's an unknown directive.
		if p.current.Type == lexer.PATH && isUnknownDirectiveCandidate(p.current.Literal) {
			text := p.current.Literal
			dirLoc := p.current.Location
			next := p.nextToken()
			if next.Type == lexer.COLON {
				return nil, p.unknownDirectiveError(text, dirLoc)
			}
			// Not actually a directive attempt - treat the rest of the line
			// as a command, using the already-consumed token as its start.
			cmdLoc := ast.SourceLocation{File: dirLoc.File, Line: dirLoc.Line, Column: dirLoc.Column}
			lineCmd, err := p.parseCommandLineContinuation(cmdLoc, []ast.CommandPart{&ast.LiteralCommand{Text: text}})
			if err != nil {
				return nil, err
			}
			if lineCmd != nil {
				recipe.Commands = append(recipe.Commands, lineCmd)
			}
			continue
		}

		// Check for block: keyword
		if p.current.Type == lexer.BLOCK {
			blockCmd, err := p.parseBlockCommand()
			if err != nil {
				return nil, err
			}
			recipe.Commands = append(recipe.Commands, blockCmd)
			continue
		}

		// Regular command line - lexer is already in command mode
		lineCmd, err := p.parseCommandLine()
		if err != nil {
			return nil, err
		}
		if lineCmd != nil {
			recipe.Commands = append(recipe.Commands, lineCmd)
		}
	}

	return recipe, nil
}

// calculateIndentLevel determines the logical indentation level from the indent string.
// Level 0 = no indent, Level 1 = recipe indent, Level 2 = block indent.
func (p *Parser) calculateIndentLevel(indent string) int {
	if indent == "" {
		return 0
	}

	level, err := p.lexer.IndentTracker().Process(indent)
	if err != nil {
		// On error, return 1 (assume recipe level)
		return 1
	}
	return level
}

// parseRecipeDirective parses a directive within a recipe (.shell, .after, .autodeps, .requires).
func (p *Parser) parseRecipeDirective(recipe *ast.Recipe) *ParseError {
	switch p.current.Type {
	case lexer.DOT_SHELL:
		return p.parseRecipeShell(recipe)
	case lexer.DOT_AFTER:
		return p.parseRecipeAfter(recipe)
	case lexer.DOT_AUTODEPS:
		return p.parseRecipeAutodeps(recipe)
	case lexer.DOT_REQUIRES:
		return p.parseRecipeRequires(recipe)
	default:
		// Validate directive is allowed at recipe scope
		if err := p.validateDirectiveScope(p.current); err != nil {
			return err
		}
		// Skip unknown directive
		p.skipToNewline()
		return nil
	}
}

// parseRecipeShell parses .shell: directive in recipe.
func (p *Parser) parseRecipeShell(recipe *ast.Recipe) *ParseError {
	if err := p.expectColon(".shell"); err != nil {
		return err
	}

	recipe.Directives.Shell = p.ParseValue()
	p.consumeNewline()
	return nil
}

// parseRecipeAfter parses .after: directive in recipe.
func (p *Parser) parseRecipeAfter(recipe *ast.Recipe) *ParseError {
	if err := p.expectColon(".after"); err != nil {
		return err
	}

	recipe.Directives.After = append(recipe.Directives.After, p.ParseValue())
	p.consumeNewline()
	return nil
}

// parseRecipeAutodeps parses .autodeps: directive in recipe.
func (p *Parser) parseRecipeAutodeps(recipe *ast.Recipe) *ParseError {
	if err := p.expectColon(".autodeps"); err != nil {
		return err
	}

	recipe.Directives.Autodeps = p.ParseValue()
	p.consumeNewline()
	return nil
}

// parseRecipeRequires parses .requires: directive in recipe.
func (p *Parser) parseRecipeRequires(recipe *ast.Recipe) *ParseError {
	if err := p.expectColon(".requires"); err != nil {
		return err
	}

	reqs, err := p.parseRequirementsList()
	if err != nil {
		return err
	}
	recipe.Directives.Requires = append(recipe.Directives.Requires, reqs...)

	p.consumeNewline()
	return nil
}
