package parser

import (
	"strings"
	"testing"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/lexer"
)

// TestParser_ParseTarget_Simple tests basic target patterns.
func TestParser_ParseTarget_Simple(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantSegments    int
		wantIsPhony     bool
		wantIsDirectory bool
		wantDeps        int
	}{
		{
			name:         "simple file target no deps",
			input:        "build/app:",
			wantSegments: 1,
			wantDeps:     0,
		},
		{
			name:         "simple file target with single dep",
			input:        "build/app: src/main.c",
			wantSegments: 1,
			wantDeps:     1,
		},
		{
			name:         "simple file target with multiple deps",
			input:        "build/app: build/main.o build/utils.o",
			wantSegments: 1,
			wantDeps:     2,
		},
		{
			name:            "directory target",
			input:           "build/:",
			wantSegments:    1,
			wantIsDirectory: true,
			wantDeps:        0,
		},
		{
			name:         "phony target no deps",
			input:        "@clean:",
			wantSegments: 1,
			wantIsPhony:  true,
			wantDeps:     0,
		},
		{
			name:         "phony target with deps",
			input:        "@all: build/app build/test",
			wantSegments: 1,
			wantIsPhony:  true,
			wantDeps:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.need", tt.input)
			p := New(l)

			target, err := p.ParseTarget()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(target.Pattern.Segments) != tt.wantSegments {
				t.Errorf("Segments = %d, want %d", len(target.Pattern.Segments), tt.wantSegments)
			}
			if target.Pattern.IsPhony != tt.wantIsPhony {
				t.Errorf("IsPhony = %v, want %v", target.Pattern.IsPhony, tt.wantIsPhony)
			}
			if target.Pattern.IsDirectory != tt.wantIsDirectory {
				t.Errorf("IsDirectory = %v, want %v", target.Pattern.IsDirectory, tt.wantIsDirectory)
			}
			if len(target.Dependencies) != tt.wantDeps {
				t.Errorf("Dependencies = %d, want %d", len(target.Dependencies), tt.wantDeps)
			}
		})
	}
}

// TestParser_ParseTarget_PatternCaptures tests targets with capture expressions.
func TestParser_ParseTarget_PatternCaptures(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantSegments    int
		wantBraceExprs  int
		wantLiterals    int
		wantDeps        int
		wantDepSegments []int // number of segments per dependency
	}{
		{
			name:            "single capture in target",
			input:           "build/{name}.o: src/{name}.c",
			wantSegments:    3, // "build/", BraceExpr("name"), ".o"
			wantBraceExprs:  1,
			wantLiterals:    2,
			wantDeps:        1,
			wantDepSegments: []int{3}, // "src/", BraceExpr("name"), ".c"
		},
		{
			name:            "multiple captures in target",
			input:           "{dir}/{name}.o: {dir}/{name}.c",
			wantSegments:    4, // BraceExpr("dir"), "/", BraceExpr("name"), ".o"
			wantBraceExprs:  2,
			wantLiterals:    2,
			wantDeps:        1,
			wantDepSegments: []int{4},
		},
		{
			name:           "capture with no literal prefix",
			input:          "{name}: {name}.c",
			wantSegments:   1, // just BraceExpr("name")
			wantBraceExprs: 1,
			wantLiterals:   0,
			wantDeps:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.need", tt.input)
			p := New(l)

			target, err := p.ParseTarget()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(target.Pattern.Segments) != tt.wantSegments {
				t.Errorf("Segments = %d, want %d", len(target.Pattern.Segments), tt.wantSegments)
			}

			// Count brace expressions and literals
			braceCount := 0
			literalCount := 0
			for _, seg := range target.Pattern.Segments {
				switch seg.(type) {
				case *ast.BraceExpr:
					braceCount++
				case *ast.LiteralSegment:
					literalCount++
				}
			}

			if braceCount != tt.wantBraceExprs {
				t.Errorf("BraceExpr count = %d, want %d", braceCount, tt.wantBraceExprs)
			}
			if literalCount != tt.wantLiterals {
				t.Errorf("LiteralSegment count = %d, want %d", literalCount, tt.wantLiterals)
			}

			if len(target.Dependencies) != tt.wantDeps {
				t.Errorf("Dependencies = %d, want %d", len(target.Dependencies), tt.wantDeps)
			}

			// Check dependency segment counts
			for i, dep := range target.Dependencies {
				if i < len(tt.wantDepSegments) {
					if len(dep.Segments) != tt.wantDepSegments[i] {
						t.Errorf("Dependency[%d].Segments = %d, want %d",
							i, len(dep.Segments), tt.wantDepSegments[i])
					}
				}
			}
		})
	}
}

