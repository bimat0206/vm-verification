package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
)

var (
	log             *logrus.Logger
	s3Client        *s3.Client
	referenceBucket string
	checkingBucket  string
)

type ViewResponse struct {
	PresignedUrl string `json:"presignedUrl"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
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
	referenceBucket = os.Getenv("REFERENCE_BUCKET")
	checkingBucket = os.Getenv("CHECKING_BUCKET")

	// Initialize AWS clients
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.WithError(err).Fatal("unable to load AWS SDK config")
	}

	s3Client = s3.NewFromConfig(cfg)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract the key from path parameters
	key, exists := request.PathParameters["key"]
	if !exists || key == "" {
		log.Error("Missing or empty key parameter")
		return createErrorResponse(400, "Missing key parameter", "The image key must be provided in the URL path")
	}

	// URL decode the key to handle special characters and spaces
	decodedKey, err := url.QueryUnescape(key)
	if err != nil {
		log.WithError(err).WithField("key", key).Error("Failed to URL decode key")
		return createErrorResponse(400, "Invalid key format", "The provided key contains invalid URL encoding")
	}

	log.WithFields(logrus.Fields{
		"original_key": key,
		"decoded_key":  decodedKey,
	}).Info("Processing image view request")

	// Determine bucket based on query parameter (default to reference)
	bucketType := request.QueryStringParameters["bucketType"]
	if bucketType == "" {
		bucketType = "reference"
	}

	var bucketName string
	switch strings.ToLower(bucketType) {
	case "reference":
		bucketName = referenceBucket
	case "checking":
		bucketName = checkingBucket
	default:
		log.WithField("bucketType", bucketType).Error("Invalid bucket type")
		return createErrorResponse(400, "Invalid bucket type", "bucketType must be 'reference' or 'checking'")
	}

	if bucketName == "" {
		log.WithField("bucketType", bucketType).Error("Bucket name not configured")
		return createErrorResponse(500, "Bucket configuration error", fmt.Sprintf("The %s bucket is not configured", bucketType))
	}

	log.WithFields(logrus.Fields{
		"bucket_name": bucketName,
		"bucket_type": bucketType,
		"key":         decodedKey,
	}).Info("Generating presigned URL")

	// Check if object exists before generating presigned URL
	_, err = s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(decodedKey),
	})
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"bucket": bucketName,
			"key":    decodedKey,
		}).Error("Object not found or access denied")
		return createErrorResponse(404, "Image not found", "The requested image does not exist or is not accessible")
	}

	// Generate presigned URL for GetObject with 1 hour expiration
	presignClient := s3.NewPresignClient(s3Client)
	presignedURL, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(decodedKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Hour // 1 hour expiration
	})

	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"bucket": bucketName,
			"key":    decodedKey,
		}).Error("Failed to generate presigned URL")
		return createErrorResponse(500, "Failed to generate presigned URL", "Unable to create temporary access URL for the image")
	}

	// Create response
	response := ViewResponse{
		PresignedUrl: presignedURL.URL,
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		log.WithError(err).Error("Failed to marshal response")
		return createErrorResponse(500, "Internal server error", "Failed to process response")
	}

	log.WithFields(logrus.Fields{
		"bucket":        bucketName,
		"key":           decodedKey,
		"presigned_url": presignedURL.URL,
	}).Info("Successfully generated presigned URL")

	// Set response headers with CORS support
	headers := map[string]string{
		"Content-Type":                     "application/json",
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
		"Access-Control-Allow-Methods":     "GET,OPTIONS",
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(responseBody),
	}, nil
}

func createErrorResponse(statusCode int, error string, message string) (events.APIGatewayProxyResponse, error) {
	errorResponse := ErrorResponse{
		Error:   error,
		Message: message,
	}

	responseBody, err := json.Marshal(errorResponse)
	if err != nil {
		log.WithError(err).Error("Failed to marshal error response")
		responseBody = []byte(`{"error": "Internal server error"}`)
	}

	headers := map[string]string{
		"Content-Type":                     "application/json",
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
		"Access-Control-Allow-Methods":     "GET,OPTIONS",
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       string(responseBody),
	}, nil
}

func main() {
	lambda.Start(handler)
}
