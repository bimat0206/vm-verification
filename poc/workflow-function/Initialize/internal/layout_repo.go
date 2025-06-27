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
	ErrLayoutTableNotDefined     = errors.New("layout table not defined in config")
	ErrLayoutNotFound            = errors.New("layout not found")
	ErrGSIQueryFailed            = errors.New("GSI query failed")
	ErrLayoutGsiNameNotDefined = errors.New("layout GSI name for reference image lookup not defined in config")
)

const (
	// DefaultGsiNameReferenceImage is an example GSI name.
	// This should ideally be configurable if it varies.
	DefaultGsiNameReferenceImage = "ReferenceImageIndex-gsi"
)

// LayoutRepository handles operations on layout metadata
type LayoutRepository struct {
	dbClient *DynamoDBClient
	logger   logger.Logger
	config   Config
}

// NewLayoutRepository creates a new LayoutRepository
func NewLayoutRepository(dbClient *DynamoDBClient, cfg Config, log logger.Logger) *LayoutRepository {
	return &LayoutRepository{
		dbClient: dbClient,
		logger:   log.WithFields(map[string]interface{}{"component": "LayoutRepository"}),
		config:   cfg,
	}
}

// VerifyLayoutExists checks if a layout exists in DynamoDB
func (r *LayoutRepository) VerifyLayoutExists(ctx context.Context, layoutId int, layoutPrefix string) (bool, error) {
	if r.config.LayoutTable == "" {
		return false, ErrLayoutTableNotDefined
	}

	r.logger.Debug("Verifying layout exists", map[string]interface{}{
		"layoutId":     layoutId,
		"layoutPrefix": layoutPrefix,
		"table":        r.config.LayoutTable,
	})

	// Create the composite key for the query
	key, err := attributevalue.MarshalMap(map[string]interface{}{
		"layoutId":     layoutId,
		"layoutPrefix": layoutPrefix,
	})
	if err != nil {
		r.logger.Error("Failed to marshal layout key", map[string]interface{}{
			"error":        err.Error(),
			"layoutId":     layoutId,
			"layoutPrefix": layoutPrefix,
		})
		return false, fmt.Errorf("failed to marshal layout key: %w", err)
	}

	// Execute the GetItem operation
	_, err = r.dbClient.GetItem(ctx, r.config.LayoutTable, key)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			r.logger.Warn("Layout not found", map[string]interface{}{
				"layoutId":     layoutId,
				"layoutPrefix": layoutPrefix,
			})
			return false, nil // Return false without error for not found
		}

		r.logger.Error("Failed to verify layout exists", map[string]interface{}{
			"error":        err.Error(),
			"layoutId":     layoutId,
			"layoutPrefix": layoutPrefix,
		})
		return false, fmt.Errorf("failed to verify layout exists: %w", err)
	}

	// Layout exists
	r.logger.Info("Layout exists", map[string]interface{}{
		"layoutId":     layoutId,
		"layoutPrefix": layoutPrefix,
	})

	return true, nil
}

// GetLayoutMetadata retrieves full layout metadata
func (r *LayoutRepository) GetLayoutMetadata(ctx context.Context, layoutId int, layoutPrefix string) (*schema.LayoutMetadata, error) {
	if r.config.LayoutTable == "" {
		return nil, ErrLayoutTableNotDefined
	}

	r.logger.Debug("Getting layout metadata", map[string]interface{}{
		"layoutId":     layoutId,
		"layoutPrefix": layoutPrefix,
		"table":        r.config.LayoutTable,
	})

	// Create the composite key for the query
	key, err := attributevalue.MarshalMap(map[string]interface{}{
		"layoutId":     layoutId,
		"layoutPrefix": layoutPrefix,
	})
	if err != nil {
		r.logger.Error("Failed to marshal layout key", map[string]interface{}{
			"error":        err.Error(),
			"layoutId":     layoutId,
			"layoutPrefix": layoutPrefix,
		})
		return nil, fmt.Errorf("failed to marshal layout key: %w", err)
	}

	// Execute the GetItem operation
	result, err := r.dbClient.GetItem(ctx, r.config.LayoutTable, key)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			r.logger.Warn("Layout metadata not found", map[string]interface{}{
				"layoutId":     layoutId,
				"layoutPrefix": layoutPrefix,
			})
			return nil, ErrLayoutNotFound
		}

		r.logger.Error("Failed to get layout metadata", map[string]interface{}{
			"error":        err.Error(),
			"layoutId":     layoutId,
			"layoutPrefix": layoutPrefix,
		})
		return nil, fmt.Errorf("failed to get layout metadata: %w", err)
	}

	var metadata schema.LayoutMetadata
	if err := attributevalue.UnmarshalMap(result.Item, &metadata); err != nil {
		r.logger.Error("Failed to unmarshal layout metadata", map[string]interface{}{
			"error":        err.Error(),
			"layoutId":     layoutId,
			"layoutPrefix": layoutPrefix,
		})
		return nil, fmt.Errorf("failed to unmarshal layout metadata: %w", err)
	}

	return &metadata, nil
}

