package parser

import (
	"testing"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/lexer"
)

func TestParser_ParseVariable_Simple(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantLazy    bool
		wantParts   int
		wantLiteral string
	}{
		{
			name:        "simple assignment",
			input:       "cc = gcc",
			wantName:    "cc",
			wantLazy:    false,
			wantParts:   1,
			wantLiteral: "gcc",
		},
		{
			name:        "assignment with spaces in value",
			input:       "cflags = -Wall -O2",
			wantName:    "cflags",
			wantLazy:    false,
			wantParts:   1,
			wantLiteral: "-Wall -O2",
		},
		{
			name:        "lazy assignment",
			input:       "lazy all_flags = -Wall",
			wantName:    "all_flags",
			wantLazy:    true,
			wantParts:   1,
			wantLiteral: "-Wall",
		},
		{
			name:        "empty value",
			input:       "empty = ",
			wantName:    "empty",
			wantLazy:    false,
			wantParts:   0,
			wantLiteral: "",
		},
		{
			name:        "path value",
			input:       "src_dir = src/main",
			wantName:    "src_dir",
			wantLazy:    false,
			wantParts:   1,
			wantLiteral: "src/main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.need", tt.input)
			p := New(l)

			v, err := p.ParseVariable()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if v.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", v.Name, tt.wantName)
			}
			if v.Lazy != tt.wantLazy {
				t.Errorf("Lazy = %v, want %v", v.Lazy, tt.wantLazy)
			}
			if len(v.Value.Parts) != tt.wantParts {
				t.Errorf("Parts count = %d, want %d", len(v.Value.Parts), tt.wantParts)
			}
			if tt.wantParts > 0 {
				if lit, ok := v.Value.Parts[0].(*ast.LiteralValue); ok {
					if lit.Text != tt.wantLiteral {
						t.Errorf("Literal = %q, want %q", lit.Text, tt.wantLiteral)
					}
				}
			}
		})
	}
}

func TestParser_ParseVariable_WithInterpolation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantName       string
		wantPartTypes  []string // "literal", "interp", "func"
		wantInterpName string
		wantInterpRaw  bool
	}{
		{
			name:           "interpolation in value",
			input:          "all_flags = {cflags} {extra}",
			wantName:       "all_flags",
			wantPartTypes:  []string{"interp", "literal", "interp"}, // space is a literal
			wantInterpName: "cflags",
		},
		{
			name:           "raw modifier",
			input:          "flags = {opts:raw}",
			wantName:       "flags",
			wantPartTypes:  []string{"interp"},
			wantInterpName: "opts",
			wantInterpRaw:  true,
		},
		{
			name:          "mixed literal and interpolation",
			input:         "cmd = gcc {flags} -o out",
			wantName:      "cmd",
			wantPartTypes: []string{"literal", "interp", "literal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.need", tt.input)
			p := New(l)

			v, err := p.ParseVariable()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if v.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", v.Name, tt.wantName)
			}

			if len(v.Value.Parts) != len(tt.wantPartTypes) {
				t.Fatalf("Parts count = %d, want %d", len(v.Value.Parts), len(tt.wantPartTypes))
			}

			for i, partType := range tt.wantPartTypes {
				switch partType {
				case "literal":
					if _, ok := v.Value.Parts[i].(*ast.LiteralValue); !ok {
						t.Errorf("Part[%d] is not LiteralValue", i)
					}
				case "interp":
					interp, ok := v.Value.Parts[i].(*ast.Interpolation)
					if !ok {
						t.Errorf("Part[%d] is not Interpolation", i)
					} else if i == 0 && tt.wantInterpName != "" {
						if interp.Name != tt.wantInterpName {
							t.Errorf("Interpolation.Name = %q, want %q", interp.Name, tt.wantInterpName)
						}
						if interp.Raw != tt.wantInterpRaw {
							t.Errorf("Interpolation.Raw = %v, want %v", interp.Raw, tt.wantInterpRaw)
						}
					}
				case "func":
					if _, ok := v.Value.Parts[i].(*ast.FunctionCall); !ok {
						t.Errorf("Part[%d] is not FunctionCall", i)
					}
				}
			}
		})
	}
}

