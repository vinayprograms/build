package errors

import (
	"strings"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

func TestFormattedError_Basic(t *testing.T) {
	err := &FormattedError{
		Code:    "E001",
		Message: "invalid character",
		Location: ast.SourceLocation{
			File:   "Buildfile",
			Line:   10,
			Column: 5,
		},
	}

	got := err.Error()
	if !strings.Contains(got, "E001") {
		t.Errorf("Error should contain code E001, got: %q", got)
	}
	if !strings.Contains(got, "invalid character") {
		t.Errorf("Error should contain message, got: %q", got)
	}
	if !strings.Contains(got, "Buildfile:10:5") {
		t.Errorf("Error should contain location, got: %q", got)
	}
}

func TestFormattedError_WithNote(t *testing.T) {
	err := &FormattedError{
		Code:     "E100",
		Message:  "unexpected token",
		Location: ast.SourceLocation{File: "Buildfile", Line: 5, Column: 1},
		Note:     "expected ':' after target name",
	}

	got := err.Error()
	if !strings.Contains(got, "note:") {
		t.Errorf("Error with note should contain 'note:', got: %q", got)
	}
	if !strings.Contains(got, "expected ':' after target name") {
		t.Errorf("Error should contain note text, got: %q", got)
	}
}

func TestFormattedError_WithHelp(t *testing.T) {
	err := &FormattedError{
		Code:     "E200",
		Message:  "undefined variable 'foo'",
		Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 10},
		Help:     "define the variable before use: foo = value",
	}

	got := err.Error()
	if !strings.Contains(got, "help:") {
		t.Errorf("Error with help should contain 'help:', got: %q", got)
	}
	if !strings.Contains(got, "define the variable before use") {
		t.Errorf("Error should contain help text, got: %q", got)
	}
}

func TestFormattedError_WithSourceContext(t *testing.T) {
	err := &FormattedError{
		Code:     "E100",
		Message:  "unexpected token",
		Location: ast.SourceLocation{File: "Buildfile", Line: 5, Column: 8},
		SourceLines: []SourceLine{
			{Number: 4, Text: ""},
			{Number: 5, Text: "build/app deps"},
			{Number: 6, Text: "    echo hello"},
		},
		CaretLine:   5,
		CaretColumn: 8,
	}

	got := err.Error()
	// Should show line numbers
	if !strings.Contains(got, "5 |") {
		t.Errorf("Error should show line number, got: %q", got)
	}
	// Should show caret pointer
	if !strings.Contains(got, "^") {
		t.Errorf("Error should show caret pointer, got: %q", got)
	}
}

func TestFormattedError_CaretPosition(t *testing.T) {
	err := &FormattedError{
		Code:     "E001",
		Message:  "invalid character",
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 5},
		SourceLines: []SourceLine{
			{Number: 1, Text: "foo{bar}"},
		},
		CaretLine:   1,
		CaretColumn: 5,
	}

	got := err.Error()
	// Caret should be under the 5th character (index 4)
	lines := strings.Split(got, "\n")
	var caretLine string
	for _, line := range lines {
		if strings.Contains(line, "^") && !strings.Contains(line, "foo") {
			caretLine = line
			break
		}
	}
	if caretLine == "" {
		t.Errorf("Error should have caret line, got: %q", got)
	}
}

func TestFormattedError_Format(t *testing.T) {
	// Test complete error format
	err := &FormattedError{
		Code:     "E100",
		Message:  "missing ':' in target definition",
		Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 10},
		SourceLines: []SourceLine{
			{Number: 2, Text: "cc = gcc"},
			{Number: 3, Text: "build/app deps"},
			{Number: 4, Text: "    gcc -o build/app deps"},
		},
		CaretLine:   3,
		CaretColumn: 10,
		Note:        "targets require ':' before dependencies",
		Help:        "change to: build/app: deps",
	}

	got := err.Format()
	// Check structure
	if !strings.HasPrefix(got, "error[E100]") {
		t.Errorf("Formatted error should start with 'error[CODE]', got: %q", got[:min(20, len(got))])
	}
}

func TestSourceLine_Format(t *testing.T) {
	line := SourceLine{Number: 10, Text: "hello world"}
	got := line.Format(3) // 3-digit line number width

	if !strings.Contains(got, " 10 |") {
		t.Errorf("SourceLine format should have ' 10 |', got: %q", got)
	}
	if !strings.Contains(got, "hello world") {
		t.Errorf("SourceLine format should contain text, got: %q", got)
	}
}

func TestFormatCaret(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		column    int
		wantCaret string
	}{
		{
			name:      "column 1",
			width:     3,
			column:    1,
			wantCaret: "    | ^",
		},
		{
			name:      "column 5",
			width:     3,
			column:    5,
			wantCaret: "    |     ^",
		},
		{
			name:      "column 10",
			width:     2,
			column:    10,
			wantCaret: "   |          ^",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCaret(tt.width, tt.column)
			if got != tt.wantCaret {
				t.Errorf("FormatCaret(%d, %d) = %q, want %q", tt.width, tt.column, got, tt.wantCaret)
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ----------------------------------------------------------------------------
// Source Extraction Tests
// ----------------------------------------------------------------------------

func TestExtractSourceLines_Middle(t *testing.T) {
	source := `line1
line2
line3
line4
line5`

	lines := ExtractSourceLines(source, 3, 1)

	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}
	if lines[0].Number != 2 || lines[0].Text != "line2" {
		t.Errorf("First line wrong: %+v", lines[0])
	}
	if lines[1].Number != 3 || lines[1].Text != "line3" {
		t.Errorf("Middle line wrong: %+v", lines[1])
	}
	if lines[2].Number != 4 || lines[2].Text != "line4" {
		t.Errorf("Last line wrong: %+v", lines[2])
	}
}

func TestExtractSourceLines_Start(t *testing.T) {
	source := `line1
line2
line3`

	lines := ExtractSourceLines(source, 1, 1)

	if len(lines) != 2 {
		t.Errorf("Expected 2 lines (can't go before line 1), got %d", len(lines))
	}
	if lines[0].Number != 1 {
		t.Errorf("First line should be 1, got %d", lines[0].Number)
	}
}

func TestExtractSourceLines_End(t *testing.T) {
	source := `line1
line2
line3`

	lines := ExtractSourceLines(source, 3, 1)

	if len(lines) != 2 {
		t.Errorf("Expected 2 lines (can't go past end), got %d", len(lines))
	}
	if lines[len(lines)-1].Number != 3 {
		t.Errorf("Last line should be 3, got %d", lines[len(lines)-1].Number)
	}
}

func TestExtractSourceLines_LargerContext(t *testing.T) {
	source := `line1
line2
line3
line4
line5
line6
line7`

	lines := ExtractSourceLines(source, 4, 2)

	if len(lines) != 5 {
		t.Errorf("Expected 5 lines (2 before + target + 2 after), got %d", len(lines))
	}
	// Should be lines 2-6
	if lines[0].Number != 2 {
		t.Errorf("First line should be 2, got %d", lines[0].Number)
	}
	if lines[4].Number != 6 {
		t.Errorf("Last line should be 6, got %d", lines[4].Number)
	}
}
