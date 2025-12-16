// Package ast_test provides tests for the AST node types.
package ast_test

import (
	"testing"

	"github.com/vinayprograms/build/internal/ast"
	"github.com/vinayprograms/build/internal/lexer"
)

func TestSourceLocation(t *testing.T) {
	loc := ast.SourceLocation{
		File:   "Buildfile",
		Line:   10,
		Column: 5,
	}

	want := "Buildfile:10:5"
	if got := loc.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestSourceLocationFromToken(t *testing.T) {
	tok := lexer.Token{
		Type:    lexer.IDENTIFIER,
		Literal: "foo",
		Location: lexer.SourceLocation{
			File:   "test.build",
			Line:   3,
			Column: 7,
		},
	}

	loc := ast.SourceLocationFromToken(tok)

	if loc.File != "test.build" {
		t.Errorf("File = %q, want %q", loc.File, "test.build")
	}
	if loc.Line != 3 {
		t.Errorf("Line = %d, want %d", loc.Line, 3)
	}
	if loc.Column != 7 {
		t.Errorf("Column = %d, want %d", loc.Column, 7)
	}
}

func TestBuildfile(t *testing.T) {
	bf := &ast.Buildfile{
		SourcePath: "Buildfile",
		Statements: []ast.Statement{},
	}

	if bf.SourcePath != "Buildfile" {
		t.Errorf("SourcePath = %q, want %q", bf.SourcePath, "Buildfile")
	}
	if len(bf.Statements) != 0 {
		t.Errorf("len(Statements) = %d, want %d", len(bf.Statements), 0)
	}
}

func TestDirectiveKind(t *testing.T) {
	tests := []struct {
		kind ast.DirectiveKind
		want string
	}{
		{ast.DirectiveShell, "shell"},
		{ast.DirectiveParallel, "parallel"},
		{ast.DirectiveDefault, "default"},
		{ast.DirectiveInclude, "include"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDirective(t *testing.T) {
	dir := &ast.Directive{
		Kind: ast.DirectiveShell,
		Value: &ast.Value{
			Parts: []ast.ValuePart{
				&ast.LiteralValue{Text: "bash"},
			},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	}

	if dir.Kind != ast.DirectiveShell {
		t.Errorf("Kind = %v, want %v", dir.Kind, ast.DirectiveShell)
	}
	// Verify it implements Statement interface
	var _ ast.Statement = dir
}

func TestRuntime(t *testing.T) {
	tests := []struct {
		runtime ast.Runtime
		want    string
	}{
		{ast.RuntimeBare, "bare"},
		{ast.RuntimeDocker, "docker"},
		{ast.RuntimePodman, "podman"},
		{ast.RuntimeDevcontainer, "devcontainer"},
		{ast.RuntimeNix, "nix"},
		{ast.RuntimeLima, "lima"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.runtime.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEnvironment(t *testing.T) {
	env := &ast.Environment{
		Name:    stringPtr("ci"),
		Runtime: runtimePtr(ast.RuntimeDocker),
		Source: &ast.Value{
			Parts: []ast.ValuePart{
				&ast.LiteralValue{Text: "./docker/ci.Dockerfile"},
			},
		},
		Args: &ast.Value{
			Parts: []ast.ValuePart{
				&ast.LiteralValue{Text: "--platform linux/amd64"},
			},
		},
		Requires: []ast.Requirement{
			{Name: "gcc", Version: ast.VersionLatest{}},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 5, Column: 1},
	}

	if *env.Name != "ci" {
		t.Errorf("Name = %q, want %q", *env.Name, "ci")
	}
	if *env.Runtime != ast.RuntimeDocker {
		t.Errorf("Runtime = %v, want %v", *env.Runtime, ast.RuntimeDocker)
	}
	// Verify it implements Statement interface
	var _ ast.Statement = env
}

func TestVersionSpec(t *testing.T) {
	tests := []struct {
		name    string
		version ast.VersionSpec
		want    string
	}{
		{"latest", ast.VersionLatest{}, "latest"},
		{"major", ast.VersionMajor{Major: 11}, "11"},
		{"major.minor", ast.VersionMajorMinor{Major: 11, Minor: 4}, "11.4"},
		{"exact", ast.VersionExact{Major: 11, Minor: 4, Patch: 0}, "11.4.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVariable(t *testing.T) {
	v := &ast.Variable{
		Name: "cc",
		Value: &ast.Value{
			Parts: []ast.ValuePart{
				&ast.LiteralValue{Text: "gcc"},
			},
		},
		Lazy:     false,
		Location: ast.SourceLocation{File: "Buildfile", Line: 10, Column: 1},
	}

	if v.Name != "cc" {
		t.Errorf("Name = %q, want %q", v.Name, "cc")
	}
	if v.Lazy != false {
		t.Errorf("Lazy = %v, want %v", v.Lazy, false)
	}
	// Verify it implements Statement interface
	var _ ast.Statement = v
}

func TestLazyVariable(t *testing.T) {
	v := &ast.Variable{
		Name: "all_flags",
		Value: &ast.Value{
			Parts: []ast.ValuePart{
				&ast.Interpolation{Name: "cflags", Raw: false},
				&ast.LiteralValue{Text: " "},
				&ast.Interpolation{Name: "extra_flags", Raw: false},
			},
		},
		Lazy:     true,
		Location: ast.SourceLocation{File: "Buildfile", Line: 15, Column: 1},
	}

	if v.Lazy != true {
		t.Errorf("Lazy = %v, want %v", v.Lazy, true)
	}
}

func TestConditional(t *testing.T) {
	cond := &ast.Conditional{
		IfBranch: ast.ConditionalBranch{
			Condition: &ast.EqualsCondition{
				Left:  &ast.Value{Parts: []ast.ValuePart{&ast.Interpolation{Name: "os"}}},
				Right: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "linux"}}},
			},
			Body: []ast.Statement{
				&ast.Variable{Name: "ldflags", Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "-lpthread"}}}},
			},
		},
		ElifBranches: []ast.ConditionalBranch{
			{
				Condition: &ast.EqualsCondition{
					Left:  &ast.Value{Parts: []ast.ValuePart{&ast.Interpolation{Name: "os"}}},
					Right: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "darwin"}}},
				},
				Body: []ast.Statement{
					&ast.Variable{Name: "ldflags", Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "-framework CoreFoundation"}}}},
				},
			},
		},
		ElseBody: []ast.Statement{
			&ast.Variable{Name: "ldflags", Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: ""}}}},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 20, Column: 1},
	}

	// Verify it implements Statement interface
	var _ ast.Statement = cond
	if len(cond.ElifBranches) != 1 {
		t.Errorf("len(ElifBranches) = %d, want %d", len(cond.ElifBranches), 1)
	}
	if len(cond.ElseBody) != 1 {
		t.Errorf("len(ElseBody) = %d, want %d", len(cond.ElseBody), 1)
	}
}

