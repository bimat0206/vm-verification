#!/bin/bash
# test.sh - Simple test runner script for ExecuteTurn1Combined Lambda

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
VERBOSE=false
COVERAGE=false
SHORT=false
PATTERN=""
TIMEOUT="10m"

# Function to print colored output
print_color() {
    local color=$1
    shift
    echo -e "${color}$@${NC}"
}

# Function to show usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help      Show this help message"
    echo "  -v, --verbose   Run tests with verbose output"
    echo "  -c, --coverage  Run tests with coverage"
    echo "  -s, --short     Run only short tests"
    echo "  -p, --pattern   Run tests matching pattern"
    echo "  -t, --timeout   Set test timeout (default: 10m)"
    echo "  -a, --all       Run all tests including integration"
    echo ""
    echo "Examples:"
    echo "  $0                    # Run basic tests"
    echo "  $0 -v -c              # Run with verbose and coverage"
    echo "  $0 -p TestHappyPath   # Run tests matching pattern"
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -s|--short)
            SHORT=true
            shift
            ;;
        -p|--pattern)
            PATTERN="$2"
            shift 2
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -a|--all)
            ALL_TESTS=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Header
print_color $GREEN "==================================="
print_color $GREEN "ExecuteTurn1Combined Test Runner"
print_color $GREEN "==================================="
echo ""

# Create test reports directory
mkdir -p test_reports

# Build test command
CMD="go test"

if [ "$VERBOSE" = true ]; then
    CMD="$CMD -v"
fi

if [ "$COVERAGE" = true ]; then
    CMD="$CMD -cover -coverprofile=coverage.out -covermode=atomic"
fi

if [ "$SHORT" = true ]; then
    CMD="$CMD -short"
fi

if [ -n "$PATTERN" ]; then
    CMD="$CMD -run $PATTERN"
fi

CMD="$CMD -timeout $TIMEOUT"

# Determine which packages to test
if [ "$ALL_TESTS" = true ]; then
    PACKAGES="./..."
else
    # Test specific packages to avoid integration tests
    PACKAGES="./internal/handler ./internal/config ./internal/models"
fi

# Show configuration
print_color $YELLOW "Configuration:"
echo "  Verbose: $VERBOSE"
echo "  Coverage: $COVERAGE"
echo "  Short: $SHORT"
echo "  Pattern: ${PATTERN:-none}"
echo "  Timeout: $TIMEOUT"
echo "  Packages: $PACKAGES"
echo ""

# Run tests
print_color $YELLOW "Running tests..."
echo "$CMD $PACKAGES"
echo ""

if $CMD $PACKAGES; then
    print_color $GREEN "✅ Tests passed!"
    
    # Generate coverage report if coverage was enabled
    if [ "$COVERAGE" = true ]; then
        echo ""
        print_color $YELLOW "Generating coverage report..."
        go tool cover -html=coverage.out -o=test_reports/coverage.html
        print_color $GREEN "Coverage report generated: test_reports/coverage.html"
        
        echo ""
        print_color $YELLOW "Coverage summary:"
        go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $3}'
    fi
else
    print_color $RED "❌ Tests failed!"
    exit 1
fi

# Generate test summary
if [ "$COVERAGE" = true ]; then
    echo ""
    print_color $YELLOW "Generating test summary..."
    {
        echo "# Test Summary"
        echo "Generated: $(date)"
        echo ""
        echo "## Configuration"
        echo "- Verbose: $VERBOSE"
        echo "- Coverage: $COVERAGE"
        echo "- Short: $SHORT"
        echo "- Pattern: ${PATTERN:-none}"
        echo "- Timeout: $TIMEOUT"
        echo ""
        echo "## Coverage Summary"
        go tool cover -func=coverage.out | tail -1
        echo ""
        echo "## Package Coverage"
        go tool cover -func=coverage.out | grep -E "^github|^workflow" | sort
    } > test_reports/summary.md
    
    print_color $GREEN "Summary saved to: test_reports/summary.md"
fi

echo ""
print_color $GREEN "==================================="
print_color $GREEN "Test execution complete!"
print_color $GREEN "===================================" 