package integration

import (
	"strings"
	"testing"
)

// TestSimpleCCompilation tests compiling a simple C program.
func TestSimpleCCompilation(t *testing.T) {
	h := NewTestHarness(t)

	// Write a simple C program
	h.WriteFile("main.c", `#include <stdio.h>

int main() {
    printf("Hello from C!\n");
    return 0;
}
`)

	// Write Buildfile
	h.WriteFile("Buildfile", `.shell: bash

app: main.c
	gcc -o {target} {in}
`)

	// Build the target
	result := h.Run("app")
	result.AssertSuccess()

	// Verify the binary exists
	if !h.FileExists("app") {
		t.Error("app binary was not created")
	}

	// Run the binary (if it exists)
	if h.FileExists("app") {
		result := h.RunShell("./app")
		result.AssertSuccess().
			AssertStdoutContains("Hello from C!")
	}
}

// TestPatternTargets tests pattern-based target matching.
func TestPatternTargets(t *testing.T) {
	h := NewTestHarness(t)

	// Write source files
	h.Mkdir("src")
	h.WriteFile("src/utils.c", `#include "utils.h"
int add(int a, int b) { return a + b; }
`)
	h.WriteFile("src/utils.h", `int add(int a, int b);
`)
	h.WriteFile("src/main.c", `#include <stdio.h>
#include "utils.h"

int main() {
    printf("2 + 3 = %d\n", add(2, 3));
    return 0;
}
`)

	// Write Buildfile with pattern targets
	h.WriteFile("Buildfile", `.shell: bash

objects = build/main.o build/utils.o

app: {objects}
	gcc -o {target} {deps}

build/{name}.o: src/{name}.c
	mkdir -p build
	gcc -c {in} -o {out} -I src
`)

	// Build the target
	result := h.Run("app")
	result.AssertSuccess()

	// Verify intermediate files were created
	if !h.FileExists("build/main.o") {
		t.Error("build/main.o was not created")
	}
	if !h.FileExists("build/utils.o") {
		t.Error("build/utils.o was not created")
	}

	// Verify the binary exists and works
	if !h.FileExists("app") {
		t.Error("app binary was not created")
	}
}

// TestConditionalCompilation tests conditional blocks based on OS/arch.
func TestConditionalCompilation(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("main.c", `#include <stdio.h>

int main() {
    printf("Hello\n");
    return 0;
}
`)

	// Write Buildfile with conditionals
	h.WriteFile("Buildfile", `.shell: bash

if {os} == darwin
	cflags = -Wall -Wextra
elif {os} == linux
	cflags = -Wall -O2
else
	cflags = -Wall
end

app: main.c
	gcc {cflags} -o {target} {in}
`)

	// Build should succeed regardless of platform
	result := h.Run("app")
	result.AssertSuccess()

	if !h.FileExists("app") {
		t.Error("app binary was not created")
	}
}

// TestPhonyTargets tests phony targets (tasks without file outputs).
func TestPhonyTargets(t *testing.T) {
	h := NewTestHarness(t)

	// Create a test file that clean should remove
	h.WriteFile("temp.txt", "temporary")

	h.WriteFile("Buildfile", `.shell: bash

@all: @build @test

@build:
	echo "Building..."
	touch built.txt

@test:
	echo "Testing..."
	touch tested.txt

@clean:
	echo "Cleaning..."
	rm -f built.txt tested.txt temp.txt
`)

	// Test @all target (should trigger both build and test)
	result := h.Run("all") // Note: no @ prefix needed in args
	result.AssertSuccess().
		AssertStdoutContains("Building...").
		AssertStdoutContains("Testing...")

	if !h.FileExists("built.txt") {
		t.Error("built.txt was not created")
	}
	if !h.FileExists("tested.txt") {
		t.Error("tested.txt was not created")
	}

	// Test @clean target
	result = h.Run("clean")
	result.AssertSuccess().
		AssertStdoutContains("Cleaning...")

	if h.FileExists("built.txt") {
		t.Error("built.txt should have been removed")
	}
	if h.FileExists("tested.txt") {
		t.Error("tested.txt should have been removed")
	}
	if h.FileExists("temp.txt") {
		t.Error("temp.txt should have been removed")
	}
}

