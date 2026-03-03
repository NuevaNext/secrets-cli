package cmd

import (
	"os"
	"testing"
)

func TestConfirm(t *testing.T) {
	// Test force flag
	if !Confirm("test", true) {
		t.Errorf("Confirm with force=true should return true")
	}

	// Test non-terminal environment
	// In most CI environments, os.Stdin is not a terminal.
	// Even if it were, we want to ensure it handles non-terminal gracefully.

	// Create a pipe to simulate non-terminal stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	if Confirm("test", false) {
		t.Errorf("Confirm without force in non-terminal should return false")
	}
}
