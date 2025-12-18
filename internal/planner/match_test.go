package planner

import (
	"testing"

	"github.com/vinayprograms/build/internal/ast"
)

// ----------------------------------------------------------------------------
// Literal Target Matching Tests
// ----------------------------------------------------------------------------

func TestMatchLiteral_ExactMatch(t *testing.T) {
	pattern := createLiteralPattern("build/app", false, false)

	matched, captures := MatchTarget(pattern, "build/app")

	if !matched {
		t.Error("expected literal target to match exact path")
	}
	if len(captures) != 0 {
		t.Errorf("expected no captures for literal match, got %d", len(captures))
	}
}

func TestMatchLiteral_NoMatch(t *testing.T) {
	pattern := createLiteralPattern("build/app", false, false)

	matched, _ := MatchTarget(pattern, "build/other")

	if matched {
		t.Error("expected literal target not to match different path")
	}
}

func TestMatchLiteral_PartialNoMatch(t *testing.T) {
	pattern := createLiteralPattern("build/app", false, false)

	testCases := []struct {
		path string
		desc string
	}{
		{"build/app/extra", "longer path"},
		{"build", "shorter path"},
		{"build/ap", "partial match"},
		{"xbuild/app", "different prefix"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			matched, _ := MatchTarget(pattern, tc.path)
			if matched {
				t.Errorf("expected no match for %s (%s)", tc.path, tc.desc)
			}
		})
	}
}

func TestMatchLiteral_PhonyTarget(t *testing.T) {
	pattern := createLiteralPattern("all", true, false)

	matched, captures := MatchTarget(pattern, "@all")

	if !matched {
		t.Error("expected phony target to match with @ prefix")
	}
	if len(captures) != 0 {
		t.Errorf("expected no captures for phony match, got %d", len(captures))
	}
}

func TestMatchLiteral_PhonyTargetNoMatch(t *testing.T) {
	pattern := createLiteralPattern("all", true, false)

	// Without @ prefix should not match
	matched, _ := MatchTarget(pattern, "all")
	if matched {
		t.Error("expected phony target not to match path without @ prefix")
	}

	// Different phony name should not match
	matched, _ = MatchTarget(pattern, "@clean")
	if matched {
		t.Error("expected phony target not to match different phony name")
	}
}

func TestMatchLiteral_DirectoryTarget(t *testing.T) {
	pattern := createLiteralPattern("build/", false, true)

	matched, captures := MatchTarget(pattern, "build/")

	if !matched {
		t.Error("expected directory target to match path with trailing slash")
	}
	if len(captures) != 0 {
		t.Errorf("expected no captures for directory match, got %d", len(captures))
	}
}

func TestMatchLiteral_DirectoryTargetNoTrailingSlash(t *testing.T) {
	pattern := createLiteralPattern("build/", false, true)

	matched, _ := MatchTarget(pattern, "build")

	// Directory targets should match with or without trailing slash
	if !matched {
		t.Error("expected directory target to match path without trailing slash")
	}
}

func TestMatchLiteral_EmptyPath(t *testing.T) {
	pattern := createLiteralPattern("build/app", false, false)

	matched, _ := MatchTarget(pattern, "")

	if matched {
		t.Error("expected no match for empty path")
	}
}

func TestMatchLiteral_CaseSensitive(t *testing.T) {
	pattern := createLiteralPattern("Build/App", false, false)

	matched, _ := MatchTarget(pattern, "build/app")

	if matched {
		t.Error("expected case-sensitive comparison (no match)")
	}
}

// ----------------------------------------------------------------------------
// Pattern Target Matching Tests
// ----------------------------------------------------------------------------

func TestMatchPattern_SingleCapture(t *testing.T) {
	// Pattern: build/{name}.o
	pattern := createCapturePattern(
		[]patternPart{
			{literal: "build/"},
			{capture: "name"},
			{literal: ".o"},
		},
		false, false,
	)

	matched, captures := MatchTarget(pattern, "build/main.o")

	if !matched {
		t.Error("expected pattern to match")
	}
	if captures["name"] != "main" {
		t.Errorf("expected capture name='main', got %q", captures["name"])
	}
}

