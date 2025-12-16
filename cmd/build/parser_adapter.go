package main

import (
	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/lexer"
	"github.com/vinayprograms/build/internal/parser"
)

// ----------------------------------------------------------------------------
// Scope Adapters
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

// ----------------------------------------------------------------------------
// Parser Adapters
// ----------------------------------------------------------------------------

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

// patternToText converts an ast.TargetPattern to its text representation.
func patternToText(p *ast.TargetPattern) string {
	if p == nil {
		return ""
	}
	var text string
	if p.IsPhony {
		text = "@"
	}
	for _, seg := range p.Segments {
		switch s := seg.(type) {
		case *ast.LiteralSegment:
			text += s.Text
		case *ast.BraceExpr:
			text += "{" + s.Identifier + "}"
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

// ----------------------------------------------------------------------------
// Conditional Adapters
// ----------------------------------------------------------------------------

// conditionalAdapter wraps ast.Conditional to implement the Conditional interface.
type conditionalAdapter struct {
	c *ast.Conditional
}

func (ca conditionalAdapter) ConditionType() string {
	return conditionTypeString(ca.c.IfBranch.Condition)
}

func (ca conditionalAdapter) ConditionLeftText() string {
	switch c := ca.c.IfBranch.Condition.(type) {
	case *ast.EqualsCondition:
		return valueToText(c.Left)
	case *ast.NotEqualsCondition:
		return valueToText(c.Left)
	}
	return ""
}

func (ca conditionalAdapter) ConditionRightText() string {
	switch c := ca.c.IfBranch.Condition.(type) {
	case *ast.EqualsCondition:
		return valueToText(c.Right)
	case *ast.NotEqualsCondition:
		return valueToText(c.Right)
	}
	return ""
}

func (ca conditionalAdapter) ConditionVarName() string {
	switch c := ca.c.IfBranch.Condition.(type) {
	case *ast.DefinedCondition:
		return c.Name
	case *ast.NotDefinedCondition:
		return c.Name
	}
	return ""
}

func (ca conditionalAdapter) IfBodyCount() int {
	return len(ca.c.IfBranch.Body)
}

func (ca conditionalAdapter) ElifCount() int {
	return len(ca.c.ElifBranches)
}

func (ca conditionalAdapter) ElifConditionType(i int) string {
	if i < 0 || i >= len(ca.c.ElifBranches) {
		return ""
	}
	return conditionTypeString(ca.c.ElifBranches[i].Condition)
}

func (ca conditionalAdapter) ElifConditionLeftText(i int) string {
	if i < 0 || i >= len(ca.c.ElifBranches) {
		return ""
	}
	switch c := ca.c.ElifBranches[i].Condition.(type) {
	case *ast.EqualsCondition:
		return valueToText(c.Left)
	case *ast.NotEqualsCondition:
		return valueToText(c.Left)
	}
	return ""
}

func (ca conditionalAdapter) ElifConditionRightText(i int) string {
	if i < 0 || i >= len(ca.c.ElifBranches) {
		return ""
	}
	switch c := ca.c.ElifBranches[i].Condition.(type) {
	case *ast.EqualsCondition:
		return valueToText(c.Right)
	case *ast.NotEqualsCondition:
		return valueToText(c.Right)
	}
	return ""
}

func (ca conditionalAdapter) ElifBodyCount(i int) int {
	if i < 0 || i >= len(ca.c.ElifBranches) {
		return 0
	}
	return len(ca.c.ElifBranches[i].Body)
}

func (ca conditionalAdapter) HasElse() bool {
	return ca.c.ElseBody != nil
}

func (ca conditionalAdapter) ElseBodyCount() int {
	if ca.c.ElseBody == nil {
		return 0
	}
	return len(ca.c.ElseBody)
}

func (ca conditionalAdapter) Location() string {
	return ca.c.Location.String()
}

// conditionTypeString returns the type of a condition as a string.
func conditionTypeString(c ast.Condition) string {
	switch c.(type) {
	case *ast.EqualsCondition:
		return "equals"
	case *ast.NotEqualsCondition:
		return "not_equals"
	case *ast.DefinedCondition:
		return "defined"
	case *ast.NotDefinedCondition:
		return "not_defined"
	default:
		return "unknown"
	}
}

// conditionalParserAdapter wraps parser.Parser to implement ConditionalParser.
type conditionalParserAdapter struct {
	p *parser.Parser
}

func (cp *conditionalParserAdapter) ParseConditional() (Conditional, error) {
	c, err := cp.p.ParseConditional()
	if err != nil {
		return nil, err
	}
	return conditionalAdapter{c: c}, nil
}

func (cp *conditionalParserAdapter) IsConditionalLine() bool {
	return cp.p.IsConditionalLine()
}

// NewConditionalParser creates a ConditionalParser from a Parser.
func NewConditionalParser(p Parser) ConditionalParser {
	if pa, ok := p.(*parserAdapter); ok {
		return &conditionalParserAdapter{p: pa.p}
	}
	panic("NewConditionalParser requires a Parser created by NewParser")
}

// ----------------------------------------------------------------------------
// Include Adapters
// ----------------------------------------------------------------------------

// includeResultAdapter wraps ast.Directive and statements to implement IncludeResult.
type includeResultAdapter struct {
	directive  *ast.Directive
	statements []ast.Statement
}

func (ia includeResultAdapter) DirectiveKind() string {
	return ia.directive.Kind.String()
}

func (ia includeResultAdapter) Path() string {
	return valueToText(ia.directive.Value)
}

func (ia includeResultAdapter) IncludedStatementCount() int {
	return len(ia.statements)
}

func (ia includeResultAdapter) IncludedStatementType(i int) string {
	if i < 0 || i >= len(ia.statements) {
		return ""
	}
	switch ia.statements[i].(type) {
	case *ast.Variable:
		return "variable"
	case *ast.Target:
		return "target"
	case *ast.Directive:
		return "directive"
	case *ast.Environment:
		return "environment"
	case *ast.Conditional:
		return "conditional"
	case *ast.Comment:
		return "comment"
	case *ast.Blank:
		return "blank"
	default:
		return "unknown"
	}
}

func (ia includeResultAdapter) IncludedStatementText(i int) string {
	if i < 0 || i >= len(ia.statements) {
		return ""
	}
	switch s := ia.statements[i].(type) {
	case *ast.Variable:
		lazyStr := ""
		if s.Lazy {
			lazyStr = "lazy "
		}
		return lazyStr + s.Name + " = " + valueToText(s.Value)
	case *ast.Target:
		return patternToText(&s.Pattern)
	case *ast.Directive:
		return "." + s.Kind.String() + ": " + valueToText(s.Value)
	case *ast.Environment:
		if s.Name != nil {
			return ".environment: " + *s.Name
		}
		return ".environment:"
	case *ast.Conditional:
		return "if ..."
	case *ast.Comment:
		return s.Text
	default:
		return ""
	}
}

func (ia includeResultAdapter) Location() string {
	return ia.directive.Location.String()
}

// includeParserAdapter wraps parser.Parser to implement IncludeParser.
type includeParserAdapter struct {
	p *parser.Parser
}

func (ip *includeParserAdapter) ParseInclude() (IncludeResult, error) {
	directive, statements, err := ip.p.ParseInclude()
	if err != nil {
		return nil, err
	}
	return includeResultAdapter{
		directive:  directive,
		statements: statements,
	}, nil
}

func (ip *includeParserAdapter) IsIncludeLine() bool {
	return ip.p.CurrentToken().Type == lexer.DOT_INCLUDE
}

// NewIncludeParser creates an IncludeParser from a Parser.
func NewIncludeParser(p Parser) IncludeParser {
	if pa, ok := p.(*parserAdapter); ok {
		return &includeParserAdapter{p: pa.p}
	}
	panic("NewIncludeParser requires a Parser created by NewParser")
}

// ----------------------------------------------------------------------------
// Buildfile Parser Adapters
// ----------------------------------------------------------------------------

// statementAdapter wraps ast.Statement to implement the Statement interface.
type statementAdapter struct {
	s ast.Statement
}

// Raw returns the underlying AST node.
func (sa statementAdapter) Raw() interface{} {
	return sa.s
}

func (sa statementAdapter) StatementType() string {
	switch sa.s.(type) {
	case *ast.Directive:
		return "directive"
	case *ast.Environment:
		return "environment"
	case *ast.Variable:
		return "variable"
	case *ast.Conditional:
		return "conditional"
	case *ast.Target:
		return "target"
	case *ast.Comment:
		return "comment"
	case *ast.Blank:
		return "blank"
	default:
		return "unknown"
	}
}

func (sa statementAdapter) Location() string {
	switch s := sa.s.(type) {
	case *ast.Directive:
		return s.Location.String()
	case *ast.Environment:
		return s.Location.String()
	case *ast.Variable:
		return s.Location.String()
	case *ast.Conditional:
		return s.Location.String()
	case *ast.Target:
		return s.Location.String()
	case *ast.Comment:
		return s.Location.String()
	case *ast.Blank:
		return s.Location.String()
	default:
		return ""
	}
}

func (sa statementAdapter) Summary() string {
	switch s := sa.s.(type) {
	case *ast.Directive:
		return "." + s.Kind.String() + ": " + valueToText(s.Value)
	case *ast.Environment:
		if s.Name != nil {
			return ".environment: " + *s.Name
		}
		return ".environment: (default)"
	case *ast.Variable:
		prefix := ""
		if s.Lazy {
			prefix = "lazy "
		}
		return prefix + s.Name + " = " + valueToText(s.Value)
	case *ast.Conditional:
		return conditionalSummary(s)
	case *ast.Target:
		return targetSummary(s)
	case *ast.Comment:
		text := s.Text
		if len(text) > 60 {
			text = text[:57] + "..."
		}
		return text
	case *ast.Blank:
		return "(blank)"
	default:
		return ""
	}
}

// conditionalSummary returns a brief summary of a conditional.
func conditionalSummary(c *ast.Conditional) string {
	switch cond := c.IfBranch.Condition.(type) {
	case *ast.EqualsCondition:
		return "if " + valueToText(cond.Left) + " == " + valueToText(cond.Right)
	case *ast.NotEqualsCondition:
		return "if " + valueToText(cond.Left) + " != " + valueToText(cond.Right)
	case *ast.DefinedCondition:
		return "ifdef " + cond.Name
	case *ast.NotDefinedCondition:
		return "ifndef " + cond.Name
	default:
		return "if ..."
	}
}

// targetSummary returns a brief summary of a target.
func targetSummary(t *ast.Target) string {
	pattern := patternToText(&t.Pattern)
	deps := ""
	if len(t.Dependencies) > 0 {
		depTexts := make([]string, len(t.Dependencies))
		for i, dep := range t.Dependencies {
			depTexts[i] = dependencyToText(&dep)
		}
		deps = " " + joinStrings(depTexts, " ")
	}
	recipe := ""
	if t.Recipe != nil {
		cmdCount := len(t.Recipe.Commands)
		if cmdCount == 1 {
			recipe = " (1 command)"
		} else if cmdCount > 1 {
			recipe = " (" + itoa(cmdCount) + " commands)"
		}
	}
	return pattern + ":" + deps + recipe
}

// dependencyToText converts a dependency to text.
func dependencyToText(d *ast.Dependency) string {
	var text string
	for _, seg := range d.Segments {
		switch s := seg.(type) {
		case *ast.LiteralSegment:
			text += s.Text
		case *ast.BraceExpr:
			text += "{" + s.Identifier + "}"
		}
	}
	return text
}

// joinStrings joins strings with a separator.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// itoa converts int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// parseErrorAdapter wraps parser.ParseError to implement the ParseError interface.
type parseErrorAdapter struct {
	e *parser.ParseError
}

func (pe parseErrorAdapter) Error() string {
	return pe.e.Error()
}

func (pe parseErrorAdapter) ErrorLocation() string {
	return pe.e.Location.String()
}

func (pe parseErrorAdapter) ErrorHint() string {
	return pe.e.Hint
}

// buildfileResultAdapter wraps the result of ParseBuildfile.
type buildfileResultAdapter struct {
	statements []ast.Statement
	errors     *parser.ParseErrors
}

func (br buildfileResultAdapter) Statements() []Statement {
	result := make([]Statement, len(br.statements))
	for i, s := range br.statements {
		result[i] = statementAdapter{s: s}
	}
	return result
}

func (br buildfileResultAdapter) ErrorCount() int {
	return len(br.errors.Errors)
}

func (br buildfileResultAdapter) GetError(i int) ParseError {
	if i < 0 || i >= len(br.errors.Errors) {
		return nil
	}
	return parseErrorAdapter{e: br.errors.Errors[i]}
}

func (br buildfileResultAdapter) HasErrors() bool {
	return br.errors.HasErrors()
}

func (br buildfileResultAdapter) AllErrors() string {
	return br.errors.Error()
}

// buildfileParserAdapter wraps parser.Parser to implement BuildfileParser.
type buildfileParserAdapter struct {
	p *parser.Parser
}

func (bp *buildfileParserAdapter) ParseBuildfile() BuildfileResult {
	stmts, errs := bp.p.ParseBuildfile()
	return buildfileResultAdapter{
		statements: stmts,
		errors:     errs,
	}
}

// NewBuildfileParser creates a BuildfileParser from a Parser.
func NewBuildfileParser(p Parser) BuildfileParser {
	if pa, ok := p.(*parserAdapter); ok {
		return &buildfileParserAdapter{p: pa.p}
	}
	panic("NewBuildfileParser requires a Parser created by NewParser")
}
