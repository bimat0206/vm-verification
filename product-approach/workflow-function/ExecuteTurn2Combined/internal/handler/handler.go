package handler

import (
    "context"

    "workflow-function/ExecuteTurn2Combined/internal/config"
    "workflow-function/ExecuteTurn2Combined/internal/models"
    "workflow-function/ExecuteTurn2Combined/internal/services"
    "workflow-function/shared/logger"
)

// Handler orchestrates Turn2 processing

type Handler struct {
    cfg    config.Config
    s3     services.S3StateManager
    bedrock services.BedrockService
    dynamo services.DynamoDBService
    prompts services.PromptService
    log    logger.Logger
}

func NewHandler(s3 services.S3StateManager, br services.BedrockService, dy services.DynamoDBService, pr services.PromptService, log logger.Logger, cfg *config.Config) (*Handler, error) {
    return &Handler{cfg: *cfg, s3: s3, bedrock: br, dynamo: dy, prompts: pr, log: log}, nil
}

// Handle processes Turn 2 and returns a StepFunctionResponse
func (h *Handler) Handle(ctx context.Context, req *models.Turn2Request) (*models.StepFunctionResponse, error) {
    h.log.Info("turn2_start", map[string]interface{}{"verificationId": req.VerificationID})

    // Load system prompt and checking image
    systemPrompt, err := h.s3.LoadSystemPrompt(ctx, req.S3Refs.Prompts.System)
    if err != nil {
        return nil, err
    }
    imageData, err := h.s3.LoadBase64Image(ctx, req.S3Refs.Images.CheckingBase64)
    if err != nil {
        return nil, err
    }
    turn1Markdown, err := h.s3.LoadJSON(ctx, req.S3Refs.Processing.Turn1Markdown, nil)
    if err != nil {
        return nil, err
    }

    // Generate prompt using Turn1 markdown as context
    prompt, err := h.prompts.GenerateTurn2Prompt(systemPrompt, imageData, turn1Markdown)
    if err != nil {
        return nil, err
    }

    // Call bedrock
    resp, err := h.bedrock.Invoke(ctx, prompt, imageData)
    if err != nil {
        return nil, err
    }

    // Store raw response
    rawRef, err := h.s3.StoreRawResponse(ctx, req.VerificationID, resp)
    if err != nil {
        return nil, err
    }

    result := &models.StepFunctionResponse{
        VerificationID: req.VerificationID,
        Status:         "TURN2_COMPLETED",
        S3References:   map[string]models.S3Reference{"rawResponse": rawRef},
    }

    h.log.Info("turn2_completed", map[string]interface{}{"verificationId": req.VerificationID})
    return result, nil
}
