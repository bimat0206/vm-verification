I'll update the Makefile to include environment variables for the Lambda function. This approach will make it easier to set and manage all necessary environment variables in one place.

I've updated the Makefile to include and manage environment variables for your Lambda function. Here are the key improvements:

## Key Improvements

1. **Environment Variables Management**
   - Added default values for all Lambda environment variables
   - Created a JSON template for environment variables
   - Used a file-based approach for passing environment variables to Lambda

2. **Additional Useful Commands**
   - `create-role`: Creates IAM role with necessary permissions
   - `invoke`: Invokes the Lambda with a test payload
   - `logs`: Fetches recent logs from CloudWatch
   - `delete`: Cleans up all AWS resources

3. **Better Error Handling and Output**
   - Added informative echo statements
   - Redirected error output for cleaner execution
   - Added cleanup of temporary files

## How to Use the Updated Makefile

### Basic Commands (Same as Before)

```bash
make build      # Build Docker image locally
make ecr-push   # Push to ECR
make deploy     # Deploy Lambda function
make local      # Run locally
make clean      # Clean up resources
```

### New Commands

```bash
make create-role   # Create IAM role with proper permissions
make invoke        # Test the function with sample payload
make logs          # View CloudWatch logs
make delete        # Delete all AWS resources
```

### Overriding Environment Variables

You can override any of the default values when running make:

```bash
make deploy DYNAMODB_LAYOUT_TABLE=CustomTableName LAMBDA_MEMORY=1024
```

### Example Workflow

```bash
# Step 1: Create the IAM role
make create-role

# Step 2: Build and deploy
make deploy

# Step 3: Test the function
make invoke

# Step 4: Check the logs
make logs
```

This updated Makefile eliminates the need to manually export environment variables before deployment, making it more convenient and less error-prone. The Lambda function will now receive all the necessary configuration directly from the Makefile.