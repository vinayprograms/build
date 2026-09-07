package eval

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

func lit(s string) *ast.LiteralCommand             { return &ast.LiteralCommand{Text: s} }
func interp(name string) *ast.CommandInterpolation { return &ast.CommandInterpolation{Name: name} }

func runSh(t *testing.T, cmd string) string {
	t.Helper()
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		t.Fatalf("sh -c %q failed: %v\n%s", cmd, err, out)
	}
	return strings.TrimRight(string(out), "\n")
}

func TestSpecExamples_InterpolatedText(t *testing.T) {
	ctx := NewContext()
	ctx.Set("dir", "my dir")
	ctx.Set("name", "build")
	cmdCtx := NewCommandContext(ctx, "build/app", nil)

	cases := []struct {
		name   string
		parts  []ast.CommandPart
		expect string
	}{
		{"ls-dir", []ast.CommandPart{lit("ls "), interp("dir")}, "ls 'my dir'"},
		{"cp-name", []ast.CommandPart{lit("cp "), interp("name"), lit(".o out/")}, "cp build.o out/"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmd := &ast.LineCommand{Parts: c.parts}
			out, err := InterpolateCommand(cmd, cmdCtx)
			if err != nil {
				t.Fatalf("interpolate error: %v", err)
			}
			if out != c.expect {
				t.Errorf("interpolated = %q, want %q", out, c.expect)
			}
		})
	}
}

func TestSpecExamples_ShellExecuted(t *testing.T) {
	ctx := NewContext()
	ctx.Set("dir", "my dir")
	ctx.Set("flag", "$HOME")
	ctx.Set("json", `{"key": "value"}`)
	cmdCtx := NewCommandContext(ctx, "build/app", nil)

	cases := []struct {
		name   string
		parts  []ast.CommandPart
		expect string
	}{
		{"echo-dq-dir", []ast.CommandPart{lit(`echo "Dir: `), interp("dir"), lit(`"`)}, "Dir: my dir"},
		{"echo-dq-flag", []ast.CommandPart{lit(`echo "Home: `), interp("flag"), lit(`"`)}, "Home: $HOME"},
		{"echo-sq-dir", []ast.CommandPart{lit(`echo 'Dir: `), interp("dir"), lit(`'`)}, "Dir: my dir"},
		{"echo-dq-json", []ast.CommandPart{lit(`echo "JSON: `), interp("json"), lit(`"`)}, `JSON: {"key": "value"}`},
		{"echo-dq-apostrophe", []ast.CommandPart{lit(`echo "It's `), interp("dir"), lit(`"`)}, "It's my dir"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmd := &ast.LineCommand{Parts: c.parts}
			out, err := InterpolateCommand(cmd, cmdCtx)
			if err != nil {
				t.Fatalf("interpolate error: %v", err)
			}
			got := runSh(t, out)
			if got != c.expect {
				t.Errorf("shell output = %q, want %q (interpolated cmd: %s)", got, c.expect, out)
			}
		})
	}
}

// TestSpecExample_Heredoc covers a block with a heredoc, where interpolated
// values inside the heredoc body are treated as double-quoted context per
// spec: $, ` and \ are neutralized by a backslash so they don't trigger
// expansion/command-substitution, matching real heredoc semantics for those
// three characters.
func TestSpecExample_Heredoc(t *testing.T) {
	ctx := NewContext()
	ctx.Set("msg", "it's a $HOME value with `cmd` inside")
	cmdCtx := NewCommandContext(ctx, "build/app", nil)

	block := &ast.BlockCommand{
		Lines: [][]ast.CommandPart{
			{lit("cat <<EOF")},
			{lit("Value: "), interp("msg")},
			{lit("EOF")},
			{lit("echo done")},
		},
	}
	out, err := InterpolateBlockCommand(block, cmdCtx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("script:\n%s", out)
	got := runSh(t, out)
	want := "Value: it's a $HOME value with `cmd` inside\ndone"
	if got != want {
		t.Errorf("heredoc output = %q, want %q\nscript was:\n%s", got, want, out)
	}
}

// TestSpecExample_DepsBareFormOutsideQuotes verifies {deps} still expands
// to multiple, individually-quoted, space-separated shell words when it
// appears bare (outside quotes) - it is context-independent by design
// since re-quoting a pre-quoted word list makes no sense.
func TestSpecExample_DepsBareFormOutsideQuotes(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "build/app", []string{"src/main.c", "my dir/utils.c"})

	cmd := &ast.LineCommand{Parts: []ast.CommandPart{lit("gcc -o app "), interp("deps")}}
	out, err := InterpolateCommand(cmd, cmdCtx)
	if err != nil {
		t.Fatal(err)
	}
	want := "gcc -o app src/main.c 'my dir/utils.c'"
	if out != want {
		t.Errorf("interpolated = %q, want %q", out, want)
	}

	// Executed, it must expand to two distinct arguments.
	script := "for f in " + strings.TrimPrefix(out, "gcc -o app ") + "; do echo [$f]; done"
	got := runSh(t, script)
	wantExec := "[src/main.c]\n[my dir/utils.c]"
	if got != wantExec {
		t.Errorf("shell expansion = %q, want %q", got, wantExec)
	}
}

