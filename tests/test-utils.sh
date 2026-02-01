#!/usr/bin/env bash
#
# test-utils.sh - Utility functions for secret-cli tests
#
# Provides colored output, test assertions, and result formatting
#

# Colors
if [[ -n "${NO_COLOR:-}" ]]; then
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    CYAN=''
    BOLD=''
    DIM=''
    NC=''
else
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BLUE='\033[0;34m'
    CYAN='\033[0;36m'
    BOLD='\033[1m'
    DIM='\033[2m'
    NC='\033[0m'
fi

# Test state
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0
CURRENT_TEST=""
VERBOSE=${VERBOSE:-false}
TEST_TIMEOUT=${TEST_TIMEOUT:-30}  # Default 30 seconds per test

# Captured output
LAST_OUTPUT=""
LAST_EXIT_CODE=0
TIMED_OUT=false

# =============================================================================
# Output Functions
# =============================================================================

# Print a header for a test group
test_group() {
    local name="$1"
    echo ""
    echo -e "${BOLD}${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}${BLUE}  $name${NC}"
    echo -e "${BOLD}${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Start a test
test_start() {
    local test_name="$1"
    CURRENT_TEST="$test_name"
    ((TESTS_RUN++)) || true
    
    if [[ "$VERBOSE" == "true" ]]; then
        echo ""
        echo -e "${CYAN}┌─ TEST: ${test_name}${NC}"
    fi
}

# Log verbose output
test_log() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${DIM}│  $*${NC}"
    fi
}

# Test passed
test_pass() {
    ((TESTS_PASSED++)) || true
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${GREEN}└─ ✓ PASSED${NC}"
    else
        echo -e "  ${GREEN}✓${NC} ${CURRENT_TEST}"
    fi
}

# Test failed with expected vs actual
test_fail() {
    local expected="$1"
    local actual="$2"
    local details="${3:-}"
    
    ((TESTS_FAILED++)) || true
    
    echo ""
    echo -e "  ${RED}✗${NC} ${CURRENT_TEST}"
    echo -e "    ${RED}FAILED${NC}"
    echo -e "    ${BOLD}Expected:${NC} ${expected}"
    echo -e "    ${BOLD}Actual:${NC}   ${actual}"
    if [[ -n "$details" ]]; then
        echo -e "    ${BOLD}Details:${NC}"
        echo "$details" | sed 's/^/      /'
    fi
    if [[ -n "$LAST_OUTPUT" ]]; then
        echo -e "    ${BOLD}Command output:${NC}"
        echo "$LAST_OUTPUT" | head -20 | sed 's/^/      /'
        if [[ $(echo "$LAST_OUTPUT" | wc -l) -gt 20 ]]; then
            echo "      ... (truncated)"
        fi
    fi
    echo ""
}

# Print final summary
test_summary() {
    echo ""
    echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}  TEST SUMMARY${NC}"
    echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  Total:  ${TESTS_RUN}"
    echo -e "  ${GREEN}Passed: ${TESTS_PASSED}${NC}"
    if [[ $TESTS_FAILED -gt 0 ]]; then
        echo -e "  ${RED}Failed: ${TESTS_FAILED}${NC}"
    else
        echo -e "  Failed: ${TESTS_FAILED}"
    fi
    echo ""
    
    if [[ $TESTS_FAILED -eq 0 ]]; then
        echo -e "  ${GREEN}${BOLD}All tests passed! ✓${NC}"
        return 0
    else
        echo -e "  ${RED}${BOLD}Some tests failed ✗${NC}"
        return 1
    fi
}

# =============================================================================
# Command Execution
# =============================================================================

# Run a command and capture output (with timeout)
run_cmd() {
    local cmd="$*"
    test_log "Running: $cmd"
    
    # Reset state before running
    LAST_EXIT_CODE=0
    TIMED_OUT=false
    
    # Capture output and exit code properly with timeout
    set +e
    LAST_OUTPUT=$(timeout "$TEST_TIMEOUT" bash -c "$cmd" 2>&1)
    LAST_EXIT_CODE=$?
    set -e
    
    # Check for timeout (exit code 124)
    if [[ $LAST_EXIT_CODE -eq 124 ]]; then
        TIMED_OUT=true
        LAST_OUTPUT="TIMEOUT: Command exceeded ${TEST_TIMEOUT}s limit"$'\n'"$LAST_OUTPUT"
    fi
    
    if [[ "$VERBOSE" == "true" ]] && [[ -n "$LAST_OUTPUT" ]]; then
        echo "$LAST_OUTPUT" | while IFS= read -r line; do
            echo -e "${DIM}│    > ${line}${NC}"
        done
    fi
    
    return $LAST_EXIT_CODE
}

