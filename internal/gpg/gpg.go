// Package gpg provides a wrapper around the gpg command-line tool.
package gpg

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GPG wraps gpg command execution
type GPG struct {
	Binary string
}

// New creates a new GPG wrapper with the specified binary path
func New(binary string) *GPG {
	if binary == "" {
		binary = "gpg"
	}
	return &GPG{Binary: binary}
}

// Key represents a GPG key
type Key struct {
	KeyID       string
	Fingerprint string
	Email       string
	Name        string
}

// run executes a gpg command and returns stdout
func (g *GPG) run(args ...string) (string, error) {
	cmd := exec.Command(g.Binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gpg error: %s: %w", stderr.String(), err)
	}

	return stdout.String(), nil
}

// ExportPublicKey exports a public key for the given email
func (g *GPG) ExportPublicKey(email string) ([]byte, error) {
	cmd := exec.Command(g.Binary, "--armor", "--export", "--", email)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to export key for %s: %s: %w", email, stderr.String(), err)
	}

	output := stdout.Bytes()
	if len(output) == 0 {
		return nil, fmt.Errorf("no public key found for %s", email)
	}

	return output, nil
}

// ExportPublicKeyToFile exports a public key to a file
func (g *GPG) ExportPublicKeyToFile(email, filePath string) error {
	key, err := g.ExportPublicKey(email)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, key, 0644); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

// ImportKey imports a key from a file
func (g *GPG) ImportKey(keyPath string) error {
	_, err := g.run("--import", keyPath)
	return err
}

// ImportKeyFromDir imports all keys from a directory
func (g *GPG) ImportKeyFromDir(keysDir string) (int, error) {
	entries, err := os.ReadDir(keysDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read keys directory: %w", err)
	}

	imported := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".asc") {
			keyPath := filepath.Join(keysDir, entry.Name())
			if err := g.ImportKey(keyPath); err != nil {
				// Log but continue - some keys may already be imported
				continue
			}
			imported++
		}
	}

	return imported, nil
}

// GetKeyID returns the key ID for an email address
func (g *GPG) GetKeyID(email string) (string, error) {
	output, err := g.run("--list-keys", "--keyid-format", "long", "--", email)
	if err != nil {
		return "", fmt.Errorf("no key found for %s", email)
	}

	// Parse output to extract key ID
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "pub") {
			// Format: pub   rsa4096/KEYID 2024-01-01 [SC]
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				keyPart := parts[1]
				if idx := strings.Index(keyPart, "/"); idx != -1 {
					return keyPart[idx+1:], nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not parse key ID for %s", email)
}

// GetFingerprint returns the fingerprint for an email address
func (g *GPG) GetFingerprint(email string) (string, error) {
	output, err := g.run("--fingerprint", "--", email)
	if err != nil {
		return "", err
	}

	// Parse fingerprint from output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Fingerprint line contains spaces between groups
		if strings.Contains(line, " ") && !strings.Contains(line, "=") && !strings.HasPrefix(line, "uid") && !strings.HasPrefix(line, "pub") && !strings.HasPrefix(line, "sub") {
			// Remove spaces to get clean fingerprint
			fp := strings.ReplaceAll(line, " ", "")
			if len(fp) == 40 { // GPG fingerprints are 40 hex chars
				return fp, nil
			}
		}
	}

	return "", fmt.Errorf("could not parse fingerprint for %s", email)
}

// KeyExists checks if a key exists for the given email
func (g *GPG) KeyExists(email string) bool {
	_, err := g.run("--list-keys", "--", email)
	return err == nil
}

// ListSecretKeys lists all secret (private) keys
func (g *GPG) ListSecretKeys() ([]Key, error) {
	output, err := g.run("--list-secret-keys", "--keyid-format", "long")
	if err != nil {
		return nil, err
	}

	return parseKeyList(output), nil
}

// ListPublicKeys lists all public keys
func (g *GPG) ListPublicKeys() ([]Key, error) {
	output, err := g.run("--list-keys", "--keyid-format", "long")
	if err != nil {
		return nil, err
	}

	return parseKeyList(output), nil
}

// parseKeyList parses gpg --list-keys output
func parseKeyList(output string) []Key {
	var keys []Key
	var currentKey *Key

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "pub") || strings.HasPrefix(line, "sec") {
			if currentKey != nil {
				keys = append(keys, *currentKey)
			}
			currentKey = &Key{}

			// Extract key ID
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				keyPart := parts[1]
				if idx := strings.Index(keyPart, "/"); idx != -1 {
					currentKey.KeyID = keyPart[idx+1:]
				}
			}
		} else if strings.HasPrefix(line, "uid") && currentKey != nil {
			// Extract name and email from uid line
			// Format: uid           [ultimate] Name <email@example.com>
			line = strings.TrimPrefix(line, "uid")
			line = strings.TrimSpace(line)

			// Remove trust level in brackets
			if idx := strings.Index(line, "]"); idx != -1 {
				line = strings.TrimSpace(line[idx+1:])
			}

			// Parse name and email
			if emailStart := strings.LastIndex(line, "<"); emailStart != -1 {
				currentKey.Name = strings.TrimSpace(line[:emailStart])
				if emailEnd := strings.LastIndex(line, ">"); emailEnd > emailStart {
					currentKey.Email = line[emailStart+1 : emailEnd]
				}
			} else {
				currentKey.Name = line
			}
		}
	}

	if currentKey != nil {
		keys = append(keys, *currentKey)
	}

	return keys
}
