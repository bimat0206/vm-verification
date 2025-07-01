package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/logging"
	"github.com/sirupsen/logrus"
)

// Global variables for AWS clients and configuration
var (
	log                          *logrus.Logger
	dynamoClient                 *dynamodb.Client
	s3Client                     *s3.Client
	verificationTableName        string
	conversationTableName        string
	stateBucket                  string
	stepFunctionsStateMachineArn string
	awsRegion                    string
)

// VerificationRecord represents a verification record from DynamoDB verification table
type VerificationRecord struct {
	VerificationID       string                 `json:"verificationId" dynamodbav:"verificationId"`
	VerificationAt       string                 `json:"verificationAt" dynamodbav:"verificationAt"`
	VerificationStatus   string                 `json:"verificationStatus" dynamodbav:"verificationStatus"`
	VerificationType     string                 `json:"verificationType" dynamodbav:"verificationType"`
	VendingMachineID     string                 `json:"vendingMachineId" dynamodbav:"vendingMachineId"`
	ReferenceImageURL    string                 `json:"referenceImageUrl" dynamodbav:"referenceImageUrl"`
	CheckingImageURL     string                 `json:"checkingImageUrl" dynamodbav:"checkingImageUrl"`
	LayoutID             *int                   `json:"layoutId,omitempty" dynamodbav:"layoutId,omitempty"`
	LayoutPrefix         *string                `json:"layoutPrefix,omitempty" dynamodbav:"layoutPrefix,omitempty"`
	OverallAccuracy      *float64               `json:"overallAccuracy,omitempty" dynamodbav:"overallAccuracy,omitempty"`
	CorrectPositions     *int                   `json:"correctPositions,omitempty" dynamodbav:"correctPositions,omitempty"`
	DiscrepantPositions  *int                   `json:"discrepantPositions,omitempty" dynamodbav:"discrepantPositions,omitempty"`
	Result               map[string]interface{} `json:"result,omitempty" dynamodbav:"result,omitempty"`
	VerificationSummary  map[string]interface{} `json:"verificationSummary,omitempty" dynamodbav:"verificationSummary,omitempty"`
	CreatedAt            string                 `json:"createdAt,omitempty" dynamodbav:"createdAt,omitempty"`
	UpdatedAt            string                 `json:"updatedAt,omitempty" dynamodbav:"updatedAt,omitempty"`
	CurrentStatus        string                 `json:"currentStatus,omitempty" dynamodbav:"currentStatus,omitempty"`
	Turn1ProcessedPath   string                 `json:"turn1ProcessedPath,omitempty" dynamodbav:"turn1ProcessedPath,omitempty"`
	Turn2ProcessedPath   string                 `json:"turn2ProcessedPath,omitempty" dynamodbav:"turn2ProcessedPath,omitempty"`
}

// S3References represents S3 file references
type S3References struct {
	Turn1Processed string `json:"turn1Processed,omitempty"`
	Turn2Processed string `json:"turn2Processed,omitempty"`
}

// StatusSummary represents verification summary information
type StatusSummary struct {
	Message             string   `json:"message"`
	VerificationAt      string   `json:"verificationAt"`
	VerificationStatus  string   `json:"verificationStatus"`
	OverallAccuracy     *float64 `json:"overallAccuracy,omitempty"`
	CorrectPositions    *int     `json:"correctPositions,omitempty"`
	DiscrepantPositions *int     `json:"discrepantPositions,omitempty"`
}

