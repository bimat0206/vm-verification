# Migration Guide: Using Shared Packages in Lambda Functions

This guide will help you migrate existing Lambda functions to use the standardized shared packages. Follow these steps to ensure consistent implementation across all workflow functions.

## Step-by-Step Migration Process

### 1. Update Dependencies in go.mod

Modify your function's `go.mod` file to include local imports of shared packages:

```go
module workflow-function/YourFunctionName

go 1.21

require (
    // existing dependencies...
    workflow-function/shared/schema v0.0.0
    workflow-function/shared/logger v0.0.0
    workflow-function/shared/s3utils v0.0.0
    workflow-function/shared/dbutils v0.0.0
)

// Add replace directives
replace workflow-function/shared/schema => ../shared/schema
replace workflow-function/shared/logger => ../shared/logger
replace workflow-function/shared/s3utils => ../shared/s3utils
replace workflow-function/shared/dbutils => ../shared/dbutils
```

### 2. Replace Custom Logger Implementation

Replace your function-specific logger with the shared logger:

```go
// Before:
logger := NewStructuredLogger()

// After:
import "workflow-function/shared/logger"

log := logger.New("verification-service", "YourFunctionName")
```

Update logging calls throughout your code:

```go
// Before:
logger.Info("Processing request", map[string]interface{}{...})

// After:
log.Info("Processing request", map[string]interface{}{...})
```

### 3. Replace S3 Utilities

Replace your function-specific S3 utilities with the shared implementation:

```go
// Before:
s3Utils := NewS3Utils(client, logger)
s3Utils.SetConfig(config)

// After: 
import "workflow-function/shared/s3utils"

s3Client := dependencies.GetS3Client()
s3Utils := s3utils.New(s3Client, log)
```

Update S3 operation calls:

```go
// Before:
err := s3Utils.ValidateImageExists(ctx, s3Url)

// After:
valid, err := s3Utils.ValidateImageExists(ctx, s3Url, 10*1024*1024)
```

### 4. Replace DynamoDB Utilities

Replace your function-specific DynamoDB utilities with the shared implementation:

```go
// Before:
dbUtils := NewDynamoDBUtils(client, logger)
dbUtils.SetConfig(config)

// After:
import "workflow-function/shared/dbutils"

dbClient := dependencies.GetDynamoDBClient()
dbConfig := dbutils.Config{
    VerificationTable: os.Getenv("DYNAMODB_VERIFICATION_TABLE"),
    LayoutTable: os.Getenv("DYNAMODB_LAYOUT_TABLE"),
    ConversationTable: os.Getenv("DYNAMODB_CONVERSATION_TABLE"),
    DefaultTTLDays: 30,
}
dbUtils := dbutils.New(dbClient, log, dbConfig)
```

Update DynamoDB operation calls:

```go
// Before:
err := dbUtils.StoreVerificationRecord(ctx, verificationContext)

// After:
err := dbUtils.StoreVerificationRecord(ctx, verificationContext)
// Interface remains the same, but implementation is standardized
```

### 5. Use Schema Package Types

Replace function-specific types with schema package types:

```go
// Before:
type VerificationContext struct {
    // Fields...
}

// After:
import "workflow-function/shared/schema"

// Use schema.VerificationContext directly
context := &schema.VerificationContext{
    // Fields...
}
```

### 6. Update Dockerfile

Modify your Dockerfile to include shared packages in the build context:

```dockerfile
# Before:
COPY . /app/function-directory/

# After:
COPY workflow-function/shared/ /app/workflow-function/shared/
COPY workflow-function/YourFunctionName/ /app/workflow-function/YourFunctionName/
```

### 7. Update Build Script

Update your `retry-docker-build.sh` script to include shared packages:

```bash
# Create temporary build directory with correct structure
TEMP_DIR=$(mktemp -d)

# Create directories
mkdir -p "$TEMP_DIR/workflow-function/shared/schema"
mkdir -p "$TEMP_DIR/workflow-function/shared/logger" 
mkdir -p "$TEMP_DIR/workflow-function/shared/s3utils"
mkdir -p "$TEMP_DIR/workflow-function/shared/dbutils"
mkdir -p "$TEMP_DIR/workflow-function/YourFunctionName"

# Copy files
cp -r "$PARENT_DIR/shared/schema"/* "$TEMP_DIR/workflow-function/shared/schema/"
cp -r "$PARENT_DIR/shared/logger"/* "$TEMP_DIR/workflow-function/shared/logger/"
cp -r "$PARENT_DIR/shared/s3utils"/* "$TEMP_DIR/workflow-function/shared/s3utils/"
cp -r "$PARENT_DIR/shared/dbutils"/* "$TEMP_DIR/workflow-function/shared/dbutils/"
cp -r "$CURRENT_DIR"/* "$TEMP_DIR/workflow-function/YourFunctionName/"

# Continue with Docker build...
```

### 8. Update Function Documentation

Update your function's README.md and CHANGELOG.md files to document the migration to shared packages:

```markdown
## [1.3.0] - 2025-05-14

### Changed
- Migrated to shared package components
- Updated logger to use standardized shared logger
- Updated S3 and DynamoDB utilities to use shared implementations
- Added schema version handling with shared schema package
```

## Testing Your Migration

After completing the migration:

1. Run local tests to ensure functionality is unchanged
2. Verify Docker build works correctly with new package structure
3. Deploy to development environment and verify integration with Step Functions
4. Check logs to ensure format is consistent with other functions

## Example: InitializeFunction Migration

The InitializeFunction has been successfully migrated to use shared packages. You can reference its implementation as a guide:

- Updated `go.mod` with appropriate dependencies and replace directives
- Migrated from custom logger to shared logger
- Replaced S3 and DynamoDB utilities with shared implementations
- Updated Docker build script to include shared packages
- Updated documentation to reflect changes

## Common Issues and Solutions

### Import Path Issues

If you encounter import path errors, ensure your `go.mod` file has the correct replace directives:

```go
replace workflow-function/shared/schema => ../shared/schema
// Additional replace directives...
```

### Docker Build Context Issues

If Docker build fails with "package not found" errors, check that your build context includes all necessary directories and that paths are correct.

### Interface Mismatches

If function signatures have changed slightly in the shared packages, you may need to update your function calls. Refer to the package documentation for the current interfaces.