func TestConditionTypes(t *testing.T) {
	t.Run("EqualsCondition", func(t *testing.T) {
		cond := &ast.EqualsCondition{
			Left:  &ast.Value{Parts: []ast.ValuePart{&ast.Interpolation{Name: "os"}}},
			Right: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "linux"}}},
		}
		// Verify it implements Condition interface
		var _ ast.Condition = cond
	})

	t.Run("NotEqualsCondition", func(t *testing.T) {
		cond := &ast.NotEqualsCondition{
			Left:  &ast.Value{Parts: []ast.ValuePart{&ast.Interpolation{Name: "os"}}},
			Right: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "windows"}}},
		}
		// Verify it implements Condition interface
		var _ ast.Condition = cond
	})

	t.Run("DefinedCondition", func(t *testing.T) {
		cond := &ast.DefinedCondition{Name: "DEBUG"}
		// Verify it implements Condition interface
		var _ ast.Condition = cond
	})

	t.Run("NotDefinedCondition", func(t *testing.T) {
		cond := &ast.NotDefinedCondition{Name: "CC"}
		// Verify it implements Condition interface
		var _ ast.Condition = cond
	})
}

func TestTarget(t *testing.T) {
	target := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments:    []ast.PatternSegment{&ast.LiteralSegment{Text: "build/app"}},
			IsPhony:     false,
			IsDirectory: false,
		},
		Dependencies: []ast.Dependency{
			{Segments: []ast.PatternSegment{&ast.LiteralSegment{Text: "build/main.o"}}},
			{Segments: []ast.PatternSegment{&ast.LiteralSegment{Text: "build/utils.o"}}},
		},
		Recipe:   nil,
		Location: ast.SourceLocation{File: "Buildfile", Line: 30, Column: 1},
	}

	if target.Pattern.IsPhony {
		t.Error("IsPhony should be false")
	}
	if len(target.Dependencies) != 2 {
		t.Errorf("len(Dependencies) = %d, want %d", len(target.Dependencies), 2)
	}
	// Verify it implements Statement interface
	var _ ast.Statement = target
}

