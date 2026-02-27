#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# Build the binary
make build

# Initialize a store
rm -rf .test-secrets
# We need to be in a git repo for init to work if not absolute path, but current dir is a git repo.
./secrets-cli init --email test@example.com --secrets-dir .test-secrets

# 1. Test path traversal in 'key add' (Creation)
echo "Testing path traversal in 'key add'..."
# Create a dummy key file
echo "dummy" > dummy.asc
if ./secrets-cli key add "../../traversal-test" --key-file dummy.asc --secrets-dir .test-secrets 2>&1 | grep -q "invalid name"; then
    echo -e "${GREEN}✓ Caught path traversal in 'key add'${NC}"
else
    echo -e "${RED}✗ Failed to catch path traversal in 'key add'${NC}"
    rm dummy.asc
    exit 1
fi
rm dummy.asc

# 2. Test path traversal in 'list' (Read)
echo "Testing path traversal in 'list'..."
if ./secrets-cli list "../../../etc" --secrets-dir .test-secrets 2>&1 | grep -q "invalid name"; then
    echo -e "${GREEN}✓ Caught path traversal in 'list'${NC}"
else
    echo -e "${RED}✗ Failed to catch path traversal in 'list'${NC}"
    exit 1
fi

# 3. Test path traversal in 'delete' (Delete)
echo "Testing path traversal in 'delete'..."
if ./secrets-cli delete "../../../etc" "passwd" --force --secrets-dir .test-secrets 2>&1 | grep -q "invalid name"; then
    echo -e "${GREEN}✓ Caught path traversal in 'delete'${NC}"
else
    echo -e "${RED}✗ Failed to catch path traversal in 'delete'${NC}"
    exit 1
fi

# Clean up
rm -rf .test-secrets
echo "All traversal tests passed!"
