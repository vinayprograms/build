// Command build is a Make-inspired build tool with readable syntax.
package main

import (
	"flag"
	"fmt"
	"os"
)

// Version information (set at build time via -ldflags).
var (
	version = "dev"
	commit  = "unknown"
)

// Exit codes per spec.
const (
	exitSuccess      = 0
	exitBuildFailure = 1
	exitUsageError   = 2
	exitParseError   = 3
	exitEnvError     = 4
)

// CLI flags.
type flags struct {
	file        string
	env         string
	jobs        int
	dryRun      bool
	verbose     bool
	checkEnv    bool
	showInstall bool
	listEnv     bool
	shell       bool
	keep        bool
	showVersion bool
	showHelp    bool
	debugLex    bool // Debug: dump lexer tokens
	debugParse  bool // Debug: dump parser scope validation
	debugVar    bool // Debug: dump variable parsing
	debugTarget bool // Debug: dump target parsing
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	f, targets, err := parseFlags(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return exitUsageError
	}

	if f.showHelp {
		printUsage()
		return exitSuccess
	}

	if f.showVersion {
		fmt.Printf("build %s (%s)\n", version, commit)
		return exitSuccess
	}

	// Find Buildfile
	buildfile := f.file
	if buildfile == "" {
		buildfile = findBuildfile()
	}

	if buildfile == "" {
		fmt.Fprintln(os.Stderr, "error: no Buildfile found")
		return exitParseError
	}

	// Debug mode: dump lexer tokens
	if f.debugLex {
		return debugLexer(buildfile)
	}

	// Debug mode: dump parser scope validation
	if f.debugParse {
		return debugParser(buildfile)
	}

	// Debug mode: dump variable parsing
	if f.debugVar {
		return debugVariables(buildfile)
	}

	// Debug mode: dump target parsing
	if f.debugTarget {
		return debugTargets(buildfile)
	}

	// For now, just print what we would do
	if f.verbose {
		fmt.Printf("Buildfile: %s\n", buildfile)
		fmt.Printf("Targets: %v\n", targets)
	}

	fmt.Println("build: lexer and parser not yet fully implemented")
	fmt.Println("use --debug-lex to test lexer components")

	return exitSuccess
}

func parseFlags(args []string) (*flags, []string, error) {
	f := &flags{}

	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	fs.StringVar(&f.file, "file", "", "Use alternate Buildfile")
	fs.StringVar(&f.file, "f", "", "Use alternate Buildfile (shorthand)")
	fs.StringVar(&f.env, "env", "", "Use named environment")
	fs.StringVar(&f.env, "e", "", "Use named environment (shorthand)")
	fs.IntVar(&f.jobs, "jobs", 1, "Parallel jobs")
	fs.IntVar(&f.jobs, "j", 1, "Parallel jobs (shorthand)")
	fs.BoolVar(&f.dryRun, "dry-run", false, "Show what would execute")
	fs.BoolVar(&f.dryRun, "n", false, "Show what would execute (shorthand)")
	fs.BoolVar(&f.verbose, "verbose", false, "Verbose output")
	fs.BoolVar(&f.verbose, "v", false, "Verbose output (shorthand)")
	fs.BoolVar(&f.checkEnv, "check-env", false, "Verify environment requirements")
	fs.BoolVar(&f.showInstall, "show-install", false, "Show install instructions")
	fs.BoolVar(&f.listEnv, "list-env", false, "List available environments")
	fs.BoolVar(&f.shell, "shell", false, "Open shell in environment")
	fs.BoolVar(&f.keep, "keep", false, "Keep sandbox running after build")
	fs.BoolVar(&f.showVersion, "version", false, "Show version")
	fs.BoolVar(&f.showVersion, "V", false, "Show version (shorthand)")
	fs.BoolVar(&f.showHelp, "help", false, "Show help")
	fs.BoolVar(&f.showHelp, "h", false, "Show help (shorthand)")
	fs.BoolVar(&f.debugLex, "debug-lex", false, "Debug: dump lexer tokens")
	fs.BoolVar(&f.debugParse, "debug-parse", false, "Debug: dump parser scope validation")
	fs.BoolVar(&f.debugVar, "debug-var", false, "Debug: dump variable parsing")
	fs.BoolVar(&f.debugTarget, "debug-target", false, "Debug: dump target parsing")

	if err := fs.Parse(args); err != nil {
		return nil, nil, err
	}

	return f, fs.Args(), nil
}

func printUsage() {
	fmt.Print(`Usage: build [options] [targets...]

A Make-inspired build tool with readable syntax.

Options:
  -f, --file PATH      Use alternate Buildfile
  -e, --env NAME       Use named environment
  -j, --jobs N         Parallel jobs (default: 1)
  -n, --dry-run        Show what would execute without running
  -v, --verbose        Verbose output
  --check-env          Verify environment requirements
  --show-install       Show install instructions for missing requirements
  --list-env           List available environments
  --shell              Open shell in sandbox environment
  --keep               Keep sandbox running after build
  -V, --version        Show version
  -h, --help           Show this help

Debug Options:
  --debug-lex          Dump lexer tokens (for development)
  --debug-parse        Dump parser scope validation (for development)
  --debug-var          Dump variable parsing (for development)
  --debug-target       Dump target parsing (for development)

Examples:
  build                Build default target
  build @test          Build phony target
  build -n             Dry run
  build -f other.build Use alternate file
`)
}

func findBuildfile() string {
	candidates := []string{"Buildfile", "buildfile", "Buildfile.build"}
	for _, name := range candidates {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}
	return ""
}

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
	lines := splitLines(string(content))
	hasError := false

	for lineNum, line := range lines {
		// Skip empty lines and comments
		trimmed := trimLeadingWhitespace(line)
		if trimmed == "" || trimmed[0] == '#' {
			continue
		}

		// Check if this looks like a variable (has = before :)
		if !looksLikeVariable(trimmed) {
			continue
		}

		// Try to parse as variable
		l := NewLexer(path, line)
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

// splitLines splits a string into lines.
func splitLines(s string) []string {
	var lines []string
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

// trimLeadingWhitespace removes leading spaces and tabs.
func trimLeadingWhitespace(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return s[i:]
		}
	}
	return ""
}

// looksLikeVariable checks if a line looks like a variable definition.
// A line is a variable if = appears before : (or : doesn't appear).
func looksLikeVariable(s string) bool {
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
	// Has equals, and either no colon or equals comes first
	return equalsPos >= 0 && (colonPos < 0 || equalsPos < colonPos)
}

// looksLikeTarget checks if a line looks like a target definition.
// A line is a target if : appears before = (or = doesn't appear).
func looksLikeTarget(s string) bool {
	// Phony targets start with @
	if len(s) > 0 && s[0] == '@' {
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
	lines := splitLines(string(content))
	hasError := false

	for lineNum, line := range lines {
		// Skip empty lines and comments
		trimmed := trimLeadingWhitespace(line)
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
