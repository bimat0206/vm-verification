package internal

import (
	"context"
	"fmt"

	"workflow-function/shared/errors"
	"workflow-function/shared/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBRepository handles direct DynamoDB operations
type DynamoDBRepository struct {
	client                 *dynamodb.Client
	tableName              string
	conversationTableName  string
	logger                 logger.Logger
}

// NewDynamoDBRepository creates a new DynamoDB repository
func NewDynamoDBRepository(client *dynamodb.Client, tableName string, conversationTableName string, log logger.Logger) *DynamoDBRepository {
	return &DynamoDBRepository{
		client:                client,
		tableName:             tableName,
		conversationTableName: conversationTableName,
		logger: log.WithFields(map[string]interface{}{
			"component": "dynamodb-repository",
		}),
	}
}

// FindPreviousVerification finds the most recent verification record
// where the reference image URL matches the provided S3 path, excluding the current verification
func (repo *DynamoDBRepository) FindPreviousVerification(ctx context.Context, imageURL string, currentVerificationID string, currentVerificationTime string) (*VerificationRecord, error) {
	repo.logger.Info("Finding previous verification", map[string]interface{}{
		"imageURL":                imageURL,
		"currentVerificationID":   currentVerificationID,
		"currentVerificationTime": currentVerificationTime,
	})

	// Create query to find verifications where reference image matches the provided URL
	// Exclude the current verification and only get verifications before the current time
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		IndexName:              aws.String("ReferenceImageIndex"), // Query by reference image
		KeyConditionExpression: aws.String("referenceImageUrl = :img AND verificationAt < :currentTime"),
		FilterExpression:       aws.String("verificationId <> :currentId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":img":         &types.AttributeValueMemberS{Value: imageURL},
			":currentTime": &types.AttributeValueMemberS{Value: currentVerificationTime},
			":currentId":   &types.AttributeValueMemberS{Value: currentVerificationID},
		},
		ScanIndexForward: aws.Bool(false), // Most recent first
		Limit:            aws.Int32(1),
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
			"imageURL":                imageURL,
			"currentVerificationID":   currentVerificationID,
			"currentVerificationTime": currentVerificationTime,
		})
		return nil, errors.NewValidationError("No previous verification found for image",
			map[string]interface{}{
				"imageURL":                imageURL,
				"currentVerificationID":   currentVerificationID,
				"currentVerificationTime": currentVerificationTime,
				"resource":                "Verification",
			})
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
		"verificationId":          record.VerificationID,
		"verificationAt":          record.VerificationAt,
		"currentVerificationID":   currentVerificationID,
		"currentVerificationTime": currentVerificationTime,
		"imageURL":                imageURL,
	})

	return &record, nil
}

// GetTurn2ProcessedPath retrieves the Turn2ProcessedPath from ConversationHistory table
func (repo *DynamoDBRepository) GetTurn2ProcessedPath(ctx context.Context, verificationID string) (string, error) {
	repo.logger.Info("Getting Turn2ProcessedPath from ConversationHistory", map[string]interface{}{
		"verificationId": verificationID,
		"table":          repo.conversationTableName,
	})

	// Query ConversationHistory table to get the Turn2ProcessedPath
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(repo.conversationTableName),
		KeyConditionExpression: aws.String("verificationId = :vid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":vid": &types.AttributeValueMemberS{Value: verificationID},
		},
		ScanIndexForward: aws.Bool(false), // Most recent first
		Limit:            aws.Int32(1),
	}

	result, err := repo.client.Query(ctx, queryInput)
	if err != nil {
		repo.logger.Error("Failed to query ConversationHistory", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": verificationID,
		})
		return "", fmt.Errorf("failed to query ConversationHistory: %w", err)
	}

	if len(result.Items) == 0 {
		repo.logger.Warn("No conversation history found", map[string]interface{}{
			"verificationId": verificationID,
		})
		return "", nil // Return empty string if no conversation history found
	}

	// Unmarshal the conversation tracker
	var conversationTracker struct {
		Turn2ProcessedPath string `dynamodbav:"turn2ProcessedPath"`
	}

	err = attributevalue.UnmarshalMap(result.Items[0], &conversationTracker)
	if err != nil {
		repo.logger.Error("Failed to unmarshal conversation tracker", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": verificationID,
		})
		return "", fmt.Errorf("failed to unmarshal conversation tracker: %w", err)
	}

	repo.logger.Info("Successfully retrieved Turn2ProcessedPath", map[string]interface{}{
		"verificationId":     verificationID,
		"turn2ProcessedPath": conversationTracker.Turn2ProcessedPath,
	})

	return conversationTracker.Turn2ProcessedPath, nil
}
