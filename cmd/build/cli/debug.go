// Debug commands for the build tool CLI.
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/vinayprograms/build/internal/ast"
)

func debugLexer(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		return exitParseError
	}

	fmt.Printf("=== Lexer Debug: %s ===\n\n", path)
	fmt.Printf("File contents (%d bytes):\n", len(content))
	fmt.Println("---")
	fmt.Print(string(content))
	if len(content) > 0 && content[len(content)-1] != '\n' {
		fmt.Println()
	}
	fmt.Println("---")
	fmt.Println()

	// Tokenize using the lexer interface
	fmt.Println("Tokens:")
	l := NewLexer(path, string(content))
	hasError := false
	for {
		tok := l.NextToken()
		fmt.Printf("  %s %s", tok.TokenLocation(), tok.TokenType())
		if tok.TokenLiteral() != "" && tok.TokenType() != "NEWLINE" && !tok.IsEOF() {
			fmt.Printf(" %q", tok.TokenLiteral())
		}
		fmt.Println()

		if tok.IsError() {
			hasError = true
		}
		if tok.IsEOF() {
			break
		}
	}

	if hasError {
		return exitParseError
	}
	return exitSuccess
}

func debugParser(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		return exitParseError
	}

	fmt.Printf("=== Parser Debug: %s ===\n\n", path)
	fmt.Printf("File contents (%d bytes):\n", len(content))
	fmt.Println("---")
	fmt.Print(string(content))
	if len(content) > 0 && content[len(content)-1] != '\n' {
		fmt.Println()
	}
	fmt.Println("---")
	fmt.Println()

	// Create lexer and parser using interfaces
	l := NewLexer(path, string(content))
	p := NewParser(l)

	fmt.Println("Scope validation test:")
	fmt.Printf("  Initial scope: %s\n", p.CurrentScope())
	fmt.Println()

	// Demonstrate scope transitions
	fmt.Println("Scope transitions:")
	fmt.Printf("  Global → %s (indent level: %d)\n", p.CurrentScope(), p.CurrentIndentLevel())

	p.EnterScope(ScopeEnvironment)
	fmt.Printf("  After entering environment → %s (indent level: %d)\n", p.CurrentScope(), p.CurrentIndentLevel())
	p.ExitScope()
	fmt.Printf("  After exiting environment → %s (indent level: %d)\n", p.CurrentScope(), p.CurrentIndentLevel())

	p.EnterScope(ScopeRecipe)
	fmt.Printf("  After entering recipe → %s (indent level: %d)\n", p.CurrentScope(), p.CurrentIndentLevel())

	p.EnterScope(ScopeBlock)
	fmt.Printf("  After entering block → %s (indent level: %d)\n", p.CurrentScope(), p.CurrentIndentLevel())
	p.ExitScope()
	fmt.Printf("  After exiting block → %s (indent level: %d)\n", p.CurrentScope(), p.CurrentIndentLevel())
	p.ExitScope()
	fmt.Printf("  After exiting recipe → %s (indent level: %d)\n", p.CurrentScope(), p.CurrentIndentLevel())
	fmt.Println()

	// Demonstrate directive validation using the validator interface
	fmt.Println("Directive scope validation:")
	validator := NewDirectiveValidator()
	testDirectives := []struct {
		tokenType string
		scope     Scope
	}{
		{"DOT_SHELL", ScopeGlobal},
		{"DOT_SHELL", ScopeRecipe},
		{"DOT_SHELL", ScopeEnvironment},
		{"DOT_USING", ScopeGlobal},
		{"DOT_USING", ScopeEnvironment},
		{"DOT_AFTER", ScopeGlobal},
		{"DOT_AFTER", ScopeRecipe},
		{"DOT_REQUIRES", ScopeGlobal},
		{"DOT_REQUIRES", ScopeEnvironment},
		{"DOT_REQUIRES", ScopeRecipe},
	}

	for _, td := range testDirectives {
		valid := validator.IsValidAtScope(td.tokenType, td.scope)
		name := validator.DirectiveName(td.tokenType)
		status := "✓"
		if !valid {
			status = "✗"
		}
		fmt.Printf("  %s %s at %s\n", status, name, td.scope)
	}

	return exitSuccess
}

func debugVariables(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		return exitParseError
	}

	fmt.Printf("=== Variable Parsing Debug: %s ===\n\n", path)
	fmt.Printf("File contents (%d bytes):\n", len(content))
	fmt.Println("---")
	fmt.Print(string(content))
	if len(content) > 0 && content[len(content)-1] != '\n' {
		fmt.Println()
	}
	fmt.Println("---")
	fmt.Println()

	// Parse line by line looking for variable definitions
	fmt.Println("Variable definitions found:")
	lines := strings.Split(string(content), "\n")
	hasError := false

	for lineNum, line := range lines {
		// Skip empty lines and comments
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" || trimmed[0] == '#' {
			continue
		}

		// Check if this looks like a variable (has = before :)
		if !looksLikeVariable(trimmed) {
			continue
		}

		// Try to parse as variable (use trimmed line to avoid indent issues)
		l := NewLexer(path, trimmed)
		p := NewParser(l)
		vp := NewVariableParser(p)

		v, err := vp.ParseVariable()
		if err != nil {
			fmt.Printf("  Line %d: ERROR: %v\n", lineNum+1, err)
			hasError = true
			continue
		}

		// Print variable info
		lazyStr := ""
		if v.IsLazy() {
			lazyStr = "lazy "
		}
		fmt.Printf("  Line %d: %s%s = ", lineNum+1, lazyStr, v.Name())

		// Print value parts
		parts := v.ValueParts()
		if len(parts) == 0 {
			fmt.Print("(empty)")
		} else {
			for i, part := range parts {
				if i > 0 {
					fmt.Print(" + ")
				}
				switch part.PartType() {
				case "literal":
					fmt.Printf("%q", part.Text())
				case "interpolation":
					if part.IsRaw() {
						fmt.Printf("{%s:raw}", part.Text())
					} else {
						fmt.Printf("{%s}", part.Text())
					}
				case "function":
					fmt.Printf("%s(...)", part.Text())
				}
			}
		}
		fmt.Println()
	}

	if hasError {
		return exitParseError
	}
	return exitSuccess
}

