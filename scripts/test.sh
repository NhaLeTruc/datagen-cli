#!/usr/bin/env bash

# Test script for datagen-cli
# Runs unit tests, integration tests, and generates coverage reports

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Test configuration
COVERAGE_DIR="${PROJECT_ROOT}/coverage"
COVERAGE_FILE="${COVERAGE_DIR}/coverage.txt"
COVERAGE_HTML="${COVERAGE_DIR}/coverage.html"
COVERAGE_THRESHOLD="${COVERAGE_THRESHOLD:-90}"

# Test types
RUN_UNIT=false
RUN_INTEGRATION=false
RUN_E2E=false
RUN_ALL=false
VERBOSE=false
SHORT=false
GENERATE_COVERAGE=false
CHECK_COVERAGE=false

# Help message
usage() {
    cat <<EOF
Usage: $0 [OPTIONS]

Run tests for datagen-cli.

OPTIONS:
    -h, --help              Show this help message
    -a, --all               Run all tests (unit, integration, e2e)
    -u, --unit              Run unit tests only
    -i, --integration       Run integration tests only
    -e, --e2e               Run end-to-end tests only
    -v, --verbose           Verbose output
    -s, --short             Run tests in short mode (skip slow tests)
    -c, --coverage          Generate coverage report
    --check-coverage        Check coverage threshold (${COVERAGE_THRESHOLD}%)
    --threshold NUM         Set coverage threshold (default: 90)
    --clean                 Clean test cache and coverage data

TEST TYPES:
    Unit Tests              Fast, isolated tests (tests/unit/)
    Integration Tests       Tests with file I/O (tests/integration/)
    End-to-End Tests        Full CLI workflows (tests/e2e/)

EXAMPLES:
    # Run all tests
    ./scripts/test.sh --all

    # Run unit tests with coverage
    ./scripts/test.sh --unit --coverage

    # Run integration tests (verbose)
    ./scripts/test.sh --integration --verbose

    # Check coverage meets threshold
    ./scripts/test.sh --all --coverage --check-coverage

    # Quick test run (short mode)
    ./scripts/test.sh --unit --short

EOF
}

# Logger functions
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

section() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# Clean test cache and coverage data
clean() {
    info "Cleaning test cache and coverage data..."
    go clean -testcache
    rm -rf "${COVERAGE_DIR}"
    info "Clean complete"
}

# Setup coverage directory
setup_coverage() {
    mkdir -p "${COVERAGE_DIR}"
}

# Run unit tests
run_unit_tests() {
    section "Running Unit Tests"

    local test_flags="-race"
    local coverage_flags=""

    if [[ "$VERBOSE" == "true" ]]; then
        test_flags="$test_flags -v"
    fi

    if [[ "$SHORT" == "true" ]]; then
        test_flags="$test_flags -short"
    fi

    if [[ "$GENERATE_COVERAGE" == "true" ]]; then
        setup_coverage
        coverage_flags="-coverprofile=${COVERAGE_DIR}/unit.coverprofile -covermode=atomic"
    fi

    info "Running unit tests in tests/unit/..."

    if go test $test_flags $coverage_flags ./tests/unit/...; then
        info "✓ Unit tests passed"
        return 0
    else
        error "✗ Unit tests failed"
        return 1
    fi
}

# Run integration tests
run_integration_tests() {
    section "Running Integration Tests"

    local test_flags="-race"
    local coverage_flags=""

    if [[ "$VERBOSE" == "true" ]]; then
        test_flags="$test_flags -v"
    fi

    if [[ "$SHORT" == "true" ]]; then
        test_flags="$test_flags -short"
    fi

    if [[ "$GENERATE_COVERAGE" == "true" ]]; then
        setup_coverage
        coverage_flags="-coverprofile=${COVERAGE_DIR}/integration.coverprofile -covermode=atomic"
    fi

    info "Running integration tests in tests/integration/..."

    if go test $test_flags $coverage_flags ./tests/integration/...; then
        info "✓ Integration tests passed"
        return 0
    else
        error "✗ Integration tests failed"
        return 1
    fi
}