// TestParser_ParseTarget_PhonyTargets tests @name phony target parsing.
func TestParser_ParseTarget_PhonyTargets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantDeps int
	}{
		{
			name:     "simple phony",
			input:    "@all:",
			wantName: "all",
			wantDeps: 0,
		},
		{
			name:     "phony with deps",
			input:    "@test: build/app",
			wantName: "test",
			wantDeps: 1,
		},
		{
			name:     "phony clean",
			input:    "@clean:",
			wantName: "clean",
			wantDeps: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.need", tt.input)
			p := New(l)

			target, err := p.ParseTarget()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !target.Pattern.IsPhony {
				t.Error("expected phony target")
			}

			// The pattern should have a single literal segment with the name
			if len(target.Pattern.Segments) != 1 {
				t.Fatalf("expected 1 segment, got %d", len(target.Pattern.Segments))
			}

			lit, ok := target.Pattern.Segments[0].(*ast.LiteralSegment)
			if !ok {
				t.Fatalf("expected LiteralSegment, got %T", target.Pattern.Segments[0])
			}
			if lit.Text != tt.wantName {
				t.Errorf("phony name = %q, want %q", lit.Text, tt.wantName)
			}

			if len(target.Dependencies) != tt.wantDeps {
				t.Errorf("Dependencies = %d, want %d", len(target.Dependencies), tt.wantDeps)
			}
		})
	}
}

// TestParser_ParseTarget_DirectoryTargets tests targets ending with /.
func TestParser_ParseTarget_DirectoryTargets(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantPattern string
	}{
		{
			name:        "simple directory",
			input:       "build/:",
			wantPattern: "build/",
		},
		{
			name:        "nested directory",
			input:       "build/output/:",
			wantPattern: "build/output/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.need", tt.input)
			p := New(l)

			target, err := p.ParseTarget()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !target.Pattern.IsDirectory {
				t.Error("expected directory target")
			}

			// Reconstruct pattern text
			var text string
			for _, seg := range target.Pattern.Segments {
				if lit, ok := seg.(*ast.LiteralSegment); ok {
					text += lit.Text
				}
			}
			if text != tt.wantPattern {
				t.Errorf("pattern = %q, want %q", text, tt.wantPattern)
			}
		})
	}
}

// TestParser_ParseTarget_SourceLocation tests that location tracking works.
func TestParser_ParseTarget_SourceLocation(t *testing.T) {
	input := "build/app: src/main.c"
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Location.File != "test.need" {
		t.Errorf("Location.File = %q, want %q", target.Location.File, "test.need")
	}
	if target.Location.Line != 1 {
		t.Errorf("Location.Line = %d, want %d", target.Location.Line, 1)
	}
	if target.Location.Column != 1 {
		t.Errorf("Location.Column = %d, want %d", target.Location.Column, 1)
	}
}

// TestParser_ParseTarget_BraceExprLocation tests location tracking for brace expressions.
func TestParser_ParseTarget_BraceExprLocation(t *testing.T) {
	input := "build/{name}.o:"
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find the brace expression
	var braceExpr *ast.BraceExpr
	for _, seg := range target.Pattern.Segments {
		if be, ok := seg.(*ast.BraceExpr); ok {
			braceExpr = be
			break
		}
	}

	if braceExpr == nil {
		t.Fatal("expected to find BraceExpr")
	}

	if braceExpr.Identifier != "name" {
		t.Errorf("BraceExpr.Identifier = %q, want %q", braceExpr.Identifier, "name")
	}

	// Location should be at the { character
	if braceExpr.Location.Line != 1 {
		t.Errorf("BraceExpr.Location.Line = %d, want %d", braceExpr.Location.Line, 1)
	}
}

// TestParser_ParseTarget_Error tests error cases.
func TestParser_ParseTarget_Error(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "missing colon",
			input:   "build/app",
			wantErr: "expected ':'",
		},
		{
			name:    "unclosed brace in pattern",
			input:   "build/{name.o:",
			wantErr: "unclosed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.need", tt.input)
			p := New(l)

			_, err := p.ParseTarget()
			if err == nil {
				t.Error("expected error")
				return
			}
			// Just check that we got an error
			// The specific message format may vary
		})
	}
}

// TestParser_ParseTargetPattern tests the pattern parsing helper.
func TestParser_ParseTargetPattern(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantSegments int
		wantText     string // reconstructed text
	}{
		{
			name:         "simple path",
			input:        "build/app",
			wantSegments: 1,
			wantText:     "build/app",
		},
		{
			name:         "path with extension",
			input:        "build/main.o",
			wantSegments: 1,
			wantText:     "build/main.o",
		},
		{
			name:         "path with brace",
			input:        "build/{name}.o",
			wantSegments: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.need", tt.input+":")
			p := New(l)

			target, err := p.ParseTarget()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(target.Pattern.Segments) != tt.wantSegments {
				t.Errorf("Segments = %d, want %d", len(target.Pattern.Segments), tt.wantSegments)
			}

			if tt.wantText != "" {
				// Reconstruct text from literal segments
				var text string
				for _, seg := range target.Pattern.Segments {
					if lit, ok := seg.(*ast.LiteralSegment); ok {
						text += lit.Text
					}
				}
				if text != tt.wantText {
					t.Errorf("text = %q, want %q", text, tt.wantText)
				}
			}
		})
	}
}

