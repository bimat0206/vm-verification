package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBClient wraps the DynamoDB SDK client
type DynamoDBClient struct {
	client *dynamodb.Client
}

// NewDynamoDBClient creates a new DynamoDB client
func NewDynamoDBClient(cfg aws.Config) *DynamoDBClient {
	return &DynamoDBClient{
		client: dynamodb.NewFromConfig(cfg),
	}
}

// QueryMostRecentVerificationByCheckingImage queries the GSI4 index to find verifications
// using the provided referenceImageUrl as the checking image
func (d *DynamoDBClient) QueryMostRecentVerificationByCheckingImage(ctx context.Context, imageURL string) (*VerificationRecord, error) {
	// Define GSI4 query parameters (checkingImageUrl-verificationAt)
	input := &dynamodb.QueryInput{
		TableName:              aws.String(getVerificationTableName()),
		IndexName:              aws.String("CheckImageIndex"), // GSI4
		KeyConditionExpression: aws.String("checkingImageUrl = :url"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":url": &types.AttributeValueMemberS{Value: imageURL},
		},
		ScanIndexForward: aws.Bool(false), // Sort by verificationAt in descending order
		Limit:            aws.Int32(1),    // Get only the most recent verification
	}

	// Execute query
	result, err := d.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query DynamoDB: %w", err)
	}

	// Check if any items were found
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no previous verification found for image: %s", imageURL)
	}

	// Unmarshal the result
	var verification VerificationRecord
	err = attributevalue.UnmarshalMap(result.Items[0], &verification)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal verification record: %w", err)
	}

	return &verification, nil
}