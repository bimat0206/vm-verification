package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"verification-service/internal/domain/models"
)

// VerificationRepository implements the repository interface for DynamoDB
type VerificationRepository struct {
	client                  *dynamodb.Client
	verificationResultsTable string
	layoutMetadataTable      string
}

// NewVerificationRepository creates a new DynamoDB-based verification repository
func NewVerificationRepository(
	region string,
	verificationResultsTable string,
	layoutMetadataTable string,
) (*VerificationRepository, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create DynamoDB client
	client := dynamodb.NewFromConfig(cfg)

	return &VerificationRepository{
		client:                  client,
		verificationResultsTable: verificationResultsTable,
		layoutMetadataTable:      layoutMetadataTable,
	}, nil
}

// CreateVerification creates a new verification record
func (r *VerificationRepository) CreateVerification(ctx context.Context, verification *models.VerificationContext) error {
	// Convert struct to map
	item, err := attributevalue.MarshalMap(verification)
	if err != nil {
		return fmt.Errorf("failed to marshal verification: %w", err)
	}

	// Store timestamps as ISO 8601 strings for better compatibility
	verificationAt, err := attributevalue.Marshal(verification.VerificationAt.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to marshal verificationAt: %w", err)
	}
	item["verificationAt"] = verificationAt

	// Add storage timestamp
	createdAt, err := attributevalue.Marshal(time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to marshal createdAt: %w", err)
	}
	item["createdAt"] = createdAt

	// Convert turn timestamps
	turnTimestamps := make(map[string]string)
	for k, v := range verification.TurnTimestamps {
		turnTimestamps[k] = v.Format(time.RFC3339)
	}
	timestampsValue, err := attributevalue.Marshal(turnTimestamps)
	if err != nil {
		return fmt.Errorf("failed to marshal turnTimestamps: %w", err)
	}
	item["turnTimestamps"] = timestampsValue

	// Put item in DynamoDB
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.verificationResultsTable),
		Item:      item,
		// Only create if it doesn't exist
		ConditionExpression: aws.String("attribute_not_exists(verificationId)"),
	})

	if err != nil {
		return fmt.Errorf("failed to put verification in DynamoDB: %w", err)
	}

	return nil
}

// GetVerification retrieves a verification by ID
func (r *VerificationRepository) GetVerification(ctx context.Context, id string) (*models.VerificationContext, error) {
	// Get item from DynamoDB
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.verificationResultsTable),
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: id},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get verification from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("verification not found: %s", id)
	}

	// Convert to struct
	verification := &models.VerificationContext{}
	if err := attributevalue.UnmarshalMap(result.Item, verification); err != nil {
		return nil, fmt.Errorf("failed to unmarshal verification: %w", err)
	}

	// Parse timestamps
	if verificationAtStr, ok := result.Item["verificationAt"].(*types.AttributeValueMemberS); ok {
		if verificationAt, err := time.Parse(time.RFC3339, verificationAtStr.Value); err == nil {
			verification.VerificationAt = verificationAt
		}
	}

	// Parse turn timestamps
	if timestampsValue, ok := result.Item["turnTimestamps"].(*types.AttributeValueMemberM); ok {
		verification.TurnTimestamps = make(map[string]time.Time)
		for k, v := range timestampsValue.Value {
			if strValue, ok := v.(*types.AttributeValueMemberS); ok {
				if parsedTime, err := time.Parse(time.RFC3339, strValue.Value); err == nil {
					verification.TurnTimestamps[k] = parsedTime
				}
			}
		}
	}

	return verification, nil
}

// UpdateVerificationStatus updates the status of a verification
func (r *VerificationRepository) UpdateVerificationStatus(ctx context.Context, id string, status models.VerificationStatus) error {
	// Update item in DynamoDB
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.verificationResultsTable),
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression: aws.String("SET #status = :status, updatedAt = :updatedAt"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":    &types.AttributeValueMemberS{Value: string(status)},
			":updatedAt": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to update verification status in DynamoDB: %w", err)
	}

	return nil
}