// TestSpecExample_RawUntouchedInEveryContext verifies {var:raw} is emitted
// completely untouched regardless of the surrounding quote context.
func TestSpecExample_RawUntouchedInEveryContext(t *testing.T) {
	ctx := NewContext()
	ctx.Set("v", `it's "raw" $HOME`)
	cmdCtx := NewCommandContext(ctx, "build/app", nil)

	raw := &ast.CommandInterpolation{Name: "v", Raw: true}

	cases := []struct {
		name  string
		parts []ast.CommandPart
	}{
		{"unquoted", []ast.CommandPart{lit("echo "), raw}},
		{"double-quoted", []ast.CommandPart{lit(`echo "`), raw, lit(`"`)}},
		{"single-quoted", []ast.CommandPart{lit(`echo '`), raw, lit(`'`)}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmd := &ast.LineCommand{Parts: c.parts}
			out, err := InterpolateCommand(cmd, cmdCtx)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out, `it's "raw" $HOME`) {
				t.Errorf("raw value was altered: %q", out)
			}
		})
	}
}

// TestSpecExample_QuoteStateCarriesAcrossBlockLines verifies an open
// double-quoted string spanning multiple lines of a block keeps
// interpolations inside it double-quoted-context formatted, since a block
// is executed as one continuous shell script.
func TestSpecExample_QuoteStateCarriesAcrossBlockLines(t *testing.T) {
	ctx := NewContext()
	ctx.Set("dir", "my dir")
	cmdCtx := NewCommandContext(ctx, "build/app", nil)

	block := &ast.BlockCommand{
		Lines: [][]ast.CommandPart{
			{lit(`echo "line one`)},
			{lit("line two: "), interp("dir")},
			{lit(`"`)},
		},
	}
	out, err := InterpolateBlockCommand(block, cmdCtx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("script:\n%s", out)
	got := runSh(t, out)
	want := "line one\nline two: my dir"
	if got != want {
		t.Errorf("output = %q, want %q\nscript was:\n%s", got, want, out)
	}
}

// TestSpecExample_DepsInsideQuotes verifies {deps} follows the quoting
// context like any other value when it appears inside quotes: no per-item
// quoting, just the joined list escaped for the surrounding quotes.
func TestSpecExample_DepsInsideQuotes(t *testing.T) {
	ctx := NewContext()
	cmdCtx := NewCommandContext(ctx, "out/app.txt", []string{"src/imp!.txt", "src/c.txt"})

	cases := []struct {
		parts []ast.CommandPart
		want  string
	}{
		{[]ast.CommandPart{lit(`echo "deps: `), interp("deps"), lit(`"`)}, `echo "deps: src/imp!.txt src/c.txt"`},
		{[]ast.CommandPart{lit(`echo 'deps: `), interp("deps"), lit(`'`)}, `echo 'deps: src/imp!.txt src/c.txt'`},
		{[]ast.CommandPart{lit(`echo `), interp("deps")}, `echo 'src/imp!.txt' src/c.txt`},
	}
	for _, c := range cases {
		out, err := InterpolateCommand(&ast.LineCommand{Parts: c.parts}, cmdCtx)
		if err != nil {
			t.Fatal(err)
		}
		if out != c.want {
			t.Errorf("interpolated = %q, want %q", out, c.want)
		}
		if got := runSh(t, out); got != "deps: src/imp!.txt src/c.txt" && got != "src/imp!.txt src/c.txt" {
			t.Errorf("executed = %q", got)
		}
	}
}