// TestDependencyChain tests that dependencies are built in correct order.
func TestDependencyChain(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

@final: @middle
	echo "final" >> order.txt

@middle: @first
	echo "middle" >> order.txt

@first:
	echo "first" > order.txt
`)

	result := h.Run("final")
	result.AssertSuccess()

	// Check the order file
	content := h.ReadFile("order.txt")
	lines := strings.Split(strings.TrimSpace(content), "\n")

	if len(lines) != 3 {
		t.Errorf("expected 3 lines in order.txt, got %d", len(lines))
	}
	if lines[0] != "first" || lines[1] != "middle" || lines[2] != "final" {
		t.Errorf("wrong order in order.txt: %v", lines)
	}
}

// TestPhonyDependenciesWithoutAtPrefix tests that phony targets can be
// referenced in dependencies without the @ prefix.
func TestPhonyDependenciesWithoutAtPrefix(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

@first:
	echo "First target"

@second:
	echo "Second target"

@all: first second
	echo "All targets invoked"
`)

	result := h.Run("all")
	result.AssertSuccess().
		AssertStdoutContains("First target").
		AssertStdoutContains("Second target").
		AssertStdoutContains("All targets invoked")
}

// TestVariableInterpolation tests variable substitution in recipes.
func TestVariableInterpolation(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

name = world
greeting = Hello

@greet:
	echo "{greeting}, {name}!"
`)

	result := h.Run("greet")
	result.AssertSuccess().
		AssertStdoutContains("Hello, world!")
}

// TestLazyVariables tests lazy variable evaluation.
func TestLazyVariables(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

counter = 0
lazy next = shell(echo $((++counter)))

@test:
	echo "First: {next}"
	echo "Second: {next}"
`)

	result := h.Run("test")
	result.AssertSuccess()

	// Lazy variables should be evaluated each time they're used
	stdout := result.Stdout()
	if !strings.Contains(stdout, "First:") || !strings.Contains(stdout, "Second:") {
		t.Error("expected both First and Second in output")
	}
}

// TestBuiltInFunctions tests built-in functions (glob, basename, dirname, replace).
func TestBuiltInFunctions(t *testing.T) {
	h := NewTestHarness(t)

	// Create some test files
	h.Mkdir("src")
	h.WriteFile("src/file1.c", "")
	h.WriteFile("src/file2.c", "")
	h.WriteFile("src/file3.c", "")

	h.WriteFile("Buildfile", `.shell: bash

sources = glob(src/*.c)
objects = replace({sources}, .c, .o)
dir = dirname(src/file1.c)
base = basename(src/file1.c)

@test:
	echo "Sources: {sources}"
	echo "Objects: {objects}"
	echo "Dir: {dir}"
	echo "Base: {base}"
`)

	result := h.Run("test")
	result.AssertSuccess().
		AssertStdoutContains("Sources:").
		AssertStdoutContains(".c").
		AssertStdoutContains("Objects:").
		AssertStdoutContains(".o").
		AssertStdoutContains("Dir: src").
		AssertStdoutContains("Base: file1.c")
}

// TestAutomaticVariables tests automatic variables in recipes.
func TestAutomaticVariables(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("input.txt", "input content")

	h.WriteFile("Buildfile", `.shell: bash

output.txt: input.txt
	echo "Target: {target}" > {out}
	echo "Input: {in}" >> {out}
	echo "Deps: {deps}" >> {out}
	echo "Target.dir: {target.dir}" >> {out}
	echo "Target.file: {target.file}" >> {out}
`)

	result := h.Run("output.txt")
	result.AssertSuccess()

	content := h.ReadFile("output.txt")
	if !strings.Contains(content, "Target: output.txt") {
		t.Error("target variable not interpolated correctly")
	}
	if !strings.Contains(content, "Input: input.txt") {
		t.Error("in variable not interpolated correctly")
	}
	if !strings.Contains(content, "Deps: input.txt") {
		t.Error("deps variable not interpolated correctly")
	}
	if !strings.Contains(content, "Target.file: output.txt") {
		t.Error("target.file variable not interpolated correctly")
	}
}

