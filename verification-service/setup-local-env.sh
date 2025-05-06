#!/bin/bash
set -e

echo "Starting local DynamoDB and S3..."
docker-compose up -d

echo "Waiting for services to start..."
sleep 5

echo "Creating DynamoDB tables..."
aws dynamodb create-table \
  --table-name dynamodb-verification-verification-results-dptnik39 \
  --attribute-definitions AttributeName=verificationId,AttributeType=S \
  --key-schema AttributeName=verificationId,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --endpoint-url http://localhost:8000 \
  --region us-east-1 || echo "Table already exists"

aws dynamodb create-table \
  --table-name dynamodb-verification-layout-metadata-dptnik39 \
  --attribute-definitions AttributeName=layoutId,AttributeType=N \
  --key-schema AttributeName=layoutId,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --endpoint-url http://localhost:8000 \
  --region us-east-1 || echo "Table already exists"

echo "Creating S3 buckets..."
aws s3 mb s3://vending-machine-verification-image-reference-a11 \
  --endpoint-url http://localhost:4566 \
  --region us-east-1 || echo "Bucket already exists"

aws s3 mb s3://vending-machine-verification-image-checking-a12 \
  --endpoint-url http://localhost:4566 \
  --region us-east-1 || echo "Bucket already exists"

aws s3 mb s3://s3-verification-results-dptnik39 \
  --endpoint-url http://localhost:4566 \
  --region us-east-1 || echo "Bucket already exists"

echo "Local environment setup complete!"
