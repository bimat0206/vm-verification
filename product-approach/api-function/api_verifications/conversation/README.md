# API Verifications Conversation Lambda Function

This is a Go-based AWS Lambda function that provides a REST API for retrieving conversation content for specific verification records. It serves as the backend for the `/api/verifications/{verificationId}/conversation` endpoint, enabling users to access processed conversation responses stored in S3.

## Features

### Core Functionality
- **Single Verification Conversation Retrieval**: Get conversation content for a specific verification ID
- **S3 Content Retrieval**: Automatically fetches markdown content from S3 using the stored path
- **Error Handling**: Comprehensive error handling for missing records, invalid S3 paths, and other failures
- **CORS Support**: Full CORS support for cross-origin requests from frontend applications

### API Endpoint
- **GET** `/api/verifications/{verificationId}/conversation` - Retrieve conversation content for a verification

### Response Format
```json
{
  "verificationId": "verification-123",
  "content": "# Conversation Content\n\nThis is the processed conversation response...",
  "contentType": "text/markdown"
}
```

### Error Responses
```json
{
  "error": "Conversation not found",
  "message": "No conversation found for verificationId: verification-123",
  "code": "HTTP_404"
}
```

## Architecture

### AWS Services Integration
- **AWS Lambda**: Serverless compute platform for the API handler
- **Amazon DynamoDB**: NoSQL database for conversation metadata storage
- **Amazon S3**: Object storage for processed conversation content files
- **AWS API Gateway**: HTTP API gateway for routing and request handling

### Data Flow
1. **Request Processing**: Extract verificationId from path parameters
2. **DynamoDB Query**: Query conversation table using verificationId
3. **S3 Retrieval**: Fetch markdown content from S3 using turn2ProcessedPath
4. **Response Formation**: Return structured JSON response with content

## Environment Variables

The function requires the following environment variables:

| Variable | Description | Required |
|----------|-------------|----------|
| `DYNAMODB_CONVERSATION_TABLE` | Name of the DynamoDB conversation table | Yes |
| `RESULTS_BUCKET` | Name of the S3 bucket containing processed responses | Yes |
| `LOG_LEVEL` | Logging level (DEBUG, INFO, WARN, ERROR) | No (default: INFO) |

## DynamoDB Table Structure

The function expects a DynamoDB conversation table with the following structure:

### Primary Key
- **Hash Key**: `verificationId` (String)
- **Range Key**: `conversationId` (String) [Optional, depending on table design]

### Required Attributes
- `verificationId` (String): Unique identifier for the verification
- `turn2ProcessedPath` (String): S3 path to the processed conversation response file
- `conversationId` (String): Unique identifier for the conversation
- `createdAt` (String): ISO 8601 timestamp of record creation
- `updatedAt` (String): ISO 8601 timestamp of last update

### Example Record
```json
{
  "verificationId": "verification-123",
  "conversationId": "conv-456",
  "turn2ProcessedPath": "conversations/verification-123/turn2-processed-response.md",
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:35:00Z"
}
```

## S3 Content Structure

The function expects processed conversation responses to be stored as markdown files in S3:

### File Format
- **Content Type**: Markdown (`.md`)
- **Encoding**: UTF-8
- **Structure**: Standard markdown format

### Example S3 Path
```
s3://results-bucket/conversations/verification-123/turn2-processed-response.md
```

## API Usage Examples

### Successful Request
```bash
curl -X GET "https://api.example.com/api/verifications/verification-123/conversation" \
  -H "Content-Type: application/json"
```

**Response (200 OK):**
```json
{
  "verificationId": "verification-123",
  "content": "# Verification Analysis\n\n## Summary\nThe verification process has been completed...",
  "contentType": "text/markdown"
}
```

### Error Cases

#### Conversation Not Found
```bash
curl -X GET "https://api.example.com/api/verifications/nonexistent-id/conversation"
```

**Response (404 Not Found):**
```json
{
  "error": "Conversation not found",
  "message": "No conversation found for verificationId: nonexistent-id",
  "code": "HTTP_404"
}
```

#### Missing Path Parameter
```bash
curl -X GET "https://api.example.com/api/verifications//conversation"
```

**Response (400 Bad Request):**
```json
{
  "error": "Missing parameter",
  "message": "verificationId path parameter is required",
  "code": "HTTP_400"
}
```

## Development

### Prerequisites
- Go 1.20 or later
- AWS CLI configured with appropriate permissions
- Docker for containerization

### Local Development
```bash
# Set environment variables
export DYNAMODB_CONVERSATION_TABLE=your-conversation-table
export RESULTS_BUCKET=your-results-bucket
export LOG_LEVEL=DEBUG

# Install dependencies
go mod download

# Run tests
go test -v

# Build binary
go build -o api-verifications-conversation *.go

# Run locally (requires AWS credentials)
./api-verifications-conversation
```

