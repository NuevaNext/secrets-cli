#!/bin/bash
# Build the binary
make build

# Use a temporary directory for secrets
export SECRETS_DIR=".test-secrets-vault-repro"
rm -rf "$SECRETS_DIR"
mkdir -p "$SECRETS_DIR"

# Create a file that should NOT be deleted
echo "DO NOT DELETE" > "should_not_be_deleted.txt"

# Initialize
./secrets-cli init --email sentinel@example.com --secrets-dir "$SECRETS_DIR"

echo "Attempting path traversal in 'vault delete'..."
# Attempt path traversal in vault delete
# vaultName = ../../should_not_be_deleted.txt
# This should resolve to $SECRETS_DIR/vaults/../../should_not_be_deleted.txt -> should_not_be_deleted.txt
./secrets-cli vault delete "../../should_not_be_deleted.txt" --force --secrets-dir "$SECRETS_DIR"

if [ ! -f "should_not_be_deleted.txt" ]; then
    echo "VULNERABILITY CONFIRMED: Path traversal successful! File 'should_not_be_deleted.txt' was deleted."
    rm -rf "$SECRETS_DIR"
    exit 1
else
    echo "Path traversal failed (File still exists)"
    rm -rf "$SECRETS_DIR"
    rm "should_not_be_deleted.txt"
    exit 0
fi
