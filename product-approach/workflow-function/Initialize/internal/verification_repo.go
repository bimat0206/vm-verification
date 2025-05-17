package internal

import (
	"context"
	"errors"
	"fmt"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// Custom error types
var (
	ErrVerificationExists = errors.New("verification record already exists")
	ErrNoTableDefined     = errors.New("verification table not defined in config")
)

// VerificationRepository handles operations on verification records
type VerificationRepository struct {
	dbClient *DynamoDBClient
	logger   logger.Logger
	config   Config
}

// NewVerificationRepository creates a new VerificationRepository
func NewVerificationRepository(dbClient *DynamoDBClient, cfg Config, log logger.Logger) *VerificationRepository {
	return &VerificationRepository{
		dbClient: dbClient,
		logger:   log.WithFields(map[string]interface{}{"component": "VerificationRepository"}),
		config:   cfg,
	}
}

// StoreVerificationRecord stores a verification record in DynamoDB with TTL
func (r *VerificationRepository) StoreVerificationRecord(ctx context.Context, verification *schema.VerificationContext) error {
	if r.config.VerificationTable == "" {
		return ErrNoTableDefined
	}
	
	// Calculate TTL for the record (30 days default)
	ttl := r.dbClient.CalculateTTL(r.config.DefaultTTLDays)
	
	// Marshal the verification context to DynamoDB attributes
	// Create a struct that includes the verification context and TTL
	type VerificationWithTTL struct {
		schema.VerificationContext
		TTL int64 `json:"ttl" dynamodbav:"ttl"`
	}
	
	// First create a copy with the TTL field
	verificationWithTTL := VerificationWithTTL{
		VerificationContext: *verification,
		TTL: ttl,
	}
	
	item, err := attributevalue.MarshalMap(verificationWithTTL)
	if err != nil {
		r.logger.Error("Failed to marshal verification record", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": verification.VerificationId,
		})
		return fmt.Errorf("failed to marshal verification record: %w", err)
	}
	
	// Create condition expression to ensure idempotency
	// Only insert if the item doesn't already exist
	conditionExpression := aws.String("attribute_not_exists(verificationId)")
	
	r.logger.Info("Storing verification record", map[string]interface{}{
		"verificationId":   verification.VerificationId,
		"verificationType": verification.VerificationType,
		"table":            r.config.VerificationTable,
	})
	
	_, err = r.dbClient.PutItem(
		ctx,
		r.config.VerificationTable,
		item,
		conditionExpression,
		nil, // No expression attribute names
		nil, // No expression attribute values
	)
	
	if err != nil {
		// Check if this is a condition check failure
		if errors.Is(err, ErrConditionFailed) {
			r.logger.Warn("Verification record already exists", map[string]interface{}{
				"verificationId": verification.VerificationId,
			})
			return ErrVerificationExists
		}
		
		r.logger.Error("Failed to store verification record", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": verification.VerificationId,
		})
		return fmt.Errorf("failed to store verification record: %w", err)
	}
	
	r.logger.Info("Verification record stored successfully", map[string]interface{}{
		"verificationId": verification.VerificationId,
	})
	
	return nil
}

// GetVerificationRecord retrieves a verification record by ID
func (r *VerificationRepository) GetVerificationRecord(ctx context.Context, verificationId string) (*schema.VerificationContext, error) {
	if r.config.VerificationTable == "" {
		return nil, ErrNoTableDefined
	}
	
	// Create the key for the query
	key, err := attributevalue.MarshalMap(map[string]string{
		"verificationId": verificationId,
	})
	if err != nil {
		r.logger.Error("Failed to marshal key", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": verificationId,
		})
		return nil, fmt.Errorf("failed to marshal key: %w", err)
	}
	
	r.logger.Debug("Getting verification record", map[string]interface{}{
		"verificationId": verificationId,
		"table":          r.config.VerificationTable,
	})
	
	result, err := r.dbClient.GetItem(ctx, r.config.VerificationTable, key)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			r.logger.Warn("Verification record not found", map[string]interface{}{
				"verificationId": verificationId,
			})
			return nil, nil // Return nil instead of error for not found
		}
		
		r.logger.Error("Failed to get verification record", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": verificationId,
		})
		return nil, fmt.Errorf("failed to get verification record: %w", err)
	}
	
	// Unmarshal the result into a VerificationContext
	var verification schema.VerificationContext
	if err := attributevalue.UnmarshalMap(result.Item, &verification); err != nil {
		r.logger.Error("Failed to unmarshal verification record", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": verificationId,
		})
		return nil, fmt.Errorf("failed to unmarshal verification record: %w", err)
	}
	
	r.logger.Debug("Retrieved verification record", map[string]interface{}{
		"verificationId":   verificationId,
		"verificationType": verification.VerificationType,
	})
	
	return &verification, nil
}

// FindPreviousVerification finds the most recent verification using the reference image
func (r *VerificationRepository) FindPreviousVerification(ctx context.Context, checkingImageUrl string) (*schema.VerificationContext, error) {
	if r.config.VerificationTable == "" {
		return nil, ErrNoTableDefined
	}
	
	r.logger.Debug("Finding previous verification by checking image URL", map[string]interface{}{
		"checkingImageUrl": checkingImageUrl,
		"table":            r.config.VerificationTable,
	})
	
	// Create the key condition expression for the query
	keyExpr := aws.String("checkingImageUrl = :url")
	
	// Create the expression attribute values
	exprValues, err := attributevalue.MarshalMap(map[string]string{
		":url": checkingImageUrl,
	})
	if err != nil {
		r.logger.Error("Failed to marshal expression values", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to marshal expression values: %w", err)
	}
	
	// Set up the DynamoDB query
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(r.config.VerificationTable),
		IndexName:                 aws.String("GSI4"), // Assuming GSI4 indexes on checkingImageUrl
		KeyConditionExpression:    keyExpr,
		ExpressionAttributeValues: exprValues,
		ScanIndexForward:          aws.Bool(false), // Sort in descending order to get most recent first
		Limit:                     aws.Int32(1),    // Only need the most recent
	}
	
	// Execute the query
	result, err := r.dbClient.Query(ctx, queryInput)
	if err != nil {
		r.logger.Error("Failed to query for previous verification", map[string]interface{}{
			"error":           err.Error(),
			"checkingImageUrl": checkingImageUrl,
		})
		return nil, fmt.Errorf("failed to query for previous verification: %w", err)
	}
	
	// Check if we found any results
	if len(result.Items) == 0 {
		r.logger.Info("No previous verification found", map[string]interface{}{
			"checkingImageUrl": checkingImageUrl,
		})
		return nil, nil // Return nil instead of error for not found
	}
	
	// Unmarshal the first (most recent) result
	var verification schema.VerificationContext
	if err := attributevalue.UnmarshalMap(result.Items[0], &verification); err != nil {
		r.logger.Error("Failed to unmarshal previous verification", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to unmarshal previous verification: %w", err)
	}
	
	r.logger.Info("Found previous verification", map[string]interface{}{
		"verificationId":   verification.VerificationId,
		"verificationType": verification.VerificationType,
		"verificationAt":   verification.VerificationAt,
	})
	
	return &verification, nil
}