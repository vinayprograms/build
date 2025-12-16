package parser

import (
	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/lexer"
)

// IsConditionalLine checks if the current token starts a conditional.
// Returns true for if, ifdef, or ifndef keywords.
func (p *Parser) IsConditionalLine() bool {
	switch p.current.Type {
	case lexer.IF, lexer.IFDEF, lexer.IFNDEF:
		return true
	default:
		return false
	}
}

// ParseConditional parses a conditional block (if/elif/else/end or ifdef/ifndef/end).
// Grammar:
//
//	conditional = if_clause { elif_clause } [ else_clause ] "end" NEWLINE ;
//	if_clause = "if" condition NEWLINE { statement } ;
//	elif_clause = "elif" condition NEWLINE { statement } ;
//	else_clause = "else" NEWLINE { statement } ;
//	condition = interpolation "==" value | interpolation "!=" value ;
//	ifdef_clause = "ifdef" identifier NEWLINE { statement } "end" NEWLINE ;
//	ifndef_clause = "ifndef" identifier NEWLINE { statement } "end" NEWLINE ;
func (p *Parser) ParseConditional() (*ast.Conditional, *ParseError) {
	loc := ast.SourceLocationFromToken(p.current)

	// Determine conditional type
	switch p.current.Type {
	case lexer.IF:
		return p.parseIfConditional(loc)
	case lexer.IFDEF:
		return p.parseIfdefConditional(loc, true)
	case lexer.IFNDEF:
		return p.parseIfdefConditional(loc, false)
	default:
		return nil, &ParseError{
			Message:  "expected 'if', 'ifdef', or 'ifndef'",
			Location: p.current.Location,
		}
	}
}

