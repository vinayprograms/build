package lexer

import (
	"testing"
)

func TestIndentChar(t *testing.T) {
	tests := []struct {
		name string
		char IndentChar
		want string
	}{
		{"unknown", IndentUnknown, "unknown"},
		{"space", IndentSpace, "space"},
		{"tab", IndentTab, "tab"},
		{"invalid", IndentChar(99), "IndentChar(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.char.String(); got != tt.want {
				t.Errorf("IndentChar.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewIndentTracker(t *testing.T) {
	tracker := NewIndentTracker()

	if tracker.Char() != IndentUnknown {
		t.Errorf("initial Char() = %v, want %v", tracker.Char(), IndentUnknown)
	}
	if tracker.Width() != 0 {
		t.Errorf("initial Width() = %d, want 0", tracker.Width())
	}
}

func TestIndentTrackerFirstIndentSetsUnit(t *testing.T) {
	tests := []struct {
		name       string
		indent     string
		wantChar   IndentChar
		wantWidth  int
		wantLevel  int
		wantErr    bool
		errMessage string
	}{
		{
			name:      "4 spaces",
			indent:    "    ",
			wantChar:  IndentSpace,
			wantWidth: 4,
			wantLevel: 1,
		},
		{
			name:      "2 spaces",
			indent:    "  ",
			wantChar:  IndentSpace,
			wantWidth: 2,
			wantLevel: 1,
		},
		{
			name:      "1 tab",
			indent:    "\t",
			wantChar:  IndentTab,
			wantWidth: 1,
			wantLevel: 1,
		},
		{
			name:      "2 tabs",
			indent:    "\t\t",
			wantChar:  IndentTab,
			wantWidth: 2,
			wantLevel: 1,
		},
		{
			name:       "mixed spaces then tab",
			indent:     "  \t",
			wantErr:    true,
			errMessage: "mixed indentation",
		},
		{
			name:       "mixed tab then spaces",
			indent:     "\t  ",
			wantErr:    true,
			errMessage: "mixed indentation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewIndentTracker()
			level, err := tracker.Process(tt.indent)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Process(%q) error = nil, want error containing %q", tt.indent, tt.errMessage)
				}
				if !containsString(err.Error(), tt.errMessage) {
					t.Errorf("error = %q, want error containing %q", err.Error(), tt.errMessage)
				}
				return
			}

			if err != nil {
				t.Fatalf("Process(%q) unexpected error: %v", tt.indent, err)
			}

			if level != tt.wantLevel {
				t.Errorf("Process(%q) level = %d, want %d", tt.indent, level, tt.wantLevel)
			}
			if tracker.Char() != tt.wantChar {
				t.Errorf("Char() = %v, want %v", tracker.Char(), tt.wantChar)
			}
			if tracker.Width() != tt.wantWidth {
				t.Errorf("Width() = %d, want %d", tracker.Width(), tt.wantWidth)
			}
		})
	}
}

func TestIndentTrackerSubsequentIndents(t *testing.T) {
	tests := []struct {
		name      string
		first     string
		second    string
		wantLevel int
		wantErr   bool
	}{
		{
			name:      "same indent level",
			first:     "    ",
			second:    "    ",
			wantLevel: 1,
		},
		{
			name:      "double indent (level 2)",
			first:     "    ",
			second:    "        ",
			wantLevel: 2,
		},
		{
			name:      "back to level 0",
			first:     "    ",
			second:    "",
			wantLevel: 0,
		},
		{
			name:    "inconsistent indent width",
			first:   "    ",   // 4 spaces = unit
			second:  "      ", // 6 spaces, not multiple of 4
			wantErr: true,
		},
		{
			name:    "different char after establishing unit",
			first:   "    ", // spaces established
			second:  "\t",   // tab
			wantErr: true,
		},
		{
			name:      "tabs level 2",
			first:     "\t",
			second:    "\t\t",
			wantLevel: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewIndentTracker()

			// Process first indent to establish the unit
			_, err := tracker.Process(tt.first)
			if err != nil {
				t.Fatalf("Process first indent %q unexpected error: %v", tt.first, err)
			}

			// Process second indent
			level, err := tracker.Process(tt.second)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Process(%q) error = nil, want error", tt.second)
				}
				return
			}

			if err != nil {
				t.Fatalf("Process(%q) unexpected error: %v", tt.second, err)
			}

			if level != tt.wantLevel {
				t.Errorf("Process(%q) level = %d, want %d", tt.second, level, tt.wantLevel)
			}
		})
	}
}

