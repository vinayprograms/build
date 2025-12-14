package main

import (
	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/lexer"
	"github.com/vinayprograms/build/internal/parser"
)

// ----------------------------------------------------------------------------
// Lexer Adapters
// ----------------------------------------------------------------------------

// tokenAdapter wraps lexer.Token to implement the Token interface.
type tokenAdapter struct {
	tok lexer.Token
}

func (t tokenAdapter) TokenType() string     { return t.tok.Type.String() }
func (t tokenAdapter) TokenLiteral() string  { return t.tok.Literal }
func (t tokenAdapter) TokenLocation() string { return t.tok.Location.String() }
func (t tokenAdapter) IsEOF() bool           { return t.tok.Type == lexer.EOF }
func (t tokenAdapter) IsError() bool         { return t.tok.Type == lexer.ERROR }

// lexerAdapter wraps lexer.Lexer to implement the Lexer interface.
type lexerAdapter struct {
	lex *lexer.Lexer
}

func (l *lexerAdapter) NextToken() Token {
	return tokenAdapter{tok: l.lex.NextToken()}
}

// NewLexer creates a new Lexer for the given source.
// This is the LexerFactory implementation using the internal/lexer package.
func NewLexer(file, input string) Lexer {
	return &lexerAdapter{lex: lexer.New(file, input)}
}

// ----------------------------------------------------------------------------
// Parser Adapters
// ----------------------------------------------------------------------------

// scopeAdapter wraps parser.Scope to implement the Scope interface.
type scopeAdapter struct {
	scope parser.Scope
}

func (s scopeAdapter) String() string { return s.scope.String() }

// Scope constants exposed for external use.
var (
	ScopeGlobal      Scope = scopeAdapter{scope: parser.ScopeGlobal}
	ScopeEnvironment Scope = scopeAdapter{scope: parser.ScopeEnvironment}
	ScopeRecipe      Scope = scopeAdapter{scope: parser.ScopeRecipe}
	ScopeBlock       Scope = scopeAdapter{scope: parser.ScopeBlock}
)

// parserAdapter wraps parser.Parser to implement the Parser interface.
type parserAdapter struct {
	p *parser.Parser
}

func (p *parserAdapter) CurrentScope() Scope {
	return scopeAdapter{scope: p.p.CurrentScope()}
}

func (p *parserAdapter) EnterScope(scope Scope) {
	// Extract the underlying parser.Scope from the adapter
	if sa, ok := scope.(scopeAdapter); ok {
		p.p.EnterScope(sa.scope)
	}
}

func (p *parserAdapter) ExitScope() Scope {
	return scopeAdapter{scope: p.p.ExitScope()}
}

func (p *parserAdapter) CurrentIndentLevel() int {
	return p.p.CurrentIndentLevel()
}

func (p *parserAdapter) HasErrors() bool {
	return p.p.HasErrors()
}

// NewParser creates a new Parser for the given Lexer.
// Note: This function needs access to the underlying lexer.Lexer,
// so it expects a lexerAdapter. For maximum flexibility, it also
// accepts a raw *lexer.Lexer.
func NewParser(lex Lexer) Parser {
	// Extract the underlying lexer.Lexer from the adapter
	if la, ok := lex.(*lexerAdapter); ok {
		return &parserAdapter{p: parser.New(la.lex)}
	}
	// This shouldn't happen in normal use, but handle gracefully
	panic("NewParser requires a Lexer created by NewLexer")
}

// NewParserFromLexer creates a Parser directly from a lexer.Lexer.
// This is useful when you need to work with both interfaces and concrete types.
func NewParserFromLexer(lex *lexer.Lexer) Parser {
	return &parserAdapter{p: parser.New(lex)}
}

// ----------------------------------------------------------------------------
// Directive Validation
// ----------------------------------------------------------------------------

// directiveValidatorImpl implements DirectiveValidator using the parser package.
type directiveValidatorImpl struct{}

