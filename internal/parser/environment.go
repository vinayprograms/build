package parser

import (
	"strings"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/lexer"
)

// ParseEnvironment parses an environment block.
// Grammar: environment_block = ".environment:" [ identifier ] NEWLINE
//
//	INDENT { env_directive NEWLINE } DEDENT ;
//
// env_directive = ".using:" value | ".source:" value | ".args:" value | ".requires:" value ;
func (p *Parser) ParseEnvironment() (*ast.Environment, *ParseError) {
	// Expect .environment: token
	if p.current.Type != lexer.DOT_ENVIRONMENT {
		return nil, &ParseError{
			Message:  "expected '.environment:' directive",
			Location: p.current.Location,
		}
	}

	loc := ast.SourceLocationFromToken(p.current)
	p.nextToken() // consume .environment

	// Expect colon
	if p.current.Type != lexer.COLON {
		return nil, &ParseError{
			Message:  "expected ':' after .environment",
			Location: p.current.Location,
		}
	}
	p.nextToken() // consume :

	env := &ast.Environment{
		Location: loc,
	}

	// Check for optional name
	if p.current.Type == lexer.IDENTIFIER || p.current.Type == lexer.STRING {
		name := strings.TrimSpace(p.current.Literal)
		if name != "" {
			env.Name = &name
		}
		p.nextToken()
	}

	// Consume newline
	if p.current.Type == lexer.NEWLINE {
		p.nextToken()
	}

	// Enter environment scope
	p.EnterScope(ScopeEnvironment)
	defer p.ExitScope()

	// Parse environment directives until we dedent
	for {
		// Check for end conditions
		if p.current.Type == lexer.EOF {
			break
		}

		// At the start of each line, we should have an INDENT token
		if p.current.Type != lexer.INDENT {
			// Not indented - end of environment block
			break
		}

		// Check indent level
		indentLevel := p.calculateIndentLevel(p.current.Literal)
		if indentLevel == 0 {
			// Back to global scope - end of environment block
			break
		}
		p.nextToken() // consume INDENT

		// Empty line
		if p.current.Type == lexer.NEWLINE {
			p.nextToken()
			continue
		}

		// Comment line
		if p.current.Type == lexer.COMMENT {
			p.nextToken()
			if p.current.Type == lexer.NEWLINE {
				p.nextToken()
			}
			continue
		}

		// Parse environment directives
		if p.current.Type.IsDotKeyword() {
			if err := p.parseEnvironmentDirective(env); err != nil {
				return nil, err
			}
			continue
		}

		// Unexpected token
		return nil, &ParseError{
			Message:  "unexpected token in environment block",
			Location: p.current.Location,
		}
	}

	return env, nil
}

// parseEnvironmentDirective parses a directive within an environment block.
func (p *Parser) parseEnvironmentDirective(env *ast.Environment) *ParseError {
	// Validate directive is allowed at environment scope
	if err := p.validateDirectiveScope(p.current); err != nil {
		return err
	}

	switch p.current.Type {
	case lexer.DOT_USING:
		return p.parseEnvUsing(env)
	case lexer.DOT_SOURCE:
		return p.parseEnvSource(env)
	case lexer.DOT_ARGS:
		return p.parseEnvArgs(env)
	case lexer.DOT_REQUIRES:
		return p.parseEnvRequires(env)
	default:
		// Invalid directive for environment scope
		return &ParseError{
			Message:  "invalid directive in environment block: " + p.current.Literal,
			Location: p.current.Location,
		}
	}
}

// parseEnvUsing parses .using: directive in environment.
func (p *Parser) parseEnvUsing(env *ast.Environment) *ParseError {
	if err := p.expectColon(".using"); err != nil {
		return err
	}

	// Parse runtime type
	runtime, err := p.parseRuntimeType()
	if err != nil {
		return err
	}
	env.Runtime = &runtime

	p.consumeNewline()
	return nil
}

// parseRuntimeType parses a runtime type name.
func (p *Parser) parseRuntimeType() (ast.Runtime, *ParseError) {
	// Get the runtime name
	var name string
	if p.current.Type == lexer.IDENTIFIER || p.current.Type == lexer.STRING {
		name = strings.TrimSpace(p.current.Literal)
		p.nextToken()
	} else {
		return ast.RuntimeBare, &ParseError{
			Message:  "expected runtime type (bare, docker, podman, devcontainer, nix, lima)",
			Location: p.current.Location,
		}
	}

	switch strings.ToLower(name) {
	case "bare":
		return ast.RuntimeBare, nil
	case "docker":
		return ast.RuntimeDocker, nil
	case "podman":
		return ast.RuntimePodman, nil
	case "devcontainer":
		return ast.RuntimeDevcontainer, nil
	case "nix":
		return ast.RuntimeNix, nil
	case "lima":
		return ast.RuntimeLima, nil
	default:
		return ast.RuntimeBare, &ParseError{
			Message:  "unknown runtime type: " + name,
			Location: p.current.Location,
			Hint:     "valid runtime types are: bare, docker, podman, devcontainer, nix, lima",
		}
	}
}

// parseEnvSource parses .source: directive in environment.
func (p *Parser) parseEnvSource(env *ast.Environment) *ParseError {
	if err := p.expectColon(".source"); err != nil {
		return err
	}

	env.Source = p.ParseValue()
	p.consumeNewline()
	return nil
}

// parseEnvArgs parses .args: directive in environment.
func (p *Parser) parseEnvArgs(env *ast.Environment) *ParseError {
	if err := p.expectColon(".args"); err != nil {
		return err
	}

	env.Args = p.ParseValue()
	p.consumeNewline()
	return nil
}

// parseEnvRequires parses .requires: directive in environment.
func (p *Parser) parseEnvRequires(env *ast.Environment) *ParseError {
	if err := p.expectColon(".requires"); err != nil {
		return err
	}

	reqs, err := p.parseRequirementsList()
	if err != nil {
		return err
	}
	env.Requires = append(env.Requires, reqs...)

	p.consumeNewline()
	return nil
}
