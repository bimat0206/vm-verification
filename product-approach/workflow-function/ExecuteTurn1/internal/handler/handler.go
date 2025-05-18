package handler

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
)

const (
	// Retry configuration
	maxRetryAttempts = 3
	baseRetryDelay   = time.Second * 2

	// Turn configuration
	turn1ID = 1
)

// Handler provides the core logic for ExecuteTurn1
type Handler struct {
	bedrockClient *bedrockruntime.Client
	s3Client      *s3.Client
	logger        logger.Logger
	hybridConfig  *schema.HybridStorageConfig
	modelId       string  // Model ID from environment configuration
}

// NewHandler constructs the ExecuteTurn1 handler with injected dependencies.
func NewHandler(
	bedrockClient *bedrockruntime.Client,
	s3Client *s3.Client,
	hybridConfig *schema.HybridStorageConfig,
	logger logger.Logger,
	modelId string,
) *Handler {
	return &Handler{
		bedrockClient: bedrockClient,
		s3Client:      s3Client,
		hybridConfig:  hybridConfig,
		logger:        logger,
		modelId:       modelId,
	}
}

// HandleRequest executes Turn 1: validates input, invokes Bedrock with retries, and updates WorkflowState.
func (h *Handler) HandleRequest(ctx context.Context, state *schema.WorkflowState) (*schema.WorkflowState, error) {
	log := h.logger.WithFields(map[string]interface{}{
		"verificationId": state.VerificationContext.VerificationId,
		"step":           "ExecuteTurn1",
		"turnId":         turn1ID,
	})
	
	log.Info("Starting ExecuteTurn1 execution", nil)

	// Ensure schema version is current
	h.ensureSchemaVersion(state, log)

	// Step 1: Validate core workflow state and configuration
	if err := h.validateCoreWorkflowState(state, log); err != nil {
		return state, err
	}

	// Step 2: Generate Base64 for images using hybrid storage
	retriever, err := h.generateBase64Images(state, log)
	if err != nil {
		return state, err
	}

	// Step 3: Validate complete workflow state including images
	if err := h.validateCompleteWorkflowState(state, log); err != nil {
		return state, err
	}

	// Step 4: Build and execute Bedrock request
	turn1Response, err := h.executeBedrockRequest(ctx, state, retriever, log)
	if err != nil {
		return state, err
	}

	// Step 5: Update workflow state with results
	h.updateWorkflowStateWithResults(state, turn1Response, log)

	log.Info("ExecuteTurn1 completed successfully", map[string]interface{}{
		"latencyMs": turn1Response.LatencyMs,
		"turnId":    turn1Response.TurnId,
		"status":    state.VerificationContext.Status,
	})

	return state, nil
}

// ensureSchemaVersion ensures the workflow state uses the current schema version
func (h *Handler) ensureSchemaVersion(state *schema.WorkflowState, log logger.Logger) {
	if state.SchemaVersion != schema.SchemaVersion {
		log.Info("Updating schema version", map[string]interface{}{
			"from": state.SchemaVersion,
			"to":   schema.SchemaVersion,
		})
		state.SchemaVersion = schema.SchemaVersion
	}
}