// looksLikeVariable checks if a line looks like a variable definition.
// A line is a variable if = appears before : (or : doesn't appear).
// Excludes conditionals (if, elif, else, end, ifdef, ifndef) and directives.
func looksLikeVariable(s string) bool {
	// Skip empty lines
	if len(s) == 0 {
		return false
	}
	// Skip directives (start with .)
	if s[0] == '.' {
		return false
	}
	// Skip conditional keywords
	if startsWithKeyword(s, "if ") || startsWithKeyword(s, "elif ") ||
		startsWithKeyword(s, "else") || startsWithKeyword(s, "end") ||
		startsWithKeyword(s, "ifdef ") || startsWithKeyword(s, "ifndef ") {
		return false
	}
	equalsPos := -1
	colonPos := -1
	for i := 0; i < len(s); i++ {
		if s[i] == '=' && equalsPos < 0 {
			// Skip == (comparison operator)
			if i+1 < len(s) && s[i+1] == '=' {
				i++ // skip the second =
				continue
			}
			equalsPos = i
		}
		if s[i] == ':' && colonPos < 0 {
			colonPos = i
		}
	}
	// Has equals, and either no colon or equals comes first
	return equalsPos >= 0 && (colonPos < 0 || equalsPos < colonPos)
}

// startsWithKeyword checks if s starts with the given keyword.
func startsWithKeyword(s, keyword string) bool {
	if len(s) < len(keyword) {
		return false
	}
	return s[:len(keyword)] == keyword
}

// looksLikeTarget checks if a line looks like a target definition.
// A line is a target if : appears before = (or = doesn't appear).
// Excludes directives (starting with .) and indented lines.
func looksLikeTarget(s string) bool {
	// Skip empty lines
	if len(s) == 0 {
		return false
	}
	// Skip directives (start with .)
	if s[0] == '.' {
		return false
	}
	// Phony targets start with @
	if s[0] == '@' {
		return true
	}
	equalsPos := -1
	colonPos := -1
	for i := 0; i < len(s); i++ {
		if s[i] == '=' && equalsPos < 0 {
			equalsPos = i
		}
		if s[i] == ':' && colonPos < 0 {
			colonPos = i
		}
	}
	// Has colon, and either no equals or colon comes first
	return colonPos >= 0 && (equalsPos < 0 || colonPos < equalsPos)
}

func debugTargets(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		return exitParseError
	}

	fmt.Printf("=== Target Parsing Debug: %s ===\n\n", path)
	fmt.Printf("File contents (%d bytes):\n", len(content))
	fmt.Println("---")
	fmt.Print(string(content))
	if len(content) > 0 && content[len(content)-1] != '\n' {
		fmt.Println()
	}
	fmt.Println("---")
	fmt.Println()

	// Parse line by line looking for target definitions
	fmt.Println("Target definitions found:")
	lines := strings.Split(string(content), "\n")
	hasError := false

	for lineNum, line := range lines {
		// Skip indented lines (not at level 0)
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			continue
		}

		// Skip empty lines and comments
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" || trimmed[0] == '#' {
			continue
		}

		// Check if this looks like a target (has : before =)
		if !looksLikeTarget(trimmed) {
			continue
		}

		// Try to parse as target
		l := NewLexer(path, line)
		p := NewParser(l)
		tp := NewTargetParser(p)

		t, err := tp.ParseTarget()
		if err != nil {
			fmt.Printf("  Line %d: ERROR: %v\n", lineNum+1, err)
			hasError = true
			continue
		}

		// Print target info
		typeStr := "file"
		if t.IsPhony() {
			typeStr = "phony"
		} else if t.IsDirectory() {
			typeStr = "dir"
		}

		fmt.Printf("  Line %d: [%s] %s", lineNum+1, typeStr, t.PatternText())

		// Print captures if any
		if t.HasCaptures() {
			fmt.Printf(" (captures: %v)", t.CaptureNames())
		}

		// Print dependencies
		depCount := t.DependencyCount()
		if depCount > 0 {
			fmt.Printf(" ← ")
			for i := 0; i < depCount; i++ {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(t.DependencyText(i))
			}
		}
		fmt.Println()
	}

	if hasError {
		return exitParseError
	}
	return exitSuccess
}

func debugRecipes(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		return exitParseError
	}

	fmt.Printf("=== Recipe Parsing Debug: %s ===\n\n", path)
	fmt.Printf("File contents (%d bytes):\n", len(content))
	fmt.Println("---")
	fmt.Print(string(content))
	if len(content) > 0 && content[len(content)-1] != '\n' {
		fmt.Println()
	}
	fmt.Println("---")
	fmt.Println()

	// Parse targets with recipes
	fmt.Println("Targets with recipes:")
	hasError := false

	// Simple approach: scan for target-like lines and parse them
	// This is a simplified approach - a full parser would parse the entire file
	lines := strings.Split(string(content), "\n")
	lineIdx := 0

	for lineIdx < len(lines) {
		line := lines[lineIdx]
		trimmed := strings.TrimLeft(line, " \t")

		// Skip non-target lines
		if trimmed == "" || trimmed[0] == '#' || !looksLikeTarget(trimmed) {
			lineIdx++
			continue
		}

		// Collect the target line and following indented lines
		targetBlock := line + "\n"
		startLine := lineIdx + 1 // 1-based
		lineIdx++

		for lineIdx < len(lines) {
			nextLine := lines[lineIdx]
			if len(nextLine) > 0 && (nextLine[0] == ' ' || nextLine[0] == '\t') {
				targetBlock += nextLine + "\n"
				lineIdx++
			} else if nextLine == "" {
				// Empty line might be within recipe
				targetBlock += "\n"
				lineIdx++
			} else {
				break
			}
		}

		// Re-create lexer and parser for this block
		l := NewLexer(path, targetBlock)
		p := NewParser(l)
		tp := NewTargetParser(p)

		t, err := tp.ParseTarget()
		if err != nil {
			fmt.Printf("  Line %d: ERROR: %v\n", startLine, err)
			hasError = true
			continue
		}

		// Print target info
		typeStr := "file"
		if t.IsPhony() {
			typeStr = "phony"
		} else if t.IsDirectory() {
			typeStr = "dir"
		}

		fmt.Printf("  Line %d: [%s] %s\n", startLine, typeStr, t.PatternText())

		// Print recipe info
		if !t.HasRecipe() {
			fmt.Println("    (no recipe)")
			continue
		}

		recipe := t.Recipe()
		fmt.Printf("    Recipe at %s:\n", recipe.Location())

		// Print directives
		if recipe.HasShellDirective() {
			fmt.Println("      .shell: (set)")
		}
		if recipe.HasAfterDirective() {
			fmt.Println("      .after: (set)")
		}
		if recipe.HasAutodepsDirective() {
			fmt.Println("      .autodeps: (set)")
		}
		reqCount := recipe.RequiresCount()
		if reqCount > 0 {
			fmt.Printf("      .requires: ")
			for i := 0; i < reqCount; i++ {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s@%s", recipe.RequirementName(i), recipe.RequirementVersion(i))
			}
			fmt.Println()
		}

		// Print commands
		cmdCount := recipe.CommandCount()
		fmt.Printf("      Commands (%d):\n", cmdCount)
		for i := 0; i < cmdCount; i++ {
			if recipe.IsBlockCommand(i) {
				fmt.Printf("        [block] %s\n", recipe.CommandText(i))
			} else {
				fmt.Printf("        [line] %s\n", recipe.CommandText(i))
			}
		}
	}

	if hasError {
		return exitParseError
	}
	return exitSuccess
}