func TestPhonyTarget(t *testing.T) {
	target := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments:    []ast.PatternSegment{&ast.LiteralSegment{Text: "all"}},
			IsPhony:     true,
			IsDirectory: false,
		},
		Dependencies: []ast.Dependency{
			{Segments: []ast.PatternSegment{&ast.LiteralSegment{Text: "build/app"}}},
		},
		Recipe:   nil,
		Location: ast.SourceLocation{File: "Buildfile", Line: 25, Column: 1},
	}

	if !target.Pattern.IsPhony {
		t.Error("IsPhony should be true")
	}
}

func TestDirectoryTarget(t *testing.T) {
	target := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments:    []ast.PatternSegment{&ast.LiteralSegment{Text: "build/"}},
			IsPhony:     false,
			IsDirectory: true,
		},
		Dependencies: []ast.Dependency{},
		Recipe: &ast.Recipe{
			Commands: []ast.Command{
				&ast.LineCommand{Parts: []ast.CommandPart{
					&ast.LiteralCommand{Text: "mkdir -p "},
					&ast.CommandInterpolation{Name: "target"},
				}},
			},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 35, Column: 1},
	}

	if !target.Pattern.IsDirectory {
		t.Error("IsDirectory should be true")
	}
}

func TestPatternTarget(t *testing.T) {
	target := &ast.Target{
		Pattern: ast.TargetPattern{
			Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "build/"},
				&ast.BraceExpr{Identifier: "name"},
				&ast.LiteralSegment{Text: ".o"},
			},
			IsPhony:     false,
			IsDirectory: false,
		},
		Dependencies: []ast.Dependency{
			{Segments: []ast.PatternSegment{
				&ast.LiteralSegment{Text: "src/"},
				&ast.BraceExpr{Identifier: "name"},
				&ast.LiteralSegment{Text: ".c"},
			}},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 40, Column: 1},
	}

	if len(target.Pattern.Segments) != 3 {
		t.Errorf("len(Pattern.Segments) = %d, want %d", len(target.Pattern.Segments), 3)
	}

	// Check BraceExpr segment
	braceExpr, ok := target.Pattern.Segments[1].(*ast.BraceExpr)
	if !ok {
		t.Fatal("Expected BraceExpr segment")
	}
	if braceExpr.Identifier != "name" {
		t.Errorf("BraceExpr.Identifier = %q, want %q", braceExpr.Identifier, "name")
	}
}

func TestPatternSegments(t *testing.T) {
	t.Run("LiteralSegment", func(t *testing.T) {
		seg := &ast.LiteralSegment{Text: "build/"}
		// Verify it implements PatternSegment interface
		var _ ast.PatternSegment = seg
	})

	t.Run("BraceExpr", func(t *testing.T) {
		seg := &ast.BraceExpr{Identifier: "name"}
		// Verify it implements PatternSegment interface
		var _ ast.PatternSegment = seg
	})
}

func TestRecipe(t *testing.T) {
	recipe := &ast.Recipe{
		Directives: ast.RecipeDirectives{
			Shell: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "bash"}}},
			After: []*ast.Value{
				{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "build/"}}},
			},
			Autodeps: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "build/app.d"}}},
			Requires: []ast.Requirement{
				{Name: "pkg-config", Version: ast.VersionLatest{}},
			},
		},
		Commands: []ast.Command{
			&ast.LineCommand{Parts: []ast.CommandPart{
				&ast.LiteralCommand{Text: "gcc -o "},
				&ast.CommandInterpolation{Name: "target"},
				&ast.LiteralCommand{Text: " "},
				&ast.CommandInterpolation{Name: "deps"},
			}},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 31, Column: 5},
	}

	if recipe.Directives.Shell == nil {
		t.Error("Shell directive should not be nil")
	}
	if len(recipe.Directives.After) != 1 {
		t.Errorf("len(After) = %d, want %d", len(recipe.Directives.After), 1)
	}
	if len(recipe.Commands) != 1 {
		t.Errorf("len(Commands) = %d, want %d", len(recipe.Commands), 1)
	}
}

