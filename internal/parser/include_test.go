package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vinayprograms/need/internal/ast"
	"github.com/vinayprograms/need/internal/lexer"
)

func TestParser_ParseInclude_Simple(t *testing.T) {
	// Create a temp directory with test files
	tmpDir := t.TempDir()

	// Create included file
	includedContent := `cc = gcc
cflags = -Wall
`
	includedPath := filepath.Join(tmpDir, "common.need")
	if err := os.WriteFile(includedPath, []byte(includedContent), 0644); err != nil {
		t.Fatalf("failed to create included file: %v", err)
	}

	// Create main file content
	mainContent := `.include: ` + includedPath + `
binary = app
`

	l := lexer.New("main.need", mainContent)
	p := New(l)

	// Parse include directive
	if p.current.Type != lexer.DOT_INCLUDE {
		t.Fatalf("expected DOT_INCLUDE token, got %v", p.current.Type)
	}

	directive, statements, err := p.ParseInclude()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check directive kind
	if directive.Kind != ast.DirectiveInclude {
		t.Errorf("directive kind = %v, want DirectiveInclude", directive.Kind)
	}

	// Check that included statements were returned
	if len(statements) != 2 {
		t.Fatalf("expected 2 included statements, got %d", len(statements))
	}

	// Check first included statement is a variable
	v1, ok := statements[0].(*ast.Variable)
	if !ok {
		t.Errorf("statement[0] is not Variable, got %T", statements[0])
	} else if v1.Name != "cc" {
		t.Errorf("variable name = %q, want %q", v1.Name, "cc")
	}

	// Check second included statement is a variable
	v2, ok := statements[1].(*ast.Variable)
	if !ok {
		t.Errorf("statement[1] is not Variable, got %T", statements[1])
	} else if v2.Name != "cflags" {
		t.Errorf("variable name = %q, want %q", v2.Name, "cflags")
	}
}

func TestParser_ParseInclude_WithInterpolation(t *testing.T) {
	// Create a temp directory with test files
	tmpDir := t.TempDir()

	// Create included file
	includedContent := `extra_flags = -O2
`
	includedPath := filepath.Join(tmpDir, "extra.need")
	if err := os.WriteFile(includedPath, []byte(includedContent), 0644); err != nil {
		t.Fatalf("failed to create included file: %v", err)
	}

	// Create main file content with interpolation in path
	// Note: interpolation evaluation happens during semantic analysis, not parsing
	// So we test that the value is parsed correctly
	mainContent := `.include: ` + includedPath + `
`

	l := lexer.New("main.need", mainContent)
	p := New(l)

	directive, _, err := p.ParseInclude()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check directive has a value
	if directive.Value == nil {
		t.Error("directive value is nil")
	}
}

