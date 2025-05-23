# ExecuteTurn1Combined Lambda - Test Documentation

## Overview

This document describes the test infrastructure for the ExecuteTurn1Combined Lambda function, focusing on the "Happy-path – layout vs checking" scenario.

## Test Structure

### 1. Test Files

- **`internal/handler/handler_happy_test.go`** - Main test file containing the happy path test
- **`internal/handler/mocks/mocks.go`** - Mock implementations for all dependencies
- **`internal/handler/mocks/fixtures.go`** - Test fixtures and helper functions

### 2. Test Entry Points

Multiple ways to run tests are provided:

#### Using Make (Recommended)
```bash
make test          # Run all unit tests
make test-verbose  # Run tests with verbose output
make test-cover    # Run tests with coverage
make test-report   # Run tests and generate HTML coverage report
make test-happy    # Run only the happy path test
make test-all      # Run all tests including integration
```

#### Using Shell Script
```bash
./test.sh                    # Run basic tests
./test.sh -v -c              # Run with verbose and coverage
./test.sh -p TestHappyPath   # Run tests matching pattern
```

#### Using Go Test Runner
```bash
go run run_tests.go          # Run with default settings
go run run_tests.go -v       # Verbose output
go run run_tests.go -cover   # With coverage
go run run_tests.go -report  # Generate HTML report
```

#### Direct Go Test
```bash
go test ./internal/handler -v
go test ./internal/handler -run TestHandleTurn1Combined_HappyPathLayoutVsChecking -v
```

## Test Implementation Details

### Happy Path Test Scenario

The test simulates a successful LAYOUT_VS_CHECKING verification flow:

1. **Input**: StepFunction event with S3 references for:
   - System prompt
   - Image metadata
   - Layout metadata
   - Initialization data
   - Checking and layout images

2. **Processing**: 
   - Loads data from S3 (mocked)
   - Generates prompt using template
   - Invokes Bedrock AI model (mocked)
   - Stores results back to S3
   - Updates DynamoDB status

3. **Assertions**:
   - Response size < 256KB
   - Token usage matches expected (500 input, 42 output)
   - All S3 operations called correctly
   - DynamoDB status updated to TURN1_COMPLETED
   - Logger recorded key events
   - Processing stages completed successfully

### Mock Implementations

All external dependencies are mocked:

- **S3StateManager**: Returns test fixtures for JSON data and base64 images
- **BedrockService**: Returns mock AI response with overallAccuracy
- **DynamoDBService**: Records status updates
- **PromptService**: Generates mock prompts with metrics
- **TemplateLoader**: Returns mock template rendering
- **Logger**: Records all log calls for verification

### Test Fixtures

Located in `internal/handler/mocks/fixtures.go`:

- System prompt JSON
- Image metadata (2 images with dimensions)
- Layout metadata (5x6 planogram)
- Initialization data
- Base64-encoded test images

## Diagnostic Reporting

On test failure, a detailed markdown report is generated at `test_reports/happy_path_failure.md` containing:

- Test execution timestamp
- Failure details
- Expected vs actual values
- Mock call records
- Environment information
- Troubleshooting recommendations

## VS Code Integration

The `.vscode/launch.json` file provides debug configurations:

- Run All Tests
- Run Happy Path Test
- Run Handler Tests
- Debug Happy Path Test
- Run Tests with Race Detector
- Run Benchmarks

## Coverage Reports

Test coverage reports are generated in multiple formats:

- **Terminal output**: Summary statistics
- **HTML report**: `test_reports/coverage.html`
- **Coverage file**: `coverage.out` (for further analysis)

## Continuous Integration

The test setup is CI-friendly:

```bash
make ci  # Run lint, coverage, and race detection tests
```

## Troubleshooting

### Common Issues

1. **Build Failures**: Ensure all dependencies are installed:
   ```bash
   go mod tidy
   ```

2. **Missing Mocks**: Regenerate mocks if interfaces change:
   ```bash
   make mock-gen  # Configure this target for your mock tool
   ```

3. **Template Not Found**: Ensure template directories exist:
   ```bash
   mkdir -p templates/turn1-layout-vs-checking
   mkdir -p templates/turn1-previous-vs-current
   ```

### Verification

Run the setup verification:
```bash
go run verify_test_setup.go
```

This checks:
- Directory structure
- Required files
- Dependencies
- Template directories

## Best Practices

1. **Run tests before committing**: `make test`
2. **Check coverage**: `make test-cover` (aim for >80%)
3. **Use race detector periodically**: `make test-race`
4. **Keep mocks updated**: Update when interfaces change
5. **Add new test scenarios**: Extend the test suite for edge cases

## Future Enhancements

1. Add more test scenarios (error cases, edge cases)
2. Implement integration tests
3. Add performance benchmarks
4. Create property-based tests
5. Add mutation testing

---

For questions or issues, refer to the main project documentation or create an issue in the repository.