func TestBlockCommand(t *testing.T) {
	block := &ast.BlockCommand{
		Lines: [][]ast.CommandPart{
			{
				&ast.LiteralCommand{Text: "if [[ -f "},
				&ast.CommandInterpolation{Name: "target"},
				&ast.LiteralCommand{Text: " ]]; then"},
			},
			{
				&ast.LiteralCommand{Text: "    rm "},
				&ast.CommandInterpolation{Name: "target"},
			},
			{
				&ast.LiteralCommand{Text: "fi"},
			},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 45, Column: 5},
	}

	// Verify it implements Command interface
	var _ ast.Command = block
	if len(block.Lines) != 3 {
		t.Errorf("len(Lines) = %d, want %d", len(block.Lines), 3)
	}
}

func TestCommandParts(t *testing.T) {
	t.Run("LiteralCommand", func(t *testing.T) {
		part := &ast.LiteralCommand{Text: "gcc -c "}
		// Verify it implements CommandPart interface
		var _ ast.CommandPart = part
	})

	t.Run("CommandInterpolation", func(t *testing.T) {
		part := &ast.CommandInterpolation{Name: "target", Raw: false}
		// Verify it implements CommandPart interface
		var _ ast.CommandPart = part
	})

	t.Run("CommandInterpolation with raw", func(t *testing.T) {
		part := &ast.CommandInterpolation{Name: "flags", Raw: true}
		if !part.Raw {
			t.Error("Raw should be true")
		}
	})
}

func TestValueParts(t *testing.T) {
	t.Run("LiteralValue", func(t *testing.T) {
		part := &ast.LiteralValue{Text: "gcc"}
		// Verify it implements ValuePart interface
		var _ ast.ValuePart = part
	})

	t.Run("Interpolation", func(t *testing.T) {
		part := &ast.Interpolation{Name: "cc", Raw: false}
		// Verify it implements ValuePart interface
		var _ ast.ValuePart = part
	})

	t.Run("Interpolation with raw", func(t *testing.T) {
		part := &ast.Interpolation{Name: "flags", Raw: true}
		if !part.Raw {
			t.Error("Raw should be true")
		}
	})

	t.Run("FunctionCall", func(t *testing.T) {
		part := &ast.FunctionCall{
			Name: ast.FuncShell,
			Args: []*ast.Value{
				{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "find src -name \"*.c\""}}},
			},
		}
		// Verify it implements ValuePart interface
		var _ ast.ValuePart = part
	})
}

