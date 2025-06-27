package errors

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Example: Enhanced DynamoDB client wrapper
type EnhancedDynamoDBClient struct {
	client        *dynamodb.Client
	defaultTable  string
	correlationID string
	component     string
}

// NewEnhancedDynamoDBClient creates a new enhanced DynamoDB client
func NewEnhancedDynamoDBClient(client *dynamodb.Client, defaultTable, component string) *EnhancedDynamoDBClient {
	return &EnhancedDynamoDBClient{
		client:       client,
		defaultTable: defaultTable,
		component:    component,
	}
}

// SetCorrelationID sets the correlation ID for tracing
func (c *EnhancedDynamoDBClient) SetCorrelationID(correlationID string) {
	c.correlationID = correlationID
}

// PutItemWithRetry demonstrates enhanced error handling with retry logic
func (c *EnhancedDynamoDBClient) PutItemWithRetry(ctx context.Context, item map[string]types.AttributeValue, tableName string) error {
	if tableName == "" {
		tableName = c.defaultTable
	}

	operation := "PutItem"
	var lastErr *WorkflowError

	for attempt := 0; attempt < 5; attempt++ {
		input := &dynamodb.PutItemInput{
			TableName: &tableName,
			Item:      item,
		}

		_, err := c.client.PutItem(ctx, input)
		if err == nil {
			// Success
			if attempt > 0 {
				log.Printf("PutItem succeeded on attempt %d", attempt+1)
			}
			return nil
		}

		// Analyze the error
		enhancedErr := AnalyzeDynamoDBError(operation, tableName, err).
			WithComponent(c.component).
			WithOperation(operation).
			WithCorrelationID(c.correlationID).
			IncrementRetryCount()

		lastErr = enhancedErr

		// Check if error is retryable
		if !IsDynamoDBRetryableError(enhancedErr) {
			log.Printf("Non-retryable error: %s", enhancedErr.Error())
			return enhancedErr
		}

		// Check retry limit
		if enhancedErr.IsRetryLimitExceeded() {
			log.Printf("Retry limit exceeded for %s", operation)
			return enhancedErr
		}

		// Calculate delay based on retry strategy
		baseDelay := time.Second
		delay := enhancedErr.GetRetryDelay(baseDelay)
		
		log.Printf("Attempt %d failed, retrying in %v: %s", attempt+1, delay, enhancedErr.Error())

		// Wait before retry
		select {
		case <-ctx.Done():
			return NewTimeoutError(operation, time.Since(time.Now())).
				WithComponent(c.component).
				WithTableName(tableName)
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return lastErr
}

// QueryWithErrorAnalysis demonstrates comprehensive error analysis
func (c *EnhancedDynamoDBClient) QueryWithErrorAnalysis(ctx context.Context, input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	operation := "Query"
	tableName := ""
	if input.TableName != nil {
		tableName = *input.TableName
	}

	result, err := c.client.Query(ctx, input)
	if err != nil {
		// Create enhanced error with full context
		enhancedErr := AnalyzeDynamoDBError(operation, tableName, err).
			WithComponent(c.component).
			WithOperation(operation).
			WithCorrelationID(c.correlationID).
			WithStackTrace()

		// Add query-specific context
		if input.IndexName != nil {
			enhancedErr.WithIndexName(*input.IndexName)
		}

		// Log error metrics for monitoring
		metrics := GetErrorMetrics(enhancedErr)
		log.Printf("Query error metrics: %+v", metrics)

		return nil, enhancedErr
	}

	return result, nil
}

// BatchWriteWithErrorHandling demonstrates batch operation error handling
func (c *EnhancedDynamoDBClient) BatchWriteWithErrorHandling(ctx context.Context, requests map[string][]types.WriteRequest) error {
	operation := "BatchWriteItem"
	
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: requests,
	}

	result, err := c.client.BatchWriteItem(ctx, input)
	if err != nil {
		enhancedErr := AnalyzeDynamoDBError(operation, "multiple", err).
			WithComponent(c.component).
			WithOperation(operation).
			WithCorrelationID(c.correlationID)

		return enhancedErr
	}

	// Handle partial failures
	if len(result.UnprocessedItems) > 0 {
		totalItems := 0
		for _, writeRequests := range requests {
			totalItems += len(writeRequests)
		}

		unprocessedCount := 0
		for _, writeRequests := range result.UnprocessedItems {
			unprocessedCount += len(writeRequests)
		}

		batchErr := NewBatchOperationError(operation, "multiple", unprocessedCount, totalItems).
			WithComponent(c.component).
			WithCorrelationID(c.correlationID)

		// Add details about unprocessed items
		batchErr.WithContext("unprocessed_tables", getTableNames(result.UnprocessedItems))

		return batchErr
	}

	return nil
}

// TransactWriteWithEnhancedErrors demonstrates transaction error handling
func (c *EnhancedDynamoDBClient) TransactWriteWithEnhancedErrors(ctx context.Context, items []types.TransactWriteItem) error {
	operation := "TransactWriteItems"
	
	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: items,
	}

	_, err := c.client.TransactWriteItems(ctx, input)
	if err != nil {
		// Check for specific transaction errors
		enhancedErr := AnalyzeDynamoDBError(operation, "transaction", err).
			WithComponent(c.component).
			WithOperation(operation).
			WithCorrelationID(c.correlationID)

		// Add transaction-specific context
		enhancedErr.WithContext("transaction_item_count", len(items))
		enhancedErr.WithContext("transaction_tables", getTransactionTables(items))

		return enhancedErr
	}

	return nil
}

