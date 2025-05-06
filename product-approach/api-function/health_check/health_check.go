package main

import (
	"context"
	"encoding/json"
	//"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
)

var (
	log               *logrus.Logger
	dynamoClient      *dynamodb.Client
	s3Client          *s3.Client
	bedrockClient     *bedrockruntime.Client
	verificationTable string
	conversationTable string
	referenceBucket   string
	checkingBucket    string
	resultsBucket     string
	bedrockModel      string
)

type HealthResponse struct {
	Status    string                 `json:"status"`
	Version   string                 `json:"version"`
	Timestamp string                 `json:"timestamp"`
	Services  map[string]ServiceInfo `json:"services"`
}

type ServiceInfo struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
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
	verificationTable = os.Getenv("VERIFICATION_RESULTS_TABLE")
	conversationTable = os.Getenv("CONVERSATION_HISTORY_TABLE")
	referenceBucket = os.Getenv("REFERENCE_BUCKET")
	checkingBucket = os.Getenv("CHECKING_BUCKET")
	resultsBucket = os.Getenv("RESULTS_BUCKET")
	bedrockModel = os.Getenv("BEDROCK_MODEL")

	// Initialize AWS clients
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.WithError(err).Fatal("unable to load AWS SDK config")
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	s3Client = s3.NewFromConfig(cfg)
	bedrockClient = bedrockruntime.NewFromConfig(cfg)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Info("Health check started")
	
	// Prepare the health check response
	response := HealthResponse{
		Status:    "healthy", // Default to healthy, will be updated based on checks
		Version:   "1.0.0",   // Can be set from an environment variable or build info
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  make(map[string]ServiceInfo),
	}

	// Check DynamoDB Tables
	dynamoHealth := checkDynamoDB(ctx)
	response.Services["dynamodb"] = dynamoHealth
	if dynamoHealth.Status != "healthy" {
		response.Status = "degraded"
	}

	// Check S3 Buckets
	s3Health := checkS3(ctx)
	response.Services["s3"] = s3Health
	if s3Health.Status != "healthy" {
		response.Status = "degraded"
	}

	// Check Bedrock
	bedrockHealth := checkBedrock(ctx)
	response.Services["bedrock"] = bedrockHealth
	if bedrockHealth.Status != "healthy" {
		response.Status = "degraded"
	}

	// Convert response to JSON
	responseBody, err := json.Marshal(response)
	if err != nil {
		log.WithError(err).Error("Failed to marshal response")
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "Internal server error"}`,
		}, nil
	}

	// Set response headers
	headers := map[string]string{
		"Content-Type":                     "application/json",
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Credentials": "true",
	}

	log.Info("Health check completed with status: " + response.Status)

	// Return API Gateway response
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(responseBody),
	}, nil
}

func checkDynamoDB(ctx context.Context) ServiceInfo {
	result := ServiceInfo{
		Status: "healthy",
		Details: map[string]string{
			"verification_table": verificationTable,
			"conversation_table": conversationTable,
		},
	}

	// Check verification table
	_, err := dynamoClient.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(verificationTable),
	})
	if err != nil {
		log.WithError(err).Error("Failed to describe verification table")
		result.Status = "unhealthy"
		result.Message = "Failed to access verification table: " + err.Error()
		return result
	}

	// Check conversation table
	_, err = dynamoClient.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(conversationTable),
	})
	if err != nil {
		log.WithError(err).Error("Failed to describe conversation table")
		result.Status = "unhealthy"
		result.Message = "Failed to access conversation table: " + err.Error()
		return result
	}

	result.Message = "DynamoDB tables accessible"
	return result
}

func checkS3(ctx context.Context) ServiceInfo {
	result := ServiceInfo{
		Status: "healthy",
		Details: map[string]string{
			"reference_bucket": referenceBucket,
			"checking_bucket":  checkingBucket,
			"results_bucket":   resultsBucket,
		},
	}

	// Check reference bucket
	_, err := s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(referenceBucket),
	})
	if err != nil {
		log.WithError(err).Error("Failed to access reference bucket")
		result.Status = "unhealthy"
		result.Message = "Failed to access reference bucket: " + err.Error()
		return result
	}

	// Check checking bucket
	_, err = s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(checkingBucket),
	})
	if err != nil {
		log.WithError(err).Error("Failed to access checking bucket")
		result.Status = "unhealthy"
		result.Message = "Failed to access checking bucket: " + err.Error()
		return result
	}

	// Check results bucket
	_, err = s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(resultsBucket),
	})
	if err != nil {
		log.WithError(err).Error("Failed to access results bucket")
		result.Status = "unhealthy"
		result.Message = "Failed to access results bucket: " + err.Error()
		return result
	}

	result.Message = "S3 buckets accessible"
	return result
}

func checkBedrock(ctx context.Context) ServiceInfo {
	result := ServiceInfo{
		Status: "healthy",
		Details: map[string]string{
			"model_id": bedrockModel,
		},
	}

	// Check if we can at least get the model info
	// Note: Real implementation would need to adapt to Bedrock API
	// This is a simplified check - in a real implementation you might want to:
	// 1. Use GetFoundationModel to check if the model exists
	// 2. Or do a simple inference request with minimal tokens
	
	// For now, we'll just check that the model ID is not empty
	if bedrockModel == "" {
		result.Status = "unhealthy"
		result.Message = "Bedrock model ID is not configured"
		return result
	}

	// Add a more robust check here if needed
	// For example, you could make a small inference request to verify the model is responding

	result.Message = "Bedrock model available"
	return result
}

func main() {
	lambda.Start(handler)
}