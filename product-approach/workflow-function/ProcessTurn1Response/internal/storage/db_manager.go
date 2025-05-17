package storage

import (
	"context"
	"fmt"
	"time"

	"workflow-function/shared/logger"
	"workflow-function/shared/schema"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DBConfig holds configuration for database operations
type DBConfig struct {
	VerificationTable  string // Table for verification records
	LayoutTable        string // Table for layout metadata
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

// StoreVerificationRecord saves a verification record to DynamoDB
func (m *DBManager) StoreVerificationRecord(ctx context.Context, verificationContext *schema.VerificationContext) error {
	// Create map for DynamoDB attributes
	item := map[string]interface{}{
		"verificationId":       verificationContext.VerificationId,
		"verificationAt":       verificationContext.VerificationAt,
		"status":               verificationContext.Status,
		"verificationType":     verificationContext.VerificationType,
		"vendingMachineId":     verificationContext.VendingMachineId,
		"referenceImageUrl":    verificationContext.ReferenceImageUrl,
		"checkingImageUrl":     verificationContext.CheckingImageUrl,
		"notificationEnabled":  verificationContext.NotificationEnabled,
		"schemaVersion":        schema.SchemaVersion,
		"ttl":                  time.Now().AddDate(0, 0, m.config.DefaultTTLDays).Unix(),
	}

	// Add optional fields
	if verificationContext.LayoutId > 0 {
		item["layoutId"] = verificationContext.LayoutId
	}
	if verificationContext.LayoutPrefix != "" {
		item["layoutPrefix"] = verificationContext.LayoutPrefix
	}
	if verificationContext.PreviousVerificationId != "" {
		item["previousVerificationId"] = verificationContext.PreviousVerificationId
	}
	if verificationContext.RequestMetadata != nil {
		item["requestId"] = verificationContext.RequestMetadata.RequestId
		item["requestTimestamp"] = verificationContext.RequestMetadata.RequestTimestamp
	}

	// Convert to DynamoDB AttributeValues
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return NewDynamoDBError(
			"Failed to marshal verification record",
			"MARSHALING_ERROR",
			false,
			err,
		)
	}

	// Conditional check to ensure idempotency
	conditionExpression := "attribute_not_exists(verificationId)"

	// Add item to DynamoDB
	_, err = m.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(m.config.VerificationTable),
		Item:                av,
		ConditionExpression: aws.String(conditionExpression),
	})

	if err != nil {
		var conditionalCheckFailedErr *types.ConditionalCheckFailedException
		if As(err, &conditionalCheckFailedErr) {
			// Item already exists - not an error for idempotency
			m.logger.Warn("Duplicate verification ID detected", map[string]interface{}{
				"verificationId": verificationContext.VerificationId,
			})
			return nil
		}
		return NewDynamoDBError(
			"Failed to store verification record",
			"WRITE_ERROR",
			true,
			err,
		)
	}

	m.logger.Info("Stored verification record", map[string]interface{}{
		"verificationId": verificationContext.VerificationId,
		"table":          m.config.VerificationTable,
		"schemaVersion":  schema.SchemaVersion,
	})

	return nil
}

// GetVerificationRecord retrieves a verification record from DynamoDB
func (m *DBManager) GetVerificationRecord(ctx context.Context, verificationId string) (*schema.VerificationContext, error) {
	m.logger.Debug("Getting verification record", map[string]interface{}{
		"verificationId": verificationId,
		"table":         m.config.VerificationTable,
	})

	result, err := m.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(m.config.VerificationTable),
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: verificationId},
		},
	})
	
	if err != nil {
		return nil, NewDynamoDBError(
			fmt.Sprintf("Failed to retrieve verification record: %s", verificationId),
			"READ_ERROR",
			true,
			err,
		)
	}
	
	if result.Item == nil || len(result.Item) == 0 {
		return nil, NewValidationError(
			fmt.Sprintf("Verification record not found: %s", verificationId),
			map[string]interface{}{"verificationId": verificationId},
		)
	}
	
	// Unmarshal the DynamoDB item into a verification context
	var verificationContext schema.VerificationContext
	err = attributevalue.UnmarshalMap(result.Item, &verificationContext)
	if err != nil {
		return nil, NewDynamoDBError(
			"Failed to unmarshal verification record",
			"UNMARSHALING_ERROR",
			false,
			err,
		)
	}
	
	m.logger.Debug("Successfully retrieved verification record", map[string]interface{}{
		"verificationId":   verificationId,
		"verificationType": verificationContext.VerificationType,
		"status":          verificationContext.Status,
	})
	
	return &verificationContext, nil
}

// UpdateVerificationStatus updates the status of a verification record
func (m *DBManager) UpdateVerificationStatus(ctx context.Context, verificationId string, status string) error {
	m.logger.Debug("Updating verification status", map[string]interface{}{
		"verificationId": verificationId,
		"newStatus":     status,
	})

	_, err := m.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(m.config.VerificationTable),
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: verificationId},
		},
		UpdateExpression: aws.String("SET #status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: status},
		},
		ConditionExpression: aws.String("attribute_exists(verificationId)"),
	})
	
	if err != nil {
		var conditionalCheckFailedErr *types.ConditionalCheckFailedException
		if As(err, &conditionalCheckFailedErr) {
			return NewValidationError(
				fmt.Sprintf("Verification record not found: %s", verificationId),
				map[string]interface{}{"verificationId": verificationId},
			)
		}
		return NewDynamoDBError(
			"Failed to update verification status",
			"UPDATE_ERROR",
			true,
			err,
		)
	}

	m.logger.Info("Updated verification status", map[string]interface{}{
		"verificationId": verificationId,
		"status":         status,
	})
	
	return nil
}