func (d directiveValidatorImpl) IsValidAtScope(tokenType string, scope Scope) bool {
	// Convert string token type back to lexer.TokenType
	tokType := tokenTypeFromString(tokenType)
	if tokType == lexer.EOF {
		return false
	}
	// Extract underlying parser.Scope
	if sa, ok := scope.(scopeAdapter); ok {
		return parser.IsDirectiveValidAtScope(tokType, sa.scope)
	}
	return false
}

func (d directiveValidatorImpl) DirectiveName(tokenType string) string {
	tokType := tokenTypeFromString(tokenType)
	return parser.DirectiveNameForError(tokType)
}

// NewDirectiveValidator creates a DirectiveValidator.
func NewDirectiveValidator() DirectiveValidator {
	return directiveValidatorImpl{}
}

// tokenTypeFromString converts a token type string back to lexer.TokenType.
// This is needed for the directive validator interface.
func tokenTypeFromString(s string) lexer.TokenType {
	// Map string names back to token types
	switch s {
	case "DOT_SHELL":
		return lexer.DOT_SHELL
	case "DOT_PARALLEL":
		return lexer.DOT_PARALLEL
	case "DOT_DEFAULT":
		return lexer.DOT_DEFAULT
	case "DOT_INCLUDE":
		return lexer.DOT_INCLUDE
	case "DOT_ENVIRONMENT":
		return lexer.DOT_ENVIRONMENT
	case "DOT_USING":
		return lexer.DOT_USING
	case "DOT_SOURCE":
		return lexer.DOT_SOURCE
	case "DOT_ARGS":
		return lexer.DOT_ARGS
	case "DOT_REQUIRES":
		return lexer.DOT_REQUIRES
	case "DOT_AFTER":
		return lexer.DOT_AFTER
	case "DOT_AUTODEPS":
		return lexer.DOT_AUTODEPS
	default:
		return lexer.EOF
	}
}

// ----------------------------------------------------------------------------
// Variable Adapters
// ----------------------------------------------------------------------------

// valuePartAdapter wraps ast.ValuePart to implement the ValuePart interface.
type valuePartAdapter struct {
	part ast.ValuePart
}

func (v valuePartAdapter) PartType() string {
	switch v.part.(type) {
	case *ast.LiteralValue:
		return "literal"
	case *ast.Interpolation:
		return "interpolation"
	case *ast.FunctionCall:
		return "function"
	default:
		return "unknown"
	}
}

func (v valuePartAdapter) Text() string {
	switch p := v.part.(type) {
	case *ast.LiteralValue:
		return p.Text
	case *ast.Interpolation:
		return p.Name
	case *ast.FunctionCall:
		return p.Name.String()
	default:
		return ""
	}
}

func (v valuePartAdapter) IsRaw() bool {
	if interp, ok := v.part.(*ast.Interpolation); ok {
		return interp.Raw
	}
	return false
}

// variableAdapter wraps ast.Variable to implement the Variable interface.
type variableAdapter struct {
	v *ast.Variable
}

func (va variableAdapter) Name() string {
	return va.v.Name
}

func (va variableAdapter) IsLazy() bool {
	return va.v.Lazy
}

func (va variableAdapter) ValueParts() []ValuePart {
	if va.v.Value == nil {
		return nil
	}
	parts := make([]ValuePart, len(va.v.Value.Parts))
	for i, p := range va.v.Value.Parts {
		parts[i] = valuePartAdapter{part: p}
	}
	return parts
}

func (va variableAdapter) Location() string {
	return va.v.Location.String()
}

// variableParserAdapter wraps parser.Parser to implement VariableParser.
type variableParserAdapter struct {
	p *parser.Parser
}

func (vp *variableParserAdapter) ParseVariable() (Variable, error) {
	v, err := vp.p.ParseVariable()
	if err != nil {
		return nil, err
	}
	return variableAdapter{v: v}, nil
}

// NewVariableParser creates a VariableParser from a Parser.
func NewVariableParser(p Parser) VariableParser {
	if pa, ok := p.(*parserAdapter); ok {
		return &variableParserAdapter{p: pa.p}
	}
	panic("NewVariableParser requires a Parser created by NewParser")
}

// ----------------------------------------------------------------------------
// Target Adapters
// ----------------------------------------------------------------------------

