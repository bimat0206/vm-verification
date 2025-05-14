// Package dbutils provides utilities for DynamoDB operations
package dbutils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// Config holds configuration for DynamoDB operations
type Config struct {
	VerificationTable  string // Table for verification records
	LayoutTable        string // Table for layout metadata
	ConversationTable  string // Table for conversation history
	DefaultTTLDays     int    // Default TTL in days for records
}

// DynamoDBUtils provides utilities for DynamoDB operations
type DynamoDBUtils struct {
	client *dynamodb.Client
	logger logger.Logger
	config Config
}

// New creates a new DynamoDBUtils instance
func New(client *dynamodb.Client, log logger.Logger, config Config) *DynamoDBUtils {
	// Set default TTL if not specified
	if config.DefaultTTLDays == 0 {
		config.DefaultTTLDays = 30
	}

	return &DynamoDBUtils{
		client: client,
		logger: log.WithFields(map[string]interface{}{
			"component": "dbutils",
		}),
		config: config,
	}
}

// ==========================================
// Verification Record Operations
// ==========================================

// StoreVerificationRecord saves a verification record to DynamoDB
func (d *DynamoDBUtils) StoreVerificationRecord(ctx context.Context, verificationContext *schema.VerificationContext) error {
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
		"ttl":                  time.Now().AddDate(0, 0, d.config.DefaultTTLDays).Unix(),
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
		return fmt.Errorf("failed to marshal DynamoDB item: %w", err)
	}

	// Conditional check to ensure idempotency
	conditionExpression := "attribute_not_exists(verificationId)"

	// Add item to DynamoDB
	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(d.config.VerificationTable),
		Item:                av,
		ConditionExpression: aws.String(conditionExpression),
	})

	if err != nil {
		var conditionalCheckFailedErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedErr) {
			// Item already exists - not an error for idempotency
			d.logger.Warn("Duplicate verification ID detected", map[string]interface{}{
				"verificationId": verificationContext.VerificationId,
			})
			return nil
		}
		return fmt.Errorf("failed to put item in DynamoDB: %w", err)
	}

	d.logger.Info("Stored verification record", map[string]interface{}{
		"verificationId": verificationContext.VerificationId,
		"table":          d.config.VerificationTable,
		"schemaVersion":  schema.SchemaVersion,
	})

	return nil
}

// GetVerificationRecord retrieves a verification record from DynamoDB
func (d *DynamoDBUtils) GetVerificationRecord(ctx context.Context, verificationId string) (*schema.VerificationContext, error) {
	d.logger.Debug("Getting verification record", map[string]interface{}{
		"verificationId": verificationId,
		"table":         d.config.VerificationTable,
	})

	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.config.VerificationTable),
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: verificationId},
		},
	})
	
	if err != nil {
		return nil, fmt.Errorf("error retrieving verification record: %w", err)
	}
	
	if result.Item == nil || len(result.Item) == 0 {
		return nil, fmt.Errorf("verification record not found: %s", verificationId)
	}
	
	// Check for schema version to determine the format (for future compatibility)
	// For now, we just use the current schema version
	
	// Unmarshal the DynamoDB item into a verification context
	var verificationContext schema.VerificationContext
	err = attributevalue.UnmarshalMap(result.Item, &verificationContext)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling verification record: %w", err)
	}
	
	d.logger.Debug("Successfully retrieved verification record", map[string]interface{}{
		"verificationId":   verificationId,
		"verificationType": verificationContext.VerificationType,
		"status":          verificationContext.Status,
	})
	
	return &verificationContext, nil
}

// UpdateVerificationStatus updates the status of a verification record
func (d *DynamoDBUtils) UpdateVerificationStatus(ctx context.Context, verificationId string, status string) error {
	d.logger.Debug("Updating verification status", map[string]interface{}{
		"verificationId": verificationId,
		"newStatus":     status,
	})

	_, err := d.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(d.config.VerificationTable),
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
		if errors.As(err, &conditionalCheckFailedErr) {
			return fmt.Errorf("verification record not found: %s", verificationId)
		}
		return fmt.Errorf("error updating verification status: %w", err)
	}

	d.logger.Info("Updated verification status", map[string]interface{}{
		"verificationId": verificationId,
		"status":         status,
	})
	
	return nil
}

