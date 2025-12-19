package executor

import (
	"testing"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/eval"
	"github.com/vinayprograms/build/internal/output"
)

// ----------------------------------------------------------------------------
// Emitter Integration Tests
// ----------------------------------------------------------------------------

// TestExecutor_EmitsCommandEvents tests that executor emits command lifecycle events.
func TestExecutor_EmitsCommandEvents(t *testing.T) {
	collector := &eventCollector{}
	emitter := output.NewEmitter(collector)

	cfg := NewShellConfig()
	exec := NewExecutor(cfg)
	exec.SetEmitter(emitter)

	ctx := eval.NewContext()
	cmdCtx := eval.NewCommandContext(ctx, "foo.o", []string{"foo.c"})

	recipe := &ast.Recipe{
		Commands: []ast.Command{
			&ast.LineCommand{
				Parts: []ast.CommandPart{
					&ast.LiteralCommand{Text: "echo hello"},
				},
			},
		},
	}

	_, err := exec.ExecuteRecipe(recipe, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have: CommandStarted, CommandOutput (if any), CommandCompleted
	if len(collector.events) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(collector.events))
	}

	// Check CommandStarted
	cs, ok := collector.events[0].(output.CommandStarted)
	if !ok {
		t.Fatalf("expected CommandStarted, got %T", collector.events[0])
	}
	if cs.Target != "foo.o" {
		t.Errorf("expected target 'foo.o', got '%s'", cs.Target)
	}
	if cs.Command != "echo hello" {
		t.Errorf("expected command 'echo hello', got '%s'", cs.Command)
	}

	// Check CommandCompleted is last
	lastEvent := collector.events[len(collector.events)-1]
	cc, ok := lastEvent.(output.CommandCompleted)
	if !ok {
		t.Fatalf("expected CommandCompleted as last event, got %T", lastEvent)
	}
	if cc.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", cc.ExitCode)
	}
}

// TestExecutor_EmitsCommandOutput tests that executor emits command output events.
func TestExecutor_EmitsCommandOutput(t *testing.T) {
	collector := &eventCollector{}
	emitter := output.NewEmitter(collector)

	cfg := NewShellConfig()
	exec := NewExecutor(cfg)
	exec.SetEmitter(emitter)

	ctx := eval.NewContext()
	cmdCtx := eval.NewCommandContext(ctx, "foo.o", []string{"foo.c"})

	recipe := &ast.Recipe{
		Commands: []ast.Command{
			&ast.LineCommand{
				Parts: []ast.CommandPart{
					&ast.LiteralCommand{Text: "echo output"},
				},
			},
		},
	}

	_, err := exec.ExecuteRecipe(recipe, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find CommandOutput event
	var found bool
	for _, ev := range collector.events {
		if co, ok := ev.(output.CommandOutput); ok {
			found = true
			if co.Stdout != "output\n" {
				t.Errorf("expected stdout 'output\\n', got '%s'", co.Stdout)
			}
		}
	}
	if !found {
		t.Error("expected CommandOutput event to be emitted")
	}
}

// TestExecutor_EmitsOnError tests that executor emits events on command failure.
func TestExecutor_EmitsOnError(t *testing.T) {
	collector := &eventCollector{}
	emitter := output.NewEmitter(collector)

	cfg := NewShellConfig()
	exec := NewExecutor(cfg)
	exec.SetEmitter(emitter)

	ctx := eval.NewContext()
	cmdCtx := eval.NewCommandContext(ctx, "foo.o", []string{"foo.c"})

	recipe := &ast.Recipe{
		Commands: []ast.Command{
			&ast.LineCommand{
				Parts: []ast.CommandPart{
					&ast.LiteralCommand{Text: "exit 42"},
				},
			},
		},
	}

	_, err := exec.ExecuteRecipe(recipe, cmdCtx)
	if err == nil {
		t.Fatal("expected error")
	}

	// Check CommandCompleted with error
	var found bool
	for _, ev := range collector.events {
		if cc, ok := ev.(output.CommandCompleted); ok {
			found = true
			if cc.ExitCode != 42 {
				t.Errorf("expected exit code 42, got %d", cc.ExitCode)
			}
		}
	}
	if !found {
		t.Error("expected CommandCompleted event to be emitted")
	}
}

// TestExecutor_DryRunEmitsDryRunEvents tests that dry-run emits DryRunCommand events.
func TestExecutor_DryRunEmitsDryRunEvents(t *testing.T) {
	collector := &eventCollector{}
	emitter := output.NewEmitter(collector)

	cfg := NewShellConfig()
	cfg.DryRun = true
	exec := NewExecutor(cfg)
	exec.SetEmitter(emitter)

	ctx := eval.NewContext()
	cmdCtx := eval.NewCommandContext(ctx, "foo.o", []string{"foo.c"})

	recipe := &ast.Recipe{
		Commands: []ast.Command{
			&ast.LineCommand{
				Parts: []ast.CommandPart{
					&ast.LiteralCommand{Text: "gcc -c foo.c"},
				},
			},
		},
	}

	_, err := exec.ExecuteRecipe(recipe, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should emit DryRunCommand event
	var found bool
	for _, ev := range collector.events {
		if drc, ok := ev.(output.DryRunCommand); ok {
			found = true
			if drc.Command != "gcc -c foo.c" {
				t.Errorf("expected command 'gcc -c foo.c', got '%s'", drc.Command)
			}
			if drc.Target != "foo.o" {
				t.Errorf("expected target 'foo.o', got '%s'", drc.Target)
			}
		}
	}
	if !found {
		t.Error("expected DryRunCommand event in dry-run mode")
	}
}

// TestExecutor_NilEmitter tests that executor works without an emitter.
func TestExecutor_NilEmitter(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)
	// No emitter set

	ctx := eval.NewContext()
	cmdCtx := eval.NewCommandContext(ctx, "foo.o", []string{"foo.c"})

	recipe := &ast.Recipe{
		Commands: []ast.Command{
			&ast.LineCommand{
				Parts: []ast.CommandPart{
					&ast.LiteralCommand{Text: "echo hello"},
				},
			},
		},
	}

	_, err := exec.ExecuteRecipe(recipe, cmdCtx)
	if err != nil {
		t.Fatalf("should work without emitter: %v", err)
	}
}

// TestExecutor_MultipleCommands tests events for multiple commands.
func TestExecutor_MultipleCommands(t *testing.T) {
	collector := &eventCollector{}
	emitter := output.NewEmitter(collector)

	cfg := NewShellConfig()
	exec := NewExecutor(cfg)
	exec.SetEmitter(emitter)

	ctx := eval.NewContext()
	cmdCtx := eval.NewCommandContext(ctx, "foo.o", []string{})

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

	_, err := exec.ExecuteRecipe(recipe, cmdCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Count CommandStarted events
	startedCount := 0
	completedCount := 0
	for _, ev := range collector.events {
		if _, ok := ev.(output.CommandStarted); ok {
			startedCount++
		}
		if _, ok := ev.(output.CommandCompleted); ok {
			completedCount++
		}
	}

	if startedCount != 2 {
		t.Errorf("expected 2 CommandStarted events, got %d", startedCount)
	}
	if completedCount != 2 {
		t.Errorf("expected 2 CommandCompleted events, got %d", completedCount)
	}
}

// eventCollector collects output events for testing.
type eventCollector struct {
	events []output.OutputEvent
}

func (c *eventCollector) WriteEvent(event output.OutputEvent) {
	c.events = append(c.events, event)
}

func (c *eventCollector) Flush() {}
