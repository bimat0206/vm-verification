package handler

import (
	"context"
	"encoding/json"  
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
	"workflow-function/shared/errors"
)

// Handler provides the core logic for ExecuteTurn1
type Handler struct {
	bedrockClient *bedrockruntime.Client
	s3Client      *s3.Client
	logger        logger.Logger
	hybridConfig  *schema.HybridStorageConfig
}

// NewHandler constructs the ExecuteTurn1 handler with injected dependencies.
func NewHandler(bedrockClient *bedrockruntime.Client, s3Client *s3.Client, hybridConfig *schema.HybridStorageConfig, logger logger.Logger) *Handler {
	return &Handler{
		bedrockClient: bedrockClient,
		s3Client:      s3Client,
		hybridConfig:  hybridConfig,
		logger:        logger,
	}
}

// HandleRequest executes Turn 1: validates input, invokes Bedrock, updates WorkflowState.
func (h *Handler) HandleRequest(ctx context.Context, state *schema.WorkflowState) (*schema.WorkflowState, error) {
	log := h.logger.WithFields(map[string]interface{}{
		"verificationId": state.VerificationContext.VerificationId,
		"step":           "ExecuteTurn1",
	})

	log.Info("Begin ExecuteTurn1 - validating input", nil)

	// Ensure schema version is set to the latest supported version
	if state.SchemaVersion == "" || state.SchemaVersion != schema.SchemaVersion {
		log.Info("Updating schema version", map[string]interface{}{
			"from": state.SchemaVersion, 
			"to": schema.SchemaVersion,
		})
		state.SchemaVersion = schema.SchemaVersion
	}

	// Validate core workflow state, image data, prompt, and Bedrock config
	if errs := schema.ValidateWorkflowState(state); len(errs) > 0 {
		wfErr := errors.NewValidationError("Invalid WorkflowState", map[string]interface{}{"validationErrors": errs.Error()})
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}
	if errs := schema.ValidateImageData(state.Images, true); len(errs) > 0 {
		wfErr := errors.NewValidationError("Invalid ImageData", map[string]interface{}{"validationErrors": errs.Error()})
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}
	if errs := schema.ValidateCurrentPrompt(state.CurrentPrompt, true); len(errs) > 0 {
		wfErr := errors.NewValidationError("Invalid CurrentPrompt", map[string]interface{}{"validationErrors": errs.Error()})
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}
	if errs := schema.ValidateBedrockConfig(state.BedrockConfig); len(errs) > 0 {
		wfErr := errors.NewValidationError("Invalid BedrockConfig", map[string]interface{}{"validationErrors": errs.Error()})
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}

	// Prepare hybrid Base64 retriever for reference image
	retriever := schema.NewHybridBase64Retriever(h.s3Client, h.hybridConfig)

	// Ensure Base64 for images (hybrid logic handled in schema)
	err := schema.HybridImageProcessor(retriever, nil).EnsureHybridBase64Generated(state.Images)
	if err != nil {
		wfErr := errors.NewInternalError("HybridBase64Generation", err)
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}

	// Build Bedrock InvokeModel input using shared schema
	modelInput, err := h.buildConverseInput(ctx, state, retriever)
	if err != nil {
		wfErr := errors.NewBedrockError("Failed to build Bedrock model input", "BEDROCK_BUILD_INPUT", false).
			WithContext("detail", err.Error())
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}

	// Invoke Bedrock Model API
	log.Info("Invoking Bedrock Model API", map[string]interface{}{
		"modelId":   aws.ToString(modelInput.ModelId),
		"promptId":  state.CurrentPrompt.PromptId,
	})

	start := time.Now()
	resp, err := h.bedrockClient.InvokeModel(ctx, modelInput)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		wfErr := errors.NewBedrockError("Bedrock Model API failed", "BEDROCK_API_FAILURE", true).
			WithContext("bedrockError", err.Error())
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}

	// Parse Bedrock response into Turn1Response
	modelId := aws.ToString(modelInput.ModelId)
	turn1Response, err := h.buildTurn1Response(state, resp, latencyMs, modelId)
	if err != nil {
		wfErr := errors.NewParsingError("Bedrock Model API response", err)
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}

	// Update workflow state
	state.Turn1Response = map[string]interface{}{"turnResponse": turn1Response}
	state.VerificationContext.Status = schema.StatusTurn1Completed
	state.ConversationState = h.updateConversationState(state, turn1Response)
	state.VerificationContext.Error = nil // clear error on success

	log.Info("Turn1 completed", map[string]interface{}{
		"latencyMs":  latencyMs,
		"turnId":     turn1Response.TurnId,
		"tokens":     turn1Response.TokenUsage,
		"status":     state.VerificationContext.Status,
	})

	return state, nil
}

// --- Helper Methods ---

// setErrorAndLog centralizes error logging and state update.
func (h *Handler) setErrorAndLog(state *schema.WorkflowState, err error, log logger.Logger, status string) {
	log.Error("Error in ExecuteTurn1", map[string]interface{}{"error": err.Error()})
	var wfErr *errors.WorkflowError
	if e, ok := err.(*errors.WorkflowError); ok {
		wfErr = e
	} else {
		wfErr = errors.WrapError(err, errors.ErrorTypeInternal, "Unexpected error", false)
	}
	
	// Convert WorkflowError to schema.ErrorInfo
	state.VerificationContext.Error = &schema.ErrorInfo{
		Code:      wfErr.Code,
		Message:   wfErr.Message,
		Details:   wfErr.Context,
		Timestamp: schema.FormatISO8601(),
	}
	state.VerificationContext.Status = status
}