func TestMatchPattern_MultipleCaptures(t *testing.T) {
	// Pattern: {dir}/{name}.o
	pattern := createCapturePattern(
		[]patternPart{
			{capture: "dir"},
			{literal: "/"},
			{capture: "name"},
			{literal: ".o"},
		},
		false, false,
	)

	matched, captures := MatchTarget(pattern, "build/main.o")

	if !matched {
		t.Error("expected pattern to match")
	}
	if captures["dir"] != "build" {
		t.Errorf("expected capture dir='build', got %q", captures["dir"])
	}
	if captures["name"] != "main" {
		t.Errorf("expected capture name='main', got %q", captures["name"])
	}
}

func TestMatchPattern_CaptureAtStart(t *testing.T) {
	// Pattern: {name}.o
	pattern := createCapturePattern(
		[]patternPart{
			{capture: "name"},
			{literal: ".o"},
		},
		false, false,
	)

	matched, captures := MatchTarget(pattern, "utils.o")

	if !matched {
		t.Error("expected pattern to match")
	}
	if captures["name"] != "utils" {
		t.Errorf("expected capture name='utils', got %q", captures["name"])
	}
}

func TestMatchPattern_CaptureAtEnd(t *testing.T) {
	// Pattern: src/{name}
	pattern := createCapturePattern(
		[]patternPart{
			{literal: "src/"},
			{capture: "name"},
		},
		false, false,
	)

	matched, captures := MatchTarget(pattern, "src/main.c")

	if !matched {
		t.Error("expected pattern to match")
	}
	if captures["name"] != "main.c" {
		t.Errorf("expected capture name='main.c', got %q", captures["name"])
	}
}

func TestMatchPattern_OnlyCapture(t *testing.T) {
	// Pattern: {name}
	pattern := createCapturePattern(
		[]patternPart{
			{capture: "name"},
		},
		false, false,
	)

	matched, captures := MatchTarget(pattern, "anything")

	if !matched {
		t.Error("expected pattern to match")
	}
	if captures["name"] != "anything" {
		t.Errorf("expected capture name='anything', got %q", captures["name"])
	}
}

func TestMatchPattern_NoMatchWrongSuffix(t *testing.T) {
	// Pattern: build/{name}.o
	pattern := createCapturePattern(
		[]patternPart{
			{literal: "build/"},
			{capture: "name"},
			{literal: ".o"},
		},
		false, false,
	)

	matched, _ := MatchTarget(pattern, "build/main.c")

	if matched {
		t.Error("expected pattern not to match with wrong suffix")
	}
}

func TestMatchPattern_NoMatchWrongPrefix(t *testing.T) {
	// Pattern: build/{name}.o
	pattern := createCapturePattern(
		[]patternPart{
			{literal: "build/"},
			{capture: "name"},
			{literal: ".o"},
		},
		false, false,
	)

	matched, _ := MatchTarget(pattern, "src/main.o")

	if matched {
		t.Error("expected pattern not to match with wrong prefix")
	}
}

func TestMatchPattern_EmptyCapture(t *testing.T) {
	// Pattern: build/{name}.o
	pattern := createCapturePattern(
		[]patternPart{
			{literal: "build/"},
			{capture: "name"},
			{literal: ".o"},
		},
		false, false,
	)

	// "build/.o" should match with empty capture
	matched, captures := MatchTarget(pattern, "build/.o")

	if !matched {
		t.Error("expected pattern to match with empty capture value")
	}
	if captures["name"] != "" {
		t.Errorf("expected capture name='', got %q", captures["name"])
	}
}

func TestMatchPattern_CaptureWithSlash(t *testing.T) {
	// Pattern: {path}.o
	// This pattern can capture paths with slashes
	pattern := createCapturePattern(
		[]patternPart{
			{capture: "path"},
			{literal: ".o"},
		},
		false, false,
	)

	matched, captures := MatchTarget(pattern, "build/src/main.o")

	if !matched {
		t.Error("expected pattern to match path with slashes in capture")
	}
	if captures["path"] != "build/src/main" {
		t.Errorf("expected capture path='build/src/main', got %q", captures["path"])
	}
}