func debugEnvironments(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		return exitParseError
	}

	fmt.Printf("=== Environment Parsing Debug: %s ===\n\n", path)
	fmt.Printf("File contents (%d bytes):\n", len(content))
	fmt.Println("---")
	fmt.Print(string(content))
	if len(content) > 0 && content[len(content)-1] != '\n' {
		fmt.Println()
	}
	fmt.Println("---")
	fmt.Println()

	// Parse environment blocks
	fmt.Println("Environment blocks found:")
	hasError := false
	envCount := 0

	// Scan for .environment: lines
	lines := strings.Split(string(content), "\n")
	lineIdx := 0

	for lineIdx < len(lines) {
		line := lines[lineIdx]
		trimmed := strings.TrimLeft(line, " \t")

		// Look for .environment: directive
		if !looksLikeEnvironment(trimmed) {
			lineIdx++
			continue
		}

		// Collect the environment block
		envBlock := line + "\n"
		startLine := lineIdx + 1 // 1-based
		lineIdx++

		for lineIdx < len(lines) {
			nextLine := lines[lineIdx]
			if len(nextLine) > 0 && (nextLine[0] == ' ' || nextLine[0] == '\t') {
				envBlock += nextLine + "\n"
				lineIdx++
			} else if nextLine == "" {
				// Empty line might be within block
				envBlock += "\n"
				lineIdx++
			} else {
				break
			}
		}

		// Parse environment block
		l := NewLexer(path, envBlock)
		p := NewParser(l)
		ep := NewEnvironmentParser(p)

		e, err := ep.ParseEnvironment()
		if err != nil {
			fmt.Printf("  Line %d: ERROR: %v\n", startLine, err)
			hasError = true
			continue
		}

		envCount++

		// Print environment info
		name := "(default)"
		if !e.IsDefault() {
			name = e.Name()
		}
		fmt.Printf("  Line %d: Environment %s\n", startLine, name)

		// Print runtime
		if e.HasRuntime() {
			fmt.Printf("    .using: %s\n", e.RuntimeType())
		}

		// Print source
		if e.HasSource() {
			fmt.Printf("    .source: %s\n", e.Source())
		}

		// Print args
		if e.HasArgs() {
			fmt.Printf("    .args: %s\n", e.Args())
		}

		// Print requires
		reqCount := e.RequiresCount()
		if reqCount > 0 {
			fmt.Printf("    .requires: ")
			for i := 0; i < reqCount; i++ {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s@%s", e.RequirementName(i), e.RequirementVersion(i))
			}
			fmt.Println()
		}
	}

	if envCount == 0 {
		fmt.Println("  (no environment blocks found)")
	}

	if hasError {
		return exitParseError
	}
	return exitSuccess
}

// looksLikeEnvironment checks if a line looks like an environment block start.
func looksLikeEnvironment(s string) bool {
	return len(s) >= 13 && s[:13] == ".environment:"
}

// looksLikeConditional checks if a line looks like a conditional start.
func looksLikeConditional(s string) bool {
	// Check for if, ifdef, ifndef keywords at start of line
	if len(s) >= 3 && s[:3] == "if " {
		return true
	}
	if len(s) >= 6 && s[:6] == "ifdef " {
		return true
	}
	if len(s) >= 7 && s[:7] == "ifndef " {
		return true
	}
	return false
}

func debugConditionals(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		return exitParseError
	}

	fmt.Printf("=== Conditional Parsing Debug: %s ===\n\n", path)
	fmt.Printf("File contents (%d bytes):\n", len(content))
	fmt.Println("---")
	fmt.Print(string(content))
	if len(content) > 0 && content[len(content)-1] != '\n' {
		fmt.Println()
	}
	fmt.Println("---")
	fmt.Println()

	// Parse conditional blocks
	fmt.Println("Conditional blocks found:")
	hasError := false
	condCount := 0

	// Scan for conditional starts
	lines := strings.Split(string(content), "\n")
	lineIdx := 0

	for lineIdx < len(lines) {
		line := lines[lineIdx]

		// Only consider non-indented lines as potential conditional starts
		// Indented lines are inside recipes/blocks and may contain shell if statements
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			lineIdx++
			continue
		}

		trimmed := strings.TrimLeft(line, " \t")

		// Look for conditional start
		if !looksLikeConditional(trimmed) {
			lineIdx++
			continue
		}

		// Collect the conditional block (up to and including 'end')
		condBlock := line + "\n"
		startLine := lineIdx + 1 // 1-based
		lineIdx++
		depth := 1 // Track nesting for nested conditionals

		for lineIdx < len(lines) && depth > 0 {
			nextLine := lines[lineIdx]
			nextTrimmed := strings.TrimLeft(nextLine, " \t")
			condBlock += nextLine + "\n"

			// Track nested conditionals
			if looksLikeConditional(nextTrimmed) {
				depth++
			} else if nextTrimmed == "end" || (len(nextTrimmed) >= 4 && nextTrimmed[:3] == "end" && (nextTrimmed[3] == ' ' || nextTrimmed[3] == '\n' || nextTrimmed[3] == '\r')) {
				depth--
			}
			lineIdx++
		}

		// Parse conditional block
		l := NewLexer(path, condBlock)
		p := NewParser(l)
		cp := NewConditionalParser(p)

		c, err := cp.ParseConditional()
		if err != nil {
			fmt.Printf("  Line %d: ERROR: %v\n", startLine, err)
			hasError = true
			continue
		}

		condCount++

		// Print conditional info
		condType := c.ConditionType()
		fmt.Printf("  Line %d: ", startLine)

		switch condType {
		case "equals":
			fmt.Printf("if %s == %s\n", c.ConditionLeftText(), c.ConditionRightText())
		case "not_equals":
			fmt.Printf("if %s != %s\n", c.ConditionLeftText(), c.ConditionRightText())
		case "defined":
			fmt.Printf("ifdef %s\n", c.ConditionVarName())
		case "not_defined":
			fmt.Printf("ifndef %s\n", c.ConditionVarName())
		}

		fmt.Printf("    if body: %d statement(s)\n", c.IfBodyCount())

		// Print elif branches
		elifCount := c.ElifCount()
		for i := 0; i < elifCount; i++ {
			elifType := c.ElifConditionType(i)
			switch elifType {
			case "equals":
				fmt.Printf("    elif %s == %s: %d statement(s)\n",
					c.ElifConditionLeftText(i), c.ElifConditionRightText(i), c.ElifBodyCount(i))
			case "not_equals":
				fmt.Printf("    elif %s != %s: %d statement(s)\n",
					c.ElifConditionLeftText(i), c.ElifConditionRightText(i), c.ElifBodyCount(i))
			}
		}

		// Print else
		if c.HasElse() {
			fmt.Printf("    else: %d statement(s)\n", c.ElseBodyCount())
		}
	}

	if condCount == 0 {
		fmt.Println("  (no conditional blocks found)")
	}

	if hasError {
		return exitParseError
	}
	return exitSuccess
}

