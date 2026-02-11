// Package pass provides a wrapper around the pass password manager.
package pass

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Pass wraps pass command execution
type Pass struct {
	StoreDir string // PASSWORD_STORE_DIR
}

// New creates a new Pass wrapper for a specific store directory
func New(storeDir string) *Pass {
	return &Pass{StoreDir: storeDir}
}

// run executes a pass command with PASSWORD_STORE_DIR set
func (p *Pass) run(args ...string) (string, error) {
	cmd := exec.Command("pass", args...)
	// Preserve existing PASSWORD_STORE_GPG_OPTS and append --trust-model always
	existingOpts := os.Getenv("PASSWORD_STORE_GPG_OPTS")
	gpgOpts := "--trust-model always"
	if existingOpts != "" {
		gpgOpts = existingOpts + " " + gpgOpts
	}
	cmd.Env = append(os.Environ(),
		"PASSWORD_STORE_DIR="+p.StoreDir,
		"PASSWORD_STORE_GPG_OPTS="+gpgOpts,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("pass error: %s", errMsg)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// runWithStdin executes a pass command with stdin input
func (p *Pass) runWithStdin(input string, args ...string) (string, error) {
	cmd := exec.Command("pass", args...)
	// Preserve existing PASSWORD_STORE_GPG_OPTS and append --trust-model always
	existingOpts := os.Getenv("PASSWORD_STORE_GPG_OPTS")
	gpgOpts := "--trust-model always"
	if existingOpts != "" {
		gpgOpts = existingOpts + " " + gpgOpts
	}
	cmd.Env = append(os.Environ(),
		"PASSWORD_STORE_DIR="+p.StoreDir,
		"PASSWORD_STORE_GPG_OPTS="+gpgOpts,
	)
	cmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("pass error: %s", errMsg)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// Init initializes the password store with GPG IDs
func (p *Pass) Init(gpgIDs []string) error {
	args := append([]string{"init", "--"}, gpgIDs...)
	_, err := p.run(args...)
	return err
}

// Insert adds or updates a secret (overwrites if exists)
func (p *Pass) Insert(name, value string) error {
	// Use insert with multiline and force to overwrite
	_, err := p.runWithStdin(value, "insert", "--multiline", "--force", "--", name)
	return err
}

// Show retrieves a secret value
func (p *Pass) Show(name string) (string, error) {
	return p.run("show", "--", name)
}

// Exists checks if a secret exists
func (p *Pass) Exists(name string) bool {
	_, err := p.run("show", "--", name)
	return err == nil
}

// Remove deletes a secret
func (p *Pass) Remove(name string) error {
	_, err := p.run("rm", "--force", "--", name)
	return err
}

// Move renames a secret
func (p *Pass) Move(oldName, newName string) error {
	_, err := p.run("mv", "--force", "--", oldName, newName)
	return err
}

// Copy copies a secret
func (p *Pass) Copy(srcName, dstName string) error {
	_, err := p.run("cp", "--force", "--", srcName, dstName)
	return err
}

// List returns all secret names in the store
func (p *Pass) List() ([]string, error) {
	return p.listDir("")
}

// listDir lists secrets recursively from a directory
func (p *Pass) listDir(prefix string) ([]string, error) {
	dir := p.StoreDir
	if prefix != "" {
		dir = filepath.Join(dir, prefix)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read store: %w", err)
	}

	var secrets []string
	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files and .gpg-id
		if strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := name
		if prefix != "" {
			fullPath = filepath.Join(prefix, name)
		}

		if entry.IsDir() {
			// Recurse into subdirectories
			subSecrets, err := p.listDir(fullPath)
			if err != nil {
				continue
			}
			secrets = append(secrets, subSecrets...)
		} else if strings.HasSuffix(name, ".gpg") {
			// Remove .gpg extension
			secretName := strings.TrimSuffix(fullPath, ".gpg")
			secrets = append(secrets, secretName)
		}
	}

	return secrets, nil
}

// ReInit re-initializes the store with new GPG IDs (re-encrypts all secrets)
func (p *Pass) ReInit(gpgIDs []string) error {
	// Write new .gpg-id file
	gpgIDPath := filepath.Join(p.StoreDir, ".gpg-id")
	content := strings.Join(gpgIDs, "\n") + "\n"
	if err := os.WriteFile(gpgIDPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write .gpg-id: %w", err)
	}

	// Re-init to re-encrypt all secrets
	args := append([]string{"init", "--"}, gpgIDs...)
	if _, err := p.run(args...); err != nil {
		return err
	}

	// Verify re-encryption succeeded for at least one secret
	secrets, err := p.List()
	if err != nil {
		return fmt.Errorf("failed to list secrets after re-init: %w", err)
	}

	// If there are secrets, verify at least the first one is encrypted correctly
	if len(secrets) > 0 {
		if err := p.VerifyEncryption(secrets[0], gpgIDs); err != nil {
			return fmt.Errorf("re-encryption verification failed: %w", err)
		}
	}

	return nil
}

// VerifyEncryption checks if a secret is encrypted for the expected GPG IDs.
// It uses a count-based approach which is more robust across GPG versions than
// trying to match exact key IDs (which can vary in format).
func (p *Pass) VerifyEncryption(secretName string, expectedGPGIDs []string) error {
secretPath := filepath.Join(p.StoreDir, secretName+".gpg")

// First, verify all expected GPG IDs exist in the keyring
for _, gpgID := range expectedGPGIDs {
cmd := exec.Command("gpg", "--list-keys", gpgID)
if err := cmd.Run(); err != nil {
return fmt.Errorf("GPG ID %s not found in keyring: %w", gpgID, err)
}
}

// Count recipients in the encrypted file
cmd := exec.Command("gpg", "--list-packets", secretPath)
var stdout bytes.Buffer
cmd.Stdout = &stdout

if err := cmd.Run(); err != nil {
return fmt.Errorf("failed to list packets: %w", err)
}

// Count how many encryption recipients are in the file
// Each ":pubkey enc packet:" line represents one recipient
keyIDRegex := regexp.MustCompile(`(?i):pubkey enc packet:`)
matches := keyIDRegex.FindAllString(stdout.String(), -1)
recipientCount := len(matches)

if recipientCount == 0 {
return fmt.Errorf("no encryption recipients found in %s", secretName)
}

// Verify the count matches
// Since we know pass was asked to encrypt to exactly these GPG IDs,
// if the recipient count matches, encryption was successful
if recipientCount != len(expectedGPGIDs) {
return fmt.Errorf("secret %s is encrypted for %d recipients, but expected %d (GPG IDs: %v)",
secretName, recipientCount, len(expectedGPGIDs), expectedGPGIDs)
}

return nil
}

func (p *Pass) GetGPGIDs() ([]string, error) {
	gpgIDPath := filepath.Join(p.StoreDir, ".gpg-id")
	data, err := os.ReadFile(gpgIDPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read .gpg-id: %w", err)
	}

	var ids []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			ids = append(ids, line)
		}
	}

	return ids, nil
}
