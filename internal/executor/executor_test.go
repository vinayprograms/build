package executor

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/eval"
)

// ----------------------------------------------------------------------------
// Shell Configuration Tests
// ----------------------------------------------------------------------------

func TestNewShellConfig_Default(t *testing.T) {
	cfg := NewShellConfig()

	// Default shell should be /bin/sh
	if cfg.Shell != "/bin/sh" {
		t.Errorf("expected default shell '/bin/sh', got '%s'", cfg.Shell)
	}
}

func TestShellConfig_FromGlobalDirective(t *testing.T) {
	cfg := NewShellConfig()
	cfg.SetShell("bash")

	if cfg.Shell != "bash" {
		t.Errorf("expected shell 'bash', got '%s'", cfg.Shell)
	}
}

func TestShellConfig_FromRecipeOverride(t *testing.T) {
	globalCfg := NewShellConfig()
	globalCfg.SetShell("bash")

	// Recipe-level override
	recipeCfg := globalCfg.WithOverride("zsh")

	if recipeCfg.Shell != "zsh" {
		t.Errorf("expected recipe shell 'zsh', got '%s'", recipeCfg.Shell)
	}

	// Global should not be affected
	if globalCfg.Shell != "bash" {
		t.Errorf("global shell should remain 'bash', got '%s'", globalCfg.Shell)
	}
}

// ----------------------------------------------------------------------------
// Line Mode Execution Tests
// ----------------------------------------------------------------------------

func TestExecuteLine_Simple(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)

	result, err := exec.ExecuteLine("echo hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "hello\n"
	if result.Stdout != expected {
		t.Errorf("expected stdout '%s', got '%s'", expected, result.Stdout)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestExecuteLine_WithVariables(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)

	// Use shell variable expansion (not our interpolation)
	result, err := exec.ExecuteLine("echo $HOME")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// $HOME should be expanded by the shell
	if result.Stdout == "$HOME\n" || result.Stdout == "" {
		t.Error("shell should have expanded $HOME")
	}
}

func TestExecuteLine_Failure(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)

	result, err := exec.ExecuteLine("exit 42")
	if err == nil {
		t.Fatal("expected error for failed command")
	}

	if result.ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", result.ExitCode)
	}
}

func TestExecuteLine_Stderr(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)

	result, err := exec.ExecuteLine("echo error >&2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Stderr, "error") {
		t.Errorf("expected stderr to contain 'error', got '%s'", result.Stderr)
	}
}

func TestExecuteLine_BashSpecific(t *testing.T) {
	// Skip if bash is not available
	if _, err := os.Stat("/bin/bash"); os.IsNotExist(err) {
		t.Skip("bash not available")
	}

	cfg := NewShellConfig()
	cfg.SetShell("/bin/bash")
	exec := NewExecutor(cfg)

	// Use bash-specific syntax (arrays)
	result, err := exec.ExecuteLine("arr=(a b c); echo ${arr[1]}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(result.Stdout) != "b" {
		t.Errorf("expected 'b', got '%s'", result.Stdout)
	}
}

// ----------------------------------------------------------------------------
// Block Mode Execution Tests
// ----------------------------------------------------------------------------

func TestExecuteBlock_Simple(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)

	script := `echo line1
echo line2`

	result, err := exec.ExecuteBlock(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "line1\nline2\n"
	if result.Stdout != expected {
		t.Errorf("expected stdout '%s', got '%s'", expected, result.Stdout)
	}
}

func TestExecuteBlock_WithIfStatement(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)

	script := `if [ 1 -eq 1 ]; then
    echo yes
else
    echo no
fi`

	result, err := exec.ExecuteBlock(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(result.Stdout) != "yes" {
		t.Errorf("expected 'yes', got '%s'", result.Stdout)
	}
}

func TestExecuteBlock_WithLoop(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)

	script := `for i in 1 2 3; do
    echo $i
done`

	result, err := exec.ExecuteBlock(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "1\n2\n3\n"
	if result.Stdout != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Stdout)
	}
}

func TestExecuteBlock_FailsOnError(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)

	script := `echo before
exit 1
echo after`

	result, err := exec.ExecuteBlock(script)
	if err == nil {
		t.Fatal("expected error for failed command")
	}

	// Should have before but not after
	if !strings.Contains(result.Stdout, "before") {
		t.Error("expected 'before' in stdout")
	}
	if strings.Contains(result.Stdout, "after") {
		t.Error("should not have 'after' in stdout")
	}
}

