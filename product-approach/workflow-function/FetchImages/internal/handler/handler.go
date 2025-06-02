// Package handler provides Lambda handler functionality for the FetchImages function
package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"workflow-function/shared/logger"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	localConfig "workflow-function/FetchImages/internal/config"
	"workflow-function/FetchImages/internal/models"
	"workflow-function/FetchImages/internal/repository"
	"workflow-function/FetchImages/internal/service"
)

// Handler handles Lambda events for the FetchImages function
type Handler struct {
	fetchService *service.FetchService
	logger       logger.Logger
	config       localConfig.Config
}

// NewHandler creates a new Handler instance with all dependencies
func NewHandler(ctx context.Context) (*Handler, error) {
	// Load AWS configuration
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	
	// Load application configuration
	cfg := localConfig.New()
	
	// Create logger
	log := logger.New("kootoro-verification", "FetchImagesFunction")
	
	// Log configuration
	log.Info("Configuration loaded", map[string]interface{}{
		"layoutTable":       cfg.LayoutTable,
		"verificationTable": cfg.VerificationTable,
		"stateBucket":       cfg.StateBucket,
		"maxImageSize":      cfg.MaxImageSize,
	})
	
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Error("Configuration validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}
	
	// Create AWS clients
	s3Client := s3.NewFromConfig(awsConfig)
	dynamoDBClient := dynamodb.NewFromConfig(awsConfig)
	
	// Create repositories
	s3Repo := repository.NewS3Repository(s3Client, log)
	dynamoDBRepo := repository.NewDynamoDBRepository(
		dynamoDBClient,
		cfg.LayoutTable,
		cfg.VerificationTable,
		log,
	)
	
	// Create S3 state manager
	stateManager, err := service.NewS3StateManager(ctx, awsConfig, cfg, log)
	if err != nil {
		log.Error("Failed to create S3 state manager", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}
	
	// Create service
	fetchService := service.NewFetchService(s3Repo, dynamoDBRepo, stateManager, log)
	
	return &Handler{
		fetchService: fetchService,
		logger:       log,
		config:       cfg,
	}, nil
}

// Handle processes a Lambda event
func (h *Handler) Handle(ctx context.Context, event interface{}) (models.FetchImagesResponse, error) {
	// Log the incoming event for debugging
	eventBytes, _ := json.Marshal(event)
	h.logger.Info("Received Lambda event", map[string]interface{}{
		"eventType": fmt.Sprintf("%T", event),
		"eventBody": string(eventBytes),
	})
	
	// Parse the input based on event type
	var req models.FetchImagesRequest
	var parseErr error
	
	switch e := event.(type) {
	case events.LambdaFunctionURLRequest:
		// Function URL invocation
		if parseErr = json.Unmarshal([]byte(e.Body), &req); parseErr != nil {
			h.logger.Error("Failed to parse Function URL input", map[string]interface{}{
				"error": parseErr.Error(),
			})
			return models.FetchImagesResponse{}, models.NewValidationError("Invalid JSON input", parseErr)
		}
	case map[string]interface{}:
		// Direct invocation from Step Function
		data, _ := json.Marshal(e)
		if parseErr = json.Unmarshal(data, &req); parseErr != nil {
			h.logger.Error("Failed to parse Step Function input", map[string]interface{}{
				"error": parseErr.Error(),
				"input": fmt.Sprintf("%+v", e),
			})
			return models.FetchImagesResponse{}, models.NewValidationError("Invalid JSON input", parseErr)
		}
	case models.FetchImagesRequest:
		// Direct struct invocation
		req = e
	default:
		// Try raw JSON unmarshal as fallback
		data, _ := json.Marshal(event)
		if parseErr = json.Unmarshal(data, &req); parseErr != nil {
			h.logger.Error("Failed to parse unknown input type", map[string]interface{}{
				"error": parseErr.Error(),
				"input": fmt.Sprintf("%+v", event),
			})
			return models.FetchImagesResponse{}, models.NewValidationError("Invalid JSON input", parseErr)
		}
	}

	// Log the parsed request for debugging
	h.logger.Info("Parsed FetchImages request", map[string]interface{}{
		"verificationId": req.VerificationId,
		"status": req.Status,
		"s3ReferencesCount": len(req.S3References),
		"hasVerificationContext": req.VerificationContext != nil,
		"schemaVersion": req.SchemaVersion,
	})

	// Log verification context details if available
	if req.VerificationContext != nil {
		h.logger.Info("Verification context found in request", map[string]interface{}{
			"verificationId": req.VerificationContext.VerificationId,
			"verificationType": req.VerificationContext.VerificationType,
			"layoutId": req.VerificationContext.LayoutId,
			"layoutPrefix": req.VerificationContext.LayoutPrefix,
			"referenceImageUrl": req.VerificationContext.ReferenceImageUrl,
			"checkingImageUrl": req.VerificationContext.CheckingImageUrl,
		})
	} else {
		h.logger.Warn("No verification context found in request", map[string]interface{}{
			"verificationId": req.VerificationId,
		})
	}

	// Log S3 references for debugging
	if req.S3References != nil {
		for key, ref := range req.S3References {
			h.logger.Info("S3 reference found", map[string]interface{}{
				"key": key,
				"bucket": ref.Bucket,
				"s3Key": ref.Key,
				"size": ref.Size,
			})
		}
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Request validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return models.FetchImagesResponse{}, models.NewValidationError("Request validation failed", err)
	}
	
	// Process the request with enhanced error handling
	response, err := h.fetchService.ProcessRequest(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to process request", map[string]interface{}{
			"error": err.Error(),
			"errorType": fmt.Sprintf("%T", err),
		})

		// Create error response with proper error isolation
		errorResponse := h.createErrorResponse(&req, err)
		return errorResponse, nil // Return error in response, not as Go error
	}

	h.logger.Info("Successfully processed request", map[string]interface{}{
		"verificationId": response.VerificationId,
		"status":         response.Status,
		"referenceCount": len(response.S3References),
	})

	return *response, nil
}

