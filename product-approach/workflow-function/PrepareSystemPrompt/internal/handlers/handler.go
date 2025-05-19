package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	
	"workflow-function/PrepareSystemPrompt/internal/adapters"
	"workflow-function/PrepareSystemPrompt/internal/config"
	"workflow-function/PrepareSystemPrompt/internal/models"
	"workflow-function/PrepareSystemPrompt/internal/processors"
)

// Handler handles Lambda requests
type Handler struct {
	config            *config.Config
	logger            logger.Logger
	s3Adapter         *adapters.S3StateAdapter
	bedrockAdapter    *adapters.BedrockAdapter
	templateProcessor *processors.TemplateProcessor
	validationProcessor *processors.ValidationProcessor
}

// NewHandler creates a new handler
func NewHandler(cfg *config.Config) (*Handler, error) {
	// Initialize logger
	log := logger.New(cfg.ComponentName, "handler")
	if cfg.Debug {
		// Add fields if debug is enabled
		log = log.WithFields(map[string]interface{}{
			"debug": cfg.Debug,
		})
	}
	
	// Initialize S3 adapter
	s3Adapter, err := adapters.NewS3StateAdapter(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 adapter: %w", err)
	}
	
	// Initialize Bedrock adapter
	bedrockAdapter := adapters.NewBedrockAdapter(cfg, log)
	
	// Initialize template processor
	templateProcessor, err := processors.NewTemplateProcessor(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create template processor: %w", err)
	}
	
	// Initialize validation processor
	validationProcessor := processors.NewValidationProcessor(cfg, log)
	
	return &Handler{
		config:            cfg,
		logger:            log,
		s3Adapter:         s3Adapter,
		bedrockAdapter:    bedrockAdapter,
		templateProcessor: templateProcessor,
		validationProcessor: validationProcessor,
	}, nil
}