func debugIncludes(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		return exitParseError
	}

	fmt.Printf("=== Include Parsing Debug: %s ===\n\n", path)
	fmt.Printf("File contents (%d bytes):\n", len(content))
	fmt.Println("---")
	fmt.Print(string(content))
	if len(content) > 0 && content[len(content)-1] != '\n' {
		fmt.Println()
	}
	fmt.Println("---")
	fmt.Println()

	// Parse include directives
	fmt.Println(".include: directives found:")
	hasError := false
	includeCount := 0

	// Scan for include directives
	lines := strings.Split(string(content), "\n")

	for lineIdx, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")

		// Look for .include:
		if len(trimmed) < 9 || trimmed[:9] != ".include:" {
			continue
		}

		// Parse the include directive
		l := NewLexer(path, line+"\n")
		p := NewParser(l)
		ip := NewIncludeParser(p)

		result, err := ip.ParseInclude()
		if err != nil {
			fmt.Printf("  Line %d: ERROR: %v\n", lineIdx+1, err)
			hasError = true
			continue
		}

		includeCount++

		// Print include info
		fmt.Printf("  Line %d: .include: %s\n", lineIdx+1, result.Path())
		fmt.Printf("    Included %d statement(s):\n", result.IncludedStatementCount())

		for i := 0; i < result.IncludedStatementCount(); i++ {
			stmtType := result.IncludedStatementType(i)
			stmtText := result.IncludedStatementText(i)
			fmt.Printf("      - [%s] %s\n", stmtType, stmtText)
		}
	}

	if includeCount == 0 {
		fmt.Println("  (no .include: directives found)")
	}

	if hasError {
		return exitParseError
	}
	return exitSuccess
}

func debugAST(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		return exitParseError
	}

	fmt.Printf("=== Full AST Debug (with error recovery): %s ===\n\n", path)
	fmt.Printf("File contents (%d bytes):\n", len(content))
	fmt.Println("---")
	fmt.Print(string(content))
	if len(content) > 0 && content[len(content)-1] != '\n' {
		fmt.Println()
	}
	fmt.Println("---")
	fmt.Println()

	// Parse complete buildfile with error recovery
	l := NewLexer(path, string(content))
	p := NewParser(l)
	bp := NewBuildfileParser(p)

	result := bp.ParseBuildfile()

	// Print errors (if any)
	if result.HasErrors() {
		fmt.Printf("Parse errors (%d):\n", result.ErrorCount())
		for i := 0; i < result.ErrorCount(); i++ {
			e := result.GetError(i)
			fmt.Printf("  %s\n", e.Error())
			if e.ErrorHint() != "" {
				fmt.Printf("    hint: %s\n", e.ErrorHint())
			}
		}
		fmt.Println()
	} else {
		fmt.Println("No parse errors.")
		fmt.Println()
	}

	// Print statements as tree
	stmts := result.Statements()
	fmt.Printf("AST (%d statements):\n", len(stmts))
	for _, stmt := range stmts {
		printStatementTree(stmt, 0)
	}

	if result.HasErrors() {
		return exitParseError
	}
	return exitSuccess
}

// printStatementTree prints a statement as a tree with indentation.
func printStatementTree(stmt Statement, indent int) {
	// Get the underlying AST node for detailed printing
	if sa, ok := stmt.(interface{ Raw() interface{} }); ok {
		printASTNode(sa.Raw(), indent)
	} else {
		ind := indentStr(indent)
		fmt.Printf("%s%s: %s\n", ind, stmt.StatementType(), stmt.Summary())
	}
}

func indentStr(level int) string {
	s := ""
	for i := 0; i < level; i++ {
		s += "  "
	}
	return s
}