// FindPreviousVerification retrieves the most recent verification that used a specific checking image
func (d *DynamoDBUtils) FindPreviousVerification(ctx context.Context, checkingImageUrl string) (*schema.VerificationContext, error) {
	d.logger.Info("Searching for previous verification", map[string]interface{}{
		"checkingImageUrl": checkingImageUrl,
		"indexName":        "CheckingImageIndex",
	})

	// Use secondary index to find verifications where this image was used as checking
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(d.config.VerificationTable),
		IndexName:              aws.String("CheckingImageIndex"), // Secondary index name
		KeyConditionExpression: aws.String("checkingImageUrl = :checkingUrl"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":checkingUrl": &types.AttributeValueMemberS{Value: checkingImageUrl},
		},
		ScanIndexForward: aws.Bool(false), // Return most recent first
		Limit:            aws.Int32(1),    // We only need the most recent one
	}
	
	result, err := d.client.Query(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("error querying previous verification: %w", err)
	}
	
	if len(result.Items) == 0 {
		d.logger.Info("No previous verification found", map[string]interface{}{
			"checkingImageUrl": checkingImageUrl,
		})
		return nil, nil // No previous verification found
	}
	
	// Unmarshal into VerificationContext
	var verification schema.VerificationContext
	err = attributevalue.UnmarshalMap(result.Items[0], &verification)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling previous verification: %w", err)
	}
	
	d.logger.Info("Found previous verification", map[string]interface{}{
		"previousVerificationId": verification.VerificationId,
		"verificationAt":         verification.VerificationAt,
		"status":                 verification.Status,
	})
	
	return &verification, nil
}

// ==========================================
// Layout Metadata Operations
// ==========================================

// VerifyLayoutExists checks if a layout exists in DynamoDB
func (d *DynamoDBUtils) VerifyLayoutExists(ctx context.Context, layoutId int, layoutPrefix string) (bool, error) {
	d.logger.Debug("Verifying layout exists", map[string]interface{}{
		"layoutId":     layoutId,
		"layoutPrefix": layoutPrefix,
		"table":        d.config.LayoutTable,
	})

	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.config.LayoutTable),
		Key: map[string]types.AttributeValue{
			"layoutId":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", layoutId)},
			"layoutPrefix": &types.AttributeValueMemberS{Value: layoutPrefix},
		},
	})
	
	if err != nil {
		return false, fmt.Errorf("error querying DynamoDB for layout: %w", err)
	}
	
	// If the item is nil or empty, the layout doesn't exist
	exists := result.Item != nil && len(result.Item) > 0
	
	d.logger.Debug("Layout existence check completed", map[string]interface{}{
		"layoutId":     layoutId,
		"layoutPrefix": layoutPrefix,
		"exists":       exists,
	})
	
	return exists, nil
}

// GetLayoutMetadata retrieves complete layout metadata from DynamoDB
func (d *DynamoDBUtils) GetLayoutMetadata(ctx context.Context, layoutId int, layoutPrefix string) (*schema.LayoutMetadata, error) {
	d.logger.Info("Getting layout metadata", map[string]interface{}{
		"layoutId":     layoutId,
		"layoutPrefix": layoutPrefix,
		"table":        d.config.LayoutTable,
	})

	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.config.LayoutTable),
		Key: map[string]types.AttributeValue{
			"layoutId":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", layoutId)},
			"layoutPrefix": &types.AttributeValueMemberS{Value: layoutPrefix},
		},
	})
	
	if err != nil {
		return nil, fmt.Errorf("error retrieving layout metadata: %w", err)
	}
	
	if result.Item == nil || len(result.Item) == 0 {
		return nil, fmt.Errorf("layout metadata not found: layoutId=%d, layoutPrefix=%s", layoutId, layoutPrefix)
	}
	
	// Unmarshal the DynamoDB item into a layout metadata struct
	var layout schema.LayoutMetadata
	err = attributevalue.UnmarshalMap(result.Item, &layout)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling layout metadata: %w", err)
	}

	// Validate the unmarshaled data
	if layout.LayoutId == 0 {
		return nil, fmt.Errorf("invalid layout metadata: layoutId is 0")
	}
	if layout.LayoutPrefix == "" {
		return nil, fmt.Errorf("invalid layout metadata: layoutPrefix is empty")
	}

	d.logger.Info("Successfully retrieved layout metadata", map[string]interface{}{
		"layoutId":        layout.LayoutId,
		"layoutPrefix":    layout.LayoutPrefix,
		"vendingMachineId": layout.VendingMachineId,
		"location":        layout.Location,
		"createdAt":       layout.CreatedAt,
		"updatedAt":       layout.UpdatedAt,
		"machineStructureFields": len(layout.MachineStructure),
		"productPositions": len(layout.ProductPositionMap),
	})
	
	return &layout, nil
}

