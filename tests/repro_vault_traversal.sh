#!/bin/bash
# Build the binary
make build

# Use a temporary directory for secrets
export SECRETS_DIR=".test-secrets-vault-repro"
rm -rf "$SECRETS_DIR"
mkdir -p "$SECRETS_DIR"
touch "$SECRETS_DIR/sensitive_file"

# Initialize
./secrets-cli init --email sentinel@example.com --secrets-dir "$SECRETS_DIR"

echo "Attempting path traversal in 'vault delete'..."
# Attempt path traversal in vault delete
# vaultName = ../sensitive_file
./secrets-cli vault delete "../sensitive_file" --force --secrets-dir "$SECRETS_DIR"

if [ ! -f "$SECRETS_DIR/sensitive_file" ]; then
    echo "VULNERABILITY CONFIRMED: Path traversal successful! File deleted at $SECRETS_DIR/sensitive_file"
    rm -rf "$SECRETS_DIR"
    exit 1
else
    echo "Path traversal failed (File still exists)"
    rm -rf "$SECRETS_DIR"
    exit 0
fi
