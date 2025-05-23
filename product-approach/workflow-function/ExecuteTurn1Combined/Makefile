# Makefile for ExecuteTurn1Combined Lambda function tests

# Variables
GOTEST := go test
GORUN := go run
COVERAGE_FILE := coverage.out
REPORT_DIR := test_reports
TEST_RUNNER := run_tests.go

# Default target
.DEFAULT_GOAL := test

# Help target
.PHONY: help
help:
	@echo "ExecuteTurn1Combined Lambda - Test Targets"
	@echo "=========================================="
	@echo "make test          - Run all unit tests"
	@echo "make test-verbose  - Run tests with verbose output"
	@echo "make test-cover    - Run tests with coverage"
	@echo "make test-report   - Run tests and generate HTML coverage report"
	@echo "make test-race     - Run tests with race detector"
	@echo "make test-bench    - Run benchmarks"
	@echo "make test-short    - Run only short tests"
	@echo "make test-handler  - Run only handler tests"
	@echo "make test-happy    - Run only happy path test"
	@echo "make test-all      - Run all tests including integration"
	@echo "make clean         - Clean test artifacts"
	@echo "make lint          - Run linters"
	@echo "make mock-gen      - Regenerate mocks (if using mockgen)"

# Create necessary directories
.PHONY: init
init:
	@mkdir -p $(REPORT_DIR)

# Basic test target
.PHONY: test
test: init
	@echo "Running unit tests..."
	@$(GOTEST) ./internal/handler ./internal/config ./internal/models

# Verbose tests
.PHONY: test-verbose
test-verbose: init
	@echo "Running tests with verbose output..."
	@$(GORUN) $(TEST_RUNNER) -v

# Tests with coverage
.PHONY: test-cover
test-cover: init
	@echo "Running tests with coverage..."
	@$(GORUN) $(TEST_RUNNER) -cover -v

# Tests with HTML coverage report
.PHONY: test-report
test-report: init
	@echo "Running tests and generating coverage report..."
	@$(GORUN) $(TEST_RUNNER) -cover -report -v

# Tests with race detector
.PHONY: test-race
test-race: init
	@echo "Running tests with race detector..."
	@$(GORUN) $(TEST_RUNNER) -race -v

# Run benchmarks
.PHONY: test-bench
test-bench: init
	@echo "Running benchmarks..."
	@$(GORUN) $(TEST_RUNNER) -bench -v

# Run short tests only
.PHONY: test-short
test-short: init
	@echo "Running short tests..."
	@$(GORUN) $(TEST_RUNNER) -short -v

# Run handler tests only
.PHONY: test-handler
test-handler: init
	@echo "Running handler tests..."
	@$(GOTEST) -v ./internal/handler

# Run happy path test specifically
.PHONY: test-happy
test-happy: init
	@echo "Running happy path test..."
	@$(GORUN) $(TEST_RUNNER) -v -specific "TestHandleTurn1Combined_HappyPathLayoutVsChecking"

# Run all tests including integration
.PHONY: test-all
test-all: init
	@echo "Running all tests including integration..."
	@$(GORUN) $(TEST_RUNNER) -integration -cover -v

# Quick test (no coverage, parallel)
.PHONY: test-quick
test-quick: init
	@echo "Running quick tests..."
	@$(GOTEST) -short -parallel 4 ./...

# Test with specific timeout
.PHONY: test-timeout
test-timeout: init
	@echo "Running tests with custom timeout..."
	@$(GORUN) $(TEST_RUNNER) -timeout 5m -v

# Clean test artifacts
.PHONY: clean
clean:
	@echo "Cleaning test artifacts..."
	@rm -f $(COVERAGE_FILE)
	@rm -rf $(REPORT_DIR)
	@rm -f *.test
	@rm -f *.prof
	@echo "✅ Clean complete"

# Lint the code
.PHONY: lint
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Running go vet..."; \
		go vet ./...; \
	fi

# Generate mocks (placeholder - customize based on your mock generation tool)
.PHONY: mock-gen
mock-gen:
	@echo "Mock generation..."
	@echo "Note: Add your mock generation commands here if using mockgen or similar tools"

# Watch for changes and run tests (requires entr or similar tool)
.PHONY: test-watch
test-watch:
	@echo "Watching for changes..."
	@if command -v entr > /dev/null; then \
		find . -name '*.go' | entr -c make test-verbose; \
	else \
		echo "entr not installed. Install with: brew install entr (macOS) or apt-get install entr (Linux)"; \
	fi

# CI target - suitable for CI/CD pipelines
.PHONY: ci
ci: lint test-cover test-race
	@echo "✅ CI checks passed"

# Show test coverage in browser
.PHONY: show-coverage
show-coverage: test-report
	@echo "Opening coverage report in browser..."
	@open $(REPORT_DIR)/coverage.html 2>/dev/null || xdg-open $(REPORT_DIR)/coverage.html 2>/dev/null || echo "Please open $(REPORT_DIR)/coverage.html manually"

# Test specific package with coverage
.PHONY: test-pkg
test-pkg: init
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-pkg PKG=./internal/handler"; \
		exit 1; \
	fi
	@echo "Testing package: $(PKG)"
	@$(GOTEST) -v -cover $(PKG)

# Generate test summary report
.PHONY: test-summary
test-summary: test-cover
	@echo "Generating test summary..."
	@echo "# Test Summary Report" > $(REPORT_DIR)/summary.md
	@echo "Generated: $$(date)" >> $(REPORT_DIR)/summary.md
	@echo "" >> $(REPORT_DIR)/summary.md
	@echo "## Coverage Summary" >> $(REPORT_DIR)/summary.md
	@go tool cover -func=$(COVERAGE_FILE) | tail -n 1 >> $(REPORT_DIR)/summary.md
	@echo "" >> $(REPORT_DIR)/summary.md
	@echo "## Detailed Coverage" >> $(REPORT_DIR)/summary.md
	@go tool cover -func=$(COVERAGE_FILE) >> $(REPORT_DIR)/summary.md
	@echo "✅ Summary generated: $(REPORT_DIR)/summary.md"

# Docker test target (if needed)
.PHONY: test-docker
test-docker:
	@echo "Running tests in Docker..."
	@docker build -f Dockerfile.test -t executeturn1combined-test .
	@docker run --rm executeturn1combined-test

.PHONY: all
all: clean lint test-cover test-report
	@echo "✅ All tasks completed"