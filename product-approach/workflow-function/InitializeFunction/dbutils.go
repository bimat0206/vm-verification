package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBUtils provides utilities for DynamoDB operations
type DynamoDBUtils struct {
	client *dynamodb.Client
	logger Logger
	config *ConfigVars
}

// NewDynamoDBUtils creates a new DynamoDBUtils instance
func NewDynamoDBUtils(client *dynamodb.Client, logger Logger) *DynamoDBUtils {
	return &DynamoDBUtils{
		client: client,
		logger: logger,
		config: &ConfigVars{},
	}
}

// SetConfig sets the configuration for DynamoDBUtils
func (d *DynamoDBUtils) SetConfig(config ConfigVars) {
	d.config = &config
}

// VerifyLayoutExists checks if a layout exists in DynamoDB
func (d *DynamoDBUtils) VerifyLayoutExists(ctx context.Context, layoutId int, layoutPrefix string) (bool, error) {
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
	return result.Item != nil && len(result.Item) > 0, nil
}

// StoreVerificationRecord saves the verification record to DynamoDB
func (d *DynamoDBUtils) StoreVerificationRecord(ctx context.Context, verificationContext *VerificationContext) error {
	// Create DynamoDB item
	item := DynamoDBVerificationItem{
		VerificationId:       verificationContext.VerificationId,
		VerificationAt:       verificationContext.VerificationAt,
		Status:               verificationContext.Status,
		VerificationType:     verificationContext.VerificationType,
		VendingMachineId:     verificationContext.VendingMachineId,
		LayoutId:             verificationContext.LayoutId,
		LayoutPrefix:         verificationContext.LayoutPrefix,
		PreviousVerificationId: verificationContext.PreviousVerificationId,
		ReferenceImageUrl:    verificationContext.ReferenceImageUrl,
		CheckingImageUrl:     verificationContext.CheckingImageUrl,
		RequestId:            verificationContext.RequestMetadata.RequestId,
		NotificationEnabled:  verificationContext.NotificationEnabled,
		// Set TTL for 30 days from now
		TTL: time.Now().AddDate(0, 0, 30).Unix(),
	}

	// Convert item to DynamoDB attribute values
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal DynamoDB item: %w", err)
	}

	// Conditional check to ensure idempotency
	conditionExpression := "attribute_not_exists(verificationId)"

	// Put item in DynamoDB
	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(d.config.VerificationTable),
		Item:                av,
		ConditionExpression: aws.String(conditionExpression),
	})

	if err != nil {
		var conditionalCheckFailedErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedErr) {
			// Item already exists, this is a duplicate request
			d.logger.Warn("Duplicate verification ID detected", map[string]interface{}{
				"verificationId": verificationContext.VerificationId,
			})
			// This is not considered an error for idempotency
			return nil
		}
		return fmt.Errorf("failed to put item in DynamoDB: %w", err)
	}

	d.logger.Info("Stored verification record", map[string]interface{}{
		"verificationId": verificationContext.VerificationId,
		"table": d.config.VerificationTable,
	})

	return nil
}

// GetVerificationRecord retrieves a verification record from DynamoDB
func (d *DynamoDBUtils) GetVerificationRecord(ctx context.Context, verificationId string) (*VerificationContext, error) {
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
	
	// Unmarshal the DynamoDB item into a verification context
	var verificationContext VerificationContext
	err = attributevalue.UnmarshalMap(result.Item, &verificationContext)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling verification record: %w", err)
	}
	
	return &verificationContext, nil
}

// UpdateVerificationStatus updates the status of a verification record
func (d *DynamoDBUtils) UpdateVerificationStatus(ctx context.Context, verificationId string, status string) error {
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
		"status": status,
	})
	
	return nil
}

// FindPreviousVerification retrieves the most recent verification that used a specific checking image
func (d *DynamoDBUtils) FindPreviousVerification(ctx context.Context, checkingImageUrl string) (*VerificationContext, error) {
	d.logger.Info("Searching for previous verification", map[string]interface{}{
		"checkingImageUrl": checkingImageUrl,
		"indexName": "CheckingImageIndex",
	})

	// Use GSI4 (CheckingImageIndex) to find verifications where this image was used as checking
	queryInput := &dynamodb.QueryInput{
		TableName: aws.String(d.config.VerificationTable),
		IndexName: aws.String("CheckingImageIndex"), // GSI4 name
		KeyConditionExpression: aws.String("checkingImageUrl = :checkingUrl"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":checkingUrl": &types.AttributeValueMemberS{Value: checkingImageUrl},
		},
		ScanIndexForward: aws.Bool(false), // Return most recent first
		Limit: aws.Int32(1), // We only need the most recent one
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
	var verification VerificationContext
	err = attributevalue.UnmarshalMap(result.Items[0], &verification)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling previous verification: %w", err)
	}
	
	d.logger.Info("Found previous verification", map[string]interface{}{
		"previousVerificationId": verification.VerificationId,
		"verificationAt": verification.VerificationAt,
		"status": verification.Status,
	})
	
	return &verification, nil
}

// BatchGetVerificationRecords retrieves multiple verification records based on IDs
func (d *DynamoDBUtils) BatchGetVerificationRecords(ctx context.Context, verificationIds []string) ([]*VerificationContext, error) {
	if len(verificationIds) == 0 {
		return []*VerificationContext{}, nil
	}

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
	results := make([]*VerificationContext, 0, len(items))

	for _, item := range items {
		var context VerificationContext
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

	return results, nil
}