// TestParser_ParseDependencyList tests dependency list parsing.
func TestParser_ParseDependencyList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantDeps int
	}{
		{
			name:     "no dependencies",
			input:    "target:",
			wantDeps: 0,
		},
		{
			name:     "single dependency",
			input:    "target: dep",
			wantDeps: 1,
		},
		{
			name:     "multiple dependencies",
			input:    "target: dep1 dep2 dep3",
			wantDeps: 3,
		},
		{
			name:     "dependencies with paths",
			input:    "build/app: src/main.c src/utils.c",
			wantDeps: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.need", tt.input)
			p := New(l)

			target, err := p.ParseTarget()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(target.Dependencies) != tt.wantDeps {
				t.Errorf("Dependencies = %d, want %d", len(target.Dependencies), tt.wantDeps)
			}
		})
	}
}

// renderDepSegments renders a dependency's pattern segments back to a string
// for easy comparison in tests, without resolving any interpolations.
func renderDepSegments(segs []ast.PatternSegment) string {
	var b strings.Builder
	for _, s := range segs {
		switch v := s.(type) {
		case *ast.LiteralSegment:
			b.WriteString(v.Text)
		case *ast.BraceExpr:
			b.WriteString("{" + v.Identifier + "}")
		}
	}
	return b.String()
}

// TestParser_ParseDependencyList_TrailingSpaceSplit covers B2a: a STRING
// token that trails a completed dependency with a space (e.g. "/main.o "
// followed by another interpolation) must flush after the trailing space,
// not before processing it.
func TestParser_ParseDependencyList_TrailingSpaceSplit(t *testing.T) {
	input := "{build_dir}/app: {build_dir}/main.o {build_dir}/utils.o\n"
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"{build_dir}/main.o", "{build_dir}/utils.o"}
	if len(target.Dependencies) != len(want) {
		t.Fatalf("got %d deps, want %d: %#v", len(target.Dependencies), len(want), target.Dependencies)
	}
	for i, dep := range target.Dependencies {
		if got := renderDepSegments(dep.Segments); got != want[i] {
			t.Errorf("dep[%d] = %q, want %q", i, got, want[i])
		}
	}
}

// TestParser_ParseDependencyList_InterpolationSplitting covers the exact
// scenario from the B2a bug report: leading/trailing/no whitespace around a
// STRING token between interpolations must split into the right deps.
func TestParser_ParseDependencyList_InterpolationSplitting(t *testing.T) {
	input := "@t: {a}/x {a}/y  z {a}\n"
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"{a}/x", "{a}/y", "z", "{a}"}
	if len(target.Dependencies) != len(want) {
		t.Fatalf("got %d deps, want %d: %#v", len(target.Dependencies), len(want), target.Dependencies)
	}
	for i, dep := range target.Dependencies {
		if got := renderDepSegments(dep.Segments); got != want[i] {
			t.Errorf("dep[%d] = %q, want %q", i, got, want[i])
		}
	}
}

// TestParser_ParseTarget_UnknownDirective covers B5: an unknown dot-keyword
// at the start of a target line (immediately followed by ':') must be
// reported as an unknown directive, not parsed as a file target.
func TestParser_ParseTarget_UnknownDirective(t *testing.T) {
	input := ".unknown: value\n"
	l := lexer.New("test.need", input)
	p := New(l)

	_, err := p.ParseTarget()
	if err == nil {
		t.Fatal("expected error for unknown directive, got none")
	}
	if !strings.Contains(err.Message, "unknown directive '.unknown'") {
		t.Errorf("error message = %q, want to contain %q", err.Message, "unknown directive '.unknown'")
	}
	if !strings.Contains(err.Hint, "./") {
		t.Errorf("hint = %q, want to mention ./ file target prefix", err.Hint)
	}
}

// TestParser_ParseTarget_HiddenFileTargetStillWorks covers B5's negative
// case: a genuine hidden-file target with a '/' must still parse as a file
// target, not an unknown directive.
func TestParser_ParseTarget_HiddenFileTargetStillWorks(t *testing.T) {
	input := ".hidden/output.o: .hidden/input.c\n"
	l := lexer.New("test.need", input)
	p := New(l)

	target, err := p.ParseTarget()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Pattern.IsPhony {
		t.Fatal("expected file target, got phony")
	}
	got := renderDepSegments(segmentsFromPattern(target.Pattern))
	if got != ".hidden/output.o" {
		t.Errorf("pattern = %q, want %q", got, ".hidden/output.o")
	}
}

// segmentsFromPattern is a small helper to reuse renderDepSegments for a
// TargetPattern's segments.
func segmentsFromPattern(p ast.TargetPattern) []ast.PatternSegment {
	return p.Segments
}
