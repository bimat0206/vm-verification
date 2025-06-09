package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sirupsen/logrus"
)

var (
	log                   *logrus.Logger
	dynamoClient          *dynamodb.Client
	verificationTableName string
	conversationTableName string
)

// VerificationRecord represents a verification record from DynamoDB
type VerificationRecord struct {
	VerificationID         string                 `json:"verificationId" dynamodbav:"verificationId"`
	VerificationAt         string                 `json:"verificationAt" dynamodbav:"verificationAt"`
	VerificationStatus     string                 `json:"verificationStatus" dynamodbav:"verificationStatus"`
	VerificationType       string                 `json:"verificationType" dynamodbav:"verificationType"`
	VendingMachineID       string                 `json:"vendingMachineId" dynamodbav:"vendingMachineId"`
	ReferenceImageURL      string                 `json:"referenceImageUrl" dynamodbav:"referenceImageUrl"`
	CheckingImageURL       string                 `json:"checkingImageUrl" dynamodbav:"checkingImageUrl"`
	LayoutID               *int                   `json:"layoutId,omitempty" dynamodbav:"layoutId,omitempty"`
	LayoutPrefix           *string                `json:"layoutPrefix,omitempty" dynamodbav:"layoutPrefix,omitempty"`
	OverallAccuracy        *float64               `json:"overallAccuracy,omitempty" dynamodbav:"overallAccuracy,omitempty"`
	CorrectPositions       *int                   `json:"correctPositions,omitempty" dynamodbav:"correctPositions,omitempty"`
	DiscrepantPositions    *int                   `json:"discrepantPositions,omitempty" dynamodbav:"discrepantPositions,omitempty"`
	Result                 map[string]interface{} `json:"result,omitempty" dynamodbav:"result,omitempty"`
	VerificationSummary    map[string]interface{} `json:"verificationSummary,omitempty" dynamodbav:"verificationSummary,omitempty"`
	CreatedAt              string                 `json:"createdAt,omitempty" dynamodbav:"createdAt,omitempty"`
	UpdatedAt              string                 `json:"updatedAt,omitempty" dynamodbav:"updatedAt,omitempty"`
	PreviousVerificationID *string                `json:"previousVerificationId,omitempty" dynamodbav:"previousVerificationId,omitempty"`
}