func TestFunctionName(t *testing.T) {
	tests := []struct {
		fn   ast.FunctionName
		want string
	}{
		{ast.FuncShell, "shell"},
		{ast.FuncGlob, "glob"},
		{ast.FuncBasename, "basename"},
		{ast.FuncDirname, "dirname"},
		{ast.FuncReplace, "replace"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.fn.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComment(t *testing.T) {
	c := &ast.Comment{
		Text:     "# This is a comment",
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	}

	if c.Text != "# This is a comment" {
		t.Errorf("Text = %q, want %q", c.Text, "# This is a comment")
	}
	// Verify it implements Statement interface
	var _ ast.Statement = c
}

func TestBlank(t *testing.T) {
	b := &ast.Blank{
		Location: ast.SourceLocation{File: "Buildfile", Line: 5, Column: 1},
	}

	// Verify it implements Statement interface
	var _ ast.Statement = b
}

func TestComplexValue(t *testing.T) {
	// Test: sources = shell(find {src_dir} -name "*.c")
	v := &ast.Value{
		Parts: []ast.ValuePart{
			&ast.FunctionCall{
				Name: ast.FuncShell,
				Args: []*ast.Value{
					{Parts: []ast.ValuePart{
						&ast.LiteralValue{Text: "find "},
						&ast.Interpolation{Name: "src_dir", Raw: false},
						&ast.LiteralValue{Text: " -name \"*.c\""},
					}},
				},
			},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 20, Column: 11},
	}

	if len(v.Parts) != 1 {
		t.Errorf("len(Parts) = %d, want %d", len(v.Parts), 1)
	}

	fc, ok := v.Parts[0].(*ast.FunctionCall)
	if !ok {
		t.Fatal("Expected FunctionCall")
	}
	if fc.Name != ast.FuncShell {
		t.Errorf("FunctionCall.Name = %v, want %v", fc.Name, ast.FuncShell)
	}
	if len(fc.Args) != 1 {
		t.Errorf("len(Args) = %d, want %d", len(fc.Args), 1)
	}
	if len(fc.Args[0].Parts) != 3 {
		t.Errorf("len(Args[0].Parts) = %d, want %d", len(fc.Args[0].Parts), 3)
	}
}

func TestLineCommand(t *testing.T) {
	cmd := &ast.LineCommand{
		Parts: []ast.CommandPart{
			&ast.LiteralCommand{Text: "gcc -o "},
			&ast.CommandInterpolation{Name: "target"},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 32, Column: 5},
	}

	// Verify it implements Command interface
	var _ ast.Command = cmd
	if len(cmd.Parts) != 2 {
		t.Errorf("len(Parts) = %d, want %d", len(cmd.Parts), 2)
	}
}

func TestBuildfileWithStatements(t *testing.T) {
	// Build a complete Buildfile AST
	bf := &ast.Buildfile{
		SourcePath: "Buildfile",
		Statements: []ast.Statement{
			&ast.Directive{
				Kind:     ast.DirectiveShell,
				Value:    &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "bash"}}},
				Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
			},
			&ast.Variable{
				Name:     "cc",
				Value:    &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "gcc"}}},
				Lazy:     false,
				Location: ast.SourceLocation{File: "Buildfile", Line: 3, Column: 1},
			},
			&ast.Target{
				Pattern: ast.TargetPattern{
					Segments:    []ast.PatternSegment{&ast.LiteralSegment{Text: "all"}},
					IsPhony:     true,
					IsDirectory: false,
				},
				Dependencies: []ast.Dependency{
					{Segments: []ast.PatternSegment{&ast.LiteralSegment{Text: "build/app"}}},
				},
				Recipe:   nil,
				Location: ast.SourceLocation{File: "Buildfile", Line: 5, Column: 1},
			},
			&ast.Target{
				Pattern: ast.TargetPattern{
					Segments:    []ast.PatternSegment{&ast.LiteralSegment{Text: "build/app"}},
					IsPhony:     false,
					IsDirectory: false,
				},
				Dependencies: []ast.Dependency{
					{Segments: []ast.PatternSegment{&ast.LiteralSegment{Text: "build/main.o"}}},
				},
				Recipe: &ast.Recipe{
					Commands: []ast.Command{
						&ast.LineCommand{Parts: []ast.CommandPart{
							&ast.CommandInterpolation{Name: "cc"},
							&ast.LiteralCommand{Text: " -o "},
							&ast.CommandInterpolation{Name: "target"},
							&ast.LiteralCommand{Text: " "},
							&ast.CommandInterpolation{Name: "deps"},
						}},
					},
				},
				Location: ast.SourceLocation{File: "Buildfile", Line: 7, Column: 1},
			},
		},
	}

	if len(bf.Statements) != 4 {
		t.Errorf("len(Statements) = %d, want %d", len(bf.Statements), 4)
	}

	// Verify type assertions work correctly
	_, ok := bf.Statements[0].(*ast.Directive)
	if !ok {
		t.Error("Expected first statement to be Directive")
	}
	_, ok = bf.Statements[1].(*ast.Variable)
	if !ok {
		t.Error("Expected second statement to be Variable")
	}
	_, ok = bf.Statements[2].(*ast.Target)
	if !ok {
		t.Error("Expected third statement to be Target")
	}
	_, ok = bf.Statements[3].(*ast.Target)
	if !ok {
		t.Error("Expected fourth statement to be Target")
	}
}

// ----------------------------------------------------------------------------
// String() method tests for debugging
// ----------------------------------------------------------------------------

func TestDirectiveString(t *testing.T) {
	dir := &ast.Directive{
		Kind: ast.DirectiveShell,
		Value: &ast.Value{
			Parts: []ast.ValuePart{&ast.LiteralValue{Text: "bash"}},
		},
		Location: ast.SourceLocation{File: "Buildfile", Line: 1, Column: 1},
	}

	got := dir.String()
	if got != ".shell: bash" {
		t.Errorf("Directive.String() = %q, want %q", got, ".shell: bash")
	}
}

