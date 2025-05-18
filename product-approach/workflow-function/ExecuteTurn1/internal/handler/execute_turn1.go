package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors" // Standard errors package for errors.As
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
	wferrors "workflow-function/shared/errors" // Renamed to avoid name collision
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

// HandleRequest executes Turn 1: validates input, invokes Bedrock with retries, and updates WorkflowState.
func (h *Handler) HandleRequest(ctx context.Context, state *schema.WorkflowState) (*schema.WorkflowState, error) {
	log := h.logger.WithFields(map[string]interface{}{
		"verificationId": state.VerificationContext.VerificationId,
		"step":           "ExecuteTurn1",
	})
	log.Info("Begin ExecuteTurn1", nil)

	// Ensure schema version
	if state.SchemaVersion != schema.SchemaVersion {
		log.Info("Updating schema version", map[string]interface{}{ "from": state.SchemaVersion, "to": schema.SchemaVersion })
		state.SchemaVersion = schema.SchemaVersion
	}

	// Validate workflow state
	if errs := schema.ValidateWorkflowState(state); len(errs) > 0 {
		wfErr := wferrors.NewValidationError("Invalid WorkflowState", map[string]interface{}{"validationErrors": errs.Error()})
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}

	// Validate current prompt
	if errs := schema.ValidateCurrentPrompt(state.CurrentPrompt, true); len(errs) > 0 {
		wfErr := wferrors.NewValidationError("Invalid CurrentPrompt", map[string]interface{}{"validationErrors": errs.Error()})
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}

	// Validate Bedrock config
	if errs := schema.ValidateBedrockConfig(state.BedrockConfig); len(errs) > 0 {
		wfErr := wferrors.NewValidationError("Invalid BedrockConfig", map[string]interface{}{"validationErrors": errs.Error()})
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}

	// Ensure Base64 for images (hybrid logic)
	retriever := schema.NewHybridBase64Retriever(h.s3Client, h.hybridConfig)
	if err := schema.HybridImageProcessor(retriever, nil).EnsureHybridBase64Generated(state.Images); err != nil {
		wfErr := wferrors.NewInternalError("HybridBase64Generation", err)
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}

	// Validate full image data now that Base64 is attached
	if errs := schema.ValidateImageData(state.Images, true); len(errs) > 0 {
		wfErr := wferrors.NewValidationError("Invalid ImageData", map[string]interface{}{"validationErrors": errs.Error()})
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}

	// Build Bedrock InvokeModel input
	modelInput, err := h.buildConverseInput(ctx, state, retriever)
	if err != nil {
		wfErr := wferrors.NewBedrockError("Failed to build Bedrock input", "BEDROCK_BUILD_INPUT", false).
			WithContext("detail", err.Error())
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}

	// Invoke Bedrock with retries
	var respOut *bedrockruntime.InvokeModelOutput
	var invokeErr error
	var latencyMs int64
	for attempt := 1; attempt <= 3; attempt++ {
		log.Info("Invoking Bedrock Model API", map[string]interface{}{ "modelId": aws.ToString(modelInput.ModelId), "attempt": attempt })
		start := time.Now()
		respOut, invokeErr = h.bedrockClient.InvokeModel(ctx, modelInput)
		latencyMs = time.Since(start).Milliseconds()
		if invokeErr == nil {
			break
		}
		// Only retry on ServiceException or ThrottlingException
		var apiErr smithy.APIError
		if errors.As(invokeErr, &apiErr) {
			code := apiErr.ErrorCode()
			if (code == "ServiceException" || code == "ThrottlingException") && attempt < 3 {
				time.Sleep(time.Duration(3*(1<<uint(attempt-1))) * time.Second)
				continue
			}
		}
		break
	}
	if invokeErr != nil {
		wfErr := wferrors.NewBedrockError("Bedrock API failure", "BEDROCK_API_FAILURE", true).
			WithContext("bedrockError", invokeErr.Error())
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}

	// Read response body (which is already []byte in the AWS SDK v2)
	bodyBytes := respOut.Body
	if err != nil {
		wfErr := wferrors.NewParsingError("Reading Bedrock response", err)
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}

	// Parse response and build TurnResponse
	turn1Response, err := h.buildTurn1Response(state, bodyBytes, latencyMs, aws.ToString(modelInput.ModelId))
	if err != nil {
		wfErr := wferrors.NewParsingError("Bedrock response parsing", err)
		h.setErrorAndLog(state, wfErr, log, schema.StatusBedrockProcessingFailed)
		return state, wfErr
	}

	// Update workflow state
	state.Turn1Response = map[string]interface{}{"turnResponse": turn1Response}
	state.VerificationContext.Status = schema.StatusTurn1Completed
	state.ConversationState = h.updateConversationState(state, turn1Response)
	state.VerificationContext.Error = nil

	log.Info("Turn1 completed", map[string]interface{}{ "latencyMs": latencyMs, "turnId": turn1Response.TurnId })
	return state, nil
}

