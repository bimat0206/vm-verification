# Basic usage with default table names
python dynamodb-mock-data-generator.py

# Use environment variables for table names
export VERIFICATION_TABLE="Production-VerificationResults"
export CONVERSATION_TABLE="Production-ConversationHistory"
python dynamodb-mock-data-generator.py

# Use command-line arguments for table names
python dynamodb-mock-data-generator.py --verification-table "Dev-VerificationResults" --conversation-table "Dev-ConversationHistory"

# Generate a small dataset for testing
python dynamodb-mock-data-generator.py --records 2 --vendor-machines 1

# Use a specific AWS region
python dynamodb-mock-data-generator.py --region us-west-2

# Use a local DynamoDB instance
python dynamodb-mock-data-generator.py --local

# Use a custom DynamoDB endpoint
python dynamodb-mock-data-generator.py --endpoint-url "http://dynamodb.custom-domain.com:8000"

# Test data generation without inserting into DynamoDB
python dynamodb-mock-data-generator.py --dry-run --verbose

# Generate a large dataset with verbose output
python dynamodb-mock-data-generator.py --records 50 --vendor-machines 10 --verbose