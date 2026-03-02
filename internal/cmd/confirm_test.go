package cmd

import (
	"testing"
)

func TestConfirm(t *testing.T) {
	// Test forced bypass
	if !Confirm("Are you sure?", true) {
		t.Error("Expected Confirm to return true when forced, even if not in a terminal")
	}

	// Test non-terminal environment (which is the case during tests)
	if Confirm("Are you sure?", false) {
		t.Error("Expected Confirm to return false in non-terminal environment")
	}
}
