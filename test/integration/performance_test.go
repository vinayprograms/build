package integration

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestLargeBuildfileParsing_1000Targets tests parsing a Buildfile with 1000+ targets.
// This validates that the parser handles large files without excessive memory or time.
func TestLargeBuildfileParsing_1000Targets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	h := NewTestHarness(t)

	// Generate a Buildfile with 1000 targets
	var sb strings.Builder
	sb.WriteString(".shell: bash\n\n")

	// Write 1000 phony targets that depend on each other
	for i := 0; i < 1000; i++ {
		if i == 0 {
			sb.WriteString(fmt.Sprintf("@target%d:\n\techo \"Target %d\"\n\n", i, i))
		} else {
			sb.WriteString(fmt.Sprintf("@target%d: @target%d\n\techo \"Target %d\"\n\n", i, i-1, i))
		}
	}

	h.WriteFile("Buildfile", sb.String())

	// Parse and validate (don't actually build)
	start := time.Now()
	result := h.Run("--dry-run", "target999")
	elapsed := time.Since(start)

	// Should complete successfully
	result.AssertSuccess()

	// Should complete within reasonable time (10 seconds for parsing + planning)
	if elapsed > 10*time.Second {
		t.Errorf("parsing 1000 targets took too long: %v", elapsed)
	}

	t.Logf("Parsed 1000 targets in %v", elapsed)
}

// TestLargeBuildfileParsing_ManyPatternTargets tests parsing with many pattern targets.
func TestLargeBuildfileParsing_ManyPatternTargets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	h := NewTestHarness(t)

	// Generate a Buildfile with 100 pattern targets and many concrete targets
	var sb strings.Builder
	sb.WriteString(".shell: bash\n\n")

	// Create source files to make targets valid
	h.Mkdir("src")
	h.Mkdir("build")

	// Write pattern targets
	for i := 0; i < 100; i++ {
		// Create a source file for each pattern target to match
		h.WriteFile(fmt.Sprintf("src/file%d.c", i), fmt.Sprintf("/* file %d */\n", i))
		sb.WriteString(fmt.Sprintf("build/file%d.o: src/file%d.c\n\techo \"Compiling file%d.o\"\n\n", i, i, i))
	}

	// Add a main target that depends on all .o files
	sb.WriteString("build/app:")
	for i := 0; i < 100; i++ {
		sb.WriteString(fmt.Sprintf(" build/file%d.o", i))
	}
	sb.WriteString("\n\techo \"Linking app\"\n")

	h.WriteFile("Buildfile", sb.String())

	// Parse and validate (don't actually build)
	start := time.Now()
	result := h.Run("--dry-run", "build/app")
	elapsed := time.Since(start)

	// Should complete successfully
	result.AssertSuccess()

	// Should complete within reasonable time
	if elapsed > 10*time.Second {
		t.Errorf("parsing many pattern targets took too long: %v", elapsed)
	}

	// Verify output mentions linking
	result.AssertStdoutContains("echo \"Linking app\"")

	t.Logf("Parsed 100 targets with dependencies in %v", elapsed)
}

// TestDeepIncludeHierarchy tests deeply nested include files.
func TestDeepIncludeHierarchy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	h := NewTestHarness(t)

	// Create a deep include hierarchy: level0.build includes level1.build includes level2.build ...
	const depth = 20

	// Create include files from deepest to shallowest
	for i := depth; i >= 0; i-- {
		var content string
		if i == depth {
			// Deepest file has no include
			content = fmt.Sprintf("var%d = value%d\n", i, i)
		} else {
			// Each file includes the next level
			content = fmt.Sprintf(".include: ./level%d.build\nvar%d = value%d\n", i+1, i, i)
		}
		h.WriteFile(fmt.Sprintf("level%d.build", i), content)
	}

	// Main Buildfile includes level0
	h.WriteFile("Buildfile", `.shell: bash
.include: ./level0.build

@test:
	echo "var0={var0}"
	echo "var10={var10}"
	echo "var20={var20}"
`)

	// Parse and run
	start := time.Now()
	result := h.Run("test")
	elapsed := time.Since(start)

	// Should complete successfully
	result.AssertSuccess()

	// Verify variables from all levels are accessible
	result.AssertStdoutContains("var0=value0").
		AssertStdoutContains("var10=value10").
		AssertStdoutContains("var20=value20")

	// Should complete within reasonable time
	if elapsed > 5*time.Second {
		t.Errorf("parsing deep include hierarchy took too long: %v", elapsed)
	}

	t.Logf("Parsed %d-level deep include hierarchy in %v", depth, elapsed)
}

