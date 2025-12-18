package executor

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/eval"
	"github.com/vinayprograms/build/internal/planner"
)

// ----------------------------------------------------------------------------
// Scheduler Tests
// ----------------------------------------------------------------------------

func TestNewScheduler(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)
	sched := NewScheduler(exec, 2)

	if sched == nil {
		t.Fatal("NewScheduler returned nil")
	}

	if sched.Workers() != 2 {
		t.Errorf("expected 2 workers, got %d", sched.Workers())
	}
}

func TestScheduler_SingleTask_NoDeps(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)
	sched := NewScheduler(exec, 1)

	ctx := eval.NewContext()
	cmdCtx := eval.NewCommandContext(ctx, "test.txt", nil)

	tasks := []planner.BuildTask{
		{
			Target:       "test.txt",
			Dependencies: nil,
			Recipe: &ast.Recipe{
				Commands: []ast.Command{
					&ast.LineCommand{
						Parts: []ast.CommandPart{
							&ast.LiteralCommand{Text: "echo test"},
						},
					},
				},
			},
			Reason: planner.BuildReasonTargetMissing,
		},
	}

	results := sched.Execute(tasks, func(target string) *eval.CommandContext {
		return cmdCtx
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Error != nil {
		t.Errorf("unexpected error: %v", results[0].Error)
	}

	if results[0].Target != "test.txt" {
		t.Errorf("expected target 'test.txt', got '%s'", results[0].Target)
	}
}

func TestScheduler_MultipleTasks_NoDeps(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)
	sched := NewScheduler(exec, 2)

	ctx := eval.NewContext()

	tasks := []planner.BuildTask{
		{
			Target:       "a.txt",
			Dependencies: nil,
			Recipe: &ast.Recipe{
				Commands: []ast.Command{
					&ast.LineCommand{
						Parts: []ast.CommandPart{
							&ast.LiteralCommand{Text: "echo a"},
						},
					},
				},
			},
			Reason: planner.BuildReasonTargetMissing,
		},
		{
			Target:       "b.txt",
			Dependencies: nil,
			Recipe: &ast.Recipe{
				Commands: []ast.Command{
					&ast.LineCommand{
						Parts: []ast.CommandPart{
							&ast.LiteralCommand{Text: "echo b"},
						},
					},
				},
			},
			Reason: planner.BuildReasonTargetMissing,
		},
	}

	results := sched.Execute(tasks, func(target string) *eval.CommandContext {
		return eval.NewCommandContext(ctx, target, nil)
	})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Both should succeed
	for _, r := range results {
		if r.Error != nil {
			t.Errorf("unexpected error for %s: %v", r.Target, r.Error)
		}
	}
}

func TestScheduler_DependencyOrdering(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)
	sched := NewScheduler(exec, 1) // Single worker to ensure ordering

	ctx := eval.NewContext()

	// B depends on A, so A should run first
	tasks := []planner.BuildTask{
		{
			Target:       "a.txt",
			Dependencies: nil,
			Recipe: &ast.Recipe{
				Commands: []ast.Command{
					&ast.LineCommand{
						Parts: []ast.CommandPart{
							&ast.LiteralCommand{Text: "echo a"},
						},
					},
				},
			},
			Reason: planner.BuildReasonTargetMissing,
		},
		{
			Target:       "b.txt",
			Dependencies: []string{"a.txt"},
			Recipe: &ast.Recipe{
				Commands: []ast.Command{
					&ast.LineCommand{
						Parts: []ast.CommandPart{
							&ast.LiteralCommand{Text: "echo b"},
						},
					},
				},
			},
			Reason: planner.BuildReasonTargetMissing,
		},
	}

	var execOrder []string
	var mu sync.Mutex

	results := sched.ExecuteWithCallback(tasks,
		func(target string) *eval.CommandContext {
			return eval.NewCommandContext(ctx, target, nil)
		},
		func(target string) {
			mu.Lock()
			execOrder = append(execOrder, target)
			mu.Unlock()
		},
	)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// A should come before B
	if len(execOrder) != 2 || execOrder[0] != "a.txt" || execOrder[1] != "b.txt" {
		t.Errorf("expected execution order [a.txt, b.txt], got %v", execOrder)
	}
}

func TestScheduler_ParallelIndependentTasks(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)
	sched := NewScheduler(exec, 4)

	ctx := eval.NewContext()

	// All tasks are independent, should run in parallel
	tasks := make([]planner.BuildTask, 4)
	for i := 0; i < 4; i++ {
		target := string(rune('a'+i)) + ".txt"
		tasks[i] = planner.BuildTask{
			Target:       target,
			Dependencies: nil,
			Recipe: &ast.Recipe{
				Commands: []ast.Command{
					&ast.LineCommand{
						Parts: []ast.CommandPart{
							&ast.LiteralCommand{Text: "echo " + target},
						},
					},
				},
			},
			Reason: planner.BuildReasonTargetMissing,
		}
	}

	start := time.Now()
	results := sched.Execute(tasks, func(target string) *eval.CommandContext {
		return eval.NewCommandContext(ctx, target, nil)
	})
	elapsed := time.Since(start)

	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}

	// Should complete quickly since tasks run in parallel
	if elapsed > 2*time.Second {
		t.Errorf("parallel execution took too long: %v", elapsed)
	}

	for _, r := range results {
		if r.Error != nil {
			t.Errorf("unexpected error for %s: %v", r.Target, r.Error)
		}
	}
}