// StoreReferenceAnalysis stores the reference analysis for a verification
func (r *VerificationRepository) StoreReferenceAnalysis(ctx context.Context, verificationID string, analysis *models.ReferenceAnalysis) error {
	// Convert to JSON to store as a single attribute
	analysisJSON, err := json.Marshal(analysis)
	if err != nil {
		return fmt.Errorf("failed to marshal reference analysis: %w", err)
	}

	// Update item in DynamoDB
	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.verificationResultsTable),
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: verificationID},
		},
		UpdateExpression: aws.String("SET referenceAnalysis = :analysis, updatedAt = :updatedAt"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":analysis":  &types.AttributeValueMemberS{Value: string(analysisJSON)},
			":updatedAt": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to store reference analysis in DynamoDB: %w", err)
	}

	return nil
}

// StoreCheckingAnalysis stores the checking analysis for a verification
func (r *VerificationRepository) StoreCheckingAnalysis(ctx context.Context, verificationID string, analysis *models.CheckingAnalysis) error {
	// Convert to JSON to store as a single attribute
	analysisJSON, err := json.Marshal(analysis)
	if err != nil {
		return fmt.Errorf("failed to marshal checking analysis: %w", err)
	}

	// Update item in DynamoDB
	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.verificationResultsTable),
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: verificationID},
		},
		UpdateExpression: aws.String("SET checkingAnalysis = :analysis, updatedAt = :updatedAt"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":analysis":  &types.AttributeValueMemberS{Value: string(analysisJSON)},
			":updatedAt": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to store checking analysis in DynamoDB: %w", err)
	}

	return nil
}

// StoreFinalResults stores the final results for a verification
func (r *VerificationRepository) StoreFinalResults(ctx context.Context, verificationID string, results *models.VerificationResult) error {
	// Convert to JSON to store as a single attribute
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to marshal final results: %w", err)
	}

	// Update item in DynamoDB
	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.verificationResultsTable),
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: verificationID},
		},
		UpdateExpression: aws.String("SET finalResults = :results, updatedAt = :updatedAt"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":results":   &types.AttributeValueMemberS{Value: string(resultsJSON)},
			":updatedAt": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to store final results in DynamoDB: %w", err)
	}

	return nil
}

// StoreResultImageURL stores the result image URL for a verification
func (r *VerificationRepository) StoreResultImageURL(ctx context.Context, verificationID string, url string) error {
	// Update item in DynamoDB
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.verificationResultsTable),
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: verificationID},
		},
		UpdateExpression: aws.String("SET resultImageUrl = :url, updatedAt = :updatedAt"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":url":       &types.AttributeValueMemberS{Value: url},
			":updatedAt": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to store result image URL in DynamoDB: %w", err)
	}

	return nil
}

// ListVerifications retrieves a list of verifications with filtering
func (r *VerificationRepository) ListVerifications(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.VerificationResult, int, error) {
	// Build query expression
	expressionBuilder := NewQueryExpressionBuilder()

	// Add filters
	if vendingMachineID, ok := filters["vendingMachineId"].(string); ok && vendingMachineID != "" {
		expressionBuilder.AddCondition("vendingMachineId", "=", vendingMachineID)
	}

	if status, ok := filters["verificationStatus"].(string); ok && status != "" {
		// Use GSI2 which has verificationStatus as partition key
		expressionBuilder.SetIndex("GSI2")
		expressionBuilder.AddCondition("verificationStatus", "=", status)
	}

	if fromDate, ok := filters["fromDate"].(string); ok && fromDate != "" {
		expressionBuilder.AddCondition("verificationAt", ">=", fromDate)
	}

	if toDate, ok := filters["toDate"].(string); ok && toDate != "" {
		expressionBuilder.AddCondition("verificationAt", "<=", toDate)
	}

	// Build query input
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(r.verificationResultsTable),
		KeyConditionExpression:    expressionBuilder.BuildKeyCondition(),
		FilterExpression:          expressionBuilder.BuildFilter(),
		ExpressionAttributeNames:  expressionBuilder.GetAttributeNames(),
		ExpressionAttributeValues: expressionBuilder.GetAttributeValues(),
		Limit:                     aws.Int32(int32(limit)),
	}

	// Use index if specified
	if expressionBuilder.Index() != "" {
		queryInput.IndexName = aws.String(expressionBuilder.Index())
	}

	// Apply pagination
	if offset > 0 {
		// In a real implementation, you would use LastEvaluatedKey for pagination
		// For simplicity, we'll skip this for now
	}

	// Execute query
	result, err := r.client.Query(ctx, queryInput)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query verifications: %w", err)
	}

	// Convert results to models
	verifications := make([]*models.VerificationResult, 0, len(result.Items))
	for _, item := range result.Items {
		// Check if finalResults exists
		finalResultsAttr, ok := item["finalResults"]
		if !ok {
			// Skip items without final results
			continue
		}

		var finalResults *models.VerificationResult
		if finalResultsStr, ok := finalResultsAttr.(*types.AttributeValueMemberS); ok {
			if err := json.Unmarshal([]byte(finalResultsStr.Value), &finalResults); err != nil {
				// Log error but continue processing other items
				fmt.Printf("Failed to unmarshal final results for a verification: %v\n", err)
				continue
			}
			verifications = append(verifications, finalResults)
		}
	}

	// Return results and total count
	return verifications, int(result.Count), nil
}