func TestParser_ParseVariable_WithFunctionCall(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantName     string
		wantFuncName ast.FunctionName
		wantArgCount int
	}{
		{
			name:         "shell function",
			input:        "sources = shell(find src -name *.c)",
			wantName:     "sources",
			wantFuncName: ast.FuncShell,
			wantArgCount: 1,
		},
		{
			name:         "glob function",
			input:        "files = glob(src/*.c)",
			wantName:     "files",
			wantFuncName: ast.FuncGlob,
			wantArgCount: 1,
		},
		{
			name:         "filename function",
			input:        "name = filename(src/main.c)",
			wantName:     "name",
			wantFuncName: ast.FuncFilename,
			wantArgCount: 1,
		},
		{
			name:         "dirname function",
			input:        "dir = dirname(src/main.c)",
			wantName:     "dir",
			wantFuncName: ast.FuncDirname,
			wantArgCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.need", tt.input)
			p := New(l)

			v, err := p.ParseVariable()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if v.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", v.Name, tt.wantName)
			}

			if len(v.Value.Parts) < 1 {
				t.Fatal("expected at least one part")
			}

			funcCall, ok := v.Value.Parts[0].(*ast.FunctionCall)
			if !ok {
				t.Fatalf("Part[0] is not FunctionCall, got %T", v.Value.Parts[0])
			}

			if funcCall.Name != tt.wantFuncName {
				t.Errorf("FunctionCall.Name = %v, want %v", funcCall.Name, tt.wantFuncName)
			}

			if len(funcCall.Args) != tt.wantArgCount {
				t.Errorf("FunctionCall.Args count = %d, want %d", len(funcCall.Args), tt.wantArgCount)
			}
		})
	}
}

func TestParser_ParseVariable_WithInterpolationInFunction(t *testing.T) {
	input := "sources = shell(find {src_dir} -name *.c)"
	l := lexer.New("test.need", input)
	p := New(l)

	v, err := p.ParseVariable()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v.Name != "sources" {
		t.Errorf("Name = %q, want %q", v.Name, "sources")
	}

	if len(v.Value.Parts) < 1 {
		t.Fatal("expected at least one part")
	}

	funcCall, ok := v.Value.Parts[0].(*ast.FunctionCall)
	if !ok {
		t.Fatalf("Part[0] is not FunctionCall, got %T", v.Value.Parts[0])
	}

	if funcCall.Name != ast.FuncShell {
		t.Errorf("FunctionCall.Name = %v, want FuncShell", funcCall.Name)
	}

	// Check that the argument contains an interpolation
	if len(funcCall.Args) < 1 {
		t.Fatal("expected function to have argument")
	}

	foundInterp := false
	for _, part := range funcCall.Args[0].Parts {
		if _, ok := part.(*ast.Interpolation); ok {
			foundInterp = true
			break
		}
	}
	if !foundInterp {
		t.Error("expected interpolation in function argument")
	}
}

func TestParser_ParseVariable_EscapedBraces(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantLiteral string
	}{
		{
			name:     "escaped open brace",
			input:    "json = {{key}}",
			wantName: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test.need", tt.input)
			p := New(l)

			v, err := p.ParseVariable()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if v.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", v.Name, tt.wantName)
			}

			// The escaped braces should be converted to literal { and }
			// Check that we have literal parts
			hasLiteralBrace := false
			for _, part := range v.Value.Parts {
				if lit, ok := part.(*ast.LiteralValue); ok {
					if lit.Text == "{" || lit.Text == "}" {
						hasLiteralBrace = true
					}
				}
			}
			if !hasLiteralBrace {
				t.Error("expected literal brace from escape sequence")
			}
		})
	}
}

func TestParser_ParseVariable_SourceLocation(t *testing.T) {
	input := "cc = gcc"
	l := lexer.New("test.need", input)
	p := New(l)

	v, err := p.ParseVariable()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v.Location.File != "test.need" {
		t.Errorf("Location.File = %q, want %q", v.Location.File, "test.need")
	}
	if v.Location.Line != 1 {
		t.Errorf("Location.Line = %d, want %d", v.Location.Line, 1)
	}
	if v.Location.Column != 1 {
		t.Errorf("Location.Column = %d, want %d", v.Location.Column, 1)
	}
}

func TestParser_ParseVariable_Error_MissingEquals(t *testing.T) {
	input := "cc gcc"
	l := lexer.New("test.need", input)
	p := New(l)

	_, err := p.ParseVariable()
	if err == nil {
		t.Error("expected error for missing equals")
	}
}