// printASTNode prints an AST node as a tree.
func printASTNode(node interface{}, indent int) {
	ind := indentStr(indent)

	switch n := node.(type) {
	case *ast.Comment:
		// Skip comments in tree view for cleaner output
		return

	case *ast.Blank:
		// Skip blanks
		return

	case *ast.Directive:
		fmt.Printf("%sDirective\n", ind)
		fmt.Printf("%s  kind: %s\n", ind, n.Kind.String())
		fmt.Printf("%s  value:\n", ind)
		printValue(n.Value, indent+2)

	case *ast.Variable:
		fmt.Printf("%sVariable\n", ind)
		fmt.Printf("%s  name: %s\n", ind, n.Name)
		fmt.Printf("%s  lazy: %v\n", ind, n.Lazy)
		fmt.Printf("%s  value:\n", ind)
		printValue(n.Value, indent+2)

	case *ast.Environment:
		fmt.Printf("%sEnvironment\n", ind)
		if n.Name != nil {
			fmt.Printf("%s  name: %s\n", ind, *n.Name)
		} else {
			fmt.Printf("%s  name: (default)\n", ind)
		}
		if n.Runtime != nil {
			fmt.Printf("%s  runtime: %s\n", ind, n.Runtime.String())
		}
		if n.Source != nil {
			fmt.Printf("%s  source:\n", ind)
			printValue(n.Source, indent+2)
		}
		if n.Args != nil {
			fmt.Printf("%s  args:\n", ind)
			printValue(n.Args, indent+2)
		}
		if len(n.Requires) > 0 {
			fmt.Printf("%s  requires:\n", ind)
			for _, req := range n.Requires {
				fmt.Printf("%s    - %s@%s\n", ind, req.Name, req.Version.String())
			}
		}

	case *ast.Conditional:
		fmt.Printf("%sConditional\n", ind)
		fmt.Printf("%s  if:\n", ind)
		printCondition(n.IfBranch.Condition, indent+2)
		fmt.Printf("%s  then: (%d statements)\n", ind, len(n.IfBranch.Body))
		for _, stmt := range n.IfBranch.Body {
			printASTNode(stmt, indent+3)
		}
		for i, elif := range n.ElifBranches {
			fmt.Printf("%s  elif[%d]:\n", ind, i)
			printCondition(elif.Condition, indent+2)
			fmt.Printf("%s  then: (%d statements)\n", ind, len(elif.Body))
			for _, stmt := range elif.Body {
				printASTNode(stmt, indent+3)
			}
		}
		if n.ElseBody != nil {
			fmt.Printf("%s  else: (%d statements)\n", ind, len(n.ElseBody))
			for _, stmt := range n.ElseBody {
				printASTNode(stmt, indent+3)
			}
		}

	case *ast.Target:
		fmt.Printf("%sTarget\n", ind)
		fmt.Printf("%s  pattern: %s\n", ind, patternStr(&n.Pattern))
		if n.Pattern.IsPhony {
			fmt.Printf("%s  phony: true\n", ind)
		}
		if n.Pattern.IsDirectory {
			fmt.Printf("%s  directory: true\n", ind)
		}
		if len(n.Dependencies) > 0 {
			fmt.Printf("%s  dependencies:\n", ind)
			for _, dep := range n.Dependencies {
				fmt.Printf("%s    - %s\n", ind, depStr(&dep))
			}
		}
		if n.Recipe != nil {
			fmt.Printf("%s  recipe:\n", ind)
			printRecipe(n.Recipe, indent+2)
		}

	default:
		fmt.Printf("%s(unknown node type)\n", ind)
	}
}

func printCondition(cond ast.Condition, indent int) {
	ind := indentStr(indent)
	switch c := cond.(type) {
	case *ast.EqualsCondition:
		fmt.Printf("%scondition: == \n", ind)
		fmt.Printf("%s  left:\n", ind)
		printValue(c.Left, indent+2)
		fmt.Printf("%s  right:\n", ind)
		printValue(c.Right, indent+2)
	case *ast.NotEqualsCondition:
		fmt.Printf("%scondition: !=\n", ind)
		fmt.Printf("%s  left:\n", ind)
		printValue(c.Left, indent+2)
		fmt.Printf("%s  right:\n", ind)
		printValue(c.Right, indent+2)
	case *ast.DefinedCondition:
		fmt.Printf("%scondition: ifdef %s\n", ind, c.Name)
	case *ast.NotDefinedCondition:
		fmt.Printf("%scondition: ifndef %s\n", ind, c.Name)
	}
}

func printValue(v *ast.Value, indent int) {
	if v == nil {
		return
	}
	ind := indentStr(indent)
	for _, part := range v.Parts {
		switch p := part.(type) {
		case *ast.LiteralValue:
			fmt.Printf("%sLiteral: %q\n", ind, p.Text)
		case *ast.Interpolation:
			if p.Raw {
				fmt.Printf("%sInterpolation: {%s:raw}\n", ind, p.Name)
			} else {
				fmt.Printf("%sInterpolation: {%s}\n", ind, p.Name)
			}
		case *ast.FunctionCall:
			fmt.Printf("%sFunctionCall: %s(\n", ind, p.Name.String())
			for _, arg := range p.Args {
				printValue(arg, indent+1)
			}
			fmt.Printf("%s)\n", ind)
		}
	}
}

func printRecipe(r *ast.Recipe, indent int) {
	ind := indentStr(indent)
	if r.Directives.Shell != nil {
		fmt.Printf("%s.shell:\n", ind)
		printValue(r.Directives.Shell, indent+1)
	}
	if len(r.Directives.After) > 0 {
		fmt.Printf("%s.after:\n", ind)
		for _, a := range r.Directives.After {
			printValue(a, indent+1)
		}
	}
	if r.Directives.Autodeps != nil {
		fmt.Printf("%s.autodeps:\n", ind)
		printValue(r.Directives.Autodeps, indent+1)
	}
	if len(r.Directives.Requires) > 0 {
		fmt.Printf("%s.requires:\n", ind)
		for _, req := range r.Directives.Requires {
			fmt.Printf("%s  - %s@%s\n", ind, req.Name, req.Version.String())
		}
	}
	fmt.Printf("%scommands: (%d)\n", ind, len(r.Commands))
	for i, cmd := range r.Commands {
		switch c := cmd.(type) {
		case *ast.LineCommand:
			fmt.Printf("%s  [%d] line: ", ind, i)
			printCommandParts(c.Parts)
			fmt.Println()
		case *ast.BlockCommand:
			fmt.Printf("%s  [%d] block:\n", ind, i)
			for j, line := range c.Lines {
				fmt.Printf("%s    [%d] ", ind, j)
				printCommandParts(line)
				fmt.Println()
			}
		}
	}
}

func printCommandParts(parts []ast.CommandPart) {
	for _, part := range parts {
		switch p := part.(type) {
		case *ast.LiteralCommand:
			fmt.Print(p.Text)
		case *ast.CommandInterpolation:
			if p.Raw {
				fmt.Printf("{%s:raw}", p.Name)
			} else {
				fmt.Printf("{%s}", p.Name)
			}
		}
	}
}

func patternStr(p *ast.TargetPattern) string {
	var s string
	for _, seg := range p.Segments {
		switch sg := seg.(type) {
		case *ast.LiteralSegment:
			s += sg.Text
		case *ast.BraceExpr:
			s += "{" + sg.Identifier + "}"
		}
	}
	return s
}