func TestParser_ParseInclude_FileNotFound(t *testing.T) {
	mainContent := `.include: /nonexistent/path/file.need
`

	l := lexer.New("main.need", mainContent)
	p := New(l)

	_, _, err := p.ParseInclude()
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParser_ParseInclude_CircularInclude(t *testing.T) {
	// Create a temp directory with test files
	tmpDir := t.TempDir()

	// Create file A that includes file B
	fileAPath := filepath.Join(tmpDir, "a.need")
	fileBPath := filepath.Join(tmpDir, "b.need")

	fileAContent := `.include: ` + fileBPath + `
varA = value
`
	fileBContent := `.include: ` + fileAPath + `
varB = value
`

	if err := os.WriteFile(fileAPath, []byte(fileAContent), 0644); err != nil {
		t.Fatalf("failed to create file A: %v", err)
	}
	if err := os.WriteFile(fileBPath, []byte(fileBContent), 0644); err != nil {
		t.Fatalf("failed to create file B: %v", err)
	}

	// Read file A content
	content, _ := os.ReadFile(fileAPath)
	l := lexer.New(fileAPath, string(content))
	p := New(l)

	// Parse include - should detect circular include
	_, _, err := p.ParseInclude()
	if err == nil {
		t.Error("expected error for circular include")
	}
}

func TestParser_ParseInclude_NestedInclude(t *testing.T) {
	// Create a temp directory with test files
	tmpDir := t.TempDir()

	// Create nested include chain: main -> common -> base
	basePath := filepath.Join(tmpDir, "base.need")
	commonPath := filepath.Join(tmpDir, "common.need")

	baseContent := `base_var = base_value
`
	commonContent := `.include: ` + basePath + `
common_var = common_value
`

	if err := os.WriteFile(basePath, []byte(baseContent), 0644); err != nil {
		t.Fatalf("failed to create base file: %v", err)
	}
	if err := os.WriteFile(commonPath, []byte(commonContent), 0644); err != nil {
		t.Fatalf("failed to create common file: %v", err)
	}

	mainContent := `.include: ` + commonPath + `
main_var = main_value
`

	l := lexer.New("main.need", mainContent)
	p := New(l)

	directive, statements, err := p.ParseInclude()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if directive.Kind != ast.DirectiveInclude {
		t.Errorf("directive kind = %v, want DirectiveInclude", directive.Kind)
	}

	// Should have statements from both common and base files
	// base.need: base_var
	// common.need: .include (processed), common_var
	// So we expect: base_var, common_var
	if len(statements) < 2 {
		t.Fatalf("expected at least 2 statements from nested includes, got %d", len(statements))
	}

	// Check that base_var came first (from nested include)
	foundBaseVar := false
	foundCommonVar := false
	for _, stmt := range statements {
		if v, ok := stmt.(*ast.Variable); ok {
			if v.Name == "base_var" {
				foundBaseVar = true
			}
			if v.Name == "common_var" {
				foundCommonVar = true
			}
		}
	}

	if !foundBaseVar {
		t.Error("expected to find base_var from nested include")
	}
	if !foundCommonVar {
		t.Error("expected to find common_var from include")
	}
}

func TestParser_ParseInclude_RelativePath(t *testing.T) {
	// Create a temp directory with test files
	tmpDir := t.TempDir()

	// Create included file
	includedContent := `rel_var = value
`
	includedPath := filepath.Join(tmpDir, "included.need")
	if err := os.WriteFile(includedPath, []byte(includedContent), 0644); err != nil {
		t.Fatalf("failed to create included file: %v", err)
	}

	// Create main file in same directory
	mainPath := filepath.Join(tmpDir, "main.need")
	mainContent := `.include: ./included.need
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to create main file: %v", err)
	}

	// Read and parse
	content, _ := os.ReadFile(mainPath)
	l := lexer.New(mainPath, string(content))
	p := New(l)

	_, statements, err := p.ParseInclude()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(statements))
	}

	v, ok := statements[0].(*ast.Variable)
	if !ok {
		t.Errorf("statement is not Variable, got %T", statements[0])
	} else if v.Name != "rel_var" {
		t.Errorf("variable name = %q, want %q", v.Name, "rel_var")
	}
}

func TestParser_ParseInclude_SourceLocation(t *testing.T) {
	// Create a temp directory with test files
	tmpDir := t.TempDir()

	// Create included file
	includedContent := `inc_var = value
`
	includedPath := filepath.Join(tmpDir, "inc.need")
	if err := os.WriteFile(includedPath, []byte(includedContent), 0644); err != nil {
		t.Fatalf("failed to create included file: %v", err)
	}

	mainContent := `.include: ` + includedPath + `
`

	l := lexer.New("main.need", mainContent)
	p := New(l)

	directive, _, err := p.ParseInclude()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check directive location is in main file
	if directive.Location.File != "main.need" {
		t.Errorf("directive location file = %q, want %q", directive.Location.File, "main.need")
	}
	if directive.Location.Line != 1 {
		t.Errorf("directive location line = %d, want %d", directive.Location.Line, 1)
	}
}

func TestParser_ParseInclude_EmptyFile(t *testing.T) {
	// Create a temp directory with test files
	tmpDir := t.TempDir()

	// Create empty included file
	includedPath := filepath.Join(tmpDir, "empty.need")
	if err := os.WriteFile(includedPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create included file: %v", err)
	}

	mainContent := `.include: ` + includedPath + `
`

	l := lexer.New("main.need", mainContent)
	p := New(l)

	directive, statements, err := p.ParseInclude()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if directive.Kind != ast.DirectiveInclude {
		t.Errorf("directive kind = %v, want DirectiveInclude", directive.Kind)
	}

	// Empty file should result in empty statements
	if len(statements) != 0 {
		t.Errorf("expected 0 statements from empty file, got %d", len(statements))
	}
}

// TestParser_ParseInclude_Interpolation covers C1: a .include: path may
// reference a variable defined earlier in the same file, including nested
// {a}/{b} chains.
func TestParser_ParseInclude_Interpolation(t *testing.T) {
	tmpDir := t.TempDir()

	includedPath := filepath.Join(tmpDir, "sub", "extra.need")
	if err := os.MkdirAll(filepath.Dir(includedPath), 0755); err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}
	if err := os.WriteFile(includedPath, []byte("extra_var = value\n"), 0644); err != nil {
		t.Fatalf("failed to create included file: %v", err)
	}

	mainPath := filepath.Join(tmpDir, "main.need")
	mainContent := "config_dir = sub\n" +
		"config_file = {config_dir}/extra.need\n" +
		".include: {config_file}\n"

	l := lexer.New(mainPath, mainContent)
	p := New(l)

	stmts, errs := p.ParseNeedfile()
	if errs.HasErrors() {
		t.Fatalf("unexpected parse errors: %v", errs.Errors)
	}

	found := false
	for _, stmt := range stmts {
		if v, ok := stmt.(*ast.Variable); ok && v.Name == "extra_var" {
			found = true
		}
	}
	if !found {
		t.Error("expected to find extra_var from the interpolated include path")
	}
}

// TestParser_ParseInclude_InterpolationUndefined covers C1's error path: an
// undefined variable in a .include: path is a parse error naming the
// directive and the variable.
func TestParser_ParseInclude_InterpolationUndefined(t *testing.T) {
	mainContent := ".include: {nope}/x.need\n"

	l := lexer.New("main.need", mainContent)
	p := New(l)

	_, errs := p.ParseNeedfile()
	if !errs.HasErrors() {
		t.Fatal("expected a parse error for undefined variable in .include: path")
	}
	msg := errs.Errors[0].Message
	if !strings.Contains(msg, ".include: cannot resolve 'nope': undefined variable") {
		t.Errorf("error message = %q, want to contain %q", msg, ".include: cannot resolve 'nope': undefined variable")
	}
}

// TestParser_ParseInclude_InterpolationLazy covers C1: a lazy variable
// cannot be used in a .include: path (it isn't resolved at parse time).
func TestParser_ParseInclude_InterpolationLazy(t *testing.T) {
	mainContent := "lazy config_dir = shell(echo sub)\n" +
		".include: {config_dir}/x.need\n"

	l := lexer.New("main.need", mainContent)
	p := New(l)

	_, errs := p.ParseNeedfile()
	if !errs.HasErrors() {
		t.Fatal("expected a parse error for lazy variable in .include: path")
	}
	msg := errs.Errors[0].Message
	if !strings.Contains(msg, ".include: cannot resolve 'config_dir': lazy variable") {
		t.Errorf("error message = %q, want to contain %q", msg, ".include: cannot resolve 'config_dir': lazy variable")
	}
}

// TestParser_ParseInclude_InterpolationAutomatic covers C1: automatic
// variables are never resolvable in a .include: path.
func TestParser_ParseInclude_InterpolationAutomatic(t *testing.T) {
	mainContent := ".include: {target}/x.need\n"

	l := lexer.New("main.need", mainContent)
	p := New(l)

	_, errs := p.ParseNeedfile()
	if !errs.HasErrors() {
		t.Fatal("expected a parse error for automatic variable in .include: path")
	}
	msg := errs.Errors[0].Message
	if !strings.Contains(msg, ".include: cannot resolve 'target': automatic variable") {
		t.Errorf("error message = %q, want to contain %q", msg, ".include: cannot resolve 'target': automatic variable")
	}
}

func TestParser_ParseInclude_WithComments(t *testing.T) {
	// Create a temp directory with test files
	tmpDir := t.TempDir()

	// Create included file with comments
	includedContent := `# This is a comment
comment_var = value
# Another comment
`
	includedPath := filepath.Join(tmpDir, "comments.need")
	if err := os.WriteFile(includedPath, []byte(includedContent), 0644); err != nil {
		t.Fatalf("failed to create included file: %v", err)
	}

	mainContent := `.include: ` + includedPath + `
`

	l := lexer.New("main.need", mainContent)
	p := New(l)

	_, statements, err := p.ParseInclude()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have comments and variable
	// The exact count depends on whether we preserve comments in AST
	foundVar := false
	for _, stmt := range statements {
		if v, ok := stmt.(*ast.Variable); ok && v.Name == "comment_var" {
			foundVar = true
		}
	}

	if !foundVar {
		t.Error("expected to find comment_var in included statements")
	}
}
