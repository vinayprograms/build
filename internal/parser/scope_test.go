package parser

import (
	"testing"
)

func TestScope_String(t *testing.T) {
	tests := []struct {
		scope Scope
		want  string
	}{
		{ScopeGlobal, "GLOBAL"},
		{ScopeEnvironment, "ENVIRONMENT"},
		{ScopeRecipe, "RECIPE"},
		{ScopeBlock, "BLOCK"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.scope.String()
			if got != tt.want {
				t.Errorf("Scope.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestScopeStack_New(t *testing.T) {
	ss := NewScopeStack()
	if ss == nil {
		t.Fatal("NewScopeStack() returned nil")
	}
	if ss.Current() != ScopeGlobal {
		t.Errorf("NewScopeStack().Current() = %v, want ScopeGlobal", ss.Current())
	}
	if ss.Depth() != 1 {
		t.Errorf("NewScopeStack().Depth() = %d, want 1", ss.Depth())
	}
}

func TestScopeStack_PushPop(t *testing.T) {
	ss := NewScopeStack()

	// Push environment scope
	ss.Push(ScopeEnvironment)
	if ss.Current() != ScopeEnvironment {
		t.Errorf("after Push(ScopeEnvironment), Current() = %v, want ScopeEnvironment", ss.Current())
	}
	if ss.Depth() != 2 {
		t.Errorf("after Push(ScopeEnvironment), Depth() = %d, want 2", ss.Depth())
	}

	// Push recipe scope (simulating a nested scenario)
	ss.Push(ScopeRecipe)
	if ss.Current() != ScopeRecipe {
		t.Errorf("after Push(ScopeRecipe), Current() = %v, want ScopeRecipe", ss.Current())
	}
	if ss.Depth() != 3 {
		t.Errorf("after Push(ScopeRecipe), Depth() = %d, want 3", ss.Depth())
	}

	// Pop back to environment
	popped := ss.Pop()
	if popped != ScopeRecipe {
		t.Errorf("Pop() = %v, want ScopeRecipe", popped)
	}
	if ss.Current() != ScopeEnvironment {
		t.Errorf("after Pop(), Current() = %v, want ScopeEnvironment", ss.Current())
	}

	// Pop back to global
	popped = ss.Pop()
	if popped != ScopeEnvironment {
		t.Errorf("Pop() = %v, want ScopeEnvironment", popped)
	}
	if ss.Current() != ScopeGlobal {
		t.Errorf("after Pop(), Current() = %v, want ScopeGlobal", ss.Current())
	}

	// Can't pop below global
	popped = ss.Pop()
	if popped != ScopeGlobal {
		t.Errorf("Pop() when at global = %v, want ScopeGlobal", popped)
	}
	if ss.Current() != ScopeGlobal {
		t.Errorf("after Pop() at global, Current() = %v, want ScopeGlobal", ss.Current())
	}
}

func TestScopeStack_IsIn(t *testing.T) {
	ss := NewScopeStack()

	// At global, only in global
	if !ss.IsIn(ScopeGlobal) {
		t.Error("at global, IsIn(ScopeGlobal) should be true")
	}
	if ss.IsIn(ScopeEnvironment) {
		t.Error("at global, IsIn(ScopeEnvironment) should be false")
	}

	// Push environment
	ss.Push(ScopeEnvironment)
	if !ss.IsIn(ScopeGlobal) {
		t.Error("at environment, IsIn(ScopeGlobal) should be true (we're inside global)")
	}
	if !ss.IsIn(ScopeEnvironment) {
		t.Error("at environment, IsIn(ScopeEnvironment) should be true")
	}
	if ss.IsIn(ScopeRecipe) {
		t.Error("at environment, IsIn(ScopeRecipe) should be false")
	}

	// Push recipe inside environment (hypothetically)
	ss.Push(ScopeRecipe)
	if !ss.IsIn(ScopeGlobal) {
		t.Error("at recipe, IsIn(ScopeGlobal) should be true")
	}
	if !ss.IsIn(ScopeEnvironment) {
		t.Error("at recipe, IsIn(ScopeEnvironment) should be true")
	}
	if !ss.IsIn(ScopeRecipe) {
		t.Error("at recipe, IsIn(ScopeRecipe) should be true")
	}
}

func TestScopeStack_Reset(t *testing.T) {
	ss := NewScopeStack()
	ss.Push(ScopeEnvironment)
	ss.Push(ScopeRecipe)
	ss.Push(ScopeBlock)

	ss.Reset()

	if ss.Current() != ScopeGlobal {
		t.Errorf("after Reset(), Current() = %v, want ScopeGlobal", ss.Current())
	}
	if ss.Depth() != 1 {
		t.Errorf("after Reset(), Depth() = %d, want 1", ss.Depth())
	}
}

func TestScopeStack_RecipeAndBlockNesting(t *testing.T) {
	// Test the typical recipe → block nesting scenario
	ss := NewScopeStack()

	// Start at global, enter recipe
	ss.Push(ScopeRecipe)
	if ss.Current() != ScopeRecipe {
		t.Fatalf("after entering recipe, Current() = %v, want ScopeRecipe", ss.Current())
	}

	// Enter block inside recipe
	ss.Push(ScopeBlock)
	if ss.Current() != ScopeBlock {
		t.Fatalf("after entering block, Current() = %v, want ScopeBlock", ss.Current())
	}

	// Should be in all three scopes
	if !ss.IsIn(ScopeGlobal) || !ss.IsIn(ScopeRecipe) || !ss.IsIn(ScopeBlock) {
		t.Error("inside block, should be in GLOBAL, RECIPE, and BLOCK")
	}

	// Exit block
	ss.Pop()
	if ss.Current() != ScopeRecipe {
		t.Errorf("after exiting block, Current() = %v, want ScopeRecipe", ss.Current())
	}

	// Exit recipe
	ss.Pop()
	if ss.Current() != ScopeGlobal {
		t.Errorf("after exiting recipe, Current() = %v, want ScopeGlobal", ss.Current())
	}
}
