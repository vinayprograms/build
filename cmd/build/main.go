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
	file          string
	env           string
	jobs          int
	dryRun        bool
	verbose       bool
	checkEnv      bool
	showInstall   bool
	listEnv       bool
	shell         bool
	keep          bool
	showVersion   bool
	showHelp      bool
	debugLex      bool // Debug: dump lexer tokens
	debugParse    bool // Debug: dump parser scope validation
	debugVar      bool // Debug: dump variable parsing
	debugTarget   bool // Debug: dump target parsing
	debugRecipe   bool // Debug: dump recipe parsing
	debugEnv      bool // Debug: dump environment parsing
	debugCond     bool // Debug: dump conditional parsing
	debugIncl     bool // Debug: dump include parsing
	debugAST      bool // Debug: dump full AST with error recovery
	debugSemantic bool // Debug: dump semantic analysis (symbol table)
	debugEval     bool // Debug: dump variable evaluation
	debugPlan     bool // Debug: dump build planning (target matching)
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

	// Debug mode: dump recipe parsing
	if f.debugRecipe {
		return debugRecipes(buildfile)
	}

	// Debug mode: dump environment parsing
	if f.debugEnv {
		return debugEnvironments(buildfile)
	}

	// Debug mode: dump conditional parsing
	if f.debugCond {
		return debugConditionals(buildfile)
	}

	// Debug mode: dump include parsing
	if f.debugIncl {
		return debugIncludes(buildfile)
	}

	// Debug mode: dump full AST with error recovery
	if f.debugAST {
		return debugAST(buildfile)
	}

	// Debug mode: dump semantic analysis (symbol table)
	if f.debugSemantic {
		return debugSemantic(buildfile)
	}

	// Debug mode: dump variable evaluation
	if f.debugEval {
		return debugEval(buildfile)
	}

	// Debug mode: dump build planning (target matching)
	if f.debugPlan {
		return debugPlan(buildfile)
	}

	// Parse the buildfile
	content, err := os.ReadFile(buildfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", buildfile, err)
		return exitParseError
	}

	l := NewLexer(buildfile, string(content))
	p := NewParser(l)
	bp := NewBuildfileParser(p)
	result := bp.ParseBuildfile()

	// Report parse errors
	if result.HasErrors() {
		fmt.Fprintf(os.Stderr, "parse errors in %s:\n", buildfile)
		for i := 0; i < result.ErrorCount(); i++ {
			e := result.GetError(i)
			fmt.Fprintf(os.Stderr, "  %s\n", e.Error())
		}
		return exitParseError
	}

	// Show what was parsed (verbose mode)
	if f.verbose {
		fmt.Printf("Buildfile: %s\n", buildfile)
		fmt.Printf("Parsed %d statements\n", len(result.Statements()))

		// Count statement types
		counts := make(map[string]int)
		for _, stmt := range result.Statements() {
			counts[stmt.StatementType()]++
		}

		for stmtType, count := range counts {
			fmt.Printf("  %s: %d\n", stmtType, count)
		}

		if len(targets) > 0 {
			fmt.Printf("Requested targets: %v\n", targets)
		}
	}

	// TODO: Semantic analysis, build planning, and execution not yet implemented
	fmt.Println("build: semantic analysis and execution not yet implemented")
	fmt.Println("use --debug-ast to see parsed AST")

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
	fs.BoolVar(&f.debugRecipe, "debug-recipe", false, "Debug: dump recipe parsing")
	fs.BoolVar(&f.debugEnv, "debug-env", false, "Debug: dump environment parsing")
	fs.BoolVar(&f.debugCond, "debug-cond", false, "Debug: dump conditional parsing")
	fs.BoolVar(&f.debugIncl, "debug-include", false, "Debug: dump include parsing")
	fs.BoolVar(&f.debugAST, "debug-ast", false, "Debug: dump full AST with error recovery")
	fs.BoolVar(&f.debugSemantic, "debug-semantic", false, "Debug: dump semantic analysis")
	fs.BoolVar(&f.debugEval, "debug-eval", false, "Debug: dump variable evaluation")
	fs.BoolVar(&f.debugPlan, "debug-plan", false, "Debug: dump build planning (target matching)")

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
  --debug-recipe       Dump recipe parsing (for development)
  --debug-env          Dump environment parsing (for development)
  --debug-cond         Dump conditional parsing (for development)
  --debug-include      Dump include parsing (for development)
  --debug-ast          Dump full AST with error recovery (for development)
  --debug-semantic     Dump semantic analysis (for development)
  --debug-eval         Dump variable evaluation (for development)
  --debug-plan         Dump build planning / target matching (for development)

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
