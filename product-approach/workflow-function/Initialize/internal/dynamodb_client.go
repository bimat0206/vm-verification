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
	
	c.logger.Debug("Putting item in DynamoDB", map[string]interface{}{
		"table": tableName,
	})
	
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
			"error": err.Error(),
		})
		return nil, err
	}
	
	return result, nil
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