// QueryExpressionBuilder helps build DynamoDB query expressions
type QueryExpressionBuilder struct {
	keyConditions        []string
	filterConditions     []string
	attributeNames       map[string]string
	attributeValues      map[string]types.AttributeValue
	nameCounter          int
	valueCounter         int
	index                string
}

// NewQueryExpressionBuilder creates a new query expression builder
func NewQueryExpressionBuilder() *QueryExpressionBuilder {
	return &QueryExpressionBuilder{
		keyConditions:    make([]string, 0),
		filterConditions: make([]string, 0),
		attributeNames:   make(map[string]string),
		attributeValues:  make(map[string]types.AttributeValue),
		nameCounter:      0,
		valueCounter:     0,
	}
}

// SetIndex sets the GSI to use
func (b *QueryExpressionBuilder) SetIndex(index string) {
	b.index = index
}

// Index returns the current index
func (b *QueryExpressionBuilder) Index() string {
	return b.index
}

// AddCondition adds a condition to the builder
func (b *QueryExpressionBuilder) AddCondition(attribute, operator, value string) {
	nameKey := fmt.Sprintf("#n%d", b.nameCounter)
	valueKey := fmt.Sprintf(":v%d", b.valueCounter)
	
	b.attributeNames[nameKey] = attribute
	b.attributeValues[valueKey] = &types.AttributeValueMemberS{Value: value}
	
	condition := fmt.Sprintf("%s %s %s", nameKey, operator, valueKey)
	
	// Determine if this is a key condition
	if attribute == "verificationId" || attribute == "verificationAt" || 
		(b.index == "GSI1" && (attribute == "layoutId")) || 
		(b.index == "GSI2" && (attribute == "verificationStatus")) {
		b.keyConditions = append(b.keyConditions, condition)
	} else {
		b.filterConditions = append(b.filterConditions, condition)
	}
	
	b.nameCounter++
	b.valueCounter++
}

// BuildKeyCondition builds the key condition expression
func (b *QueryExpressionBuilder) BuildKeyCondition() *string {
	if len(b.keyConditions) == 0 {
		return nil
	}
	
	expr := b.keyConditions[0]
	for _, cond := range b.keyConditions[1:] {
		expr += " AND " + cond
	}
	
	return aws.String(expr)
}

// BuildFilter builds the filter expression
func (b *QueryExpressionBuilder) BuildFilter() *string {
	if len(b.filterConditions) == 0 {
		return nil
	}
	
	expr := b.filterConditions[0]
	for _, cond := range b.filterConditions[1:] {
		expr += " AND " + cond
	}
	
	return aws.String(expr)
}

// GetAttributeNames returns the expression attribute names
func (b *QueryExpressionBuilder) GetAttributeNames() map[string]string {
	return b.attributeNames
}

// GetAttributeValues returns the expression attribute values
func (b *QueryExpressionBuilder) GetAttributeValues() map[string]types.AttributeValue {
	return b.attributeValues
}