// VerifyLayoutExists checks if a layout exists in DynamoDB
func (m *DBManager) VerifyLayoutExists(ctx context.Context, layoutId int, layoutPrefix string) (bool, error) {
	m.logger.Debug("Verifying layout exists", map[string]interface{}{
		"layoutId":     layoutId,
		"layoutPrefix": layoutPrefix,
		"table":        m.config.LayoutTable,
	})

	result, err := m.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(m.config.LayoutTable),
		Key: map[string]types.AttributeValue{
			"layoutId":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", layoutId)},
			"layoutPrefix": &types.AttributeValueMemberS{Value: layoutPrefix},
		},
	})
	
	if err != nil {
		return false, NewDynamoDBError(
			"Failed to check if layout exists",
			"READ_ERROR",
			true,
			err,
		)
	}
	
	// If the item is nil or empty, the layout doesn't exist
	exists := result.Item != nil && len(result.Item) > 0
	
	m.logger.Debug("Layout existence check completed", map[string]interface{}{
		"layoutId":     layoutId,
		"layoutPrefix": layoutPrefix,
		"exists":       exists,
	})
	
	return exists, nil
}

// GetLayoutMetadata retrieves complete layout metadata from DynamoDB
func (m *DBManager) GetLayoutMetadata(ctx context.Context, layoutId int, layoutPrefix string) (*schema.LayoutMetadata, error) {
	m.logger.Info("Getting layout metadata", map[string]interface{}{
		"layoutId":     layoutId,
		"layoutPrefix": layoutPrefix,
		"table":        m.config.LayoutTable,
	})

	result, err := m.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(m.config.LayoutTable),
		Key: map[string]types.AttributeValue{
			"layoutId":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", layoutId)},
			"layoutPrefix": &types.AttributeValueMemberS{Value: layoutPrefix},
		},
	})
	
	if err != nil {
		return nil, NewDynamoDBError(
			"Failed to retrieve layout metadata",
			"READ_ERROR",
			true,
			err,
		)
	}
	
	if result.Item == nil || len(result.Item) == 0 {
		return nil, NewValidationError(
			fmt.Sprintf("Layout metadata not found: layoutId=%d, layoutPrefix=%s", layoutId, layoutPrefix),
			map[string]interface{}{
				"layoutId":     layoutId,
				"layoutPrefix": layoutPrefix,
			},
		)
	}
	
	// Unmarshal the DynamoDB item into a layout metadata struct
	var layout schema.LayoutMetadata
	err = attributevalue.UnmarshalMap(result.Item, &layout)
	if err != nil {
		return nil, NewDynamoDBError(
			"Failed to unmarshal layout metadata",
			"UNMARSHALING_ERROR",
			false,
			err,
		)
	}

	// Validate the unmarshaled data
	if layout.LayoutId == 0 {
		return nil, NewValidationError(
			"Invalid layout metadata: layoutId is 0",
			map[string]interface{}{"layoutId": layout.LayoutId},
		)
	}
	if layout.LayoutPrefix == "" {
		return nil, NewValidationError(
			"Invalid layout metadata: layoutPrefix is empty",
			map[string]interface{}{"layoutPrefix": layout.LayoutPrefix},
		)
	}

	m.logger.Info("Successfully retrieved layout metadata", map[string]interface{}{
		"layoutId":            layout.LayoutId,
		"layoutPrefix":        layout.LayoutPrefix,
		"vendingMachineId":    layout.VendingMachineId,
		"location":            layout.Location,
		"createdAt":           layout.CreatedAt,
		"updatedAt":           layout.UpdatedAt,
		"machineStructureFields": len(layout.MachineStructure),
		"productPositions":    len(layout.ProductPositionMap),
	})
	
	return &layout, nil
}

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

// GetConversationHistory retrieves conversation history for a verification
func (m *DBManager) GetConversationHistory(ctx context.Context, verificationId string) (*schema.ConversationState, error) {
	// Check if table name is configured
	if m.config.ConversationTable == "" {
		return nil, NewValidationError(
			"Conversation table name not configured",
			map[string]interface{}{"component": "db_manager"},
		)
	}

	m.logger.Debug("Getting conversation history", map[string]interface{}{
		"verificationId": verificationId,
		"table":         m.config.ConversationTable,
	})

	result, err := m.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(m.config.ConversationTable),
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: verificationId},
		},
	})
	
	if err != nil {
		return nil, NewDynamoDBError(
			fmt.Sprintf("Failed to retrieve conversation history: %s", verificationId),
			"READ_ERROR",
			true,
			err,
		)
	}
	
	if result.Item == nil || len(result.Item) == 0 {
		return nil, NewValidationError(
			fmt.Sprintf("Conversation history not found: %s", verificationId),
			map[string]interface{}{"verificationId": verificationId},
		)
	}
	
	// Unmarshal the DynamoDB item into a conversation state
	var conversationState schema.ConversationState
	err = attributevalue.UnmarshalMap(result.Item, &conversationState)
	if err != nil {
		return nil, NewDynamoDBError(
			"Failed to unmarshal conversation history",
			"UNMARSHALING_ERROR",
			false,
			err,
		)
	}
	
	m.logger.Debug("Successfully retrieved conversation history", map[string]interface{}{
		"verificationId": verificationId,
		"currentTurn":    conversationState.CurrentTurn,
		"maxTurns":       conversationState.MaxTurns,
	})
	
	return &conversationState, nil
}