// parseIfConditional parses an if/elif/else/end block.
func (p *Parser) parseIfConditional(loc ast.SourceLocation) (*ast.Conditional, *ParseError) {
	p.nextToken() // consume 'if'

	// Parse if condition
	cond, err := p.parseCondition()
	if err != nil {
		return nil, err
	}

	// Expect newline after condition
	if p.current.Type != lexer.NEWLINE {
		return nil, &ParseError{
			Message:  "expected newline after condition",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume newline

	// Parse if body
	ifBody, err := p.parseConditionalBody()
	if err != nil {
		return nil, err
	}

	conditional := &ast.Conditional{
		IfBranch: ast.ConditionalBranch{
			Condition: cond,
			Body:      ifBody,
		},
		Location: loc,
	}

	// Parse elif branches
	for p.current.Type == lexer.ELIF {
		elifBranch, err := p.parseElifBranch()
		if err != nil {
			return nil, err
		}
		conditional.ElifBranches = append(conditional.ElifBranches, *elifBranch)
	}

	// Parse else body if present
	if p.current.Type == lexer.ELSE {
		p.nextToken() // consume 'else'

		// Expect newline after else
		if p.current.Type != lexer.NEWLINE {
			return nil, &ParseError{
				Message:  "expected newline after 'else'",
				Location: p.current.Location,
			}
		}
		p.nextToken() // consume newline

		elseBody, err := p.parseConditionalBody()
		if err != nil {
			return nil, err
		}
		conditional.ElseBody = elseBody
	}

	// Expect end
	if p.current.Type != lexer.END {
		return nil, &ParseError{
			Message:  "expected 'end' to close conditional",
			Location: p.current.Location,
			Hint:     "every 'if' must be closed with 'end'",
		}
	}
	p.nextToken() // consume 'end'

	// Consume trailing newline if present
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	return conditional, nil
}

// parseIfdefConditional parses an ifdef/ifndef/end block.
func (p *Parser) parseIfdefConditional(loc ast.SourceLocation, isDefined bool) (*ast.Conditional, *ParseError) {
	p.nextToken() // consume 'ifdef' or 'ifndef'

	// Expect identifier
	if p.current.Type != lexer.IDENTIFIER {
		return nil, &ParseError{
			Message:  "expected identifier after 'ifdef'/'ifndef'",
			Location: p.current.Location,
		}
	}

	varName := p.current.Literal
	p.nextToken()

	// Create condition
	var cond ast.Condition
	if isDefined {
		cond = &ast.DefinedCondition{Name: varName}
	} else {
		cond = &ast.NotDefinedCondition{Name: varName}
	}

	// Expect newline
	if p.current.Type != lexer.NEWLINE {
		return nil, &ParseError{
			Message:  "expected newline after identifier",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume newline

	// Parse body
	body, err := p.parseConditionalBody()
	if err != nil {
		return nil, err
	}

	// Expect end
	if p.current.Type != lexer.END {
		return nil, &ParseError{
			Message:  "expected 'end' to close conditional",
			Location: p.current.Location,
			Hint:     "every 'ifdef'/'ifndef' must be closed with 'end'",
		}
	}
	p.nextToken() // consume 'end'

	// Consume trailing newline if present
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	return &ast.Conditional{
		IfBranch: ast.ConditionalBranch{
			Condition: cond,
			Body:      body,
		},
		Location: loc,
	}, nil
}

// parseElifBranch parses a single elif branch.
func (p *Parser) parseElifBranch() (*ast.ConditionalBranch, *ParseError) {
	p.nextToken() // consume 'elif'

	// Parse condition
	cond, err := p.parseCondition()
	if err != nil {
		return nil, err
	}

	// Expect newline
	if p.current.Type != lexer.NEWLINE {
		return nil, &ParseError{
			Message:  "expected newline after condition",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume newline

	// Parse body
	body, err := p.parseConditionalBody()
	if err != nil {
		return nil, err
	}

	return &ast.ConditionalBranch{
		Condition: cond,
		Body:      body,
	}, nil
}

// parseCondition parses a condition expression.
// Grammar: condition = interpolation "==" value | interpolation "!=" value ;
func (p *Parser) parseCondition() (ast.Condition, *ParseError) {
	// Parse left side (should be an interpolation or value starting with interpolation)
	left := p.parseConditionValue()
	if left == nil || len(left.Parts) == 0 {
		return nil, &ParseError{
			Message:  "expected condition expression",
			Location: p.current.Location,
		}
	}

	// Expect comparison operator
	var isEquals bool
	switch p.current.Type {
	case lexer.DOUBLE_EQUALS:
		isEquals = true
	case lexer.NOT_EQUALS:
		isEquals = false
	default:
		return nil, &ParseError{
			Message:  "expected '==' or '!=' in condition",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume operator

	// Parse right side
	right := p.parseConditionValue()
	if right == nil || len(right.Parts) == 0 {
		return nil, &ParseError{
			Message:  "expected value after comparison operator",
			Location: p.current.Location,
		}
	}

	if isEquals {
		return &ast.EqualsCondition{
			Left:  left,
			Right: right,
		}, nil
	}
	return &ast.NotEqualsCondition{
		Left:  left,
		Right: right,
	}, nil
}

// parseConditionValue parses a value in a condition.
// Stops at comparison operators, newline, or EOF.
func (p *Parser) parseConditionValue() *ast.Value {
	loc := ast.SourceLocationFromToken(p.current)
	var parts []ast.ValuePart

	for p.current.Type != lexer.DOUBLE_EQUALS &&
		p.current.Type != lexer.NOT_EQUALS &&
		p.current.Type != lexer.NEWLINE &&
		p.current.Type != lexer.EOF {

		switch p.current.Type {
		case lexer.INTERP_START:
			interp := p.parseInterpolation()
			if interp != nil {
				parts = append(parts, interp)
			}

		case lexer.STRING:
			parts = append(parts, &ast.LiteralValue{Text: p.current.Literal})
			p.nextToken()

		case lexer.IDENTIFIER:
			parts = append(parts, &ast.LiteralValue{Text: p.current.Literal})
			p.nextToken()

		case lexer.PATH:
			parts = append(parts, &ast.LiteralValue{Text: p.current.Literal})
			p.nextToken()

		case lexer.ESCAPE_LBRACE:
			parts = append(parts, &ast.LiteralValue{Text: "{"})
			p.nextToken()

		case lexer.ESCAPE_RBRACE:
			parts = append(parts, &ast.LiteralValue{Text: "}"})
			p.nextToken()

		default:
			// Stop on unexpected tokens
			break
		}
	}

	if len(parts) == 0 {
		return nil
	}

	return &ast.Value{
		Parts:    parts,
		Location: loc,
	}
}

// parseConditionalBody parses statements until end, elif, or else.
func (p *Parser) parseConditionalBody() ([]ast.Statement, *ParseError) {
	var statements []ast.Statement

	for p.current.Type != lexer.END &&
		p.current.Type != lexer.ELIF &&
		p.current.Type != lexer.ELSE &&
		p.current.Type != lexer.EOF {

		stmt, err := p.parseBodyStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	return statements, nil
}

// parseBodyStatement parses a single statement in a conditional body.
func (p *Parser) parseBodyStatement() (ast.Statement, *ParseError) {
	switch p.current.Type {
	case lexer.NEWLINE:
		// Skip blank lines
		p.nextToken()
		return nil, nil

	case lexer.COMMENT:
		comment := p.parseComment()
		p.consumeNewline()
		return comment, nil

	case lexer.LAZY:
		// Lazy variable definition
		v, err := p.ParseVariable()
		if err != nil {
			return nil, err
		}
		// Consume newline if present
		if p.current.Type == lexer.NEWLINE {
			p.nextToken()
		}
		return v, nil

	case lexer.IDENTIFIER:
		// Variable definition
		v, err := p.ParseVariable()
		if err != nil {
			return nil, err
		}
		// Consume newline if present
		if p.current.Type == lexer.NEWLINE {
			p.nextToken()
		}
		return v, nil

	case lexer.IF, lexer.IFDEF, lexer.IFNDEF:
		// Nested conditional
		nested, err := p.ParseConditional()
		if err != nil {
			return nil, err
		}
		return nested, nil

	default:
		// Skip unknown tokens for now
		p.nextToken()
		return nil, nil
	}
}