func TestIndentTrackerEmptyAndWhitespaceLines(t *testing.T) {
	tracker := NewIndentTracker()

	// Empty line before any indentation
	level, err := tracker.Process("")
	if err != nil {
		t.Fatalf("Process empty string unexpected error: %v", err)
	}
	if level != 0 {
		t.Errorf("empty line level = %d, want 0", level)
	}
	// Should not establish indent unit
	if tracker.Char() != IndentUnknown {
		t.Errorf("Char() after empty = %v, want IndentUnknown", tracker.Char())
	}

	// Now establish unit with 4 spaces
	level, err = tracker.Process("    ")
	if err != nil {
		t.Fatalf("Process 4 spaces unexpected error: %v", err)
	}
	if level != 1 {
		t.Errorf("4 spaces level = %d, want 1", level)
	}

	// Empty line should still work and not affect tracking
	level, err = tracker.Process("")
	if err != nil {
		t.Fatalf("Process empty after indent unexpected error: %v", err)
	}
	if level != 0 {
		t.Errorf("empty line level = %d, want 0", level)
	}
}

func TestIndentTrackerReset(t *testing.T) {
	tracker := NewIndentTracker()

	// Establish indent with spaces
	_, err := tracker.Process("    ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tracker.Char() != IndentSpace {
		t.Errorf("Char() = %v, want IndentSpace", tracker.Char())
	}

	// Reset
	tracker.Reset()

	if tracker.Char() != IndentUnknown {
		t.Errorf("after Reset Char() = %v, want IndentUnknown", tracker.Char())
	}
	if tracker.Width() != 0 {
		t.Errorf("after Reset Width() = %d, want 0", tracker.Width())
	}
}

func TestIndentTrackerLevel0AlwaysWorks(t *testing.T) {
	tracker := NewIndentTracker()

	// Establish indent with tabs
	_, err := tracker.Process("\t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Level 0 (no indent) should always work, even with empty string
	level, err := tracker.Process("")
	if err != nil {
		t.Fatalf("Process empty string unexpected error: %v", err)
	}
	if level != 0 {
		t.Errorf("level = %d, want 0", level)
	}
}

func TestIndentTrackerDeepIndentation(t *testing.T) {
	tracker := NewIndentTracker()

	// Establish 2-space unit
	level, err := tracker.Process("  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if level != 1 {
		t.Errorf("level 1 = %d, want 1", level)
	}

	// Level 2
	level, err = tracker.Process("    ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if level != 2 {
		t.Errorf("level 2 = %d, want 2", level)
	}

	// Level 3
	level, err = tracker.Process("      ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if level != 3 {
		t.Errorf("level 3 = %d, want 3", level)
	}
}

func TestIndentTrackerPartialIndentError(t *testing.T) {
	tracker := NewIndentTracker()

	// Establish 4-space unit
	_, err := tracker.Process("    ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 3 spaces is not a valid multiple of 4
	_, err = tracker.Process("   ")
	if err == nil {
		t.Error("Process 3 spaces should error when unit is 4")
	}

	// 5 spaces is not a valid multiple of 4
	_, err = tracker.Process("     ")
	if err == nil {
		t.Error("Process 5 spaces should error when unit is 4")
	}
}

func TestIndentTrackerOnlyTwoCharTypes(t *testing.T) {
	// Per spec, only spaces and tabs are valid indent characters
	tracker := NewIndentTracker()

	// Exotic whitespace should be rejected (e.g., non-breaking space U+00A0)
	_, err := tracker.Process("\u00A0\u00A0")
	if err == nil {
		t.Error("exotic whitespace should be rejected")
	}
}

func TestIndentError(t *testing.T) {
	err := &IndentError{
		Message: "test error",
		Line:    5,
		Column:  3,
	}

	expected := "indentation error at line 5, column 3: test error"
	if got := err.Error(); got != expected {
		t.Errorf("IndentError.Error() = %q, want %q", got, expected)
	}
}

func TestIndentTrackerMixedAfterEstablished(t *testing.T) {
	// After establishing spaces as the indent char,
	// any tabs in the indentation should be an error
	tracker := NewIndentTracker()

	_, err := tracker.Process("    ") // 4 spaces
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Now try mixed (spaces followed by tab)
	_, err = tracker.Process("    \t") // 4 spaces + tab
	if err == nil {
		t.Error("Process mixed indent should error")
	}
}

// containsString checks if s contains substr
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
