package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"

	appconfig "api_images_upload_render/config"
	"api_images_upload_render/dynamodb"
	"api_images_upload_render/logger"
	"api_images_upload_render/prefix"
	"api_images_upload_render/renderer"
	"api_images_upload_render/s3utils"
)

var (
	log                *logrus.Logger
	s3Client           *s3.Client
	referenceBucket    string
	checkingBucket     string
	jsonRenderPath     string
	jsonRenderBucket   string
	jsonRenderS3Path   string
)

// UploadResponse represents the response structure for the file upload API
type UploadResponse struct {
	Success      bool           `json:"success"`
	Message      string         `json:"message"`
	Files        []UploadedFile `json:"files,omitempty"`
	Errors       []string       `json:"errors,omitempty"`
	RenderResult *RenderResult  `json:"renderResult,omitempty"`
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

// RenderResult represents the result of JSON rendering
type RenderResult struct {
	Rendered      bool   `json:"rendered"`
	LayoutID      int64  `json:"layoutId,omitempty"`
	LayoutPrefix  string `json:"layoutPrefix,omitempty"`
	ProcessedKey  string `json:"processedKey,omitempty"`
	Message       string `json:"message,omitempty"`
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

// isBinaryFile determines if a file is likely binary based on its extension
func isBinaryFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	binaryExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
		".tiff": true,
		".tif":  true,
		".pdf":  true,
	}
	return binaryExtensions[ext]
}

// validateBinaryFileIntegrity checks if binary file data appears to be corrupted
func validateBinaryFileIntegrity(filename string, fileBytes []byte) error {
	if !isBinaryFile(filename) {
		return nil // Not a binary file, skip validation
	}

	// Check for common binary file signatures
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		// JPEG files should start with FF D8 FF
		if len(fileBytes) < 3 || fileBytes[0] != 0xFF || fileBytes[1] != 0xD8 || fileBytes[2] != 0xFF {
			return fmt.Errorf("JPEG file signature missing or corrupted - expected FF D8 FF, got %02X %02X %02X",
				fileBytes[0], fileBytes[1], fileBytes[2])
		}
	case ".png":
		// PNG files should start with 89 50 4E 47 0D 0A 1A 0A
		pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		if len(fileBytes) < len(pngSignature) || !bytes.Equal(fileBytes[:len(pngSignature)], pngSignature) {
			return fmt.Errorf("PNG file signature missing or corrupted")
		}
	case ".gif":
		// GIF files should start with "GIF87a" or "GIF89a"
		if len(fileBytes) < 6 || (!bytes.Equal(fileBytes[:6], []byte("GIF87a")) && !bytes.Equal(fileBytes[:6], []byte("GIF89a"))) {
			return fmt.Errorf("GIF file signature missing or corrupted")
		}
	case ".pdf":
		// PDF files should start with "%PDF"
		if len(fileBytes) < 4 || !bytes.Equal(fileBytes[:4], []byte("%PDF")) {
			return fmt.Errorf("PDF file signature missing or corrupted")
		}
	}

	// Check for UTF-8 replacement characters which indicate binary corruption
	if bytes.Contains(fileBytes, []byte{0xEF, 0xBF, 0xBD}) {
		return fmt.Errorf("binary file contains UTF-8 replacement characters - data is corrupted")
	}

	return nil
}

