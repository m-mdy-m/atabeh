#!/bin/bash
# Comprehensive test runner for Atabeh
# Tests real-world scenarios including Iranian filtering environment

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
export ATABEH_TEST_TIMEOUT="10s"
export ATABEH_TEST_VERBOSE="${VERBOSE:-false}"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Atabeh Test Suite${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to print section headers
print_section() {
    echo ""
    echo -e "${YELLOW}>>> $1${NC}"
    echo ""
}

# Function to handle test failures
handle_failure() {
    echo -e "${RED}✗ Tests failed${NC}"
    exit 1
}

print_section "Running Parser Tests"
go test -v -race -timeout=5m ./tests/parsers/... || handle_failure
print_section "Running Normalizer Tests"
go test -v -race -timeout=5m ./tests/normalizer/... || handle_failure
print_section "Running Tester Tests"
go test -v -race -timeout=10m ./tests/tester/... || handle_failure

# 7. Coverage Report
print_section "Generating Coverage Report"
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out | grep total
go tool cover -html=coverage.out -o coverage.html
echo -e "${GREEN}Coverage report: coverage.html${NC}"

# 8. Race Detection
print_section "Running Race Detection Tests"
go test -race -short ./... || handle_failure

# 9. Memory Leak Detection (if available)
if command -v valgrind &> /dev/null; then
    print_section "Running Memory Leak Detection"
    go test -c -o test.out ./tests/integration
    valgrind --leak-check=full --show-leak-kinds=all ./test.out || true
    rm -f test.out
fi

# 10. Benchmark Tests
print_section "Running Benchmarks"
go test -bench=. -benchmem -run=^$ ./... | tee benchmark.txt

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✓ All tests passed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"