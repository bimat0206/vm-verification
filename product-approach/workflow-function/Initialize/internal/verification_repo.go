package internal

import (
	"context"
	"errors"
	"fmt"

	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Custom error types
var (
	ErrVerificationExists  = errors.New("verification record already exists")
	ErrNoTableDefined      = errors.New("verification table not defined in config")
	ErrMissingVerificationId = errors.New("verification ID is missing")
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

// StoreMinimalRecord stores a minimal verification record in DynamoDB with S3 reference
func (r *VerificationRepository) StoreMinimalRecord(
	ctx context.Context, 
	verification *schema.VerificationContext,
	s3Reference *s3state.Reference,
) error {
	if r.config.VerificationTable == "" {
		return ErrNoTableDefined
	}

	// Validate the verification context has a valid ID
	if verification == nil {
		r.logger.Error("Verification context is nil", nil)
		return fmt.Errorf("verification context cannot be nil")
	}
	
	// Crucial validation: ensure verificationId is present and not empty
	if verification.VerificationId == "" {
		r.logger.Error("VerificationId is missing", map[string]interface{}{
			"verification": fmt.Sprintf("%+v", verification),
		})
		return ErrMissingVerificationId
	}
	
	// Also ensure verificationAt is present and not empty
	if verification.VerificationAt == "" {
		r.logger.Error("VerificationAt is missing", map[string]interface{}{
			"verification": verification.VerificationId,
		})
		return fmt.Errorf("missing required field: verificationAt")
	}
	
	// Log verification context before processing
	r.logger.Debug("Storing minimal verification record", map[string]interface{}{
		"verificationId": verification.VerificationId,
		"verificationType": verification.VerificationType,
		"status": verification.Status,
	})
	
	// Calculate TTL for the record (30 days default)
	ttl := r.dbClient.CalculateTTL(r.config.DefaultTTLDays)
	
	// Minimal record with only essential fields and S3 reference
	type MinimalVerificationRecord struct {
		// Primary key fields with lowercase names
		VerificationId    string `json:"verificationId" dynamodbav:"verificationId"`
		VerificationAt    string `json:"verificationAt" dynamodbav:"verificationAt"`
		
		// Status for tracking
		Status            string `json:"status,omitempty" dynamodbav:"status,omitempty"`
		
		// Basic metadata
		VerificationType  string `json:"verificationType,omitempty" dynamodbav:"verificationType,omitempty"`
		VendingMachineId  string `json:"vendingMachineId,omitempty" dynamodbav:"vendingMachineId,omitempty"`
		
		// Image URLs for verification
		ReferenceImageUrl string `json:"referenceImageUrl,omitempty" dynamodbav:"referenceImageUrl,omitempty"`
		CheckingImageUrl  string `json:"checkingImageUrl,omitempty" dynamodbav:"checkingImageUrl,omitempty"`
		
		// Layout information for LAYOUT_VS_CHECKING verification type
		LayoutId          int    `json:"layoutId,omitempty" dynamodbav:"layoutId,omitempty"`
		LayoutPrefix      string `json:"layoutPrefix,omitempty" dynamodbav:"layoutPrefix,omitempty"`
		
		// S3 reference
		S3StateBucket     string `json:"s3StateBucket,omitempty" dynamodbav:"s3StateBucket,omitempty"`
		S3StateKey        string `json:"s3StateKey,omitempty" dynamodbav:"s3StateKey,omitempty"`
		S3StateSize       int64  `json:"s3StateSize,omitempty" dynamodbav:"s3StateSize,omitempty"`
		
		// TTL field for DynamoDB
		TTL int64 `json:"ttl" dynamodbav:"ttl"`
	}
	
	// Map verification context to record ensuring proper field casing
	record := MinimalVerificationRecord{
		VerificationId:      verification.VerificationId,
		VerificationAt:      verification.VerificationAt,
		Status:              verification.Status,
		VerificationType:    verification.VerificationType,
		VendingMachineId:    verification.VendingMachineId,
		ReferenceImageUrl:   verification.ReferenceImageUrl,
		CheckingImageUrl:    verification.CheckingImageUrl,
		LayoutId:            verification.LayoutId,
		LayoutPrefix:        verification.LayoutPrefix,
		TTL:                 ttl,
	}
	
	// Add S3 reference if provided
	if s3Reference != nil {
		record.S3StateBucket = s3Reference.Bucket
		record.S3StateKey = s3Reference.Key
		record.S3StateSize = s3Reference.Size
	}
	
	// Marshal the record with proper field casing
	item, err := attributevalue.MarshalMap(record)
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
	
	r.logger.Info("Storing minimal verification record", map[string]interface{}{
		"verificationId":     verification.VerificationId,
		"verificationType":   verification.VerificationType,
		"table":              r.config.VerificationTable,
		"referenceImageUrl":  record.ReferenceImageUrl,
		"checkingImageUrl":   record.CheckingImageUrl,
		"layoutId":           record.LayoutId,
		"layoutPrefix":       record.LayoutPrefix,
		"s3StateBucket":      record.S3StateBucket,
		"s3StateKey":         record.S3StateKey,
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

// Legacy methods for backward compatibility
// StoreVerificationRecord stores a full verification record in DynamoDB with TTL
func (r *VerificationRepository) StoreVerificationRecord(ctx context.Context, verification *schema.VerificationContext) error {
	r.logger.Warn("Using legacy StoreVerificationRecord method - consider using StoreMinimalRecord", map[string]interface{}{
		"verificationId": verification.VerificationId,
	})
	
	if r.config.VerificationTable == "" {
		return ErrNoTableDefined
	}

	// Validate the verification context has a valid ID
	if verification == nil {
		r.logger.Error("Verification context is nil", nil)
		return fmt.Errorf("verification context cannot be nil")
	}
	
	// Crucial validation: ensure verificationId is present and not empty
	if verification.VerificationId == "" {
		r.logger.Error("VerificationId is missing", map[string]interface{}{
			"verification": fmt.Sprintf("%+v", verification),
		})
		return ErrMissingVerificationId
	}
	
	// Also ensure verificationAt is present and not empty
	if verification.VerificationAt == "" {
		r.logger.Error("VerificationAt is missing", map[string]interface{}{
			"verification": verification.VerificationId,
		})
		return fmt.Errorf("missing required field: verificationAt")
	}
	
	// Calculate TTL for the record (30 days default)
	ttl := r.dbClient.CalculateTTL(r.config.DefaultTTLDays)
	
	// Use a structure specifically designed for DynamoDB storage
	// with lowercase field names for the primary key fields
	type VerificationRecord struct {
		// Primary key fields with lowercase names
		VerificationId    string `json:"verificationId" dynamodbav:"verificationId"`
		VerificationAt    string `json:"verificationAt" dynamodbav:"verificationAt"`
		
		// All other fields from VerificationContext
		Status            string `json:"status,omitempty" dynamodbav:"status,omitempty"`
		VerificationType  string `json:"verificationType,omitempty" dynamodbav:"verificationType,omitempty"`
		VendingMachineId  string `json:"vendingMachineId,omitempty" dynamodbav:"vendingMachineId,omitempty"`
		ReferenceImageUrl string `json:"referenceImageUrl,omitempty" dynamodbav:"referenceImageUrl,omitempty"`
		CheckingImageUrl  string `json:"checkingImageUrl,omitempty" dynamodbav:"checkingImageUrl,omitempty"`
		LayoutId          int    `json:"layoutId,omitempty" dynamodbav:"layoutId,omitempty"`
		LayoutPrefix      string `json:"layoutPrefix,omitempty" dynamodbav:"layoutPrefix,omitempty"`
		PreviousVerificationId string `json:"previousVerificationId,omitempty" dynamodbav:"previousVerificationId,omitempty"`
		NotificationEnabled bool   `json:"notificationEnabled,omitempty" dynamodbav:"notificationEnabled,omitempty"`
		
		// Embedded fields
		ResourceValidation *schema.ResourceValidation `json:"resourceValidation,omitempty" dynamodbav:"resourceValidation,omitempty"`
		ConversationType    string                   `json:"conversationType,omitempty" dynamodbav:"conversationType,omitempty"`
		TurnConfig          *schema.TurnConfig      `json:"turnConfig,omitempty" dynamodbav:"turnConfig,omitempty"`
		TurnTimestamps      *schema.TurnTimestamps  `json:"turnTimestamps,omitempty" dynamodbav:"turnTimestamps,omitempty"`
		RequestMetadata     *schema.RequestMetadata `json:"requestMetadata,omitempty" dynamodbav:"requestMetadata,omitempty"`
		Error               *schema.ErrorInfo       `json:"error,omitempty" dynamodbav:"error,omitempty"`
		
		// TTL field for DynamoDB
		TTL int64 `json:"ttl" dynamodbav:"ttl"`
	}
	
	// Map verification context to record ensuring proper field casing
	record := VerificationRecord{
		VerificationId:      verification.VerificationId,
		VerificationAt:      verification.VerificationAt,
		Status:              verification.Status,
		VerificationType:    verification.VerificationType,
		VendingMachineId:    verification.VendingMachineId,
		ReferenceImageUrl:   verification.ReferenceImageUrl,
		CheckingImageUrl:    verification.CheckingImageUrl,
		LayoutId:            verification.LayoutId,
		LayoutPrefix:        verification.LayoutPrefix,
		PreviousVerificationId: verification.PreviousVerificationId,
		NotificationEnabled: verification.NotificationEnabled,
		ResourceValidation:  verification.ResourceValidation,
		ConversationType:    verification.ConversationType,
		TurnConfig:          verification.TurnConfig,
		TurnTimestamps:      verification.TurnTimestamps,
		RequestMetadata:     verification.RequestMetadata,
		Error:               verification.Error,
		TTL:                 ttl,
	}
	
	// Marshal the record with proper field casing
	item, err := attributevalue.MarshalMap(record)
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
	
	// Validate verificationId is not empty
	if verificationId == "" {
		r.logger.Error("GetVerificationRecord called with empty verificationId", nil)
		return nil, ErrMissingVerificationId
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
	
	// Validate that verificationId was properly unmarshaled
	if verification.VerificationId == "" {
		r.logger.Warn("VerificationId missing after unmarshal, setting it explicitly", map[string]interface{}{
			"verificationId": verificationId,
		})
		verification.VerificationId = verificationId
	}
	
	r.logger.Debug("Retrieved verification record", map[string]interface{}{
		"verificationId":   verification.VerificationId,
		"verificationType": verification.VerificationType,
	})
	
	return &verification, nil
}

// FindPreviousVerification finds the most recent verification using the reference image
func (r *VerificationRepository) FindPreviousVerification(ctx context.Context, checkingImageUrl string) (*schema.VerificationContext, error) {
	if r.config.VerificationTable == "" {
		return nil, ErrNoTableDefined
	}
	
	// Validate checkingImageUrl is not empty
	if checkingImageUrl == "" {
		r.logger.Error("FindPreviousVerification called with empty checkingImageUrl", nil)
		return nil, fmt.Errorf("checkingImageUrl cannot be empty")
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
	
	// Validate that verificationId is present after unmarshaling
	if verification.VerificationId == "" {
		r.logger.Warn("VerificationId missing in query result", map[string]interface{}{
			"checkingImageUrl": checkingImageUrl,
		})
		// This is a more serious issue since we don't know the actual ID
		// In this case we'll return the error rather than trying to fix it
		return nil, fmt.Errorf("previous verification record has missing verificationId")
	}
	
	r.logger.Info("Found previous verification", map[string]interface{}{
		"verificationId":   verification.VerificationId,
		"verificationType": verification.VerificationType,
		"verificationAt":   verification.VerificationAt,
	})
	
	return &verification, nil
}