// cleanJSONContent removes multipart form data boundaries from JSON file content
func cleanJSONContent(content []byte) ([]byte, bool) {
	contentStr := string(content)

	// Check if content contains multipart boundaries
	if !strings.Contains(contentStr, "WebKitFormBoundary") {
		return content, false // No cleaning needed
	}

	log.WithField("originalSize", len(content)).Info("Detected multipart boundaries in JSON file, cleaning...")

	// Remove boundary headers at the start
	boundaryHeaderRegex := regexp.MustCompile(`^------WebKitFormBoundary[a-zA-Z0-9]+\r?\n`)
	contentStr = boundaryHeaderRegex.ReplaceAllString(contentStr, "")

	contentDispositionRegex := regexp.MustCompile(`^Content-Disposition: form-data; name="[^"]*"; filename="[^"]*"\r?\n`)
	contentStr = contentDispositionRegex.ReplaceAllString(contentStr, "")

	contentTypeRegex := regexp.MustCompile(`^Content-Type: application/json\r?\n`)
	contentStr = contentTypeRegex.ReplaceAllString(contentStr, "")

	// Remove empty lines after headers
	emptyLineRegex := regexp.MustCompile(`^\r?\n`)
	contentStr = emptyLineRegex.ReplaceAllString(contentStr, "")

	// Remove boundary footers at the end
	boundaryFooterRegex := regexp.MustCompile(`\r?\n------WebKitFormBoundary[a-zA-Z0-9]+--\r?\n?$`)
	contentStr = boundaryFooterRegex.ReplaceAllString(contentStr, "")

	// More aggressive cleaning - remove any remaining boundary lines
	remainingBoundaryRegex := regexp.MustCompile(`------WebKitFormBoundary[a-zA-Z0-9]+[^\n]*\n?`)
	contentStr = remainingBoundaryRegex.ReplaceAllString(contentStr, "")

	remainingContentDispositionRegex := regexp.MustCompile(`Content-Disposition:[^\n]*\n?`)
	contentStr = remainingContentDispositionRegex.ReplaceAllString(contentStr, "")

	remainingContentTypeRegex := regexp.MustCompile(`Content-Type:[^\n]*\n?`)
	contentStr = remainingContentTypeRegex.ReplaceAllString(contentStr, "")

	// Trim any remaining whitespace
	contentStr = strings.TrimSpace(contentStr)

	cleanedContent := []byte(contentStr)
	log.WithFields(logrus.Fields{
		"originalSize": len(content),
		"cleanedSize":  len(cleanedContent),
		"sizeReduction": len(content) - len(cleanedContent),
	}).Info("Successfully cleaned JSON content")

	return cleanedContent, true
}

// validateJSONContent validates that the content is valid JSON
func validateJSONContent(content []byte) error {
	var jsonData interface{}
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return fmt.Errorf("invalid JSON content: %v", err)
	}
	return nil
}

// processJSONFile processes JSON files to remove multipart boundaries and validate content
func processJSONFile(fileName string, fileBytes []byte) ([]byte, error) {
	// Check if this is a JSON file
	if !strings.HasSuffix(strings.ToLower(fileName), ".json") {
		return fileBytes, nil // Not a JSON file, return as-is
	}

	log.WithFields(logrus.Fields{
		"fileName": fileName,
		"fileSize": len(fileBytes),
	}).Info("Processing JSON file")

	// Clean the content
	cleanedContent, wasCleaned := cleanJSONContent(fileBytes)

	// Validate the JSON content
	if err := validateJSONContent(cleanedContent); err != nil {
		if wasCleaned {
			log.WithError(err).Error("JSON validation failed after cleaning")
			return nil, fmt.Errorf("JSON validation failed after cleaning: %v", err)
		} else {
			log.WithError(err).Error("JSON validation failed")
			return nil, fmt.Errorf("JSON validation failed: %v", err)
		}
	}

	if wasCleaned {
		log.WithField("fileName", fileName).Info("JSON file successfully cleaned and validated")
	} else {
		log.WithField("fileName", fileName).Info("JSON file is already clean and valid")
	}

	return cleanedContent, nil
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
	jsonRenderPath = os.Getenv("JSON_RENDER_PATH")

	if referenceBucket == "" || checkingBucket == "" {
		log.Fatal("REFERENCE_BUCKET and CHECKING_BUCKET environment variables are required")
	}

	if jsonRenderPath == "" {
		jsonRenderPath = "raw" // Default to "raw" folder
		jsonRenderBucket = referenceBucket
		jsonRenderS3Path = "raw"
	} else {
		// Parse S3 URI format: s3://bucket-name/path/
		bucket, path, err := parseS3URI(jsonRenderPath)
		if err != nil {
			log.WithError(err).Fatal("Invalid JSON_RENDER_PATH format. Expected: s3://bucket-name/path/ or simple path")
		}
		jsonRenderBucket = bucket
		jsonRenderS3Path = path
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
	}).Info("Upload and render request received")

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
		"bucketType":        bucketType,
		"bucketName":        bucketName,
		"uploadPath":        uploadPath,
		"jsonRenderPath":    jsonRenderPath,
		"jsonRenderBucket":  jsonRenderBucket,
		"jsonRenderS3Path":  jsonRenderS3Path,
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
		"rendered":      response.RenderResult != nil && response.RenderResult.Rendered,
	}).Info("Upload and render request completed")

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

