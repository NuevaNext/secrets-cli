#!/bin/bash
# Test for re-encryption verification with untrusted keys
# This test simulates the real-world bug where adding a member with an untrusted key
# would silently fail to re-encrypt secrets.

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo "Testing re-encryption with untrusted GPG key..."

# Create test directory
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"

# Initialize git
git init --quiet
git config user.email "alice@test.com"
git config user.name "Alice"

# Generate Alice's key (trusted, in local keyring)
echo "Generating Alice's key..."
gpg --batch --gen-key <<EOF
%no-protection
Key-Type: RSA
Key-Length: 2048
Subkey-Type: RSA
Subkey-Length: 2048
Name-Real: Alice Test
Name-Email: alice@test.com
Expire-Date: 0
%commit
EOF

# Generate Charlie's key in a SEPARATE keyring (simulates external key)
echo "Generating Charlie's key (external/untrusted)..."
CHARLIE_GNUPGHOME=$(mktemp -d)
export GNUPGHOME="$CHARLIE_GNUPGHOME"
gpg --batch --gen-key <<EOF
%no-protection
Key-Type: RSA
Key-Length: 2048
Subkey-Type: RSA
Subkey-Length: 2048
Name-Real: Charlie External
Name-Email: charlie@external.com
Expire-Date: 0
%commit
EOF

# Export Charlie's public key
gpg --armor --export charlie@external.com > /tmp/charlie.asc

# Switch back to main keyring
unset GNUPGHOME

# Import Charlie's key (will be UNTRUSTED)
echo "Importing Charlie's key as untrusted..."
gpg --import /tmp/charlie.asc

# Verify Charlie's key is untrusted
TRUST_LEVEL=$(gpg --list-keys --with-colons charlie@external.com | grep '^pub:' | cut -d: -f2)
if [[ "$TRUST_LEVEL" != "-" && "$TRUST_LEVEL" != "q" ]]; then
    echo -e "${RED}SETUP ERROR: Charlie's key should be untrusted but has trust level: $TRUST_LEVEL${NC}"
    exit 1
fi
echo "✓ Charlie's key is untrusted (trust level: $TRUST_LEVEL)"

# Initialize secrets-cli
echo "Initializing secrets-cli..."
secrets-cli init --email alice@test.com

# Create vault and add a secret
echo "Creating vault and secret..."
secrets-cli vault create test
echo "secret-value-123" | secrets-cli --email alice@test.com set test my/secret

# Get Charlie's key ID
CHARLIE_KEY_ID=$(gpg --list-keys --with-colons charlie@external.com | grep '^sub:' | grep ':e:' | head -1 | cut -d: -f5)
echo "Charlie's encryption key ID: $CHARLIE_KEY_ID"

# Add Charlie's key to secrets-cli
echo "Adding Charlie to vault..."
secrets-cli key add charlie@external.com

# THIS IS THE CRITICAL TEST: Add Charlie as a member
# With the bug, this would succeed but NOT actually re-encrypt
secrets-cli --email alice@test.com vault add-member test charlie@external.com

# VERIFICATION: Check if the secret is actually encrypted for Charlie
echo "Verifying re-encryption..."
SECRET_FILE=".secrets/vaults/test/.password-store/my/secret.gpg"

# Get the list of recipient key IDs from the encrypted file
RECIPIENTS=$(gpg --list-packets "$SECRET_FILE" 2>&1 | grep 'keyid' | awk '{print $NF}')

# Check if Charlie's key ID is in the recipients
if echo "$RECIPIENTS" | grep -q "$CHARLIE_KEY_ID"; then
    echo -e "${GREEN}✓ SUCCESS: Secret is encrypted for Charlie's key ($CHARLIE_KEY_ID)${NC}"
    echo "Recipients: $RECIPIENTS"
else
    echo -e "${RED}✗ FAILURE: Secret is NOT encrypted for Charlie's key${NC}"
    echo "Expected key ID: $CHARLIE_KEY_ID"
    echo "Found recipients: $RECIPIENTS"
    exit 1
fi

# Cleanup
cd /
rm -rf "$TEST_DIR" "$CHARLIE_GNUPGHOME" /tmp/charlie.asc

echo -e "${GREEN}All tests passed!${NC}"
