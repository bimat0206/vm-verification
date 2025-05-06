# Verification Service

This service provides verification capabilities for vending machine images.

## Prerequisites

- Docker
- Docker Compose
- AWS CLI

## Local Development Setup

1. Clone the repository
2. Navigate to the verification-service directory
3. Run the setup script to start local DynamoDB and S3:

```bash
./setup-local-env.sh
```

This will:
- Start local DynamoDB and S3 using Docker Compose
- Create the necessary DynamoDB tables
- Create the required S3 buckets

## Running the Service

To run the service locally:

```bash
./run-docker.sh
```

This will:
- Check if the local environment is running and start it if needed
- Build the Docker image
- Run the service with the appropriate environment variables

## Testing the API

To test the API endpoints:

```bash
./test-api.sh
```

This will:
- Send a test request to the verification endpoint
- Check the health endpoint

## Troubleshooting

### DynamoDB Connection Issues

If you see errors related to DynamoDB connections:

1. Make sure the local DynamoDB is running:
   ```bash
   docker ps | grep dynamodb-local
   ```

2. Check if the tables exist:
   ```bash
   aws dynamodb list-tables --endpoint-url http://localhost:8000 --region us-east-1
   ```

### S3 Connection Issues

If you see errors related to S3 connections:

1. Make sure the local S3 is running:
   ```bash
   docker ps | grep s3-local
   ```

2. Check if the buckets exist:
   ```bash
   aws s3 ls --endpoint-url http://localhost:4566 --region us-east-1
   ```

## Architecture

The service follows a clean architecture pattern:

- `cmd/server`: Entry point for the application
- `internal/api`: API handlers and middleware
- `internal/app`: Application services
- `internal/domain`: Domain models and services
- `internal/infrastructure`: Infrastructure implementations (DynamoDB, S3, etc.)
- `internal/config`: Configuration management
