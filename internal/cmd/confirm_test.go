package cmd

import (
	"os"
	"testing"
)

func TestConfirm(t *testing.T) {
	// Test forced confirmation
	if !Confirm("test", true) {
		t.Errorf("Confirm should return true when forced")
	}

	// Test non-interactive environment
	// In the test environment, os.Stdin might not be a terminal
	fileInfo, err := os.Stdin.Stat()
	isTerminal := err == nil && (fileInfo.Mode()&os.ModeCharDevice) != 0

	if !isTerminal {
		if Confirm("test", false) {
			t.Errorf("Confirm should return false in non-interactive environment")
		}
	}
}
