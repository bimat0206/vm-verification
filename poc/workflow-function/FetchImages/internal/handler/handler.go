// Package handler provides Lambda handler functionality for the FetchImages function
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"workflow-function/shared/errors"
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

// Handle processes a Lambda event with comprehensive error logging
func (h *Handler) Handle(ctx context.Context, event interface{}) (models.FetchImagesResponse, error) {
	startTime := time.Now()
	correlationID := fmt.Sprintf("fetch-images-%d", startTime.UnixNano())

	// Log the incoming event for debugging
	eventBytes, _ := json.Marshal(event)
	h.logger.Info("Received Lambda event", map[string]interface{}{
		"eventType":     fmt.Sprintf("%T", event),
		"eventBody":     string(eventBytes),
		"correlationId": correlationID,
		"timestamp":     startTime.Format(time.RFC3339),
	})
	
	// Parse the input based on event type
	var req models.FetchImagesRequest
	var parseErr error
	
	switch e := event.(type) {
	case events.LambdaFunctionURLRequest:
		// Function URL invocation
		if parseErr = json.Unmarshal([]byte(e.Body), &req); parseErr != nil {
			enhancedErr := errors.NewValidationError("Failed to parse Function URL input", map[string]interface{}{
				"correlationId": correlationID,
				"eventType":     "LambdaFunctionURLRequest",
				"bodyLength":    len(e.Body),
				"parseError":    parseErr.Error(),
			}).WithComponent("FetchImagesHandler").WithCorrelationID(correlationID)

			h.logger.Error("Failed to parse Function URL input", map[string]interface{}{
				"error":         parseErr.Error(),
				"correlationId": correlationID,
				"bodyLength":    len(e.Body),
			})
			return models.FetchImagesResponse{}, enhancedErr
		}
	case map[string]interface{}:
		// Direct invocation from Step Function
		data, _ := json.Marshal(e)
		if parseErr = json.Unmarshal(data, &req); parseErr != nil {
			enhancedErr := errors.NewValidationError("Failed to parse Step Function input", map[string]interface{}{
				"correlationId": correlationID,
				"eventType":     "StepFunctionMap",
				"dataLength":    len(data),
				"parseError":    parseErr.Error(),
				"inputKeys":     getMapKeys(e),
			}).WithComponent("FetchImagesHandler").WithCorrelationID(correlationID)

			h.logger.Error("Failed to parse Step Function input", map[string]interface{}{
				"error":         parseErr.Error(),
				"correlationId": correlationID,
				"input":         fmt.Sprintf("%+v", e),
				"dataLength":    len(data),
			})
			return models.FetchImagesResponse{}, enhancedErr
		}
	case models.FetchImagesRequest:
		// Direct struct invocation
		req = e
		h.logger.Info("Direct struct invocation", map[string]interface{}{
			"correlationId":  correlationID,
			"verificationId": req.VerificationId,
		})
	default:
		// Try raw JSON unmarshal as fallback
		data, _ := json.Marshal(event)
		if parseErr = json.Unmarshal(data, &req); parseErr != nil {
			enhancedErr := errors.NewValidationError("Failed to parse unknown input type", map[string]interface{}{
				"correlationId": correlationID,
				"eventType":     fmt.Sprintf("%T", event),
				"dataLength":    len(data),
				"parseError":    parseErr.Error(),
			}).WithComponent("FetchImagesHandler").WithCorrelationID(correlationID)

			h.logger.Error("Failed to parse unknown input type", map[string]interface{}{
				"error":         parseErr.Error(),
				"correlationId": correlationID,
				"eventType":     fmt.Sprintf("%T", event),
				"input":         fmt.Sprintf("%+v", event),
			})
			return models.FetchImagesResponse{}, enhancedErr
		}
	}
	
	// Validate request with enhanced error context
	if err := req.Validate(); err != nil {
		enhancedErr := errors.NewValidationError("Request validation failed", map[string]interface{}{
			"correlationId":   correlationID,
			"verificationId":  req.VerificationId,
			"validationError": err.Error(),
		}).WithComponent("FetchImagesHandler").WithCorrelationID(correlationID)

		h.logger.Error("Request validation failed", map[string]interface{}{
			"error":          err.Error(),
			"correlationId":  correlationID,
			"verificationId": req.VerificationId,
		})
		return models.FetchImagesResponse{}, enhancedErr
	}

	// Process the request with enhanced error handling
	response, err := h.fetchService.ProcessRequest(ctx, &req)
	if err != nil {
		// Check if it's already a WorkflowError
		if workflowErr, ok := err.(*errors.WorkflowError); ok {
			// Add handler context to existing error
			workflowErr.WithCorrelationID(correlationID).WithComponent("FetchImagesHandler")

			h.logger.Error("Failed to process request", map[string]interface{}{
				"error":          err.Error(),
				"correlationId":  correlationID,
				"verificationId": req.VerificationId,
				"errorType":      workflowErr.Type,
				"errorCode":      workflowErr.Code,
				"retryable":      workflowErr.Retryable,
			})
			return models.FetchImagesResponse{}, workflowErr
		}

		// Wrap non-WorkflowError with enhanced context
		enhancedErr := errors.NewInternalError("FetchImagesHandler", err).
			WithCorrelationID(correlationID).
			WithContext("verificationId", req.VerificationId).
			WithContext("operation", "ProcessRequest")

		h.logger.Error("Failed to process request", map[string]interface{}{
			"error":          err.Error(),
			"correlationId":  correlationID,
			"verificationId": req.VerificationId,
		})
		return models.FetchImagesResponse{}, enhancedErr
	}

	// Log successful completion with metrics
	processingTime := time.Since(startTime)
	h.logger.Info("Successfully processed request", map[string]interface{}{
		"verificationId":   response.VerificationId,
		"status":           response.Status,
		"referenceCount":   len(response.S3References),
		"correlationId":    correlationID,
		"processingTimeMs": processingTime.Milliseconds(),
		"timestamp":        time.Now().Format(time.RFC3339),
	})
	
	return *response, nil
}

// Helper function to get map keys for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}