// internal/services/dynamodb.go - FIXED VERSION
package services

import (
	"context"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	
	// FIXED: Using shared errors package instead of internal
	"workflow-function/shared/errors"
)

// DynamoDBService defines status-tracking and conversation-history operations.
type DynamoDBService interface {
	// UpdateVerificationStatus sets the current status and token usage metrics.
	UpdateVerificationStatus(ctx context.Context, verificationID string, status models.VerificationStatus, metrics models.TokenUsage) error
	// RecordConversationTurn logs the details of a single conversation turn.
	RecordConversationTurn(ctx context.Context, turn *models.ConversationTurn) error
}

type dynamoClient struct {
	client            *dynamodb.Client
	verificationTable string
	conversationTable string
}

// NewDynamoDBService constructs a DynamoDBService using AWS config and table names.
func NewDynamoDBService(cfg *config.Config) DynamoDBService {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(cfg.AWS.Region))
	if err != nil {
		panic("failed to load AWS config for DynamoDB: " + err.Error())
	}
	client := dynamodb.NewFromConfig(awsCfg)
	return &dynamoClient{
		client:            client,
		verificationTable: cfg.AWS.DynamoDBVerificationTable,
		conversationTable: cfg.AWS.DynamoDBConversationTable,
	}
}

// UpdateVerificationStatus sets the verification's currentStatus and tokenUsage.
func (d *dynamoClient) UpdateVerificationStatus(ctx context.Context, verificationID string, status models.VerificationStatus, metrics models.TokenUsage) error {
	// Marshal metrics struct into DynamoDB attribute map
	avMetrics, err := attributevalue.MarshalMap(metrics)
	if err != nil {
		// FIXED: Using shared error WrapError function instead of WrapRetryable
		// WrapError(err, errorType, message, retryable)
		return errors.WrapError(err, errors.ErrorTypeDynamoDB, 
			"failed to marshal token usage metrics", true) // true = retryable
	}

	input := &dynamodb.UpdateItemInput{
		TableName: &d.verificationTable,
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: verificationID},
		},
		UpdateExpression: awsString("SET currentStatus = :status, tokenUsage = :metrics"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":  &types.AttributeValueMemberS{Value: string(status)},
			":metrics": &types.AttributeValueMemberM{Value: avMetrics},
		},
	}

	if _, err := d.client.UpdateItem(ctx, input); err != nil {
		// FIXED: Using ErrorTypeDynamoDB instead of StageDynamoDB
		// The shared package categorizes by error type, not execution stage
		return errors.WrapError(err, errors.ErrorTypeDynamoDB, 
			"failed to update verification status", true).
			// ENHANCED: Adding context for better debugging
			WithContext("table", d.verificationTable).
			WithContext("verificationId", verificationID).
			WithContext("operation", "UpdateItem")
	}
	return nil
}

// RecordConversationTurn inserts a new item into the conversation history table.
func (d *dynamoClient) RecordConversationTurn(ctx context.Context, turn *models.ConversationTurn) error {
	item, err := attributevalue.MarshalMap(turn)
	if err != nil {
		// FIXED: Same pattern - WrapError with ErrorTypeDynamoDB and retryable=true
		return errors.WrapError(err, errors.ErrorTypeDynamoDB, 
			"failed to marshal conversation turn", true).
			// ENHANCED: Adding context for debugging
			WithContext("turn_id", turn.TurnID).
			WithContext("verification_id", turn.VerificationID)
	}

	input := &dynamodb.PutItemInput{
		TableName: &d.conversationTable,
		Item:      item,
	}

	if _, err := d.client.PutItem(ctx, input); err != nil {
		// FIXED: Consistent error handling pattern
		return errors.WrapError(err, errors.ErrorTypeDynamoDB, 
			"failed to record conversation turn", true).
			WithContext("table", d.conversationTable).
			WithContext("turn_id", turn.TurnID).
			WithContext("verification_id", turn.VerificationID).
			WithContext("operation", "PutItem")
	}
	return nil
}

// awsString is a helper to take a Go string pointer.
func awsString(s string) *string {
	return &s
}