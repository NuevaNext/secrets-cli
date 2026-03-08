#!/bin/bash
SECRET_CLI_BIN=$(pwd)/secrets-cli
WORKSPACE=$(mktemp -d)
cd "$WORKSPACE"
echo "sensitive data" > sensitive.asc
mkdir -p .secrets/keys
echo "Testing for path traversal vulnerability in 'key remove'..."
$SECRET_CLI_BIN key remove "../../sensitive"
if [ ! -f sensitive.asc ]; then
    echo "VULNERABILITY CONFIRMED: sensitive.asc was deleted!"
    rm -rf "$WORKSPACE"
    # End script with non-zero
    (exit 1)
else
    echo "SUCCESS: sensitive.asc was NOT deleted."
    rm -rf "$WORKSPACE"
    # End script with zero
    (exit 0)
fi