func TestEnvironmentString(t *testing.T) {
	tests := []struct {
		name string
		env  *ast.Environment
		want string
	}{
		{
			name: "named environment",
			env: &ast.Environment{
				Name:    stringPtr("ci"),
				Runtime: runtimePtr(ast.RuntimeDocker),
			},
			want: ".environment: ci (docker)",
		},
		{
			name: "default environment",
			env: &ast.Environment{
				Name:    nil,
				Runtime: runtimePtr(ast.RuntimeBare),
			},
			want: ".environment: (default) (bare)",
		},
		{
			name: "no runtime",
			env: &ast.Environment{
				Name:    stringPtr("dev"),
				Runtime: nil,
			},
			want: ".environment: dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.env.String(); got != tt.want {
				t.Errorf("Environment.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVariableString(t *testing.T) {
	tests := []struct {
		name string
		v    *ast.Variable
		want string
	}{
		{
			name: "immediate variable",
			v: &ast.Variable{
				Name:  "cc",
				Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "gcc"}}},
				Lazy:  false,
			},
			want: "cc = gcc",
		},
		{
			name: "lazy variable",
			v: &ast.Variable{
				Name:  "flags",
				Value: &ast.Value{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "-Wall"}}},
				Lazy:  true,
			},
			want: "lazy flags = -Wall",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("Variable.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTargetString(t *testing.T) {
	tests := []struct {
		name   string
		target *ast.Target
		want   string
	}{
		{
			name: "simple target",
			target: &ast.Target{
				Pattern: ast.TargetPattern{
					Segments: []ast.PatternSegment{&ast.LiteralSegment{Text: "build/app"}},
				},
				Dependencies: []ast.Dependency{
					{Segments: []ast.PatternSegment{&ast.LiteralSegment{Text: "main.o"}}},
				},
			},
			want: "build/app: main.o",
		},
		{
			name: "phony target",
			target: &ast.Target{
				Pattern: ast.TargetPattern{
					Segments: []ast.PatternSegment{&ast.LiteralSegment{Text: "all"}},
					IsPhony:  true,
				},
				Dependencies: []ast.Dependency{},
			},
			want: "@all:",
		},
		{
			name: "pattern target",
			target: &ast.Target{
				Pattern: ast.TargetPattern{
					Segments: []ast.PatternSegment{
						&ast.LiteralSegment{Text: "build/"},
						&ast.BraceExpr{Identifier: "name"},
						&ast.LiteralSegment{Text: ".o"},
					},
				},
				Dependencies: []ast.Dependency{
					{Segments: []ast.PatternSegment{
						&ast.LiteralSegment{Text: "src/"},
						&ast.BraceExpr{Identifier: "name"},
						&ast.LiteralSegment{Text: ".c"},
					}},
				},
			},
			want: "build/{name}.o: src/{name}.c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.target.String(); got != tt.want {
				t.Errorf("Target.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRecipeString(t *testing.T) {
	recipe := &ast.Recipe{
		Commands: []ast.Command{
			&ast.LineCommand{Parts: []ast.CommandPart{
				&ast.LiteralCommand{Text: "gcc -o target deps"},
			}},
			&ast.LineCommand{Parts: []ast.CommandPart{
				&ast.LiteralCommand{Text: "echo done"},
			}},
		},
	}

	got := recipe.String()
	want := "Recipe(2 commands)"
	if got != want {
		t.Errorf("Recipe.String() = %q, want %q", got, want)
	}
}

func TestValueString(t *testing.T) {
	tests := []struct {
		name  string
		value *ast.Value
		want  string
	}{
		{
			name: "literal only",
			value: &ast.Value{
				Parts: []ast.ValuePart{&ast.LiteralValue{Text: "hello"}},
			},
			want: "hello",
		},
		{
			name: "with interpolation",
			value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.LiteralValue{Text: "prefix-"},
					&ast.Interpolation{Name: "var"},
					&ast.LiteralValue{Text: "-suffix"},
				},
			},
			want: "prefix-{var}-suffix",
		},
		{
			name: "with raw interpolation",
			value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.Interpolation{Name: "flags", Raw: true},
				},
			},
			want: "{flags:raw}",
		},
		{
			name: "with function call",
			value: &ast.Value{
				Parts: []ast.ValuePart{
					&ast.FunctionCall{
						Name: ast.FuncShell,
						Args: []*ast.Value{
							{Parts: []ast.ValuePart{&ast.LiteralValue{Text: "ls"}}},
						},
					},
				},
			},
			want: "shell(ls)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.String(); got != tt.want {
				t.Errorf("Value.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func runtimePtr(r ast.Runtime) *ast.Runtime {
	return &r
}
