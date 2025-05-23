# Template Testing Summary

## Overview

This project now includes comprehensive unit tests for all Turn-1 templates used in the ExecuteTurn1Combined Lambda function. These tests ensure that templates can be executed without runtime errors by validating:

1. All template functions are properly registered in the FuncMap
2. All required context fields are available
3. Templates compile and execute successfully

## Problem Background

The ExecuteTurn1Combined Lambda was experiencing `InternalException` errors during runtime because:
- Go templates referenced functions (like `{{add .RowCount -1}}`) that weren't registered in the FuncMap
- The shared `templateloader` package was being used, but the Lambda wasn't properly configuring it with all required functions
- Errors only surfaced during actual Lambda execution, making debugging difficult

## Test Structure

### Files Created

1. **`internal/template_testutil/loader.go`**
   - Discovers all `.tmpl` files in the templates directory
   - Provides a complete FuncMap with all required template functions
   - Returns template metadata for testing

2. **`internal/template_testutil/context.go`**
   - Defines `MockContext` struct with all fields used by templates
   - Provides factory method `NewMockContext()` with sensible defaults
   - Includes conversion to `map[string]interface{}` for template execution

3. **`templates/templates_test.go`**
   - Main test that executes all discovered templates
   - Captures and reports any template execution failures
   - Generates markdown report when failures occur
   - Uses embed.FS to include templates in test binary

4. **`templates/templates_failure_demo_test.go`** (build tag: `failuredemo`)
   - Demonstrates failure scenarios with missing functions/fields
   - Not part of regular test suite

## Template Functions Included

The test FuncMap includes all functions currently used by templates:

### Math Functions
- `add(a, b int) int` - Addition (e.g., `{{add .RowCount -1}}`)
- `sub(a, b int) int` - Subtraction
- `mul(a, b int) int` - Multiplication
- `div(a, b int) int` - Division

### String Functions
- `split(s, sep string) []string` - Split string by separator
- `join(elems []string, sep string) string` - Join strings with separator
- `upper(s string) string` - Convert to uppercase
- `lower(s string) string` - Convert to lowercase
- `trim(s string) string` - Trim whitespace

### Comparison Functions
- `gt(a, b interface{}) bool` - Greater than
- `lt(a, b interface{}) bool` - Less than
- `eq(a, b interface{}) bool` - Equal
- `ne(a, b interface{}) bool` - Not equal
- `ge(a, b interface{}) bool` - Greater than or equal
- `le(a, b interface{}) bool` - Less than or equal

### Array/Slice Functions
- `index(slice interface{}, idx int) interface{}` - Access array element (e.g., `{{index .RowLabels 0}}`)
- `len(v interface{}) int` - Get length of array/string
- `first(slice interface{}) interface{}` - Get first element
- `last(slice interface{}) interface{}` - Get last element

### Formatting Functions
- `printf(format string, a ...interface{}) string` - Format string (e.g., `{{printf "%.1f" .HoursSinceLastVerification}}`)

### Utility Functions
- `default(def, val interface{}) interface{}` - Return default if value is nil/empty

## Mock Context Structure

The test uses a comprehensive mock context (`MockContext`) that includes all fields referenced by templates:

```go
type MockContext struct {
    // Machine identification
    VendingMachineID string           // "VM-001"
    VendingMachineId string           // "VM-001" (duplicate for compatibility)
    Location         string           // "Building A - Floor 2"
    
    // Machine configuration
    RowCount    int                   // 6
    ColumnCount int                   // 10
    RowLabels   []string              // ["A", "B", "C", "D", "E", "F"]
    
    // Verification history
    PreviousVerificationAt     string // "2024-01-20T10:30:00Z"
    HoursSinceLastVerification float64 // 24.5
    PreviousVerificationStatus string  // "VERIFIED"
    
    // Verification results
    VerificationSummary struct {
        OverallAccuracy       float64  // 95.5
        MissingProducts       int      // 3
        IncorrectProductTypes int      // 2
        EmptyPositionsCount   int      // 5
    }
    
    // System metadata
    SystemPromptVersion string        // "1.0.0"
    VerificationType    string        // "layout-vs-checking"
    SystemPrompt        string        // "You are a vending machine..."
    TemplateVersion     string        // "v1.0"
    CreatedAt           string        // "2024-01-21T11:00:00Z"
    
    // Layout-specific fields
    LayoutId       string             // "layout-123"
    LayoutPrefix   string             // "standard"
    LayoutMetadata interface{}        // map with category, region, etc.
}
```

## Template Usage Examples

### turn1-layout-vs-checking/v1.0.tmpl
Key template expressions used:
- `{{if .VendingMachineID}}{{.VendingMachineID}}{{end}}`
- `{{if .Location}} at {{.Location}}{{end}}`
- `{{.RowCount}} rows ({{.RowLabels}})`
- `{{.ColumnCount}} slots per row`
- `{{index .RowLabels 0}}` - First row label
- `{{index .RowLabels (add .RowCount -1)}}` - Last row label (requires `add` function)

