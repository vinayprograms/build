// Command build is a Make-inspired build tool with readable syntax.
package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vinayprograms/build/internal/ast"
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
	quiet         bool   // Suppress non-error output
	color         string // Color mode: auto, always, never
	progress      string // Progress mode: auto, always, never
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

func Main() {
	// Initialize error package file reader for source context
	initFileReader()
	os.Exit(Run(os.Args[1:]))
}

func Run(args []string) int {
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

	// Parse the buildfile (with caching)
	result, content, err := ParseBuildfileWithCache(buildfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", buildfile, err)
		return exitParseError
	}

	// Report parse errors
	if result.HasErrors() {
		source := content
		for i := 0; i < result.ErrorCount(); i++ {
			e := result.GetError(i)
			formatted := formatParseErrorFromInterface(e, source)
			fmt.Fprint(os.Stderr, formatted.Format())
		}
		return exitParseError
	}

	// Handle --check-env flag
	if f.checkEnv {
		buildfileDir := filepath.Dir(buildfile)
		return checkEnvironment(result, f.env, buildfileDir, f.verbose, f.showInstall)
	}

	// Handle --list-env flag
	if f.listEnv {
		return listEnvironments(result)
	}

	// Resolve target arguments
	resolvedTargets, err := ResolveTargetArgs(targets, result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return exitUsageError
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

		if len(resolvedTargets) > 0 {
			fmt.Printf("Resolved targets: %v\n", resolvedTargets)
		}
	}

	// Run semantic analysis
	collectResult := CollectSymbols(result)
	if collectResult.HasErrors() {
		source := content
		for _, e := range collectResult.Errors() {
			formatted := FormatSemanticError(e, source)
			fmt.Fprint(os.Stderr, formatted.Format())
		}
		return exitParseError
	}

	captureResult := ValidateCaptures(collectResult)
	if captureResult.HasErrors() {
		source := content
		for _, e := range captureResult.Errors() {
			formatted := FormatSemanticError(e, source)
			fmt.Fprint(os.Stderr, formatted.Format())
		}
		return exitParseError
	}

	// Convert Statement interfaces to ast.Statement for reference validation
	astStmts := make([]ast.Statement, len(result.Statements()))
	for i, stmt := range result.Statements() {
		if sa, ok := stmt.(statementAdapter); ok {
			astStmts[i] = sa.s
		}
	}

	refResult := ValidateReferences(collectResult, astStmts, captureResult)
	if refResult.HasErrors() {
		source := content
		for _, e := range refResult.Errors() {
			formatted := FormatSemanticError(e, source)
			fmt.Fprint(os.Stderr, formatted.Format())
		}
		return exitParseError
	}

	depResult := ValidateDependencies(collectResult)
	if depResult.HasErrors() {
		source := content
		for _, e := range depResult.Errors() {
			formatted := FormatSemanticError(e, source)
			fmt.Fprint(os.Stderr, formatted.Format())
		}
		return exitParseError
	}

	// Evaluate variables
	evalResult := EvaluateVariables(result)
	if evalResult.HasErrors() {
		source := content
		for _, e := range evalResult.Errors() {
			formatted := FormatEvaluationError(e, source)
			fmt.Fprint(os.Stderr, formatted.Format())
		}
		return exitParseError
	}

	// Extract targets for planning
	symbolTable := collectResult.SymbolTable()
	sta, ok := symbolTable.(*symbolTableAdapter)
	if !ok {
		fmt.Fprintln(os.Stderr, "error: unexpected symbol table type")
		return exitParseError
	}
	astTargets := sta.st.Targets

	// Plan and execute builds for each target
	ctx := evalResult.Context()
	fs := newRealFileSystem()

	// Get global shell directive
	globalShell := "/bin/sh"
	for _, stmt := range result.Statements() {
		if stmt.StatementType() == "directive" {
			summary := stmt.Summary()
			if len(summary) > 7 && summary[:7] == ".shell:" {
				globalShell = summary[8:] // Skip ".shell: "
				break
			}
		}
	}

	// Configure executor
	shellConfig := NewShellConfig()
	shellConfig.SetShell(globalShell)
	shellConfig.SetDryRun(f.dryRun)
	shellConfig.SetVerbose(f.verbose)
	shellConfig.SetQuiet(f.quiet)

	// Create output reporter based on flags
	reporter := NewNormalReporterWithConfig(os.Stdout, f.verbose, f.quiet, f.color)

	// Determine number of workers for parallel execution
	parallelDirective := GetParallelDirective(result)
	numWorkers := ResolveWorkerCount(f.jobs, parallelDirective)

	// Process each resolved target
	hasFailure := false
	totalTasks := 0
	failedTasks := 0

	for _, target := range resolvedTargets {
		// Plan the build
		planResult := PlanBuild(target, astTargets, ctx, fs)
		if planResult.Error() != nil {
			fmt.Fprintf(os.Stderr, "planning error for %s: %v\n", target, planResult.Error())
			return exitParseError
		}

		if planResult.TaskCount() == 0 {
			reporter.NothingToBuild(target)
			continue
		}

		// Set total for progress display
		if ra, ok := reporter.(*normalReporterAdapter); ok {
			ra.SetTotal(planResult.TaskCount())
		}

		// Create executor
		executor := NewExecutor(shellConfig)

		// Set up emitter for colored output
		emitter := CreateOutputEmitter(f.verbose, f.quiet, f.color)
		executor.SetEmitter(emitter)

		// Use scheduler for parallel execution
		if numWorkers > 1 {
			scheduler := NewScheduler(executor, numWorkers)
			scheduler.SetKeepGoing(f.keep)

			results := scheduler.ExecutePlan(planResult, ctx, reporter)

			for _, r := range results {
				totalTasks++
				if r.Skipped {
					continue
				}

				// Report command output
				for _, er := range r.Results {
					reporter.CommandOutput(er.Command(), er.Stdout(), er.Stderr())
				}

				if r.Error != nil {
					reporter.BuildCompleted(r.Target, false, r.Error.Error())
					hasFailure = true
					failedTasks++
					continue
				}

				// Check for command failures
				cmdFailed := false
				for _, er := range r.Results {
					if er.ExitCode() != 0 {
						errMsg := fmt.Sprintf("command failed with exit code %d: %s", er.ExitCode(), er.Command())
						reporter.BuildCompleted(r.Target, false, errMsg)
						hasFailure = true
						failedTasks++
						cmdFailed = true
						break
					}
				}

				if !cmdFailed {
					reporter.BuildCompleted(r.Target, true, "")
				}
			}
		} else {
			// Sequential execution (original code path)
			for i := 0; i < planResult.TaskCount(); i++ {
				task := planResult.Task(i)
				totalTasks++

				// Announce the build
				reporter.BuildStarted(task.Target())

				// Get the recipe
				recipe := task.Recipe()
				if recipe == nil {
					reporter.BuildCompleted(task.Target(), true, "")
					continue
				}

				// Get recipe shell override if present
				recipeShell := GetRecipeShell(recipe, globalShell, ctx)
				if recipeShell != globalShell {
					executor = NewExecutor(shellConfig.WithOverride(recipeShell))
					executor.SetEmitter(emitter)
				}

				// Set target for event context
				executor.SetTarget(task.Target())

				// Create command context with automatic variables
				cmdCtx := NewCommandContext(
					ctx,
					task.Target(),
					task.Deps(),
				)
				// Set captures if present
				if cca, ok := cmdCtx.(*commandContextAdapter); ok {
					cca.SetCaptures(task.Captures())
				}

				// Execute the recipe (output is printed by executor during execution)
				results, err := executor.ExecuteRecipe(recipe, cmdCtx)

				if err != nil {
					reporter.BuildCompleted(task.Target(), false, err.Error())
					hasFailure = true
					failedTasks++
					break
				}

				// Check for command failures
				cmdFailed := false
				for _, r := range results {
					if r.ExitCode() != 0 {
						errMsg := fmt.Sprintf("command failed with exit code %d: %s", r.ExitCode(), r.Command())
						reporter.BuildCompleted(task.Target(), false, errMsg)
						hasFailure = true
						failedTasks++
						cmdFailed = true
						break
					}
				}

				if cmdFailed {
					break
				}

				reporter.BuildCompleted(task.Target(), true, "")
			}
		}

		if hasFailure && !f.keep {
			break
		}
	}

	// Show summary if there were tasks
	if totalTasks > 0 {
		reporter.Summary(totalTasks, failedTasks)
	}

	if hasFailure {
		return exitBuildFailure
	}

	return exitSuccess
}