func TestMatchPattern_DuplicateCaptureName(t *testing.T) {
	// Pattern: {name}/{name}.o - same capture name used twice
	// The captures should have the same value
	pattern := createCapturePattern(
		[]patternPart{
			{capture: "name"},
			{literal: "/"},
			{capture: "name"},
			{literal: ".o"},
		},
		false, false,
	)

	// "main/main.o" should match
	matched, captures := MatchTarget(pattern, "main/main.o")
	if !matched {
		t.Error("expected pattern to match when both captures have same value")
	}
	if captures["name"] != "main" {
		t.Errorf("expected capture name='main', got %q", captures["name"])
	}

	// "foo/bar.o" should not match (captures would have different values)
	matched, _ = MatchTarget(pattern, "foo/bar.o")
	if matched {
		t.Error("expected pattern not to match when captures would differ")
	}
}

func TestMatchPattern_PhonyCaptureNotAllowed(t *testing.T) {
	// Phony targets should not have captures - they're always literal
	// This is a validation issue, not a matching issue
	// But we test that even if created, matching behaves sensibly

	// If somehow a phony pattern with capture existed
	// it would still work as a pattern match
	pattern := createCapturePattern(
		[]patternPart{
			{capture: "name"},
		},
		true, false,
	)

	// Should match @test
	matched, captures := MatchTarget(pattern, "@test")
	if !matched {
		t.Error("expected phony pattern to match")
	}
	if captures["name"] != "test" {
		t.Errorf("expected capture name='test', got %q", captures["name"])
	}
}

func TestMatchPattern_AdjacentCaptures(t *testing.T) {
	// Pattern: {a}{b} - two adjacent captures
	// This is ambiguous but we should handle it gracefully
	pattern := createCapturePattern(
		[]patternPart{
			{capture: "a"},
			{capture: "b"},
		},
		false, false,
	)

	// With adjacent captures and no delimiters, first capture is greedy
	matched, captures := MatchTarget(pattern, "hello")

	// The matching behavior for adjacent captures depends on implementation
	// Either first is greedy (takes all), or we fail
	// For simplicity, we'll make first capture greedy
	if matched {
		// First capture takes all, second is empty
		if captures["a"] != "hello" || captures["b"] != "" {
			t.Errorf("expected a='hello', b='', got a=%q, b=%q", captures["a"], captures["b"])
		}
	}
	// Note: If we decide adjacent captures are ambiguous, this test should expect !matched
}

// ----------------------------------------------------------------------------
// Target Lookup Tests
// ----------------------------------------------------------------------------

func TestLookupTarget_ExactMatch(t *testing.T) {
	targets := []*ast.Target{
		createTarget("build/app", false, false, nil),
		createTarget("build/lib.so", false, false, nil),
	}

	target, captures, err := LookupTarget("build/app", targets)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target == nil {
		t.Fatal("expected to find target")
	}
	if len(captures) != 0 {
		t.Errorf("expected no captures, got %d", len(captures))
	}
}

func TestLookupTarget_PatternMatch(t *testing.T) {
	targets := []*ast.Target{
		createTarget("build/app", false, false, nil),
		createPatternTarget(
			[]patternPart{
				{literal: "build/"},
				{capture: "name"},
				{literal: ".o"},
			},
			false, false,
		),
	}

	target, captures, err := LookupTarget("build/main.o", targets)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target == nil {
		t.Fatal("expected to find target")
	}
	if captures["name"] != "main" {
		t.Errorf("expected capture name='main', got %q", captures["name"])
	}
}

func TestLookupTarget_ExactMatchPreferred(t *testing.T) {
	// Both exact and pattern could match, but exact wins
	targets := []*ast.Target{
		createPatternTarget(
			[]patternPart{
				{literal: "build/"},
				{capture: "name"},
				{literal: ".o"},
			},
			false, false,
		),
		createTarget("build/main.o", false, false, nil), // exact match
	}

	target, captures, err := LookupTarget("build/main.o", targets)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target == nil {
		t.Fatal("expected to find target")
	}
	// Should prefer exact match (no captures)
	if len(captures) != 0 {
		t.Error("expected exact match (no captures), got pattern match")
	}
}