// buildConverseInput prepares the Bedrock payload according to Converse API spec
func (h *Handler) buildConverseInput(
	ctx context.Context,
	state *schema.WorkflowState,
	retriever *schema.HybridBase64Retriever,
) (*bedrockruntime.InvokeModelInput, error) {
	if len(state.CurrentPrompt.Messages) == 0 {
		return nil, fmt.Errorf("no messages in CurrentPrompt")
	}
	msg := state.CurrentPrompt.Messages[0]

	// Build content array: text then image
	var content []interface{}
	// Text block
	if len(msg.Content) > 0 && msg.Content[0].Text != "" {
		content = append(content, msg.Content[0].Text)
	}
	// Image block if present
	if len(msg.Content) > 1 && msg.Content[1].Image != nil {
		info := h.getReferenceImageInfo(state.Images)
		b64, err := retriever.RetrieveBase64Data(info)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch Base64 from %s: %w", info.URL, err)
		}
		raw, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("invalid Base64 data: %w", err)
		}
		// Append image object
		content = append(content, map[string]interface{}{ 
			"image": map[string]interface{}{ 
				"format": msg.Content[1].Image.Format,
				"source": map[string]interface{}{ 
					"bytes": raw,
				},
			},
		})
	}

	// Assemble payload map
	payload := map[string]interface{}{
		"messages": []map[string]interface{}{ {"role": msg.Role, "content": content} },
		"max_tokens":       state.BedrockConfig.MaxTokens,
		"anthropic_version": state.BedrockConfig.AnthropicVersion,
	}

	// Marshal to JSON
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Build InvokeModelInput
	return &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(state.BedrockConfig.AnthropicVersion),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        bodyBytes,
	}, nil
}

// buildTurn1Response parses the Bedrock JSON and constructs a TurnResponse
func (h *Handler) buildTurn1Response(
	state *schema.WorkflowState,
	body []byte,
	latencyMs int64,
	modelId string,
) (*schema.TurnResponse, error) {
	var resp struct {
		Content []struct{ Type, Text string } `json:"content"`
		Usage   struct{ InputTokens, OutputTokens int } `json:"usage"`
		StopReason string                                `json:"stop_reason"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal Bedrock response: %w", err)
	}

	// Concatenate text blocks
	var fullText string
	for _, c := range resp.Content {
		if c.Type == "text" {
			fullText += c.Text
		}
	}

	// Token usage
	tokens := &schema.TokenUsage{
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
		TotalTokens:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
	}

	// Build TurnResponse
	return &schema.TurnResponse{
		TurnId:    1,
		Timestamp: schema.FormatISO8601(),
		Prompt:    state.CurrentPrompt.Messages[0].Content[0].Text,
		Response: schema.BedrockApiResponse{
			Content:    fullText,
			StopReason: resp.StopReason,
			ModelId:    modelId,
		},
		LatencyMs:  latencyMs,
		TokenUsage: tokens,
		Stage:      schema.StatusTurn1Completed,
	}, nil
}

// updateConversationState appends the turn to history
func (h *Handler) updateConversationState(
	state *schema.WorkflowState,
	turn *schema.TurnResponse,
) *schema.ConversationState {
	cs := state.ConversationState
	if cs == nil {
		cs = &schema.ConversationState{CurrentTurn: 1, MaxTurns: 2, History: []interface{}{}}
	}
	cs.History = append(cs.History, *turn)
	cs.CurrentTurn = 1
	return cs
}

// getReferenceImageInfo returns the reference baseline image info
func (h *Handler) getReferenceImageInfo(images *schema.ImageData) *schema.ImageInfo {
	if images.Reference != nil {
		return images.Reference
	}
	return images.ReferenceImage // legacy fallback
}

// setErrorAndLog centralizes error logging and state update
func (h *Handler) setErrorAndLog(
	state *schema.WorkflowState,
	err error,
	log logger.Logger,
	status string,
) {
	log.Error("ExecuteTurn1 error", map[string]interface{}{"error": err.Error()})
	var wfErr *wferrors.WorkflowError
	if e, ok := err.(*wferrors.WorkflowError); ok {
		wfErr = e
	} else {
		wfErr = wferrors.WrapError(err, wferrors.ErrorTypeInternal, "unexpected", false)
	}
	state.VerificationContext.Error = &schema.ErrorInfo{
		Code:      wfErr.Code,
		Message:   wfErr.Message,
		Details:   wfErr.Context,
		Timestamp: schema.FormatISO8601(),
	}
	state.VerificationContext.Status = status
}
