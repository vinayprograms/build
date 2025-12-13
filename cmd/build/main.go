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

	// Test indentation tracking on each line
	fmt.Println("Indentation analysis:")
	if err := analyzeIndentation(string(content)); err != nil {
		fmt.Fprintf(os.Stderr, "indentation error: %v\n", err)
		return exitParseError
	}

	fmt.Println()

	// Test interpolation detection
	fmt.Println("Interpolation analysis:")
	analyzeInterpolations(string(content))

	return exitSuccess
}

func analyzeIndentation(content string) error {
	// Import would create cycle, so inline the logic for now
	// In future, this will use the actual lexer
	lines := splitLines(content)

	var indentChar byte
	var indentWidth int

	for i, line := range lines {
		if len(line) == 0 {
			fmt.Printf("  line %d: (empty)\n", i+1)
			continue
		}

		// Count leading whitespace
		indent := 0
		for indent < len(line) && (line[indent] == ' ' || line[indent] == '\t') {
			indent++
		}

		if indent == len(line) {
			fmt.Printf("  line %d: (whitespace only)\n", i+1)
			continue
		}

		if indent == 0 {
			fmt.Printf("  line %d: level=0\n", i+1)
			continue
		}

		// First indented line establishes the unit
		if indentChar == 0 {
			indentChar = line[0]
			indentWidth = indent
			charName := "spaces"
			if indentChar == '\t' {
				charName = "tabs"
			}
			fmt.Printf("  line %d: level=1 (established %d %s as unit)\n", i+1, indentWidth, charName)
			continue
		}

		// Check consistency
		for j := 0; j < indent; j++ {
			if line[j] != indentChar {
				return fmt.Errorf("line %d: mixed indentation", i+1)
			}
		}

		if indent%indentWidth != 0 {
			return fmt.Errorf("line %d: indent %d is not multiple of unit %d", i+1, indent, indentWidth)
		}

		level := indent / indentWidth
		fmt.Printf("  line %d: level=%d\n", i+1, level)
	}

	return nil
}

func analyzeInterpolations(content string) {
	// Simple scan for { characters and test boundary rules
	lines := splitLines(content)

	for lineNum, line := range lines {
		for i := 0; i < len(line); i++ {
			if line[i] != '{' {
				continue
			}

			// Determine previous character
			var prev byte
			atSOL := i == 0
			if i > 0 {
				prev = line[i-1]
			}

			// Check for escape
			if i+1 < len(line) && line[i+1] == '{' {
				fmt.Printf("  line %d col %d: {{ (escaped brace)\n", lineNum+1, i+1)
				i++ // skip next {
				continue
			}

			// Check boundary
			isBoundary := atSOL || prev == ' ' || prev == '\t' || prev == ':' || prev == '=' || prev == '/' || prev == '"' || prev == '\'' || prev == '(' || prev == ')' || prev == ',' || prev == '>' || prev == '<'
			if !isBoundary {
				fmt.Printf("  line %d col %d: { not at boundary (prev=%q)\n", lineNum+1, i+1, prev)
				continue
			}

			// Check identifier start
			if i+1 >= len(line) {
				fmt.Printf("  line %d col %d: { at end of line\n", lineNum+1, i+1)
				continue
			}

			next := line[i+1]
			isIdentStart := (next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') || next == '_'
			if !isIdentStart {
				fmt.Printf("  line %d col %d: { not followed by identifier (next=%q)\n", lineNum+1, i+1, next)
				continue
			}

			// Scan identifier
			j := i + 2
			for j < len(line) {
				c := line[j]
				isIdent := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '.'
				if !isIdent {
					break
				}
				j++
			}

			name := line[i+1 : j]
			raw := false

			// Check for modifier or close
			if j < len(line) && line[j] == ':' {
				// Check for :raw
				if j+4 <= len(line) && line[j:j+4] == ":raw" {
					raw = true
					j += 4
				}
			}

			if j < len(line) && line[j] == '}' {
				if raw {
					fmt.Printf("  line %d col %d: {%s:raw} (interpolation with raw)\n", lineNum+1, i+1, name)
				} else {
					fmt.Printf("  line %d col %d: {%s} (interpolation)\n", lineNum+1, i+1, name)
				}
			} else {
				fmt.Printf("  line %d col %d: {%s (unclosed)\n", lineNum+1, i+1, name)
			}
		}
	}
}

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