func TestParser_ParseVariable_Error_MissingIdentifier(t *testing.T) {
	// Start with equals - should error on missing identifier
	input := "= gcc"
	l := lexer.New("test.need", input)
	p := New(l)

	_, err := p.ParseVariable()
	if err == nil {
		t.Error("expected error for missing identifier")
	}
}

func TestParser_ParseValue_Empty(t *testing.T) {
	// Test parsing empty value (just newline)
	input := "\n"
	l := lexer.New("test.need", input)
	p := New(l)

	v := p.ParseValue()
	if len(v.Parts) != 0 {
		t.Errorf("Parts count = %d, want 0", len(v.Parts))
	}
}

func TestParser_ParseValue_WithComment(t *testing.T) {
	// Value parsing should stop at comment
	input := "gcc # this is a comment"
	l := lexer.New("test.need", input)
	p := New(l)

	v := p.ParseValue()
	if len(v.Parts) != 1 {
		t.Fatalf("Parts count = %d, want 1", len(v.Parts))
	}

	lit, ok := v.Parts[0].(*ast.LiteralValue)
	if !ok {
		t.Fatalf("Part[0] is not LiteralValue")
	}
	// The literal should not include the comment
	if lit.Text == "gcc # this is a comment" {
		t.Error("comment should not be included in value")
	}
}

func TestParser_ParseInterpolation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantRaw  bool
	}{
		{
			name:     "simple interpolation",
			input:    "{var}",
			wantName: "var",
			wantRaw:  false,
		},
		{
			name:     "interpolation with raw modifier",
			input:    "{flags:raw}",
			wantName: "flags",
			wantRaw:  true,
		},
		{
			name:     "dotted identifier",
			input:    "{target.dir}",
			wantName: "target.dir",
			wantRaw:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Wrap in assignment to get proper value parsing context
			input := "x = " + tt.input
			l := lexer.New("test.need", input)
			p := New(l)

			v, err := p.ParseVariable()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(v.Value.Parts) < 1 {
				t.Fatal("expected at least one part")
			}

			interp, ok := v.Value.Parts[0].(*ast.Interpolation)
			if !ok {
				t.Fatalf("Part[0] is not Interpolation, got %T", v.Value.Parts[0])
			}

			if interp.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", interp.Name, tt.wantName)
			}
			if interp.Raw != tt.wantRaw {
				t.Errorf("Raw = %v, want %v", interp.Raw, tt.wantRaw)
			}
		})
	}
}

func TestParser_ParseFunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantFunc ast.FunctionName
		wantArgs int
	}{
		{
			name:     "shell function",
			input:    "shell(echo hello)",
			wantFunc: ast.FuncShell,
			wantArgs: 1,
		},
		{
			name:     "glob function",
			input:    "glob(*.go)",
			wantFunc: ast.FuncGlob,
			wantArgs: 1,
		},
		{
			name:     "filename function",
			input:    "filename(path/to/file.txt)",
			wantFunc: ast.FuncFilename,
			wantArgs: 1,
		},
		{
			name:     "dirname function",
			input:    "dirname(path/to/file.txt)",
			wantFunc: ast.FuncDirname,
			wantArgs: 1,
		},
		{
			name:     "replace function with three args",
			input:    "replace({sources}, .c, .o)",
			wantFunc: ast.FuncReplace,
			wantArgs: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Wrap in assignment
			input := "x = " + tt.input
			l := lexer.New("test.need", input)
			p := New(l)

			v, err := p.ParseVariable()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(v.Value.Parts) < 1 {
				t.Fatal("expected at least one part")
			}

			funcCall, ok := v.Value.Parts[0].(*ast.FunctionCall)
			if !ok {
				t.Fatalf("Part[0] is not FunctionCall, got %T", v.Value.Parts[0])
			}

			if funcCall.Name != tt.wantFunc {
				t.Errorf("Name = %v, want %v", funcCall.Name, tt.wantFunc)
			}

			if len(funcCall.Args) != tt.wantArgs {
				t.Errorf("Args count = %d, want %d", len(funcCall.Args), tt.wantArgs)
			}
		})
	}
}

