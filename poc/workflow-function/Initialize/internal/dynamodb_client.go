package internal

import (
	"context"
	"errors"
	"time"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"workflow-function/shared/logger"
)

// DynamoDB error types
var (
	ErrItemNotFound       = errors.New("item not found in DynamoDB")
	ErrConditionFailed    = errors.New("condition check failed")
	ErrTableNotSpecified  = errors.New("table name not specified")
	ErrMarshalingFailed   = errors.New("failed to marshal item")
	ErrUnmarshalingFailed = errors.New("failed to unmarshal item")
	ErrMissingPrimaryKey  = errors.New("missing primary key attributes")
)

// DynamoDBClient wraps DynamoDB operations with consistent error handling
type DynamoDBClient struct {
	client *dynamodb.Client
	logger logger.Logger
	config Config
}

// NewDynamoDBClient creates a new DynamoDB client wrapper
func NewDynamoDBClient(awsConfig aws.Config, cfg Config, log logger.Logger) *DynamoDBClient {
	client := dynamodb.NewFromConfig(awsConfig)
	
	return &DynamoDBClient{
		client: client,
		logger: log.WithFields(map[string]interface{}{"component": "DynamoDBClient"}),
		config: cfg,
	}
}

// GetItem retrieves an item from DynamoDB
func (c *DynamoDBClient) GetItem(ctx context.Context, tableName string, key map[string]types.AttributeValue) (*dynamodb.GetItemOutput, error) {
	if tableName == "" {
		return nil, ErrTableNotSpecified
	}
	
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	}
	
	c.logger.Debug("Getting item from DynamoDB", map[string]interface{}{
		"table": tableName,
		"key":   key,
	})
	
	result, err := c.client.GetItem(ctx, input)
	if err != nil {
		c.logger.Error("Failed to get item from DynamoDB", map[string]interface{}{
			"table": tableName,
			"key":   key,
			"error": err.Error(),
		})
		return nil, err
	}
	
	// Check if item exists (empty result)
	if result.Item == nil || len(result.Item) == 0 {
		return nil, ErrItemNotFound
	}
	
	return result, nil
}

// PutItem stores an item in DynamoDB with conditional expression
func (c *DynamoDBClient) PutItem(
	ctx context.Context, 
	tableName string, 
	item map[string]types.AttributeValue,
	conditionExpression *string,
	expressionAttrNames map[string]string,
	expressionAttrValues map[string]types.AttributeValue,
) (*dynamodb.PutItemOutput, error) {
	if tableName == "" {
		return nil, ErrTableNotSpecified
	}
	
	// Validate and fix primary key issues before attempting to put the item
	item = c.ensurePrimaryKeys(tableName, item)
	
	// Log the keys present in the item for debugging
	keys := make([]string, 0, len(item))
	for k := range item {
		keys = append(keys, k)
	}
	
	c.logger.Debug("Putting item in DynamoDB", map[string]interface{}{
		"table": tableName,
		"keys": keys,
	})
	
	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	}
	
	// Add conditional expression if provided
	if conditionExpression != nil {
		input.ConditionExpression = conditionExpression
	}
	
	// Add expression attribute names if provided
	if len(expressionAttrNames) > 0 {
		input.ExpressionAttributeNames = expressionAttrNames
	}
	
	// Add expression attribute values if provided
	if len(expressionAttrValues) > 0 {
		input.ExpressionAttributeValues = expressionAttrValues
	}
	
	result, err := c.client.PutItem(ctx, input)
	if err != nil {
		var conditionCheckErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionCheckErr) {
			c.logger.Warn("Condition check failed for DynamoDB put", map[string]interface{}{
				"table": tableName,
				"error": err.Error(),
			})
			return nil, ErrConditionFailed
		}
		
		c.logger.Error("Failed to put item in DynamoDB", map[string]interface{}{
			"table": tableName,
			"keys": keys,
			"error": err.Error(),
		})
		return nil, err
	}
	
	return result, nil
}

// ensurePrimaryKeys ensures the required primary keys exist in the item
// and fixes any case-sensitive key issues
func (c *DynamoDBClient) ensurePrimaryKeys(tableName string, item map[string]types.AttributeValue) map[string]types.AttributeValue {
	// Check for VerificationResults table
	if tableName == c.config.VerificationTable {
		// Handle case-sensitive keys for verificationId/VerificationId
		_, hasLowerVerificationId := item["verificationId"]
		upperVerificationId, hasUpperVerificationId := item["VerificationId"]
		
		if !hasLowerVerificationId && hasUpperVerificationId {
			// Copy from uppercase to lowercase which is the expected key name
			c.logger.Info("Fixed case sensitivity for verificationId", nil)
			item["verificationId"] = upperVerificationId
		}
		
		// Handle case-sensitive keys for verificationAt/VerificationAt
		_, hasLowerVerificationAt := item["verificationAt"]
		upperVerificationAt, hasUpperVerificationAt := item["VerificationAt"]
		
		if !hasLowerVerificationAt && hasUpperVerificationAt {
			// Copy from uppercase to lowercase which is the expected key name
			c.logger.Info("Fixed case sensitivity for verificationAt", nil)
			item["verificationAt"] = upperVerificationAt
		}
		
		// Validate both keys are present
		if _, hasVerificationId := item["verificationId"]; !hasVerificationId {
			c.logger.Error("Missing required partition key 'verificationId'", nil)
		}
		
		if _, hasVerificationAt := item["verificationAt"]; !hasVerificationAt {
			c.logger.Error("Missing required sort key 'verificationAt'", nil)
		}
	}
	
	// Check for layoutId and layoutPrefix for LayoutMetadata table
	if tableName == c.config.LayoutTable {
		// Handle case-sensitive keys for layoutId/LayoutId
		_, hasLowerLayoutId := item["layoutId"]
		upperLayoutId, hasUpperLayoutId := item["LayoutId"]
		
		if !hasLowerLayoutId && hasUpperLayoutId {
			// Copy from uppercase to lowercase
			c.logger.Info("Fixed case sensitivity for layoutId", nil)
			item["layoutId"] = upperLayoutId
		}
		
		// Handle case-sensitive keys for layoutPrefix/LayoutPrefix
		_, hasLowerLayoutPrefix := item["layoutPrefix"]
		upperLayoutPrefix, hasUpperLayoutPrefix := item["LayoutPrefix"]
		
		if !hasLowerLayoutPrefix && hasUpperLayoutPrefix {
			// Copy from uppercase to lowercase
			c.logger.Info("Fixed case sensitivity for layoutPrefix", nil)
			item["layoutPrefix"] = upperLayoutPrefix
		}
	}
	
	return item
}

