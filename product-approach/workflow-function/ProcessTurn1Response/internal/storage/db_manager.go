package storage

import (
	"context"
	//"errors"
	"time"
	 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// DBConfig holds configuration for database operations
type DBConfig struct {
	ConversationTable  string // Table for conversation history
	DefaultTTLDays     int    // Default TTL in days for records
}

// DBManager provides utilities for database operations
type DBManager struct {
	client *dynamodb.Client
	logger logger.Logger
	config DBConfig
}

// NewDBManager creates a new DBManager instance
func NewDBManager(client *dynamodb.Client, log logger.Logger, config DBConfig) *DBManager {
	// Set default TTL if not specified
	if config.DefaultTTLDays == 0 {
		config.DefaultTTLDays = 30
	}

	return &DBManager{
		client: client,
		logger: log.WithFields(map[string]interface{}{
			"component": "db_manager",
		}),
		config: config,
	}
}

// Note: This is a simplified version of DBManager focused only on storing conversation history
// All other methods have been removed as they are unnecessary for the ProcessTurn1Response function

// StoreVerificationRecord has been removed as it's not needed for this function

// GetVerificationRecord has been removed as it's not needed for this function

// UpdateVerificationStatus has been removed as it's not needed for this function

// VerifyLayoutExists has been removed as it's not needed for this function

// GetLayoutMetadata has been removed as it's not needed for this function

// StoreConversationHistory stores conversation history for a verification
func (m *DBManager) StoreConversationHistory(ctx context.Context, verificationId string, conversationState *schema.ConversationState) error {
	// Check if table name is configured
	if m.config.ConversationTable == "" {
		return NewValidationError(
			"Conversation table name not configured",
			map[string]interface{}{"component": "db_manager"},
		)
	}

	m.logger.Debug("Storing conversation history", map[string]interface{}{
		"verificationId": verificationId,
		"currentTurn":    conversationState.CurrentTurn,
		"maxTurns":       conversationState.MaxTurns,
	})

	// Create map for DynamoDB attributes
	item := map[string]interface{}{
		"verificationId":     verificationId,
		"currentTurn":        conversationState.CurrentTurn,
		"maxTurns":           conversationState.MaxTurns,
		"history":            conversationState.History,
		"referenceAnalysis":  conversationState.ReferenceAnalysis,
		"checkingAnalysis":   conversationState.CheckingAnalysis,
		"timestamp":          time.Now().UTC().Format(time.RFC3339),
		"schemaVersion":      schema.SchemaVersion,
		"ttl":                time.Now().AddDate(0, 0, m.config.DefaultTTLDays).Unix(),
	}

	// Convert to DynamoDB AttributeValues
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return NewDynamoDBError(
			"Failed to marshal conversation history",
			"MARSHALING_ERROR",
			false,
			err,
		)
	}

	// Add item to DynamoDB
	_, err = m.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(m.config.ConversationTable),
		Item:      av,
	})

	if err != nil {
		return NewDynamoDBError(
			"Failed to store conversation history",
			"WRITE_ERROR",
			true,
			err,
		)
	}

	m.logger.Info("Stored conversation history", map[string]interface{}{
		"verificationId": verificationId,
		"currentTurn":    conversationState.CurrentTurn,
		"maxTurns":       conversationState.MaxTurns,
	})

	return nil
}

// GetConversationHistory has been removed as it's not needed for this function
