package internal

import (
	"context"
	"fmt"
	//"time"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"workflow-function/shared/logger"
)

// DynamoDBRepository handles direct DynamoDB operations
type DynamoDBRepository struct {
	client    *dynamodb.Client
	tableName string
	logger    logger.Logger
}

// NewDynamoDBRepository creates a new DynamoDB repository
func NewDynamoDBRepository(client *dynamodb.Client, tableName string, log logger.Logger) *DynamoDBRepository {
	return &DynamoDBRepository{
		client:    client,
		tableName: tableName,
		logger:    log.WithFields(map[string]interface{}{
			"component": "dynamodb-repository",
		}),
	}
}

// FindPreviousVerification finds the most recent verification using the given image URL as a checking image
func (repo *DynamoDBRepository) FindPreviousVerification(ctx context.Context, imageURL string) (*VerificationRecord, error) {
	repo.logger.Info("Finding previous verification", map[string]interface{}{
		"imageURL": imageURL,
	})

	// Create query to find verifications where checking image matches the reference image
	queryInput := &dynamodb.QueryInput{
		TableName: aws.String(repo.tableName),
		IndexName: aws.String("CheckingImageIndex"), // Assuming this GSI exists
		KeyConditionExpression: aws.String("CheckingImageUrl = :img"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":img": &types.AttributeValueMemberS{Value: imageURL},
		},
		ScanIndexForward: aws.Bool(false), // Most recent first
		Limit: aws.Int32(1),
	}
	
	// Execute query
	result, err := repo.client.Query(ctx, queryInput)
	if err != nil {
		repo.logger.Error("Failed to query DynamoDB", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to query DynamoDB: %w", err)
	}
	
	// Check if any items returned
	if len(result.Items) == 0 {
		repo.logger.Warn("No previous verification found", map[string]interface{}{
			"imageURL": imageURL,
		})
		return nil, NewResourceNotFoundError("Verification", imageURL)
	}
	
	// Unmarshal the first (most recent) item
	var record VerificationRecord
	err = attributevalue.UnmarshalMap(result.Items[0], &record)
	if err != nil {
		repo.logger.Error("Failed to unmarshal DynamoDB item", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to unmarshal DynamoDB item: %w", err)
	}
	
	repo.logger.Info("Found previous verification", map[string]interface{}{
		"verificationId": record.VerificationID,
		"verificationAt": record.VerificationAt,
	})
	
	return &record, nil
}