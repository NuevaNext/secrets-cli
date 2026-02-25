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
    echo "Path traversal failed (expected)"
fi

echo "Attempting path traversal in 'vault delete'..."
mkdir -p "$SECRETS_DIR/vaults"
mkdir -p "$SECRETS_DIR/dangerous"
./secrets-cli vault delete "../dangerous" --force --secrets-dir "$SECRETS_DIR" 2>&1 | grep "invalid name"
if [ $? -eq 0 ]; then
    echo "Path traversal in 'vault delete' failed (expected)"
else
    echo "VULNERABILITY CONFIRMED: Path traversal in 'vault delete' successful!"
    rm -rf "$SECRETS_DIR"
    exit 1
fi

echo "Attempting path traversal in 'copy' (new-name flag)..."
mkdir -p "$SECRETS_DIR/vaults/src"
mkdir -p "$SECRETS_DIR/vaults/dst"
./secrets-cli copy "src" "secret" "dst" --new-name "../new" --secrets-dir "$SECRETS_DIR" 2>&1 | grep "invalid secret name"
if [ $? -eq 0 ]; then
    echo "Path traversal in 'copy' (new-name) failed (expected)"
else
    echo "VULNERABILITY CONFIRMED: Path traversal in 'copy' successful!"
    rm -rf "$SECRETS_DIR"
    exit 1
fi

rm -rf "$SECRETS_DIR"
echo "ALL TRAVERSAL TESTS PASSED"
exit 0
