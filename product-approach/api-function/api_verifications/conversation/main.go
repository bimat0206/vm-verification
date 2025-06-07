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
	verificationTableName  string
	conversationTableName  string
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
	// Processed paths from verification metadata
	Turn1ProcessedPath   string                 `json:"turn1ProcessedPath,omitempty" dynamodbav:"turn1ProcessedPath,omitempty"`
	Turn2ProcessedPath   string                 `json:"turn2ProcessedPath,omitempty" dynamodbav:"turn2ProcessedPath,omitempty"`
}

// ConversationContent represents content from a single turn
type ConversationContent struct {
	Turn        int    `json:"turn"`
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
	S3Path      string `json:"s3Path"`
}

// ConversationResponse represents the API response structure
type ConversationResponse struct {
	VerificationID string                `json:"verificationId"`
	Turn1Content   *ConversationContent  `json:"turn1Content,omitempty"`
	Turn2Content   *ConversationContent  `json:"turn2Content,omitempty"`
	Contents       []ConversationContent `json:"contents"`
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

	// Query DynamoDB verification table for processed paths
	verificationRecord, err := getVerificationRecord(ctx, verificationId)
	if err != nil {
		log.WithError(err).Error("Failed to get verification record")
		return createErrorResponse(500, "Failed to get verification record", err.Error(), headers)
	}

	if verificationRecord == nil {
		return createErrorResponse(404, "Verification not found", fmt.Sprintf("No verification found for verificationId: %s", verificationId), headers)
	}

	// Prepare response structure
	response := ConversationResponse{
		VerificationID: verificationId,
		Contents:       []ConversationContent{},
	}

	// Retrieve Turn1 content if path exists
	if verificationRecord.Turn1ProcessedPath != "" {
		turn1Content, err := getS3Content(ctx, verificationRecord.Turn1ProcessedPath)
		if err != nil {
			log.WithError(err).Warn("Failed to retrieve Turn1 S3 content")
		} else {
			turn1 := ConversationContent{
				Turn:        1,
				Content:     turn1Content,
				ContentType: "text/markdown",
				S3Path:      verificationRecord.Turn1ProcessedPath,
			}
			response.Turn1Content = &turn1
			response.Contents = append(response.Contents, turn1)
		}
	}

	// Retrieve Turn2 content if path exists
	if verificationRecord.Turn2ProcessedPath != "" {
		turn2Content, err := getS3Content(ctx, verificationRecord.Turn2ProcessedPath)
		if err != nil {
			log.WithError(err).Warn("Failed to retrieve Turn2 S3 content")
		} else {
			turn2 := ConversationContent{
				Turn:        2,
				Content:     turn2Content,
				ContentType: "text/markdown",
				S3Path:      verificationRecord.Turn2ProcessedPath,
			}
			response.Turn2Content = &turn2
			response.Contents = append(response.Contents, turn2)
		}
	}

	// Check if we have any content
	if len(response.Contents) == 0 {
		log.WithFields(logrus.Fields{
			"verificationId": verificationId,
		}).Error("No processed paths found in verification record")
		return createErrorResponse(404, "Content not found", "No processed conversation content paths found", headers)
	}

	// Convert response to JSON
	responseBody, err := json.Marshal(response)
	if err != nil {
		log.WithError(err).Error("Failed to marshal response")
		return createErrorResponse(500, "Internal server error", "Failed to process response", headers)
	}

	log.WithFields(logrus.Fields{
		"verificationId": verificationId,
		"turn1Available": response.Turn1Content != nil,
		"turn2Available": response.Turn2Content != nil,
		"totalContents":  len(response.Contents),
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



func getVerificationRecord(ctx context.Context, verificationId string) (*VerificationRecord, error) {
	log.WithFields(logrus.Fields{
		"verificationId": verificationId,
		"tableName":      verificationTableName,
	}).Debug("Querying DynamoDB for verification record")

	// Query DynamoDB using verificationId as the key
	input := &dynamodb.QueryInput{
		TableName:              aws.String(verificationTableName),
		KeyConditionExpression: aws.String("verificationId = :verificationId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":verificationId": &types.AttributeValueMemberS{Value: verificationId},
		},
		Limit: aws.Int32(1), // We only need one record
	}

	result, err := dynamoClient.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query verification table: %w", err)
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
		"turn1ProcessedPath": record.Turn1ProcessedPath,
		"turn2ProcessedPath": record.Turn2ProcessedPath,
	}).Debug("Successfully retrieved verification record")

	return &record, nil
}

func getS3Content(ctx context.Context, s3Path string) (string, error) {
	if s3Path == "" {
		return "", fmt.Errorf("s3 path is empty")
	}

	var bucket, key string

	// Parse S3 path to extract bucket and key
	if strings.HasPrefix(s3Path, "s3://") {
		// Extract bucket and key from full S3 URI: s3://bucket/key
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