// ----------------------------------------------------------------------------
// Recipe Execution Tests
// ----------------------------------------------------------------------------

func TestExecuteRecipe_SingleLineCommand(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)

	ctx := eval.NewContext()
	cmdCtx := eval.NewCommandContext(ctx, "build/app", []string{"main.c"})

	recipe := &ast.Recipe{
		Commands: []ast.Command{
			&ast.LineCommand{
				Parts: []ast.CommandPart{
					&ast.LiteralCommand{Text: "echo building "},
					&ast.CommandInterpolation{Name: "target", Raw: true},
				},
			},
		},
	}

	results, err := exec.ExecuteRecipe(recipe, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if strings.TrimSpace(results[0].Stdout) != "building build/app" {
		t.Errorf("expected 'building build/app', got '%s'", results[0].Stdout)
	}
}

func TestExecuteRecipe_MultipleLineCommands(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)

	ctx := eval.NewContext()
	cmdCtx := eval.NewCommandContext(ctx, "build/app", []string{"main.c"})

	recipe := &ast.Recipe{
		Commands: []ast.Command{
			&ast.LineCommand{
				Parts: []ast.CommandPart{
					&ast.LiteralCommand{Text: "echo first"},
				},
			},
			&ast.LineCommand{
				Parts: []ast.CommandPart{
					&ast.LiteralCommand{Text: "echo second"},
				},
			},
		},
	}

	results, err := exec.ExecuteRecipe(recipe, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if strings.TrimSpace(results[0].Stdout) != "first" {
		t.Errorf("first command: expected 'first', got '%s'", results[0].Stdout)
	}

	if strings.TrimSpace(results[1].Stdout) != "second" {
		t.Errorf("second command: expected 'second', got '%s'", results[1].Stdout)
	}
}

func TestExecuteRecipe_BlockCommand(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)

	ctx := eval.NewContext()
	cmdCtx := eval.NewCommandContext(ctx, "build/app", []string{})

	recipe := &ast.Recipe{
		Commands: []ast.Command{
			&ast.BlockCommand{
				Lines: [][]ast.CommandPart{
					{&ast.LiteralCommand{Text: "echo start"}},
					{&ast.LiteralCommand{Text: "echo end"}},
				},
			},
		},
	}

	results, err := exec.ExecuteRecipe(recipe, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result (block is single execution), got %d", len(results))
	}

	expected := "start\nend\n"
	if results[0].Stdout != expected {
		t.Errorf("expected '%s', got '%s'", expected, results[0].Stdout)
	}
}

func TestExecuteRecipe_StopsOnFirstError(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)

	ctx := eval.NewContext()
	cmdCtx := eval.NewCommandContext(ctx, "build/app", []string{})

	recipe := &ast.Recipe{
		Commands: []ast.Command{
			&ast.LineCommand{
				Parts: []ast.CommandPart{
					&ast.LiteralCommand{Text: "echo first"},
				},
			},
			&ast.LineCommand{
				Parts: []ast.CommandPart{
					&ast.LiteralCommand{Text: "exit 1"},
				},
			},
			&ast.LineCommand{
				Parts: []ast.CommandPart{
					&ast.LiteralCommand{Text: "echo third"},
				},
			},
		},
	}

	results, err := exec.ExecuteRecipe(recipe, cmdCtx)
	if err == nil {
		t.Fatal("expected error for failed command")
	}

	// Should have executed first two commands, not the third
	if len(results) != 2 {
		t.Errorf("expected 2 results (stopped at failure), got %d", len(results))
	}
}

func TestExecuteRecipe_ShellOverride(t *testing.T) {
	// Skip if bash is not available
	if _, err := os.Stat("/bin/bash"); os.IsNotExist(err) {
		t.Skip("bash not available")
	}

	cfg := NewShellConfig()
	exec := NewExecutor(cfg)

	ctx := eval.NewContext()
	cmdCtx := eval.NewCommandContext(ctx, "build/app", []string{})

	shell := "bash"
	recipe := &ast.Recipe{
		Directives: ast.RecipeDirectives{
			Shell: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.LiteralValue{Text: shell},
				},
			},
		},
		Commands: []ast.Command{
			&ast.LineCommand{
				Parts: []ast.CommandPart{
					// Use bash-specific syntax
					&ast.LiteralCommand{Text: "echo ${BASH_VERSION:0:1}"},
				},
			},
		},
	}

	results, err := exec.ExecuteRecipe(recipe, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have used bash (output would be different with sh)
	if results[0].Stdout == "" {
		t.Error("expected non-empty output from bash version")
	}
}