func TestParser_ParseFunctionCall_NestedParentheses(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantFunc  ast.FunctionName
		wantArgs  int
		wantInArg string // expected text within argument
	}{
		{
			name:      "shell with subshell",
			input:     "shell(echo $(date))",
			wantFunc:  ast.FuncShell,
			wantArgs:  1,
			wantInArg: "$(date)",
		},
		{
			name:      "shell with arithmetic expansion",
			input:     "shell(echo $((1 + 2)))",
			wantFunc:  ast.FuncShell,
			wantArgs:  1,
			wantInArg: "$((1 + 2))",
		},
		{
			name:      "shell with nested command",
			input:     "shell(bash -c 'echo (test)')",
			wantFunc:  ast.FuncShell,
			wantArgs:  1,
			wantInArg: "(test)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "x = " + tt.input
			l := lexer.New("test.need", input)
			p := New(l)

			v, err := p.ParseVariable()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(v.Value.Parts) < 1 {
				t.Fatal("expected at least one part")
			}

			funcCall, ok := v.Value.Parts[0].(*ast.FunctionCall)
			if !ok {
				t.Fatalf("Part[0] is not FunctionCall, got %T", v.Value.Parts[0])
			}

			if funcCall.Name != tt.wantFunc {
				t.Errorf("Name = %v, want %v", funcCall.Name, tt.wantFunc)
			}

			if len(funcCall.Args) != tt.wantArgs {
				t.Errorf("Args count = %d, want %d", len(funcCall.Args), tt.wantArgs)
			}

			// Check that the nested parentheses are preserved in the argument
			if len(funcCall.Args) > 0 {
				argText := ""
				for _, part := range funcCall.Args[0].Parts {
					if lit, ok := part.(*ast.LiteralValue); ok {
						argText += lit.Text
					}
				}
				if !containsSubstring(argText, tt.wantInArg) {
					t.Errorf("Argument text %q does not contain %q", argText, tt.wantInArg)
				}
			}
		})
	}
}

func TestParser_ParseFunctionCall_ReplaceWithMultipleArgs(t *testing.T) {
	input := "objs = replace({sources}, .c, .o)"
	l := lexer.New("test.need", input)
	p := New(l)

	v, err := p.ParseVariable()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v.Name != "objs" {
		t.Errorf("Name = %q, want %q", v.Name, "objs")
	}

	if len(v.Value.Parts) < 1 {
		t.Fatal("expected at least one part")
	}

	funcCall, ok := v.Value.Parts[0].(*ast.FunctionCall)
	if !ok {
		t.Fatalf("Part[0] is not FunctionCall, got %T", v.Value.Parts[0])
	}

	if funcCall.Name != ast.FuncReplace {
		t.Errorf("Name = %v, want FuncReplace", funcCall.Name)
	}

	if len(funcCall.Args) != 3 {
		t.Fatalf("Args count = %d, want 3", len(funcCall.Args))
	}

	// First arg should contain interpolation
	hasInterp := false
	for _, part := range funcCall.Args[0].Parts {
		if _, ok := part.(*ast.Interpolation); ok {
			hasInterp = true
			break
		}
	}
	if !hasInterp {
		t.Error("expected first argument to contain interpolation")
	}

	// Second arg should be ".c"
	if len(funcCall.Args[1].Parts) < 1 {
		t.Fatal("expected second argument to have parts")
	}
	lit, ok := funcCall.Args[1].Parts[0].(*ast.LiteralValue)
	if !ok || !containsSubstring(lit.Text, ".c") {
		t.Errorf("second arg should contain '.c', got %v", funcCall.Args[1].Parts)
	}

	// Third arg should be ".o"
	if len(funcCall.Args[2].Parts) < 1 {
		t.Fatal("expected third argument to have parts")
	}
	lit, ok = funcCall.Args[2].Parts[0].(*ast.LiteralValue)
	if !ok || !containsSubstring(lit.Text, ".o") {
		t.Errorf("third arg should contain '.o', got %v", funcCall.Args[2].Parts)
	}
}

// containsSubstring is a helper for checking substrings
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && contains(s, substr)))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestParser_ParseValue_UnclosedInterpolation covers B4: an unclosed
// interpolation must surface a parse error (via p.Errors()) with the
// lexer's diagnostic message, not be silently absorbed as literal text.
func TestParser_ParseValue_UnclosedInterpolation(t *testing.T) {
	input := "msg = hello {world\n"
	l := lexer.New("test.need", input)
	p := New(l)

	_, errs := p.ParseNeedfile()
	if !errs.HasErrors() {
		t.Fatal("expected a parse error for unclosed interpolation, got none")
	}
	found := false
	for _, e := range errs.Errors {
		if e.Message == "unclosed interpolation: {world" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected an error with message %q, got %v", "unclosed interpolation: {world", errs.Errors)
	}
}