### Building and Testing
```bash
# Format code
./deploy.sh go-fmt

# Download dependencies
./deploy.sh go-deps

# Run tests
./deploy.sh go-test

# Build binary
./deploy.sh go-build

# Clean up
./deploy.sh go-clean
```

## Deployment

### Prerequisites
- AWS CLI configured with deployment permissions
- Docker installed and running
- ECR repository created (manually or via Terraform)
- Lambda function created (manually or via Terraform)

### Environment Variables Setup
Before deploying, you must set the required environment variables:

```bash
# Required variables
export ECR_REPO="123456789012.dkr.ecr.us-east-1.amazonaws.com/your-repo-name"
export LAMBDA_FUNCTION_NAME="your-lambda-function-name"

# Optional variables (with defaults)
export AWS_REGION="us-east-1"
export IMAGE_TAG="latest"
```

### Deployment Commands
```bash
# Full deployment (recommended)
./deploy.sh

# Build Docker image only
./deploy.sh build

# Build and push to ECR
./deploy.sh push

# Update Lambda function only
./deploy.sh update

# Test deployed function
./deploy.sh test

# Show help and examples
./deploy.sh help
```

### Example Deployment
```bash
# Set your specific values
export ECR_REPO="879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-api-conversation"
export LAMBDA_FUNCTION_NAME="kootoro-dev-lambda-api-conversation"

# Deploy
./deploy.sh
```

### Deployment Process
1. **Dependency Check**: Verifies AWS CLI, Docker, and Go are available
2. **Environment Validation**: Checks required environment variables are set
3. **ECR Login**: Authenticates with Amazon ECR using AWS CLI
4. **Build and Test**: Compiles Go code and runs tests
5. **Docker Build**: Creates optimized container image
6. **ECR Push**: Uploads image to Amazon ECR
7. **Lambda Update**: Updates function with new image
8. **Function Test**: Validates deployment with test invocation

## IAM Permissions

The Lambda function requires the following IAM permissions:

### DynamoDB Permissions
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:Query",
        "dynamodb:GetItem"
      ],
      "Resource": [
        "arn:aws:dynamodb:region:account:table/conversation-table-name",
        "arn:aws:dynamodb:region:account:table/conversation-table-name/index/*"
      ]
    }
  ]
}
```

### S3 Permissions
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject"
      ],
      "Resource": [
        "arn:aws:s3:::results-bucket-name/*"
      ]
    }
  ]
}
```

## Monitoring and Logging

### CloudWatch Logs
- All requests and responses are logged with structured JSON format
- Error conditions include detailed error messages and stack traces
- Debug logging available for troubleshooting

### Metrics
- Function duration and memory usage via CloudWatch
- Error rates and success rates
- DynamoDB and S3 operation metrics

### Log Levels
- **DEBUG**: Detailed operation logs including DynamoDB queries and S3 operations
- **INFO**: Request/response logging and major operation status
- **WARN**: Non-critical issues and fallback operations
- **ERROR**: Error conditions and failures

## Troubleshooting

### Common Issues

1. **"Conversation not found" errors**: 
   - Verify verificationId exists in conversation table
   - Check DynamoDB table name in environment variables

2. **"Failed to retrieve S3 content" errors**: 
   - Verify S3 bucket name and permissions
   - Check turn2ProcessedPath format in DynamoDB record

3. **"Access denied" errors**: 
   - Check IAM permissions for DynamoDB and S3 access
   - Verify Lambda execution role has required policies

4. **Empty or malformed responses**: 
   - Check S3 file content and encoding
   - Verify markdown file format

### Debug Mode

Enable debug logging by setting `LOG_LEVEL=DEBUG` to see detailed operation logs including:
- DynamoDB query parameters and results
- S3 operation details and content lengths
- Request/response processing steps

## Integration with Frontend

### JavaScript/TypeScript Example
```javascript
async function getConversation(verificationId) {
  try {
    const response = await fetch(`/api/verifications/${verificationId}/conversation`);
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    
    const data = await response.json();
    return data.content; // Markdown content
  } catch (error) {
    console.error('Failed to fetch conversation:', error);
    throw error;
  }
}
```

### React Component Example
```jsx
import { useState, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';

function ConversationViewer({ verificationId }) {
  const [content, setContent] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    async function fetchConversation() {
      try {
        const response = await fetch(`/api/verifications/${verificationId}/conversation`);
        const data = await response.json();
        
        if (response.ok) {
          setContent(data.content);
        } else {
          setError(data.message);
        }
      } catch (err) {
        setError('Failed to load conversation');
      } finally {
        setLoading(false);
      }
    }

    fetchConversation();
  }, [verificationId]);

  if (loading) return <div>Loading conversation...</div>;
  if (error) return <div>Error: {error}</div>;

  return (
    <div className="conversation-content">
      <ReactMarkdown>{content}</ReactMarkdown>
    </div>
  );
}
```
