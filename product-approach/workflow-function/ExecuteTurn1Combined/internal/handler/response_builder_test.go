package handler

import (
	"testing"

	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/schema"
)

func TestBuildCombinedTurnResponseSetsDynamoFlag(t *testing.T) {
	cfg := config.Config{}
	cfg.AWS.BedrockModel = "test-model"
	builder := NewResponseBuilder(cfg)

	req := &models.Turn1Request{
		VerificationID:      "verif1",
		VerificationContext: models.VerificationContext{VerificationType: schema.VerificationTypeLayoutVsChecking},
	}

	invoke := &models.BedrockResponse{
		Raw:        []byte("ok"),
		RequestID:  "req",
		TokenUsage: models.TokenUsage{InputTokens: 1, OutputTokens: 1, ThinkingTokens: 1, TotalTokens: 3},
	}

	resp := builder.BuildCombinedTurnResponse(
		req,
		"prompt",
		models.S3Reference{},
		models.S3Reference{},
		models.S3Reference{},
		invoke,
		nil,
		10,
		5,
		false,
	)

	summary := resp.ContextEnrichment["summary"].(ExecutionSummary)
	if summary.DynamodbUpdated {
		t.Errorf("expected dynamo flag false, got true")
	}
}