func depStr(d *ast.Dependency) string {
	var s string
	for _, seg := range d.Segments {
		switch sg := seg.(type) {
		case *ast.LiteralSegment:
			s += sg.Text
		case *ast.BraceExpr:
			s += "{" + sg.Identifier + "}"
		}
	}
	return s
}

func debugSemantic(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		return exitParseError
	}

	fmt.Printf("=== Semantic Analysis Debug: %s ===\n\n", path)
	fmt.Printf("File contents (%d bytes):\n", len(content))
	fmt.Println("---")
	fmt.Print(string(content))
	if len(content) > 0 && content[len(content)-1] != '\n' {
		fmt.Println()
	}
	fmt.Println("---")
	fmt.Println()

	// Parse the buildfile first
	l := NewLexer(path, string(content))
	p := NewParser(l)
	bp := NewBuildfileParser(p)
	result := bp.ParseBuildfile()

	// Report parse errors
	if result.HasErrors() {
		fmt.Printf("Parse errors (%d):\n", result.ErrorCount())
		for i := 0; i < result.ErrorCount(); i++ {
			e := result.GetError(i)
			fmt.Printf("  %s\n", e.Error())
		}
		fmt.Println()
	}

	// Symbol Collection (Pass 1)
	fmt.Println("Symbol Collection (Pass 1):")
	fmt.Println()

	collectResult := CollectSymbols(result)
	st := collectResult.SymbolTable()

	// Print variables
	fmt.Println("Variables:")
	if st.VariableCount() == 0 && st.ConditionalVarCount() == 0 {
		fmt.Println("  (none)")
	} else {
		// Print unconditional variables
		for i := 0; i < st.VariableCount(); i++ {
			name := st.VariableName(i)
			loc := st.VariableLocation(i)
			lazy := st.VariableIsLazy(i)
			lazyStr := ""
			if lazy {
				lazyStr = " (lazy)"
			}
			fmt.Printf("  %s%s at %s\n", name, lazyStr, loc)
		}
	}
	fmt.Println()

	// Print conditional variables (variables defined in if/elif/else branches)
	fmt.Println("Conditional Variables:")
	if st.ConditionalVarCount() == 0 {
		fmt.Println("  (none)")
	} else {
		for i := 0; i < st.ConditionalVarCount(); i++ {
			name := st.ConditionalVarName(i)
			defCount := st.ConditionalVarDefCount(i)
			fmt.Printf("  %s (%d definitions):\n", name, defCount)
			for j := 0; j < defCount; j++ {
				loc := st.ConditionalVarDefLocation(i, j)
				branch := st.ConditionalVarDefBranch(i, j)
				fmt.Printf("    [%s] at %s\n", branch, loc)
			}
		}
	}
	fmt.Println()

	// Print targets
	fmt.Println("Targets:")
	if st.TargetCount() == 0 {
		fmt.Println("  (none)")
	} else {
		for i := 0; i < st.TargetCount(); i++ {
			pattern := st.TargetPattern(i)
			loc := st.TargetLocation(i)
			fmt.Printf("  %s at %s\n", pattern, loc)
		}
	}
	fmt.Println()

	// Print environments
	fmt.Println("Environments:")
	if st.EnvironmentCount() == 0 {
		fmt.Println("  (none)")
	} else {
		for i := 0; i < st.EnvironmentCount(); i++ {
			name := st.EnvironmentName(i)
			loc := st.EnvironmentLocation(i)
			if name == "" {
				name = "(default)"
			}
			fmt.Printf("  %s at %s\n", name, loc)
		}
	}
	fmt.Println()

	// Print automatic variables
	fmt.Println("Automatic Variables (built-in):")
	for _, v := range []string{"target", "deps", "in", "out", "stem", "target.dir", "target.file"} {
		fmt.Printf("  %s\n", v)
	}
	fmt.Println()

	// Print built-in variables
	fmt.Println("Built-in Variables:")
	for _, v := range []string{"os", "arch"} {
		fmt.Printf("  %s\n", v)
	}
	fmt.Println()

	// Print semantic errors (if any)
	if collectResult.HasErrors() {
		fmt.Printf("Semantic errors (%d):\n", collectResult.ErrorCount())
		for _, e := range collectResult.Errors() {
			fmt.Printf("  %s\n", e.Error())
		}
		return exitParseError
	}

	fmt.Println("No semantic errors (symbol collection).")
	fmt.Println()

	// Capture Validation (Pass 2)
	fmt.Println("Capture Validation (Pass 2):")
	fmt.Println()

	captureResult := ValidateCaptures(collectResult)

	// Print captures
	fmt.Println("Pattern Targets with Captures:")
	if captureResult.CaptureCount() == 0 {
		fmt.Println("  (none)")
	} else {
		for i := 0; i < captureResult.CaptureCount(); i++ {
			pattern := captureResult.TargetPattern(i)
			names := captureResult.CaptureNames(i)
			fmt.Printf("  %s → captures: %v\n", pattern, names)
		}
	}
	fmt.Println()

	// Print interpolations
	fmt.Println("Targets with Variable Interpolations in Patterns:")
	if captureResult.InterpolationCount() == 0 {
		fmt.Println("  (none)")
	} else {
		for i := 0; i < captureResult.InterpolationCount(); i++ {
			pattern := captureResult.InterpolationTargetPattern(i)
			names := captureResult.InterpolationNames(i)
			fmt.Printf("  %s → interpolations: %v\n", pattern, names)
		}
	}
	fmt.Println()

	// Print capture validation errors (if any)
	if captureResult.HasErrors() {
		fmt.Printf("Capture validation errors (%d):\n", captureResult.ErrorCount())
		for _, e := range captureResult.Errors() {
			fmt.Printf("  %s\n", e.Error())
		}
		return exitParseError
	}

	fmt.Println("No capture validation errors.")
	fmt.Println()

	// Reference Validation (Pass 3)
	fmt.Println("Reference Validation (Pass 3):")
	fmt.Println()

	// Convert Statement interface slice to ast.Statement slice
	astStmts := make([]ast.Statement, len(result.Statements()))
	for i, stmt := range result.Statements() {
		if sa, ok := stmt.(statementAdapter); ok {
			astStmts[i] = sa.s
		}
	}

	refResult := ValidateReferences(collectResult, astStmts, captureResult)

	// Print reference validation errors (if any)
	if refResult.HasErrors() {
		fmt.Printf("Reference validation errors (%d):\n", refResult.ErrorCount())
		for _, e := range refResult.Errors() {
			fmt.Printf("  %s\n", e.Error())
		}
		return exitParseError
	}

	fmt.Println("No reference validation errors.")
	fmt.Println()

	// Dependency Graph Validation (Pass 4)
	fmt.Println("Dependency Graph Validation (Pass 4):")
	fmt.Println()

	depResult := ValidateDependencies(collectResult)

	// Print graph nodes
	fmt.Println("Dependency Graph:")
	if depResult.NodeCount() == 0 {
		fmt.Println("  (no nodes)")
	} else {
		for i := 0; i < depResult.NodeCount(); i++ {
			nodeName := depResult.NodeName(i)
			edgeCount := depResult.NodeEdgeCount(i)
			if edgeCount == 0 {
				fmt.Printf("  %s → (no dependencies)\n", nodeName)
			} else {
				deps := make([]string, edgeCount)
				for j := 0; j < edgeCount; j++ {
					deps[j] = depResult.NodeEdge(i, j)
				}
				fmt.Printf("  %s → %v\n", nodeName, deps)
			}
		}
	}
	fmt.Println()

	// Print pattern targets
	fmt.Println("Pattern Targets (rules, not concrete nodes):")
	if depResult.PatternTargetCount() == 0 {
		fmt.Println("  (none)")
	} else {
		for i := 0; i < depResult.PatternTargetCount(); i++ {
			fmt.Printf("  %s\n", depResult.PatternTargetPattern(i))
		}
	}
	fmt.Println()

	// Print unsatisfied dependencies
	fmt.Println("Unsatisfied Dependencies (may be source files or pattern matches):")
	if depResult.UnsatisfiedDepsCount() == 0 {
		fmt.Println("  (none)")
	} else {
		for i := 0; i < depResult.UnsatisfiedDepsCount(); i++ {
			target := depResult.UnsatisfiedDepsTarget(i)
			deps := depResult.UnsatisfiedDepsList(i)
			fmt.Printf("  %s needs: %v\n", target, deps)
		}
	}
	fmt.Println()

	// Print dependency validation errors (if any)
	if depResult.HasErrors() {
		fmt.Printf("Dependency validation errors (%d):\n", depResult.ErrorCount())
		for _, e := range depResult.Errors() {
			fmt.Printf("  %s\n", e.Error())
		}
		return exitParseError
	}

	fmt.Println("No dependency validation errors.")

	if result.HasErrors() {
		return exitParseError
	}
	return exitSuccess
}