// targetAdapter wraps ast.Target to implement the Target interface.
type targetAdapter struct {
	t *ast.Target
}

func (ta targetAdapter) PatternText() string {
	var text string
	for _, seg := range ta.t.Pattern.Segments {
		switch s := seg.(type) {
		case *ast.LiteralSegment:
			text += s.Text
		case *ast.BraceExpr:
			text += "{" + s.Identifier + "}"
		}
	}
	return text
}

func (ta targetAdapter) IsPhony() bool {
	return ta.t.Pattern.IsPhony
}

func (ta targetAdapter) IsDirectory() bool {
	return ta.t.Pattern.IsDirectory
}

func (ta targetAdapter) DependencyCount() int {
	return len(ta.t.Dependencies)
}

func (ta targetAdapter) DependencyText(i int) string {
	if i < 0 || i >= len(ta.t.Dependencies) {
		return ""
	}
	var text string
	for _, seg := range ta.t.Dependencies[i].Segments {
		switch s := seg.(type) {
		case *ast.LiteralSegment:
			text += s.Text
		case *ast.BraceExpr:
			text += "{" + s.Identifier + "}"
		}
	}
	return text
}

func (ta targetAdapter) HasCaptures() bool {
	for _, seg := range ta.t.Pattern.Segments {
		if _, ok := seg.(*ast.BraceExpr); ok {
			return true
		}
	}
	return false
}

func (ta targetAdapter) CaptureNames() []string {
	var names []string
	for _, seg := range ta.t.Pattern.Segments {
		if be, ok := seg.(*ast.BraceExpr); ok {
			names = append(names, be.Identifier)
		}
	}
	return names
}

func (ta targetAdapter) Location() string {
	return ta.t.Location.String()
}

func (ta targetAdapter) HasRecipe() bool {
	return ta.t.Recipe != nil
}

func (ta targetAdapter) Recipe() Recipe {
	if ta.t.Recipe == nil {
		return nil
	}
	return recipeAdapter{r: ta.t.Recipe}
}

// recipeAdapter wraps ast.Recipe to implement the Recipe interface.
type recipeAdapter struct {
	r *ast.Recipe
}

func (ra recipeAdapter) CommandCount() int {
	return len(ra.r.Commands)
}

func (ra recipeAdapter) CommandText(i int) string {
	if i < 0 || i >= len(ra.r.Commands) {
		return ""
	}
	cmd := ra.r.Commands[i]
	switch c := cmd.(type) {
	case *ast.LineCommand:
		return commandPartsToText(c.Parts)
	case *ast.BlockCommand:
		var lines []string
		for _, line := range c.Lines {
			lines = append(lines, commandPartsToText(line))
		}
		return "block:\n" + joinLines(lines, "        ")
	}
	return ""
}

func (ra recipeAdapter) IsBlockCommand(i int) bool {
	if i < 0 || i >= len(ra.r.Commands) {
		return false
	}
	_, ok := ra.r.Commands[i].(*ast.BlockCommand)
	return ok
}

func (ra recipeAdapter) HasShellDirective() bool {
	return ra.r.Directives.Shell != nil
}

func (ra recipeAdapter) HasAfterDirective() bool {
	return len(ra.r.Directives.After) > 0
}

func (ra recipeAdapter) HasAutodepsDirective() bool {
	return ra.r.Directives.Autodeps != nil
}

func (ra recipeAdapter) RequiresCount() int {
	return len(ra.r.Directives.Requires)
}

func (ra recipeAdapter) RequirementName(i int) string {
	if i < 0 || i >= len(ra.r.Directives.Requires) {
		return ""
	}
	return ra.r.Directives.Requires[i].Name
}

func (ra recipeAdapter) RequirementVersion(i int) string {
	if i < 0 || i >= len(ra.r.Directives.Requires) {
		return ""
	}
	return ra.r.Directives.Requires[i].Version.String()
}

func (ra recipeAdapter) Location() string {
	return ra.r.Location.String()
}

