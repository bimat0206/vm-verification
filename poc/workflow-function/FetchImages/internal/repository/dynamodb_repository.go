// Package repository provides data access implementations for the FetchImages function
package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"workflow-function/shared/logger"
	"workflow-function/shared/schema"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBRepository handles DynamoDB operations
type DynamoDBRepository struct {
	client           *dynamodb.Client
	layoutTable      string
	verificationTable string
	logger           logger.Logger
}

// NewDynamoDBRepository creates a new DynamoDBRepository instance
func NewDynamoDBRepository(
	client *dynamodb.Client,
	layoutTable string,
	verificationTable string,
	log logger.Logger,
) *DynamoDBRepository {
	return &DynamoDBRepository{
		client:           client,
		layoutTable:      layoutTable,
		verificationTable: verificationTable,
		logger:           log.WithFields(map[string]interface{}{"component": "DynamoDBRepository"}),
	}
}

// GetTableName returns the verification table name used for debugging
func (r *DynamoDBRepository) GetTableName() string {
	return r.verificationTable
}

// ValidateLayoutExists checks if a layout exists before attempting to fetch it
func (r *DynamoDBRepository) ValidateLayoutExists(
	ctx context.Context, 
	layoutId int, 
	layoutPrefix string,
) (bool, error) {
	r.logger.Info("Validating layout existence", map[string]interface{}{
		"layoutId":     layoutId,
		"layoutPrefix": layoutPrefix,
		"table":        r.layoutTable,
	})

	// Create the key for the DynamoDB query
	key := map[string]types.AttributeValue{
		"layoutId": &types.AttributeValueMemberN{
			Value: strconv.Itoa(layoutId),
		},
		"layoutPrefix": &types.AttributeValueMemberS{
			Value: layoutPrefix,
		},
	}
	
	// Get the item from DynamoDB
	getItemOutput, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.layoutTable),
		Key:       key,
		ProjectionExpression: aws.String("layoutId"), // Only retrieve the key to minimize data transfer
	})
	
	if err != nil {
		return false, fmt.Errorf("failed to check if layout exists: %w", err)
	}
	
	exists := getItemOutput.Item != nil
	
	r.logger.Info("Layout existence validation result", map[string]interface{}{
		"layoutId":     layoutId,
		"layoutPrefix": layoutPrefix,
		"exists":       exists,
	})
	
	// Check if the item exists
	return exists, nil
}

// FetchLayoutMetadata retrieves layout metadata from DynamoDB
func (r *DynamoDBRepository) FetchLayoutMetadata(
	ctx context.Context, 
	layoutId int, 
	layoutPrefix string,
) (*schema.LayoutMetadata, error) {
	// Create the key for the DynamoDB query
	key := map[string]types.AttributeValue{
		"layoutId": &types.AttributeValueMemberN{
			Value: strconv.Itoa(layoutId),
		},
		"layoutPrefix": &types.AttributeValueMemberS{
			Value: layoutPrefix,
		},
	}
	
	r.logger.Info("Fetching layout metadata", map[string]interface{}{
		"layoutId":     layoutId,
		"layoutPrefix": layoutPrefix,
		"table":        r.layoutTable,
	})
	
	// Get the item from DynamoDB
	getItemOutput, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.layoutTable),
		Key:       key,
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve layout metadata from DynamoDB: %w", err)
	}
	
	// Check if the item exists
	if getItemOutput.Item == nil {
		return nil, fmt.Errorf("layout not found: layoutId=%d, layoutPrefix=%s", layoutId, layoutPrefix)
	}
	
	// Unmarshal the DynamoDB item into a LayoutMetadata struct
	var layout schema.LayoutMetadata
	err = attributevalue.UnmarshalMap(getItemOutput.Item, &layout)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal layout metadata: %w", err)
	}
	
	r.logger.Info("Successfully fetched layout metadata", map[string]interface{}{
		"layoutId":          layout.LayoutId,
		"layoutPrefix":      layout.LayoutPrefix,
		"vendingMachineId":  layout.VendingMachineId,
		"hasStructure":      layout.MachineStructure != nil,
		"hasPositions":      layout.ProductPositionMap != nil,
	})
	
	return &layout, nil
}

// FetchHistoricalVerification retrieves historical verification data from DynamoDB
func (r *DynamoDBRepository) FetchHistoricalVerification(
	ctx context.Context, 
	verificationId string,
) (map[string]interface{}, error) {
	r.logger.Info("Fetching historical verification", map[string]interface{}{
		"verificationId": verificationId,
		"table":          r.verificationTable,
	})
	
	// Use a Query operation instead of GetItem to handle composite key tables
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(r.verificationTable),
		KeyConditionExpression: aws.String("verificationId = :vid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":vid": &types.AttributeValueMemberS{Value: verificationId},
		},
		Limit: aws.Int32(1), // We only need the most recent record
	}
	
	// Execute the query
	queryOutput, err := r.client.Query(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical verification: %w", err)
	}
	
	// Check if any items were returned
	if len(queryOutput.Items) == 0 {
		return nil, fmt.Errorf("verification not found: verificationId=%s", verificationId)
	}
	
	// Create a map to unmarshal the first (and should be only) DynamoDB item
	var result map[string]interface{}
	err = attributevalue.UnmarshalMap(queryOutput.Items[0], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal verification record: %w", err)
	}
	
	// Add calculated fields
	if verificationAt, ok := result["verificationAt"].(string); ok {
		// Calculate hours since verification
		verTime, err := time.Parse(time.RFC3339, verificationAt)
		if err == nil {
			hoursSince := time.Since(verTime).Hours()
			result["hoursSinceLastVerification"] = hoursSince
		}
	}
	
	r.logger.Info("Successfully fetched historical verification", map[string]interface{}{
		"verificationId": verificationId,
		"status":         result["status"],
		"recordCount":    len(queryOutput.Items),
	})
	
	return result, nil
}

// UpdateVerificationStatus updates the verification status in DynamoDB
func (r *DynamoDBRepository) UpdateVerificationStatus(
	ctx context.Context, 
	verificationId string, 
	status string,
) error {
	// Create the key for the DynamoDB update
	key := map[string]types.AttributeValue{
		"verificationId": &types.AttributeValueMemberS{
			Value: verificationId,
		},
	}
	
	// Create the update expression
	updateExpression := "SET #status = :status"
	expressionAttributeNames := map[string]string{
		"#status": "status",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":status": &types.AttributeValueMemberS{
			Value: status,
		},
	}
	
	r.logger.Info("Updating verification status", map[string]interface{}{
		"verificationId": verificationId,
		"status":         status,
		"table":          r.verificationTable,
	})
	
	// Update the item in DynamoDB
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.verificationTable),
		Key:                       key,
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
	})
	
	if err != nil {
		return fmt.Errorf("failed to update verification status: %w", err)
	}
	
	r.logger.Info("Successfully updated verification status", map[string]interface{}{
		"verificationId": verificationId,
		"status":         status,
	})
	
	return nil
}