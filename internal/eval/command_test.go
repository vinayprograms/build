package eval

import (
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

// ----------------------------------------------------------------------------
// CommandContext Tests
// ----------------------------------------------------------------------------

func TestNewCommandContext(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "build/app", []string{"src/main.c", "src/utils.c"})

	if cmdCtx == nil {
		t.Fatal("NewCommandContext returned nil")
	}

	// Verify automatic variables are set
	tests := []struct {
		name     string
		expected string
	}{
		{"target", "build/app"},
		{"out", "build/app"},
		{"deps", "src/main.c src/utils.c"},
		{"in", "src/main.c"},
		{"target.dir", "build"},
		{"target.file", "app"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := cmdCtx.Get(tt.name)
			if !ok {
				t.Errorf("automatic variable '%s' not defined", tt.name)
				return
			}
			if val != tt.expected {
				t.Errorf("expected '%s' = '%s', got '%s'", tt.name, tt.expected, val)
			}
		})
	}
}

func TestCommandContext_NoDependencies(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "output.txt", nil)

	// deps and in should be empty for targets with no dependencies
	if val, ok := cmdCtx.Get("deps"); ok && val != "" {
		t.Errorf("expected empty deps, got '%s'", val)
	}

	if val, ok := cmdCtx.Get("in"); ok && val != "" {
		t.Errorf("expected empty in, got '%s'", val)
	}
}

func TestCommandContext_PhonyTarget(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "@clean", nil)

	// For phony targets, target.dir should be "." and target.file should be name without @
	tests := []struct {
		name     string
		expected string
	}{
		{"target", "@clean"},
		{"out", "@clean"},
		{"target.dir", "."},
		{"target.file", "clean"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := cmdCtx.Get(tt.name)
			if !ok {
				t.Errorf("automatic variable '%s' not defined", tt.name)
				return
			}
			if val != tt.expected {
				t.Errorf("expected '%s' = '%s', got '%s'", tt.name, tt.expected, val)
			}
		})
	}
}

func TestCommandContext_WithStem(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "build/utils.o", []string{"src/utils.c"})
	cmdCtx.SetStem("utils")

	val, ok := cmdCtx.Get("stem")
	if !ok {
		t.Fatal("stem not defined after SetStem")
	}
	if val != "utils" {
		t.Errorf("expected stem = 'utils', got '%s'", val)
	}
}

func TestCommandContext_WithCaptures(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "build/utils.o", []string{"src/utils.c"})
	cmdCtx.SetCaptures(map[string]string{
		"name": "utils",
		"ext":  "c",
	})

	// Captures should be accessible
	tests := []struct {
		name     string
		expected string
	}{
		{"name", "utils"},
		{"ext", "c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := cmdCtx.GetCapture(tt.name)
			if !ok {
				t.Errorf("capture '%s' not defined", tt.name)
				return
			}
			if val != tt.expected {
				t.Errorf("expected '%s' = '%s', got '%s'", tt.name, tt.expected, val)
			}
		})
	}
}

func TestCommandContext_DirectoryTarget(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "build/output/", []string{})

	// target.dir should be the directory itself (without trailing slash)
	// target.file should be empty for directory targets
	if val, _ := cmdCtx.Get("target.dir"); val != "build/output" {
		t.Errorf("expected target.dir = 'build/output', got '%s'", val)
	}

	if val, _ := cmdCtx.Get("target.file"); val != "" {
		t.Errorf("expected empty target.file for directory target, got '%s'", val)
	}
}

func TestCommandContext_RootTarget(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "app", []string{"main.c"})

	// For files in root directory
	if val, _ := cmdCtx.Get("target.dir"); val != "." {
		t.Errorf("expected target.dir = '.', got '%s'", val)
	}

	if val, _ := cmdCtx.Get("target.file"); val != "app" {
		t.Errorf("expected target.file = 'app', got '%s'", val)
	}
}

