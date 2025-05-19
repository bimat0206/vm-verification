// Package handler provides Lambda handler functionality for the FetchImages function
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"workflow-function/shared/logger"
	
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
	
	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Request validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return models.FetchImagesResponse{}, models.NewValidationError("Request validation failed", err)
	}
	
	// Process the request
	response, err := h.fetchService.ProcessRequest(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to process request", map[string]interface{}{
			"error": err.Error(),
		})
		return models.FetchImagesResponse{}, err
	}
	
	h.logger.Info("Successfully processed request", map[string]interface{}{
		"verificationId": response.VerificationId,
		"status":         response.Status,
		"referenceCount": len(response.S3References),
	})
	
	return *response, nil
}