// ----------------------------------------------------------------------------
// Dry Run Tests
// ----------------------------------------------------------------------------

func TestDryRun_PrintsCommands(t *testing.T) {
	cfg := NewShellConfig()
	cfg.DryRun = true
	exec := NewExecutor(cfg)

	var output bytes.Buffer
	exec.SetOutput(&output)

	ctx := eval.NewContext()
	cmdCtx := eval.NewCommandContext(ctx, "build/app", []string{"main.c"})

	recipe := &ast.Recipe{
		Commands: []ast.Command{
			&ast.LineCommand{
				Parts: []ast.CommandPart{
					&ast.LiteralCommand{Text: "gcc -o "},
					&ast.CommandInterpolation{Name: "target", Raw: true},
					&ast.LiteralCommand{Text: " "},
					&ast.CommandInterpolation{Name: "in", Raw: true},
				},
			},
		},
	}

	_, err := exec.ExecuteRecipe(recipe, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should print the command, not execute it
	out := output.String()
	if !strings.Contains(out, "gcc -o build/app main.c") {
		t.Errorf("expected dry-run output to contain command, got: %s", out)
	}
}

func TestDryRun_DoesNotExecute(t *testing.T) {
	cfg := NewShellConfig()
	cfg.DryRun = true
	exec := NewExecutor(cfg)

	// This command would fail if executed
	result, err := exec.ExecuteLine("exit 1")
	if err != nil {
		t.Fatalf("dry-run should not return error: %v", err)
	}

	// Exit code should be 0 in dry-run mode
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0 in dry-run, got %d", result.ExitCode)
	}
}

func TestDryRun_WouldBuildPrefix(t *testing.T) {
	cfg := NewShellConfig()
	cfg.DryRun = true
	exec := NewExecutor(cfg)

	var output bytes.Buffer
	exec.SetOutput(&output)

	ctx := eval.NewContext()
	cmdCtx := eval.NewCommandContext(ctx, "build/app", []string{"main.c"})

	recipe := &ast.Recipe{
		Commands: []ast.Command{
			&ast.LineCommand{
				Parts: []ast.CommandPart{
					&ast.LiteralCommand{Text: "gcc -o app main.c"},
				},
			},
		},
	}

	_, err := exec.ExecuteRecipe(recipe, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := output.String()
	// Should have "Would build: target" prefix before command
	if !strings.Contains(out, "Would build: build/app") {
		t.Errorf("expected dry-run output to have 'Would build: build/app', got: %s", out)
	}
}

// ----------------------------------------------------------------------------
// Verbose Mode Tests
// ----------------------------------------------------------------------------

func TestVerbose_PrintsCommand(t *testing.T) {
	cfg := NewShellConfig()
	cfg.Verbose = true
	exec := NewExecutor(cfg)

	var output bytes.Buffer
	exec.SetOutput(&output)

	_, err := exec.ExecuteLine("echo hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should print the command before executing
	out := output.String()
	if !strings.Contains(out, "echo hello") {
		t.Errorf("expected verbose output to contain command, got: %s", out)
	}
}

// ----------------------------------------------------------------------------
// Shell Validation Tests
// ----------------------------------------------------------------------------

func TestValidateShell_ValidShell(t *testing.T) {
	cfg := NewShellConfig()
	cfg.Shell = "/bin/sh"

	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected no error for valid shell, got: %v", err)
	}
}

func TestValidateShell_MissingShell(t *testing.T) {
	cfg := NewShellConfig()
	cfg.Shell = "/nonexistent/shell"

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing shell")
	}

	// Should be a ShellNotFoundError
	if _, ok := err.(*ShellNotFoundError); !ok {
		t.Errorf("expected ShellNotFoundError, got %T", err)
	}
}

func TestValidateShell_ShellInPath(t *testing.T) {
	// Validate a shell that exists in PATH (like "sh" or "bash")
	cfg := NewShellConfig()
	cfg.Shell = "sh"

	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected no error for shell in PATH, got: %v", err)
	}
}

func TestNewExecutor_ValidatesShell(t *testing.T) {
	cfg := NewShellConfig()
	cfg.Shell = "/nonexistent/shell"

	_, err := NewExecutorWithValidation(cfg)
	if err == nil {
		t.Error("expected error for missing shell")
	}
}