### turn1-previous-vs-current/v1.0.tmpl
Key template expressions used:
- `{{printf "%.1f" .HoursSinceLastVerification}}` - Format hours with 1 decimal
- `{{if .VerificationSummary}}...{{end}}` - Conditional block
- `{{.VerificationSummary.OverallAccuracy}}%`
- All the same row/column expressions as layout template

## Test Results

✅ **All templates pass execution tests**

- `turn1-layout-vs-checking/v1.0.tmpl` - PASS
- `turn1-previous-vs-current/v1.0.tmpl` - PASS

## Running the Tests

```bash
# Run template tests only
go test ./templates -v

# Run all tests (including template tests)
go test ./... -v

# Run with failure demo (not in CI)
go test -tags=failuredemo ./templates -v
```

## CI Integration

The template tests are automatically included in the standard `go test ./...` command, ensuring they run in CI pipelines without additional configuration.

## Error Detection

When templates fail, the test:

1. Identifies the specific template that failed
2. Extracts missing functions or fields using regex patterns
3. Generates a markdown report at `./test_reports/template_failures.md`
4. Fails the test with clear error messages

### Example Error Patterns Detected

- **Missing functions**: `function "add" not defined`
  - Common cause: FuncMap doesn't include required math/string functions
  - Example: `{{add .RowCount -1}}` without `add` in FuncMap

- **Missing map keys**: `map has no entry for key "RowCount"`
  - Common cause: Context object missing required fields
  - Example: `{{.RowCount}}` when RowCount isn't in context

- **Nil pointer access**: `<nil> is not a field of struct type`
  - Common cause: Accessing nested field on nil object
  - Example: `{{.VerificationSummary.OverallAccuracy}}` when VerificationSummary is nil

- **Invalid field access**: `can't evaluate field X in type Y`
  - Common cause: Typo in field name or wrong type
  - Example: `{{.RowCounts}}` instead of `{{.RowCount}}`

### Sample Failure Report

When failures occur, `./test_reports/template_failures.md` contains:

```markdown
# Template Execution Failures Report

Generated: 2024-01-21T15:30:00Z
Total failures: 1

## turn1-layout-vs-checking

### Template: turn1-layout-vs-checking/v1.0.tmpl

**Version:** v1.0

**Error:**
```
template: v1.0.tmpl:9: function "add" not defined
```

**Missing items:**
- function: add

## Summary of Missing Items

### Unique missing items across all templates:
- function: add
```

## Benefits

1. **Early Detection**: Template errors are caught at test time instead of runtime
2. **Clear Diagnostics**: Missing functions/fields are explicitly identified
3. **Version Safety**: New template versions are automatically tested
4. **CI Integration**: Tests run automatically in build pipelines
5. **Documentation**: Failures generate readable markdown reports

## Troubleshooting Guide

### Common Issues and Solutions

1. **Test fails with "function X not defined"**
   - Check if function is in `BuildTestFuncMap()` in `loader.go`
   - Verify function signature matches template usage
   - Add missing function to FuncMap

2. **Test fails with "map has no entry for key X"**
   - Check if field exists in `MockContext` struct
   - Verify field is included in `ToMap()` conversion
   - Add missing field with appropriate test data

3. **Test passes but Lambda still fails**
   - Ensure production code uses same FuncMap as tests
   - Check if `templateloader` is properly configured in production
   - Verify context building in `prompt.go` matches mock context

4. **Templates not discovered by tests**
   - Verify embed directive includes all template patterns
   - Check template file extensions (.tmpl)
   - Ensure templates are in correct directory structure

### Debugging Template Issues

To debug a specific template failure:

1. Run the failure demo test:
   ```bash
   go test -tags=failuredemo ./templates -v
   ```

2. Check the generated report:
   ```bash
   cat ./test_reports/template_failures.md
   ```

3. Examine the specific template:
   ```bash
   cat templates/turn1-layout-vs-checking/v1.0.tmpl | grep -n "add"
   ```

4. Verify function registration in production:
   ```bash
   grep -r "FuncMap" internal/
   ```

## Integration with Production Code

To ensure tests accurately reflect production behavior:

1. **Template Loader Configuration**
   - Production uses `workflow-function/shared/templateloader`
   - Verify `DefaultFunctions` includes all required functions
   - Check `NewPromptService` properly initializes the loader

2. **Context Building**
   - Production context built in `buildTemplateContext()` 
   - Ensure all fields in mock match production context
   - Verify field names and types are consistent

3. **Error Handling**
   - Production errors wrapped with `errors.WrapError`
   - Test error patterns match production error messages
   - Validate error classification in `classifyTemplateError()`

## Future Considerations

- Add performance benchmarks for template execution
- Include token count estimation validation
- Test template output format compliance
- Add integration tests with actual prompt service
- Validate template versioning compatibility
- Add regression tests for historical template versions