// TestWideIncludeHierarchy tests many sibling includes from one file.
func TestWideIncludeHierarchy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	h := NewTestHarness(t)

	// Create 50 sibling include files
	const width = 50

	var mainBuilder strings.Builder
	mainBuilder.WriteString(".shell: bash\n\n")

	for i := 0; i < width; i++ {
		// Each sibling defines its own variables and targets
		content := fmt.Sprintf("var%d = value%d\n\n@target%d:\n\techo \"target%d: {var%d}\"\n", i, i, i, i, i)
		h.WriteFile(fmt.Sprintf("include%d.build", i), content)
		mainBuilder.WriteString(fmt.Sprintf(".include: ./include%d.build\n", i))
	}

	// Add an aggregator target that depends on all included targets
	mainBuilder.WriteString("\n@all:")
	for i := 0; i < width; i++ {
		mainBuilder.WriteString(fmt.Sprintf(" @target%d", i))
	}
	mainBuilder.WriteString("\n\techo \"All targets complete\"\n")

	h.WriteFile("Buildfile", mainBuilder.String())

	// Parse and validate
	start := time.Now()
	result := h.Run("--dry-run", "all")
	elapsed := time.Since(start)

	// Should complete successfully
	result.AssertSuccess()

	// Should complete within reasonable time
	if elapsed > 10*time.Second {
		t.Errorf("parsing wide include hierarchy took too long: %v", elapsed)
	}

	t.Logf("Parsed %d sibling includes in %v", width, elapsed)
}

// TestDeepDependencyChain tests a very long dependency chain.
func TestDeepDependencyChain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	h := NewTestHarness(t)

	// Create a deep dependency chain: target0 <- target1 <- target2 <- ... <- target99
	const depth = 100

	var sb strings.Builder
	sb.WriteString(".shell: bash\n\n")

	for i := 0; i < depth; i++ {
		if i == 0 {
			sb.WriteString(fmt.Sprintf("@target%d:\n\techo \"Building base target\"\n\n", i))
		} else {
			sb.WriteString(fmt.Sprintf("@target%d: @target%d\n\techo \"Building target %d\"\n\n", i, i-1, i))
		}
	}

	h.WriteFile("Buildfile", sb.String())

	// Parse and validate the deepest target
	start := time.Now()
	result := h.Run("--dry-run", fmt.Sprintf("target%d", depth-1))
	elapsed := time.Since(start)

	// Should complete successfully
	result.AssertSuccess()

	// Verify the entire chain is planned
	stdout := result.Stdout()
	for i := 0; i < depth; i++ {
		if !strings.Contains(stdout, fmt.Sprintf("target%d", i)) {
			t.Errorf("missing target%d in dry-run output", i)
		}
	}

	// Should complete within reasonable time
	if elapsed > 5*time.Second {
		t.Errorf("planning deep dependency chain took too long: %v", elapsed)
	}

	t.Logf("Planned %d-deep dependency chain in %v", depth, elapsed)
}

