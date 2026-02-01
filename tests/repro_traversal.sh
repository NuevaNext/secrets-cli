#!/bin/bash
# Build the binary
make build

# Use a temporary directory for secrets
export SECRETS_DIR=".test-secrets-repro"
rm -rf "$SECRETS_DIR"

# Initialize
./secrets-cli init --email sentinel@example.com --secrets-dir "$SECRETS_DIR"

echo "Attempting path traversal in 'key add'..."
# Attempt path traversal in key add
# email = ../traversed
./secrets-cli key add "../traversed" --key-file README.md --secrets-dir "$SECRETS_DIR"

if [ -f "$SECRETS_DIR/traversed.asc" ]; then
    echo "VULNERABILITY CONFIRMED: Path traversal successful! File created at $SECRETS_DIR/traversed.asc"
    rm -rf "$SECRETS_DIR"
    exit 1
else
    echo "Path traversal failed"
    rm -rf "$SECRETS_DIR"
    exit 0
fi
