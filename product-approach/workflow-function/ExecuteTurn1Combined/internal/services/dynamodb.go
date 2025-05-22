// internal/services/dynamodb.go
package services

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"ExecuteTurn1Combined/internal/config"
	"ExecuteTurn1Combined/internal/errors"
	"ExecuteTurn1Combined/internal/models"
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
func NewDynamoDBService(cfg config.Config) DynamoDBService {
	awsCfg, err := config.LoadDefaultAWSConfig(context.Background(), config.WithRegion(cfg.AWS.Region))
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
		return errors.WrapRetryable(err, errors.StageDynamoDB, "failed to marshal token usage metrics")
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
		return errors.WrapRetryable(err, errors.StageDynamoDB, "failed to update verification status")
	}
	return nil
}

// RecordConversationTurn inserts a new item into the conversation history table.
func (d *dynamoClient) RecordConversationTurn(ctx context.Context, turn *models.ConversationTurn) error {
	item, err := attributevalue.MarshalMap(turn)
	if err != nil {
		return errors.WrapRetryable(err, errors.StageDynamoDB, "failed to marshal conversation turn")
	}

	input := &dynamodb.PutItemInput{
		TableName: &d.conversationTable,
		Item:      item,
	}

	if _, err := d.client.PutItem(ctx, input); err != nil {
		return errors.WrapRetryable(err, errors.StageDynamoDB, "failed to record conversation turn")
	}
	return nil
}

// awsString is a helper to take a Go string pointer.
func awsString(s string) *string {
	return &s
}