func debugEval(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		return exitParseError
	}

	fmt.Printf("=== Variable Evaluation Debug: %s ===\n\n", path)
	fmt.Printf("File contents (%d bytes):\n", len(content))
	fmt.Println("---")
	fmt.Print(string(content))
	if len(content) > 0 && content[len(content)-1] != '\n' {
		fmt.Println()
	}
	fmt.Println("---")
	fmt.Println()

	// Parse the buildfile first
	l := NewLexer(path, string(content))
	p := NewParser(l)
	bp := NewBuildfileParser(p)
	result := bp.ParseBuildfile()

	// Report parse errors
	if result.HasErrors() {
		fmt.Printf("Parse errors (%d):\n", result.ErrorCount())
		for i := 0; i < result.ErrorCount(); i++ {
			e := result.GetError(i)
			fmt.Printf("  %s\n", e.Error())
		}
		fmt.Println()
		return exitParseError
	}

	// Evaluate variables
	fmt.Println("Variable Evaluation:")
	fmt.Println()

	evalResult := EvaluateVariables(result)

	// Print built-in variables
	fmt.Println("Built-in Variables:")
	ctx := evalResult.Context()
	if osVal, ok := ctx.Get("os"); ok {
		fmt.Printf("  os = %s\n", osVal)
	}
	if archVal, ok := ctx.Get("arch"); ok {
		fmt.Printf("  arch = %s\n", archVal)
	}
	fmt.Println()

	// Print evaluated variables
	fmt.Println("Evaluated Variables:")
	if evalResult.EvaluatedCount() == 0 {
		fmt.Println("  (none)")
	} else {
		for i := 0; i < evalResult.EvaluatedCount(); i++ {
			name := evalResult.EvaluatedName(i)
			value := evalResult.EvaluatedValue(i)
			fmt.Printf("  %s = %q\n", name, value)
		}
	}
	fmt.Println()

	// Print lazy variables
	fmt.Println("Lazy Variables (not yet evaluated):")
	if evalResult.LazyCount() == 0 {
		fmt.Println("  (none)")
	} else {
		for i := 0; i < evalResult.LazyCount(); i++ {
			name := evalResult.LazyName(i)
			fmt.Printf("  %s (lazy)\n", name)
		}
	}
	fmt.Println()

	// Print evaluation errors (if any)
	if evalResult.HasErrors() {
		fmt.Printf("Evaluation errors (%d):\n", evalResult.ErrorCount())
		for _, e := range evalResult.Errors() {
			fmt.Printf("  %s\n", e.Error())
		}
		return exitParseError
	}

	fmt.Println("No evaluation errors.")
	return exitSuccess
}