func TestScheduler_DiamondDependency(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)
	sched := NewScheduler(exec, 2)

	ctx := eval.NewContext()

	// Diamond: D depends on B and C, B and C depend on A
	// A
	// |\
	// B C
	// |/
	// D
	tasks := []planner.BuildTask{
		{Target: "a.txt", Dependencies: nil, Recipe: makeEchoRecipe("a"), Reason: planner.BuildReasonTargetMissing},
		{Target: "b.txt", Dependencies: []string{"a.txt"}, Recipe: makeEchoRecipe("b"), Reason: planner.BuildReasonTargetMissing},
		{Target: "c.txt", Dependencies: []string{"a.txt"}, Recipe: makeEchoRecipe("c"), Reason: planner.BuildReasonTargetMissing},
		{Target: "d.txt", Dependencies: []string{"b.txt", "c.txt"}, Recipe: makeEchoRecipe("d"), Reason: planner.BuildReasonTargetMissing},
	}

	var execOrder []string
	var mu sync.Mutex

	results := sched.ExecuteWithCallback(tasks,
		func(target string) *eval.CommandContext {
			return eval.NewCommandContext(ctx, target, nil)
		},
		func(target string) {
			mu.Lock()
			execOrder = append(execOrder, target)
			mu.Unlock()
		},
	)

	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}

	// A should be first, D should be last
	if execOrder[0] != "a.txt" {
		t.Errorf("expected a.txt first, got %s", execOrder[0])
	}
	if execOrder[3] != "d.txt" {
		t.Errorf("expected d.txt last, got %s", execOrder[3])
	}
}

func TestScheduler_FailureCancellation(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)
	sched := NewScheduler(exec, 1)

	ctx := eval.NewContext()

	// First task fails, second should not run
	tasks := []planner.BuildTask{
		{
			Target:       "fail.txt",
			Dependencies: nil,
			Recipe: &ast.Recipe{
				Commands: []ast.Command{
					&ast.LineCommand{
						Parts: []ast.CommandPart{
							&ast.LiteralCommand{Text: "exit 1"},
						},
					},
				},
			},
			Reason: planner.BuildReasonTargetMissing,
		},
		{
			Target:       "after.txt",
			Dependencies: []string{"fail.txt"},
			Recipe:       makeEchoRecipe("after"),
			Reason:       planner.BuildReasonTargetMissing,
		},
	}

	results := sched.Execute(tasks, func(target string) *eval.CommandContext {
		return eval.NewCommandContext(ctx, target, nil)
	})

	// First should fail, second should be skipped
	var failedCount, skippedCount int
	for _, r := range results {
		if r.Error != nil {
			failedCount++
		}
		if r.Skipped {
			skippedCount++
		}
	}

	if failedCount != 1 {
		t.Errorf("expected 1 failure, got %d", failedCount)
	}

	if skippedCount != 1 {
		t.Errorf("expected 1 skipped, got %d", skippedCount)
	}
}

func TestScheduler_NoRecipe(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)
	sched := NewScheduler(exec, 1)

	ctx := eval.NewContext()

	// Task with no recipe (source file)
	tasks := []planner.BuildTask{
		{
			Target:       "source.c",
			Dependencies: nil,
			Recipe:       nil, // No recipe - source file
			Reason:       planner.BuildReasonTargetMissing,
		},
	}

	results := sched.Execute(tasks, func(target string) *eval.CommandContext {
		return eval.NewCommandContext(ctx, target, nil)
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	// Should succeed (no-op)
	if results[0].Error != nil {
		t.Errorf("unexpected error: %v", results[0].Error)
	}
}

func TestScheduler_ParallelWorkerCount(t *testing.T) {
	cfg := NewShellConfig()
	exec := NewExecutor(cfg)
	sched := NewScheduler(exec, 4)

	ctx := eval.NewContext()

	// Create tasks that take some time to measure parallelism
	var concurrentCount int32
	var maxConcurrent int32

	tasks := make([]planner.BuildTask, 8)
	for i := 0; i < 8; i++ {
		target := string(rune('a'+i)) + ".txt"
		tasks[i] = planner.BuildTask{
			Target:       target,
			Dependencies: nil,
			Recipe: &ast.Recipe{
				Commands: []ast.Command{
					&ast.LineCommand{
						Parts: []ast.CommandPart{
							&ast.LiteralCommand{Text: "sleep 0.1"},
						},
					},
				},
			},
			Reason: planner.BuildReasonTargetMissing,
		}
	}

	results := sched.ExecuteWithCallback(tasks,
		func(target string) *eval.CommandContext {
			return eval.NewCommandContext(ctx, target, nil)
		},
		func(target string) {
			current := atomic.AddInt32(&concurrentCount, 1)
			if current > atomic.LoadInt32(&maxConcurrent) {
				atomic.StoreInt32(&maxConcurrent, current)
			}
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt32(&concurrentCount, -1)
		},
	)

	if len(results) != 8 {
		t.Fatalf("expected 8 results, got %d", len(results))
	}

	// Max concurrent should be limited by worker count (4)
	if atomic.LoadInt32(&maxConcurrent) > 4 {
		t.Errorf("max concurrent exceeded worker count: %d", maxConcurrent)
	}
}

// Helper to create an echo recipe
func makeEchoRecipe(msg string) *ast.Recipe {
	return &ast.Recipe{
		Commands: []ast.Command{
			&ast.LineCommand{
				Parts: []ast.CommandPart{
					&ast.LiteralCommand{Text: "echo " + msg},
				},
			},
		},
	}
}