// buildConverseInput prepares a simplified JSON-based input for the Bedrock Converse API
func (h *Handler) buildConverseInput(ctx context.Context, state *schema.WorkflowState, retriever *schema.HybridBase64Retriever) (*bedrockruntime.InvokeModelInput, error) {
	// Ensure prompt messages exist
	if len(state.CurrentPrompt.Messages) == 0 {
		return nil, fmt.Errorf("no Bedrock messages found in current prompt")
	}

	// Build simplified payload structure
	type Message struct {
		Role    string                   `json:"role"`
		Content []map[string]interface{} `json:"content"`
	}

	type RequestPayload struct {
		ModelID         string    `json:"model_id"`
		Messages        []Message `json:"messages"`
		MaxTokens       int       `json:"max_tokens"`
		Temperature     float64   `json:"temperature"`
		AnthropicVersion string   `json:"anthropic_version"`
	}

	// Prepare messages in the appropriate format
	var messages []Message
	for _, msg := range state.CurrentPrompt.Messages {
		var contentBlocks []map[string]interface{}

		for _, content := range msg.Content {
			switch content.Type {
			case "text":
				contentBlocks = append(contentBlocks, map[string]interface{}{
					"type": "text",
					"text": content.Text,
				})
			case "image":
				if content.Image != nil {
					base64Data, err := retriever.RetrieveBase64Data(h.getReferenceImageInfo(state.Images))
					if err != nil {
						return nil, fmt.Errorf("failed to retrieve image Base64: %w", err)
					}
					
					contentBlocks = append(contentBlocks, map[string]interface{}{
						"type": "image",
						"source": map[string]interface{}{
							"type":   "base64",
							"media_type": content.Image.Format,
							"data": base64Data,
						},
					})
				}
			}
		}

		messages = append(messages, Message{
			Role:    msg.Role,
			Content: contentBlocks,
		})
	}

	// Create the payload
	payload := RequestPayload{
		ModelID:    state.BedrockConfig.AnthropicVersion,
		Messages:   messages,
		MaxTokens:  state.BedrockConfig.MaxTokens,
		Temperature: state.BedrockConfig.Temperature,
		AnthropicVersion: state.BedrockConfig.AnthropicVersion,
	}

	// Marshal to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create the InvokeModelInput using AnthropicVersion as the model ID
	input := &bedrockruntime.InvokeModelInput{
		ModelId:          aws.String(state.BedrockConfig.AnthropicVersion),
		ContentType:      aws.String("application/json"),
		Accept:           aws.String("application/json"),
		Body:             payloadBytes,
	}

	return input, nil
}

// buildTurn1Response parses the Bedrock response and builds a TurnResponse.
func (h *Handler) buildTurn1Response(state *schema.WorkflowState, resp *bedrockruntime.InvokeModelOutput, latencyMs int64, modelId string) (*schema.TurnResponse, error) {
	// Parse response JSON
	var responseData struct {
		Type      string `json:"type"`
		Content   []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		StopSequence string `json:"stop_sequence"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(resp.Body, &responseData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Bedrock response: %w", err)
	}

	// Extract content and thinking
	var content, thinking string
	for _, item := range responseData.Content {
		if item.Type == "text" {
			content += item.Text
		}
	}

	// Extract thinking from content if present
	thinkingPattern := regexp.MustCompile(`(?s)<thinking>(.*?)</thinking>`)
	if matches := thinkingPattern.FindStringSubmatch(content); len(matches) > 1 {
		thinking = matches[1]
	}

	// Build token usage
	tokenUsage := &schema.TokenUsage{
		InputTokens:  responseData.Usage.InputTokens,
		OutputTokens: responseData.Usage.OutputTokens,
		TotalTokens:  responseData.Usage.InputTokens + responseData.Usage.OutputTokens,
	}

	// Build the response
	return &schema.TurnResponse{
		TurnId:    1,
		Timestamp: schema.FormatISO8601(),
		Prompt:    state.CurrentPrompt.Text,
		ImageUrls: map[string]string{
			"reference": h.getReferenceImageInfo(state.Images).URL,
		},
		Response: schema.BedrockApiResponse{
			Content:    content,
			Thinking:   thinking,
			StopReason: responseData.StopReason,
			ModelId:    modelId,
			RequestId:  "request-id-not-available", // Response metadata not accessible
		},
		LatencyMs:  latencyMs,
		TokenUsage: tokenUsage,
		Stage:      schema.StatusTurn1Completed,
	}, nil
}

// updateConversationState updates the conversation state/history after Turn1.
func (h *Handler) updateConversationState(state *schema.WorkflowState, turn1Response *schema.TurnResponse) *schema.ConversationState {
	cs := state.ConversationState
	if cs == nil {
		cs = &schema.ConversationState{
			CurrentTurn: 1,
			MaxTurns:    2,
			History:     []interface{}{},
		}
	}
	cs.History = append(cs.History, *turn1Response)
	cs.CurrentTurn = 1
	return cs
}

// getReferenceImageInfo returns the reference ImageInfo from ImageData.
func (h *Handler) getReferenceImageInfo(images *schema.ImageData) *schema.ImageInfo {
	if images.Reference != nil {
		return images.Reference
	}
	return images.ReferenceImage // legacy fallback
}