// ListVerificationsResponse represents the API response structure
type ListVerificationsResponse struct {
	Results    []VerificationRecord `json:"results"`
	Pagination PaginationInfo       `json:"pagination"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Total      int  `json:"total"`
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	NextOffset *int `json:"nextOffset,omitempty"`
}

// QueryParams represents the query parameters for filtering
type QueryParams struct {
	VerificationStatus string
	VendingMachineID   string
	FromDate           string
	ToDate             string
	Limit              int
	Offset             int
	SortBy             string
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func init() {
	// Initialize logger
	log = logrus.New()
	logLevel := os.Getenv("LOG_LEVEL")
	if level, err := logrus.ParseLevel(logLevel); err == nil {
		log.SetLevel(level)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
	log.SetFormatter(&logrus.JSONFormatter{})

	// Load environment variables
	verificationTableName = os.Getenv("DYNAMODB_VERIFICATION_TABLE")
	if verificationTableName == "" {
		log.Fatal("DYNAMODB_VERIFICATION_TABLE environment variable is required")
	}

	conversationTableName = os.Getenv("DYNAMODB_CONVERSATION_TABLE")
	if conversationTableName == "" {
		log.Fatal("DYNAMODB_CONVERSATION_TABLE environment variable is required")
	}

	log.WithFields(logrus.Fields{
		"verificationTable": verificationTableName,
		"conversationTable": conversationTableName,
	}).Info("Environment variables loaded successfully")

	// Initialize AWS DynamoDB client
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.WithError(err).Fatal("unable to load AWS SDK config")
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.WithFields(logrus.Fields{
		"method": request.HTTPMethod,
		"path":   request.Path,
		"params": request.QueryStringParameters,
	}).Info("List verifications request received")

	// Set CORS headers
	headers := map[string]string{
		"Content-Type":                     "application/json",
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
		"Access-Control-Allow-Methods":     "GET,OPTIONS",
	}

	// Handle OPTIONS request for CORS
	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    headers,
			Body:       "",
		}, nil
	}

	// Only allow GET requests
	if request.HTTPMethod != "GET" {
		return createErrorResponse(405, "Method not allowed", "Only GET requests are supported", headers)
	}

	// Parse query parameters
	queryParams, err := parseQueryParams(request.QueryStringParameters)
	if err != nil {
		log.WithError(err).Error("Failed to parse query parameters")
		return createErrorResponse(400, "Invalid query parameters", err.Error(), headers)
	}

	log.WithFields(logrus.Fields{
		"queryParams": queryParams,
	}).Info("Processing list verifications request")

	// Query DynamoDB for verification records
	response, err := queryVerifications(ctx, queryParams)
	if err != nil {
		log.WithError(err).Error("Failed to query verifications")
		return createErrorResponse(500, "Failed to query verifications", err.Error(), headers)
	}

	// Convert response to JSON
	responseBody, err := json.Marshal(response)
	if err != nil {
		log.WithError(err).Error("Failed to marshal response")
		return createErrorResponse(500, "Internal server error", "Failed to process response", headers)
	}

	log.WithFields(logrus.Fields{
		"resultCount": len(response.Results),
		"total":       response.Pagination.Total,
	}).Info("List verifications request completed successfully")

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(responseBody),
	}, nil
}

func parseQueryParams(params map[string]string) (*QueryParams, error) {
	queryParams := &QueryParams{
		Limit:  20,                    // default limit
		Offset: 0,                     // default offset
		SortBy: "verificationAt:desc", // default sort
	}

	if params == nil {
		return queryParams, nil
	}

	// Parse verification status
	if status := params["verificationStatus"]; status != "" {
		if status != "CORRECT" && status != "INCORRECT" {
			return nil, fmt.Errorf("invalid verificationStatus: must be CORRECT or INCORRECT")
		}
		queryParams.VerificationStatus = status
	}

	// Parse vending machine ID
	if machineID := params["vendingMachineId"]; machineID != "" {
		queryParams.VendingMachineID = machineID
	}

	// Parse date range
	if fromDate := params["fromDate"]; fromDate != "" {
		if _, err := time.Parse(time.RFC3339, fromDate); err != nil {
			return nil, fmt.Errorf("invalid fromDate format: must be RFC3339")
		}
		queryParams.FromDate = fromDate
	}

	if toDate := params["toDate"]; toDate != "" {
		if _, err := time.Parse(time.RFC3339, toDate); err != nil {
			return nil, fmt.Errorf("invalid toDate format: must be RFC3339")
		}
		queryParams.ToDate = toDate
	}

	// Parse limit
	if limitStr := params["limit"]; limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			return nil, fmt.Errorf("invalid limit: must be between 1 and 100")
		}
		queryParams.Limit = limit
	}

	// Parse offset
	if offsetStr := params["offset"]; offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return nil, fmt.Errorf("invalid offset: must be non-negative integer")
		}
		queryParams.Offset = offset
	}

	// Parse sort by
	if sortBy := params["sortBy"]; sortBy != "" {
		validSorts := map[string]bool{
			"verificationAt:desc":  true,
			"verificationAt:asc":   true,
			"overallAccuracy:desc": true,
			"overallAccuracy:asc":  true,
		}
		if !validSorts[sortBy] {
			return nil, fmt.Errorf("invalid sortBy: must be one of verificationAt:desc, verificationAt:asc, overallAccuracy:desc, overallAccuracy:asc")
		}
		queryParams.SortBy = sortBy
	}

	return queryParams, nil
}

func createErrorResponse(statusCode int, error, message string, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	errorResp := ErrorResponse{
		Error:   error,
		Message: message,
		Code:    fmt.Sprintf("HTTP_%d", statusCode),
	}

	body, _ := json.Marshal(errorResp)

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       string(body),
	}, nil
}

func queryVerifications(ctx context.Context, params *QueryParams) (*ListVerificationsResponse, error) {
	var items []map[string]types.AttributeValue
	var err error

	// Determine which query strategy to use based on filters
	if params.VerificationStatus != "" {
		// Use VerificationStatusIndex GSI for status-based queries
		items, err = queryByStatus(ctx, params)
	} else {
		// Use scan for general queries (less efficient but more flexible)
		items, err = scanTable(ctx, params)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query DynamoDB: %w", err)
	}

	// Convert DynamoDB items to VerificationRecord structs
	var records []VerificationRecord
	for _, item := range items {
		var record VerificationRecord
		err := attributevalue.UnmarshalMap(item, &record)
		if err != nil {
			log.WithError(err).Warn("Failed to unmarshal verification record")
			continue
		}
		records = append(records, record)
	}

	// Apply additional filtering that couldn't be done in DynamoDB
	records = applyAdditionalFilters(records, params)

	// Sort records
	records = sortRecords(records, params.SortBy)

	// Calculate total count (for pagination)
	total := len(records)

	// Apply pagination
	start := params.Offset
	end := start + params.Limit
	if start > len(records) {
		start = len(records)
	}
	if end > len(records) {
		end = len(records)
	}

	paginatedRecords := records[start:end]

	// Calculate next offset
	var nextOffset *int
	if end < len(records) {
		next := end
		nextOffset = &next
	}

	return &ListVerificationsResponse{
		Results: paginatedRecords,
		Pagination: PaginationInfo{
			Total:      total,
			Limit:      params.Limit,
			Offset:     params.Offset,
			NextOffset: nextOffset,
		},
	}, nil
}

func queryByStatus(ctx context.Context, params *QueryParams) ([]map[string]types.AttributeValue, error) {
	// Build key condition for GSI query
	keyCondition := "verificationStatus = :status"
	expressionAttributeValues := map[string]types.AttributeValue{
		":status": &types.AttributeValueMemberS{Value: params.VerificationStatus},
	}

	// Add date range filter if specified
	if params.FromDate != "" || params.ToDate != "" {
		if params.FromDate != "" && params.ToDate != "" {
			keyCondition += " AND verificationAt BETWEEN :fromDate AND :toDate"
			expressionAttributeValues[":fromDate"] = &types.AttributeValueMemberS{Value: params.FromDate}
			expressionAttributeValues[":toDate"] = &types.AttributeValueMemberS{Value: params.ToDate}
		} else if params.FromDate != "" {
			keyCondition += " AND verificationAt >= :fromDate"
			expressionAttributeValues[":fromDate"] = &types.AttributeValueMemberS{Value: params.FromDate}
		} else if params.ToDate != "" {
			keyCondition += " AND verificationAt <= :toDate"
			expressionAttributeValues[":toDate"] = &types.AttributeValueMemberS{Value: params.ToDate}
		}
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(verificationTableName),
		IndexName:                 aws.String("VerificationStatusIndex"),
		KeyConditionExpression:    aws.String(keyCondition),
		ExpressionAttributeValues: expressionAttributeValues,
		ScanIndexForward:          aws.Bool(false), // Default to descending order
	}

	// Add filter expression for vending machine ID if specified
	if params.VendingMachineID != "" {
		input.FilterExpression = aws.String("vendingMachineId = :machineId")
		expressionAttributeValues[":machineId"] = &types.AttributeValueMemberS{Value: params.VendingMachineID}
		input.ExpressionAttributeValues = expressionAttributeValues
	}

	log.WithFields(logrus.Fields{
		"indexName":    "VerificationStatusIndex",
		"keyCondition": keyCondition,
		"filterExpr":   input.FilterExpression,
	}).Debug("Querying DynamoDB with GSI")

	result, err := dynamoClient.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query VerificationStatusIndex: %w", err)
	}

	return result.Items, nil
}

func scanTable(ctx context.Context, params *QueryParams) ([]map[string]types.AttributeValue, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(verificationTableName),
	}

	var filterExpressions []string
	expressionAttributeValues := make(map[string]types.AttributeValue)

	// Add filters
	if params.VendingMachineID != "" {
		filterExpressions = append(filterExpressions, "vendingMachineId = :machineId")
		expressionAttributeValues[":machineId"] = &types.AttributeValueMemberS{Value: params.VendingMachineID}
	}

	if params.FromDate != "" {
		filterExpressions = append(filterExpressions, "verificationAt >= :fromDate")
		expressionAttributeValues[":fromDate"] = &types.AttributeValueMemberS{Value: params.FromDate}
	}

	if params.ToDate != "" {
		filterExpressions = append(filterExpressions, "verificationAt <= :toDate")
		expressionAttributeValues[":toDate"] = &types.AttributeValueMemberS{Value: params.ToDate}
	}

	if len(filterExpressions) > 0 {
		input.FilterExpression = aws.String(strings.Join(filterExpressions, " AND "))
		input.ExpressionAttributeValues = expressionAttributeValues
	}

	log.WithFields(logrus.Fields{
		"filterExpr": input.FilterExpression,
	}).Debug("Scanning DynamoDB table")

	result, err := dynamoClient.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan table: %w", err)
	}

	return result.Items, nil
}

func applyAdditionalFilters(records []VerificationRecord, params *QueryParams) []VerificationRecord {
	// This function applies filters that couldn't be efficiently done in DynamoDB
	// Currently, all major filters are handled in DynamoDB queries
	return records
}

func sortRecords(records []VerificationRecord, sortBy string) []VerificationRecord {
	switch sortBy {
	case "verificationAt:desc":
		// DynamoDB already returns in desc order by default, but let's ensure it
		sort.Slice(records, func(i, j int) bool {
			return records[i].VerificationAt > records[j].VerificationAt
		})
	case "verificationAt:asc":
		sort.Slice(records, func(i, j int) bool {
			return records[i].VerificationAt < records[j].VerificationAt
		})
	case "overallAccuracy:desc":
		sort.Slice(records, func(i, j int) bool {
			// Handle nil values - put them at the end
			if records[i].OverallAccuracy == nil && records[j].OverallAccuracy == nil {
				return false
			}
			if records[i].OverallAccuracy == nil {
				return false
			}
			if records[j].OverallAccuracy == nil {
				return true
			}
			return *records[i].OverallAccuracy > *records[j].OverallAccuracy
		})
	case "overallAccuracy:asc":
		sort.Slice(records, func(i, j int) bool {
			// Handle nil values - put them at the end
			if records[i].OverallAccuracy == nil && records[j].OverallAccuracy == nil {
				return false
			}
			if records[i].OverallAccuracy == nil {
				return false
			}
			if records[j].OverallAccuracy == nil {
				return true
			}
			return *records[i].OverallAccuracy < *records[j].OverallAccuracy
		})
	default:
		// Default to verificationAt:desc
		sort.Slice(records, func(i, j int) bool {
			return records[i].VerificationAt > records[j].VerificationAt
		})
	}

	return records
}

func main() {
	lambda.Start(handler)
}
