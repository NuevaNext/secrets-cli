// Package pass provides a wrapper around the pass password manager.
package pass

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	cmd.Env = append(os.Environ(), "PASSWORD_STORE_DIR="+p.StoreDir)

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
	cmd.Env = append(os.Environ(), "PASSWORD_STORE_DIR="+p.StoreDir)
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
	args := append([]string{"init"}, gpgIDs...)
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
	args := append([]string{"init"}, gpgIDs...)
	_, err := p.run(args...)
	return err
}

// GetGPGIDs reads the current GPG IDs from the store
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