// StatusResponse represents the API response structure
type StatusResponse struct {
	VerificationID      string                 `json:"verificationId"`
	Status              string                 `json:"status"`
	CurrentStatus       string                 `json:"currentStatus,omitempty"`
	VerificationStatus  string                 `json:"verificationStatus,omitempty"`
	S3References        S3References           `json:"s3References"`
	Summary             StatusSummary          `json:"summary"`
	LLMResponse         string                 `json:"llmResponse,omitempty"`
	LLMAnalysis         string                 `json:"llmAnalysis,omitempty"`
	VerificationSummary map[string]interface{} `json:"verificationSummary,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// Custom logger adapter for AWS SDK
type awsLogAdapter struct {
	logger *logrus.Logger
}

func (a *awsLogAdapter) Logf(classification logging.Classification, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	switch classification {
	case logging.Debug:
		a.logger.Debug(msg)
	case logging.Warn:
		a.logger.Warn(msg)
	default:
		a.logger.Info(msg)
	}
}

// Initialize logger with proper configuration
func initLogger() {
	log = logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		log.WithError(err).Warn("Invalid log level, defaulting to INFO")
		level = logrus.InfoLevel
	}
	log.SetLevel(level)
}

// Load and validate environment variables
func loadEnvironmentVariables() error {
	verificationTableName = os.Getenv("DYNAMODB_VERIFICATION_TABLE")
	if verificationTableName == "" {
		return fmt.Errorf("DYNAMODB_VERIFICATION_TABLE environment variable is required")
	}

	conversationTableName = os.Getenv("DYNAMODB_CONVERSATION_TABLE")
	if conversationTableName == "" {
		return fmt.Errorf("DYNAMODB_CONVERSATION_TABLE environment variable is required")
	}

	stateBucket = os.Getenv("STATE_BUCKET")
	if stateBucket == "" {
		log.Warn("STATE_BUCKET environment variable is not set")
	}

	stepFunctionsStateMachineArn = os.Getenv("STEP_FUNCTIONS_STATE_MACHINE_ARN")
	if stepFunctionsStateMachineArn == "" {
		log.Warn("STEP_FUNCTIONS_STATE_MACHINE_ARN environment variable is not set")
	}

	// Use AWS_REGION as primary, fall back to REGION if not set
	awsRegion = os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = os.Getenv("REGION")
	}
	if awsRegion == "" {
		awsRegion = os.Getenv("AWS_DEFAULT_REGION")
	}
	if awsRegion == "" {
		// Default to us-east-1 if no region is specified
		awsRegion = "us-east-1"
		log.Warn("No AWS region specified, defaulting to us-east-1")
	}

	log.WithFields(logrus.Fields{
		"verificationTable":         verificationTableName,
		"conversationTable":         conversationTableName,
		"stateBucket":               stateBucket,
		"stepFunctionsStateMachine": stepFunctionsStateMachineArn,
		"awsRegion":                 awsRegion,
	}).Info("Environment variables loaded")

	return nil
}

// Initialize AWS clients with proper error handling
func initAWSClients(ctx context.Context) error {
	// Create custom HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create a custom endpoint resolver that properly handles DynamoDB endpoints
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		// Force explicit endpoints for services to avoid resolution issues
		switch service {
		case dynamodb.ServiceID:
			return aws.Endpoint{
				URL:           fmt.Sprintf("https://dynamodb.%s.amazonaws.com", region),
				SigningRegion: region,
				PartitionID:   "aws",
				HostnameImmutable: true,
			}, nil
		case s3.ServiceID:
			return aws.Endpoint{
				URL:           fmt.Sprintf("https://s3.%s.amazonaws.com", region),
				SigningRegion: region,
				PartitionID:   "aws",
				HostnameImmutable: true,
			}, nil
		default:
			// Fallback to default endpoint resolution
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		}
	})

	// Load AWS configuration with custom resolver and minimal options
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(awsRegion),
		config.WithHTTPClient(httpClient),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithRetryMaxAttempts(3),
		config.WithLogger(&awsLogAdapter{logger: log}),
		config.WithClientLogMode(aws.LogRetries|aws.LogRequestWithBody|aws.LogResponseWithBody),
	)
	if err != nil {
		return fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	// Initialize DynamoDB client
	dynamoClient = dynamodb.NewFromConfig(cfg)

	// Initialize S3 client
	s3Client = s3.NewFromConfig(cfg)

	// Test DynamoDB connectivity
	testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	_, err = dynamoClient.DescribeTable(testCtx, &dynamodb.DescribeTableInput{
		TableName: aws.String(verificationTableName),
	})
	if err != nil {
		log.WithError(err).Warn("Failed to describe DynamoDB table - endpoint may not be properly configured")
	} else {
		log.Info("Successfully connected to DynamoDB")
	}

	log.Info("AWS clients initialized successfully")
	return nil
}

func init() {
	// Initialize logger first
	initLogger()

	// Load environment variables
	if err := loadEnvironmentVariables(); err != nil {
		log.WithError(err).Fatal("Failed to load environment variables")
	}

	// Initialize AWS clients
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := initAWSClients(ctx); err != nil {
		log.WithError(err).Fatal("Failed to initialize AWS clients")
	}
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.WithFields(logrus.Fields{
		"method":    request.HTTPMethod,
		"path":      request.Path,
		"params":    request.PathParameters,
		"requestId": request.RequestContext.RequestID,
	}).Info("Get verification status request received")

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

	// Extract verificationId from path parameters or query parameters
	verificationId := request.PathParameters["verificationId"]
	if verificationId == "" {
		verificationId = request.QueryStringParameters["verificationId"]
	}
	if verificationId == "" {
		return createErrorResponse(400, "Missing parameter", "verificationId path or query parameter is required", headers)
	}

	log.WithFields(logrus.Fields{
		"verificationId": verificationId,
	}).Info("Processing get verification status request")

	// Query DynamoDB verification table with timeout
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	verificationRecord, err := getVerificationRecord(queryCtx, verificationId)
	if err != nil {
		log.WithError(err).Error("Failed to get verification record")
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return createErrorResponse(504, "Gateway timeout", "Request timed out while querying database", headers)
		}
		return createErrorResponse(500, "Failed to get verification record", err.Error(), headers)
	}

	if verificationRecord == nil {
		return createErrorResponse(404, "Verification not found", fmt.Sprintf("No verification found for verificationId: %s", verificationId), headers)
	}

	// Build response
	response, err := buildStatusResponse(ctx, verificationRecord)
	if err != nil {
		log.WithError(err).Error("Failed to build status response")
		return createErrorResponse(500, "Failed to build status response", err.Error(), headers)
	}

	// Convert response to JSON
	responseBody, err := json.Marshal(response)
	if err != nil {
		log.WithError(err).Error("Failed to marshal response")
		return createErrorResponse(500, "Internal server error", "Failed to process response", headers)
	}

	log.WithFields(logrus.Fields{
		"verificationId":     verificationId,
		"status":             response.Status,
		"currentStatus":      response.CurrentStatus,
		"verificationStatus": response.VerificationStatus,
	}).Info("Get verification status request completed successfully")

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(responseBody),
	}, nil
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

func getVerificationRecord(ctx context.Context, verificationId string) (*VerificationRecord, error) {
	log.WithFields(logrus.Fields{
		"verificationId": verificationId,
		"tableName":      verificationTableName,
		"region":         awsRegion,
	}).Debug("Querying DynamoDB for verification record")

	// Query DynamoDB using verificationId as the hash key
	input := &dynamodb.QueryInput{
		TableName:              aws.String(verificationTableName),
		KeyConditionExpression: aws.String("verificationId = :verificationId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":verificationId": &types.AttributeValueMemberS{Value: verificationId},
		},
		ScanIndexForward: aws.Bool(false), // Sort by verificationAt in descending order
		Limit:            aws.Int32(1),     // Only get the most recent record
	}

	// Execute query with retry logic
	var result *dynamodb.QueryOutput
	var err error
	
	for i := 0; i < 3; i++ {
		result, err = dynamoClient.Query(ctx, input)
		if err == nil {
			break
		}
		
		// Log the error and retry if it's a retryable error
		log.WithError(err).WithField("attempt", i+1).Warn("DynamoDB query failed, retrying...")
		
		// Check if error is retryable
		if !isRetryableError(err) {
			break
		}
		
		// Wait before retry with exponential backoff
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(i+1) * time.Second):
			continue
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to query verification table after retries: %w", err)
	}

	if len(result.Items) == 0 {
		log.WithFields(logrus.Fields{
			"verificationId": verificationId,
		}).Info("No verification record found")
		return nil, nil
	}

	// Unmarshal the first item
	var record VerificationRecord
	err = attributevalue.UnmarshalMap(result.Items[0], &record)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal verification record: %w", err)
	}

	log.WithFields(logrus.Fields{
		"verificationId":     record.VerificationID,
		"verificationAt":     record.VerificationAt,
		"currentStatus":      record.CurrentStatus,
		"verificationStatus": record.VerificationStatus,
		"turn1ProcessedPath": record.Turn1ProcessedPath,
		"turn2ProcessedPath": record.Turn2ProcessedPath,
	}).Debug("Successfully retrieved verification record")

	return &record, nil
}

// Helper function to check if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for common retryable error patterns
	errStr := err.Error()
	retryablePatterns := []string{
		"ResolveEndpointV2",
		"connection refused",
		"timeout",
		"TooManyRequestsException",
		"ServiceUnavailable",
		"RequestLimitExceeded",
		"ThrottlingException",
		"ProvisionedThroughputExceededException",
		"InternalServerError",
	}
	
	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	
	return false
}

func buildStatusResponse(ctx context.Context, record *VerificationRecord) (*StatusResponse, error) {
	// Determine overall status based on currentStatus and verificationStatus
	status := determineOverallStatus(record.CurrentStatus, record.VerificationStatus)

	// Build S3 references
	s3References := S3References{
		Turn1Processed: record.Turn1ProcessedPath,
		Turn2Processed: record.Turn2ProcessedPath,
	}

	// Build summary
	summary := StatusSummary{
		VerificationAt:      record.VerificationAt,
		VerificationStatus:  record.VerificationStatus,
		OverallAccuracy:     record.OverallAccuracy,
		CorrectPositions:    record.CorrectPositions,
		DiscrepantPositions: record.DiscrepantPositions,
	}

	// Set summary message based on status
	switch status {
	case "COMPLETED":
		if record.VerificationStatus == "CORRECT" {
			summary.Message = "Verification completed successfully - No discrepancies found"
		} else if record.VerificationStatus == "INCORRECT" {
			summary.Message = "Verification completed - Discrepancies detected"
		} else {
			summary.Message = "Verification completed"
		}
	case "RUNNING":
		summary.Message = "Verification is currently in progress"
	case "FAILED":
		summary.Message = "Verification failed during processing"
	default:
		summary.Message = "Verification status unknown"
	}

	response := &StatusResponse{
		VerificationID:      record.VerificationID,
		Status:              status,
		CurrentStatus:       record.CurrentStatus,
		VerificationStatus:  record.VerificationStatus,
		S3References:        s3References,
		Summary:             summary,
		VerificationSummary: record.VerificationSummary,
	}

	// If verification is completed, retrieve LLM response from S3
	if status == "COMPLETED" && record.Turn2ProcessedPath != "" {
		s3Ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		
		llmResponse, err := getS3Content(s3Ctx, record.Turn2ProcessedPath)
		if err != nil {
			log.WithError(err).Warn("Failed to retrieve LLM response from S3")
		} else {
			response.LLMResponse = llmResponse
			response.LLMAnalysis = llmResponse // Also populate llmAnalysis field for frontend compatibility
		}
	}

	return response, nil
}

func determineOverallStatus(currentStatus, verificationStatus string) string {
	// Map currentStatus to overall status
	switch currentStatus {
	case "COMPLETED":
		return "COMPLETED"
	case "RUNNING", "TURN1_COMPLETED", "TURN2_RUNNING", "PROCESSING":
		return "RUNNING"
	case "FAILED", "ERROR":
		return "FAILED"
	default:
		// If currentStatus is empty or unknown, check verificationStatus
		if verificationStatus == "PENDING" || verificationStatus == "" {
			return "RUNNING"
		}
		return "COMPLETED"
	}
}

func getS3Content(ctx context.Context, s3Path string) (string, error) {
	if s3Path == "" {
		return "", fmt.Errorf("s3 path is empty")
	}

	var bucket, key string

	// Parse S3 path to extract bucket and key
	if strings.HasPrefix(s3Path, "s3://") {
		parts := strings.SplitN(strings.TrimPrefix(s3Path, "s3://"), "/", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid S3 path format: %s", s3Path)
		}
		bucket = parts[0]
		key = parts[1]
	} else {
		return "", fmt.Errorf("S3 path must start with s3://: %s", s3Path)
	}

	log.WithFields(logrus.Fields{
		"s3Path": s3Path,
		"bucket": bucket,
		"key":    key,
	}).Debug("Retrieving content from S3")

	// Get object from S3
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := s3Client.GetObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	// Read the content
	content, err := io.ReadAll(result.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read S3 object content: %w", err)
	}

	log.WithFields(logrus.Fields{
		"s3Path":        s3Path,
		"contentLength": len(content),
	}).Debug("Successfully retrieved S3 content")

	return string(content), nil
}

func main() {
	lambda.Start(handler)
}