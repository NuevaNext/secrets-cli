#!/usr/bin/env bash
#
# run-tests.sh - Build and run secret-cli end-to-end tests in a Docker container
#
# Usage:
#   ./tests/run-tests.sh              # Run all tests
#   ./tests/run-tests.sh -v           # Run with verbose output
#   ./tests/run-tests.sh -t <name>    # Run specific test
#   ./tests/run-tests.sh -l           # List available tests
#   ./tests/run-tests.sh --rebuild    # Force rebuild container image
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Configuration
IMAGE_NAME="secret-cli-tests"
CONTAINER_NAME="secret-cli-test-run"

# Colors
if [[ -n "${NO_COLOR:-}" ]]; then
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
else
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
fi

# =============================================================================
# Functions
# =============================================================================

log_step() {
    echo -e "${BLUE}→${NC} $*"
}

log_info() {
    echo -e "${GREEN}✓${NC} $*"
}

log_error() {
    echo -e "${RED}✗${NC} $*" >&2
}

cleanup_container() {
    # Remove container if it exists
    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
    fi
}

build_image() {
    local force_rebuild="${1:-false}"
    
    # Check if image exists and we're not forcing rebuild
    if [[ "$force_rebuild" != "true" ]] && docker image inspect "$IMAGE_NAME" >/dev/null 2>&1; then
        # Check if any source files are newer than the image
        local image_created
        image_created=$(docker image inspect "$IMAGE_NAME" --format '{{.Created}}' 2>/dev/null || echo "1970-01-01")
        
        local needs_rebuild=false
        for file in "${PROJECT_DIR}/secret-cli" "${SCRIPT_DIR}/Dockerfile" "${SCRIPT_DIR}/test-utils.sh" "${SCRIPT_DIR}/e2e-tests.sh"; do
            if [[ -f "$file" ]] && [[ "$file" -nt "${SCRIPT_DIR}/.image-timestamp" ]]; then
                needs_rebuild=true
                break
            fi
        done
        
        if [[ "$needs_rebuild" != "true" ]] && [[ -f "${SCRIPT_DIR}/.image-timestamp" ]]; then
            log_info "Using cached image (use --rebuild to force)"
            return 0
        fi
    fi
    
    log_step "Building test container image..."
    
    docker build \
        -t "$IMAGE_NAME" \
        -f "${SCRIPT_DIR}/Dockerfile" \
        "$PROJECT_DIR"
    
    # Touch timestamp file
    touch "${SCRIPT_DIR}/.image-timestamp"
    
    log_info "Image built successfully"
}

run_tests() {
    local test_args=("$@")
    
    log_step "Running tests in container..."
    echo ""
    
    # Clean up any existing container
    cleanup_container
    
    # Run tests
    local exit_code=0
    docker run \
        --rm \
        --name "$CONTAINER_NAME" \
        "$IMAGE_NAME" \
        "${test_args[@]}" || exit_code=$?
    
    return $exit_code
}

show_help() {
    cat <<EOF
Usage: $(basename "$0") [options]

Run secret-cli end-to-end tests in an isolated Docker container.

Options:
  -v, --verbose      Enable verbose test output
  -t, --test <name>  Run a specific test by name
  -l, --list         List all available tests
  --rebuild          Force rebuild of the test container image
  -h, --help         Show this help message

Examples:
  $(basename "$0")                    # Run all tests
  $(basename "$0") -v                 # Run all tests with verbose output
  $(basename "$0") -t vault_create    # Run only the vault_create test
  $(basename "$0") -l                 # List all tests
  $(basename "$0") --rebuild -v       # Rebuild image and run verbose

The test container is automatically cleaned up after each run.
EOF
}

# =============================================================================
# Main
# =============================================================================

main() {
    local force_rebuild=false
    local test_args=()
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --rebuild)
                force_rebuild=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--verbose|-t|--test|-l|--list)
                test_args+=("$1")
                if [[ "$1" == "-t" || "$1" == "--test" ]]; then
                    test_args+=("$2")
                    shift
                fi
                shift
                ;;
            *)
                echo "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Check for Docker
    if ! command -v docker &>/dev/null; then
        log_error "Docker is required but not installed."
        exit 1
    fi
    
    # Ensure cleanup on exit
    trap cleanup_container EXIT
    
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║       secret-cli End-to-End Test Runner                    ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    # Build image
    build_image "$force_rebuild"
    
    # Run tests
    run_tests "${test_args[@]}"
}

main "$@"