// parseMultipartForm parses multipart form data from API Gateway request
func parseMultipartForm(request events.APIGatewayProxyRequest) ([]byte, string, error) {
	contentType := request.Headers["content-type"]
	if contentType == "" {
		contentType = request.Headers["Content-Type"]
	}

	if !strings.Contains(contentType, "multipart/form-data") {
		return nil, "", fmt.Errorf("request must be multipart/form-data")
	}

	// Get the body
	body := request.Body
	var bodyBytes []byte
	var err error

	if request.IsBase64Encoded {
		// Decode base64 body
		bodyBytes, err = base64.StdEncoding.DecodeString(body)
		if err != nil {
			return nil, "", fmt.Errorf("failed to decode base64 body: %v", err)
		}
		log.WithField("bodySize", len(bodyBytes)).Info("Decoded base64 encoded body")
	} else {
		// For non-base64 encoded bodies, we need to be careful with binary data
		// API Gateway should base64 encode binary data, but if it doesn't,
		// converting string to bytes can corrupt binary data
		bodyBytes = []byte(body)
		log.WithField("bodySize", len(bodyBytes)).Warn("Processing non-base64 encoded body - binary data may be corrupted")

		// Check if the body contains potential binary corruption indicators
		if strings.Contains(body, "\ufffd") {
			log.Error("Detected UTF-8 replacement characters in request body - binary data is corrupted")
			return nil, "", fmt.Errorf("binary data corruption detected - ensure API Gateway is configured for base64 encoding")
		}
	}

	// Parse the content type to get the boundary
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse content type: %v", err)
	}

	boundary, ok := params["boundary"]
	if !ok {
		return nil, "", fmt.Errorf("no boundary found in content type")
	}

	// Create a multipart reader
	reader := multipart.NewReader(bytes.NewReader(bodyBytes), boundary)

	// Parse the multipart form
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("failed to read multipart part: %v", err)
		}

		// Check if this is the file part
		if part.FormName() == "file" {
			// Read the file content
			fileBytes, err := io.ReadAll(part)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read file content: %v", err)
			}

			// Get the filename
			fileName := part.FileName()
			if fileName == "" {
				return nil, "", fmt.Errorf("no filename provided")
			}

			// Log file information for debugging
			log.WithFields(logrus.Fields{
				"fileName":     fileName,
				"fileSize":     len(fileBytes),
				"isBase64":     request.IsBase64Encoded,
				"contentType":  part.Header.Get("Content-Type"),
			}).Info("Successfully parsed file from multipart form")

			return fileBytes, fileName, nil
		}
	}

	return nil, "", fmt.Errorf("no file found in multipart form")
}

func processFileUpload(ctx context.Context, request events.APIGatewayProxyRequest, bucketName, uploadPath string) (*UploadResponse, error) {
	// Parse multipart form data
	fileBytes, fileName, err := parseMultipartForm(request)
	if err != nil {
		log.WithError(err).Error("Failed to parse multipart form")
		return &UploadResponse{
			Success: false,
			Message: "Failed to parse multipart form",
			Errors:  []string{err.Error()},
		}, nil
	}

	// If fileName is not provided in multipart, try to get it from query parameters
	if fileName == "" {
		fileName = request.QueryStringParameters["fileName"]
		if fileName == "" {
			return &UploadResponse{
				Success: false,
				Message: "fileName parameter is required",
				Errors:  []string{"Missing fileName parameter"},
			}, nil
		}
	}

	// Validate file type
	if !isAllowedFileType(fileName) {
		return &UploadResponse{
			Success: false,
			Message: "File type not allowed",
			Errors:  []string{fmt.Sprintf("File type not allowed: %s", filepath.Ext(fileName))},
		}, nil
	}

	// Validate binary file integrity
	if err := validateBinaryFileIntegrity(fileName, fileBytes); err != nil {
		log.WithError(err).Error("Binary file integrity validation failed")
		return &UploadResponse{
			Success: false,
			Message: "File data is corrupted",
			Errors:  []string{fmt.Sprintf("Binary file validation failed: %v", err)},
		}, nil
	}

	// Create S3 key
	s3Key := fileName
	if uploadPath != "" {
		s3Key = uploadPath + "/" + fileName
	}

	// Process JSON files to remove multipart boundaries and validate content
	processedFileBytes, err := processJSONFile(fileName, fileBytes)
	if err != nil {
		log.WithError(err).Error("Failed to process JSON file")
		return &UploadResponse{
			Success: false,
			Message: "Failed to process JSON file",
			Errors:  []string{err.Error()},
		}, nil
	}

	// Use processed file bytes for upload
	fileBytes = processedFileBytes

	// Upload to S3
	contentLength := int64(len(fileBytes))
	if contentLength > MaxFileSize {
		return &UploadResponse{
			Success: false,
			Message: "File too large",
			Errors:  []string{fmt.Sprintf("File size %d exceeds maximum %d bytes", contentLength, MaxFileSize)},
		}, nil
	}

	contentTypeHeader := detectContentType(fileName)

	err = s3utils.UploadFile(ctx, s3Client, bucketName, s3Key, fileBytes, contentTypeHeader)
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

	response := &UploadResponse{
		Success: true,
		Message: "File uploaded successfully",
		Files:   []UploadedFile{uploadedFile},
	}

	// Check if we should render this file (JSON in the configured render path)
	if shouldRenderFile(s3Key, fileName, bucketName) {
		renderResult := processJSONRender(ctx, bucketName, s3Key)
		response.RenderResult = renderResult
		
		if renderResult.Rendered {
			response.Message = "File uploaded and rendered successfully"
		} else {
			response.Message = fmt.Sprintf("File uploaded successfully. Render: %s", renderResult.Message)
		}
	}

	return response, nil
}

