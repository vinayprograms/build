package planner

import (
	"testing"
	"time"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/eval"
	"github.com/vinayprograms/build/internal/output"
)

// TestPlanner_EmitsStalenessCheckedEvents tests that planner emits staleness check events.
func TestPlanner_EmitsStalenessCheckedEvents(t *testing.T) {
	collector := &emitterEventCollector{}
	emitter := output.NewEmitter(collector)

	fs := &mockFileSystem{
		exists:  map[string]bool{"src/main.c": true},
		missing: map[string]bool{"main.o": true},
	}

	targets := []*ast.Target{
		createTargetWithRecipe("main.o", false, false, []string{"src/main.c"}, &ast.Recipe{}),
	}

	ctx := eval.NewContext()

	_, err := PlanBuildWithEmitter("main.o", targets, ctx, fs, emitter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have emitted a StalenessChecked event
	var found bool
	for _, ev := range collector.events {
		if sc, ok := ev.(output.StalenessChecked); ok {
			found = true
			if sc.Target != "main.o" {
				t.Errorf("expected target 'main.o', got '%s'", sc.Target)
			}
			// Target missing, should rebuild
			if sc.Action != "rebuild" {
				t.Errorf("expected action 'rebuild', got '%s'", sc.Action)
			}
		}
	}
	if !found {
		t.Error("expected StalenessChecked event to be emitted")
	}
}

// TestPlanner_EmitsSkipForUpToDate tests that planner emits skip events for up-to-date targets.
func TestPlanner_EmitsSkipForUpToDate(t *testing.T) {
	collector := &emitterEventCollector{}
	emitter := output.NewEmitter(collector)

	now := time.Now()
	fs := &mockFileSystem{
		exists: map[string]bool{
			"src/main.c": true,
			"main.o":     true,
		},
		mtimes: map[string]time.Time{
			"src/main.c": now.Add(-time.Hour), // Source is older
			"main.o":     now,                 // Target is newer
		},
	}

	targets := []*ast.Target{
		createTargetWithRecipe("main.o", false, false, []string{"src/main.c"}, &ast.Recipe{}),
	}

	ctx := eval.NewContext()

	plan, err := PlanBuildWithEmitter("main.o", targets, ctx, fs, emitter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Plan should have no tasks (up to date)
	if len(plan.Tasks) != 0 {
		t.Errorf("expected 0 tasks (up to date), got %d", len(plan.Tasks))
	}

	// Should have emitted a StalenessChecked event with "skip"
	var found bool
	for _, ev := range collector.events {
		if sc, ok := ev.(output.StalenessChecked); ok {
			found = true
			if sc.Action != "skip" {
				t.Errorf("expected action 'skip', got '%s'", sc.Action)
			}
		}
	}
	if !found {
		t.Error("expected StalenessChecked event to be emitted")
	}
}

// TestPlanner_EmitsRebuildReason tests that planner provides accurate rebuild reasons.
func TestPlanner_EmitsRebuildReason(t *testing.T) {
	tests := []struct {
		name           string
		fs             *mockFileSystem
		expectedReason string
	}{
		{
			name: "target missing",
			fs: &mockFileSystem{
				exists:  map[string]bool{"src/main.c": true},
				missing: map[string]bool{"main.o": true},
			},
			expectedReason: "target missing",
		},
		{
			name: "dependency newer",
			fs: &mockFileSystem{
				exists: map[string]bool{
					"src/main.c": true,
					"main.o":     true,
				},
				mtimes: map[string]time.Time{
					"src/main.c": time.Now(),
					"main.o":     time.Now().Add(-time.Hour),
				},
			},
			expectedReason: "dependency newer than target",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := &emitterEventCollector{}
			emitter := output.NewEmitter(collector)

			targets := []*ast.Target{
				createTargetWithRecipe("main.o", false, false, []string{"src/main.c"}, &ast.Recipe{}),
			}

			ctx := eval.NewContext()

			_, err := PlanBuildWithEmitter("main.o", targets, ctx, tt.fs, emitter)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Find the StalenessChecked event and verify reason
			var found bool
			for _, ev := range collector.events {
				if sc, ok := ev.(output.StalenessChecked); ok {
					found = true
					if sc.Reason != tt.expectedReason {
						t.Errorf("expected reason '%s', got '%s'", tt.expectedReason, sc.Reason)
					}
				}
			}
			if !found {
				t.Error("expected StalenessChecked event to be emitted")
			}
		})
	}
}

// TestPlanner_EmitsPhonyRebuild tests that phony targets always emit rebuild.
func TestPlanner_EmitsPhonyRebuild(t *testing.T) {
	collector := &emitterEventCollector{}
	emitter := output.NewEmitter(collector)

	fs := &mockFileSystem{
		exists: map[string]bool{},
	}

	targets := []*ast.Target{
		createTargetWithRecipe("clean", true, false, nil, &ast.Recipe{}),
	}

	ctx := eval.NewContext()

	_, err := PlanBuildWithEmitter("@clean", targets, ctx, fs, emitter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have emitted a StalenessChecked event with phony reason
	var found bool
	for _, ev := range collector.events {
		if sc, ok := ev.(output.StalenessChecked); ok {
			found = true
			if sc.Reason != "phony target" {
				t.Errorf("expected reason 'phony target', got '%s'", sc.Reason)
			}
			if sc.Action != "rebuild" {
				t.Errorf("expected action 'rebuild', got '%s'", sc.Action)
			}
		}
	}
	if !found {
		t.Error("expected StalenessChecked event to be emitted")
	}
}

// TestPlanner_NilEmitter tests that planner works without an emitter.
func TestPlanner_NilEmitter(t *testing.T) {
	fs := &mockFileSystem{
		exists:  map[string]bool{"src/main.c": true},
		missing: map[string]bool{"main.o": true},
	}

	targets := []*ast.Target{
		createTargetWithRecipe("main.o", false, false, []string{"src/main.c"}, &ast.Recipe{}),
	}

	ctx := eval.NewContext()

	// Should work without emitter
	_, err := PlanBuildWithEmitter("main.o", targets, ctx, fs, nil)
	if err != nil {
		t.Fatalf("should work without emitter: %v", err)
	}
}

// emitterEventCollector collects output events for testing.
type emitterEventCollector struct {
	events []output.OutputEvent
}

func (c *emitterEventCollector) WriteEvent(event output.OutputEvent) {
	c.events = append(c.events, event)
}

func (c *emitterEventCollector) Flush() {}