func debugPlan(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		return exitParseError
	}

	fmt.Printf("=== Build Planning Debug: %s ===\n\n", path)
	fmt.Printf("File contents (%d bytes):\n", len(content))
	fmt.Println("---")
	fmt.Print(string(content))
	if len(content) > 0 && content[len(content)-1] != '\n' {
		fmt.Println()
	}
	fmt.Println("---")
	fmt.Println()

	// Parse the buildfile first
	l := NewLexer(path, string(content))
	p := NewParser(l)
	bp := NewBuildfileParser(p)
	result := bp.ParseBuildfile()

	// Report parse errors
	if result.HasErrors() {
		fmt.Printf("Parse errors (%d):\n", result.ErrorCount())
		for i := 0; i < result.ErrorCount(); i++ {
			e := result.GetError(i)
			fmt.Printf("  %s\n", e.Error())
		}
		fmt.Println()
		return exitParseError
	}

	// Symbol collection (needed to get targets)
	collectResult := CollectSymbols(result)
	if collectResult.HasErrors() {
		fmt.Printf("Semantic errors (%d):\n", collectResult.ErrorCount())
		for _, e := range collectResult.Errors() {
			fmt.Printf("  %s\n", e.Error())
		}
		return exitParseError
	}

	// Get the underlying symbol table
	sta, ok := collectResult.SymbolTable().(*symbolTableAdapter)
	if !ok {
		fmt.Fprintln(os.Stderr, "error: unexpected symbol table type")
		return exitParseError
	}
	targets := sta.st.Targets

	fmt.Println("Target Pattern Matching:")
	fmt.Println()

	// Print all defined targets
	fmt.Println("Defined Targets:")
	if len(targets) == 0 {
		fmt.Println("  (none)")
	} else {
		for i, target := range targets {
			isPattern := false
			for _, seg := range target.Pattern.Segments {
				if _, ok := seg.(*ast.BraceExpr); ok {
					isPattern = true
					break
				}
			}
			patternType := "literal"
			if isPattern {
				patternType = "pattern"
			}
			if target.Pattern.IsPhony {
				patternType = "phony"
			}
			if target.Pattern.IsDirectory {
				patternType = "directory"
			}
			fmt.Printf("  [%d] %s (%s)\n", i, patternStr(&target.Pattern), patternType)
		}
	}
	fmt.Println()

	// Target lookup based on defined patterns
	fmt.Println("Target Lookup Examples (from Buildfile patterns):")
	// Generate test paths based on actual patterns
	for _, target := range targets {
		if target.Pattern.IsPhony {
			testPath := "@" + patternStr(&target.Pattern)
			lookupResult := LookupTargetByPath(testPath, targets)
			if lookupResult.Found() {
				fmt.Printf("  %s → found (phony)\n", testPath)
			}
		}
	}
	fmt.Println()

	// Evaluate variables for dependency resolution
	evalResult := EvaluateVariables(result)
	if evalResult.HasErrors() {
		fmt.Printf("Evaluation errors (%d):\n", evalResult.ErrorCount())
		for _, e := range evalResult.Errors() {
			fmt.Printf("  %s\n", e.Error())
		}
	}
	ctx := evalResult.Context()

	// Dependency resolution examples
	fmt.Println("Dependency Resolution:")
	fmt.Println()

	for _, target := range targets {
		if len(target.Dependencies) == 0 {
			continue
		}
		targetPath := patternStr(&target.Pattern)
		if target.Pattern.IsPhony {
			targetPath = "@" + targetPath
		}

		fmt.Printf("  %s:\n", targetPath)

		// Get captures from target pattern (using example values)
		captures := make(map[string]string)
		for _, seg := range target.Pattern.Segments {
			if be, ok := seg.(*ast.BraceExpr); ok {
				// Use example value for captures
				captures[be.Identifier] = "example"
			}
		}

		// Resolve dependencies
		resolveResult := ResolveDependencyPaths(target.Dependencies, captures, ctx)
		if resolveResult.Error() != nil {
			fmt.Printf("    ERROR: %v\n", resolveResult.Error())
		} else {
			paths := resolveResult.Paths()
			for _, p := range paths {
				fmt.Printf("    → %s\n", p)
			}
		}
	}
	fmt.Println()

	// Command Interpolation examples
	fmt.Println("Command Interpolation (automatic variable resolution):")
	fmt.Println()

	// Find global shell directive
	globalShell := "/bin/sh"
	for _, stmt := range result.Statements() {
		if stmt.StatementType() == "directive" {
			summary := stmt.Summary()
			if len(summary) > 7 && summary[:7] == ".shell:" {
				globalShell = summary[8:] // Skip ".shell: "
			}
		}
	}

	for _, target := range targets {
		if target.Recipe == nil || len(target.Recipe.Commands) == 0 {
			continue
		}

		targetPath := patternStr(&target.Pattern)
		if target.Pattern.IsPhony {
			targetPath = "@" + targetPath
		}

		fmt.Printf("  %s:\n", targetPath)

		// Get captures from target pattern (using example values)
		captures := make(map[string]string)
		for _, seg := range target.Pattern.Segments {
			if be, ok := seg.(*ast.BraceExpr); ok {
				captures[be.Identifier] = "example"
			}
		}

		// Resolve dependencies for command context
		resolveResult := ResolveDependencyPaths(target.Dependencies, captures, ctx)
		var deps []string
		if resolveResult.Error() == nil {
			deps = resolveResult.Paths()
		}

		// Create command context with automatic variables
		cmdCtx := NewCommandContext(ctx, targetPath, deps)
		cmdCtx.SetCaptures(captures)

		// Show shell configuration
		recipeShell := GetRecipeShell(target.Recipe, globalShell, ctx)
		fmt.Printf("    Shell: %s", recipeShell)
		if target.Recipe.Directives.Shell != nil {
			fmt.Printf(" (recipe override)")
		}
		fmt.Println()

		// Show automatic variables
		fmt.Printf("    Automatic variables:\n")
		for _, name := range cmdCtx.AutomaticVarNames() {
			if val, ok := cmdCtx.GetAutomatic(name); ok && val != "" {
				fmt.Printf("      {%s} = %q\n", name, val)
			}
		}
		if len(captures) > 0 {
			fmt.Printf("    Captures:\n")
			for name, val := range captures {
				fmt.Printf("      {%s} = %q\n", name, val)
			}
		}

		// Interpolate each command
		fmt.Printf("    Commands:\n")
		for i, cmd := range target.Recipe.Commands {
			switch c := cmd.(type) {
			case *ast.LineCommand:
				interpResult := InterpolateLineCommand(c, cmdCtx)
				if interpResult.Error() != nil {
					fmt.Printf("      [%d] ERROR: %v\n", i, interpResult.Error())
				} else {
					fmt.Printf("      [%d] %s\n", i, interpResult.Interpolated())
				}
			case *ast.BlockCommand:
				interpResult := InterpolateBlockCommand(c, cmdCtx)
				if interpResult.Error() != nil {
					fmt.Printf("      [%d] block: ERROR: %v\n", i, interpResult.Error())
				} else {
					fmt.Printf("      [%d] block:\n", i)
					// Print each line of the block indented
					lines := splitLines(interpResult.Interpolated())
					for _, line := range lines {
						fmt.Printf("          %s\n", line)
					}
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("Build planning complete.")
	return exitSuccess
}

// splitLines splits a string into lines.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
