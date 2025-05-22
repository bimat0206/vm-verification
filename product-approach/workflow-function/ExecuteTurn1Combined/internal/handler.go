// internal/handler.go
package handler

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"ExecuteTurn1Combined/internal/config"
	"ExecuteTurn1Combined/internal/errors"
	"ExecuteTurn1Combined/internal/logger"
	"ExecuteTurn1Combined/internal/models"
	bedrocksvc "ExecuteTurn1Combined/internal/services/bedrock"
	dynsvc "ExecuteTurn1Combined/internal/services/dynamodb"
	promptsvc "ExecuteTurn1Combined/internal/services/prompt"
	s3svc "ExecuteTurn1Combined/internal/services/s3"
)

// Handler orchestrates the ExecuteTurn1Combined workflow.
type Handler struct {
	cfg           config.Config
	s3            s3svc.S3StateManager
	bedrock       bedrocksvc.BedrockService
	dynamo        dynsvc.DynamoDBService
	promptService promptsvc.PromptService
	log           *logger.Logger
}

// NewHandler wires together all dependencies for the Lambda.
func NewHandler(
	s3Mgr s3svc.S3StateManager,
	bedrockClient bedrocksvc.BedrockService,
	dynamoClient dynsvc.DynamoDBService,
	promptGen promptsvc.PromptService,
	log *logger.Logger,
	cfg *config.Config,
) *Handler {
	return &Handler{
		cfg:           *cfg,
		s3:            s3Mgr,
		bedrock:       bedrockClient,
		dynamo:        dynamoClient,
		promptService: promptGen,
		log:           log,
	}
}

// Handle executes a single Turn-1 verification cycle.
func (h *Handler) Handle(ctx context.Context, req *models.Turn1Request) (*models.Turn1Response, error) {
	start := time.Now()
	lg := h.log.WithContext(ctx).WithFields(
		"verificationId", req.VerificationID,
		"turnId", 1,
	)
	lg.Info("Starting ExecuteTurn1Combined")

	// --- 1) Concurrently load system prompt & base64 image from S3 ---
	var (
		systemPrompt string
		base64Img    string
		loadErr      error
	)
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		sp, err := h.s3.LoadSystemPrompt(ctx, req.S3Refs.Prompts.System)
		if err != nil {
			loadErr = errors.WrapRetryable(err, errors.StageContextLoading, "failed to load system prompt")
			return
		}
		systemPrompt = sp
	}()

	go func() {
		defer wg.Done()
		img, err := h.s3.LoadBase64Image(ctx, req.S3Refs.Images.ReferenceBase64)
		if err != nil {
			loadErr = errors.WrapRetryable(err, errors.StageContextLoading, "failed to load reference image")
			return
		}
		base64Img = img
	}()

	wg.Wait()
	if loadErr != nil {
		lg.Error("resource loading error", "error", loadErr)
		return nil, loadErr
	}

	// --- 2) Generate Turn-1 prompt ---
	turnPrompt, err := h.promptService.GenerateTurn1Prompt(ctx, req.VerificationContext, systemPrompt)
	if err != nil {
		wrapped := errors.WrapNonRetryable(err, errors.StagePromptGeneration, "prompt generation failed")
		lg.Error("prompt generation error", "error", wrapped)
		return nil, wrapped
	}

	// --- 3) Invoke Bedrock Converse API ---
	resp, err := h.bedrock.Converse(ctx, systemPrompt, turnPrompt, base64Img)
	if err != nil {
		if errors.IsRetryable(err) {
			stepErr := errors.ToStepFnError(err)
			lg.Warn("bedrock retryable error", "error", err)
			return nil, stepErr
		}
		wrapped := errors.WrapNonRetryable(err, errors.StageBedrockCall, "bedrock invocation failed")
		stepErr := errors.ToStepFnError(wrapped)
		lg.Error("bedrock non-retryable error", "error", wrapped)
		return nil, stepErr
	}

	// --- 4) Persist raw & processed responses to S3 ---
	rawRef, err := h.s3.StoreRawResponse(ctx, req.VerificationID, resp.Raw)
	if err != nil {
		wrap := errors.WrapRetryable(err, errors.StageProcessing, "store raw response failed")
		lg.Warn("s3 raw-store warning", "error", wrap)
		return nil, wrap
	}
	procRef, err := h.s3.StoreProcessedAnalysis(ctx, req.VerificationID, resp.Processed)
	if err != nil {
		wrap := errors.WrapRetryable(err, errors.StageProcessing, "store processed analysis failed")
		lg.Warn("s3 processed-store warning", "error", wrap)
		return nil, wrap
	}

	// --- 5) DynamoDB updates (fire & forget) ---
	go func() {
		if err := h.dynamo.UpdateVerificationStatus(ctx, req.VerificationID, models.StatusTurn1Completed, resp.TokenUsage); err != nil {
			lg.Warn("dynamodb status update failed", "error", err)
		}
	}()
	go func() {
		turn := &models.ConversationTurn{
			VerificationID:   req.VerificationID,
			TurnID:           1,
			RawResponseRef:   rawRef,
			ProcessedRef:     procRef,
			TokenUsage:       resp.TokenUsage,
			BedrockRequestID: resp.RequestID,
			Timestamp:        time.Now().UTC(),
		}
		if err := h.dynamo.RecordConversationTurn(ctx, turn); err != nil {
			lg.Warn("dynamodb record conversation failed", "error", err)
		}
	}()

	// --- 6) Build and return response envelope ---
	summary := models.Summary{
		AnalysisStage:    models.StageReferenceAnalysis,
		ProcessingTimeMs: time.Since(start).Milliseconds(),
		TokenUsage:       resp.TokenUsage,
		BedrockRequestID: resp.RequestID,
	}
	output := &models.Turn1Response{
		S3Refs: models.Turn1ResponseS3Refs{
			RawResponse:       rawRef,
			ProcessedResponse: procRef,
		},
		Status:  models.StatusTurn1Completed,
		Summary: summary,
	}

	lg.Info("Completed ExecuteTurn1Combined", "durationMs", summary.ProcessingTimeMs)
	return output, nil
}

// HandleTurn1Combined is the Lambda entrypoint invoked by Step Functions. It
// expects the input payload to conform to the Turn1Request structure.
func (h *Handler) HandleTurn1Combined(ctx context.Context, event json.RawMessage) (*models.Turn1Response, error) {
	var req models.Turn1Request
	if err := json.Unmarshal(event, &req); err != nil {
		return nil, errors.WrapNonRetryable(err, errors.StageValidation, "invalid input payload")
	}
	return h.Handle(ctx, &req)
}