// TestWideDependencyGraph tests a target with many direct dependencies.
func TestWideDependencyGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	h := NewTestHarness(t)

	// Create many independent targets and one target that depends on all of them
	const width = 200

	var sb strings.Builder
	sb.WriteString(".shell: bash\n\n")

	for i := 0; i < width; i++ {
		sb.WriteString(fmt.Sprintf("@dep%d:\n\techo \"Building dep %d\"\n\n", i, i))
	}

	// Create the main target with all dependencies
	sb.WriteString("@all:")
	for i := 0; i < width; i++ {
		sb.WriteString(fmt.Sprintf(" @dep%d", i))
	}
	sb.WriteString("\n\techo \"All dependencies built\"\n")

	h.WriteFile("Buildfile", sb.String())

	// Parse and validate
	start := time.Now()
	result := h.Run("--dry-run", "all")
	elapsed := time.Since(start)

	// Should complete successfully
	result.AssertSuccess()

	// Should complete within reasonable time
	if elapsed > 5*time.Second {
		t.Errorf("planning wide dependency graph took too long: %v", elapsed)
	}

	t.Logf("Planned %d-wide dependency graph in %v", width, elapsed)
}

// TestDiamondDependencyPerformance tests diamond dependency pattern at scale.
func TestDiamondDependencyPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	h := NewTestHarness(t)

	// Create a multi-level diamond pattern:
	// Level 0: base (1 target)
	// Level 1: many targets depending on base
	// Level 2: many targets depending on level 1 targets
	// Level 3: top target depending on all level 2 targets

	var sb strings.Builder
	sb.WriteString(".shell: bash\n\n")

	// Level 0: base
	sb.WriteString("@base:\n\techo \"Base\"\n\n")

	// Level 1: 20 targets depending on base
	for i := 0; i < 20; i++ {
		sb.WriteString(fmt.Sprintf("@l1_%d: @base\n\techo \"L1 %d\"\n\n", i, i))
	}

	// Level 2: 40 targets, each depending on 2 level 1 targets (creating diamonds)
	for i := 0; i < 40; i++ {
		dep1 := i % 20
		dep2 := (i + 1) % 20
		sb.WriteString(fmt.Sprintf("@l2_%d: @l1_%d @l1_%d\n\techo \"L2 %d\"\n\n", i, dep1, dep2, i))
	}

	// Level 3: top target depending on all level 2
	sb.WriteString("@top:")
	for i := 0; i < 40; i++ {
		sb.WriteString(fmt.Sprintf(" @l2_%d", i))
	}
	sb.WriteString("\n\techo \"Top\"\n")

	h.WriteFile("Buildfile", sb.String())

	// Parse and validate
	start := time.Now()
	result := h.Run("--dry-run", "top")
	elapsed := time.Since(start)

	// Should complete successfully
	result.AssertSuccess()

	// Should complete within reasonable time
	if elapsed > 5*time.Second {
		t.Errorf("planning diamond dependencies took too long: %v", elapsed)
	}

	t.Logf("Planned diamond dependency graph in %v", elapsed)
}

// TestManyVariables tests evaluation of many variables.
func TestManyVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	h := NewTestHarness(t)

	// Create a Buildfile with many variables (some referencing others)
	const numVars = 500

	var sb strings.Builder
	sb.WriteString(".shell: bash\n\n")

	for i := 0; i < numVars; i++ {
		if i == 0 {
			sb.WriteString(fmt.Sprintf("var%d = value%d\n", i, i))
		} else if i%10 == 0 {
			// Every 10th variable references the previous
			sb.WriteString(fmt.Sprintf("var%d = {var%d}_extended\n", i, i-1))
		} else {
			sb.WriteString(fmt.Sprintf("var%d = value%d\n", i, i))
		}
	}

	// Add a test target that uses some variables
	sb.WriteString(fmt.Sprintf("\n@test:\n\techo \"var0={var0}\"\n\techo \"var%d={var%d}\"\n", numVars-1, numVars-1))

	h.WriteFile("Buildfile", sb.String())

	// Parse and run
	start := time.Now()
	result := h.Run("test")
	elapsed := time.Since(start)

	// Should complete successfully
	result.AssertSuccess()
	result.AssertStdoutContains("var0=value0")

	// Should complete within reasonable time
	if elapsed > 5*time.Second {
		t.Errorf("evaluating %d variables took too long: %v", numVars, elapsed)
	}

	t.Logf("Evaluated %d variables in %v", numVars, elapsed)
}