# Run end-to-end tests
run_e2e_tests() {
    section "Running End-to-End Tests"

    local test_flags="-race"

    if [[ "$VERBOSE" == "true" ]]; then
        test_flags="$test_flags -v"
    fi

    if [[ "$SHORT" == "true" ]]; then
        test_flags="$test_flags -short"
    fi

    info "Running e2e tests in tests/e2e/..."

    # E2E tests typically don't contribute to coverage
    if go test $test_flags ./tests/e2e/...; then
        info "✓ End-to-end tests passed"
        return 0
    else
        error "✗ End-to-end tests failed"
        return 1
    fi
}

# Merge coverage profiles
merge_coverage() {
    if [[ ! -d "${COVERAGE_DIR}" ]]; then
        warn "No coverage data found"
        return 1
    fi

    info "Merging coverage profiles..."

    # Find all coverage profiles
    local profiles=($(find "${COVERAGE_DIR}" -name "*.coverprofile"))

    if [[ ${#profiles[@]} -eq 0 ]]; then
        warn "No coverage profiles found"
        return 1
    fi

    # Merge profiles
    echo "mode: atomic" > "${COVERAGE_FILE}"
    for profile in "${profiles[@]}"; do
        tail -n +2 "$profile" >> "${COVERAGE_FILE}"
    done

    info "✓ Coverage profiles merged to ${COVERAGE_FILE}"
}

# Generate coverage report
generate_coverage_report() {
    if [[ ! -f "${COVERAGE_FILE}" ]]; then
        warn "Coverage file not found: ${COVERAGE_FILE}"
        return 1
    fi

    section "Generating Coverage Report"

    # Generate HTML coverage report
    info "Generating HTML coverage report..."
    go tool cover -html="${COVERAGE_FILE}" -o "${COVERAGE_HTML}"
    info "✓ HTML coverage report: ${COVERAGE_HTML}"

    # Show coverage summary
    info "Coverage summary:"
    go tool cover -func="${COVERAGE_FILE}" | tail -n 1

    # Extract total coverage percentage
    local coverage_pct
    coverage_pct=$(go tool cover -func="${COVERAGE_FILE}" | tail -n 1 | awk '{print $3}' | sed 's/%//')

    echo ""
    info "Total coverage: ${coverage_pct}%"

    # Store coverage percentage for threshold check
    echo "$coverage_pct" > "${COVERAGE_DIR}/coverage.pct"
}

# Check coverage threshold
check_coverage_threshold() {
    if [[ ! -f "${COVERAGE_DIR}/coverage.pct" ]]; then
        error "Coverage data not found. Run with --coverage first."
    fi

    local coverage_pct
    coverage_pct=$(cat "${COVERAGE_DIR}/coverage.pct")

    section "Checking Coverage Threshold"

    info "Coverage: ${coverage_pct}%"
    info "Threshold: ${COVERAGE_THRESHOLD}%"

    # Compare using bc for floating point comparison
    if command -v bc >/dev/null 2>&1; then
        if (( $(echo "$coverage_pct >= $COVERAGE_THRESHOLD" | bc -l) )); then
            info "✓ Coverage meets threshold"
            return 0
        else
            error "✗ Coverage below threshold (${coverage_pct}% < ${COVERAGE_THRESHOLD}%)"
            return 1
        fi
    else
        # Fallback to integer comparison
        local coverage_int=${coverage_pct%.*}
        if [[ $coverage_int -ge $COVERAGE_THRESHOLD ]]; then
            info "✓ Coverage meets threshold"
            return 0
        else
            error "✗ Coverage below threshold (${coverage_pct}% < ${COVERAGE_THRESHOLD}%)"
            return 1
        fi
    fi
}

# Show test summary
show_summary() {
    section "Test Summary"

    local total=0
    local passed=0
    local failed=0

    if [[ "$RUN_UNIT" == "true" ]] || [[ "$RUN_ALL" == "true" ]]; then
        total=$((total + 1))
        echo "  Unit Tests:        ✓ Passed"
        passed=$((passed + 1))
    fi

    if [[ "$RUN_INTEGRATION" == "true" ]] || [[ "$RUN_ALL" == "true" ]]; then
        total=$((total + 1))
        echo "  Integration Tests: ✓ Passed"
        passed=$((passed + 1))
    fi

    if [[ "$RUN_E2E" == "true" ]] || [[ "$RUN_ALL" == "true" ]]; then
        total=$((total + 1))
        echo "  E2E Tests:         ✓ Passed"
        passed=$((passed + 1))
    fi

    echo ""
    info "Results: ${passed}/${total} test suites passed"

    if [[ "$GENERATE_COVERAGE" == "true" ]]; then
        local coverage_pct
        coverage_pct=$(cat "${COVERAGE_DIR}/coverage.pct" 2>/dev/null || echo "N/A")
        info "Coverage: ${coverage_pct}%"
    fi
}

# Main test function
main() {
    local clean_first=false

    # Parse command-line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -a|--all)
                RUN_ALL=true
                shift
                ;;
            -u|--unit)
                RUN_UNIT=true
                shift
                ;;
            -i|--integration)
                RUN_INTEGRATION=true
                shift
                ;;
            -e|--e2e)
                RUN_E2E=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -s|--short)
                SHORT=true
                shift
                ;;
            -c|--coverage)
                GENERATE_COVERAGE=true
                shift
                ;;
            --check-coverage)
                CHECK_COVERAGE=true
                GENERATE_COVERAGE=true
                shift
                ;;
            --threshold)
                COVERAGE_THRESHOLD="$2"
                shift 2
                ;;
            --clean)
                clean_first=true
                shift
                ;;
            *)
                error "Unknown option: $1\nUse --help for usage information"
                ;;
        esac
    done

    # Default to all tests if none specified
    if [[ "$RUN_UNIT" == "false" ]] && [[ "$RUN_INTEGRATION" == "false" ]] && [[ "$RUN_E2E" == "false" ]] && [[ "$RUN_ALL" == "false" ]]; then
        RUN_ALL=true
    fi

    # Navigate to project root
    cd "$PROJECT_ROOT"

    # Clean if requested
    if [[ "$clean_first" == "true" ]]; then
        clean
    fi

    # Print test configuration
    info "Test Configuration:"
    echo "  Unit Tests:        $(if [[ "$RUN_UNIT" == "true" ]] || [[ "$RUN_ALL" == "true" ]]; then echo "enabled"; else echo "disabled"; fi)"
    echo "  Integration Tests: $(if [[ "$RUN_INTEGRATION" == "true" ]] || [[ "$RUN_ALL" == "true" ]]; then echo "enabled"; else echo "disabled"; fi)"
    echo "  E2E Tests:         $(if [[ "$RUN_E2E" == "true" ]] || [[ "$RUN_ALL" == "true" ]]; then echo "enabled"; else echo "disabled"; fi)"
    echo "  Coverage:          $(if [[ "$GENERATE_COVERAGE" == "true" ]]; then echo "enabled"; else echo "disabled"; fi)"
    echo "  Verbose:           $(if [[ "$VERBOSE" == "true" ]]; then echo "yes"; else echo "no"; fi)"
    echo "  Short Mode:        $(if [[ "$SHORT" == "true" ]]; then echo "yes"; else echo "no"; fi)"
    echo ""

    # Run tests
    local test_failed=false

    if [[ "$RUN_UNIT" == "true" ]] || [[ "$RUN_ALL" == "true" ]]; then
        if ! run_unit_tests; then
            test_failed=true
        fi
    fi

    if [[ "$RUN_INTEGRATION" == "true" ]] || [[ "$RUN_ALL" == "true" ]]; then
        if ! run_integration_tests; then
            test_failed=true
        fi
    fi

    if [[ "$RUN_E2E" == "true" ]] || [[ "$RUN_ALL" == "true" ]]; then
        if ! run_e2e_tests; then
            test_failed=true
        fi
    fi

    # Exit early if tests failed
    if [[ "$test_failed" == "true" ]]; then
        error "Tests failed"
        exit 1
    fi

    # Generate coverage report
    if [[ "$GENERATE_COVERAGE" == "true" ]]; then
        merge_coverage
        generate_coverage_report
    fi

    # Check coverage threshold
    if [[ "$CHECK_COVERAGE" == "true" ]]; then
        if ! check_coverage_threshold; then
            exit 1
        fi
    fi

    # Show summary
    show_summary

    echo ""
    info "All tests passed! ✓"
}

# Run main function
main "$@"
