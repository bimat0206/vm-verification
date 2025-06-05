package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
)

var (
	log                     *logrus.Logger
	dynamoClient           *dynamodb.Client
	s3Client               *s3.Client
	conversationTableName  string
	resultsBucketName      string
)

// MetadataPath represents a path structure in DynamoDB metadata
type MetadataPath struct {
	S string `json:"S" dynamodbav:"S"`
}

// ConversationMetadata represents the metadata field structure from DynamoDB
type ConversationMetadata struct {
	Turn1ProcessedPath *MetadataPath `json:"turn1ProcessedPath,omitempty" dynamodbav:"turn1ProcessedPath,omitempty"`
	Turn2ProcessedPath *MetadataPath `json:"turn2ProcessedPath,omitempty" dynamodbav:"turn2ProcessedPath,omitempty"`
}

// ConversationRecord represents a conversation record from DynamoDB
type ConversationRecord struct {
	VerificationID     string                `json:"verificationId" dynamodbav:"verificationId"`
	ConversationID     string                `json:"conversationId" dynamodbav:"conversationId"`
	Metadata           *ConversationMetadata `json:"metadata,omitempty" dynamodbav:"metadata,omitempty"`
	Turn2ProcessedPath string                `json:"turn2ProcessedPath,omitempty" dynamodbav:"turn2ProcessedPath,omitempty"` // Keep for backward compatibility
	CreatedAt          string                `json:"createdAt,omitempty" dynamodbav:"createdAt,omitempty"`
	UpdatedAt          string                `json:"updatedAt,omitempty" dynamodbav:"updatedAt,omitempty"`
	// Extracted paths for easier access
	Turn1ProcessedPathValue string `json:"-" dynamodbav:"-"`
	Turn2ProcessedPathValue string `json:"-" dynamodbav:"-"`
}

// ConversationResponse represents the API response structure
type ConversationResponse struct {
	VerificationID string `json:"verificationId"`
	Content        string `json:"content"`
	ContentType    string `json:"contentType"`
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
	conversationTableName = os.Getenv("DYNAMODB_CONVERSATION_TABLE")
	if conversationTableName == "" {
		log.Fatal("DYNAMODB_CONVERSATION_TABLE environment variable is required")
	}

	resultsBucketName = os.Getenv("RESULTS_BUCKET")
	if resultsBucketName == "" {
		log.Fatal("RESULTS_BUCKET environment variable is required")
	}

	// Initialize AWS clients
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.WithError(err).Fatal("unable to load AWS SDK config")
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	s3Client = s3.NewFromConfig(cfg)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.WithFields(logrus.Fields{
		"method": request.HTTPMethod,
		"path":   request.Path,
		"params": request.PathParameters,
	}).Info("Get conversation request received")

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

	// Extract verificationId from path parameters
	verificationId := request.PathParameters["verificationId"]
	if verificationId == "" {
		return createErrorResponse(400, "Missing parameter", "verificationId path parameter is required", headers)
	}

	log.WithFields(logrus.Fields{
		"verificationId": verificationId,
	}).Info("Processing get conversation request")

	// Query DynamoDB for conversation record
	conversationRecord, err := getConversationRecord(ctx, verificationId)
	if err != nil {
		log.WithError(err).Error("Failed to get conversation record")
		return createErrorResponse(500, "Failed to get conversation record", err.Error(), headers)
	}

	if conversationRecord == nil {
		return createErrorResponse(404, "Conversation not found", fmt.Sprintf("No conversation found for verificationId: %s", verificationId), headers)
	}

	// Determine which path to use for S3 content retrieval
	s3Path := conversationRecord.Turn2ProcessedPathValue
	if s3Path == "" {
		// Fallback to the direct field for backward compatibility
		s3Path = conversationRecord.Turn2ProcessedPath
	}

	if s3Path == "" {
		log.WithFields(logrus.Fields{
			"verificationId": verificationId,
		}).Error("No turn2ProcessedPath found in conversation record")
		return createErrorResponse(404, "Content not found", "No processed conversation content path found", headers)
	}

	// Retrieve content from S3
	content, err := getS3Content(ctx, s3Path)
	if err != nil {
		log.WithError(err).Error("Failed to retrieve S3 content")
		return createErrorResponse(500, "Failed to retrieve conversation content", err.Error(), headers)
	}

	// Create response
	response := ConversationResponse{
		VerificationID: verificationId,
		Content:        content,
		ContentType:    "text/markdown",
	}

	// Convert response to JSON
	responseBody, err := json.Marshal(response)
	if err != nil {
		log.WithError(err).Error("Failed to marshal response")
		return createErrorResponse(500, "Internal server error", "Failed to process response", headers)
	}

	log.WithFields(logrus.Fields{
		"verificationId": verificationId,
		"contentLength":  len(content),
	}).Info("Get conversation request completed successfully")

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

func getConversationRecord(ctx context.Context, verificationId string) (*ConversationRecord, error) {
	log.WithFields(logrus.Fields{
		"verificationId": verificationId,
		"tableName":      conversationTableName,
	}).Debug("Querying DynamoDB for conversation record")

	// Query DynamoDB using verificationId as the key
	// Assuming verificationId is the primary key or there's a GSI for it
	input := &dynamodb.QueryInput{
		TableName:              aws.String(conversationTableName),
		KeyConditionExpression: aws.String("verificationId = :verificationId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":verificationId": &types.AttributeValueMemberS{Value: verificationId},
		},
		Limit: aws.Int32(1), // We only need one record
	}

	result, err := dynamoClient.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversation table: %w", err)
	}

	if len(result.Items) == 0 {
		log.WithFields(logrus.Fields{
			"verificationId": verificationId,
		}).Info("No conversation record found")
		return nil, nil
	}

	// Unmarshal the first item
	var record ConversationRecord
	err = attributevalue.UnmarshalMap(result.Items[0], &record)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversation record: %w", err)
	}

	// Extract paths from metadata if available
	if record.Metadata != nil {
		if record.Metadata.Turn1ProcessedPath != nil {
			record.Turn1ProcessedPathValue = record.Metadata.Turn1ProcessedPath.S
		}
		if record.Metadata.Turn2ProcessedPath != nil {
			record.Turn2ProcessedPathValue = record.Metadata.Turn2ProcessedPath.S
		}
	}

	log.WithFields(logrus.Fields{
		"verificationId":          record.VerificationID,
		"conversationId":          record.ConversationID,
		"turn1ProcessedPath":      record.Turn1ProcessedPathValue,
		"turn2ProcessedPath":      record.Turn2ProcessedPathValue,
		"legacyTurn2ProcessedPath": record.Turn2ProcessedPath,
		"hasMetadata":             record.Metadata != nil,
	}).Debug("Successfully retrieved conversation record")

	return &record, nil
}

func getS3Content(ctx context.Context, s3Path string) (string, error) {
	if s3Path == "" {
		return "", fmt.Errorf("s3 path is empty")
	}

	log.WithFields(logrus.Fields{
		"s3Path": s3Path,
		"bucket": resultsBucketName,
	}).Debug("Retrieving content from S3")

	// Parse S3 path to extract the key
	// The path might be in format: s3://bucket/key or just the key
	key := s3Path
	if strings.HasPrefix(s3Path, "s3://") {
		// Extract key from full S3 URI
		parts := strings.SplitN(strings.TrimPrefix(s3Path, "s3://"), "/", 2)
		if len(parts) == 2 {
			key = parts[1]
		}
	}

	// Get object from S3
	input := &s3.GetObjectInput{
		Bucket: aws.String(resultsBucketName),
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
