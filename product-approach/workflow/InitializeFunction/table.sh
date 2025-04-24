#!/bin/bash

# AWS CLI command to create VerificationResults table
aws dynamodb create-table \
  --table-name VerificationResults \
  --attribute-definitions \
    AttributeName=verificationId,AttributeType=S \
    AttributeName=verificationAt,AttributeType=S \
    AttributeName=layoutId,AttributeType=N \
    AttributeName=verificationStatus,AttributeType=S \
    AttributeName=comparisonId,AttributeType=S \
  --key-schema \
    AttributeName=verificationId,KeyType=HASH \
    AttributeName=verificationAt,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST \
  --global-secondary-indexes \
    "[
      {
        \"IndexName\": \"GSI1\",
        \"KeySchema\": [
          {\"AttributeName\": \"layoutId\", \"KeyType\": \"HASH\"},
          {\"AttributeName\": \"verificationAt\", \"KeyType\": \"RANGE\"}
        ],
        \"Projection\": {
          \"ProjectionType\": \"ALL\"
        }
      },
      {
        \"IndexName\": \"GSI2\",
        \"KeySchema\": [
          {\"AttributeName\": \"verificationStatus\", \"KeyType\": \"HASH\"},
          {\"AttributeName\": \"verificationAt\", \"KeyType\": \"RANGE\"}
        ],
        \"Projection\": {
          \"ProjectionType\": \"INCLUDE\",
          \"NonKeyAttributes\": [
            \"vendingMachineId\", 
            \"location\", 
            \"verificationSummary\"
          ]
        }
      },
      {
        \"IndexName\": \"ComparisonIdIndex\",
        \"KeySchema\": [
          {\"AttributeName\": \"comparisonId\", \"KeyType\": \"HASH\"}
        ],
        \"Projection\": {
          \"ProjectionType\": \"ALL\"
        }
      }
    ]" \
  --sse-specification Enabled=true,SSEType=KMS \
  --tags Key=Environment,Value=Production Key=Project,Value=VendingVerification \
  --region us-east-1

# Wait for table to be created and active
echo "Waiting for VerificationResults table to become active..."
aws dynamodb wait table-exists --table-name VerificationResults --region us-east-1

# Enable Time-To-Live (TTL) for automatic item expiration
aws dynamodb update-time-to-live \
  --table-name VerificationResults \
  --time-to-live-specification "Enabled=true, AttributeName=TTL" \
  --region us-east-1

echo "VerificationResults table created successfully!"