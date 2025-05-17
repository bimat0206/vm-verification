package internal

import (
	"context"
	"errors"
	"fmt"
	
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"workflow-function/shared/logger"
)

// Custom error types
var (
	ErrLayoutTableNotDefined = errors.New("layout table not defined in config")
	ErrLayoutNotFound        = errors.New("layout not found")
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
// This is included for future use, not currently used in the initialization process
func (r *LayoutRepository) GetLayoutMetadata(ctx context.Context, layoutId int, layoutPrefix string) (map[string]interface{}, error) {
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
	
	// Unmarshal the result into a map
	var metadata map[string]interface{}
	if err := attributevalue.UnmarshalMap(result.Item, &metadata); err != nil {
		r.logger.Error("Failed to unmarshal layout metadata", map[string]interface{}{
			"error":        err.Error(),
			"layoutId":     layoutId,
			"layoutPrefix": layoutPrefix,
		})
		return nil, fmt.Errorf("failed to unmarshal layout metadata: %w", err)
	}
	
	return metadata, nil
}