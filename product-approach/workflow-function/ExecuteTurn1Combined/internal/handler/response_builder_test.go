package handler

import (
	"testing"
	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/schema"
)

func TestBuildCombinedTurnResponse(t *testing.T) {
	// Setup
	cfg := createResponseTestConfig()
	responseBuilder := NewResponseBuilder(cfg)
	
	// Create test data
	req := &models.Turn1Request{
		VerificationID: "test-verification-id",
		VerificationContext: models.VerificationContext{
			VerificationType: schema.VerificationTypeLayoutVsChecking,
		},
		S3Refs: models.Turn1RequestS3Refs{
			Prompts: models.PromptRefs{
				System: models.S3Reference{
					Bucket: "test-bucket",
					Key:    "system-prompt.json",
				},
			},
			Images: models.ImageRefs{
				ReferenceBase64: models.S3Reference{
					Bucket: "test-bucket",
					Key:    "reference-base64.json",
				},
			},
		},
	}
	
	renderedPrompt := "test prompt"
	promptRef := models.S3Reference{
		Bucket: "test-bucket",
		Key:    "test-verification-id/prompts/turn1-prompt.json",
		ETag:   "test-etag",
		Size:   100,
	}
	rawRef := models.S3Reference{
		Bucket: "test-bucket",
		Key:    "test-raw-ref",
		ETag:   "test-etag",
		Size:   200,
	}
	procRef := models.S3Reference{
		Bucket: "test-bucket",
		Key:    "test-proc-ref",
		ETag:   "test-etag",
		Size:   300,
	}
	
	// Create mock Bedrock response
	invoke := &models.BedrockResponse{
		Raw:       []byte("test raw response"),
		Processed: map[string]interface{}{"key": "value"},
		RequestID: "test-request-id",
		TokenUsage: models.TokenUsage{
			InputTokens:  100,
			OutputTokens: 200,
			TotalTokens:  300,
		},
	}
	
	// Create processing stages
	stages := []schema.ProcessingStage{
		{
			StageName: "test-stage",
			StartTime: "2025-01-01T00:00:00Z",
			EndTime:   "2025-01-01T00:01:00Z",
			Duration:  60000,
			Status:    "completed",
		},
	}
	
	// Test
	dynamoOK := true
	resp := responseBuilder.BuildCombinedTurnResponse(
		req, renderedPrompt, promptRef, rawRef, procRef,
		invoke, stages, 1000, &dynamoOK,
	)
	
	// Assertions
	
	// 1. Assert schema version is set correctly in contextEnrichment
	schemaVersion, ok := resp.ContextEnrichment["schema_version"]
	if !ok {
		t.Errorf("schema_version not found in contextEnrichment")
	}
	if schemaVersion != schema.SchemaVersion {
		t.Errorf("Expected schema_version to be %s, got %s", schema.SchemaVersion, schemaVersion)
	}
	
	// 2. Assert schema version is "2.1.0"
	if schema.SchemaVersion != "2.1.0" {
		t.Errorf("Expected schema.SchemaVersion to be 2.1.0, got %s", schema.SchemaVersion)
	}
	
	// 3. Assert S3References are set correctly
	s3Refs, ok := resp.ContextEnrichment["s3_references"].(S3ReferenceTree)
	if !ok {
		t.Errorf("s3_references not found in contextEnrichment or not of type S3ReferenceTree")
	} else {
		// 4. Assert Turn1Prompt key ends with turn1-prompt.json
		if s3Refs.Prompts.Turn1Prompt.Key != promptRef.Key {
			t.Errorf("Expected Turn1Prompt.Key to be %s, got %s", promptRef.Key, s3Refs.Prompts.Turn1Prompt.Key)
		}
		
		if s3Refs.Prompts.Turn1Prompt.Key == "" {
			t.Errorf("Turn1Prompt.Key should not be empty")
		}
		
		// Check if the key ends with turn1-prompt.json
		if !endsWithSubstring(s3Refs.Prompts.Turn1Prompt.Key, "turn1-prompt.json") {
			t.Errorf("Expected Turn1Prompt.Key to end with turn1-prompt.json, got %s", s3Refs.Prompts.Turn1Prompt.Key)
		}
		
		// 5. Assert S3 reference structure matches expected format
		// Check initialization reference
		if s3Refs.Initialization.Bucket == "" {
			t.Errorf("Initialization bucket should not be empty")
		}
		if !endsWithSubstring(s3Refs.Initialization.Key, "initialization.json") {
			t.Errorf("Expected Initialization.Key to end with initialization.json, got %s", s3Refs.Initialization.Key)
		}
		
		// Check images metadata reference
		if s3Refs.Images.Metadata.Bucket == "" {
			t.Errorf("Images.Metadata bucket should not be empty")
		}
		if !endsWithSubstring(s3Refs.Images.Metadata.Key, "images/metadata.json") {
			t.Errorf("Expected Images.Metadata.Key to end with images/metadata.json, got %s", s3Refs.Images.Metadata.Key)
		}
		
		// Check processing references
		if s3Refs.Processing.LayoutMetadata.Bucket == "" {
			t.Errorf("Processing.LayoutMetadata bucket should not be empty")
		}
		if !endsWithSubstring(s3Refs.Processing.LayoutMetadata.Key, "processing/layout-metadata.json") {
			t.Errorf("Expected Processing.LayoutMetadata.Key to end with processing/layout-metadata.json, got %s", s3Refs.Processing.LayoutMetadata.Key)
		}
		
		if s3Refs.Processing.HistoricalContext.Bucket == "" {
			t.Errorf("Processing.HistoricalContext bucket should not be empty")
		}
		if !endsWithSubstring(s3Refs.Processing.HistoricalContext.Key, "processing/historical-context.json") {
			t.Errorf("Expected Processing.HistoricalContext.Key to end with processing/historical-context.json, got %s", s3Refs.Processing.HistoricalContext.Key)
		}
		
		// Check response references
		if s3Refs.Responses.Turn1Raw.Key != rawRef.Key {
			t.Errorf("Expected Responses.Turn1Raw.Key to be %s, got %s", rawRef.Key, s3Refs.Responses.Turn1Raw.Key)
		}
		
		if s3Refs.Responses.Turn1Processed.Key != procRef.Key {
			t.Errorf("Expected Responses.Turn1Processed.Key to be %s, got %s", procRef.Key, s3Refs.Responses.Turn1Processed.Key)
		}
	}
	
	// 6. Assert summary fields are set correctly
	summary, ok := resp.ContextEnrichment["summary"].(ExecutionSummary)
	if !ok {
		t.Errorf("summary not found in contextEnrichment or not of type ExecutionSummary")
	} else {
		// Check summary fields
		if summary.AnalysisStage != "REFERENCE_ANALYSIS" {
			t.Errorf("Expected summary.AnalysisStage to be REFERENCE_ANALYSIS, got %s", summary.AnalysisStage)
		}
		
		if summary.VerificationType != req.VerificationContext.VerificationType {
			t.Errorf("Expected summary.VerificationType to be %s, got %s", req.VerificationContext.VerificationType, summary.VerificationType)
		}
		
		if summary.ProcessingTimeMs != 1000 {
			t.Errorf("Expected summary.ProcessingTimeMs to be 1000, got %d", summary.ProcessingTimeMs)
		}
		
		if summary.TokenUsage.Input != invoke.TokenUsage.InputTokens {
			t.Errorf("Expected summary.TokenUsage.Input to be %d, got %d", invoke.TokenUsage.InputTokens, summary.TokenUsage.Input)
		}
		
		if summary.TokenUsage.Output != invoke.TokenUsage.OutputTokens {
			t.Errorf("Expected summary.TokenUsage.Output to be %d, got %d", invoke.TokenUsage.OutputTokens, summary.TokenUsage.Output)
		}
		
		if summary.TokenUsage.Total != invoke.TokenUsage.TotalTokens {
			t.Errorf("Expected summary.TokenUsage.Total to be %d, got %d", invoke.TokenUsage.TotalTokens, summary.TokenUsage.Total)
		}
		
		if summary.BedrockRequestId != invoke.RequestID {
			t.Errorf("Expected summary.BedrockRequestId to be %s, got %s", invoke.RequestID, summary.BedrockRequestId)
		}
		
		if !summary.DynamodbUpdated {
			t.Errorf("Expected summary.DynamodbUpdated to be true, got false")
		}
		
		if !summary.ConversationTracked {
			t.Errorf("Expected summary.ConversationTracked to be true, got false")
		}
		
		if !summary.S3StorageCompleted {
			t.Errorf("Expected summary.S3StorageCompleted to be true, got false")
		}
	}
}

// Helper function to check if a string ends with a substring
func endsWithSubstring(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}

// Helper function to create test config for response_builder_test
func createResponseTestConfig() config.Config {
	cfg := config.Config{}
	
	cfg.AWS.Region = "us-west-2"
	cfg.AWS.S3Bucket = "test-bucket"
	cfg.AWS.BedrockModel = "anthropic.claude-v2"
	cfg.AWS.AnthropicVersion = "bedrock-2023-05-31"
	cfg.AWS.DynamoDBVerificationTable = "test-verification-table"
	cfg.AWS.DynamoDBConversationTable = "test-conversation-table"
	
	cfg.Processing.MaxTokens = 24000
	cfg.Processing.BudgetTokens = 16000
	cfg.Processing.ThinkingType = "enable"
	cfg.Processing.MaxRetries = 3
	cfg.Processing.BedrockConnectTimeoutSec = 10
	cfg.Processing.BedrockCallTimeoutSec = 30
	
	cfg.Logging.Level = "INFO"
	cfg.Logging.Format = "json"
	
	cfg.Prompts.TemplateVersion = "1.0"
	cfg.Prompts.TemplateBasePath = "/opt/templates"
	
	return cfg
}