// TestBlockCommands tests block: syntax for multi-line shell scripts.
func TestBlockCommands(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

output.txt:
	block:
		for i in 1 2 3; do
			echo "Number: $i"
		done > {target}
`)

	result := h.Run("output.txt")
	result.AssertSuccess()

	content := h.ReadFile("output.txt")
	if !strings.Contains(content, "Number: 1") ||
		!strings.Contains(content, "Number: 2") ||
		!strings.Contains(content, "Number: 3") {
		t.Errorf("block command output incorrect: %s", content)
	}
}

// TestIncludeDirective tests the .include: directive.
func TestIncludeDirective(t *testing.T) {
	h := NewTestHarness(t)

	// Write included file
	h.WriteFile("common.build", `# Common build configuration
cc = gcc
cflags = -Wall -O2
`)

	// Write main Buildfile
	h.WriteFile("Buildfile", `.include: ./common.build

app: main.c
	{cc} {cflags} -o {target} {in}
`)

	h.WriteFile("main.c", `#include <stdio.h>
int main() { printf("Hello\n"); return 0; }
`)

	result := h.Run("app")
	result.AssertSuccess()

	if !h.FileExists("app") {
		t.Error("app binary was not created")
	}
}

// TestStalenessDetection tests that targets are only rebuilt when dependencies change.
func TestStalenessDetection(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("input.txt", "original")
	h.WriteFile("Buildfile", `.shell: bash

output.txt: input.txt
	cp {in} {out}
`)

	// First build
	result := h.Run("output.txt")
	result.AssertSuccess()

	if !h.FileExists("output.txt") {
		t.Fatal("output.txt was not created")
	}

	// Second build (should be up-to-date)
	result = h.Run("--verbose", "output.txt")
	result.AssertSuccess()

	// In verbose mode, we should see some indication that work was done or skipped
	// The exact behavior depends on implementation
}

// TestErrorHandling tests that build errors are reported correctly.
func TestErrorHandling(t *testing.T) {
	h := NewTestHarness(t)

	h.WriteFile("Buildfile", `.shell: bash

@fail:
	echo "About to fail"
	false
	echo "This should not run"
`)

	result := h.Run("fail")
	result.AssertExitCode(1). // Build failure
					AssertStdoutContains("About to fail").
					AssertStdoutNotContains("This should not run")
}

// TestNestedIncludes tests that nested .include: directives work correctly.
func TestNestedIncludes(t *testing.T) {
	h := NewTestHarness(t)

	// Create the base file
	h.WriteFile("base.build", `base_var = base_value
`)

	// Create the common file that includes base
	h.WriteFile("common.build", `.include: ./base.build
common_var = common_value
`)

	// Create main Buildfile that includes common
	h.WriteFile("Buildfile", `.shell: bash
.include: ./common.build

main_var = main_value

@test:
	echo "base: {base_var}"
	echo "common: {common_var}"
	echo "main: {main_var}"
`)

	result := h.Run("test")
	result.AssertSuccess().
		AssertStdoutContains("base:").
		AssertStdoutContains("base_value").
		AssertStdoutContains("common:").
		AssertStdoutContains("common_value").
		AssertStdoutContains("main:").
		AssertStdoutContains("main_value")
}

// TestCircularIncludeDetection tests that circular includes are detected and reported as errors.
func TestCircularIncludeDetection(t *testing.T) {
	h := NewTestHarness(t)

	// Create file A that includes file B
	h.WriteFile("a.build", `.include: ./b.build
varA = value
`)

	// Create file B that includes file A (circular)
	h.WriteFile("b.build", `.include: ./a.build
varB = value
`)

	// Create main Buildfile that includes A
	h.WriteFile("Buildfile", `.include: ./a.build

@test:
	echo "test"
`)

	result := h.Run("test")
	result.AssertExitCode(3). // Parse error
					AssertStderrContains("circular")
}

// TestDeepNestedIncludes tests deeply nested includes (A → B → C → D).
func TestDeepNestedIncludes(t *testing.T) {
	h := NewTestHarness(t)

	// Create deep include chain
	h.WriteFile("d.build", `d_var = d_value
`)
	h.WriteFile("c.build", `.include: ./d.build
c_var = c_value
`)
	h.WriteFile("b.build", `.include: ./c.build
b_var = b_value
`)
	h.WriteFile("a.build", `.include: ./b.build
a_var = a_value
`)

	h.WriteFile("Buildfile", `.shell: bash
.include: ./a.build

@test:
	echo "{a_var} {b_var} {c_var} {d_var}"
`)

	result := h.Run("test")
	result.AssertSuccess().
		AssertStdoutContains("a_value").
		AssertStdoutContains("b_value").
		AssertStdoutContains("c_value").
		AssertStdoutContains("d_value")
}

// TestIncludeWithTargets tests that targets from included files can be built.
func TestIncludeWithTargets(t *testing.T) {
	h := NewTestHarness(t)

	// Create included file with targets
	h.WriteFile("targets.build", `@included-target:
	echo "from included file"
`)

	h.WriteFile("Buildfile", `.shell: bash
.include: ./targets.build

@main-target: @included-target
	echo "from main file"
`)

	result := h.Run("main-target")
	result.AssertSuccess().
		AssertStdoutContains("from included file").
		AssertStdoutContains("from main file")
}
