package pass

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestVerifyEncryption tests the VerifyEncryption function with real GPG files
func TestVerifyEncryption(t *testing.T) {
	// Skip if GPG is not available
	if _, err := exec.LookPath("gpg"); err != nil {
		t.Skip("gpg not available in PATH")
	}

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pass-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Generate a test GPG key
	keyEmail := "test@example.com"
	generateTestKey(t, keyEmail)

	// Create a Pass instance
	p := &Pass{StoreDir: tmpDir}

	// Create a test secret file
	secretName := "test-secret"
	secretPath := filepath.Join(tmpDir, secretName+".gpg")

	// Encrypt a test message to our key
	cmd := exec.Command("gpg", "--batch", "--yes", "--encrypt", "--recipient", keyEmail, "--output", secretPath)
	cmd.Stdin = strings.NewReader("test secret value")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create encrypted file: %v", err)
	}

	// Test 1: Verify with correct GPG ID
	t.Run("CorrectGPGID", func(t *testing.T) {
		err := p.VerifyEncryption(secretName, []string{keyEmail})
		if err != nil {
			t.Errorf("Expected verification to succeed, got error: %v", err)
		}
	})

	// Test 2: Verify with wrong GPG ID
	t.Run("WrongGPGID", func(t *testing.T) {
		err := p.VerifyEncryption(secretName, []string{"wrong@example.com"})
		if err == nil {
			t.Error("Expected verification to fail for wrong GPG ID")
		}
	})

	// Test 3: Verify with multiple GPG IDs (only one correct)
	t.Run("MultipleGPGIDsPartial", func(t *testing.T) {
		err := p.VerifyEncryption(secretName, []string{keyEmail, "other@example.com"})
		if err == nil {
			t.Error("Expected verification to fail when not all recipients are present")
		}
	})
}

// generateTestKey generates a GPG key for testing
func generateTestKey(t *testing.T, email string) {
	t.Helper()

	// Check if key already exists
	cmd := exec.Command("gpg", "--list-keys", email)
	if cmd.Run() == nil {
		// Key exists, skip generation
		return
	}

	// Generate a new key
	keySpec := `
%no-protection
Key-Type: RSA
Key-Length: 2048
Name-Real: Test User
Name-Email: ` + email + `
Expire-Date: 0
%commit
`

	cmd = exec.Command("gpg", "--batch", "--gen-key")
	cmd.Stdin = strings.NewReader(keySpec)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to generate test key: %v\nStderr: %s", err, stderr.String())
	}
}