// StoreLayoutMetadata stores layout metadata in DynamoDB
func (d *DynamoDBUtils) StoreLayoutMetadata(ctx context.Context, layout *schema.LayoutMetadata) error {
	// Validate required fields
	if layout.LayoutId == 0 {
		return fmt.Errorf("layoutId is required")
	}
	if layout.LayoutPrefix == "" {
		return fmt.Errorf("layoutPrefix is required")
	}

	// Set timestamps if not provided
	now := time.Now().UTC().Format(time.RFC3339)
	if layout.CreatedAt == "" {
		layout.CreatedAt = now
	}
	layout.UpdatedAt = now

	// Convert to DynamoDB AttributeValues
	av, err := attributevalue.MarshalMap(layout)
	if err != nil {
		return fmt.Errorf("failed to marshal layout metadata: %w", err)
	}

	// Store in DynamoDB
	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.config.LayoutTable),
		Item:      av,
	})

	if err != nil {
		return fmt.Errorf("failed to store layout metadata: %w", err)
	}

	d.logger.Info("Stored layout metadata", map[string]interface{}{
		"layoutId":        layout.LayoutId,
		"layoutPrefix":    layout.LayoutPrefix,
		"vendingMachineId": layout.VendingMachineId,
		"table":           d.config.LayoutTable,
	})

	return nil
}

// ==========================================
// Batch Operations
// ==========================================

// BatchGetVerificationRecords retrieves multiple verification records based on IDs
func (d *DynamoDBUtils) BatchGetVerificationRecords(ctx context.Context, verificationIds []string) ([]*schema.VerificationContext, error) {
	if len(verificationIds) == 0 {
		return []*schema.VerificationContext{}, nil
	}

	d.logger.Debug("Batch getting verification records", map[string]interface{}{
		"count": len(verificationIds),
		"table": d.config.VerificationTable,
	})

	// Convert verification IDs to DynamoDB keys
	keys := make([]map[string]types.AttributeValue, len(verificationIds))
	for i, id := range verificationIds {
		keys[i] = map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: id},
		}
	}

	// Batch get items from DynamoDB
	response, err := d.client.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			d.config.VerificationTable: {
				Keys: keys,
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error batch getting verification records: %w", err)
	}

	// Process the results
	items := response.Responses[d.config.VerificationTable]
	results := make([]*schema.VerificationContext, 0, len(items))

	for _, item := range items {
		var context schema.VerificationContext
		if err := attributevalue.UnmarshalMap(item, &context); err != nil {
			d.logger.Warn("Failed to unmarshal verification record", map[string]interface{}{
				"error": err.Error(),
			})
			continue
		}
		results = append(results, &context)
	}

	// Handle unprocessed keys if any
	if len(response.UnprocessedKeys) > 0 {
		d.logger.Warn("Some verification records couldn't be retrieved", map[string]interface{}{
			"unprocessedCount": len(response.UnprocessedKeys[d.config.VerificationTable].Keys),
		})
	}

	d.logger.Debug("Batch get completed", map[string]interface{}{
		"requested": len(verificationIds),
		"retrieved": len(results),
	})

	return results, nil
}