// Example: Error monitoring and alerting
func (c *EnhancedDynamoDBClient) MonitorErrors(errors []error) {
	if len(errors) == 0 {
		return
	}

	// Aggregate errors for analysis
	summary := AggregateErrors(errors)

	// Log summary for monitoring
	log.Printf("Error Summary: Total=%d, Retryable=%d, Critical=%d, MostCommon=%s",
		summary.TotalErrors, summary.RetryableCount, summary.CriticalCount, summary.MostCommon)

	// Check for critical error patterns
	if summary.CriticalCount > 0 {
		log.Printf("ALERT: %d critical errors detected", summary.CriticalCount)
		// Send alert to monitoring system
	}

	// Check for high error rate
	if float64(summary.RetryableCount)/float64(summary.TotalErrors) > 0.8 {
		log.Printf("WARNING: High rate of retryable errors (%d/%d) - possible capacity issue",
			summary.RetryableCount, summary.TotalErrors)
	}

	// Log suggestions for error resolution
	if len(summary.Suggestions) > 0 {
		log.Printf("Error resolution suggestions: %v", summary.Suggestions)
	}
}

// Example: Custom retry logic with circuit breaker pattern
type CircuitBreaker struct {
	failures    int
	maxFailures int
	timeout     time.Duration
	lastFailure time.Time
	state       string // "closed", "open", "half-open"
}

func (c *EnhancedDynamoDBClient) PutItemWithCircuitBreaker(ctx context.Context, item map[string]types.AttributeValue, tableName string, cb *CircuitBreaker) error {
	// Check circuit breaker state
	if cb.state == "open" && time.Since(cb.lastFailure) < cb.timeout {
		return NewDynamoDBError("PutItem", tableName, fmt.Errorf("circuit breaker open")).
			WithComponent(c.component).
			WithSuggestions("Wait for circuit breaker to reset", "Check service health")
	}

	err := c.PutItemWithRetry(ctx, item, tableName)
	if err != nil {
		// Update circuit breaker on failure
		cb.failures++
		cb.lastFailure = time.Now()
		
		if cb.failures >= cb.maxFailures {
			cb.state = "open"
			log.Printf("Circuit breaker opened due to %d failures", cb.failures)
		}
		
		return err
	}

	// Reset circuit breaker on success
	if cb.state == "half-open" || cb.failures > 0 {
		cb.failures = 0
		cb.state = "closed"
		log.Printf("Circuit breaker reset after successful operation")
	}

	return nil
}

// Helper functions
func getTableNames(unprocessedItems map[string][]types.WriteRequest) []string {
	tables := make([]string, 0, len(unprocessedItems))
	for table := range unprocessedItems {
		tables = append(tables, table)
	}
	return tables
}

func getTransactionTables(items []types.TransactWriteItem) []string {
	tableSet := make(map[string]bool)
	
	for _, item := range items {
		if item.Put != nil && item.Put.TableName != nil {
			tableSet[*item.Put.TableName] = true
		}
		if item.Update != nil && item.Update.TableName != nil {
			tableSet[*item.Update.TableName] = true
		}
		if item.Delete != nil && item.Delete.TableName != nil {
			tableSet[*item.Delete.TableName] = true
		}
		if item.ConditionCheck != nil && item.ConditionCheck.TableName != nil {
			tableSet[*item.ConditionCheck.TableName] = true
		}
	}
	
	tables := make([]string, 0, len(tableSet))
	for table := range tableSet {
		tables = append(tables, table)
	}
	return tables
}

// Example usage in a Lambda function
func ExampleLambdaHandler(ctx context.Context, event interface{}, dynamoClient *dynamodb.Client) error {
	// Initialize enhanced client
	client := NewEnhancedDynamoDBClient(dynamoClient, "UserTable", "UserService")
	client.SetCorrelationID("req-123-456")

	// Example item
	item := map[string]types.AttributeValue{
		"userId": &types.AttributeValueMemberS{Value: "user123"},
		"name":   &types.AttributeValueMemberS{Value: "John Doe"},
	}

	// Put item with enhanced error handling
	if err := client.PutItemWithRetry(ctx, item, ""); err != nil {
		// Error is already enhanced with context
		log.Printf("Failed to put item: %s", err.Error())
		
		// Get error metrics for monitoring
		metrics := GetErrorMetrics(err)
		// Send metrics to CloudWatch or other monitoring system
		_ = metrics
		
		return err
	}

	return nil
}
