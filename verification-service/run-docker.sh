#!/bin/bash
set -e

# Check if local environment is running
if ! docker ps | grep -q "dynamodb-local"; then
  echo "Local environment not running. Starting it now..."
  ./setup-local-env.sh
fi

# Build the Docker image
docker build -t verification-service .

# Run the container with AWS credentials and endpoint overrides
docker run -p 3000:3000 \
  -e AWS_ACCESS_KEY_ID=test \
  -e AWS_SECRET_ACCESS_KEY=test \
  -e AWS_REGION=us-east-1 \
  -e AWS_ENDPOINT_URL_DYNAMODB=http://host.docker.internal:8000 \
  -e AWS_ENDPOINT_URL_S3=http://host.docker.internal:4566 \
  -e LOG_LEVEL=DEBUG \
  verification-service