// GetLayoutByReferenceImage retrieves layout metadata by its referenceImageUrl using a GSI.
func (r *LayoutRepository) GetLayoutByReferenceImage(ctx context.Context, referenceImageUrl string) (*schema.LayoutMetadata, error) {
	if r.config.LayoutTable == "" {
		return nil, ErrLayoutTableNotDefined
	}
	// NOTE: The GSI name should ideally come from config (e.g., r.config.LayoutGsiName)
	// For now, using a hardcoded example name.
	gsiName := DefaultGsiNameReferenceImage
	if gsiName == "" { // Or if r.config.LayoutGsiName == ""
		r.logger.Error("Layout GSI name for reference image lookup is not configured.", nil)
		return nil, ErrLayoutGsiNameNotDefined
	}

	r.logger.Debug("Getting layout metadata by referenceImageUrl", map[string]interface{}{
		"referenceImageUrl": referenceImageUrl,
		"table":             r.config.LayoutTable,
		"gsiName":           gsiName,
	})

	exprValues, err := attributevalue.MarshalMap(map[string]string{
		":refImgUrl": referenceImageUrl,
	})
	if err != nil {
		r.logger.Error("Failed to marshal expression values for referenceImageUrl query", map[string]interface{}{
			"error":             err.Error(),
			"referenceImageUrl": referenceImageUrl,
		})
		return nil, fmt.Errorf("failed to marshal expression values: %w", err)
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(r.config.LayoutTable),
		IndexName:                 aws.String(gsiName),
		KeyConditionExpression:    aws.String("referenceImageUrl = :refImgUrl"),
		ExpressionAttributeValues: exprValues,
		Limit:                     aws.Int32(1), // Assuming referenceImageUrl is unique enough for the active layout
	}

	result, err := r.dbClient.Query(ctx, queryInput)
	if err != nil {
		r.logger.Error("Failed to query layout by referenceImageUrl", map[string]interface{}{
			"error":             err.Error(),
			"referenceImageUrl": referenceImageUrl,
			"gsiName":           gsiName,
		})
		return nil, fmt.Errorf("%w: %v", ErrGSIQueryFailed, err)
	}

	if len(result.Items) == 0 {
		r.logger.Warn("Layout not found for the provided referenceImageUrl", map[string]interface{}{
			"referenceImageUrl": referenceImageUrl,
			"gsiName":           gsiName,
		})
		return nil, ErrLayoutNotFound // Or return nil, nil if "not found" is not an error for the caller
	}

	var metadata schema.LayoutMetadata
	if err := attributevalue.UnmarshalMap(result.Items[0], &metadata); err != nil {
		r.logger.Error("Failed to unmarshal layout metadata from GSI query result", map[string]interface{}{
			"error":             err.Error(),
			"referenceImageUrl": referenceImageUrl,
		})
		return nil, fmt.Errorf("failed to unmarshal layout metadata from GSI: %w", err)
	}

	r.logger.Info("Layout metadata retrieved successfully by referenceImageUrl", map[string]interface{}{
		"layoutId":          metadata.LayoutId,
		"layoutPrefix":      metadata.LayoutPrefix,
		"referenceImageUrl": referenceImageUrl,
	})

	return &metadata, nil
}