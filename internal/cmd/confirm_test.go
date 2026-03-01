package cmd

import (
	"os"
	"testing"
)

func TestConfirm(t *testing.T) {
	// Test with force=true (should always be true)
	if !Confirm("test", true) {
		t.Error("Confirm(..., true) should return true")
	}

	// Test with force=false in non-interactive environment (should be false)
	// Note: go test usually runs in a non-interactive environment
	fileInfo, err := os.Stdin.Stat()
	if err == nil && (fileInfo.Mode()&os.ModeCharDevice) == 0 {
		if Confirm("test", false) {
			t.Error("Confirm(..., false) should return false in non-interactive environment")
		}
	}
}
