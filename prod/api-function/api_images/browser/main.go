package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
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

// BrowserResponse represents the response structure for the image browser API
type BrowserResponse struct {
	CurrentPath string      `json:"currentPath"`
	ParentPath  string      `json:"parentPath,omitempty"`
	Items       []BrowserItem `json:"items"`
	TotalItems  int         `json:"totalItems"`
}

// BrowserItem represents a single item (file or folder) in the browser
type BrowserItem struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Type         string    `json:"type"` // "folder" or "image"
	Size         int64     `json:"size,omitempty"`
	LastModified string    `json:"lastModified,omitempty"`
	ContentType  string    `json:"contentType,omitempty"`
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
	}).Info("Image browser request received")

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

	// Extract path from the URL path parameters
	// The path structure is /api/images/browser/{path+}
	pathParam := ""
	if pathParams, exists := request.PathParameters["path"]; exists {
		pathParam = pathParams
	}

	// URL decode the path
	decodedPath, err := url.QueryUnescape(pathParam)
	if err != nil {
		log.WithError(err).Error("Failed to decode path parameter")
		decodedPath = pathParam
	}

	// Clean and normalize the path
	browsePath := cleanPath(decodedPath)

	log.WithFields(logrus.Fields{
		"bucketType":   bucketType,
		"bucketName":   bucketName,
		"browsePath":   browsePath,
		"originalPath": pathParam,
	}).Info("Processing browser request")

	// Browse the S3 bucket
	response, err := browseS3Bucket(ctx, bucketName, browsePath)
	if err != nil {
		log.WithError(err).Error("Failed to browse S3 bucket")
		return createErrorResponse(500, "Failed to browse bucket", err.Error(), headers)
	}

	// Convert response to JSON
	responseBody, err := json.Marshal(response)
	if err != nil {
		log.WithError(err).Error("Failed to marshal response")
		return createErrorResponse(500, "Internal server error", "Failed to process response", headers)
	}

	log.WithFields(logrus.Fields{
		"itemCount": len(response.Items),
		"path":      response.CurrentPath,
	}).Info("Browser request completed successfully")

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(responseBody),
	}, nil
}

func browseS3Bucket(ctx context.Context, bucketName, path string) (*BrowserResponse, error) {
	// Ensure path ends with / for prefix matching (except for root)
	prefix := path
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	log.WithFields(logrus.Fields{
		"bucket": bucketName,
		"prefix": prefix,
	}).Debug("Listing S3 objects")

	// List objects with the given prefix
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucketName),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"), // Use delimiter to get folder-like structure
		MaxKeys:   aws.Int32(1000), // Limit results
	}

	result, err := s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 objects: %w", err)
	}

	var items []BrowserItem

	// Add folders (common prefixes)
	for _, commonPrefix := range result.CommonPrefixes {
		if commonPrefix.Prefix == nil {
			continue
		}

		folderPath := strings.TrimSuffix(*commonPrefix.Prefix, "/")
		folderName := filepath.Base(folderPath)

		if folderName == "" {
			continue
		}

		items = append(items, BrowserItem{
			Name: folderName,
			Path: folderPath,
			Type: "folder",
		})
	}

	// Add files (objects)
	for _, object := range result.Contents {
		if object.Key == nil {
			continue
		}

		objectKey := *object.Key

		// Skip if this is just the prefix itself (folder marker)
		if objectKey == prefix {
			continue
		}

		// Get the file name (last part of the path)
		fileName := filepath.Base(objectKey)
		if fileName == "" || fileName == "." {
			continue
		}

		// Check if this is an image file
		itemType := "file"
		if isImageFile(fileName) {
			itemType = "image"
		}

		// Format last modified time
		lastModified := ""
		if object.LastModified != nil {
			lastModified = object.LastModified.Format(time.RFC3339)
		}

		// Get file size
		size := int64(0)
		if object.Size != nil {
			size = *object.Size
		}

		items = append(items, BrowserItem{
			Name:         fileName,
			Path:         strings.TrimSuffix(objectKey, "/"),
			Type:         itemType,
			Size:         size,
			LastModified: lastModified,
		})
	}

	// Sort items: folders first, then files, both alphabetically
	sort.Slice(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type == "folder" // folders come first
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	// Calculate parent path
	parentPath := ""
	if path != "" {
		parentPath = filepath.Dir(path)
		if parentPath == "." {
			parentPath = ""
		}
	}

	return &BrowserResponse{
		CurrentPath: path,
		ParentPath:  parentPath,
		Items:       items,
		TotalItems:  len(items),
	}, nil
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
		".tiff": true,
		".tif":  true,
	}
	return imageExtensions[ext]
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