func parseFlags(args []string) (*flags, []string, error) {
	f := &flags{
		color:    "auto", // Default value
		progress: "auto", // Default value
	}

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
	fs.BoolVar(&f.quiet, "quiet", false, "Suppress non-error output")
	fs.BoolVar(&f.quiet, "q", false, "Suppress non-error output (shorthand)")
	fs.StringVar(&f.color, "color", "auto", "Color output (auto, always, never)")
	fs.StringVar(&f.progress, "progress", "auto", "Progress output (auto, always, never)")
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
  -q, --quiet          Suppress non-error output
  --color=MODE         Color output: auto, always, never (default: auto)
  --progress=MODE      Progress output: auto, always, never (default: auto)
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
  build                    Build default target
  build @test              Build phony target
  build -n                 Dry run (show what would execute)
  build -v                 Verbose output (show staleness checks)
  build -f other.build     Use alternate file
  build -j4                Build with 4 parallel jobs
  build @clean @build      Build multiple targets in order

Environment Commands:
  build --list-env         List all defined environments
  build --check-env        Check if requirements are satisfied
  build --check-env -e ci  Check named environment "ci"
  build --check-env --show-install
                           Show install commands for missing tools

Example Buildfile:
  .shell: bash

  cc = gcc
  sources = glob(src/*.c)
  objects = replace({sources}, .c, .o)

  build/app: {objects}
      {cc} -o {target} {deps}

  build/{name}.o: src/{name}.c
      {cc} -c {in} -o {out}

  @clean:
      rm -rf build/

  @test: build/app
      ./build/app --test

For more information, see: https://github.com/vinayprograms/build
`)
}

// realFileSystem implements FileSystem using the actual file system.
type realFileSystem struct{}

func newRealFileSystem() FileSystem {
	return &realFileSystem{}
}

func (r *realFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (r *realFileSystem) ModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

func findBuildfile() string {
	candidates := []string{"Buildfile", "buildfile", "Buildfile.build"}

	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Search up the directory tree
	for {
		// Check each candidate in current directory
		for _, name := range candidates {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				// If found in current directory, return just the name
				if dir == mustGetwd() {
					return name
				}
				// Otherwise return the full path
				return path
			}
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return ""
}

// mustGetwd returns the current working directory, panicking on error.
func mustGetwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}
