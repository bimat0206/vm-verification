package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// UploadResponse represents the response structure for the file upload API
type UploadResponse struct {
	Success   bool     `json:"success"`
	Message   string   `json:"message"`
	Files     []UploadedFile `json:"files,omitempty"`
	Errors    []string `json:"errors,omitempty"`
}

// UploadedFile represents information about an uploaded file
type UploadedFile struct {
	OriginalName string `json:"originalName"`
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	ContentType  string `json:"contentType"`
	Bucket       string `json:"bucket"`
	URL          string `json:"url,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

const (
	MaxFileSize = 10 * 1024 * 1024 // 10MB limit for Lambda
)

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

	if referenceBucket == "" || checkingBucket == "" {
		log.Fatal("REFERENCE_BUCKET and CHECKING_BUCKET environment variables are required")
	}

	// Initialize AWS S3 client
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.WithError(err).Fatal("unable to load AWS SDK config")
	}

	s3Client = s3.NewFromConfig(cfg)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.WithFields(logrus.Fields{
		"method": request.HTTPMethod,
		"path":   request.Path,
		"params": request.QueryStringParameters,
	}).Info("File upload request received")

	// Set CORS headers
	headers := map[string]string{
		"Content-Type":                     "application/json",
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
		"Access-Control-Allow-Methods":     "POST,OPTIONS",
	}

	// Handle OPTIONS request for CORS
	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    headers,
			Body:       "",
		}, nil
	}

	// Only allow POST requests
	if request.HTTPMethod != "POST" {
		return createErrorResponse(405, "Method not allowed", "Only POST requests are supported", headers)
	}

	// Extract bucket type from query parameters
	bucketType := request.QueryStringParameters["bucketType"]
	if bucketType == "" {
		bucketType = "reference" // default to reference bucket
	}

	// Validate bucket type
	var bucketName string
	switch bucketType {
	case "reference":
		bucketName = referenceBucket
	case "checking":
		bucketName = checkingBucket
	default:
		return createErrorResponse(400, "Invalid bucket type",
			fmt.Sprintf("Bucket type must be 'reference' or 'checking', got: %s", bucketType), headers)
	}

	// Extract path from query parameters for organizing uploads
	uploadPath := ""
	if pathParam, exists := request.QueryStringParameters["path"]; exists {
		uploadPath = cleanPath(pathParam)
	}

	log.WithFields(logrus.Fields{
		"bucketType": bucketType,
		"bucketName": bucketName,
		"uploadPath": uploadPath,
	}).Info("Processing upload request")

	// Process file upload
	response, err := processFileUpload(ctx, request, bucketName, uploadPath)
	if err != nil {
		log.WithError(err).Error("Failed to process file upload")
		return createErrorResponse(500, "Upload failed", err.Error(), headers)
	}

	// Convert response to JSON
	responseBody, err := json.Marshal(response)
	if err != nil {
		log.WithError(err).Error("Failed to marshal response")
		return createErrorResponse(500, "Internal server error", "Failed to process response", headers)
	}

	log.WithFields(logrus.Fields{
		"uploadedFiles": len(response.Files),
		"success":       response.Success,
	}).Info("Upload request completed")

	statusCode := 200
	if !response.Success {
		statusCode = 400
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       string(responseBody),
	}, nil
}

func processFileUpload(ctx context.Context, request events.APIGatewayProxyRequest, bucketName, uploadPath string) (*UploadResponse, error) {
	// Check if request has multipart content
	contentType := request.Headers["content-type"]
	if contentType == "" {
		contentType = request.Headers["Content-Type"]
	}

	if !strings.Contains(contentType, "multipart/form-data") {
		return &UploadResponse{
			Success: false,
			Message: "Request must be multipart/form-data",
			Errors:  []string{"Invalid content type"},
		}, nil
	}

	// Parse multipart form data from base64 encoded body
	body := request.Body
	if request.IsBase64Encoded {
		// For API Gateway, we need to handle base64 encoded body
		return &UploadResponse{
			Success: false,
			Message: "Base64 encoded uploads not yet supported",
			Errors:  []string{"Use direct binary upload"},
		}, nil
	}

	// For now, we'll handle direct file upload via presigned URLs or direct S3 upload
	// This is a simplified version that expects file content in the body
	fileName := request.QueryStringParameters["fileName"]
	if fileName == "" {
		return &UploadResponse{
			Success: false,
			Message: "fileName parameter is required",
			Errors:  []string{"Missing fileName parameter"},
		}, nil
	}

	// Validate file type
	if !isAllowedFileType(fileName) {
		return &UploadResponse{
			Success: false,
			Message: "File type not allowed",
			Errors:  []string{fmt.Sprintf("File type not allowed: %s", filepath.Ext(fileName))},
		}, nil
	}

	// Create S3 key
	s3Key := fileName
	if uploadPath != "" {
		s3Key = uploadPath + "/" + fileName
	}

	// Upload to S3
	contentLength := int64(len(body))
	if contentLength > MaxFileSize {
		return &UploadResponse{
			Success: false,
			Message: "File too large",
			Errors:  []string{fmt.Sprintf("File size %d exceeds maximum %d bytes", contentLength, MaxFileSize)},
		}, nil
	}

	contentTypeHeader := detectContentType(fileName)
	
	putInput := &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(s3Key),
		Body:          strings.NewReader(body),
		ContentType:   aws.String(contentTypeHeader),
		ContentLength: aws.Int64(contentLength),
	}

	log.WithFields(logrus.Fields{
		"bucket":      bucketName,
		"key":         s3Key,
		"size":        contentLength,
		"contentType": contentTypeHeader,
	}).Info("Uploading file to S3")

	_, err := s3Client.PutObject(ctx, putInput)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	uploadedFile := UploadedFile{
		OriginalName: fileName,
		Key:          s3Key,
		Size:         contentLength,
		ContentType:  contentTypeHeader,
		Bucket:       bucketName,
	}

	return &UploadResponse{
		Success: true,
		Message: "File uploaded successfully",
		Files:   []UploadedFile{uploadedFile},
	}, nil
}

func isAllowedFileType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
		".tiff": true,
		".tif":  true,
		".pdf":  true,
		".txt":  true,
		".json": true,
		".csv":  true,
		".xml":  true,
	}
	return allowedExtensions[ext]
}

func detectContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".webp": "image/webp",
		".tiff": "image/tiff",
		".tif":  "image/tiff",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".json": "application/json",
		".csv":  "text/csv",
		".xml":  "application/xml",
	}
	
	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}
	return "application/octet-stream"
}

func cleanPath(path string) string {
	// Remove leading and trailing slashes
	path = strings.Trim(path, "/")

	// Handle empty path
	if path == "" {
		return ""
	}

	// Clean the path to remove any .. or . components
	cleaned := filepath.Clean(path)

	// Ensure we don't go above root
	if strings.HasPrefix(cleaned, "..") {
		return ""
	}

	// Handle case where cleaned path is just "."
	if cleaned == "." {
		return ""
	}

	return cleaned
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

func main() {
	lambda.Start(handler)
}