// HandleRequest handles Lambda requests
func (h *Handler) HandleRequest(ctx context.Context, event json.RawMessage) (json.RawMessage, error) {
	start := time.Now()
	h.logger.Info("Received request", map[string]interface{}{
		"timestamp": start.Format(time.RFC3339),
	})
	
	// Parse input
	var input models.Input
	if err := json.Unmarshal(event, &input); err != nil {
		h.logger.Error("Failed to parse input", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("invalid input format: %w", err)
	}
	
	// Validate input
	if err := h.validationProcessor.ValidateInput(&input); err != nil {
		h.logger.Error("Input validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("input validation failed: %w", err)
	}
	
	// Process based on input type
	switch input.Type {
	case models.InputTypeS3Reference:
		return h.processS3ReferenceInput(ctx, &input)
	case models.InputTypeDirectJSON:
		return h.processDirectJSONInput(ctx, &input)
	default:
		return nil, fmt.Errorf("unsupported input type")
	}
}

// processS3ReferenceInput processes an input with S3 references
func (h *Handler) processS3ReferenceInput(ctx context.Context, input *models.Input) (json.RawMessage, error) {
	start := time.Now()
	h.logger.Info("Processing S3 reference input", map[string]interface{}{
		"verificationId": input.GetVerificationID(),
	})
	
	// Load state from S3
	state, datePartition, err := h.s3Adapter.LoadStateFromEnvelope(ctx, input.S3Envelope)
	if err != nil {
		h.logger.Error("Failed to load state from S3", map[string]interface{}{
			"error": err.Error(),
			"verificationId": input.GetVerificationID(),
		})
		return nil, fmt.Errorf("failed to load state from S3: %w", err)
	}
	
	// Extract verification context and metadata
	verificationContext := state.VerificationContext
	if verificationContext == nil {
		// This should not happen now that we added recovery in the adapter,
		// but we'll handle it just in case
		h.logger.Error("Verification context is nil after loading from S3", map[string]interface{}{
			"verificationId": input.GetVerificationID(),
			"state_schema_version": state.SchemaVersion,
			"envelope": fmt.Sprintf("%+v", input.S3Envelope),
		})
		
		// Try to create a minimal verification context if possible
		if input.S3Envelope != nil && input.S3Envelope.VerificationID != "" {
			h.logger.Info("Creating minimal verification context from envelope", map[string]interface{}{
				"verificationId": input.S3Envelope.VerificationID,
				"references": fmt.Sprintf("%v", input.S3Envelope.References),
			})
			
			// Attempt to extract verification type from references
			verificationType := schema.VerificationTypeLayoutVsChecking // Default
			for key := range input.S3Envelope.References {
				if strings.Contains(strings.ToLower(key), "layout") {
					verificationType = schema.VerificationTypeLayoutVsChecking
					break
				} else if strings.Contains(strings.ToLower(key), "previous") {
					verificationType = schema.VerificationTypePreviousVsCurrent
					break
				}
			}
			
			// Create basic verification context
			verificationContext = &schema.VerificationContext{
				VerificationId: input.S3Envelope.VerificationID,
				Status:         "INITIALIZED",
				VerificationType: verificationType,
				VerificationAt: time.Now().Format(time.RFC3339),
			}
			
			// Update the state
			state.VerificationContext = verificationContext
			
			// Try to get more context from the envelope
			for key, ref := range input.S3Envelope.References {
				if strings.Contains(key, "initialization") && ref != nil {
					h.logger.Info("Found initialization reference, attempting to load directly", map[string]interface{}{
						"key": key,
						"bucket": ref.Bucket,
						"file_path": ref.Key,
					})
					
					// Try to update state directly from S3
					if err := h.s3Adapter.UpdateVerificationContextFromS3(state, ref); err == nil {
						h.logger.Info("Successfully updated verification context from S3", map[string]interface{}{
							"verification_id": state.VerificationContext.VerificationId,
							"verification_type": state.VerificationContext.VerificationType,
						})
						verificationContext = state.VerificationContext
						break
					}
				}
			}
		} else {
			return nil, fmt.Errorf("verification context is nil in state and could not be created: missing verification ID in envelope")
		}
	}
	
	layoutMetadata := state.LayoutMetadata
	historicalContext := state.HistoricalContext
	
	// For LAYOUT_VS_CHECKING, ensure layout metadata is loaded
	if verificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking && layoutMetadata == nil {
		h.logger.Info("Layout metadata is required for LAYOUT_VS_CHECKING but not found in state", map[string]interface{}{
			"verificationId": verificationContext.VerificationId,
		})
		
		// Load layout metadata from S3
		lm, err := h.s3Adapter.LoadLayoutMetadata(datePartition, verificationContext.VerificationId)
		if err == nil && lm != nil {
			h.logger.Info("Successfully loaded layout metadata", map[string]interface{}{
				"verificationId": verificationContext.VerificationId,
			})
			layoutMetadata = lm
			// Update state with loaded metadata
			state.LayoutMetadata = layoutMetadata
		} else {
			h.logger.Error("Failed to load layout metadata", map[string]interface{}{
				"error": err.Error(),
				"verificationId": verificationContext.VerificationId,
			})
			return nil, fmt.Errorf("failed to generate system prompt: failed to build template data: layout metadata is required for LAYOUT_VS_CHECKING")
		}
	}
	
	// Generate system prompt
	prompt, version, err := h.templateProcessor.GenerateSystemPromptContent(
		verificationContext, layoutMetadata, historicalContext)
	if err != nil {
		h.logger.Error("Failed to generate system prompt", map[string]interface{}{
			"error": err.Error(),
			"verificationId": verificationContext.VerificationId,
		})
		return nil, fmt.Errorf("failed to generate system prompt: %w", err)
	}
	
	// Create system prompt object
	systemPrompt := h.bedrockAdapter.CreateSystemPrompt(
		prompt, version, verificationContext.VerificationId)
	
	// Store system prompt in S3
	promptRef, err := h.s3Adapter.StoreSystemPrompt(
		datePartition, verificationContext.VerificationId, systemPrompt)
	if err != nil {
		h.logger.Error("Failed to store system prompt", map[string]interface{}{
			"error": err.Error(),
			"verificationId": verificationContext.VerificationId,
		})
		return nil, fmt.Errorf("failed to store system prompt: %w", err)
	}
	
	// Update verification context status
	verificationContext.Status = schema.StatusPromptPrepared
	
	// Update state
	state.SystemPrompt = systemPrompt
	state.VerificationContext = verificationContext
	
	// Update state in S3
	stateRef, err := h.s3Adapter.UpdateWorkflowState(
		datePartition, verificationContext.VerificationId, state)
	if err != nil {
		h.logger.Error("Failed to update workflow state", map[string]interface{}{
			"error": err.Error(),
			"verificationId": verificationContext.VerificationId,
		})
		return nil, fmt.Errorf("failed to update workflow state: %w", err)
	}
	
	// Create envelope
	envelope := h.s3Adapter.CreateS3Envelope(verificationContext.VerificationId)
	envelope.Status = schema.StatusPromptPrepared
	envelope.AddReference("processing_initialization", stateRef)
	envelope.AddReference("prompts_system", promptRef)
	
	// Create processing time in milliseconds
	processingTimeMs := time.Since(start).Milliseconds()
	
	// Add summary
	envelope.Summary = models.CreateSummary(systemPrompt, verificationContext.VerificationType, processingTimeMs)
	
	// Create response
	response := models.BuildResponseWithContext(envelope, verificationContext, datePartition)
	
	// Convert to JSON
	responseJSON, err := response.ToJSON()
	if err != nil {
		h.logger.Error("Failed to serialize response", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to serialize response: %w", err)
	}
	
	h.logger.Info("Request completed successfully", map[string]interface{}{
		"processingTimeMs": processingTimeMs,
		"verificationId": verificationContext.VerificationId,
	})
	
	return responseJSON, nil
}

// processDirectJSONInput processes an input with direct JSON
func (h *Handler) processDirectJSONInput(ctx context.Context, input *models.Input) (json.RawMessage, error) {
	start := time.Now()
	verificationContext := input.VerificationContext
	
	// Check if verification context is nil
	if verificationContext == nil {
		h.logger.Error("Verification context is nil in direct JSON input", nil)
		return nil, fmt.Errorf("verification context is nil in direct JSON input")
	}
	
	h.logger.Info("Processing direct JSON input", map[string]interface{}{
		"verificationId": verificationContext.VerificationId,
		"verificationType": verificationContext.VerificationType,
	})
	
	// Generate date partition
	datePartition := h.config.CurrentDatePartition()
	
	// Generate system prompt
	prompt, version, err := h.templateProcessor.GenerateSystemPromptContent(
		verificationContext, input.LayoutMetadata, input.HistoricalContext)
	if err != nil {
		h.logger.Error("Failed to generate system prompt", map[string]interface{}{
			"error": err.Error(),
			"verificationId": verificationContext.VerificationId,
		})
		return nil, fmt.Errorf("failed to generate system prompt: %w", err)
	}
	
	// Create system prompt object
	systemPrompt := h.bedrockAdapter.CreateSystemPrompt(
		prompt, version, verificationContext.VerificationId)
	
	// Create workflow state for storage
	state := input.CreateWorkflowState()
	if state == nil {
		state = &schema.WorkflowState{
			SchemaVersion:      "1.0.0",
			VerificationContext: verificationContext,
		}
	}
	
	// Update verification context status
	verificationContext.Status = schema.StatusPromptPrepared
	state.VerificationContext = verificationContext
	state.SystemPrompt = systemPrompt
	
	// Store state in S3
	stateRef, err := h.s3Adapter.UpdateWorkflowState(
		datePartition, verificationContext.VerificationId, state)
	if err != nil {
		h.logger.Error("Failed to store workflow state", map[string]interface{}{
			"error": err.Error(),
			"verificationId": verificationContext.VerificationId,
		})
		return nil, fmt.Errorf("failed to store workflow state: %w", err)
	}
	
	// Store system prompt in S3
	promptRef, err := h.s3Adapter.StoreSystemPrompt(
		datePartition, verificationContext.VerificationId, systemPrompt)
	if err != nil {
		h.logger.Error("Failed to store system prompt", map[string]interface{}{
			"error": err.Error(),
			"verificationId": verificationContext.VerificationId,
		})
		return nil, fmt.Errorf("failed to store system prompt: %w", err)
	}
	
	// Create envelope
	envelope := h.s3Adapter.CreateS3Envelope(verificationContext.VerificationId)
	envelope.Status = schema.StatusPromptPrepared
	envelope.AddReference("processing_initialization", stateRef)
	envelope.AddReference("prompts_system", promptRef)
	
	// Create processing time in milliseconds
	processingTimeMs := time.Since(start).Milliseconds()
	
	// Add summary
	envelope.Summary = models.CreateSummary(systemPrompt, verificationContext.VerificationType, processingTimeMs)
	
	// Create response
	response := models.BuildResponseWithContext(envelope, verificationContext, datePartition)
	
	// Convert to JSON
	responseJSON, err := response.ToJSON()
	if err != nil {
		h.logger.Error("Failed to serialize response", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to serialize response: %w", err)
	}
	
	h.logger.Info("Request completed successfully", map[string]interface{}{
		"processingTimeMs": processingTimeMs,
		"verificationId": verificationContext.VerificationId,
	})
	
	return responseJSON, nil
}