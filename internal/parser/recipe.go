package parser

import (
	"strconv"
	"strings"

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
	// Check if next line is indented (recipe start)
	if p.current.Type != lexer.INDENT {
		return nil, nil // No recipe
	}

	// Store the recipe start location
	loc := ast.SourceLocationFromToken(p.current)

	// Enter recipe scope
	p.enterScope(ScopeRecipe)
	defer p.exitScope()

	recipe := &ast.Recipe{
		Location: loc,
	}

	// Parse recipe lines until we dedent
	for {
		// Check for end of recipe (no indent, newline, or EOF)
		if p.current.Type == lexer.EOF {
			break
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
	p.nextToken() // consume .shell

	// Expect colon
	if p.current.Type != lexer.COLON {
		return &ParseError{
			Message:  "expected ':' after .shell",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume :

	// Parse value
	value := p.ParseValue()
	recipe.Directives.Shell = value

	// Consume newline
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	return nil
}

// parseRecipeAfter parses .after: directive in recipe.
func (p *Parser) parseRecipeAfter(recipe *ast.Recipe) *ParseError {
	p.nextToken() // consume .after

	// Expect colon
	if p.current.Type != lexer.COLON {
		return &ParseError{
			Message:  "expected ':' after .after",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume :

	// Parse value
	value := p.ParseValue()
	recipe.Directives.After = append(recipe.Directives.After, value)

	// Consume newline
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	return nil
}

// parseRecipeAutodeps parses .autodeps: directive in recipe.
func (p *Parser) parseRecipeAutodeps(recipe *ast.Recipe) *ParseError {
	p.nextToken() // consume .autodeps

	// Expect colon
	if p.current.Type != lexer.COLON {
		return &ParseError{
			Message:  "expected ':' after .autodeps",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume :

	// Parse value
	value := p.ParseValue()
	recipe.Directives.Autodeps = value

	// Consume newline
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	return nil
}

// parseRecipeRequires parses .requires: directive in recipe.
func (p *Parser) parseRecipeRequires(recipe *ast.Recipe) *ParseError {
	p.nextToken() // consume .requires

	// Expect colon
	if p.current.Type != lexer.COLON {
		return &ParseError{
			Message:  "expected ':' after .requires",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume :

	// Parse requirements list
	reqs, err := p.parseRequirementsList()
	if err != nil {
		return err
	}
	recipe.Directives.Requires = append(recipe.Directives.Requires, reqs...)

	// Consume newline
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	return nil
}

// parseRequirementsList parses a space-separated list of requirements.
// Each requirement has the form: name[@version]
func (p *Parser) parseRequirementsList() ([]ast.Requirement, *ParseError) {
	var reqs []ast.Requirement

	// Collect all text from the line
	var text string
	for p.current.Type != lexer.NEWLINE && p.current.Type != lexer.COMMENT && p.current.Type != lexer.EOF {
		text += p.current.Literal
		p.nextToken()
	}

	// Split on whitespace
	parts := strings.Fields(text)
	for _, part := range parts {
		req, err := parseRequirement(part)
		if err != nil {
			return nil, &ParseError{
				Message:  err.Error(),
				Location: p.current.Location,
			}
		}
		reqs = append(reqs, req)
	}

	return reqs, nil
}

// parseRequirement parses a single requirement string like "gcc@11.4".
func parseRequirement(s string) (ast.Requirement, error) {
	// Split on @
	parts := strings.SplitN(s, "@", 2)
	name := parts[0]

	var version ast.VersionSpec = ast.VersionLatest{}

	if len(parts) == 2 {
		verStr := parts[1]
		if verStr == "latest" {
			version = ast.VersionLatest{}
		} else {
			version = parseVersionSpec(verStr)
		}
	}

	return ast.Requirement{
		Name:    name,
		Version: version,
	}, nil
}

// parseVersionSpec parses a version string into a VersionSpec.
func parseVersionSpec(s string) ast.VersionSpec {
	parts := strings.Split(s, ".")

	switch len(parts) {
	case 1:
		// Major only
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			return ast.VersionLatest{}
		}
		return ast.VersionMajor{Major: major}

	case 2:
		// Major.Minor
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			return ast.VersionLatest{}
		}
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return ast.VersionLatest{}
		}
		return ast.VersionMajorMinor{Major: major, Minor: minor}

	case 3:
		// Exact version
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			return ast.VersionLatest{}
		}
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return ast.VersionLatest{}
		}
		patch, err := strconv.Atoi(parts[2])
		if err != nil {
			return ast.VersionLatest{}
		}
		return ast.VersionExact{Major: major, Minor: minor, Patch: patch}

	default:
		return ast.VersionLatest{}
	}
}

// parseCommandLine parses a single command line.
// Grammar: command = { command_part } ;
// Grammar: command_part = STRING | interpolation ;
func (p *Parser) parseCommandLine() (*ast.LineCommand, *ParseError) {
	loc := ast.SourceLocationFromToken(p.current)
	var parts []ast.CommandPart

	// Parse command parts until newline
	for p.current.Type != lexer.NEWLINE && p.current.Type != lexer.EOF && p.current.Type != lexer.COMMENT {
		switch p.current.Type {
		case lexer.STRING:
			parts = append(parts, &ast.LiteralCommand{Text: p.current.Literal})
			p.nextToken()

		case lexer.INTERP_START:
			interp, err := p.parseCommandInterpolation()
			if err != nil {
				return nil, err
			}
			if interp != nil {
				parts = append(parts, interp)
			}

		case lexer.ESCAPE_LBRACE:
			parts = append(parts, &ast.LiteralCommand{Text: "{"})
			p.nextToken()

		case lexer.ESCAPE_RBRACE:
			parts = append(parts, &ast.LiteralCommand{Text: "}"})
			p.nextToken()

		case lexer.PATH, lexer.IDENTIFIER:
			parts = append(parts, &ast.LiteralCommand{Text: p.current.Literal})
			p.nextToken()

		default:
			// Other tokens as literal text
			parts = append(parts, &ast.LiteralCommand{Text: p.current.Literal})
			p.nextToken()
		}
	}

	// Consume newline
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	if len(parts) == 0 {
		return nil, nil
	}

	return &ast.LineCommand{
		Parts:    parts,
		Location: loc,
	}, nil
}

// parseCommandInterpolation parses an interpolation in a command ({var} or {var:raw}).
func (p *Parser) parseCommandInterpolation() (*ast.CommandInterpolation, *ParseError) {
	if p.current.Type != lexer.INTERP_START {
		return nil, nil
	}
	loc := ast.SourceLocationFromToken(p.current)
	p.nextToken() // consume {

	// Expect identifier
	if p.current.Type != lexer.IDENTIFIER {
		return nil, &ParseError{
			Message:  "expected identifier in interpolation",
			Location: p.current.Location,
		}
	}

	name := p.current.Literal
	p.nextToken()

	// Check for :raw modifier
	raw := false
	if p.current.Type == lexer.INTERP_MOD {
		if p.current.Literal == ":raw" {
			raw = true
		}
		p.nextToken()
	}

	// Expect close brace
	if p.current.Type != lexer.INTERP_END {
		return nil, &ParseError{
			Message:  "expected '}' to close interpolation",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume }

	return &ast.CommandInterpolation{
		Name:     name,
		Raw:      raw,
		Location: loc,
	}, nil
}

// parseBlockCommand parses a block: statement with nested lines.
// Grammar: block_stmt = "block:" NEWLINE INDENT { raw_line } DEDENT ;
func (p *Parser) parseBlockCommand() (*ast.BlockCommand, *ParseError) {
	loc := ast.SourceLocationFromToken(p.current)

	if p.current.Type != lexer.BLOCK {
		return nil, &ParseError{
			Message:  "expected 'block:' keyword",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume block

	// May have a colon after block (block: or block)
	if p.current.Type == lexer.COLON {
		p.nextToken() // consume :
	}

	// Consume newline
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	// Enter block scope
	p.enterScope(ScopeBlock)
	defer p.exitScope()

	var lines [][]ast.CommandPart

	// Parse block lines until dedent
	for {
		if p.current.Type == lexer.EOF {
			break
		}

		// Check for indent
		if p.current.Type != lexer.INDENT {
			break
		}

		// Check indent level - block content should be at level 2
		indentLevel := p.calculateIndentLevel(p.current.Literal)
		if indentLevel < 2 {
			// Back to recipe level - end of block
			break
		}

		// Block content is always command lines - switch to command mode
		p.lexer.SetCommandMode()
		p.nextToken() // consume INDENT

		// Empty line in block
		if p.current.Type == lexer.NEWLINE {
			p.nextToken()
			continue
		}

		// Comment line in block
		if p.current.Type == lexer.COMMENT {
			p.nextToken()
			if p.current.Type == lexer.NEWLINE {
				p.nextToken()
			}
			continue
		}

		// Parse block line
		lineParts, err := p.parseBlockLine()
		if err != nil {
			return nil, err
		}
		if len(lineParts) > 0 {
			lines = append(lines, lineParts)
		}
	}

	return &ast.BlockCommand{
		Lines:    lines,
		Location: loc,
	}, nil
}

// parseBlockLine parses a single line in a block.
func (p *Parser) parseBlockLine() ([]ast.CommandPart, *ParseError) {
	var parts []ast.CommandPart

	// Parse parts until newline
	for p.current.Type != lexer.NEWLINE && p.current.Type != lexer.EOF && p.current.Type != lexer.COMMENT {
		switch p.current.Type {
		case lexer.STRING:
			parts = append(parts, &ast.LiteralCommand{Text: p.current.Literal})
			p.nextToken()

		case lexer.INTERP_START:
			interp, err := p.parseCommandInterpolation()
			if err != nil {
				return nil, err
			}
			if interp != nil {
				parts = append(parts, interp)
			}

		case lexer.ESCAPE_LBRACE:
			parts = append(parts, &ast.LiteralCommand{Text: "{"})
			p.nextToken()

		case lexer.ESCAPE_RBRACE:
			parts = append(parts, &ast.LiteralCommand{Text: "}"})
			p.nextToken()

		default:
			parts = append(parts, &ast.LiteralCommand{Text: p.current.Literal})
			p.nextToken()
		}
	}

	// Consume newline
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	return parts, nil
}

// skipToNewline advances past the current line.
func (p *Parser) skipToNewline() {
	for p.current.Type != lexer.NEWLINE && p.current.Type != lexer.EOF {
		p.nextToken()
	}
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}
}
