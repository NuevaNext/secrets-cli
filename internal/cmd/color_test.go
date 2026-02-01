package cmd

import (
	"os"
	"testing"
)

func TestColor(t *testing.T) {
	// Test NO_COLOR
	os.Setenv("NO_COLOR", "1")
	if res := color("test", "32"); res != "test" {
		t.Errorf("color() with NO_COLOR expected 'test', got %q", res)
	}
	os.Unsetenv("NO_COLOR")

	// Test non-TTY (default in tests)
	if res := color("test", "32"); res != "test" {
		t.Errorf("color() in non-TTY expected 'test', got %q", res)
	}
}