// parseS3URI parses an S3 URI and returns bucket and path
func parseS3URI(s3URI string) (bucket, path string, err error) {
	// Handle simple path format (backward compatibility)
	if !strings.HasPrefix(s3URI, "s3://") {
		return "", strings.Trim(s3URI, "/"), nil
	}

	// Remove s3:// prefix
	s3URI = strings.TrimPrefix(s3URI, "s3://")

	// Split bucket and path
	parts := strings.SplitN(s3URI, "/", 2)
	if len(parts) < 1 || parts[0] == "" {
		return "", "", fmt.Errorf("invalid S3 URI format: missing bucket name")
	}

	bucket = parts[0]
	if len(parts) > 1 {
		path = strings.Trim(parts[1], "/")
	}

	return bucket, path, nil
}

// shouldRenderFile determines if a file should be rendered based on path and type
func shouldRenderFile(s3Key, fileName string, bucketName string) bool {
	// Check if file is in the configured render bucket and path, and is a JSON file
	if jsonRenderBucket != "" && bucketName != jsonRenderBucket {
		return false
	}

	// Check if file is in the configured render path and is a JSON file
	if jsonRenderS3Path != "" {
		return strings.HasPrefix(s3Key, jsonRenderS3Path+"/") && filepath.Ext(fileName) == ".json"
	}

	// Fallback to simple path check
	return strings.HasPrefix(s3Key, jsonRenderPath+"/") && filepath.Ext(fileName) == ".json"
}