func TestCommandContext_InheritsVariables(t *testing.T) {
	ctx := NewContext()
	ctx.Set("CC", "gcc")
	ctx.Set("CFLAGS", "-Wall")

	cmdCtx := NewCommandContext(ctx, "build/app", []string{"main.c"})

	// Should be able to access parent context variables
	if val, ok := cmdCtx.Get("CC"); !ok || val != "gcc" {
		t.Errorf("expected CC = 'gcc', got '%s'", val)
	}

	if val, ok := cmdCtx.Get("CFLAGS"); !ok || val != "-Wall" {
		t.Errorf("expected CFLAGS = '-Wall', got '%s'", val)
	}
}

// ----------------------------------------------------------------------------
// Command Interpolation Tests
// ----------------------------------------------------------------------------

func TestInterpolateCommand_LiteralOnly(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "build/app", []string{"main.c"})

	cmd := &ast.LineCommand{
		Parts: []ast.CommandPart{
			&ast.LiteralCommand{Text: "echo hello"},
		},
	}

	result, err := InterpolateCommand(cmd, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "echo hello" {
		t.Errorf("expected 'echo hello', got '%s'", result)
	}
}

func TestInterpolateCommand_AutomaticVariables(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "build/app", []string{"main.c", "utils.c"})

	cmd := &ast.LineCommand{
		Parts: []ast.CommandPart{
			&ast.LiteralCommand{Text: "gcc -o "},
			&ast.CommandInterpolation{Name: "target", Raw: false},
			&ast.LiteralCommand{Text: " "},
			&ast.CommandInterpolation{Name: "deps", Raw: false},
		},
	}

	result, err := InterpolateCommand(cmd, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// deps is pre-quoted, target/deps without special chars don't get quoted
	expected := "gcc -o build/app main.c utils.c"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestInterpolateCommand_RawModifier(t *testing.T) {
	ctx := NewContext()
	ctx.Set("FLAGS", "-Wall -O2")
	cmdCtx := NewCommandContext(ctx, "build/app", []string{"main.c"})

	cmd := &ast.LineCommand{
		Parts: []ast.CommandPart{
			&ast.LiteralCommand{Text: "gcc "},
			&ast.CommandInterpolation{Name: "FLAGS", Raw: true},
			&ast.LiteralCommand{Text: " -o "},
			&ast.CommandInterpolation{Name: "target", Raw: false},
		},
	}

	result, err := InterpolateCommand(cmd, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Raw modifier should NOT quote the value; target has no special chars
	expected := "gcc -Wall -O2 -o build/app"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestInterpolateCommand_CaptureVariables(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "build/utils.o", []string{"src/utils.c"})
	cmdCtx.SetCaptures(map[string]string{"name": "utils"})

	cmd := &ast.LineCommand{
		Parts: []ast.CommandPart{
			&ast.LiteralCommand{Text: "gcc -c src/"},
			&ast.CommandInterpolation{Name: "name", Raw: false},
			&ast.LiteralCommand{Text: ".c -o build/"},
			&ast.CommandInterpolation{Name: "name", Raw: false},
			&ast.LiteralCommand{Text: ".o"},
		},
	}

	result, err := InterpolateCommand(cmd, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "gcc -c src/utils.c -o build/utils.o"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestInterpolateCommand_StemVariable(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "build/test.o", []string{"src/test.c"})
	cmdCtx.SetStem("test")

	cmd := &ast.LineCommand{
		Parts: []ast.CommandPart{
			&ast.LiteralCommand{Text: "echo Building "},
			&ast.CommandInterpolation{Name: "stem", Raw: false},
		},
	}

	result, err := InterpolateCommand(cmd, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "echo Building test"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestInterpolateCommand_UserVariables(t *testing.T) {
	ctx := NewContext()
	ctx.Set("CC", "gcc")
	ctx.Set("CFLAGS", "-Wall -O2")
	cmdCtx := NewCommandContext(ctx, "build/app", []string{"main.c"})

	cmd := &ast.LineCommand{
		Parts: []ast.CommandPart{
			&ast.CommandInterpolation{Name: "CC", Raw: true},
			&ast.LiteralCommand{Text: " "},
			&ast.CommandInterpolation{Name: "CFLAGS", Raw: true},
			&ast.LiteralCommand{Text: " -o "},
			&ast.CommandInterpolation{Name: "target", Raw: false},
			&ast.LiteralCommand{Text: " "},
			&ast.CommandInterpolation{Name: "in", Raw: false},
		},
	}

	result, err := InterpolateCommand(cmd, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "gcc -Wall -O2 -o build/app main.c"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestInterpolateCommand_UndefinedVariable(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "build/app", []string{"main.c"})

	cmd := &ast.LineCommand{
		Parts: []ast.CommandPart{
			&ast.LiteralCommand{Text: "echo "},
			&ast.CommandInterpolation{Name: "UNDEFINED", Raw: false},
		},
	}

	_, err := InterpolateCommand(cmd, cmdCtx)
	if err == nil {
		t.Fatal("expected error for undefined variable")
	}
}

func TestInterpolateCommand_BuiltinVariables(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "build/app", []string{})

	cmd := &ast.LineCommand{
		Parts: []ast.CommandPart{
			&ast.LiteralCommand{Text: "echo "},
			&ast.CommandInterpolation{Name: "os", Raw: true},
			&ast.LiteralCommand{Text: "-"},
			&ast.CommandInterpolation{Name: "arch", Raw: true},
		},
	}

	result, err := InterpolateCommand(cmd, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain os and arch values (platform-specific)
	if result == "" {
		t.Error("expected non-empty result with os and arch")
	}
}

func TestInterpolateCommand_TargetDirAndFile(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "build/sub/app.exe", []string{})

	cmd := &ast.LineCommand{
		Parts: []ast.CommandPart{
			&ast.LiteralCommand{Text: "mkdir -p "},
			&ast.CommandInterpolation{Name: "target.dir", Raw: false},
			&ast.LiteralCommand{Text: " && echo "},
			&ast.CommandInterpolation{Name: "target.file", Raw: false},
		},
	}

	result, err := InterpolateCommand(cmd, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "mkdir -p build/sub && echo app.exe"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

// ----------------------------------------------------------------------------
// Shell Quoting Tests
// ----------------------------------------------------------------------------

func TestShellQuote_Simple(t *testing.T) {
	// Simple alphanumeric strings don't need quoting
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},               // no special chars - not quoted
		{"path/to/file", "path/to/file"}, // slashes are fine - not quoted
		{"", ""},                         // empty string - not quoted
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ShellQuote(tt.input)
			if result != tt.expected {
				t.Errorf("ShellQuote(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestShellQuote_SpecialCharacters(t *testing.T) {
	// Strings with special characters need quoting
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "'hello world'"},           // spaces - quoted as whole
		{"file.c file.h", "'file.c file.h'"},       // multiple items - quoted as whole
		{"$HOME", "'$HOME'"},                       // dollar sign (no expansion)
		{"a*b", "'a*b'"},                           // glob
		{"test(1)", "'test(1)'"},                   // parens
		{"a>b", "'a>b'"},                           // redirect
		{"a&b", "'a&b'"},                           // background
		{"a|b", "'a|b'"},                           // pipe
		{`hello"world`, `'hello"world'`},           // double quote in string
		{"it's", `'it'"'"'s'`},                     // single quote requires escaping
		{"it's a \"test\"", `'it'"'"'s a "test"'`}, // mixed quotes
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ShellQuote(tt.input)
			if result != tt.expected {
				t.Errorf("ShellQuote(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Block Command Interpolation Tests
// ----------------------------------------------------------------------------

func TestInterpolateBlockCommand(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "build/app", []string{"main.c"})

	block := &ast.BlockCommand{
		Lines: [][]ast.CommandPart{
			{
				&ast.LiteralCommand{Text: "if [[ -f "},
				&ast.CommandInterpolation{Name: "target", Raw: false},
				&ast.LiteralCommand{Text: " ]]; then"},
			},
			{
				&ast.LiteralCommand{Text: "    rm "},
				&ast.CommandInterpolation{Name: "target", Raw: false},
			},
			{
				&ast.LiteralCommand{Text: "fi"},
			},
		},
	}

	result, err := InterpolateBlockCommand(block, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "if [[ -f build/app ]]; then\n    rm build/app\nfi"
	if result != expected {
		t.Errorf("expected:\n%s\n\ngot:\n%s", expected, result)
	}
}