// createErrorResponse creates a FetchImagesResponse with error information
// This isolates current processing errors from inherited workflow errors
func (h *Handler) createErrorResponse(req *models.FetchImagesRequest, err error) models.FetchImagesResponse {
	h.logger.Info("Creating error response with error isolation", map[string]interface{}{
		"verificationId": req.VerificationId,
		"errorType": fmt.Sprintf("%T", err),
		"errorMessage": err.Error(),
	})

	// Create base response structure
	response := models.FetchImagesResponse{
		VerificationId: req.VerificationId,
		S3References:   req.S3References, // Preserve existing references
		Status:         "FETCH_IMAGES_FAILED",
		Summary: map[string]interface{}{
			"imagesFetched": false,
			"errorSource": "FetchImages",
			"errorStage": "processing",
			"errorType": fmt.Sprintf("%T", err),
			"errorMessage": err.Error(),
		},
	}

	// Add error categorization
	if validationErr, ok := err.(*models.ValidationError); ok {
		response.Summary["errorCategory"] = "validation"
		response.Summary["validationDetails"] = validationErr.Message
	} else if processingErr, ok := err.(*models.ProcessingError); ok {
		response.Summary["errorCategory"] = "processing"
		response.Summary["processingDetails"] = processingErr.Message
	} else if notFoundErr, ok := err.(*models.NotFoundError); ok {
		response.Summary["errorCategory"] = "not_found"
		response.Summary["notFoundDetails"] = notFoundErr.Message
	} else {
		response.Summary["errorCategory"] = "unknown"
	}

	h.logger.Info("Error response created", map[string]interface{}{
		"verificationId": response.VerificationId,
		"status": response.Status,
		"errorCategory": response.Summary["errorCategory"],
	})

	return response
}