// processJSONRender handles the rendering of JSON layout files
func processJSONRender(ctx context.Context, bucketName, s3Key string) *RenderResult {
	log.WithFields(logrus.Fields{
		"bucket": bucketName,
		"key":    s3Key,
	}).Info("Starting JSON render process")

	// Check if it's a valid layout JSON
	isLayout, err := s3utils.IsLayoutJSON(ctx, s3Client, bucketName, s3Key)
	if err != nil {
		log.WithError(err).Error("Failed to check if file is layout JSON")
		return &RenderResult{
			Rendered: false,
			Message:  fmt.Sprintf("Failed to validate JSON: %v", err),
		}
	}

	if !isLayout {
		log.Info("File is not a valid layout JSON, skipping render")
		return &RenderResult{
			Rendered: false,
			Message:  "File is not a valid layout JSON",
		}
	}

	// Check file size
	fileSize, err := s3utils.GetObjectSize(ctx, bucketName, s3Key)
	if err != nil {
		log.WithError(err).Error("Failed to get object size")
		return &RenderResult{
			Rendered: false,
			Message:  fmt.Sprintf("Failed to get file size: %v", err),
		}
	}

	// 10MB limit
	if fileSize > 10*1024*1024 {
		log.WithField("fileSize", fileSize).Error("File size exceeds limit")
		return &RenderResult{
			Rendered: false,
			Message:  fmt.Sprintf("File size exceeds limit (10MB): %d bytes", fileSize),
		}
	}

	// Update config with bucket name
	appconfig.UpdateBucketName(bucketName)

	// Initialize logger for render process
	renderLogger := logger.NewLogger(s3Client, bucketName)

	// Download and parse the layout JSON
	layout, err := s3utils.DownloadAndParseLayout(ctx, s3Client, bucketName, s3Key)
	if err != nil {
		log.WithError(err).Error("Failed to download or parse layout")
		renderLogger.Error(ctx, 0, fmt.Sprintf("failed to download or parse layout: %v", err))
		return &RenderResult{
			Rendered: false,
			Message:  fmt.Sprintf("Failed to parse layout: %v", err),
		}
	}

	// Validate layout data
	if layout.LayoutID <= 0 {
		log.Error("Invalid layout ID")
		renderLogger.Error(ctx, layout.LayoutID, "invalid layout ID")
		return &RenderResult{
			Rendered: false,
			Message:  fmt.Sprintf("Invalid layout ID: %d", layout.LayoutID),
		}
	}

	if len(layout.SubLayoutList) == 0 || len(layout.SubLayoutList[0].TrayList) == 0 {
		log.Error("Layout contains no trays")
		renderLogger.Error(ctx, layout.LayoutID, "layout contains no trays")
		return &RenderResult{
			Rendered: false,
			Message:  "Layout contains no trays",
		}
	}

	// Generate unique layoutPrefix
	prefixGenerator := prefix.NewGenerator()
	layoutPrefix, err := prefixGenerator.GenerateLayoutPrefix()
	if err != nil {
		log.WithError(err).Error("Failed to generate layout prefix")
		renderLogger.Error(ctx, layout.LayoutID, fmt.Sprintf("failed to generate layout prefix: %v", err))
		return &RenderResult{
			Rendered: false,
			Message:  fmt.Sprintf("Failed to generate layout prefix: %v", err),
		}
	}

	renderLogger.Info(ctx, layout.LayoutID, fmt.Sprintf("generated layout prefix: %s", layoutPrefix))

	// Render the layout to an image
	imgBytes, err := renderer.RenderLayoutToBytes(*layout)
	if err != nil {
		log.WithError(err).Error("Failed to render layout")
		renderLogger.Error(ctx, layout.LayoutID, fmt.Sprintf("failed to render layout: %v", err))
		return &RenderResult{
			Rendered: false,
			Message:  fmt.Sprintf("Failed to render layout: %v", err),
		}
	}

	// Generate processed key
	processedKey := s3utils.GenerateProcessedKey(s3Key, layout.LayoutID, layoutPrefix)

	// Upload the image to S3
	err = s3utils.UploadImage(ctx, s3Client, bucketName, processedKey, imgBytes)
	if err != nil {
		log.WithError(err).Error("Failed to upload image")
		renderLogger.Error(ctx, layout.LayoutID, fmt.Sprintf("failed to upload image: %v", err))
		return &RenderResult{
			Rendered: false,
			Message:  fmt.Sprintf("Failed to upload image: %v", err),
		}
	}

	renderLogger.Info(ctx, layout.LayoutID, fmt.Sprintf("successfully generated and uploaded image to %s", processedKey))

	// Store metadata in DynamoDB (optional)
	err = storeLayoutMetadata(ctx, layout, layoutPrefix, bucketName, processedKey, s3Key, renderLogger)
	if err != nil {
		log.WithError(err).Warn("Failed to store layout metadata, but render was successful")
		// Don't fail the entire operation if metadata storage fails
	}

	return &RenderResult{
		Rendered:     true,
		LayoutID:     layout.LayoutID,
		LayoutPrefix: layoutPrefix,
		ProcessedKey: processedKey,
		Message:      "Layout rendered successfully",
	}
}

// storeLayoutMetadata stores layout metadata in DynamoDB
func storeLayoutMetadata(ctx context.Context, layout *renderer.Layout, layoutPrefix, s3Bucket, processedKey, sourceKey string, renderLogger *logger.Logger) error {
	// Get DynamoDB table name from environment variable
	tableName := os.Getenv("DYNAMODB_LAYOUT_TABLE")
	if tableName == "" {
		renderLogger.Info(ctx, layout.LayoutID, "DYNAMODB_LAYOUT_TABLE not set, skipping metadata storage")
		return nil // Not an error if table is not configured
	}

	// Get AWS region
	dynamoRegion := os.Getenv("AWS_REGION")
	if dynamoRegion == "" {
		dynamoRegion = "us-east-1" // Default region
	}

	// Initialize DynamoDB manager
	dbManager, err := dynamodb.NewManager(ctx, dynamoRegion, tableName)
	if err != nil {
		renderLogger.Error(ctx, layout.LayoutID, fmt.Sprintf("failed to initialize DynamoDB manager: %v", err))
		return err
	}

	// Store layout metadata in DynamoDB
	err = dbManager.StoreLayoutMetadata(ctx, layout, layoutPrefix, s3Bucket, processedKey, sourceKey)
	if err != nil {
		// If it's a conditional check failure, the item already exists, which is not a fatal error
		if !strings.Contains(err.Error(), "already exists") {
			renderLogger.Error(ctx, layout.LayoutID, fmt.Sprintf("failed to store layout metadata: %v", err))
			return err
		}
		renderLogger.Info(ctx, layout.LayoutID, fmt.Sprintf("layout metadata already exists: %v", err))
	} else {
		renderLogger.Info(ctx, layout.LayoutID, "successfully stored layout metadata in DynamoDB")
	}

	return nil
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