// commandPartsToText converts command parts to text.
func commandPartsToText(parts []ast.CommandPart) string {
	var text string
	for _, part := range parts {
		switch p := part.(type) {
		case *ast.LiteralCommand:
			text += p.Text
		case *ast.CommandInterpolation:
			if p.Raw {
				text += "{" + p.Name + ":raw}"
			} else {
				text += "{" + p.Name + "}"
			}
		}
	}
	return text
}

// joinLines joins lines with a prefix.
func joinLines(lines []string, prefix string) string {
	var result string
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += prefix + line
	}
	return result
}

// targetParserAdapter wraps parser.Parser to implement TargetParser.
type targetParserAdapter struct {
	p *parser.Parser
}

func (tp *targetParserAdapter) ParseTarget() (Target, error) {
	t, err := tp.p.ParseTarget()
	if err != nil {
		return nil, err
	}
	return targetAdapter{t: t}, nil
}

func (tp *targetParserAdapter) IsTargetLine() bool {
	return tp.p.IsTargetLine()
}

// NewTargetParser creates a TargetParser from a Parser.
func NewTargetParser(p Parser) TargetParser {
	if pa, ok := p.(*parserAdapter); ok {
		return &targetParserAdapter{p: pa.p}
	}
	panic("NewTargetParser requires a Parser created by NewParser")
}

// ----------------------------------------------------------------------------
// Environment Adapters
// ----------------------------------------------------------------------------

// environmentAdapter wraps ast.Environment to implement the Environment interface.
type environmentAdapter struct {
	e *ast.Environment
}

func (ea environmentAdapter) Name() string {
	if ea.e.Name == nil {
		return ""
	}
	return *ea.e.Name
}

func (ea environmentAdapter) IsDefault() bool {
	return ea.e.Name == nil
}

func (ea environmentAdapter) RuntimeType() string {
	if ea.e.Runtime == nil {
		return ""
	}
	return ea.e.Runtime.String()
}

func (ea environmentAdapter) HasRuntime() bool {
	return ea.e.Runtime != nil
}

func (ea environmentAdapter) Source() string {
	if ea.e.Source == nil {
		return ""
	}
	return valueToText(ea.e.Source)
}

func (ea environmentAdapter) HasSource() bool {
	return ea.e.Source != nil
}

func (ea environmentAdapter) Args() string {
	if ea.e.Args == nil {
		return ""
	}
	return valueToText(ea.e.Args)
}

func (ea environmentAdapter) HasArgs() bool {
	return ea.e.Args != nil
}

func (ea environmentAdapter) RequiresCount() int {
	return len(ea.e.Requires)
}

func (ea environmentAdapter) RequirementName(i int) string {
	if i < 0 || i >= len(ea.e.Requires) {
		return ""
	}
	return ea.e.Requires[i].Name
}

func (ea environmentAdapter) RequirementVersion(i int) string {
	if i < 0 || i >= len(ea.e.Requires) {
		return ""
	}
	return ea.e.Requires[i].Version.String()
}

func (ea environmentAdapter) Location() string {
	return ea.e.Location.String()
}

// valueToText converts an ast.Value to its text representation.
func valueToText(v *ast.Value) string {
	if v == nil {
		return ""
	}
	var text string
	for _, part := range v.Parts {
		switch p := part.(type) {
		case *ast.LiteralValue:
			text += p.Text
		case *ast.Interpolation:
			if p.Raw {
				text += "{" + p.Name + ":raw}"
			} else {
				text += "{" + p.Name + "}"
			}
		case *ast.FunctionCall:
			text += p.Name.String() + "(...)"
		}
	}
	return text
}

// environmentParserAdapter wraps parser.Parser to implement EnvironmentParser.
type environmentParserAdapter struct {
	p *parser.Parser
}

func (ep *environmentParserAdapter) ParseEnvironment() (Environment, error) {
	e, err := ep.p.ParseEnvironment()
	if err != nil {
		return nil, err
	}
	return environmentAdapter{e: e}, nil
}

// NewEnvironmentParser creates an EnvironmentParser from a Parser.
func NewEnvironmentParser(p Parser) EnvironmentParser {
	if pa, ok := p.(*parserAdapter); ok {
		return &environmentParserAdapter{p: pa.p}
	}
	panic("NewEnvironmentParser requires a Parser created by NewParser")
}