// BatchGetLayoutMetadata retrieves multiple layout metadata records
func (d *DynamoDBUtils) BatchGetLayoutMetadata(ctx context.Context, layoutKeys []schema.LayoutKey) ([]*schema.LayoutMetadata, error) {
	if len(layoutKeys) == 0 {
		return []*schema.LayoutMetadata{}, nil
	}

	d.logger.Debug("Batch getting layout metadata", map[string]interface{}{
		"count": len(layoutKeys),
		"table": d.config.LayoutTable,
	})

	// Convert layout keys to DynamoDB keys
	keys := make([]map[string]types.AttributeValue, len(layoutKeys))
	for i, key := range layoutKeys {
		keys[i] = map[string]types.AttributeValue{
			"layoutId":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", key.LayoutId)},
			"layoutPrefix": &types.AttributeValueMemberS{Value: key.LayoutPrefix},
		}
	}

	// Batch get items from DynamoDB
	response, err := d.client.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			d.config.LayoutTable: {
				Keys: keys,
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error batch getting layout metadata: %w", err)
	}

	// Process the results
	items := response.Responses[d.config.LayoutTable]
	results := make([]*schema.LayoutMetadata, 0, len(items))

	for _, item := range items {
		var layout schema.LayoutMetadata
		if err := attributevalue.UnmarshalMap(item, &layout); err != nil {
			d.logger.Warn("Failed to unmarshal layout metadata", map[string]interface{}{
				"error": err.Error(),
			})
			continue
		}
		results = append(results, &layout)
	}

	// Handle unprocessed keys if any
	if len(response.UnprocessedKeys) > 0 {
		d.logger.Warn("Some layout metadata records couldn't be retrieved", map[string]interface{}{
			"unprocessedCount": len(response.UnprocessedKeys[d.config.LayoutTable].Keys),
		})
	}

	d.logger.Debug("Batch get layout metadata completed", map[string]interface{}{
		"requested": len(layoutKeys),
		"retrieved": len(results),
	})

	return results, nil
}

// ==========================================
// Conversation History Operations
// ==========================================

// StoreConversationHistory stores conversation history for a verification
func (d *DynamoDBUtils) StoreConversationHistory(ctx context.Context, verificationId string, conversationState *schema.ConversationState) error {
	// Check if table name is configured
	if d.config.ConversationTable == "" {
		return errors.New("conversation table name not configured")
	}

	d.logger.Debug("Storing conversation history", map[string]interface{}{
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
		"ttl":                time.Now().AddDate(0, 0, d.config.DefaultTTLDays).Unix(),
	}

	// Convert to DynamoDB AttributeValues
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal conversation history: %w", err)
	}

	// Add item to DynamoDB
	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.config.ConversationTable),
		Item:      av,
	})

	if err != nil {
		return fmt.Errorf("failed to store conversation history: %w", err)
	}

	d.logger.Info("Stored conversation history", map[string]interface{}{
		"verificationId": verificationId,
		"currentTurn":    conversationState.CurrentTurn,
		"maxTurns":       conversationState.MaxTurns,
	})

	return nil
}

// GetConversationHistory retrieves conversation history for a verification
func (d *DynamoDBUtils) GetConversationHistory(ctx context.Context, verificationId string) (*schema.ConversationState, error) {
	// Check if table name is configured
	if d.config.ConversationTable == "" {
		return nil, errors.New("conversation table name not configured")
	}

	d.logger.Debug("Getting conversation history", map[string]interface{}{
		"verificationId": verificationId,
		"table":         d.config.ConversationTable,
	})

	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.config.ConversationTable),
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: verificationId},
		},
	})
	
	if err != nil {
		return nil, fmt.Errorf("error retrieving conversation history: %w", err)
	}
	
	if result.Item == nil || len(result.Item) == 0 {
		return nil, fmt.Errorf("conversation history not found: %s", verificationId)
	}
	
	// Unmarshal the DynamoDB item into a conversation state
	var conversationState schema.ConversationState
	err = attributevalue.UnmarshalMap(result.Item, &conversationState)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling conversation history: %w", err)
	}
	
	d.logger.Debug("Successfully retrieved conversation history", map[string]interface{}{
		"verificationId": verificationId,
		"currentTurn":    conversationState.CurrentTurn,
		"maxTurns":       conversationState.MaxTurns,
	})
	
	return &conversationState, nil
}