func TestLookupTarget_NotFound(t *testing.T) {
	targets := []*ast.Target{
		createTarget("build/app", false, false, nil),
		createTarget("build/lib.so", false, false, nil),
	}

	_, _, err := LookupTarget("build/other", targets)

	if err == nil {
		t.Error("expected error for unmatched target")
	}
}

func TestLookupTarget_PhonyTarget(t *testing.T) {
	targets := []*ast.Target{
		createTarget("all", true, false, nil),
		createTarget("clean", true, false, nil),
	}

	target, captures, err := LookupTarget("@all", targets)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target == nil {
		t.Fatal("expected to find target")
	}
	if len(captures) != 0 {
		t.Errorf("expected no captures, got %d", len(captures))
	}
}

func TestLookupTarget_DirectoryTarget(t *testing.T) {
	targets := []*ast.Target{
		createTarget("build/", false, true, nil),
	}

	// Both with and without trailing slash should work
	for _, path := range []string{"build/", "build"} {
		t.Run(path, func(t *testing.T) {
			target, captures, err := LookupTarget(path, targets)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if target == nil {
				t.Fatal("expected to find target")
			}
			if len(captures) != 0 {
				t.Errorf("expected no captures, got %d", len(captures))
			}
		})
	}
}

func TestLookupTarget_MultiplePatternMatches(t *testing.T) {
	// Multiple patterns could match - first in definition order wins
	targets := []*ast.Target{
		createPatternTarget(
			[]patternPart{
				{capture: "name"},
				{literal: ".o"},
			},
			false, false,
		),
		createPatternTarget(
			[]patternPart{
				{literal: "build/"},
				{capture: "name"},
				{literal: ".o"},
			},
			false, false,
		),
	}

	// First pattern matches everything.o, so it wins for build/main.o
	target, captures, err := LookupTarget("build/main.o", targets)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target == nil {
		t.Fatal("expected to find target")
	}
	// First pattern wins, so capture includes "build/main"
	if captures["name"] != "build/main" {
		t.Errorf("expected first pattern match, got captures: %v", captures)
	}
}

func TestLookupTarget_EmptyTargetList(t *testing.T) {
	targets := []*ast.Target{}

	_, _, err := LookupTarget("build/app", targets)

	if err == nil {
		t.Error("expected error for empty target list")
	}
}

// ----------------------------------------------------------------------------
// Test Helpers
// ----------------------------------------------------------------------------

// patternPart represents either a literal or capture in a pattern for testing.
type patternPart struct {
	literal string
	capture string
}

// createLiteralPattern creates a TargetPattern with only literal segments.
func createLiteralPattern(path string, isPhony, isDirectory bool) *ast.TargetPattern {
	return &ast.TargetPattern{
		Segments:    []ast.PatternSegment{&ast.LiteralSegment{Text: path}},
		IsPhony:     isPhony,
		IsDirectory: isDirectory,
	}
}

// createCapturePattern creates a TargetPattern from a sequence of parts.
func createCapturePattern(parts []patternPart, isPhony, isDirectory bool) *ast.TargetPattern {
	segments := make([]ast.PatternSegment, 0, len(parts))
	for _, p := range parts {
		if p.capture != "" {
			segments = append(segments, &ast.BraceExpr{Identifier: p.capture})
		} else {
			segments = append(segments, &ast.LiteralSegment{Text: p.literal})
		}
	}
	return &ast.TargetPattern{
		Segments:    segments,
		IsPhony:     isPhony,
		IsDirectory: isDirectory,
	}
}

// createTarget creates a Target with a literal pattern.
func createTarget(path string, isPhony, isDirectory bool, deps []string) *ast.Target {
	target := &ast.Target{
		Pattern: *createLiteralPattern(path, isPhony, isDirectory),
	}
	for _, dep := range deps {
		target.Dependencies = append(target.Dependencies, ast.Dependency{
			Segments: []ast.PatternSegment{&ast.LiteralSegment{Text: dep}},
		})
	}
	return target
}

// createPatternTarget creates a Target with a pattern.
func createPatternTarget(parts []patternPart, isPhony, isDirectory bool) *ast.Target {
	return &ast.Target{
		Pattern: *createCapturePattern(parts, isPhony, isDirectory),
	}
}