# Run command expecting success (exit code 0)
run_expecting_success() {
    local cmd="$*"
    run_cmd "$cmd"
    local exit_code=$?
    
    if [[ $exit_code -ne 0 ]]; then
        return 1
    fi
    return 0
}

# Run command expecting failure (non-zero exit code)
run_expecting_failure() {
    local cmd="$*"
    run_cmd "$cmd"
    local exit_code=$?
    
    if [[ $exit_code -eq 0 ]]; then
        return 1
    fi
    return 0
}

# =============================================================================
# Assertions
# =============================================================================

# Assert command succeeds
assert_success() {
    local cmd="$*"
    if ! run_expecting_success "$cmd"; then
        if [[ "$TIMED_OUT" == "true" ]]; then
            test_fail "exit code 0 (success)" "TIMEOUT after ${TEST_TIMEOUT}s"
        else
            test_fail "exit code 0 (success)" "exit code $LAST_EXIT_CODE"
        fi
        return 1
    fi
    return 0
}

# Assert command fails
assert_failure() {
    local cmd="$*"
    if ! run_expecting_failure "$cmd"; then
        test_fail "non-zero exit code (failure)" "exit code 0 (success)"
        return 1
    fi
    return 0
}

# Assert output contains string
assert_output_contains() {
    local expected="$1"
    if [[ "$LAST_OUTPUT" != *"$expected"* ]]; then
        test_fail "output containing '$expected'" "output not containing '$expected'" "$LAST_OUTPUT"
        return 1
    fi
    return 0
}

# Assert output does NOT contain string
assert_output_not_contains() {
    local unexpected="$1"
    if [[ "$LAST_OUTPUT" == *"$unexpected"* ]]; then
        test_fail "output NOT containing '$unexpected'" "output containing '$unexpected'" "$LAST_OUTPUT"
        return 1
    fi
    return 0
}

# Assert output matches regex
assert_output_matches() {
    local pattern="$1"
    if ! echo "$LAST_OUTPUT" | grep -qE "$pattern"; then
        test_fail "output matching pattern '$pattern'" "no match found" "$LAST_OUTPUT"
        return 1
    fi
    return 0
}

# Assert file exists
assert_file_exists() {
    local file="$1"
    if [[ ! -f "$file" ]]; then
        test_fail "file '$file' exists" "file does not exist"
        return 1
    fi
    return 0
}

# Assert directory exists
assert_dir_exists() {
    local dir="$1"
    if [[ ! -d "$dir" ]]; then
        test_fail "directory '$dir' exists" "directory does not exist"
        return 1
    fi
    return 0
}

# Assert file contains string
assert_file_contains() {
    local file="$1"
    local expected="$2"
    if [[ ! -f "$file" ]]; then
        test_fail "file '$file' containing '$expected'" "file does not exist"
        return 1
    fi
    if ! grep -q "$expected" "$file"; then
        test_fail "file containing '$expected'" "string not found in file"
        return 1
    fi
    return 0
}

# Assert exit code
assert_exit_code() {
    local expected="$1"
    if [[ $LAST_EXIT_CODE -ne $expected ]]; then
        test_fail "exit code $expected" "exit code $LAST_EXIT_CODE"
        return 1
    fi
    return 0
}

# Assert output equals exactly
assert_output_equals() {
    local expected="$1"
    if [[ "$LAST_OUTPUT" != "$expected" ]]; then
        test_fail "'$expected'" "'$LAST_OUTPUT'"
        return 1
    fi
    return 0
}

# Get a specific line from last output
get_output_line() {
    local line_num="$1"
    echo "$LAST_OUTPUT" | sed -n "${line_num}p"
}

# Get output as variable (for further processing)
get_output() {
    echo "$LAST_OUTPUT"
}