// validatePrimaryKey checks if the required primary key field(s) exist in the item
// This is a critical validation to prevent DynamoDB ValidationException errors
func (c *DynamoDBClient) validatePrimaryKey(tableName string, item map[string]types.AttributeValue) error {
	// Check for verificationId and verificationAt for VerificationResults table
	if tableName == c.config.VerificationTable {
		// Check lowercase and uppercase variants for verificationId
		_, hasLowerVerificationId := item["verificationId"]
		_, hasUpperVerificationId := item["VerificationId"]
		
		if !hasLowerVerificationId && !hasUpperVerificationId {
			c.logger.Error("Missing primary key 'verificationId' for table", map[string]interface{}{
				"table": tableName,
			})
			return ErrMissingPrimaryKey
		}
		
		// Check lowercase and uppercase variants for verificationAt
		_, hasLowerVerificationAt := item["verificationAt"]
		_, hasUpperVerificationAt := item["VerificationAt"]
		
		if !hasLowerVerificationAt && !hasUpperVerificationAt {
			c.logger.Error("Missing sort key 'verificationAt' for table", map[string]interface{}{
				"table": tableName,
			})
			return ErrMissingPrimaryKey
		}
	}
	
	// Check for layoutId and layoutPrefix for LayoutMetadata table
	if tableName == c.config.LayoutTable {
		// Check lowercase and uppercase variants for layoutId
		_, hasLowerLayoutId := item["layoutId"]
		_, hasUpperLayoutId := item["LayoutId"]
		
		if !hasLowerLayoutId && !hasUpperLayoutId {
			c.logger.Error("Missing partition key 'layoutId' for table", map[string]interface{}{
				"table": tableName,
			})
			return ErrMissingPrimaryKey
		}
		
		// Check lowercase and uppercase variants for layoutPrefix
		_, hasLowerLayoutPrefix := item["layoutPrefix"]
		_, hasUpperLayoutPrefix := item["LayoutPrefix"]
		
		if !hasLowerLayoutPrefix && !hasUpperLayoutPrefix {
			c.logger.Error("Missing sort key 'layoutPrefix' for table", map[string]interface{}{
				"table": tableName,
			})
			return ErrMissingPrimaryKey
		}
	}
	
	return nil
}

// Query executes a query against a DynamoDB table or index
func (c *DynamoDBClient) Query(
	ctx context.Context,
	input *dynamodb.QueryInput,
) (*dynamodb.QueryOutput, error) {
	if input.TableName == nil || *input.TableName == "" {
		return nil, ErrTableNotSpecified
	}
	
	c.logger.Debug("Querying DynamoDB", map[string]interface{}{
		"table": *input.TableName,
		"index": aws.ToString(input.IndexName),
	})
	
	result, err := c.client.Query(ctx, input)
	if err != nil {
		c.logger.Error("Failed to query DynamoDB", map[string]interface{}{
			"table": *input.TableName,
			"index": aws.ToString(input.IndexName),
			"error": err.Error(),
		})
		return nil, err
	}
	
	return result, nil
}

// CalculateTTL returns a TTL timestamp for the specified number of days from now
func (c *DynamoDBClient) CalculateTTL(daysFromNow int) int64 {
	if daysFromNow <= 0 {
		// Use default from config
		daysFromNow = c.config.DefaultTTLDays
	}
	
	// Calculate expiration time
	expiry := time.Now().UTC().AddDate(0, 0, daysFromNow)
	return expiry.Unix()
}

// Client returns the underlying DynamoDB client
func (c *DynamoDBClient) Client() *dynamodb.Client {
	return c.client
}

// MarshalMap is a convenience wrapper around the attributevalue.MarshalMap function
// with additional validation for required fields
func MarshalMap(item interface{}) (map[string]types.AttributeValue, error) {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return nil, ErrMarshalingFailed
	}
	
	return av, nil
}

// UnmarshalMap is a convenience wrapper around the attributevalue.UnmarshalMap function
func UnmarshalMap(av map[string]types.AttributeValue, out interface{}) error {
	err := attributevalue.UnmarshalMap(av, out)
	if err != nil {
		return ErrUnmarshalingFailed
	}
	return nil
}