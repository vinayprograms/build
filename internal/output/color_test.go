package output

import (
	"os"
	"testing"
)

func TestShouldUseColor_Always(t *testing.T) {
	if !ShouldUseColor("always") {
		t.Error("expected ShouldUseColor('always') to return true")
	}
}

func TestShouldUseColor_Never(t *testing.T) {
	if ShouldUseColor("never") {
		t.Error("expected ShouldUseColor('never') to return false")
	}
}

func TestShouldUseColor_Auto_NoColor(t *testing.T) {
	old := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", old)

	os.Setenv("NO_COLOR", "1")
	if ShouldUseColor("auto") {
		t.Error("expected ShouldUseColor('auto') to return false when NO_COLOR is set")
	}
}

func TestShouldUseColor_Auto_ForceColor(t *testing.T) {
	oldNo := os.Getenv("NO_COLOR")
	oldForce := os.Getenv("FORCE_COLOR")
	defer func() {
		os.Setenv("NO_COLOR", oldNo)
		os.Setenv("FORCE_COLOR", oldForce)
	}()

	os.Unsetenv("NO_COLOR")
	os.Setenv("FORCE_COLOR", "1")
	if !ShouldUseColor("auto") {
		t.Error("expected ShouldUseColor('auto') to return true when FORCE_COLOR is set")
	}
}

func TestColorize_Enabled(t *testing.T) {
	result := Colorize("hello", ColorRed, true)
	expected := "\033[31mhello\033[0m"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestColorize_Disabled(t *testing.T) {
	result := Colorize("hello", ColorRed, false)
	if result != "hello" {
		t.Errorf("expected 'hello' unchanged, got %q", result)
	}
}

func TestBold_Enabled(t *testing.T) {
	result := Bold("hello", true)
	expected := "\033[1mhello\033[0m"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestBold_Disabled(t *testing.T) {
	result := Bold("hello", false)
	if result != "hello" {
		t.Errorf("expected 'hello' unchanged, got %q", result)
	}
}

func TestDim_Enabled(t *testing.T) {
	result := Dim("hello", true)
	expected := "\033[2mhello\033[0m"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestDim_Disabled(t *testing.T) {
	result := Dim("hello", false)
	if result != "hello" {
		t.Errorf("expected 'hello' unchanged, got %q", result)
	}
}

func TestColorizeStatus_Success(t *testing.T) {
	result := ColorizeStatus("OK", true, true)
	expected := "\033[32mOK\033[0m"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestColorizeStatus_Failure(t *testing.T) {
	result := ColorizeStatus("FAIL", false, true)
	expected := "\033[1;31mFAIL\033[0m"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestColorizeStatus_Disabled(t *testing.T) {
	if ColorizeStatus("OK", true, false) != "OK" {
		t.Error("expected unchanged text when colors disabled")
	}
	if ColorizeStatus("FAIL", false, false) != "FAIL" {
		t.Error("expected unchanged text when colors disabled")
	}
}
